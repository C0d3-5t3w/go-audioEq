# Basic Makefile for building the Go VST plugin

# Variables
PLUGIN_NAME=goAudioEq
SOURCE_DIR=./cmd
OUTPUT_DIR=./build
GO=go
# Adjust extension based on OS
ifeq ($(OS),Windows_NT)
    OUTPUT_EXT=.dll
    # Add Windows specific build flags if needed
    LDFLAGS=
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
        OUTPUT_EXT=.so
        LDFLAGS=-Wl,-Bsymbolic # Example Linux linker flags
    endif
    ifeq ($(UNAME_S),Darwin)
        OUTPUT_EXT=.dylib # VSTs on macOS often use .vst extension, but build produces .dylib
        LDFLAGS= # Add macOS specific build flags if needed
    endif
endif
OUTPUT_FILE=$(OUTPUT_DIR)/$(PLUGIN_NAME)$(OUTPUT_EXT)
VST_OUTPUT_FILE=$(OUTPUT_DIR)/$(PLUGIN_NAME).vst # Common VST extension

# Default target
all: build

# Build the plugin as a C-shared library
build:
	@echo "Building $(PLUGIN_NAME) VST plugin..."
	@mkdir -p $(OUTPUT_DIR)
	$(GO) build -buildmode=c-shared -ldflags="$(LDFLAGS)" -o $(OUTPUT_FILE) $(SOURCE_DIR)/$(PLUGIN_NAME).go
	@echo "Build complete: $(OUTPUT_FILE)"
	# Optional: Rename or copy to standard VST extension (handled by BinaryToVst.sh potentially)
	# @cp $(OUTPUT_FILE) $(VST_OUTPUT_FILE)
	# @echo "Copied to $(VST_OUTPUT_FILE)"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(OUTPUT_DIR)
	@rm -f $(SOURCE_DIR)/*.h # Remove generated C header file

# Run tests (if any)
test:
	@echo "Running tests..."
	$(GO) test ./...

.PHONY: all build clean test

