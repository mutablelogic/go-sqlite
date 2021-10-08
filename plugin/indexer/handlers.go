package main

import (
	"context"
	"net/http"
	"regexp"

	// Packages
	router "github.com/mutablelogic/go-server/pkg/httprouter"
	indexer "github.com/mutablelogic/go-sqlite/pkg/indexer"
	version "github.com/mutablelogic/go-sqlite/pkg/version"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type PingResponse struct {
	Version map[string]string `json:"version"`
	Indexes []IndexResponse   `json:"indexes"`
}

type IndexResponse struct {
	Name    string      `json:"name"`
	Path    string      `json:"path,omitempty"`
	Count   int64       `json:"count,omitempty"`
	Modtime interface{} `json:"reindexed,omitempty"`
	Status  string      `json:"status,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// ROUTES

var (
	reRoutePing = regexp.MustCompile(`^/?$`)
)

///////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	maxResultLimit = 1000
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (p *plugin) AddHandlers(ctx context.Context, provider Provider) error {
	// Add handler for ping
	if err := provider.AddHandlerFuncEx(ctx, reRoutePing, p.ServePing); err != nil {
		return err
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// HANDLERS

func (p *plugin) ServePing(w http.ResponseWriter, req *http.Request) {
	// Get a connection
	conn := p.pool.Get()
	if conn == nil {
		router.ServeError(w, http.StatusBadGateway, "No connection")
		return
	}
	defer p.pool.Put(conn)

	// Retrieve indexes with count of documents in each
	index, err := indexer.ListIndexWithCount(req.Context(), conn, p.store.Schema())
	if err != nil {
		router.ServeError(w, http.StatusBadGateway, err.Error())
		return
	}

	// Add known indexes to the response - these may not yet have any rows in the
	// database
	for _, idx := range p.index {
		name := idx.Name()
		if _, exists := index[name]; !exists {
			index[name] = 0
		}
	}

	// Populate response
	response := PingResponse{
		Version: version.Version(),
		Indexes: make([]IndexResponse, 0, len(index)),
	}

	// Add all indexes into the response, adding their modtime and
	// status
	for name, count := range index {
		response.Indexes = append(response.Indexes, IndexResponse{
			Name:    name,
			Count:   count,
			Path:    p.pathForIndex(name),
			Modtime: p.modtimeForIndex(name),
			Status:  p.statusForIndex(name),
		})
	}

	// Serve response
	router.ServeJSON(w, response, http.StatusOK, 2)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (p *plugin) pathForIndex(name string) string {
	if idx, exists := p.index[name]; exists {
		return idx.Path()
	} else {
		return ""
	}
}

func (p *plugin) modtimeForIndex(name string) interface{} {
	if t, exists := p.modtime[name]; exists && t.IsZero() == false {
		return t
	} else {
		return nil
	}
}

func (p *plugin) statusForIndex(name string) string {
	if idx, exists := p.index[name]; !exists {
		return ""
	} else if idx.IsIndexing() {
		return "indexing"
	} else if t, exists := p.modtime[name]; exists && t.IsZero() == false {
		return "ready"
	} else {
		return "pending"
	}
}
