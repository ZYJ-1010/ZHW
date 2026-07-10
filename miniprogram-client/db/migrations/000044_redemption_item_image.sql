alter table redemption_items
  add column if not exists image_url varchar(500) default '';
