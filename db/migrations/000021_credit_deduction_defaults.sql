insert into credit_deduction_rules (rule_code, change_value, enabled, description, created_at, updated_at)
values
  ('quit_after_confirm', -10, true, 'Deduct credit when a member exits after service confirmation starts.', now(), now()),
  ('quit_after_started', -10, true, 'Deduct credit when a member exits after the game has started.', now(), now()),
  ('low_review', -5, true, 'Deduct credit when a participant receives a low score review.', now(), now())
on conflict (rule_code) do update set
  change_value = excluded.change_value,
  enabled = excluded.enabled,
  description = excluded.description,
  updated_at = now();
