alter table games
  add column if not exists start_reason text not null default '',
  add column if not exists started_by_user_id bigint,
  add column if not exists started_at timestamptz;
