# Lens

[![GoDoc](https://godoc.org/github.com/RTradeLtd/Lens?status.svg)](https://godoc.org/github.com/RTradeLtd/Lens) [![codecov](https://codecov.io/gh/RTradeLtd/Lens/branch/master/graph/badge.svg)](https://codecov.io/gh/RTradeLtd/Lens) [![Build Status](https://travis-ci.com/RTradeLtd/Lens.svg?branch=master)](https://travis-ci.com/RTradeLtd/Lens) [![Go Report Card](https://goreportcard.com/badge/github.com/RTradeLtd/Lens)](https://goreportcard.com/report/github.com/RTradeLtd/Lens)

Lens is an opt-in search engine and data collection tool to aid content discovery of the distributed web. Initially integrated with TEMPORAL, Lens will allow users to optionally have the data they upload be searched and indexed and be awarded with RTC for participating in the data collection process. Users can then search for "keywords" of content, such as "document" or "api". Lens will then use this keyword to retrieve all content which matched.

Searching through Lens will be facilitated through the TEMPORAL web interface. Optionally, we will have a service independent from TEMPORAL which users can submit content to have it be indexed. This however, is not compensated with RTC. In order to receive the RTC, you must participate through Lens indexing within the TEMPORAL web interface.

## Supported Formats

Only IPFS CIDs are supported, and they must be plaintext files. We attempt to determine the content type via mime type sniffing, and use that to determine whether or not we can analyze the content.

Please see the following table for supported content types that we can index. Note if the type is listed as `<type>/*` it means that any "sub type" of that mime type is supported.

| Mime Type        | Support Level | Tested Types             |
|------------------|---------------|--------------------------|
| `text/*`         | Alpha         | `text/plain`, `text/html`|
| `image/*`        | Alpha         | `image/jpeg`             |
| `application/pdf`| Alpha         | `application/pdf`        |

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

For image indexing, we currently run the images against pre-trained InceptionV5 tensorflow models. In the future we will more than likely migrate to models we train ourselves, leveraging our extensive GPU computing infrastructure.

## Searching

1) When receiving a search request, we are simply provided with a list of keywords to search through.
2) Using these keywords, we then search through badgerds to see if these keywords have been seen before. If they have, we then pull a list of all lens identifiers that can be matched by this keyword. 
3) After repeating step 2 for all keywords, we then search through badgerds to find the objects that the lens identifiers refer to
4) The user is then sent a list of all object names (ie, ipfs content hashes) for which.
