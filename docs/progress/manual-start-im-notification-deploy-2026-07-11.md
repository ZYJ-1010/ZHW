# 组局开局 IM 与成员通知生产发布记录 - 2026-07-11

## 发布内容

- 用户执行组局开局后，幂等创建或同步该局 IM 房间。
- 在 IM 历史中持久化“《局名称》局开始了”文本消息。
- 通过现有 WebSocket 房间 Hub 向在线成员实时广播该消息。
- 给全部局成员创建 `game_started` App 站内通知，不创建微信订阅消息任务。

## 发布方式

- 先下载生产当前 `game_handler.go`，以生产版本为基线叠加本次改动，未覆盖本地其他未发布功能。
- 生产备份：`/opt/zhw-mini/.deploy-backups/manual-start-notifications-20260711-193349`。
- 发布文件 SHA-256：`adc43d4db6212c6a9a1934728958296e4c2a72293f650d3872eaa283fb182f95`。
- 重新构建并启动生产 `go-api` 容器成功。

## 测试结果

- 使用生产发布文件覆盖编译运行定向测试：
  - `TestManualStartPersistsBroadcastsAndNotifiesMembers` 通过。
  - `TestIMWebSocketFlow` 通过。
  - `internal/im`、`internal/notifications` 测试通过。
- 生产内网与公网 `/api/app/health` 均返回 HTTP 200。
- 生产 `/api/app/games/1/manual-start` 未登录请求返回 HTTP 401，确认路由在线并受鉴权保护。
- 生产源码哈希与已测试发布包一致，容器日志无启动或运行错误。
- 为避免污染真实业务数据，未在生产创建测试用户或测试组局；开局完整业务链路由生产发布文件的本地集成测试覆盖。

## 通知范围说明

- 本能力仅使用局内 IM、WebSocket 和 App 自身站内通知。
- 不触发微信订阅通知。

## 19:44 修正发布

- 移除开局通知的 `NeedWechat` 和微信任务数据，不再创建 `game_started` 微信订阅任务。
- 将 `game_started` 归入个人“消息”页的“组局动态”分区。
- 点击开局消息直接跳转 `pages/im/room/index?gameId={gameId}`，进入对应局内 IM。
- 修正发布生产备份：`/opt/zhw-mini/.deploy-backups/manual-start-inapp-only-20260711-194426`。
- 发布包定向测试、生产编译、内外网健康检查及线上文件哈希校验均通过。
- 数据库只读核查确认：错误版本短暂在线期间创建的 `game_started` 微信订阅任务数量为 `0`。
