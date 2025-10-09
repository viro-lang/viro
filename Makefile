.PHONY: build test clean all install

# Binary name
BINARY_NAME=viro
BUILD_DIR=.
CMD_DIR=./cmd/viro

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

all: test build

build:
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)

install:
	$(GOBUILD) -o $(GOPATH)/bin/$(BINARY_NAME) $(CMD_DIR)

deps:
	$(GOMOD) download
	$(GOMOD) verify
