package main

import (
	"C"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
	"unsafe"

	// Modules
	importer "github.com/djthorpe/go-sqlite/pkg/importer"
	sqlite3 "github.com/djthorpe/go-sqlite/sys/sqlite3"

	// Namespace Imports
	. "github.com/djthorpe/go-sqlite"
)
import (
	"errors"
	"math"
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

	//db.SetTraceHook(trace, sqlite3.SQLITE_TRACE_PROFILE)

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
		importer, err := importer.NewImporter(config, url, writer)
		if err != nil {
			fmt.Fprintln(os.Stderr, importer.URL(), ": ", err)
			continue
		}

		// Create the decoder, guess mimetype instead of supplying it
		decoder, err := importer.Decoder("")
		if err != nil {
			fmt.Fprintln(os.Stderr, importer.URL(), ": ", err)
			continue
		}

		// Reset the counter
		log.Println(" import:", importer.URL())
		log.Println("     ...decoder", decoder)

		// Read and write rows
		start, mark := time.Now(), time.Now()
		for {
			if err := importer.ReadWrite(decoder); err == io.EOF {
				break
			} else if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				fmt.Fprintln(os.Stderr, importer.URL(), ": ", err)
				break
			}
			if time.Since(mark) > 5*time.Second {
				log.Printf("     ...written %d rows", writer.Count())
				mark = time.Now()
			}
		}

		// Report
		since := time.Since(start)
		ops_per_sec := math.Round(float64(writer.Count()) * 1000 / float64(since.Milliseconds()))
		log.Printf("     ...written %d rows in %v (%.0f ops/s)", writer.Count(), since.Truncate(time.Millisecond), ops_per_sec)
	}
}

func logger(name string) *log.Logger {
	if *flagQuiet {
		return log.New(io.Discard, name, 0)
	} else {
		return log.New(os.Stderr, name, 0)
	}
}
func trace(t sqlite3.TraceType, a, b unsafe.Pointer) int {
	switch t {
	case sqlite3.SQLITE_TRACE_STMT:
		fmt.Println("STMT => ", (*sqlite3.Statement)(a), C.GoString((*C.char)(b)))
	case sqlite3.SQLITE_TRACE_PROFILE:
		ms := time.Duration(time.Duration(*(*int64)(b)) * time.Nanosecond)
		fmt.Println("PROF => ", (*sqlite3.Statement)(a), ms)
	case sqlite3.SQLITE_TRACE_ROW:
		fmt.Println("ROW  => ", (*sqlite3.Statement)(a))
	case sqlite3.SQLITE_TRACE_CLOSE:
		fmt.Println("CLSE => ", (*sqlite3.Conn)(a))
	}
	return 0
}
