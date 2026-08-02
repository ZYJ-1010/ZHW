#!/bin/sh
set -eu

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<'SQL'
create table if not exists schema_migrations (
  version varchar(255) primary key,
  applied_at timestamptz not null default now()
);
SQL

for file in /app/db/migrations/*.sql; do
  version=$(basename "$file")
  applied=$(psql -At --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" \
    -c "select 1 from schema_migrations where version = '$version'")

  if [ "$applied" = "1" ]; then
    continue
  fi

  echo "applying migration: $version"
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -f "$file"
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" \
    -c "insert into schema_migrations(version) values ('$version')"
done
