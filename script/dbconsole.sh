#!/bin/sh

set -e

db_container=$1

dbname=registrar_development
connection_opts='-h "$POSTGRES_PORT_5432_TCP_ADDR" -p "$POSTGRES_PORT_5432_TCP_PORT" -U postgres'

function psql_run() {
  docker run -it --rm \
    --link $db_container:postgres \
    postgres \
    sh -c "exec $1"
}

function dbconsole() {
  psql_run "psql $connection_opts -d $dbname"
}

dbconsole
