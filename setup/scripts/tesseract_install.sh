#! /bin/bash

set -e

# detect other platforms
if [[ "$OSTYPE" == "darwin"* ]]; then
  brew install tesseract
else
  sudo apt-get update -qq
  sudo apt-get install -y -qq libtesseract-dev libleptonica-dev tesseract-ocr-eng
fi
