-- 历史版本曾将手机号验证视为实名完成，导致少量未完成个人实名认证的
-- 用户被开通行家或领路人。停用这些角色，待个人实名认证通过后再由后台开通。
update user_roles role
set status = 'inactive'
where role.role_code in ('expert', 'guide')
  and role.status = 'active'
  and not exists (
    select 1
    from identity_verification_records identity_record
    where identity_record.user_id = role.user_id
      and identity_record.status = 'verified'
      and coalesce(identity_record.real_name_ciphertext, '') <> ''
      and coalesce(identity_record.id_card_ciphertext, '') <> ''
  )
  and not exists (
    select 1
    from realname_auth_records legacy_record
    where legacy_record.user_id = role.user_id
      and legacy_record.status = 'verified'
  );

update guide_qualification_records qualification
set condition_met = false,
    payment_met = false,
    guide_open_status = 'waiting_condition',
    updated_at = now()
where not exists (
  select 1
  from user_roles role
  where role.user_id = qualification.user_id
    and role.role_code = 'guide'
    and role.status = 'active'
);
