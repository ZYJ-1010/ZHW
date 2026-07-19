alter table if exists role_applications
  add column if not exists eligibility_snapshot jsonb not null default '{}'::jsonb;
