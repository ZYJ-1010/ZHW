create table if not exists game_status_logs (
  id bigserial primary key,
  game_id bigint not null,
  from_status varchar(32) not null,
  to_status varchar(32) not null,
  operator_user_id bigint,
  reason text,
  created_at timestamptz not null default now()
);
create index if not exists idx_game_status_logs_game_created on game_status_logs(game_id, created_at desc);
