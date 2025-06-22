#!/bin/sh

# wait-for-db.sh
# Wait for PostgreSQL database to be ready

set -e

host="$1"
shift
port="$1"
shift
cmd="$@"

until pg_isready -h "$host" -p "$port" -U "${DB_USER:-admin}"; do
  >&2 echo "Database is unavailable - sleeping"
  sleep 1
done

>&2 echo "Database is up - executing command"
exec $cmd 