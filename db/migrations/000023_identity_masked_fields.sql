alter table identity_verification_records
  add column if not exists real_name_masked varchar(128),
  add column if not exists id_card_masked varchar(32);
