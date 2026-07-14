alter table game_applications
  add column if not exists role varchar(32) not null default 'player';

alter table game_invitations
  add column if not exists role varchar(32) not null default 'guide';
