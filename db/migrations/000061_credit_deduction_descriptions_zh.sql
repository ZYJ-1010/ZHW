update credit_deduction_rules
set description = case rule_code
  when 'quit_after_confirm' then '成员确认服务后主动退出时扣除信用分。'
  when 'quit_after_started' then '局已开局后主动退出时扣除信用分。'
  when 'player_cancel_service' then '玩家取消已确认服务时扣除信用分。'
  when 'expert_cancel_service' then '行家取消已确认服务时扣除信用分。'
  when 'low_review' then '收到低分评价时扣除信用分。'
  when 'report_confirmed' then '举报核实成立时扣除信用分。'
  when 'malicious_report' then '恶意举报核实成立时扣除信用分。'
  else description
end,
updated_at = now()
where rule_code in (
  'quit_after_confirm',
  'quit_after_started',
  'player_cancel_service',
  'expert_cancel_service',
  'low_review',
  'report_confirmed',
  'malicious_report'
);
