package main

import (
	"context"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
)

func Walk(ctx context.Context, root string, ch chan<- Event) error {
	return filepath.WalkDir(root, func(path string, file fs.DirEntry, err error) error {
		// Propogate errors
		if err != nil {
			return err
		}
		// Ignore hidden files and folders
		if strings.HasPrefix(file.Name(), ".") {
			if file.IsDir() {
				return filepath.SkipDir
			}
			return err
		}
		// Emit event
		if info, err := file.Info(); err != nil {
			log.Println(err)
		} else {
			ch <- Event{path, info, EventTypeAdded}
		}
		// Return context error
		return ctx.Err()
	})
}
