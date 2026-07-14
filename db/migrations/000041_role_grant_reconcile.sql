insert into user_roles (user_id, role_code, status, created_at)
select distinct user_id, role_code, 'active', now()
from role_applications
where status = 'approved'
on conflict (user_id, role_code) do update set status = 'active';

insert into guide_qualification_records (user_id, condition_met, payment_met, guide_open_status, updated_at)
select distinct user_id, true, true, 'opened', now()
from role_applications
where status = 'approved' and role_code = 'guide'
on conflict (user_id) do update set
  condition_met = true,
  payment_met = true,
  guide_open_status = 'opened',
  updated_at = now();
