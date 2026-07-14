alter table identity_verification_records
  add column if not exists real_name_ciphertext text,
  add column if not exists real_name_initials varchar(8);

comment on column identity_verification_records.real_name_ciphertext is
  'AES-GCM encrypted full real name; never expose through public profile APIs';
comment on column identity_verification_records.real_name_initials is
  'Uppercase initials of the first two real-name Han characters for authorized in-game avatars';
