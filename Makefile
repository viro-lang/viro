.PHONY: build test clean all install test-summary

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
	@$(GOTEST) ./...

clean:
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)

pack: build 
	upx -9 $(BUILD_DIR)/$(BINARY_NAME)

deps:
	$(GOMOD) download
	$(GOMOD) verify
