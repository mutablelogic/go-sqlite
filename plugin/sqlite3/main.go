package main

import (
	"context"
	"fmt"
	"time"

	// Packages
	sqlite3 "github.com/mutablelogic/go-sqlite/pkg/sqlite3"

	// Namespace imports
	. "github.com/mutablelogic/go-server"
	. "github.com/mutablelogic/go-sqlite"

	// Some sort of hack
	_ "gopkg.in/yaml.v3"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Databases map[string]string `yaml:"databases"`
	Max       int               `yaml:"max"`
	Create    bool              `yaml:"create"`
	Trace     bool              `yaml:"trace"`
}

type plugin struct {
	pool SQPool
	errs chan error
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create the module
func New(ctx context.Context, provider Provider) Plugin {
	p := new(plugin)

	// Get configuration
	var cfg Config
	if err := provider.GetConfig(ctx, &cfg); err != nil {
		provider.Print(ctx, err)
		return nil
	}
	// Check for databases
	if len(cfg.Databases) == 0 {
		provider.Print(ctx, fmt.Errorf("no databases defined"))
		return nil
	}
	// Create the pool
	poolcfg := sqlite3.NewConfig().
		WithMaxConnections(cfg.Max).
		WithCreate(cfg.Create)
	for name, path := range cfg.Databases {
		poolcfg = poolcfg.WithSchema(name, path)
	}
	if cfg.Trace {
		poolcfg = poolcfg.WithTrace(func(q string, d time.Duration) {
			if d >= 0 {
				provider.Printf(ctx, "TRACE %q => %v", q, d)
			}
		})
	}

	// Create a channel for errors
	p.errs = make(chan error)

	// Create a pool
	if pool, err := sqlite3.OpenPool(poolcfg, p.errs); err != nil {
		provider.Print(ctx, err)
		close(p.errs)
		p.errs = nil
		return nil
	} else {
		p.pool = pool
	}

	// Return success
	return p
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *plugin) String() string {
	str := "<sqlite3"
	if p.pool != nil {
		str += fmt.Sprint(" ", p.pool)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Name() string {
	return "sqlite3"
}

func (p *plugin) Run(ctx context.Context, provider Provider) error {
	// Add handlers
	if err := p.AddHandlers(ctx, provider); err != nil {
		return err
	}

	// Run until cancelled - print any errors from the connection pool
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			break FOR_LOOP
		case err := <-p.errs:
			if err != nil {
				provider.Print(ctx, err)
			}
		}
	}

	// Close the pool
	if err := p.pool.Close(); err != nil {
		provider.Print(ctx, err)
	}

	// Close error channel
	close(p.errs)

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PUT AND GET

func (p *plugin) Get() SQConnection {
	return p.pool.Get()
}

func (p *plugin) Put(conn SQConnection) {
	p.pool.Put(conn)
}
