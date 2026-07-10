create table if not exists payment_orders (
  id bigserial primary key,
  order_no varchar(64) not null unique,
  user_id bigint not null,
  game_id bigint,
  amount_cent bigint not null default 0,
  pay_status varchar(32) not null default 'free_no_pay',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists revenue_templates (
  id bigserial primary key,
  name varchar(100) not null,
  game_type varchar(32) not null default 'free',
  platform_bps int not null default 0,
  creator_bps int not null default 0,
  member_bps int not null default 10000,
  status varchar(32) not null default 'active',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint ck_revenue_template_bps check (platform_bps + creator_bps + member_bps = 10000)
);

create table if not exists revenue_rules (
  id bigserial primary key,
  template_id bigint not null,
  rule_code varchar(64) not null,
  rule_value varchar(128) not null,
  created_at timestamptz not null default now(),
  unique(template_id, rule_code)
);

create table if not exists revenue_records (
  id bigserial primary key,
  revenue_record_no varchar(64) not null,
  game_id bigint not null,
  template_id bigint not null,
  status varchar(32) not null default 'pending_settlement',
  total_amount_cent bigint not null default 0,
  frozen_reason text,
  settled_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint uk_revenue_record_no unique(revenue_record_no),
  unique(game_id)
);

create table if not exists revenue_record_items (
  id bigserial primary key,
  revenue_record_id bigint not null,
  user_id bigint,
  role varchar(32) not null,
  amount_cent bigint not null default 0,
  created_at timestamptz not null default now()
);

create table if not exists user_income_accounts (
  id bigserial primary key,
  user_id bigint not null unique,
  total_cent bigint not null default 0,
  pending_cent bigint not null default 0,
  settled_cent bigint not null default 0,
  updated_at timestamptz not null default now()
);

create table if not exists income_logs (
  id bigserial primary key,
  user_id bigint not null,
  revenue_record_id bigint,
  change_value_cent bigint not null,
  reason varchar(128) not null,
  created_at timestamptz not null default now()
);

create table if not exists settlement_records (
  id bigserial primary key,
  revenue_record_id bigint not null,
  method varchar(32) not null default 'offline',
  proof_no varchar(128),
  amount_cent bigint not null default 0,
  created_at timestamptz not null default now()
);
