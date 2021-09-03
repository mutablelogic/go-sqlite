package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	// Modules
	indexer "github.com/djthorpe/go-sqlite/pkg/indexer"
	sqlite "github.com/djthorpe/go-sqlite/pkg/sqlite"

	// Import namespaces
	. "github.com/djthorpe/go-sqlite"
)

var (
	flagDatabase = flag.String("db", "", "Database file")
	flagReset    = flag.Bool("reset", false, "Reset database schema")
	flagReindex  = flag.Bool("reindex", true, "Reindex database")
	flagExclude  = flag.String("exclude", "", "Path and file extension exclusions")
)

func main() {
	flag.Parse()

	path, err := GetPath()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// Make database connection
	conn, err := sqlite.Open(*flagDatabase, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
	defer conn.Close()

	// Create index manager with main schema
	flags := SQLITE_FLAG_NONE
	if *flagReset {
		flags |= SQLITE_FLAG_DELETEIFEXISTS
	}
	idx, err := indexer.NewManager(conn, "main", flags)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// Handle signal
	ctx := HandleSignal()

	// Run indexer in background
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Running ", idx)
		if err := idx.Run(ctx, RenderFile); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	// Make a new indexer
	main, err := idx.NewIndexer("main", path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// Add exclude paths and extensions
	for _, exclude := range strings.Fields(*flagExclude) {
		if err := idx.Exclude(main, exclude); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
	}

	// Kick off a reindex, which is cancelled if ctx is cancelled
	if *flagReindex {
		time.Sleep(time.Second)
		if err := idx.Reindex(main); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
	}

	// Wait for end of goroutines
	wg.Wait()

	// Shutdown
	fmt.Println("Shutdown")
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

func GetPath() (string, error) {
	if flag.NArg() > 1 {
		return "", fmt.Errorf("usage: %s (<path>)", filepath.Base(flag.CommandLine.Name()))
	}
	if flag.NArg() == 1 {
		return flag.Arg(0), nil
	}
	return os.Getwd()
}

func RenderFile(ctx context.Context, evt indexer.IndexerEvent) error {
	fmt.Println("TODO: ", evt)
	return nil
}
