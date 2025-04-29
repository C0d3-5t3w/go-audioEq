#!/bin/bash

# Simple script to rename the built shared library to a .vst extension
# and potentially move it to a common VST plugin location.

PLUGIN_NAME="goAudioEq"
BUILD_DIR="./build"
TARGET_EXT=".vst"

# Detect OS and set default VST paths
VST2_PATH_MACOS="$HOME/Library/Audio/Plug-Ins/VST"
VST2_PATH_LINUX="$HOME/.vst"
# VST2_PATH_WINDOWS="/Program Files/Steinberg/VSTPlugins" # Example

# Find the built library file (.so, .dll, .dylib)
BUILT_LIB=$(find "$BUILD_DIR" -name "${PLUGIN_NAME}.so" -o -name "${PLUGIN_NAME}.dll" -o -name "${PLUGIN_NAME}.dylib" | head -n 1)

if [ -z "$BUILT_LIB" ]; then
    echo "Error: Built library not found in $BUILD_DIR"
    exit 1
fi

LIB_BASENAME=$(basename "$BUILT_LIB")
TARGET_FILE="$BUILD_DIR/${PLUGIN_NAME}${TARGET_EXT}"

echo "Found library: $BUILT_LIB"

# Rename the library to .vst in the build directory
echo "Renaming $LIB_BASENAME to ${PLUGIN_NAME}${TARGET_EXT}..."
mv "$BUILT_LIB" "$TARGET_FILE"
if [ $? -ne 0 ]; then
    echo "Error: Failed to rename library."
    exit 1
fi
echo "Renamed to $TARGET_FILE"

# Optional: Copy to system VST folder
INSTALL_PATH=""
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    INSTALL_PATH="$VST2_PATH_LINUX"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    INSTALL_PATH="$VST2_PATH_MACOS"
# elif [[ "$OSTYPE" == "cygwin" || "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
    # INSTALL_PATH="$VST2_PATH_WINDOWS" # Uncomment and adjust for Windows
    : # Placeholder for Windows
fi

if [ ! -z "$INSTALL_PATH" ]; then
    echo "Attempting to install to $INSTALL_PATH..."
    mkdir -p "$INSTALL_PATH"
    cp "$TARGET_FILE" "$INSTALL_PATH/"
    if [ $? -eq 0 ]; then
        echo "Successfully copied $TARGET_FILE to $INSTALL_PATH"
    else
        echo "Warning: Failed to copy to $INSTALL_PATH. Try running with sudo or manually copy the file."
    fi
else
    echo "Skipping installation to system VST folder (OS not detected or path not set)."
fi

echo "Packaging complete. You may need to restart your DAW."
exit 0
