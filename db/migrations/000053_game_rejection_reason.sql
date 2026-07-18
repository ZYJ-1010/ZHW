alter table games
  add column if not exists reject_reason text not null default '';
