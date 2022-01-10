---
id: docker
title: Docker Images
---

## Supported tags and respective `Dockerfile` links

- [`latest`, `v0.8.0-alpha.1`, `v0.8.0`, `v0.8`, `v0`](https://github.com/ory/kratos/blob/master/.docker/Dockerfile-alpine)

## Image Variants

The `Kratos` Docker images use Alpine Linux as their base image and come with
SQLite support build in.

## How to use these images

In order to make the provided Docker images as useful as possible they can be
configured through a set of supported Environment variables. In addition the
default configuration directory can be bound to a directory of choice to make it
simple to pass in your own configuration files.

## Do Not Use `latest`

Please, always use a tagged version and never use `latest` Docker images. This
ensures that your deployment does not unexpectedly update with an incompatible
version!

## Running Migrations

To run SQL Migrations, which are required for new installations and when
upgrading, run:

```shell
docker -e DSN="<your database URL>" run oryd/kratos:<version> migrate sql -e
```

### Environment Variables

#### `DSN`

This environment variable allows you to specify the database source name. As the
`DSN` normally consists of the url to the database system and the credentials to
access the database it is recommended to specify the `DSN` using a Environment
variable.

**Example:**

```
docker run -e DSN="memory" oryd/kratos:<version>
```

#### `SECRETS_DEFAULT`

This environment variable allows you to specify the secret used to sign and
verify signatures and encrypt things:

**Example:**

`docker run -e SECRETS_DEFAULT="CHANGE-ME" oryd/kratos:<version>`

### Volumes

The provided Kratos Docker images currently do not include a default
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
   oryd/kratos:<version>
```

In general we only recommend this approach for local development.

#### Creating custom Docker image

You can create your own, custom Kratos Docker images which embeds your
configuration files by simply using the official Kratos Docker images as the
Base Image and just adding your configuration file(s) as shown in the example
below:

```dockerfile
FROM oryd/kratos:latest
COPY contrib/quickstart/kratos/email-password/kratos.yml /home/ory
```

**Note that in both cases**, you must supply the location of the configuration
file using the `--config` flag when running the container.

```
$ docker run <theimage> --config /home/ory/kratos.yml
```
