create table if not exists admin_applications (
  id bigserial primary key,
  name varchar(64) not null,
  contact varchar(120) not null,
  desired_role varchar(64) not null references admin_roles(role_code),
  reason varchar(500) not null default '',
  status varchar(32) not null default 'pending',
  review_remark varchar(500),
  reviewed_by bigint references admin_users(id),
  reviewed_at timestamptz,
  linked_admin_user_id bigint references admin_users(id),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint admin_applications_status_check
    check (status in ('pending', 'approved', 'rejected', 'closed')),
  constraint admin_applications_review_fields_check
    check (
      (status = 'pending' and reviewed_by is null and reviewed_at is null and linked_admin_user_id is null)
      or
      (status <> 'pending' and reviewed_by is not null and reviewed_at is not null)
    ),
  constraint admin_applications_link_check
    check (
      (status = 'approved' and linked_admin_user_id is not null)
      or
      (status <> 'approved' and linked_admin_user_id is null)
    ),
  constraint admin_applications_rejection_remark_check
    check (
      status not in ('rejected', 'closed')
      or nullif(btrim(review_remark), '') is not null
    )
);

create index if not exists idx_admin_applications_status_created
  on admin_applications(status, created_at desc);

create index if not exists idx_admin_applications_linked_admin
  on admin_applications(linked_admin_user_id)
  where linked_admin_user_id is not null;
