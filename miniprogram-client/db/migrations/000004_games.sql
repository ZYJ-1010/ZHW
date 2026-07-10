create table if not exists games (
  id bigserial primary key,
  creator_user_id bigint not null,
  title varchar(100) not null,
  game_type varchar(32) not null default 'free',
  game_source varchar(32) not null default 'app',
  status varchar(32) not null default 'draft',
  min_players int not null default 5,
  max_players int not null default 8,
  current_players int not null default 0,
  city_code varchar(32),
  city_name varchar(64),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists game_locations (
  id bigserial primary key,
  game_id bigint not null unique,
  longitude decimal(10,7),
  latitude decimal(10,7),
  address varchar(255),
  poi_title varchar(100),
  city_code varchar(32),
  city_name varchar(64),
  created_at timestamptz not null default now()
);

create index if not exists idx_games_status_city on games(status, city_code);
create index if not exists idx_games_creator_created on games(creator_user_id, created_at desc);

create table if not exists game_service_confirms (
  id bigserial primary key,
  game_id bigint not null unique,
  status varchar(32) not null default 'pending',
  completed_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists game_service_confirm_items (
  id bigserial primary key,
  confirm_id bigint not null,
  game_id bigint not null,
  user_id bigint not null,
  note text,
  file_ids jsonb not null default '[]'::jsonb,
  created_at timestamptz not null default now(),
  unique(game_id, user_id)
);

alter table game_service_confirm_items add column if not exists file_ids jsonb not null default '[]'::jsonb;
create index if not exists idx_game_service_confirm_items_confirm on game_service_confirm_items(confirm_id);
