# 评价流程生产部署记录 - 2026-07-12

## 发布范围

- 修复多人局提交一条评价后提前关闭整局评价入口的问题。
- 服务管理与玩家管理接口按剩余待评价对象返回 `canReview`，整局评价入口按 `gameId` 加载全部待评价对象。
- 评价目标角色由服务端按局成员角色归一化，支持玩家、行家、领路人和主领路人。
- 新增评价 NPS 0–10 分校验与持久化。
- 本次仅发布 Go API 后端并修改 PostgreSQL；未发布小程序前端、后台网页、Nginx 或 Redis。

## 发布方式

- 先下载生产当前目标源码逐文件审计，只对生产版本叠加本次评价补丁。
- 发布期间发现并发上线的评价积分修复，重新以最新生产源码创建第二版隔离候选，保留 `SubmittedReviewPoints` 与积分发放逻辑。
- 隔离候选镜像：`zhw-go-api-review:20260712-1630-v2`。
- 候选镜像 ID：`sha256:39afb5aceb3d245b3fbf252fe1b495d838bca148af66e6656d536c2cd6dd7d64`。
- 生产镜像 ID：`sha256:f25f8e357a02d6491a2d0bea47956d0c4b2da447f8e165f6b145cd06d225f6f1`。
- 生产备份目录：`/opt/zhw-mini/.deploy-backups/review-20260712-1630`。
- 数据库完整备份：`zhw_mini_before.dump`，SHA-256 为 `4e8bcc51a4e474f12b075b251d29a4a64289252a526e2ce07f7e596f6cf3b1fb`。

## 数据库迁移

- 已执行 `db/migrations/000049_review_nps_score.sql`。
- `reviews.nps_score` 类型为 `smallint`。
- `reviews_nps_score_check` 约束已生效，允许空值或 0–10。

## 生产文件哈希

- `services/go-api/internal/appapi/game_handler.go`：`48625425ae3fc827e71bb49b1ee57780e7482adad0733260412a89ccdd6b93a7`
- `services/go-api/internal/reviews/service.go`：`f0b368baedc85d3ac9443544dfd7d701183ea3557a2df37c535bd5344071c714`
- `services/go-api/internal/reviews/sql_repository.go`：`a113b97ea50ff36f499be2112c6f42d099344852e56ca7414734be15b07264c5`
- `db/migrations/000049_review_nps_score.sql`：`c294c19121eeab4f4a998359f003a0d27896ae00fff67ecd002c53aa20cd52bc`

## 验证结果

- `deploy-go-api-1` 状态为 `running`，重启次数为 `0`。
- PostgreSQL 容器状态为 `healthy`。
- 内网与公网 `GET /api/app/health` 均返回 HTTP 200。
- 公网未登录请求 `GET /api/app/reviews/available` 返回 HTTP 401，路由与鉴权正常。
- 启动日志未发现 `panic`、`fatal` 或数据库字段错误。

## 回滚方式

1. 从 `/opt/zhw-mini/.deploy-backups/review-20260712-1630` 恢复三个 Go 源码文件。
2. 在 `/opt/zhw-mini/deploy` 仅重新构建并启动 `go-api`。
3. `nps_score` 为可空新增列，代码回滚后可保留；如必须完整恢复数据库，使用备份目录中的 `zhw_mini_before.dump`。
