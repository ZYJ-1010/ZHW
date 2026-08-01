alter table experience_logs
  add column if not exists idempotency_key varchar(160);

create unique index if not exists uk_experience_logs_idempotency
  on experience_logs (idempotency_key)
  where idempotency_key is not null;
