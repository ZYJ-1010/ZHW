#!/bin/sh
set -eu

for file in /app/db/migrations/*.sql; do
  echo "running migration: $file"
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -f "$file"
done

for file in /app/db/seeds/*.sql; do
  echo "running seed: $file"
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -f "$file"
done
