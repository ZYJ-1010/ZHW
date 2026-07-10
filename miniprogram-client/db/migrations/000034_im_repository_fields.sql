alter table chat_rooms
  add column if not exists engine varchar(32) not null default 'local',
  add column if not exists openim_group_id varchar(128),
  add column if not exists archived_at timestamptz,
  add column if not exists archive_reason text;
