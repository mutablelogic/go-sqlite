# Paths to packages
GO=$(shell which go)

# Paths to locations, etc
BUILD_DIR = "build"
BUILD_MODULE = "github.com/djthorpe/go-server"
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitSource=${BUILD_MODULE}
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitTag=$(shell git describe --tags)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitHash=$(shell git rev-parse HEAD)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GoBuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BUILD_FLAGS = -ldflags "-s -w $(BUILD_LD_FLAGS)" 
BUILD_VERSION = $(shell git describe --tags)
PLUGIN_DIR = $(wildcard plugin/*)

.PHONY: all server plugins dependencies mkdir clean 

all: clean server plugins

server: dependencies mkdir
	@echo Build server
	@${GO} build -o ${BUILD_DIR}/server ${BUILD_FLAGS} github.com/djthorpe/go-server/cmd/server

plugins: $(PLUGIN_DIR)
	@echo Build plugin httpserver 
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/httpserver.plugin ${BUILD_FLAGS} github.com/djthorpe/go-server/plugin/httpserver
	@echo Build plugin log 
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/log.plugin ${BUILD_FLAGS} github.com/djthorpe/go-server/plugin/log

$(PLUGIN_DIR): FORCE
	@echo Build plugin $(notdir $@)
	@${GO} build -buildmode=plugin -o ${BUILD_DIR}/$(notdir $@).plugin ${BUILD_FLAGS} ./$@

FORCE:

dependencies:
ifeq (,${GO})
        $(error "Missing go binary")
endif

mkdir:
	@install -d ${BUILD_DIR}

clean:
	@rm -fr $(BUILD_DIR)
	@${GO} mod tidy
	@${GO} clean
