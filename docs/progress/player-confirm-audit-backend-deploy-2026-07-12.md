# 玩家确认与行家审核后台重新部署记录 - 2026-07-12

## 部署范围

- 仅重新构建并强制重建生产 `go-api` 容器。
- 未上传小程序前端，未修改生产数据库、Nginx、Redis、PostgreSQL 或后台网页。
- 发布前确认生产 `game_invite_handler.go`、`review_handler.go` 与本地目标版本 SHA-256 一致。
- 生产 `game_handler.go` 包含更新的取消通知逻辑，因此保留生产最新版本，未使用本地旧文件覆盖。

## 发布结果

- 生产镜像：`sha256:db360fc744ef75ba4c135f64e3f25ccb75a19e100808953b63342d4eba71997e`。
- 容器 `deploy-go-api-1` 已重建，状态为 `running`，重启次数为 `0`。
- 内网 `GET http://127.0.0.1:8080/api/app/health` 返回 HTTP 200。
- 公网 `GET https://api.haowan.net.cn/api/app/health` 返回 HTTP 200，服务状态为 `ok`。
- 启动日志显示 `go-api listening on :8080`，未发现启动错误。

## 回滚位置

- 生产源码备份：`/opt/zhw-mini/.deploy-backups/player-confirm-ui-20260712-121632`。
- 备份包含 `game_invite_handler.go`、`game_handler.go`、`review_handler.go` 及 `SHA256SUMS`。
