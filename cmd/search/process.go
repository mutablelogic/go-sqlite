package main

import (
	"context"
	"log"
	"path/filepath"
)

func Process(ctx context.Context, root string, ch <-chan Event) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case evt := <-ch:
			if relpath, err := filepath.Rel(root, evt.Path); err != nil {
				log.Println(err)
			} else if err := process(evt, relpath); err != nil {
				log.Println(err)
			}
		}
	}
}

func process(evt Event, relpath string) error {
	switch evt.Type {
	case EventTypeAdded:
		log.Println("Added", relpath)
	case EventTypeRemoved:
		log.Println("Removed", relpath)
	case EventTypeRenamed:
		log.Println("Renamed", relpath)
	case EventTypeChanged:
		log.Println("Changed", relpath)
	}
	return nil
}
