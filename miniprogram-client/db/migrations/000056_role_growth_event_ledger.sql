-- 三角色等级指标使用可追溯的事实事件，不直接持久化不可解释的累计分数。
create table if not exists growth_event_ledger (
  id bigserial primary key,
  user_id bigint not null,
  game_id bigint,
  role_code varchar(32) not null,
  event_code varchar(64) not null,
  idempotency_key varchar(128) not null unique,
  occurred_at timestamptz not null,
  created_at timestamptz not null default now()
);

create index if not exists idx_growth_event_ledger_user_role_time
  on growth_event_ledger (user_id, role_code, occurred_at desc);
create index if not exists idx_growth_event_ledger_game
  on growth_event_ledger (game_id, event_code);
