package schema

import (
	"fmt"
	"strconv"
	"time"

	// Packages
	tui "github.com/mutablelogic/go-server/pkg/tui"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (o ObjectList) String() string {
	return types.Stringify(o)
}

func (v Volume) String() string {
	return types.Stringify(v)
}

func (v VolumeList) String() string {
	return types.Stringify(v)
}

////////////////////////////////////////////////////////////////////////////////
// TABLE

// Header returns the columns of an object, as a table's row
func (o Object) Header() []string {
	return []string{"Volume", "Path", "Title", "Kind", "Type", "Size", "Modified", "Rank"}
}

// Width returns the width of a column, which is its own
func (o Object) Width(int) int {
	return 0
}

// Cell returns the column of an object
func (o Object) Cell(col int) tui.CellText {
	switch col {
	case 0:
		return o.Volume
	case 1:
		return o.Path
	case 2:
		return o.Title
	case 3:
		return o.Kind
	case 4:
		return o.Type
	case 5:
		return FormatSize(o.Size)
	case 6:
		return o.ModTime.Local().Format(time.DateTime)
	case 7:
		if o.Rank == nil {
			return ""
		}
		return strconv.FormatFloat(*o.Rank, 'g', 3, 64)
	}
	return ""
}

// Header returns the columns of a volume, as a table's row
func (v Volume) Header() []string {
	return []string{"ID", "Path", "Objects", "Status", "Run", "Started", "Finished", "Error"}
}

// Width returns the width of a column, which is its own
func (v Volume) Width(int) int {
	return 0
}

// Cell returns the column of a volume
func (v Volume) Cell(col int) tui.CellText {
	switch col {
	case 0:
		return v.ID
	case 1:
		return v.Path
	case 2:
		return v.Objects
	case 3:
		return string(v.Status)
	case 4:
		return v.Run
	case 5:
		return formatTime(v.Started)
	case 6:
		return formatTime(v.Finished)
	case 7:
		return v.Error
	}
	return ""
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// FormatSize returns a size in bytes, or in KiB, MiB, GiB or TiB, with one
// decimal place, when it's larger
func FormatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit && exp < 3; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(size)/float64(div), "KMGT"[exp])
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// formatTime returns a time, in the local time zone, or nothing
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Local().Format(time.DateTime)
}
