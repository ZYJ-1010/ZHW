do $$
declare
  current_official_id bigint;
  target_official_id constant bigint := 88888;
  ref record;
begin
  select id
    into current_official_id
  from users
  where account_type = 'platform_official'
  for update;

  if current_official_id is null or current_official_id = target_official_id then
    return;
  end if;

  if exists (select 1 from users where id = target_official_id) then
    raise exception '平台官方豹子号 % 已被其他用户占用', target_official_id;
  end if;

  -- 用户关联表尚未建立外键级联规则，逐列迁移后再更新主用户 ID，确保邀请码、
  -- 邀请关系和所有历史运营记录继续归属同一个平台官方账号。
  for ref in
    select table_schema, table_name, column_name
    from information_schema.columns
    where table_schema = 'public'
      and table_name <> 'users'
      and column_name in (
        'user_id', 'owner_user_id', 'inviter_user_id', 'invitee_user_id',
        'creator_user_id', 'main_guide_user_id', 'operator_user_id',
        'leader_user_id', 'member_user_id', 'connected_user_id',
        'reporter_user_id', 'reviewer_user_id', 'sender_user_id',
        'target_user_id', 'uploader_user_id', 'player_user_id',
        'expert_user_id', 'started_by_user_id'
      )
  loop
    execute format(
      'update %I.%I set %I = $1 where %I = $2',
      ref.table_schema, ref.table_name, ref.column_name, ref.column_name
    ) using target_official_id, current_official_id;
  end loop;

  update users
  set id = target_official_id,
      updated_at = now()
  where id = current_official_id;
end
$$;
