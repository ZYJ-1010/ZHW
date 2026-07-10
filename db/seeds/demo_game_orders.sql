insert into users(id, status, nickname, realname_status, created_at, updated_at)
values
  (91001, 'active', '李明', 'verified', now(), now()),
  (91002, 'active', '张专家', 'verified', now(), now()),
  (91003, 'active', '王引荐', 'verified', now(), now()),
  (91004, 'active', '王导师', 'verified', now(), now()),
  (91005, 'active', '刘设计师', 'verified', now(), now())
on conflict (id) do update set
  nickname = excluded.nickname,
  realname_status = excluded.realname_status,
  updated_at = now();

insert into games(id, creator_user_id, main_guide_user_id, title, game_type, game_source, status, min_players, max_players, current_players, city_code, city_name, created_at, updated_at)
values
  (92001, 91002, 91003, '产品架构咨询', 'free', 'app', 'in_progress', 5, 8, 5, '500100', '重庆市', now() - interval '10 days', now()),
  (92002, 91004, 91003, '品牌定位咨询', 'free', 'app', 'completed', 5, 8, 5, '500100', '重庆市', now() - interval '20 days', now()),
  (92003, 91005, 91003, 'UI设计服务', 'free', 'app', 'canceled', 5, 8, 5, '500100', '重庆市', now() - interval '25 days', now())
on conflict (id) do update set
  title = excluded.title,
  game_type = excluded.game_type,
  game_source = excluded.game_source,
  status = excluded.status,
  min_players = excluded.min_players,
  max_players = excluded.max_players,
  current_players = excluded.current_players,
  updated_at = now();

insert into game_members(game_id, user_id, role, status, joined_at)
values
  (92001, 91001, 'member', 'active', now() - interval '9 days'),
  (92001, 91002, 'expert', 'active', now() - interval '10 days'),
  (92001, 91003, 'guide', 'active', now() - interval '10 days'),
  (92002, 91001, 'member', 'active', now() - interval '19 days'),
  (92002, 91004, 'expert', 'active', now() - interval '20 days'),
  (92002, 91003, 'guide', 'active', now() - interval '20 days'),
  (92003, 91001, 'member', 'quit', now() - interval '24 days'),
  (92003, 91005, 'expert', 'active', now() - interval '25 days'),
  (92003, 91003, 'guide', 'active', now() - interval '25 days')
on conflict (game_id, user_id) do update set
  role = excluded.role,
  status = excluded.status;

insert into reports(id, game_id, reporter_user_id, target_user_id, report_type, content, status, handle_result, handled_at, created_at)
values
  (
    93001,
    92001,
    91001,
    91002,
    'service_dispute',
    '交易为平台正常订单，并非私下交易，可提供订单截图作为证明',
    'appealed',
    '{"result":"平台已收到申诉，正在复核订单和沟通记录。","appealContent":"交易为平台正常订单，并非私下交易，可提供订单截图作为证明","outcome":"processing","rewardPoints":0,"creditChange":0}',
    now() - interval '2 days',
    now() - interval '3 days'
  ),
  (
    93002,
    92002,
    91001,
    91004,
    'im_message',
    '沟通为正常服务交流，不存在骚扰行为，聊天记录可查证',
    'appealed',
    '{"result":"客服已进入会话核查。","appealContent":"沟通为正常服务交流，不存在骚扰行为，聊天记录可查证","outcome":"processing","rewardPoints":0,"creditChange":0}',
    now() - interval '4 days',
    now() - interval '5 days'
  ),
  (
    93003,
    92002,
    91003,
    91004,
    'user_complaint',
    '个人资料真实有效，可提供身份证明及学历证明',
    'handled',
    '{"result":"申诉已通过，本次举报不再影响信用。","appealContent":"个人资料真实有效，可提供身份证明及学历证明","outcome":"appeal_approved","rewardPoints":0,"creditChange":0}',
    now() - interval '7 days',
    now() - interval '8 days'
  ),
  (
    93004,
    92003,
    91003,
    91005,
    'service_dispute',
    '因不可抗力因素取消，非主观恶意',
    'closed',
    '{"result":"申诉材料不足，维持原处理结果。","appealContent":"因不可抗力因素取消，非主观恶意","outcome":"appeal_rejected","rewardPoints":0,"creditChange":-2,"creditTargetUserId":91005}',
    now() - interval '12 days',
    now() - interval '13 days'
  )
on conflict (id) do update set
  game_id = excluded.game_id,
  reporter_user_id = excluded.reporter_user_id,
  target_user_id = excluded.target_user_id,
  report_type = excluded.report_type,
  content = excluded.content,
  status = excluded.status,
  handle_result = excluded.handle_result,
  handled_at = excluded.handled_at;
