package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/hashicorp/go-multierror"
)

////////////////////////////////////////////////////////////////////////////////

var (
	flagDatabase = flag.String("db", "", "Filename for database")
)

////////////////////////////////////////////////////////////////////////////////

func main() {
	var root string

	flag.Parse()

	// Check arg
	if flag.NArg() > 1 {
		fmt.Fprintln(os.Stderr, "Expected one argument")
		os.Exit(-1)
	}

	// Get directory to search
	if flag.NArg() == 1 {
		root = flag.Arg(0)
	} else if wd, err := os.Getwd(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	} else {
		root = wd
	}
	fmt.Println("Watching for file changes:", root)

	// Open database
	db, err := OpenDatabase(*flagDatabase)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
	defer db.Close()

	// Wait until both go routines are completed
	var wg sync.WaitGroup
	var result error
	ch := make(chan Event)

	// Walk the directory tree
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := Walk(HandleSignal(), root, ch); err != nil {
			result = multierror.Append(result, err)
		}
	}()

	// Watch for changes
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := Watch(HandleSignal(), root, ch); err != nil {
			result = multierror.Append(result, err)
		}
	}()

	// Accept changes
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := Process(HandleSignal(), root, ch); err != nil {
			result = multierror.Append(result, err)
		}
	}()

	// Wait for all go routines to complete
	wg.Wait()
	if result != nil {
		fmt.Fprintln(os.Stderr, result)
		os.Exit(-1)
	}
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

// Handle signals - call cancel on returned context when interrupt received
func HandleSignal() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		<-ch
		cancel()
	}()
	return ctx
}
