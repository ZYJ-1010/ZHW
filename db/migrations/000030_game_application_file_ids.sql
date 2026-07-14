alter table game_applications
  add column if not exists file_ids jsonb not null default '[]'::jsonb;
