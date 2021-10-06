# go-sqlite

This module provides an interface for sqlite, including:

  * Opening in-memory databases and persistent file-based databases;
  * Transactions (committing changes and rolling back on errors);
  * Adding custom functions, authentication and authorization;
  * Reflection on databases (schemas, tables, columns, indexes, etc);
  * Executing arbitrary statements or building statements programmatically;
  * A pool of connections to run sqlite in a highliy concurrent environment such as a webservice;
  * A backend REST API for sqlite;
  * A generalized importer to import data from other data sources in different formats;
  * A generalized file indexer to index files in a directory tree and provide a REST API
    to search them;
  * A frontend web application to explore and interact with databases.

Presently the module is in development and the API is subject to change.

| If you want to...                    |  Folder         | Documentation |
|--------------------------------------|-----------------|---------------|
| Use the lower-level sqlite3 bindings similar to the [C API](https://www.sqlite.org/capi3ref.html) | [sys/sqlite3](https://github.com/mutablelogic/go-sqlite/tree/master/sys/sqlite3) | [README.md](https://github.com/mutablelogic/go-sqlite/blob/master/sys/sqlite3/README.md) |
| Use high-concurrency high-level interface including statement caching and connection pool | [pkg/sqlite3](https://github.com/mutablelogic/go-sqlite/tree/master/pkg/sqlite3) | [README.md](https://github.com/mutablelogic/go-sqlite/blob/master/pkg/sqlite3/README.md) |
| Implement or use a REST API to sqlite3 | [plugin/sqlite3](https://github.com/mutablelogic/go-sqlite/tree/master/plugin/sqlite3) | [README.md](https://github.com/mutablelogic/go-sqlite/blob/master/plugin/sqlite3/README.md) |
| Develop or use a front-end web service to the REST API backend | [npm/sqlite3](https://github.com/mutablelogic/go-sqlite/tree/master/npm/sqlite3) | [README.md](https://github.com/mutablelogic/go-sqlite/blob/master/npm/sqlite3/README.md) |
| Use an "object" interface to persist structured data | [pkg/sqobj](https://github.com/mutablelogic/go-sqlite/tree/master/pkg/sqobj) | [README.md](https://github.com/mutablelogic/go-sqlite/blob/master/pkg/sqobj/README.md) |
| Use a statement builder to programmatically write SQL statements | [pkg/lang](https://github.com/mutablelogic/go-sqlite/tree/master/pkg/lang) | [README.md](https://github.com/mutablelogic/go-sqlite/blob/master/pkg/lang/README.md) |
| Implement a generalized data importer from CSV, JSON, Excel, etc | [pkg/importer](https://github.com/mutablelogic/go-sqlite/tree/master/pkg/importer) | [README.md](https://github.com/mutablelogic/go-sqlite/blob/master/pkg/importer/README.md) |
| Implement a search indexer | [pkg/indexer](https://github.com/mutablelogic/go-sqlite/tree/master/pkg/indexer) | [README.md](https://github.com/mutablelogic/go-sqlite/blob/master/pkg/indexer/README.md) |
| Tokenize SQL statements for syntax colouring (for example) | [pkg/tokenizer](https://github.com/mutablelogic/go-sqlite/tree/master/pkg/tokenizer) | [README.md](https://github.com/mutablelogic/go-sqlite/blob/master/pkg/tokenizer/README.md) |
| See example command-line tools | [cmd](https://github.com/mutablelogic/go-sqlite/tree/master/cmd) | [README.md](https://github.com/mutablelogic/go-sqlite/blob/master/cmd/README.md) |

## Requirements

  * [go1.17](https://golang.org/dl/) or later;
  * Tested on Debian Linux (32- and 64- bit) on ARM and macOS on x64
    architectures.

## Building

There are some examples in the `cmd` folder of the main repository on how to use
the package. The various make targets are:

  * `make all` will perform tests, build all examples, the backend API and the frontend web application;
  * `make test` will perform tests;
  * `make cmd` will build example command-line tools into the `build` folder;
  * `make server plugins` will install the backend server and required plugins in the `build` folder;
  * `make npm` will compile the frontend web application in a 'dist' folder for each npm module located in the `npm` folder;
  * `make clean` will remove all build artifacts.

## Contributing & Distribution

__This module is currently in development and subject to change.__

Please do file feature requests and bugs [here](https://github.com/mutablelogic/go-sqlite/issues).
The license is Apache 2 so feel free to redistribute. Redistributions in either source
code or binary form must reproduce the copyright notice, and please link back to this
repository for more information:

> Copyright (c) 2021, David Thorpe, All rights reserved.
