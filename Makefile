# Path parameters
GO ?= $(shell which go 2>/dev/null)

# Locations
BUILD_DIR ?= build
CMD_DIR := $(filter-out cmd/_%,$(wildcard cmd/*))

# Set OS and Architecture
ARCH ?= $(shell arch | tr A-Z a-z | sed 's/x86_64/amd64/' | sed 's/i386/amd64/' | sed 's/armv7l/arm/' | sed 's/aarch64/arm64/')
OS ?= $(shell uname | tr A-Z a-z)
VERSION ?= $(shell git describe --tags --always | sed 's/^v//')

# Set build flags
VERSION_PKG = github.com/mutablelogic/go-server/pkg/version
BUILD_LD_FLAGS += -X $(VERSION_PKG).GitTag=$(shell git describe --tags --always)
BUILD_LD_FLAGS += -X $(VERSION_PKG).GitBranch=$(shell git name-rev HEAD --name-only --always)
BUILD_LD_FLAGS += -X $(VERSION_PKG).BuildDate=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BUILD_FLAGS = -ldflags "-s -w ${BUILD_LD_FLAGS}"

# SQLite is pure Go, from ncruces/go-sqlite3, so nothing needs cgo. The
# repo's go.work builds filesearch with this go-sqlite, rather than its
# release, and a go.work in a parent folder isn't used, unless GOWORK is set
export CGO_ENABLED ?= 0
export GOWORK ?= $(CURDIR)/go.work

###############################################################################
# ALL

.PHONY: all
all: build

###############################################################################
# BUILD

# Build the packages, and the commands in the cmd directory, when there are
# any
.PHONY: build
build: tidy
	@echo Build
	@${GO} build ./...
	@$(if $(CMD_DIR),$(MAKE) --no-print-directory $(CMD_DIR),true)
	@$(MAKE) --no-print-directory filesearch

# The flags of the C libraries which cgo doesn't allow by default, for
# filesearch with MEDIA=1: the -Xpreprocessor flag of LibRaw built with
# OpenMP, as Homebrew's is, which is before -fopenmp
CGO_CFLAGS_ALLOW ?= -Xpreprocessor
CGO_LDFLAGS_ALLOW ?= -Xpreprocessor
export CGO_CFLAGS_ALLOW CGO_LDFLAGS_ALLOW

# On macOS, Homebrew's ffmpeg@8, which is keg-only, so isn't found otherwise,
# when it's installed. The linker warns about the libraries which more than
# one of the C libraries link, which isn't a problem.
ifeq ($(OS),darwin)
BREW ?= $(shell command -v brew 2>/dev/null)
BREW_FFMPEG := $(if $(BREW),$(wildcard $(shell $(BREW) --prefix ffmpeg@8 2>/dev/null)/lib/pkgconfig))
ifneq ($(BREW_FFMPEG),)
export PKG_CONFIG_PATH := $(BREW_FFMPEG)$(if $(PKG_CONFIG_PATH),:$(PKG_CONFIG_PATH))
endif
export CGO_LDFLAGS += -Wl,-no_warn_duplicate_libraries
endif

# Build filesearch, of its own module, which reads documents, email, text
# and images, or, with MEDIA=1, audio, video and photos' EXIF, HEIF and RAW
# too, with cgo, and the C libraries which go-metadata needs
.PHONY: filesearch
filesearch: go-dep mkdir
	@echo Build command filesearch $(if $(MEDIA),with media,)
	@cd filesearch && $(if $(MEDIA),CGO_ENABLED=1,) ${GO} build $(if $(MEDIA),-tags media,) ${BUILD_FLAGS} -o ../${BUILD_DIR}/filesearch ./cmd/filesearch

$(CMD_DIR): go-dep mkdir
	@echo Build command $(notdir $@) GOOS=${OS} GOARCH=${ARCH}
	@GOOS=${OS} GOARCH=${ARCH} ${GO} build ${BUILD_FLAGS} -o ${BUILD_DIR}/$(notdir $@) ./$@

###############################################################################
# TEST

.PHONY: test
test: vet unit-test coverage-test

.PHONY: vet
vet: go-dep
	@echo Vet
	@${GO} vet ./...

# The race detector needs cgo on Linux, though SQLite doesn't
.PHONY: unit-test
unit-test: go-dep
	@echo Unit Tests
	@CGO_ENABLED=1 ${GO} test -race ./...
	@cd filesearch && CGO_ENABLED=1 ${GO} test -race ./...

# Coverage is of the packages, other than the tests which implementations pass,
# which test the root package, as well as the implementation
COVER_PKG = $(shell GOWORK=off ${GO} list ./... | grep -v /test$$ | paste -sd, -)

.PHONY: coverage-test
coverage-test: go-dep mkdir
	@echo Test Coverage
	@${GO} test -coverpkg=$(COVER_PKG) -coverprofile ${BUILD_DIR}/coverprofile.out ./...
	@${GO} tool cover -func ${BUILD_DIR}/coverprofile.out > ${BUILD_DIR}/coverage.txt

###############################################################################
# LICENSE

GOFILES_LICENSE = $(shell find . -name '*.go' -not -path './build/*' -not -path './.git/*' -not -path './_old/*')

# Add the license header to the Go files which don't have it
.PHONY: license
license: go-dep mkdir
	@GOBIN=$(abspath $(BUILD_DIR)) ${GO} install github.com/google/addlicense@latest
	@${BUILD_DIR}/addlicense -c "David Thorpe" -l apache -y 2026 $(GOFILES_LICENSE)

###############################################################################
# CLEAN

.PHONY: mkdir
mkdir:
	@install -d $(BUILD_DIR)

.PHONY: tidy
tidy: go-dep
	@echo 'go tidy'
	@$(GO) mod tidy
	@cd filesearch && $(GO) mod tidy

.PHONY: clean
clean: tidy
	@echo 'clean'
	@rm -fr $(BUILD_DIR)
	@$(GO) clean

###############################################################################
# DEPENDENCIES

.PHONY: go-dep
go-dep:
	@test -f "${GO}" && test -x "${GO}"  || (echo "Missing go binary" && exit 1)
