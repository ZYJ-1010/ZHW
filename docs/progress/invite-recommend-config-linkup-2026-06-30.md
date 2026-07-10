# 邀请与系统推荐配置后端化 - 2026-06-30

## 本轮处理

- `services/go-api/internal/appapi/game_invite_handler.go`
  - 扩展 `GET /api/app/game-invites/player-config` 返回邀请页初始化配置，并优先读取 `system_configs.game.invite_config`：
    - `defaultBudget`
    - `defaultTitle`
    - `defaultDetail`
    - `playerIntroTemplate`
    - `expert`
    - `activityTypes`
    - `rewardRateConfig`
  - 保留原有 `minPlayerCount`、`maxPlayerCount`，兼容现有测试和页面。

- `db/seeds/system_configs.sql`
  - 新增 `game.invite_config` 测试数据。
  - 邀请页默认标题、详情、预算、活动类型、奖励比例和引荐语模板已落入数据库 seed。

- `miniprogram-client/pages/game/invite/index.js`
  - 移除页面内写死的默认行家、活动类型、标题、详情、引荐语。
  - `onLoad` 和第二步玩家选择都复用 `GET /api/app/game-invites/player-config` 初始化数据。
  - 页面样式不变，只调整数据来源。

- `miniprogram-client/pages/game/system-recommend/index.js`
  - 移除前端默认推荐分类。
  - 推荐分类只读取 `GET /api/app/game-invites/system-recommendations` 的 `categories`。

- `miniprogram-client/api/mock-data.js`
  - 补齐开发 mock 的邀请配置字段，方便无后端开发时页面不崩。
  - 真实联调与线上仍以 Go API 返回为准。

## 验证

- `node --check miniprogram-client/pages/game/system-recommend/index.js`
- `node --check miniprogram-client/pages/game/invite/index.js`
- `node --check miniprogram-client/api/mock-data.js`
- `go test -count=1 ./internal/appapi -run "Test.*Invite|Test.*Recommendation|Test.*Replay"`
- `go test -count=1 ./internal/systemconfig`

## 下一步

- 继续清 `game/confirm` 的 `DEFAULT_REPLAY_CONTEXT`。
- 继续清 `game/payment`、`game/apply`、`profile` 里会影响真实联调的业务默认数据。
