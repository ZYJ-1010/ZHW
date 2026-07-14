alter table role_applications
  add column if not exists certificate_no varchar(64),
  add column if not exists certified_at timestamptz;

update role_applications
set certificate_no = 'ZHW-' || lpad(id::text, 5, '0') || '-' || extract(year from updated_at)::int,
    certified_at = updated_at
where status = 'approved'
  and (certificate_no is null or certificate_no = '');

create unique index if not exists uk_role_applications_certificate_no
  on role_applications(certificate_no)
  where certificate_no is not null;
