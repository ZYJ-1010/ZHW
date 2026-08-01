-- 每条扣分流水只能关联一次信用申诉。历史重复申诉保留原举报记录，
-- 但扣分流水只绑定最早提交的一条，后续审批据此拒绝重复恢复。
with first_credit_appeal as (
  select distinct on (r.reporter_user_id, r.credit_log_id)
    r.id as appeal_id,
    r.reporter_user_id as user_id,
    r.credit_log_id
  from reports r
  where r.report_type = 'credit_appeal'
    and r.credit_log_id is not null
  order by r.reporter_user_id, r.credit_log_id, r.created_at asc, r.id asc
)
update credit_logs cl
set appeal_id = first_credit_appeal.appeal_id
from first_credit_appeal
where cl.id = first_credit_appeal.credit_log_id
  and cl.user_id = first_credit_appeal.user_id
  and cl.change_value < 0
  and cl.appeal_id is null;

create index if not exists idx_credit_logs_appeal_id
  on credit_logs (appeal_id)
  where appeal_id is not null;
