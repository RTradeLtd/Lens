# 🔍 Lens

> 一个分布式互联网中的搜索引擎

Lens既是服务于分布式互联网的搜索引擎，又是一个数据收集工具。它公开了一个简单小巧的API接口用于智能地查找[IPFS](https://ipfs.io/)上的内容。

[![GoDoc](https://godoc.org/github.com/RTradeLtd/Lens?status.svg)](https://godoc.org/github.com/RTradeLtd/Lens)
[![Build Status](https://travis-ci.com/RTradeLtd/Lens.svg?branch=master)](https://travis-ci.com/RTradeLtd/Lens)
[![codecov](https://codecov.io/gh/RTradeLtd/Lens/branch/master/graph/badge.svg)](https://codecov.io/gh/RTradeLtd/Lens) 
[![Go Report Card](https://goreportcard.com/badge/github.com/RTradeLtd/Lens)](https://goreportcard.com/report/github.com/RTradeLtd/Lens)
[![Latest Release](https://img.shields.io/github/release/RTradeLtd/Lens.svg?colorB=red)](https://github.com/RTradeLtd/Lens/releases)

## 多语言

[![](https://img.shields.io/badge/Lang-English-blue.svg)](README.md)  [![jaywcjlove/sb](https://jaywcjlove.github.io/sb/lang/chinese.svg)](README-zh.md)

## 特性与用例

Lens最初是与Temporal配合使用的，用户可以在使用Temporal时选择是否将他们上传的数据被Lens索引，并在贡献数据的同时获得RTC奖励。然后，用户可以使用一个简单易用的API来搜索数据内容。

在[Temporal web](https://temporal.cloud/lens)中使用Lens进行搜索将会非常有益，并且可以获得RTC通证奖励。当然，我们也赋予了Lens独立部署和使用的服务，用户可以单独使用Lens进行内容录入和搜索服务，但这种方式并不能获取RTC通证奖励。


### API编程接口

Lens基于[gRPC](https://grpc.io/)暴露了一个简单的API接口。 定义如下：
[`RTradeLtd/grpc`](https://github.com/RTradeLtd/grpc/blob/master/lensv2/service.proto).

Lens API的核心RPCs如下：

```proto
service LensV2 {
  rpc Index(IndexReq)   returns (IndexResp)  {}
  rpc Search(SearchReq) returns (SearchResp) {}
  rpc Remove(RemoveReq) returns (RemoveResp) {}
}
```

可以在[`RTradeLtd/grpc`](https://github.com/RTradeLtd/grpc)中找到。

### 编码支持

只支持IPFS[CIDs](https://github.com/multiformats/cid) 作为搜索输入值, 并且搜索结果仅支持图片，文本，和pdf文件。我们正尝试通过数据类型智能嗅探技术来解析更多内容类型。

下面表格中是我们所支持检索的文件格式：

| Mime Type        | Support Level | Tested Types             |
|------------------|---------------|--------------------------|
| `text/*`         | Beta          | `text/plain`, `text/html`|
| `image/*`        | Beta          | `image/jpeg`             |
| `application/pdf`| Beta          | `application/pdf`        |

## 部署

基于Docker命令行的部署方式如下
[`rtradetech/lens`](https://cloud.docker.com/u/rtradetech/repository/docker/rtradetech/lens)


```sh
$> docker pull rtradetech/lens:latest
```

A[`docker-compose`](https://docs.docker.com/compose/) [configuration](/lens.yml)
配置信息如下：

```sh
$> wget -O lens.yml https://raw.githubusercontent.com/RTradeLtd/Lens/master/lens.yml
$> LENS=latest BASE=/my/dir docker-compose -f lens.yml up
```

## 参与开发

这个项目依赖于:

* [Go 1.11+](https://golang.org/dl/)
* [dep](https://github.com/golang/dep#installation)
* [Tesseract](https://github.com/tesseract-ocr/tesseract#installing-tesseract)
* [Tensorflow](https://www.tensorflow.org/install)
* [go-fitz](https://github.com/gen2brain/go-fitz#install)

使用 `go get` 获取代码库:

```sh
$> go get github.com/RTradeLtd/Lens
```

通过我们所提供的 [`make dep`](https://github.com/RTradeLtd/Lens/blob/master/Makefile#L13)可以一键安装所需依赖。
