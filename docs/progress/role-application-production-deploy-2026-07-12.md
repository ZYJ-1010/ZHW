# 角色申请校验生产发布记录 - 2026-07-12

## 发布范围

- 角色已激活时禁止重复申请。
- 同角色审核拒绝后 7 天内禁止重新申请。
- 行家与领路人的申请历史按角色独立校验。
- 新增 `role already active` 和 `role application reapply cooldown` 的 HTTP 409 错误映射。

## 发布方式

- 发布前下载并核对生产目标文件，确认差异仅包含本次角色申请校验。
- 生产源码备份：`/opt/zhw-mini/.deploy-backups/role-application-20260712-090734`。
- 在 `/tmp/zhw-role-application-20260712-090734` 使用生产当前源码叠加本次两个文件，隔离 Docker 构建成功后再替换生产源码。
- 仅重建并启动 `go-api`，未修改数据库、Nginx、Redis、PostgreSQL 或后台网页容器。

## 验证结果

- 生产文件 SHA-256 与本地目标文件一致。
- `go-api` 容器状态为 `running`，重启次数为 `0`。
- 内网与公网 `GET /api/app/health` 均返回 HTTP 200。
- 公网未登录请求 `GET /api/app/role-applications/status-config` 返回 HTTP 401，鉴权保护正常。
- 发布后角色、角色状态配置、权益配置和首页接口持续返回 HTTP 200。
- 启动日志未发现 `panic`、`fatal` 或启动错误。
