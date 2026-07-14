alter table user_behavior_logs
  add column if not exists event_code varchar(64),
  add column if not exists business_type varchar(64),
  add column if not exists business_id bigint,
  add column if not exists source varchar(32) not null default 'app',
  add column if not exists device varchar(128),
  add column if not exists ip varchar(64),
  add column if not exists occurred_at timestamptz;

update user_behavior_logs
set
  event_code = coalesce(event_code, event_type),
  business_type = coalesce(business_type, target_type),
  business_id = coalesce(business_id, target_id),
  occurred_at = coalesce(occurred_at, created_at)
where event_code is null
   or business_type is null
   or occurred_at is null;

alter table user_behavior_logs
  alter column event_code set not null,
  alter column business_type set not null,
  alter column occurred_at set not null;

create index if not exists idx_user_behavior_logs_event_code_occurred
  on user_behavior_logs(event_code, occurred_at desc);

create index if not exists idx_user_behavior_logs_user_occurred
  on user_behavior_logs(user_id, occurred_at desc);

create index if not exists idx_user_behavior_logs_business
  on user_behavior_logs(business_type, business_id, occurred_at desc);
