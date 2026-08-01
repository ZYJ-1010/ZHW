alter table chat_messages
  add column if not exists acked_by bigint[] not null default '{}',
  add column if not exists read_by bigint[] not null default '{}';
