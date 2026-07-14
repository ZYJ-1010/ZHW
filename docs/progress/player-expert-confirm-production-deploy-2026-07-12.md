# 玩家确认后行家确认流程生产部署记录 - 2026-07-12

## 部署范围

- 玩家确认与行家确认按同一 `inviteGroupId` 分组，避免同局不同邀请组串状态。
- 玩家确认后向对应行家发送“玩家已确认，等待你的确认”通知，并跳转到该行家的确认邀请。
- 行家确认页分别返回玩家与行家的确认状态；玩家确认后显示“已确认”，行家按钮可用。
- 本次仅部署外层 `services/go-api` 后端；未部署小程序前端，未修改内层历史后端副本、数据库、Nginx、Redis 或 PostgreSQL。

## 发布与回滚

- 生产目标文件：`/opt/zhw-mini/services/go-api/internal/appapi/game_invite_handler.go`。
- 发布前生产文件 SHA-256：`afc534564ef4b328a0974bedee8c77733ca2ab3f11dce77349338ce97787f4ab`。
- 发布后生产文件 SHA-256：`be5ef0e4fa8ca5a2ac97fdde0494f3f64f03ccd1ee92ce1e83b96904b82e1a85`。
- 生产备份目录：`/opt/zhw-mini/.deploy-backups/player-expert-confirm-20260712-155855`。
- 生产镜像 ID：`sha256:5a8144c6dbb5367f3ec2fdeee1c67e91c54824881b3f554425b10ebe1080279c`。
- 回滚时恢复备份目录中的 `game_invite_handler.go`，然后在 `/opt/zhw-mini/deploy` 仅重新构建并启动 `go-api`。

## 验证结果

- 本地 HTTP 回归通过：玩家确认后行家收到通知，行家确认页返回 `canConfirm=true`，行家确认后返回成功页。
- 前端 `pages/game/audit-detail/index.js` 语法检查通过。
- 生产容器 `deploy-go-api-1` 状态为 `running`，重启次数为 `0`。
- 内网 `GET http://127.0.0.1:8080/api/app/health` 返回 HTTP 200。
- 公网 `GET https://api.haowan.net.cn/api/app/health` 返回 HTTP 200。
- 公网未登录请求邀请进度接口返回 HTTP 401，鉴权保护正常。
- 启动日志显示 `go-api listening on :8080`，未发现启动错误。

