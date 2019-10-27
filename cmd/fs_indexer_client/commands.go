/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package main

import (
	"fmt"
	"os"
	"strconv"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	sq "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////

type Command struct {
	name string
	call func(*gopi.Flags, []string, sq.FSIndexerIndexClient, sq.FSIndexerQueryClient) error
}

////////////////////////////////////////////////////////////////////////////////

var (
	Commands = []Command{
		Command{"list", ListCommand},
		Command{"add", AddCommand},
		Command{"delete", DeleteCommand},
		Command{"query", QueryCommand},
	}
)

////////////////////////////////////////////////////////////////////////////////

func GetCommand(app *gopi.AppInstance) (*Command, []string) {
	if args := app.AppFlags.Args(); len(args) == 0 {
		// Return default command
		return &Commands[0], []string{}
	} else {
		// Find command
		for i := range Commands {
			if Commands[i].name == args[0] {
				return &Commands[i], args[1:]
			}
		}

		// Return not found
		return nil, nil
	}
}

func RunCommand(app *gopi.AppInstance, indexer sq.FSIndexerIndexClient, query sq.FSIndexerQueryClient) error {
	if command, args := GetCommand(app); command == nil {
		return gopi.ErrHelp
	} else if err := indexer.Ping(); err != nil {
		return err
	} else {
		return command.call(app.AppFlags, args, indexer, query)
	}
}

////////////////////////////////////////////////////////////////////////////////

func ListCommand(_ *gopi.Flags, args []string, indexer sq.FSIndexerIndexClient, _ sq.FSIndexerQueryClient) error {
	if len(args) != 0 {
		return fmt.Errorf("%w: Too many arguments", gopi.ErrBadParameter)
	}
	if jobs, err := indexer.List(); err != nil {
		return err
	} else {
		PrintIndexes(os.Stdout, jobs)
	}

	return nil
}

func AddCommand(_ *gopi.Flags, args []string, indexer sq.FSIndexerIndexClient, _ sq.FSIndexerQueryClient) error {
	if len(args) == 0 {
		return fmt.Errorf("%w: Missing index path", gopi.ErrBadParameter)
	}

	// Index each path
	indexes := make([]sq.FSIndex, 0, len(args))
	for _, path := range args {
		if index, err := indexer.AddIndex(path, false); err != nil {
			return fmt.Errorf("%w: AddIndex failed for path %v", err, strconv.Quote(path))
		} else {
			indexes = append(indexes, index)
		}
	}

	// Print out the indexes
	PrintIndexes(os.Stdout, indexes)

	// Return success
	return nil
}

func DeleteCommand(_ *gopi.Flags, args []string, indexer sq.FSIndexerIndexClient, _ sq.FSIndexerQueryClient) error {
	if len(args) == 0 {
		return fmt.Errorf("%w: Missing index path", gopi.ErrBadParameter)
	}

	// Delete each index
	for _, arg := range args {
		if id, err := strconv.ParseInt(arg, 10, 64); err != nil {
			return fmt.Errorf("%w: Invalid index %v", gopi.ErrBadParameter, strconv.Quote(arg))
		} else if err := indexer.DeleteIndex(id); err != nil {
			return fmt.Errorf("%w: DeleteIndex failed for %v", err, strconv.Quote(arg))
		}
	}

	// Return success
	return nil
}

func QueryCommand(flags *gopi.Flags, args []string, _ sq.FSIndexerIndexClient, query sq.FSIndexerQueryClient) error {
	if len(args) != 0 {
		return fmt.Errorf("%w: Too many arguments", gopi.ErrBadParameter)
	}
	limit, _ := flags.GetUint("limit")
	if response, err := query.Query(uint64(limit)); err != nil {
		return err
	} else {
		PrintFiles(os.Stdout, response.Files())
		fmt.Printf("%v of %v files\n", len(response.Files()), response.Count())
	}

	// Return success
	return nil
}
