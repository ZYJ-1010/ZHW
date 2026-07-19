insert into admin_roles(role_code, role_name)
values
  ('super_admin', '超级管理员'),
  ('operation_manager', '运营管理员'),
  ('user_manager', '用户管理员'),
  ('finance_manager', '财务管理员'),
  ('customer_manager', '客服管理员'),
  ('data_analyst', '数据分析员'),
  ('game_manager', '组局管理员')
on conflict (role_code) do nothing;

insert into admin_permissions(permission_code, permission_name)
values
  ('admin_user:view', '查看管理员'),
  ('admin_user:create', '创建管理员'),
  ('ai:data:read', '查看 AI 数据准备'),
  ('ai:data:export', '导出 AI 数据'),
  ('ai:data:seed', '生成 AI 验收测试数据'),
  ('analytics:funnel:view', '查看漏斗分析'),
  ('data:behavior:read', '查看行为事件'),
  ('connection:read', '查看人脉关系'),
  ('delivery:manage', '管理交付文档'),
  ('analytics:retention:view', '查看留存分析'),
  ('analytics:timeline:view', '查看行为时间线'),
  ('game:view', '查看组局'),
  ('game:read', '查看组局详情'),
  ('game:progress:manage', '管理组局进度'),
  ('game:create_admin', '后台创建组局'),
  ('game:update_status', '调整组局状态'),
  ('identity:read', '查看实名核验'),
  ('identity:sensitive:read', '查看实名完整敏感字段'),
  ('identity:update', '审核实名核验'),
  ('im:room:view', '查看 IM 房间'),
  ('im:message:view_dispute', '查看争议 IM 消息'),
  ('im:message:hide', '隐藏违规 IM 消息'),
  ('member_report:read', '查看会员报表'),
  ('notification:wechat:view', '查看微信订阅消息任务'),
  ('operation_log:view_self', '查看自己的操作日志'),
  ('operation_log:view_full', '查看完整操作日志'),
  ('points:read', '查看积分流水'),
  ('profile:read', '查看用户画像'),
  ('profile:sensitive:read', '查看用户画像敏感字段'),
  ('redemption:manage', '管理积分兑换'),
  ('report:assign', '分配举报申诉'),
  ('report:handle', '处理举报申诉'),
  ('report:close', '关闭举报申诉'),
  ('report:view', '查看举报申诉'),
  ('report_export:create', '创建报表导出'),
  ('revenue:freeze', '冻结分润记录'),
  ('revenue:generate', '生成分润记录'),
  ('revenue:record:view', '查看分润记录'),
  ('revenue:simulate', '分润试算'),
  ('revenue:template:update', '修改分润模板'),
  ('revenue:template:view', '查看分润模板'),
  ('role:view', '查看角色'),
  ('role:update', '修改角色'),
  ('settlement:offline:create', '登记线下结算'),
  ('system_config:update', '修改系统配置'),
  ('team:read', '查看团队会员'),
  ('testcase:manage', '管理测试用例'),
  ('user:read', '查看用户基础信息'),
  ('user:view', '查看用户'),
  ('user:update_status', '修改用户状态'),
  ('user:export', '导出用户')
on conflict (permission_code) do nothing;

insert into admin_permissions(permission_code, permission_name)
values ('testcase:read', 'View test cases and test runs')
on conflict (permission_code) do nothing;

insert into admin_permissions(permission_code, permission_name)
values ('admin_user:update', 'Update admin users')
on conflict (permission_code) do nothing;

insert into admin_permissions(permission_code, permission_name)
values
  ('feedback:view', 'View user feedback'),
  ('feedback:reply', 'Reply user feedback')
on conflict (permission_code) do nothing;

insert into admin_permissions(permission_code, permission_name)
values
  ('content:sensitive_word:view', 'View sensitive words'),
  ('content:sensitive_word:create', 'Create sensitive words'),
  ('content:sensitive_word:update', 'Update sensitive words'),
  ('content:sensitive_word:import', 'Import sensitive words'),
  ('content:risk_log:view', 'View content risk logs')
on conflict (permission_code) do nothing;

insert into admin_users(username, password_hash, status)
values
  ('admin', 'pbkdf2-sha256$60000$7a68772d6d696e692d61646d696e2d6c6f63616c2d7631$8a063af851459547d9091cd7dbe46291f722d4a0a6b1e2d6e7f75eff38e77cb5', 'active'),
  ('data_analyst', 'pbkdf2-sha256$60000$7a68772d6d696e692d61646d696e2d6c6f63616c2d7631$8a063af851459547d9091cd7dbe46291f722d4a0a6b1e2d6e7f75eff38e77cb5', 'active'),
  ('operator', 'pbkdf2-sha256$60000$7a68772d6d696e692d61646d696e2d6c6f63616c2d7631$8a063af851459547d9091cd7dbe46291f722d4a0a6b1e2d6e7f75eff38e77cb5', 'active'),
  ('user_manager', 'pbkdf2-sha256$60000$7a68772d6d696e692d61646d696e2d6c6f63616c2d7631$8a063af851459547d9091cd7dbe46291f722d4a0a6b1e2d6e7f75eff38e77cb5', 'active'),
  ('finance', 'pbkdf2-sha256$60000$7a68772d6d696e692d61646d696e2d6c6f63616c2d7631$8a063af851459547d9091cd7dbe46291f722d4a0a6b1e2d6e7f75eff38e77cb5', 'active'),
  ('customer', 'pbkdf2-sha256$60000$7a68772d6d696e692d61646d696e2d6c6f63616c2d7631$8a063af851459547d9091cd7dbe46291f722d4a0a6b1e2d6e7f75eff38e77cb5', 'active'),
  ('game_manager', 'pbkdf2-sha256$60000$7a68772d6d696e692d61646d696e2d6c6f63616c2d7631$8a063af851459547d9091cd7dbe46291f722d4a0a6b1e2d6e7f75eff38e77cb5', 'active')
