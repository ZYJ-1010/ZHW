# IM 样式配套后端生产部署记录 - 2026-07-12

## 发布范围

- 群 IM 房间接口补齐局标题、成员数量、成员详情、当前用户和协作入口。
- IM 会话接口补齐当前用户和局标题。
- IM 消息补齐消息 ID、发送者 ID、角色文案、头像字段、消息类型和文件名兼容字段。
- WebSocket 建连消息与 HTTP 房间接口使用同一套房间数据结构。

## 部署结果

- 仅替换 `services/go-api/internal/appapi/im_handler.go` 和 `services/go-api/internal/appapi/im_ws.go`。
- 新镜像构建成功，镜像 ID：`sha256:e0bbfcff88c0b70d4c1f4e217aae4b67794120e2c2f2665d3e116519bfef295f`。
- 容器 `deploy-go-api-1` 状态为 `running`，重启次数为 `0`。
- 内网和公网 `GET /api/app/health` 均返回 HTTP 200，服务状态为 `ok`。
- 未重建数据库、Redis、Nginx、管理后台或其他服务。

## 发布后文件哈希

- `im_handler.go`: `1528ddd5006838b3426bac75ba09dd8dd143862fd263ce71e357344d6abafc1e`
- `im_ws.go`: `a21e256652b1f15f27c4c4ac9120eb334434d57105329975871e55c9f61e63b9`

## 回滚位置

- `/opt/zhw-mini/.deploy-backups/im-style-20260712-131812`
- 恢复备份中的两个文件到 `/opt/zhw-mini/services/go-api/internal/appapi/` 后，仅重建并启动 `go-api`。

