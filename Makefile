# Makefile for goldmark-adf
#
# This module requires Go 1.25+ with the experimental json/v2 package.

# Build configuration
GOEXPERIMENT := jsonv2
export GOEXPERIMENT

# Source files for dependency tracking
GO_FILES := $(shell find . -name '*.go' -type f)
GO_MOD := go.mod

# Stamp file directory
STAMP_DIR := .stamps

# Default target
.PHONY: all
all: fmt vet test build

# Create stamp directory
$(STAMP_DIR):
	mkdir -p $(STAMP_DIR)

# Format - stamp file tracks last successful run
$(STAMP_DIR)/fmt: $(GO_FILES) | $(STAMP_DIR)
	go fmt ./...
	touch $@

.PHONY: fmt
fmt: $(STAMP_DIR)/fmt

# Vet - depends on fmt being run first
$(STAMP_DIR)/vet: $(GO_FILES) $(GO_MOD) $(STAMP_DIR)/fmt | $(STAMP_DIR)
	go vet ./...
	touch $@

.PHONY: vet
vet: $(STAMP_DIR)/vet

# Build - validates the module compiles
$(STAMP_DIR)/build: $(GO_FILES) $(GO_MOD) | $(STAMP_DIR)
	go build ./...
	touch $@

.PHONY: build
build: $(STAMP_DIR)/build

# Test - always runs, go test handles its own caching
# Supports passing arguments for specific tests/packages
# Usage: make test ARGS="-run TestName ./path/to/package"
#        make test ARGS="-v ./adfschema/..."
ARGS ?= ./...
.PHONY: test
test: $(STAMP_DIR)/vet
	go test $(ARGS)

# Clean - remove stamp files
.PHONY: clean
clean:
	rm -rf $(STAMP_DIR)
