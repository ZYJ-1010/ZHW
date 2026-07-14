# 2026-06-30 全链路对接进度

## 本轮已完成
- 后台生产就绪检查补齐联调字段：
  - `ReadinessItem` 现在返回 `name`、`ready`、`critical`、`details`。
  - 新增 `wechat_url_link` 检查，用于确认邀请链接入口是否配置 HTTPS。
  - 腾讯地图检查增加非敏感 `details`，明确当前采用 `server_proxy` 模式，不向前端暴露地图 Key。
- 后台系统页对接新旧 readiness 字段：
  - 同时兼容 `status/required` 和 `ready/critical`。
  - `details` 会合并显示到消息列，便于上线联调排查。
- 后台局类型配置增加后端防跑偏校验：
  - `typeFilters` 允许 `all` 作为筛选项。
  - 可创建局类型必须属于后端已实现枚举：`free`、`standard`、`public_welfare`、`aa`、`crowdfund`、`deposit`、`condition`。
  - 避免后台配置出小程序可见、但后端创建失败的无效局类型。
- 地图附近局详情跳转接入真实局：
  - 小程序地图页从后端附近局接口拿到 `id` 后，会生成 `pages/game/detail/index?id=...`。
  - 保留 demo 数据无详情跳转，避免假数据误入真实详情页。

## 已验证
- `go test ./...`
- `go test ./internal/common/config -count=1`
- `go test ./internal/appapi -run TestAdminGameCategoryConfigFeedsAppHTTP -count=1`
- `node --check admin-web/src/main.js`
- `node --check miniprogram-client/pages/map/index.js`
- `npm run build` in `admin-web`

## 当前进度判断
- 小程序后端与后台主业务链路约 `82%`。
- 前后端全链路对接约 `78%`。
- 剩余优先级最高的缺口：
  - 地图页组队动作已接局详情和加入意图；玩家点位详情仍需产品确认私聊、弹层或详情页形态。
  - 消息页语音/图片/文件动作需要继续对接 OpenIM 文件消息能力。
  - 首页地球/关系网仍需进一步使用后端真实关系数据。
  - 部分个人中心页面仍有静态兜底数据，需要逐页替换为已有后端接口。
