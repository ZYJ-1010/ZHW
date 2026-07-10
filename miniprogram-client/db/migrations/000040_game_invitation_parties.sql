alter table game_invitations
  add column if not exists player_user_id bigint,
  add column if not exists expert_user_id bigint;

create index if not exists idx_game_invitations_player_user on game_invitations(player_user_id);
create index if not exists idx_game_invitations_expert_user on game_invitations(expert_user_id);

update game_invitations
set player_user_id = target_user_id
where player_user_id is null
  and coalesce(role, 'member') <> 'expert';

update game_invitations
set expert_user_id = target_user_id
where expert_user_id is null
  and coalesce(role, 'member') = 'expert';
