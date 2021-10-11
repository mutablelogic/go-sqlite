package sqlite3

import (
	"fmt"
	"unsafe"

	multierror "github.com/hashicorp/go-multierror"
)

///////////////////////////////////////////////////////////////////////////////
// CGO

/*
#include <stdlib.h>
#include <sqlite3.h>
#include <pthread.h>
#include <assert.h>
#include <stdio.h>

// sqlite library needs to be compiled with -DSQLITE_ENABLE_UNLOCK_NOTIFY
// https://www.sqlite.org/unlock_notify.html

// A pointer to an instance of this structure is passed as the user-context
// pointer when registering for an unlock-notify callback.
typedef struct {
	int fired;                         // True after unlock event has occurred
	pthread_cond_t cond;               // Condition variable to wait on
	pthread_mutex_t mutex;             // Mutex to protect structure
} UnlockNotification;

// This function allocates a new UnlockNotification structure and returns a pointer to it.
static UnlockNotification* alloc_unlock_notification(void) {
	return (UnlockNotification* )malloc(sizeof(UnlockNotification));
}

// This function is an unlock-notify callback registered with SQLite.
static void unlock_notify_cb(void** apArg, int nArg){
	for(int i = 0; i < nArg; i++) {
		UnlockNotification* un = (UnlockNotification* )apArg[i];
		printf("   unlocking un=%p\n", un);
		fflush(stdout);
		pthread_mutex_lock(&un->mutex);
		un->fired = 1;
		pthread_cond_signal(&un->cond);
		pthread_mutex_unlock(&un->mutex);
	}
}

// This function assumes that an SQLite API call (either sqlite3_prepare_v2()
// or sqlite3_step()) has just returned SQLITE_LOCKED. The argument is the
// associated database connection.
//
// This function calls sqlite3_unlock_notify() to register for an
// unlock-notify callback, then blocks until that callback is delivered
// and returns SQLITE_OK. The caller should then retry the failed operation.
//
// Or, if sqlite3_unlock_notify() indicates that to block would deadlock
// the system, then this function returns SQLITE_LOCKED immediately. In
// this case the caller should not retry the operation and should roll
// back the current transaction (if any).
static int wait_for_unlock_notify(sqlite3* db){
	UnlockNotification un;
	un.fired = 0;
	pthread_mutex_init(&un.mutex, 0);
	pthread_cond_init(&un.cond, 0);

	printf("wait_for_unlock_notify %p un=%p\n",db,&un);
	fflush(stdout);

	// Register for an unlock-notify callback.
	// The call to sqlite3_unlock_notify() always returns either SQLITE_LOCKED
	// or SQLITE_OK.
	//
	// If SQLITE_LOCKED was returned, then the system is deadlocked. In this
	// case this function needs to return SQLITE_LOCKED to the caller so
	// that the current transaction can be rolled back. Otherwise, block
	// until the unlock-notify callback is invoked, then return SQLITE_OK.
	int rc = sqlite3_unlock_notify(db, unlock_notify_cb, (void* )&un);
	assert(rc==SQLITE_LOCKED || rc==SQLITE_OK);
	if(rc == SQLITE_OK) {
		pthread_mutex_lock(&un.mutex);
		if (!un.fired) {
			pthread_cond_wait(&un.cond, &un.mutex);
		}
		pthread_mutex_unlock(&un.mutex);
	}

	printf("unlocked %p un=%p rc=%d\n",db,&un,rc);
	fflush(stdout);

	// Destroy the mutex and condition variables.
	pthread_cond_destroy(&un.cond);
	pthread_mutex_destroy(&un.mutex);

	return rc;
}

// This code is a wrapper around sqlite3_step
static int _sqlite3_blocking_step(sqlite3_stmt* stmt, UnlockNotification* un) {
	int rc;
	sqlite3* db = sqlite3_db_handle(stmt);
	for (;;) {
		printf("make step db=%p\n",db);
		fflush(stdout);
		rc = sqlite3_step(stmt);
		if (sqlite3_extended_errcode(db) != SQLITE_LOCKED_SHAREDCACHE) {
     		break;
	    }
		printf("calling wait_for_unlock_notify db=%p\n",db);
		fflush(stdout);
		rc = wait_for_unlock_notify(db);
		if (rc != SQLITE_OK) {
			break;
		}
		printf("resetting db=%p\n",db);
		fflush(stdout);
		sqlite3_reset(stmt);
	}
	return rc;
}

// This code is a wrapper around sqlite3_prepare_v2
static int _sqlite3_blocking_prepare_v2(
  sqlite3* db,             // Database handle
  UnlockNotification* un,  // IN: Unlock notification object
  const char* sql,         // IN: UTF-8 encoded SQL statement
  int nSql,                // IN: Length of zSql in bytes
  sqlite3_stmt** stmt,     // OUT: A pointer to the prepared statement
  const char** pz          // OUT: End of parsed string
){
	int rc;
	for (;;) {
		rc = sqlite3_prepare_v2(db, sql, nSql, stmt, pz);
		if (sqlite3_extended_errcode(db) != SQLITE_LOCKED_SHAREDCACHE) {
     		break;
	    }
		rc = wait_for_unlock_notify(db);
		if (rc != SQLITE_OK) {
			break;
		}
	}
	return rc;
}


// This code is a wrapper around sqlite3_reset
static int _sqlite3_blocking_reset(sqlite3_stmt* stmt, UnlockNotification* un) {
	int rc;
	sqlite3* db = sqlite3_db_handle(stmt);
	for (;;) {
		rc = sqlite3_reset(stmt);
		if (sqlite3_extended_errcode(db) != SQLITE_LOCKED_SHAREDCACHE) {
     		break;
	    }
		rc = wait_for_unlock_notify(db);
		if (rc != SQLITE_OK) {
			break;
		}
	}
	return rc;
}
*/
import "C"

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Statement C.sqlite3_stmt

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s *Statement) String() string {
	str := "<statement"
	if s.IsBusy() {
		str += " busy"
	}
	/*
		if s.IsExplain() {
			str += " explain"
		}
	*/
	if s.IsReadonly() {
		str += " readonly"
	}
	if num_params := s.NumParams(); num_params > 0 {
		str += fmt.Sprint(" num_params=", num_params)
		params := make([]string, num_params)
		for i := 0; i < num_params; i++ {
			params[i] = s.ParamName(i + 1)
		}
		str += fmt.Sprintf(" params=%q", params)
	}
	if data_count := s.DataCount(); data_count > 0 {
		str += fmt.Sprint(" data_count=", data_count)
	}
	if col_count := s.ColumnCount(); col_count > 0 {
		str += fmt.Sprint(" col_count=", col_count)
		cols := make([]string, col_count)
		for i := 0; i < col_count; i++ {
			cols[i] = fmt.Sprintf("%v.%v.%v %v", s.ColumnDatabaseName(i), s.ColumnTableName(i), s.ColumnName(i), s.ColumnType(i))
		}
		str += fmt.Sprintf(" cols=%q", cols)
	}
	if sql := s.SQL(); sql != "" {
		str += fmt.Sprintf(" sql=%q", sql)
	}
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return next prepared statement, or first is nil
func (c *Conn) NextStatement(s *Statement) *Statement {
	if s := C.sqlite3_next_stmt((*C.sqlite3)(c), (*C.sqlite3_stmt)(s)); s == nil {
		return nil
	} else {
		return (*Statement)(s)
	}
}

// Prepare query
func (c *Conn) Prepare(query string) (*Statement, string, error) {
	var cQuery, cExtra *C.char
	var s *C.sqlite3_stmt

	// Populate CStrings
	if query != "" {
		cQuery = C.CString(query)
		defer C.free(unsafe.Pointer(cQuery))
	}

	// Prepare statement
	un := C.alloc_unlock_notification()
	defer C.free(unsafe.Pointer(un))
	if err := SQError(C._sqlite3_blocking_prepare_v2((*C.sqlite3)(c), un, cQuery, -1, &s, &cExtra)); err != SQLITE_OK {
		return nil, "", err.With(C.GoString(C.sqlite3_errmsg((*C.sqlite3)(c))))
	}

	// Return prepared statement and extra string
	return (*Statement)(s), C.GoString(cExtra), nil
}

// Bind parameters
func (s *Statement) Bind(v ...interface{}) error {

	// Check state
	if s.IsBusy() {
		return SQLITE_BUSY
	}

	// Reset bind parameters
	if err := s.ClearBindings(); err != nil {
		return err
	}

	// Bind parameters
	var result error
	for i, v := range v {
		if err := s.BindInterface(i+1, v); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Return any errors
	return result
}

// Return connection object from statement
func (s *Statement) Conn() *Conn {
	return (*Conn)(C.sqlite3_db_handle((*C.sqlite3_stmt)(s)))
}

// Reset statement
func (s *Statement) Reset() error {
	un := C.alloc_unlock_notification()
	defer C.free(unsafe.Pointer(un))
	err := SQError(C._sqlite3_blocking_reset((*C.sqlite3_stmt)(s), un))
	if (err & 0xFF) == SQLITE_LOCKED {
		fmt.Println("TODO: Locked Reset", int(err))
	}
	if err != SQLITE_OK {
		return err.With(C.GoString(C.sqlite3_errmsg((*C.sqlite3)(s.Conn()))))
	} else {
		return nil
	}
}

// IsBusy returns true if in middle of execution
func (s *Statement) IsBusy() bool {
	return intToBool(int(C.sqlite3_stmt_busy((*C.sqlite3_stmt)(s))))
}

// IsExplain returns true if the  statement S is an EXPLAIN statement or an EXPLAIN QUERY PLAN
/*
func (s *Statement) IsExplain() bool {
	return intToBool(int(C.sqlite3_stmt_isexplain((*C.sqlite3_stmt)(s))))
}
*/

// IsReadonly returns true if the statement makes no direct changes to the content of the database file.
func (s *Statement) IsReadonly() bool {
	return intToBool(int(C.sqlite3_stmt_readonly((*C.sqlite3_stmt)(s))))
}

// Finalize prepared statement
func (s *Statement) Finalize() error {
	if err := SQError(C.sqlite3_finalize((*C.sqlite3_stmt)(s))); err != SQLITE_OK {
		return err
	} else {
		return nil
	}
}

// Step statement
func (s *Statement) Step() error {
	un := C.alloc_unlock_notification()
	defer C.free(unsafe.Pointer(un))
	err := SQError(C._sqlite3_blocking_step((*C.sqlite3_stmt)(s), un))
	if (err & 0xFF) == SQLITE_LOCKED {
		fmt.Println("TODO Locked Step", int(err))
	}
	return err
}

// Return number of parameters expected for a statement
func (s *Statement) NumParams() int {
	return int(C.sqlite3_bind_parameter_count((*C.sqlite3_stmt)(s)))
}

// Returns parameter name for the nth parameter, which is an empty
// string if an unnamed parameter (?) or the parameter name otherwise (:a)
func (s *Statement) ParamName(index int) string {
	var cName *C.char
	cName = C.sqlite3_bind_parameter_name((*C.sqlite3_stmt)(s), C.int(index))
	if cName == nil {
		return ""
	} else {
		return C.GoString(cName)
	}
}

// Returns parameter index for a name, or zero
func (s *Statement) ParamIndex(name string) int {
	var cName *C.char

	// Set CString
	cName = C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	// Get parameter index and return it
	return int(C.sqlite3_bind_parameter_index((*C.sqlite3_stmt)(s), cName))
}

// Returns SQL associated with a statement
func (s *Statement) SQL() string {
	return C.GoString(C.sqlite3_sql((*C.sqlite3_stmt)(s)))
}

// Returns SQL associated with a statement, expanded with bound parameters
func (s *Statement) ExpandedSQL() string {
	return C.GoString(C.sqlite3_expanded_sql((*C.sqlite3_stmt)(s)))
}
