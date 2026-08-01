create table if not exists user_login_activity (
  user_id bigint primary key,
  last_app_login_at timestamptz not null,
  updated_at timestamptz not null default now()
);

create index if not exists idx_user_login_activity_last_app_login on user_login_activity (last_app_login_at desc);
