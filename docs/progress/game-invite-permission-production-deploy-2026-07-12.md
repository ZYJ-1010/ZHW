# 局引荐权限生产发布记录 - 2026-07-12

## 发布范围

- 续局引荐最终提交接口增加服务端权限兜底。
- 仅原局创建者或原局主领路人可发起引荐。
- 普通玩家、行家、普通领路人及局外用户返回 HTTP 403。

## 发布方式

- 生产目标文件备份：`/opt/zhw-mini/.deploy-backups/invite-permission-20260712-093644`。
- 在 `/tmp/zhw-invite-permission-20260712-093644` 基于生产当前 Go API 源码叠加本次两个文件，并完成隔离 Docker 构建。
- 仅替换 `services/go-api/internal/games/service.go` 和对应单元测试文件。
- 仅重建并启动 `go-api`，未修改数据库、Nginx、Redis、PostgreSQL 或后台网页容器。

## 验证结果

- 本地 `go test ./internal/games` 通过。
- 本地 `appapi` 包编译通过。
- 隔离候选镜像与生产 Go API 镜像均构建成功。
- 生产文件 SHA-256 与本地目标文件一致。
- `go-api` 容器状态为 `running`，重启次数为 `0`。
- 内网与公网 `GET /api/app/health` 均返回 HTTP 200。
- 公网未登录请求 `GET /api/app/game-invites/permission?gameId=1` 返回 HTTP 401，鉴权保护正常。
- 启动日志未发现 `panic`、`fatal` 或启动错误。

## 回滚位置

- 如需回滚，将备份目录中的两个文件恢复至 `/opt/zhw-mini/services/go-api/internal/games/`，再仅重建并启动 `go-api`。
