#! /bin/bash

set -e

# detect other platforms
if [[ "$OSTYPE" == "darwin"* ]]; then
  brew install imagemagick@6
  exit 0
fi

# set release and download URL
OUTPUT_DIR="tmp/imagemagick"
OUTPUT_FILE="$OUTPUT_DIR/ImageMagick.tar.gz"
mkdir -p "$OUTPUT_DIR"
VERSION="6.9.10-14"
URL="http://www.imagemagick.org/download/ImageMagick-$VERSION.tar.gz"
echo "[INFO] ImageMagick $VERSION will be downloaded"

# download and install
echo "[INFO] Downloading ImageMagick"
wget "$URL" -O "$OUTPUT_FILE"
echo "[INFO] Extracting ImageMagick"
tar -xzf "$OUTPUT_FILE" -C "$OUTPUT_DIR"
echo "[INFO] Installing ImageMagick"
cd OUTPUT_DIR/ImageMagick
sudo ./configure
sudo checkinstall
echo "[INFO] Updating linkers"
sudo ldconfig
