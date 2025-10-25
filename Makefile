.PHONY: build test clean all install grammar test-summary

# Binary name
BINARY_NAME=viro
BUILD_DIR=.
CMD_DIR=./cmd/viro
GRAMMAR_FILE=grammar/viro.peg
PARSER_OUTPUT=internal/parse/peg/parser.go

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

all: test build

grammar:
	pigeon -o $(PARSER_OUTPUT) $(GRAMMAR_FILE)

build: grammar
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

test-summary:
	@echo "Running tests and summarizing results..."
	@$(GOTEST) -v ./... 2>&1 | awk '/^--- (PASS|FAIL):/ {total++; if ($$2 ~ /^FAIL/) failed++} END {passed = total - failed; print "Total tests:", total ", Passed:", passed ", Failed:", failed}'

clean:
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	$(GOBUILD) -o $(GOPATH)/bin/$(BINARY_NAME) $(CMD_DIR)

deps:
	$(GOMOD) download
	$(GOMOD) verify
