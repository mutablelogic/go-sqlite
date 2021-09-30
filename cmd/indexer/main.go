package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"

	"github.com/mutablelogic/go-sqlite/pkg/indexer"
)

var (
	flagInclude = flag.String("include", "", "Paths, names and extensions to include")
	flagExclude = flag.String("exclude", "", "Paths, names and extensions to exclude")
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "missing path argument")
		os.Exit(-1)
	}
	path := flag.Arg(0)
	if stat, err := os.Stat(path); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	} else if !stat.IsDir() {
		fmt.Fprintln(os.Stderr, "Not a directory")
		os.Exit(-1)
	}
	indexer, err := indexer.NewIndexer("indexer", path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// Inclusions
	for _, include := range strings.FieldsFunc(*flagInclude, sep) {
		if err := indexer.Include(include); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
	}

	// Exclusions
	for _, exclude := range strings.FieldsFunc(*flagExclude, sep) {
		if err := indexer.Exclude(exclude); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
	}

	var wg sync.WaitGroup

	ctx := HandleSignal()
	errs := make(chan error)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case err := <-errs:
				fmt.Fprintln(os.Stderr, err)
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := indexer.Run(ctx, errs); err != nil {
			errs <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-time.After(time.Second)
		if err := indexer.Walk(ctx); err != nil {
			errs <- err
		}
	}()

	// Wait for all goroutines to finish
	wg.Wait()
	os.Exit(0)
}

func HandleSignal() context.Context {
	// Handle signals - call cancel when interrupt received
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		cancel()
	}()
	return ctx
}

func sep(r rune) bool {
	return r == ',' || unicode.IsSpace(r)
}
