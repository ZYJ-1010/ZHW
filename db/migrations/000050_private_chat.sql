create table if not exists private_chat_conversations (
  id bigserial primary key,
  user_a_id bigint not null,
  user_b_id bigint not null,
  source_game_id bigint,
  status varchar(32) not null default 'active',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint ck_private_chat_distinct_users check (user_a_id <> user_b_id),
  unique(user_a_id, user_b_id)
);

create table if not exists private_chat_messages (
  id bigserial primary key,
  conversation_id bigint not null,
  sender_user_id bigint not null,
  target_user_id bigint not null,
  source_game_id bigint,
  message_type varchar(32) not null default 'text',
  content text not null default '',
  status varchar(32) not null default 'sent',
  created_at timestamptz not null default now()
);

create index if not exists idx_private_chat_messages_conversation_created
  on private_chat_messages(conversation_id, created_at asc, id asc);

create index if not exists idx_private_chat_messages_sender_created
  on private_chat_messages(sender_user_id, created_at desc);
