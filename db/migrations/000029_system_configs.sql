create table if not exists system_configs (
  id bigserial primary key,
  config_key varchar(128) not null unique,
  config_value jsonb not null default '{}'::jsonb,
  status varchar(32) not null default 'active',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists idx_system_configs_status on system_configs(status);
