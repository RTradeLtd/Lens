#! /bin/bash

set -e

# detect other platforms
if [[ "$OSTYPE" == "darwin"* ]]; then
  brew install tesseract
else
  sudo apt-get -qq update
  sudo apt-get install -y tesseract-ocr libleptonica-dev libtesseract-dev libjpeg8
fi
