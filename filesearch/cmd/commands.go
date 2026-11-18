//go:build !client

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
	tui "github.com/mutablelogic/go-server/pkg/tui"
	types "github.com/mutablelogic/go-server/pkg/types"
	sqlite "github.com/mutablelogic/go-sqlite"
	filesearch "github.com/mutablelogic/go-sqlite/filesearch"
	schema "github.com/mutablelogic/go-sqlite/filesearch/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Commands are the commands of a file search manager, of its database
type Commands struct {
	Index        IndexCmd        `cmd:"" name:"index" help:"Index a folder, as a volume, of its files' metadata, content and artwork, and sweep what's gone." group:"FILESEARCH"`
	Search       SearchCmd       `cmd:"" name:"search" help:"Search the files which are indexed, of words, or list them, without any." group:"FILESEARCH"`
	Volumes      VolumesCmd      `cmd:"" name:"volumes" help:"List the volumes, with the status of their indexing." group:"FILESEARCH"`
	VolumeDelete VolumeDeleteCmd `cmd:"" name:"volume-delete" help:"Delete a volume, by its id, or path, with what's indexed of it." group:"FILESEARCH"`
	Info         InfoCmd         `cmd:"" name:"info" help:"Show what's indexed of a file, of its metadata." group:"FILESEARCH"`
	Cat          CatCmd          `cmd:"" name:"cat" help:"Write a file's content, as Markdown, or its artwork." group:"FILESEARCH"`
}

// VolumeDeleteCmd deletes a volume
type VolumeDeleteCmd struct {
	Flags  `embed:""`
	Volume string `arg:"" name:"volume" help:"The volume's id, or path"`
}

// InfoCmd shows what's indexed of a file
type InfoCmd struct {
	Flags `embed:""`
	Path  string `arg:"" name:"path" help:"The file's path, or the end of it, such as its name, which is of one file"`
}

// CatCmd writes a file's content, as Markdown, or its artwork
type CatCmd struct {
	Flags   `embed:""`
	Path    string `arg:"" name:"path" help:"The file's path, or the end of it, such as its name, which is of one file"`
	Artwork bool   `name:"artwork" help:"Its artwork, rather than its content"`
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
	Type   []string `name:"type" help:"Files of any of the types: a kind, or the start, or end, of a content type, such as video, or pdf, or, with a /, a content type, such as video/mp4"`
	Order  string   `name:"order" help:"rank, the best first, which is a search's default, date, the newest first, which is a list's, path, size, the largest first, or title" enum:",rank,date,path,size,title" default:""`
	Cutoff float64  `name:"cutoff" help:"The lowest rank, of a search, between 0 and 1, which is higher the better a file matches, or 0, for none" default:"0.15"`
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
			Types:       cmd.Type,
			Order:       schema.Order(cmd.Order),
		}
		if len(cmd.Words) > 0 {
			query := strings.Join(cmd.Words, " ")
			req.Query = &query
			req.Cutoff = &cmd.Cutoff
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

// Run deletes the volume, by its id, or path
func (cmd *VolumeDeleteCmd) Run(ctx server.Cmd) error {
	return cmd.WithManager(ctx, nil, func(manager *filesearch.Manager) error {
		id, err := strconv.ParseInt(cmd.Volume, 10, 64)
		if err != nil {
			volume, err := manager.VolumeByPath(ctx.Context(), cmd.Volume)
			if err != nil {
				return err
			}
			id = volume.ID
		}
		volume, err := manager.DeleteVolume(ctx.Context(), id)
		if err != nil {
			return err
		}
		fmt.Printf("Deleted %s, of %d files\n", volume.Path, volume.Objects)
		return nil
	})
}

// Run shows what's indexed of the file, as a table of its fields, and its
// metadata, or JSON, with --debug
func (cmd *InfoCmd) Run(ctx server.Cmd) error {
	return cmd.WithManager(ctx, nil, func(manager *filesearch.Manager) error {
		object, err := manager.GetObject(ctx.Context(), cmd.Path)
		if err != nil {
			return err
		}
		if ctx.IsDebug() {
			fmt.Println(types.Stringify(object))
			return nil
		}
		path, err := manager.ObjectPath(ctx.Context(), object)
		if err != nil {
			return err
		}
		rows := []info{
			{"Path", path},
			{"Title", object.Title},
			{"Kind", object.Kind},
			{"Type", object.Type},
			{"Size", schema.FormatSize(object.Size)},
			{"Modified", object.ModTime.Local().Format(time.DateTime)},
			{"Indexed", object.Indexed.Local().Format(time.DateTime)},
		}
		keys := make([]string, 0, len(object.Meta))
		for key := range object.Meta {
			keys = append(keys, key)
		}
		slices.Sort(keys)
		for _, key := range keys {
			rows = append(rows, info{key, infoValue(object.Meta[key])})
		}
		_, err = tui.TableFor[info](tui.SetWidth(ctx.IsTerm())).Write(os.Stdout, rows...)
		return err
	})
}

// Run writes the file's content, as Markdown, which is formatted, on a
// terminal, or its artwork, which is drawn, on a terminal, or else is the
// image's bytes
func (cmd *CatCmd) Run(ctx server.Cmd) error {
	return cmd.WithManager(ctx, nil, func(manager *filesearch.Manager) error {
		object, err := manager.GetObject(ctx.Context(), cmd.Path)
		if err != nil {
			return err
		}
		width := ctx.IsTerm()

		// Its artwork, which is drawn, or written
		if cmd.Artwork {
			artwork, err := manager.Artwork(ctx.Context(), object.ID)
			if errors.Is(err, sqlite.ErrNotFound) {
				return sqlite.ErrNotFound.Withf("%s has no artwork", object.Path)
			} else if err != nil {
				return err
			}
			if width <= 0 {
				_, err := os.Stdout.Write(artwork)
				return err
			}
			img, _, err := image.Decode(bytes.NewReader(artwork))
			if err != nil {
				return err
			}
			_, err = tui.Image(tui.SetWidth(width)).Write(os.Stdout, img)
			return err
		}

		// Its content, as Markdown, which is formatted, or written
		markdown, err := manager.Markdown(ctx.Context(), object.ID)
		if err != nil {
			return err
		}
		if markdown == "" {
			return sqlite.ErrNotFound.Withf("%s has no content", object.Path)
		}
		if width <= 0 {
			_, err := fmt.Fprint(os.Stdout, markdown)
			return err
		}
		_, err = tui.MarkdownFormatter(tui.SetWidth(width)).Write(os.Stdout, markdown)
		return err
	})
}

///////////////////////////////////////////////////////////////////////////////
// INFO

// info is a field of a file, as a table's row
type info struct {
	Key, Value string
}

func (info) Header() []string { return []string{"Key", "Value"} }
func (info) Width(int) int    { return 0 }
func (i info) Cell(col int) tui.CellText {
	if col == 0 {
		return i.Key
	}
	return i.Value
}

// infoValue returns a value of metadata, as text: a list, of its values,
// and others, of their JSON
func infoValue(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case []any:
		values := make([]string, 0, len(v))
		for _, v := range v {
			values = append(values, infoValue(v))
		}
		return strings.Join(values, ", ")
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprint(v)
	}
	return string(data)
}
