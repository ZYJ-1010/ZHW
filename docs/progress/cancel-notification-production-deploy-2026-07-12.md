# 取消组局通知生产部署记录 - 2026-07-12

## 发布范围

- 玩家取消组局、行家取消组局时生成中文通知标题，并在通知内容中携带取消人和取消原因。
- 玩家、行家、领路人收到取消通知后，可进入取消通知详情页；通知动作路由到 `pages/game/guide-cancel/index`。
- 取消通知详情接口返回取消方角色、名称、原因、组局编号和时间线。
- 本次仅发布 Go API 后端；未修改生产数据库、Nginx、Redis、PostgreSQL、后台网页或小程序前端容器。

## 发布方式

- 先下载生产当前源码做差异审计，只向生产版本应用本次取消通知补丁，保留其他会话已经上线的修改。
- 实际替换文件：`services/go-api/internal/appapi/game_handler.go`、`services/go-api/internal/appapi/notification_handler.go`。
- `game_invite_handler.go` 在生产环境已与本地目标版本一致，因此没有重复替换。
- 候选镜像：`zhw-go-api-cancel-notification:20260712-112325`，镜像 ID `sha256:381da1e6f07d9f9401ad99fbfcb2412970f2acbf433ab39b79d5fd0d7cd4d985`。
- 生产镜像 ID：`sha256:689472763070545a894db4b2e2f8f1d2470ab03165669f6af33bb4a9b2bafe72`。
- 生产备份目录：`/opt/zhw-mini/.deploy-backups/cancel-notification-20260712-112325/services/go-api/internal/appapi`。
- 发布后源码 SHA-256：`game_handler.go` 为 `39aa129accf8671df4fd2676d6379150e55b3dd3e33b5384803ffe6d8f9aaf5f`，`notification_handler.go` 为 `2ce20cae1998ed5a5bd651b7628dc7b6893e14e99522b3bb041d6d59a72f8005`。

## 验证结果

- `deploy-go-api-1` 状态为 `running`，重启次数为 `0`。
- 内网 `GET /api/app/health` 返回成功。
- 公网 `GET https://api.haowan.net.cn/api/app/health` 返回 HTTP 200，响应状态为 `ok`。
- 未登录访问取消通知详情接口返回 HTTP 401，鉴权保护正常。
- 生产源码已复核存在 `gameCancelDetail` 和 `serviceCancelNotificationDetail` 逻辑。

## 回滚位置

- 将备份目录中的 `game_handler.go` 和 `notification_handler.go` 恢复到 `/opt/zhw-mini/services/go-api/internal/appapi/`。
- 在 `/opt/zhw-mini/deploy` 仅重新构建并启动 `go-api`。
