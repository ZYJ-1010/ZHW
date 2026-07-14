alter table game_applications
  drop constraint if exists game_applications_game_id_user_id_status_key;

create unique index if not exists uk_game_applications_pending
  on game_applications(game_id, user_id)
  where status = 'pending';
