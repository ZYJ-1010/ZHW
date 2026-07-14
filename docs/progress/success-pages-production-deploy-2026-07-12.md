# 领路人、行家组局成功页生产发布记录 - 2026-07-12

## 发布范围

- 行家组局成功详情返回接口文案、真实服务活动信息、参与人数和双方服务确认时间线。
- 领路人组局成功详情增加双方确认成功条件、接口文案和积分奖励展示。
- 保留生产当前其他后端改动，仅发布 `game_handler.go` 中与两个成功页相关的选择性补丁。
- 未修改数据库、Nginx、Redis、PostgreSQL、后台网页或小程序前端容器。

## 发布方式

- 发布前生产源码 SHA-256：`1460673cafbd2f421f895ac3f226ab52b497c41e1143d9f0331e25d845e1c8e1`。
- 生产备份目录：`/opt/zhw-mini/.deploy-backups/success-pages-20260712-111341`。
- 先下载生产当前 `game_handler.go` 做差异审计，排除本地其他尚未上线的通知文案改动。
- 在生产当前源码上应用成功页选择性补丁，仅重建并启动 `go-api`。
- 发布后生产源码 SHA-256：`f3674f852b35c56ffb9df75a4b03a5c48ca7b5c56b976dfdc9b6ebed52c86b95`。

## 验证结果

- Go API Docker 镜像构建成功。
- `go-api` 容器运行正常，重启次数为 `0`。
- 内网 `GET /api/app/health` 返回 HTTP 200。
- 公网 `GET https://api.haowan.net.cn/api/app/health` 返回 HTTP 200。
- 未登录访问行家、领路人成功详情接口均返回 HTTP 401，鉴权保护正常。
- 最近 100 行启动日志未发现 `panic`、`fatal`、段错误或监听错误。

## 回滚位置

- 将 `/opt/zhw-mini/.deploy-backups/success-pages-20260712-111341/game_handler.go` 恢复到 `/opt/zhw-mini/services/go-api/internal/appapi/game_handler.go`。
- 在 `/opt/zhw-mini/deploy` 仅重建并启动 `go-api`。