on conflict (username) do nothing;

insert into admin_user_roles(admin_user_id, role_id)
select u.id, r.id
from admin_users u
join admin_roles r on r.role_code = 'super_admin'
where u.username = 'admin'
on conflict do nothing;

insert into admin_user_roles(admin_user_id, role_id)
select u.id, r.id
from admin_users u
join admin_roles r on r.role_code = case u.username
  when 'data_analyst' then 'data_analyst'
  when 'operator' then 'operation_manager'
  when 'user_manager' then 'user_manager'
  when 'finance' then 'finance_manager'
  when 'customer' then 'customer_manager'
  when 'game_manager' then 'game_manager'
end
where u.username in ('data_analyst', 'operator', 'user_manager', 'finance', 'customer', 'game_manager')
on conflict do nothing;

insert into admin_role_permissions(role_id, permission_id)
select r.id, p.id
from admin_roles r
cross join admin_permissions p
where r.role_code = 'super_admin'
on conflict do nothing;

insert into admin_role_permissions(role_id, permission_id)
select r.id, p.id
from admin_roles r
join admin_permissions p on p.permission_code in (
  'ai:data:read',
  'analytics:funnel:view',
  'analytics:retention:view',
  'analytics:timeline:view',
  'data:behavior:read',
  'member_report:read',
  'operation_log:view_self',
  'report:view',
  'report_export:create',
  'testcase:read'
)
where r.role_code = 'data_analyst'
on conflict do nothing;

insert into admin_role_permissions(role_id, permission_id)
select r.id, p.id
from admin_roles r
join admin_permissions p on p.permission_code in (
  'game:progress:manage',
  'game:read',
  'game:view',
  'notification:wechat:view',
  'operation_log:view_self',
  'report:assign',
  'report:handle',
  'report:close',
  'report:view',
  'user:read',
  'user:view'
)
where r.role_code = 'operation_manager'
on conflict do nothing;

insert into admin_permissions(permission_code, permission_name)
values
  ('im:room:read', 'Read IM rooms'),
  ('im:room:archive', 'Archive IM rooms'),
  ('im:room:retry_create', 'Retry IM room creation'),
  ('system_config:read', 'Read system config'),
  ('invite_code:read', 'Read invite codes'),
  ('invite_code:manage', 'Manage invite codes')
on conflict (permission_code) do nothing;

insert into admin_role_permissions(role_id, permission_id)
select r.id, p.id
from admin_roles r
join admin_permissions p on p.permission_code in (
  'connection:read',
  'identity:read',
  'identity:sensitive:read',
  'identity:update',
  'invite_code:manage',
  'invite_code:read',
  'member_report:read',
  'operation_log:view_self',
  'points:read',
  'profile:read',
  'profile:sensitive:read',
  'redemption:manage',
  'role:update',
  'role:view',
  'team:read',
  'user:read',
  'user:view'
)
where r.role_code = 'user_manager'
on conflict do nothing;

insert into admin_role_permissions(role_id, permission_id)
select r.id, p.id
from admin_roles r
join admin_permissions p on p.permission_code in (
  'operation_log:view_self',
  'report_export:create',
  'revenue:freeze',
  'revenue:generate',
  'revenue:record:view',
  'revenue:simulate',
  'revenue:template:update',
  'revenue:template:view',
  'settlement:offline:create'
)
where r.role_code = 'finance_manager'
on conflict do nothing;

insert into admin_role_permissions(role_id, permission_id)
select r.id, p.id
from admin_roles r
join admin_permissions p on p.permission_code in (
  'im:message:view_dispute',
  'im:room:read',
  'operation_log:view_self',
  'report:assign',
  'report:close',
  'report:handle',
  'report:view',
  'user:read',
  'user:view'
)
where r.role_code = 'customer_manager'
on conflict do nothing;

insert into admin_role_permissions(role_id, permission_id)
select r.id, p.id
from admin_roles r
join admin_permissions p on p.permission_code in (
  'feedback:reply',
  'feedback:view'
)
where r.role_code in ('customer_manager', 'operation_manager')
on conflict do nothing;

insert into admin_role_permissions(role_id, permission_id)
select r.id, p.id
from admin_roles r
join admin_permissions p on p.permission_code in (
  'game:create_admin',
  'game:progress:manage',
  'game:read',
  'game:update_status',
  'game:view',
  'operation_log:view_self',
  'im:room:read'
)
where r.role_code = 'game_manager'
on conflict do nothing;

insert into admin_role_permissions(role_id, permission_id)
select r.id, p.id
from admin_roles r
join admin_permissions p on p.permission_code in (
  'content:risk_log:view',
  'content:sensitive_word:update',
  'operation_log:view_self',
  'content:sensitive_word:view'
)
where r.role_code = 'operation_manager'
on conflict do nothing;

insert into admin_role_permissions(role_id, permission_id)
select r.id, p.id
from admin_roles r
join admin_permissions p on p.permission_code in (
  'content:risk_log:view',
  'content:sensitive_word:update',
  'operation_log:view_self'
)
where r.role_code = 'data_analyst'
on conflict do nothing;

insert into admin_role_permissions(role_id, permission_id)
select r.id, p.id
from admin_roles r
join admin_permissions p on p.permission_code in (
  'im:room:read',
  'im:room:archive',
  'im:room:retry_create',
  'system_config:read',
  'invite_code:read',
  'invite_code:manage'
)
where r.role_code = 'super_admin'
on conflict do nothing;
