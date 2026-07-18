create table if not exists enterprise_certifications (
  id bigserial primary key,
  user_id bigint not null,
  company_name varchar(200) not null,
  unified_social_credit_code varchar(32) not null,
  legal_person varchar(100) not null,
  business_license_file_id bigint not null,
  public_account_file_id bigint not null,
  status varchar(32) not null default 'pending',
  reject_reason text,
  review_admin_id bigint,
  review_remark text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint chk_enterprise_cert_status check (status in ('pending','approved','rejected'))
);

create unique index if not exists uk_enterprise_cert_pending
  on enterprise_certifications(user_id)
  where status = 'pending';

create index if not exists idx_enterprise_cert_status_created
  on enterprise_certifications(status, created_at desc);
