#!/bin/sh
set -x
#docker run -it -e DSN="sqlite3://:memory:?_fk=true" --mount type=bind,source=/Users/andreas/Documents/dev/ory.sh/kratos-config,target=/home/ory oryd/mykratos:latest-sqlite
docker run -it -e DSN="memory" --mount type=bind,source=/Users/andreas/Documents/dev/ory.sh/kratos-config,target=/home/ory oryd/mykratos:latest-sqlite