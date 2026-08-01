create table if not exists user_newbie_guide_progress (
  user_id bigint primary key,
  profile_reminder_count integer not null default 0 check (profile_reminder_count >= 0),
  last_profile_reminder_at timestamptz,
  updated_at timestamptz not null default now()
);

comment on table user_newbie_guide_progress is '新手引导资料提醒展示进度；不代表任务已完成';
