alter table game_applications
  add column if not exists reject_reason text not null default '';
