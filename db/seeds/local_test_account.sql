-- Local-only seed for WeChat DevTools testing.
-- Login code: local-expert-guide
-- Mock openid resolved by the Go API: mock_openid_local-expert-guide

with upsert_user as (
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
  values (
    10001,
    'active',
    '本机测试行家领路人',
    'https://static.haowan.net.cn/miniprogram/pages/home/player/assets/ranking-avatar-03.png',
    '138****0001',
    'verified',
    now(),
    now()
  )
  on conflict (id) do update set
    status = excluded.status,
    nickname = excluded.nickname,
    avatar_url = excluded.avatar_url,
    mobile_masked = excluded.mobile_masked,
    realname_status = excluded.realname_status,
    updated_at = now()
  returning id
)
insert into user_wechat_accounts (user_id, openid, unionid, created_at, updated_at)
select id, 'mock_openid_local-expert-guide', 'mock_unionid_local-expert-guide', now(), now()
from upsert_user
on conflict (openid) do update set
  user_id = excluded.user_id,
  unionid = excluded.unionid,
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
values (
  10001,
  'unknown',
  '本机联调用测试账号，已拥有行家和领路人身份。',
  '["创业交流","AI工具","城市探索"]'::jsonb,
  '110100',
  '北京',
  now(),
  now()
)
on conflict (user_id) do update set
  bio = excluded.bio,
  interest_tags = excluded.interest_tags,
  city_code = excluded.city_code,
  city_name = excluded.city_name,
  updated_at = now();

insert into identity_verification_records (
  user_id,
  status,
  phone_masked,
  real_name_masked,
  id_card_masked,
  sms_verified,
  phone_verified,
  face_verified,
  wechat_realname_consistency,
  updated_at
)
values (
  10001,
  'verified',
  '138****0001',
  '测*',
  '110***********1234',
  true,
  true,
  true,
  'not_supported',
  now()
)
on conflict (user_id) do update set
  status = excluded.status,
  phone_masked = excluded.phone_masked,
  real_name_masked = excluded.real_name_masked,
  id_card_masked = excluded.id_card_masked,
  sms_verified = excluded.sms_verified,
  phone_verified = excluded.phone_verified,
  face_verified = excluded.face_verified,
  wechat_realname_consistency = excluded.wechat_realname_consistency,
  updated_at = now();

insert into user_roles (user_id, role_code, status, created_at)
values
  (10001, 'expert', 'active', now()),
  (10001, 'guide', 'active', now())
on conflict (user_id, role_code) do update set status = 'active';

insert into role_applications (
  user_id,
  role_code,
  status,
  reason,
  ability_description,
  proof_file_ids,
  review_admin_id,
  review_remark,
  certificate_no,
  certified_at,
  created_at,
  updated_at
)
values
  (10001, 'expert', 'approved', '本机测试行家身份', '用于本机测试行家页面、行家服务和组局能力。', '[]'::jsonb, 1, 'local seed approved', 'ZHW-10001-E-2026', now(), now(), now()),
  (10001, 'guide', 'approved', '本机测试领路人身份', '用于本机测试领路人页面、主行家/主领路人相关能力。', '[]'::jsonb, 1, 'local seed approved', 'ZHW-10001-G-2026', now(), now(), now())
on conflict do nothing;

update role_applications
set
  certificate_no = case role_code
    when 'expert' then 'ZHW-10001-E-2026'
    when 'guide' then 'ZHW-10001-G-2026'
  end,
  certified_at = coalesce(certified_at, updated_at)
where user_id = 10001
  and role_code in ('expert', 'guide')
  and status = 'approved'
  and certificate_no is null;

insert into guide_qualification_records (
  user_id,
  condition_met,
  payment_met,
  guide_open_status,
  updated_at
)
values (10001, true, true, 'opened', now())
on conflict (user_id) do update set
  condition_met = excluded.condition_met,
  payment_met = excluded.payment_met,
  guide_open_status = excluded.guide_open_status,
  updated_at = now();

insert into expert_skill_profiles (
  user_id,
  skill_tree,
  service_tags,
  case_file_ids,
  updated_at
)
values (
  10001,
  '["AI工具落地","创业复盘","活动策划"]'::jsonb,
  '["AI交流","项目诊断","资源链接"]'::jsonb,
  '[]'::jsonb,
  now()
)
on conflict (user_id) do update set
  skill_tree = excluded.skill_tree,
  service_tags = excluded.service_tags,
  case_file_ids = excluded.case_file_ids,
  updated_at = now();

