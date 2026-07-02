create table if not exists game_favorites (
  id bigserial primary key,
  user_id bigint not null,
  game_id bigint not null,
  created_at timestamptz not null default now(),
  unique(user_id, game_id)
);

create table if not exists game_milestones (
  id bigserial primary key,
  game_id bigint not null,
  title varchar(100) not null,
  status varchar(32) not null default 'pending',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists game_checkins (
  id bigserial primary key,
  game_id bigint not null,
  user_id bigint not null,
  milestone_id bigint,
  checkin_type varchar(32) not null,
  content text,
  file_ids jsonb not null default '[]'::jsonb,
  status varchar(32) not null default 'valid',
  created_at timestamptz not null default now()
);

create table if not exists game_retrospectives (
  id bigserial primary key,
  game_id bigint not null,
  user_id bigint not null,
  content text not null,
  again_intent varchar(32),
  created_at timestamptz not null default now()
);

create table if not exists game_continue_drafts (
  id bigserial primary key,
  original_game_id bigint not null,
  draft_game_id bigint not null,
  creator_user_id bigint not null,
  title varchar(100) not null,
  status varchar(32) not null default 'draft',
  created_at timestamptz not null default now(),
  unique(original_game_id, draft_game_id)
);

create table if not exists team_relations (
  id bigserial primary key,
  leader_user_id bigint not null,
  member_user_id bigint not null,
  relation_level int not null default 1,
  source varchar(32) not null,
  status varchar(32) not null default 'active',
  created_at timestamptz not null default now(),
  unique(leader_user_id, member_user_id, relation_level)
);

create table if not exists member_report_snapshots (
  id bigserial primary key,
  user_id bigint not null,
  report_type varchar(32) not null,
  period varchar(32) not null,
  metrics jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now()
);

create index if not exists idx_member_report_snapshots_user_period on member_report_snapshots(user_id, report_type, period);

create table if not exists redemption_items (
  id bigserial primary key,
  name varchar(100) not null,
  description text,
  points_cost int not null,
  stock int not null default 0,
  status varchar(32) not null default 'active',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists redemption_orders (
  id bigserial primary key,
  order_no varchar(64) not null unique,
  user_id bigint not null,
  item_id bigint not null,
  points_cost int not null,
  status varchar(32) not null default 'pending',
  review_admin_id bigint,
  review_reason text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists idx_redemption_orders_user_created on redemption_orders(user_id, created_at desc);

create table if not exists user_connections (
  id bigserial primary key,
  user_id bigint not null,
  connected_user_id bigint not null,
  relation_type varchar(32) not null,
  source_type varchar(32) not null,
  source_id bigint,
  strength_score int not null default 0,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  unique(user_id, connected_user_id, relation_type, source_type, source_id)
);

create index if not exists idx_user_connections_user on user_connections(user_id, relation_type, updated_at desc);

create table if not exists connection_follow_logs (
  id bigserial primary key,
  connection_id bigint not null,
  operator_user_id bigint not null,
  follow_type varchar(32) not null,
  content text not null,
  next_follow_at timestamptz,
  created_at timestamptz not null default now()
);

create index if not exists idx_connection_follow_logs_connection on connection_follow_logs(connection_id, created_at desc);

create table if not exists expert_skill_profiles (
  id bigserial primary key,
  user_id bigint not null unique,
  skill_tree jsonb not null default '[]'::jsonb,
  service_tags jsonb not null default '[]'::jsonb,
  case_file_ids jsonb not null default '[]'::jsonb,
  updated_at timestamptz not null default now()
);

create table if not exists guide_resource_profiles (
  id bigserial primary key,
  user_id bigint not null unique,
  resource_tags jsonb not null default '[]'::jsonb,
  industry_tags jsonb not null default '[]'::jsonb,
  city_codes jsonb not null default '[]'::jsonb,
  connection_scale varchar(64),
  updated_at timestamptz not null default now()
);

create table if not exists delivery_documents (
  id bigserial primary key,
  doc_type varchar(64) not null,
  title varchar(200) not null,
  status varchar(32) not null default 'draft',
  created_at timestamptz not null default now()
);

create table if not exists test_cases (
  id bigserial primary key,
  module varchar(64) not null,
  case_name varchar(200) not null,
  priority varchar(16) not null,
  expected_result text not null,
  created_at timestamptz not null default now()
);

create table if not exists test_runs (
  id bigserial primary key,
  case_id bigint not null,
  result varchar(32) not null,
  actual_result text,
  request_id varchar(128),
  created_at timestamptz not null default now()
);
