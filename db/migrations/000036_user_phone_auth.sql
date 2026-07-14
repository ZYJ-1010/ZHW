alter table users
  add column if not exists mobile_hash varchar(80);

create unique index if not exists uk_users_mobile_hash
  on users(mobile_hash)
  where mobile_hash is not null and mobile_hash <> '';
