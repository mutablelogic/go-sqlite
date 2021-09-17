package sqlite3

import (
	// Modules
	"context"

	sqlite3 "github.com/djthorpe/go-sqlite/sys/sqlite3"

	// Namespace Imports
	. "github.com/djthorpe/go-errors"
	. "github.com/djthorpe/go-sqlite"
)

////////////////////////////////////////////////////////////////////////////////
// PRIVATE FUNCTIONS

func (p *Pool) auth(ctx context.Context, action sqlite3.SQAction, args [4]string) error {
	switch action {
	case sqlite3.SQLITE_CREATE_INDEX:
		return p.Auth.CanExec(ctx, SQLITE_AUTH_INDEX|SQLITE_AUTH_CREATE, args[2], args[1], args[0])
	case sqlite3.SQLITE_CREATE_TABLE: //           2   /* Table Name      NULL            */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_TABLE|SQLITE_AUTH_CREATE, args[2], args[0])
	case sqlite3.SQLITE_CREATE_TEMP_INDEX: //      3   /* Index Name      Table Name      */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_INDEX|SQLITE_AUTH_CREATE|SQLITE_AUTH_TEMP, args[2], args[1], args[0])
	case sqlite3.SQLITE_CREATE_TEMP_TABLE: //      4   /* Table Name      NULL            */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_TABLE|SQLITE_AUTH_CREATE|SQLITE_AUTH_TEMP, args[2], args[0])
	case sqlite3.SQLITE_CREATE_TEMP_TRIGGER: //    5   /* Trigger Name    Table Name      */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_TRIGGER|SQLITE_AUTH_CREATE|SQLITE_AUTH_TEMP, args[2], args[1], args[0])
	case sqlite3.SQLITE_CREATE_TEMP_VIEW: //       6   /* View Name       NULL            */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_VIEW|SQLITE_AUTH_CREATE|SQLITE_AUTH_TEMP, args[2], args[1], args[0])
	case sqlite3.SQLITE_CREATE_TRIGGER: //         7   /* Trigger Name    Table Name      */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_TRIGGER|SQLITE_AUTH_CREATE, args[2], args[1], args[0])
	case sqlite3.SQLITE_CREATE_VIEW: //            8   /* View Name       NULL            */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_VIEW|SQLITE_AUTH_CREATE, args[2], args[1], args[0])
	case sqlite3.SQLITE_DELETE: //                 9   /* Table Name      NULL            */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_TABLE|SQLITE_AUTH_DELETE, args[2], args[0])
	case sqlite3.SQLITE_DROP_INDEX: //            10   /* Index Name      Table Name      */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_INDEX|SQLITE_AUTH_DROP, args[2], args[1], args[0])
	case sqlite3.SQLITE_DROP_TABLE: //            11   /* Table Name      NULL            */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_TABLE|SQLITE_AUTH_DROP, args[2], args[0])
	case sqlite3.SQLITE_DROP_TEMP_INDEX: //       12   /* Index Name      Table Name      */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_INDEX|SQLITE_AUTH_DROP|SQLITE_AUTH_TEMP, args[2], args[1], args[0])
	case sqlite3.SQLITE_DROP_TEMP_TABLE: //       13   /* Table Name      NULL            */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_TABLE|SQLITE_AUTH_DROP|SQLITE_AUTH_TEMP, args[2], args[0])
	case sqlite3.SQLITE_DROP_TEMP_TRIGGER: //     14   /* Trigger Name    Table Name      */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_TRIGGER|SQLITE_AUTH_DROP|SQLITE_AUTH_TEMP, args[2], args[1], args[0])
	case sqlite3.SQLITE_DROP_TEMP_VIEW: //        15   /* View Name       NULL            */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_VIEW|SQLITE_AUTH_DROP|SQLITE_AUTH_TEMP, args[2], args[1], args[0])
	case sqlite3.SQLITE_DROP_TRIGGER: //          16   /* Trigger Name    Table Name      */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_TRIGGER|SQLITE_AUTH_DROP, args[2], args[1], args[0])
	case sqlite3.SQLITE_DROP_VIEW: //             17   /* View Name       NULL            */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_VIEW|SQLITE_AUTH_DROP, args[2], args[1], args[0])
	case sqlite3.SQLITE_INSERT: //                18   /* Table Name      NULL            */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_TABLE|SQLITE_AUTH_INSERT, args[2], args[0])
	case sqlite3.SQLITE_PRAGMA:
		//                19   /* Pragma Name     1st arg or NULL */
		if args[1] == "" {
			return p.Auth.CanExec(ctx, SQLITE_AUTH_PRAGMA, args[0])
		} else {
			return p.Auth.CanExec(ctx, SQLITE_AUTH_PRAGMA, args[0], args[1])
		}
	case sqlite3.SQLITE_SELECT: //                21   /* NULL            NULL            */
		return p.Auth.CanSelect(ctx)
	case sqlite3.SQLITE_ALTER_TABLE: //           26   /* Database Name   Table Name      */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_TABLE|SQLITE_AUTH_ALTER, args[0], args[1])
	case sqlite3.SQLITE_CREATE_VTABLE: //         29   /* Table Name      Module Name     */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_VTABLE|SQLITE_AUTH_CREATE, args[2], args[0], args[1])
	case sqlite3.SQLITE_DROP_VTABLE: //           30   /* Table Name      Module Name     */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_VTABLE|SQLITE_AUTH_DROP, args[2], args[0], args[1])
	case sqlite3.SQLITE_ANALYZE: //               28   /* Table Name      NULL            */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_TABLE|SQLITE_AUTH_ANALYZE, args[2], args[0])
	case sqlite3.SQLITE_FUNCTION: //              31   /* NULL            Function Name   */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_FUNCTION, args[1])
	case sqlite3.SQLITE_TRANSACTION: //           22   /* Operation       NULL            */
		switch args[0] {
		case "BEGIN":
			return p.Auth.CanTransaction(ctx, SQLITE_AUTH_TRANSACTION|SQLITE_AUTH_BEGIN)
		case "ROLLBACK":
			return p.Auth.CanTransaction(ctx, SQLITE_AUTH_TRANSACTION|SQLITE_AUTH_ROLLBACK)
		case "COMMIT":
			return p.Auth.CanTransaction(ctx, SQLITE_AUTH_TRANSACTION|SQLITE_AUTH_COMMIT)
		}
		// TODO: Op is BEGIN, ROLLBACK or COMMIT so use this
	case sqlite3.SQLITE_READ: //                  20   /* Table Name      Column Name     */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_TABLE|SQLITE_AUTH_READ, args[2], args[0], args[1])
	case sqlite3.SQLITE_UPDATE: //                23   /* Table Name      Column Name     */
		return p.Auth.CanExec(ctx, SQLITE_AUTH_TABLE|SQLITE_AUTH_UPDATE, args[2], args[0], args[1])
		// TODO		case sqlite3.SQLITE_SAVEPOINT: //             32   /* Operation       Savepoint Name  */
		// TODO case sqlite3.SQLITE_ATTACH: //                24   /* Filename        NULL            */
		// TODO case sqlite3.SQLITE_DETACH: //                25   /* Database Name   NULL            */
		// TODO case sqlite3.SQLITE_REINDEX: //               27   /* Index Name      NULL            */
	}

	// Report an error
	p.err(ErrNotImplemented.With("Auth: ", action))

	// Return allow by default
	return nil
}
