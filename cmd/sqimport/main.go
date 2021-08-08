package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	// Modules
	sq "github.com/djthorpe/go-sqlite"
	sqlite "github.com/djthorpe/go-sqlite/pkg/sqlite"
	multierror "github.com/hashicorp/go-multierror"
)

var (
	flagLocation  = flag.String("tz", "Local", "Timezone name")
	flagOverwrite = flag.Bool("overwrite", false, "Overwrite existing tables")
	flagSeparator = flag.String("separator", "", "Field separator")
	flagComment   = flag.String("comment", "#", "Comment character")
)

////////////////////////////////////////////////////////////////////////////////

func main() {
	flag.Parse()

	// Check number of arguments
	if flag.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "Usage: sqimport <sqlite-database> <file>...")
		os.Exit(1)
	}

	// Load location
	loc, err := time.LoadLocation(*flagLocation)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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
		table, err := NewTable(arg)
		if err != nil {
			result = multierror.Append(result, err)
		}
		if err := read(db, table); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Print import errors
	if result != nil {
		fmt.Fprintln(os.Stderr, result)
	}
}

func read(db sq.SQConnection, table *table) error {
	for {
		row, err := table.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if row != nil {
			fmt.Println("INSERT ROW=", row)
		}
	}
}
