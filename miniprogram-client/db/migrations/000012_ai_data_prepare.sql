create table if not exists ai_data_snapshots (
  id bigserial primary key,
  behavior_log_count int not null default 0,
  favorite_count int not null default 0,
  review_count int not null default 0,
  connection_count int not null default 0,
  expert_profile_count int not null default 0,
  guide_profile_count int not null default 0,
  im_message_count int not null default 0,
  data_ready boolean not null default false,
  im_export_enabled boolean not null default false,
  created_at timestamptz not null default now()
);

create table if not exists ai_data_export_tasks (
  id bigserial primary key,
  export_type varchar(64) not null,
  status varchar(32) not null default 'blocked',
  blocked_reason text,
  created_by_admin_id bigint,
  created_at timestamptz not null default now(),
  completed_at timestamptz
);

create index if not exists idx_ai_data_export_tasks_type_created on ai_data_export_tasks(export_type, created_at desc);
