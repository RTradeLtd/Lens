# 🔍 Lens

> Search engine for the distributed web

Lens is an opt-in search engine and data collection tool to aid content discovery
of the distributed web. It exposes a simple, minimal API for intelligently indexing
searching content on [IPFS](https://ipfs.io/).

[![GoDoc](https://godoc.org/github.com/RTradeLtd/Lens?status.svg)](https://godoc.org/github.com/RTradeLtd/Lens)
[![Build Status](https://travis-ci.com/RTradeLtd/Lens.svg?branch=master)](https://travis-ci.com/RTradeLtd/Lens)
[![codecov](https://codecov.io/gh/RTradeLtd/Lens/branch/master/graph/badge.svg)](https://codecov.io/gh/RTradeLtd/Lens) 
[![Go Report Card](https://goreportcard.com/badge/github.com/RTradeLtd/Lens)](https://goreportcard.com/report/github.com/RTradeLtd/Lens)
[![Latest Release](https://img.shields.io/github/release/RTradeLtd/Lens.svg?colorB=red)](https://github.com/RTradeLtd/Lens/releases)

## Features and Usage

Initially integrated with Temporal, Lens will allow users to optionally have the
data they upload be searched and indexed and be awarded with RTC for participating
in the data collection process. Users can then search for content using a
simple-to-use API.

Searching through Lens will be facilitated through [Temporal web](https://temporal.cloud/lens).
Optionally, we will have a service independent from Temporal which users can
submit content to have it be indexed. This however, is not compensated with RTC.
In order to receive the RTC, you must participate through Lens indexing within
the Temporal web interface.

### API

Lens exposes a simple API via [gRPC](https://grpc.io/). The definitions are in
[`RTradeLtd/grpc`](https://github.com/RTradeLtd/grpc/blob/master/lensv2/service.proto).

The Lens API, summarized, currently consists of three core RPCs:

```proto
service LensV2 {
  rpc Index(IndexReq)   returns (IndexResp)  {}
  rpc Search(SearchReq) returns (SearchResp) {}
  rpc Remove(RemoveReq) returns (RemoveResp) {}
}
```

Golang bindings for the Lens API can be found in
[`RTradeLtd/grpc`](https://github.com/RTradeLtd/grpc).

### Supported Formats

Only IPFS [CIDs](https://github.com/multiformats/cid) are supported, and they
must be plaintext files. We attempt to determine the content type via mime type
sniffing, and use that to determine whether or not we can analyze the content.

Please see the following table for supported content types that we can index.
Note if the type is listed as `<type>/*` it means that any "sub type" of that
mime type is supported.

| Mime Type        | Support Level | Tested Types             |
|------------------|---------------|--------------------------|
| `text/*`         | Beta          | `text/plain`, `text/html`|
| `image/*`        | Beta          | `image/jpeg`             |
| `application/pdf`| Beta          | `application/pdf`        |

## Deployment

The recommended way to deploy a Lens instance is via the
[`rtradetech/lens`](https://cloud.docker.com/u/rtradetech/repository/docker/rtradetech/lens)
Docker image.

```sh
$> docker pull rtradetech/lens:latest
```

A [`docker-compose`](https://docs.docker.com/compose/) [configuration](/lens.yml)
is available that also starts up other prerequisites:

```sh
$> wget -O lens.yml https://raw.githubusercontent.com/RTradeLtd/Lens/master/lens.yml
$> LENS=latest BASE=/my/dir docker-compose -f lens.yml up
```

## Development

This project requires:

* [Go 1.11+](https://golang.org/dl/)
* [dep](https://github.com/golang/dep#installation)
* [Tesseract](https://github.com/tesseract-ocr/tesseract#installing-tesseract)
* [Tensorflow](https://www.tensorflow.org/install)
* [go-fitz](https://github.com/gen2brain/go-fitz#install)

To fetch the codebase, use `go get`:

```sh
$> go get github.com/RTradeLtd/Lens
```

A rudimentary Makefile target [`make dep`](https://github.com/RTradeLtd/Lens/blob/master/Makefile#L13)
is available for installing the required dependencies.
