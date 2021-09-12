# Paths to packages
GO=$(shell which go)
NPM=$(shell which npm)

# Paths to locations, etc
BUILD_DIR = "build"
PLUGIN_DIR = $(wildcard plugin/*)
NPM_DIR = $(wildcard npm/*)

# Build flags
BUILD_MODULE = "github.com/djthorpe/go-server"
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitSource=${BUILD_MODULE}
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitTag=$(shell git describe --tags)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GitHash=$(shell git rev-parse HEAD)
BUILD_LD_FLAGS += -X $(BUILD_MODULE)/pkg/config.GoBuildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BUILD_FLAGS = -ldflags "-s -w $(BUILD_LD_FLAGS)" 

.PHONY: all server npm plugins dependencies mkdir clean 

all: clean plugins server npm

server: dependencies mkdir
	@echo Build server
	@${GO} build -o ${BUILD_DIR}/server ${BUILD_FLAGS} github.com/djthorpe/go-server/cmd/server

npm: $(NPM_DIR)

$(NPM_DIR): FORCE
	@echo Build npm $(notdir $@)
	@cd $@ && ${NPM} run build

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
