-- 将旧版本仅包含任务局、探索局的配置补齐为一期要求的四类局。
do $$
declare
  cfg jsonb;
  categories jsonb;
begin
  select config_value into cfg
    from system_configs
   where config_key = 'game.category_config';

  if cfg is null or jsonb_typeof(cfg->'primaryCategories') <> 'array' then
    return;
  end if;

  categories := cfg->'primaryCategories';
  if not exists (
    select 1 from jsonb_array_elements(categories) item where item->>'key' = 'social'
  ) then
    categories := categories || '[{"key":"social","name":"社交局","icon":"category-social","visible":true,"order":30,"children":[{"key":"meal","name":"饭局","visible":true,"order":10,"selectable":true},{"key":"board_game","name":"桌游局","visible":true,"order":20,"selectable":true},{"key":"friend","name":"交友局","visible":true,"order":30,"selectable":true}]}]'::jsonb;
  end if;
  if not exists (
    select 1 from jsonb_array_elements(categories) item where item->>'key' = 'growth'
  ) then
    categories := categories || '[{"key":"growth","name":"成长局","icon":"category-growth","visible":true,"order":40,"children":[{"key":"reading","name":"读书局","visible":true,"order":10,"selectable":true},{"key":"fitness","name":"健身局","visible":true,"order":20,"selectable":true},{"key":"study","name":"学习共修局","visible":true,"order":30,"selectable":true}]}]'::jsonb;
  end if;

  update system_configs
     set config_value = jsonb_set(cfg, '{primaryCategories}', categories, true),
         updated_at = now()
   where config_key = 'game.category_config';
end $$;
