create table if not exists app_auth_sessions (
  id bigserial primary key,
  token_hash varchar(128) not null unique,
  user_id bigint not null,
  session_kind varchar(32) not null,
  expires_at timestamptz not null,
  revoked_at timestamptz,
  created_at timestamptz not null default now()
);

create index if not exists idx_app_auth_sessions_user_kind on app_auth_sessions(user_id, session_kind, expires_at desc);

create table if not exists identity_verification_records (
  id bigserial primary key,
  user_id bigint not null unique,
  status varchar(64) not null,
  phone_masked varchar(32),
  phone_encrypted text,
  phone_hash varchar(128),
  sms_verified boolean not null default false,
  phone_verified boolean not null default false,
  face_verified boolean not null default false,
  wechat_realname_consistency varchar(64) not null default 'pending',
  failure_reason text,
  last_sms_sent_at timestamptz,
  sms_daily_key varchar(16),
  sms_daily_count int not null default 0,
  last_face_token_hash varchar(128),
  faceid_request_id varchar(128),
  faceid_result_digest varchar(128),
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists idx_identity_verification_records_status on identity_verification_records(status, updated_at desc);

create table if not exists identity_sms_code_records (
  id bigserial primary key,
  user_id bigint not null,
  scene varchar(64) not null default 'strong_identity',
  phone_masked varchar(32),
  code_hash varchar(128) not null,
  sent_at timestamptz not null default now(),
  expires_at timestamptz not null,
  verified_at timestamptz,
  verify_failed_count int not null default 0,
  created_at timestamptz not null default now()
);

create index if not exists idx_identity_sms_code_records_user_scene on identity_sms_code_records(user_id, scene, sent_at desc);

create table if not exists identity_faceid_sessions (
  id bigserial primary key,
  user_id bigint not null,
  face_token_hash varchar(128) not null unique,
  status varchar(64) not null default 'processing',
  detect_auth_payload_digest varchar(128),
  callback_payload_digest varchar(128),
  failure_reason text,
  created_at timestamptz not null default now(),
  completed_at timestamptz
);

create index if not exists idx_identity_faceid_sessions_user_status on identity_faceid_sessions(user_id, status, created_at desc);
