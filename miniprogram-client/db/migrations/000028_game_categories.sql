alter table games
  add column if not exists primary_category varchar(32),
  add column if not exists primary_category_text varchar(32),
  add column if not exists secondary_category varchar(32),
  add column if not exists secondary_category_text varchar(32),
  add column if not exists type varchar(32);

create index if not exists idx_games_primary_category on games(primary_category);
create index if not exists idx_games_secondary_category on games(secondary_category);
create index if not exists idx_games_type on games(type);
