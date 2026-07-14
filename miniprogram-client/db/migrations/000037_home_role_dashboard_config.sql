insert into system_configs (config_key, config_value, status)
values (
  'home.role_dashboard_config',
  '{
    "version": "2026-07-04",
    "roles": {
      "player": {
        "quickActions": [
          {
            "id": "create",
            "title": "发起组局",
            "desc": "创建你的带局房间",
            "icon": "📍",
            "theme": "pink",
            "route": "pages/game/create/index"
          },
          {
            "id": "lobby",
            "title": "局前大厅",
            "desc": "准备就绪加入一局",
            "theme": "cyan",
            "route": "pages/game/hall/index",
            "routeIcon": true
          }
        ]
      },
      "guide": {
        "quickActions": [
          {
            "id": "lobby",
            "title": "局前大厅",
            "desc": "准备加入一局",
            "theme": "pink",
            "route": "pages/game/hall/index",
            "routeIcon": true
          },
          {
            "id": "invite",
            "title": "我的邀约",
            "desc": "管理连接的玩家",
            "icon": "📍",
            "theme": "cyan",
            "route": "pages/profile/service-center/invite/overview/index",
            "routeIcon": true
          }
        ],
        "network": {
          "title": "我的关系网络",
          "status": "实时连接中",
          "locationText": "核心区",
          "items": [
            { "id": "relations", "key": "relations", "icon": "👑", "name": "累计连接" },
            { "id": "strong", "key": "strongRelations", "icon": "🎓", "name": "强关系" },
            { "id": "nodes", "key": "onlineNodes", "icon": "👶", "name": "动态节点" },
            { "id": "income", "key": "income", "icon": "🏛️", "name": "本周收益" },
            { "id": "more", "key": "more", "icon": "+", "name": "更多", "desc": "待加入", "dashed": true }
          ],
          "buttons": [
            { "text": "管理我的连接", "route": "pages/relation/network/index", "primary": true },
            { "text": "查看分润", "route": "pages/profile/service-center/invite/income/index" }
          ]
        },
        "recommendation": {
          "title": "推荐行家"
        }
      }
    }
  }'::jsonb,
  'active'
)
on conflict (config_key) do nothing;
