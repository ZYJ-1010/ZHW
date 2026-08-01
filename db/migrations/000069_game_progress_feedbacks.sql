-- 局协作进度反馈必须持久化；此前仅保存在 Go 进程内，服务重启后会丢失。
create table if not exists game_progress_feedbacks (
  id bigserial primary key,
  game_id bigint not null,
  user_id bigint not null,
  progress int not null check (progress between 0 and 100),
  content varchar(500),
  file_ids jsonb not null default '[]'::jsonb,
  created_at timestamptz not null default now()
);

create index if not exists idx_game_progress_feedbacks_game_created
  on game_progress_feedbacks(game_id, created_at asc, id asc);
