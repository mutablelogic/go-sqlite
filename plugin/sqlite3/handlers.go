package main

import (
	"context"
	"fmt"
	"html"
	"html/template"
	"io"
	"net/http"
	"regexp"
	"strconv"

	// Modules
	router "github.com/djthorpe/go-server/pkg/httprouter"
	sqlite3 "github.com/djthorpe/go-sqlite/pkg/sqlite3"

	// Namespace imports
	. "github.com/djthorpe/go-server"
	. "github.com/djthorpe/go-sqlite"
	. "github.com/djthorpe/go-sqlite/pkg/lang"

	// Some sort of hack
	_ "gopkg.in/yaml.v3"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type PingResponse struct {
	Version string       `json:"version"`
	Modules []string     `json:"modules"`
	Schemas []string     `json:"schemas"`
	Pool    PoolResponse `json:"pool"`
}

type PoolResponse struct {
	Cur int32 `json:"cur"`
	Max int32 `json:"max"`
}

type SchemaResponse struct {
	Schema   string                `json:"schema"`
	Filename string                `json:"filename,omitempty"`
	Memory   bool                  `json:"memory,omitempty"`
	Tables   []SchemaTableResponse `json:"tables,omitempty"`
}

type SchemaTableResponse struct {
	Name    string   `json:"name"`
	Schema  string   `json:"schema"`
	Indexes []string `json:"indexes,omitempty"`
}

type SqlRequest struct {
	Sql string `json:"sql"`
}

type SqlResultResponse struct {
	Sql []string `json:"sql"`
}

type SyntaxResponse struct {
	Html     []template.HTML `json:"html,omitempty"`
	Complete bool            `json:"complete"`
}

///////////////////////////////////////////////////////////////////////////////
// ROUTES

var (
	reRoutePing     = regexp.MustCompile(`^/?$`)
	reRouteSchema   = regexp.MustCompile(`^/([a-zA-Z][a-zA-Z0-9_-]+)/?$`)
	reRouteSyntaxer = regexp.MustCompile(`^/-/syntax/?$`)
	reRouteQuery    = regexp.MustCompile(`^/-/q/?$`)
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

	// Add handler for schema
	if err := provider.AddHandlerFuncEx(ctx, reRouteSchema, p.ServeSchema); err != nil {
		return err
	}

	// Add handler for SQL syntax checker
	if err := provider.AddHandlerFuncEx(ctx, reRouteSyntaxer, p.ServeSyntaxer, http.MethodPost); err != nil {
		return err
	}

	// Add handler for queries
	if err := provider.AddHandlerFuncEx(ctx, reRouteQuery, p.ServeQuery, http.MethodPost); err != nil {
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
		Schemas: []string{},
		Modules: []string{},
	}
	response.Version = sqlite3.Version()
	response.Schemas = append(response.Schemas, conn.Schemas()...)
	response.Modules = append(response.Schemas, conn.Modules()...)
	response.Pool = PoolResponse{Cur: p.Cur(), Max: p.Max()}

	// Serve response
	router.ServeJSON(w, response, http.StatusOK, 2)
}

func (p *plugin) ServeSchema(w http.ResponseWriter, req *http.Request) {
	// Decode params, params[0] is the schema name
	params := router.RequestParams(req)

	// Get a connection
	conn := p.Get(req.Context())
	if conn == nil {
		router.ServeError(w, http.StatusBadGateway, "No connection")
		return
	}
	defer p.Put(conn)

	// Check for schema
	if stringSliceContainsElement(conn.Schemas(), params[0]) == false {
		router.ServeError(w, http.StatusNotFound, "Schema not found", strconv.Quote(params[0]))
		return
	}

	// Populate response
	response := SchemaResponse{
		Schema:   params[0],
		Filename: conn.Filename(params[0]),
		Tables:   []SchemaTableResponse{},
	}

	if err := conn.(*sqlite3.Conn).Exec(Q("PRAGMA database_list;"), func(row, col []string) bool {
		fmt.Printf("%q => %q\n", col, row)
		return false
	}); err != nil {
		router.ServeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Set memory flag
	if response.Filename == "" {
		response.Memory = true
	}

	// Populate tables
	for _, table := range conn.Tables(params[0]) {
		response.Tables = append(response.Tables, SchemaTableResponse{
			Name:   table,
			Schema: params[0],
		})
	}

	// Serve response
	router.ServeJSON(w, response, http.StatusOK, 2)
}

func (p *plugin) ServeSyntaxer(w http.ResponseWriter, req *http.Request) {
	// Decode request
	query := SqlRequest{}
	if err := router.RequestBody(req, &query); err != nil {
		router.ServeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get a connection
	conn := p.Get(req.Context())
	if conn == nil {
		router.ServeError(w, http.StatusBadGateway, "No connection")
		return
	}
	defer p.Put(conn)

	// Tokenize input
	html, err := tokenize(query.Sql)
	if err != nil {
		router.ServeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Populate response
	response := SyntaxResponse{
		Html:     html,
		Complete: sqlite3.IsComplete(query.Sql),
	}

	// Serve response
	router.ServeJSON(w, response, http.StatusOK, 2)
}

func (p *plugin) ServeQuery(w http.ResponseWriter, req *http.Request) {
	// Decode request
	query := SqlRequest{}
	if err := router.RequestBody(req, &query); err != nil {
		router.ServeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get a connection
	conn := p.Get(req.Context())
	if conn == nil {
		router.ServeError(w, http.StatusBadGateway, "No connection")
		return
	}
	defer p.Put(conn)

	// Perform query
	response := make([]SqlResultResponse, 0)
	if err := conn.Do(req.Context(), SQLITE_TXN_DEFAULT, func(txn SQTransaction) error {
		_, err := txn.Query(Q(query.Sql))
		if err != nil {
			return err
		}
		// Return success
		return nil
	}); err != nil {
		router.ServeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Serve response
	router.ServeJSON(w, response, http.StatusOK, 2)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func stringSliceContainsElement(v []string, elem string) bool {
	for _, v := range v {
		if v == elem {
			return true
		}
	}
	return false
}

// tokenize will return an array of html spans, one for each token in the input
func tokenize(v string) ([]template.HTML, error) {
	result := []template.HTML{}

	// Iterate through the tokenizer
	t := sqlite3.NewTokenizer(v)
	for {
		token, err := t.Next()
		if token == nil || err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		switch t := token.(type) {
		case sqlite3.KeywordToken:
			result = appendtoken(result, "keyword", t)
		case sqlite3.TypeToken:
			result = appendtoken(result, "type", t)
		case sqlite3.NameToken:
			result = appendtoken(result, "name", t)
		case sqlite3.ValueToken:
			result = appendtoken(result, "value", t)
		case sqlite3.PuncuationToken:
			result = appendtoken(result, "puncuation", t)
		case sqlite3.WhitespaceToken:
			result = appendtoken(result, "space", t)
		default:
			result = appendtoken(result, "", t)
		}
	}

	// Return success
	return result, nil
}

// Append token adds a html span to the result slice
func appendtoken(result []template.HTML, class string, value interface{}) []template.HTML {
	v := fmt.Sprint(value)
	if class != "" {
		return append(result, template.HTML("<span class="+strconv.Quote(class)+">"+html.EscapeString(v)+"</span>"))
	} else {
		return append(result, template.HTML("<span>"+html.EscapeString(v)+"</span>"))
	}
}
