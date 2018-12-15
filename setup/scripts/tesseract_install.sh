#! /bin/bash

set -e

LINKER=ldconfig

# detect other platforms
if [[ "$OSTYPE" == "darwin"* ]]; then
  LINKER=update_dyld_shared_cache
  echo "[INFO] Installing Tesseract using Homebrew"
  brew install tesseract
else
  echo "[INFO] Installing Tesseract using apt-get"
  sudo apt-get update -qq
  sudo apt-get install -y -qq libtesseract-dev libleptonica-dev tesseract-ocr-eng
fi

echo "[INFO] Updating linker"
sudo $LINKER
