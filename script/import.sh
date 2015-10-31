#!/bin/bash

set -e

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

db_container=$1

dbuser=registrar
dbname=registrar_development
connection_opts='-h "$POSTGRES_PORT_5432_TCP_ADDR" -p "$POSTGRES_PORT_5432_TCP_PORT" -U postgres'

function postgres_run() {
  docker run -it --rm \
    --link $db_container:postgres \
    --volume $DIR:/db \
    postgres \
    sh -c "exec $1"
}

function psql_run() {
  echo "$1" | docker run -i --rm \
    --link $db_container:postgres \
    --volume $DIR:/db \
    postgres \
    sh -c "exec psql $connection_opts -d $dbname"
}

function createuser() {
  echo "Creating user '$dbuser'..."
  postgres_run "createuser $connection_opts -DRS --login --encrypted --interactive --pwprompt $dbuser"
}

function createdb() {
  postgres_run "createdb $connection_opts $dbname"
}

function dbimport() {
  postgres_run "psql $connection_opts -d $dbname < /db/schema.sql"
}

function grant() {
  psql_run "GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO $dbuser;"
  psql_run "GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO $dbuser;"
}

createuser
createdb
dbimport
grant
