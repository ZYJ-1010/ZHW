# 地图附近局与玩家点位对接进度

日期：2026-06-30

## 本轮完成

- 小程序地图页「去组队」从待接入提示改为进入局详情页，并携带 `intent=join`，复用现有报名链路。
- 小程序地图页玩家 marker 增加「查看资料」入口，真实玩家数据返回 `userId/id` 时跳转到成员详情页。
- 后端 `/api/app/games/nearby` 增加附近玩家数据：
  - `onlinePlayers`
  - `offlinePlayers`
  - `onlinePlayerCount`
  - `offlinePlayerCount`
- LBS 服务新增 `NearbyUsers`，基于用户最近位置按半径筛选附近用户。
- SQL LBS 仓库新增 `LatestLocations`，用于读取每个用户最新位置。
- 附近玩家展示名称不回退手机号，昵称为空时显示 `玩家{id}`。

## 验证

- `node --check miniprogram-client/pages/map/index.js`
- `node --check miniprogram-client/pages/game/detail/index.js`
- `go test -count=1 ./internal/lbs`
- `go test -count=1 ./internal/appapi -run TestLocationAndNearbyGamesFlow`
- `go test -count=1 ./...`
- `npm run build` in `admin-web`
- `ponytail:` debt scan: no markers

## 当前判断

- 小程序地图主链路从“附近局展示”推进为“附近局 + 附近玩家点位 + 组队入口 + 玩家资料入口”。
- 后端/小程序/后台整体业务对接进度约 81%。
- 后端核心业务进度约 96.8%，剩余主要是线上三方配置、真实微信/IM/OpenIM/地图服务联调，以及若干前端页面的待接入口。
