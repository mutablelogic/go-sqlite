//go:build !client

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
	tui "github.com/mutablelogic/go-server/pkg/tui"
	sqlite "github.com/mutablelogic/go-sqlite"
	filesearch "github.com/mutablelogic/go-sqlite/filesearch"
	schema "github.com/mutablelogic/go-sqlite/filesearch/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Commands are the commands of a file search manager, of its database
type Commands struct {
	Index   IndexCmd   `cmd:"" name:"index" help:"Index a folder, as a volume, of its files' metadata, content and artwork, and sweep what's gone." group:"FILESEARCH"`
	Search  SearchCmd  `cmd:"" name:"search" help:"Search the files which are indexed, of words, or list them, without any." group:"FILESEARCH"`
	Volumes VolumesCmd `cmd:"" name:"volumes" help:"List the volumes, with the status of their indexing." group:"FILESEARCH"`
}

// IndexCmd indexes a folder, as a volume
type IndexCmd struct {
	Flags     `embed:""`
	Path      string `arg:"" name:"folder" help:"The folder, which is a volume" type:"existingdir"`
	Recursive bool   `name:"recursive" help:"Index the folder's folders, as well as its files" default:"true" negatable:""`
}

// SearchCmd searches the files which are indexed
type SearchCmd struct {
	Flags  `embed:""`
	Words  []string `arg:"" name:"words" help:"Words, which are all in the files' text, as prefixes" optional:""`
	Volume []int64  `name:"volume" help:"The volumes to search, or all of them"`
	Offset uint64   `name:"offset" help:"The offset of the first file"`
	Limit  *uint64  `name:"limit" help:"The most files, which is at most 100"`
}

// VolumesCmd lists the volumes
type VolumesCmd struct {
	Flags  `embed:""`
	Offset uint64  `name:"offset" help:"The offset of the first volume"`
	Limit  *uint64 `name:"limit" help:"The most volumes, which is at most 100"`
}

///////////////////////////////////////////////////////////////////////////////
// COMMANDS

// Run indexes the folder, with its progress, on a terminal, and says the
// files which couldn't be read, and then what was indexed, marked and
// swept, once it's finished
func (cmd *IndexCmd) Run(ctx server.Cmd) error {
	return cmd.WithManager(ctx, nil, func(manager *filesearch.Manager) error {
		start := time.Now()
		progress := newIndexing(ctx.IsTerm())
		volume, err := manager.IndexVolume(ctx.Context(), schema.IndexVolumeRequest{Path: cmd.Path, Recursive: cmd.Recursive}, progress.Progress)
		progress.Close()
		if volume != nil {
			status := string(volume.Status)
			if errors.Is(err, context.Canceled) {
				status = "interrupted"
			}
			if p, err := manager.Progress(volume.ID); err == nil {
				fmt.Printf("%s: %s, of %d files: %d indexed, %d marked, %d swept, %d which couldn't be read, in %v\n",
					volume.Path, status, p.Total, p.Indexed, p.Marked, p.Swept, progress.Errors(), time.Since(start).Round(time.Millisecond))
			}
		}
		return err
	})
}

// Run searches the files, of the words, or lists them, without any
func (cmd *SearchCmd) Run(ctx server.Cmd) error {
	return cmd.WithManager(ctx, nil, func(manager *filesearch.Manager) error {
		req := schema.ObjectListRequest{
			OffsetLimit: sqlite.OffsetLimit{Offset: cmd.Offset, Limit: cmd.Limit},
			Volumes:     cmd.Volume,
		}
		if len(cmd.Words) > 0 {
			query := strings.Join(cmd.Words, " ")
			req.Query = &query
		}
		list, err := manager.ListObjects(ctx.Context(), req)
		if err != nil {
			return err
		}
		if ctx.IsDebug() {
			fmt.Println(list)
			return nil
		}
		table := tui.TableFor[schema.Object](tui.SetWidth(ctx.IsTerm()))
		if _, err := table.Write(os.Stdout, list.Body...); err != nil {
			return err
		}
		_, err = tui.TableSummary("files", uint(list.Count), uint(len(list.Body)), list.Offset, list.Limit).Write(os.Stdout)
		return err
	})
}

// Run lists the volumes
func (cmd *VolumesCmd) Run(ctx server.Cmd) error {
	return cmd.WithManager(ctx, nil, func(manager *filesearch.Manager) error {
		list, err := manager.ListVolumes(ctx.Context(), schema.VolumeListRequest{
			OffsetLimit: sqlite.OffsetLimit{Offset: cmd.Offset, Limit: cmd.Limit},
		})
		if err != nil {
			return err
		}
		if ctx.IsDebug() {
			fmt.Println(list)
			return nil
		}
		table := tui.TableFor[schema.Volume](tui.SetWidth(ctx.IsTerm()))
		if _, err := table.Write(os.Stdout, list.Body...); err != nil {
			return err
		}
		_, err = tui.TableSummary("volumes", uint(list.Count), uint(len(list.Body)), list.Offset, list.Limit).Write(os.Stdout)
		return err
	})
}
