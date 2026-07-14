# 组局满员与开局 IM 流程生产部署记录 - 2026-07-12

## 发布范围

- 玩家与行家组局成功且局人数达到上限时，仅向组局发起人创建一次“组局成功，可以开局”通知。
- 仅满员但玩家、行家尚未全部确认时不发送开局通知；仅组局成功但未满员时也不发送。
- 发起人实际开始组局后，沿用已上线链路向局 IM 写入“局开始了”消息，并向每个局成员创建 `game_started` 消息。
- 本次仅发布 Go API 后端；未上传小程序前端，未修改数据库、Nginx、Redis、PostgreSQL 或后台网页。

## 发布方式

- 发布前连续两次下载生产目标源码并核对哈希，确认期间生产文件未变化。
- 生产与本地差异仅包含本次功能段，未发现其他冲突。
- 候选镜像：`zhw-go-api-group-start-flow:20260712-132214`，镜像 ID `sha256:86ada8be456c095486a9c86fddd7a5b74d800575bf9b6c0a7054415e04a16583`。
- 生产 Go API 镜像 ID：`sha256:a9c3241c1b1a52465db9e43e703089f16ab9e05fc74c2270a4cc14afabb0fa16`。
- 生产回滚备份：`/opt/zhw-mini/.deploy-backups/group-start-flow-20260712-132214`。

## 生产文件哈希

- `services/go-api/internal/appapi/game_handler.go`：`7fd9809a794bf03c456c2585af3081ecb30ef792c1cc573d07cfdd83bdf8d372`
- `services/go-api/internal/appapi/game_invite_handler.go`：`686b1fcfd9921e7fe2cae96ba943284add46d4b1f1653b3b95bb672ff762a2ae`

## 验证结果

- 发布前定向 Go 测试通过：满员通知、开局 IM 消息、成员通知和消息中心 IM 入口均已覆盖。
- `deploy-go-api-1` 状态为 `running`，重启次数为 `0`。
- 最近启动日志未发现 `panic` 或 `fatal`。
- 内网 `GET /api/app/health` 返回 `ok`。
- 公网 `GET https://api.haowan.net.cn/api/app/health` 返回 HTTP 200。
- 未登录访问消息中心返回 HTTP 401，鉴权保护正常。

## 回滚方式

将备份目录中的两个 Go 源码文件恢复到 `/opt/zhw-mini` 对应位置，然后在 `/opt/zhw-mini/deploy` 仅重新构建并启动 `go-api`。
