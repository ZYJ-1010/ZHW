# 创建局地点搜索对接进度 - 2026-06-30

## 本次完成

- 小程序创建局的地点入口改为打开后端地图搜索面板，不再要求前端直接绑定腾讯地图 Key。
- `miniprogram-client/api/modules/location.js` 增加 `/api/app/map/search` 和 `/api/app/map/reverse-geocode` 调用。
- `miniprogram-client/services/location.js` 增加地图搜索/逆地理解析服务封装，并统一抛出可读错误。
- `pages/game/create` 增加地点关键词搜索、结果选择、地点写回 `locationInfo` 的流程。
- 创建局发布 payload 改为优先提交 `cityCode/cityName/longitude/latitude`，保持和后端 `games.CreateRequest` 字段一致。
- mock 环境补齐 `/api/app/map/search`，本地开发也能跑通创建局地点搜索。

## 验证

- `node --check miniprogram-client/pages/game/create/index.js`
- `node --check miniprogram-client/services/location.js`
- `node --check miniprogram-client/api/modules/location.js`
- `node --check miniprogram-client/api/mock.js`
- `go test -count=1 ./internal/appapi -run TestTencentMapProxyHTTP`
- `go test -count=1 ./...`
- `npm run build` in `admin-web`

## 剩余关联缺口

- 创建局后端模型当前没有 `address` 字段，线上如果需要展示详细门牌地址，需要补数据库字段和 `games.CreateRequest`。
- 地图搜索结果当前用 `city` 作为 `cityName`，若后续要做更严格同城筛选，需要接入行政区划编码映射。
