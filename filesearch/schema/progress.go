package schema

import (
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Progress is how far a run of a volume has indexed it, while it's
// indexing, or how it did, once it's finished
type Progress struct {
	Volume  int64     `json:"volume"`
	Path    string    `json:"path"`              // The volume's path
	Run     int64     `json:"run"`               // The run which is indexing it
	Started time.Time `json:"started"`           // When the run started
	Current string    `json:"current,omitempty"` // The file which is being indexed, within the volume
	Files   uint64    `json:"files"`             // The files which were seen
	Indexed uint64    `json:"indexed"`           // Of which were new, or changed, and were indexed
	Marked  uint64    `json:"marked"`            // Of which were as they were, and were marked
	Swept   uint64    `json:"swept"`             // The objects which weren't seen, which were swept, once it's indexed

	// When it finished, or nil while it's indexing
	Finished *time.Time `json:"finished,omitempty"`
}
