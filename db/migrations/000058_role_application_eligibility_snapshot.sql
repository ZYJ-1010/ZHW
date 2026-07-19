alter table role_applications
  add column if not exists eligibility_snapshot jsonb not null default '{}'::jsonb;

comment on column role_applications.eligibility_snapshot is '提交申请时的实名、企业认证、次数、信用分和计划书资格快照';
