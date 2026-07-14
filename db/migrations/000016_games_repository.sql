alter table games
  add column if not exists longitude decimal(10,7),
  add column if not exists latitude decimal(10,7);

create table if not exists game_members (
  game_id bigint not null,
  user_id bigint not null,
  role varchar(32) not null default 'member',
  status varchar(32) not null default 'active',
  quit_reason varchar(64),
  credit_deducted boolean not null default false,
  credit_log_id bigint,
  joined_at timestamptz not null default now(),
  primary key (game_id, user_id)
);

create index if not exists idx_game_members_user on game_members(user_id, joined_at desc);

create table if not exists game_applications (
  id bigserial primary key,
  game_id bigint not null,
  user_id bigint not null,
  status varchar(32) not null default 'pending',
  reason text,
  created_at timestamptz not null default now(),
  reviewed_at timestamptz,
  unique(game_id, user_id, status)
);

create index if not exists idx_game_applications_user on game_applications(user_id, created_at desc);
create index if not exists idx_game_applications_game_status on game_applications(game_id, status);
