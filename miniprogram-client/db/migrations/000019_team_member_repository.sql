create table if not exists teams (
  id bigserial primary key,
  leader_user_id bigint not null unique,
  name varchar(100) not null,
  status varchar(32) not null default 'active',
  created_at timestamptz not null default now()
);

create index if not exists idx_teams_status_created on teams(status, created_at desc);

create table if not exists member_report_memberships (
  user_id bigint primary key,
  plan_name varchar(100) not null,
  invited_count int not null default 0,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists idx_member_report_memberships_updated on member_report_memberships(updated_at desc);
