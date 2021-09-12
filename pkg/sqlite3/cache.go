package sqlite3

////////////////////////////////////////////////////////////////////////////////
// TYPES

// PoolCache caches prepared statements and profiling information for
// statements so it's possible to see slow queries, etc.
type PoolCache struct {
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return a prepared statement from the cache
func (cache *PoolCache) Prepare(q string) (*Results, error) {

}
