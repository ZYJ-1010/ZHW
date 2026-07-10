create table if not exists realname_auth_records (
  id bigserial primary key,
  user_id bigint not null,
  real_name_encrypted text not null,
  id_card_encrypted text not null,
  mobile_encrypted text,
  status varchar(32) not null default 'pending',
  reject_reason text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists user_roles (
  id bigserial primary key,
  user_id bigint not null,
  role_code varchar(32) not null,
  status varchar(32) not null default 'active',
  created_at timestamptz not null default now(),
  unique(user_id, role_code)
);

create table if not exists role_applications (
  id bigserial primary key,
  user_id bigint not null,
  role_code varchar(32) not null,
  status varchar(32) not null default 'pending',
  reason text,
  ability_description text,
  proof_file_ids jsonb not null default '[]'::jsonb,
  reject_reason text,
  review_admin_id bigint,
  review_remark text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint chk_role_applications_role_code check (role_code in ('expert','guide'))
);

create unique index if not exists uk_role_applications_pending
  on role_applications(user_id, role_code)
  where status = 'pending';

create table if not exists membership_plans (
  id bigserial primary key,
  plan_code varchar(64) not null unique,
  plan_name varchar(100) not null,
  status varchar(32) not null default 'active',
  created_at timestamptz not null default now()
);

create table if not exists user_memberships (
  id bigserial primary key,
  user_id bigint not null,
  plan_code varchar(64) not null,
  status varchar(32) not null default 'active',
  started_at timestamptz not null default now(),
  expires_at timestamptz
);

create table if not exists guide_qualification_rules (
  id bigserial primary key,
  rule_code varchar(64) not null unique,
  min_invite_count int not null default 0,
  min_credit_score int not null default 0,
  min_completed_games int not null default 0,
  payment_required boolean not null default true,
  status varchar(32) not null default 'active',
  updated_at timestamptz not null default now()
);

create table if not exists guide_qualification_records (
  user_id bigint primary key,
  condition_met boolean not null default false,
  payment_met boolean not null default false,
  guide_open_status varchar(32) not null default 'waiting_condition',
  updated_at timestamptz not null default now()
);
