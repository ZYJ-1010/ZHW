create table if not exists files (
  id bigserial primary key,
  uploader_user_id bigint,
  biz_type varchar(64) not null,
  object_id bigint,
  file_name varchar(255) not null,
  mime_type varchar(128),
  size_bytes bigint not null default 0,
  sha256 varchar(128),
  storage_key varchar(512) not null,
  access_level varchar(32) not null default 'private',
  expires_at timestamptz,
  created_at timestamptz not null default now()
);

create unique index if not exists uk_files_storage_key on files(storage_key);
create index if not exists idx_files_biz_object on files(biz_type, object_id);
create index if not exists idx_files_uploader_created on files(uploader_user_id, created_at desc);
