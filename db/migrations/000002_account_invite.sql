create table if not exists users (
  id bigserial primary key,
  status varchar(32) not null default 'active',
  nickname varchar(100),
  avatar_url text,
  mobile_encrypted text,
  mobile_masked varchar(32),
  realname_status varchar(32) not null default 'pending',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists user_wechat_accounts (
  id bigserial primary key,
  user_id bigint not null,
  openid varchar(128) not null unique,
  unionid varchar(128),
  session_key_encrypted text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists invite_codes (
  id bigserial primary key,
  code varchar(64) not null unique,
  owner_user_id bigint,
  status varchar(32) not null default 'active',
  max_uses int,
  used_count int not null default 0,
  expires_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists invite_relations (
  id bigserial primary key,
  invite_code_id bigint not null,
  inviter_user_id bigint,
  invitee_user_id bigint not null unique,
  bind_source varchar(32) not null,
  created_at timestamptz not null default now()
);

create index if not exists idx_invite_relations_inviter on invite_relations(inviter_user_id);

