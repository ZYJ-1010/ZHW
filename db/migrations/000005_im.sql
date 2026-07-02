create table if not exists chat_rooms (
  id bigserial primary key,
  game_id bigint not null unique,
  status varchar(32) not null default 'active',
  created_at timestamptz not null default now()
);

create table if not exists chat_room_members (
  id bigserial primary key,
  room_id bigint not null,
  user_id bigint not null,
  status varchar(32) not null default 'active',
  created_at timestamptz not null default now(),
  unique(room_id, user_id)
);

create table if not exists chat_messages (
  id bigserial primary key,
  room_id bigint not null,
  sender_user_id bigint not null,
  client_msg_id varchar(128),
  message_type varchar(32) not null,
  content text,
  file_id bigint,
  status varchar(32) not null default 'sent',
  created_at timestamptz not null default now()
);

create index if not exists idx_chat_messages_room_created on chat_messages(room_id, created_at desc);
create unique index if not exists uk_room_client_msg
  on chat_messages(room_id, sender_user_id, client_msg_id)
  where client_msg_id is not null and client_msg_id <> '';

create table if not exists sensitive_words (
  id bigserial primary key,
  word varchar(80) not null unique,
  level varchar(32) not null default 'medium',
  action varchar(32) not null default 'block',
  status varchar(32) not null default 'active',
  created_at timestamptz not null default now()
);

create table if not exists content_risk_logs (
  id bigserial primary key,
  game_id bigint,
  room_id bigint,
  message_id bigint,
  sender_user_id bigint,
  content text,
  word varchar(80),
  action varchar(32) not null,
  status varchar(32) not null,
  source varchar(32) not null default 'sensitive_word',
  created_at timestamptz not null default now()
);

create index if not exists idx_content_risk_logs_game_created
  on content_risk_logs(game_id, created_at desc);
