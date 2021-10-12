# Paths to packages
GO=$(shell which go)

# Paths to locations, etc
BUILD_DIR := build
PLUGIN_DIR := $(wildcard plugin/*)
CMD_DIR := $(filter-out cmd/README.md, $(wildcard cmd/*))
SQLITE_MODULE = "github.com/mutablelogic/go-sqlite"
SERVER_MODULE = "github.com/mutablelogic/go-server"

# Build flags
BUILD_LD_FLAGS += -X $(SQLITE_MODULE)/pkg/config.GitSource=${SQLITE_MODULE}
BUILD_LD_FLAGS += -X $(SQLITE_MODULE)/pkg/config.GitTag=$(shell git describe --tags)
BUILD_LD_FLAGS += -X $(SQLITE_MODULE)/pkg/config.GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_LD_FLAGS += -X $(SQLITE_MODULE)/pkg/config.GitHash=$(shell git rev-parse HEAD)
BUILD_LD_FLAGS += -X $(SQLITE_MODULE)/pkg/config.GoBuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BUILD_FLAGS = -ldflags "-s -w ${BUILD_LD_FLAGS}" 

all: clean server plugins cmd

server: dependencies
	@echo Build server
	@${GO} build -o ${BUILD_DIR}/server ${BUILD_FLAGS} ${SERVER_MODULE}/cmd/server

plugins: dependencies $(PLUGIN_DIR)
	@echo Build plugin httpserver 
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/httpserver.plugin ${BUILD_FLAGS} ${SERVER_MODULE}/plugin/httpserver
	@echo Build plugin log 
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/log.plugin ${BUILD_FLAGS} ${SERVER_MODULE}/plugin/log
	@echo Build plugin env 
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/env.plugin ${BUILD_FLAGS} ${SERVER_MODULE}/plugin/env
	@echo Build plugin static 
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/static.plugin ${BUILD_FLAGS} ${SERVER_MODULE}/plugin/static
	@echo Build plugin renderer 
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/renderer.plugin ${BUILD_FLAGS} ${SERVER_MODULE}/plugin/renderer
	@echo Build plugin text-renderer 
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/text-renderer.plugin ${BUILD_FLAGS} ${SERVER_MODULE}/plugin/text-renderer
	@echo Build plugin markdown-renderer 
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/markdown-renderer.plugin ${BUILD_FLAGS} ${SERVER_MODULE}/plugin/markdown-renderer 

cmd: dependencies $(CMD_DIR)

$(CMD_DIR): FORCE
	@echo Build cmd $(notdir $@)
	@${GO} build -o ${BUILD_DIR}/$(notdir $@) ${BUILD_FLAGS} ./$@

$(PLUGIN_DIR): FORCE
	@echo Build plugin $(notdir $@)
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/$(notdir $@).plugin ${BUILD_FLAGS} ./$@

FORCE:

test:
	@echo Test sys/sqlite3
	@${GO} test ./sys/sqlite3
	@echo Test pkg/sqlite3
	@${GO} test ./pkg/sqlite3
	@echo Test pkg/tokenizer
	@${GO} test ./pkg/tokenizer
	@echo Test pkg/lang
	@${GO} test ./pkg/lang
	@echo Test pkg/importer
	@${GO} test ./pkg/importer
	@echo Test pkg/indexer
	@${GO} test ./pkg/indexer
	@echo Test pkg/quote
	@${GO} test ./pkg/quote
	@echo Test pkg/sqobj
	@${GO} test ./pkg/sqobj

dependencies: mkdir
ifeq (,${GO})
        $(error "Missing go binary")
endif

mkdir:
	@install -d ${BUILD_DIR}

clean:
	@echo Clean
	@rm -fr $(BUILD_DIR)
	@${GO} mod tidy
	@${GO} clean

