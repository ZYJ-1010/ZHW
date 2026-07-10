alter table invite_relations
  drop constraint if exists invite_relations_invitee_user_id_key;

create index if not exists idx_invite_relations_code
  on invite_relations(invite_code_id);

create index if not exists idx_invite_relations_invitee
  on invite_relations(invitee_user_id);
