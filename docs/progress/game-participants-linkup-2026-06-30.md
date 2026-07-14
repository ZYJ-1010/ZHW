# 局参与者列表对接

日期：2026-06-30

## 本轮完成

- 小程序 API/service 层补齐 `GET /api/app/games/{gameId}/members`。
- 参与者页进入时按 `gameId` 拉取真实成员列表。
- 将后端成员 DTO 映射为现有 `participant-card` 展示字段，不改页面样式。
- 点击参与者进入成员详情页；缺少成员 id 时兜底进入个人页。
- mock 环境补齐 `/api/app/games/{gameId}/members`，本地联调字段与真实接口一致。

## 验证

- `node --check miniprogram-client/api/modules/game.js`
- `node --check miniprogram-client/services/game.js`
- `node --check miniprogram-client/pages/game/participants/index.js`
- `node --check miniprogram-client/api/mock.js`
- `go test -count=1 ./internal/appapi -run "Test.*Members|Test.*Participant|Test.*GameApplications"`

