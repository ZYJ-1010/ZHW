# 交付确认页后台生产发布记录 - 2026-07-12

## 发布范围

- `GET /api/app/games/{gameId}/success-detail` 增加 `deliveryMode=free|paid`，供小程序以后台局类型选择交付确认页配置。
- 交付确认相关通知处理路由增加 `mode=free|paid`，并补齐 `service_confirm_remind` 处理入口。
- `POST /api/app/games/{gameId}/service-confirm-items` 与主确认接口统一接收和校验 `confirmItemKeys`。
- 保留生产现有角色权限、行家先确认、玩家后确认、领路人不参与、幂等确认和评价流转逻辑。
- 未修改数据库、Nginx、Redis、PostgreSQL、后台网页或小程序前端发布物。

## 发布方式

- 先读取生产当前 Go API 源码，确认本地三个目标文件还包含其他未发布差异，没有整文件覆盖生产版本。
- 在 `/tmp/zhw-delivery-confirm-20260712-110638/services-go-api` 复制生产当前源码并叠加本次最小补丁。
- 隔离候选镜像 `zhw-go-api-delivery-confirm:20260712-110638` 构建成功，镜像 ID 为 `sha256:ce02c9d7aaee182b71d661ef4ab424844a63676c29487ddf9c58ab51d90bc742`。
- 生产源码备份目录：`/opt/zhw-mini/.deploy-backups/delivery-confirm-20260712-110638`，包含三个原始文件及发布前后 SHA-256 清单。
- 仅重新构建并启动 `go-api`；生产镜像 ID 为 `sha256:4e5000c2b44ebafc56f4a70776f4b20a244be62504ece860936a07626e0f84d3`。

## 验证结果

- 容器 `deploy-go-api-1` 状态为 `running`，重启次数为 `0`。
- 内网 `GET http://127.0.0.1:8080/api/app/health` 返回 HTTP 200。
- 公网 `GET https://api.haowan.net.cn/api/app/health` 返回 HTTP 200，服务状态为 `ok`。
- 公网未登录请求交付详情与交付确认接口均返回 HTTP 401，鉴权保护正常。
- 生产源码已核实包含 `deliveryMode`、`deliveryNotificationRoute` 和第二确认入口的 `ConfirmItemKeys` 校验。
- 启动与验证日志未发现 `panic`、`fatal`、端口占用或重启异常。

## 回滚位置

- 恢复 `/opt/zhw-mini/.deploy-backups/delivery-confirm-20260712-110638/services/go-api/internal/appapi/` 下的三个文件到 `/opt/zhw-mini/services/go-api/internal/appapi/`。
- 在 `/opt/zhw-mini/deploy` 执行 `docker compose -f docker-compose.prod.yml --env-file env.prod build go-api`，然后执行 `docker compose -f docker-compose.prod.yml --env-file env.prod up -d --no-deps go-api`。
- 本次无数据库变更，回滚无需恢复数据库。

## 部署后接口回归

- 项目文档中的本地交付夹具 token 在生产环境均返回 HTTP 401，确认这些 token 未被带入生产。
- 选择生产免费局 `gameId=92005` 的现有成员 `userId=10001`，创建仅用于测试的 15 分钟临时会话；未创建用户、局、成员或确认记录。
- 已登录请求 `GET /api/app/games/92005/success-detail` 返回 HTTP 200：`deliveryMode=free`，`completion.canConfirm=false`，`completion.waitingText=组局已结束，可以进入评价`，`deliveryProof.primaryContactKey=contact_expert`，参与人数为 6。
- 已登录请求 `GET /api/app/notifications` 返回 HTTP 200。
- 请求 `POST /api/app/games/92005/service-confirm-items` 并传入无效 `confirmItemKeys`，返回 HTTP 422 / `code=42200`，在进入业务确认前被拦截。
- 回归后复核：局 `92005` 状态仍为 `pending_review`；临时 `codex_test` 会话数量为 0；go-api 容器仍为 `running`，重启次数为 0。
- 生产数据库当前没有可供只读验证的付费局成员数据，因此未对真实付费局执行接口请求；付费模式继续由同一 `deliveryMode` 分支和本地编译/定向测试覆盖。
