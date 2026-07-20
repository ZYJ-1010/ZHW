create table if not exists game_create_drafts (
  id bigserial primary key,
  creator_user_id bigint not null,
  title varchar(100) not null default '未命名组局',
  payload jsonb not null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists idx_game_create_drafts_creator_updated
  on game_create_drafts(creator_user_id, updated_at desc, id desc);
