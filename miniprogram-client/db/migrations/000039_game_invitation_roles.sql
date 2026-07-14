alter table game_invitations
  add column if not exists role varchar(32) not null default 'member';
