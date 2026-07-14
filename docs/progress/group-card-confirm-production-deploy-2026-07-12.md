# 局卡片头像与玩家自动确认生产部署记录 - 2026-07-12

## 部署范围

- 修复领路人不是局创建者时，玩家已自动通过但行家侧仍显示“待审核”的问题。
- 局列表按 `min(已入局人数, 3)` 返回成员昵称和用户资料中的昵称头像 URL。
- 组局详情的组局者返回用户昵称和昵称头像 URL。
- 修复小程序创建页把 `http://tmp/` 微信临时封面误当成公网 URL 的问题；该前端修复需随小程序版本发布。
- 测试账号种子不再写入 `mock://avatar/...`，改用可访问的测试昵称头像。

## 生产数据修正

- 组局 `92012` 原封面为失效的 `http://tmp/...`，已替换为可访问的默认局封面。
- 用户 `10002`、`10003`、`10004`、`10005` 的昵称头像已由 `mock://` 更新为可访问 URL。
- 邀请和申请数据未改动；玩家邀请原本已经是 `accepted`，对应申请原本已经是 `approved`。

## 发布与回滚

- 生产备份：`/opt/zhw-mini/.deploy-backups/group-card-confirm-20260712-162331`。
- 发布文件：`services/go-api/internal/appapi/game_handler.go`、`game_invite_handler.go`。
- 发布后文件 SHA-256：
  - `game_handler.go`: `db7eb061e4c5514335dd1c75ab0f853570ae576eb892473d289a0ef55612e901`
  - `game_invite_handler.go`: `e6d2b96541fb5a40324a06b948f69b05be5060e073ff677cf7f313fa1e0e84b6`
- 生产镜像：`sha256:7d8d4d71573dca037e3655ddecc88b14f5a7948aebfe69c53d250f541f2c2f11`。
- 回滚时恢复备份目录内两份源码，根据 `DATA.before.txt` 恢复目标数据，再仅重建并启动 `go-api`。

## 验证结果

- 候选发布包的两条分组邀请 HTTP 回归通过。
- 生产容器 `deploy-go-api-1` 状态为 `running`，重启次数为 `0`。
- 公网健康接口返回 HTTP 200。
- 生产组局 `92012` 返回有效 `coverImage`，封面资源 HEAD 返回 HTTP 200。
- 生产组局 `92012` 当前 4 人，接口返回 3 个 `playerAvatars`，每项包含 `userId`、昵称 `name` 和昵称头像 `avatarUrl`。
- 生产邀请 `10` 返回 `canConfirm=true`；玩家 `10005` 返回 `status=approved`、`statusText=已确认`、`confirmed=true`。

