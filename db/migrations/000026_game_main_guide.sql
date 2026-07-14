alter table games
  add column if not exists main_guide_user_id bigint;

create index if not exists idx_games_main_guide on games(main_guide_user_id);
