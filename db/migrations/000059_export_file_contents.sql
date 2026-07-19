-- 导出内容由服务端生成并持久化，避免只登记 files 元数据而没有实际 CSV。
create table if not exists export_file_contents (
  file_id bigint primary key references files(id) on delete cascade,
  content bytea not null,
  created_at timestamptz not null default now()
);
