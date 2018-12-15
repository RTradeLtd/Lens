#! /bin/bash

set -e

# setup
TENSORFLOW_VERSION="1.12.0"
TENSORFLOW_DIR="tmp/tensorflow"
mkdir -p $TENSORFLOW_DIR

# set defaults
TENSORFLOW_DIST="cpu"
TENSORFLOW_PLATFORM="linux"
TENSORFLOW_ARCH="x86_64"
LINKER=ldconfig

# detect other platforms
if [[ "$OSTYPE" == "darwin"* ]]; then
  echo "[INFO] Darwin system detected - setting custom config"
  TENSORFLOW_PLATFORM="darwin"
  LINKER=update_dyld_shared_cache
fi

# set release and download URL
TENSORFLOW="$TENSORFLOW_DIST-$TENSORFLOW_PLATFORM-$TENSORFLOW_ARCH-$TENSORFLOW_VERSION"
TENSORFLOW_URL="https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-$TENSORFLOW.tar.gz"
echo "[INFO] Tensorflow $TENSORFLOW will be downloaded"

# download and install
echo "[INFO] Installing tensorflow"
curl -L $TENSORFLOW_URL | sudo tar -C /usr/local -xz

echo "[INFO] Updating linker"
sudo $LINKER
