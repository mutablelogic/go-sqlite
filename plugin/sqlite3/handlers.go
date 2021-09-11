package main

import (
	"context"
	"net/http"
	"regexp"

	// Modules
	router "github.com/djthorpe/go-server/pkg/httprouter"
	sqlite3 "github.com/djthorpe/go-sqlite/pkg/sqlite3"

	// Namespace imports
	. "github.com/djthorpe/go-server"

	// Some sort of hack
	_ "gopkg.in/yaml.v3"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type PingResponse struct {
	Version string   `json:"version"`
	Modules []string `json:"modules"`
	Schemas []string `json:"schemas"`
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
	conn := p.Get(req.Context())
	if conn == nil {
		router.ServeError(w, http.StatusBadGateway, "No connection")
		return
	}
	defer p.Put(conn)

	// Populate response
	response := PingResponse{
		Modules: []string{},
		Schemas: []string{},
	}
	response.Version = sqlite3.Version()
	response.Modules = append(response.Modules, conn.Modules()...)
	response.Schemas = append(response.Schemas, conn.Schemas()...)

	// Serve response
	router.ServeJSON(w, response, http.StatusOK, 0)
}
