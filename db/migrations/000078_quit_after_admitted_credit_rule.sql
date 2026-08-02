insert into credit_deduction_rules (rule_code, change_value, enabled, description, created_at, updated_at)
values
  ('quit_after_admitted', -10, true, '成员入局后、开局前主动退出时扣除信用分。', now(), now())
on conflict (rule_code) do nothing;
