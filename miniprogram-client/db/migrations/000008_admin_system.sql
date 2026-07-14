create table if not exists admin_users (
  id bigserial primary key,
  username varchar(64) not null unique,
  password_hash varchar(255) not null,
  status varchar(32) not null default 'active',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists admin_roles (
  id bigserial primary key,
  role_code varchar(64) not null unique,
  role_name varchar(100) not null,
  created_at timestamptz not null default now()
);

create table if not exists admin_permissions (
  id bigserial primary key,
  permission_code varchar(100) not null unique,
  permission_name varchar(100) not null,
  created_at timestamptz not null default now()
);

create table if not exists admin_user_roles (
  admin_user_id bigint not null,
  role_id bigint not null,
  created_at timestamptz not null default now(),
  primary key (admin_user_id, role_id)
);

create table if not exists admin_role_permissions (
  role_id bigint not null,
  permission_id bigint not null,
  created_at timestamptz not null default now(),
  primary key (role_id, permission_id)
);

create table if not exists operation_logs (
  id bigserial primary key,
  admin_user_id bigint,
  action varchar(100) not null,
  target_type varchar(64),
  target_id varchar(64),
  request_id varchar(128),
  ip varchar(64),
  detail_json jsonb,
  created_at timestamptz not null default now()
);

create index if not exists idx_operation_logs_admin_created on operation_logs(admin_user_id, created_at desc);
create index if not exists idx_operation_logs_action_created on operation_logs(action, created_at desc);

create table if not exists reports (
  id bigserial primary key,
  game_id bigint not null,
  reporter_user_id bigint not null,
  target_user_id bigint,
  report_type varchar(64) not null,
  content text,
  status varchar(32) not null default 'pending',
  chat_message_id bigint,
  file_id bigint,
  review_id bigint,
  revenue_record_id bigint,
  revenue_frozen boolean not null default false,
  revenue_freeze_note varchar(128),
  handler_admin_id bigint,
  handle_result text,
  handled_at timestamptz,
  created_at timestamptz not null default now()
);

create index if not exists idx_reports_game_status on reports(game_id, status);
create index if not exists idx_reports_reporter on reports(reporter_user_id, created_at desc);

create table if not exists notifications (
  id bigserial primary key,
  user_id bigint not null,
  notify_type varchar(64) not null,
  title varchar(120) not null,
  content text,
  biz_type varchar(64),
  biz_id bigint,
  status varchar(32) not null default 'unread',
  need_wechat boolean not null default false,
  wechat_state varchar(32),
  wechat_template_id varchar(128),
  wechat_task_id bigint,
  created_at timestamptz not null default now(),
  read_at timestamptz
);

create index if not exists idx_notifications_user_status_created on notifications(user_id, status, created_at desc);
create index if not exists idx_notifications_biz on notifications(biz_type, biz_id);

create table if not exists wechat_subscribe_templates (
  id bigserial primary key,
  scene varchar(64) not null unique,
  template_id varchar(128) not null,
  title varchar(120) not null,
  status varchar(32) not null default 'active',
  created_at timestamptz not null default now()
);

create table if not exists wechat_subscribe_tasks (
  id bigserial primary key,
  notification_id bigint not null,
  user_id bigint not null,
  scene varchar(64) not null,
  template_id varchar(128) not null,
  status varchar(32) not null default 'pending',
  request_payload jsonb,
  result_code varchar(64),
  result_message text,
  created_at timestamptz not null default now(),
  sent_at timestamptz
);

create index if not exists idx_wechat_subscribe_tasks_status_created on wechat_subscribe_tasks(status, created_at);
create index if not exists idx_wechat_subscribe_tasks_notification on wechat_subscribe_tasks(notification_id);
