package sqlite3_test

import (
	"context"
	"strings"
	"testing"

	// Module imports
	sqlite3 "github.com/mutablelogic/go-sqlite/sys/sqlite3"

	// Namespace Imports
	. "github.com/mutablelogic/go-sqlite"
	. "github.com/mutablelogic/go-sqlite/pkg/lang"
	. "github.com/mutablelogic/go-sqlite/pkg/sqlite3"
)

func Test_Auth_001(t *testing.T) {
	errs, cancel := catchErrors(t)
	defer cancel()

	// Create the pool
	pool, err := OpenPool(PoolConfig{
		Schemas: map[string]string{"main": ":memory:"},
		Auth:    NewAuth(t),
	}, errs)
	if err != nil {
		t.Error(err)
	}
	defer pool.Close()

	// Get conn
	conn := pool.Get(context.Background())
	if conn == nil {
		t.Fatal("conn is nil")
	}
	defer pool.Put(conn)

	// Make various requests
	if err := conn.(*Conn).Exec(N("table_a").CreateTable(N("a").WithType("TEXT").WithPrimary()), nil); err != nil {
		t.Error(err)
	}
	for _, schema := range conn.Schemas() {
		conn.Tables(schema)
	}

	// Insert a row
	conn.(*Conn).Exec(N("table_a").Insert().DefaultValues(), nil)

	// Delete a row
	conn.(*Conn).Exec(N("table_a").Delete(true), nil)

	// Do a transaction
	conn.(*Conn).Begin(sqlite3.SQLITE_TXN_DEFAULT)
	conn.(*Conn).Rollback()
	conn.(*Conn).Begin(sqlite3.SQLITE_TXN_IMMEDIATE)
	conn.(*Conn).Commit()
}

type Auth struct {
	*testing.T
}

func NewAuth(t *testing.T) *Auth {
	a := new(Auth)
	a.T = t
	return a
}

func (a *Auth) CanSelect(_ context.Context) error {
	a.Log("CanSelect")
	return nil
}

// Transaction and savepoint (or empty string if transaction)
func (a *Auth) CanTransaction(_ context.Context, flag SQAuthFlag) error {
	a.Log("CanBegin: ", flag)
	return nil
}

// Return nil to allow operation given object and operation, schema plus arguments
func (a *Auth) CanExec(_ context.Context, flag SQAuthFlag, schema string, args ...string) error {
	if flag.Is(SQLITE_AUTH_TABLE) {
		if strings.HasPrefix(args[0], "sqlite_") {
			return nil
		}
	}
	if flag.Is(SQLITE_AUTH_INDEX) {
		if strings.HasPrefix(args[1], "sqlite_") {
			return nil
		}
	}
	a.Logf("CanExec: %v %q %q", flag, schema, args)
	return nil
}
