#! /bin/bash

set -e

# detect other platforms
if [[ "$OSTYPE" == "darwin"* ]]; then
  brew install tesseract
else
  sudo apt install tesseract-ocr
fi
