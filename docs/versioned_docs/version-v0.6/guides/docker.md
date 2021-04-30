---
id: docker
title: Docker Images
---

## Supported tags and respective `Dockerfile` links

- [`latest`, `v0.4.3-alpha.1`, `v0.4.3`, `v0.4`, `v0`](https://github.com/ory/kratos/blob/v0.4.3-alpha.1/.docker/Dockerfile)
- [`latest-sqlite`, `v0.4.3-alpha.1-sqlite`, `v0.4.3-sqlite`, `v0.4-sqlite`, `v0-sqlite`](https://github.com/ory/kratos/blob/v0.4.3-alpha.1/.docker/Dockerfile-sqlite)

## Image Variants

The `Kratos` Docker images come in two different flavors, one with and one
without SQLite support. All Docker images with the postfix
`kratos:<version>-sqlite` in the tag are compiled with embed SQLite support and
uses libmusl. All Docker images (`kratos:<version>`) without the postfix
`-sqlite` are compiled without SQLite support and therefore also don't include
libmusl.

If you don't make use of the embedded SQLite support we recommend to use the
Docker images without SQLite support as they are smaller in size, include fewer
libraries and therefore have a smaller attack surface.

## How to use these images

In order to make the provided Docker images as useful as possible they can be
configured through a set of supported Environment variables. In addition the
default configuration directory can be bound to a directory of choice to make it
simple to pass in your own configuration files.

### Environment Variables

#### `DSN`

This environment variable allows you to specify the database source name. As the
`DSN` normally consists of the url to the database system and the credentials to
access the database it is recommended to specify the `DSN` using a Environment
variable.

**Example:**

`docker run -e DSN="memory" oryd/kratos:latest`

#### `SECRETS_DEFAULT`

This environment variable allows you to specify the secret used to sign and
verify signatures and encrypt things:

**Example:**

`docker run -e SECRETS_DEFAULT="CHANGE-ME" oryd/kratos:v0.4.3-alpha.1`

### Volumes

If the file `$HOME/.kratos.yaml` exists, it will be used as the configuration
file. The provided Kratos Docker images currently do not include a default
configuration file, but make it easy to pass in your own configuration file(s)
by either binding a local directory or by creating your own custom Docker Image
and adding the configuration file(s) to the custom image.

#### Binding host directory

**Example:** In this example we start the standard Docker container with SQLite
support and use the quickstart email-password example configuration files by
bind mounting the local directory. This example assumes that you checked out the
Kratos Git repo and execute the Docker command in the Kratos Git repo directory:

```
docker run -it -e DSN="memory" \
       --mount type=bind,source="$(pwd)"/contrib/quickstart/kratos/email-password,target=/home/ory \
       oryd/kratos:latest-sqlite
```

In general we only recommend this approach for local development.

#### Creating custom Docker image

You can create your own, custom Kratos Docker images which embeds your
configuration files by simply using the official Kratos Docker images as the
Base Image and just adding your configuration file(s) as shown in the example
below:

```dockerfile
FROM oryd/kratos:latest
COPY contrib/quickstart/kratos/email-password/kratos.yml /ory/home
```

### Examples

Below you find different examples how to use the official Kratos Docker images.
