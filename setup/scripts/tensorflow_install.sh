#! /bin/bash

TENSORFLOW_URL="https://storage.googleapis.com/tensorflow/libtensorflow/libtensorflow-cpu-linux-x86_64-1.11.0.tar.gz"
OUTPUT_FILE="tensorflow.tar.gz"

echo "[INFO] Downloading tensorflow"
wget "$TENSORFLOW_URL" -O "$OUTPUT_FILE"

echo "[INFO] Extracting tensorflow"
sudo tar zxvf "$OUTPUT_FILE" -C /usr/local

echo "[INFO] Verifying install"
sudo ldconfig