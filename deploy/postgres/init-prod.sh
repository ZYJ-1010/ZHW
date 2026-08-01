#!/bin/sh
set -eu

for file in /app/db/migrations/*.sql; do
  echo "running migration: $file"
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -f "$file"
done

# 新库初始化完成后写入迁移台账。后续版本由 migrate 服务按台账执行，
# 不再依赖 PostgreSQL 数据目录是否首次创建。
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<'SQL'
create table if not exists schema_migrations (
  version varchar(255) primary key,
  applied_at timestamptz not null default now()
);
SQL

for file in /app/db/migrations/*.sql; do
  version=$(basename "$file")
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" \
    -v version="$version" \
    -c "insert into schema_migrations(version) values (:'version') on conflict (version) do nothing"
done

for file in /app/db/seeds/*.sql; do
  echo "running seed: $file"
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -f "$file"
done
