insert into credit_deduction_rules (rule_code, change_value, enabled, description, created_at, updated_at)
values
  ('player_cancel_service', -3, true, 'Deduct credit when a player cancels an active service.', now(), now()),
  ('expert_cancel_service', -5, true, 'Deduct credit when an expert cancels an active service.', now(), now())
on conflict (rule_code) do update set
  change_value = excluded.change_value,
  enabled = excluded.enabled,
  description = excluded.description,
  updated_at = now();
