# Lens

[![GoDoc](https://godoc.org/github.com/RTradeLtd/Lens?status.svg)](https://godoc.org/github.com/RTradeLtd/Lens) [![codecov](https://codecov.io/gh/RTradeLtd/Lens/branch/master/graph/badge.svg)](https://codecov.io/gh/RTradeLtd/Lens) [![Build Status](https://travis-ci.com/RTradeLtd/Lens.svg?branch=master)](https://travis-ci.com/RTradeLtd/Lens)

Lens is an opt-in search engine and data collection tool to aid content discovery of the distributed web. Initially integrated with TEMPORAL, Lens will allow users to optionally have the data they upload be searched and indexed and be awarded with RTC for participating in the data collection process. Users can then search for "keywords" of content, such as "document" or "api". Lens will then use this keyword to retrieve all content which matched.

Searching through Lens will be facilitated through the TEMPORAL web interface. Optionally, we will have a service independent from TEMPORAL which users can submit content to have it be indexed. This however, is not compensated with RTC. In order to receive the RTC, you must participate through Lens indexing within the TEMPORAL web interface.

## Installation

1) Get RTFS dependencies `go get -u -v github.com/RTradeLtd/RTFS`
2) Get Lens `go get -u -v github.com/RTradeLtd/Lens`
3) `make cli` this will generate the lens binary (does basically nothing at this point)

## Processing

Currently, we will only support indexing of content from IPFS. Within this, right now it is liimited to processing of content for which we can extract text from. This text will then be sent through TextRank to extract keywords which can then be used to search for the content. In the future we will expand to support additional content types. It could even be extended such that we do image indexing, leveraging our GPU computing infrastructure to process images, or video data.

## Testing

1) Build the testenvironment with `make testenv`
2) Build the command line tool with `make cli`
3) Start the testenvironment with `docker-compose -f lens.yml up`
4) In a seperate shell, run `./temporal-lens client` to show a small example of a basic index request, and search request

Note that the first time running `./temporal-lens client` with a fresh badgerds instance will cause the search to fail. Simply re-run `./temporal-lens client` and the search will succeed. The badgerds instance is stored in `/data/lens/badgerds-lens` to enable easy backup.