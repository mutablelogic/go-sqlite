package sqlite

import (
	"fmt"
	"net/url"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type OffsetLimit struct {
	Offset uint64  `json:"offset,omitempty" help:"Offset within the result set" jsonschema:"Offset for pagination" example:"0"`
	Limit  *uint64 `json:"limit,omitempty" help:"Number of items to return" jsonschema:"Limit for pagination" example:"100"`
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Bind sets the offset and limit SQL fragment on the bind object.
func (r *OffsetLimit) Bind(bind *Bind, max uint64) {
	if (r.Limit != nil && *r.Limit > max) || r.Limit == nil {
		r.Limit = &max
	}
	if r.Offset != 0 && r.Limit != nil {
		bind.Set("offsetlimit", fmt.Sprintf("LIMIT %v OFFSET %v", *r.Limit, r.Offset))
	} else if r.Limit != nil {
		bind.Set("offsetlimit", fmt.Sprintf("LIMIT %v", *r.Limit))
	} else if r.Offset != 0 {
		bind.Set("offsetlimit", fmt.Sprintf("OFFSET %v", r.Offset))
	} else {
		bind.Set("offsetlimit", "")
	}
}

// Query sets offset and limit query parameters on the provided map.
func (r *OffsetLimit) Query(params url.Values) {
	if r.Limit != nil {
		params.Set("limit", fmt.Sprintf("%v", types.Value(r.Limit)))
	}
	if r.Offset != 0 {
		params.Set("offset", fmt.Sprintf("%v", r.Offset))
	}
}

// Clamp restricts the limit to the number of rows available after offset.
func (r *OffsetLimit) Clamp(total uint64) {
	if r.Limit == nil {
		return
	}

	// Available rows are constrained by total count and current offset.
	available := total
	if r.Offset >= total {
		available = 0
	} else {
		available = total - r.Offset
	}

	*r.Limit = min(*r.Limit, available)
}
