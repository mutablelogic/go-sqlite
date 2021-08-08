/*
	SQLite client
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package main

import (
	"fmt"
	"io"

	// Frameworks
	sq "github.com/djthorpe/sqlite"
	tablewriter "github.com/olekukonko/tablewriter"
)

////////////////////////////////////////////////////////////////////////////////

func PrintIndexes(fh io.Writer, list []sq.FSIndex) {
	table := tablewriter.NewWriter(fh)
	table.SetHeader([]string{"id", "path", "count", "status"})
	for _, index := range list {
		table.Append([]string{
			fmt.Sprint(index.Id()),
			index.Name(),
			fmt.Sprint(index.Count()),
			fmt.Sprint(index.Status()),
		})
	}
	table.Render()
}

func PrintFiles(fh io.Writer, list []sq.FSFile) {
	table := tablewriter.NewWriter(fh)
	table.SetHeader([]string{"id", "index", "path", "mimetype", "size"})
	for _, file := range list {
		table.Append([]string{
			fmt.Sprint(file.Id()),
			"TODO",
			file.Path(),
			file.MimeType(),
			fmt.Sprint(file.Size()),
		})
	}
	table.Render()
}
