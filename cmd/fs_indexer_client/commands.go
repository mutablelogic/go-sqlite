/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package main

import (
	"fmt"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	sq "github.com/djthorpe/sqlite"
)

////////////////////////////////////////////////////////////////////////////////

type Command struct {
	name string
	call func([]string, sq.FSIndexerIndexClient) error
}

////////////////////////////////////////////////////////////////////////////////

var (
	Commands = []Command{
		Command{"list", ListCommand},
		Command{"add", IndexCommand},
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

func RunCommand(app *gopi.AppInstance, indexer sq.FSIndexerIndexClient) error {
	if command, args := GetCommand(app); command == nil {
		return gopi.ErrHelp
	} else if err := indexer.Ping(); err != nil {
		return err
	} else {
		return command.call(args, indexer)
	}
}

////////////////////////////////////////////////////////////////////////////////

func ListCommand(args []string, indexer sq.FSIndexerIndexClient) error {
	if len(args) != 0 {
		return fmt.Errorf("%w: Too many arguments", gopi.ErrBadParameter)
	}
	if jobs, err := indexer.List(); err != nil {
		return err
	} else {
		fmt.Println("LIST", jobs)
	}

	return nil
}

func IndexCommand(args []string, indexer sq.FSIndexerIndexClient) error {
	if len(args) == 0 {
		return fmt.Errorf("%w: Missing index path", gopi.ErrBadParameter)
	}
	// Index each path
	for _, path := range args {
		if index, err := indexer.Index(path, false); err != nil {
			return err
		} else {
			fmt.Println(path, "=>", index)
		}
	}
	// Return success
	return nil
}
