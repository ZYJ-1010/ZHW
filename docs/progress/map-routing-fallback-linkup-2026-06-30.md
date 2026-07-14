# 地图页入口路由兜底对接

日期：2026-06-30

## 本轮完成

- 附近玩法点有真实 `gameId` 时继续进入局详情。
- 演示点或缺少真实 `gameId` 时跳局前大厅，不再停留在“详情页待接入”。
- 去组队动作统一带 `intent=join&from=nearbyMap`，真实局进详情，兜底进大厅。
- 附近玩家有真实用户 id 时进入成员详情，无真实 id 时进入个人页。
- 地图页底部导航恢复到首页、消息、我的、元宇宙等真实路由。

## 验证

- `node --check miniprogram-client/pages/map/index.js`
- `go test -count=1 ./internal/appapi -run "TestNearby|TestMap|TestTencentMapProxyHTTP|TestHomeNearby"`

