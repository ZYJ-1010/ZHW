create table if not exists user_task_progress (
  user_id bigint not null,
  task_code varchar(100) not null,
  completed_at timestamptz not null default now(),
  primary key (user_id, task_code)
);
create index if not exists idx_user_task_progress_user on user_task_progress(user_id, completed_at desc);
