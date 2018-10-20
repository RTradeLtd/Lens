# Lens

[![GoDoc](https://godoc.org/github.com/RTradeLtd/Lens?status.svg)](https://godoc.org/github.com/RTradeLtd/Lens) [![codecov](https://codecov.io/gh/RTradeLtd/Lens/branch/master/graph/badge.svg)](https://codecov.io/gh/RTradeLtd/Lens) [![Build Status](https://travis-ci.com/RTradeLtd/Lens.svg?branch=master)](https://travis-ci.com/RTradeLtd/Lens)

Lens is an opt-in search engine and data collection tool to aid content discovery of the distributed web. Initially integrated with TEMPORAL, Lens will allow users to optionally have the data they upload be searched and indexed and be awarded with RTC for participating in the data collection process. Users can then search for "keywords" of content, such as "document" or "api". Lens will then use this keyword to retrieve all content which matched.

Searching through Lens will be facilitated through the TEMPORAL web interface. Optionally, we will have a service independent from TEMPORAL which users can submit content to have it be indexed. This however, is not compensated with RTC. In order to receive the RTC, you must participate through Lens indexing within the TEMPORAL web interface.

## Installation

NOTE: All commands after `#1` need to be ran from the root directroy of this repo, ie `$GOPATH/src/github.com/RTradeLtd/Lens`

1) Download this repository via git or `go get -u -v github.com/RTradeLtd/Lens/...`
2) Run `docker-compose -f lens.yml up` to create a test environment with default settings (lens grpc server of `0.0.0.0:9998`, with an ipfs api of `127.0.0.1:5001`)
3) Build the client with `make cli`
4) Run the client with `./temporal-lens client` and follow any instructions (if asked to set `CONFIG_DAG` set it to `test/config.json`)

Note, after starting up the lens container once, you may change its configuration settings located at `/data/lens/config.json`

## Processing

Currently, we will only support indexing of content from IPFS. Within this, right now it is liimited to processing of content for which we can extract text from. This text will then be sent through TextRank to extract keywords which can then be used to search for the content. In the future we will expand to support additional content types. It could even be extended such that we do image indexing, leveraging our GPU computing infrastructure to process images, or video data.

## Testing

1) Build the testenvironment with `make testenv`
2) Build the command line tool with `make cli`
3) Start the testenvironment with `docker-compose -f lens.yml up`
4) In a seperate shell, run `./temporal-lens client` to show a small example of a basic index request, and search request

Note that the first time running `./temporal-lens client` with a fresh badgerds instance will cause the search to fail. Simply re-run `./temporal-lens client` and the search will succeed. The badgerds instance is stored in `/data/lens/badgerds-lens` to enable easy backup.

To use the lens gRPC API Server, set the env var `LENS_IP` to the IP address of the lens GRPC server, and `LENS_PORT` to the port of the lens GRPC server