insert into guide_resource_profiles (
  user_id,
  resource_tags,
  industry_tags,
  city_codes,
  connection_scale,
  updated_at
)
values (
  10001,
  '["本地场地","创业者社群","AI工具圈"]'::jsonb,
  '["互联网","文娱活动","企业服务"]'::jsonb,
  '["110100","310100"]'::jsonb,
  '100-500人',
  now()
)
on conflict (user_id) do update set
  resource_tags = excluded.resource_tags,
  industry_tags = excluded.industry_tags,
  city_codes = excluded.city_codes,
  connection_scale = excluded.connection_scale,
  updated_at = now();

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
values (
  'LOCAL-EXPERT-GUIDE',
  10001,
  'active',
  1,
  0,
  'link',
  now(),
  now()
)
on conflict (code) do update set
  owner_user_id = excluded.owner_user_id,
  status = excluded.status,
  max_uses = excluded.max_uses,
  entry_type = excluded.entry_type,
  updated_at = now();

with invite as (
  select id from invite_codes where code = 'LOCAL-EXPERT-GUIDE'
)
insert into invite_relations (
  invite_code_id,
  inviter_user_id,
  invitee_user_id,
  bind_source,
  created_at
)
select id, 10001, 10001, 'local_seed', now()
from invite
where not exists (
  select 1
  from invite_relations
  where invite_code_id = invite.id and invitee_user_id = 10001
);

with upsert_game as (
  insert into games (
    id,
    creator_user_id,
    main_guide_user_id,
    title,
    game_type,
    game_source,
    status,
    primary_category,
    primary_category_text,
    secondary_category,
    secondary_category_text,
    type,
    min_players,
    max_players,
    current_players,
    city_code,
    city_name,
    address,
    longitude,
    latitude,
    created_at,
    updated_at
  )
  values (
    10001,
    10001,
    10001,
    '本机测试局：行家领路人联调',
    'free',
    'app',
    'in_progress',
    'social',
    '社交局',
    'ai_exchange',
    'AI交流',
    'free',
    5,
    8,
    5,
    '110100',
    '北京',
    '北京市朝阳区本机测试点',
    116.397128,
    39.916527,
    now(),
    now()
  )
  on conflict (id) do update set
    creator_user_id = excluded.creator_user_id,
    main_guide_user_id = excluded.main_guide_user_id,
    title = excluded.title,
    game_type = excluded.game_type,
    game_source = excluded.game_source,
    status = excluded.status,
    primary_category = excluded.primary_category,
    primary_category_text = excluded.primary_category_text,
    secondary_category = excluded.secondary_category,
    secondary_category_text = excluded.secondary_category_text,
    type = excluded.type,
    min_players = excluded.min_players,
    max_players = excluded.max_players,
    current_players = excluded.current_players,
    city_code = excluded.city_code,
    city_name = excluded.city_name,
    address = excluded.address,
    longitude = excluded.longitude,
    latitude = excluded.latitude,
    updated_at = now()
  returning id
)
insert into game_members (game_id, user_id, role, status, joined_at)
select id, 10001, 'creator', 'active', now()
from upsert_game
on conflict (game_id, user_id) do update set
  role = excluded.role,
  status = excluded.status,
  quit_reason = null;

with room as (
  insert into chat_rooms (game_id, status, created_at)
  values (10001, 'active', now())
  on conflict (game_id) do update set status = 'active'
  returning id
)
insert into chat_room_members (room_id, user_id, status, created_at)
select id, 10001, 'active', now()
from room
on conflict (room_id, user_id) do update set status = 'active';

insert into chat_messages (
  room_id,
  sender_user_id,
  client_msg_id,
  message_type,
  content,
  status,
  created_at
)
select
  id,
  10001,
  'local-seed-welcome',
  'text',
  '本机测试消息：这个账号已经拥有行家和领路人身份，可用于测试 IM、组局和身份页面。',
  'sent',
  now()
from chat_rooms
where game_id = 10001
on conflict (room_id, sender_user_id, client_msg_id)
where client_msg_id is not null and client_msg_id <> ''
do update set
  content = excluded.content,
  status = excluded.status;

select setval(pg_get_serial_sequence('users', 'id'), greatest((select max(id) from users), 1), true);
select setval(pg_get_serial_sequence('games', 'id'), greatest((select max(id) from games), 1), true);
select setval(pg_get_serial_sequence('chat_rooms', 'id'), greatest((select max(id) from chat_rooms), 1), true);
