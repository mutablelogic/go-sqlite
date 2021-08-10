package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	// Modules
	sq "github.com/djthorpe/go-sqlite"
	sqlite "github.com/djthorpe/go-sqlite/pkg/sqlite"
	multierror "github.com/hashicorp/go-multierror"
)

var (
	flagLocation  = flag.String("tz", "Local", "Timezone name")
	flagOverwrite = flag.Bool("overwrite", false, "Overwrite existing tables")
	flagQuiet     = flag.Bool("quiet", false, "Suppress output")
	flagScan      = flag.Bool("scan", true, "Adjust data types for columns")
	//flagSeparator = flag.String("separator", "", "Field separator")
	//flagComment   = flag.String("comment", "#", "Comment character")
)

////////////////////////////////////////////////////////////////////////////////

func main() {
	flag.Parse()

	// Check number of arguments
	if flag.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "Usage: sqimport <sqlite-database> <file>...")
		os.Exit(1)
	}

	// Create log
	log := logger(filepath.Base(flag.CommandLine.Name()) + " ")

	// Load location
	loc, err := time.LoadLocation(*flagLocation)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else {
		log.Println("timezone:", loc)
	}

	// Open database
	db, err := sqlite.Open(flag.Arg(0), loc)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer db.Close()

	// Read files
	var result error
	for _, arg := range flag.Args()[1:] {
		// Create a table writer
		writer := NewWriter(db)
		if *flagOverwrite {
			writer.Overwrite = true
		}
		log.Println("writer:", writer)

		// Create a table reader
		table, err := NewTable(arg, writer)
		if err != nil {
			result = multierror.Append(result, err)
		}

		// Read in data
		if err := read(db, table, log); err != nil {
			result = multierror.Append(result, err)
		}
		// Scan and adjust data types
		if err := scan(db, table); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Print import errors
	if result != nil {
		fmt.Fprintln(os.Stderr, result)
	}
}

func logger(name string) *log.Logger {
	if *flagQuiet {
		return log.New(io.Discard, name, 0)
	} else {
		return log.New(os.Stderr, name, 0)
	}
}

func read(db sq.SQConnection, table *table, log *log.Logger) error {
	l := false
	for {
		err := table.Read(db)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if !l {
			log.Println("table: ", table)
			l = true
		}
	}
}

func scan(db sq.SQConnection, table *table) error {
	return table.Scan(db)
}
