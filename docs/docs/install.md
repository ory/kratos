---
id: install
title: Installation
---

Installing Ory Kratos on any system is straight forward. We provide prebuilt
binaries, Docker images, and support various package managers.

## Docker

We recommend using Docker to run Ory Kratos:

```shell
$ docker pull oryd/kratos:v0.8.0-alpha.3
$ docker run --rm -it oryd/kratos help
```

You can find more detailed information on the official Kratos docker images
[here](guides/docker.md).

## macOS

You can install Ory Kratos using [homebrew](https://brew.sh/) on macOS:

```shell
$ brew install ory/tap/kratos
$ kratos help
```

## Linux

On linux, you can use `bash <(curl ...)` to fetch the latest stable binary
using:

```shell
$ bash <(curl https://raw.githubusercontent.com/ory/meta/master/install.sh) -d -b . kratos v0.8.0-alpha.3
$ ./kratos help
```

You may want to move Ory Kratos to your `$PATH`:

```shell
$ sudo mv ./kratos /usr/local/bin/
$ kratos help
```

## Windows

You can install Ory Kratos using [scoop](https://scoop.sh) on Windows
(Powershell is required):

```shell
> scoop bucket add ory https://github.com/ory/scoop.git
> scoop install kratos
> kratos help
```

## Download Binaries

The client and server binaries are downloadable at the
[releases tab](https://github.com/ory/kratos/releases). There is currently no
installer available. You have to add the Kratos binary to the PATH environment
variable yourself or put the binary in a location that is already in your
`$PATH` (e.g. `/usr/local/bin`).

Once installed, you should be able to run:

```shell
$ kratos help
```

## Building From Source

If you wish to compile Ory Kratos yourself, you need to install and set up
[Go 1.12+](https://golang.org/) and add `$GOPATH/bin` to your `$PATH`.

The following commands will check out the latest release tag of Ory Kratos,
compile it, and set up flags so that `kratos version` works as expected. Please
note that this will only work with POSIX-compliant shells like `bash` or `sh`.

```shell
$ git clone https://github.com/ory/kratos.git
$ cd kratos
$ GO111MODULE=on make install
$ $(go env GOPATH)/bin/kratos help
```
