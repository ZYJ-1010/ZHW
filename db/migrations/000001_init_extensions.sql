create extension if not exists "uuid-ossp";
create extension if not exists "pg_trgm";

create table if not exists schema_migrations_marker (
  id bigserial primary key,
  version varchar(64) not null unique,
  description varchar(255) not null,
  applied_at timestamptz not null default now()
);

insert into schema_migrations_marker(version, description)
values ('000001', 'init extensions and migration marker')
on conflict (version) do nothing;

