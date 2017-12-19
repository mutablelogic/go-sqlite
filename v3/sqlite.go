/*
	SQLite client
	(c) Copyright David Thorpe 2017
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE file
*/

package v3

import (
	"database/sql/driver"
	"fmt"

	gopi "github.com/djthorpe/gopi"
	sqlite "github.com/djthorpe/sqlite"
	sqlite_driver "github.com/mattn/go-sqlite3"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Client defines the configuration parameters for connecting to SQLite Database
type Client struct {
	DSN string
}

type client struct {
	log  gopi.Logger
	conn driver.Conn
}

type column struct {
	n  string            // column name
	t  sqlite.Type       // sql type
	f  sqlite.Flag       // flags
	kv map[string]string // key-value tag pairs
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	_V3_TAG = "sql"
)

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open returns a client object
func (config Client) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<sqlite.v3.Client>Open{ dsn='%v' }", config.DSN)

	this := new(client)
	this.log = log

	d := &sqlite_driver.SQLiteDriver{}
	if conn, err := d.Open(config.DSN); err != nil {
		return nil, err
	} else {
		this.conn = conn
	}

	// Return success
	return this, nil
}

// Close releases any resources associated with the client connection
func (this *client) Close() error {
	this.log.Debug("<sqlite.v3.Client>Close{ }")
	return this.conn.Close()
}

////////////////////////////////////////////////////////////////////////////////
// INTERFACE

func (this *client) Do(statement sqlite.Statement) error {
	this.log.Debug("<sqlite.v3.Client>Do{ %v }", statement.SQL())
	if s, err := this.conn.Prepare(statement.SQL()); err != nil {
		return err
	} else if _, err := s.Exec([]driver.Value{}); err != nil {
		return err
	} else {
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// COLUMN

func (this *column) Name() string {
	return this.n
}

func (this *column) Type() sqlite.Type {
	return this.t
}

func (this *column) Flags() sqlite.Flag {
	return this.f
}

func (this *column) String() string {
	return fmt.Sprintf("<sqlite.v3.Column>{ name=%v type=%v flags=%v tags=%v }", this.n, this.t, this.f, this.kv)
}
