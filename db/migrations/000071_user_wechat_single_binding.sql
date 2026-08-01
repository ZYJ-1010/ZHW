delete from user_wechat_accounts duplicate
using user_wechat_accounts retained
where duplicate.user_id = retained.user_id
  and duplicate.id > retained.id;

create unique index if not exists ux_user_wechat_accounts_user_id
on user_wechat_accounts (user_id);
