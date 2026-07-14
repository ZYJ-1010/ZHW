-- Local-only accounts for full frontend-to-backend flow testing.
-- Login codes:
--   local-player-1 ... local-player-8
-- Invite codes:
--   LOCAL-PLAYER-1 ... LOCAL-PLAYER-8

insert into users (
  id,
  status,
  nickname,
  avatar_url,
  mobile_masked,
  realname_status,
  created_at,
  updated_at
)
select
  10001 + n,
  'active',
  'Local Player ' || n,
  'https://static.haowan.net.cn/miniprogram/pages/home/player/assets/ranking-avatar-0' || (((n - 1) % 3) + 1) || '.png',
  '139****00' || lpad(n::text, 2, '0'),
  'verified',
  now(),
  now()
from generate_series(1, 8) as n
on conflict (id) do update set
  status = excluded.status,
  nickname = excluded.nickname,
  avatar_url = excluded.avatar_url,
  mobile_masked = excluded.mobile_masked,
  realname_status = excluded.realname_status,
  updated_at = now();

insert into user_profiles (
  user_id,
  gender,
  bio,
  interest_tags,
  city_code,
  city_name,
  created_at,
  updated_at
)
select
  10001 + n,
  'unknown',
  'Local flow player account ' || n,
  '["local-flow","frontend-backend","game-test"]'::jsonb,
  '330100',
  'Hangzhou',
  now(),
  now()
from generate_series(1, 8) as n
on conflict (user_id) do update set
  bio = excluded.bio,
  interest_tags = excluded.interest_tags,
  city_code = excluded.city_code,
  city_name = excluded.city_name,
  updated_at = now();

insert into user_wechat_accounts (
  user_id,
  openid,
  unionid,
  created_at,
  updated_at
)
select
  10001 + n,
  'mock_openid_local-player-' || n,
  'mock_unionid_local-player-' || n,
  now(),
  now()
from generate_series(1, 8) as n
on conflict (openid) do update set
  user_id = excluded.user_id,
  unionid = excluded.unionid,
  updated_at = now();

insert into identity_verification_records (
  user_id,
  status,
  phone_masked,
  sms_verified,
  phone_verified,
  face_verified,
  wechat_realname_consistency,
  updated_at
)
select
  10001 + n,
  'verified',
  '139****00' || lpad(n::text, 2, '0'),
  true,
  true,
  true,
  'not_supported',
  now()
from generate_series(1, 8) as n
on conflict (user_id) do update set
  status = excluded.status,
  phone_masked = excluded.phone_masked,
  sms_verified = excluded.sms_verified,
  phone_verified = excluded.phone_verified,
  face_verified = excluded.face_verified,
  wechat_realname_consistency = excluded.wechat_realname_consistency,
  updated_at = now();

-- User 10003 is the dedicated guide account used by WeChat compile modes.
insert into user_roles (user_id, role_code, status, created_at)
values (10003, 'guide', 'active', now())
on conflict (user_id, role_code) do update set status = 'active';

insert into role_applications (
  user_id, role_code, status, reason, ability_description, proof_file_ids,
  review_admin_id, review_remark, certificate_no, certified_at, created_at, updated_at
)
select
  10003, 'guide', 'approved', '本机测试领路人身份',
  '用于本机测试领路人页面、邀约玩家和组局能力。', '[]'::jsonb,
  1, 'local seed approved', 'ZHW-10003-G-2026', now(), now(), now()
where not exists (
  select 1 from role_applications where user_id = 10003 and role_code = 'guide' and status = 'approved'
);

update role_applications
set certificate_no = 'ZHW-10003-G-2026', certified_at = coalesce(certified_at, updated_at)
where user_id = 10003 and role_code = 'guide' and status = 'approved' and certificate_no is null;

insert into invite_codes (
  code,
  owner_user_id,
  status,
  max_uses,
  used_count,
  entry_type,
  created_at,
  updated_at
)
select
  'LOCAL-PLAYER-' || n,
  10001,
  'active',
  1,
  0,
  'link',
  now(),
  now()
from generate_series(1, 8) as n
on conflict (code) do update set
  owner_user_id = excluded.owner_user_id,
  status = excluded.status,
  max_uses = excluded.max_uses,
  entry_type = excluded.entry_type,
  updated_at = now();

insert into invite_relations (
  invite_code_id,
  inviter_user_id,
  invitee_user_id,
  bind_source,
  created_at
)
select
  c.id,
  10001,
  10001 + n,
  'local_flow_seed',
  now()
from generate_series(1, 8) as n
join invite_codes c on c.code = 'LOCAL-PLAYER-' || n
where not exists (
  select 1 from invite_relations r where r.invitee_user_id = 10001 + n
);

insert into points_accounts (
  user_id,
  available_points,
  frozen_points,
  total_earned_points,
  updated_at
)
select
  10000 + n,
  5000,
  0,
  5000,
  now()
from generate_series(1, 9) as n
on conflict (user_id) do update set
  available_points = greatest(points_accounts.available_points, excluded.available_points),
  frozen_points = 0,
  total_earned_points = greatest(points_accounts.total_earned_points, excluded.total_earned_points),
  updated_at = now();

insert into redemption_items (
  id,
  name,
  description,
  points_cost,
  stock,
  status,
  created_at,
  updated_at
)
values (
  10001,
  'Local Flow Gift',
  'Local-only item for order detail and cancel tests.',
  100,
  99,
  'active',
  now(),
  now()
)
on conflict (id) do update set
  name = excluded.name,
  description = excluded.description,
  points_cost = excluded.points_cost,
  stock = greatest(redemption_items.stock, excluded.stock),
  status = excluded.status,
  updated_at = now();

insert into revenue_templates (
  id,
  name,
  game_type,
  platform_bps,
  creator_bps,
  member_bps,
  status,
  created_at,
  updated_at
)
values (
  10001,
  'Local Free Game Settlement',
  'free',
  0,
  5000,
  5000,
  'active',
  now(),
  now()
)
on conflict (id) do update set
  name = excluded.name,
  game_type = excluded.game_type,
  platform_bps = excluded.platform_bps,
  creator_bps = excluded.creator_bps,
  member_bps = excluded.member_bps,
  status = excluded.status,
  updated_at = now();

select setval(pg_get_serial_sequence('users', 'id'), greatest((select max(id) from users), 1), true);
select setval(pg_get_serial_sequence('invite_codes', 'id'), greatest((select max(id) from invite_codes), 1), true);
select setval(pg_get_serial_sequence('redemption_items', 'id'), greatest((select max(id) from redemption_items), 1), true);
select setval(pg_get_serial_sequence('revenue_templates', 'id'), greatest((select max(id) from revenue_templates), 1), true);
