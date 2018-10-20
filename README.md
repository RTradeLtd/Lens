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

## Supported Formats

Only IPFS CIDs are supported, and they must be plaintext files. We attempt to determine the content type via mime type sniffing, and if it results in anything other than `text/plain` and indexing requests will be rejected.

## Processing

We support two types of processing, index and search requests

### Indexing

1) When receiving an index request, we check to make sure the object to be indexed is a supported data type.
2) We then attempt to determine the mime type of whatever object is being indexed, and validate it to make sure its a supported format.
3) We then extract consumable data from the object through an `xtractor` service.
4) After extracting usable data, we then send it to an `analyzer` service which is responsible for analyzer content to create meta-data
5) After the meta-data is generated, we then pass it onto the core of the lens service
6) The lens service is responsible for creating lens objects, which are valid IPLD objects, and storing them within IPFS, and within a local badgerds instance

The following objects are created during an indexing request:

Keyword Object:

* A keyword object contains all of the Lens Identifiers for content that can be searched for with this keyword

Object:

* An object is content that was indexed, and includes a Lens Identifier for this content within the lens system (note, this is simply to enable easy lookup and is not valid outside of Lens)
* Also includes are all the keywords that can be used to search for this particular content

## Searching

1) When receiving a search request, we are simply provided with a list of keywords to search through.
2) Using these keywords, we then search through badgerds to see if these keywords have been seen before. If they have, we then pull a list of all lens identifiers that can be matched by this keyword. 
3) After repeating step 2 for all keywords, we then search through badgerds to find the objects that the lens identifiers refer to
4) The user is then sent a list of all object names (ie, ipfs content hashes) for which 

## Testing

1) Build the testenvironment with `make testenv`
2) Build the command line tool with `make cli`
3) Start the testenvironment with `docker-compose -f lens.yml up`
4) In a seperate shell, run `./temporal-lens client` to show a small example of a basic index request, and search request

Note that the first time running `./temporal-lens client` with a fresh badgerds instance will cause the search to fail. Simply re-run `./temporal-lens client` and the search will succeed. The badgerds instance is stored in `/data/lens/badgerds-lens` to enable easy backup.

To use the lens gRPC API Server, set the env var `LENS_IP` to the IP address of the lens GRPC server, and `LENS_PORT` to the port of the lens GRPC server