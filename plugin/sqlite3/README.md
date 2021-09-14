# sqlite3 API plugin

The API plugin provides an interface into the database through HTTP calls. It is a plugin to
the monolithic server [`github.com/djthorpe/go-server`](github.com/djthorpe/go-server) 
and as such the server needs to be installed in addition to the plugin. Instructions on how
to install the server and necessary plugins is described below.

This package is part of a wider project, `github.com/djthorpe/go-sqlite`.
Please see the [module documentation](https://github.com/djthorpe/go-sqlite/blob/master/README.md)
for more information.

## Running the API backend

The simplest way to install the backend is to run the following commands:

```bash
[bash] git clone github.com/djthorpe/go-sqlite.git
[bash] cd go-sqlite
[bash] make server plugins
```

This will put the following binaries in the `build` directory:

  * `server` is the basic monolith server binary;
  * `httpserver.plugin` is the HTTP server plugin;
  * `log.plugin` provides logging for HTTP requests;
  * `static.plugin` provides static file serving. This is not necessary for the
    backend, but can be used to serve a frontend (see [here](../../npm/sqlite3) for more
    information on the frontend).

To run the server, there is an example configuration file in the `etc` folder:

```bash
[bash] ./build/server etc/server.yaml
```

On Macintosh, you may need to use the `DYLD_LIBRARY_PATH` environment variable to
locate the correct sqlite3 library. For example,

```bash
[bash] brew install sqlite3
[bash] DYLD_LIBRARY_PATH="/usr/local/opt/sqlite/lib" \
  ./build/server etc/server.yaml
```

You can override the port by passing the `-addr` flag:

```bash
[bash] ./build/server -addr :9001 etc/server.yaml
```

Press CTRL+C to stop the server.

## REST API calls

Requests can generally be `application/json` or `application/x-www-form-urlencoded`, which
needs to be indicated in the `Content-Type` header. Responses are always in `application/json`.

| Endpoint Path      | Method    | Name     | Description |
|--------------------|-----------|----------|-------------|
| /                  | GET       | Ping     | Return version, schema, connection pool and module information
| /`schema`          | GET       | Schema   | Return information about a schema: tables, indexes, tiggers and views
| /`schema`/`table`  | GET       | Table    | Return rows of the table or view
| /-/q               | POST      | Query    | Execute a query
| /-/tokenizer       | POST      | Tokenize | Tokenize a query for syntax colouring

## Error Responses

Errors are returned when the status code is not `200 OK`. A typical error response will look like this:

```json
{
   "reason" : "1 error occurred:\n\t* SQL logic error\n\n",
   "code" : 400
}
```

## Plugin Configuration

TODO

## Requests and Responses

### Ping Request and Response

There are no query arguments for this call. Typically a response will look like this:

```json
{
  "version": "3.36.0",
  "modules": [
    "json_tree",
    "json_each",
    "fts3",
    "fts4",
    "fts3tokenize",
    "fts5vocab",
    "fts5",
    "rtree",
    "rtree_i32",
    "fts4aux",
    "geopoly"
  ],
  "schemas": [
    "main",
    "test"
  ],
  "pool": {
    "cur": 1,
    "max": 50
  }
}
```

### Schema Request and Response

There are no query arguments for this call. Typically a response will provide you with information
in the schemas. For example, a typical response may look like this:

TODO

### Table Request and Response

TODO

### Query Request and Response

TODO

### Tokenizer Request and Response

