# Go parameters
GOCMD=go
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOGENERATE=$(GOCMD) generate

# App parameters
GOPI=github.com/djthorpe/gopi
GOLDFLAGS += -X $(GOPI).GitTag=$(shell git describe --tags)
GOLDFLAGS += -X $(GOPI).GitBranch=$(shell git name-rev HEAD --name-only --always)
GOLDFLAGS += -X $(GOPI).GitHash=$(shell git rev-parse HEAD)
GOLDFLAGS += -X $(GOPI).GoBuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GOFLAGS = -ldflags "-s -w $(GOLDFLAGS)" 

all: test install

install: sq_import fs_indexer

sq_import:
	$(GOINSTALL) $(GOFLAGS) ./cmd/sq_import/...

fs_indexer:
	$(GOINSTALL) $(GOFLAGS) ./cmd/fs_indexer/...

fs_indexer_service: protogen
	$(GOINSTALL) $(GOFLAGS) ./cmd/fs_indexer_service/...

protogen:
	$(GOGENERATE) ./rpc/protobuf/...

test: 
	$(GOTEST) -v .
	$(GOTEST) -v ./sys/...

clean: 
	$(GOCLEAN)