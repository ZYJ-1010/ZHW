# 后台开局地图搜索对接进度 - 2026-06-30

## 本次完成

- 后台开局表单增加地点搜索入口。
- 调用已有 `GET /api/admin/map/search`，复用后端地图代理，不让后台前端直接持有地图 Key。
- 地点搜索结果可一键填充 `cityCode/cityName/address/longitude/latitude`。
- 后台表单保留手动编辑能力，便于线上联调时修正地图返回结果。
- 新增少量样式，复用后台现有面板/按钮风格，没有改动整体布局。

## 验证

- `node --check admin-web/src/main.js`
- `npm run build` in `admin-web`
- `go test -count=1 ./internal/appapi -run TestTencentMapProxyHTTP`
- ponytail debt/audit scan: no marker, no obvious new abstraction.

## 进度判断

- 后端业务核心约 97.5%。
- 后台和小程序地图/开局链路约 86%。
- 全链路对接约 84.5%。
