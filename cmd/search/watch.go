package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	// Modules
	notify "github.com/rjeczalik/notify"
)

func Watch(ctx context.Context, root string, ch chan<- Event) error {
	// Check path argument
	if stat, err := os.Stat(root); err != nil {
		return err
	} else if !stat.IsDir() {
		return fmt.Errorf("invalid path: %q", root)
	}

	// Set up a watchpoint listening for supported events within a folder
	c := make(chan notify.EventInfo, notifyChannelSize)
	root = filepath.Join(root, "...")
	if err := notify.Watch(root, c, notify.Create, notify.Remove, notify.Write, notify.Rename); err != nil {
		return err
	}
	defer notify.Stop(c)

	// Dispatch events to index files and folders
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			break FOR_LOOP
		case evt := <-c:
			// Calculate relative path
			info, err := os.Stat(evt.Path())
			if err != nil {
				log.Println(err)
			}
			switch evt.Event() {
			case notify.Create:
				ch <- Event{evt.Path(), info, EventTypeAdded}
			case notify.Remove:
				ch <- Event{evt.Path(), info, EventTypeRemoved}
			case notify.Rename:
				ch <- Event{evt.Path(), info, EventTypeRenamed}
			case notify.Write:
				ch <- Event{evt.Path(), info, EventTypeChanged}
			}
		}
	}

	// Return success
	return nil
}
