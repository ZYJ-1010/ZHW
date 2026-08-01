alter table games
  add column if not exists allow_guide_escort boolean not null default false;
