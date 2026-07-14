# 原局引荐邀请生产部署记录 - 2026-07-12

## 发布范围

- 新增 `POST /api/app/game-invites/current`。
- 领路人引荐玩家和行家时，邀请记录绑定来源局 `gameId`，不再创建 `replayGameId`。
- 玩家和行家分别生成站内邀请通知及微信通知任务。
- 配对邀请继续执行“玩家先确认、行家后确认”；双方接受后直接进入原局，无需发起人再次审批。
- 本次仅发布 Go API 后端；小程序前端样式和调用改动未随本次后端部署发布。

## 发布方式

- 下载生产当前 `server.go` 和 `game_invite_handler.go` 与本地目标文件进行差异审计，确认差异仅为本次路由与处理函数。
- 隔离候选镜像：`zhw-go-api-current-game-invite:20260712-141953`。
- 候选镜像 ID：`sha256:eda19d6d5a3d6f773ce6fce7deda3550dd00cc33434caae199c4ab400a243844`。
- 生产镜像 ID：`sha256:9b2853cbfda55caff9105f84000ba7e2a82f13dac3407d7447a26b6d16274a59`。
- 生产备份目录：`/opt/zhw-mini/.deploy-backups/current-game-invite-20260712-141953`。
- 仅重建并启动 `go-api`，未修改 PostgreSQL、Redis、Nginx、数据库结构或后台网页。

## 文件校验

- `services/go-api/internal/appapi/server.go`：`f7e71e6091f1c3d77937a6a06e4b2994a1909fc56aa6d109106ccc66bce80481`
- `services/go-api/internal/appapi/game_invite_handler.go`：`85d3f793429a1fcc118434bfc95a35ae96455502cddab561a61efc4703c1931f`
- 生产文件 SHA-256 与本地发布文件一致。

## 验证结果

- 本地 `appapi` 包编译检查通过。
- 分组邀请领域测试通过。
- 新增原局邀请 HTTP 定向测试通过。
- 生产容器 `deploy-go-api-1` 状态为 `running`，重启次数为 `0`。
- 内网 `GET /api/app/health` 返回 HTTP 200。
- 公网 `GET https://api.haowan.net.cn/api/app/health` 返回 HTTP 200。
- 公网未登录请求 `POST /api/app/game-invites/current` 返回 HTTP 401，路由与鉴权正常。
- 启动日志未发现 `panic`、`fatal` 或监听错误。

## 回滚方式

- 将 `/opt/zhw-mini/.deploy-backups/current-game-invite-20260712-141953/services/go-api/internal/appapi/` 中的 `server.go` 和 `game_invite_handler.go` 恢复到生产对应目录。
- 在 `/opt/zhw-mini/deploy` 仅重新构建并启动 `go-api`。
