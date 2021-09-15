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
	importer "github.com/djthorpe/go-sqlite/pkg/importer"
	sqlite3 "github.com/djthorpe/go-sqlite/sys/sqlite3"

	// Namespace Imports
	. "github.com/djthorpe/go-sqlite"
)

var (
	flagOverwrite = flag.Bool("overwrite", false, "Overwrite existing tables")
	flagQuiet     = flag.Bool("quiet", false, "Suppress non-error output")
	flagHeader    = flag.Bool("header", true, "CSV contains header row")
	flagDelimiter = flag.String("delimiter", "", "Field delimiter")
	flagComment   = flag.String("comment", "#", "Comment character")
	flagTrimSpace = flag.Bool("trimspace", true, "Trim leading space of a field")
)

////////////////////////////////////////////////////////////////////////////////

func main() {
	flag.Parse()

	// Name of tool, logger
	name := filepath.Base(flag.CommandLine.Name())
	log := logger(name + " ")

	// Check number of arguments
	if flag.NArg() < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %v <sqlite-database> <url>...\n", name)
		os.Exit(1)
	}

	// Open database
	db, err := sqlite3.OpenPathEx(flag.Arg(0), sqlite3.SQLITE_OPEN_CREATE, "")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer db.Close()

	// Report on the database
	log.Println("database:", db.Filename(sqlite3.DefaultSchema))

	// Create a configuration
	config := SQImportConfig{
		Header:    *flagHeader,
		TrimSpace: *flagTrimSpace,
		Overwrite: *flagOverwrite,
	}
	if *flagDelimiter != "" {
		config.Delimiter = rune((*flagDelimiter)[0])
	}
	if *flagComment != "" {
		config.Comment = rune((*flagComment)[0])
	}

	// Create an SQL Writer
	writer, err := importer.NewSQLWriter(config, db)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Read files
	for _, url := range flag.Args()[1:] {
		// Create an importer
		importer, err := importer.NewImporter(config, url)
		if err != nil {
			fmt.Fprintln(os.Stderr, importer.URL(), ": ", err)
			continue
		}

		// Create the decoder
		decoder, err := importer.Decoder("")
		if err != nil {
			fmt.Fprintln(os.Stderr, importer.URL(), ": ", err)
			continue
		}
		defer decoder.Close()

		// Reset the counter
		log.Println(" import:", importer.URL())
		log.Println("     ...decoder", decoder)

		// Call Begin for writer to get writing function
		fn, err := writer.Begin(importer.Name(), sqlite3.DefaultSchema, []string{"continent"})
		if err != nil {
			fmt.Fprintln(os.Stderr, importer.URL(), ": ", err)
			continue
		}

		// Read and write rows
		start, mark := time.Now(), time.Now()
		for {
			if err := importer.ReadWrite(decoder, fn); err == io.EOF {
				writer.End(true) // commit
				break
			} else if err != nil {
				writer.End(false) // rollback
				fmt.Fprintln(os.Stderr, importer.URL(), ": ", err)
				break
			}
			if time.Since(mark) > 5*time.Second {
				log.Printf("     ...written %d rows", 0)
				mark = time.Now()
			}
		}

		// Report
		since := time.Since(start)
		//ops_per_sec := math.Round(float64(writer.Count()) * 1000 / float64(since.Milliseconds()))
		log.Printf("     ...written %d rows in %v (%.0f ops/s)", 0, since.Truncate(time.Millisecond), 0)
	}
}

func logger(name string) *log.Logger {
	if *flagQuiet {
		return log.New(io.Discard, name, 0)
	} else {
		return log.New(os.Stderr, name, 0)
	}
}
