---
id: debug-docker-delve-ory-kratos
title: Debugging Ory Kratos in Docker with Delve
---

Very often, there is a need to debug Kratos being deployed as a Docker image. To
support this, Kratos ships with a couple of files:

- The `Dockerfile-debug` file, which you can find in the `.docker` directory.
- The `docker-compose.template.dbg` file, which you can find in the same
  directory. This file defines a template for a service, one would like to debug
  in Docker
- and a supplementary `debug-entrypoint.sh` skript, located in the `script`
  directory.

Actually, these files do not include any Kratos specifica and thus can be used
for any Golang based project. As you already could infer, this support is meant
to be used in a docker-compose setup as described below. You can however run it
as a standalone Docker container as well. You can find some information on how
to achieve this at the end of this document.

## As part of a docker-compose setup

Imagine you have the following project structure:

- docker-compose - a directory containing your `docker-compose.yaml` file
- kratos - a directory containing the Kratos code
- kratos-frontend - a directory containing a frontend application for Kratos

The `docker-compose.yml` mentioned above could look as follows:

```yaml
version: '3.7'

volumes:
  postgres-db:

services:
  postgresd:
    image: postgres:9.6
    ports:
      - '5432:5432'
    volumes:
      - type: volume
        source: postgres-db
        target: /var/lib/postgresql/data
        read_only: false
    environment:
      - PGDATA=/var/lib/postgresql/data/pgdata
      - POSTGRES_PASSWORD=secret
      - POSTGRES_USER=kratos

  kratos-migrate:
    image: kratos
    build:
      context: ../kratos
      dockerfile: ./.docker/Dockerfile-build
    environment:
      - DSN=postgres://kratos:secret@postgresd:5432/kratos?sslmode=disable&max_conns=20&max_idle_conns=4
    volumes:
      - type: bind
        source: path-to-kratos-config
        target: /etc/config/kratos
    command: migrate sql -e --yes
    depends_on:
      - postgresd

  kratos:
    image: kratos
    build:
      context: ../kratos
      dockerfile: ./.docker/Dockerfile-build
    depends_on:
      - kratos-migrate
    ports:
      - '4433:4433' # public
      - '4434:4434' # admin
    command: serve -c /etc/config/kratos/kratos.yml --watch-courier --dev
    volumes:
      - type: bind
        source: path-to-kratos-config
        target: /etc/config/kratos

  kratos-frontend:
    image: kratos-frontend
    build:
      context: ../kratos-kratos-frontend
      dockerfile: ./Dockerfile
    env_file:
      - file-containing-all-required-configuration.env
```

To enable debugging of Kratos without changing the above docker-compose file,
you can do the following (from the docker-compose directory):

```bash
SERVICE_NAME=kratos SERVICE_ROOT=../kratos REMOTE_DEBUGGING_PORT=9999 envsubst < ../kratos/.docker/docker-compose.template.dbg \
  > docker-compose.kratos.tmp
docker-compose -f docker-compose.yaml -f docker-compose.kratos.tmp up --build -d kratos
```

The first line will create an overwrite docker-compose file to have a debug
configuration for the kratos service. The second line will start a debug
container by

- mounting your `kratos` directory into the resulting Docker container,
- downloading Delve,
- building Kratos inside the container,
- starting it in Delve with the arguments, you've defined in your regular
  docker-compose file - in the example above, this would be
  `serve -c /etc/config/kratos/kratos.yml --watch-courier --dev` - and
- watching for changes on any go file within the mounted code base.

Each time you change a .go file, the Delve process will be stopped, Kratos will
be recompiled and Delve will be started again. With other words, you'll have to
re-connect with your debugger again after each change.

As you can see from the above usage, the `docker-compose.template.dbg` template
expects the following variables to be defined:

- `SERVICE_ROOT` - the root directory of the service to be started in the debug
  mode.
- `SERVICE_NAME` - the name of the service from the docker-compose file.
- `REMOTE_DEBUGGING_PORT` - the host port, the Delve listening port should be
  exposed as. This is the port you should connect your remote debugger to.

If you run docker-compose this way, the container run with debugging enabled
will wait until the debugger connects. If your IDE supports remote debugging,
set host to `localhost` and port to the value, you've used for
`REMOTE_DEBUGGING_PORT` in your remote debugging configuration.

## As a standalone Docker container

If you just would like to start Kratos in a container in debug mode, you can
just use the `Dockerfile-debug` file instead of the regular `Dockerfile`. Make
however sure your build context in the root directory of Kratos and not the
`.docker` directory. In your IDE the debug configuration has to reference that
file. In addition, you'll have to expose the Delve service port 40000 under the
port 8000, as well as the actual port of the service, you'll like to access from
your host, configure the bind mounts and set the run options to
`--security-opt="apparmor=unconfined" --cap-add=SYS_PTRACE`.
