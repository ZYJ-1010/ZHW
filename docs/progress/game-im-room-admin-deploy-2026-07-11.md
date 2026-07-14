# 后台组局 IM 房间入口生产发布记录 - 2026-07-11

## 本次修改

- 组局详情“运营明细”新增“创建/同步 IM 房间”入口。
- 新增 `POST /api/admin/games/{gameId}/im-room`，使用 `game:update_status` 权限。
- 接口校验组局存在后幂等调用 `EnsureRoom`；已有房间同步成员，无房间则创建，并记录后台操作日志。

## 验证与发布

- JavaScript 语法检查通过。
- Go `internal/appapi` 只编译检查和 `internal/im` 测试通过；完整 `internal/appapi` 套件仍有本分支既有的无关失败。
- 发布前下载并核对生产目标文件，只在生产当前版本上叠加本次 3 个文件改动，没有覆盖线上其他差异。
- 已备份生产文件到 `/opt/zhw-mini/.deploy-backups/game-im-room-20260711-191603`。
- `go-api` 与 `admin-web` 镜像构建、容器重建成功。
- 内网与公网 `/api/app/health` 均返回成功。
- 生产后台 `main.js` 已包含“创建/同步 IM 房间”；未登录调用新接口返回 HTTP 403，确认接口已注册并受后台鉴权保护。
