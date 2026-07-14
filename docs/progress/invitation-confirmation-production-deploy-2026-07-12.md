# 玩家确认与行家审核邀请生产发布记录 - 2026-07-12

## 发布范围

- 配对组局邀请执行“玩家先确认、行家后审核、双方确认后直接成局”。
- 玩家未确认时，行家确认接口返回冲突状态及“请等待玩家先确认组局”。
- 普通邀请继续保留原申请审核流程。
- 消息中心的邀请通知仅提供详情入口，接受和拒绝必须在邀请详情页完成。
- 邀请进度接口返回玩家/行家资料、双方状态、确认禁用原因、倒计时和成功页路由。

## 数据库迁移

- 发布前完成 PostgreSQL 完整备份。
- 执行幂等迁移：
  - `000046_game_invitation_detail_fields.sql`
  - `000047_game_invitation_groups.sql`
  - `000048_game_invitation_parties.sql`
- 已确认 `game_invitations` 包含邀请分组、玩家/行家用户及服务详情字段。

## 发布方式

- 基于生产最新源码生成最小候选版本，避免覆盖并发发布内容。
- 候选镜像：`zhw-go-api:invitation-confirmation-20260712-110936`。
- 生产备份：`/opt/zhw-mini/.deploy-backups/invitation-confirmation-20260712-110936`。
- 仅重建 `go-api`，未重建 PostgreSQL、Redis、Nginx 或后台网页容器。

## 验证结果

- 隔离候选 Docker 镜像构建成功。
- 生产 `go-api` 镜像构建与容器重建成功。
- 容器状态为 `running`，重启次数为 `0`。
- 内网与公网 `GET /api/app/health` 均返回 HTTP 200。
- 未登录访问邀请进度接口返回 HTTP 401。
- 五个发布文件的生产 SHA-256 与候选版本一致。
- 启动日志未发现 `panic`、`fatal` 或启动错误。

## 回滚位置

- 源码与数据库备份均位于 `/opt/zhw-mini/.deploy-backups/invitation-confirmation-20260712-110936`。
- 回滚源码后，仅需重新构建并启动 `go-api`。
