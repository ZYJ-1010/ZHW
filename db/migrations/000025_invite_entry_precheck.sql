alter table invite_codes
  add column if not exists entry_type varchar(32) not null default 'link';

create index if not exists idx_invite_codes_entry_type on invite_codes(entry_type);
