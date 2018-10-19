# Lens

[![GoDoc](https://godoc.org/github.com/RTradeLtd/Lens?status.svg)](https://godoc.org/github.com/RTradeLtd/Lens) [![codecov](https://codecov.io/gh/RTradeLtd/Lens/branch/master/graph/badge.svg)](https://codecov.io/gh/RTradeLtd/Lens) [![Build Status](https://travis-ci.com/RTradeLtd/Lens.svg?branch=master)](https://travis-ci.com/RTradeLtd/Lens)

Lens is an opt-in search engine and data collection tool to aid content discovery of the distributed web.

Currently we only support indexing of textual data, which we accomplish via `textrank`. Additional analyzers, and supported formats will be added over time.

## Textual Processing

The only language supported is english (for the moment) however we can easily add new language with packages like https://github.com/stopwords-iso

## Testing

Run `make testenv` to spin up our test IPFS node
Run `go test -v lens_test.go` to demo a basic end-to-end test