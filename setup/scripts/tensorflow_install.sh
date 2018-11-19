#! /bin/bash

set -e

# setup
TENSORFLOW_VERSION="1.12.0"
TENSORFLOW_DIR="tmp/tensorflow"
OUTPUT_FILE="$TENSORFLOW_DIR/tensorflow.tar.gz"
mkdir -p $TENSORFLOW_DIR

# set defaults
TENSORFLOW_DIST="cpu"
TENSORFLOW_PLATFORM="linux"
TENSORFLOW_ARCH="x86_64"
LINKER=ldconfig

# detect other platforms
if [[ "$OSTYPE" == "darwin"* ]]; then
  TENSORFLOW_PLATFORM="darwin"
  LINKER=update_dyld_shared_cache
fi

# set release and download URL
TENSORFLOW="$TENSORFLOW_DIST-$TENSORFLOW_PLATFORM-$TENSORFLOW_ARCH-$TENSORFLOW_VERSION"
TENSORFLOW_URL="https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-$TENSORFLOW.tar.gz"
echo "[INFO] Tensorflow $TENSORFLOW will be downloaded"

# download and install
echo "[INFO] Downloading tensorflow"
wget "$TENSORFLOW_URL" -O "$OUTPUT_FILE"
echo "[INFO] Extracting tensorflow"
tar -xzf "$OUTPUT_FILE" -C "$TENSORFLOW_DIR"
echo "[INFO] Installing tensorflow"
sudo cp -r "$TENSORFLOW_DIR/lib" "/usr/local/lib/"
sudo cp -r "$TENSORFLOW_DIR/include/" "/usr/local/include"
echo "[INFO] Updating linkers"
sudo $LINKER
