alter table reports
  add column if not exists credit_log_id bigint;

create index if not exists idx_reports_credit_log_id on reports(credit_log_id);
