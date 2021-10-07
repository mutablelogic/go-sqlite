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

	// Packages
	"github.com/mutablelogic/go-sqlite/pkg/config"
	"github.com/mutablelogic/go-sqlite/pkg/indexer"
	"github.com/mutablelogic/go-sqlite/pkg/sqlite3"
)

var (
	flagName     = flag.String("name", "index", "Index name")
	flagInclude  = flag.String("include", "", "Paths, names and extensions to include")
	flagExclude  = flag.String("exclude", "", "Paths, names and extensions to exclude")
	flagWorkers  = flag.Uint("workers", 0, "Number of indexing workers")
	flagDatabase = flag.String("db", ":memory:", "Path to sqlite database")
	flagVersion  = flag.Bool("version", false, "Display version")
)

func main() {
	var wg sync.WaitGroup
	ctx := HandleSignal()

	// Parse flags
	flag.Parse()
	if *flagVersion {
		config.PrintVersion(flag.CommandLine.Output())
		os.Exit(0)
	}
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

	// Open indexer to path
	idx, err := indexer.NewIndexer(*flagName, path, indexer.NewQueue())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// Indexer inclusions
	for _, include := range strings.FieldsFunc(*flagInclude, sep) {
		if err := idx.Include(include); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
	}

	// Indexer exclusions
	for _, exclude := range strings.FieldsFunc(*flagExclude, sep) {
		if err := idx.Exclude(exclude); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
	}

	// Create database pool
	errs := make(chan error)
	pool, err := sqlite3.OpenPool(sqlite3.PoolConfig{
		Max:    int32(*flagWorkers),
		Create: true,
		Schemas: map[string]string{
			"main": *flagDatabase,
		},
	}, errs)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
	defer pool.Close()

	// Create store
	store := indexer.NewStore(pool, "main", idx.Queue(), *flagWorkers)
	if store == nil {
		fmt.Fprintln(os.Stderr, "failed to create store")
		os.Exit(-1)
	}

	// Error routine persists until error channel is closed
	go func() {
		for err := range errs {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := idx.Run(ctx, errs); err != nil {
			errs <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := store.Run(ctx, errs); err != nil {
			errs <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-time.After(time.Second)
		err := idx.Walk(ctx, func(err error) {
			if err != nil {
				errs <- fmt.Errorf("reindexing completed with errors: %w", err)
			}
		})
		if err != nil {
			errs <- fmt.Errorf("reindexing cannot start: %w", err)
		}
	}()

	// Wait for all goroutines to finish
	wg.Wait()

	// Close error channel
	close(errs)
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
