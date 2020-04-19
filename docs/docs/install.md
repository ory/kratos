---
id: install
title: Installation
---

Installing ORY Kratos on any system is straight forward. We provide pre-built
binaries, Docker Images and support various package managers.

## Docker

We recommend using Docker to run ORY Kratos:

```shell
$ docker pull oryd/kratos:v0.1.1-alpha.1
$ docker run --rm -it oryd/kratos help
```

## macOS

You can install ORY Kratos using [homebrew](https://brew.sh/) on macOS:

```shell
$ brew tap ory/kratos
$ brew install ory/kratos/kratos
$ kratos help
```

## Linux

On linux, you can use `bash <(curl ...)` to fetch the latest stable binary
using:

```shell
$ bash <(curl https://raw.githubusercontent.com/ory/kratos/master/install.sh) -b . v0.1.1-alpha.1
$ ./kratos help
```

You may want to move ORY Kratos to your `$PATH`:

```shell
$ sudo mv ./kratos /usr/local/bin/
$ kratos help
```

## Windows

You can install ORY Kratos using [scoop](https://scoop.sh) on Windows:

```shell
> scoop bucket add ory-kratos https://github.com/ory/scoop-kratos.git
> scoop install kratos
> kratos help
```

## Download Binaries

The client and server **binaries are downloadable at the
[releases tab](https://github.com/ory/kratos/releases)**. There is currently no
installer available. You have to add the Kratos binary to the PATH environment
variable yourself or put the binary in a location that is already in your
`$PATH` (e.g. `/usr/local/bin`, ...).

Once installed, you should be able to run:

```shell
$ kratos help
```

## Building from Source

If you wish to compile ORY Kratos yourself, you need to install and set up
[Go 1.12+](https://golang.org/) and add `$GOPATH/bin` to your `$PATH`.

The following commands will check out the latest release tag of ORY Kratos and
compile it and set up flags so that `kratos version` works as expected. Please
note that this will only work with a linux shell like bash or sh.

```shell
$ go get -d -u github.com/ory/kratos
$ cd $(go env GOPATH)/src/github.com/ory/kratos
$ GO111MODULE=on make install
$ $(go env GOPATH)/bin/kratos help
```
