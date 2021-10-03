# Paths to packages
GO=$(shell which go)
SED=$(shell which sed)
NPM=$(shell which npm)

# Paths to locations, etc
BUILD_DIR := "build"
PLUGIN_DIR := $(wildcard plugin/*)
NPM_DIR := $(wildcard npm/*)
CMD_DIR := $(filter-out cmd/README.md, $(wildcard cmd/*))

# Build flags
BUILD_MODULE = "github.com/mutablelogic/go-server"
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitSource=${BUILD_MODULE}
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitTag=$(shell git describe --tags)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitHash=$(shell git rev-parse HEAD)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GoBuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BUILD_FLAGS = -ldflags "-s -w $(BUILD_LD_FLAGS)" 
BUILD_VERSION = $(shell git describe --tags)
BUILD_ARCH = $(shell $(GO) env GOARCH)
BUILD_PLATFORM = $(shell $(GO) env GOOS)

all: clean server plugins npm cmd

server: dependencies mkdir
	@echo Build server
	@${GO} build -o ${BUILD_DIR}/server ${BUILD_FLAGS} github.com/mutablelogic/go-server/cmd/server

plugins: $(PLUGIN_DIR)
	@echo Build plugin httpserver 
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/httpserver.plugin ${BUILD_FLAGS} github.com/mutablelogic/go-server/plugin/httpserver
	@echo Build plugin log 
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/log.plugin ${BUILD_FLAGS} github.com/mutablelogic/go-server/plugin/log
	@echo Build plugin static 
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/static.plugin ${BUILD_FLAGS} github.com/mutablelogic/go-server/plugin/static

npm: $(NPM_DIR)

cmd: dependencies mkdir $(CMD_DIR)
  
$(NPM_DIR): FORCE
	@echo Build npm $(notdir $@)
	@cd $@ && ${NPM} run build

$(CMD_DIR): FORCE
	@echo Build cmd $(notdir $@)
	@${GO} build -o ${BUILD_DIR}/$(notdir $@) ${BUILD_FLAGS} ./$@

$(PLUGIN_DIR): FORCE
	@echo Build plugin $(notdir $@)
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/$(notdir $@).plugin ${BUILD_FLAGS} ./$@

FORCE:


deb: nfpm go-server-sqlite3-deb

go-server-sqlite3-deb: plugin/sqlite3
	@echo Package go-server-sqlite3 deb
	@${SED} \
		-e 's/^version:.*$$/version: $(BUILD_VERSION)/'  \
		-e 's/^arch:.*$$/arch: $(BUILD_ARCH)/' \
		-e 's/^platform:.*$$/platform: $(BUILD_PLATFORM)/' \
		etc/nfpm/go-server-sqlite3/nfpm.yaml > $(BUILD_DIR)/go-server-sqlite3-nfpm.yaml
	@nfpm pkg -f $(BUILD_DIR)/go-server-sqlite3-nfpm.yaml --packager deb --target $(BUILD_DIR)

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


nfpm:
	@echo Installing nfpm
	@${GO} mod tidy
	@${GO} install github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.3.1	

dependencies:
ifeq (,${GO})
        $(error "Missing go binary")
endif
ifeq (,${NPM})
        $(error "Missing npm binary")
endif
ifeq (,${SED})
        $(error "Missing sed binary")
endif

mkdir:
	@install -d ${BUILD_DIR}

clean:
	@rm -fr $(BUILD_DIR)
	@${GO} mod tidy
	@${GO} clean
