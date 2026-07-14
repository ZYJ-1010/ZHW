alter table reviews
  add column if not exists nps_score smallint;

do $$
begin
  if not exists (
    select 1
    from pg_constraint
    where conname = 'reviews_nps_score_check'
  ) then
    alter table reviews
      add constraint reviews_nps_score_check
      check (nps_score is null or nps_score between 0 and 10);
  end if;
end $$;
