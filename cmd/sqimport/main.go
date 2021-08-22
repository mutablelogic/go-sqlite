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
	. "github.com/djthorpe/go-sqlite"
	sqimport "github.com/djthorpe/go-sqlite/pkg/sqimport"
	sqlite "github.com/djthorpe/go-sqlite/pkg/sqlite"
)

var (
	flagLocation  = flag.String("tz", "Local", "Timezone name")
	flagOverwrite = flag.Bool("overwrite", false, "Overwrite existing tables")
	flagQuiet     = flag.Bool("quiet", false, "Suppress output")
	flagHeader    = flag.Bool("header", true, "CSV contains header row")
	flagDelimiter = flag.String("delimiter", "", "Field delimiter")
	flagComment   = flag.String("comment", "#", "Comment character")
	flagTrimSpace = flag.Bool("trimspace", true, "Trim leading space of a field")
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

	// Create a configuration
	config := SQImportConfig{
		Header:    *flagHeader,
		TrimSpace: *flagTrimSpace,
	}
	if *flagDelimiter != "" {
		config.Delimiter = rune((*flagDelimiter)[0])
	}
	if *flagComment != "" {
		config.Comment = rune((*flagComment)[0])
	}

	// Create an SQL Writer
	writer, err := sqimport.NewSQLWriter(config, db)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Read files
	for _, url := range flag.Args()[1:] {
		// Create an importer
		importer, err := sqimport.NewImporter(config, url, writer)
		if err != nil {
			fmt.Fprintln(os.Stderr, importer.URL(), ": ", err)
			continue
		}
		for {
			if err := importer.Read(); err == io.EOF {
				break
			} else if err != nil {
				fmt.Fprintln(os.Stderr, importer.URL(), ": ", err)
				break
			}
		}
	}
}

func logger(name string) *log.Logger {
	if *flagQuiet {
		return log.New(io.Discard, name, 0)
	} else {
		return log.New(os.Stderr, name, 0)
	}
}
