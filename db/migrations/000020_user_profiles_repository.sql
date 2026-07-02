create table if not exists user_profiles (
  id bigserial primary key,
  user_id bigint not null unique,
  gender varchar(32),
  bio text,
  interest_tags jsonb not null default '[]'::jsonb,
  city_code varchar(32),
  city_name varchar(64),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists idx_user_profiles_city on user_profiles(city_code, updated_at desc);

alter table users add column if not exists avatar_file_id bigint;
create index if not exists idx_users_avatar_file on users(avatar_file_id);
