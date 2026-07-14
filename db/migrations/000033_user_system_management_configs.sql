create table if not exists user_system_management_configs (
  user_id bigint not null,
  config_key varchar(80) not null,
  config_value jsonb not null default '{}'::jsonb,
  updated_at timestamptz not null default now(),
  primary key (user_id, config_key)
);

create index if not exists idx_user_system_management_configs_key on user_system_management_configs(config_key, updated_at desc);
