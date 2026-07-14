create table if not exists reviews (
  id bigserial primary key,
  game_id bigint not null,
  reviewer_user_id bigint not null,
  target_user_id bigint not null,
  target_role varchar(32) not null default 'member',
  score int not null,
  content text,
  again_intent varchar(32),
  created_at timestamptz not null default now(),
  constraint uk_review_once unique(game_id, reviewer_user_id, target_user_id, target_role)
);

create table if not exists review_reminders (
  id bigserial primary key,
  game_id bigint not null,
  user_id bigint not null,
  status varchar(32) not null default 'pending',
  deadline_at timestamptz not null,
  created_at timestamptz not null default now(),
  unique(game_id, user_id)
);

alter table review_reminders
  add column if not exists reminder_type varchar(64) not null default 'review_remind',
  add column if not exists notification_id bigint,
  add column if not exists sent_at timestamptz;

create index if not exists idx_review_reminders_user_status on review_reminders(user_id, status, created_at desc);
create index if not exists idx_review_reminders_game_type on review_reminders(game_id, reminder_type);

create table if not exists user_growth_profiles (
  user_id bigint primary key,
  level int not null default 1,
  experience int not null default 0,
  review_count int not null default 0,
  updated_at timestamptz not null default now()
);

create table if not exists experience_logs (
  id bigserial primary key,
  user_id bigint not null,
  game_id bigint,
  change_value int not null,
  reason varchar(128) not null,
  created_at timestamptz not null default now()
);

create table if not exists credit_logs (
  id bigserial primary key,
  user_id bigint not null,
  game_id bigint,
  change_value int not null,
  before_score int not null,
  after_score int not null,
  reason varchar(128) not null,
  created_at timestamptz not null default now()
);

create table if not exists daily_credit_scores (
  id bigserial primary key,
  user_id bigint not null,
  score_date date not null,
  current_score int not null default 100,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint uk_user_credit_date unique(user_id, score_date)
);

create table if not exists credit_deduction_rules (
  id bigserial primary key,
  rule_code varchar(64) not null unique,
  change_value int not null,
  enabled boolean not null default true,
  description text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists points_accounts (
  id bigserial primary key,
  user_id bigint not null unique,
  available_points int not null default 0,
  frozen_points int not null default 0,
  total_earned_points int not null default 0,
  updated_at timestamptz not null default now()
);

create table if not exists points_logs (
  id bigserial primary key,
  user_id bigint not null,
  game_id bigint,
  change_value int not null,
  reason varchar(128) not null,
  created_at timestamptz not null default now()
);

create table if not exists user_footprints (
  id bigserial primary key,
  user_id bigint not null,
  game_id bigint not null,
  action varchar(64) not null,
  created_at timestamptz not null default now()
);

create table if not exists achievements (
  id bigserial primary key,
  code varchar(64) not null unique,
  title varchar(100) not null,
  created_at timestamptz not null default now()
);

create table if not exists user_achievements (
  id bigserial primary key,
  user_id bigint not null,
  achievement_code varchar(64) not null,
  achieved_at timestamptz not null default now(),
  unique(user_id, achievement_code)
);
