# 首页动态地球与关系网对接进度

日期：2026-06-30

## 本轮完成

- 玩家首页读取后端 `/api/app/home` 的 `earth` 和 `visualization` 数据。
- 地球卡不再只依赖静态文案：
  - `earth.nodes` / `earth.onlineCount` 生成动态节点数描述。
  - `earth.heatPoints` / `nearbySummary.checkedInCount` 生成城市打卡/热力标签。
  - `earth.nearbyCount` / `nearbySummary.nearbyGameCount` 生成附近局标签。
- 关系网/元宇宙卡接入 `visualization.network`：
  - `network.nodes` 生成头像占位与节点数量。
  - `network.edges` 生成关系连接标签。
  - badge 优先展示动态节点数。
- 普通首页保留 `earth` 和 `visualization` 原始 payload，方便后续页面直接使用。
- mock 首页数据补齐 `earth` 和 `visualization.network`，本地 mock 与真实后端字段一致。

## 验证

- `node --check miniprogram-client/pages/home/player/index.js`
- `node --check miniprogram-client/pages/home/index.js`
- `node --check miniprogram-client/api/mock-data.js`
- `go test -count=1 ./internal/appapi -run 'TestAppHomeAndConnectionNetworkPayloadHTTP|TestLocationAndNearbyGamesFlow'`
- `go test -count=1 ./...`
- `npm run build` in `admin-web`
- `ponytail:` debt scan: no markers

## 当前判断

- 文档要求的“用户界面的地球和关系网是动态的”已从后端 payload 延伸到小程序玩家首页数据层。
- 后端/小程序/后台整体业务对接进度约 82%。
- 后端核心业务进度约 96.8%；剩余重点仍是线上三方联调、OpenIM 实例化联调、地图服务生产 Key 配置，以及少量前端待接入口。
