alter table operation_logs
  add column if not exists before_json jsonb,
  add column if not exists after_json jsonb;
