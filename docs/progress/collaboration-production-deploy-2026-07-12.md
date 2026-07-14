# 局内协作后台生产发布记录 - 2026-07-12

## 发布范围

- 局内协作结束请求重复调用时不重复创建通知或局内提醒。
- 首次结束本局时写入 `service_confirm_remind` 局内系统卡片。
- 协作成员列表将 App 端创建者显示为“发起人”。
- IM 服务允许持久化 `service_confirm_remind` 系统消息类型。
- `games/service.go` 的完成请求幂等实现发布前已在线，生产哈希与本地一致，本次未重复覆盖。

## 发布方式

- 生产当前源码快照后生成最小函数段落补丁，未用本地完整 `game_handler.go` 覆盖生产文件。
- 隔离目录候选镜像构建成功：`zhw-go-api-collaboration:20260712-111055`。
- 候选镜像：`sha256:ff66d5fa43a1ffea9844273e848fd66049aee31751beb7f3d912d36c11286ffd`。
- 生产回滚备份：`/opt/zhw-mini/.deploy-backups/collaboration-20260712-111055`。
- 仅重建并启动 `go-api`，未修改 PostgreSQL、Redis、Nginx、后台网页或数据库结构。

## 生产文件哈希

- `services/go-api/internal/appapi/game_handler.go`：`89743181488b3c59ad6276087ea4c462a266c4b99d94786210599e1da9e5c7a9`
- `services/go-api/internal/im/service.go`：`a82818972fa967d4f1bb16485a68b23e5e1587125e38f0df5942fe2b6afb3101`

## 验证结果

- 候选镜像及生产 Go API 镜像均构建成功。
- `go-api` 容器状态为 `running`，重启次数为 `0`。
- 内网与公网 `GET /api/app/health` 均返回成功。
- 未登录访问协作聚合接口返回 HTTP 401。
- 未登录调用完成请求接口返回 HTTP 401。
- 启动日志未发现 `panic` 或 `fatal`。
- 发布前相关 Go 定向测试通过；生产验证未创建测试用户或修改真实组局数据。

## 回滚方式

恢复备份目录中的以下文件到 `/opt/zhw-mini` 对应位置，再仅重建并启动 `go-api`：

- `services/go-api/internal/appapi/game_handler.go`
- `services/go-api/internal/im/service.go`
