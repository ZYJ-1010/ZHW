create table if not exists user_login_passwords (
  user_id bigint primary key references users(id) on delete cascade,
  password_hash text not null,
  updated_at timestamptz not null default now()
);
