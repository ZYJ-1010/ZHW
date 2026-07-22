create table if not exists invite_quota_requests (
  id bigserial primary key,
  owner_user_id bigint not null,
  quantity integer not null check (quantity > 0 and quantity <= 1000),
  reason text not null default '',
  status varchar(32) not null default 'pending',
  audit_reason text not null default '',
  reviewed_by bigint,
  created_at timestamptz not null default now(),
  reviewed_at timestamptz
);

alter table invite_quota_requests
  drop constraint if exists invite_quota_requests_quantity_check;
alter table invite_quota_requests
  add constraint invite_quota_requests_quantity_check check (quantity > 0 and quantity <= 1000);

create index if not exists idx_invite_quota_requests_owner_status
  on invite_quota_requests(owner_user_id, status, id desc);

create index if not exists idx_invite_quota_requests_status
  on invite_quota_requests(status, id desc);
