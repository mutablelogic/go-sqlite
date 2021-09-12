# Paths to packages
GO=$(shell which go)
NPM=$(shell which npm)

# Paths to locations, etc
BUILD_DIR := "build"
PLUGIN_DIR := $(wildcard plugin/*)
NPM_DIR := $(wildcard npm/*)
CMD_DIR := $(filter-out cmd/README.md, $(wildcard cmd/*))

# Build flags
BUILD_MODULE = "github.com/djthorpe/go-server"
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitSource=${BUILD_MODULE}
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitTag=$(shell git describe --tags)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitHash=$(shell git rev-parse HEAD)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GoBuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BUILD_FLAGS = -ldflags "-s -w $(BUILD_LD_FLAGS)" 

.PHONY: all test server npm cmd plugins dependencies mkdir clean 

all: clean plugins server npm cmd

server: dependencies mkdir
	@echo Build server
	@${GO} build -o ${BUILD_DIR}/server ${BUILD_FLAGS} github.com/djthorpe/go-server/cmd/server

npm: $(NPM_DIR)

cmd: dependencies mkdir $(CMD_DIR)
  
$(NPM_DIR): FORCE
	@echo Build npm $(notdir $@)
	@cd $@ && ${NPM} run build

$(CMD_DIR): FORCE
	@echo Build cmd $(notdir $@)
	@${GO} build -o ${BUILD_DIR}/$(notdir $@) ${BUILD_FLAGS} ./$@

plugins: $(PLUGIN_DIR)
	@echo Build plugin httpserver 
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/httpserver.plugin ${BUILD_FLAGS} github.com/djthorpe/go-server/plugin/httpserver
	@echo Build plugin log 
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/log.plugin ${BUILD_FLAGS} github.com/djthorpe/go-server/plugin/log
	@echo Build plugin static 
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/static.plugin ${BUILD_FLAGS} github.com/djthorpe/go-server/plugin/static

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

dependencies:
ifeq (,${GO})
        $(error "Missing go binary")
endif
ifeq (,${NPM})
        $(error "Missing npm binary")
endif

mkdir:
	@install -d ${BUILD_DIR}

clean:
	@rm -fr $(BUILD_DIR)
	@${GO} mod tidy
	@${GO} clean
