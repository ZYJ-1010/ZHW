alter table points_logs
  add column if not exists before_points int,
  add column if not exists after_points int,
  add column if not exists biz_type varchar(64),
  add column if not exists biz_id bigint;

create index if not exists idx_points_logs_user_created on points_logs(user_id, created_at desc, id desc);
create index if not exists idx_points_logs_biz on points_logs(biz_type, biz_id);

create index if not exists idx_redemption_items_status_created on redemption_items(status, created_at desc);
create index if not exists idx_redemption_orders_status_created on redemption_orders(status, created_at desc);
