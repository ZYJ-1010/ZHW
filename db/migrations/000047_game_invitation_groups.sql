alter table game_invitations
  add column if not exists invite_group_id text;

create index if not exists idx_game_invitations_group
  on game_invitations(invite_group_id);
