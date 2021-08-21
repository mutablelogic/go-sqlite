
# Go parameters
GO=go
GOFLAGS = -ldflags "-s -w $(GOLDFLAGS)" 
BUILDDIR = build
TAGS = 
COMMAND = $(wildcard cmd/*)

# All targets
all: test commands

# Rules for building
.PHONY: commands $(COMMAND)
commands: mkdir $(COMMAND)

$(COMMAND): 
	@echo "Building ${BUILDDIR}/$@"
	@$(GO) build -o ${BUILDDIR}/$@ -tags "$(TAGS)" ${GOFLAGS} ./$@

.PHONY: test
test:
	@$(GO) test -tags "$(TAGS)" ./pkg/...

.PHONY: mkdir
mkdir:
	@install -d $(BUILDDIR)

.PHONY: clean
clean: 
	@rm -fr $(BUILDDIR)
	$(GO) mod tidy
	$(GO) clean
