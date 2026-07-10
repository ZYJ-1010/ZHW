alter table reviews
  add column if not exists tags jsonb not null default '[]'::jsonb;

create index if not exists idx_reviews_again_intent on reviews(again_intent);
