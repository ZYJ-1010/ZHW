create table if not exists user_location_records (
  id bigserial primary key,
  user_id bigint not null,
  longitude decimal(10,7) not null,
  latitude decimal(10,7) not null,
  accuracy_meter decimal(10,2) not null default 0,
  accuracy_warning boolean not null default false,
  city_code varchar(32),
  city_name varchar(64),
  address varchar(255),
  source varchar(32) not null default 'gps',
  created_at timestamptz not null default now()
);

create index if not exists idx_user_location_records_user_created on user_location_records(user_id, created_at desc, id desc);
create index if not exists idx_user_location_records_city_created on user_location_records(city_code, created_at desc);
