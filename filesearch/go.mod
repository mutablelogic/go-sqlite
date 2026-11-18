module github.com/mutablelogic/go-sqlite/filesearch

go 1.27.2

// The parent module, while it's developed alongside it
replace github.com/mutablelogic/go-sqlite => ../

require (
	github.com/mutablelogic/go-sqlite v1.0.51
	github.com/stretchr/testify v1.12.1
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/mutablelogic/go-server v1.8.16 // indirect
	github.com/ncruces/go-sqlite3 v0.37.0 // indirect
	github.com/ncruces/go-sqlite3-wasm/v6 v6.5.35401 // indirect
	github.com/ncruces/julianday v1.0.0 // indirect
	go.opentelemetry.io/otel v1.47.0 // indirect
	go.opentelemetry.io/otel/trace v1.47.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/sys v0.49.0 // indirect
)
