# 入局角色与重复拒绝生产发布记录 - 2026-07-12

## 发布范围

- 入局申请后端校验受保护角色：申请 `guide` 或 `expert` 时，账号必须已开通对应平台身份。
- 领路人申请审批通过后按本局角色进入；本局尚无主领路人时成为主领路人。
- 修正 `game_applications` 唯一约束，仅限制同一用户、同一局同时存在一条 `pending` 申请，允许保留多次 `rejected` 历史。

## 发布方式

- 生产备份目录：`/opt/zhw-mini/.deploy-backups/game-role-repeat-reject-20260712-095935`。
- 完整数据库备份：上述目录下 `db/zhw_mini.dump`。
- 基于生产当前 Go API 源码叠加本次最小角色校验片段，在隔离目录完成 Docker 候选镜像构建。
- 执行迁移 `db/migrations/000045_game_application_pending_unique.sql`。
- 仅重建并启动 `go-api`，未重建 PostgreSQL、Redis、Nginx 或后台网页容器。
- 小程序报名页角色提交改动保留在本地项目，开发者工具重新编译后生效；正式版仍需按微信发布流程上传。

## 验证结果

- 候选镜像和生产 Go API 镜像均构建成功。
- 新索引 `uk_game_applications_pending` 已存在。
- 在事务中将申请 `8` 更新为历史申请 `7` 的 `rejected` 状态成功，随后回滚；证明第二次拒绝不再触发唯一键冲突。
- 申请 `8` 的生产真实状态仍为 `pending/member`，验证过程未修改业务数据。
- `go-api` 容器状态为 `running`，重启次数为 `0`。
- 内网与公网 `GET /api/app/health` 均返回 HTTP 200。
- 未登录调用申请接口返回 HTTP 401，鉴权保护正常。
- 启动日志未发现 `panic`、`fatal` 或启动错误。

## 回滚位置

- Go 源码回滚：恢复备份目录中的 `services/go-api/internal/appapi/game_handler.go` 后，仅重建并启动 `go-api`。
- 数据库完整回滚：使用 `db/zhw_mini.dump`。迁移本身为放宽历史状态约束，如仅回滚代码通常无需恢复旧数据库约束。
