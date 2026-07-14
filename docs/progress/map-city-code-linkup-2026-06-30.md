# 地图行政区划编码对接进度 - 2026-06-30

## 本次完成

- `lbs.MapPlace` 增加 `cityCode`。
- 腾讯地图搜索、地址解析、逆地址解析把 `ad_info.adcode` 映射为 `cityCode`。
- 小程序 mock 地图搜索结果补齐 `cityCode`，本地创建局联调不会缺同城筛选字段。
- app OpenAPI 地图搜索说明同步 `cityCode`。

## 验证

- `node --check miniprogram-client/api/mock.js`
- `node --check admin-web/src/main.js`
- `go test -count=1 ./internal/lbs`
- `go test -count=1 ./internal/appapi -run TestTencentMapProxyHTTP`
- `go test -count=1 ./...`
- `npm run build` in `admin-web`
- ponytail debt scan: no marker.

## 进度判断

- 后端业务核心约 97.8%。
- 地图/开局链路约 88%。
- 全链路对接约 85%。
