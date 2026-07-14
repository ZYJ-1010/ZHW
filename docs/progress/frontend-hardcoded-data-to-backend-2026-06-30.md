# 前端硬编码数据迁移到后端数据源 - 2026-06-30

## 本轮处理

- `miniprogram-client/pages/profile/service-center/my-games/index.js`
  - 删除页面内置的 5 条“我的局”业务示例卡片。
  - 页面卡片改为只来自后端接口：
    - `GET /api/app/games/player/manage`
    - `GET /api/app/games/my/manage`
    - `GET /api/app/games/favorites/my`
  - 接口失败时展示空态，不再用前端假订单掩盖联调问题。

- `miniprogram-client/api/modules/game.js`
  - 新增 `getMyFavoriteGames()`，对接收藏局列表。

- `miniprogram-client/services/game.js`
  - 新增 `getMyFavoriteGames()` 服务封装。

- `miniprogram-client/pages/game/create/index.js`
  - 删除前端默认局类型 `DEFAULT_GAME_TYPES`。
  - 创建页局类型只读取 `GET /api/app/games/category-config`。
  - 后台配置加载失败或为空时不允许发布，避免前端硬编码类型绕过后台配置。
  - 删除前端默认收益模板 `DEFAULT_PROFIT_TEMPLATES`。
  - 收益模板只读取 `GET /api/app/games/profit-templates`，接口失败或为空时不允许发布。

- `miniprogram-client/api/mock.js`
  - 补齐开发 mock 下的 `GET /api/app/games/favorites/my`，仅复用已有 mock 管理数据，保证本地无后端时页面不报 404。

## 数据来源口径

- 真实联调和线上环境必须关闭小程序 mock，业务数据来自 Go API 和数据库。
- 后台已有受控验收数据入口 `POST /api/admin/ai-data/acceptance-fixture`，用于准备用户、局、收藏、评价、IM 等验收数据。
- 小程序页面不应再维护业务示例卡片；后续继续把列表、类型、推荐、筛选、文案配置迁到后端接口或后台系统配置。

## 5 天收口优先级

1. 清理小程序主链路页面业务硬编码：创建局、邀请、我的局、地图附近局、消息、个人中心。
2. 后台配置类数据进入 `system_configs` 或专门业务表，由后台维护，小程序只读接口结果。
3. 测试/验收数据由后台受控夹具、seed 或管理端创建，不写在页面组件里。
4. 保持前端视觉样式不动，只改路由、接口、数据映射和必要空态。
5. 每个迁移切片都跑 JS 语法检查和对应后端测试，避免为了赶工制造隐性断链。

## 本轮验证

- `node --check miniprogram-client/pages/profile/service-center/my-games/index.js`
- `node --check miniprogram-client/pages/game/create/index.js`
- `node --check miniprogram-client/api/modules/game.js`
- `node --check miniprogram-client/services/game.js`
- `node --check miniprogram-client/api/mock.js`
- `go test -count=1 ./internal/appapi -run "Test.*Favorite|Test.*Category|TestGameCategory|TestAdmin.*Game"`
- `go test -count=1 ./internal/appapi -run "Test.*Revenue|Test.*Profit|Test.*Category"`
