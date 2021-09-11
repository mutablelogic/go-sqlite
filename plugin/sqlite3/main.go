package main

import (
	"context"

	// Packages
	sqlite3 "github.com/djthorpe/go-sqlite/pkg/sqlite3"

	// Namespace imports
	. "github.com/djthorpe/go-server"
	. "github.com/djthorpe/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type plugin struct {
	SQPool
	errs chan error
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create the module
func New(ctx context.Context, provider Provider) Plugin {
	p := new(plugin)

	// Get configuration
	cfg := sqlite3.PoolConfig{}
	if err := provider.GetConfig(ctx, &cfg); err != nil {
		provider.Print(ctx, err)
		return nil
	}

	// Create a channel for errors
	p.errs = make(chan error)

	// Create a pool
	if pool, err := sqlite3.OpenPool(cfg, p.errs); err != nil {
		provider.Print(ctx, err)
		return nil
	} else {
		p.SQPool = pool
	}

	// Return success
	return p
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

	// Run until cancelled
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
	if err := p.SQPool.Close(); err != nil {
		provider.Print(ctx, err)
	}

	// Close error channel
	close(p.errs)

	// Return success
	return nil
}
