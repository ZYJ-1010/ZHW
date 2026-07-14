# 局详情动作对接

日期：2026-06-30

## 本轮完成

- 局详情“感兴趣”接入 `/api/app/games/{gameId}/favorite`。
- 小程序 API/service 层补齐 `favoriteGame(gameId)`。
- mock 环境补齐收藏接口，保持本地联调路径一致。
- 局详情地点点击跳地图路线页并携带 `gameId`。
- 局详情评价统计跳评价页，参与者点击跳参与者列表。
- 底部工具栏：
  - 分享继续走原分享弹窗。
  - 引荐跳邀请玩家页。
  - 签到跳地图打卡页。
  - 打招呼跳打招呼页。

## 验证

- `node --check miniprogram-client/api/modules/game.js`
- `node --check miniprogram-client/services/game.js`
- `node --check miniprogram-client/pages/game/detail/index.js`
- `node --check miniprogram-client/api/mock.js`
- `go test -count=1 ./internal/appapi -run "Test.*Favorite|Test.*Review|Test.*SuccessDetail|Test.*GameDetail"`

