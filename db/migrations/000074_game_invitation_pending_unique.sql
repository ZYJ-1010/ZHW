alter table game_invitations
  drop constraint if exists game_invitations_game_id_target_user_id_status_key;

create unique index if not exists uk_game_invitations_pending
  on game_invitations(game_id, target_user_id)
  where status = 'pending';
