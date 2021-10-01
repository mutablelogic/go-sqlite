package main

import (
	"context"
	"errors"
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
	"github.com/mutablelogic/go-sqlite/pkg/indexer"
	"github.com/mutablelogic/go-sqlite/pkg/sqlite3"

	// Namespace imports
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
)

var (
	flagName     = flag.String("name", "index", "Index name")
	flagInclude  = flag.String("include", "", "Paths, names and extensions to include")
	flagExclude  = flag.String("exclude", "", "Paths, names and extensions to exclude")
	flagWorkers  = flag.Uint("workers", 10, "Number of indexing workers")
	flagDatabase = flag.String("db", ":memory:", "Path to sqlite database")
)

func main() {
	var wg sync.WaitGroup
	ctx := HandleSignal()

	// Parse flags
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

	// Open indexer to path
	indexer, err := indexer.NewIndexer(*flagName, path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

	// Indexer inclusions
	for _, include := range strings.FieldsFunc(*flagInclude, sep) {
		if err := indexer.Include(include); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}
	}

	// Indexer exclusions
	for _, exclude := range strings.FieldsFunc(*flagExclude, sep) {
		if err := indexer.Exclude(exclude); err != nil {
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

	// Create schema
	if err := CreateSchema(ctx, pool); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}

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
		err := indexer.Walk(ctx, func(err error) {
			if err != nil {
				errs <- fmt.Errorf("reindexing completed with errors: %w", err)
			} else {
				errs <- fmt.Errorf("reindexing completed")
			}
		})
		if err != nil {
			errs <- fmt.Errorf("reindexing cannot start: %w", err)
		}
	}()

	for i := uint(0); i < *flagWorkers; i++ {
		wg.Add(1)
		go func(i uint) {
			defer wg.Done()
			conn := pool.Get(ctx)
			if conn == nil {
				return
			}
			defer pool.Put(conn)
			for {
				select {
				case <-ctx.Done():
					return
				default:
					if evt := indexer.Next(); evt != nil {
						if err := Process(ctx, conn, evt); err != nil {
							errs <- err
						}
					}
				}
			}
		}(i)
	}

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

func CreateSchema(ctx context.Context, pool SQPool) error {
	conn := pool.Get(ctx)
	if conn == nil {
		return errors.New("Unable to get a connection from pool")
	}
	defer pool.Put(conn)

	// Create table
	return conn.Do(ctx, 0, func(txn SQTransaction) error {
		if _, err := txn.Query(Q(`CREATE TABLE IF NOT EXISTS files (
			name       TEXT NOT NULL,
			path       TEXT NOT NULL,
			PRIMARY KEY (name, path)
		)`)); err != nil {
			return err
		}
		return nil
	})
}

func Process(ctx context.Context, conn SQConnection, evt *indexer.QueueEvent) error {
	return conn.Do(ctx, 0, func(txn SQTransaction) error {
		switch evt.Event {
		case indexer.EventAdd:
			if result, err := txn.Query(Q(`REPLACE INTO files (name, path) VALUES (?, ?)`), evt.Name, evt.Path); err != nil {
				return err
			} else if result.LastInsertId() > 0 {
				fmt.Println("ADDED:", evt)
			}
		case indexer.EventRemove:
			if result, err := txn.Query(Q(`DELETE FROM files WHERE name=? AND path=?`), evt.Name, evt.Path); err != nil {
				return err
			} else if result.RowsAffected() == 1 {
				fmt.Println("REMOVED:", evt)
			}
		}
		return nil
	})
}
