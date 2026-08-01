-- 信用账户为永久账户，不按自然日重置。保留 daily_credit_scores 仅供历史数据迁移和旧版本兼容。
create table if not exists credit_accounts (
  user_id bigint primary key,
  current_score int not null default 100 check (current_score >= 0),
  updated_at timestamptz not null default now()
);

-- 将每位用户最近一条历史日信用记录迁入永久账户；已存在的永久账户不被覆盖。
insert into credit_accounts (user_id, current_score, updated_at)
select distinct on (user_id) user_id, current_score, updated_at
from daily_credit_scores
order by user_id, score_date desc, updated_at desc
on conflict (user_id) do nothing;

alter table credit_logs
  add column if not exists rule_code varchar(64),
  add column if not exists source_type varchar(64),
  add column if not exists source_id varchar(64),
  add column if not exists role_snapshot varchar(32),
  add column if not exists operator_type varchar(32),
  add column if not exists operator_id bigint,
  add column if not exists rule_version varchar(64),
  add column if not exists appeal_id bigint,
  add column if not exists idempotency_key varchar(128);

create unique index if not exists uk_credit_logs_idempotency
  on credit_logs (user_id, idempotency_key)
  where idempotency_key is not null;

create index if not exists idx_credit_accounts_score on credit_accounts (current_score);
