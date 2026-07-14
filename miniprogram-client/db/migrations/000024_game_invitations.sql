create table if not exists game_invitations (
  id bigserial primary key,
  game_id bigint not null,
  inviter_user_id bigint not null,
  target_user_id bigint not null,
  status varchar(32) not null default 'pending',
  message text,
  application_id bigint,
  created_at timestamptz not null default now(),
  responded_at timestamptz,
  unique(game_id, target_user_id, status)
);

create index if not exists idx_game_invitations_target on game_invitations(target_user_id, created_at desc);
create index if not exists idx_game_invitations_game_status on game_invitations(game_id, status);
