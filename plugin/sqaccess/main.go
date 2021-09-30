package main

import (
	"context"
	"fmt"

	// Packages
	"github.com/mutablelogic/go-sqlite/pkg/sqlite3"
	"github.com/mutablelogic/go-sqlite/pkg/sqobj"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Database string `yaml:"database"`
}

type User struct {
	Username string `sqlite:"user,primary"`
	Hash     string `sqlite:"hash"`
}

type plugin struct {
	Config
	SQPool
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	clUser = sqobj.MustRegisterClass(N("User"), User{})
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create the module
func New(ctx context.Context, provider Provider) Plugin {
	p := new(plugin)

	// Get configuration
	cfg := Config{}
	if err := provider.GetConfig(ctx, &cfg); err != nil {
		provider.Print(ctx, err)
		return nil
	} else {
		p.Config = cfg
	}

	// Get sqlite3
	if pool := provider.GetPlugin(ctx, "sqlite3").(SQPool); pool == nil {
		provider.Print(ctx, "no sqlite3 plugin")
		return nil
	} else {
		p.SQPool = pool
	}

	// Get a connection
	conn := p.Get(ctx)
	if conn == nil {
		provider.Print(ctx, "no sqlite3 connection")
		return nil
	}
	defer p.Put(conn)

	// Create the database
	if sqobj, err := sqobj.With(conn.(*sqlite3.Conn), p.Database, clUser); err != nil {
		provider.Print(ctx, err)
		return nil
	} else {
		fmt.Println(sqobj)
	}

	// Return success
	return p
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *plugin) String() string {
	str := "<sqaccess"
	if p.Config.Database != "" {
		str += fmt.Sprintf(" database=%q", p.Config.Database)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Name() string {
	return "sqaccess"
}

func (p *plugin) Run(ctx context.Context, provider Provider) error {

	// Run until cancelled - print any errors from the connection pool
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			break FOR_LOOP
		}
	}

	// Return success
	return nil
}
