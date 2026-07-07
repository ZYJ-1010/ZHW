create table if not exists user_phone_accounts (
  id bigserial primary key,
  user_id bigint not null unique,
  phone_hash varchar(128) not null unique,
  phone_masked varchar(32),
  password_plain varchar(128),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists idx_user_phone_accounts_user on user_phone_accounts(user_id);
