package ncruces

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"unicode"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
	sqlite "github.com/mutablelogic/go-sqlite"
	sqlite3 "github.com/ncruces/go-sqlite3"
	attribute "go.opentelemetry.io/otel/attribute"
	codes "go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	trace "go.opentelemetry.io/otel/trace"
)

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// exec executes SQL, without arguments or rows, such as BEGIN, which is
// traced
func (p *pool) exec(ctx context.Context, db *sqlite3.Conn, sql string) error {
	return p.trace(ctx, "", sql, func(map[string]any) error {
		return db.Exec(sql)
	})
}

// trace calls the function, which executes the SQL, in a span, of the name,
// or of the SQL, when it's empty, with the arguments which the function
// bound, and its error
func (p *pool) trace(ctx context.Context, name, sql string, fn func(args map[string]any) error) error {
	if p.opts.Tracer == nil {
		return fn(nil)
	}
	if name == "" {
		name = spanName(sql)
	}
	_, span := p.opts.Tracer.Start(ctx, name,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(semconv.DBSystemSqlite, attribute.String("db.statement", strings.TrimSpace(sql))),
	)
	defer span.End()

	// Execute it, and say what its arguments were, and its error, which isn't
	// one when no rows were found
	args := make(map[string]any)
	err := fn(args)
	for name, value := range args {
		span.SetAttributes(argAttribute(name, value))
	}
	endSpan(span, err)
	return err
}

// startOpen starts the span of opening a database, of its path
func startOpen(ctx context.Context, tracer trace.Tracer, path string) (context.Context, trace.Span) {
	if tracer == nil {
		return ctx, nil
	}
	return tracer.Start(ctx, "sqlite.open",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(semconv.DBSystemSqlite, attribute.String("arg.path", path)),
	)
}

// endSpan ends the span, of its error, which isn't one when no rows were
// found
func endSpan(span trace.Span, err error) {
	if span == nil {
		return
	}
	if err != nil && !errors.Is(err, sqlite.ErrNotFound) {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// argAttribute returns the attribute of an argument, arg. and its name, of
// its value, of a string, a number or a bool, of named types and pointers
// too, or else its JSON
func argAttribute(name string, value any) attribute.KeyValue {
	key := "arg." + name
	rv := reflect.ValueOf(value)
	for rv.Kind() == reflect.Pointer && !rv.IsNil() {
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.String:
		return attribute.String(key, rv.String())
	case reflect.Bool:
		return attribute.Bool(key, rv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return attribute.Int64(key, rv.Int())
	case reflect.Float32, reflect.Float64:
		return attribute.Float64(key, rv.Float())
	}
	return attribute.String(key, types.Stringify(value))
}

// spanName returns the name of the span of SQL, which has no name:
// sqlite.transaction, sqlite.commit or sqlite.rollback, of a transaction's
// statements, including a savepoint's, or else sqlite.query
func spanName(sql string) string {
	words := strings.FieldsFunc(strings.ToLower(sql), func(r rune) bool {
		return unicode.IsSpace(r) || r == ';'
	})
	switch {
	case len(words) == 0:
		return "sqlite.query"
	case words[0] == "begin", words[0] == "savepoint":
		return "sqlite.transaction"
	case words[0] == "commit", words[0] == "end", words[0] == "release":
		return "sqlite.commit"
	case words[0] == "rollback":
		return "sqlite.rollback"
	}
	return "sqlite.query"
}
