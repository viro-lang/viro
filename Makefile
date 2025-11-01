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

test:
	@$(GOTEST) ./...

clean:
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)

install: pack
	cp ./viro $(GOPATH)/bin/$(BINARY_NAME)

pack: build 
	upx -9 viro

deps:
	$(GOMOD) download
	$(GOMOD) verify
