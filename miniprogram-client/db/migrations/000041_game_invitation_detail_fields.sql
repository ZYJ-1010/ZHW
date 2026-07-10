alter table game_invitations
  add column if not exists service_type varchar(120),
  add column if not exists service_duration_text varchar(64),
  add column if not exists demand_detail text,
  add column if not exists budget_amount_cent bigint not null default 0,
  add column if not exists expected_time_text varchar(64);
