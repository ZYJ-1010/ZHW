alter table users
  add column if not exists account_type varchar(32) not null default 'player';

update users
set account_type = 'player'
where account_type is null or btrim(account_type) = '';

create unique index if not exists idx_users_platform_official_account
  on users (account_type)
  where account_type = 'platform_official';

-- 平台官方推荐人仅承接后台发放的邀请码：无微信、手机号和实名认证材料，
-- 不会获得任何 C 端登录入口或用户侧邀请能力。
insert into users (status, nickname, realname_status, account_type, created_at, updated_at)
values ('active', '平台官方', 'system', 'platform_official', now(), now())
on conflict (account_type) where account_type = 'platform_official'
do update set
  status = 'active',
  nickname = '平台官方',
  realname_status = 'system',
  updated_at = now();
