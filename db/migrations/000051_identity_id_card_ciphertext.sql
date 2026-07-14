alter table identity_verification_records
  add column if not exists id_card_ciphertext text;

comment on column identity_verification_records.id_card_ciphertext is
  'Encrypted full ID card number for new manual realname submissions; old rows keep masked value only.';
