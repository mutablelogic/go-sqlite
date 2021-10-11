package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	// Packages
	"github.com/hashicorp/go-multierror"
	"github.com/mutablelogic/go-sqlite/pkg/indexer"
	"github.com/mutablelogic/go-sqlite/pkg/sqlite3"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-server"
	. "github.com/mutablelogic/go-sqlite"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	Workers uint              `json:"workers"`
	Paths   map[string]string `yaml:"index"`
	Schema  string            `yaml:"database"`
}

type plugin struct {
	pool    SQPool
	errs    chan error
	store   *indexer.Store
	index   map[string]*indexer.Indexer
	modtime map[string]time.Time
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultCapacity = 1024           // Default capacity for indexing queue
	deltaIndexDelta = 24 * time.Hour // Reindexing is done at most once per day
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create the module
func New(ctx context.Context, provider Provider) Plugin {
	p := new(plugin)
	p.index = make(map[string]*indexer.Indexer)
	p.modtime = make(map[string]time.Time)

	// Get configuration
	var cfg Config
	if err := provider.GetConfig(ctx, &cfg); err != nil {
		provider.Print(ctx, err)
		return nil
	}

	// Check for paths
	if len(cfg.Paths) == 0 {
		provider.Print(ctx, "no paths defined")
		return nil
	}

	// Get sqlite3
	if pool, ok := provider.GetPlugin(ctx, "sqlite3").(SQPool); !ok {
		provider.Print(ctx, "no sqlite3 plugin found")
		return nil
	} else {
		p.pool = pool
	}

	// Create a channel for errors
	p.errs = make(chan error)

	// Check schema
	schema, err := p.hasSchema(cfg.Schema)
	if err != nil {
		provider.Printf(ctx, "schema not found: %q", schema)
		return nil
	}

	// Create a queue, indexers and stores
	// TODO: Add a renderer interface for the store
	q := indexer.NewQueueWithCapacity(defaultCapacity)
	if q == nil {
		provider.Print(ctx, "unable to create queue")
		return nil
	} else if store := indexer.NewStore(p.pool, schema, q, nil, cfg.Workers); store == nil {
		provider.Print(ctx, "unable to create store")
		return nil
	} else {
		p.store = store
	}
	for name, path := range cfg.Paths {
		if idx, err := indexer.NewIndexer(name, path, q); err != nil {
			provider.Print(ctx, err)
			return nil
		} else {
			p.index[name] = idx
		}
	}

	// Return success
	return p
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *plugin) String() string {
	str := "<indexer"
	if p.store != nil {
		str += fmt.Sprint(" store=", p.store)
	}
	if len(p.index) > 0 {
		str += " indexes=["
		for _, index := range p.index {
			str += fmt.Sprintf(" %v", index)
		}
		str += " ]"
	}
	if p.pool != nil {
		str += fmt.Sprint(" pool=", p.pool)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Name() string {
	return "indexer"
}

func (p *plugin) Run(ctx context.Context, provider Provider) error {
	var wg sync.WaitGroup
	var results error

	// Error routine persists until error channel is closed
	go func() {
		for err := range p.errs {
			provider.Print(ctx, err)
		}
	}()

	// Add handlers
	if err := p.AddHandlers(ctx, provider); err != nil {
		return err
	}

	// Run indexer processes
	// TODO: Report error when indexer can't be started
	for _, idx := range p.index {
		wg.Add(1)
		go func(idx *indexer.Indexer) {
			defer wg.Done()
			if err := idx.Run(ctx, p.errs); err != nil {
				results = multierror.Append(results, err)
			}
		}(idx)
	}

	// Run store process
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := p.store.Run(ctx, p.errs); err != nil {
			results = multierror.Append(results, err)
		}
	}()

	// Timer routine for reindexing
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Run until cancelled - print any errors from the indexer
FOR_LOOP:
	for {
		select {
		case <-ticker.C:
			if index := p.nextReindex(deltaIndexDelta); index != nil {
				if err := index.Walk(ctx, func(err error) {
					p.modtime[index.Name()] = time.Now()
					if err != nil {
						p.errs <- fmt.Errorf("reindexing completed with errors: %w", err)
					}
				}); err != nil {
					p.errs <- fmt.Errorf("reindexing cannot start: %w", err)
				}
			}
			ticker.Reset(time.Minute)
		case <-ctx.Done():
			break FOR_LOOP
		}
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Close error channel
	close(p.errs)

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (p *plugin) hasSchema(v string) (string, error) {
	conn := p.pool.Get()
	if conn == nil {
		return "", ErrInternalAppError.With("unable to get connection")
	}
	defer p.pool.Put(conn)

	if v == "" {
		v = sqlite3.DefaultSchema
	}
	for _, schema := range conn.Schemas() {
		if schema == v {
			return v, nil
		}
	}

	return "", ErrNotFound.Withf("schema not found: %q", v)
}

// Return the next index to be reindexed
func (p *plugin) nextReindex(delta time.Duration) *indexer.Indexer {
	results := make([]*indexer.Indexer, 0, len(p.index))
	for name, index := range p.index {
		if modtime, exists := p.modtime[name]; !exists {
			results = append(results, index)
			p.modtime[name] = time.Time{}
		} else if time.Since(modtime) > delta {
			results = append(results, index)
		}
	}
	// Return nil if nothing needs reindexed
	if len(results) == 0 {
		return nil
	}
	// Return the index which was earliest last indexed
	sort.Slice(results, func(a, b int) bool {
		namea := results[a].Name()
		nameb := results[b].Name()
		return p.modtime[namea].Before(p.modtime[nameb])
	})
	// Return first index
	return results[0]
}
