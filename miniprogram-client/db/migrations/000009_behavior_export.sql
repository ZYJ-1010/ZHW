create table if not exists user_behavior_logs (
  id bigserial primary key,
  user_id bigint,
  event_type varchar(64) not null,
  target_type varchar(64) not null,
  target_id bigint,
  page_path varchar(255),
  keyword varchar(255),
  extra jsonb,
  created_at timestamptz not null default now()
);

create index if not exists idx_user_behavior_logs_user_created on user_behavior_logs(user_id, created_at desc);
create index if not exists idx_user_behavior_logs_event_created on user_behavior_logs(event_type, created_at desc);

create table if not exists export_templates (
  id bigserial primary key,
  code varchar(64) not null unique,
  name varchar(120) not null,
  export_type varchar(64) not null,
  file_name varchar(180) not null,
  columns_json jsonb not null,
  description text,
  enabled boolean not null default true,
  created_at timestamptz not null default now()
);

insert into export_templates (code, name, export_type, file_name, columns_json, description, enabled)
values
  (
    'reports_default',
    '举报申诉报表',
    'reports',
    'reports-export.csv',
    '["report_id","game_id","reporter_user_id","report_type","status","created_at"]'::jsonb,
    '固定格式导出举报申诉列表',
    true
  ),
  (
    'operation_logs_default',
    '操作日志报表',
    'operation_logs',
    'operation-logs-export.csv',
    '["log_id","admin_user_id","action","target_type","target_id","created_at"]'::jsonb,
    '固定格式导出后台操作日志',
    true
  )
on conflict (code) do update set
  name = excluded.name,
  export_type = excluded.export_type,
  file_name = excluded.file_name,
  columns_json = excluded.columns_json,
  description = excluded.description,
  enabled = excluded.enabled;

create table if not exists export_tasks (
  id bigserial primary key,
  task_no varchar(64) not null unique,
  template_code varchar(64) not null,
  export_type varchar(64) not null,
  status varchar(32) not null default 'pending',
  file_id bigint,
  created_by bigint,
  filters_json jsonb,
  fail_reason text,
  created_at timestamptz not null default now(),
  finished_at timestamptz
);

create index if not exists idx_export_tasks_created_by_status on export_tasks(created_by, status, created_at desc);
create index if not exists idx_export_tasks_status_created on export_tasks(status, created_at);
