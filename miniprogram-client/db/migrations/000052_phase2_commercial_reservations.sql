-- 二期商业化仅做架构预留：本期不创建有偿局、不发起支付、不进行资金分配。
alter table games
  add column if not exists distribution_method varchar(32) not null default 'none',
  add column if not exists payment_status varchar(32) not null default 'not_required';

alter table payment_orders
  add column if not exists order_type varchar(32) not null default 'game_join',
  add column if not exists business_type varchar(32) not null default 'free_game',
  add column if not exists provider_user_id bigint,
  add column if not exists distribution_method varchar(32) not null default 'none',
  add column if not exists fulfillment_status varchar(32) not null default 'not_required',
  add column if not exists metadata jsonb not null default '{}'::jsonb;

create table if not exists commercial_service_orders (
  id bigserial primary key,
  order_no varchar(64) not null unique,
  order_type varchar(32) not null,
  game_id bigint,
  buyer_user_id bigint not null,
  provider_user_id bigint not null,
  amount_cent bigint not null default 0,
  payment_status varchar(32) not null default 'created',
  fulfillment_status varchar(32) not null default 'pending',
  distribution_method varchar(32) not null default 'equal_split',
  platform_fee_bps int not null default 0,
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  check (order_type in ('one_to_one', 'custom_service')),
  check (amount_cent >= 0),
  check (platform_fee_bps between 0 and 10000)
);

create table if not exists commercial_order_allocations (
  id bigserial primary key,
  commercial_order_id bigint not null,
  recipient_user_id bigint not null,
  role varchar(32) not null,
  amount_cent bigint not null default 0,
  allocation_status varchar(32) not null default 'pending',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  unique(commercial_order_id, recipient_user_id, role)
);

create index if not exists idx_commercial_service_orders_buyer on commercial_service_orders(buyer_user_id, created_at desc);
create index if not exists idx_commercial_service_orders_provider on commercial_service_orders(provider_user_id, created_at desc);
create index if not exists idx_commercial_order_allocations_order on commercial_order_allocations(commercial_order_id);
