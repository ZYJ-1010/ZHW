create table if not exists payment_callbacks (
  id bigserial primary key,
  provider varchar(32) not null default 'wechat_pay',
  callback_type varchar(64) not null default 'payment_notify',
  event_id varchar(128),
  order_no varchar(64),
  out_trade_no varchar(128),
  transaction_id varchar(128),
  raw_headers_json jsonb,
  raw_body text,
  payload_digest varchar(128),
  verify_status varchar(32) not null default 'placeholder_verified',
  process_status varchar(32) not null default 'received',
  error_message text,
  created_at timestamptz not null default now(),
  processed_at timestamptz
);

create unique index if not exists uk_payment_callbacks_order_event
  on payment_callbacks(order_no, event_id)
  where order_no is not null and event_id is not null;

create index if not exists idx_payment_callbacks_trade
  on payment_callbacks(out_trade_no, callback_type);

create index if not exists idx_payment_callbacks_order_created
  on payment_callbacks(order_no, created_at desc);

create index if not exists idx_payment_callbacks_status_created
  on payment_callbacks(process_status, created_at desc);
