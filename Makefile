.PHONY: build test clean all install test-summary install-syntax

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

install-syntax:
	@echo "Installing Viro syntax highlighting for Neovim..."
	@mkdir -p ~/.config/nvim/syntax
	@cp viro.vim ~/.config/nvim/syntax/
	@echo "✓ Syntax file installed to ~/.config/nvim/syntax/viro.vim"
	@echo ""
	@echo "Add this to your ~/.config/nvim/init.lua:"
	@echo "  vim.api.nvim_create_autocmd({\"BufRead\", \"BufNewFile\"}, {"
	@echo "    pattern = \"*.viro\","
	@echo "    command = \"set filetype=viro\","
	@echo "  })"
	@echo ""
	@echo "Or for init.vim:"
	@echo "  au BufRead,BufNewFile *.viro set filetype=viro"
