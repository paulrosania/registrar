#!/bin/sh

db_container=$1

connection_opts='-h "$POSTGRES_PORT_5432_TCP_ADDR" -p "$POSTGRES_PORT_5432_TCP_PORT" -U postgres'

function psql_run() {
  docker run -it --rm \
    --link $db_container:postgres \
    postgres \
    sh -c "exec $1"
}

function createuser() {
  psql_run "createuser $connection_opts -DRS --login --encrypted --interactive --pwprompt"
}

createuser
