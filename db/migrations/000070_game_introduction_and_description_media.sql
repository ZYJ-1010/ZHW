-- 局介绍与局详情媒体为正式发布数据，不使用微信临时路径或页面缓存保存。
alter table games add column if not exists introduction varchar(200);
alter table games add column if not exists description_media jsonb not null default '[]'::jsonb;
