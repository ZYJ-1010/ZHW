alter table games
  add column if not exists allowed_roles jsonb not null default '["player", "expert", "guide"]'::jsonb;

update games
set allowed_roles = '["player", "expert", "guide"]'::jsonb
where allowed_roles is null or jsonb_typeof(allowed_roles) <> 'array' or jsonb_array_length(allowed_roles) = 0;
