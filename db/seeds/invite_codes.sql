insert into invite_codes(code, owner_user_id, status, max_uses, used_count, entry_type, created_at, updated_at)
values
  ('ENJOY2026', null, 'active', 1, 0, 'link', now(), now())
on conflict (code) do update set
  status = excluded.status,
  max_uses = excluded.max_uses,
  entry_type = excluded.entry_type,
  updated_at = now();
