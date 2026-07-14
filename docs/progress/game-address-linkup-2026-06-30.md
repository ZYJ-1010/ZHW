# 组局详细地址字段对接进度 - 2026-06-30

## 本次完成

- `games.Game` 和 `games.CreateRequest` 增加 `address` 字段。
- 创建局校验增加 `address <= 255`，并在创建前 trim。
- SQL 仓库创建、更新、详情、列表查询全部读写 `games.address`。
- 新增迁移 `db/migrations/000031_game_address.sql`。
- 小程序创建局发布 payload 增加 `address`，承接地图搜索结果。
- mock 创建局保存并返回 `address`。
- 后台开局表单增加详细地址、经度、纬度输入；组局列表优先展示详细地址。
- app/admin OpenAPI 同步 `address` 字段。

## 验证

- `node --check miniprogram-client/pages/game/create/index.js`
- `node --check miniprogram-client/api/mock.js`
- `node --check admin-web/src/main.js`
- `go test -count=1 ./internal/games`
- `go test -count=1 ./internal/appapi -run 'TestPlayerCanCreateGameWithoutVerifiedIdentity|TestAppCreateGameRequiresFreeType|TestAdminCanCreateDocumentedGameTypes'`
- `go test -count=1 ./...`
- `npm run build` in `admin-web`

## 进度判断

- 后端业务核心约 97.5%。
- 小程序/后台/后端全链路约 84%。
- 地图相关链路从“只有经纬度/城市”推进到“POI 地址可创建、可持久化、后台可核对”。
