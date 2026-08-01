create table if not exists auth_phone_code_states (
  phone_key varchar(128) primary key,
  phone_masked varchar(32),
  scene varchar(32),
  code_hash varchar(128),
  sent_at timestamptz,
  expires_at timestamptz,
  failed_attempts integer not null default 0 check (failed_attempts >= 0),
  consumed_at timestamptz,
  claim_token varchar(128),
  day_key varchar(8),
  day_count integer not null default 0 check (day_count >= 0),
  pending_token varchar(128),
  pending_scene varchar(32),
  pending_code_hash varchar(128),
  pending_expires_at timestamptz,
  pending_started_at timestamptz,
  updated_at timestamptz not null default now()
);

create index if not exists idx_auth_phone_code_states_pending
  on auth_phone_code_states (pending_started_at)
  where pending_token is not null;

comment on table auth_phone_code_states is '手机号登录、邀请注册及密码重置验证码共享状态；仅保存手机号和验证码的不可逆哈希';
