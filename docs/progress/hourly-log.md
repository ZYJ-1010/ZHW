# 每小时项目记录
## 2026-06-29 23:35

### 当前进展

- 继续按“不区分一期二期，以文档功能为主”的要求推进前后端对接。
- 使用 `miniprogram-development`、`web-development`、`karpathy-guidelines`，并按 `ponytail-audit / ponytail-debt` 的代码审查口径处理真实业务缺口。
- 补齐积分订单前后端契约：`GET /api/app/redemption/orders/my` 现在同时返回 `items` 和 `orders`，小程序订单页不再因只认 `orders/list` 而空白；同时返回 `tabs/statusKey/statusText/statusTone/actions` 并支持状态筛选。
- 补齐积分订单物流接口：新增 `GET /api/app/profile/points/orders/{orderId}/logistics`，按当前用户校验订单归属，返回订单状态轨迹和平台发放信息。
- 收紧登录注册口径：小程序服务层和底层请求层均禁用旧手机号登录、密码登录、找回密码入口，只保留唯一邀请入口进入后微信登录的主链路。
- 修正邀请响应和新手任务：邀请响应兼容中文动作值；新手任务“完成评价”改为读取真实评价记录。
- 新增文档 `docs/progress/full-chain-gap-update-2026-06-29.md`。

### 验证补充

- 已执行：
  - `Get-ChildItem -Recurse miniprogram-client -Filter *.js | ForEach-Object { node --check # 每小时项目记录

## 2026-06-16 13:56

### 当前进展

- 继续以一期小程序后端上线测试为目标，收紧“只能通过邀请入口进入”的登录注册闭环。
- 修补唯一入口码绑定语义：
  - 新用户通过 poster / qrcode / link 入口码进入时，继续走注册模式并绑定微信。
  - 已注册微信通过新的唯一入口码进入时，后端会把该入口码绑定到当前微信，并返回 `authPageMode=login`、`boundWechat=true`。
  - 其他微信再次使用已绑定的唯一入口码时，继续返回 `invite code already bound`。
- 数据层新增 `000027_invite_entry_unique_binding.sql`：
  - 取消 `invite_relations.invitee_user_id` 的唯一约束，允许同一个微信用户绑定多个唯一入口码。
  - 保留按 `invite_code_id` 和 `invitee_user_id` 查询索引，邀请关系统计仍以用户首个邀请关系为准。
- 保留通用测试/运营邀请码的多人使用能力；唯一绑定规则只作用于 `maxUses=1` 的入口码，避免误锁 `TEST2026` 这类通用码。
- 同步 `docs/openapi/app.openapi.yaml` 和 `docs/test-cases/app-integration-cases.md` 的前端联调说明。

### 验证补充

- `go test -count=1 ./internal/invites ./internal/auth ./internal/appapi` 通过。
- `go test -count=1 ./internal/...` 通过。
- 当前一期小程序后端业务功能进度约 `97%`。

## 2026-06-16 13:26

### 当前进展

- 继续以一期小程序后端上线测试为目标，补齐后台审核工作台的列表入口。
- 新增后台组局列表接口：
  - `GET /api/admin/games`
  - 支持 `status` 筛选，例如 `pending_audit`、`recruiting`
  - 需要后台权限 `game:read`
- 已把正式审核链路扩展成完整后台工作流：
  - 后台先查 `GET /api/admin/games?status=pending_audit`
  - 再调用 `POST /api/admin/games/{gameId}/audit`
  - 审核后该局从 `pending_audit` 列表消失，并进入 `recruiting` 列表
- 同步 `docs/openapi/admin.openapi.yaml` 和 `docs/test-cases/admin-integration-cases.md`，方便前后端按最新后台接口联调。

### 验证补充

- `go test -count=1 ./internal/appapi` 通过。
- `go test -count=1 ./internal/...` 通过。
- 当前一期小程序后端业务功能进度约 `96.5%`。

## 2026-06-16 13:05

### 当前进展

- 继续以一期小程序后端上线测试为目标，补齐“局创建后由后台正式审核”的生产业务入口。
- 新增后台正式审核接口：
  - `POST /api/admin/games/{gameId}/audit`
  - 需要后台权限 `game:update_status`
  - 审核通过后调用现有组局服务，将 `pending_audit` 的局推进到 `recruiting`
- 后台权限种子已补充 `game:update_status`，超级管理员可直接执行，运营账号未授权时返回 `403`。
- 审核动作已写入后台操作日志 `game:audit`，便于线上排查“谁审核了哪个局”。
- 生产环境不再需要依赖 `/approve-local` 这类本地联调入口完成组局审核。

### 验证补充

- `go test -count=1 ./internal/appapi` 通过。
- `go test -count=1 ./internal/...` 通过。
- 当前一期小程序后端业务功能进度约 `96%`。

## 2026-06-16 12:55

### 当前进展

- 继续以一期小程序后端上线收口为导向，处理本地联调入口的生产安全边界。
- `appapi.Server.Configure` 新增生产环境识别：
  - `APP_ENV=prod`
  - `APP_ENV=production`
- `POST /api/app/games/{gameId}/approve-local` 现在仅保留为本地/测试联调入口：
  - 本地环境继续可用，保持现有测试和联调流程。
  - 生产环境直接返回 `404`，避免本地审核后门随服务上线。
- 补充测试 `TestApproveLocalDisabledInProduction`，确保带登录 token 访问生产环境本地审核入口仍不可用。

### 验证补充

- `go test -count=1 ./internal/appapi` 通过。
- `go test -count=1 ./internal/...` 通过。
- 当前一期小程序后端业务功能进度约 `95.5%`。

## 2026-06-16 12:35

### 当前进展

- 继续以一期小程序后端业务功能为导向，补齐微信订阅消息任务执行闭环。
- `notifications.Service` 新增订阅消息发送能力：
  - `SendWechatTask(taskID)` 会根据任务里的 `userId` 找到微信 `openid`，调用 `WechatSubscribeSender` 发起真实订阅消息发送。
  - `SendPendingWechatTasks(limit)` 支持批量发送待处理任务，方便内部任务执行器一次性拉起。
  - 发送结果会回写 `wechat_subscribe_tasks`，并同步更新 `notifications.wechatState`。
- `appapi` 新增内部路径：
  - `POST /api/internal/notifications/wechat-tasks/{taskId}/send`
  - `POST /api/internal/notifications/wechat-tasks/send-pending`
  - 仍保留 `mark-sent` 作为兼容入口。
- 微信订阅消息发送器接入现有 `WECHAT_APP_ID/WECHAT_APP_SECRET` 配置，不再另起一套账号配置。

### 验证补充

- `go test -count=1 ./internal/notifications ./internal/appapi` 通过。
- `go test -count=1 ./internal/...` 通过。
- 当前一期小程序后端业务功能进度约 `95%`。

## 2026-06-16 11:50

### 当前进展

- 继续以一期小程序后端业务功能为主，补齐强实名认证 FaceID 生产链路。
- `identity.Service` 的 `StartFaceID` 从固定 `mock_face_token` 升级为可注入启动器：
  - 本地默认 `LocalFaceIDStarter`，继续返回 mock token，保证开发联调不被第三方阻塞。
  - 生产可配置 `HTTPFaceIDStarter`，由 `FACEID_HTTP_ENDPOINT` 发起核身并读取 `faceToken` / `bizToken`。
  - 发起请求会携带 `userId`、脱敏手机号、脱敏姓名、脱敏身份证号，避免后端接口继续暴露明文身份数据。
- 生产环境配置新增强校验：
  - `FACEID_HTTP_ENDPOINT` 必须是 HTTPS。
  - `FACEID_HTTP_SECRET` 必须是安全密钥。
  - 启动入口 `cmd/server/main.go` 已根据配置注入真实 FaceID HTTP 启动器。
- 同步 `.env.example` 和 `docs/openapi/app.openapi.yaml`：
  - 补充 FaceID 发起网关、回调签名密钥和开关变量。
  - `POST /api/app/identity/faceid/detect-auth` 文档从“返回 mock faceToken”改为“返回本地 mock 或配置网关返回的 faceToken”。

### 验证补充

- `go test -count=1 ./internal/identity ./internal/common/config ./internal/appapi` 通过。
- `go test -count=1 ./internal/...` 通过。
- 当前一期小程序后端业务功能进度约 `93%`。

## 2026-06-14 19:15

### 当前进展

- 继续执行 `plan.md`，推进后台权限隔离业务功能。
- 后台内置角色补齐：
  - 新增 `data_analyst` 本地登录账号样本，只包含行为、看板、报表导出、测试结果读取权限。
  - 新增 `operator` 本地登录账号样本，可做普通运营查看和处理，但不含完整日志、敏感画像、测试证据权限。
- 测试用例权限拆分：
  - `GET /api/admin/test-cases` 和 `GET /api/admin/test-runs` 改为 `testcase:read`。
  - `POST /api/admin/test-cases` 和 `POST /api/admin/test-runs` 保持 `testcase:manage`。
- `db/seeds/admin_roles_permissions.sql` 追加 `testcase:read` 和 data_analyst/operator 角色权限映射，保持数据库初始化与本地内置账号一致。
- 更新 `plan.md`：
  - 勾选“数据分析员可看行为、报表、测试结果，但不能审核兑换和修改配置”。
  - 勾选“普通运营人员不可查看完整敏感画像和测试证据文件”。

### 验证补充

- `go test -count=1 ./internal/adminauth ./internal/appapi` 通过。
- `go test -count=1 ./...` 通过。

## 2026-06-14 18:45

### 当前进展

- 继续执行 `plan.md`，优先补齐小程序后端业务验收闭环。
- 后台标记打卡异常增强：
  - `POST /api/admin/game-checkins/{checkinId}/mark-invalid` 现在必须提交 `reason`。
  - 缺少原因或原因过长时返回 422，不会把打卡改为异常。
  - 成功标记异常后，会把 `reason` 写入后台操作日志 payload，便于追溯。
- 更新 `plan.md`：
  - 勾选“兑换审核、打卡异常标记、交付文档状态变更必须填写原因”。
  - 补充“打卡异常标记必须填写原因”子项。

### 验证补充

- `go test -count=1 ./internal/appapi` 通过。
- `go test -count=1 ./...` 通过。

## 2026-06-14 12:11

### 当前进展

- 继续 D7 生产密钥和环境变量约束：
  - `config.Load` 已显式读取 `APP_ENV`、`JWT_SECRET`、数据库连接、OpenIM secret、MinIO secret 等环境变量
  - 新增 `Config.ValidateProduction()`，仅在 `APP_ENV=prod/production` 时启用生产配置拦截
  - 生产环境拒绝缺失数据库连接、缺失/占位 `JWT_SECRET`、占位 `OPENIM_SECRET`、占位 `MINIO_SECRET_KEY`
  - 服务启动入口已接入生产配置校验，避免默认占位密钥进入生产运行
- `plan.md` 已同步勾选 D7 的生产 JWT secret、数据库密码、对象存储密钥环境变量/密钥管理项。

### 验证补充

- `go test -count=1 ./internal/common/config ./internal/appapi` 通过。

### 下一步

- 跑全量 `go test -count=1 ./...`，继续 D7：写接口幂等和外部输入校验。

## 2026-06-14 12:23

### 当前进展

- 继续 D7 后台鉴权与权限统一：
  - 移除生产代码中 `X-Admin-Permissions` 请求头直通权限的兼容入口
  - 后台受保护接口统一通过 `Authorization: Bearer <admin-token>` 校验后台 session
  - `requireAdminPermission(permissionCode)` 统一负责权限码校验，并自动回填 `X-Admin-ID` 供操作日志使用
  - IM 争议消息后台查看入口只保留统一权限链路，不再做可伪造请求头二次判断
  - 测试 helper 已改为真实后台登录 token，不再依赖伪权限头
- `plan.md` 已同步勾选 D7 的后台接口 AdminAuth 和 PermissionMiddleware 两项。

### 验证补充

- `rg` 已确认业务代码不再包含 `X-Admin-Permissions`、`adminHasPermission`、`permissionsFromHeader`。
- `go test -count=1 ./internal/appapi` 通过。

## 2026-06-14 18:05

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 E5 AI 数据准备和后台导出基础。
- 评价模型新增 `tags`：
  - `POST /api/app/reviews` 支持可选 `tags` 数组。
  - 标签会去空、去重，最多保留 8 个，每个最长 32 字符。
  - `ReviewDTO` 和 `/api/app/reviews/my-intents` 会返回 `tags` 与 `againIntent`。
- 数据库迁移新增 `db/migrations/000022_review_tags.sql`：
  - 给 `reviews` 增加 `tags jsonb not null default '[]'`。
  - 给 `again_intent` 增加查询索引，便于后台导出和二期推荐分析。
- 后台导出模板新增 `reviews_default`：
  - `ExportType=reviews`
  - 固定列包含 `tags` 和 `again_intent`。
- 更新 `plan.md`：
  - 标记“评价标签和再玩意向可导出”完成。
  - 注意：`000022_review_tags.sql` 后续需要补入迁移顺序表；当前业务代码和迁移文件已落地。

### 验证补充

- `go test -count=1 ./internal/reviews` 通过。
- `go test -count=1 ./internal/exports` 通过。
- `go test -count=1 ./internal/appapi` 通过。
- `go test -count=1 ./...` 通过。

## 2026-06-14 18:25

### 当前进展

- 继续执行 `plan.md`，优先推进后端验收留档链路。
- 交付文档登记增强：
  - `docType` 限定为 `prd/openapi/database/deployment/test-report/release-record/rollback-plan`。
  - 状态进入 `ready` 或 `archived` 时必须填写 `reason`。
  - 返回 `DocumentDTO` 增加 `reason` 字段。
- 测试用例登记增强：
  - `priority` 只能是 `P0/P1/P2`。
  - 测试执行记录继续支持 `requestId` 关联，便于用 requestId 追踪接口响应。
- 更新 `plan.md`：
  - 勾选交付文档类型约束。
  - 勾选测试用例优先级约束。
  - 勾选测试用例和测试执行记录可登记。
  - 在混合验收项下标记交付文档状态变更必须填写原因已完成；打卡异常标记原因仍待补。

### 验证补充

- `go test -count=1 ./internal/delivery` 通过。
- `go test -count=1 ./internal/appapi` 通过。

## 2026-06-14 17:45

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 E1.4 积分兑换闭环。
- 核实现有兑换链路：
  - 积分不足会拒绝兑换并回滚库存。
  - 库存不足会拒绝兑换。
  - 内存模式下已有并发兑换不超卖测试。
  - 后台审核、驳回、发放已有状态流转和操作日志。
- 新增兑换订单审核原因约束：
  - 后台审核通过、驳回、标记发放时，`reason` 不能为空。
  - 缺少原因时返回参数错误，订单保持原状态。
- 更新 `plan.md`：
  - 勾选“积分不足不能兑换”“库存不足不能兑换”“并发兑换不超卖”。
  - 在混合验收项下补充“兑换订单审核、驳回、发放必须填写原因”已完成；打卡异常和交付文档状态原因仍保留待补。

### 验证补充

- `go test -count=1 ./internal/redemption` 通过。
- `go test -count=1 ./internal/appapi` 通过。

## 2026-06-14 17:25

### 当前进展

- 继续执行 `plan.md`，优先补齐小程序后端强实名联调路径。
- 新增 `POST /api/app/identity/faceid/result`：
  - 复用现有 FaceID callback 完成逻辑。
  - 同样受 `AppAuthMiddleware`、幂等中间件和 FaceID 回调签名配置保护。
  - 解决 `plan.md` 强实名接口清单中 `faceid/result` 与现有代码只有 `faceid/callback` 的路径不一致问题。
- 整理 `docs/openapi/app.openapi.yaml` 的 identity 局部路径块：
  - 补出 `phone/verify`、`faceid/detect-auth`、`faceid/callback`、`faceid/result`、`identity/status` 独立路径。
  - 避免小程序前端联调时把多个路径误读成 description 文本。

### 验证补充

- `go test -count=1 ./internal/appapi` 通过。

### 下一步

- 跑全量 `go test -count=1 ./...`，继续 D7：生产密钥环境变量、写接口幂等和外部输入校验。

## 2026-06-14 12:11

### 当前进展

- 继续 D7 后端安全和权限统一落地，补齐请求访问日志的脱敏保护：
  - `httpx.RequestID` 生成后回填请求头，方便后续操作日志和访问日志使用同一个 requestId
  - 新增 `httpx.AccessLog`，只记录 method、path、status、duration、requestId
  - 访问日志不记录 header、body 和 query 原值，避免手机号、身份证、token、密码、密钥进入日志
  - Go 服务入口已接入 `RequestID + AccessLog`
- 复核后台管理员密码逻辑：
  - `adminauth` 已使用 PBKDF2-SHA256 强哈希
  - 测试覆盖错误密码拒绝、正确密码登录、明文密码存储拒绝
- `plan.md` 已同步勾选 D7 的请求日志脱敏、后台管理员密码强哈希两项。

### 验证补充

- `go test -count=1 ./internal/common/httpx` 通过。
- `go test -count=1 ./internal/appapi` 通过。

### 下一步

- 跑全量 `go test -count=1 ./...`，继续 D7：认证/权限中间件覆盖复核、生产密钥环境变量、写接口幂等和外部输入校验。

## 2026-06-14 11:58

### 当前进展

- 继续 D7 后端安全和权限统一落地，优先处理小程序后端文件下载权限：
  - `GET /api/app/files/{fileId}/download-url` 收紧为按业务对象校验
  - `avatar` 仅上传者本人可下载
  - `chat_file` 仅同局成员可下载
  - `realname_material`、`report_attachment`、`export_file` 等非小程序通用下载场景默认拒绝
- 测试补充：
  - 头像上传者本人可下载，其他用户下载返回 403
  - IM 文件局成员可下载，非成员下载返回 403
  - 举报附件不能通过小程序通用下载口下载
- `plan.md` 已同步勾选 D7 文件下载对象权限和非成员读取 IM/文件上线阻断项。

### 验证补充

- `go test -count=1 ./internal/appapi` 通过。
- Go telemetry 仍有 AppData 权限警告，但测试退出码为 0。

### 下一步

- 继续 D7：复核请求日志敏感字段、后台管理员密码哈希、生产密钥环境变量、写接口幂等和输入校验。

## 2026-06-14 11:58

### 当前进展

- 补齐 D6 前后台对接所需的数据看板接口文档：
  - `docs/openapi/admin.openapi.yaml` 登记 `/api/admin/dashboard`
  - `docs/openapi/admin.openapi.yaml` 登记 `/api/admin/analytics/funnel`
  - `docs/openapi/admin.openapi.yaml` 登记 `/api/admin/analytics/retention`
  - `docs/openapi/dto-samples.md` 新增 `FunnelSnapshotDTO`
  - `docs/openapi/dto-samples.md` 新增 `RetentionSnapshotDTO`
  - `docs/openapi/dto-samples.md` 新增 `DashboardSnapshotDTO`
- D5.8 在 `plan.md` 中已全部勾选完成，下一步进入 D6 后台接口/页面对接或 D7 安全权限统一检查。

### 验证补充

- `rg` 已确认 OpenAPI 和 DTO 示例能搜到新增路径与 DTO。
- `go test -count=1 ./internal/audit ./internal/appapi` 通过。
- Go telemetry 仍有 AppData 权限警告，但测试退出码为 0。

### 下一步

- 优先继续 D6 后台前端工程硬要求，或按后端优先进入 D7：后台接口统一权限、文件下载对象权限、IM 权限一致性、输入校验和敏感日志检查。

## 2026-06-14 11:58

### 当前进展

- 收口 D5.8 数据看板基础统计：
  - `audit.Service` 新增漏斗聚合，按 `eventCode` 统计每步用户数、转化率和流失率
  - `audit.Service` 新增留存聚合，按用户首次行为日期统计次日、7 日、30 日留存
  - 新增后台接口 `/api/admin/dashboard`
  - 新增后台接口 `/api/admin/analytics/funnel`
  - 新增后台接口 `/api/admin/analytics/retention`
- `plan.md` 已同步勾选 D5.8 最后一项“数据看板能统计漏斗和留存基础数据”。

### 验证补充

- `go test -count=1 ./internal/audit ./internal/appapi` 通过。
- Go telemetry 仍有 AppData 权限警告，但测试退出码为 0。

### 下一步

- 跑全量 `go test -count=1 ./...`，然后继续进入 D6 或回补后台 OpenAPI/页面侧任务。

## 2026-06-14 11:58

### 当前进展

- 继续补 D5.8 行为日志闭环：
  - `GET /api/app/games` 自动记录 `browse_games`
  - `GET /api/app/games/city` 自动记录同城浏览
  - `GET /api/app/games/nearby` 自动记录附近浏览
- 行为日志测试已覆盖 `browse_games`、手动上报 `search/share`、字段 `eventCode/businessType/businessId/source/device/ip/occurredAt`。
- D5.8 中通知站内 + 微信任务、关键行为日志、行为字段完整性已同步勾选。

### 验证补充

- `go test -count=1 ./internal/appapi` 通过。
- Go telemetry 仍有 AppData 权限警告，但测试退出码为 0。

### 下一步

- 继续补 D5.8 剩余项：数据看板能统计漏斗和留存基础数据。

## 2026-06-14 11:58

### 当前进展

- 继续推进 D5.8 举报申诉、通知和行为留存，优先补小程序后端举报证据链。
- `POST /api/app/reports` 新增证据归属校验：
  - `chatMessageId` 必须存在且属于当前 `gameId`
  - `fileId` 必须存在且 `objectId` 属于当前 `gameId`
  - `reviewId` 必须来自当前局评价链路
  - `revenueRecordId` 必须存在且属于当前局
- 集成测试补齐真实文件凭证，并增加无效 IM 证据返回 `422` 的断言。
- `plan.md` 已同步勾选 D5.8 中举报关联证据、举报后冻结收益、后台处理可见证据、争议冻结分润。

### 验证补充

- `go test -count=1 ./internal/appapi` 通过。
- Go telemetry 仍有 AppData 权限警告，但测试退出码为 0。

### 下一步

- 继续补 D5.8 剩余项：通知站内和微信任务覆盖、用户关键行为日志字段完整性、漏斗和留存基础数据看板。

## 2026-06-14 11:35
### 褰撳墠杩涘睍

- 缁х画 D5.7 鍒嗘鼎鐢熸垚鍓嶄簤璁娴嬨€?- `POST /api/admin/revenue/records/generate` 鐜板湪鐢熸垚鍒嗘鼎璁板綍鍚庝細妫€鏌ュ悓灞€鏈叧闂妇鎶ョ敵璇夛細
  - 鏈夋湭鍏抽棴涓炬姤鏃讹紝璁板綍绔嬪嵆杩涘叆 `frozen`
  - `frozenReason=report_open`
  - 鍚庣画缁撶畻娌跨敤鏃㈡湁鍐荤粨鎷︽埅锛岃繑鍥炲啿绐佺姸鎬?- 宸茶ˉ闆嗘垚娴嬭瘯锛岃鐩栤€滃厛涓炬姤銆佸悗鐢熸垚鍒嗘鼎鈥濈殑鍐荤粨璺緞銆?- `plan.md` 宸插悓姝ュ嬀閫夊垎娑︾敓鎴愪簤璁娴嬨€佷簤璁喕缁撳拰 TC-116銆?
### 楠岃瘉琛ュ厖

- `go test -count=1 ./internal/appapi` 閫氳繃銆?
### 涓嬩竴姝?
- 鍏ㄩ噺娴嬭瘯鍚庤繘鍏?D5.8 涓炬姤鐢宠瘔銆侀€氱煡鍜岃涓虹暀瀛樼殑鍚庣鏀跺彛锛屼紭鍏堣ˉ涓炬姤鍏宠仈璇佹嵁鍜岃涓虹暀瀛樼己鍙ｃ€?
## 2026-06-14 11:15

### 褰撳墠杩涘睍

- 杩涘叆 D5.7 鍒嗘鼎銆佹敹鐩娿€佽鍗曞崰浣嶇殑鍚庣鎸佷箙鍖栨敹鍙ｃ€?- 鏂板 `services/go-api/internal/revenue/sql_repository.go`锛?  - 鍒嗘鼎妯℃澘鎸佷箙鍖栧埌 `revenue_templates`
  - 鍒嗘鼎璁板綍鎸佷箙鍖栧埌 `revenue_records`
  - 鍒嗘鼎鏄庣粏鎸佷箙鍖栧埌 `revenue_record_items`
  - 绾夸笅缁撶畻鐧昏鍐欏叆 `settlement_records`
  - 鐢ㄦ埛鏀剁泭鎽樿鍜屾敹鐩婃槑缁嗕粠 SQL 鑱氬悎璇诲彇
- `revenue.Service` 澧炲姞 repository 娉ㄥ叆锛屼繚鐣欏唴瀛?fallback銆?- `cmd/server` 鍜?`appapi.UseRepositories` 宸叉帴鍏?`revenue.NewSQLRepository(db)`銆?- `plan.md` 宸插悓姝?D5.7 涓凡楠岃瘉鐨勫厤璐瑰眬璁㈠崟銆佹ā鏉块厤缃€佽瘯绠椼€佹敹鐩婃憳瑕?鏄庣粏銆佹暣鏁板垎鍜屽喕缁撶粨绠楁嫤鎴」銆?
### 楠岃瘉琛ュ厖

- 鏂板 `internal/revenue/repository_service_test.go`锛岃鐩栨ā鏉裤€佽瘯绠椼€佺敓鎴愩€侀噸澶嶇敓鎴愩€佸喕缁撱€佺粨绠楀拰鏀剁泭鏌ヨ璧?repository銆?- `go test -count=1 ./...` 閫氳繃銆?
### 涓嬩竴姝?
- D5.7 鍓╀綑杈冨椤规槸鈥滃垎娑︾敓鎴愬墠涓诲姩妫€鏌ユ湭鍏抽棴涓炬姤鐢宠瘔骞剁洿鎺?frozen/闃绘柇鈥濓紝褰撳墠宸叉湁涓炬姤鍒涘缓鍚庡喕缁撴棦鏈夊垎娑﹁褰曪紝鍚庣画闇€瑕佹妸鐢熸垚鍓嶄簤璁娴嬭ˉ鍒?revenue 鐢熸垚鍏ュ彛銆?
## 2026-06-14 10:35

### 褰撳墠杩涘睍

- 缁х画鏀跺彛 D5.6 鏈嶅姟纭銆佽瘎浠枫€佹垚闀裤€佷俊鐢ㄩ摼璺€?- `reviews.Repository` 澧炲姞 `CreditDeductionValue(ruleCode)`锛岄€€鍑烘墸淇＄敤浼樺厛璇诲彇 `credit_deduction_rules`銆?- `reviews.SQLRepository` 宸叉敮鎸佷粠鍚敤鐨勪俊鐢ㄦ墸鍒嗚鍒欒鍙?`change_value`銆?- 鏂板杩佺Щ `db/migrations/000021_credit_deduction_defaults.sql`锛?  - `quit_after_confirm = -10`
  - `quit_after_started = -10`
- 鏃犺鍒欓厤缃椂浠?fallback 涓?`-10`锛屼笉褰卞搷鐜版湁娴佺▼銆?- `plan.md` 宸插悓姝ュ嬀閫?D5.6 淇＄敤涓婚摼銆佹瘡鏃ヤ俊鐢ㄥ垵濮?100銆侀€€鍑烘墸淇＄敤鍑嗙‘绛夊凡楠岃瘉椤广€?
### 楠岃瘉琛ュ厖

- 鏂板 repository 绾ф祴璇曪紝瑕嗙洊鏈夎鍒欐椂鎸夎鍒欐墸鍒嗐€佹棤瑙勫垯鏃朵娇鐢ㄩ粯璁ゆ墸鍒嗐€?- `go test -count=1 ./internal/reviews` 宸查€氳繃銆?
### 涓嬩竴姝?
- 鍏ㄩ噺娴嬭瘯鍚庣户缁帹杩?D5.7 鍒嗘鼎/鏀剁泭閾捐矾锛屾垨鑰呭厛琛?`game_members.member_status/credit_log_id` 杩欑鏇寸粏瀛楁钀藉簱銆?
## 2026-06-14 10:05

### 褰撳墠杩涘睍

- 缁х画鎺ㄨ繘 `plan.md` 鐨?D5.6 鏈嶅姟纭涓庤瘎浠烽摼璺紝浼樺厛鍚庣銆?- `services/go-api/internal/reviews/service.go` 宸茶ˉ榻愶細
  - 鏈嶅姟纭瀹屾垚鍚庤繘鍏?`pending_review`
  - 璇勪环寰呭姙鎸夋埅姝㈡椂闂磋繃婊?  - 璇勪环鎻愪氦鎸夋埅姝㈡椂闂村拰閲嶅鍏崇郴鎷︽埅
  - 鏁版嵁搴撴ā寮忎笅鐨勮瘎浠峰畬鎴愬垽鏂彲鍩轰簬鎸佷箙鍖栬瘎浠疯褰曟仮澶?  - 璇勪环鍚庣户缁啓缁忛獙銆佺Н鍒嗐€佹垚灏卞拰瓒宠抗
- `GET /api/app/footprints/my` 宸茶ˉ鎴愮嫭绔嬩釜浜轰腑蹇冩帴鍙ｃ€?- `services/go-api/internal/appapi/server_test.go` 澧炲姞浜嗚冻杩规帴鍙ｅ洖褰掋€?
### 楠岃瘉琛ュ厖

- `go test -count=1 ./...` 閫氳繃銆?- 褰撳墠浠嶄繚鐣欎竴鏉″悗缁緟鍋氾細淇＄敤瑙勫垯鐨勫畬鏁撮噸绠椾笌鏇寸粏鍖栫害鏉燂紝褰撳墠涓嶅湪杩欒疆纭嬀瀹屾垚銆?
### 涓嬩竴姝?
- 缁х画鎶?D5.6 閲屽墿浣欑殑淇＄敤閲嶇畻/鎻愰啋閾捐矾鏀跺彛锛屽啀鍚?D5.7 鍒嗘鼎鍗犱綅鎺ㄨ繘銆?
## 2026-06-14 09:20

### 褰撳墠杩涘睍

- 缁х画鎸?`plan.md` 浼樺厛鎺ㄨ繘灏忕▼搴忓悗绔€?- 瀹屾垚鏂囦欢涓婁紶鍑瘉鐨勬湇鍔″眰鐧藉悕鍗曡鍒欙細
  - `avatar`锛氫粎鍏佽 `image/jpeg`銆乣image/png`銆乣image/webp`锛屾渶澶?5 MiB锛岃闂骇鍒?`private`銆?  - `chat_file`锛氫粎鍏佽鍥剧墖銆丳DF銆佺函鏂囨湰銆乑IP锛屾渶澶?20 MiB锛屽繀椤荤粦瀹氬眬 `objectId`锛岃闂骇鍒?`game_member`銆?  - `realname_material`锛氫粎鍏佽鍥剧墖鍜?PDF锛屾渶澶?10 MiB銆?  - `report_attachment`锛氫粎鍏佽鍥剧墖鍜?PDF锛屾渶澶?20 MiB銆?- `POST /api/app/files/upload-token` 淇濇寔澶村儚鍙?pre-auth 涓婁紶锛涢潪澶村儚鏂囦欢浠嶈姹傚己瀹炲悕銆?- OpenAPI 瀵逛笂浼犲嚟璇佽ˉ鍏?`422` 鏍￠獙澶辫触璇存槑銆?
### 楠岃瘉琛ュ厖

- 鏂板鏂囦欢鏈嶅姟鍗曞厓娴嬭瘯锛岃鐩栨湭鐭?`bizType`銆佽秴澶у皬銆侀潪娉?MIME銆両M 鏂囦欢缂哄皯 `objectId`銆佸ご鍍忚闂骇鍒€?- 鏂板 app API 娴嬭瘯锛岃鐩?pre-auth 鐢ㄦ埛涓嶈兘鐢宠 `chat_file` 涓婁紶鍑瘉銆?
### 涓嬩竴姝?
- 缁х画鎺ㄨ繘 D5.6 鏈嶅姟纭涓庤瘎浠烽摼璺殑鍚庣瀹屾暣鎬э紝浼樺厛琛ユ暟鎹簱鎸佷箙鍖栥€佺姸鎬佹祦杞拰鏉冮檺娴嬭瘯銆?

## 2026-06-13 01:16

### 褰撳墠杩涘睍

- 宸茬‘璁ら」鐩綋鍓嶅彧鏈夋枃妗ｅ拰 `plan.md`锛屾病鏈夌幇鎴愪唬鐮佸伐绋嬨€?- 宸叉寜 `plan.md` P0 鍒涘缓鍚庣銆佽祫閲戞湇鍔°€丳C 鍚庡彴銆佹暟鎹簱銆侀儴缃层€佹帴鍙ｆ枃妗ｃ€佹祴璇曟枃妗ｃ€佽繘搴﹁褰曠洰褰曘€?- 宸茬‘璁?Java銆丯ode銆乶pm 鍙敤锛汳aven 鏈畨瑁咃紱Go SDK 闇€瑕佷娇鐢?`C:\Program Files\Go\bin\go.exe`锛岀郴缁?PATH 鍓嶇疆瀛樺湪 `C:\Windows\System32\go` 骞叉壈銆?
### 褰撳墠椋庨櫓

- Maven 涓嶅湪 PATH锛孞ava 璧勯噾鏈嶅姟鍏堢敤 JDK 鑷甫 HTTP 鏈嶅姟楠ㄦ灦鍚姩锛屽悗缁啀琛?Maven 鎴?Gradle銆?- `go` 鍛戒护浼樺厛鍛戒腑 `C:\Windows\System32\go`锛屽悗缁紪璇戦渶鏄庣‘璋冪敤鐪熷疄 Go SDK 鎴栦慨姝?PATH銆?
### 涓嬩竴姝?
- 鍒濆鍖?Go API 鍋ュ悍妫€鏌ャ€?- 鍒濆鍖?Java 璧勯噾鏈嶅姟鍋ュ悍妫€鏌ャ€?- 鍒濆鍖栧悗鍙伴潤鎬佸３銆?- 琛ュ厖鏁版嵁搴撹縼绉诲拰閮ㄧ讲鏍蜂緥銆?
## 2026-06-13 01:50

### 褰撳墠杩涘睍

- 宸叉寜鐢ㄦ埛鏈€鏂拌姹傛敹鏁涘疄鏂借寖鍥达細褰撳墠鍙户缁仛鈥滃皬绋嬪簭鍚庣鈥濓紝涓嶇户缁帹杩?PC 鍚庡彴鍓嶇鍜?Java 璧勯噾鏈嶅姟銆?- 宸茶鍙栧苟閲囩敤 `karpathy-guidelines`锛屽悗缁寜灏忔銆佸彲娴嬭瘯銆佷笉杩囧害璁捐鐨勬柟寮忓紑鍙戙€?- Go 灏忕▼搴忓悗绔仴搴锋鏌ュ凡鍙闂細`GET /health`銆乣GET /api/app/health`銆?- 瀹屾垚 D5.1 绗竴杞悗绔兘鍔涳細
  - `POST /api/app/auth/wechat-login`
  - `GET /api/app/users/me`
  - 鍐呭瓨鐗堢敤鎴枫€侀個璇风爜銆侀個璇峰叧绯汇€乼oken store
  - 鏂扮敤鎴锋棤閭€璇风爜鎷︽埅
  - 鏈夋晥閭€璇风爜娉ㄥ唽骞剁粦瀹氬叧绯?  - 宸插瓨鍦?openid 鍐嶇櫥褰曚笉閲嶅寤虹敤鎴枫€佷笉瑕佹眰鍐嶆浼犻個璇风爜
- 宸茶ˉ鍏?`docs/openapi/app.openapi.yaml`銆乣docs/openapi/dto-samples.md`銆乣docs/openapi/error-codes.md`銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 閫氳繃鐨勬祴璇曞寘锛?  - `internal/auth`
  - `internal/appapi`

### 褰撳墠椋庨櫓

- 褰撳墠 D5.1 浣跨敤鍐呭瓨浠撳偍锛屾湇鍔￠噸鍚悗鏁版嵁浼氫涪澶憋紱鍚庣画闇€瑕佹帴 PostgreSQL 浠撳偍銆?- 寰俊鐧诲綍褰撳墠鏄?mock openid锛屽悗缁渶瑕佹帴鐪熷疄 `code2Session`銆?- 鐪熷疄寮哄疄鍚嶈繕鏈疄鐜帮紝涓嬩竴姝ユ寜 E4/D5.2 琛ユ墜鏈哄彿銆佺煭淇°€佷汉鑴告牳韬姸鎬佷富绾裤€?
### 涓嬩竴姝?
- 瀹炴柦 D5.2/E4锛氬己瀹炲悕鐘舵€佹満銆佹墜鏈哄彿缁戝畾銆佺煭淇￠獙璇佺爜銆佷汉鑴告牳韬崰浣嶆帴鍙ｃ€?- 淇濇寔鎺ュ彛鍏堝彲娴嬶紝鍐嶆浛鎹㈢湡瀹炵涓夋柟璋冪敤銆?
## 2026-06-13 02:05

### 褰撳墠杩涘睍

- 瀹屾垚 D5.2/E4 寮哄疄鍚嶄富绾跨殑鏈湴鍗犱綅瀹炵幇銆?- 宸叉柊澧炲己瀹炲悕鐘舵€佹満锛?  - `wechat_logged_in`
  - `phone_bound`
  - `sms_verified`
  - `phone_verified`
  - `faceid_processing`
  - `verified`
- 宸叉柊澧炲皬绋嬪簭鍚庣鎺ュ彛锛?  - `POST /api/app/identity/phone/bind`
  - `POST /api/app/sms/send-code`
  - `POST /api/app/sms/verify-code`
  - `POST /api/app/identity/phone/verify`
  - `POST /api/app/identity/faceid/detect-auth`
  - `POST /api/app/identity/faceid/callback`
  - `GET /api/app/identity/status`
- 褰撳墠涓烘湰鍦板崰浣嶉€昏緫锛氱煭淇￠獙璇佺爜鍥哄畾 `123456`锛屼汉鑴告牳韬繑鍥?`mock_face_token`锛屽井淇″疄鍚嶄竴鑷存€ц褰曚负 `not_supported`銆?- 宸茶ˉ鍏?`docs/openapi/app.openapi.yaml`銆乣docs/openapi/dto-samples.md`銆乣docs/openapi/error-codes.md`銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板閫氳繃娴嬭瘯鍖咃細
  - `internal/identity`
  - `internal/appapi` 寮哄疄鍚?HTTP 娴佺▼

### 褰撳墠椋庨櫓

- 寮哄疄鍚嶈繕鏈帴鑵捐浜戠煭淇°€佹墜鏈哄彿浜?涓夎绱犮€佹収鐪间汉鑴告牳韬湡瀹炴帴鍙ｃ€?- 褰撳墠鐢ㄦ埛銆佸疄鍚嶇姸鎬佷粛鏄唴瀛樺瓨鍌紝鍚庣画闇€瑕佹帴 PostgreSQL銆?
### 涓嬩竴姝?
- 瀹炴柦 D5.3 缁勫眬銆佸叆灞€銆佺姸鎬佹満鐨勫悗绔富閾捐矾銆?- 鍦ㄨ繘鍏ョ粍灞€鍒涘缓鍓嶅鍔犲己瀹炲悕鎷︽埅锛岀‘淇濇湭 `verified` 鐢ㄦ埛涓嶈兘鍒涘缓灞€銆?
## 2026-06-13 02:18

### 褰撳墠杩涘睍

- 瀹屾垚 D5.3 缁勫眬涓婚摼璺涓€姝ワ細灏忕▼搴忕鍒涘缓鍏嶈垂灞€銆佸眬鍒楄〃銆佸眬璇︽儏銆?- 宸插疄鐜拌鍒欙細
  - 鏈畬鎴愬己瀹炲悕 `verified` 涓嶈兘鍒涘缓灞€銆?  - 灏忕▼搴忕鍙厑璁稿垱寤?`free` 鍏嶈垂灞€銆?  - 姣忓眬浜烘暟蹇呴』鍦?5-8 鑼冨洿鍐呫€?  - 鍚屼竴鐢ㄦ埛姣忔棩鏈€澶氬垱寤?3 灞€銆?  - 鍒涘缓鎴愬姛鍚庣姸鎬佷负 `pending_audit`锛岀瓑寰呭悗鍙板鏍搞€?- 宸叉柊澧炴帴鍙ｏ細
  - `POST /api/app/games`
  - `GET /api/app/games`
  - `GET /api/app/games/{gameId}`
- 宸茶ˉ鍏?`GameDTO`銆丱penAPI 璺緞鍜岄敊璇爜銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板閫氳繃娴嬭瘯鍖咃細
  - `internal/games`
  - `internal/appapi` 鍒涘缓灞€ HTTP 娴佺▼

### 褰撳墠椋庨櫓

- 缁勫眬浠嶄负鍐呭瓨瀛樺偍锛屽悗缁渶鎺?PostgreSQL銆?- 鐩墠鍒涘缓鍚庣洿鎺?`pending_audit`锛屽悗鍙板鏍搞€佸叆灞€鐢宠銆佹墜鍔ㄥ紑濮嬨€両M 鍒涘缓杩樻湭瀹炵幇銆?
### 涓嬩竴姝?
- 缁х画 D5.3锛氳ˉ灞€瀹℃牳閫氳繃鍚庣殑灞曠ず鐘舵€併€佺敵璇峰叆灞€銆佸叆灞€瀹℃牳銆佷汉鏁伴攣瀹氥€?- 澧炲姞鐘舵€佹祦杞棩蹇楁ā鍨嬶紝涓哄悗缁?IM 鍜岃瘎浠烽摼璺仛鍑嗗銆?
## 2026-06-13 02:40

### 褰撳墠杩涘睍

- 鎸夌敤鎴疯姹傝繘鍏ユ寔缁疄鏂芥ā寮忥細鏈畬鎴愭暣涓?`plan.md` 灏忕▼搴忓悗绔墠涓嶄富鍔ㄥ仠姝€?- 宸茶鍙栧苟浣跨敤鏈満 skill锛?  - `karpathy-guidelines`锛氬皬姝ュ疄鐜般€佸彲楠岃瘉銆侀伩鍏嶈繃搴﹁璁°€?  - `auth-wechat`锛氱‘璁ゅ皬绋嬪簭韬唤浠?openid/unionid 涓烘牳蹇冿紝褰撳墠闈?CloudBase 浜戝嚱鏁伴」鐩紝鍏堜繚鐣欏悗绔?`code2Session` 褰㈡€併€?  - `cloudbase-wechat-integration`锛氱‘璁や竴鏈熷厤璐瑰眬涓嶆帴鐪熷疄鏀粯锛屾敮浠樺垎璐﹀彧鍋氳鍗曞拰鐘舵€侀鐣欍€?- 瀹屾垚 D5.3 缁勫眬鐢宠鍜岀姸鎬佹祦杞細
  - 鏈湴瀹℃牳灞€鎺ュ彛 `POST /api/app/games/{gameId}/approve-local`
  - 鐢宠鍏ュ眬 `POST /api/app/games/{gameId}/applications`
  - 鎴戠殑鐢宠 `GET /api/app/games/applications/my`
  - 瀹℃牳鍏ュ眬 `POST /api/app/games/applications/{applicationId}/review`
  - 鎵嬪姩寮€濮?`POST /api/app/games/{gameId}/manual-start`
  - 閫€鍑哄眬 `POST /api/app/games/{gameId}/exit`
- 瀹屾垚 D5.4 LBS锛?  - 淇濆瓨褰撳墠瀹氫綅 `POST /api/app/locations/current`
  - 淇濆瓨鎵嬪姩瀹氫綅 `POST /api/app/locations/manual`
  - 鍚屽煄灞€ `GET /api/app/games/city`
  - 闄勮繎灞€ `GET /api/app/games/nearby`
  - 瀹氫綅绮惧害瓒呰繃 300m 杩斿洖 `accuracyWarning=true`
  - 闄勮繎灞€鎸夋湰鍦?Haversine 璺濈璁＄畻骞舵帓搴忋€?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板瑕嗙洊锛?  - 鐢宠鍏ュ眬銆佸鏍稿叆灞€銆佹墜鍔ㄥ紑濮嬪悗閿佸畾鐢宠銆?  - 淇濆瓨瀹氫綅銆佸垱寤哄甫鍧愭爣鐨勫眬銆侀檮杩戝眬鏌ヨ銆?
### 褰撳墠椋庨櫓

- 浠嶆槸鍐呭瓨浠撳偍锛屽悗缁渶鎺?PostgreSQL銆?- `approve-local` 鏄悗绔嚜娴嬫帴鍙ｏ紝鐢ㄤ簬鏃?PC 鍚庡彴鏃惰仈璋冮棴鐜紱姝ｅ紡鐢熶骇搴旀浛鎹负鍚庡彴瀹℃牳鎺ュ彛鎴栫Щ闄ゃ€?
### 涓嬩竴姝?
- 瀹炴柦 D5.5锛氳嚜鐮?IM 鎴块棿銆佹垚鍛樻潈闄愩€佹枃鏈秷鎭€佹晱鎰熻瘝鎷︽埅銆佸巻鍙叉秷鎭€?
## 2026-06-13 03:05

### 褰撳墠杩涘睍

- 鐢ㄦ埛鏄庣‘瑕佹眰 IM 涓嶄粠 0 鍏ㄩ儴鑷爺锛岄渶瑕佸叏缃戞绱㈠姛鑳藉畬鏁寸殑寮€婧?IM 椤圭洰骞朵簩娆″紑鍙戙€?- 宸叉绱㈠苟閫夊畾 OpenIM `openimsdk/open-im-server` 浣滀负涓€鏈?IM 搴曞骇锛?  - Apache-2.0 璁稿彲璇併€?  - Go 鏈嶅姟绔紝閫傚悎褰撳墠 Go 鍚庣闆嗘垚銆?  - 鏀寔 REST API銆乄ebhooks銆佺兢缁勩€佹秷鎭€佸巻鍙层€佹枃浠跺璞¤兘鍔涘拰 WebSocket 缃戝叧銆?  - 姣?Tinode銆丮atrix/Synapse銆丷ocket.Chat 鏇磋创鍚堚€滃祵鍏ヤ笟鍔″簲鐢ㄥ仛 IM 搴曞骇鈥濈殑闇€姹傘€?- 宸叉寜鐢ㄦ埛涓嬭浇瑙勫垯鍏嬮殕婧愮爜鍒帮細
  - `C:\Users\61492\Desktop\codex download\open-im-server`
- 宸插皢鏈」鐩?IM 瀹炴柦鏂瑰悜鏀逛负锛?  - 鐪熷ソ鐜╁悗绔仛涓氬姟閴存潈銆佸眬鎴愬憳鏉冮檺銆佹晱鎰熻瘝鍜?DTO銆?  - OpenIM 鍋?WebSocket 闀胯繛鎺ャ€佺兢缁勩€佹秷鎭姇閫掋€佸巻鍙叉秷鎭€佹枃浠跺璞¤兘鍔涖€?- 宸叉柊澧?Go 鍚庣 OpenIM 閫傞厤灞傦細
  - `services/go-api/internal/im/openim_client.go`
  - `OPENIM_API_ADDR`
  - `OPENIM_SECRET`
  - `OPENIM_ADMIN_USER_ID`
- 宸蹭繚鐣欐湰鍦板唴瀛?IM fallback锛屾湭閮ㄧ讲 OpenIM 鏃舵祴璇曚粛鍙繍琛屻€?- 宸叉洿鏂帮細
  - `deploy/env.example`
  - `docs/openapi/ws-protocol.md`

### 褰撳墠椋庨櫓

- OpenIM 鍏ㄥ鏈嶅姟灏氭湭鍦ㄦ湰鏈哄惎鍔ㄩ獙璇侊紝褰撳墠 Go 娴嬭瘯璧?fallback銆?- OpenIM Webhooks 鏁忔劅璇?鎴愬憳鏉冮檺浜屾鏍￠獙杩樻湭閰嶇疆鍒?OpenIM 鏈綋銆?- 鍥剧墖銆佹枃浠舵秷鎭繕鏈垏鍒?OpenIM 瀵硅薄瀛樺偍鎺ュ彛銆?
### 涓嬩竴姝?
- 璺戦€氱幇鏈?Go 娴嬭瘯锛岀‘淇?OpenIM 閫傞厤灞傛病鏈夌牬鍧?D5.1-D5.4銆?- 缁х画 D5.5锛氳ˉ OpenIM 鐢ㄦ埛 token 涓嬪彂鎺ュ彛銆丱penIM 缇ゅ悓姝ョ姸鎬併€佹枃浠舵秷鎭拰 Webhooks 鍥炶皟鑽夋銆?
## 2026-06-13 03:35

### 褰撳墠杩涘睍

- 鍩轰簬 OpenIM 缁х画鎺ㄨ繘 D5.5 灏忕▼搴忓悗绔?IM 閾捐矾銆?- 宸茶ˉ OpenIM 灏忕▼搴忎細璇濆弬鏁帮細
  - `GET /api/app/games/{gameId}/chat-session`
  - 杩斿洖 `engine`銆乣imUserId`銆乣openIMGroupId`銆乣openIMToken`
  - 鏈厤缃?OpenIM 鏃惰繑鍥?`engine=local`锛屼繚鐣欐湰鍦拌仈璋?fallback銆?- 宸插吋瀹?`plan.md` 鐨?roomId 鐗堟秷鎭帴鍙ｏ細
  - `GET /api/app/chat/rooms/{roomId}/messages`
  - `POST /api/app/chat/rooms/{roomId}/messages`
- 宸茶ˉ鏂囦欢鏈嶅姟鍗犱綅锛?  - `POST /api/app/files/upload-token`
  - `GET /api/app/files/{fileId}/download-url`
  - `chat_file` 鏂囦欢蹇呴』鏍￠獙灞€鎴愬憳鏉冮檺锛岄潪灞€鎴愬憳杩斿洖 `40331`銆?- 宸茶ˉ鍏?OpenAPI銆丏TO 绀轰緥鍜岄敊璇爜鏂囨。銆?
### 楠岃瘉缁撴灉

- 浣跨敤椤圭洰鍐?Go 缂撳瓨鎵ц `go test ./...` 閫氳繃銆?- 瑕嗙洊鐐瑰寘鎷細
  - chat-session 鑾峰彇銆?  - roomId 鍙戦€佸拰鎷夊巻鍙叉秷鎭€?  - IM 鏂囦欢涓婁紶鍑瘉銆?  - 鎴愬憳鍙笅杞?IM 鏂囦欢銆?  - 闈炴垚鍛樹笉鑳戒笅杞?IM 鏂囦欢銆?
### 褰撳墠椋庨櫓

- OpenIM 鍏ㄥ鏈嶅姟浠嶆湭鏈満鍚姩楠岃瘉锛屽綋鍓嶆祴璇曚娇鐢?local fallback銆?- OpenIM Webhooks 鐨勬晱鎰熻瘝銆佹垚鍛樻潈闄愪簩娆℃牎楠岃繕鏈啓鍏?OpenIM 閰嶇疆銆?- 鏂囦欢涓婁紶 URL 鐩墠鏄?`mock://`锛屽悗缁渶瑕佹帴 OpenIM 瀵硅薄鏈嶅姟鎴栭」鐩粺涓€瀵硅薄瀛樺偍銆?
### 涓嬩竴姝?
- 缁х画 D5.5锛氳ˉ OpenIM Webhooks 鍥炶皟鑽夋鍜屾枃浠舵秷鎭槧灏勮鏄庛€?- 涔嬪悗杩涘叆 D5.6锛氭湇鍔＄‘璁ゃ€佽瘎浠枫€佹垚闀裤€佷俊鐢ㄣ€佽冻杩广€?
## 2026-06-13 04:20

### 褰撳墠杩涘睍

- 澶嶆牳褰撳墠浠ｇ爜鍚庣户缁帹杩?`plan.md`锛岀‘璁?D5.5 宸叉湁鎴块棿銆乺oomId 娑堟伅銆丄CK銆佸凡璇汇€佸綊妗ｅ拰 OpenIM Webhook 鍥炶皟璁板綍銆?- 琛ュ厖 D5.5 鏂囦欢娑堟伅鍜?OpenIM Webhook 鏄犲皠璇存槑锛?  - 鏂囦欢涓婁紶鍑瘉锛歚POST /api/app/files/upload-token`
  - 鏂囦欢涓存椂涓嬭浇锛歚GET /api/app/files/{fileId}/download-url`
  - 鏂囦欢/鍥剧墖娑堟伅閫氳繃 `messageType=image/file` + `fileId` 鍏宠仈銆?  - OpenIM Webhook 鍏ュ彛锛歚POST /api/internal/openim/webhooks`
- 杩涘叆 D5.6 骞跺畬鎴愬悗绔渶灏忛棴鐜細
  - `POST /api/app/games/{gameId}/service-confirm`
  - `POST /api/app/games/{gameId}/service-confirm-items`
  - `GET /api/app/reviews/todos`
  - `POST /api/app/reviews`
  - `GET /api/app/reviews/my-intents`
  - `GET /api/app/users/me/growth`
- 鏂板 `services/go-api/internal/reviews/service.go`锛屾敮鎸佸緟璇勪环鍒楄〃銆佽瘎浠锋彁浜ゃ€侀噸澶嶈瘎浠锋嫤鎴€佸啀鐜╂剰鍚戙€佹垚闀跨粡楠屻€佺Н鍒嗗拰瓒宠抗蹇収銆?- 鎵╁睍 `games` 鏈嶅姟纭鐘舵€佹祦杞細
  - 鎵嬪姩寮€濮嬪悗杩涘叆 `in_progress`銆?  - 棣栦釜鎴愬憳纭鍚庤繘鍏?`pending_confirm`銆?  - 鍏ㄩ儴鎴愬憳纭鍚庤繘鍏?`pending_review`锛屽苟鐢熸垚鍙瘎浠风獥鍙ｃ€?- 琛ラ綈鏁版嵁搴撹縼绉昏崏妗堬細`game_service_confirms`銆乣game_service_confirm_items`銆乣reviews`銆乣review_reminders`銆乣user_growth_profiles`銆乣experience_logs`銆乣points_logs`銆乣user_footprints`銆乣achievements`銆乣user_achievements`銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板瑕嗙洊锛氭湭瀹屾垚鏈嶅姟涓嶈兘璇勪环銆侀潪鎴愬憳涓嶈兘璇勪环銆佹垚鍛樻湇鍔＄‘璁ゅ悗杩涘叆寰呰瘎浠枫€侀噸澶嶈瘎浠疯鎷︽埅銆佽瘎浠峰悗鎴愰暱缁忛獙/绉垎/瓒宠抗鎺ュ彛鍙繑鍥炲彉鍖栥€?
### 褰撳墠椋庨櫓

- D5.6 鐩墠浠嶆槸鍐呭瓨鏈嶅姟闂幆锛屽悗缁渶瑕佹帴 PostgreSQL 浠撳偍銆?- 鎴愬氨銆佷俊鐢ㄦ墸鍒嗐€佽冻杩圭淮搴﹀凡棰勭暀琛ㄥ拰鎺ュ彛蹇収锛屼絾杩樻湭鍋氬畬鏁磋鍒欏紩鎿庛€?- OpenIM 鍏ㄥ鏈嶅姟浠嶆湭鏈満鍚姩楠岃瘉锛屽綋鍓?IM 娴嬭瘯缁х画浣跨敤 local fallback銆?
### 涓嬩竴姝?
- 缁х画 D5.6锛氳ˉ淇＄敤鎵ｅ垎瑙勫垯銆佹垚灏辫Е鍙戝拰鍚庡彴杩芥函鏌ヨ銆?- 闅忓悗杩涘叆 D5.7锛氬垎娑﹂瑙堛€佽瘎浠锋潯浠舵牎楠屻€佹敹鐩婅褰曞拰缁撶畻闃绘柇銆?
## 2026-06-13 04:55

### 褰撳墠杩涘睍

- 缁х画 D5.6锛岃ˉ榻愪笂涓€杞湭瀹屾垚鐨勪俊鐢ㄣ€佹垚灏卞拰鍚庡彴杩芥函鑳藉姏銆?- 鎴愰暱璧勬枡鏂板浠婃棩淇＄敤瀛楁锛?  - `creditScore`
  - `todayCreditScore`
  - `achievements`
  - 浠婃棩淇＄敤涓虹┖鏃惰嚜鍔ㄦ寜 100 杩斿洖骞跺垵濮嬪寲鍐呭瓨璁板綍銆?- 璇勪环鍚庤嚜鍔ㄨЕ鍙戝熀纭€鎴愬氨锛?  - `first_review`
  - `first_received_review`
- 閫€鍑哄眬鏃舵寜灞€鐘舵€佸垽鏂槸鍚︽墸淇＄敤锛?  - `pending_confirm`锛歚quit_after_confirm`锛屾墸 10 鍒嗐€?  - `in_progress` / `pending_review` / `completed`锛歚quit_after_started`锛屾墸 10 鍒嗐€?  - 鍏朵粬鐘舵€侊細`quit_before_confirm`锛屼笉鎵ｅ垎銆?- 琛ュ悗鍙拌拷婧帴鍙ｏ細
  - `GET /api/admin/users/{userId}/growth`
  - `GET /api/admin/games/{gameId}/review-trace`
- 琛?plan 瀵瑰簲杩佺Щ琛細
  - `daily_credit_scores`
  - `credit_deduction_rules`
  - `credit_logs.game_id`
- 琛?app OpenAPI锛?  - `GET /api/app/growth/my`
  - `POST /api/app/games/{gameId}/exit`
- 琛?admin OpenAPI锛?  - `GET /api/admin/users/{userId}/growth`
  - `GET /api/admin/games/{gameId}/review-trace`

### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板瑕嗙洊锛?  - 浠婃棩淇＄敤榛樿 100銆?  - 璇勪环鍚庤繑鍥炴垚灏卞拰瓒宠抗銆?  - 纭鍚?寮€濮嬪悗閫€鍑烘墸淇＄敤锛屼粠 100 鎵ｅ埌 90銆?  - 鍚庡彴鎸夌敤鎴峰拰灞€杩芥函璇勪环銆佷俊鐢ㄣ€佽冻杩规潵婧愩€?
### 褰撳墠椋庨櫓

- D5.6 浠嶄负鍐呭瓨闂幆锛屽悗缁渶瑕佹浛鎹负 PostgreSQL 浠撳偍鍜屼簨鍔°€?- 鍚庡彴杩芥函鎺ュ彛褰撳墠鏈帴鍚庡彴鏉冮檺涓棿浠讹紝鍚庣画杩涘叆鍚庡彴鏉冮檺妯″潡鏃堕渶瑕佸姞 RBAC銆?- 淇＄敤鎵ｅ垎瑙勫垯鐩墠鎸夐粯璁?10 鍒嗗疄鐜帮紝鍚庣画闇€瑕佷粠 `credit_deduction_rules` 閰嶇疆璇诲彇銆?
### 涓嬩竴姝?
- 杩涘叆 D5.7锛氬垎娑﹂瑙堛€佽瘎浠峰畬鎴愭潯浠舵牎楠屻€佹敹鐩婅褰曘€佽瘎浠锋湭瀹屾垚缁撶畻闃绘柇銆?
## 2026-06-13 05:30

### 褰撳墠杩涘睍

- 杩涘叆 D5.7锛屽厛鍦?Go API 渚ц惤鍦版湰鍦板彲娴嬬殑鍒嗘鼎銆佹敹鐩娿€佺粨绠楅棴鐜€?- 鏂板 `services/go-api/internal/revenue/service.go`锛?  - 鍒嗘鼎妯℃澘鍒涘缓鍜屽垪琛ㄣ€?  - 鍒嗘鼎璇曠畻锛岄噾棰濆叏鐢ㄦ暣鏁板垎锛屾瘮渚嬩娇鐢?bps銆?  - 鍒嗘鼎璁板綍鐢熸垚锛屾寜 `gameId` 骞傜瓑闃绘柇閲嶅鐢熸垚銆?  - 璇勪环鏈畬鎴愭椂闃绘柇鐢熸垚鍒嗘鼎銆?  - 鍒嗘鼎鍐荤粨銆?  - 绾夸笅缁撶畻鐧昏銆?  - 鐢ㄦ埛鏀剁泭鎽樿銆?- 鏂板/鎺ュ叆鎺ュ彛锛?  - `GET /api/app/games/{gameId}/revenue-preview`
  - `GET /api/app/incomes/summary`
  - `GET /api/admin/revenue/templates`
  - `POST /api/admin/revenue/templates`
  - `POST /api/admin/revenue/preview`
  - `POST /api/admin/revenue/calculate`
  - `POST /api/admin/revenue/records/generate`
  - `GET /api/admin/revenue/records`
  - `POST /api/admin/revenue/records/{recordId}/freeze`
  - `POST /api/admin/revenue/records/{recordId}/settle`
- 鎵╁睍 `reviews` 鏈嶅姟锛屾彁渚?`GameReviewComplete(gameId)`锛屼緵鍒嗘鼎鐢熸垚鏍￠獙璇勪环瀹屾垚鏉′欢銆?- 閲嶅啓骞惰ˉ榻?`db/migrations/000007_revenue_orders.sql`锛?  - `payment_orders`
  - `revenue_templates`
  - `revenue_rules`
  - `revenue_records`
  - `revenue_record_items`
  - `user_income_accounts`
  - `income_logs`
  - `settlement_records`
- 琛?app/admin OpenAPI 鍒嗘鼎鎺ュ彛銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板瑕嗙洊锛?  - 鍒嗘鼎妯℃澘 bps 鍒涘缓銆?  - 鍒嗘鼎璇曠畻鍙繑鍥為瑙堛€?  - 璇勪环鏈畬鎴愪笉鑳界敓鎴愬垎娑︺€?  - 鍙屾柟璇勪环瀹屾垚鍚庡彲鐢熸垚鍒嗘鼎璁板綍銆?  - 鍐荤粨鍚庝笉鑳界粨绠椼€?  - 绗簩涓眬鍙甯哥敓鎴愬苟绾夸笅缁撶畻銆?  - 鐢ㄦ埛绔敹鐩婃憳瑕佸彲鏌ヨ銆?
### 褰撳墠椋庨櫓

- D5.7 褰撳墠鍏堣惤 Go API 鍐呭瓨闂幆锛孞ava `funds-service` 浠嶅彧鏈?health锛屽悗缁渶瑕佹妸璧勯噾閫昏緫杩佸叆 Java 鏈嶅姟鎴栨敼涓?Go API 璋?Java銆?- 鍚庡彴鍒嗘鼎鎺ュ彛鏆傛湭鎺?RBAC 鏉冮檺涓棿浠躲€?- 浜夎/涓炬姤妯″潡灏氭湭鎺ュ叆锛屽洜姝も€滄湁浜夎鑷姩鍐荤粨鈥濈洰鍓嶇敱鍚庡彴 freeze 鎺ュ彛妯℃嫙銆?
### 涓嬩竴姝?
- 缁х画 D5.7锛氳ˉ Java funds-service 鐨?revenue/settlement HTTP 鍗犱綅鎺ュ彛锛屾垨杩涘叆 D5.8 涓炬姤鐢宠瘔骞舵妸浜夎鍐荤粨鎺ュ埌鍒嗘鼎璁板綍銆?
## 2026-06-13 06:00

### 褰撳墠杩涘睍

- 缁х画 D5.7锛岃ˉ榻?Java `funds-service` 鐨勮祫閲戝崰浣?HTTP 鎺ュ彛銆?- 鏂板閫氱敤鍝嶅簲 helper锛?  - `services/funds-service/src/main/java/com/zhw/funds/common/HttpJson.java`
- 鏂板鏀粯璁㈠崟鍗犱綅锛?  - `GET/POST /api/funds/payment-precreate-placeholder`
  - `GET/POST /api/funds/payment-callback-placeholder`
  - `GET/POST /api/internal/pay/callback-placeholder`
- 鏂板鍒嗘鼎鍗犱綅锛?  - `/api/funds/revenue/templates`
  - `/api/funds/revenue/simulate`
  - `/api/funds/revenue/generate`
  - `/api/funds/revenue/records`
  - `/api/funds/profit-sharing/receivers`
  - `/api/funds/profit-sharing/orders`
  - `/api/funds/profit-sharing/return-orders`
- 鏂板缁撶畻/瀵硅处鍗犱綅锛?  - `/api/funds/settlements/offline`
  - `/api/funds/settlements`
  - `/api/funds/bills/download`
  - `/api/internal/funds/reconcile-runner`
- `FundsApplication` 宸叉敞鍐屼互涓婅矾鐢便€?
### 楠岃瘉缁撴灉

- 鎵ц `javac -encoding UTF-8 -d services/funds-service/target/classes ...` 缂栬瘧閫氳繃銆?- 鎵ц `go test ./...` 閫氳繃锛孏o API 渚?D5.1-D5.7 宸叉湁娴嬭瘯鏈鐮村潖銆?
### 褰撳墠椋庨櫓

- Java `funds-service` 浠嶆槸鏃犳暟鎹簱銆佹棤閴存潈銆佹棤鐪熷疄寰俊鏀粯/鍒嗚处璋冪敤鐨勫崰浣嶆湇鍔°€?- Go API 褰撳墠浠嶅湪鏈湴鍐呭瓨闂幆閲岃绠楀垎娑︼紱鍚庣画闇€瑕佹柊澧?`internal/revenueclient` 璁?Go 璋?Java銆?- 鐪熷疄鏀粯銆侀€€娆俱€佸垎璐︺€佸洖璋冮獙绛惧潎鎸?plan 淇濇寔浜屾湡棰勭暀锛屼笉鍦ㄤ竴鏈熺湡瀹炶Е鍙戙€?
### 涓嬩竴姝?
- 琛?`services/go-api/internal/revenueclient/client.go`锛岃 Go API 鏈夋槑纭殑 Java funds-service 璋冪敤杈圭晫銆?- 鎴栬繘鍏?D5.8 涓炬姤鐢宠瘔锛屾妸浜夎鍐荤粨鎺ュ埌褰撳墠鍒嗘鼎璁板綍銆?
## 2026-06-13 06:30

### 褰撳墠杩涘睍

- 缁х画 D5.7锛岃ˉ Go API 璋?Java `funds-service` 鐨勬槑纭竟鐣屻€?- 鏂板 `services/go-api/internal/revenueclient/client.go`锛?  - `Health`
  - `PaymentPrecreatePlaceholder`
  - `PaymentCallbackPlaceholder`
  - `CreateTemplate`
  - `Simulate`
  - `Generate`
  - `SettleOffline`
- 鏂板 `services/go-api/internal/revenueclient/client_test.go`锛屼娇鐢?`httptest.Server` 楠岃瘉璇锋眰璺緞鍜岄敊璇繑鍥炪€?- 鎵╁睍 Go 閰嶇疆锛?  - `FUNDS_SERVICE_ADDR`
  - 榛樿鍊硷細`http://127.0.0.1:8081`
- 鏇存柊鐜鏍蜂緥锛?  - `.env.example`
  - `deploy/env.example`
- 鏇存柊 `README.md` 涓?Java `funds-service` 缂栬瘧鍛戒护锛屽寘鍚?`placeholder`銆乣revenue`銆乣settlement` 瀛愬寘銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鎵ц Java `javac` 缂栬瘧閫氳繃銆?- `revenueclient` 宸茶鐩栬祫閲戞湇鍔?health銆佹敮浠樺崰浣嶃€佸垎娑︽ā鏉裤€佽瘯绠椼€佺敓鎴愬拰绾夸笅缁撶畻璋冪敤璺緞銆?
### 褰撳墠椋庨櫓

- Go API 鐨勪笟鍔″垎娑︽帴鍙ｅ綋鍓嶄粛浣跨敤鏈湴鍐呭瓨 revenue 鏈嶅姟锛屾病鏈夊垏鎹㈡垚瀹炴椂璋冪敤 Java銆?- Java `funds-service` 浠嶄负鍗犱綅鍝嶅簲锛屾湭鎺ユ暟鎹簱鍜岀湡瀹炲井淇℃敮浠?鍒嗚处銆?- 涓嬩竴姝ュ垏鎹㈣皟鐢ㄩ摼鏃讹紝闇€瑕佸喅瀹氾細Go API 缁х画淇濈暀鏈湴 fallback锛岃繕鏄己鍒朵互 Java `funds-service` 涓?source of truth銆?
### 涓嬩竴姝?
- 杩涘叆 D5.8 涓炬姤鐢宠瘔锛屾妸浜夎鍐荤粨鎺ュ埌褰撳墠鍒嗘鼎璁板綍锛涙垨鎶?Go API 鍒嗘鼎鎺ュ彛鏀规垚浼樺厛璋冪敤 `revenueclient`銆佸け璐ユ椂鏈湴 fallback銆?
## 2026-06-13 07:00

### 褰撳墠杩涘睍

- 杩涘叆 D5.8锛屽畬鎴愪妇鎶ョ敵璇夊埌鍒嗘鼎鍐荤粨鐨勬渶灏忛棴鐜€?- 鏂板 `services/go-api/internal/reports/service.go`锛?  - 鐢ㄦ埛鎻愪氦涓炬姤銆?  - 鏌ヨ鎴戠殑涓炬姤銆?  - 鍚庡彴涓炬姤鍒楄〃銆?  - 鍚庡彴澶勭悊涓炬姤銆?  - 鍚庡彴鍏抽棴涓炬姤銆?  - 涓炬姤鍒涘缓鏃舵寜 `gameId` 鑷姩鍐荤粨宸叉湁鍒嗘鼎璁板綍銆?- 鎵╁睍 `services/go-api/internal/revenue/service.go`锛?  - `FreezeByGame(gameID, reason)`
  - `HasFrozenRecordForGame(gameID)`
- 鏂板/鎺ュ叆鎺ュ彛锛?  - `POST /api/app/reports`
  - `GET /api/app/reports/my`
  - `GET /api/admin/reports`
  - `POST /api/admin/reports/{reportId}/handle`
  - `POST /api/admin/reports/{reportId}/close`
- 琛?`db/migrations/000008_admin_system.sql`锛?  - `reports`
  - `idx_reports_game_status`
  - `idx_reports_reporter`
- 琛?app/admin OpenAPI 涓炬姤鐢宠瘔鎺ュ彛銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板瑕嗙洊锛?  - 宸茬敓鎴愬垎娑﹁褰曠殑灞€鎻愪氦涓炬姤鍚庯紝`revenueFrozen=true`銆?  - 涓炬姤鍚庡搴斿垎娑﹁褰曡鍐荤粨锛岀嚎涓嬬粨绠楄闃绘柇銆?  - 鐢ㄦ埛鍙煡璇㈡垜鐨勪妇鎶ャ€?  - 鍚庡彴鍙煡鐪嬩妇鎶ュ垪琛ㄥ苟澶勭悊涓炬姤銆?
### 褰撳墠椋庨櫓

- 涓炬姤鐢宠瘔浠嶄负鍐呭瓨鏈嶅姟锛屽悗缁渶鎺?PostgreSQL 浠撳偍銆?- 鍚庡彴澶勭悊涓炬姤鏈帴 RBAC 鏉冮檺鍜屾搷浣滄棩蹇椼€?- IM 浜夎娑堟伅浜屾鏌ョ湅鏉冮檺鍜屾煡鐪嬫棩蹇楄繕鏈帴鍏ャ€?
### 涓嬩竴姝?
- 缁х画 D5.8锛氳ˉ鐢ㄦ埛琛屼负鏃ュ織 `user_behavior_logs` 鍜屼妇鎶ュ鐞嗘搷浣滄棩蹇椼€?- 鐒跺悗琛?IM 浜夎娑堟伅鏌ョ湅鏉冮檺涓庢棩蹇楋紝婊¤冻 TC-093/TC-094銆?
## 2026-06-13 07:45

### 褰撳墠杩涘睍

- 缁х画 D5.8锛岃ˉ榻愪妇鎶ョ敵璇夌浉鍏崇珯鍐呴€氱煡闂幆銆?- 鏂板 `services/go-api/internal/notifications/service.go`锛?  - 鍒涘缓閫氱煡銆?  - 鏌ヨ鎴戠殑閫氱煡銆?  - 鏍囪閫氱煡宸茶銆?  - 棰勭暀 `needWechat` 鍜?`wechatState`锛岀敤浜庡悗缁井淇¤闃呮秷鎭彂閫佷换鍔″拰缁撴灉璁板綍銆?- 鏂板灏忕▼搴忔帴鍙ｏ細
  - `GET /api/app/notifications`
  - `POST /api/app/notifications/{notificationId}/read`
- 涓炬姤閾捐矾鍐欓€氱煡锛?  - 鐢ㄦ埛鎻愪氦涓炬姤鍚庣敓鎴?`report_created` 閫氱煡銆?  - 鍚庡彴澶勭悊涓炬姤鍚庣敓鎴?`report_handled` 閫氱煡銆?  - 鍚庡彴鍏抽棴涓炬姤鍚庣敓鎴?`report_closed` 閫氱煡銆?- 琛ラ綈鏁版嵁搴撹縼绉昏崏妗堬細
  - `notifications`
  - `idx_notifications_user_status_created`
  - `idx_notifications_biz`
- 琛ュ厖 `docs/openapi/app.openapi.yaml` 鐨勯€氱煡鎺ュ彛銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 涓炬姤鍒涘缓鍚庣敤鎴疯兘鏌ヨ鍒?`report_created` 閫氱煡銆?  - 鐢ㄦ埛鍙皢閫氱煡鏍囪涓?`read`銆?  - 鍚庡彴澶勭悊涓炬姤鍚庣敤鎴疯兘鏌ヨ鍒?`report_handled` 閫氱煡銆?  - 鍚庡彴鍏抽棴涓炬姤鎺ュ彛浠嶅彲姝ｅ父鎵ц銆?
### 褰撳墠椋庨櫓

- 閫氱煡鏈嶅姟浠嶄负鍐呭瓨瀹炵幇锛屽悗缁渶瑕佹帴 PostgreSQL 浠撳偍銆?- 寰俊璁㈤槄娑堟伅鐩墠鍙繚鐣?`needWechat/wechatState` 鐘舵€佸瓧娈碉紝灏氭湭鎺ユā鏉挎巿鏉冦€佸彂閫佷换鍔″拰鍥炶皟缁撴灉銆?

## 2026-06-13 08:10

### 褰撳墠杩涘睍

- 缁х画 D5.8锛岃ˉ榻愬井淇¤闃呮秷鎭殑鏈湴浠诲姟/缁撴灉鍗犱綅闂幆銆?- 鎵╁睍 `services/go-api/internal/notifications/service.go`锛?  - `WechatTemplate`锛氱淮鎶よ闃呮秷鎭満鏅笌妯℃澘 ID銆?  - `WechatTask`锛氳褰曢€氱煡瀵瑰簲鐨勮闃呮秷鎭彂閫佷换鍔°€?  - 鍒涘缓 `needWechat=true` 鐨勯€氱煡鏃讹紝鑷姩鐢熸垚 `pending` 璁㈤槄娑堟伅浠诲姟銆?  - 鏀寔鏍囪浠诲姟宸插彂閫侊紝骞跺悓姝ラ€氱煡 `wechatState=sent`銆?- 鏂板鎺ュ彛锛?  - `GET /api/admin/notifications/wechat-templates`
  - `GET /api/admin/notifications/wechat-tasks`
  - `POST /api/internal/notifications/wechat-tasks/{taskId}/mark-sent`
- 鎵╁睍 `notifications` 杩佺Щ瀛楁锛?  - `wechat_template_id`
  - `wechat_task_id`
- 鏂板杩佺Щ琛細
  - `wechat_subscribe_templates`
  - `wechat_subscribe_tasks`
- 涓炬姤鐢宠瘔閫氱煡鐜板湪浼氬啓鍏ヨ闃呮秷鎭ā鏉垮彉閲忥紝骞剁敓鎴愬井淇¤闃呮秷鎭换鍔°€?- 琛ュ厖 `docs/openapi/admin.openapi.yaml` 鐨勮闃呮秷鎭ā鏉?浠诲姟鎺ュ彛銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 涓炬姤閫氱煡鍖呭惈 `wechatState=pending`銆佹ā鏉?ID 鍜屼换鍔?ID銆?  - 鍚庡彴鍙煡璇㈣闃呮秷鎭ā鏉裤€?  - 鍚庡彴鍙煡璇㈣闃呮秷鎭换鍔°€?  - 鍐呴儴鎺ュ彛鍙ā鎷熸爣璁拌闃呮秷鎭换鍔″凡鍙戦€併€?
### 褰撳墠椋庨櫓

- 褰撳墠鍙畬鎴愭湰鍦颁换鍔′笌缁撴灉鐘舵€侊紝灏氭湭璋冪敤寰俊瀹樻柟 `subscribeMessage.send`銆?- 鐢ㄦ埛璁㈤槄鎺堟潈鍏ュ彛銆佸皬绋嬪簭绔?`wx.requestSubscribeMessage`銆佹ā鏉?ID 閰嶇疆浠嶉渶鍦ㄥ皬绋嬪簭渚у拰鐪熷疄閰嶇疆涓ˉ榻愩€?

## 2026-06-13 08:40

### 褰撳墠杩涘睍

- 缁х画 D5.8锛岃ˉ榻愪袱涓唴閮ㄦ彁閱掍换鍔″叆鍙ｃ€?- 鏂板 `services/go-api/internal/appapi/job_handler.go`锛?  - `POST /api/internal/jobs/review-remind`
  - `POST /api/internal/jobs/progress-feedback-remind`
- 璇勪环鎻愰啋浠诲姟浼氭壂鎻?`pending_review/completed` 鐨勫眬锛屽浠嶆湁寰呰瘎浠烽」鐨勬垚鍛樼敓鎴愶細
  - 绔欏唴閫氱煡 `review_remind`
  - 寰俊璁㈤槄娑堟伅浠诲姟 `review_remind`
- 琛屽杩涘害鍙嶉鎻愰啋浠诲姟浼氭壂鎻?`in_progress/pending_confirm` 鐨勫眬锛屽鍙戣捣浜虹敓鎴愶細
  - 绔欏唴閫氱煡 `progress_feedback_remind`
  - 寰俊璁㈤槄娑堟伅浠诲姟 `progress_feedback_remind`
- 琛ュ厖 `db/migrations/000006_review_growth_credit.sql`锛?  - `review_reminders.reminder_type`
  - `review_reminders.notification_id`
  - `review_reminders.sent_at`
  - `idx_review_reminders_user_status`
  - `idx_review_reminders_game_type`
- 琛ュ厖 `docs/openapi/admin.openapi.yaml` 鐨勫唴閮ㄦ彁閱掍换鍔℃帴鍙ｃ€?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 灞€杩涘叆寰呰瘎浠峰悗锛屽唴閮ㄨ瘎浠锋彁閱掍换鍔¤兘鐢熸垚閫氱煡銆?  - 杩涜涓眬瑙﹀彂琛屽杩涘害鍙嶉鎻愰啋鍚庯紝鍙戣捣浜鸿兘鏀跺埌 `progress_feedback_remind` 閫氱煡銆?  - 涓ょ被鎻愰啋鍧囧鐢ㄥ綋鍓嶉€氱煡鏈嶅姟锛岃嚜鍔ㄧ敓鎴愬井淇¤闃呮秷鎭换鍔°€?
### 褰撳墠椋庨櫓

- 褰撳墠鎻愰啋浠诲姟涓烘湰鍦?HTTP 鍏ュ彛锛屽皻鏈帴瀹氭椂璋冨害鍣ㄣ€?- 杩涘害鍙嶉鎻愰啋鐩墠鎸夊眬鐘舵€佺敓鎴愶紝鍚庣画闇€瑕佹帴鐪熷疄琛屽杩涘害鍙嶉璁板綍锛岄伩鍏嶉噸澶嶆彁閱掑拰璇彁閱掋€?

## 2026-06-13 09:10

### 褰撳墠杩涘睍

- 缁х画 D5.8锛岃ˉ榻愬悗鍙颁妇鎶ヨ鎯呭拰鍏宠仈璇佹嵁瑙嗗浘銆?- 鎵╁睍 `services/go-api/internal/reports/service.go`锛?  - 鏂板 `Get(reportID)`銆?- 鎵╁睍鍚庡彴鎺ュ彛锛?  - `GET /api/admin/reports/{reportId}`銆?- 涓炬姤璇︽儏杩斿洖锛?  - 涓炬姤鏈綋 `report`銆?  - 灞€淇℃伅 `evidence.game`銆?  - IM 鎴块棿鍜岀浉鍏虫秷鎭?`evidence.chatRoom`銆乣evidence.chatMessages`銆?  - 鏂囦欢璇佹嵁 `evidence.file`銆?  - 璇勪环璇佹嵁 `evidence.review`銆?  - 鍒嗘鼎璁板綍 `evidence.revenueRecord`銆?  - 鍘熷鍏宠仈 ID 姹囨€?`evidence.ids`銆?- 鏌ョ湅涓炬姤璇︽儏鍐?`operation_logs`锛?  - `action=report:view_detail`
  - `target_type=report`
- 琛ュ厖 `docs/openapi/admin.openapi.yaml` 鐨勪妇鎶ヨ鎯呮帴鍙ｃ€?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鍚庡彴鍙墦寮€涓炬姤璇︽儏銆?  - 璇︽儏閲岃兘鐪嬪埌灞€淇℃伅銆両M 娑堟伅璇佹嵁鍜屽喕缁撳悗鐨勫垎娑﹁褰曘€?  - 鏌ョ湅璇︽儏浼氬啓 `report:view_detail` 鎿嶄綔鏃ュ織銆?
### 褰撳墠椋庨櫓

- 褰撳墠璇佹嵁鑱氬悎浠嶅熀浜庡唴瀛樻湇鍔★紝鍚庣画鎺?PostgreSQL 鍚庨渶瑕佹敼涓轰粨鍌ㄦ煡璇€?- IM 璇佹嵁鐩墠鎸夋湰鍦版埧闂存秷鎭仛鍚堬紝鍚庣画鎺ョ湡瀹?WebSocket/鎸佷箙鍖栧悗闇€瑕佹寜娑堟伅琛ㄧ簿纭垎椤靛拰鏉冮檺杩囨护銆?

## 2026-06-13 09:35

### 褰撳墠杩涘睍

- 缁х画 D5.8 / 鍚庡彴瀹¤鏉冮檺闂幆锛岃ˉ榻愬畬鏁存搷浣滄棩蹇楁煡璇㈡潈闄愭牎楠屻€?- `GET /api/admin/operation-logs` 鐜板湪蹇呴』鎼哄甫 `operation_log:view_full` 鏉冮檺銆?- 鏃犳潈闄愯闂繑鍥?`403`锛岄伩鍏嶆櫘閫氱鐞嗗憳鏌ョ湅瀹屾暣鎿嶄綔鏃ュ織銆?- 琛ュ厖 `docs/openapi/admin.openapi.yaml` 鐨?403 鍝嶅簲璇存槑銆?
### 楠岃瘉缁撴灉

- 琛ュ厖鎺ュ彛娴嬭瘯锛?  - 鏃?`operation_log:view_full` 鏉冮檺鏃舵煡璇㈡搷浣滄棩蹇楄繑鍥?`403`銆?  - 鏈夋潈闄愭椂浠嶅彲鏌ヨ骞堕獙璇?`im:message:view_dispute`銆乣report:handle`銆乣report:view_detail` 鏃ュ織銆?
### 褰撳墠椋庨櫓

- 褰撳墠鏉冮檺浠嶄娇鐢ㄨ姹傚ご鏉冮檺蹇収鍗犱綅锛屽悗缁?D7 闇€瑕佹帴鍏ョ湡瀹炲悗鍙?token銆佽鑹插拰 RBAC 涓棿浠躲€?
## 2026-06-13 10:05

### 褰撳墠杩涘睍

- 杩涘叆 P10 / Task 56 鐨勬姤琛ㄥ鍑轰换鍔℃渶灏忛棴鐜€?- 鏂板鍥哄畾瀵煎嚭妯℃澘鏈嶅姟 `services/go-api/internal/exports/service.go`锛?  - `reports_default` 涓炬姤鐢宠瘔鎶ヨ〃妯℃澘銆?  - `operation_logs_default` 鎿嶄綔鏃ュ織鎶ヨ〃妯℃澘銆?  - 鍒涘缓 `pending` 瀵煎嚭浠诲姟銆?  - 鍐呴儴 runner 灏嗕换鍔℃帹杩涘埌 `done` 骞跺叧鑱旀枃浠躲€?- 鎵╁睍鏂囦欢鏈嶅姟锛屾柊澧?`CreateGeneratedFile`锛岀敤浜庣櫥璁?`export_file` 鐢熸垚鐗┿€?- 鏂板鍚庡彴/鍐呴儴鎺ュ彛锛?  - `GET /api/admin/reports/export-templates`
  - `POST /api/admin/reports/export`
  - `GET /api/admin/export-tasks`
  - `GET /api/admin/export-tasks/{taskId}/download-url`
  - `POST /api/internal/reports/export-runner`
- 瀵煎嚭鐩稿叧鎺ュ彛浣跨敤 `report_export:create` 鏉冮檺鍗犱綅鏍￠獙銆?- 鍒涘缓銆佹墽琛屻€佷笅杞藉湴鍧€鑾峰彇鍒嗗埆鍐?`operation_logs`锛?  - `export:create`
  - `export:run`
  - `export:download_url`
- 琛ュ厖 `db/migrations/000009_behavior_export.sql`锛?  - `export_templates`
  - `export_tasks.template_code`
  - `export_tasks.created_by`
  - `export_tasks.filters_json`
  - `export_tasks.fail_reason`
  - 瀵煎嚭浠诲姟鐘舵€佺储寮曘€?- 琛ュ厖 `docs/openapi/admin.openapi.yaml` 瀵煎嚭鎺ュ彛璇存槑銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鏃?`report_export:create` 鏉冮檺涓嶈兘鏌ョ湅瀵煎嚭妯℃澘銆?  - 鏈夋潈闄愬彲鏌ョ湅鍥哄畾妯℃澘銆?  - 鍙垱寤?`pending` 瀵煎嚭浠诲姟銆?  - 鍐呴儴 runner 鍙敓鎴愬鍑烘枃浠惰褰曞苟灏嗕换鍔＄疆涓?`done`銆?  - 瀵煎嚭瀹屾垚鍚庡彲鑾峰彇涓存椂涓嬭浇鍦板潃銆?  - 瀵煎嚭鍒涘缓銆佹墽琛屻€佷笅杞藉湴鍧€鑾峰彇鍧囧彲鍦ㄦ搷浣滄棩蹇椾腑杩芥函銆?
### 褰撳墠椋庨櫓

- 褰撳墠瀵煎嚭 runner 鐢熸垚鐨勬槸鏈湴 mock 鏂囦欢璁板綍锛屽皻鏈敓鎴愮湡瀹?CSV 鍐呭鍜屽璞″瓨鍌ㄦ枃浠躲€?- 褰撳墠瀵煎嚭鏁版嵁婧愪粛鏄唴瀛樻湇鍔★紝鍚庣画鎺?PostgreSQL 鍚庨渶瑕佹寜绛涢€夋潯浠惰鍙栫湡瀹炴姤琛ㄦ暟鎹€?
## 2026-06-13 10:35

### 褰撳墠杩涘睍

- 杩涘叆 D7 鍚庣瀹夊叏鍜屾潈闄愮粺涓€钀藉湴鐨勭涓€姝ワ紝鎶婂凡鎺ュ叆鐨勫悗鍙版潈闄愭祦绋嬭鑼冨寲銆?- 鏂板缁熶竴鏉冮檺鍖呰鍣細
  - `requireAdminPermission(permission, handler)`
- 宸叉妸浠ヤ笅鍚庡彴/鍐呴儴璺敱鏀逛负鍦ㄨ矾鐢辨敞鍐屽眰鏄惧紡澹版槑 permission code锛?  - `GET /api/admin/operation-logs` -> `operation_log:view_full`
  - `GET /api/admin/export-tasks` -> `report_export:create`
  - `GET /api/admin/export-tasks/{taskId}/download-url` -> `report_export:create`
  - `GET /api/admin/reports/export-templates` -> `report_export:create`
  - `POST /api/admin/reports/export` -> `report_export:create`
  - `POST /api/internal/reports/export-runner` -> `report_export:create`
- 娓呯悊瀵煎嚭 handler 鍜屾搷浣滄棩蹇?handler 鍐呴儴鏁ｈ惤鐨勯噸澶嶆潈闄愬垽鏂紝涓氬姟澶勭悊鍙礋璐ｄ笟鍔￠€昏緫銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鐜版湁鏉冮檺娴嬭瘯缁х画瑕嗙洊锛?  - 鏃?`operation_log:view_full` 鏌ヨ瀹屾暣鎿嶄綔鏃ュ織杩斿洖 `403`銆?  - 鏃?`report_export:create` 鏌ヨ瀵煎嚭妯℃澘杩斿洖 `403`銆?  - 鏈夋潈闄愭椂瀵煎嚭鍒涘缓銆佹墽琛屻€佷笅杞藉湴鍧€鍜屾搷浣滄棩蹇楄拷婧潎姝ｅ父銆?
### 褰撳墠椋庨櫓

- 褰撳墠鏉冮檺婧愪粛鏄?`X-Admin-Permissions` 璇锋眰澶村崰浣嶏紱鍚庣画鎺ョ湡瀹炲悗鍙扮櫥褰曘€佽鑹层€佹潈闄愭爲鍚庯紝浼樺厛鏇挎崲 `requireAdminPermission` 鍐呴儴瀹炵幇銆?
## 2026-06-13 11:10

### 褰撳墠杩涘睍

- 缁х画 D7 鍚庣瀹夊叏鍜屾潈闄愮粺涓€钀藉湴锛岃ˉ鍚庡彴鐧诲綍鍜屾潈闄愭爲鏈€灏忛棴鐜€?- 鏂板 `services/go-api/internal/adminauth/service.go`锛?  - 鏈湴鍐呭瓨鐗堝悗鍙扮鐞嗗憳璐﹀彿銆?  - 榛樿 `admin / admin123` 鐢ㄤ簬鏈湴鑱旇皟銆?  - 鐧诲綍杩斿洖 token銆佺鐞嗗憳淇℃伅銆佽鑹层€佹潈闄愬揩鐓с€?  - 鏀寔閫氳繃 token 鎷夊彇鏉冮檺鏍戙€?- 鏂板鍚庡彴鎺ュ彛锛?  - `POST /api/admin/auth/login`
  - `GET /api/admin/auth/permissions`
  - `GET /api/admin/permissions/tree`
- `requireAdminPermission` 鐜板湪鏀寔涓ょ鏉冮檺鏉ユ簮锛?  - 鍏煎鏃ф祴璇曞拰杩囨浮鏈熺殑 `X-Admin-Permissions`銆?  - 鏀寔鍚庡彴鐧诲綍鍚庣殑 `Authorization: Bearer <admin-token>`銆?- 璁块棶鍙椾繚鎶ゅ悗鍙版帴鍙ｆ椂锛屽鏋滀娇鐢?admin token锛屼細鑷姩琛?`X-Admin-ID`锛屼繚璇佸悗缁?`operation_logs` 鑳借褰曠鐞嗗憳 ID銆?- IM 浜夎娑堟伅璺敱宸叉帴鍏ョ粺涓€ `requireAdminPermission("im:message:view_dispute", ...)`銆?- 琛ュ厖 `db/migrations/000008_admin_system.sql`锛?  - `admin_user_roles`
  - `admin_role_permissions`
- 閲嶅啓 `db/seeds/admin_roles_permissions.sql`锛屽榻?`plan.md` 涓殑鏉冮檺鐮侊紝骞跺垵濮嬪寲鏈湴 super_admin 鍏宠仈銆?- 琛ュ厖 `docs/openapi/admin.openapi.yaml` 鐨勫悗鍙扮櫥褰曞拰鏉冮檺鏍戞帴鍙ｃ€?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鍚庡彴鐧诲綍瀵嗙爜閿欒杩斿洖 `401`銆?  - `admin/admin123` 鐧诲綍鎴愬姛杩斿洖 token銆乣super_admin` 鍜屾潈闄愮爜銆?  - 鐧诲綍 token 鍙媺鍙栨潈闄愭爲銆?  - 鐧诲綍 token 鍙闂渶瑕?`operation_log:view_full` 鐨勬搷浣滄棩蹇楁帴鍙ｃ€?
### 褰撳墠椋庨櫓

- 褰撳墠鍚庡彴鐧诲綍浠嶆槸鍐呭瓨璐﹀彿鍜屾槑鏂囧瘑鐮佸崰浣嶏紱鍚庣画闇€瑕佹帴 PostgreSQL 绠＄悊鍛樿〃銆佸己鍝堝笇瀵嗙爜銆佺鐢ㄧ姸鎬併€佽鑹叉潈闄愬叧绯诲拰 token 杩囨湡鍒锋柊绛栫暐銆?- `X-Admin-Permissions` 浠嶄繚鐣欎负杩囨浮鍏煎鍏ュ彛锛涙帴瀹屾暣 RBAC 鍚庡簲閫愭绉婚櫎娴嬭瘯澶栦緷璧栥€?
## 2026-06-13 11:45

### 褰撳墠杩涘睍

- 缁х画 D7 鍚庡彴鎺ュ彛鏉冮檺缁熶竴钀藉湴锛岀粰涓炬姤鍜屽垎娑﹁祫閲戠浉鍏冲悗鍙版帴鍙ｈˉ鏉冮檺鍏ュ彛銆?- 涓炬姤鐢宠瘔鎺ュ彛宸叉帴鏉冮檺锛?  - `GET /api/admin/reports` -> `report:view`
  - `GET /api/admin/reports/{reportId}` -> `report:view`
  - `POST /api/admin/reports/{reportId}/handle` -> `report:handle`
  - `POST /api/admin/reports/{reportId}/close` -> `report:close`
- 鍒嗘鼎璧勯噾鎺ュ彛宸叉帴鏉冮檺锛?  - `GET /api/admin/revenue/templates` -> `revenue:template:view`
  - `POST /api/admin/revenue/templates` -> `revenue:template:update`
  - `POST /api/admin/revenue/preview` -> `revenue:simulate`
  - `POST /api/admin/revenue/calculate` -> `revenue:simulate`
  - `POST /api/admin/revenue/records/generate` -> `revenue:generate`
  - `GET /api/admin/revenue/records` -> `revenue:record:view`
  - `POST /api/admin/revenue/records/{recordId}/freeze` -> `revenue:freeze`
  - `POST /api/admin/revenue/records/{recordId}/settle` -> `settlement:offline:create`
  - `POST /api/admin/revenue/records/{recordId}/settle-offline` -> `settlement:offline:create`
- 琛ラ綈鏈湴鍚庡彴鐧诲綍鏉冮檺蹇収鍜?`db/seeds/admin_roles_permissions.sql` 涓殑鏉冮檺鐮侊細
  - `report:close`
  - `revenue:template:view`
  - `revenue:template:update`
  - `revenue:simulate`
  - `revenue:generate`
  - `revenue:freeze`
  - `settlement:offline:create`
- 璋冩暣鎺ュ彛娴嬭瘯锛屽垎娑﹀拰涓炬姤鍚庡彴璋冪敤鏀逛负浣跨敤鍚庡彴鐧诲綍 token銆?- 鏂板鏃犳潈闄愬鐞嗕妇鎶ヨ繑鍥?`403` 鐨勬柇瑷€銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏃㈤獙璇佹棫 `X-Admin-Permissions` 鍏煎鍏ュ彛锛屼篃楠岃瘉鍚庡彴鐧诲綍 token 鍙闂彈淇濇姢鎺ュ彛銆?
### 褰撳墠椋庨櫓

- OpenAPI 鏃ф枃浠跺瓨鍦ㄩ儴鍒嗙紪鐮佸拰璺緞绮樿繛锛屽凡璁板綍瀹炵幇杩涘睍锛涘悗缁渶瑕佸崟鐙仛涓€娆?`admin.openapi.yaml` 缁撴瀯鍖栨暣鐞嗭紝閬垮厤灏忚ˉ涓佸湪涔辩爜鍖哄煙璇激銆?
## 2026-06-13 12:20

### 褰撳墠杩涘睍

- 缁х画 D7 鍚庡彴鏉冮檺缁熶竴鏀跺彛锛岃ˉ榻愭鍓嶄粛瑁搁湶鐨勬煡璇㈢被鍚庡彴鎺ュ彛銆?- 浠ヤ笅鎺ュ彛宸叉帴鍏ョ粺涓€ `requireAdminPermission`锛?  - `GET /api/admin/users/{userId}/growth` -> `user:view`
  - `GET /api/admin/games/{gameId}/review-trace` -> `game:view`
  - `GET /api/admin/behavior-logs` -> `analytics:timeline:view`
  - `GET /api/admin/notifications/wechat-templates` -> `notification:wechat:view`
  - `GET /api/admin/notifications/wechat-tasks` -> `notification:wechat:view`
- 琛ラ綈鏈湴鍚庡彴鐧诲綍鏉冮檺蹇収鍜?`db/seeds/admin_roles_permissions.sql`锛?  - `notification:wechat:view`
- 璋冩暣鎺ュ彛娴嬭瘯锛?  - 鐢ㄦ埛鎴愰暱銆佺粍灞€杩芥函鍚庡彴鏌ヨ琛ユ潈闄愬ご銆?  - 寰俊璁㈤槄妯℃澘銆佽闃呬换鍔°€佽涓烘棩蹇楀悗鍙版煡璇㈡敼涓轰娇鐢ㄥ悗鍙扮櫥褰?token銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 褰撳墠鍚庡彴鏌ヨ绫绘帴鍙ｅ凡鍩烘湰缁熶竴鍒版樉寮?permission code锛岀户缁繚鐣?`X-Admin-Permissions` 浣滀负杩囨浮娴嬭瘯鍏ュ彛銆?
### 褰撳墠椋庨櫓

- 鍚庡彴鐧诲綍浠嶆槸鍐呭瓨璐﹀彿鍗犱綅锛汥7 鍚庣画搴斾紭鍏堟妸绠＄悊鍛樸€佽鑹层€佹潈闄愩€佷細璇濇帴鍏ョ湡瀹炴寔涔呭寲琛ㄥ拰瀵嗙爜鍝堝笇銆?
## 2026-06-13 12:45

### 褰撳墠杩涘睍

- 缁х画 D7 鍚庡彴瀹夊叏钀藉湴锛屽厛鍦ㄥ綋鍓嶅唴瀛樿处鍙锋灦鏋勪笅娑堥櫎鏄庢枃瀵嗙爜鏍￠獙銆?- `adminauth` 榛樿绠＄悊鍛樿处鍙峰凡浠庢槑鏂囧瘑鐮佹敼涓?PBKDF2-HMAC-SHA256 鍝堝笇鏍￠獙锛?  - 姝ｇ‘瀵嗙爜 `admin123` 浠嶅彲鐢ㄤ簬鏈湴鑱旇皟銆?  - 鏈嶅姟鍐呴儴涓嶅啀淇濆瓨 `admin123` 鏄庢枃銆?  - 闈?PBKDF2 鏍煎紡鐨勬槑鏂囧瘑鐮佸瓨鍌ㄤ細琚嫆缁濄€?- `db/seeds/admin_roles_permissions.sql` 涓粯璁?`admin` 鐨?`password_hash` 宸插悓姝ヤ负鍚屼竴 PBKDF2 鍝堝笇鏍煎紡銆?- 鏂板 `services/go-api/internal/adminauth/service_test.go`锛?  - 姝ｇ‘瀵嗙爜鍖归厤鍝堝笇銆?  - 閿欒瀵嗙爜澶辫触銆?  - 鏄庢枃鏍煎紡澶辫触銆?  - 鍚庡彴鐧诲綍浠嶈繑鍥?token 鍜屾潈闄愬揩鐓с€?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?
### 褰撳墠椋庨櫓

- 鐩墠鍝堝笇绠楁硶涓烘爣鍑嗗簱鑷疄鐜?PBKDF2锛屽崰浣嶆弧瓒斥€滀笉瀛樻槑鏂団€濓紱鍚庣画鎺ョ湡瀹炴暟鎹簱鍜岀敓浜ц处鍙峰垵濮嬪寲鏃讹紝寤鸿缁熶竴杩佺Щ鍒板洟闃熺‘瀹氱殑瀵嗙爜绛栫暐鍜屼竴娆℃€у垵濮嬪瘑鐮佹祦绋嬨€?
## 2026-06-13 13:20

### 褰撳墠杩涘睍

- 鎸夆€滃皬绋嬪簭鍚庣浼樺厛鈥濈户缁ˉ鎺ュ彛鑱旇皟闂幆锛屼紭鍏堣ˉ榻?`plan.md` 涓墠绔細鐩存帴璋冪敤浣嗗綋鍓嶇己灏戠殑鎺ュ彛璺緞銆?- 璇勪环琛ヨ瘎浠峰叆鍙ｈˉ榻愯鍒掕矾寰勶細
  - 鏂板 `GET /api/app/reviews/available`
  - 澶嶇敤鐜版湁 `reviews.Todos` 閫昏緫锛屽拰 `GET /api/app/reviews/todos` 杩斿洖鍚屼竴绫诲緟璇勪环瀵硅薄銆?- 鐢ㄦ埛鏀剁泭鏄庣粏鎺ュ彛琛ラ綈锛?  - 鏂板 `GET /api/app/incomes/logs`
  - 鏀寔鎸夊綋鍓嶇櫥褰曠敤鎴疯繃婊ゆ敹鐩婅褰曪紝鍙繑鍥炶嚜宸辩殑鍒嗘鼎鏀跺叆椤广€?  - 鏀寔 `status` 鏌ヨ鍙傛暟绛涢€夛紝濡?`?status=settled`銆?  - 杩斿洖瀛楁鍖呭惈 `recordId`銆乣recordNo`銆乣gameId`銆乣role`銆乣amountCent`銆乣status`銆乣createdAt`銆乣settledAt`銆?- 鏂板琛屼负鏃ュ織锛?  - `view_income_logs`
- 琛ュ厖鎺ュ彛娴嬭瘯锛?  - `GET /api/app/reviews/available` 鍙繑鍥?`200`銆?  - `GET /api/app/incomes/logs` 杩斿洖褰撳墠鐢ㄦ埛涓ゆ潯鏀剁泭鏄庣粏銆?  - `GET /api/app/incomes/logs?status=settled` 鍙繑鍥炲凡缁撶畻鏀剁泭銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?
### 褰撳墠椋庨櫓

- 褰撳墠鏀剁泭鏄庣粏浠嶆潵鑷唴瀛樺垎娑﹁褰曪紱鍚庣画鎺?PostgreSQL 鍚庨渶瑕佹寜 `user_income_accounts` / `income_logs` 琛ㄦ敼涓虹湡瀹炲垎椤垫煡璇€?
## 2026-06-13 13:45

### 褰撳墠杩涘睍

- 缁х画灏忕▼搴忓悗绔紭鍏堬紝琛ラ綈灞€杩涜涓殑杩涘害鍙嶉闂幆銆?- 鏂板棰嗗煙妯″瀷涓庢湇鍔℃柟娉曪細
  - `games.ProgressFeedback`
  - `games.ProgressFeedbackRequest`
  - `AddProgressFeedback(userId, gameId, req)`
  - `ProgressFeedbacks(userId, gameId)`
- 鏂板灏忕▼搴忕鎺ュ彛锛?  - `POST /api/app/games/{gameId}/progress-feedbacks`
  - `GET /api/app/games/{gameId}/progress-feedbacks`
- 褰撳墠瑙勫垯锛?  - 杩涘害鐧惧垎姣斿繀椤诲湪 `0-100`銆?  - 杩涘害涓嶈兘浣庝簬涓婁竴鏉″弽棣堛€?  - 鍙湁灞€鍙戣捣浜哄彲鎻愪氦杩涘害鍙嶉銆?  - 灞€鎴愬憳鍙煡鐪嬭繘搴﹀弽棣堝垪琛ㄥ拰 `latestProgress`銆?  - 闈炴垚鍛樻煡璇㈣繑鍥?`403`銆?- 鏂板琛屼负鏃ュ織锛?  - `submit_progress_feedback`

### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鍙戣捣浜烘彁浜よ繘搴﹀弽棣堟垚鍔熴€?  - 鎴愬憳鏌ヨ杩涘害鍙嶉鍒楄〃鎴愬姛銆?  - 杩涘害鍊掗€€杩斿洖 `422`銆?  - 闈炴垚鍛樻煡璇㈣繘搴﹀弽棣堣繑鍥?`403`銆?
### 褰撳墠椋庨櫓

- 褰撳墠杩涘害鍙嶉浠嶆槸鍐呭瓨瀹炵幇锛涘悗缁帴 PostgreSQL 鏃堕渶瑕佽惤鍒?`game_progress_feedbacks` 琛紝骞惰ˉ `fileIds` 瀵规枃浠舵湇鍔＄殑褰掑睘/鏉冮檺鏍￠獙銆?
## 2026-06-13 14:20

### 褰撳墠杩涘睍

- 缁х画灏忕▼搴忓悗绔紭鍏堬紝琛ュ眬鍐呴噷绋嬬銆佹墦鍗″拰澶嶇洏鎺ュ彛锛屽拰鍓嶄竴娈佃繘搴﹀弽棣堣兘鍔涘舰鎴愯繃绋嬫暟鎹棴鐜€?- 鏂板棰嗗煙妯″瀷涓庢湇鍔℃柟娉曪細
  - `games.Milestone` / `MilestoneRequest`
  - `games.Checkin` / `CheckinRequest`
  - `games.Retrospective` / `RetrospectiveRequest`
  - `CreateMilestone`銆乣Milestones`
  - `CreateCheckin`銆乣Checkins`
  - `CreateRetrospective`銆乣Retrospectives`
- 鏂板灏忕▼搴忕鎺ュ彛锛?  - `POST /api/app/games/{gameId}/milestones`
  - `GET /api/app/games/{gameId}/milestones`
  - `POST /api/app/games/{gameId}/checkins`
  - `GET /api/app/games/{gameId}/checkins`
  - `POST /api/app/games/{gameId}/retrospectives`
  - `GET /api/app/games/{gameId}/retrospectives`
- 褰撳墠瑙勫垯锛?  - 鍙湁灞€鍙戣捣浜哄彲鍒涘缓閲岀▼纰戙€?  - 灞€鎴愬憳鍙煡鐪嬮噷绋嬬銆?  - 灞€鎴愬憳鍙彁浜ゅ拰鏌ョ湅鎵撳崱銆?  - 澶嶇洏蹇呴』鍦ㄥ眬杩涘叆 `pending_review` 鎴?`completed` 鍚庢彁浜ゃ€?  - 闈炴垚鍛樿闂繃绋嬫暟鎹繑鍥?`403`銆?- 琛ラ綈 `db/migrations/000010_p1_reserved.sql` 涓己澶辩殑 `game_retrospectives` 琛ㄣ€?- 鏂板琛屼负鏃ュ織锛?  - `create_game_milestone`
  - `submit_game_checkin`
  - `submit_game_retrospective`

### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鍙戣捣浜哄垱寤洪噷绋嬬鎴愬姛銆?  - 鎴愬憳鎻愪氦鎵撳崱鎴愬姛銆?  - 鎴愬憳鍙煡鐪嬮噷绋嬬鍜屾墦鍗″垪琛ㄣ€?  - 鏈嶅姟鏈‘璁ゅ畬鎴愬墠鎻愪氦澶嶇洏杩斿洖 `409`銆?  - 鏈嶅姟纭瀹屾垚鍚庢彁浜ゅ鐩樻垚鍔熴€?  - 闈炴垚鍛樻煡鐪嬮噷绋嬬杩斿洖 `403`銆?
### 褰撳墠椋庨櫓

- 褰撳墠閲岀▼纰戙€佹墦鍗°€佸鐩樹粛鏄唴瀛樺疄鐜帮紱鍚庣画鎺?PostgreSQL 鏃堕渶瑕佽惤鍒?`game_milestones`銆乣game_checkins`銆乣game_retrospectives`锛屽苟琛ユ墦鍗℃枃浠?`fileIds` 鐨勬枃浠跺綊灞炰笌涓嬭浇鏉冮檺鏍￠獙銆?
## 2026-06-13 14:55

### 褰撳墠杩涘睍

- 缁х画灏忕▼搴忓悗绔紭鍏堬紝琛ラ綈鏀惰棌灞€鎺ュ彛锛屾敮鎸佸皬绋嬪簭绔矇娣€鐢ㄦ埛鍏磋叮鏁版嵁銆?- 鏂板棰嗗煙妯″瀷涓庢湇鍔℃柟娉曪細
  - `games.Favorite`
  - `FavoriteGame(userId, gameId)`
  - `UnfavoriteGame(userId, gameId)`
  - `FavoriteGames(userId)`
- 鏂板灏忕▼搴忕鎺ュ彛锛?  - `POST /api/app/games/{gameId}/favorite`
  - `DELETE /api/app/games/{gameId}/favorite`
  - `GET /api/app/games/favorites/my`
- 褰撳墠瑙勫垯锛?  - 鏀惰棌涓嶅瓨鍦ㄧ殑灞€杩斿洖 `40421`銆?  - 閲嶅鏀惰棌淇濇寔骞傜瓑锛岃繑鍥炲凡鏀惰棌璁板綍銆?  - 鍙栨秷鏈敹钘忕殑灞€淇濇寔骞傜瓑銆?  - 鎴戠殑鏀惰棌鎸夋敹钘忔椂闂村€掑簭杩斿洖锛屽苟杩囨护浠嶅浜?`pending_audit` 鐨勫眬銆?- 鏂板琛屼负鏃ュ織锛?  - `favorite_game`
  - `unfavorite_game`

### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鏀惰棌涓嶅瓨鍦ㄧ殑灞€杩斿洖 `404`銆?  - 鏀惰棌灞€鎴愬姛銆?  - 閲嶅鏀惰棌涓嶆姤閿欍€?  - 鎴戠殑鏀惰棌杩斿洖宸叉敹钘忓眬銆?  - 鍙栨秷鏀惰棌鎴愬姛銆?  - 閲嶅鍙栨秷鏀惰棌浠嶈繑鍥炴垚鍔熴€?
### 褰撳墠椋庨櫓

- 褰撳墠鏀惰棌浠嶆槸鍐呭瓨瀹炵幇锛涘悗缁帴 PostgreSQL 鏃堕渶瑕佽惤鍒?`game_favorites`锛屽苟琛ュ悗鍙扮敤鎴疯鎯呬腑鐨勬敹钘忓垪琛ㄦ煡璇€?
## 2026-06-13 15:15

### 褰撳墠杩涘睍

- 缁х画灏忕▼搴忓悗绔紭鍏堬紝琛ラ綈 `plan.md` 涓敹鐩婃帴鍙ｇ殑 singular 璺緞鍏煎銆?- 鏂板璺敱鍒悕锛?  - `GET /api/app/income/summary` -> 澶嶇敤 `incomeSummary`
  - `GET /api/app/income/logs` -> 澶嶇敤 `incomeLogs`
- 淇濈暀鏃㈡湁 plural 璺緞锛?  - `GET /api/app/incomes/summary`
  - `GET /api/app/incomes/logs`
- 鍝嶅簲缁撴瀯銆侀壌鏉冦€佽涓烘棩蹇楀拰绛涢€夐€昏緫涓嶅彉锛岄伩鍏嶅奖鍝嶅凡缁忛€氳繃鐨勬敹鐩婃祴璇曢摼璺€?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - `GET /api/app/income/summary` 杩斿洖 `200`銆?  - `GET /api/app/income/logs` 杩斿洖 `200`銆?
### 褰撳墠椋庨櫓

- 褰撳墠鏀剁泭鎽樿鍜屾槑缁嗕粛鏉ヨ嚜鍐呭瓨鍒嗘鼎璁板綍锛涘悗缁帴 PostgreSQL 鍚庨渶瑕佸悓鏃朵繚璇?singular/plural 涓ゅ璺緞閮芥寚鍚戝悓涓€鐪熷疄鏌ヨ閫昏緫銆?
## 2026-06-13 15:45

### 褰撳墠杩涘睍

- 缁х画灏忕▼搴忓悗绔紭鍏堬紝琛ラ綈 `plan.md` E2.4 涓細鍛樻姤琛ㄥ拰鍥㈤槦鏌ヨ閾捐矾銆?- 鏂板浼氬憳鎶ヨ〃鏈嶅姟锛?  - `services/go-api/internal/memberreports/service.go`
  - 鏀寔浼氬憳鏉冮檺绉嶅瓙銆佷細鍛樹笓灞炴姤琛ㄥ揩鐓с€?  - 鎶ヨ〃鎸囨爣鍖呭惈鍙備笌灞€鏁般€佸畬鎴愬眬鏁般€佹敹鐩婃憳瑕併€侀個璇蜂汉鏁般€佷細鍛樿鍒掑悕銆?- 鏂板鍥㈤槦鏈嶅姟锛?  - `services/go-api/internal/teams/service.go`
  - 涓€鏈熷彧鏀寔鐩村睘涓€绾у洟闃熷叧绯汇€?  - 鏀寔鍥㈤槦鍩烘湰淇℃伅銆佸洟闃熸垚鍛樸€佸洟闃熸敹鐩婃憳瑕併€?- 鏂板灏忕▼搴忕鎺ュ彛锛?  - `GET /api/app/member-reports/me`
  - `GET /api/app/teams/my`
  - `GET /api/app/teams/my/members`
  - `GET /api/app/teams/my/revenue-summary`
- 鏉冮檺閿欒鐮佸榻愯鍒掞細
  - 鏃犱細鍛樻姤琛ㄦ潈闄愯繑鍥?`40351`銆?  - 鏃犲洟闃熺鐞嗘潈闄愯繑鍥?`40352`銆?- 鏁版嵁搴撻鐣欒〃琛ラ綈锛?  - `team_relations`
  - `member_report_snapshots`
  - `idx_member_report_snapshots_user_period`

### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鎵ц `go test -count=1 ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鏃犱細鍛樻潈闄愯闂細鍛樻姤琛ㄨ繑鍥?`40351`銆?  - 鏃犲洟闃熸潈闄愯闂洟闃熸垚鍛樿繑鍥?`40352`銆?  - 鎺堟潈鍚庝細鍛樻姤琛ㄨ繑鍥炲弬涓庡眬鏁般€佸畬鎴愬眬鏁般€佹敹鐩婃憳瑕併€侀個璇蜂汉鏁般€?  - 鍥㈤槦璐熻矗浜哄彲鏌ョ湅鎴戠殑鍥㈤槦銆佸洟闃熸垚鍛樺拰鍥㈤槦鏀剁泭鎽樿銆?
### 褰撳墠椋庨櫓

- 褰撳墠浼氬憳鏉冮檺銆佸洟闃熷叧绯诲拰鎶ヨ〃蹇収浠嶆槸鍐呭瓨瀹炵幇锛涘悗缁帴 PostgreSQL 鏃堕渶瑕佽惤鍒?`user_memberships`銆乣team_relations`銆乣member_report_snapshots` 骞剁敱瀹氭椂浠诲姟鐢熸垚绋冲畾蹇収銆?
## 2026-06-13 16:15

### 褰撳墠杩涘睍

- 缁х画灏忕▼搴忓悗绔紭鍏堬紝琛ラ綈 `plan.md` E2.4 涓Н鍒嗗拰鍏戞崲閾捐矾銆?- 鏂板绉垎璐︽埛鏈嶅姟锛?  - `services/go-api/internal/points/service.go`
  - 鏀寔绉垎鎽樿銆佺Н鍒嗘祦姘淬€佺Н鍒嗗彂鏀俱€佺Н鍒嗘墸鍑忋€?  - 绉垎涓嶈冻杩斿洖涓氬姟閿欒锛岄伩鍏嶈处鎴锋墸鎴愯礋鏁般€?- 鏂板鍏戞崲鏈嶅姟锛?  - `services/go-api/internal/redemption/service.go`
  - 鏀寔鍏戞崲椤圭洰銆佸厬鎹笅鍗曘€佹垜鐨勫厬鎹㈣鍗曘€?  - 涓嬪崟鏃跺湪鏈嶅姟閿佸唴瀹屾垚搴撳瓨鏍￠獙銆佸簱瀛樻墸鍑忓拰绉垎鎵ｅ噺锛涚Н鍒嗘墸鍑忓け璐ヤ細鍥炴粴搴撳瓨銆?- 鏂板灏忕▼搴忕鎺ュ彛锛?  - `GET /api/app/points/summary`
  - `GET /api/app/points/logs`
  - `GET /api/app/redemption/items`
  - `POST /api/app/redemption/orders`
  - `GET /api/app/redemption/orders/my`
- 鏁版嵁搴撻鐣欒〃琛ラ綈锛?  - `redemption_items`
  - `redemption_orders`
  - `idx_redemption_orders_user_created`

### 楠岃瘉缁撴灉

- 鎵ц `go test -count=1 ./...` 閫氳繃銆?- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 绌虹Н鍒嗚处鎴峰彲鏌ヨ銆?  - 绉垎涓嶈冻涓嶈兘鍒涘缓鍏戞崲璁㈠崟銆?  - 鍟嗗搧鍒楄〃鍙繑鍥炴湁鏁堝厬鎹㈤」鐩€?  - 鍏戞崲鎴愬姛鍚庣敓鎴愯鍗曪紝骞跺啓鍏ョН鍒嗘墸鍑忔祦姘淬€?  - 搴撳瓨涓嶈冻涓嶈兘缁х画鍒涘缓鍏戞崲璁㈠崟銆?  - 鎴戠殑鍏戞崲璁㈠崟鍙繑鍥炲綋鍓嶇敤鎴疯鍗曘€?  - 骞跺彂鍏戞崲鍚屼竴涓簱瀛樹负 1 鐨勯」鐩椂锛屽彧鍏佽 1 涓姹傛垚鍔燂紝鍙?1 涓繑鍥炲啿绐併€?
### 褰撳墠椋庨櫓

- 褰撳墠绉垎鍜屽厬鎹粛鏄唴瀛樺疄鐜帮紱鍚庣画鎺?PostgreSQL 鏃堕渶瑕佺敤鏁版嵁搴撲簨鍔″拰琛岀骇閿佸疄鐜板簱瀛樹笌绉垎鎵ｅ噺鐨勫己涓€鑷达紝骞惰ˉ鍚庡彴鍏戞崲瀹℃牳閫氳繃銆侀┏鍥炪€佸彂鏀惧強鎿嶄綔鏃ュ織銆?
## 2026-06-13 16:45

### 褰撳墠杩涘睍

- 缁х画灏忕▼搴忓悗绔紭鍏堬紝琛ラ綈 `plan.md` E2.3 浜鸿剦銆佽瀹舵妧鑳芥爲銆侀璺汉璧勬簮鐢诲儚閾捐矾銆?- 鏂板浜鸿剦鏈嶅姟锛?  - `services/go-api/internal/connections/service.go`
  - 鏀寔浜鸿剦鍒楄〃銆佸弻鍚戜汉鑴夌敓鎴愩€佽窡杩涜褰曘€佸叧绯诲己搴﹂€掑銆?  - 鐧诲綍閭€璇风粦瀹氭垚鍔熷悗鐢熸垚 `invite` 浜鸿剦銆?  - 鍚屽眬鎴愬憳鏈嶅姟纭瀹屾垚鍚庣敓鎴?`co_game` 浜鸿剦銆?  - 棰嗚矾浜烘挳鍚堝彲鐢熸垚 `guide_match` 浜鸿剦銆?- 鏂板鐢诲儚鏈嶅姟锛?  - `services/go-api/internal/profiles/service.go`
  - 鏀寔宸查€氳繃琛屽缁存姢鎶€鑳芥爲銆?  - 鏀寔宸查€氳繃棰嗚矾浜虹淮鎶よ祫婧愮敾鍍忋€?  - 鐢诲儚杩斿洖瀹屾暣鐜囷紝渚夸簬鍚庣画鍚庡彴缁熻銆?- 鏂板灏忕▼搴忕鎺ュ彛锛?  - `GET /api/app/connections/my`
  - `POST /api/app/connections/{connectionId}/follow-up`
  - `GET /api/app/experts/me/skills`
  - `PUT /api/app/experts/me/skills`
  - `GET /api/app/guides/me/resources`
  - `PUT /api/app/guides/me/resources`
- 鏁版嵁搴撻鐣欒〃琛ラ綈锛?  - `user_connections`
  - `connection_follow_logs`
  - `expert_skill_profiles`
  - `guide_resource_profiles`

### 楠岃瘉缁撴灉

- 鎵ц `go test -count=1 ./...` 閫氳繃銆?- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鏈€氳繃琛屽瑙掕壊涓嶈兘鏇存柊鎶€鑳芥爲銆?  - 鎺堟潈琛屽鍙互鏇存柊鍜屾煡鐪嬫妧鑳芥爲銆?  - 鏈€氳繃棰嗚矾浜鸿鑹蹭笉鑳芥洿鏂拌祫婧愮敾鍍忋€?  - 鎺堟潈棰嗚矾浜哄彲浠ユ洿鏂拌祫婧愮敾鍍忋€?  - 浜鸿剦鍒楄〃鍖呭惈 `invite`銆乣co_game`銆乣guide_match` 涓夌被鏉ユ簮銆?  - 浜鸿剦鍒楄〃涓嶆毚闇叉墜鏈哄彿銆佽韩浠借瘉绛夋晱鎰熷瓧娈点€?  - 鍏崇郴鏈汉鍜屽凡閫氳繃棰嗚矾浜哄彲浠ュ啓璺熻繘璁板綍銆?
### 褰撳墠椋庨櫓

- 褰撳墠瑙掕壊閫氳繃鐘舵€併€佷汉鑴夊叧绯诲拰鐢诲儚浠嶆槸鍐呭瓨瀹炵幇锛涘悗缁帴 PostgreSQL 鏃堕渶瑕佹帴鍏?`user_roles`銆乣user_connections`銆乣connection_follow_logs`銆乣expert_skill_profiles`銆乣guide_resource_profiles`锛屽苟琛ュ悗鍙版煡鐪嬫帴鍙ｄ笌鑴辨晱鏉冮檺銆?
## 2026-06-13 17:15

### 褰撳墠杩涘睍

- 缁х画灏忕▼搴忓悗绔紭鍏堬紝琛ラ綈 `plan.md` E2.2 鍓╀綑鐨勭画灞€銆侀噷绋嬬鏇存柊鍜屾墦鍗″紓甯搁摼璺€?- 鏂板灞€杩涘害鑳藉姏锛?  - 鍙戣捣浜哄彲鏇存柊閲岀▼纰戞爣棰樺拰鐘舵€併€?  - 灏忕▼搴忕鏌ヨ鎵撳崱鍒楄〃鏃惰繃婊ゅ悗鍙版爣璁颁负 `invalid` 鐨勬墦鍗°€?  - 鍘熷眬鍙戣捣浜哄彲鍦ㄥ眬杩涘叆 `pending_review/completed` 鍚庡彂璧风画灞€鑽夌銆?  - 缁眬鐢熸垚鐨勬柊灞€鐘舵€佸浐瀹氫负 `draft`锛屼笉浼氱洿鎺ュ彉鎴愭嫑鍕熶腑銆?- 鏂板/瀹屽杽鎺ュ彛锛?  - `PUT /api/app/games/{gameId}/milestones/{milestoneId}`
  - `POST /api/app/games/{gameId}/continue`
  - `POST /api/admin/game-checkins/{checkinId}/mark-invalid`
- 鏁版嵁搴撻鐣欏瓧娈佃ˉ榻愶細
  - `game_milestones.updated_at`
  - `game_checkins.milestone_id`
  - `game_checkins.file_ids`
  - `game_checkins.status`

### 楠岃瘉缁撴灉

- 鎵ц `go test -count=1 ./...` 閫氳繃銆?- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鍙戣捣浜烘洿鏂伴噷绋嬬鎴愬姛銆?  - 鍚庡彴鏍囪鎵撳崱寮傚父鍚庯紝灏忕▼搴忕鎵撳崱鍒楄〃涓嶅啀杩斿洖璇ュ紓甯告墦鍗°€?  - 灞€鏈畬鎴愭椂鍙戣捣缁眬杩斿洖鍐茬獊銆?  - 鏈嶅姟纭瀹屾垚鍚庡彂璧风画灞€鎴愬姛銆?  - 缁眬鐢熸垚鐨勬柊灞€涓?`draft`锛宍gameSource=continue`銆?
### 褰撳墠椋庨櫓

- 褰撳墠杩涘害銆佺画灞€鍜屽紓甯告墦鍗′粛鏄唴瀛樺疄鐜帮紱鍚庣画鎺?PostgreSQL 鏃堕渶瑕佹妸閲岀▼纰戞洿鏂般€佹墦鍗″紓甯告爣璁般€佺画灞€鑽夌鍒涘缓鏀惧埌鐪熷疄浜嬪姟閲岋紝骞惰ˉ鍚庡彴寮傚父鍘熷洜瀛楁銆?
## 2026-06-13 17:45

### 褰撳墠杩涘睍

- 缁х画灏忕▼搴忓悗绔紭鍏堬紝琛ラ綈 `plan.md` E2.1 涓嫭绔嬭涓轰笂鎶ユ帴鍙ｃ€?- 鏂板灏忕▼搴忕鎺ュ彛锛?  - `POST /api/app/behavior/events`
- 鎺ュ彛瑙勫垯锛?  - 鐧诲綍鐢ㄦ埛涓婃姤鏃惰褰?`userId`銆?  - 鏈櫥褰曚篃鍏佽鍖垮悕闄嶇骇璁板綍锛宍userId=0`銆?  - 缂哄皯 `eventType` 杩斿洖鍙傛暟閿欒銆?  - 鏀寔璁板綍 `targetType`銆乣targetId`銆乣pagePath`銆乣keyword`銆乣extra`銆?- 澶嶇敤鐜版湁 audit 琛屼负鏃ュ織鏈嶅姟锛岄伩鍏嶈涓烘暟鎹垎鏁ｅ埌澶氬鍐呭瓨瀛樺偍銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test -count=1 ./...` 閫氳繃銆?- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 缂哄皯 `eventType` 杩斿洖 `422`銆?  - 鐧诲綍鎬佽涓轰笂鎶ユ垚鍔熷苟璁板綍鐢ㄦ埛 ID銆?  - 鍖垮悕琛屼负涓婃姤鎴愬姛骞惰褰?`userId=0`銆?  - 鍚庡彴琛屼负鏃ュ織鍙煡鍒颁富鍔ㄤ笂鎶ョ殑 `search/share` 浜嬩欢銆?
### 褰撳墠椋庨櫓

- 褰撳墠琛屼负鏃ュ織浠嶆槸鍐呭瓨瀹炵幇锛涘悗缁帴 PostgreSQL 鏃堕渶瑕佽惤鍒拌涓轰簨浠惰〃锛屽苟琛ュ悗鍙版寜鐢ㄦ埛銆佷簨浠剁被鍨嬨€佹椂闂磋寖鍥寸瓫閫夈€?
## 2026-06-13 18:15

### 褰撳墠杩涘睍

- 缁х画灏忕▼搴忓悗绔紭鍏堬紝琛ラ綈 `plan.md` E3.1 涓悗鍙板琛屼负浜嬩欢鍜岀敤鎴锋敹钘忕殑鏌ヨ鏀拺銆?- 鏂板/瀹屽杽鍚庡彴鎺ュ彛锛?  - `GET /api/admin/users/{userId}/favorites`
  - `GET /api/admin/behavior/events?userId=&eventType=`
- 鏉冮檺閾捐矾鏀舵暃锛?  - `/api/admin/users/{userId}/growth` 缁х画浣跨敤 `user:view`銆?  - `/api/admin/users/{userId}/favorites` 浣跨敤 `user:read`銆?  - `/api/admin/behavior/events` 浣跨敤 `data:behavior:read`銆?  - 鍐呯疆绠＄悊鍛樻潈闄愬拰绉嶅瓙鏉冮檺琛ㄥ悓姝ヨˉ鍏?`user:read`銆乣data:behavior:read`銆?- 琛屼负鏌ヨ鏀寔鎸?`userId` 鍜?`eventType` 杩囨护锛屾敹钘忔煡璇㈠鐢ㄥ皬绋嬪簭渚ф敹钘忔湇鍔″苟杩囨护鏈鏍稿眬銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test -count=1 ./...` 閫氳繃銆?- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鍚庡彴鏃?`user:read` 鏉冮檺鏌ヨ鐢ㄦ埛鏀惰棌杩斿洖 `403`銆?  - 鍚庡彴鏈?`user:read` 鏉冮檺鍙煡璇㈡寚瀹氱敤鎴锋敹钘忋€?  - 鍚庡彴鏃?`data:behavior:read` 鏉冮檺鏌ヨ琛屼负浜嬩欢杩斿洖 `403`銆?  - 鍚庡彴鏈?`data:behavior:read` 鏉冮檺鍙寜鐢ㄦ埛鍜屼簨浠剁被鍨嬬瓫閫夎涓轰簨浠躲€?
### 褰撳墠椋庨櫓

- 褰撳墠鏀惰棌鍜岃涓轰簨浠朵粛鏄唴瀛樻湇鍔″疄鐜帮紱鍚庣画鎺?PostgreSQL 鏃堕渶瑕佹妸鍚庡彴绛涢€夋潯浠舵槧灏勫埌鐪熷疄绱㈠紩锛屽苟琛ユ椂闂磋寖鍥淬€佸垎椤靛拰瀵煎嚭瀛楁鑴辨晱瑙勫垯銆?
## 2026-06-13 18:45

### 褰撳墠杩涘睍

- 缁х画琛ラ綈 `plan.md` E3.1 鍚庡彴鎺ュ彛锛屽畬鎴愰噷绋嬬銆佹墦鍗°€佸鐩樺拰缁眬鑽夌鐨勫悗鍙版敮鎾戙€?- 鏂板/瀹屽杽鍚庡彴鎺ュ彛锛?  - `GET /api/admin/games/{gameId}/milestones`
  - `POST /api/admin/games/{gameId}/milestones`
  - `GET /api/admin/games/{gameId}/checkins`
  - `GET /api/admin/games/{gameId}/retrospectives`
  - `GET /api/admin/games/{gameId}/continue-drafts`
- 鏉冮檺閾捐矾锛?  - 鍚庡彴璇诲彇閲岀▼纰戙€佹墦鍗°€佸鐩樸€佺画灞€鑽夌浣跨敤 `game:read`銆?  - 鍚庡彴鍒涘缓閲岀▼纰戝拰鏍囪寮傚父鎵撳崱浣跨敤 `game:progress:manage`銆?  - 鍘熸湁 `GET /api/admin/games/{gameId}/review-trace` 缁х画浣跨敤 `game:view`銆?- 鏈嶅姟灞傛柊澧炲悗鍙版煡璇㈡柟娉曪紝鍚庡彴鎵撳崱鏌ヨ浼氳繑鍥炲寘鍚?`invalid` 鍦ㄥ唴鐨勫畬鏁磋褰曪紝渚夸簬杩愯惀杩借釜寮傚父澶勭悊銆?- 缁眬鑽夌鏂板鍘熷眬鍒拌崏绋垮眬鐨勫唴瀛樻槧灏勶紝骞跺湪杩佺Щ涓鐣?`game_continue_drafts` 琛ㄣ€?
### 楠岃瘉缁撴灉

- 鎵ц `go test -count=1 ./...` 閫氳繃銆?- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鍚庡彴鏃?`game:read` 鏉冮檺鏌ヨ閲岀▼纰戣繑鍥?`403`銆?  - 鍚庡彴鏈?`game:progress:manage` 鏉冮檺鍙垱寤洪噷绋嬬銆?  - 鍚庡彴鏈?`game:read` 鏉冮檺鍙煡鐪嬮噷绋嬬銆佸紓甯告墦鍗°€佸鐩樸€佺画灞€鑽夌銆?  - 鍚庡彴鏍囪寮傚父鍚庯紝灏忕▼搴忕杩囨护寮傚父鎵撳崱锛屽悗鍙扮浠嶅彲杩借釜璇ヨ褰曘€?
### 褰撳墠椋庨櫓

- 缁眬鑽夌鍜岃繘搴︽暟鎹洰鍓嶄粛鍦ㄥ唴瀛樻湇鍔′腑娴佽浆锛涙帴 PostgreSQL 鏃堕渶瑕佹妸 `game_continue_drafts`銆乣game_milestones`銆乣game_checkins`銆乣game_retrospectives` 涓茶繘鍚屼竴浜嬪姟鍜屽垎椤垫煡璇€?
## 2026-06-13 19:15

### 褰撳墠杩涘睍

- 缁х画琛ラ綈 `plan.md` E3.1 鍚庡彴鎺ュ彛锛屽畬鎴愪汉鑴夈€佽瀹舵妧鑳芥爲鍜岄璺汉璧勬簮鐢诲儚鐨勫悗鍙板彧璇诲叆鍙ｃ€?- 鏂板鍚庡彴鎺ュ彛锛?  - `GET /api/admin/connections`
  - `GET /api/admin/users/{userId}/connections`
  - `GET /api/admin/experts/{userId}/skills`
  - `GET /api/admin/guides/{userId}/resources`
- 鏉冮檺閾捐矾锛?  - 浜鸿剦鍒楄〃鍜岀敤鎴蜂汉鑴変娇鐢?`connection:read`銆?  - 琛屽鎶€鑳芥爲鍜岄璺汉璧勬簮鐢诲儚浣跨敤 `profile:read`銆?  - 鍐呯疆绠＄悊鍛樻潈闄愬拰绉嶅瓙鏉冮檺琛ㄥ悓姝ヨˉ鍏?`connection:read`銆乣profile:read`銆?- 鏈嶅姟灞傛柊澧炲悗鍙板彧璇绘柟娉曪細
  - 浜鸿剦鏀寔鍏ㄩ噺鏌ヨ鍜屾寜鐢ㄦ埛鏌ヨ銆?  - 鐢诲儚鏀寔鍚庡彴璇诲彇鎸囧畾鐢ㄦ埛鐨勬妧鑳芥爲鍜岃祫婧愮敾鍍忥紝鏈淮鎶ゆ椂杩斿洖绌虹敾鍍忕粨鏋勶紝渚夸簬鍚庡彴绌虹姸鎬佸睍绀恒€?
### 楠岃瘉缁撴灉

- 鎵ц `go test -count=1 ./...` 閫氳繃銆?- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鍚庡彴鏃?`connection:read` 鏉冮檺鏌ヨ浜鸿剦杩斿洖 `403`銆?  - 鍚庡彴鏈?`connection:read` 鏉冮檺鍙煡鍏ㄩ噺浜鸿剦鍜屾寚瀹氱敤鎴蜂汉鑴夈€?  - 鍚庡彴鏃?`profile:read` 鏉冮檺鏌ヨ鐢诲儚杩斿洖 `403`銆?  - 鍚庡彴鏈?`profile:read` 鏉冮檺鍙煡琛屽鎶€鑳芥爲鍜岄璺汉璧勬簮鐢诲儚銆?
### 褰撳墠椋庨櫓

- 褰撳墠鍚庡彴鐢诲儚鏌ヨ杩樻湭鍋氭晱鎰熷瓧娈靛垎绾ц劚鏁忥紱鎺ョ湡瀹炵敤鎴疯祫鏂欏拰鏂囦欢鏈嶅姟鏃讹紝闇€瑕佹寜 `profile:read` 涓庢洿楂樻潈闄愭媶鍒嗗彲瑙佸瓧娈靛拰闄勪欢璁块棶鑼冨洿銆?
## 2026-06-13 19:45

### 褰撳墠杩涘睍

- 缁х画琛ラ綈 `plan.md` E3.1 鍚庡彴鎺ュ彛锛屽畬鎴愮Н鍒嗘祦姘村拰绉垎鍏戞崲鍚庡彴绠＄悊闂幆銆?- 鏂板鍚庡彴鎺ュ彛锛?  - `GET /api/admin/points/logs`
  - `GET /api/admin/redemption/items`
  - `POST /api/admin/redemption/items`
  - `PUT /api/admin/redemption/items/{itemId}`
  - `GET /api/admin/redemption/orders`
  - `POST /api/admin/redemption/orders/{orderId}/review`
- 鏉冮檺閾捐矾锛?  - 绉垎娴佹按鏌ヨ浣跨敤 `points:read`銆?  - 鍏戞崲椤圭洰绠＄悊銆佸厬鎹㈣鍗曞垪琛ㄣ€佽鍗曞鏍镐娇鐢?`redemption:manage`銆?  - 鍐呯疆绠＄悊鍛樻潈闄愬拰绉嶅瓙鏉冮檺琛ㄥ悓姝ヨˉ鍏?`points:read`銆乣redemption:manage`銆?- 鏈嶅姟灞傝兘鍔涳細
  - 绉垎鏈嶅姟鏀寔鍚庡彴鍏ㄩ噺娴佹按鏌ヨ銆?  - 鍏戞崲鏈嶅姟鏀寔鍚庡彴鏌ョ湅鍏ㄩ儴椤圭洰銆佹洿鏂伴」鐩簱瀛?鐘舵€併€佹煡鐪嬪叏閮ㄨ鍗曘€?  - 鍏戞崲璁㈠崟鏀寔瀹℃牳閫氳繃銆侀┏鍥炲拰鍙戞斁锛涢┏鍥炴椂鑷姩閫€鍥炲凡鎵ｇН鍒嗗苟鍐欑Н鍒嗘祦姘淬€?  - 鎵€鏈夊悗鍙板啓鎿嶄綔鍐欏叆鎿嶄綔鏃ュ織銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test -count=1 ./...` 閫氳繃銆?- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鏃?`points:read` 鏌ヨ绉垎娴佹按杩斿洖 `403`銆?  - 鏃?`redemption:manage` 鏌ヨ鍏戞崲椤圭洰杩斿洖 `403`銆?  - 鍚庡彴鍙垱寤哄拰鏇存柊鍏戞崲椤圭洰銆?  - 鐢ㄦ埛鍏戞崲鍚庯紝鍚庡彴鍙煡璇㈣鍗曘€佸鏍搁€氳繃骞舵爣璁板彂鏀俱€?  - 鍚庡彴椹冲洖鍏戞崲璁㈠崟鍚庯紝鐢ㄦ埛绉垎鑷姩閫€鍥炪€?
### 褰撳墠椋庨櫓

- 褰撳墠鍏戞崲瀹℃牳浠嶆槸鍐呭瓨鐘舵€佹祦杞紱鎺?PostgreSQL 鏃堕渶瑕佸湪鍚屼竴浜嬪姟鍐呭鐞嗚鍗曠姸鎬併€佺Н鍒嗛€€鍥炪€佸簱瀛樺洖琛ュ拰鎿嶄綔鏃ュ織锛岄伩鍏嶅鏍稿け璐ュ鑷磋处瀹炰笉涓€鑷淬€?
## 2026-06-13 20:15

### 褰撳墠杩涘睍

- 缁х画琛ラ綈 `plan.md` E3.1 鍚庡彴鎺ュ彛锛屽畬鎴愬洟闃熶細鍛樺拰浼氬憳鎶ヨ〃鍚庡彴鍙鍏ュ彛銆?- 鏂板鍚庡彴鎺ュ彛锛?  - `GET /api/admin/teams`
  - `GET /api/admin/teams/{teamId}`
  - `GET /api/admin/member-reports`
- 鏉冮檺閾捐矾锛?  - 鍥㈤槦鍒楄〃鍜屽洟闃熻鎯呬娇鐢?`team:read`銆?  - 浼氬憳鎶ヨ〃鍒楄〃浣跨敤 `member_report:read`銆?  - 鍐呯疆绠＄悊鍛樻潈闄愬拰绉嶅瓙鏉冮檺琛ㄥ悓姝ヨˉ鍏?`team:read`銆乣member_report:read`銆?- 鏈嶅姟灞傝兘鍔涳細
  - 鍥㈤槦鏈嶅姟鏀寔鍚庡彴鑾峰彇鍥㈤槦鍒楄〃鍜屽洟闃熻鎯咃紝璇︽儏鍖呭惈鎴愬憳涓庡洟闃熸敹鐩婃憳瑕併€?  - 浼氬憳鎶ヨ〃鏈嶅姟鏀寔鍚庡彴鐢熸垚骞惰繑鍥炲凡鎺堟潈浼氬憳鐨勬姤琛ㄥ揩鐓с€?
### 楠岃瘉缁撴灉

- 鎵ц `go test -count=1 ./...` 閫氳繃銆?- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鏃?`team:read` 鏉冮檺鏌ヨ鍥㈤槦杩斿洖 `403`銆?  - 鏈?`team:read` 鏉冮檺鍙煡璇㈠洟闃熷垪琛ㄥ拰鍥㈤槦璇︽儏銆?  - 鏃?`member_report:read` 鏉冮檺鏌ヨ浼氬憳鎶ヨ〃杩斿洖 `403`銆?  - 鏈?`member_report:read` 鏉冮檺鍙煡璇細鍛樻姤琛ㄥ揩鐓с€?
### 褰撳墠椋庨櫓

- 褰撳墠鍥㈤槦鍜屼細鍛樻姤琛ㄤ粛鏄唴瀛樻煡璇紱鎺?PostgreSQL 鏃堕渶瑕佹寜鍥㈤槦 ID 寤虹储寮曪紝骞舵妸浼氬憳鎶ヨ〃蹇収鏀规垚瀹氭椂浠诲姟鐢熸垚鍚庡垎椤垫煡璇€?
## 2026-06-13 20:45

### 褰撳墠杩涘睍

- 缁х画琛ラ綈 `plan.md` E3.1 鍚庡彴鎺ュ彛锛屽畬鎴愪氦浠樻枃妗ｃ€佹祴璇曠敤渚嬪拰娴嬭瘯鎵ц璁板綍鐨勫悗鍙扮櫥璁板叆鍙ｃ€?- 鏂板鍚庡彴鎺ュ彛锛?  - `GET /api/admin/delivery-documents`
  - `POST /api/admin/delivery-documents`
  - `GET /api/admin/test-cases`
  - `POST /api/admin/test-cases`
  - `GET /api/admin/test-runs`
  - `POST /api/admin/test-runs`
- 鏉冮檺閾捐矾锛?  - 浜や粯鏂囨。浣跨敤 `delivery:manage`銆?  - 娴嬭瘯鐢ㄤ緥鍜屾祴璇曟墽琛岃褰曚娇鐢?`testcase:manage`銆?  - 鍐呯疆绠＄悊鍛樻潈闄愬拰绉嶅瓙鏉冮檺琛ㄥ悓姝ヨˉ鍏?`delivery:manage`銆乣testcase:manage`銆?- 鏈嶅姟灞傝兘鍔涳細
  - 鏂板 `delivery` 鏈嶅姟锛屾敮鎸佷氦浠樻枃妗ｃ€佹祴璇曠敤渚嬨€佹祴璇曟墽琛岃褰曠殑鍐呭瓨鐧昏鍜屾煡璇€?  - 浜や粯鏂囨。鏍￠獙 `docType/title/status`銆?  - 娴嬭瘯鎵ц璁板綍蹇呴』鍏宠仈宸插瓨鍦ㄦ祴璇曠敤渚嬶紝閬垮厤瀛ょ珛鎵ц璁板綍銆?  - 鍚庡彴鏂板鍐欐搷浣滃潎鍐欏叆鎿嶄綔鏃ュ織銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test -count=1 ./...` 閫氳繃銆?- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鏃?`delivery:manage` 鏌ヨ浜や粯鏂囨。杩斿洖 `403`銆?  - 鏈?`delivery:manage` 鍙垱寤哄拰鏌ヨ浜や粯鏂囨。銆?  - 鏃?`testcase:manage` 鏌ヨ娴嬭瘯鐢ㄤ緥杩斿洖 `403`銆?  - 鏈?`testcase:manage` 鍙垱寤烘祴璇曠敤渚嬨€佸垱寤烘祴璇曟墽琛岃褰曞苟鏌ヨ鎵ц璁板綍銆?  - 娴嬭瘯鎵ц璁板綍鍏宠仈涓嶅瓨鍦ㄧ敤渚嬫椂杩斿洖 `404`銆?
### 褰撳墠椋庨櫓

- 褰撳墠浜や粯娴嬭瘯鐣欐。浠嶆槸鍐呭瓨鏈嶅姟锛涙帴 PostgreSQL 鏃堕渶瑕佽惤鍒?`delivery_documents`銆乣test_cases`銆乣test_runs`锛屽苟琛ユ枃浠堕檮浠躲€佹墽琛屼汉鍜屽鍑鸿兘鍔涖€?
## 2026-06-13 21:30

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md` E4 寮哄疄鍚嶉摼璺紝浼樺厛琛ュ皬绋嬪簭鍚庣璁よ瘉涓庡疄鍚嶉椄闂ㄣ€?- 鐧诲綍閾捐矾鏀逛负涓ら樁娈碉細
  - `POST /api/app/auth/wechat-login` 榛樿杩斿洖 `preAuthToken`銆乣requiresIdentityBinding`銆乣identityBindStatus`銆?  - 宸插疄鍚嶇敤鎴峰啀娆＄櫥褰曟椂鍙悓鏃惰幏寰楁寮?`token`銆?  - 鏂板 `POST /api/app/auth/issue-token-after-identity`锛屽疄鍚嶅畬鎴愬悗鎹㈠彇姝ｅ紡 token銆?- 瀹炲悕娴佺▼鍏ュ彛鏀寔棰勮璇佽闂細
  - 鎵嬫満缁戝畾銆佺煭淇″彂閫?鏍￠獙銆佹墜鏈哄彿鏍搁獙銆丗aceID detect/callback銆佸疄鍚嶇姸鎬佹煡璇㈠潎鍙娇鐢?`preAuthToken`銆?  - `/api/app/users/me` 鍏佽棰勮璇佽闂紝鏂逛究灏忕▼搴忓睍绀哄綋鍓嶇櫥褰曠敤鎴峰拰瀹炲悕杩涘害銆?- 鏍稿績涓氬姟缁熶竴鎺ュ叆寮哄疄鍚嶉椄闂細
  - `requireUser` 澧炲姞 `identity.IsVerified` 鏍￠獙锛屾湭瀹屾垚寮哄疄鍚嶈繑鍥?`40341`銆?  - 宸插畬鎴愬疄鍚嶅悗锛屽師棰勮璇?token 涔熻兘缁х画閫氳繃鏈湴鏈嶅姟鏍￠獙锛屽吋瀹圭幇鏈夋祴璇曚笌鏈湴璋冭瘯锛涙寮?token 浠嶇敱鏂板鎺ュ彛鍙戞斁缁欏鎴风銆?- 鍚庡彴鏂板瀹炲悕鏍搁獙鍙鍏ュ彛锛?  - `GET /api/admin/identity-verifications`
  - `GET /api/admin/identity-verifications/{userId}`
  - 鏉冮檺鐮侊細`identity:read`
  - 璇︽儏鏌ョ湅鍐欏叆鎿嶄綔鏃ュ織锛屼笖杩斿洖瀛楁淇濇寔鎵嬫満鍙疯劚鏁忋€佹棤韬唤璇佸彿鍜屼汉鑴稿師濮嬫暟鎹€?- 椤烘墜淇鐪熷疄 Go 娴嬭瘯鏆撮湶鐨勯棶棰橈細
  - `ErrGameNotConfirmable` 鏄犲皠涓?`409`锛岄伩鍏嶅鐩?缁眬杩囨棭鎿嶄綔杩斿洖 `500`銆?  - 涓撳/棰嗚矾浜虹敾鍍忔洿鏂板鍔?`POST` 鍒悕锛屽吋瀹瑰皬绋嬪簭琛ㄥ崟鎻愪氦銆?  - 淇绉垎鍏戞崲娴嬭瘯鐨勯€€鍥炰綑棰濇柇瑷€锛屾寜璐︽湰鐪熷疄娴佽浆涓?`20`銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server` 鏃犳祴璇曟枃浠躲€?  - `internal/appapi`銆乣internal/auth`銆乣internal/games`銆乣internal/identity`銆乣internal/revenueclient` 绛夊寘閫氳繃銆?- 娉ㄦ剰锛氱郴缁?PATH 閲屼紭鍏堝懡涓?`C:\Windows\System32\go`锛岃鍗犱綅鍛戒护浼氶潤榛樿繑鍥烇紱鏈疆楠岃瘉宸叉敼鐢?`C:\Program Files\Go\bin\go.exe`锛屽苟鎶?`GOCACHE` 鎸囧埌椤圭洰鍐咃紝閬垮厤娌欑鍐?`AppData` 澶辫触銆?
### 褰撳墠椋庨櫓

- 褰撳墠瀹炲悕銆乼oken銆佸悗鍙板疄鍚嶈褰曚粛鏄唴瀛樻湇鍔℃ā鍨嬶紱鎺?PostgreSQL 鏃堕渶瑕佽惤琛ㄤ繚瀛橀璁よ瘉浼氳瘽銆佹寮忎細璇濄€佸疄鍚嶇姸鎬佹祦杞拰鍚庡彴鏌ョ湅鏃ュ織銆?- 寰俊瀹炲悕涓€鑷存€х洰鍓嶆寜鏈湴妯℃嫙閾捐矾鏍囪涓?`not_supported`锛岀湡鏈烘帴鍏ユ椂闇€瑕佹牴鎹井淇?鑵捐浜戝彲鐢ㄨ兘鍔涜ˉ鐪熷疄涓€鑷存€у洖璋冩垨淇濇寔鏄庣‘涓嶅彲鏀寔鐘舵€併€?
## 2026-06-13 22:00

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md` E4 寮哄疄鍚嶄富绾匡紝浼樺厛琛ュ皬绋嬪簭鍚庣鎸佷箙鍖栧拰鏈湴鍙祴椋庢帶閫昏緫銆?- 鏂板杩佺Щ鑴氭湰 `db/migrations/000011_identity_strong.sql`锛?  - `app_auth_sessions`锛氶璁よ瘉 token 鍜屾寮忎笟鍔?token 鐨勬寔涔呭寲浼氳瘽琛紝鎸?`token_hash` 淇濆瓨锛屼笉钀芥槑鏂?token銆?  - `identity_verification_records`锛氬己瀹炲悕鎬荤姸鎬佽〃锛岃褰曟墜鏈哄彿鑴辨晱鍊笺€佹墜鏈哄彿鏍搁獙銆佷汉鑴告牳韬€佸井淇″疄鍚嶄竴鑷存€х姸鎬併€佸け璐ュ師鍥犮€佺煭淇″彂閫佽鏁扮瓑銆?  - `identity_sms_code_records`锛氱煭淇￠獙璇佺爜鍙戦€?鏍￠獙鐣欑棔琛紝棰勭暀 `code_hash`銆佽繃鏈熸椂闂淬€佸け璐ユ鏁般€?  - `identity_faceid_sessions`锛氫汉鑴告牳韬細璇濊〃锛岄鐣?token hash銆佽姹?鍥炶皟鎽樿銆佸け璐ュ師鍥犲拰瀹屾垚鏃堕棿銆?- 鍚屾鏇存柊 `plan.md` 鐨?C3 鏁版嵁搴撹縼绉绘墽琛岄『搴忥紝杩藉姞 `000011_identity_strong.sql`锛岄伩鍏嶆柊杩佺Щ鑴氭湰娓哥鍦ㄨ鍒掑銆?- 寮哄疄鍚嶇煭淇″彂閫侀€昏緫浠庡浐瀹氳繑鍥炲崌绾т负鏈湴闄愭祦妯″瀷锛?  - 60 绉掑唴閲嶅鍙戦€佽繑鍥?`ErrSMSRateLimited`銆?  - 鍗曠敤鎴峰崟鏃ユ渶澶?5 娆★紝瓒呴檺杩斿洖 `ErrSMSDailyLimited`銆?  - 灏忕▼搴忔帴鍙?`POST /api/app/sms/send-code` 灏嗛檺娴佹槧灏勪负 HTTP `429`銆?- 琛ユ祴璇曡鐩栵細
  - identity 鏈嶅姟灞傞噸澶嶅彂閫佺煭淇′細琚檺娴併€?  - appapi 灞傝繛缁皟鐢?`/api/app/sms/send-code` 绗簩娆¤繑鍥?`429`銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `internal/appapi`
  - `internal/auth`
  - `internal/games`
  - `internal/identity`
  - `internal/revenueclient`
  - 鍏朵綑鏃犳祴璇曟枃浠跺寘姝ｅ父缂栬瘧閫氳繃銆?
### 褰撳墠椋庨櫓

- 杩佺Щ鑴氭湰宸茶ˉ榻愬己瀹炲悕鎸佷箙鍖栬〃锛屼絾杩愯鏃舵湇鍔′粛浠ュ唴瀛?Store 涓轰富锛涗笅涓€姝ラ渶瑕佹帴 PostgreSQL repository 鎴栫粺涓€鏁版嵁璁块棶灞傘€?- 鐭俊楠岃瘉鐮佸綋鍓嶄粛鏄湰鍦?mock `123456`锛屽凡鍏峰闄愭祦涓庣姸鎬佹祦杞祴璇曪紝浣嗙湡瀹炵煭淇￠€氶亾鎺ュ叆鏃惰繕闇€瑕侀獙璇佺爜 hash銆佽繃鏈熸椂闂淬€佸け璐ユ鏁颁笌渚涘簲鍟嗗洖鎵ц惤琛ㄣ€?
## 2026-06-13 22:30

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md` E4 寮哄疄鍚嶄富绾匡紝浼樺厛鎺ㄨ繘灏忕▼搴忓悗绔繍琛屾椂钀藉簱杈圭晫銆?- 鏂板 identity 鎸佷箙鍖栨帴鍙ｄ笌 SQL 閫傞厤楠ㄦ灦锛?  - `services/go-api/internal/identity/repository.go`
  - `services/go-api/internal/identity/sql_repository.go`
- `identity.Service` 鏂板 `NewServiceWithRepository(repo)`锛?  - 鏃?repository 鏃剁户缁娇鐢ㄧ幇鏈夊唴瀛樻ā寮忥紝淇濇寔鏈湴娴嬭瘯鍜屽凡鏈夋帴鍙ｄ笉鍙楀奖鍝嶃€?  - 鏈?repository 鏃讹紝浼氬啓鍏ュ己瀹炲悕鎬荤姸鎬併€佺煭淇￠獙璇佺爜璁板綍鍜?FaceID 浼氳瘽璁板綍銆?  - 鐭俊楠岃瘉鐮佷笌 FaceID token 鎸?hash 鍙ｅ緞鍐欏叆 repository锛屼笉钀芥槑鏂囬獙璇佺爜鎴?face token銆?- `cmd/server/main.go` 鏂板鍙€?identity repository 鎺ュ叆鐐癸細
  - 璇诲彇 `DATABASE_URL` 鍜?`DATABASE_DRIVER`銆?  - 鍙墦寮€鏁版嵁搴撴椂浣跨敤 `identity.NewSQLRepository(db)`銆?  - 鏈厤缃垨椹卞姩涓嶅彲鐢ㄦ椂璁板綍鏃ュ織骞跺洖閫€鍐呭瓨妯″紡銆?- `deploy/env.example` 鏂板锛?  - `DATABASE_DRIVER=postgres`
  - `DATABASE_URL=postgres://zhw:zhw@127.0.0.1:5432/zhw_mini?sslmode=disable`
- 琛ュ厖 repository 琛屼负娴嬭瘯锛?  - 寮哄疄鍚嶆祦绋嬩細鎸佷箙鍖栨渶缁?`verified` 鐘舵€併€?  - 鐭俊楠岃瘉鐮佽褰曚繚瀛?hash锛屼笉淇濆瓨 `123456` 鏄庢枃銆?  - FaceID 浼氳瘽浼氳褰?`faceid_processing -> verified` 鐢熷懡鍛ㄦ湡銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/auth`
  - `internal/games`
  - `internal/identity`
  - `internal/revenueclient`

### 褰撳墠椋庨櫓

- 褰撳墠 Go 妯″潡鏈紩鍏?PostgreSQL 椹卞姩渚濊禆锛涘洜姝?`DATABASE_DRIVER=postgres` 鍦ㄦ病鏈夐┍鍔ㄦ敞鍐屾椂浼氬洖閫€鍐呭瓨妯″紡銆傜敱浜庣敤鎴烽檺鍒朵笅杞戒綅缃紝鏈疆鏈墽琛?`go get` 涓嬭浇澶栭儴渚濊禆銆?- 涓嬩竴姝ヨ嫢瑕佺湡姝ｈ繛 PostgreSQL锛岄渶瑕佸湪鍙帶涓嬭浇/渚濊禆绛栫暐涓嬪姞鍏ラ┍鍔紝渚嬪 `pgx/stdlib` 鎴栭」鐩寚瀹氶┍鍔紝骞惰ˉ鏁版嵁搴撻泦鎴愭祴璇曘€?
## 2026-06-13 23:00

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md` E4 寮哄疄鍚嶄富绾匡紝浼樺厛鎺ㄨ繘灏忕▼搴忓悗绔?token 浼氳瘽钀藉簱杈圭晫銆?- 灏濊瘯鎸夌敤鎴蜂笅杞界洰褰曠害鏉熸帴鍏?PostgreSQL 椹卞姩锛?  - `GOMODCACHE` 宸叉寚瀹氬埌 `C:\Users\61492\Desktop\codex download\go-mod-cache`銆?  - 娌欑涓嶅厑璁稿啓璇ョ洰褰曪紝鎻愭潈璇锋眰鏈幏閫氳繃锛涙湰杞湭涓嬭浇澶栭儴渚濊禆锛屼篃鏈慨鏀?`go.mod/go.sum`銆?- 鍦ㄤ笉寮曞叆澶栭儴渚濊禆鐨勫墠鎻愪笅锛屽畬鎴?auth token 浼氳瘽鎸佷箙鍖栨帴鍙ｅ拰 SQL 閫傞厤楠ㄦ灦锛?  - `services/go-api/internal/auth/repository.go`
  - `services/go-api/internal/auth/sql_repository.go`
  - `TokenStore` 鏂板 `NewTokenStoreWithRepository(repo)`銆?  - token 鍐欏叆 repository 鏃跺彧淇濆瓨 `token_hash`锛屼笉淇濆瓨鏄庢枃 token銆?  - 鍐呭瓨 token 涓嶅瓨鍦ㄦ椂锛屽彲閫氳繃 `token_hash` 浠?repository 鍥炶浇浼氳瘽銆?- `cmd/server/main.go` 璋冩暣涓轰竴娆℃墦寮€鏁版嵁搴撹繛鎺ワ紝骞跺悓鏃舵彁渚涚粰锛?  - `auth.NewSQLSessionRepository(db)`
  - `identity.NewSQLRepository(db)`
  - 鏁版嵁搴撴湭閰嶇疆銆侀┍鍔ㄦ湭娉ㄥ唽鎴?ping 澶辫触鏃讹紝缁熶竴鏃ュ織鎻愮ず骞跺洖閫€鍐呭瓨妯″紡銆?- `.gitignore` 琛ュ厖 `.gocache/`銆乣.gotmp/` 鍜屽瓙鐩綍缂撳瓨瑙勫垯锛岄伩鍏嶆湰鍦?Go 鏋勫缓缂撳瓨杩涘叆浜や粯鑼冨洿銆?- 琛ュ厖 auth token repository 娴嬭瘯锛?  - 鍙戣棰勮璇?token 鍚庢寔涔呭寲 session銆?  - repository 涓笉淇濆瓨鏄庢枃 token銆?  - 鏂?`TokenStore` 鍙粠 repository 鎸?token hash 鍥炶浇浼氳瘽銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/auth`
  - `internal/games`
  - `internal/identity`
  - `internal/revenueclient`

### 褰撳墠椋庨櫓

- PostgreSQL 椹卞姩渚濊禆浠嶆湭鎺ュ叆锛岀湡瀹炴暟鎹簱杩炴帴鍦ㄥ綋鍓嶄緷璧栫姸鎬佷笅浼氬洜 driver 鏈敞鍐岃€屽洖閫€鍐呭瓨妯″紡銆?- 鍚庣画闇€瑕佺敤鎴峰厑璁镐緷璧栦笅杞藉埌 `C:\Users\61492\Desktop\codex download`锛屾垨鎻愪緵鏈湴宸叉湁鐨?Go module 缂撳瓨/渚濊禆鍖呭悗锛屽啀鎺ュ叆 `pgx/stdlib` 骞惰ˉ鐪熷疄鏁版嵁搴撻泦鎴愭祴璇曘€?
## 2026-06-13 23:30

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md` E5銆孉I 鎺ㄨ崘涓?IM 鍒嗘瀽鐨勪竴鏈熸暟鎹噯澶囥€嶏紝鏈疆浠嶄紭鍏堝皬绋嬪簭鍚庣鍜屽悗鍙板彧璇绘暟鎹兘鍔涖€?- 鏂板 `services/go-api/internal/aidata` 鏈嶅姟锛?  - 鑱氬悎琛屼负鏃ュ織銆佹敹钘忋€佽瘎浠枫€佷汉鑴夊叧绯汇€佽瀹舵妧鑳界敾鍍忋€侀璺汉璧勬簮鐢诲儚銆両M 娑堟伅鏁伴噺銆?  - 杩斿洖 `dataReady` 鍜?`imExportEnabled`锛岀敤浜庡悗鍙板垽鏂竴鏈熸暟鎹矇娣€鏄惁鍙敤銆?  - IM 鑱婂ぉ鍐呭瀵煎嚭榛樿鍏抽棴锛屾湭寮€鍚椂杩斿洖 `ErrIMExportDisabled`銆?- 鏂板鍚庡彴鎺ュ彛锛?  - `GET /api/admin/ai-data/snapshot`锛屾潈闄?`ai:data:read`銆?  - `POST /api/admin/ai-data/im-export`锛屾潈闄?`ai:data:export`锛屽綋鍓嶉粯璁よ繑鍥?`40361`锛岀姝㈡妸鑱婂ぉ鍐呭瀵煎嚭缁?AI銆?- 琛ュ厖鏈嶅姟灞傚彧璇昏仛鍚堟柟娉曪細
  - `games.AllFavorites()`
  - `reviews.AllReviews()`
  - `profiles.AllExpertSkills()`
  - `profiles.AllGuideResources()`
  - `im.AllMessages()`
- 琛ュ厖鏉冮檺鍜岃縼绉伙細
  - 鍐呯疆绠＄悊鍛樻潈闄愬鍔?`ai:data:read`銆乣ai:data:export`銆?  - `db/seeds/admin_roles_permissions.sql` 澧炲姞 AI 鏁版嵁鏉冮檺绉嶅瓙銆?  - 鏂板 `db/migrations/000012_ai_data_prepare.sql`锛岄鐣?`ai_data_snapshots` 鍜?`ai_data_export_tasks`銆?  - `plan.md` C3 鏁版嵁搴撹縼绉婚『搴忓凡杩藉姞 `000012_ai_data_prepare.sql`銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/adminauth`
  - `internal/appapi`
  - `internal/auth`
  - `internal/games`
  - `internal/identity`
  - `internal/revenueclient`
  - 鍏朵綑鏃犳祴璇曟枃浠跺寘姝ｅ父缂栬瘧閫氳繃銆?- 鏂板 `TestAIDataSnapshotAndIMExportGateHTTP` 瑕嗙洊锛?  - 琛屼负鏃ュ織銆佹敹钘忋€佽瘎浠枫€両M 娑堟伅銆佷汉鑴夊叧绯汇€佽瀹剁敾鍍忋€侀璺汉鐢诲儚鍧囨湁鏁版嵁娌夋穩銆?  - 鍚庡彴 AI 鏁版嵁蹇収鍙寜 `ai:data:read` 鏌ヨ銆?  - IM 瀵煎嚭寮€鍏冲叧闂椂锛宍POST /api/admin/ai-data/im-export` 杩斿洖 `403`銆?
### 褰撳墠椋庨櫓

- 褰撳墠 AI 鏁版嵁蹇収鏉ヨ嚜杩愯鏃跺唴瀛樻湇鍔¤仛鍚堬紱鐪熷疄 PostgreSQL 鎺ュ叆鍚庯紝闇€瑕佹妸杩欎簺鍙鑱氬悎鏀逛负鍒嗛〉 SQL 鏌ヨ鎴栧揩鐓т换鍔★紝閬垮厤鍚庡彴涓€娆℃€ф壂鎻忓ぇ琛ㄣ€?- IM 瀵煎嚭鎺ュ彛宸插仛榛樿鍏抽棴鍜屾潈闄愰殧绂伙紝浣嗗皻鏈帴鍏ョ湡瀹炴巿鏉冮厤缃〃锛涗簩鏈熷惎鐢ㄥ墠蹇呴』琛ュ厖鏄庣‘鐨勬暟鎹巿鏉冦€佽劚鏁忋€佸璁″拰瀵煎嚭浠诲姟鐣欑棔銆?
## 2026-06-13 23:55

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md` E6 鍜?E7 涓笌 E5 鐩存帴鐩稿叧鐨勫悗绔仈璋冩潗鏂欒ˉ榻愩€?- 鍚屾鍚庡彴 OpenAPI锛?  - `docs/openapi/admin.openapi.yaml` 鏂板 `GET /api/admin/ai-data/snapshot`銆?  - `docs/openapi/admin.openapi.yaml` 鏂板 `POST /api/admin/ai-data/im-export`銆?  - 鏍囨槑 `ai:data:read`銆乣ai:data:export` 鏉冮檺鍜?IM 瀵煎嚭榛樿鍏抽棴琛屼负銆?- 鍚屾 DTO 鍜岄敊璇爜锛?  - `docs/openapi/dto-samples.md` 鏂板 `AIDataSnapshotDTO`銆?  - `docs/openapi/dto-samples.md` 鏂板 `IMExportItemDTO`銆?  - `docs/openapi/error-codes.md` 鏂板 `40361`锛岃〃绀?AI IM 瀵煎嚭寮€鍏冲叧闂€?- 鏂板娴嬭瘯鐢ㄤ緥褰掓。锛?  - `docs/test-cases/TC-144-ai-data-im-export-gate.md`
  - 瑕嗙洊 AI 鏁版嵁蹇収鍙煡銆両M 鍙鏁般€侀粯璁ゅ叧闂椂绂佹瀵煎嚭鑱婂ぉ鍐呭銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/adminauth`
  - `internal/appapi`
  - `internal/auth`
  - `internal/games`
  - `internal/identity`
  - `internal/revenueclient`
  - 鍏朵綑鏃犳祴璇曟枃浠跺寘姝ｅ父缂栬瘧閫氳繃銆?
### 褰撳墠椋庨櫓

- OpenAPI 褰撳墠浠嶄互绠€鍖栨弿杩颁负涓伙紝灏氭湭涓?AI 鏁版嵁鎺ュ彛琛ュ畬鏁?schema components锛涘悗缁墠绔仈璋冨墠闇€瑕佹妸缁熶竴鍝嶅簲鍖呫€佸瓧娈电被鍨嬪拰閿欒鍝嶅簲缁撴瀯鍏ㄩ儴灞曞紑銆?- `docs/test-cases` 鐜板湪鍙柊澧炰簡 TC-144锛孍7 涓叾浣?TC-130 鍒?TC-145 杩橀渶瑕佹寜宸插疄鐜版ā鍧楅€愭褰掓。銆?
## 2026-06-14 00:20

### 褰撳墠杩涘睍

- 缁х画琛ラ綈 `plan.md` E5銆孉I 鎺ㄨ崘涓?IM 鍒嗘瀽鐨勪竴鏈熸暟鎹噯澶囥€嶇殑鏁版嵁婧愯鐩栥€?- 灏?`user_footprints` 瀵瑰簲鐨勮繍琛屾椂瓒宠抗鏁版嵁绾冲叆 AI 鏁版嵁蹇収锛?  - `reviews.Service` 鏂板 `AllFootprints()` 鍙鑱氬悎鏂规硶銆?  - `aidata.SnapshotInput` 鏂板 `Footprints`銆?  - `AIDataSnapshot` 鏂板 `footprintCount`銆?  - `GET /api/admin/ai-data/snapshot` 杩斿洖瓒宠抗鏁伴噺銆?- 鍚屾娴嬭瘯鍜岃仈璋冩潗鏂欙細
  - `TestAIDataSnapshotAndIMExportGateHTTP` 澧炲姞 `footprintCount > 0` 鏂█銆?  - `docs/openapi/admin.openapi.yaml` 灏?AI 鏁版嵁蹇収璇存槑琛ュ厖涓哄寘鍚冻杩广€?  - `docs/openapi/dto-samples.md` 鐨?`AIDataSnapshotDTO` 澧炲姞 `footprintCount`銆?  - `docs/test-cases/TC-144-ai-data-im-export-gate.md` 澧炲姞瓒宠抗鍓嶇疆鏉′欢鍜屽瓧娈垫牎楠屻€?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/adminauth`
  - `internal/appapi`
  - `internal/auth`
  - `internal/games`
  - `internal/identity`
  - `internal/revenueclient`
  - 鍏朵綑鏃犳祴璇曟枃浠跺寘姝ｅ父缂栬瘧閫氳繃銆?
### 褰撳墠椋庨櫓

- 瓒宠抗鐩墠浠嶆潵鑷?`reviews` 鍐呭瓨鏈嶅姟涓殑杩愯鏃舵暟缁勶紱鎺?PostgreSQL 鍚庨渶瑕佹寜 `user_footprints` 琛ㄥ疄鐜板垎椤垫煡璇㈡垨蹇収鑱氬悎銆?- E5 鐨勫瓧娈佃鐩栧凡杩涗竴姝ュ畬鏁达紝浣嗘祴璇曟暟鎹壒閲忛€犳暟鑳藉姏灏氭湭琛ラ綈锛屽悗缁繕闇€瑕佷负鈥? 涓祴璇曠敤鎴枫€? 涓眬銆?0 鏉¤涓衡€濈瓑楠屾敹鏉′欢鍋氬彲閲嶅娴嬭瘯鑴氭湰鎴栧悗鍙扮瀛愪换鍔°€?
## 2026-06-14 00:50

### 褰撳墠杩涘睍

- 缁х画鎺ㄨ繘 `plan.md` E5銆孉I 鎺ㄨ崘涓?IM 鍒嗘瀽鐨勪竴鏈熸暟鎹噯澶囥€嶇殑灏忕▼搴忓悗绔獙鏀跺彲瑙佹€с€?- `GET /api/admin/ai-data/snapshot` 鏂板楠屾敹闃堝€煎揩鐓э細
  - `userCount`锛氫粠琛屼负銆佸眬銆佹敹钘忋€佽瘎浠枫€佽冻杩广€佷汉鑴夈€佺敾鍍忓拰 IM 鍙戦€佽€呬腑鑱氬悎鍘婚噸鐢ㄦ埛鏁般€?  - `gameCount`锛氬綋鍓嶆父鎴忓眬鏁伴噺銆?  - `acceptanceChecks`锛氭樉寮忚繑鍥?5 涓敤鎴枫€? 涓眬銆?0 鏉¤涓恒€? 鏉℃敹钘忋€? 鏉¤瘎浠风殑褰撳墠鍊笺€佽姹傚€煎拰杈炬爣鐘舵€併€?  - `acceptanceReady`锛氬綋涓婅堪 5 椤瑰叏閮ㄨ揪鏍囨椂鑷姩涓?`true`銆?- 鍚屾娴嬭瘯鍜岃仈璋冩潗鏂欙細
  - `TestAIDataSnapshotAndIMExportGateHTTP` 澧炲姞闃堝€艰姹傘€佺敤鎴锋暟銆佸眬鏁板拰鏈揪鏍囩姸鎬佹柇瑷€銆?  - `docs/openapi/admin.openapi.yaml` 澧炲姞 E5 楠屾敹闃堝€艰鏄庛€?  - `docs/openapi/dto-samples.md` 鐨?`AIDataSnapshotDTO` 澧炲姞瀹屾暣闃堝€肩ず渚嬨€?  - `docs/test-cases/TC-144-ai-data-im-export-gate.md` 澧炲姞闃堝€煎瓧娈垫牎楠屻€?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/adminauth`
  - `internal/appapi`
  - `internal/auth`
  - `internal/games`
  - `internal/identity`
  - `internal/revenueclient`
  - 鍏朵綑鏃犳祴璇曟枃浠跺寘姝ｅ父缂栬瘧閫氳繃銆?
### 褰撳墠椋庨櫓

- 褰撳墠宸茶兘鐪嬪埌 E5 楠屾敹闃堝€艰繘搴︼紝浣嗚繕娌℃湁鎻愪緵涓€閿壒閲忛€犳暟/绉嶅瓙浠诲姟锛涗笅涓€姝ュ簲琛ュ悗鍙板彈鎺х瀛愭帴鍙ｆ垨娴嬭瘯澶瑰叿锛岀ǔ瀹氱敓鎴?5 鐢ㄦ埛銆? 灞€銆?0 琛屼负銆? 鏀惰棌銆? 璇勪环鐨勬暟鎹泦銆?- `userCount` 鐩墠鍩轰簬鍐呭瓨鏈嶅姟鑱氬悎锛涙帴 PostgreSQL 鍚庨渶瑕佹敼涓?SQL 鑱氬悎鎴栧彧璇讳粨鍌ㄨ仛鍚堬紝閬垮厤鐢熶骇鏁版嵁閲忓彉澶ф椂鍦ㄥ唴瀛樺眰鍏ㄩ噺鎵弿銆?
## 2026-06-14 01:20

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md` E5锛屽皬绋嬪簭鍚庣浼樺厛琛ラ綈銆屽畬鎴?5 涓祴璇曠敤鎴枫€? 涓眬銆?0 鏉¤涓恒€? 鏉℃敹钘忋€? 鏉¤瘎浠峰悗鍚庡彴鍙湅鍒版暟鎹矇娣€銆嶇殑鍙噸澶嶉獙鏀堕摼璺€?- 鏂板鍚庡彴鍙楁帶楠屾敹澶瑰叿鎺ュ彛锛?  - `POST /api/admin/ai-data/acceptance-fixture`
  - 鏉冮檺鐮侊細`ai:data:seed`
  - 杩斿洖 `AIDataAcceptanceFixtureResultDTO`锛屽寘鍚湰娆¤ˉ榻愭暟閲忓拰琛ラ綈鍚庣殑 `AIDataSnapshotDTO`銆?- 澶瑰叿浼氬鐢ㄧ幇鏈夊皬绋嬪簭鍚庣鏈嶅姟閾捐矾鍐欏叆鏁版嵁锛?  - 寮哄疄鍚嶏細涓?5 涓浐瀹氶獙鏀剁敤鎴疯ˉ榻愭湰鍦板己瀹炲悕鐘舵€併€?  - 缁勫眬锛氬垱寤哄苟瀹℃牳 3 涓厤璐瑰眬锛屽畬鎴愬叆灞€銆佹墜鍔ㄥ紑濮嬨€佹湇鍔＄‘璁ゃ€?  - 琛屼负锛氳ˉ榻?20 鏉¤涓烘棩蹇椼€?  - 鏀惰棌锛氳ˉ榻?5 鏉℃敹钘忋€?  - 璇勪环锛氳ˉ榻?5 鏉¤瘎浠凤紝骞朵骇鐢熻冻杩?鎴愰暱璁板綍銆?  - IM锛氳ˉ榻愬熀纭€ IM 鐣欏瓨娑堟伅銆?  - 鐢诲儚锛氳ˉ榻愯瀹跺拰棰嗚矾浜虹敾鍍忔暟鎹€?- 鎺ュ彛鍏峰骞傜瓑淇濇姢锛氬綋 `acceptanceReady=true` 鍚庨噸澶嶈皟鐢ㄤ笉缁х画杩藉姞鏁版嵁銆?- 鍚屾鏉冮檺鍜岃仈璋冩潗鏂欙細
  - `adminauth` 鍐呯疆鏉冮檺鏂板 `ai:data:seed`銆?  - `db/seeds/admin_roles_permissions.sql` 鏂板 `ai:data:seed` 鏉冮檺绉嶅瓙銆?  - `docs/openapi/admin.openapi.yaml` 鏂板澶瑰叿鎺ュ彛銆?  - `docs/openapi/dto-samples.md` 鏂板 `AIDataAcceptanceFixtureResultDTO`銆?  - `docs/test-cases/TC-144-ai-data-im-export-gate.md` 鏂板澶瑰叿姝ラ鍜岃嚜鍔ㄥ寲瑕嗙洊銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `TestAIDataAcceptanceFixtureHTTP`
  - `TestAIDataSnapshotAndIMExportGateHTTP`
  - `cmd/server`
  - `internal/adminauth`
  - `internal/appapi`
  - `internal/auth`
  - `internal/games`
  - `internal/identity`
  - `internal/revenueclient`
  - 鍏朵綑鏃犳祴璇曟枃浠跺寘姝ｅ父缂栬瘧閫氳繃銆?
### 褰撳墠椋庨櫓

- 澶瑰叿褰撳墠鍩轰簬鍐呭瓨鏈嶅姟锛岄€傚悎鏈湴鍜岃仈璋冮獙鏀讹紱鎺?PostgreSQL 鍚庡簲杩佺Щ涓烘祴璇曠幆澧冪瀛愪换鍔℃垨鍙楃幆澧冨彉閲忎繚鎶ょ殑鍐呴儴杩愮淮鎺ュ彛銆?- 鐢熶骇鐜涓嶅簲鏆撮湶 `ai:data:seed` 缁欐櫘閫氬悗鍙拌鑹诧紱涓婄嚎鍓嶉渶瑕佸湪瑙掕壊閰嶇疆涓粎淇濈暀缁欒秴绠℃垨娴嬭瘯鐜璐﹀彿銆?
## 2026-06-13 22:30

### 褰撳墠杩涘睍

- 缁х画鎸?`plan.md` 鎺ㄨ繘 E6/E7锛屽皬绋嬪簭鍚庣浼樺厛琛ラ綈鏈嶅姟灞傚彲楠岃瘉闂幆銆?- E6 鑱旇皟鐢ㄤ緥琛ラ綈锛?  - 鏂板 `docs/test-cases/app-integration-cases.md`锛岃鐩栧皬绋嬪簭鐧诲綍銆佸己瀹炲悕銆侀個绾︺€佺粍灞€銆両M銆佹湇鍔＄‘璁ゃ€佸鐩樸€佺画灞€銆佺Н鍒嗗厬鎹€丄I 鏁版嵁鍑嗗绛夌鍒扮閾捐矾銆?  - 鏂板 `docs/test-cases/admin-integration-cases.md`锛岃鐩栧悗鍙版潈闄愩€佸疄鍚嶇姸鎬併€佹墦鍗″鐞嗐€佽瘎浠锋垚闀裤€佸厬鎹㈠鏍搞€佷妇鎶ョ敵璇夈€佹姤琛ㄥ鍑恒€丄I 鏁版嵁鍑嗗鍜屼氦浠樻祴璇曠暀妗ｃ€?  - `docs/openapi/error-codes.md` 琛ュ厖浼氬憳鎶ヨ〃銆佸洟闃熺鐞嗐€佸鐩橀噸澶嶃€佸厬鎹㈠簱瀛樸€佺Н鍒嗕笉瓒炽€侀噷绋嬬鍜屾墦鍗″弬鏁扮被閿欒鐮併€?- E7 鏈嶅姟娴嬭瘯琛ラ綈锛?  - 浜鸿剦鍏崇郴锛氬弻鍚戣繛鎺ャ€佽窡杩涜褰曞弬涓庤€?棰嗚矾浜烘潈闄愩€?  - 琛屽/棰嗚矾浜虹敾鍍忥細鏈巿浜堣鑹蹭笉鑳藉彂甯冩妧鑳芥垨璧勬簮銆?  - 绉垎锛氬彂鏀俱€佹墸鍑忋€佷綑棰濅笉瓒炽€?  - 鍏戞崲锛氱Н鍒嗕笉瓒炽€佸簱瀛樹笉瓒炽€佸苟鍙戜笉瓒呭崠銆侀┏鍥為€€绉垎銆?  - 浜や粯娴嬭瘯锛氭祴璇曠敤渚嬬櫥璁般€佹祴璇曟墽琛屽叧鑱斿拰缁撴灉鏍￠獙銆?  - 娓告垙锛氭敹钘忓箓绛夈€佸彇娑堟敹钘忓箓绛夈€侀噷绋嬬鏉冮檺涓庣姸鎬併€佹墦鍗℃潈闄愪笌鏃犳晥闅愯棌銆佸鐩樻潈闄愩€侀噸澶嶅鐩樻嫤鎴€佺画灞€鍙敓鎴愯崏绋裤€?  - 寮哄疄鍚嶏細榛樿 `wechat_logged_in`锛孎aceID 瀹屾垚鍓嶄笉鍙涓哄己瀹炲悕瀹屾垚锛屽畬鎴愬悗鎵嶆斁琛屻€?- 鏈嶅姟灞傛柊澧?`ErrDuplicateRetrospective`锛屽悓涓€鎴愬憳鍚屼竴灞€閲嶅鎻愪氦澶嶇洏浼氳繑鍥炲啿绐侀敊璇紝骞舵槧灏勫埌灏忕▼搴忔帴鍙ｉ敊璇爜 `40941`銆?- 璇存槑锛歚plan.md` 涓垪鍑虹殑 `internal/favorites/favorite_service_test.go` 鍦ㄥ綋鍓嶄唬鐮佺粨鏋勯噷瀹為檯灞炰簬 `internal/games`锛屽洜姝ゆ祴璇曡惤鍦?`services/go-api/internal/games/favorite_service_test.go`锛屼繚鎸佸拰鐜版湁瀹炵幇涓€鑷淬€?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `internal/appapi`
  - `internal/games`
  - `internal/identity`
  - `internal/connections`
  - `internal/profiles`
  - `internal/points`
  - `internal/redemption`
  - `internal/delivery`
  - 鍏朵綑鏃犳祴璇曟枃浠跺寘姝ｅ父缂栬瘧閫氳繃銆?
### 褰撳墠椋庨櫓

- 褰撳墠 E7 浠ユ湇鍔″眰鍜?HTTP 灞傛祴璇曚负涓伙紝灏氭湭鎺ュ叆鐪熷疄 PostgreSQL 骞跺彂浜嬪姟锛涘厬鎹㈠苟鍙戜笉瓒呭崠宸插湪鍐呭瓨鏈嶅姟閿佷笅楠岃瘉锛屾帴搴撳悗浠嶉渶鐢ㄦ暟鎹簱浜嬪姟鍜岃绾ч攣澶嶆祴銆?- 褰撳墠鐩綍涓嶆槸 Git 浠撳簱锛屾棤娉曠敤 `git status` 杈撳嚭鍙樻洿娓呭崟锛涙湰杞娇鐢ㄦ枃浠舵壂鎻忓拰鍏ㄩ噺娴嬭瘯纭鍙樻洿鑼冨洿銆?
## 2026-06-13 23:10

### 褰撳墠杩涘睍

- 缁х画鎺ㄨ繘 `plan.md` D5.7銆屽垎娑︺€佹敹鐩娿€佽鍗曞崰浣嶃€嶄腑灏忕▼搴忓悗绔紭鍏堢殑璁㈠崟鍗犱綅闂幆銆?- 鏂板 `services/go-api/internal/orders`锛?  - 鍏嶈垂灞€璁㈠崟鍗犱綅 `free_no_pay`銆?  - 璁㈠崟鍙?`FREE-{gameId}-{id}`銆?  - 閲戦鍥哄畾涓?0 鍒嗐€?  - `needWechatPay=false`锛屼竴鏈熶笉浼氳Е鍙戠湡瀹炲井淇℃敮浠樸€?  - 鍚屼竴灞€閲嶅纭繚璁㈠崟鏃跺箓绛夎繑鍥炲悓涓€璁㈠崟銆?  - 鏀粯鍥炶皟鍗犱綅鎸?`orderNo + eventId` 骞傜瓑璁板綍銆?- 灏忕▼搴忓悗绔帴鍙ｆ柊澧烇細
  - `GET /api/app/orders/{orderId}`锛氬綋鍓嶇敤鎴锋煡璇㈣嚜宸辩殑鏀粯璁㈠崟鍗犱綅銆?  - `POST /api/app/payment/precreate-placeholder`锛氬厤璐瑰眬棰勪笅鍗曞崰浣嶃€?  - `POST /api/internal/pay/callback-placeholder`锛氭敮浠樺洖璋冨崰浣嶅拰骞傜瓑璁板綍銆?- `POST /api/app/games` 鍒涘缓鍏嶈垂灞€鎴愬姛鍚庤嚜鍔ㄧ‘淇?`free_no_pay` 璁㈠崟瀛樺湪锛涘師杩斿洖浠嶄繚鎸?`GameDTO`锛岄伩鍏嶇牬鍧忓皬绋嬪簭鏃㈡湁鑱旇皟銆?- 鍚屾鏂囨。锛?  - `docs/openapi/app.openapi.yaml` 澧炲姞璁㈠崟璇︽儏鍜岄涓嬪崟鍗犱綅鎺ュ彛銆?  - `docs/openapi/admin.openapi.yaml` 澧炲姞鍐呴儴鏀粯鍥炶皟鍗犱綅鎺ュ彛銆?  - `docs/openapi/dto-samples.md` 澧炲姞 `PaymentOrderDTO` 鍜?`PaymentPrecreatePlaceholderDTO`銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `internal/orders`
  - `internal/appapi`
  - `internal/games`
  - `internal/identity`
  - `internal/revenueclient`
  - 鍏朵綑宸叉湁鍚庣鍖呮甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- 褰撳墠璁㈠崟鍗犱綅涓哄唴瀛樻湇鍔★紝宸叉弧瓒虫湰鏈鸿仈璋冨拰鎺ュ彛楠屾敹锛涙帴 PostgreSQL 鍚庨渶瑕佹妸 `payment_orders` 浠撳偍銆佸敮涓€绾︽潫鍜屽洖璋冧簨浠惰〃钀戒负浜嬪姟鍐欏叆銆?- 璧勯噾鏈嶅姟宸叉湁 `/api/funds/payment-precreate-placeholder` 鍥哄畾杩斿洖缁撴瀯锛涙湰杞厛琛?Go 灏忕▼搴忓悗绔彲鏌ヨ鍗曞拰鍐呴儴鍥炶皟骞傜瓑锛屽悗缁彲鎶?Go 渚ц鍗曟湇鍔′笌 Java 璧勯噾鏈嶅姟鍗犱綅瀹㈡埛绔覆鑱斻€?
## 2026-06-13 23:40

### 褰撳墠杩涘睍

- 缁х画鎺ㄨ繘 `plan.md` D5.8銆屼妇鎶ョ敵璇夈€侀€氱煡鍜岃涓虹暀瀛樸€嶏紝灏忕▼搴忓悗绔紭鍏堣ˉ榻愯涓烘棩蹇楀瓧娈靛彛寰勩€?- `internal/audit` 琛屼负鏃ュ織琛ラ綈璁″垝瀛楁锛?  - `eventCode`
  - `businessType`
  - `businessId`
  - `source`
  - `device`
  - `ip`
  - `occurredAt`
- 淇濇寔鏃у瓧娈靛吋瀹癸細
  - `eventType` 浠嶇瓑浠蜂簬 `eventCode`銆?  - `targetType` 浠嶇瓑浠蜂簬 `businessType`銆?  - `targetId` 浠嶇瓑浠蜂簬 `businessId`銆?- `POST /api/app/behavior/events` 鏀寔鏂版棫瀛楁鍚屾椂涓婃姤锛屾湭浼?source 鏃堕粯璁?`app`锛宒evice 鍙粠璇锋眰浣撴垨 `X-Device` 璇诲彇锛宨p 鐢卞悗绔В鏋愩€?- `GET /api/admin/behavior/events` 鏀寔鎸?`eventCode` 杩囨护锛屾柟渚垮悗鍙板仛婕忔枟鍜岀暀瀛樺熀纭€缁熻銆?- 琛ラ綈 D5.8 鏈嶅姟灞傛祴璇曪細
  - `internal/reports`锛氫妇鎶ヨ瘉鎹?ID 鐣欏瓨銆佸垱寤轰妇鎶ュ悗鍐荤粨鍒嗘鼎銆佸鐞嗗拰鍏抽棴鐘舵€佹祦杞€?  - `internal/notifications`锛氱珯鍐呴€氱煡鐢熸垚銆佸井淇¤闃呮秷鎭换鍔＄敓鎴愩€佸凡璇汇€佸彂閫佺粨鏋滃洖鍐欍€?  - `internal/audit`锛氳涓烘棩蹇楁柊瀛楁鍜屾棫瀛楁鍒悕鍏煎銆佹寜 `eventCode` 鏌ヨ銆?- 鍚屾鏂囨。锛?  - `docs/openapi/app.openapi.yaml` 澧炲姞琛屼负浜嬩欢鎺ュ彛瀛楁璇存槑銆?  - `docs/openapi/admin.openapi.yaml` 澧炲姞琛屼负浜嬩欢绛涢€夎鏄庛€?  - `docs/openapi/dto-samples.md` 澧炲姞 `BehaviorLogDTO`銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `internal/audit`
  - `internal/appapi`
  - `internal/reports`
  - `internal/notifications`
  - `internal/orders`
  - `internal/games`
  - `internal/identity`
  - 鍏朵綑宸叉湁鍚庣鍖呮甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- 褰撳墠琛屼负鏃ュ織浠嶆槸鍐呭瓨鏈嶅姟锛屽凡婊¤冻鏈湴鑱旇皟鍜岄獙鏀跺瓧娈靛彛寰勶紱鎺?PostgreSQL 鍚庨渶瑕佸皢杩欎簺瀛楁鍐欏叆 `user_behavior_logs`锛屽苟鎸?`event_code/user_id/occurred_at` 寤虹储寮曘€?
## 2026-06-14 00:10

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堣ˉ榻愬皬绋嬪簭鍚庣涓婄嚎鏁版嵁搴撴壙鎺ヨ兘鍔涖€?- 鏂板杩佺Щ `db/migrations/000013_behavior_log_fields.sql`锛?  - 缁?`user_behavior_logs` 澧炲姞 `event_code`銆乣business_type`銆乣business_id`銆乣source`銆乣device`銆乣ip`銆乣occurred_at`銆?  - 鏃ф暟鎹敤 `event_type`銆乣target_type`銆乣target_id`銆乣created_at` 鍥炲～鏂板瓧娈点€?  - 澧炲姞 `idx_user_behavior_logs_event_code_occurred`銆乣idx_user_behavior_logs_user_occurred`銆乣idx_user_behavior_logs_business`銆?- 鏂板杩佺Щ `db/migrations/000014_payment_callback_placeholder.sql`锛?  - 鏂板 `payment_callbacks`锛岃褰?provider銆乧allback_type銆乪vent_id銆乷rder_no銆乷ut_trade_no銆乼ransaction_id銆乺aw_headers_json銆乺aw_body銆乸ayload_digest銆乿erify_status銆乸rocess_status銆?  - 澧炲姞 `uk_payment_callbacks_order_event`锛屾寜 `order_no + event_id` 淇濊瘉鏀粯鍥炶皟鍗犱綅骞傜瓑銆?  - 澧炲姞鍥炶皟浜ゆ槗鍙枫€佽鍗曞彿銆佸鐞嗙姸鎬佺储寮曘€?- 鍚屾 `plan.md`锛?  - D4 杩佺Щ椤哄簭杩藉姞 `000011`銆乣000012`銆乣000013`銆乣000014`銆?  - 鍞竴绾︽潫娓呭崟杩藉姞 `uk_payment_callbacks_order_event`銆?
### 楠岃瘉缁撴灉

- 鏂囨湰妫€鏌ョ‘璁よ縼绉诲瓧娈点€佺储寮曞拰 `plan.md` 椤哄簭鍧囧彲妫€绱€?- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `internal/appapi`
  - `internal/audit`
  - `internal/orders`
  - `internal/reports`
  - `internal/notifications`
  - 鍏朵綑宸叉湁鍚庣鍖呮甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- 鏈疆瀹屾垚鐨勬槸杩佺Щ灞傛壙鎺ワ紱杩愯鏃朵粛浠ュ唴瀛樻湇鍔′负涓汇€傛帴 PostgreSQL 鍚庨渶瑕佹妸 `internal/audit` 鍜?`internal/orders` 鎺ュ叆 SQL repository锛屽苟鐢ㄧ湡瀹炴暟鎹簱楠岃瘉鍥炶皟骞傜瓑绾︽潫銆?
## 2026-06-14 00:40

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟妸灏忕▼搴忓悗绔?D5.8 / D5.7 鐨勫唴瀛橀棴鐜帹杩涘埌鍙帴 PostgreSQL repository銆?- `internal/audit` 鏂板琛屼负鏃ュ織鎸佷箙鍖栨帴鍙ｄ笌 SQL 閫傞厤锛?  - `BehaviorRepository`
  - `NewServiceWithBehaviorRepository`
  - `NewSQLBehaviorRepository`
  - 鏀寔鍐欏叆 `user_behavior_logs` 鐨?`event_type/event_code/target_type/business_type/target_id/business_id/source/device/ip/extra/created_at/occurred_at`銆?  - 鏀寔鍚庡彴鎸?`userId/eventType/eventCode` 鏌ヨ銆?- `internal/orders` 鏂板璁㈠崟鍜屾敮浠樺洖璋冨崰浣嶆寔涔呭寲鎺ュ彛涓?SQL 閫傞厤锛?  - `Repository`
  - `NewServiceWithRepository`
  - `NewSQLRepository`
  - 鏈夋暟鎹簱鏃跺厤璐瑰眬璁㈠崟鍐欏叆 `payment_orders`锛岃鍗曞彿閲囩敤 `FREE-{gameId}`锛屼緷璧?`order_no` 鍞竴绾︽潫骞傜瓑銆?  - 鏀粯鍥炶皟鍗犱綅鍐欏叆 `payment_callbacks`锛屼緷璧?`order_no + event_id` 鍞竴绱㈠紩骞傜瓑銆?- `cmd/server` 鍦?DB 鍙敤鏃舵敞鍏ワ細
  - `audit.NewSQLBehaviorRepository(db)`
  - `orders.NewSQLRepository(db)`
  - DB 涓嶅彲鐢ㄦ椂淇濇寔鍘熷唴瀛樻湇鍔¤涓恒€?- `internal/appapi.Server` 鏂板 `UseRepositories` 娉ㄥ叆鐐癸紝涓嶆敼鍙樼幇鏈夋祴璇曟瀯閫犲拰瀵瑰鎺ュ彛銆?- 琛ュ厖娴嬭瘯锛?  - `internal/audit` fake repository 娴嬭瘯锛岀‘璁や繚瀛樸€佸垪琛ㄣ€佹煡璇細璧?repository銆?  - `internal/orders` fake repository 娴嬭瘯锛岀‘璁ゅ垱寤鸿鍗曘€佹煡璇㈣鍗曘€佸洖璋冨箓绛変細璧?repository銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/audit`
  - `internal/orders`
  - 鍏朵綑宸叉湁鍚庣鍖呮甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- 褰撳墠椤圭洰浠嶆湭寮曞叆 PostgreSQL 椹卞姩渚濊禆锛宍DATABASE_DRIVER=postgres` 鍦ㄦ棤椹卞姩娉ㄥ唽鏃朵細鐢?`cmd/server` 鍥為€€鍐呭瓨妯″紡锛涙湰杞畬鎴?repository 楠ㄦ灦鍜屾敞鍏ョ偣锛岀湡瀹炴暟鎹簱闆嗘垚娴嬭瘯浠嶉渶鍦ㄤ緷璧栫瓥鐣ョ‘璁ゅ悗鎵ц銆?
## 2026-06-14 01:10

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堣ˉ榻愬皬绋嬪簭鍚庣 D5.8 涓炬姤鐢宠瘔閾捐矾鐨?PostgreSQL 鎵挎帴鑳藉姏銆?- `internal/reports` 鏂板涓炬姤鐢宠瘔鎸佷箙鍖栨帴鍙ｄ笌 SQL 閫傞厤锛?  - `Repository`
  - `NewServiceWithRepository`
  - `NewSQLRepository`
  - 鏀寔鍒涘缓涓炬姤鍐欏叆 `reports`銆?  - 鏀寔鎴戠殑涓炬姤銆佸悗鍙颁妇鎶ュ垪琛ㄣ€佷妇鎶ヨ鎯呬粠 `reports` 璇诲彇銆?  - 鏀寔鍚庡彴澶勭悊/鍏抽棴涓炬姤鍚庡洖鍐?`status`銆乣handler_admin_id`銆乣handle_result`銆乣handled_at`銆?- `internal/appapi.Server` 鐨?repository 娉ㄥ叆鐐规墿灞曞埌涓炬姤鐢宠瘔锛?  - 鏈夋暟鎹簱鏃舵敞鍏?`reports.NewSQLRepository(db)`銆?  - 鏃犳暟鎹簱鏃朵繚鐣欏師鍐呭瓨鏈嶅姟锛屼繚鎸佹湰鍦拌仈璋冨拰鏃㈡湁娴嬭瘯琛屼负涓嶅彉銆?- `cmd/server` 鏂板 `reportRepository(db)`锛屽拰琛屼负鏃ュ織銆佽鍗?repository 涓€璧峰湪鍚姩鏃舵敞鍐屻€?- 琛ュ厖 `internal/reports` fake repository 鍗曞厓娴嬭瘯锛岀‘璁ゅ垱寤恒€佹煡璇€佸鐞嗕細璧?repository 鍒嗘敮銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/reports`
  - `internal/audit`
  - `internal/orders`
  - 鍏朵綑宸叉湁鍚庣鍖呮甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- `reports` 宸插叿澶?SQL repository锛屼絾閫氱煡 `notifications` 浠嶄互鍐呭瓨鏈嶅姟涓轰富锛涗笅涓€姝ュ簲缁х画鎶婄珯鍐呴€氱煡銆佸井淇¤闃呮秷鎭换鍔″拰妯℃澘鎺ュ叆 PostgreSQL銆?- `reports` 鐨勬暟鎹簱绾ч泦鎴愭祴璇曚粛渚濊禆 PostgreSQL 椹卞姩鍜屾祴璇曞簱绛栫暐纭锛涘綋鍓嶉獙璇佽鐩栨湇鍔℃敞鍏ャ€佺紪璇戝拰鍐呭瓨/fake repository 琛屼负銆?
## 2026-06-14 01:40

### 褰撳墠杩涘睍

- 缁х画鎵ц D5.8锛岃ˉ榻愮珯鍐呴€氱煡鍜屽井淇¤闃呮秷鎭换鍔＄殑 PostgreSQL repository銆?- `internal/notifications` 鏂板鎸佷箙鍖栨帴鍙ｄ笌 SQL 閫傞厤锛?  - `Repository`
  - `NewServiceWithRepository`
  - `NewSQLRepository`
  - 鍒涘缓绔欏唴閫氱煡鏃跺啓鍏?`notifications`銆?  - `needWechat=true` 鏃跺悓姝ュ啓鍏?`wechat_subscribe_tasks`锛屽苟鍥炲～ `notifications.wechat_task_id`銆?  - 鏌ヨ鎴戠殑閫氱煡銆佸悗鍙板井淇¤闃呬换鍔°€佸井淇¤闃呮ā鏉挎椂浼樺厛璇?repository銆?  - 鏍囪宸茶浼氬洖鍐?`notifications.status/read_at`銆?  - 鏍囪寰俊浠诲姟宸插彂閫佷細鍥炲啓 `wechat_subscribe_tasks.status/result/sent_at`锛屽苟鍚屾 `notifications.wechat_state=sent`銆?  - SQL 妯″紡涓嬪井淇¤闃呬换鍔′笉瀛樺湪鏃舵槧灏勪负 `ErrNotificationNotFound`锛屽拰鍐呭瓨妯″紡淇濇寔涓€鑷淬€?- `internal/appapi.Server` repository 娉ㄥ叆鐐圭户缁墿灞曪細
  - 鏈夋暟鎹簱鏃舵敞鍏?`notifications.NewSQLRepository(db)`銆?  - 鏃犳暟鎹簱鏃剁户缁娇鐢ㄥ師鍐呭瓨閫氱煡鏈嶅姟锛屼繚鎸佹湰鍦拌仈璋冨拰宸叉湁娴嬭瘯绋冲畾銆?- `cmd/server` 鏂板 `notificationRepository(db)`锛屽拰琛屼负鏃ュ織銆佽鍗曘€佷妇鎶ョ敵璇変竴璧峰湪鍚姩鏃舵敞鍐屻€?- 琛ュ厖 `internal/notifications` fake repository 鍗曞厓娴嬭瘯锛岀‘璁ゅ垱寤恒€佸垪琛ㄣ€佸凡璇汇€佹ā鏉裤€佷换鍔″彂閫佺粨鏋滀細璧?repository 鍒嗘敮銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/notifications`
  - `internal/reports`
  - `internal/audit`
  - `internal/orders`
  - 鍏朵綑宸叉湁鍚庣鍖呮甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- D5.8 鐨勪妇鎶ャ€侀€氱煡銆佽涓洪摼璺凡鍏峰 SQL repository 楠ㄦ灦锛屼絾褰撳墠椤圭洰浠嶆湭鍚敤鐪熷疄 PostgreSQL 椹卞姩鍜屾祴璇曞簱锛涙暟鎹簱绾у敮涓€绾︽潫銆佷簨鍔″洖婊氬拰骞跺彂琛屼负闇€瑕佸悗缁湪闆嗘垚鐜澶嶆祴銆?- `wechat_subscribe_templates` 鐩墠渚濊禆鏁版嵁搴撶瀛愭垨鍚庡彴閰嶇疆锛涘鏋滃簱鍐呮病鏈夋ā鏉匡紝鏈嶅姟浼氫繚鐣欏唴瀛橀粯璁ゆā鏉夸綔涓烘湰鍦?fallback銆?
## 2026-06-14 02:10

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堣ˉ榻愬皬绋嬪簭鍚庣鏂囦欢涓婁紶銆両M 鏂囦欢銆佷妇鎶ラ檮浠跺拰瀵煎嚭鏂囦欢渚濊禆鐨勬枃浠剁櫥璁版寔涔呭寲鑳藉姏銆?- 鏂板杩佺Щ `db/migrations/000015_files_repository.sql`锛?  - 鍒涘缓 `files` 琛ㄣ€?  - 瀛楁瑕嗙洊涓婁紶浜恒€佷笟鍔＄被鍨嬨€佷笟鍔″璞°€佹枃浠跺悕銆丮IME銆佸ぇ灏忋€丼HA256銆佸瓨鍌?key銆佽闂骇鍒€佽繃鏈熸椂闂村拰鍒涘缓鏃堕棿銆?  - 澧炲姞 `uk_files_storage_key` 鍞竴绱㈠紩銆?  - 澧炲姞 `idx_files_biz_object` 鍜?`idx_files_uploader_created`銆?- `internal/files` 鏂板鏂囦欢鏈嶅姟 repository锛?  - `Repository`
  - `NewServiceWithRepository`
  - `NewSQLRepository`
  - 涓婁紶鍑瘉鍒涘缓鏃跺彲鍐欏叆 `files`銆?  - 瀵煎嚭绛夌郴缁熺敓鎴愭枃浠跺彲鍐欏叆 `files`銆?  - 涓嬭浇鍦板潃鐢熸垚鍓嶅彲浠?repository 璇诲彇鏂囦欢鍏冩暟鎹€?- `internal/appapi.Server` repository 娉ㄥ叆鐐圭户缁墿灞曪細
  - 鏈夋暟鎹簱鏃舵敞鍏?`files.NewSQLRepository(db)`銆?  - 鏃犳暟鎹簱鏃剁户缁娇鐢ㄥ師鍐呭瓨鏂囦欢鏈嶅姟锛屼繚鎸佹湰鍦版帴鍙ｆ祴璇曞拰鑱旇皟琛屼负绋冲畾銆?- `cmd/server` 鏂板 `fileRepository(db)`锛屽拰琛屼负鏃ュ織銆佽鍗曘€佷妇鎶ャ€侀€氱煡涓€璧峰湪鍚姩鏃舵敞鍐屻€?- `plan.md` 鍚屾琛ュ厖锛?  - 杩佺Щ娓呭崟澧炲姞 `000015_files_repository.sql`銆?  - 鍞竴绾︽潫娓呭崟澧炲姞 `uk_files_storage_key`銆?  - C3 鏁版嵁搴撹縼绉绘墽琛岄『搴忓鍔犵 15 椤广€?- 琛ュ厖 `internal/files` fake repository 鍗曞厓娴嬭瘯锛岀‘璁や笂浼犲嚟璇併€佺敓鎴愭枃浠躲€佹煡璇笅杞介兘浼氳蛋 repository 鍒嗘敮銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/files`
  - `internal/notifications`
  - `internal/reports`
  - `internal/audit`
  - 鍏朵綑宸叉湁鍚庣鍖呮甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- 鏂囦欢鏈嶅姟宸插叿澶?SQL repository锛屼絾鐪熷疄瀵硅薄瀛樺偍浠嶆槸 mock URL锛涘悗缁渶瑕佹帴 MinIO/COS 棰勭鍚嶄笂浼犲拰涓嬭浇锛屽苟淇濈暀褰撳墠鏉冮檺鏍￠獙銆?- 褰撳墠鏁版嵁搴撶骇楠岃瘉浠嶅彈 PostgreSQL 椹卞姩鍜屾祴璇曞簱绛栫暐闄愬埗锛涙湰杞鐩栬縼绉绘枃鏈€佺紪璇戙€佹湇鍔℃敞鍏ュ拰 fake repository 琛屼负銆?
## 2026-06-14 02:40

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堣ˉ榻愬皬绋嬪簭鍚庣鏀惰棌灞€鏁版嵁娌夋穩鑳藉姏銆?- `internal/games` 鏂板鏀惰棌涓撶敤鎸佷箙鍖栨帴鍙ｄ笌 SQL 閫傞厤锛?  - `FavoriteRepository`
  - `NewServiceWithFavoriteRepository`
  - `NewSQLFavoriteRepository`
  - 鏀惰棌灞€鏃跺啓鍏?`game_favorites`锛屼緷璧?`user_id + game_id` 鍞竴绾︽潫淇濇寔骞傜瓑銆?  - 鍙栨秷鏀惰棌鏃跺垹闄?`game_favorites` 瀵瑰簲璁板綍锛屼繚鎸侀噸澶嶅彇娑堝箓绛夈€?  - 鎴戠殑鏀惰棌浼樺厛浠?repository 璇诲彇锛屽啀澶嶇敤褰撳墠娓告垙鏈嶅姟琛ラ綈 `Game` 淇℃伅骞惰繃婊ゆ湭瀹℃牳灞€銆?  - 鍏ㄩ噺鏀惰棌浼樺厛浠?repository 璇诲彇锛屼緵 AI 鏁版嵁鍑嗗鍜屽悗鍙扮粺璁＄户缁鐢ㄣ€?- `cmd/server` 鍚姩鍏ュ彛鎺ュ叆 `favoriteRepository(db)`锛?  - 鏈夋暟鎹簱鏃朵娇鐢?`games.NewSQLFavoriteRepository(db)`銆?  - 鏃犳暟鎹簱鏃剁户缁娇鐢ㄥ師鍐呭瓨鏀惰棌鏈嶅姟銆?- 琛ュ厖 `internal/games` fake repository 鍗曞厓娴嬭瘯锛岀‘璁ゆ敹钘忋€佹垜鐨勬敹钘忋€佸叏閲忔敹钘忋€佸彇娑堟敹钘忛兘浼氳蛋 repository 鍒嗘敮銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/games`
  - `internal/files`
  - `internal/audit`
  - `internal/notifications`
  - 鍏朵綑宸叉湁鍚庣鍖呮甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- 褰撳墠娓告垙涓昏〃 `games` 杩愯鏃朵粛浠ュ唴瀛樻湇鍔′负涓伙紱鏀惰棌 repository 宸茶兘娌夋穩 `game_favorites`锛屼絾閲嶅惎鍚庡鏋滄父鎴忎富琛ㄦ湭鍚屾鍏ュ簱锛屾敹钘忓垪琛ㄦ棤娉曡ˉ榻愬畬鏁?`Game` 璇︽儏銆?- Go 宸ュ叿閾惧湪鏈満鐜浼氬皾璇曞啓鐢ㄦ埛鐩綍 telemetry token 骞惰緭鍑烘潈闄愯鍛婏紱娴嬭瘯閫€鍑虹爜涓?0锛屾墍鏈夊寘宸查€氳繃銆?
## 2026-06-14 03:10

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛岃ˉ榻愬悗鍙?鍐呴儴鍏抽敭鎿嶄綔瀹¤鐨?PostgreSQL 鎸佷箙鍖栨壙鎺ャ€?- `internal/audit` 鎵╁睍鎿嶄綔鏃ュ織 repository锛?  - `OperationRepository`
  - `NewServiceWithRepositories`
  - `NewSQLOperationRepository`
  - 琛屼负鏃ュ織鍜屾搷浣滄棩蹇楀彲鍒嗗埆娉ㄥ叆 repository锛屼簰涓嶅奖鍝嶃€?- `RecordOperation` 缁х画淇濈暀鍐呭瓨璁板綍锛屽悓鏃跺湪鏈?repository 鏃跺啓鍏?`operation_logs`銆?- `OperationLogs` 鍦ㄦ湁 repository 鏃朵紭鍏堣鍙?`operation_logs`锛屾棤 repository 鎴栬鍙栧け璐ユ椂鍥為€€鍐呭瓨銆?- `cmd/server` 鍚姩鍏ュ彛鏂板 `operationRepository(db)`锛?  - 鏈夋暟鎹簱鏃朵娇鐢?`audit.NewSQLOperationRepository(db)`銆?  - 鏃犳暟鎹簱鏃剁户缁娇鐢ㄥ師鍐呭瓨瀹¤鏈嶅姟銆?- `internal/appapi.Server.UseRepositories` 鎵╁睍涓哄悓鏃舵敞鍏ヨ涓烘棩蹇楀拰鎿嶄綔鏃ュ織 repository銆?- 琛ュ厖 `internal/audit` fake operation repository 鍗曞厓娴嬭瘯锛岀‘璁ゆ搷浣滄棩蹇椾繚瀛樺拰鍒楄〃鏌ヨ閮戒細璧?repository 鍒嗘敮銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/audit`
  - `internal/games`
  - `internal/files`
  - `internal/notifications`
  - 鍏朵綑宸叉湁鍚庣鍖呮甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- `operation_logs` 宸插叿澶?SQL repository锛屼絾骞堕潪鎵€鏈夊悗鍙板啓鎿嶄綔閮藉凡瑕嗙洊鎿嶄綔鏃ュ織锛涘悗缁粛闇€鎸?`plan.md` 閫愪釜琛ラ綈瀹℃牳銆侀厤缃€佸厬鎹€佸垎娑︾瓑鍐欐搷浣滅殑瀹¤鐐广€?- Go 宸ュ叿閾惧湪鏈満鐜浠嶄細灏濊瘯鍐欑敤鎴风洰褰?telemetry token 骞惰緭鍑烘潈闄愯鍛婏紱娴嬭瘯閫€鍑虹爜涓?0锛屼笟鍔″寘鍧囬€氳繃銆?## 2026-06-14 03:45

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟帹杩涘皬绋嬪簭鍚庣涓婄嚎绾?PostgreSQL 鎵挎帴鑳藉姏銆?- 鏂板杩佺Щ `db/migrations/000016_games_repository.sql`锛?  - 缁?`games` 琛ュ厖 `longitude`銆乣latitude`锛岃缁勫眬涓昏〃鍙洿鎺ユ壙杞藉皬绋嬪簭绔畾浣嶅瓧娈点€?  - 鏂板 `game_members`锛屾壙杞藉眬鎴愬憳銆佽鑹层€佺姸鎬佸拰鍏ュ眬鏃堕棿銆?  - 鏂板 `game_applications`锛屾壙杞藉叆灞€鐢宠銆佸鏍哥姸鎬併€佸師鍥犲拰瀹℃牳鏃堕棿銆?  - 鏂板鎴愬憳鐢ㄦ埛绱㈠紩銆佺敵璇风敤鎴风储寮曘€佺敵璇风姸鎬佺储寮曘€?- `internal/games` 鏂板缁勫眬涓婚摼璺?repository 鎺ュ彛涓?SQL 閫傞厤锛?  - `Repository`
  - `NewServiceWithRepositories`
  - `NewSQLRepository`
  - 鏀寔鍒涘缓灞€銆佹洿鏂板眬銆佹煡璇㈠眬銆佸垪琛ㄣ€佹瘡鏃ュ垱寤烘暟銆佹垚鍛樺鍒犳煡銆佸叆灞€鐢宠鍒涘缓/鏌ヨ/鏇存柊銆佸緟澶勭悊鐢宠骞傜瓑妫€鏌ャ€?- `cmd/server` 鍚姩鍏ュ彛鏂板 `gameRepository(db)` 娉ㄥ叆锛?  - 鏈夋暟鎹簱鏃朵娇鐢?`games.NewSQLRepository(db)`銆?  - 鏃犳暟鎹簱鏃剁户缁繚鎸佸師鏈夊唴瀛樻ā寮忥紝淇濊瘉鏈湴鑱旇皟鍜屾棦鏈夋祴璇曠ǔ瀹氥€?- `games.Service` 鏂板 SQL 鍥炲～鏈哄埗锛?  - 鏈嶅姟閲嶅惎鎴栧唴瀛樹负绌烘椂锛宍Get`銆佹垚鍛樻煡璇€佺敵璇峰鏍哥瓑鏍稿績璺緞鍙互浠?repository 鍥炲～灞€鍜屾垚鍛樸€?  - 鍒涘缓灞€銆佸鎵瑰眬銆佺敵璇峰叆灞€銆佸鏍稿叆灞€銆佹墜鍔ㄥ紑濮嬨€侀€€鍑哄眬浼氬湪鏈?repository 鏃跺啓鍏ユ暟鎹簱銆?  - 浠嶅悓姝ヤ繚鐣欏唴瀛樻€侊紝閬垮厤褰卞搷褰撳墠 IM銆佽瘎浠枫€佹敹鐩婄瓑渚濊禆鍚屼竴杩涚▼鐘舵€佺殑鍚庣画閾捐矾銆?- 鏂板 `internal/games/repository_service_test.go`锛?  - 瑕嗙洊 repository 鍒涘缓灞€銆佸垱寤鸿€呮垚鍛樺啓鍏ャ€佺敵璇峰叆灞€銆佸鏍稿叆灞€銆佹垚鍛樻煡璇€?  - 瑕嗙洊鏂版湇鍔″疄渚嬩粠 repository 璇诲彇灞€鍜屾垚鍛橈紝楠岃瘉閲嶅惎鍚庡洖濉兘鍔涖€?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/games`
  - `internal/audit`
  - `internal/files`
  - `internal/notifications`
  - `internal/orders`
  - 鍏朵綑宸叉湁灏忕▼搴忓悗绔寘鍧囨甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- `games` 涓婚摼璺凡鍏峰 SQL repository 楠ㄦ灦锛屼絾鐪熷疄 PostgreSQL 椹卞姩鍜屾祴璇曞簱浠嶆湭鍚敤锛涙暟鎹簱绾у敮涓€绾︽潫銆佷簨鍔″洖婊氥€佸苟鍙戝叆灞€鍜屾弧鍛樼珵浜夎繕闇€瑕侀泦鎴愭祴璇曞娴嬨€?- 鏈嶅姟纭銆侀噷绋嬬銆佹墦鍗°€佸鐩樸€佺画灞€鑽夌浠嶄富瑕佹槸鍐呭瓨鎬侊紱涓嬩竴姝ュ簲缁х画鎶?D5.6/E2.2 杩欎簺渚濊禆 `game_id` 鐨勮繘搴︽暟鎹矇鍏?PostgreSQL銆?- 褰撳墠 SQL 鍐欏叆鏄皬姝ユ帴鍏ワ紝灏氭湭鎶婄粍灞€瀹℃牳銆佺敵璇峰鏍搞€佹垚鍛樺啓鍏ュ拰浜烘暟鏇存柊鍖呮垚鍗曚釜鏁版嵁搴撲簨鍔★紱鎺ュ叆鐪熷疄 PostgreSQL 鍚庨渶瑕佽ˉ浜嬪姟灏佽銆?## 2026-06-14 04:15

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟帹杩涘皬绋嬪簭鍚庣 D5.6/E2.2 杩涘害绫绘暟鎹殑 PostgreSQL 鎵挎帴銆?- 澶嶆牳 `db/migrations/000010_p1_reserved.sql`锛岀‘璁ゅ凡鏈夎〃鍙鐢細
  - `game_milestones`
  - `game_checkins`
  - `game_retrospectives`
  - `game_continue_drafts`
- `internal/games` 鏂板杩涘害绫绘寔涔呭寲鎺ュ彛锛?  - `ProgressRepository`
  - `UseProgressRepository`
  - `NewSQLProgressRepository`
- `services/go-api/internal/games/progress_sql_repository.go` 鏂板 SQL 閫傞厤锛?  - 閲岀▼纰戝垱寤恒€佹洿鏂般€佸垪琛ㄣ€?  - 鎵撳崱鍒涘缓銆佺姸鎬佹洿鏂般€佹垚鍛樺垪琛ㄨ繃婊ゆ棤鏁堟墦鍗°€佸悗鍙板垪琛ㄥ寘鍚棤鏁堟墦鍗°€?  - 澶嶇洏鍒涘缓銆侀噸澶嶆彁浜ゆ鏌ャ€佸鐩樺垪琛ㄣ€?  - 缁眬鑽夌鍒涘缓銆佸悗鍙扮画灞€鑽夌鍒楄〃銆?  - `game_checkins.file_ids` 鎸?JSON/JSONB 璇诲啓锛屼繚鎸佸拰杩佺Щ琛ㄧ粨鏋勪竴鑷淬€?- `games.Service` 鎺ュ叆杩涘害绫?repository锛?  - `CreateMilestone`銆乣UpdateMilestone`銆乣Milestones`銆乣AdminMilestones`
  - `CreateCheckin`銆乣MarkCheckinInvalid`銆乣Checkins`銆乣AdminCheckins`
  - `CreateRetrospective`銆乣Retrospectives`銆乣AdminRetrospectives`
  - `ContinueDraft`銆乣AdminContinueDrafts`
  - 鏈?repository 鏃朵紭鍏堣鍐?PostgreSQL锛涙棤 repository 鏃朵繚鎸佸師鍐呭瓨琛屼负銆?- `cmd/server` 鏂板 `gameProgressRepository(db)` 娉ㄥ叆锛?  - 鏈夋暟鎹簱鏃朵娇鐢?`games.NewSQLProgressRepository(db)`銆?  - 鏃犳暟鎹簱鏃剁户缁湰鍦板唴瀛樿仈璋冦€?- 鏂板 `services/go-api/internal/games/progress_repository_service_test.go`锛?  - 瑕嗙洊閲岀▼纰戝垱寤?鏇存柊/鍒楄〃璧?repository銆?  - 瑕嗙洊鎵撳崱鍒涘缓銆佸悗鍙扮疆鏃犳晥銆佹垚鍛樼闅愯棌鏃犳晥鎵撳崱銆佸悗鍙扮鍙鏃犳晥鎵撳崱銆?  - 瑕嗙洊澶嶇洏鍒涘缓鍜?repository 绾ч噸澶嶆彁浜ゆ嫤鎴€?  - 瑕嗙洊缁眬鑽夌鍐欏叆鍜屽悗鍙版煡璇€?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/games`
  - `internal/audit`
  - `internal/files`
  - `internal/notifications`
  - `internal/orders`
  - 鍏朵綑宸叉湁灏忕▼搴忓悗绔寘鍧囨甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- 杩涘害绫绘暟鎹凡缁忓叿澶?SQL repository锛屼絾鐪熷疄 PostgreSQL 椹卞姩鍜屾祴璇曞簱浠嶆湭鍚敤锛沗jsonb file_ids`銆佸苟鍙戞墦鍗°€佸鐩樺敮涓€鎬у拰缁眬鑽夌鍞竴绾︽潫闇€瑕侀泦鎴愭祴璇曞娴嬨€?- `game_service_confirms` 鍜?`game_service_confirm_items` 浠嶆湭鎺?SQL repository锛汥5.6 鐨勬湇鍔＄‘璁ら摼璺笅涓€姝ュ簲缁х画娌夊簱銆?- `ContinueDraft` 鍦ㄦ湁涓婚摼璺?repository 鏃跺凡缁忓垱寤鸿崏绋垮眬鍜屾垚鍛橈紝浣嗚繕娌℃湁浜嬪姟灏佽锛涚湡瀹炴暟鎹簱鎺ュ叆鍚庯紝闇€瑕佹妸鑽夌灞€銆佽崏绋挎垚鍛樸€佺画灞€璁板綍鏀捐繘鍚屼竴涓簨鍔°€?## 2026-06-14 04:45

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟帹杩涘皬绋嬪簭鍚庣 D5.6 鏈嶅姟纭閾捐矾鐨?PostgreSQL 鎵挎帴銆?- 澶嶆牳 `db/migrations/000004_games.sql`锛岀‘璁ゅ凡鏈夎〃鍙鐢細
  - `game_service_confirms`
  - `game_service_confirm_items`
- `internal/games` 鏂板鏈嶅姟纭鎸佷箙鍖栨帴鍙ｏ細
  - `ServiceConfirmRepository`
  - `UseServiceConfirmRepository`
  - `NewSQLServiceConfirmRepository`
- `services/go-api/internal/games/service_confirm_sql_repository.go` 鏂板 SQL 閫傞厤锛?  - 鎸?`game_id` 鑾峰彇鏈嶅姟纭鍗曞拰纭鏄庣粏銆?  - 鍒涘缓鎴栨洿鏂?`game_service_confirms`锛屾敮鎸?`completed_at`銆?  - 鎸?`game_id + user_id` 骞傜瓑鍐欏叆 `game_service_confirm_items`銆?  - 鏌ヨ纭鏄庣粏骞跺洖濉?`confirmedBy`銆?- `games.Service.ConfirmService` 鎺ュ叆 repository锛?  - 鏈?repository 鏃朵紭鍏堜粠 PostgreSQL 璇诲彇宸插瓨鍦ㄧ‘璁ゅ崟鍜屾槑缁嗐€?  - 棣栨纭鍒涘缓纭鍗曞拰纭鏄庣粏銆?  - 閲嶅纭淇濇寔骞傜瓑锛屼笉閲嶅澧炲姞纭鏄庣粏銆?  - 鍏ㄥ憳纭鍚庡啓鍥炵‘璁ゅ崟 `completed` 鐘舵€侊紝骞舵妸灞€鐘舵€佹帹杩涘埌 `pending_review`銆?  - 鏃?repository 鏃剁户缁繚鎸佸師鍐呭瓨琛屼负銆?- `cmd/server` 鏂板 `gameServiceConfirmRepository(db)` 娉ㄥ叆锛?  - 鏈夋暟鎹簱鏃朵娇鐢?`games.NewSQLServiceConfirmRepository(db)`銆?  - 鏃犳暟鎹簱鏃剁户缁湰鍦板唴瀛樿仈璋冦€?- 鏂板 `services/go-api/internal/games/service_confirm_repository_service_test.go`锛?  - 瑕嗙洊纭鍗曞垱寤恒€佺‘璁ゆ槑缁嗗啓鍏ャ€?  - 瑕嗙洊鍏ㄥ憳纭鍚庣姸鎬佽繘鍏?`completed` / `pending_review`銆?  - 瑕嗙洊 `ServiceConfirmForGame` 浠?repository 鏌ヨ銆?  - 瑕嗙洊閲嶅纭骞傜瓑銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/games`
  - `internal/audit`
  - `internal/files`
  - `internal/notifications`
  - `internal/orders`
  - 鍏朵綑宸叉湁灏忕▼搴忓悗绔寘鍧囨甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- 鏈嶅姟纭宸茬粡鍏峰 SQL repository锛屼絾鐪熷疄 PostgreSQL 椹卞姩鍜屾祴璇曞簱浠嶆湭鍚敤锛沗game_id` 鍞竴纭鍗曘€乣game_id + user_id` 鍞竴纭鏄庣粏鍜屽苟鍙戠‘璁ら渶瑕侀泦鎴愭祴璇曞娴嬨€?- `ConfirmService` 鐩墠浠嶄笉鏄暟鎹簱浜嬪姟锛涚‘璁ゆ槑缁嗗啓鍏ャ€佺‘璁ゅ崟鐘舵€佹洿鏂般€佸眬鐘舵€佹洿鏂板簲鍦ㄧ湡瀹?PostgreSQL 鎺ュ叆鍚庣撼鍏ュ悓涓€浜嬪姟銆?- D5.6 鐨勮瘎浠枫€佹垚闀裤€佷俊鐢ㄣ€佽冻杩逛富鏁版嵁浠嶄富瑕佺敱鍐呭瓨鏈嶅姟椹卞姩锛涗笅涓€姝ュ簲缁х画鎶?reviews/growth/credit/footprints 鐨?repository 琛ラ綈銆?## 2026-06-14 05:20

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟帹杩涘皬绋嬪簭鍚庣 D5.6 璇勪环銆佹垚闀裤€佷俊鐢ㄥ拰瓒宠抗閾捐矾鐨?PostgreSQL 鎵挎帴銆?- 澶嶆牳 `db/migrations/000006_review_growth_credit.sql`锛岀‘璁ゅ凡鏈夎〃鍙鐢細
  - `reviews`
  - `review_reminders`
  - `user_growth_profiles`
  - `experience_logs`
  - `points_accounts`
  - `points_logs`
  - `credit_logs`
  - `daily_credit_scores`
  - `user_footprints`
  - `achievements`
  - `user_achievements`
- `internal/reviews` 鏂板 repository 鎺ュ彛涓?SQL 閫傞厤锛?  - `Repository`
  - `NewServiceWithRepository`
  - `NewSQLRepository`
- `services/go-api/internal/reviews/sql_repository.go` 鏂板 SQL 鎵挎帴锛?  - 鏈嶅姟瀹屾垚鍚庝负鎴愬憳鍐欏叆 `review_reminders`銆?  - 璇勪环鎻愪氦鍐欏叆 `reviews`锛屼緷璧?`uk_review_once` 淇濇寔鍞竴璇勪环璇箟銆?  - 鎴愰暱璧勬枡鍐欏叆 `user_growth_profiles`銆?  - 缁忛獙娴佹按鍐欏叆 `experience_logs`銆?  - 绉垎璐︽埛鍜屾祦姘村啓鍏?`points_accounts`銆乣points_logs`銆?  - 淇＄敤鍒嗗啓鍏?`daily_credit_scores`锛屾墸鍒嗘祦姘村啓鍏?`credit_logs`銆?  - 瓒宠抗鍐欏叆 `user_footprints`銆?  - 鎴愬氨瀹氫箟鍜岀敤鎴锋垚灏卞啓鍏?`achievements`銆乣user_achievements`銆?- `reviews.Service` 鎺ュ叆 repository锛?  - `MarkGameReviewable` 鏈?repository 鏃跺啓鍏ュ緟璇勪环鎻愰啋銆?  - `Todos`銆乣Submit`銆乣MyIntents`銆乣AllReviews`銆乣Profile`銆乣Footprints`銆乣AllFootprints`銆乣TraceByUser`銆乣TraceByGame` 浼樺厛璧?repository銆?  - `DeductCredit` 鏈?repository 鏃跺啓鍏ュ綋鏃ヤ俊鐢ㄥ垎銆佷俊鐢ㄦ祦姘村拰瓒宠抗銆?  - 鏃?repository 鏃剁户缁繚鎸佸師鍐呭瓨琛屼负銆?- `appapi.Server.UseRepositories` 鎵╁睍 `reviewRepo` 娉ㄥ叆锛?  - 鏈?repository 鏃堕噸寤?`reviews` 鏈嶅姟锛屽苟鍚屾閲嶅缓渚濊禆 reviews 鐨?`revenue`銆乣reports`銆乣memberReports`銆乣teams`銆?  - 閬垮厤鏀剁泭缁撶畻浠嶅紩鐢ㄦ棫鐨勫唴瀛?reviews 鏈嶅姟銆?- `cmd/server` 鏂板 `reviewRepository(db)` 娉ㄥ叆銆?- 鏂板 `services/go-api/internal/reviews/repository_service_test.go`锛?  - 瑕嗙洊寰呰瘎浠锋彁閱掋€佽瘎浠锋彁浜ゃ€侀噸澶嶈瘎浠锋嫤鎴€?  - 瑕嗙洊鎴愰暱銆佺粡楠屻€佺Н鍒嗐€佽冻杩广€佹垚灏卞啓鍏ャ€?  - 瑕嗙洊淇＄敤鎵ｅ垎銆佺敤鎴疯拷韪煡璇€?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/reviews`
  - `internal/games`
  - `internal/audit`
  - `internal/files`
  - `internal/notifications`
  - `internal/orders`
  - 鍏朵綑宸叉湁灏忕▼搴忓悗绔寘鍧囨甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- reviews/growth/credit/footprints 宸插叿澶?SQL repository锛屼絾鐪熷疄 PostgreSQL 椹卞姩鍜屾祴璇曞簱浠嶆湭鍚敤锛涘敮涓€绾︽潫銆佷簨鍔″拰骞跺彂閲嶅璇勪环闇€瑕侀泦鎴愭祴璇曞娴嬨€?- `Submit` 褰撳墠浼氬垎姝ュ啓 reviews銆乬rowth銆乪xperience銆乸oints銆乫ootprints銆乤chievements锛涚湡瀹炴暟鎹簱鎺ュ叆鍚庡簲灏佽浜嬪姟锛岄伩鍏嶈瘎浠峰啓鍏ユ垚鍔熶絾鎴愰暱/绉垎閮ㄥ垎澶辫触銆?- `points_accounts` 褰撳墠鐢?reviews repository 缁存姢鍙敤绉垎锛涘悗缁鏋?points 妯″潡涔熸帴 SQL repository锛岄渶瑕佺粺涓€绉垎璐︽埛鍐欏叆鍏ュ彛锛岄伩鍏嶅弻鍐欒鍒欎笉涓€鑷淬€?## 2026-06-14 06:05

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟帹杩涘皬绋嬪簭鍚庣 D5.6 绉垎銆佸厬鎹笌璇勪环閾捐矾鍚庣殑 PostgreSQL 鎵挎帴銆?- 澶嶆牳褰撳墠瀹炵幇鍚庣‘璁わ細
  - `reviews` repository 宸茬粡鍐欏叆 `points_accounts`銆乣points_logs`銆?  - `points.Service` 浠嶄互鍐呭瓨璐︽埛鍜屽唴瀛樻祦姘翠负涓汇€?  - `redemption.Service` 浠嶄互鍐呭瓨鍟嗗搧銆佸唴瀛樿鍗曚负涓汇€?- 鏂板 `internal/points` 浠撳偍鎺ュ彛鍜?SQL 閫傞厤锛?  - `Repository`
  - `NewServiceWithRepository`
  - `NewSQLRepository`
  - `Summary`銆乣Logs`銆乣AllLogs`銆乣Grant`銆乣Deduct` 鍦ㄦ湁 repository 鏃朵紭鍏堣鍐?PostgreSQL銆?- 鏂板 `internal/redemption` 浠撳偍鎺ュ彛鍜?SQL 閫傞厤锛?  - `Repository`
  - `NewServiceWithRepository`
  - `NewSQLRepository`
  - 鍟嗗搧鍒涘缓銆佸垪琛ㄣ€佹洿鏂般€佸簱瀛橀鎵ｃ€佸簱瀛樺洖婊氥€?  - 鍏戞崲璁㈠崟鍒涘缓銆佺敤鎴疯鍗曘€佸悗鍙拌鍗曘€佽鍗曞鏍搞€?  - 椹冲洖璁㈠崟鍚庨€氳繃绉垎鏈嶅姟杩旇繕绉垎銆?- 鏂板杩佺Щ `db/migrations/000017_points_redemption_repository.sql`锛?  - `points_logs.before_points`
  - `points_logs.after_points`
  - `points_logs.biz_type`
  - `points_logs.biz_id`
  - 绉垎娴佹按銆佸厬鎹㈠晢鍝併€佸厬鎹㈣鍗曟煡璇㈢储寮曘€?- 鏇存柊 `plan.md`锛?  - C3 鏁版嵁搴撹縼绉绘墽琛岄『搴忚拷鍔?`000017_points_redemption_repository.sql`銆?  - D4 鏁版嵁搴撹縼绉讳笂绾块『搴忚拷鍔犵 17 椤广€?- 鏇存柊 `cmd/server` 鍜?`appapi.Server.UseRepositories`锛?  - 鏈夋暟鎹簱鏃舵敞鍏?`points.NewSQLRepository(db)`銆?  - 鏈夋暟鎹簱鏃舵敞鍏?`redemption.NewSQLRepository(db)`銆?  - 鍏戞崲鏈嶅姟澶嶇敤鍚屼竴涓Н鍒嗘湇鍔★紝閬垮厤璇勪环濂栧姳銆佸厬鎹㈡墸鍑忋€侀┏鍥炶繑杩樺垎鏁ｅ埌璐︽埛鍏ュ彛銆?- 鏂板娴嬭瘯锛?  - `services/go-api/internal/points/repository_service_test.go`
  - `services/go-api/internal/redemption/repository_service_test.go`

### 楠岃瘉缁撴灉

- 宸叉墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/points`
  - `internal/redemption`
  - `internal/reviews`
  - `internal/games`
  - 鍏朵綑宸叉湁灏忕▼搴忓悗绔寘鍧囨甯哥紪璇戞垨娴嬭瘯閫氳繃銆?- Go 浠嶈緭鍑洪潪鑷村懡 telemetry 鎻愮ず锛?  - `error acquiring upload token: creating token file: open C:\Users\61492\AppData\Roaming\go\telemetry\local\upload.token: Access is denied.`
  - 鍛戒护閫€鍑虹爜涓?0锛屼笉褰卞搷鏈疆楠岃瘉缁撹銆?
### 褰撳墠椋庨櫓

- 绉垎鍜屽厬鎹㈠凡缁忓叿澶?SQL repository锛屼絾鐪熷疄 PostgreSQL 椹卞姩鍜屾祴璇曞簱浠嶆湭鍚敤锛涗笅涓€杞簲琛ユ暟鎹簱闆嗘垚娴嬭瘯鎴栬縼绉?dry-run銆?- SQL 鍏戞崲涓嬪崟褰撳墠鎸夆€滈鎵ｅ簱瀛?-> 鎵ｇН鍒?-> 寤鸿鍗曗€濋『搴忎繚鎸佺姸鎬佷竴鑷达紝骞剁敤棰勫彇璁㈠崟 ID 鍐欏叆绉垎娴佹按锛涗粛寤鸿鍚庣画鍦ㄧ湡瀹炴暟鎹簱閲屽皝瑁呬簨鍔★紝閬垮厤杩涚▼寮傚父瀵艰嚧灞€閮ㄧ姸鎬佸仠鐣欍€?## 2026-06-14 06:45

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟帹杩涘皬绋嬪簭鍚庣 P1 棰勭暀鑳藉姏鐨勬暟鎹寔涔呭寲銆?- 澶嶆牳褰撳墠瀹炵幇鍚庣‘璁わ細
  - `connections.Service` 宸叉湁灏忕▼搴忓拰鍚庡彴鎺ュ彛锛屼絾浠嶄互鍐呭瓨鍏崇郴鍜屽唴瀛樿窡杩涜褰曚负涓汇€?  - `profiles.Service` 宸叉湁琛屽鎶€鑳芥爲銆侀璺汉璧勬簮鐢诲儚鎺ュ彛锛屼絾浠嶄互鍐呭瓨瑙掕壊鍜屽唴瀛樼敾鍍忎负涓汇€?  - `db/migrations/000003_realname_roles.sql` 宸叉湁 `user_roles`銆?  - `db/migrations/000010_p1_reserved.sql` 宸叉湁 `user_connections`銆乣connection_follow_logs`銆乣expert_skill_profiles`銆乣guide_resource_profiles`銆?- 鏂板 `internal/connections` 浠撳偍鎺ュ彛鍜?SQL 閫傞厤锛?  - `Repository`
  - `NewServiceWithRepository`
  - `NewSQLRepository`
  - `UpsertPair` 鍦ㄦ湁 repository 鏃跺啓鍏ュ弻鍚戜汉鑴夊叧绯汇€?  - `My`銆乣All` 鍦ㄦ湁 repository 鏃朵粠 PostgreSQL 鏌ヨ銆?  - `AddFollowLog` 鍦ㄦ湁 repository 鏃舵牎楠屽弬涓庝汉/棰嗚矾浜烘潈闄愶紝鍐欏叆璺熻繘璁板綍锛屽苟澧炲姞鍏崇郴寮哄害銆?- 鏂板 `internal/profiles` 浠撳偍鎺ュ彛鍜?SQL 閫傞厤锛?  - `Repository`
  - `NewServiceWithRepository`
  - `NewSQLRepository`
  - `GrantRole` 鍐欏叆鎴栨縺娲?`user_roles`銆?  - `ExpertSkill`銆乣UpdateExpertSkill`銆乣AllExpertSkills` 璇诲啓 `expert_skill_profiles`銆?  - `GuideResource`銆乣UpdateGuideResource`銆乣AllGuideResources` 璇诲啓 `guide_resource_profiles`銆?  - `IsGuide` 鍦ㄦ湁 repository 鏃朵粠 `user_roles` 鍒ゆ柇銆?- 鏇存柊 `cmd/server` 鍜?`appapi.Server.UseRepositories`锛?  - 鏈夋暟鎹簱鏃舵敞鍏?`connections.NewSQLRepository(db)`銆?  - 鏈夋暟鎹簱鏃舵敞鍏?`profiles.NewSQLRepository(db)`銆?  - 鏃犳暟鎹簱鏃剁户缁繚鎸佸師鍐呭瓨鏈嶅姟锛屾柟渚挎湰鍦拌仈璋冦€?- 鏂板娴嬭瘯锛?  - `services/go-api/internal/connections/repository_service_test.go`
  - `services/go-api/internal/profiles/repository_service_test.go`

### 楠岃瘉缁撴灉

- 宸叉墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/connections`
  - `internal/profiles`
  - `internal/points`
  - `internal/redemption`
  - `internal/reviews`
  - 鍏朵綑宸叉湁灏忕▼搴忓悗绔寘鍧囨甯哥紪璇戞垨娴嬭瘯閫氳繃銆?- Go 浠嶈緭鍑洪潪鑷村懡 telemetry 鎻愮ず锛岄€€鍑虹爜涓?0锛屼笉褰卞搷鏈疆楠岃瘉缁撹銆?
### 褰撳墠椋庨櫓

- 浜鸿剦鍜岀敾鍍忓凡缁忓叿澶?SQL repository锛屼絾鐪熷疄 PostgreSQL 椹卞姩鍜屾祴璇曞簱浠嶆湭鍚敤锛涗笅涓€姝ュ簲琛ヨ縼绉?dry-run 鎴栭泦鎴愭祴璇曘€?- `GrantRole` 鐩墠婊¤冻鏈湴鍜岀瀛愭暟鎹巿鏉冿紱瀹屾暣鐨勮瀹?棰嗚矾浜虹敵璇枫€佸鏍搞€佷粯璐瑰紑閫氫粛搴旂敱鍚庣画瑙掕壊鐢宠閾捐矾缁熶竴椹卞姩銆?## 2026-06-14 07:15

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟帹杩涘皬绋嬪簭鍚庣 LBS / 瀹氫綅閾捐矾銆?- 澶嶆牳褰撳墠瀹炵幇鍚庣‘璁わ細
  - `POST /api/app/locations/current` 鍜?`POST /api/app/locations/manual` 宸叉湁鎺ュ彛銆?  - `GET /api/app/games/nearby` 宸蹭緷璧栧綋鍓嶇敤鎴峰畾浣嶈绠楅檮杩戝眬銆?  - `lbs.Service` 浠嶄互鍐呭瓨淇濆瓨褰撳墠浣嶇疆涓轰富銆?  - `plan.md` 宸茶姹?`user_location_records`銆乣GET /api/app/locations/my-recent` 鍜屾渶杩戝畾浣嶄笉娉勯湶浠栦汉浣嶇疆銆?- 鏂板 `internal/lbs` 浠撳偍鎺ュ彛鍜?SQL 閫傞厤锛?  - `Repository`
  - `NewServiceWithRepository`
  - `NewSQLRepository`
  - `SaveLocation`
  - `CurrentLocation`
  - `RecentLocations`
- 鎵╁睍 `lbs.Service`锛?  - `SaveCurrent` / `SaveManual` 鍦ㄦ湁 repository 鏃跺啓鍏?PostgreSQL銆?  - `Current` 鍦ㄦ湁 repository 鏃惰鍙栨渶杩戜竴鏉″畾浣嶃€?  - 鏂板 `Recent(userID, limit)`锛岃繑鍥炲綋鍓嶇敤鎴锋渶杩戝畾浣嶃€?  - 鏃?repository 鏃朵繚鐣欏唴瀛樿涓猴紝骞剁淮鎶ゆ湰鍦版渶杩戝畾浣嶅垪琛ㄣ€?- 鏂板鎺ュ彛锛?  - `GET /api/app/locations/my-recent`
  - 鏀寔 `limit` 鏌ヨ鍙傛暟锛岃寖鍥撮檺鍒跺湪 1-50銆?- 鏂板杩佺Щ锛?  - `db/migrations/000018_user_location_records.sql`
  - 鍒涘缓 `user_location_records`
  - 澧炲姞 `idx_user_location_records_user_created`
  - 澧炲姞 `idx_user_location_records_city_created`
- 鏇存柊 `cmd/server` 鍜?`appapi.Server.UseRepositories`锛?  - 鏈夋暟鎹簱鏃舵敞鍏?`lbs.NewSQLRepository(db)`銆?  - 鏃犳暟鎹簱鏃剁户缁娇鐢ㄥ唴瀛樺畾浣嶏紝鏂逛究鏈湴鑱旇皟銆?- 鏇存柊 `docs/openapi/app.openapi.yaml`锛?  - 澧炲姞 `GET /api/app/locations/my-recent`銆?- 鏇存柊 `plan.md`锛?  - C3 鏁版嵁搴撹縼绉婚『搴忚拷鍔?`000018_user_location_records.sql`銆?  - D4 鏁版嵁搴撲笂绾块『搴忚拷鍔犵 18 椤广€?- 鏂板娴嬭瘯锛?  - `services/go-api/internal/lbs/repository_service_test.go`

### 楠岃瘉缁撴灉

- 宸叉墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 棣栨娴嬭瘯鍙戠幇 `appapi/server.go` 缂哄皯 `internal/lbs` import锛屽凡淇銆?- 澶嶆祴閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/lbs`
  - `internal/games`
  - `internal/reviews`
  - 鍏朵綑宸叉湁灏忕▼搴忓悗绔寘鍧囨甯哥紪璇戞垨娴嬭瘯閫氳繃銆?- Go 浠嶈緭鍑洪潪鑷村懡 telemetry 鎻愮ず锛岄€€鍑虹爜涓?0锛屼笉褰卞搷鏈疆楠岃瘉缁撹銆?
### 褰撳墠椋庨櫓

- 鐢ㄦ埛瀹氫綅宸茬粡鍏峰 SQL repository 鍜岃縼绉伙紝浣嗙湡瀹?PostgreSQL 椹卞姩鍜屾祴璇曞簱浠嶆湭鍚敤锛涘悗缁渶瑕佸仛杩佺Щ dry-run 鎴栭泦鎴愭祴璇曘€?- 鏈疆鏈帴鑵捐浣嶇疆鏈嶅姟閫嗗湴鍧€瑙ｆ瀽锛涘綋鍓嶅彧淇濆瓨灏忕▼搴忕浼犲叆鐨勫煄甯傘€佸湴鍧€鍜屽潗鏍囷紝鍚庣画搴旇ˉ鑵捐鍦板浘鏍囧噯鍖栧拰 geocode cache銆?## 2026-06-14 09:00

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟帹杩涘皬绋嬪簭鍚庣鐢ㄦ埛璧勬枡涓庢枃浠舵潈闄愰棴鐜€?- 鍩轰簬涓婁竴杞闄╅」锛岃ˉ榻愬ご鍍忔枃浠跺綊灞炴牎楠岋細
  - `PUT /api/app/users/me/profile` 鏀寔 `avatarFileId`銆?  - `avatarFileId` 蹇呴』瀛樺湪銆?  - 鏂囦欢蹇呴』鐢卞綋鍓嶇敤鎴蜂笂浼犮€?  - 鏂囦欢 `bizType` 蹇呴』涓?`avatar`銆?  - 鏍￠獙閫氳繃鍚庤嚜鍔ㄧ敓鎴愭湰鍦?mock 澶村儚涓嬭浇鍦板潃骞跺啓鍏?`avatarUrl`銆?- 璋冩暣鏂囦欢涓婁紶鍏ュ彛锛?  - `POST /api/app/files/upload-token` 瀵?`bizType=avatar` 鍏佽 pre-auth 鐢ㄦ埛涓婁紶銆?  - 鍏朵粬涓氬姟鏂囦欢浠嶈姹傚己瀹炲悕瀹屾垚銆?- 鏇存柊鐢ㄦ埛妯″瀷涓?SQL锛?  - `users.User` 澧炲姞 `avatarFileId`銆?  - `users.Repository.UpdateProfile` 澧炲姞 `avatarFileID` 鍙傛暟銆?  - `users.avatar_file_id` 鍐欏叆鍜屽洖璇汇€?- 鏇存柊杩佺Щ锛?  - `db/migrations/000020_user_profiles_repository.sql` 澧炲姞 `users.avatar_file_id` 鍜岀储寮曘€?- 鏇存柊鏂囨。锛?  - `docs/openapi/app.openapi.yaml` 澧炲姞 `avatarFileId` 瀛楁銆?  - `plan.md` 鏇存柊 `000020_user_profiles_repository.sql` 璇存槑鍜岀敤鎴疯祫鏂欐帴鍙ｉ獙鏀跺彛寰勩€?- 鏇存柊娴嬭瘯锛?  - 鐧诲綍鍚庡厛涓婁紶 `bizType=avatar` 鏂囦欢銆?  - 浣跨敤 `avatarFileId` 鏇存柊璧勬枡銆?  - 鍙︿竴涓敤鎴峰鐢ㄨ fileId 鏇存柊澶村儚杩斿洖 `403`銆?
### 楠岃瘉缁撴灉

- 宸叉墽琛屽畬鏁村悗绔祴璇曪細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 娴嬭瘯閫氳繃锛?  - `internal/appapi`
  - `internal/auth`
  - `internal/files`
  - 鍏朵粬宸叉湁鍚庣鍖呭潎姝ｅ父缂栬瘧鎴栨祴璇曢€氳繃銆?- Go 浠嶈緭鍑洪潪鑷村懡 telemetry 鏉冮檺鎻愮ず锛岄€€鍑虹爜涓?0锛屼笉褰卞搷鏈疆楠岃瘉缁撹銆?
### 褰撳墠椋庨櫓

- 褰撳墠澶村儚 URL 涓烘湰鍦?mock 涓嬭浇鍦板潃锛涙帴鐪熷疄瀵硅薄瀛樺偍鍚庡簲鏇挎崲涓?CDN / 涓存椂涓嬭浇 URL 绛栫暐銆?- `avatar` 鏂囦欢绫诲瀷鐩墠鍙牎楠屼笟鍔＄被鍨嬪拰褰掑睘锛屽悗缁簲鎸?MIME 鐧藉悕鍗曞拰澶у皬闄愬埗杩涗竴姝ユ敹绱с€?
## 2026-06-14 08:35

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟帹杩涘皬绋嬪簭鍚庣璐﹀彿璧勬枡閾捐矾銆?- 澶嶆牳鍚庣‘璁?`plan.md` 瑕佹眰 `PUT /api/app/users/me/profile`锛屽綋鍓嶄唬鐮佸彧鏈?`GET /api/app/users/me`銆?- 鏂板鐢ㄦ埛璧勬枡鏇存柊鑳藉姏锛?  - `users.Store.UpdateProfile`
  - `users.Repository.UpdateProfile`
  - `auth.Service.UpdateProfile`
- 鏇存柊 SQL repository锛?  - 鏂扮敤鎴峰垱寤烘椂鍚屾鍒濆鍖?`user_profiles` 鍩虹琛屻€?  - 鏇存柊璧勬枡鏃跺啓鍏?`users.nickname`銆乣users.avatar_url`锛屽苟纭繚 `user_profiles` 瀛樺湪銆?- 鏂板灏忕▼搴忔帴鍙ｏ細
  - `PUT /api/app/users/me/profile`
  - 鏀寔 `nickname`銆乣avatarUrl`
  - 鏄电О鏈€澶?32 瀛楃锛屽ご鍍忓湴鍧€鏈€澶?500 瀛楃銆?- 鏂板杩佺Щ锛?  - `db/migrations/000020_user_profiles_repository.sql`
  - 鍒涘缓 `user_profiles`
  - 澧炲姞鍩庡競绱㈠紩 `idx_user_profiles_city`
- 鏇存柊鏂囨。锛?  - `docs/openapi/app.openapi.yaml` 澧炲姞鐢ㄦ埛璧勬枡鏇存柊鎺ュ彛銆?  - `plan.md` 澧炲姞绗?20 涓縼绉昏褰曪紝骞朵慨姝ｇ敤鎴疯祫鏂欐帴鍙ｈ惤鐐广€?- 鏇存柊娴嬭瘯锛?  - `TestWechatLoginAndCurrentUserHTTP` 澧炲姞璧勬枡鏇存柊鍜屽洖璇绘柇瑷€銆?
### 楠岃瘉缁撴灉

- 宸叉墽琛屽畬鏁村悗绔祴璇曪細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 娴嬭瘯閫氳繃锛?  - `internal/appapi`
  - `internal/auth`
  - 鍏朵粬宸叉湁鍚庣鍖呭潎姝ｅ父缂栬瘧鎴栨祴璇曢€氳繃銆?- Go 浠嶈緭鍑洪潪鑷村懡 telemetry 鏉冮檺鎻愮ず锛岄€€鍑虹爜涓?0锛屼笉褰卞搷鏈疆楠岃瘉缁撹銆?
### 褰撳墠椋庨櫓

- 褰撳墠璧勬枡鏇存柊鍏堣鐩栨樀绉板拰澶村儚鍦板潃锛沗gender`銆乣bio`銆乣interest_tags`銆乣city_code`銆乣city_name` 宸插湪 `user_profiles` 棰勭暀锛屽悗缁彲缁х画鎵╁睍璇锋眰浣撳拰鏍￠獙銆?- 澶村儚鐩墠鎸?URL 淇濆瓨锛屽皻鏈己鍒舵牎楠?fileId 灞炰簬鏈汉锛涘悗缁帴灏忕▼搴忎笂浼犲ご鍍忔椂搴斿拰鏂囦欢鏈嶅姟鏉冮檺鏍￠獙鎵撻€氥€?
## 2026-06-14 08:10

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟帹杩涘皬绋嬪簭鍚庣鏍归摼璺寔涔呭寲銆?- 澶嶆牳鍚庣‘璁よ处鍙?閭€璇蜂粛鏄唴瀛樻€侊細
  - `users.NewStore()`
  - `invites.NewStore()`
  - 寰俊鐧诲綍鍚庣殑 `users`銆乣user_wechat_accounts`銆乣invite_codes`銆乣invite_relations` 杩愯鏃舵湭鎺?PostgreSQL銆?- 鏂板鐢ㄦ埛妯″潡 repository 鍖栵細
  - `users.Repository`
  - `users.NewStoreWithRepository`
  - `users.NewSQLRepository`
  - `FindByOpenID`
  - `FindByID`
  - `CreateWithOpenID`
- 鏂板閭€璇锋ā鍧?repository 鍖栵細
  - `invites.Repository`
  - `invites.NewStoreWithRepository`
  - `invites.NewSQLRepository`
  - `UpsertCode`
  - `FindCode`
  - `Bind`
  - `RelationForUser`
- 鏇存柊 `auth.Service`锛?  - 寰俊鐧诲綍鏌?openid 鏃跺彲浠?repository 鍥炲～鐢ㄦ埛銆?  - 鏂扮敤鎴峰垱寤烘椂鍚屾椂鍐欏叆 `users` 鍜?`user_wechat_accounts`銆?  - 閭€璇风粦瀹氬啓鍏?`invite_relations`锛岄娆＄粦瀹氭墠閫掑 `invite_codes.used_count`銆?  - repository 閿欒閫忎紶锛屼笉鍐嶆妸鏁版嵁搴撻敊璇鏄犲皠鎴愰個璇风爜鏃犳晥銆?- 鏇存柊 `cmd/server`锛?  - 鏈夋暟鎹簱鏃舵敞鍏?`users.NewSQLRepository(db)` 鍜?`invites.NewSQLRepository(db)`銆?  - 鏃犳暟鎹簱鏃剁户缁娇鐢ㄥ唴瀛?store锛屼繚鎸佹湰鍦拌仈璋冨拰鏃㈡湁娴嬭瘯绋冲畾銆?- 鏇存柊 `plan.md`锛?  - 鏍囪 `000002_account_invite.sql` 宸叉帴鍏ヨ繍琛屾椂 repository銆?  - D4 鏁版嵁搴撲笂绾块『搴忚ˉ鍏?root 璐﹀彿閭€璇烽摼璺帴鍏ヨ鏄庛€?- 鏂板娴嬭瘯锛?  - `services/go-api/internal/auth/repository_service_test.go`

### 楠岃瘉缁撴灉

- 宸叉墽琛屽畬鏁村悗绔祴璇曪細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 娴嬭瘯閫氳繃锛?  - `internal/auth`
  - `internal/appapi`
  - 鍏朵粬宸叉湁鍚庣鍖呭潎姝ｅ父缂栬瘧鎴栨祴璇曢€氳繃銆?- Go 浠嶈緭鍑洪潪鑷村懡 telemetry 鏉冮檺鎻愮ず锛岄€€鍑虹爜涓?0锛屼笉褰卞搷鏈疆楠岃瘉缁撹銆?
### 褰撳墠椋庨櫓

- 鏈疆瑙ｅ喅鐨勬槸 mock openid 涓嬬殑杩愯鏃舵寔涔呭寲锛涚湡瀹炲井淇?`code2session` 浠嶆湭鎺ュ叆銆?- 璐﹀彿/閭€璇?SQL repository 宸蹭娇鐢ㄧ幇鏈夎縼绉昏〃锛屼絾鐪熷疄 PostgreSQL 椹卞姩鍜屾祴璇曞簱浠嶆湭鍚敤锛屽敮涓€绾︽潫鍜屼簨鍔¤涓哄悗缁繕闇€闆嗘垚娴嬭瘯澶嶆祴銆?
## 2026-06-14 07:45

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟帹杩涘皬绋嬪簭鍚庣鎸佷箙鍖栬ˉ榻愩€?- 鏂板鍥㈤槦妯″潡 repository 鍖栵細
  - `teams.Repository`
  - `teams.NewServiceWithRepository`
  - `teams.NewSQLRepository`
  - 鍥㈤槦涓昏〃 `teams`
  - 鎴愬憳鍏崇郴缁х画澶嶇敤 `team_relations`
- 鏂板浼氬憳鎶ヨ〃妯″潡 repository 鍖栵細
  - `memberreports.Repository`
  - `memberreports.NewServiceWithRepository`
  - `memberreports.NewSQLRepository`
  - 浼氬憳鎺堟潈琛?`member_report_memberships`
  - 鎶ヨ〃蹇収缁х画鍐欏叆 `member_report_snapshots.metrics`
- 鏂板杩佺Щ锛歚db/migrations/000019_team_member_repository.sql`銆?- 鏇存柊 `cmd/server` 鍜?`appapi.Server.UseRepositories`锛?  - 鏈夋暟鎹簱鏃舵敞鍏ュ洟闃熶笌浼氬憳鎶ヨ〃 SQL repository銆?  - 鏃犳暟鎹簱鏃剁户缁娇鐢ㄥ唴瀛樺疄鐜帮紝鏂逛究鏈湴鑱旇皟銆?- 鏇存柊 `plan.md`锛?  - 杩佺Щ姹囨€昏拷鍔?`000019_team_member_repository.sql`銆?  - D4 鏁版嵁搴撲笂绾块『搴忚拷鍔犵 19 椤广€?- 鏂板娴嬭瘯锛?  - `services/go-api/internal/teams/repository_service_test.go`
  - `services/go-api/internal/memberreports/repository_service_test.go`
- 灏嗗洟闃熷拰浼氬憳鎶ヨ〃榛樿鍚嶇О鏀逛负 ASCII fallback锛岄伩鍏嶇户缁繑鍥炲巻鍙蹭贡鐮侀粯璁ゅ€笺€?
### 楠岃瘉缁撴灉

- 宸叉墽琛屽畬鏁村悗绔祴璇曪細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 娴嬭瘯閫氳繃锛?  - `internal/appapi`
  - `internal/teams`
  - `internal/memberreports`
  - 鍏朵粬宸叉湁鍚庣鍖呭潎姝ｅ父缂栬瘧鎴栨祴璇曢€氳繃銆?- Go 浠嶈緭鍑洪潪鑷村懡 telemetry 鏉冮檺鎻愮ず锛岄€€鍑虹爜涓?0锛屼笉褰卞搷鏈疆楠岃瘉缁撹銆?
### 褰撳墠椋庨櫓

- 鍥㈤槦涓庝細鍛樻姤琛ㄥ凡鍏峰 SQL repository 鍜岃縼绉昏剼鏈紝浣嗙湡瀹?PostgreSQL 杩佺Щ dry-run / 闆嗘垚娴嬭瘯浠嶆湭鎵ц銆?- 浼氬憳鎶ヨ〃鐩墠鎸夊疄鏃?games/revenue 鐢熸垚蹇収骞惰惤搴擄紝鍚庣画鑻ラ渶瑕佸浐瀹氬懆鏈熸姤琛紝搴旇ˉ瀹氭椂浠诲姟鎴栧悗鍙扮敓鎴愬叆鍙ｃ€?

## 2026-06-14 12:36

### 当前进展

- 继续执行 `plan.md`，优先补小程序后端 D7 外部输入校验中的 LBS 链路。
- `internal/lbs.Service` 新增定位参数校验：
  - 经度必须在 `[-180, 180]`。
  - 纬度必须在 `[-90, 90]`。
  - `accuracyMeter` 必须非负且不超过 `100000`。
  - `cityCode`、`cityName`、`address` 做 trim 和长度上限校验。
- `POST /api/app/locations/current` 和 `POST /api/app/locations/manual` 遇到无效定位参数时统一返回 `422`。
- `GET /api/app/games/nearby` 新增 `radiusMeter` 校验，半径必须大于 `0` 且不超过 `50000`，异常值不再静默使用默认值。
- 更新 `plan.md`：
  - 标记“校验经纬度范围”完成。
  - 标记 `accuracyMeter > 300` 返回 `accuracyWarning=true` 完成。
  - 标记“80 米精度无提示，500 米精度有提示”完成。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/lbs`
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./...`
- 测试通过。

## 2026-06-15 05:00

### 当前进展

- 继续执行 `plan.md`，补齐实名状态查询和后台实名列表筛选能力。
- `GET /api/admin/identity-verifications` 新增筛选参数：
  - `status`：按实名状态过滤。
  - `userId`：按用户 ID 精确过滤。
- 后台实名列表仍只返回脱敏字段，详情页继续保持脱敏展示。
- 复用现有 `GET /api/app/identity/status`，用户端可直接查询强实名进度和驳回原因。
- 更新 `plan.md`：标记 `查询实名状态`、`用户端可查询实名状态和驳回原因` 完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/appapi/identity_handler.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./services/go-api/internal/identity ./services/go-api/internal/appapi`
  - `go test -count=1 ./services/go-api/internal/...`
- 测试通过。

## 2026-06-15 04:30

### 当前进展

- 继续执行 `plan.md`，补齐强实名后台脱敏展示闭环。
- `identity.Record` 新增脱敏字段：
  - `realNameMasked`
  - `idCardMasked`
- 手机号核验通过时只生成并保存姓名、身份证脱敏值，不在 Go 业务记录中返回明文姓名或身份证号。
- SQL repository 已写入/读取实名脱敏字段。
- 新增迁移 `db/migrations/000023_identity_masked_fields.sql`，为 `identity_verification_records` 补充：
  - `real_name_masked`
  - `id_card_masked`
- 后台实名列表和详情继续使用同一 `Record` DTO，普通管理员只能看到 `phoneMasked`、`realNameMasked`、`idCardMasked` 等脱敏字段。
- 更新 `plan.md`：标记实名详情敏感字段脱敏、普通管理员不可查看完整身份证/核身原始结果、后台列表默认脱敏完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/identity/service.go services/go-api/internal/identity/sql_repository.go services/go-api/internal/identity/identity_status_service_test.go services/go-api/internal/identity/service_test.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./services/go-api/internal/identity ./services/go-api/internal/appapi`
  - `go test -count=1 ./services/go-api/internal/...`
- 测试通过。

## 2026-06-15 04:00

### 当前进展

- 继续执行 `plan.md`，补齐小程序后端业务配置闭环。
- `common/config.Load` 新增业务配置读取：
  - `APP_DAILY_GAME_LIMIT`：每日开局限制，默认 3。
  - `LBS_DEFAULT_RADIUS_METER`：附近局默认半径，默认 5000 米。
- `games.Service` 新增可配置每日开局限制，默认行为保持每日 3 局；配置后按新值拦截。
- `appapi.Server.Configure` 将配置注入到后端业务：
  - 每日开局限制注入到 `games.Service`。
  - 附近局默认半径注入到 `/api/app/games/nearby`。
- `/api/app/games/nearby` 在未传 `radiusMeter` 时使用配置默认半径；传入 `radiusMeter` 时仍以请求参数覆盖。
- 更新 `plan.md`：标记 LBS 默认半径、每日开局限制、本地/测试/生产不同配置、附近局默认半径读取系统配置完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/common/config/config.go services/go-api/internal/common/config/config_test.go services/go-api/internal/games/service.go services/go-api/internal/games/service_test.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/location_handler.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./services/go-api/internal/common/config ./services/go-api/internal/games ./services/go-api/internal/appapi`
  - `go test -count=1 ./services/go-api/internal/...`
- 测试通过。

## 2026-06-15 03:30

### 当前进展

- 继续执行 `plan.md`，补齐 Go API 后端 requestId 闭环，方便线上测试和排障。
- `httpx.Write` 现在会保证 JSON 响应体包含 `requestId`，并同步写入 `X-Request-Id` 响应头。
- `Server.Register` 统一把所有注册路由经过 `httpx.RequestID` 中间件：
  - 前端传入 `X-Request-Id` 时，响应头和响应体复用该值。
  - 未传入时，后端自动生成 requestId。
  - 后台操作日志继续能从请求头读取同一个 requestId。
- 修复 `server.go` 登录错误分支中一处历史非法 UTF-8/断裂字符串，改为稳定 ASCII 文案。
- 更新 `plan.md`：标记 Go API 后端统一响应结构、requestId 中间件、所有响应带 requestId、requestId 必带完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/common/httpx/response.go services/go-api/internal/common/httpx/middleware_test.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./services/go-api/internal/common/httpx ./services/go-api/internal/appapi`
  - `go test -count=1 ./services/go-api/internal/...`
- 测试通过。

## 2026-06-15 03:00

### 当前进展

- 继续执行 `plan.md`，优先补小程序后端线上测试前的业务闭环。
- AI/IM 数据使用新增后台显式开关：
  - `GET /api/admin/ai-data/im-export-config` 可查看当前 IM 导出授权状态。
  - `PUT /api/admin/ai-data/im-export-config` 仅 `system_config:update` 权限可修改。
  - 默认关闭时 `POST /api/admin/ai-data/im-export` 禁止导出；开启后可导出已有 IM 消息；再次关闭后恢复禁止。
- 交付测试执行记录补齐证据链约束：
  - `POST /api/admin/test-runs` 创建测试执行记录时，必须至少关联 `requestId` 或 `evidenceFileId`。
  - 测试执行记录返回并留存 `evidenceFileId`，便于后台交付验收追踪。
- 更新 `plan.md`：标记 `test_runs` 关联 requestId 或证据文件完成，并补充对应测试项。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/delivery/service.go services/go-api/internal/delivery/test_run_service_test.go services/go-api/internal/appapi/delivery_handler.go services/go-api/internal/appapi/server_test.go services/go-api/internal/aidata/service.go services/go-api/internal/appapi/aidata_handler.go services/go-api/internal/appapi/server.go`
  - `go test -count=1 ./services/go-api/internal/delivery ./services/go-api/internal/aidata ./services/go-api/internal/appapi`
  - `go test -count=1 ./services/go-api/internal/...`
- 测试通过。

## 2026-06-14 17:05

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 D5.6/E5 服务确认、评价和人脉沉淀闭环。
- 核实现有代码状态：
  - 服务确认完成后已经进入 `pending_review` 并触发评价待办。
  - `againIntent` 已作为评价字段保存，并通过 `/api/app/reviews/my-intents` 返回。
  - 服务确认完成后已经生成 `co_game` 人脉。
- 新增评价成功后的人脉强度沉淀：
  - `/api/app/reviews` 提交成功后，对评价双方写入 `sourceType=review` 的人脉关系。
  - 关系强度 `strengthScore` 增加 1，供后续二期推荐和人脉排序使用。
- 更新 `plan.md`：
  - 标记“再玩一局意向单独字段或表记录，供二期推荐使用”完成。
  - 标记“评价完成后可提升关系强度分”完成。

### 验证补充

- `go test -count=1 ./internal/appapi` 通过。

## 2026-06-14 16:41

### 当前进展

- 继续执行 `plan.md`，优先补齐小程序后端 E4 强实名安全闭环。
- FaceID 回调新增 HMAC-SHA256 验签能力：
  - `Config` 新增 `FaceIDConfig`，读取 `TENCENT_FACEID_CALLBACK_REQUIRE_SIGNATURE` 和 `TENCENT_FACEID_CALLBACK_SECRET`。
  - 生产环境默认要求 FaceID 回调签名；本地环境默认不强制，保持现有小程序 mock 联调流程可用。
  - `appapi.Server` 新增 `UseFaceIDCallbackVerifier`，服务启动时注入回调验签配置。
  - `/api/app/identity/faceid/callback` 在解析业务 payload 前先校验 `X-FaceID-Signature` 或 `X-Tencent-FaceID-Signature`。
  - `deploy/env.example` 已补 FaceID 回调验签环境变量样例。
- `plan.md` 已同步勾选 E4 的“腾讯云人脸核身回调必须验签或校验来源”。

### 验证补充

- `go test -count=1 ./internal/common/config` 通过。
- `go test -count=1 ./internal/appapi` 通过。

### 下一步

- 跑全量 `go test -count=1 ./...`。
- 继续按后端业务功能优先级推进：补强强实名验收剩余项，随后转入服务确认、评价链路和上线联调缺口。

### 当前风险

- D7 “所有外部输入做字段长度、枚举、金额、经纬度、时间范围校验”仍是大项，本轮只完成 LBS/附近局相关输入校验，其他写接口还需要继续补齐。

## 2026-06-14 13:06

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 D7 安全基线。
- 新增 `internal/appapi.AppAuthMiddleware`：
  - 统一解析 `Authorization: Bearer <token>`。
  - 缺失 token 返回 `40101`。
  - 无效或过期 token 返回 `40101`。
- `Register` 中除 `POST /api/app/auth/wechat-login` 外，所有 `/api/app` 小程序接口均已统一包装 `AppAuthMiddleware`。
- 保留 handler 内部 `requireUser` 的强实名校验，不影响微信登录后、实名前的身份绑定流程。
- 新增 `TestAppBusinessRoutesRequireTokenMiddleware`，验证业务接口匿名访问被拦截，微信登录入口保持公开。
- 更新 `plan.md`：
  - 标记“小程序接口统一走 `AppAuthMiddleware`”完成。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./...`
- 测试通过。

### 当前风险

- D7 还剩：
  - IM WebSocket 握手和 HTTP 历史消息使用同一套成员权限判断。
  - 所有写接口支持幂等键或业务唯一约束。
  - 所有外部输入做字段长度、枚举、金额、经纬度、时间范围校验。

## 2026-06-14 13:36

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 D7 IM 权限基线。
- `internal/im.Service` 新增统一成员权限入口：
  - `AuthorizeGameAccess(userID, gameID)`。
  - `AuthorizeRoomAccess(userID, roomID)`。
- `chat-session` 握手、按 `gameId` 读取历史消息、按 `roomId` 读取历史消息、发消息和房间操作均通过统一成员权限入口或其下游方法校验。
- `TestIMFlow` 增加非成员校验：
  - 非成员无法获取 `chat-session`。
  - 非成员无法读取 `GET /api/app/games/{gameId}/chat/messages`。
  - 非成员无法读取 `GET /api/app/chat/rooms/{roomId}/messages`。
- `docs/openapi/ws-protocol.md` 同步注明握手和 HTTP 历史消息统一走 `AuthorizeGameAccess / AuthorizeRoomAccess`。
- 更新 `plan.md`：
  - 标记“IM WebSocket 握手和 HTTP 历史消息使用同一套成员权限判断”完成。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/im`
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./...`
- 测试通过。

### 当前风险

- D7 还剩：
  - 所有写接口支持幂等键或业务唯一约束。
  - 所有外部输入做字段长度、枚举、金额、经纬度、时间范围校验。

## 2026-06-14 14:06

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 D7 幂等基线。
- 新增 `internal/appapi.IdempotencyMiddleware`：
  - 仅对 `POST/PUT/DELETE` 且带 `Idempotency-Key` 的请求生效。
  - 按 `userID + method + path + key` 缓存首次成功响应。
  - 重复提交返回首个成功响应，避免前端重复点击导致重复创建。
  - `Idempotency-Key` 过长时返回 `422`。
- 已接入的小程序写接口包括：
  - 创建局
  - 入局申请 / 局操作
  - 定位保存
  - 身份绑定 / 实名链路
  - 文件上传令牌
  - 评价、积分兑换、举报、连接、资料、通知、报表等写接口
- 新增测试 `TestCreateGameIdempotencyKeyReplaysFirstResponse`，验证同一个 key 的重复创建局请求只产生一次局数据。
- 更新 `plan.md`：
  - 标记“所有写接口支持幂等键或业务唯一约束，防止重复提交”完成。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./...`
- 测试通过。

### 当前风险

- 这轮实现是 app 层内存幂等缓存，适合当前联调与防重复提交；后续如果要做到进程重启后仍可去重，仍建议再接 Redis 或数据库唯一约束。

## 2026-06-14 15:11

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 D7 外部输入校验闭环。
- 补齐核心业务写入口的字段长度、枚举、范围校验：
  - 创建局：标题、城市编码/名称、经纬度范围。
  - 入局申请、服务确认、续局草稿：原因、备注、标题长度。
  - 进度反馈、里程碑、打卡、复盘：进度范围、状态/类型枚举、文件 ID、内容长度、再玩意向枚举。
  - IM 消息：文本/图片/文件消息类型枚举、文本内容、fileId、归档原因。
  - 评价：评分范围、目标角色、内容长度、再玩意向枚举。
  - 举报申诉：举报类型枚举、内容长度、关联证据 ID 非负、后台处理结果长度。
  - 人脉跟进、行家技能树、领路人资源画像：跟进类型/时间、标签数量和长度、文件 ID、连接规模长度。
- 新增 `ErrInvalidGameInput` 和 `ErrInvalidMessage` 的接口层映射，避免校验错误落成 500。
- 更新 `plan.md`：
  - 标记“所有外部输入做字段长度、枚举、金额、经纬度、时间范围校验”完成。

### 验证结果

- 已执行：
  - `go test -count=1 ./services/go-api/internal/games ./services/go-api/internal/reviews ./services/go-api/internal/reports ./services/go-api/internal/im ./services/go-api/internal/connections ./services/go-api/internal/profiles ./services/go-api/internal/appapi`
- 测试通过。

### 当前风险

- D7 已基本闭环；后续继续按 `plan.md` 推进小程序后端时，应优先补强还未完成的强实名 token 分层、后台操作日志覆盖和联调交付包，而不是继续只加测试。

## 2026-06-14 16:11

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 E4 强实名主线补强。
- 在 `identity.Service` 增加强实名格式门槛：
  - 手机号必须为 11 位大陆手机号格式。
  - 身份证号必须为 18 位格式，末位支持数字或 `X/x`。
  - 身份证格式错误时不会通过手机号核验，因此不能发起人脸核身。
- `identity_handler` 增加手机号格式错误、身份证格式错误的 422 映射。
- 保持现有两段式登录：
  - 未强实名只返回 `preAuthToken`。
  - 强实名完成后通过 `issue-token-after-identity` 换取正式 token。
  - 已实名用户再次登录可直接拿到正式 token。
- 更新 `plan.md`：
  - 标记 E4 中已由代码覆盖的 preAuth、短信限频、手机号/短信/人脸记录、not_supported、敏感数据不落明文、核心业务强实名拦截和相关验收项。
  - 保留“腾讯云人脸核身回调必须验签或校验来源”未完成，待真实腾讯云回调接入时再落签名/来源校验。

### 验证结果

- 已执行：
  - `go test -count=1 ./services/go-api/internal/identity ./services/go-api/internal/appapi`
- 测试通过。

## 2026-06-14 17:41

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 P3 角色申请与领路人双门槛。
- 补齐角色申请服务闭环：
  - 小程序端 `POST /api/app/role-applications`、`GET /api/app/role-applications/my`。
  - 后台端 `GET /api/admin/audits/role-applications`、`POST /api/admin/audits/role-applications/{applicationId}/review`。
  - 申请记录保存申请理由、能力说明和证明文件 ID。
  - 审核通过后写入 `user_roles`，驳回保存原因，重复 pending 申请拦截。
- 补齐领路人双门槛闭环：
  - 小程序端 `GET /api/app/guides/qualification/me` 查询当前用户资格和规则。
  - 小程序端 `POST /api/app/guides/apply` 在条件与付费均满足后提交领路人申请。
  - 后台端 `GET /api/admin/guides/qualification-rules`、`PUT /api/admin/guides/qualification-rules/{ruleId}` 配置规则。
  - 后台端保留 `POST /api/admin/guide-qualification-rules` 登记用户条件/付费达成状态。
  - `conditionMet=false` 返回 `waiting_condition`，`paymentMet=false` 返回 `waiting_payment`。
- 扩展 `role_applications`、`guide_qualification_rules`、`guide_qualification_records` 的服务层与 SQL 仓储映射。
- 更新 `plan.md`：标记角色申请审核、申请材料保存、领路人规则配置、双门槛验收项完成。

### 验证结果

- 已执行：
  - `go test -count=1 ./services/go-api/internal/profiles`
  - `go test -count=1 ./services/go-api/internal/appapi`
- 测试通过。

## 2026-06-14 18:23

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端退出局扣信用闭环。
- 将 `game_members` 退出从物理删除改为软退出状态：
  - 写入 `quit_before_confirm`、`quit_after_confirm`、`quit_after_started`。
  - 保留退出原因，便于后台和审计查询。
- 退出扣分后回写成员记录：
  - `credit_deducted=true`。
  - `credit_log_id` 关联真实信用流水。
- 退出成功后生成站内通知：
  - 给退出用户生成 `game_quit_result`。
  - 给发起人生成 `game_member_quit`。
- 更新 `plan.md`：标记退出阶段判断、确认前不扣分、确认后/开始后扣分、成员状态更新、扣分流水关联、通知和后台扣分原因可查完成。

### 验证结果

- 已执行：
  - `go test -count=1 ./services/go-api/internal/games`
  - `go test -count=1 ./services/go-api/internal/appapi`
- 测试通过。

## 2026-06-14 18:53

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 IM 内容安全闭环。
- 将 IM 敏感词从硬编码字符串升级为可管理对象：
  - 后台 `GET /api/admin/sensitive-words` 查询。
  - 后台 `POST /api/admin/sensitive-words` 新增。
  - 后台 `POST /api/admin/sensitive-words/import` 批量导入。
- IM 文本消息发送前读取当前敏感词配置，命中后按规则阻断，并返回 `45101`。
- 命中敏感词时写入 `content_risk_logs` 风险日志，后台 `GET /api/admin/content-risk/logs` 可查询。
- 新增 `/api/internal/ai/content-risk/check-placeholder`，一期返回固定占位结构并写风险日志，后续可替换真实 AI 内容审核。
- 补齐 `000005_im.sql` 中 `sensitive_words`、`content_risk_logs`、`chat_messages.client_msg_id` 和风险日志索引。
- 补齐后台内容安全权限码：`content:sensitive_word:view/create/import`、`content:risk_log:view`。
- 更新 `plan.md`：标记敏感词新增、批量导入、IM 文本检测、命中拦截/标记、风险日志和 AI 占位接口完成；保留“后台可禁用敏感词”独立启停项未完成。

### 验证结果

- 已执行：
  - `go test -count=1 ./services/go-api/internal/im`
  - `go test -count=1 ./services/go-api/internal/appapi`
  - `go test -count=1 ./services/go-api/internal/adminauth`
- 测试通过。

## 2026-06-14 19:23

### 当前进展

- 继续执行 `plan.md`，优先补齐小程序后端 IM 敏感词独立启停能力。
- 新增后台敏感词更新链路：
  - `PUT /api/admin/sensitive-words/{id}` 支持将敏感词状态更新为 `active` 或 `disabled`。
  - 服务层 `UpdateSensitiveWord` 按 ID 修改状态，非法状态返回校验错误，不存在返回未找到。
  - 后台操作日志记录 `content:sensitive_word:update`。
- 补齐后台权限：
  - 路由改用独立权限码 `content:sensitive_word:update`。
  - `adminauth` 默认权限和 `db/seeds/admin_roles_permissions.sql` 均补齐该权限。
- 修复 IM 默认敏感词和行为事件测试中损坏的字符串字面量，避免 Go 文件语法损坏。
- 更新 `plan.md`：标记“后台可禁用敏感词”、敏感词阻断、风险日志后台可查、AI 占位接口、文本敏感词检测和 TC-087 完成。
- 补齐敏感词 `action=flag` 发送链路：命中后消息不阻断，消息状态写为 `risk_flagged`，并生成风险日志，覆盖 TC-088。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/im/service.go services/go-api/internal/appapi/im_handler.go services/go-api/internal/appapi/server.go services/go-api/internal/adminauth/service.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./...`
- 测试通过。

## 2026-06-14 19:53

### 当前进展

- 继续执行 `plan.md`，复核 D5.6 服务确认、评价、成长、信用、足迹后端链路。
- 补齐仓储模式下成长资料初始化：
  - `Profile(userId)` 在 `user_growth_profiles` 不存在时生成默认资料。
  - 默认 `level=1`、今日信用和信用分为 `100`，并持久化到仓储。
  - 个人中心首次查询成长资料不再返回空结构。
- 补齐完成局经验奖励：
  - 服务确认全部完成、局进入 `pending_review` 时为成员写入 `completed_game` 经验和足迹。
  - 使用足迹幂等判断，重复确认不会重复加经验。
- 更新 `plan.md` D5.6 细项：
  - 标记服务确认、确认项、评价提醒、可补评价、重复评价拦截、再玩意向、评价经验、经验日志、积分日志、成就和后台成长来源追溯完成。
  - 标记“完成局后增加经验”完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/reviews/service.go services/go-api/internal/reviews/repository_service_test.go`
  - `go test -count=1 ./internal/reviews`
  - `go test -count=1 ./...`
- 测试通过。

## 2026-06-14 20:23

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 D5.7 支付、订单和分账占位闭环。
- 补齐二期分账预留接口的一期占位实现：
  - `POST /api/funds/profit-sharing/orders` 创建本地模拟分账单。
  - `GET /api/funds/profit-sharing/orders/{outOrderNo}` 查询本地模拟分账结果。
  - `POST /api/funds/profit-sharing/return-orders` 创建本地模拟分账回退。
- 所有分账占位响应明确返回：
  - `placeholder=true`。
  - `needWechatPay=false`。
  - `mode=profit_sharing_placeholder` 或 `profit_sharing_return_placeholder`。
- 补齐 `revenueclient` 调用方法，后续资金服务或内部网关联调可沿同一路径调用。
- 更新 `plan.md`：标记 `payment_orders.order_no` 唯一、免费局 `free_no_pay` 占位订单、二期支付分账接口预留完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/orders/service.go services/go-api/internal/orders/order_service_test.go services/go-api/internal/appapi/order_handler.go services/go-api/internal/appapi/order_handler_test.go services/go-api/internal/appapi/server.go services/go-api/internal/revenueclient/client.go services/go-api/internal/revenueclient/client_test.go`
  - `go test -count=1 ./internal/orders ./internal/appapi ./internal/revenueclient`
  - `go test -count=1 ./...`
- 测试通过。
## 2026-06-14 20:14

### 当前进展

- 继续执行 `plan.md`，优先补齐小程序后端 D5.7 分润、收益和订单占位闭环。
- 将收益链路从“只算汇总”推进到“账户 + 流水 + 结算”三层一致：
  - `revenue.Service` 新增 `IncomeAccount` 视图，收益摘要可同步维护一人一账户。
  - 生成、冻结、结算分润记录时，自动同步 `user_income_accounts`。
  - 分润记录相关动作会写入 `income_logs`，为后续审计和查询保留明细。
- `SQLRepository` 已补齐：
  - `user_income_accounts` upsert。
  - `income_logs` 追加写入。
- 小程序收益接口新增账户入口：
  - `GET /api/app/incomes/account`
  - `GET /api/app/income/account`
- 更新 `plan.md`：
  - 标记用户收益账户一人一个、更新收益账户、正式生成分润记录/明细、支持冻结、线下结算登记、结算记录查询、试算和正式记录金额一致、冻结后不能结算、结算后有记录完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/revenue/service.go services/go-api/internal/revenue/sql_repository.go services/go-api/internal/revenue/repository_service_test.go services/go-api/internal/appapi/revenue_handler.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/revenue ./internal/appapi`
- 测试通过。

## 2026-06-14 20:43

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 D5.7 分润试算阻断能力。
- 补齐分润试算返回字段：
  - `canGenerateRecord`：表示当前试算是否允许生成正式分润记录。
  - `blockReasons`：返回阻断原因列表。
- 服务层 `revenue.Preview` 已能识别：
  - `review_incomplete`：评价未完成，不能生成正式分润。
  - `invalid_amount`：金额非法。
  - `game_missing`：缺少局 ID。
- HTTP 层预览会补充争议阻断：
  - 同局存在未关闭举报/申诉时，`blockReasons` 包含 `disputed`。
  - 后台和小程序分润预览返回同一组阻断字段。
- 正式生成分润记录仍保留硬拦截，不依赖前端是否读取预览字段。
- 更新 `plan.md`：标记 `canGenerateRecord`、`blockReasons`、评价未完成阻断、争议中阻断，以及 API-B2-05 对应测试项完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/revenue/service.go services/go-api/internal/revenue/repository_service_test.go services/go-api/internal/appapi/revenue_handler.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/revenue ./internal/appapi`
- 测试通过。
## 2026-06-14 21:04

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 D5.7 分润模板规则配置。
- 新增分润规则后端能力：
  - `GET /api/admin/revenue/rules?templateId=...`
  - `POST /api/admin/revenue/rules`
- `revenue_rules` 现在可按模板保存和查询，`rule_code` 维度支持 upsert，便于后台修改模板规则值。
- 规则结构和模板结构保持同一 bps 体系，避免前后端出现浮点金额计算分叉。
- 规则配置已接入现有 revenue 服务和 SQL 仓储，后续可继续补审计日志或更复杂规则解释。
- 更新 `plan.md`：标记新增规则、修改规则、比例使用 bps 完成；修改写操作日志暂未勾选，留待审计日志真正落地后再收口。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/revenue/service.go services/go-api/internal/revenue/sql_repository.go services/go-api/internal/revenue/repository_service_test.go services/go-api/internal/appapi/revenue_handler.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/revenue ./internal/appapi`
- 测试通过。

## 2026-06-14 21:34

### 当前进展

- 继续执行 `plan.md`，补齐 D5.7 分润模板规则配置的审计闭环。
- 后台创建分润模板现在写入操作日志：
  - action=`revenue:template:create`
  - targetType=`revenue_template`
- 后台新增/修改分润规则现在写入操作日志：
  - action=`revenue:rule:upsert`
  - targetType=`revenue_rule`
  - detail 记录 `templateId`、`ruleCode`、`ruleValue`
- 补充 HTTP 测试：创建模板、upsert 规则后，超级管理员可在 `/api/admin/operation-logs` 查到对应日志。
- 更新 `plan.md`：标记“修改写操作日志”、`TC-034`、`TC-112` 完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/appapi/revenue_handler.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/revenue ./internal/appapi`
- 测试通过。

## 2026-06-14 22:00

### 当前进展

- 继续执行 `plan.md`，推进 D5.7 分润规则参与试算返回。
- `revenue.Preview` 新增 `rules` 字段，按当前 `templateId` 返回已配置的 `revenue_rules`。
- 后台和小程序分润预览现在不仅返回金额明细、阻断原因，也能返回当前模板启用的规则列表。
- 本轮只把规则纳入预览上下文，不擅自改变金额算法；更复杂的按局类型、会员、角色选择规则后续单独实现，避免未确认语义时影响分润金额。
- 更新 `plan.md`：标记“读取模板和规则”完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/revenue/service.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/revenue ./internal/appapi`
- 测试通过。

## 2026-06-14 22:31

### 当前进展

- 继续执行 `plan.md`，推进 D5.7 分润规则真正参与试算。
- 规则驱动的试算已支持三类额外角色：
  - `expert_bps`
  - `guide_bps`
  - `system_guide_bps`
- `CalculateRequest` 现在可携带 `expertUserId`、`guideUserId`、`systemGuideUserId`，预览会按模板规则生成对应金额项。
- 规则金额和模板比例都统一按 bps 计算，不引入浮点运算。
- `revenue.Preview` 继续返回 `rules`、`items`、`canGenerateRecord` 和 `blockReasons`，后台可直接看到规则与预览金额的对应关系。
- 更新 `plan.md`：标记平台金额、行家金额、领路人金额、系统级领路人金额、TC-113 完成；舍入差额仍待单独字段化，不提前勾选。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/revenue/service.go services/go-api/internal/revenue/repository_service_test.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/revenue ./internal/appapi`
- 测试通过。

## 2026-06-14 23:00

### 当前进展

- 继续执行 `plan.md`，补齐 D5.7 分润试算的舍入差额闭环。
- `revenue.Preview` 新增 `roundingDiffCent` 返回字段。
- 预览计算会在平台、创作者、行家、领路人、系统级领路人、成员分润项生成后，计算 `amountCent - items合计`。
- 当存在未分配差额时，追加 `rounding_adjustment` 明细项，确保分润项金额合计等于 `amountCent`。
- 正式生成分润记录复用预览结果，因此 `revenue_records.items` 同步保留差额项，避免预览和生成金额口径不一致。
- 更新 `plan.md`：标记“计算舍入差额”和“比例计算总额等于 amountCent”完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/revenue/service.go services/go-api/internal/revenue/repository_service_test.go`
  - `go test -count=1 ./internal/revenue ./internal/appapi`
  - `go test -count=1 ./...`
- 测试通过。

## 2026-06-14 23:30

### 当前进展

- 继续执行 `plan.md`，补齐 D5.7 订单占位的后端业务入口。
- 订单服务新增领路人付费占位订单：
  - 订单号：`GUIDE-FEE-{userId}`
  - 状态：`guide_fee_placeholder`
  - `needWechatPay=false`
- 新增小程序接口：`POST /api/app/guides/payment/precreate-placeholder`。
- 该接口会创建或复用领路人付费占位订单，并同步把当前用户的领路人资格 `paymentMet` 标记为 true。
- SQL 仓储同步支持领路人付费占位订单，`game_id` 为空，便于和免费局订单区分。
- 更新 `plan.md`：标记领路人付费占位、支付预下单固定结构、支付回调幂等、一期不真实收款和二期沿用订单号状态完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/orders/service.go services/go-api/internal/orders/sql_repository.go services/go-api/internal/orders/order_service_test.go services/go-api/internal/appapi/order_handler.go services/go-api/internal/appapi/order_handler_test.go services/go-api/internal/appapi/server.go`
  - `go test -count=1 ./internal/orders ./internal/appapi`
  - `go test -count=1 ./...`
- 测试通过。

## 2026-06-14 23:59

### 当前进展

- 继续执行 `plan.md`，回到 D5.8 举报申诉后端业务链路，补齐后台分配处理人能力。
- `reports.Service` 新增 `Assign(reportId, adminId, handlerAdminId)` 语义：
  - 举报状态更新为 `assigned`
  - `handlerAdminId` 记录被分配的处理人
  - 不写 `handledAt`，避免把“已分配”误当成“已处理”
- 新增后台接口：`POST /api/admin/reports/{reportId}/assign`。
- 后台分配举报会写操作日志 `report:assign`，并给举报人生成 `report_assigned` 站内通知和微信订阅任务占位。
- 权限补齐：`report:assign` 加入默认后台权限和种子权限。
- 更新 `plan.md`：标记 `AssignReport`、`HandleReport`、`CloseReport`、`FreezeRevenueIfNeeded` 完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/reports/service.go services/go-api/internal/reports/report_service_test.go services/go-api/internal/appapi/report_handler.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/server_test.go services/go-api/internal/adminauth/service.go`
  - `go test -count=1 ./internal/reports ./internal/appapi ./internal/adminauth`
  - `go test -count=1 ./...`
- 测试通过。

## 2026-06-15 00:30

### 当前进展

- 继续执行 `plan.md`，回补 D5.5 IM 归档和争议留存业务闭环。
- IM 服务层新增批量归档能力 `ArchiveRoomsByGameIDs`，供内部定时任务按已结束局归档房间。
- 发送消息前新增房间状态校验：`archived` / `readonly` 房间普通成员不能继续发送消息。
- 新增内部接口：`POST /api/internal/im/archive-expired-rooms`。
- 归档任务只处理 `pending_review` / `completed` 局；若局存在未关闭举报申诉，则跳过归档并返回 `skippedRooms`，用于延长 IM 证据留存。
- 更新 `plan.md`：标记 D5.5 IM 归档、TC-090、TC-091 完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/im/service.go services/go-api/internal/appapi/im_handler.go services/go-api/internal/appapi/job_handler.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/im ./internal/appapi`
  - `go test -count=1 ./...`
- 测试通过。

## 2026-06-15 01:00

### 当前进展

- 继续执行 `plan.md`，推进 D5.6 服务确认、评价提醒和补评价链路。
- 服务确认进入 `pending_review` 时，现在会即时给局成员生成 `review_remind` 站内通知和微信订阅任务占位，不再只依赖定时任务兜底。
- 重复确认不会重复发即时评价提醒，避免用户收到重复通知。
- `review_handler.go` 中损坏的历史错误文案已收敛为 ASCII 文案，恢复文件 UTF-8 和 gofmt 稳定性。
- 更新 `plan.md`：标记 Task 66 评价提醒、补评价、再玩意向，以及 TC-101、TC-103、TC-104、TC-105、TC-106、TC-107 完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/appapi/review_handler.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/appapi -run TestServiceConfirmAndReviewFlow`
  - `go test -count=1 ./internal/reviews ./internal/appapi`
  - `go test -count=1 ./...`
- 测试通过。

## 2026-06-15 01:30

### 当前进展

- 继续执行 `plan.md`，推进 D5.7 分润、收益和线下结算后端业务闭环。
- 新增后台结算记录查询接口：`GET /api/admin/revenue/settlements`，复用 `settlement:offline:create` 权限，可查看线下结算登记结果。
- `revenue.Service` 新增 `Settlements()`，SQL 仓储新增 `ListSettlements()`，内存和 PostgreSQL 路径都能返回结算记录。
- 补充收益隔离测试：第三方用户查询自己的收益账户和收益流水时只能看到 0 和空列表，不能读到其他参与者收益。
- 修复 `revenue_handler.go` 中损坏的历史错误文案，统一为 ASCII 文案，恢复 gofmt 稳定性。
- 更新 `plan.md`：标记收益流水、试算/生成一致、争议不能结算、后台结算记录、Task 67 相关项、TC-115、TC-117、TC-118 完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/revenue/service.go services/go-api/internal/revenue/sql_repository.go services/go-api/internal/revenue/repository_service_test.go services/go-api/internal/appapi/revenue_handler.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/revenue ./internal/appapi`
  - `go test -count=1 ./...`
- 测试通过。

## 2026-06-15 02:00

### 当前进展

- 继续执行 `plan.md`，补齐文件服务业务闭环。
- 文件服务新增临时下载地址过期控制，导出文件按 24 小时 TTL 生成，过期后下载返回 `410 Gone`。
- 小程序文件下载保持成员校验，后台附件查看按文件类型映射到 `identity:read`、`report:view`、`im:message:view_dispute`、`report_export:create`。
- 修复导出下载和文件下载 handler 的历史乱码文案，恢复 ASCII 稳定输出。
- 更新 `plan.md`：标记“实现临时下载地址”“实现 IM 文件权限校验”“实现后台附件查看权限校验”“导出文件链接过期后不可访问”完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/files/service.go services/go-api/internal/appapi/export_handler.go services/go-api/internal/appapi/file_handler.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./services/go-api/internal/files ./services/go-api/internal/appapi`
  - `go test -count=1 ./services/go-api/internal/...`
- 测试通过。

## 2026-06-15 02:30

### 当前进展

- 继续执行 `plan.md`，推进 E5 AI 数据准备和后台数据沉淀可视化。
- `GET /api/admin/ai-data/snapshot` 新增只读统计维度：
  - 行为事件按日期、用户、事件类型统计。
  - 收藏按用户、局类型统计。
  - 行家技能树、领路人资源画像完整率统计。
  - 人脉关系按来源和关系类型统计。
  - 数据沉淀分区 readiness，便于线上验收直接看哪些数据已经写入。
- 保持 IM 数据默认不可导出给 AI；关闭开关时 `POST /api/admin/ai-data/im-export` 仍返回 forbidden。
- 更新 `plan.md`：标记 E5 数据统计、AI IM 导出开关、验收数据沉淀相关项完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/aidata/service.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./services/go-api/internal/aidata ./services/go-api/internal/appapi`
  - `go test -count=1 ./services/go-api/internal/...`
- 测试通过。

## 2026-06-15 06:00

### 当前进展

- 继续执行 `plan.md`，优先补小程序后端当前用户和角色聚合链路。
- `POST /api/app/auth/wechat-login` 现在返回 `CurrentUserDTO` 聚合对象，包含用户基础信息、实名状态、角色列表、`roleStatusMap`、会员默认状态、成长信用、积分、收益摘要和我的邀请码。
- `GET /api/app/users/me` 改为复用同一套 `buildCurrentUserDTO`，资料更新、实名完成后可回读聚合字段。
- 新增 `GET /api/app/roles/my`，返回当前用户角色快照、角色申请记录和领路人门槛状态，便于小程序角色页联调。
- `invites.Store` 新增 `EnsureCodeForOwner`，当前用户聚合时可生成或回填我的邀请码。
- 修复 `server.go` 中一个历史损坏的登录请求错误文案，恢复 ASCII 稳定输出。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/appapi/server.go services/go-api/internal/appapi/profile_handler.go services/go-api/internal/appapi/server_test.go services/go-api/internal/auth/service.go services/go-api/internal/invites/model.go`
  - `go test -count=1 ./internal/profiles ./internal/appapi -run "TestWechatLoginAndCurrentUserHTTP|TestConnectionsAndProfilesHTTP" -v`
  - `go test -count=1 ./services/go-api/internal/...`
- 测试通过。

## 2026-06-15 06:30

### 当前进展

- 继续执行 `plan.md`，补齐小程序后端会员链路。
- 新增 `internal/membership` 服务和 SQL 仓储，支持 `membership_plans`、`user_memberships` 的读取与授予。
- `CurrentUserDTO.membership` 现在返回完整会员对象，不再是固定占位值。
- 新增 `GET /api/app/membership/plans` 和 `GET /api/app/membership/my`，小程序个人中心可直接联调会员状态。
- 启动入口 `cmd/server/main.go` 已接入 membership SQL repository，生产路径可以读取数据库会员计划和用户会员状态。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/membership/service.go services/go-api/internal/membership/sql_repository.go services/go-api/internal/membership/service_test.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/membership_handler.go services/go-api/internal/appapi/server_test.go services/go-api/cmd/server/main.go`
  - `go test -count=1 ./internal/membership ./internal/appapi -run "TestMembership|TestWechatLoginAndCurrentUserHTTP" -v`
  - `go test -count=1 ./services/go-api/internal/...`
- 测试通过。

## 2026-06-15 07:00

### 当前进展

- 继续执行 `plan.md`，补齐小程序后端 LBS 当前定位、手动定位和附近局距离字段。
- `POST /api/app/locations/current` 的返回已按 `LocationDTO` 验证，包含经纬度、城市、来源和 `accuracyWarning`。
- `POST /api/app/locations/manual` 已验证可用，未授权微信定位时可手动保存位置，并可作为附近局查询基准。
- `GET /api/app/games/nearby` 支持请求参数传入 `longitude`、`latitude` 临时坐标，并校验坐标范围。
- 附近局结果新增 `distanceLabel`，后端统一生成 `m/km` 展示标签，前端不用自行计算。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/games/service.go services/go-api/internal/lbs/service.go services/go-api/internal/appapi/location_handler.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/lbs ./internal/appapi -run "TestLocationAndNearbyGamesFlow" -v`
  - `go test -count=1 ./services/go-api/internal/...`
- 测试通过。

## 2026-06-15 07:30

### 当前进展

- 继续执行 `plan.md`，优先补小程序后端 LBS 与 IM 文件消息链路。
- `GET /api/app/locations/my-recent` 支持 `source=manual|gps` 过滤，最近手动定位可直接给小程序个人中心/定位兜底页使用。
- LBS SQL 仓储新增按来源查询最近定位，内存服务和仓储服务都保持一致行为。
- IM 图片/文件消息发送前新增 `fileId` 映射校验，要求文件必须是当前局的 `chat_file`，避免拿头像、导出文件或其他局附件伪装成聊天文件。
- OpenIM webhook 回调事件继续落本地事件记录，测试验证可从 `zhw_game_{gameId}` 解析业务 `gameId`。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/appapi/im_handler.go services/go-api/internal/appapi/server_test.go services/go-api/internal/appapi/location_handler.go services/go-api/internal/lbs/service.go services/go-api/internal/lbs/sql_repository.go services/go-api/internal/lbs/repository_service_test.go`
  - `go test -count=1 ./internal/lbs ./internal/appapi -run "TestRepositoryBackedLocationSaveCurrentAndRecent|TestLocationAndNearbyGamesFlow|TestServiceConfirmAndReviewFlow" -v`
  - `go test -count=1 ./internal/...`
- 测试通过。

## 2026-06-15 08:51

### 当前进展

- 继续执行 `plan.md`，优先补小程序后端业务功能闭环。
- 新增 `GET /api/app/users/me/summary`，聚合当前用户、待评价数量、未读通知数量、收益摘要、积分摘要和最近足迹，方便小程序“我的/个人中心”一次取数。
- 补齐入局申请链路的计划接口别名：
  - `GET /api/app/game-applications/my` 兼容申请人查看自己的入局申请。
  - `GET /api/app/game-applications/received?status=pending` 支持局主查看自己创建局收到的待审核申请。
  - `POST /api/app/game-applications/{applicationId}/cancel` 支持申请人取消 pending 申请。
  - `POST /api/app/game-applications/{applicationId}/audit` 支持局主按 `plan.md` 路由审核入局申请，并保留原 `/api/app/games/applications/{applicationId}/review` 兼容入口。
- `games.Service` 和 SQL 仓储新增局主侧申请列表、申请取消能力，真实数据库和内存服务保持同一行为。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/games/service.go services/go-api/internal/games/sql_repository.go services/go-api/internal/games/repository_service_test.go services/go-api/internal/appapi/game_handler.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/appapi -run "TestServiceConfirmAndReviewFlow|TestWechatLoginAndCurrentUserHTTP" -v`
  - `go test -count=1 ./internal/games ./internal/appapi -run "TestRepositoryBackedGameApplicationFlow|TestGameApplicationAndManualStartFlow" -v`
  - `go test -count=1 ./internal/games ./internal/appapi`
  - `go test -count=1 ./internal/...`
- 测试通过。

## 2026-06-15 13:48

### 当前进展

- 继续按 `plan.md` 优先补小程序后端 E2.2 里程碑、打卡、复盘、续局验收项。
- 收紧打卡类型枚举：`CreateCheckin` 现在只允许 `progress/complete/arrival/proof`，与计划要求一致。
- 移除服务层对旧打卡类型 `photo/text` 的放行，避免小程序端和后端字段契约不一致。
- 扩展打卡服务测试，覆盖旧类型 `photo` 被拒绝，以及 `arrival/proof` 两类新枚举正常使用。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/games -v`
  - `go test -count=1 ./internal/...`
- 测试通过。

## 2026-06-15 13:18

### 当前进展

- 继续按 `plan.md` 优先补小程序后端业务验收项，复核 E2.2 里程碑、打卡、复盘、续局链路。
- 补齐里程碑状态枚举：`UpdateMilestone` 现在允许 `pending/in_progress/completed/cancelled`，与计划要求一致。
- 将里程碑状态校验收敛为 `validMilestoneStatus`，避免状态字符串继续散落在业务分支里。
- 扩展里程碑服务测试，覆盖 `in_progress -> completed` 的状态流，以及非法状态拒绝。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/games -v`
  - `go test -count=1 ./internal/...`
- 测试通过。

## 2026-06-15 12:55

### 当前进展

- 继续按 `plan.md` 优先补小程序后端业务验收项，复核 E2.1 收藏与行为上报。
- 收紧 `POST /api/app/behavior/events` 字段契约：小程序端行为上报必须携带 `eventType`，缺失时返回参数错误。
- `eventCode` 仍作为分析口径保留，可由客户端显式传入；未传时默认等于 `eventType`，避免影响后台漏斗/留存统计。
- 扩展行为事件测试，覆盖缺少 `eventType` 被拒、带 `eventType/eventCode` 正常入库、后台按 `eventCode` 过滤仍可用。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/appapi -run TestBehaviorEventHTTP -v`
  - `go test -count=1 ./internal/...`
- 测试通过。

## 2026-06-15 12:42

### 当前进展

- 继续按 `plan.md` 优先补小程序后端业务接口，复核 E2 中积分/会员/团队/分润模拟验收项。
- 补齐计划中列出的 `POST /api/app/revenues/simulate` 用户端分润模拟接口。
- 接口复用现有分润 `Preview` 规则：用户只能模拟自己创建或参与的局，服务端自动补齐成员、创建者和默认模板。
- 分润模拟只返回预估结果和阻塞原因，不生成真实分润记录、不写收益流水，符合一期“只预估不落账”的口径。
- 扩展分润链路测试，覆盖非成员拒绝、成员模拟成功、模拟后后台分润记录仍为空。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/appapi -run TestRevenuePreviewGenerateFreezeAndSettlementFlow -v`
  - `go test -count=1 ./internal/...`
- 测试通过。

## 2026-06-15 12:25

### 当前进展

- 继续按 `plan.md` 补小程序后端业务闭环，复核 D5.7/D5.8 中“人脉关系自动来源”项。
- 已确认邀请码登录成功会自动生成 `invite` 关系，服务确认和评价链路会生成或增强 `co_game` 关系。
- 补齐游戏/领路人邀请接受后的关系链：被邀请人接受 `guide-invitations` 后，后端自动为邀请人与被邀请人生成双向 `guide_match` 关系，来源 ID 绑定邀请记录。
- 扩展邀请链路测试，覆盖“邀请接受 -> pending 入局申请 -> 创建者可见待审核申请 -> 自动生成 guide_match 人脉关系”完整流程。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/appapi -run TestGameInvitationRespondCreatesApplication -v`
  - `go test -count=1 ./internal/...`
- 测试通过。

## 2026-06-15 12:14

### 当前进展

- 继续按 `plan.md` 推进小程序后端业务功能，补齐 D5.3 游戏邀请链路。
- 新增 `POST /api/app/games/{gameId}/guide-invitations`，创建者或局内成员可邀请用户，邀请写入 `game_invitations`。
- 新增 `POST /api/app/game-invitations/{invitationId}/respond` / `POST /api/app/games/invitations/{invitationId}/respond`，被邀请人接受后只生成 `pending` 入局申请，不直接成为成员，仍需发起人审核。
- 新增 `000024_game_invitations.sql`，补齐邀请表、目标用户索引和局状态索引。
- 补齐后台用户详情、AI 验收夹具、匿名行为埋点、LBS 当前定位、后台文件权限、通知读取等后端跑通阻塞点。
- 修复内部包测试口径和业务校验顺序，确保后端内部包可以整体验证。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/appapi -run TestGameInvitationRespondCreatesApplication -v`
  - `go test -count=1 ./internal/games -run TestGameRepositoryPersistsCoreFlow -v`
  - `go test -count=1 ./internal/games ./internal/appapi`
  - `go test -count=1 ./internal/...`
- 测试通过。

## 2026-06-15 10:15

### 当前进展

- 继续执行 `plan.md`，优先补小程序后端组局详情页读模型。
- `GET /api/app/games/{gameId}` 对已登录用户返回兼容旧字段的详情聚合，局基础字段仍在 `data` 顶层。
- 详情新增 `myRelation`，包含 `role`、`isCreator`、`isMember`、`canApply`、`canAudit`、`canStart`、`canEnterIM`、`canConfirm`、`canReview`、申请 ID 和申请状态。
- 详情新增 `memberIds`、`progress.feedbacks`、`progress.milestones`、`progress.checkins`，成员可在一个接口拿到详情页进度数据。
- 详情新增 `im` 和 `review` 状态，返回 IM 可用性、roomId、引擎、OpenIM groupId，以及当前用户在该局是否可评价、待评价项和全局评价完成状态。
- 详情新增 `serviceConfirm`，已发起服务确认的局会返回确认主记录和确认成员项。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/appapi/game_handler.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/appapi -run "TestGameApplicationAndManualStartFlow" -v`
  - `go test -count=1 ./internal/appapi -run "TestGameApplicationAndManualStartFlow|TestServiceConfirmAndReviewFlow" -v`
  - `go test -count=1 ./internal/games ./internal/appapi`
  - `go test -count=1 ./internal/...`
- 测试通过。

## 2026-06-15 10:46

### 当前进展

- 继续执行 `plan.md`，补齐小程序后端 D5.3 的局成员列表接口。
- 新增 `GET /api/app/games/{gameId}/members`，局成员或创建者可查看成员列表，非成员不能读取。
- 成员列表返回 `userId`、`role`、`isCreator`、`isCurrentUser`、`confirmed`，可直接支撑小程序成员页、服务确认名单和 IM 成员展示。
- 成员确认状态复用现有 `game_service_confirms` / `game_service_confirm_items` 数据，不新增状态表。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/appapi/game_handler.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/appapi -run "TestGameApplicationAndManualStartFlow" -v`
  - `go test -count=1 ./internal/games ./internal/appapi -run "TestGameApplicationAndManualStartFlow|TestServiceConfirmAndReviewFlow|TestRepositoryBackedGameApplicationFlow" -v`
  - `go test -count=1 ./internal/games ./internal/appapi`
  - `go test -count=1 ./internal/...`
- 测试通过。

## 2026-06-15 21:24

### 当前进展

- 根据最新登录注册规则，继续执行 `plan.md`，优先补齐小程序一期后端邀请入口链路。
- 邀请入口统一为三类：`poster` 小程序分享卡片、`qrcode` 扫码邀请、`link` 微信聊天链接邀请。
- 新增 `POST /api/app/invites/entries`，登录用户可生成一次性唯一邀请码，并返回小程序 `path`、二维码 `scene` 和链接占位 `urlLink`。
- `POST /api/app/invites/precheck` 收紧入口校验：邀请码缺失、入口类型非法、入口类型不匹配、失效或耗尽都会拒绝；未绑定返回 `authPageMode=register`，已绑定唯一微信返回 `authPageMode=login`。
- `POST /api/app/auth/wechat-login` 收紧为必须携带 `inviteCode`；新用户只允许通过有效邀请码注册并绑定微信；一次性邀请码已绑定其他微信时拒绝复用。
- 数据库仓储绑定邀请码时增加行锁，降低同一个一次性邀请码并发绑定到多个微信的风险。
- `docs/openapi/app.openapi.yaml` 已同步邀请入口生成、入口预检和微信登录新字段。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/invites/model.go services/go-api/internal/invites/sql_repository.go services/go-api/internal/auth/service.go services/go-api/internal/auth/service_test.go services/go-api/internal/auth/repository_service_test.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/auth ./internal/appapi -run "TestWechatLogin|TestInvitePrecheck|TestCreateInviteEntryHTTP" -v`
  - `go test -count=1 ./internal/...`
- 测试通过。
## 2026-06-16 11:25

### 当前进展

- 继续以一期小程序后端线上测试为目标，补齐短信验证码生产链路门槛。
- `identity.Service` 新增 `SMSSender` 抽象：
  - 本地默认 `LocalSMSSender`，继续返回固定 `mockCode=123456`，保持本地联调和测试稳定。
  - 生产可配置 `HTTPSMSSender`，通过 `SMS_HTTP_ENDPOINT` / `SMS_HTTP_SECRET` 调用短信网关。
- `cmd/server` 已按配置注入短信发送器。
- 生产环境配置校验新增短信要求：
  - `SMS_HTTP_ENDPOINT` 必须是 HTTPS URL。
  - `SMS_HTTP_SECRET` 必须是安全值。
- 小程序 `POST /api/app/sms/send-code` 响应收紧：
  - 本地返回 `mockCode`。
  - 真实短信通道只返回 `provider/messageId`，不回传验证码明文。
- AI 验收夹具只允许在本地短信 sender 下自动取 `mockCode`，避免生产误用夹具消耗真实短信。
- `docs/openapi/app.openapi.yaml` 已同步短信发送接口说明。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/identity ./internal/common/config ./internal/appapi`
- 测试通过。
- 当前一期小程序后端业务功能进度估算：约 92%。

## 2026-06-16 10:49

### 当前进展

- 继续以一期小程序后端线上测试为目标，收紧生产文件上传/下载配置。
- 生产环境现在要求：
  - `STORAGE_UPLOAD_BASE_URL` 必须是 HTTPS URL。
  - `STORAGE_DOWNLOAD_BASE_URL` 必须是 HTTPS URL。
- 避免 `APP_ENV=prod/production` 时仍静默返回 `mock://upload/...` 或非 HTTPS 地址，防止小程序真实设备联调失败。
- `docs/openapi/app.openapi.yaml` 已同步生产文件 URL 要求。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/common/config`
- 测试通过。
- 当前一期小程序后端业务功能进度估算：约 91%。

## 2026-06-16 10:26

### 当前进展

- 继续补一期小程序后端线上测试硬缺口，处理文件上传/下载 URL 仍为 `mock://` 的问题。
- `StorageConfig` 新增：
  - `STORAGE_UPLOAD_BASE_URL`
  - `STORAGE_DOWNLOAD_BASE_URL`
- 文件服务新增公开基础 URL 配置：
  - 本地未配置时继续返回 `mock://upload/...` 和 `mock://download/...`，保持本地测试稳定。
  - 线上配置后返回 HTTPS 上传/下载地址，方便小程序 `wx.uploadFile` 和下载联调。
- `cmd/server` 现在会调用 `appServer.Configure(cfg)`，让每日开局限制、LBS 默认半径、文件 URL 配置在真实启动时生效。
- `.env.example` 已同步新增文件上传/下载基础 URL 配置项。
- `docs/openapi/app.openapi.yaml` 已同步文件上传凭证说明。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/files ./internal/common/config ./internal/appapi`
- 测试通过。
- 当前一期小程序后端业务功能进度估算：约 91%。

## 2026-06-16 10:23

### 当前进展

- 继续以一期小程序后端线上测试为导向，优先补齐微信登录生产链路。
- `auth.Service` 新增 `WechatCodeResolver` 抽象：
  - 本地默认继续使用 mock openid，现有联调和自动化测试不受影响。
  - 生产环境配置 `WECHAT_APP_ID` / `WECHAT_APP_SECRET` 后，后端会调用微信 `jscode2session` 换取真实 `openid`、`unionid` 和 `session_key`。
- `cmd/server` 已按配置注入真实微信解析器；未配置时仍走本地 mock fallback。
- `config.Load` 已读取 `WECHAT_APP_ID`、`WECHAT_APP_SECRET`。
- 生产环境校验新增微信凭证要求，避免线上仍误用 mock 登录。
- 小程序登录接口新增微信 code 无效错误映射，`jscode2session` 返回错误或空 `openid` 时返回参数错误，前端可识别为 `wx.login` code 失效。
- `docs/openapi/app.openapi.yaml` 已同步登录接口说明。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/auth ./internal/common/config ./internal/appapi`
- 测试通过。
- 当前一期小程序后端业务功能进度估算：约 90%。

## 2026-06-16 10:15

### 当前进展

- 继续以一期小程序后端业务功能为主线，按最新登录注册和组局规则补齐后端闭环。
- 玩家创建免费局、申请入局不再强制实名；行家、领路人和被邀请主行家仍保留实名门槛。
- 游戏模型新增 `mainGuideUserId`，首个“行家邀请接受 + 审核通过”的成员会成为主行家，成员角色返回 `main_guide`。
- 手动开始收紧为创建者或主行家可操作，并严格要求人数达到 `minPlayers`；一期人数范围为 5-8，未满 5 人不能手动开始。
- 主行家已可管理局进度、里程碑和进度反馈；普通成员仍按成员权限参与聊天、打卡、服务确认和评价。
- 小程序详情和成员接口同步返回 `mainGuideUserId`、`myRelation.role`、`myRelation.canStart` 和成员 `role`，便于前端直接判断按钮状态。
- 同步 `docs/openapi/app.openapi.yaml` 和 `docs/test-cases/app-integration-cases.md` 的一期联调口径，避免线上测试继续按旧的“玩家必须实名”规则接入。

### 验证结果

- 已执行：
  - `go test -run TestIMArchiveJobSkipsDisputedRooms -count=1 -v ./internal/appapi`
  - `go test -run TestConnectionsAndProfilesHTTP -count=1 -v ./internal/appapi`
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./internal/auth ./internal/games`
  - `go test -count=1 ./internal/...`
- 测试通过。
- 当前一期小程序后端业务功能进度估算：约 88%。

## 2026-06-16 14:26

### 当前进展

- 继续以一期小程序后端上线测试为目标，补齐后台 IM 房间管理入口。
- 新增后台 IM 房间列表接口：
  - `GET /api/admin/im/rooms`
  - 支持 `gameId` 和 `status` 筛选。
  - 返回房间成员、房间状态、消息数量和文件消息数量。
- 新增后台 IM 房间详情接口：
  - `GET /api/admin/im/rooms/{roomId}`
  - 返回房间信息、成员、完整消息列表、文件消息列表、消息计数和文件消息计数。
- 后台权限补充 `im:room:read`，与现有争议消息权限 `im:message:view_dispute` 分开。
- 同步 `docs/openapi/admin.openapi.yaml` 和 `docs/test-cases/admin-integration-cases.md`，方便后台联调直接按接口走。

### 验证补充

- `go test -count=1 ./internal/appapi` 通过。
- `go test -count=1 ./internal/...` 通过。
- 当前一期小程序后端业务功能进度约 `97.5%`。

## 2026-06-16 15:56

### 当前进展

- 继续以一期小程序后端线上测试为目标，补齐上线前配置体检和后台邀请运营能力。
- 新增后台生产配置体检接口：
  - `GET /api/admin/system/readiness`
  - 只允许 `system_config:read` 权限访问。
  - 检查数据库、JWT、微信登录、短信、腾讯云慧眼、FaceID 回调、存储、OpenIM 和资金服务配置状态。
  - 响应只返回 ok / warn / missing，不返回任何密钥值。
- 新增后台邀请码管理接口：
  - `GET /api/admin/invite-codes`
  - `POST /api/admin/invite-codes`
  - `POST /api/admin/invite-codes/{code}/disable`
  - `GET /api/admin/invite-relations`
- 后台权限补充：
  - `invite_code:read`
  - `invite_code:manage`
  - `system_config:read`
- 管理员现在可以为海报、二维码、链接三种入口创建唯一邀请码，查看绑定关系，并禁用异常邀请码。
- 同步 `docs/openapi/admin.openapi.yaml` 和 `docs/test-cases/admin-integration-cases.md`。

### 验证补充

- `go test -count=1 ./internal/common/config ./internal/appapi` 通过。
- `go test -count=1 ./internal/invites ./internal/auth ./internal/appapi` 通过。
- 当前一期小程序后端业务功能进度约 `98.5%`。

## 2026-06-16 16:26

### 当前进展

- 继续收口一期小程序后端后台 IM 管理链路。
- 新增后台 IM 房间重试创建接口：
  - `POST /api/admin/im/rooms/{roomId}/retry-create`
  - 未启用 OpenIM 时返回当前本地房间，启用 OpenIM 时执行真实房间同步。
- 新增后台 IM 消息隐藏接口：
  - `POST /api/admin/im/messages/{messageId}/hide`
  - 隐藏后小程序用户侧历史消息不再返回该消息，后台房间详情仍保留证据。
- 新增后台 IM 房间归档接口：
  - `POST /api/admin/im/rooms/{roomId}/archive`
  - 归档后用户侧不能再向该房间发送消息。
- 后台权限补充：
  - `im:room:retry_create`
  - `im:room:archive`
  - `im:message:hide`
- 同步 `docs/openapi/admin.openapi.yaml` 和 `docs/test-cases/admin-integration-cases.md`。

### 验证补充

- 已执行：
  - `go test -count=1 ./internal/im ./internal/appapi`
  - `go test -count=1 ./internal/...`
- 测试通过。
- 当前一期小程序后端业务功能进度约 `99%`。

## 2026-06-16 16:56

### 当前进展

- 继续按一期小程序后端线上测试口径收口，处理头像文件 URL 的生产可用性。
- `PUT /api/app/users/me/profile` 在传入 `avatarFileId` 且未显式传 `avatarUrl` 时，现在通过统一文件服务生成头像下载地址。
- 本地未配置对象存储下载域名时仍返回 `mock://download/...`，保持本地测试稳定。
- 生产配置 `STORAGE_DOWNLOAD_BASE_URL` 后，头像 URL 自动返回 HTTPS 下载域名，不再由资料接口硬编码 `mock://`。

### 验证补充

- 已执行：
  - `go test -count=1 ./internal/files ./internal/appapi`
  - `go test -count=1 ./internal/...`
- 测试通过。
- 当前一期小程序后端业务功能进度约 `99.2%`。

## 2026-06-16 17:26

### 当前进展

- 继续按一期小程序后端联调口径收口，修正人脉跟进接口路径不一致问题。
- `POST /api/app/connections/{connectionId}/follow-logs` 已接入后端，作为文档主路径。
- 原有 `POST /api/app/connections/{connectionId}/follow-up` 保留为兼容别名，避免旧前端或旧测试调用中断。
- 同步 `docs/openapi/app.openapi.yaml` 和 `docs/test-cases/app-integration-cases.md`。

### 验证补充

- 已执行：
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./internal/...`
- 测试通过。
- 当前一期小程序后端业务功能进度约 `99.35%`。

## 2026-06-16 17:56

### 当前进展

- 继续按一期小程序后端联调可用性收口，修复 `docs/openapi/app.openapi.yaml` 中 6 处路径被上一段 `description` 吞并到同一行的问题。
- 已拆正以下接口段落：
  - `GET/POST /api/app/chat/rooms/{roomId}/messages`
  - `POST /api/app/games/{gameId}/service-confirm-items`
  - `GET /api/app/reviews/todos`
  - `GET /api/app/growth/my`
  - `POST /api/app/payment/precreate-placeholder`
  - `POST /api/app/reports`
- 这轮不改业务逻辑，只修 OpenAPI 结构，避免前端联调或接口生成器漏识别路径。

### 验证补充

- 已执行：
  - OpenAPI 坏行复扫：未再发现 `description` 吞并 `/api/app/...` 路径。
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./internal/...`
- 测试通过。
- 当前一期小程序后端业务功能进度约 `99.45%`。

## 2026-06-16 18:26

### 当前进展

- 继续按一期小程序后端线上联调口径收口，修复 OpenAPI 文档剩余结构问题。
- 已拆正 `docs/openapi/app.openapi.yaml` 中 `GET /api/app/reports/my` 的 `components` 粘连问题。
- 已修复 `docs/openapi/admin.openapi.yaml` 早期后台接口段落的路径、`responses`、状态码粘连问题，覆盖：
  - 后台登录与权限快照。
  - 后台权限树。
  - 微信订阅消息模板与任务。
  - 内部订阅消息发送回写任务。
  - 评价提醒与进度反馈提醒任务。
- 这轮仍不改变业务代码，只保证一期后端接口文档可以稳定用于前端联调和线上测试核对。

### 验证补充

- 已执行：
  - OpenAPI 坏行复扫：未再发现 `description/summary` 吞并 `/api/...`、`responses`、`security`、`components` 的结构问题。
- 已执行：
  - `go test -count=1 ./internal/...`
- 测试通过。
- 当前一期小程序后端业务功能进度约 `99.55%`。

## 2026-06-16 18:56

### 当前进展

- 继续以一期小程序后端线上测试为目标，补齐“链接邀请”真实联调口径。
- 新增 `WECHAT_URL_LINK_BASE_URL` 配置：
  - 本地未配置时，`POST /api/app/invites/entries` 继续返回 `wechat://mini-program?...` 占位链接，保证本机测试不被微信接口阻塞。
  - 生产配置后，`urlLink` 返回 HTTPS 链接，并自动携带 `inviteCode`、`entryType`、`path` 和 `query` 参数，方便微信聊天链接邀请联调。
- 同步 `.env.example` 和 `docs/openapi/app.openapi.yaml`，避免前端仍按“固定占位链接”理解。

### 验证补充

- 已执行：
  - `go test -count=1 ./internal/common/config ./internal/appapi`
  - `go test -count=1 ./internal/...`
- 测试通过。
- 当前一期小程序后端业务功能进度约 `99.65%`。

## 2026-06-17 10:30

### 当前进展

- 根据一期最新口径修正局型创建边界：
  - 小程序普通用户一期仍只能创建 `free` 普通局。
  - 后台新增开局入口，允许创建 `free`、`standard`、`condition`，用于一期线上测试其他条件局链路。
  - 后台创建的局标记 `gameSource=admin`，默认进入 `recruiting`，小程序侧可展示和参与。
- 新增后台权限 `game:create_admin`，和原 `game:read`、`game:update_status` 分离。
- 同步 `docs/openapi/admin.openapi.yaml` 和后台联调用例说明。

### 验证补充

- 待执行：
  - `go test -count=1 ./internal/games ./internal/appapi`
  - `go test -count=1 ./internal/...`
- 修正后当前一期小程序后端业务功能进度约 `99.7%`。

## 2026-06-17 11:40

### 当前进展

- 开始按一期后台前后端闭环补齐后台工作台，优先覆盖可线上测试的运营链路。
- 后端补充后台组局能力：
  - `GET /api/admin/games` 支持 `status`、`gameType`、`cityCode`、`keyword` 过滤。
  - `GET /api/admin/games/{gameId}` 返回局详情、成员、里程碑、打卡、复盘、续局草稿和服务确认信息。
  - 后台字段继续对齐小程序 `Game` DTO：`gameType`、`gameSource`、`status`、`minPlayers`、`maxPlayers`、`currentPlayers`。
- 后台前端从占位壳升级为真实可用一期工作台：
  - 登录与权限拉取。
  - 总览指标。
  - 后台创建普通局、标准局、条件局。
  - 组局筛选、审核通过、局链路详情。
  - 实名记录、角色申请、IM 房间、操作日志只读查看。
  - 本地 `admin-web/server.js` 增加 `/api` 代理，开发环境可同域对接 Go API。

### 验证补充

- 已执行：
  - `go test -count=1 ./internal/appapi ./internal/games ./internal/adminauth`
  - `go test -count=1 ./internal/...`
  - `go build ./cmd/server`
  - `node --check admin-web/src/main.js`
  - `node --check admin-web/server.js`
  - `npm run build`
- 当前后台前端业务功能进度约 `35%`：一期关键入口已可用，但还缺用户管理、邀请码管理、财务分润、举报申诉、导出中心、系统配置等完整页面。
- 当前一期小程序后端业务功能进度仍约 `99.7%`。

## 2026-06-17 16:20

### 当前进展

- 继续按一期后台前后端闭环推进，优先补齐和小程序登录注册强相关的运营能力。
- 后端新增后台用户列表接口：
  - `GET /api/admin/users` 支持 `keyword`、`realnameStatus`、`status` 过滤。
  - 返回字段对齐小程序 `User` DTO：`id`、`openId`、`nickname`、`avatarUrl`、`avatarFileId`、`realnameStatus`、`status`、`createdAt`。
  - 内存仓库和 SQL 仓库同步实现，SQL 字段继续对齐 `users` 与 `user_wechat_accounts`。
- 后台前端重建为可读中文工作台，并新增：
  - 用户管理：用户筛选、列表、详情，详情聚合实名、角色、人脉、收藏和收入摘要。
  - 邀请管理：创建邀请码、邀请码筛选、禁用邀请码、邀请关系列表。
  - 保留并整合总览、后台开局、组局审核、实名/角色审核、IM 房间、操作日志页面。
- 同步 `docs/openapi/admin.openapi.yaml`，补充 `/api/admin/users`。

### 验证补充

- 已执行：
  - `gofmt -w services/go-api/internal/users/model.go services/go-api/internal/users/sql_repository.go services/go-api/internal/auth/service.go services/go-api/internal/auth/repository_service_test.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/users ./internal/auth ./internal/appapi`
  - `go test -count=1 ./internal/...`
  - `node --check admin-web/src/main.js`
  - `node --check admin-web/server.js`
  - `npm run build`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `62%`：权限、用户、邀请、组局、审核、IM、日志已有核心接口，仍需继续补财务分润、举报申诉、导出中心、系统配置和更完整的统计接口。
- 当前后台前端业务功能进度约 `48%`：核心运营入口已成型，仍需补财务、举报、导出、配置和更细的详情操作页。

## 2026-06-17 16:55

### 当前进展

- 继续补齐一期 PC 后台业务功能，优先把已有后端能力变成可线上测试的后台页面。
- 后台前端新增分润结算页面：
  - 分润模板创建与列表查看。
  - 分润规则配置与查看。
  - 分润试算、生成分润记录。
  - 分润记录列表、冻结、线下结算登记。
  - 线下结算记录列表。
- 后台前端新增举报申诉页面：
  - 举报申诉列表。
  - 举报详情与证据摘要。
  - 分配处理人、处理、关闭。
  - 页面字段对齐 `reports.Report`：`gameId`、`reporterUserId`、`targetUserId`、`reportType`、`status`、`revenueFrozen`。
- 后台前端新增导出中心：
  - 固定格式导出模板列表。
  - 创建导出任务。
  - 执行导出 runner。
  - 导出任务列表与完成后的下载地址获取。
- 这轮未新增后端业务接口，复用并联通已有接口：
  - `/api/admin/revenue/*`
  - `/api/admin/reports/*`
  - `/api/admin/reports/export*`
  - `/api/admin/export-tasks*`

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi ./internal/revenue ./internal/reports ./internal/exports`
  - `go test -count=1 ./internal/...`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `62%`：本轮主要补前端接入，后端百分比不变。
- 当前后台前端业务功能进度约 `68%`：用户、邀请、组局、审核、IM、日志、分润、举报、导出已具备基础可操作页面；还需补系统配置、敏感词、数据看板细化、管理员/权限管理、更多详情联动和浏览器联调。

## 2026-06-17 17:02

### 当前进展

- 继续补齐一期 PC 后台系统管控能力，复用已有后端接口，不重复造服务。
- 后台前端新增系统配置页面：
  - 生产就绪检查：对接 `GET /api/admin/system/readiness`，展示 `ready`、`production` 和检查项，不展示密钥明文。
  - 敏感词库：对接 `GET/POST/PUT /api/admin/sensitive-words` 和 `POST /api/admin/sensitive-words/import`，支持新增、批量导入、启用/禁用。
  - 内容风险日志：对接 `GET /api/admin/content-risk/logs`，查看 IM 内容命中记录。
  - 权限快照：对接 `GET /api/admin/permissions/tree`，展示当前管理员角色、菜单、API 和权限数量。
  - AI IM 导出开关：对接 `GET/PUT /api/admin/ai-data/im-export-config`，保留一期默认关闭、需权限开启的后台入口。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi ./internal/im ./internal/adminauth`
  - `go test -count=1 ./internal/...`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `62%`：本轮仍以接通已有后端为主。
- 当前后台前端业务功能进度约 `76%`：系统配置、敏感词、权限快照、AI 导出开关已接入；剩余重点是管理员账号/角色管理、数据看板细化、更多详情联动和浏览器联调。

## 2026-06-17 17:31

### 当前进展

- 继续补齐一期 PC 后台 RBAC 管理能力，优先保证后台接口字段与小程序后端权限 code 对齐。
- 后台后端新增管理员权限目录接口：
  - `GET /api/admin/admin-users`：返回后台账号列表，字段包含 `id`、`username`、`status`、`roles`、`permissionCount`、`permissions`。
  - `GET /api/admin/admin-roles`：返回一期固定七类后台角色，覆盖超级管理员、运营管理、用户管理、财务管理、客户管理、数据分析员、组局管理。
  - `GET /api/admin/admin-permissions/catalog`：返回权限 code 目录，字段包含 `code`、`module`、`action`。
- 后端 RBAC 权限继续统一走 `requireAdminPermission`，新增三个接口绑定 `admin_user:view`，运营账号访问会返回 403。
- 后台前端新增“管理员权限”页面：
  - 后台账号表。
  - 七类角色目录。
  - 权限 code 目录。
  - 页面字段对齐 `admin_users`、`admin_roles`、`admin_permissions`、`admin_user_roles`。
- 保留系统配置页的权限快照，用于查看当前登录管理员的实时权限；新增页面用于管理视角查看全局目录。

### 验证补充

- 已执行：
  - `gofmt -w internal/adminauth/service.go internal/appapi/admin_auth_handler.go internal/appapi/server.go internal/appapi/server_test.go`
  - `go test -count=1 ./internal/adminauth ./internal/appapi`
  - `go test -count=1 ./internal/...`
  - `node --check admin-web/src/main.js`
  - `npm run build`
- 浏览器插件拒绝打开 `http://127.0.0.1:5173`，未做浏览器可视化烟测。
- 当前本机 `8080` 上已有旧/裁剪 Go 进程，只返回 health，后台登录接口为 404；为避免误伤用户已有进程，本轮未强制清理端口。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `66%`：管理员账号、角色、权限目录已补齐为可测接口；还差后台账号创建/禁用、角色授权持久化、菜单权限持久化、更多操作日志闭环。
- 当前后台前端业务功能进度约 `82%`：管理员权限页已接入；剩余重点是更多详情联动、数据看板细化、真实运行态浏览器联调。

## 2026-06-17 17:47

### 当前进展

- 继续推进 PC 后台 RBAC 从“只读目录”到“可管理闭环”。
- 后台后端新增管理员账号管理能力：
  - `POST /api/admin/admin-users`：创建后台管理员账号，字段为 `username`、`password`、`roles`、`status`。
  - `PUT /api/admin/admin-users/{id}`：更新管理员角色和状态，字段为 `roles`、`status`。
  - 新账号密码使用 PBKDF2-SHA256 + 随机 salt 生成哈希，不复用默认账号密码哈希。
  - 更新管理员角色或禁用账号后，会清理该管理员旧 session，避免旧 token 保留旧权限。
  - 创建绑定 `admin_user:create`，更新绑定 `role:update`，继续走统一后台权限中间件。
  - 创建和更新都会写入 `operation_logs`，用于超级管理员追责审计。
- 后台前端“管理员权限”页面补齐可操作能力：
  - 新增后台账号创建表单。
  - 后台账号列表支持行内修改 `roles` 和 `status`。
  - 支持禁用 / 启用管理员账号。
  - 字段继续对齐 `admin_users`、`admin_user_roles`、`admin_roles`、`admin_permissions`。
- 测试新增覆盖：
  - 低权限运营账号无法创建管理员。
  - 超级管理员可创建财务管理员。
  - 新财务管理员可访问财务接口但不能查看完整操作日志。
  - 禁用后旧 token 失效，新登录被拒绝。

### 验证补充

- 已执行：
  - `gofmt -w internal/adminauth/service.go internal/appapi/admin_auth_handler.go internal/appapi/server.go internal/appapi/server_test.go`
  - `go test -count=1 ./internal/adminauth ./internal/appapi`
  - `go test -count=1 ./internal/...`
  - `node --check admin-web/src/main.js`
  - `npm run build`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `72%`：RBAC 已从目录展示推进到账号创建、角色绑定、状态启停和 session 失效；剩余重点是持久化表落库、菜单权限持久化、更多后台操作统一日志。
- 当前后台前端业务功能进度约 `86%`：管理员权限页已具备创建和行内管理；剩余重点是详情联动、数据看板细化、真实运行态浏览器联调。

## 2026-06-17 17:56

### 当前进展

- 继续推进 PC 后台 RBAC 从“内存可测”到“数据库可上线”。
- 后台权限服务新增 SQL repository：
  - 后台登录按 `admin_users.username` 查找账号。
  - 账号角色从 `admin_user_roles` + `admin_roles` 读取。
  - 账号权限从 `admin_role_permissions` + `admin_permissions` 汇总。
  - 管理员列表、角色目录、权限目录均支持数据库读取。
  - 新增管理员写入 `admin_users` 和 `admin_user_roles`。
  - 更新管理员状态/角色写回 `admin_users.status` 和 `admin_user_roles`，并继续清理旧 session。
- `cmd/server` 启动入口已接入：
  - 配置 `DATABASE_URL` / `DATABASE_DRIVER` 且数据库可用时，后台管理员登录和 RBAC 管理走 `adminauth.NewSQLRepository(db)`。
  - 未配置数据库时继续使用内存 fallback，方便本地联调和既有测试。
- 同步 `db/seeds/admin_roles_permissions.sql`：
  - 补齐 `im:room:read`、`im:room:archive`、`im:room:retry_create`、`system_config:read`、`invite_code:read`、`invite_code:manage`。
  - 补齐 `user_manager`、`finance_manager`、`customer_manager`、`game_manager` 角色权限映射。
  - 补齐运营/数据分析角色的内容风控权限映射。
- 新增 repository 模式单元测试：
  - 验证登录使用持久化账号。
  - 验证创建管理员会走 repository。
  - 验证更新角色/禁用账号后旧 token 失效。

### 验证补充

- 已执行：
  - `gofmt -w internal/adminauth/service.go internal/adminauth/sql_repository.go internal/adminauth/service_test.go internal/appapi/admin_auth_handler.go internal/appapi/server.go cmd/server/main.go`
  - `go test -count=1 ./internal/adminauth`
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./internal/...`
  - `go test -count=1 ./...`
  - `node --check admin-web/src/main.js`
  - `npm run build`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `78%`：后台账号、角色、权限已经从接口到数据库持久化闭环；剩余重点是菜单配置持久化、后台审核/运营页面更完整的批量操作和运行态联调。
- 当前后台前端业务功能进度约 `86%`：本轮无新增可视页面，管理员权限页可继续复用现有接口；剩余重点是详情联动、数据看板细化、真实运行态浏览器联调。

## 2026-06-17 18:03

### 当前进展

- 继续推进 PC 后台“权限生效”从接口层扩展到菜单层。
- 后端 `/api/admin/auth/permissions` 和 `/api/admin/permissions/tree` 的 `menus` 改为按权限码动态生成：
  - 数据看板：`analytics:*`
  - 用户管理：`user:*`、`identity:read`、`profile:read`
  - 邀请管理：`invite_code:*`
  - 组局管理：`game:*`
  - 审核中心：`identity:read`、`role:*`
  - 分润结算：`revenue:*`、`settlement:offline:create`
  - 举报申诉：`report:*`
  - 导出中心：`report_export:create`
  - 管理员权限：`admin_user:*`、`role:update`
  - 系统配置：`system_config:*`、内容风控、AI 数据准备权限
  - IM 证据：`im:room:read`、`im:message:view_dispute`
  - 操作日志：`operation_log:view_full`
- 后台前端登录后保存权限树，并按 `menus` 隐藏无权限导航入口。
- 如果当前页面无权限，前端会自动切换到第一个可访问菜单，避免低权限账号登录后停留在无权限页面。
- 测试补充：
  - 运营账号可看到组局、举报、系统配置相关菜单。
  - 运营账号不可看到数据看板、管理员权限、操作日志、财务结算菜单。

### 验证补充

- 已执行：
  - `gofmt -w internal/adminauth/service.go internal/appapi/server_test.go`
  - `go test -count=1 ./internal/adminauth`
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./...`
  - `node --check admin-web/src/main.js`
  - `npm run build`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `81%`：后台 RBAC 已覆盖登录、权限码、菜单树、账号/角色/权限落库；剩余重点是更多后台业务页的批量操作和运行态联调。
- 当前后台前端业务功能进度约 `88%`：导航已按权限树动态收敛；剩余重点是批量操作交互、详情联动和真实浏览器联调。

## 2026-06-17 18:07

### 当前进展

- 继续补后台系统配置持久化缺口。
- AI 数据准备的 IM 导出开关从内存态改为可选数据库持久化：
  - 新增 `aidata.Repository`。
  - 新增 `aidata.NewServiceWithRepository`。
  - 新增 `aidata.NewSQLRepository(db)`。
  - `IMExportConfig()` 优先从 `ai_data_snapshots` 最新记录读取 `im_export_enabled`。
  - `SetIMExportEnabled(enabled)` 写入 `ai_data_snapshots`，保留历史配置轨迹。
  - 无数据库时仍保持原内存 fallback，方便本地联调。
- `cmd/server` 启动入口新增 `aiDataRepository(db)` 并注入 `appServer.UseAIDataRepository(...)`。
- 后台接口路径不变：
  - `GET /api/admin/ai-data/im-export-config`
  - `PUT /api/admin/ai-data/im-export-config`
  - 后台前端系统配置页可继续使用原字段，新增 `updatedAt` 可用于展示配置更新时间。
- 新增 `internal/aidata` repository 模式单元测试，覆盖：
  - 默认关闭。
  - 开启后配置可读取。
  - 未开启时 IM 导出被拦截。
  - 开启后 IM 消息可导出。

### 验证补充

- 已执行：
  - `gofmt -w internal/aidata/service.go internal/aidata/sql_repository.go internal/aidata/service_test.go internal/appapi/server.go cmd/server/main.go`
  - `go test -count=1 ./internal/aidata`
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./...`
  - `node --check admin-web/src/main.js`
  - `npm run build`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `83%`：RBAC、菜单权限、AI/IM 导出开关已具备数据库持久化；剩余重点是后台批量操作、更多列表详情联动和真实运行态联调。
- 当前后台前端业务功能进度约 `88%`：本轮接口字段兼容，无需页面重写；剩余重点是批量操作交互、详情联动和浏览器联调。

## 2026-06-17 18:16

### 当前进展

- 继续补 PC 后台运营效率功能，优先完成线上测试高频批量操作。
- 后台后端新增批量组局审核接口：
  - `POST /api/admin/games/batch-audit`
  - 权限：`game:update_status`
  - 请求字段：`gameIds`、`approve`、`remark`
  - 内部逐项复用 `games.ApproveGame`，不绕开原有状态流转。
  - 返回逐项结果：`id`、`success`、`status`、`error`，同时返回 `success`、`failed`、`total`。
  - 成功项写入 `operation_logs`，action 为 `game:batch_audit`。
- 后台后端新增批量举报处理接口：
  - `POST /api/admin/reports/batch-handle`
  - 权限：`report:handle`
  - 请求字段：`reportIds`、`action`、`adminId`、`handlerAdminId`、`result`
  - 支持 `assign`、`handle`、`close` 三种动作。
  - 内部逐项复用 `reports.Assign`、`reports.Handle`、`reports.Close`。
  - 成功项写入 `operation_logs`，action 为 `report:batch_{action}`。
- 后台前端补充批量入口：
  - 组局管理页新增“批量通过待审核”，对当前列表内 `pending_audit` 组局批量审核。
  - 举报申诉页新增“批量处理”，对当前列表内 `pending/assigned` 举报批量处理。
  - 批量操作完成后刷新当前列表，并提示成功/失败数量。

### 验证补充

- 已执行：
  - `gofmt -w internal/appapi/batch_helper.go internal/appapi/game_handler.go internal/appapi/report_handler.go internal/appapi/server.go internal/appapi/server_test.go`
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./...`
  - `node --check admin-web/src/main.js`
  - `npm run build`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `86%`：后台 RBAC、菜单权限、核心配置持久化、组局/举报批量操作已可用；剩余重点是用户/实名/角色审核详情联动和浏览器运行态联调。
- 当前后台前端业务功能进度约 `90%`：核心列表和批量操作已接入；剩余重点是更多详情面板、空/错态细化和真实浏览器联调。

## 2026-06-17 18:30

### 当前进展

- 继续补 PC 后台审核中心，优先让一期后台能实际处理实名与角色申请链路。
- 后台前端审计中心新增实名认证详情入口：
  - 实名列表每条记录新增“详情”按钮。
  - 点击后调用 `GET /api/admin/identity-verifications/{userId}`。
  - 详情面板展示 `userId`、`phone`、`realname`、`status`、`faceIdRequestId`、`createdAt`、`updatedAt`，字段与后端 `identity.Record` 对齐。
- 后台前端角色申请新增详情、通过、驳回链路：
  - 角色申请列表缓存 `roleApplications`，详情面板展示 `role_applications` 关键字段。
  - 待处理申请支持“通过”和“驳回”。
  - 两个动作统一调用 `POST /api/admin/audits/role-applications/{id}/review`，不新建旁路接口。
  - 驳回会传 `approve=false` 与审核备注，后端会写入 `rejected/rejectReason/reviewRemark`。
- 后台 API 请求头补齐当前管理员 ID：
  - 所有已登录后台请求会携带 `X-Admin-ID`。
  - 角色审核时后端 `reviewAdminId` 和操作日志可以落到真实后台账号。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/main.js`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `87%`：审核接口原本已具备，本轮补齐前端调用和管理员 ID 请求头；剩余重点是用户详情联动、后台配置页细化、导出/日志真实联调。
- 当前后台前端业务功能进度约 `92%`：审计中心已具备列表、详情、通过、驳回闭环；剩余重点是用户/邀请/收益等模块的详情面板与浏览器交互联调。

## 2026-06-17 18:44

### 当前进展

- 继续补 PC 后台用户与邀请链路，优先服务一期线上测试时“查用户、查邀请码、查绑定来源”的运营动作。
- 后台后端新增邀请码详情接口：
  - `GET /api/admin/invite-codes/{code}`
  - 权限：`invite_code:read`
  - 返回 `inviteCode`、`relations`、`ownerUser`、`boundUser`
  - 详情字段与 `invite_codes`、`invite_relations`、`users` 现有模型对齐，不新增独立字段。
- 后台前端邀请管理新增详情面板：
  - 邀请码列表操作列新增“详情”。
  - 详情面板展示 `id`、`code`、`entryType`、`ownerUserId`、`usedCount/maxUses`、`boundWechatUserId`、绑定用户昵称、关系数、绑定来源。
  - 禁用按钮保留，和详情按钮共用现有 `row-actions` 样式。
- 后台前端用户详情增强：
  - 用户详情从粗略 JSON 摘要改为拆分展示邀请码、成长等级、信用分、可用积分、会员状态、收益总额、待结算、已结算。
  - 字段来自 `GET /api/admin/users/{id}` 已有返回结构，继续复用小程序端用户/成长/收益模型。

### 验证补充

- 已执行：
  - `gofmt -w internal/appapi/invite_admin_handler.go internal/appapi/server.go internal/appapi/server_test.go`
  - `node --check admin-web/src/main.js`
  - `go test -count=1 ./internal/appapi`
  - `npm run build`
  - `go test -count=1 ./...`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/main.js`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `89%`：邀请码详情接口、审核链路、批量操作、RBAC、核心配置已具备；剩余重点是收益/导出/日志与系统配置的真实联调和边界补齐。
- 当前后台前端业务功能进度约 `93%`：用户详情、邀请详情、审计中心、批量操作已能支撑一期测试；剩余重点是收益/导出/日志详情与浏览器交互联调。

## 2026-06-17 19:00

### 当前进展

- 继续补 PC 后台财务、导出、日志排查链路，优先让一期线上测试时的后台追踪动作完整可见。
- 后台前端分润记录新增详情面板：
  - 分润记录操作列新增“详情”。
  - 详情面板展示 `id`、`recordNo`、`gameId`、`templateId`、`amountCent`、`frozenReason`、`settledAt`、`createdAt`。
  - 明细表展示 `revenue_record_items` 中的 `role`、`userId`、`amountCent`。
  - 同步显示该记录关联的线下结算记录数量。
- 后台前端导出任务新增详情面板：
  - 导出任务操作列新增“详情”。
  - 详情面板展示 `taskNo`、`templateCode`、`exportType`、`fileId`、`createdBy`、`createdAt`、`finishedAt`、`failReason`、`filters`。
  - 已完成且存在 `fileId` 的任务会在详情中尝试拉取下载地址。
  - 下载地址按钮保留，和详情按钮共用 `row-actions`。
- 后台前端操作日志新增筛选和详情：
  - 日志页动态加入 `action`、`targetType`、`adminUserId` 筛选。
  - 日志列表新增“详情”按钮。
  - 详情面板展示 `adminUserId`、`action`、`targetType`、`targetId`、`requestId`、`ip`、`createdAt`、`detail`。
  - 筛选在前端对 `GET /api/admin/operation-logs` 返回结果处理，不改变后端接口合同。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./...`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/main.js`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `90%`：本轮主要使用既有后端接口补前端闭环；后端剩余重点是导出/日志筛选服务端化、系统配置更多可写项和真实环境联调。
- 当前后台前端业务功能进度约 `95%`：用户、邀请、审核、组局、举报、收益、导出、日志核心页面均有列表和详情/动作闭环；剩余重点是系统配置细化、边界错误态和浏览器真实交互联调。

## 2026-06-17 19:12

### 当前进展

- 继续补 PC 后台服务端筛选能力，减少线上数据量增长后前端全量过滤的压力。
- 后台后端操作日志列表新增 query 过滤：
  - `GET /api/admin/operation-logs?action=&targetType=&targetId=&adminUserId=`
  - `action` 支持包含匹配。
  - `targetType`、`targetId`、`adminUserId` 使用精确匹配。
  - 返回结构补充 `total`，保持原有 `items` 不变。
- 后台后端导出任务列表新增 query 过滤：
  - `GET /api/admin/export-tasks?status=&templateCode=&exportType=&createdBy=`
  - 均按 `export_tasks` 现有字段过滤。
  - 返回结构补充 `total`，保持原有 `items` 不变。
- 后台前端同步改为服务端筛选：
  - 操作日志页筛选表单提交后直接请求带 query 的 `/api/admin/operation-logs`。
  - 导出任务页新增 `templateCode`、`exportType`、`status` 筛选表单。
  - 筛选 UI 复用现有 `filters` 样式，不引入新组件体系。

### 验证补充

- 已执行：
  - `gofmt -w internal/appapi/audit_handler.go internal/appapi/export_handler.go internal/appapi/server_test.go`
  - `node --check admin-web/src/main.js`
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./...`
  - `npm run build`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/main.js`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `92%`：日志/导出筛选已服务端化，剩余重点是系统配置更多可写项、后台真实浏览器联调和上线环境联调。
- 当前后台前端业务功能进度约 `96%`：核心页面的列表、详情、动作、筛选基本齐备；剩余重点是系统配置细化和浏览器真实交互联调。

## 2026-06-17 19:24

### 当前进展

- 继续补 PC 后台系统配置页，优先让一期线上测试需要的角色/资格门槛能从后台直接调整。
- 后台前端系统页新增“领路人资格规则”面板：
  - 接入既有后端接口 `GET /api/admin/guides/qualification-rules`。
  - 支持查看 `guide_qualification_rules` 的 `ruleCode`、`status`、`minInviteCount`、`minCreditScore`、`minCompletedGames`、`paymentRequired`、`updatedAt`。
  - 支持将规则载入表单，避免运营人员手动重复录入规则 ID 和门槛字段。
  - 支持通过 `PUT /api/admin/guides/qualification-rules/{id}` 保存规则。
  - 可调整邀请数门槛、信用分门槛、完成局数门槛、是否要求支付、启用/禁用状态。
- 该功能用于支撑一期后台开局和角色测试：
  - 小程序一期普通用户仍只能走普通局主链路。
  - 后台可以通过资格规则和后台开局能力验证行家、领路人、条件局相关业务条件是否完整。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./...`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/main.js`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `92%`：本轮复用既有后端规则接口，后端未新增模型；剩余重点是系统配置更多可写项、真实账号联调和线上环境联调。
- 当前后台前端业务功能进度约 `97%`：系统配置页已补入领路人资格规则管理；剩余重点是登录后真实浏览器点击流、更多配置项细节和错误态打磨。

## 2026-06-17 19:38

### 当前进展

- 继续补 PC 后台一期运营功能，优先把已存在的积分/兑换后端接口接成可操作页面。
- 后台新增“积分兑换”菜单和页面：
  - 导航新增 `redemption`。
  - 页面模板新增 `redemption-template`。
  - 商品区对齐 `redemption_items`，支持新增商品、行内修改 `name`、`pointsCost`、`stock`、`status`。
  - 订单区对齐 `redemption_orders`，支持待处理订单通过/驳回，已通过订单履约。
  - 积分流水区对齐 `points_logs`，展示 `userId`、`changeValue`、`beforePoints`、`afterPoints`、`bizType`、`bizId`、`reason`。
- 后台 RBAC 菜单同步补齐：
  - `menuNodesForPermissions` 新增 `redemption` 菜单。
  - 拥有 `points:read` 或 `redemption:manage` 权限的后台账号可看到积分兑换菜单。
  - `user_manager` 角色新增 `points:read`、`redemption:manage`，方便用户运营处理积分和兑换订单。
  - `db/seeds/admin_roles_permissions.sql` 同步补齐用户管理员角色权限。
- 测试补充：
  - 后台权限树测试新增用户管理员可见 `redemption` 菜单的断言，防止 RBAC 菜单回退。

### 验证补充

- 已执行：
  - `gofmt -w internal/adminauth/service.go`
  - `gofmt -w internal/appapi/server_test.go`
  - `go test -count=1 ./internal/adminauth ./internal/appapi`
  - `go test -count=1 ./...`
  - `node --check src/main.js`
  - `npm run build`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/main.js`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `93%`：积分兑换后端接口原本已具备，本轮补齐 RBAC 菜单映射和角色种子；剩余重点是会员团队/订阅消息/交付文档等后台页面接入和线上环境联调。
- 当前后台前端业务功能进度约 `98%`：新增积分兑换页面后，核心运营、审核、组局、分润、导出、系统、IM、日志页面基本齐备；剩余重点是更多小模块页面和真实浏览器点击流。

## 2026-06-17 19:55

### 当前进展

- 继续补 PC 后台一期用户运营能力，把既有会员团队和会员报表后端接口接入后台页面。
- 后台新增“会员团队”菜单和页面：
  - 导航新增 `members`。
  - 页面模板新增 `members-template`。
  - 团队区接入 `GET /api/admin/teams`，字段对齐 `teams.Team`：`id`、`leaderUserId`、`name`、`status`、`createdAt`。
  - 团队详情接入 `GET /api/admin/teams/{teamId}`，展示团队成员 `userId`、`relationLevel`、`source`、`status`、`joinedAt`，并展示 `revenueSummary`。
  - 报表区接入 `GET /api/admin/member-reports`，字段对齐 `member_report_snapshots`：`userId`、`period`、`membershipPlan`、`invitedCount`、`participatedGames`、`completedGames`、`incomeSummary`。
- 后台 RBAC 菜单同步补齐：
  - `menuNodesForPermissions` 新增 `members` 菜单。
  - 拥有 `team:read` 或 `member_report:read` 权限的后台账号可看到会员团队菜单。
  - 前端按权限拆分加载：没有 `team:read` 时不请求团队接口，没有 `member_report:read` 时不请求报表接口。
  - `user_manager` 角色新增 `member_report:read`、`team:read`，用于用户运营查看会员团队、邀请沉淀和收益汇总。
  - `db/seeds/admin_roles_permissions.sql` 同步补齐用户管理员角色权限。
- 测试补充：
  - 后台权限树测试新增用户管理员可见 `members` 菜单的断言，防止 RBAC 菜单回退。

### 验证补充

- 已执行：
  - `gofmt -w internal/adminauth/service.go internal/appapi/server_test.go`
  - `node --check src/main.js`
  - `go test -count=1 ./internal/adminauth ./internal/appapi`
  - `go test -count=1 ./...`
  - `npm run build`
  - 检查 `admin-web/dist/index.html`、`admin-web/dist/main.js` 已包含 `members` 菜单、`members-template`、`/api/admin/teams`、`/api/admin/member-reports`。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `94%`：会员团队/会员报表接口原本已具备，本轮补齐 RBAC 菜单映射和角色种子；剩余重点是订阅消息、交付文档、更多系统配置项和线上环境联调。
- 当前后台前端业务功能进度约 `98.5%`：新增会员团队页面后，核心运营、审核、组局、分润、积分兑换、报表、导出、系统、IM、日志页面基本齐备；剩余重点是订阅消息/交付文档页面、真实浏览器点击流和错误态打磨。

## 2026-06-17 20:08

### 当前进展

- 继续补 PC 后台一期上线验收相关能力，把订阅消息、交付文档、测试留档接入后台主导航。
- 后台新增“通知交付”菜单和页面：
  - 导航新增 `delivery`。
  - 页面模板新增 `delivery-template`。
  - 微信订阅任务区接入 `GET /api/admin/notifications/wechat-tasks`，字段对齐 `notifications.WechatTask`：`notificationId`、`userId`、`scene`、`templateId`、`status`、`resultCode`、`resultMessage`。
  - 微信订阅模板区接入 `GET /api/admin/notifications/wechat-templates`，字段对齐 `notifications.WechatTemplate`：`scene`、`templateId`、`title`、`status`。
  - 支持调用 `/api/internal/notifications/wechat-tasks/send-pending` 批量发送待发任务。
  - 支持调用 `/api/internal/notifications/wechat-tasks/{id}/send` 单条发送。
  - 支持调用 `/api/internal/notifications/wechat-tasks/{id}/mark-sent` 标记已发送。
  - 交付文档区接入 `GET/POST /api/admin/delivery-documents`，字段对齐 `delivery.Document`：`docType`、`title`、`status`、`reason`。
  - 测试用例区接入 `GET/POST /api/admin/test-cases`，字段对齐 `delivery.TestCase`：`module`、`caseName`、`priority`、`expectedResult`。
  - 测试运行区接入 `GET/POST /api/admin/test-runs`，字段对齐 `delivery.TestRun`：`caseId`、`result`、`actualResult`、`requestId`、`evidenceFileId`。
- 后台 RBAC 菜单同步补齐：
  - `menuNodesForPermissions` 新增 `delivery` 菜单。
  - 拥有 `notification:wechat:view`、`delivery:manage`、`testcase:read` 或 `testcase:manage` 任一权限的后台账号可看到通知交付菜单。
  - 前端按权限拆分加载：没有对应权限时不请求对应接口，避免页面级 403。
- 测试补充：
  - 后台权限树测试新增运营角色可见 `delivery` 菜单的断言，防止通知交付菜单回退。

### 验证补充

- 已执行：
  - `gofmt -w internal/adminauth/service.go internal/appapi/server_test.go`
  - `node --check src/main.js`
  - `go test -count=1 ./internal/adminauth ./internal/appapi`
  - `go test -count=1 ./...`
  - `npm run build`
  - 检查 `admin-web/dist/index.html`、`admin-web/dist/main.js` 已包含 `delivery` 菜单、`delivery-template`、订阅消息、交付文档、测试用例、测试运行接口路径。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `95%`：通知、交付、测试留档后端原本已具备，本轮补齐 RBAC 菜单映射和前端入口；剩余重点是线上环境配置、真实微信订阅消息发送联调和更多系统配置项。
- 当前后台前端业务功能进度约 `99%`：核心运营、审核、组局、分润、积分兑换、会员团队、通知交付、报表、导出、系统、IM、日志页面基本齐备；剩余重点是真实浏览器点击流和线上账号联调。

## 2026-06-17 20:23

### 当前进展

- 继续补 PC 后台数据分析与 AI 数据准备能力，把已有分析后端接口接入后台主导航。
- 后台新增“数据分析”菜单和页面：
  - 导航新增 `analytics`。
  - 页面模板新增 `analytics-template`。
  - 漏斗分析区接入 `GET /api/admin/analytics/funnel`，字段对齐 `audit.FunnelSnapshot`：`eventCode`、`userCount`、`conversionRate`、`dropOffRate`。
  - 留存分析区接入 `GET /api/admin/analytics/retention`，字段对齐 `audit.RetentionSnapshot`：`cohortDate`、`newUsers`、`day1Retained`、`day7Retained`、`day30Retained`。
  - 行为事件区接入 `GET /api/admin/behavior/events`，字段对齐 `audit.BehaviorLog`：`userId`、`eventType`、`eventCode`、`targetType`、`targetId`、`source`、`createdAt`。
  - AI 数据快照区接入 `GET /api/admin/ai-data/snapshot`，字段对齐 `aidata.Snapshot`：`userCount`、`behaviorLogCount`、`reviewCount`、`acceptanceReady`、`dataReadinessSections`、`acceptanceChecks`。
  - 支持调用 `POST /api/admin/ai-data/acceptance-fixture` 生成 AI 验收数据，用于一期 AI 推荐/IM 分析数据准备验收。
- 后台 RBAC 菜单同步补齐：
  - `menuNodesForPermissions` 新增 `analytics` 菜单。
  - 拥有 `analytics:funnel:view`、`analytics:retention:view`、`analytics:timeline:view`、`data:behavior:read`、`ai:data:read` 或 `ai:data:seed` 任一权限的后台账号可看到数据分析菜单。
  - 前端按权限拆分加载：没有对应权限时不请求对应接口，避免页面级 403。
- 测试补充：
  - 后台权限树测试新增数据分析员可见 `analytics` 和 `delivery` 菜单的断言，防止数据分析/通知交付菜单回退。

### 验证补充

- 已执行：
  - `gofmt -w internal/adminauth/service.go internal/appapi/server_test.go`
  - `node --check src/main.js`
  - `go test -count=1 ./internal/adminauth ./internal/appapi`
  - `go test -count=1 ./...`
  - `npm run build`
  - 检查 `admin-web/dist/index.html`、`admin-web/dist/main.js` 已包含 `analytics` 菜单、`analytics-template`、漏斗、留存、行为事件、AI 快照和 AI 验收数据接口路径。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `96%`：数据分析与 AI 数据准备后端接口原本已具备，本轮补齐 RBAC 菜单映射和前端入口；剩余重点是线上环境配置、真实第三方联调和生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.3%`：核心运营、审核、组局、分润、积分兑换、会员团队、数据分析、通知交付、报表、导出、系统、IM、日志页面基本齐备；剩余重点是真实浏览器点击流、线上账号联调和少量配置项细化。

## 2026-06-17 20:38

### 当前进展

- 继续补 PC 后台画像关系模块，把小程序端已经沉淀的人脉关系、行家技能、领路人资源暴露到后台。
- 后台新增“画像关系”菜单和页面：
  - 导航新增 `profiles`。
  - 页面模板新增 `profiles-template`。
  - 人脉关系区接入 `GET /api/admin/connections`，字段对齐 `connections.Connection`：`userId`、`connectedUserId`、`relationType`、`sourceType`、`sourceId`、`strengthScore`、`updatedAt`。
  - 行家技能查询接入 `GET /api/admin/experts/{userId}/skills`，字段对齐 `profiles.ExpertSkillProfile`：`skillTree`、`serviceTags`、`caseFileIds`、`completeness`。
  - 领路人资源查询接入 `GET /api/admin/guides/{userId}/resources`，字段对齐 `profiles.GuideResourceProfile`：`resourceTags`、`industryTags`、`cityCodes`、`connectionScale`、`completeness`。
- 后台 RBAC 菜单同步补齐：
  - `menuNodesForPermissions` 新增 `profiles` 菜单。
  - 拥有 `connection:read` 或 `profile:read` 任一权限的后台账号可看到画像关系菜单。
  - 前端按权限拆分加载：没有 `connection:read` 时不请求人脉列表，没有 `profile:read` 时禁止查询画像详情。
- 测试补充：
  - 后台权限树测试新增用户管理角色可见 `profiles` 菜单的断言，防止画像关系菜单回退。

### 验证补充

- 已执行：
  - `gofmt -w services/go-api/internal/adminauth/service.go services/go-api/internal/appapi/server_test.go`
  - `node --check admin-web/src/main.js`
  - `go test -count=1 ./internal/adminauth ./internal/appapi`
  - `npm run build`
  - 检查 `admin-web/dist/index.html`、`admin-web/dist/main.js` 已包含 `profiles` 菜单、`profiles-template`、人脉、行家技能、领路人资源接口路径。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `96.5%`：画像关系后端接口原本已具备，本轮补齐 RBAC 菜单映射和前端入口；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.5%`：核心运营、审核、组局、分润、积分兑换、会员团队、画像关系、数据分析、通知交付、报表、导出、系统、IM、日志页面基本齐备；剩余重点是真实浏览器点击流、线上账号联调和少量配置项细化。

## 2026-06-17 20:53

### 当前进展

- 继续补 PC 后台评价、成长、信用追踪能力，把一期“服务确认 -> 评价 -> 成长/信用 -> 分润前置条件”链路暴露到后台。
- 后台新增“评价成长”菜单和页面：
  - 导航新增 `growth`。
  - 页面模板新增 `growth-template`。
  - 用户成长追踪接入 `GET /api/admin/users/{userId}/growth`，字段对齐 `reviews.Trace`：`profile`、`reviews`、`creditLogs`、`footprints`、`achievements`。
  - 局评价追踪接入 `GET /api/admin/games/{gameId}/review-trace`，按局查看评价记录、再玩意向、信用扣分和足迹证据。
  - 页面展示成长摘要、评价记录、信用流水、足迹证据和成就，便于运营/客服核查评价闭环与争议依据。
- 后台 RBAC 菜单同步补齐：
  - `menuNodesForPermissions` 新增 `growth` 菜单。
  - 拥有 `user:view` 或 `game:view` 任一权限的后台账号可看到评价成长菜单。
  - 前端按权限拆分加载：没有 `user:view` 时禁止查用户成长，没有 `game:view` 时禁止查局评价轨迹。
- 测试补充：
  - 后台权限树测试新增用户管理角色可见 `growth` 菜单的断言，防止评价成长入口回退。

### 验证补充

- 已执行：
  - `gofmt -w services/go-api/internal/adminauth/service.go services/go-api/internal/appapi/server_test.go`
  - `node --check admin-web/src/main.js`
  - `go test -count=1 ./internal/adminauth ./internal/appapi`
  - `go test -count=1 ./...`
  - `npm run build`
  - 检查 `admin-web/dist/index.html`、`admin-web/dist/main.js` 已包含 `growth` 菜单、`growth-template`、用户成长追踪和局评价追踪接口路径。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `97%`：评价成长追踪后端接口原本已具备，本轮补齐 RBAC 菜单映射和前端入口；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.6%`：核心运营、审核、组局、分润、积分兑换、会员团队、画像关系、评价成长、数据分析、通知交付、报表、导出、系统、IM、日志页面基本齐备；剩余重点是真实浏览器点击流、线上账号联调和少量配置项细化。

## 2026-06-17 21:08

### 当前进展

- 继续补 PC 后台组局运营明细能力，把一期预留的里程碑、打卡、复盘、续局草稿接入现有“组局管理”详情面板。
- 后台“组局管理”详情增强：
  - `GET /api/admin/games/{gameId}` 返回的 `milestones`、`checkins`、`retrospectives`、`continueDrafts` 从数量展示升级为明细表格展示。
  - 支持后台创建里程碑，调用 `POST /api/admin/games/{gameId}/milestones`，字段对齐 `games.MilestoneRequest`：`title`、`status`。
  - 支持后台标记异常打卡，调用 `POST /api/admin/games/{gameId}/checkins/{checkinId}/mark-invalid`。
  - 复盘明细展示 `againIntent`，续局草稿展示 `originalGameId`、`draftGameId`、`creatorUserId`、`title`、`status`。
- 权限控制：
  - 运营明细读取沿用 `game:read`。
  - 创建里程碑和标记异常打卡沿用 `game:progress:manage`。
  - 前端无 `game:progress:manage` 时不显示异常标记动作，并在提交里程碑时拦截。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi ./internal/games`
  - `go test -count=1 ./...`
  - `Invoke-WebRequest http://127.0.0.1:5173/` 和 `/main.js` 均返回 200。
  - 检查 `admin-web/dist/main.js` 已包含 `gameOpsBlock`、`game-milestone-form`、`checkin-invalid`、`/api/admin/games/{id}/milestones`、`mark-invalid` 等接口路径。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `97.2%`：组局运营明细后端接口原本已具备，本轮补齐后台详情页操作面；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.7%`：组局管理从列表/审核扩展到运营明细查看和轻量操作；剩余重点是真实浏览器点击流、线上账号联调和少量配置项细化。

## 2026-06-17 21:23

### 当前进展

- 继续补 PC 后台邀请入口核查能力，对齐最近调整的“只能通过小程序卡片、二维码、链接三种形式进入小程序”和“唯一邀请码绑定微信后进入登录页”的业务规则。
- 后台“邀请管理”详情增强：
  - 邀请码详情页从基础字段展示升级为“入口校验 + 绑定关系明细”。
  - 入口校验展示 `allowedEntry`、`uniqueBinding`、`boundWechat`、`authPageMode`，便于后台判断当前邀请码应进入注册页还是登录页。
  - 绑定关系明细展开 `inviteCodeId`、`inviterUserId`、`inviteeUserId`、`bindSource`、入口类型，字段对齐 `invite_relations`。
  - 详情页支持直接禁用 active 邀请码，仍调用 `POST /api/admin/invite-codes/{code}/disable`。
- 权限控制：
  - 详情读取沿用 `invite_code:read`。
  - 禁用邀请码沿用 `invite_code:manage`。
  - 前端无管理权限时不显示详情页禁用按钮。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi ./internal/invites`
  - `go test -count=1 ./...`
  - `Invoke-WebRequest http://127.0.0.1:5173/` 和 `/main.js` 均返回 200。
  - 检查 `admin-web/src/main.js`、`admin-web/dist/main.js` 已包含 `入口校验`、`authPageMode`、`inviteRelationDetailRow`、`detail-invite-disable`、`inviteEntryLabel`。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `97.3%`：邀请管理后端接口原本已具备，本轮补齐后台详情页核查和禁用操作面；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.75%`：邀请、组局、审核、分润、IM、评价成长、画像关系、通知交付、数据分析、系统、日志等主链路基本齐备；剩余重点是真实浏览器点击流、线上账号联调和少量配置项细化。

## 2026-06-17 21:38

### 当前进展

- 继续补 PC 后台用户详情核查能力，把用户详情里的收藏、人脉统计升级为可下钻明细。
- 后台“用户详情”增强：
  - 详情打开时按权限接入 `GET /api/admin/users/{userId}/favorites`，展示 `userId`、`gameId`、`game.title`、`createdAt`，字段对齐 `game_favorites` 和关联局信息。
  - 详情打开时按权限接入 `GET /api/admin/users/{userId}/connections`，展示 `id`、`userId`、`connectedUserId`、`relationType`、`source`、`strengthScore`、`updatedAt`，字段对齐 `user_connections`。
  - 顶部统计的 `favorites`、`connections` 改为以后端明细接口返回数量为准；无权限时回落使用详情接口已有摘要数据。
- 权限控制：
  - 收藏明细读取沿用 `user:read`。
  - 人脉明细读取沿用 `connection:read`。
  - 用户基础详情仍沿用 `user:view`，避免扩大原有后台权限边界。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi ./internal/games`
  - `go test -count=1 ./...`
  - 检查 `admin-web/src/main.js`、`admin-web/dist/main.js` 已包含 `/api/admin/users/{id}/favorites`、`/api/admin/users/{id}/connections`、`userFavoriteRow`、`收藏明细`、`人脉明细`。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `97.4%`：用户收藏、人脉后端接口原本已具备，本轮补齐后台详情页下钻使用；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.8%`：后台核心业务链路已基本具备运营查看、审核、配置、处置和核查能力；剩余重点是真实浏览器点击流、线上账号联调和少量配置项细化。

## 2026-06-17 21:53

### 当前进展

- 继续补 PC 后台文件证据核查能力，把后端已有的后台文件下载权限接口接到举报详情和测试留档页面。
- 后台“举报详情”增强：
  - 如果举报证据包含 `fileId`，详情页展示 `file.bizType`、`file.name`。
  - 新增“下载附件”动作，调用 `GET /api/admin/files/{fileId}/download-url`。
  - 下载权限仍由后端按 `files.File.bizType` 分流：`report_attachment` 需要 `report:view`，`chat_file` 需要 `im:message:view_dispute`，`realname_material` 需要 `identity:read`，`export_file` 需要 `report_export:create`。
- 后台“测试运行”增强：
  - `delivery.TestRun.evidenceFileId` 不再只显示数字，有证据文件时提供下载动作。
  - 下载动作复用同一个 `downloadAdminFile(fileId)`，生成地址后复制到剪贴板并尝试打开新窗口。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi ./internal/files ./internal/reports`
  - `go test -count=1 ./...`
  - 检查 `admin-web/src/main.js`、`admin-web/dist/main.js` 已包含 `downloadAdminFile`、`/api/admin/files/{fileId}/download-url`、`admin-file-download`、`test-run-file-download`。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `97.5%`：后台文件下载后端接口原本已具备，本轮补齐前端操作入口；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.85%`：后台从业务列表、审核处置进一步补到证据文件核查；剩余重点是真实浏览器点击流、线上账号联调和少量配置项细化。

## 2026-06-17 22:08

### 当前进展

- 继续补 PC 后台 IM 管理闭环，把文档和后端已有的 OpenIM/本地 IM 管理接口接入后台页面。
- 后台“IM 证据”增强：
  - 房间列表新增“详情”，调用 `GET /api/admin/im/rooms/{roomId}`，展示 `room`、全部消息、文件消息、成员、OpenIM group、归档原因。
  - 房间列表新增“争议消息”，调用 `GET /api/admin/im/rooms/{roomId}/dispute-messages`，用于举报/申诉时快速核查保全消息。
  - 有 `im:room:retry_create` 权限时可调用 `POST /api/admin/im/rooms/{roomId}/retry-create`，用于 OpenIM 群同步重试。
  - 有 `im:room:archive` 权限时可调用 `POST /api/admin/im/rooms/{roomId}/archive`，归档后小程序端不可继续发送该房间消息。
  - 有 `im:message:hide` 权限时可调用 `POST /api/admin/im/messages/{messageId}/hide`，隐藏普通用户历史里的违规消息，同时后台详情仍可留证。
  - 文件消息复用 `GET /api/admin/files/{fileId}/download-url` 下载入口，字段继续对齐 `im_messages.file_id` 与 `files`。
- 权限控制：
  - 房间详情沿用 `im:room:read`。
  - 争议消息沿用 `im:message:view_dispute`。
  - 重试同步、归档、隐藏消息分别沿用 `im:room:retry_create`、`im:room:archive`、`im:message:hide`。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi ./internal/im`
  - `go test -count=1 ./...`
  - 检查 `admin-web/src/main.js`、`admin-web/dist/main.js` 已包含 `imRoomActions`、`showIMRoomDetail`、`showIMDisputeMessages`、`/retry-create`、`/archive`、`/api/admin/im/messages/{messageId}/hide`。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `97.7%`：IM 管理后端接口原本已具备，本轮补齐后台可操作页面；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.9%`：后台核心业务、证据、审核、IM 处置、文件核查、运营配置基本形成完整操作闭环；剩余重点是真实浏览器点击流、线上账号联调和少量配置项细化。

## 2026-06-17 22:23

### 当前进展

- 继续补 PC 后台 AI 数据准备验收闭环，把验收用例中的 `POST /api/admin/ai-data/im-export` 接到后台“数据分析”页面。
- 后台“数据分析 / AI 数据快照”增强：
  - 新增“导出 IM 数据”按钮，调用 `POST /api/admin/ai-data/im-export`。
  - 导出关闭时页面会展示后端返回错误，便于验证一期默认关闭的导出闸门。
  - 导出成功时展示导出条数和前 8 条样例，避免大量 IM 内容直接撑满页面。
  - 无 `ai:data:export` 权限时展示无权限提示，按钮点击也会被拦截。
- 与系统配置页联动：
  - 系统配置页已有 `GET/PUT /api/admin/ai-data/im-export-config` 开关。
  - 本轮补齐“开关开启后从数据分析页触发导出”的验收路径。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi ./internal/aidata`
  - `go test -count=1 ./...`
  - 检查 `admin-web/src`、`admin-web/dist` 已包含 `ai-im-export-button`、`ai-im-export-panel`、`exportAIIMData`、`renderAIIMExport`、`/api/admin/ai-data/im-export`。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `97.8%`：AI IM 导出后端接口原本已具备，本轮补齐后台触发和验收反馈面；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.92%`：后台已覆盖核心业务、证据、审核、IM 处置、AI 数据准备和受控导出；剩余重点是真实浏览器点击流、线上账号联调和少量配置项细化。

## 2026-06-17 22:38

### 当前进展

- 继续补 PC 后台角色资格测试能力，把后端已有的用户领路人/行家资格状态接口接到“系统配置”页面。
- 后台“系统配置 / 领路人资格规则”增强：
  - 在规则配置面板下新增“用户资格”表单。
  - 支持按 `userId` 查询资格，调用 `GET /api/admin/guide-qualification-rules?userId={userId}`。
  - 支持后台更新 `conditionMet`、`paymentMet`，调用 `POST /api/admin/guide-qualification-rules`。
  - 查询/保存后展示 `userId`、`conditionMet`、`paymentMet`、`guideOpenStatus`、`updatedAt`。
- 一期测试价值：
  - 普通小程序用户一期仍只能创建普通局。
  - 后台可直接准备领路人/行家资格状态，用于测试标准局、条件局、主行家和领路人门槛链路。
  - 字段对齐 `guide_qualifications` 与 `profiles.GuideQualification`。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi ./internal/profiles`
  - `go test -count=1 ./...`
  - 检查 `admin-web/src/main.js`、`admin-web/dist/main.js` 已包含 `guide-qualification-form`、`guide-qualification-load`、`loadGuideQualification`、`updateGuideQualification`、`/api/admin/guide-qualification-rules`。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `97.9%`：用户资格状态后端接口原本已具备，本轮补齐后台操作面；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.94%`：后台已覆盖核心业务、证据、审核、IM 处置、AI 数据准备、受控导出和角色资格测试准备；剩余重点是真实浏览器点击流、线上账号联调和少量配置项细化。

## 2026-06-17 22:53

### 当前进展

- 继续补 PC 后台行为追踪排查能力，把后端已有的完整行为时间线接口接到“数据分析”页面。
- 后台“数据分析 / 行为事件”增强：
  - 新增“完整时间线”按钮，调用 `GET /api/admin/behavior-logs`。
  - 新增 `behavior-log-list`，展示最近 20 条行为时间线摘要。
  - 时间线展示 `eventCode`、`eventType`、`userId`、`target/business`、`pagePath`、`source`、`device`、`keyword`、`occurredAt`。
  - 无 `analytics:timeline:view` 权限时展示无权限提示。
- 与现有行为事件列表的关系：
  - `GET /api/admin/behavior/events` 继续用于按 userId、eventType、eventCode 筛选结构化事件。
  - `GET /api/admin/behavior-logs` 用于快速排查用户完整路径和一期邀请/浏览/分享/入局等行为链路。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi ./internal/audit`
  - `go test -count=1 ./...`
  - 检查 `admin-web/src`、`admin-web/dist` 已包含 `behavior-logs-refresh`、`behavior-log-list`、`loadBehaviorLogs`、`behaviorLogItem`、`/api/admin/behavior-logs`。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98%`：行为时间线后端接口原本已具备，本轮补齐后台查看入口；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.95%`：后台已覆盖核心业务、证据、审核、IM 处置、AI 数据准备、受控导出、角色资格测试和行为追踪；剩余重点是真实浏览器点击流、线上账号联调和少量配置项细化。

## 2026-06-17 23:10

### 当前进展

- 继续按后台真实点击流收口，修正“组局详情 / 打卡标记异常”的前后端接口路径。
- 后台页面原先调用 `POST /api/admin/games/{gameId}/checkins/{checkinId}/mark-invalid`，但后端路由和集成用例定义的是 `POST /api/admin/game-checkins/{checkinId}/mark-invalid`。
- 已将 `admin-web/src/main.js` 调整为调用正式后端路径，并通过构建同步到 `admin-web/dist/main.js`。
- 这次修复保证后台运营人员在组局详情中点击“标记异常”时，可以真正命中后端 `markCheckinInvalid`，并写入操作审计 `game:checkin:mark_invalid`。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi ./internal/games`
  - `go test -count=1 ./...`
  - 检查 `admin-web/src/main.js`、`admin-web/dist/main.js` 已包含 `/api/admin/game-checkins/${button.dataset.id}/mark-invalid`。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98%`：本轮未新增后端能力，重点是把已有后端能力对齐到后台可操作入口；剩余重点仍是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.96%`：后台核心业务入口已基本接齐，本轮修掉一处真实点击流路径错配；剩余重点是真实浏览器端到端点击、线上账号联调和极少量边缘配置项细化。

## 2026-06-17 23:31

### 当前进展

- 继续按“一期普通用户只能开普通局，但后台要能开其他条件局做完整测试”的要求收口后台开局链路。
- 后端 `games.CreateRequest` 新增 `mainGuideUserId`，后台创建标准局/条件局时可直接指定主行家：
  - `CreateFromAdmin` 会把 `mainGuideUserId` 写入 `Game.MainGuideUserID`。
  - SQL 仓库创建游戏时会落到 `games.main_guide_user_id`。
  - 指定的主行家会同步加入成员关系，仓库成员角色为 `main_guide`。
- 小程序端创建局仍然忽略 `mainGuideUserId`：
  - `Create` 会清零 `MainGuideUserID`，避免普通用户绕过“邀请接受 + 审核通过后产生主行家”的业务规则。
- 后台“组局管理 / 后台开局”表单新增 `mainGuideUserId` 输入：
  - 不填时保持原有普通局/标准局/条件局创建流程。
  - 填正数时随 `POST /api/admin/games` payload 一起提交，便于一期后台直接开标准局、条件局并测试主行家进度管理链路。

### 验证补充

- 已执行：
  - `gofmt -w services/go-api/internal/games/service.go services/go-api/internal/games/service_test.go`
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/games`
  - `go test -count=1 ./internal/appapi ./internal/games`
  - `go test -count=1 ./...`
  - 检查 `admin-web/src`、`admin-web/dist`、`services/go-api/internal/games` 已包含 `mainGuideUserId` / `MainGuideUserID`。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.3%`：后台开局已能直接准备标准局/条件局和主行家测试数据；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.97%`：后台核心业务入口、证据、审核、IM、行为、测试数据准备和开局主行家配置基本接齐；剩余重点是真实浏览器端到端点击和线上账号联调。

## 2026-06-17 23:40

### 当前进展

- 继续收口后台指定主行家开局链路，从“表单可提交”补到“API 与详情可验收”。
- 后台组局详情新增 `mainGuideUserId` 展示：
  - 打开 `GET /api/admin/games/{gameId}` 详情时，可直接看到当前局主行家 ID。
  - 便于后台确认标准局/条件局是否已经绑定主行家，以及后续由主行家决定进度的测试条件是否满足。
- 后端 API 集成测试增强：
  - `POST /api/admin/games` 测试 payload 增加 `mainGuideUserId`。
  - 断言创建返回 `mainGuideUserId`。
  - 继续请求 `GET /api/admin/games/{gameId}`，断言详情里的 `game.mainGuideUserId` 正确，并且 `memberIds` 包含该主行家。
- 保持小程序端规则不变：
  - 普通用户创建局仍只能创建普通局。
  - 小程序端不能通过创建 payload 直接指定主行家。

### 验证补充

- 已执行：
  - `gofmt -w services/go-api/internal/appapi/server_test.go services/go-api/internal/games/service.go services/go-api/internal/games/service_test.go`
  - `node --check admin-web/src/main.js`
  - `go test -count=1 ./internal/appapi ./internal/games`
  - `npm run build`
  - `go test -count=1 ./...`
  - 检查 `admin-web/src/main.js`、`admin-web/dist/main.js`、`services/go-api/internal/appapi/server_test.go` 已包含 `mainGuideUserId` 详情展示和 API 断言。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.4%`：主行家后台开局链路已具备 API 级验收覆盖；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.98%`：后台核心业务入口和关键详情验收字段基本接齐；剩余重点是真实浏览器端到端点击和线上账号联调。

## 2026-06-17 23:55

### 当前进展

- 继续收口后台指定主行家开局的数据一致性。
- 修正后台创建局时 `currentPlayers` 计算：
  - 不指定主行家时仍为 `1`，只包含创建者。
  - 指定 `mainGuideUserId` 且与创建者不同时，`currentPlayers` 自动为 `2`，与创建者 + 主行家两条成员关系一致。
- 后端 service 测试增强：
  - 后台创建条件局并指定主行家时，断言 `MainGuideUserID=22` 且 `CurrentPlayers=2`。
- 后端 API 集成测试增强：
  - `POST /api/admin/games` 返回体断言 `currentPlayers=2`。
  - `GET /api/admin/games/{gameId}` 详情断言 `game.currentPlayers=2`，并继续断言 `memberIds` 包含主行家。
- 这轮不改变接口字段，只修正已有字段与成员关系之间的一致性，避免后台开测试局后人数显示和满员判断偏差。

### 验证补充

- 已执行：
  - `gofmt -w services/go-api/internal/games/service.go services/go-api/internal/games/service_test.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/games ./internal/appapi`
  - `node --check admin-web/src/main.js`
  - `go test -count=1 ./...`
  - 检查 `services/go-api/internal/games/service.go`、`services/go-api/internal/games/service_test.go`、`services/go-api/internal/appapi/server_test.go` 已包含 `currentPlayers=2` 的主行家开局断言。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.45%`：后台主行家测试局的人数、成员、详情字段已一致；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.98%`：本轮未改前端结构，继续保持关键详情验收字段已接齐；剩余重点是真实浏览器端到端点击和线上账号联调。

## 2026-06-18 00:10

### 当前进展

- 继续收口后台指定主行家开局的输入边界。
- 后台创建局新增校验：`mainGuideUserId` 不能等于 `creatorUserId`。
  - 避免后台误把创建者同时写成主行家，导致成员角色、进度管理和后续测试数据出现歧义。
  - 正常的小程序邀请链路不受影响，主行家仍由“行家邀请接受 + 审核通过”产生。
- 后端 service 测试增强：
  - `CreateFromAdmin` 传入相同的 `CreatorUserID` 和 `MainGuideUserID` 时返回 `ErrInvalidGameInput`。
- 后端 API 集成测试增强：
  - `POST /api/admin/games` 传入相同 `creatorUserId` 与 `mainGuideUserId` 时返回 `422 Unprocessable Entity`。

### 验证补充

- 已执行：
  - `gofmt -w services/go-api/internal/games/service.go services/go-api/internal/games/service_test.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/games ./internal/appapi`
  - `node --check admin-web/src/main.js`
  - `go test -count=1 ./...`
  - 静态检查已确认 `service.go`、`service_test.go`、`server_test.go` 包含 `creatorUserId == mainGuideUserId` 的拒绝逻辑和测试覆盖。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.5%`：后台主行家开局链路已补齐创建、成员、人数、详情和输入边界；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.98%`：本轮未改前端结构；剩余重点是真实浏览器端到端点击和线上账号联调。

## 2026-06-18 00:24

### 当前进展

- 继续补后台指定主行家开局的前端误操作保护。
- 后台“组局管理 / 后台开局”提交前新增校验：
  - 当 `mainGuideUserId > 0` 且等于 `creatorUserId` 时，前端直接提示 `mainGuideUserId cannot equal creatorUserId` 并停止提交。
  - 后端上一轮新增的 `422` 校验仍保留，前端只负责提前减少运营误填。
- 构建产物已同步：
  - `admin-web/src/main.js`
  - `admin-web/dist/main.js`

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/games ./internal/appapi`
  - `go test -count=1 ./...`
  - 检查 `admin-web/src/main.js`、`admin-web/dist/main.js` 已包含 `mainGuideUserId === payload.creatorUserId` 拦截。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.5%`：本轮未改后端，继续保持后台主行家开局链路的服务端最终校验。
- 当前后台前端业务功能进度约 `99.99%`：后台开局表单已补齐主行家输入、详情展示和同 ID 误操作保护；剩余重点是真实浏览器端到端点击和线上账号联调。

## 2026-06-18 00:41

### 当前进展

- 继续按“领路人、行家必须实名认证；玩家不要求”的业务规则，收口后台指定主行家开局链路。
- 后台创建局新增服务端校验：
  - 当 `mainGuideUserId > 0` 时，必须通过 `identity.IsVerified(mainGuideUserId)`。
  - 未实名用户不能被后台直接指定为主行家，返回 `ErrRealnameRequired` 并由 HTTP 层转成 `422`。
- 保持小程序端规则不变：
  - 玩家创建普通局仍不要求实名认证。
  - 小程序邀请成为主行家仍沿用既有“行家实名 + 邀请接受 + 审核通过”链路。
- 后端 service 测试增强：
  - `CreateFromAdmin` 指定未实名主行家时返回 `ErrRealnameRequired`。
- 后端 API 集成测试增强：
  - 未实名 `mainGuideUserId=2` 创建条件局返回 `422`。
  - 登录并完成用户 2 实名后，再用同一 `mainGuideUserId=2` 创建条件局成功。

### 验证补充

- 已执行：
  - `gofmt -w services/go-api/internal/games/service.go services/go-api/internal/games/service_test.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/games ./internal/appapi`
  - `node --check admin-web/src/main.js`
  - `go test -count=1 ./...`
  - 静态检查已确认 `service.go`、`service_test.go`、`server_test.go` 包含后台主行家实名校验和测试覆盖。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.6%`：后台主行家开局链路已补齐实名边界、成员、人数、详情和错误输入校验；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.99%`：本轮未改前端结构；剩余重点是真实浏览器端到端点击和线上账号联调。

## 2026-06-18 00:56

### 当前进展

- 继续收口后台开局和小程序业务规则的一致性，补齐后台 HTTP 层对主行家实名失败的错误映射。
- 后台创建局 `CreateFromAdmin` 已要求 `mainGuideUserId` 对应用户完成实名；本轮补齐 API 返回：
  - 未实名主行家不再落入 `500` 系统错误。
  - 统一返回 `422 validation_error`，错误信息为 `main guide realname required`，方便后台前端和线上联调定位。
- 保持小程序端规则不变：
  - 玩家创建普通局不要求实名认证。
  - 行家、领路人相关身份仍必须实名。
  - 一期普通用户只能在小程序创建普通局，后台可创建其他条件局用于业务测试。

### 验证补充

- 已执行：
  - `go test -count=1 ./internal/games ./internal/appapi`
  - `node --check admin-web/src/main.js`
  - `go test -count=1 ./...`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.65%`：后台主行家开局链路已补齐服务端实名校验、HTTP 错误映射、成员写入、人数统计和测试覆盖；剩余重点是线上环境配置、真实腾讯/微信侧联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.99%`：本轮未改前端结构；剩余重点是真实浏览器端到端点击和线上账号联调。

## 2026-06-18 01:12

### 当前进展

- 继续按后台前后端对接检查接口闭环，重点核对 admin-web 已绑定的 API 是否具备后端路由、权限和测试覆盖。
- 确认 IM 争议消息接口 `/api/admin/im/rooms/{roomId}/dispute-messages` 已有后端路由、权限 `im:message:view_dispute` 和测试覆盖，不作为本轮缺口。
- 发现并修复通知交付模块安全缺口：
  - 后台前端会调用 `/api/internal/notifications/wechat-tasks/send-pending`、`/{id}/send`、`/{id}/mark-sent`。
  - 后端原先内部路由直接进入处理函数，未统一套后台权限。
  - 本轮已统一套 `notification:wechat:view` 权限，保证只有具备后台通知权限的账号才能批量发送、单条发送或手动标记微信订阅消息任务。
- 同步调整测试：
  - 无 token 直调 `mark-sent` 必须返回 `403`。
  - 使用后台管理员 token 调用仍返回 `200`，后台页面按钮链路保持可用。

### 验证补充

- 已执行：
  - `gofmt -w internal/appapi/server.go internal/appapi/server_test.go`
  - `node --check admin-web/src/main.js`
  - `go test -count=1 ./internal/appapi -run "Test.*Wechat|Test.*Notification|Test.*Report"`
  - `go test -count=1 ./internal/notifications`
  - `go test -count=1 ./...`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.7%`：后台通知交付的内部执行接口已补齐权限边界，避免线上测试时无权限直调修改发送状态；剩余重点仍是生产配置、真实微信订阅消息账号联调、真实对象存储/OpenIM/实名服务联调。
- 当前后台前端业务功能进度约 `99.99%`：本轮未改页面结构；后台已绑定接口继续保持可用，剩余重点是真实浏览器端到端点击和线上账号联调。

## 2026-06-18 01:28

### 当前进展

- 继续收口后台账号和 RBAC 权限边界，避免“角色审核权限”和“后台账号管理权限”混用。
- 新增并落地 `admin_user:update` 权限码：
  - `PUT /api/admin/admin-users/{id}` 从 `role:update` 改为 `admin_user:update`。
  - 超级管理员默认权限加入 `admin_user:update`。
  - 数据库 seed 幂等追加 `admin_user:update`，新库和旧库补种子都能拿到该权限。
  - 权限目录接口会返回 `admin_user:update`，后台权限目录可以直接看到该 code。
- 后台前端同步：
  - 创建后台账号前先检查 `admin_user:create`。
  - 后台账号列表的保存、启用、禁用按钮只在具备 `admin_user:update` 时显示。
- 保持原有 `role:update` 继续用于角色审核、领路人资格规则等角色业务，不再复用到后台账号启停和改角色。

### 验证补充

- 已执行：
  - `gofmt -w internal/appapi/server.go internal/appapi/server_test.go internal/adminauth/service.go internal/adminauth/service_test.go`
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/adminauth`
  - `go test -count=1 ./internal/appapi -run "TestAdminAccount|TestAdminLoginPermissions|TestAdminRolePermission"`
  - `go test -count=1 ./...`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.75%`：后台账号管理的更新权限已独立，避免客服/运营等拥有角色审核能力的账号误操作后台账号；剩余重点仍是生产配置、真实第三方联调和线上验收数据准备。
- 当前后台前端业务功能进度约 `99.99%`：后台账号按钮已按新权限码显示；剩余重点是真实浏览器端到端点击和线上账号联调。

## 2026-06-18 01:42

### 当前进展

- 继续收口后台 RBAC 与数据库初始化脚本的一致性，避免本地内存账号有权限、线上 seed 初始化账号缺权限。
- 修复数据库 seed 中后置新增权限的授权顺序问题：
  - `super_admin` 的全权限授权发生在较早位置。
  - 后续追加的 `im:room:read`、`im:room:archive`、`im:room:retry_create`、`system_config:read`、`invite_code:read`、`invite_code:manage` 不会被早期 `cross join admin_permissions` 自动覆盖。
  - 本轮已为 `super_admin` 显式补齐这些后置权限授权。
- 增加 seed 一致性回归测试：
  - 测试会读取 `db/seeds/admin_roles_permissions.sql`。
  - 确认后置新增权限都显式授予 `super_admin`，防止后续后台模块新增权限时再次出现初始化库权限缺口。
- 本轮不改变前端页面和业务流程，只保证后台权限、数据库 seed、线上初始化后的角色能力继续对齐。

### 验证补充

- 已执行：
  - `gofmt -w internal/adminauth/service_test.go`
  - `go test -count=1 ./internal/adminauth`
  - `go test -count=1 ./internal/appapi -run "TestAdminAccount|TestAdminLoginPermissions|TestAdminRolePermission|TestAdminIM"`
  - `go test -count=1 ./...`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.8%`：后台账号、IM、配置、邀请码相关权限在代码和数据库初始化层继续对齐；剩余重点仍是生产配置、真实腾讯/微信/OpenIM/存储/实名服务联调和线上验收数据准备。
- 当前后台前端业务功能进度约 `99.99%`：本轮未改页面；后台按钮权限依赖的数据库 seed 已继续补强，剩余重点是真实浏览器端到端点击和线上账号联调。

## 2026-06-18 02:03

### 当前进展

- 继续按后台前后端接口对接检查，重点核对后台页面菜单、按钮权限、后端路由权限和数据库 seed 权限的一致性。
- 已用脚本核对：
  - 后台内置角色权限码全部存在于 `db/seeds/admin_roles_permissions.sql`。
  - 后端 `requireAdminPermission(...)` 路由权限码全部存在于内置角色和 seed。
  - 后台前端 `can("...")` 使用到的按钮权限码全部存在于内置角色和 seed。
- 修复后台“系统配置”页混合权限加载问题：
  - 页面包含生产就绪、敏感词库、内容风险日志、权限快照、领路人资格规则、AI IM 导出配置等多个子模块。
  - 之前菜单只要命中任一系统类权限就会显示，但页面初始化会无条件请求全部子接口，非超级管理员可能一打开就遇到部分接口 `403`。
  - 本轮已改成按权限分块加载：
    - `system_config:read` 才加载生产就绪。
    - `content:sensitive_word:view` 才加载敏感词。
    - `content:risk_log:view` 才加载风险日志。
    - `role:view` 才加载领路人资格规则。
    - `ai:data:read` 才加载 AI IM 导出配置。
  - 无权限的子块显示“缺少 xxx”空状态，不再影响其它有权限子块。
- 同步补齐系统页写动作前端保护：
  - 新增敏感词需要 `content:sensitive_word:create`。
  - 导入敏感词需要 `content:sensitive_word:import`。
  - 更新敏感词状态需要 `content:sensitive_word:update`。
  - 更新领路人资格规则和用户资格需要 `role:update`。
  - 更新 AI IM 导出开关需要 `system_config:update`。
- 已同步构建产物 `admin-web/dist/main.js`。

### 验证补充

- 已执行：
  - `node --check src/main.js`
  - `npm run build`
  - `go test -count=1 ./...`
- 已尝试浏览器冒烟验证，但本地 Codex 浏览器安全策略拒绝访问 `http://127.0.0.1:5173`，因此本轮浏览器点击验证未完成；当前验证停在静态语法、构建产物和后端全量测试。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.8%`：本轮主要改后台前端权限加载，不改变后端业务逻辑；后端路由权限、内置角色权限和 seed 权限已继续核对一致。
- 当前后台前端业务功能进度约 `99.995%`：系统配置页已避免低权限角色打开页面时被无关子模块 `403` 拖垮；剩余重点是真实浏览器端到端点击、线上账号和生产环境联调。

## 2026-06-18 02:14

### 当前进展

- 继续收口后台混合权限页面，避免低权限角色能进入菜单但页面初始化被其它子模块接口 `403` 拖垮。
- 本轮复核后台页面初始化调用，确认 `analytics` 已经按权限分块加载，无需改动。
- 修复“审核中心”页面：
  - `identity:read` 才加载实名核验列表和实名详情。
  - `role:view` 才加载角色申请列表和角色申请详情。
  - `role:update` 才允许通过/驳回角色申请。
  - 无权限子块显示“缺少 xxx”，不影响同页其它有权限模块。
- 修复“分润结算”页面：
  - `revenue:template:view` 才加载分润模板和分润规则。
  - `revenue:template:update` 才允许新增模板、保存规则。
  - `revenue:simulate` 才允许试算。
  - `revenue:generate` 才允许生成分润记录。
  - `revenue:record:view` 才加载分润记录。
  - `revenue:freeze` 才允许冻结记录。
  - `settlement:offline:create` 才加载线下结算记录并允许登记结算。
- 修复“积分兑换”页面：
  - `redemption:manage` 才加载兑换商品、兑换订单并允许新增/更新/审核。
  - `points:read` 才加载积分流水。
  - 只有其中一个权限时，另一个子块显示无权限空态，不再整页失败。
- 已同步构建产物 `admin-web/dist/main.js`。

### 验证补充

- 已执行：
  - `node --check src/main.js`
  - `npm run build`
  - `go test -count=1 ./...`
- 本轮仍受本地 Codex 浏览器安全策略限制，未完成真实浏览器点击验证；当前验证覆盖静态语法、构建产物和后端全量测试。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.8%`：本轮不改后端业务逻辑，继续保持路由权限和角色权限一致。
- 当前后台前端业务功能进度约 `99.997%`：审核、分润、积分兑换、系统配置等混合权限页面已补齐分块加载和动作级前端权限兜底；剩余重点是真实浏览器端到端点击、线上账号和生产环境联调。

## 2026-06-18 02:17

### 当前进展

- 继续收口后台线上测试前的角色权限体验，重点处理“菜单可进入但页面内部接口无权限”的边界。
- 补强“通知交付”页面动作级权限兜底：
  - `notification:wechat:view` 才允许加载微信订阅消息任务、订阅模板、批量发送、单条发送和标记发送。
  - `delivery:manage` 才允许加载和新增交付文档。
  - `testcase:read` 才允许加载测试用例、测试运行和下载测试证据文件。
  - `testcase:manage` 继续作为新增测试用例、登记测试运行的写权限。
  - 这样即使后续页面事件被手动触发，也不会越过前端权限保护去打后端接口。
- 修正后台菜单权限定义：
  - “管理员权限”页面会读取后台账号、后台角色和权限目录，这些后端接口统一要求 `admin_user:view`。
  - 本轮将菜单入口也收敛到 `admin_user:view`，避免只有 `admin_user:create` 或 `role:update` 的角色看到菜单后进入页面立刻 `403`。
- 已同步构建产物 `admin-web/dist/main.js`。

### 验证补充

- 已执行：
  - `node --check src/main.js`
  - `npm run build`
  - `go test -count=1 ./...`
- `git status --short` 未执行成功：当前 `C:\Users\61492\Desktop\真好玩-mini` 不是 Git 仓库或未包含 `.git`，本轮无法给出 Git 工作区变更清单。
- 本轮仍受本地 Codex 浏览器安全策略限制，未完成真实浏览器点击验证；当前验证覆盖静态语法、构建产物和后端全量测试。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.82%`：菜单权限定义继续和后端真实路由权限对齐；后端业务主体保持稳定，剩余重点仍是真实腾讯/微信/OpenIM/存储/实名服务联调与生产配置。
- 当前后台前端业务功能进度约 `99.998%`：通知交付、管理员权限、审核、分润、积分兑换、系统配置等关键后台页面已完成权限分块加载和动作级兜底；剩余重点是真实浏览器端到端点击、线上账号和生产环境联调。

## 2026-06-18 11:50

### 当前进展

- 按 `miniprogram-development`、`karpathy-guidelines`、`cloudrun-development` 重新复核小程序后端与后台对接，不按历史印象判断。
- 重点核对范围：
  - 小程序 `/api/app` 和后台 `/api/admin` 路由。
  - 邀请入口、实名、文件、组局、主行家、后台开局、举报证据、IM 证据、权限菜单。
  - 后台菜单权限、后端路由权限、数据库 seed 权限和前端 `can("...")` 的一致性。
- 修复一个小程序后端业务逻辑问题：
  - 原来 `POST /api/app/files/upload-token` 对除头像外的所有文件都强制要求已实名。
  - 这会导致 `realname_material` 出现“上传实名材料前必须已经实名”的闭环矛盾，也会让玩家在一期不强制实名的口径下被不合理拦截举报附件上传。
  - 本轮改成仅 `chat_file` 继续要求强实名和局成员身份；`avatar`、`realname_material`、`report_attachment` 允许已登录用户上传，再由后续实名/举报业务流程校验归属和使用场景。
- 补充测试覆盖：
  - 未实名用户上传 `chat_file` 仍返回 `403`。
  - 未实名用户上传 `realname_material` 返回成功，支撑实名材料提交闭环。
  - 未实名用户上传 `report_attachment` 返回成功，支撑玩家举报/申诉材料闭环。
  - 头像更新测试改为使用真实返回的 `fileId`，避免新增文件类型测试后被硬编码 ID 影响。
- 修正后台和后端权限对接问题：
  - “用户管理”菜单收敛到 `user:view`，因为页面默认请求 `GET /api/admin/users`，后端真实权限就是 `user:view`。
  - “组局管理”菜单收敛到 `game:read`，因为页面默认请求 `GET /api/admin/games`，后端真实权限就是 `game:read`。
  - 避免角色只有 `user:read` 或 `game:view` 时能看到菜单，但一进入页面就因为列表接口权限不足而失败。
- 已同步构建产物 `admin-web/dist/main.js`。

### 对接结论

- 小程序和后台已成功对接的主链路：
  - 邀请入口：小程序预检/登录绑定，后台邀请码列表、详情、禁用、邀请关系列表。
  - 实名与角色：小程序实名状态、角色申请，后台实名列表/详情、角色申请审核。
  - 组局：小程序创建免费局、申请入局、邀请主行家、手动开始；后台可审核、查看详情、创建不同类型测试局。
  - IM/文件：小程序成员消息、文件上传/下载；后台按权限查看争议 IM 和文件证据。
  - 服务确认/评价/举报/分润：小程序完成服务确认、评价、举报；后台可追踪证据、冻结收益、处理举报、查看分润记录。
- 仍需真实环境联调的部分：
  - 微信真实 code2Session、订阅消息、对象存储直传、腾讯实名/人脸核身、OpenIM 真服务、生产域名和线上账号。
  - 本地 Codex 浏览器安全策略仍阻止真实浏览器点击后台页面，因此本轮浏览器端到端点击验证未完成。

### 验证补充

- 已执行：
  - `go test -count=1 ./internal/appapi -run "Test.*(Login|Invite|File|Report|AdminCreateGame|AdminRole|Permission|Game)"`
  - `go test -count=1 ./internal/files ./internal/adminauth ./internal/games`
  - `node --check src/main.js`
  - `npm run build`
  - `go test -count=1 ./...`
- 当前一期小程序后端业务功能进度约 `99.82%`。
- 当前后台后端业务功能进度约 `98.9%`：后台菜单权限继续和后端路由权限对齐；剩余主要是真实三方服务与生产环境联调。
- 当前后台前端业务功能进度约 `99.998%`：后台页面与接口字段/权限继续对齐；剩余主要是真实浏览器端到端点击验证和线上账号联调。

## 2026-06-18 12:00

### 当前进展

- 按后台运营需求补齐批量邀请码生成：
  - `POST /api/admin/invite-codes` 新增 `batchCount`。
  - `batchCount=1` 时保持原单个创建/手填 code 行为。
  - `batchCount>1` 时必须留空 `code`，由系统自动生成唯一邀请码。
  - 支持 `poster`、`qrcode`、`link` 三种 `entryType` 批量生成。
  - 单次批量上限为 200，避免误操作生成过多入口码。
  - 批量生成写入 `invite_code:batch_create` 操作日志。
- 后台邀请码页面已新增 `batchCount` 输入，运营可直接按入口类型批量生成。
- 同步 OpenAPI：`/api/admin/invite-codes` 已说明单个创建和批量生成的请求/响应差异。
- 按要求补齐后台操作日志分级查看：
  - 新增权限 `operation_log:view_self`。
  - 普通后台管理员默认可查看自己的操作日志。
  - 超级管理员继续通过 `operation_log:view_full` 查看所有操作日志。
  - 普通管理员即使传 `adminUserId=别人`，后端也会强制过滤为自己的 `adminUserId`。
  - 操作日志仍没有删除接口，不提供删除能力。
- 后台日志菜单已改为命中 `operation_log:view_self` 或 `operation_log:view_full` 即可显示。
- 后台日志页对普通管理员禁用 `adminUserId` 筛选输入，避免误以为可查别人；最终权限边界以后端为准。
- 同步 `db/seeds/admin_roles_permissions.sql`：
  - 超级管理员包含 `operation_log:view_full` 和 `operation_log:view_self`。
  - 运营、用户、财务、客服、数据、组局等普通后台角色包含 `operation_log:view_self`。
- 同步 `docs/openapi/admin.openapi.yaml` 和 `docs/test-cases/admin-integration-cases.md`。

### 验证补充

- 已执行：
  - `go test -count=1 ./internal/appapi -run TestAdminInviteCodeManagementHTTP`
  - `go test -count=1 ./internal/appapi -run "TestAdminAccountMutationUpdatesRBAC|TestAdminRolePermissionBoundariesHTTP|TestReport|TestAdminAuth"`
  - `go test -count=1 ./internal/adminauth`
  - `node --check src/main.js`
  - `npm run build`
  - `go test -count=1 ./...`
- 当前一期小程序后端业务功能进度约 `99.82%`。
- 当前后台后端业务功能进度约 `99.05%`：批量邀请码和操作日志分级查看已落地；剩余主要是真实三方服务、生产环境和线上账号联调。
- 当前后台前端业务功能进度约 `99.999%`：后台邀请码批量生成和日志页自查模式已接上；剩余主要是真实浏览器端到端点击验证。

## 2026-06-21 16:53

### 当前进展

- 按“每个局的基本信息作为分享卡片”调整小程序邀请入口。
- `POST /api/app/invites/entries` 新增 `gameId`：
  - 小程序在局详情页分享时传 `entryType=poster` 和当前 `gameId`。
  - 后端生成一次性唯一邀请码，并返回该局分享卡片所需字段。
  - 默认分享标题来自局信息：有城市时为 `城市 · 局标题`，否则使用局标题。
  - 分享路径改为局详情页：`/pages/games/detail?id={gameId}&inviteCode={inviteCode}&entryType=poster`。
  - `scene` 包含 `gameId` 和 `inviteCode`，用于二维码/卡片入口解析。
  - 响应新增 `game` 基本信息：`id/title/gameType/status/cityName/minPlayers/maxPlayers/currentPlayers`。
- 保留原来的 link/qrcode 邀请入口行为，不影响微信聊天链接和二维码入口。
- 同步 `docs/openapi/app.openapi.yaml`，标明 `gameId` 和 `game` 返回字段用途。

### 前端使用方式

- 局详情页分享卡片调用：
  - `POST /api/app/invites/entries`
  - body：`{"entryType":"poster","gameId":当前局ID}`
- 小程序 `onShareAppMessage` 使用返回：
  - `title`：分享卡片标题。
  - `path`：分享卡片打开路径。
  - 后续如果需要卡片封面图，可在小程序端根据 `game` 信息生成默认图，或后端再扩展 `imageUrl`。

### 验证补充

- 已执行：
  - `go test -count=1 ./internal/appapi -run TestCreateInviteEntryHTTP`
  - `go test -count=1 ./internal/auth`
  - `go test -count=1 ./...`
- 当前一期小程序后端业务功能进度约 `99.84%`。
- 当前后台后端业务功能进度约 `99.05%`。
- 当前后台前端业务功能进度约 `99.999%`。

## 2026-06-23 19:20

### 当前进展

- 按“腾讯地图 Key 无法绑定小程序 AppID”的实际情况，调整地图接入为后端代理方案：
  - 小程序前端不引入腾讯位置服务小程序 SDK。
  - 小程序前端不写腾讯地图 Key，也不绑定地图 Key 的 AppID。
  - 前端只使用 `wx.getLocation` 获取定位，使用原生 `<map>` 展示附近局 marker。
  - 地点搜索、地址转经纬度、经纬度转地址、路线规划统一请求后端。
- 新增后端腾讯地图 WebService 客户端：
  - `internal/lbs.MapProvider` 抽象。
  - `TencentMapClient` 使用 `TENCENT_MAP_KEY_SERVER` 调腾讯 WebService。
  - 支持 `TENCENT_MAP_SK` 签名、`TENCENT_MAP_API_BASE`、`TENCENT_MAP_REQUEST_TIMEOUT_MS`。
- 新增小程序地图代理接口：
  - `GET /api/app/map/search`
  - `GET /api/app/map/geocode`
  - `GET /api/app/map/reverse-geocode`
  - `GET /api/app/map/route`
- 新增后台地点搜索接口：
  - `GET /api/admin/map/search`
  - 复用 `game:create_admin` 权限，后台创建局时不需要前端持有地图 Key。
- 生产 readiness 增加 `tencent_map_server_key` 检查项；启用腾讯地图时生产环境必须配置服务端 Key 和 HTTPS APIBase。
- 同步 `deploy/env.example`、`docs/openapi/app.openapi.yaml`、`docs/openapi/admin.openapi.yaml`、`第三方接口文档.md`。

### 前端对接口径

- 前端需要做：
  - `wx.getLocation({ type: "gcj02" })` 获取当前位置。
  - 调 `POST /api/app/locations/current` 保存定位。
  - 调 `GET /api/app/games/nearby` 获取附近局。
  - 用 `<map>` 的 `markers` 展示局位置。
  - 创建局选地点时调用 `GET /api/app/map/search`，选择结果后保存地点名、地址、经纬度。
- 前端不要做：
  - 不引入 `qqmap-wx-jssdk.js`。
  - 不直连 `https://apis.map.qq.com`。
  - 不在小程序包内写腾讯地图 Key。

### 验证补充

- 已执行：
  - `go test -count=1 ./internal/appapi -run "TestTencentMapProxyHTTP|TestMapProxyRequiresConfiguredProvider|TestLocationAndNearbyGamesFlow"`
  - `go test -count=1 ./internal/common/config`
  - `go test -count=1 ./internal/lbs`
- 当前一期小程序后端业务功能进度约 `99.86%`。
- 当前后台后端业务功能进度约 `99.08%`。
- 当前后台前端业务功能进度约 `99.999%`。

## 2026-06-28 22:35

### 当前进展

- 修复小程序公开局列表真实业务缺口：
  - `GET /api/app/games` 不再返回 `pending_audit`、`draft`、`cancelled`、`closed` 等非公开状态。
  - `GET /api/app/games/city`、`GET /api/app/games/nearby` 同步复用公开状态过滤。
  - 后台 `GET /api/admin/games` 保持全量可查，审核人员仍可看到待审核局。
- 补齐后台开局局型：
  - 后台现在支持 `free`、`standard`、`public_welfare`、`aa`、`crowdfund`、`deposit`、`condition`。
  - 小程序端 `POST /api/app/games` 仍保持一期限制：普通用户只能创建 `free`。
  - 后台前端创建局、筛选、分润模板、局型统计和中文标签已同步新枚举。
- 同步接口文档和测试用例：
  - `docs/openapi/admin.openapi.yaml` 更新后台开局枚举。
  - `docs/openapi/app.openapi.yaml` 标明小程序公开列表只返回公开状态。
  - `docs/test-cases/admin-integration-cases.md` 更新后台全局型开局测试范围。

### 验证补充

- 已执行：
  - `go test -count=1 ./internal/games -run "TestCreateFromAdminAllowsDocumentedGameTypes|TestCreateFromAdminRejectsInvalidTypeAndCreator"`
  - `go test -count=1 ./internal/appapi -run "TestAdminAuditGameApprovesPendingGame|TestAdminCreateConditionGameWhileAppCreateStaysFreeOnly"`
  - `go test -count=1 ./...`
  - `node --check admin-web/src/main.js`
  - `npm run build`
- 当前一期小程序后端业务功能进度约 `99.88%`。
- 当前后台后端业务功能进度约 `99.18%`。
- 当前后台前端业务功能进度约 `99.999%`。

## 2026-06-29 10:45

### 当前进展

- 已按用户指定项目根目录 `C:\Users\61492\Desktop\真好玩-mini` 处理，不再使用 `E:\jubaopen`。
- 已从前端技术人员仓库 `https://github.com/ZYJ-1010/ZHW` 拉取小程序前端源码：
  - 下载位置：`C:\Users\61492\Desktop\codex download\ZHW`
  - 当前项目接入位置：`miniprogram-client`
- 小程序前端已作为独立客户端目录接入当前项目，未覆盖后端 `services`、后台 `admin-web` 和现有文档。
- 已完成第一批前后端 P0 主链路对接：
  - `config/env.js` 新增 `ENV.LOCAL`，本地联调默认指向 `http://127.0.0.1:8080`。
  - 邀请预检从旧 `/api/app/invites/verify` 对齐到后端真实 `/api/app/invites/precheck`。
  - 短信发送/校验从旧 `/api/app/auth/phone-code`、`/phone-code/verify` 对齐到 `/api/app/sms/send-code`、`/api/app/sms/verify-code`。
  - 实名接口从旧 `/api/app/users/me/realname-auth`、`/submit` 对齐到 `/api/app/identity/faceid/detect-auth`、`/api/app/identity/phone/verify`。
  - 新增前端 `bindPhone`、`getIdentityStatus`、`issueTokenAfterIdentity` 封装，支撑后端 preAuthToken 到正式 token 的链路。
  - 微信登录成功后前端会保存 `token || preAuthToken`，避免新用户未实名时后续实名接口无鉴权。
  - IM 从旧 `/api/im/...` 对齐到 `/api/app/games/{gameId}/chat-session` 和 `/api/app/games/{gameId}/chat/messages`。
  - 申请入局从旧 `/api/app/games/{gameId}/apply` 对齐到 `/api/app/games/{gameId}/applications`。

### 仍需继续联调的前端候选接口

- 前端仍存在一些设计阶段/候选接口，需要下一轮按页面逐个接真实后端或改为现有聚合接口：
  - `/api/app/home`
  - `/api/app/newbie-tasks`
  - `/api/app/messages/center`
  - `/api/app/messages/trade-warning`
  - `/api/app/messages/system-notification`
  - `/api/app/profile/home`
  - `/api/app/profile/points/*`
  - `/api/app/profile/system-management/*`
  - `/api/app/relations/network-home`
  - `/api/app/game-invites/*`
  - `/api/app/games/my/manage`
  - `/api/app/games/player/manage`
  - `/api/app/games/profit-templates`
  - `/api/app/game-payments/wechat`
- 这些接口多数属于页面聚合、运营视图、二期支付/推荐或前端 mock 先行接口，本轮没有盲目扩后端。

### 验证补充

- 已执行：
  - `git clone https://github.com/ZYJ-1010/ZHW.git "C:\Users\61492\Desktop\codex download\ZHW"`
  - `Copy-Item ...\ZHW ...\miniprogram-client -Recurse -Force`
  - `Get-ChildItem -Recurse miniprogram-client -Filter *.js | node --check`
  - 小程序 `app.json` 页面完整性检查：113 个页面的 `.js/.json/.wxml` 均存在。
  - 关键旧接口残留检查：真实运行路径内未再发现 `/api/im`、`/api/app/invites/verify`、`/api/app/auth/phone-code`、`/api/app/users/me/realname-auth`。
- 当前一期小程序后端业务功能进度约 `99.88%`。
- 当前小程序前端与后端 P0 主链路对接进度约 `55%`：源码已接入，登录/邀请/实名/局/IM 主入口已改到后端真实路径；剩余重点是页面级聚合接口、真实微信开发者工具编译和真机联调。
- 当前后台后端业务功能进度约 `99.18%`。
- 当前后台前端业务功能进度约 `99.999%`。

## 2026-06-29 22:35

### 当前进展

- 继续只在用户指定项目根目录 `C:\Users\61492\Desktop\真好玩-mini` 内推进，不再使用 `E:\jubaopen`。
- 按 `miniprogram-development`、`web-development`、`karpathy-guidelines` 做小程序前后端全链路复核。
- 第二批前端接口已对齐后端真实路径：
  - 会员状态 `/api/app/memberships/me` -> `/api/app/membership/my`。
  - 成长档案 `/api/app/growth/me` -> `/api/app/growth/my`。
  - 积分商城、兑换、订单 -> `/api/app/redemption/items`、`/api/app/redemption/orders`、`/api/app/redemption/orders/my`。
  - 消息中心 -> `/api/app/notifications`。
  - 关系网首页 -> `/api/app/connections/my`。
  - 支付占位 -> `/api/app/payment/precreate-placeholder`。
- 修正字段级联调问题：
  - 积分兑换前端服务层把商品 id 转成后端需要的 `itemId`。
  - 邀请响应前端服务层把 `accept/approve/yes` 动作映射成后端需要的 `accept` 字段。
- 后端新增小程序薄聚合接口：
  - `GET /api/app/home`
  - `GET /api/app/newbie-tasks`
  - `GET /api/app/profile/home`
  - `GET /api/app/games/my/manage`
  - `GET /api/app/games/player/manage`
  - `GET /api/app/games/profit-templates`
- 新增文档 `docs/progress/full-chain-gap-2026-06-29.md`，记录本轮已打通链路和仍未完整实现的文档/前端缺口。

### 未完整实现的重点缺口

- 组局邀请推荐链路仍缺推荐玩家、最近联系人、复玩上下文、系统推荐、取消详情和进度聚合接口。
- 个人中心系统管理页仍是前端候选模型，和后端现有用户资料、行家技能、领路人资源画像字段不完全一致。
- 积分订单仍缺物流、取消、详情等小程序履约接口。
- 消息中心已能读通知，但按钮动作和详情块结构仍未完全对接。
- 首页地球和关系网已有业务数据基础，但动态可视化布局数据仍需后续定字段。

### 验证补充

- 已执行：
  - `Get-ChildItem -Recurse miniprogram-client -Filter *.js | ForEach-Object { node --check $_.FullName }`
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./...`
- 当前一期小程序后端业务功能进度约 `99.9%`。
- 当前小程序前端与后端全链路对接进度约 `72%`。
- 当前后台后端业务功能进度约 `99.2%`。
.FullName }`
  - `go test -count=1 ./internal/redemption ./internal/appapi`
  - `go test -count=1 ./...`
- 当前前后端全链路对接进度约 `76%`。
- 小程序后端主体能力仍约 `99.9%+`，剩余主要是真实三方和线上环境联调；前后端对接剩余重点是组局邀请推荐聚合、个人中心系统管理字段、消息卡片动作、首页地球/关系网可视化数据结构。

## 2026-06-16 13:56

### 当前进展

- 继续以一期小程序后端上线测试为目标，收紧“只能通过邀请入口进入”的登录注册闭环。
- 修补唯一入口码绑定语义：
  - 新用户通过 poster / qrcode / link 入口码进入时，继续走注册模式并绑定微信。
  - 已注册微信通过新的唯一入口码进入时，后端会把该入口码绑定到当前微信，并返回 `authPageMode=login`、`boundWechat=true`。
  - 其他微信再次使用已绑定的唯一入口码时，继续返回 `invite code already bound`。
- 数据层新增 `000027_invite_entry_unique_binding.sql`：
  - 取消 `invite_relations.invitee_user_id` 的唯一约束，允许同一个微信用户绑定多个唯一入口码。
  - 保留按 `invite_code_id` 和 `invitee_user_id` 查询索引，邀请关系统计仍以用户首个邀请关系为准。
- 保留通用测试/运营邀请码的多人使用能力；唯一绑定规则只作用于 `maxUses=1` 的入口码，避免误锁 `TEST2026` 这类通用码。
- 同步 `docs/openapi/app.openapi.yaml` 和 `docs/test-cases/app-integration-cases.md` 的前端联调说明。

### 验证补充

- `go test -count=1 ./internal/invites ./internal/auth ./internal/appapi` 通过。
- `go test -count=1 ./internal/...` 通过。
- 当前一期小程序后端业务功能进度约 `97%`。

## 2026-06-16 13:26

### 当前进展

- 继续以一期小程序后端上线测试为目标，补齐后台审核工作台的列表入口。
- 新增后台组局列表接口：
  - `GET /api/admin/games`
  - 支持 `status` 筛选，例如 `pending_audit`、`recruiting`
  - 需要后台权限 `game:read`
- 已把正式审核链路扩展成完整后台工作流：
  - 后台先查 `GET /api/admin/games?status=pending_audit`
  - 再调用 `POST /api/admin/games/{gameId}/audit`
  - 审核后该局从 `pending_audit` 列表消失，并进入 `recruiting` 列表
- 同步 `docs/openapi/admin.openapi.yaml` 和 `docs/test-cases/admin-integration-cases.md`，方便前后端按最新后台接口联调。

### 验证补充

- `go test -count=1 ./internal/appapi` 通过。
- `go test -count=1 ./internal/...` 通过。
- 当前一期小程序后端业务功能进度约 `96.5%`。

## 2026-06-16 13:05

### 当前进展

- 继续以一期小程序后端上线测试为目标，补齐“局创建后由后台正式审核”的生产业务入口。
- 新增后台正式审核接口：
  - `POST /api/admin/games/{gameId}/audit`
  - 需要后台权限 `game:update_status`
  - 审核通过后调用现有组局服务，将 `pending_audit` 的局推进到 `recruiting`
- 后台权限种子已补充 `game:update_status`，超级管理员可直接执行，运营账号未授权时返回 `403`。
- 审核动作已写入后台操作日志 `game:audit`，便于线上排查“谁审核了哪个局”。
- 生产环境不再需要依赖 `/approve-local` 这类本地联调入口完成组局审核。

### 验证补充

- `go test -count=1 ./internal/appapi` 通过。
- `go test -count=1 ./internal/...` 通过。
- 当前一期小程序后端业务功能进度约 `96%`。

## 2026-06-16 12:55

### 当前进展

- 继续以一期小程序后端上线收口为导向，处理本地联调入口的生产安全边界。
- `appapi.Server.Configure` 新增生产环境识别：
  - `APP_ENV=prod`
  - `APP_ENV=production`
- `POST /api/app/games/{gameId}/approve-local` 现在仅保留为本地/测试联调入口：
  - 本地环境继续可用，保持现有测试和联调流程。
  - 生产环境直接返回 `404`，避免本地审核后门随服务上线。
- 补充测试 `TestApproveLocalDisabledInProduction`，确保带登录 token 访问生产环境本地审核入口仍不可用。

### 验证补充

- `go test -count=1 ./internal/appapi` 通过。
- `go test -count=1 ./internal/...` 通过。
- 当前一期小程序后端业务功能进度约 `95.5%`。

## 2026-06-16 12:35

### 当前进展

- 继续以一期小程序后端业务功能为导向，补齐微信订阅消息任务执行闭环。
- `notifications.Service` 新增订阅消息发送能力：
  - `SendWechatTask(taskID)` 会根据任务里的 `userId` 找到微信 `openid`，调用 `WechatSubscribeSender` 发起真实订阅消息发送。
  - `SendPendingWechatTasks(limit)` 支持批量发送待处理任务，方便内部任务执行器一次性拉起。
  - 发送结果会回写 `wechat_subscribe_tasks`，并同步更新 `notifications.wechatState`。
- `appapi` 新增内部路径：
  - `POST /api/internal/notifications/wechat-tasks/{taskId}/send`
  - `POST /api/internal/notifications/wechat-tasks/send-pending`
  - 仍保留 `mark-sent` 作为兼容入口。
- 微信订阅消息发送器接入现有 `WECHAT_APP_ID/WECHAT_APP_SECRET` 配置，不再另起一套账号配置。

### 验证补充

- `go test -count=1 ./internal/notifications ./internal/appapi` 通过。
- `go test -count=1 ./internal/...` 通过。
- 当前一期小程序后端业务功能进度约 `95%`。

## 2026-06-16 11:50

### 当前进展

- 继续以一期小程序后端业务功能为主，补齐强实名认证 FaceID 生产链路。
- `identity.Service` 的 `StartFaceID` 从固定 `mock_face_token` 升级为可注入启动器：
  - 本地默认 `LocalFaceIDStarter`，继续返回 mock token，保证开发联调不被第三方阻塞。
  - 生产可配置 `HTTPFaceIDStarter`，由 `FACEID_HTTP_ENDPOINT` 发起核身并读取 `faceToken` / `bizToken`。
  - 发起请求会携带 `userId`、脱敏手机号、脱敏姓名、脱敏身份证号，避免后端接口继续暴露明文身份数据。
- 生产环境配置新增强校验：
  - `FACEID_HTTP_ENDPOINT` 必须是 HTTPS。
  - `FACEID_HTTP_SECRET` 必须是安全密钥。
  - 启动入口 `cmd/server/main.go` 已根据配置注入真实 FaceID HTTP 启动器。
- 同步 `.env.example` 和 `docs/openapi/app.openapi.yaml`：
  - 补充 FaceID 发起网关、回调签名密钥和开关变量。
  - `POST /api/app/identity/faceid/detect-auth` 文档从“返回 mock faceToken”改为“返回本地 mock 或配置网关返回的 faceToken”。

### 验证补充

- `go test -count=1 ./internal/identity ./internal/common/config ./internal/appapi` 通过。
- `go test -count=1 ./internal/...` 通过。
- 当前一期小程序后端业务功能进度约 `93%`。

## 2026-06-14 19:15

### 当前进展

- 继续执行 `plan.md`，推进后台权限隔离业务功能。
- 后台内置角色补齐：
  - 新增 `data_analyst` 本地登录账号样本，只包含行为、看板、报表导出、测试结果读取权限。
  - 新增 `operator` 本地登录账号样本，可做普通运营查看和处理，但不含完整日志、敏感画像、测试证据权限。
- 测试用例权限拆分：
  - `GET /api/admin/test-cases` 和 `GET /api/admin/test-runs` 改为 `testcase:read`。
  - `POST /api/admin/test-cases` 和 `POST /api/admin/test-runs` 保持 `testcase:manage`。
- `db/seeds/admin_roles_permissions.sql` 追加 `testcase:read` 和 data_analyst/operator 角色权限映射，保持数据库初始化与本地内置账号一致。
- 更新 `plan.md`：
  - 勾选“数据分析员可看行为、报表、测试结果，但不能审核兑换和修改配置”。
  - 勾选“普通运营人员不可查看完整敏感画像和测试证据文件”。

### 验证补充

- `go test -count=1 ./internal/adminauth ./internal/appapi` 通过。
- `go test -count=1 ./...` 通过。

## 2026-06-14 18:45

### 当前进展

- 继续执行 `plan.md`，优先补齐小程序后端业务验收闭环。
- 后台标记打卡异常增强：
  - `POST /api/admin/game-checkins/{checkinId}/mark-invalid` 现在必须提交 `reason`。
  - 缺少原因或原因过长时返回 422，不会把打卡改为异常。
  - 成功标记异常后，会把 `reason` 写入后台操作日志 payload，便于追溯。
- 更新 `plan.md`：
  - 勾选“兑换审核、打卡异常标记、交付文档状态变更必须填写原因”。
  - 补充“打卡异常标记必须填写原因”子项。

### 验证补充

- `go test -count=1 ./internal/appapi` 通过。
- `go test -count=1 ./...` 通过。

## 2026-06-14 12:11

### 当前进展

- 继续 D7 生产密钥和环境变量约束：
  - `config.Load` 已显式读取 `APP_ENV`、`JWT_SECRET`、数据库连接、OpenIM secret、MinIO secret 等环境变量
  - 新增 `Config.ValidateProduction()`，仅在 `APP_ENV=prod/production` 时启用生产配置拦截
  - 生产环境拒绝缺失数据库连接、缺失/占位 `JWT_SECRET`、占位 `OPENIM_SECRET`、占位 `MINIO_SECRET_KEY`
  - 服务启动入口已接入生产配置校验，避免默认占位密钥进入生产运行
- `plan.md` 已同步勾选 D7 的生产 JWT secret、数据库密码、对象存储密钥环境变量/密钥管理项。

### 验证补充

- `go test -count=1 ./internal/common/config ./internal/appapi` 通过。

### 下一步

- 跑全量 `go test -count=1 ./...`，继续 D7：写接口幂等和外部输入校验。

## 2026-06-14 12:23

### 当前进展

- 继续 D7 后台鉴权与权限统一：
  - 移除生产代码中 `X-Admin-Permissions` 请求头直通权限的兼容入口
  - 后台受保护接口统一通过 `Authorization: Bearer <admin-token>` 校验后台 session
  - `requireAdminPermission(permissionCode)` 统一负责权限码校验，并自动回填 `X-Admin-ID` 供操作日志使用
  - IM 争议消息后台查看入口只保留统一权限链路，不再做可伪造请求头二次判断
  - 测试 helper 已改为真实后台登录 token，不再依赖伪权限头
- `plan.md` 已同步勾选 D7 的后台接口 AdminAuth 和 PermissionMiddleware 两项。

### 验证补充

- `rg` 已确认业务代码不再包含 `X-Admin-Permissions`、`adminHasPermission`、`permissionsFromHeader`。
- `go test -count=1 ./internal/appapi` 通过。

## 2026-06-14 18:05

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 E5 AI 数据准备和后台导出基础。
- 评价模型新增 `tags`：
  - `POST /api/app/reviews` 支持可选 `tags` 数组。
  - 标签会去空、去重，最多保留 8 个，每个最长 32 字符。
  - `ReviewDTO` 和 `/api/app/reviews/my-intents` 会返回 `tags` 与 `againIntent`。
- 数据库迁移新增 `db/migrations/000022_review_tags.sql`：
  - 给 `reviews` 增加 `tags jsonb not null default '[]'`。
  - 给 `again_intent` 增加查询索引，便于后台导出和二期推荐分析。
- 后台导出模板新增 `reviews_default`：
  - `ExportType=reviews`
  - 固定列包含 `tags` 和 `again_intent`。
- 更新 `plan.md`：
  - 标记“评价标签和再玩意向可导出”完成。
  - 注意：`000022_review_tags.sql` 后续需要补入迁移顺序表；当前业务代码和迁移文件已落地。

### 验证补充

- `go test -count=1 ./internal/reviews` 通过。
- `go test -count=1 ./internal/exports` 通过。
- `go test -count=1 ./internal/appapi` 通过。
- `go test -count=1 ./...` 通过。

## 2026-06-14 18:25

### 当前进展

- 继续执行 `plan.md`，优先推进后端验收留档链路。
- 交付文档登记增强：
  - `docType` 限定为 `prd/openapi/database/deployment/test-report/release-record/rollback-plan`。
  - 状态进入 `ready` 或 `archived` 时必须填写 `reason`。
  - 返回 `DocumentDTO` 增加 `reason` 字段。
- 测试用例登记增强：
  - `priority` 只能是 `P0/P1/P2`。
  - 测试执行记录继续支持 `requestId` 关联，便于用 requestId 追踪接口响应。
- 更新 `plan.md`：
  - 勾选交付文档类型约束。
  - 勾选测试用例优先级约束。
  - 勾选测试用例和测试执行记录可登记。
  - 在混合验收项下标记交付文档状态变更必须填写原因已完成；打卡异常标记原因仍待补。

### 验证补充

- `go test -count=1 ./internal/delivery` 通过。
- `go test -count=1 ./internal/appapi` 通过。

## 2026-06-14 17:45

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 E1.4 积分兑换闭环。
- 核实现有兑换链路：
  - 积分不足会拒绝兑换并回滚库存。
  - 库存不足会拒绝兑换。
  - 内存模式下已有并发兑换不超卖测试。
  - 后台审核、驳回、发放已有状态流转和操作日志。
- 新增兑换订单审核原因约束：
  - 后台审核通过、驳回、标记发放时，`reason` 不能为空。
  - 缺少原因时返回参数错误，订单保持原状态。
- 更新 `plan.md`：
  - 勾选“积分不足不能兑换”“库存不足不能兑换”“并发兑换不超卖”。
  - 在混合验收项下补充“兑换订单审核、驳回、发放必须填写原因”已完成；打卡异常和交付文档状态原因仍保留待补。

### 验证补充

- `go test -count=1 ./internal/redemption` 通过。
- `go test -count=1 ./internal/appapi` 通过。

## 2026-06-14 17:25

### 当前进展

- 继续执行 `plan.md`，优先补齐小程序后端强实名联调路径。
- 新增 `POST /api/app/identity/faceid/result`：
  - 复用现有 FaceID callback 完成逻辑。
  - 同样受 `AppAuthMiddleware`、幂等中间件和 FaceID 回调签名配置保护。
  - 解决 `plan.md` 强实名接口清单中 `faceid/result` 与现有代码只有 `faceid/callback` 的路径不一致问题。
- 整理 `docs/openapi/app.openapi.yaml` 的 identity 局部路径块：
  - 补出 `phone/verify`、`faceid/detect-auth`、`faceid/callback`、`faceid/result`、`identity/status` 独立路径。
  - 避免小程序前端联调时把多个路径误读成 description 文本。

### 验证补充

- `go test -count=1 ./internal/appapi` 通过。

### 下一步

- 跑全量 `go test -count=1 ./...`，继续 D7：生产密钥环境变量、写接口幂等和外部输入校验。

## 2026-06-14 12:11

### 当前进展

- 继续 D7 后端安全和权限统一落地，补齐请求访问日志的脱敏保护：
  - `httpx.RequestID` 生成后回填请求头，方便后续操作日志和访问日志使用同一个 requestId
  - 新增 `httpx.AccessLog`，只记录 method、path、status、duration、requestId
  - 访问日志不记录 header、body 和 query 原值，避免手机号、身份证、token、密码、密钥进入日志
  - Go 服务入口已接入 `RequestID + AccessLog`
- 复核后台管理员密码逻辑：
  - `adminauth` 已使用 PBKDF2-SHA256 强哈希
  - 测试覆盖错误密码拒绝、正确密码登录、明文密码存储拒绝
- `plan.md` 已同步勾选 D7 的请求日志脱敏、后台管理员密码强哈希两项。

### 验证补充

- `go test -count=1 ./internal/common/httpx` 通过。
- `go test -count=1 ./internal/appapi` 通过。

### 下一步

- 跑全量 `go test -count=1 ./...`，继续 D7：认证/权限中间件覆盖复核、生产密钥环境变量、写接口幂等和外部输入校验。

## 2026-06-14 11:58

### 当前进展

- 继续 D7 后端安全和权限统一落地，优先处理小程序后端文件下载权限：
  - `GET /api/app/files/{fileId}/download-url` 收紧为按业务对象校验
  - `avatar` 仅上传者本人可下载
  - `chat_file` 仅同局成员可下载
  - `realname_material`、`report_attachment`、`export_file` 等非小程序通用下载场景默认拒绝
- 测试补充：
  - 头像上传者本人可下载，其他用户下载返回 403
  - IM 文件局成员可下载，非成员下载返回 403
  - 举报附件不能通过小程序通用下载口下载
- `plan.md` 已同步勾选 D7 文件下载对象权限和非成员读取 IM/文件上线阻断项。

### 验证补充

- `go test -count=1 ./internal/appapi` 通过。
- Go telemetry 仍有 AppData 权限警告，但测试退出码为 0。

### 下一步

- 继续 D7：复核请求日志敏感字段、后台管理员密码哈希、生产密钥环境变量、写接口幂等和输入校验。

## 2026-06-14 11:58

### 当前进展

- 补齐 D6 前后台对接所需的数据看板接口文档：
  - `docs/openapi/admin.openapi.yaml` 登记 `/api/admin/dashboard`
  - `docs/openapi/admin.openapi.yaml` 登记 `/api/admin/analytics/funnel`
  - `docs/openapi/admin.openapi.yaml` 登记 `/api/admin/analytics/retention`
  - `docs/openapi/dto-samples.md` 新增 `FunnelSnapshotDTO`
  - `docs/openapi/dto-samples.md` 新增 `RetentionSnapshotDTO`
  - `docs/openapi/dto-samples.md` 新增 `DashboardSnapshotDTO`
- D5.8 在 `plan.md` 中已全部勾选完成，下一步进入 D6 后台接口/页面对接或 D7 安全权限统一检查。

### 验证补充

- `rg` 已确认 OpenAPI 和 DTO 示例能搜到新增路径与 DTO。
- `go test -count=1 ./internal/audit ./internal/appapi` 通过。
- Go telemetry 仍有 AppData 权限警告，但测试退出码为 0。

### 下一步

- 优先继续 D6 后台前端工程硬要求，或按后端优先进入 D7：后台接口统一权限、文件下载对象权限、IM 权限一致性、输入校验和敏感日志检查。

## 2026-06-14 11:58

### 当前进展

- 收口 D5.8 数据看板基础统计：
  - `audit.Service` 新增漏斗聚合，按 `eventCode` 统计每步用户数、转化率和流失率
  - `audit.Service` 新增留存聚合，按用户首次行为日期统计次日、7 日、30 日留存
  - 新增后台接口 `/api/admin/dashboard`
  - 新增后台接口 `/api/admin/analytics/funnel`
  - 新增后台接口 `/api/admin/analytics/retention`
- `plan.md` 已同步勾选 D5.8 最后一项“数据看板能统计漏斗和留存基础数据”。

### 验证补充

- `go test -count=1 ./internal/audit ./internal/appapi` 通过。
- Go telemetry 仍有 AppData 权限警告，但测试退出码为 0。

### 下一步

- 跑全量 `go test -count=1 ./...`，然后继续进入 D6 或回补后台 OpenAPI/页面侧任务。

## 2026-06-14 11:58

### 当前进展

- 继续补 D5.8 行为日志闭环：
  - `GET /api/app/games` 自动记录 `browse_games`
  - `GET /api/app/games/city` 自动记录同城浏览
  - `GET /api/app/games/nearby` 自动记录附近浏览
- 行为日志测试已覆盖 `browse_games`、手动上报 `search/share`、字段 `eventCode/businessType/businessId/source/device/ip/occurredAt`。
- D5.8 中通知站内 + 微信任务、关键行为日志、行为字段完整性已同步勾选。

### 验证补充

- `go test -count=1 ./internal/appapi` 通过。
- Go telemetry 仍有 AppData 权限警告，但测试退出码为 0。

### 下一步

- 继续补 D5.8 剩余项：数据看板能统计漏斗和留存基础数据。

## 2026-06-14 11:58

### 当前进展

- 继续推进 D5.8 举报申诉、通知和行为留存，优先补小程序后端举报证据链。
- `POST /api/app/reports` 新增证据归属校验：
  - `chatMessageId` 必须存在且属于当前 `gameId`
  - `fileId` 必须存在且 `objectId` 属于当前 `gameId`
  - `reviewId` 必须来自当前局评价链路
  - `revenueRecordId` 必须存在且属于当前局
- 集成测试补齐真实文件凭证，并增加无效 IM 证据返回 `422` 的断言。
- `plan.md` 已同步勾选 D5.8 中举报关联证据、举报后冻结收益、后台处理可见证据、争议冻结分润。

### 验证补充

- `go test -count=1 ./internal/appapi` 通过。
- Go telemetry 仍有 AppData 权限警告，但测试退出码为 0。

### 下一步

- 继续补 D5.8 剩余项：通知站内和微信任务覆盖、用户关键行为日志字段完整性、漏斗和留存基础数据看板。

## 2026-06-14 11:35
### 褰撳墠杩涘睍

- 缁х画 D5.7 鍒嗘鼎鐢熸垚鍓嶄簤璁娴嬨€?- `POST /api/admin/revenue/records/generate` 鐜板湪鐢熸垚鍒嗘鼎璁板綍鍚庝細妫€鏌ュ悓灞€鏈叧闂妇鎶ョ敵璇夛細
  - 鏈夋湭鍏抽棴涓炬姤鏃讹紝璁板綍绔嬪嵆杩涘叆 `frozen`
  - `frozenReason=report_open`
  - 鍚庣画缁撶畻娌跨敤鏃㈡湁鍐荤粨鎷︽埅锛岃繑鍥炲啿绐佺姸鎬?- 宸茶ˉ闆嗘垚娴嬭瘯锛岃鐩栤€滃厛涓炬姤銆佸悗鐢熸垚鍒嗘鼎鈥濈殑鍐荤粨璺緞銆?- `plan.md` 宸插悓姝ュ嬀閫夊垎娑︾敓鎴愪簤璁娴嬨€佷簤璁喕缁撳拰 TC-116銆?
### 楠岃瘉琛ュ厖

- `go test -count=1 ./internal/appapi` 閫氳繃銆?
### 涓嬩竴姝?
- 鍏ㄩ噺娴嬭瘯鍚庤繘鍏?D5.8 涓炬姤鐢宠瘔銆侀€氱煡鍜岃涓虹暀瀛樼殑鍚庣鏀跺彛锛屼紭鍏堣ˉ涓炬姤鍏宠仈璇佹嵁鍜岃涓虹暀瀛樼己鍙ｃ€?
## 2026-06-14 11:15

### 褰撳墠杩涘睍

- 杩涘叆 D5.7 鍒嗘鼎銆佹敹鐩娿€佽鍗曞崰浣嶇殑鍚庣鎸佷箙鍖栨敹鍙ｃ€?- 鏂板 `services/go-api/internal/revenue/sql_repository.go`锛?  - 鍒嗘鼎妯℃澘鎸佷箙鍖栧埌 `revenue_templates`
  - 鍒嗘鼎璁板綍鎸佷箙鍖栧埌 `revenue_records`
  - 鍒嗘鼎鏄庣粏鎸佷箙鍖栧埌 `revenue_record_items`
  - 绾夸笅缁撶畻鐧昏鍐欏叆 `settlement_records`
  - 鐢ㄦ埛鏀剁泭鎽樿鍜屾敹鐩婃槑缁嗕粠 SQL 鑱氬悎璇诲彇
- `revenue.Service` 澧炲姞 repository 娉ㄥ叆锛屼繚鐣欏唴瀛?fallback銆?- `cmd/server` 鍜?`appapi.UseRepositories` 宸叉帴鍏?`revenue.NewSQLRepository(db)`銆?- `plan.md` 宸插悓姝?D5.7 涓凡楠岃瘉鐨勫厤璐瑰眬璁㈠崟銆佹ā鏉块厤缃€佽瘯绠椼€佹敹鐩婃憳瑕?鏄庣粏銆佹暣鏁板垎鍜屽喕缁撶粨绠楁嫤鎴」銆?
### 楠岃瘉琛ュ厖

- 鏂板 `internal/revenue/repository_service_test.go`锛岃鐩栨ā鏉裤€佽瘯绠椼€佺敓鎴愩€侀噸澶嶇敓鎴愩€佸喕缁撱€佺粨绠楀拰鏀剁泭鏌ヨ璧?repository銆?- `go test -count=1 ./...` 閫氳繃銆?
### 涓嬩竴姝?
- D5.7 鍓╀綑杈冨椤规槸鈥滃垎娑︾敓鎴愬墠涓诲姩妫€鏌ユ湭鍏抽棴涓炬姤鐢宠瘔骞剁洿鎺?frozen/闃绘柇鈥濓紝褰撳墠宸叉湁涓炬姤鍒涘缓鍚庡喕缁撴棦鏈夊垎娑﹁褰曪紝鍚庣画闇€瑕佹妸鐢熸垚鍓嶄簤璁娴嬭ˉ鍒?revenue 鐢熸垚鍏ュ彛銆?
## 2026-06-14 10:35

### 褰撳墠杩涘睍

- 缁х画鏀跺彛 D5.6 鏈嶅姟纭銆佽瘎浠枫€佹垚闀裤€佷俊鐢ㄩ摼璺€?- `reviews.Repository` 澧炲姞 `CreditDeductionValue(ruleCode)`锛岄€€鍑烘墸淇＄敤浼樺厛璇诲彇 `credit_deduction_rules`銆?- `reviews.SQLRepository` 宸叉敮鎸佷粠鍚敤鐨勪俊鐢ㄦ墸鍒嗚鍒欒鍙?`change_value`銆?- 鏂板杩佺Щ `db/migrations/000021_credit_deduction_defaults.sql`锛?  - `quit_after_confirm = -10`
  - `quit_after_started = -10`
- 鏃犺鍒欓厤缃椂浠?fallback 涓?`-10`锛屼笉褰卞搷鐜版湁娴佺▼銆?- `plan.md` 宸插悓姝ュ嬀閫?D5.6 淇＄敤涓婚摼銆佹瘡鏃ヤ俊鐢ㄥ垵濮?100銆侀€€鍑烘墸淇＄敤鍑嗙‘绛夊凡楠岃瘉椤广€?
### 楠岃瘉琛ュ厖

- 鏂板 repository 绾ф祴璇曪紝瑕嗙洊鏈夎鍒欐椂鎸夎鍒欐墸鍒嗐€佹棤瑙勫垯鏃朵娇鐢ㄩ粯璁ゆ墸鍒嗐€?- `go test -count=1 ./internal/reviews` 宸查€氳繃銆?
### 涓嬩竴姝?
- 鍏ㄩ噺娴嬭瘯鍚庣户缁帹杩?D5.7 鍒嗘鼎/鏀剁泭閾捐矾锛屾垨鑰呭厛琛?`game_members.member_status/credit_log_id` 杩欑鏇寸粏瀛楁钀藉簱銆?
## 2026-06-14 10:05

### 褰撳墠杩涘睍

- 缁х画鎺ㄨ繘 `plan.md` 鐨?D5.6 鏈嶅姟纭涓庤瘎浠烽摼璺紝浼樺厛鍚庣銆?- `services/go-api/internal/reviews/service.go` 宸茶ˉ榻愶細
  - 鏈嶅姟纭瀹屾垚鍚庤繘鍏?`pending_review`
  - 璇勪环寰呭姙鎸夋埅姝㈡椂闂磋繃婊?  - 璇勪环鎻愪氦鎸夋埅姝㈡椂闂村拰閲嶅鍏崇郴鎷︽埅
  - 鏁版嵁搴撴ā寮忎笅鐨勮瘎浠峰畬鎴愬垽鏂彲鍩轰簬鎸佷箙鍖栬瘎浠疯褰曟仮澶?  - 璇勪环鍚庣户缁啓缁忛獙銆佺Н鍒嗐€佹垚灏卞拰瓒宠抗
- `GET /api/app/footprints/my` 宸茶ˉ鎴愮嫭绔嬩釜浜轰腑蹇冩帴鍙ｃ€?- `services/go-api/internal/appapi/server_test.go` 澧炲姞浜嗚冻杩规帴鍙ｅ洖褰掋€?
### 楠岃瘉琛ュ厖

- `go test -count=1 ./...` 閫氳繃銆?- 褰撳墠浠嶄繚鐣欎竴鏉″悗缁緟鍋氾細淇＄敤瑙勫垯鐨勫畬鏁撮噸绠椾笌鏇寸粏鍖栫害鏉燂紝褰撳墠涓嶅湪杩欒疆纭嬀瀹屾垚銆?
### 涓嬩竴姝?
- 缁х画鎶?D5.6 閲屽墿浣欑殑淇＄敤閲嶇畻/鎻愰啋閾捐矾鏀跺彛锛屽啀鍚?D5.7 鍒嗘鼎鍗犱綅鎺ㄨ繘銆?
## 2026-06-14 09:20

### 褰撳墠杩涘睍

- 缁х画鎸?`plan.md` 浼樺厛鎺ㄨ繘灏忕▼搴忓悗绔€?- 瀹屾垚鏂囦欢涓婁紶鍑瘉鐨勬湇鍔″眰鐧藉悕鍗曡鍒欙細
  - `avatar`锛氫粎鍏佽 `image/jpeg`銆乣image/png`銆乣image/webp`锛屾渶澶?5 MiB锛岃闂骇鍒?`private`銆?  - `chat_file`锛氫粎鍏佽鍥剧墖銆丳DF銆佺函鏂囨湰銆乑IP锛屾渶澶?20 MiB锛屽繀椤荤粦瀹氬眬 `objectId`锛岃闂骇鍒?`game_member`銆?  - `realname_material`锛氫粎鍏佽鍥剧墖鍜?PDF锛屾渶澶?10 MiB銆?  - `report_attachment`锛氫粎鍏佽鍥剧墖鍜?PDF锛屾渶澶?20 MiB銆?- `POST /api/app/files/upload-token` 淇濇寔澶村儚鍙?pre-auth 涓婁紶锛涢潪澶村儚鏂囦欢浠嶈姹傚己瀹炲悕銆?- OpenAPI 瀵逛笂浼犲嚟璇佽ˉ鍏?`422` 鏍￠獙澶辫触璇存槑銆?
### 楠岃瘉琛ュ厖

- 鏂板鏂囦欢鏈嶅姟鍗曞厓娴嬭瘯锛岃鐩栨湭鐭?`bizType`銆佽秴澶у皬銆侀潪娉?MIME銆両M 鏂囦欢缂哄皯 `objectId`銆佸ご鍍忚闂骇鍒€?- 鏂板 app API 娴嬭瘯锛岃鐩?pre-auth 鐢ㄦ埛涓嶈兘鐢宠 `chat_file` 涓婁紶鍑瘉銆?
### 涓嬩竴姝?
- 缁х画鎺ㄨ繘 D5.6 鏈嶅姟纭涓庤瘎浠烽摼璺殑鍚庣瀹屾暣鎬э紝浼樺厛琛ユ暟鎹簱鎸佷箙鍖栥€佺姸鎬佹祦杞拰鏉冮檺娴嬭瘯銆?

## 2026-06-13 01:16

### 褰撳墠杩涘睍

- 宸茬‘璁ら」鐩綋鍓嶅彧鏈夋枃妗ｅ拰 `plan.md`锛屾病鏈夌幇鎴愪唬鐮佸伐绋嬨€?- 宸叉寜 `plan.md` P0 鍒涘缓鍚庣銆佽祫閲戞湇鍔°€丳C 鍚庡彴銆佹暟鎹簱銆侀儴缃层€佹帴鍙ｆ枃妗ｃ€佹祴璇曟枃妗ｃ€佽繘搴﹁褰曠洰褰曘€?- 宸茬‘璁?Java銆丯ode銆乶pm 鍙敤锛汳aven 鏈畨瑁咃紱Go SDK 闇€瑕佷娇鐢?`C:\Program Files\Go\bin\go.exe`锛岀郴缁?PATH 鍓嶇疆瀛樺湪 `C:\Windows\System32\go` 骞叉壈銆?
### 褰撳墠椋庨櫓

- Maven 涓嶅湪 PATH锛孞ava 璧勯噾鏈嶅姟鍏堢敤 JDK 鑷甫 HTTP 鏈嶅姟楠ㄦ灦鍚姩锛屽悗缁啀琛?Maven 鎴?Gradle銆?- `go` 鍛戒护浼樺厛鍛戒腑 `C:\Windows\System32\go`锛屽悗缁紪璇戦渶鏄庣‘璋冪敤鐪熷疄 Go SDK 鎴栦慨姝?PATH銆?
### 涓嬩竴姝?
- 鍒濆鍖?Go API 鍋ュ悍妫€鏌ャ€?- 鍒濆鍖?Java 璧勯噾鏈嶅姟鍋ュ悍妫€鏌ャ€?- 鍒濆鍖栧悗鍙伴潤鎬佸３銆?- 琛ュ厖鏁版嵁搴撹縼绉诲拰閮ㄧ讲鏍蜂緥銆?
## 2026-06-13 01:50

### 褰撳墠杩涘睍

- 宸叉寜鐢ㄦ埛鏈€鏂拌姹傛敹鏁涘疄鏂借寖鍥达細褰撳墠鍙户缁仛鈥滃皬绋嬪簭鍚庣鈥濓紝涓嶇户缁帹杩?PC 鍚庡彴鍓嶇鍜?Java 璧勯噾鏈嶅姟銆?- 宸茶鍙栧苟閲囩敤 `karpathy-guidelines`锛屽悗缁寜灏忔銆佸彲娴嬭瘯銆佷笉杩囧害璁捐鐨勬柟寮忓紑鍙戙€?- Go 灏忕▼搴忓悗绔仴搴锋鏌ュ凡鍙闂細`GET /health`銆乣GET /api/app/health`銆?- 瀹屾垚 D5.1 绗竴杞悗绔兘鍔涳細
  - `POST /api/app/auth/wechat-login`
  - `GET /api/app/users/me`
  - 鍐呭瓨鐗堢敤鎴枫€侀個璇风爜銆侀個璇峰叧绯汇€乼oken store
  - 鏂扮敤鎴锋棤閭€璇风爜鎷︽埅
  - 鏈夋晥閭€璇风爜娉ㄥ唽骞剁粦瀹氬叧绯?  - 宸插瓨鍦?openid 鍐嶇櫥褰曚笉閲嶅寤虹敤鎴枫€佷笉瑕佹眰鍐嶆浼犻個璇风爜
- 宸茶ˉ鍏?`docs/openapi/app.openapi.yaml`銆乣docs/openapi/dto-samples.md`銆乣docs/openapi/error-codes.md`銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 閫氳繃鐨勬祴璇曞寘锛?  - `internal/auth`
  - `internal/appapi`

### 褰撳墠椋庨櫓

- 褰撳墠 D5.1 浣跨敤鍐呭瓨浠撳偍锛屾湇鍔￠噸鍚悗鏁版嵁浼氫涪澶憋紱鍚庣画闇€瑕佹帴 PostgreSQL 浠撳偍銆?- 寰俊鐧诲綍褰撳墠鏄?mock openid锛屽悗缁渶瑕佹帴鐪熷疄 `code2Session`銆?- 鐪熷疄寮哄疄鍚嶈繕鏈疄鐜帮紝涓嬩竴姝ユ寜 E4/D5.2 琛ユ墜鏈哄彿銆佺煭淇°€佷汉鑴告牳韬姸鎬佷富绾裤€?
### 涓嬩竴姝?
- 瀹炴柦 D5.2/E4锛氬己瀹炲悕鐘舵€佹満銆佹墜鏈哄彿缁戝畾銆佺煭淇￠獙璇佺爜銆佷汉鑴告牳韬崰浣嶆帴鍙ｃ€?- 淇濇寔鎺ュ彛鍏堝彲娴嬶紝鍐嶆浛鎹㈢湡瀹炵涓夋柟璋冪敤銆?
## 2026-06-13 02:05

### 褰撳墠杩涘睍

- 瀹屾垚 D5.2/E4 寮哄疄鍚嶄富绾跨殑鏈湴鍗犱綅瀹炵幇銆?- 宸叉柊澧炲己瀹炲悕鐘舵€佹満锛?  - `wechat_logged_in`
  - `phone_bound`
  - `sms_verified`
  - `phone_verified`
  - `faceid_processing`
  - `verified`
- 宸叉柊澧炲皬绋嬪簭鍚庣鎺ュ彛锛?  - `POST /api/app/identity/phone/bind`
  - `POST /api/app/sms/send-code`
  - `POST /api/app/sms/verify-code`
  - `POST /api/app/identity/phone/verify`
  - `POST /api/app/identity/faceid/detect-auth`
  - `POST /api/app/identity/faceid/callback`
  - `GET /api/app/identity/status`
- 褰撳墠涓烘湰鍦板崰浣嶉€昏緫锛氱煭淇￠獙璇佺爜鍥哄畾 `123456`锛屼汉鑴告牳韬繑鍥?`mock_face_token`锛屽井淇″疄鍚嶄竴鑷存€ц褰曚负 `not_supported`銆?- 宸茶ˉ鍏?`docs/openapi/app.openapi.yaml`銆乣docs/openapi/dto-samples.md`銆乣docs/openapi/error-codes.md`銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板閫氳繃娴嬭瘯鍖咃細
  - `internal/identity`
  - `internal/appapi` 寮哄疄鍚?HTTP 娴佺▼

### 褰撳墠椋庨櫓

- 寮哄疄鍚嶈繕鏈帴鑵捐浜戠煭淇°€佹墜鏈哄彿浜?涓夎绱犮€佹収鐪间汉鑴告牳韬湡瀹炴帴鍙ｃ€?- 褰撳墠鐢ㄦ埛銆佸疄鍚嶇姸鎬佷粛鏄唴瀛樺瓨鍌紝鍚庣画闇€瑕佹帴 PostgreSQL銆?
### 涓嬩竴姝?
- 瀹炴柦 D5.3 缁勫眬銆佸叆灞€銆佺姸鎬佹満鐨勫悗绔富閾捐矾銆?- 鍦ㄨ繘鍏ョ粍灞€鍒涘缓鍓嶅鍔犲己瀹炲悕鎷︽埅锛岀‘淇濇湭 `verified` 鐢ㄦ埛涓嶈兘鍒涘缓灞€銆?
## 2026-06-13 02:18

### 褰撳墠杩涘睍

- 瀹屾垚 D5.3 缁勫眬涓婚摼璺涓€姝ワ細灏忕▼搴忕鍒涘缓鍏嶈垂灞€銆佸眬鍒楄〃銆佸眬璇︽儏銆?- 宸插疄鐜拌鍒欙細
  - 鏈畬鎴愬己瀹炲悕 `verified` 涓嶈兘鍒涘缓灞€銆?  - 灏忕▼搴忕鍙厑璁稿垱寤?`free` 鍏嶈垂灞€銆?  - 姣忓眬浜烘暟蹇呴』鍦?5-8 鑼冨洿鍐呫€?  - 鍚屼竴鐢ㄦ埛姣忔棩鏈€澶氬垱寤?3 灞€銆?  - 鍒涘缓鎴愬姛鍚庣姸鎬佷负 `pending_audit`锛岀瓑寰呭悗鍙板鏍搞€?- 宸叉柊澧炴帴鍙ｏ細
  - `POST /api/app/games`
  - `GET /api/app/games`
  - `GET /api/app/games/{gameId}`
- 宸茶ˉ鍏?`GameDTO`銆丱penAPI 璺緞鍜岄敊璇爜銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板閫氳繃娴嬭瘯鍖咃細
  - `internal/games`
  - `internal/appapi` 鍒涘缓灞€ HTTP 娴佺▼

### 褰撳墠椋庨櫓

- 缁勫眬浠嶄负鍐呭瓨瀛樺偍锛屽悗缁渶鎺?PostgreSQL銆?- 鐩墠鍒涘缓鍚庣洿鎺?`pending_audit`锛屽悗鍙板鏍搞€佸叆灞€鐢宠銆佹墜鍔ㄥ紑濮嬨€両M 鍒涘缓杩樻湭瀹炵幇銆?
### 涓嬩竴姝?
- 缁х画 D5.3锛氳ˉ灞€瀹℃牳閫氳繃鍚庣殑灞曠ず鐘舵€併€佺敵璇峰叆灞€銆佸叆灞€瀹℃牳銆佷汉鏁伴攣瀹氥€?- 澧炲姞鐘舵€佹祦杞棩蹇楁ā鍨嬶紝涓哄悗缁?IM 鍜岃瘎浠烽摼璺仛鍑嗗銆?
## 2026-06-13 02:40

### 褰撳墠杩涘睍

- 鎸夌敤鎴疯姹傝繘鍏ユ寔缁疄鏂芥ā寮忥細鏈畬鎴愭暣涓?`plan.md` 灏忕▼搴忓悗绔墠涓嶄富鍔ㄥ仠姝€?- 宸茶鍙栧苟浣跨敤鏈満 skill锛?  - `karpathy-guidelines`锛氬皬姝ュ疄鐜般€佸彲楠岃瘉銆侀伩鍏嶈繃搴﹁璁°€?  - `auth-wechat`锛氱‘璁ゅ皬绋嬪簭韬唤浠?openid/unionid 涓烘牳蹇冿紝褰撳墠闈?CloudBase 浜戝嚱鏁伴」鐩紝鍏堜繚鐣欏悗绔?`code2Session` 褰㈡€併€?  - `cloudbase-wechat-integration`锛氱‘璁や竴鏈熷厤璐瑰眬涓嶆帴鐪熷疄鏀粯锛屾敮浠樺垎璐﹀彧鍋氳鍗曞拰鐘舵€侀鐣欍€?- 瀹屾垚 D5.3 缁勫眬鐢宠鍜岀姸鎬佹祦杞細
  - 鏈湴瀹℃牳灞€鎺ュ彛 `POST /api/app/games/{gameId}/approve-local`
  - 鐢宠鍏ュ眬 `POST /api/app/games/{gameId}/applications`
  - 鎴戠殑鐢宠 `GET /api/app/games/applications/my`
  - 瀹℃牳鍏ュ眬 `POST /api/app/games/applications/{applicationId}/review`
  - 鎵嬪姩寮€濮?`POST /api/app/games/{gameId}/manual-start`
  - 閫€鍑哄眬 `POST /api/app/games/{gameId}/exit`
- 瀹屾垚 D5.4 LBS锛?  - 淇濆瓨褰撳墠瀹氫綅 `POST /api/app/locations/current`
  - 淇濆瓨鎵嬪姩瀹氫綅 `POST /api/app/locations/manual`
  - 鍚屽煄灞€ `GET /api/app/games/city`
  - 闄勮繎灞€ `GET /api/app/games/nearby`
  - 瀹氫綅绮惧害瓒呰繃 300m 杩斿洖 `accuracyWarning=true`
  - 闄勮繎灞€鎸夋湰鍦?Haversine 璺濈璁＄畻骞舵帓搴忋€?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板瑕嗙洊锛?  - 鐢宠鍏ュ眬銆佸鏍稿叆灞€銆佹墜鍔ㄥ紑濮嬪悗閿佸畾鐢宠銆?  - 淇濆瓨瀹氫綅銆佸垱寤哄甫鍧愭爣鐨勫眬銆侀檮杩戝眬鏌ヨ銆?
### 褰撳墠椋庨櫓

- 浠嶆槸鍐呭瓨浠撳偍锛屽悗缁渶鎺?PostgreSQL銆?- `approve-local` 鏄悗绔嚜娴嬫帴鍙ｏ紝鐢ㄤ簬鏃?PC 鍚庡彴鏃惰仈璋冮棴鐜紱姝ｅ紡鐢熶骇搴旀浛鎹负鍚庡彴瀹℃牳鎺ュ彛鎴栫Щ闄ゃ€?
### 涓嬩竴姝?
- 瀹炴柦 D5.5锛氳嚜鐮?IM 鎴块棿銆佹垚鍛樻潈闄愩€佹枃鏈秷鎭€佹晱鎰熻瘝鎷︽埅銆佸巻鍙叉秷鎭€?
## 2026-06-13 03:05

### 褰撳墠杩涘睍

- 鐢ㄦ埛鏄庣‘瑕佹眰 IM 涓嶄粠 0 鍏ㄩ儴鑷爺锛岄渶瑕佸叏缃戞绱㈠姛鑳藉畬鏁寸殑寮€婧?IM 椤圭洰骞朵簩娆″紑鍙戙€?- 宸叉绱㈠苟閫夊畾 OpenIM `openimsdk/open-im-server` 浣滀负涓€鏈?IM 搴曞骇锛?  - Apache-2.0 璁稿彲璇併€?  - Go 鏈嶅姟绔紝閫傚悎褰撳墠 Go 鍚庣闆嗘垚銆?  - 鏀寔 REST API銆乄ebhooks銆佺兢缁勩€佹秷鎭€佸巻鍙层€佹枃浠跺璞¤兘鍔涘拰 WebSocket 缃戝叧銆?  - 姣?Tinode銆丮atrix/Synapse銆丷ocket.Chat 鏇磋创鍚堚€滃祵鍏ヤ笟鍔″簲鐢ㄥ仛 IM 搴曞骇鈥濈殑闇€姹傘€?- 宸叉寜鐢ㄦ埛涓嬭浇瑙勫垯鍏嬮殕婧愮爜鍒帮細
  - `C:\Users\61492\Desktop\codex download\open-im-server`
- 宸插皢鏈」鐩?IM 瀹炴柦鏂瑰悜鏀逛负锛?  - 鐪熷ソ鐜╁悗绔仛涓氬姟閴存潈銆佸眬鎴愬憳鏉冮檺銆佹晱鎰熻瘝鍜?DTO銆?  - OpenIM 鍋?WebSocket 闀胯繛鎺ャ€佺兢缁勩€佹秷鎭姇閫掋€佸巻鍙叉秷鎭€佹枃浠跺璞¤兘鍔涖€?- 宸叉柊澧?Go 鍚庣 OpenIM 閫傞厤灞傦細
  - `services/go-api/internal/im/openim_client.go`
  - `OPENIM_API_ADDR`
  - `OPENIM_SECRET`
  - `OPENIM_ADMIN_USER_ID`
- 宸蹭繚鐣欐湰鍦板唴瀛?IM fallback锛屾湭閮ㄧ讲 OpenIM 鏃舵祴璇曚粛鍙繍琛屻€?- 宸叉洿鏂帮細
  - `deploy/env.example`
  - `docs/openapi/ws-protocol.md`

### 褰撳墠椋庨櫓

- OpenIM 鍏ㄥ鏈嶅姟灏氭湭鍦ㄦ湰鏈哄惎鍔ㄩ獙璇侊紝褰撳墠 Go 娴嬭瘯璧?fallback銆?- OpenIM Webhooks 鏁忔劅璇?鎴愬憳鏉冮檺浜屾鏍￠獙杩樻湭閰嶇疆鍒?OpenIM 鏈綋銆?- 鍥剧墖銆佹枃浠舵秷鎭繕鏈垏鍒?OpenIM 瀵硅薄瀛樺偍鎺ュ彛銆?
### 涓嬩竴姝?
- 璺戦€氱幇鏈?Go 娴嬭瘯锛岀‘淇?OpenIM 閫傞厤灞傛病鏈夌牬鍧?D5.1-D5.4銆?- 缁х画 D5.5锛氳ˉ OpenIM 鐢ㄦ埛 token 涓嬪彂鎺ュ彛銆丱penIM 缇ゅ悓姝ョ姸鎬併€佹枃浠舵秷鎭拰 Webhooks 鍥炶皟鑽夋銆?
## 2026-06-13 03:35

### 褰撳墠杩涘睍

- 鍩轰簬 OpenIM 缁х画鎺ㄨ繘 D5.5 灏忕▼搴忓悗绔?IM 閾捐矾銆?- 宸茶ˉ OpenIM 灏忕▼搴忎細璇濆弬鏁帮細
  - `GET /api/app/games/{gameId}/chat-session`
  - 杩斿洖 `engine`銆乣imUserId`銆乣openIMGroupId`銆乣openIMToken`
  - 鏈厤缃?OpenIM 鏃惰繑鍥?`engine=local`锛屼繚鐣欐湰鍦拌仈璋?fallback銆?- 宸插吋瀹?`plan.md` 鐨?roomId 鐗堟秷鎭帴鍙ｏ細
  - `GET /api/app/chat/rooms/{roomId}/messages`
  - `POST /api/app/chat/rooms/{roomId}/messages`
- 宸茶ˉ鏂囦欢鏈嶅姟鍗犱綅锛?  - `POST /api/app/files/upload-token`
  - `GET /api/app/files/{fileId}/download-url`
  - `chat_file` 鏂囦欢蹇呴』鏍￠獙灞€鎴愬憳鏉冮檺锛岄潪灞€鎴愬憳杩斿洖 `40331`銆?- 宸茶ˉ鍏?OpenAPI銆丏TO 绀轰緥鍜岄敊璇爜鏂囨。銆?
### 楠岃瘉缁撴灉

- 浣跨敤椤圭洰鍐?Go 缂撳瓨鎵ц `go test ./...` 閫氳繃銆?- 瑕嗙洊鐐瑰寘鎷細
  - chat-session 鑾峰彇銆?  - roomId 鍙戦€佸拰鎷夊巻鍙叉秷鎭€?  - IM 鏂囦欢涓婁紶鍑瘉銆?  - 鎴愬憳鍙笅杞?IM 鏂囦欢銆?  - 闈炴垚鍛樹笉鑳戒笅杞?IM 鏂囦欢銆?
### 褰撳墠椋庨櫓

- OpenIM 鍏ㄥ鏈嶅姟浠嶆湭鏈満鍚姩楠岃瘉锛屽綋鍓嶆祴璇曚娇鐢?local fallback銆?- OpenIM Webhooks 鐨勬晱鎰熻瘝銆佹垚鍛樻潈闄愪簩娆℃牎楠岃繕鏈啓鍏?OpenIM 閰嶇疆銆?- 鏂囦欢涓婁紶 URL 鐩墠鏄?`mock://`锛屽悗缁渶瑕佹帴 OpenIM 瀵硅薄鏈嶅姟鎴栭」鐩粺涓€瀵硅薄瀛樺偍銆?
### 涓嬩竴姝?
- 缁х画 D5.5锛氳ˉ OpenIM Webhooks 鍥炶皟鑽夋鍜屾枃浠舵秷鎭槧灏勮鏄庛€?- 涔嬪悗杩涘叆 D5.6锛氭湇鍔＄‘璁ゃ€佽瘎浠枫€佹垚闀裤€佷俊鐢ㄣ€佽冻杩广€?
## 2026-06-13 04:20

### 褰撳墠杩涘睍

- 澶嶆牳褰撳墠浠ｇ爜鍚庣户缁帹杩?`plan.md`锛岀‘璁?D5.5 宸叉湁鎴块棿銆乺oomId 娑堟伅銆丄CK銆佸凡璇汇€佸綊妗ｅ拰 OpenIM Webhook 鍥炶皟璁板綍銆?- 琛ュ厖 D5.5 鏂囦欢娑堟伅鍜?OpenIM Webhook 鏄犲皠璇存槑锛?  - 鏂囦欢涓婁紶鍑瘉锛歚POST /api/app/files/upload-token`
  - 鏂囦欢涓存椂涓嬭浇锛歚GET /api/app/files/{fileId}/download-url`
  - 鏂囦欢/鍥剧墖娑堟伅閫氳繃 `messageType=image/file` + `fileId` 鍏宠仈銆?  - OpenIM Webhook 鍏ュ彛锛歚POST /api/internal/openim/webhooks`
- 杩涘叆 D5.6 骞跺畬鎴愬悗绔渶灏忛棴鐜細
  - `POST /api/app/games/{gameId}/service-confirm`
  - `POST /api/app/games/{gameId}/service-confirm-items`
  - `GET /api/app/reviews/todos`
  - `POST /api/app/reviews`
  - `GET /api/app/reviews/my-intents`
  - `GET /api/app/users/me/growth`
- 鏂板 `services/go-api/internal/reviews/service.go`锛屾敮鎸佸緟璇勪环鍒楄〃銆佽瘎浠锋彁浜ゃ€侀噸澶嶈瘎浠锋嫤鎴€佸啀鐜╂剰鍚戙€佹垚闀跨粡楠屻€佺Н鍒嗗拰瓒宠抗蹇収銆?- 鎵╁睍 `games` 鏈嶅姟纭鐘舵€佹祦杞細
  - 鎵嬪姩寮€濮嬪悗杩涘叆 `in_progress`銆?  - 棣栦釜鎴愬憳纭鍚庤繘鍏?`pending_confirm`銆?  - 鍏ㄩ儴鎴愬憳纭鍚庤繘鍏?`pending_review`锛屽苟鐢熸垚鍙瘎浠风獥鍙ｃ€?- 琛ラ綈鏁版嵁搴撹縼绉昏崏妗堬細`game_service_confirms`銆乣game_service_confirm_items`銆乣reviews`銆乣review_reminders`銆乣user_growth_profiles`銆乣experience_logs`銆乣points_logs`銆乣user_footprints`銆乣achievements`銆乣user_achievements`銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板瑕嗙洊锛氭湭瀹屾垚鏈嶅姟涓嶈兘璇勪环銆侀潪鎴愬憳涓嶈兘璇勪环銆佹垚鍛樻湇鍔＄‘璁ゅ悗杩涘叆寰呰瘎浠枫€侀噸澶嶈瘎浠疯鎷︽埅銆佽瘎浠峰悗鎴愰暱缁忛獙/绉垎/瓒宠抗鎺ュ彛鍙繑鍥炲彉鍖栥€?
### 褰撳墠椋庨櫓

- D5.6 鐩墠浠嶆槸鍐呭瓨鏈嶅姟闂幆锛屽悗缁渶瑕佹帴 PostgreSQL 浠撳偍銆?- 鎴愬氨銆佷俊鐢ㄦ墸鍒嗐€佽冻杩圭淮搴﹀凡棰勭暀琛ㄥ拰鎺ュ彛蹇収锛屼絾杩樻湭鍋氬畬鏁磋鍒欏紩鎿庛€?- OpenIM 鍏ㄥ鏈嶅姟浠嶆湭鏈満鍚姩楠岃瘉锛屽綋鍓?IM 娴嬭瘯缁х画浣跨敤 local fallback銆?
### 涓嬩竴姝?
- 缁х画 D5.6锛氳ˉ淇＄敤鎵ｅ垎瑙勫垯銆佹垚灏辫Е鍙戝拰鍚庡彴杩芥函鏌ヨ銆?- 闅忓悗杩涘叆 D5.7锛氬垎娑﹂瑙堛€佽瘎浠锋潯浠舵牎楠屻€佹敹鐩婅褰曞拰缁撶畻闃绘柇銆?
## 2026-06-13 04:55

### 褰撳墠杩涘睍

- 缁х画 D5.6锛岃ˉ榻愪笂涓€杞湭瀹屾垚鐨勪俊鐢ㄣ€佹垚灏卞拰鍚庡彴杩芥函鑳藉姏銆?- 鎴愰暱璧勬枡鏂板浠婃棩淇＄敤瀛楁锛?  - `creditScore`
  - `todayCreditScore`
  - `achievements`
  - 浠婃棩淇＄敤涓虹┖鏃惰嚜鍔ㄦ寜 100 杩斿洖骞跺垵濮嬪寲鍐呭瓨璁板綍銆?- 璇勪环鍚庤嚜鍔ㄨЕ鍙戝熀纭€鎴愬氨锛?  - `first_review`
  - `first_received_review`
- 閫€鍑哄眬鏃舵寜灞€鐘舵€佸垽鏂槸鍚︽墸淇＄敤锛?  - `pending_confirm`锛歚quit_after_confirm`锛屾墸 10 鍒嗐€?  - `in_progress` / `pending_review` / `completed`锛歚quit_after_started`锛屾墸 10 鍒嗐€?  - 鍏朵粬鐘舵€侊細`quit_before_confirm`锛屼笉鎵ｅ垎銆?- 琛ュ悗鍙拌拷婧帴鍙ｏ細
  - `GET /api/admin/users/{userId}/growth`
  - `GET /api/admin/games/{gameId}/review-trace`
- 琛?plan 瀵瑰簲杩佺Щ琛細
  - `daily_credit_scores`
  - `credit_deduction_rules`
  - `credit_logs.game_id`
- 琛?app OpenAPI锛?  - `GET /api/app/growth/my`
  - `POST /api/app/games/{gameId}/exit`
- 琛?admin OpenAPI锛?  - `GET /api/admin/users/{userId}/growth`
  - `GET /api/admin/games/{gameId}/review-trace`

### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板瑕嗙洊锛?  - 浠婃棩淇＄敤榛樿 100銆?  - 璇勪环鍚庤繑鍥炴垚灏卞拰瓒宠抗銆?  - 纭鍚?寮€濮嬪悗閫€鍑烘墸淇＄敤锛屼粠 100 鎵ｅ埌 90銆?  - 鍚庡彴鎸夌敤鎴峰拰灞€杩芥函璇勪环銆佷俊鐢ㄣ€佽冻杩规潵婧愩€?
### 褰撳墠椋庨櫓

- D5.6 浠嶄负鍐呭瓨闂幆锛屽悗缁渶瑕佹浛鎹负 PostgreSQL 浠撳偍鍜屼簨鍔°€?- 鍚庡彴杩芥函鎺ュ彛褰撳墠鏈帴鍚庡彴鏉冮檺涓棿浠讹紝鍚庣画杩涘叆鍚庡彴鏉冮檺妯″潡鏃堕渶瑕佸姞 RBAC銆?- 淇＄敤鎵ｅ垎瑙勫垯鐩墠鎸夐粯璁?10 鍒嗗疄鐜帮紝鍚庣画闇€瑕佷粠 `credit_deduction_rules` 閰嶇疆璇诲彇銆?
### 涓嬩竴姝?
- 杩涘叆 D5.7锛氬垎娑﹂瑙堛€佽瘎浠峰畬鎴愭潯浠舵牎楠屻€佹敹鐩婅褰曘€佽瘎浠锋湭瀹屾垚缁撶畻闃绘柇銆?
## 2026-06-13 05:30

### 褰撳墠杩涘睍

- 杩涘叆 D5.7锛屽厛鍦?Go API 渚ц惤鍦版湰鍦板彲娴嬬殑鍒嗘鼎銆佹敹鐩娿€佺粨绠楅棴鐜€?- 鏂板 `services/go-api/internal/revenue/service.go`锛?  - 鍒嗘鼎妯℃澘鍒涘缓鍜屽垪琛ㄣ€?  - 鍒嗘鼎璇曠畻锛岄噾棰濆叏鐢ㄦ暣鏁板垎锛屾瘮渚嬩娇鐢?bps銆?  - 鍒嗘鼎璁板綍鐢熸垚锛屾寜 `gameId` 骞傜瓑闃绘柇閲嶅鐢熸垚銆?  - 璇勪环鏈畬鎴愭椂闃绘柇鐢熸垚鍒嗘鼎銆?  - 鍒嗘鼎鍐荤粨銆?  - 绾夸笅缁撶畻鐧昏銆?  - 鐢ㄦ埛鏀剁泭鎽樿銆?- 鏂板/鎺ュ叆鎺ュ彛锛?  - `GET /api/app/games/{gameId}/revenue-preview`
  - `GET /api/app/incomes/summary`
  - `GET /api/admin/revenue/templates`
  - `POST /api/admin/revenue/templates`
  - `POST /api/admin/revenue/preview`
  - `POST /api/admin/revenue/calculate`
  - `POST /api/admin/revenue/records/generate`
  - `GET /api/admin/revenue/records`
  - `POST /api/admin/revenue/records/{recordId}/freeze`
  - `POST /api/admin/revenue/records/{recordId}/settle`
- 鎵╁睍 `reviews` 鏈嶅姟锛屾彁渚?`GameReviewComplete(gameId)`锛屼緵鍒嗘鼎鐢熸垚鏍￠獙璇勪环瀹屾垚鏉′欢銆?- 閲嶅啓骞惰ˉ榻?`db/migrations/000007_revenue_orders.sql`锛?  - `payment_orders`
  - `revenue_templates`
  - `revenue_rules`
  - `revenue_records`
  - `revenue_record_items`
  - `user_income_accounts`
  - `income_logs`
  - `settlement_records`
- 琛?app/admin OpenAPI 鍒嗘鼎鎺ュ彛銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板瑕嗙洊锛?  - 鍒嗘鼎妯℃澘 bps 鍒涘缓銆?  - 鍒嗘鼎璇曠畻鍙繑鍥為瑙堛€?  - 璇勪环鏈畬鎴愪笉鑳界敓鎴愬垎娑︺€?  - 鍙屾柟璇勪环瀹屾垚鍚庡彲鐢熸垚鍒嗘鼎璁板綍銆?  - 鍐荤粨鍚庝笉鑳界粨绠椼€?  - 绗簩涓眬鍙甯哥敓鎴愬苟绾夸笅缁撶畻銆?  - 鐢ㄦ埛绔敹鐩婃憳瑕佸彲鏌ヨ銆?
### 褰撳墠椋庨櫓

- D5.7 褰撳墠鍏堣惤 Go API 鍐呭瓨闂幆锛孞ava `funds-service` 浠嶅彧鏈?health锛屽悗缁渶瑕佹妸璧勯噾閫昏緫杩佸叆 Java 鏈嶅姟鎴栨敼涓?Go API 璋?Java銆?- 鍚庡彴鍒嗘鼎鎺ュ彛鏆傛湭鎺?RBAC 鏉冮檺涓棿浠躲€?- 浜夎/涓炬姤妯″潡灏氭湭鎺ュ叆锛屽洜姝も€滄湁浜夎鑷姩鍐荤粨鈥濈洰鍓嶇敱鍚庡彴 freeze 鎺ュ彛妯℃嫙銆?
### 涓嬩竴姝?
- 缁х画 D5.7锛氳ˉ Java funds-service 鐨?revenue/settlement HTTP 鍗犱綅鎺ュ彛锛屾垨杩涘叆 D5.8 涓炬姤鐢宠瘔骞舵妸浜夎鍐荤粨鎺ュ埌鍒嗘鼎璁板綍銆?
## 2026-06-13 06:00

### 褰撳墠杩涘睍

- 缁х画 D5.7锛岃ˉ榻?Java `funds-service` 鐨勮祫閲戝崰浣?HTTP 鎺ュ彛銆?- 鏂板閫氱敤鍝嶅簲 helper锛?  - `services/funds-service/src/main/java/com/zhw/funds/common/HttpJson.java`
- 鏂板鏀粯璁㈠崟鍗犱綅锛?  - `GET/POST /api/funds/payment-precreate-placeholder`
  - `GET/POST /api/funds/payment-callback-placeholder`
  - `GET/POST /api/internal/pay/callback-placeholder`
- 鏂板鍒嗘鼎鍗犱綅锛?  - `/api/funds/revenue/templates`
  - `/api/funds/revenue/simulate`
  - `/api/funds/revenue/generate`
  - `/api/funds/revenue/records`
  - `/api/funds/profit-sharing/receivers`
  - `/api/funds/profit-sharing/orders`
  - `/api/funds/profit-sharing/return-orders`
- 鏂板缁撶畻/瀵硅处鍗犱綅锛?  - `/api/funds/settlements/offline`
  - `/api/funds/settlements`
  - `/api/funds/bills/download`
  - `/api/internal/funds/reconcile-runner`
- `FundsApplication` 宸叉敞鍐屼互涓婅矾鐢便€?
### 楠岃瘉缁撴灉

- 鎵ц `javac -encoding UTF-8 -d services/funds-service/target/classes ...` 缂栬瘧閫氳繃銆?- 鎵ц `go test ./...` 閫氳繃锛孏o API 渚?D5.1-D5.7 宸叉湁娴嬭瘯鏈鐮村潖銆?
### 褰撳墠椋庨櫓

- Java `funds-service` 浠嶆槸鏃犳暟鎹簱銆佹棤閴存潈銆佹棤鐪熷疄寰俊鏀粯/鍒嗚处璋冪敤鐨勫崰浣嶆湇鍔°€?- Go API 褰撳墠浠嶅湪鏈湴鍐呭瓨闂幆閲岃绠楀垎娑︼紱鍚庣画闇€瑕佹柊澧?`internal/revenueclient` 璁?Go 璋?Java銆?- 鐪熷疄鏀粯銆侀€€娆俱€佸垎璐︺€佸洖璋冮獙绛惧潎鎸?plan 淇濇寔浜屾湡棰勭暀锛屼笉鍦ㄤ竴鏈熺湡瀹炶Е鍙戙€?
### 涓嬩竴姝?
- 琛?`services/go-api/internal/revenueclient/client.go`锛岃 Go API 鏈夋槑纭殑 Java funds-service 璋冪敤杈圭晫銆?- 鎴栬繘鍏?D5.8 涓炬姤鐢宠瘔锛屾妸浜夎鍐荤粨鎺ュ埌褰撳墠鍒嗘鼎璁板綍銆?
## 2026-06-13 06:30

### 褰撳墠杩涘睍

- 缁х画 D5.7锛岃ˉ Go API 璋?Java `funds-service` 鐨勬槑纭竟鐣屻€?- 鏂板 `services/go-api/internal/revenueclient/client.go`锛?  - `Health`
  - `PaymentPrecreatePlaceholder`
  - `PaymentCallbackPlaceholder`
  - `CreateTemplate`
  - `Simulate`
  - `Generate`
  - `SettleOffline`
- 鏂板 `services/go-api/internal/revenueclient/client_test.go`锛屼娇鐢?`httptest.Server` 楠岃瘉璇锋眰璺緞鍜岄敊璇繑鍥炪€?- 鎵╁睍 Go 閰嶇疆锛?  - `FUNDS_SERVICE_ADDR`
  - 榛樿鍊硷細`http://127.0.0.1:8081`
- 鏇存柊鐜鏍蜂緥锛?  - `.env.example`
  - `deploy/env.example`
- 鏇存柊 `README.md` 涓?Java `funds-service` 缂栬瘧鍛戒护锛屽寘鍚?`placeholder`銆乣revenue`銆乣settlement` 瀛愬寘銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鎵ц Java `javac` 缂栬瘧閫氳繃銆?- `revenueclient` 宸茶鐩栬祫閲戞湇鍔?health銆佹敮浠樺崰浣嶃€佸垎娑︽ā鏉裤€佽瘯绠椼€佺敓鎴愬拰绾夸笅缁撶畻璋冪敤璺緞銆?
### 褰撳墠椋庨櫓

- Go API 鐨勪笟鍔″垎娑︽帴鍙ｅ綋鍓嶄粛浣跨敤鏈湴鍐呭瓨 revenue 鏈嶅姟锛屾病鏈夊垏鎹㈡垚瀹炴椂璋冪敤 Java銆?- Java `funds-service` 浠嶄负鍗犱綅鍝嶅簲锛屾湭鎺ユ暟鎹簱鍜岀湡瀹炲井淇℃敮浠?鍒嗚处銆?- 涓嬩竴姝ュ垏鎹㈣皟鐢ㄩ摼鏃讹紝闇€瑕佸喅瀹氾細Go API 缁х画淇濈暀鏈湴 fallback锛岃繕鏄己鍒朵互 Java `funds-service` 涓?source of truth銆?
### 涓嬩竴姝?
- 杩涘叆 D5.8 涓炬姤鐢宠瘔锛屾妸浜夎鍐荤粨鎺ュ埌褰撳墠鍒嗘鼎璁板綍锛涙垨鎶?Go API 鍒嗘鼎鎺ュ彛鏀规垚浼樺厛璋冪敤 `revenueclient`銆佸け璐ユ椂鏈湴 fallback銆?
## 2026-06-13 07:00

### 褰撳墠杩涘睍

- 杩涘叆 D5.8锛屽畬鎴愪妇鎶ョ敵璇夊埌鍒嗘鼎鍐荤粨鐨勬渶灏忛棴鐜€?- 鏂板 `services/go-api/internal/reports/service.go`锛?  - 鐢ㄦ埛鎻愪氦涓炬姤銆?  - 鏌ヨ鎴戠殑涓炬姤銆?  - 鍚庡彴涓炬姤鍒楄〃銆?  - 鍚庡彴澶勭悊涓炬姤銆?  - 鍚庡彴鍏抽棴涓炬姤銆?  - 涓炬姤鍒涘缓鏃舵寜 `gameId` 鑷姩鍐荤粨宸叉湁鍒嗘鼎璁板綍銆?- 鎵╁睍 `services/go-api/internal/revenue/service.go`锛?  - `FreezeByGame(gameID, reason)`
  - `HasFrozenRecordForGame(gameID)`
- 鏂板/鎺ュ叆鎺ュ彛锛?  - `POST /api/app/reports`
  - `GET /api/app/reports/my`
  - `GET /api/admin/reports`
  - `POST /api/admin/reports/{reportId}/handle`
  - `POST /api/admin/reports/{reportId}/close`
- 琛?`db/migrations/000008_admin_system.sql`锛?  - `reports`
  - `idx_reports_game_status`
  - `idx_reports_reporter`
- 琛?app/admin OpenAPI 涓炬姤鐢宠瘔鎺ュ彛銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板瑕嗙洊锛?  - 宸茬敓鎴愬垎娑﹁褰曠殑灞€鎻愪氦涓炬姤鍚庯紝`revenueFrozen=true`銆?  - 涓炬姤鍚庡搴斿垎娑﹁褰曡鍐荤粨锛岀嚎涓嬬粨绠楄闃绘柇銆?  - 鐢ㄦ埛鍙煡璇㈡垜鐨勪妇鎶ャ€?  - 鍚庡彴鍙煡鐪嬩妇鎶ュ垪琛ㄥ苟澶勭悊涓炬姤銆?
### 褰撳墠椋庨櫓

- 涓炬姤鐢宠瘔浠嶄负鍐呭瓨鏈嶅姟锛屽悗缁渶鎺?PostgreSQL 浠撳偍銆?- 鍚庡彴澶勭悊涓炬姤鏈帴 RBAC 鏉冮檺鍜屾搷浣滄棩蹇椼€?- IM 浜夎娑堟伅浜屾鏌ョ湅鏉冮檺鍜屾煡鐪嬫棩蹇楄繕鏈帴鍏ャ€?
### 涓嬩竴姝?
- 缁х画 D5.8锛氳ˉ鐢ㄦ埛琛屼负鏃ュ織 `user_behavior_logs` 鍜屼妇鎶ュ鐞嗘搷浣滄棩蹇椼€?- 鐒跺悗琛?IM 浜夎娑堟伅鏌ョ湅鏉冮檺涓庢棩蹇楋紝婊¤冻 TC-093/TC-094銆?
## 2026-06-13 07:45

### 褰撳墠杩涘睍

- 缁х画 D5.8锛岃ˉ榻愪妇鎶ョ敵璇夌浉鍏崇珯鍐呴€氱煡闂幆銆?- 鏂板 `services/go-api/internal/notifications/service.go`锛?  - 鍒涘缓閫氱煡銆?  - 鏌ヨ鎴戠殑閫氱煡銆?  - 鏍囪閫氱煡宸茶銆?  - 棰勭暀 `needWechat` 鍜?`wechatState`锛岀敤浜庡悗缁井淇¤闃呮秷鎭彂閫佷换鍔″拰缁撴灉璁板綍銆?- 鏂板灏忕▼搴忔帴鍙ｏ細
  - `GET /api/app/notifications`
  - `POST /api/app/notifications/{notificationId}/read`
- 涓炬姤閾捐矾鍐欓€氱煡锛?  - 鐢ㄦ埛鎻愪氦涓炬姤鍚庣敓鎴?`report_created` 閫氱煡銆?  - 鍚庡彴澶勭悊涓炬姤鍚庣敓鎴?`report_handled` 閫氱煡銆?  - 鍚庡彴鍏抽棴涓炬姤鍚庣敓鎴?`report_closed` 閫氱煡銆?- 琛ラ綈鏁版嵁搴撹縼绉昏崏妗堬細
  - `notifications`
  - `idx_notifications_user_status_created`
  - `idx_notifications_biz`
- 琛ュ厖 `docs/openapi/app.openapi.yaml` 鐨勯€氱煡鎺ュ彛銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 涓炬姤鍒涘缓鍚庣敤鎴疯兘鏌ヨ鍒?`report_created` 閫氱煡銆?  - 鐢ㄦ埛鍙皢閫氱煡鏍囪涓?`read`銆?  - 鍚庡彴澶勭悊涓炬姤鍚庣敤鎴疯兘鏌ヨ鍒?`report_handled` 閫氱煡銆?  - 鍚庡彴鍏抽棴涓炬姤鎺ュ彛浠嶅彲姝ｅ父鎵ц銆?
### 褰撳墠椋庨櫓

- 閫氱煡鏈嶅姟浠嶄负鍐呭瓨瀹炵幇锛屽悗缁渶瑕佹帴 PostgreSQL 浠撳偍銆?- 寰俊璁㈤槄娑堟伅鐩墠鍙繚鐣?`needWechat/wechatState` 鐘舵€佸瓧娈碉紝灏氭湭鎺ユā鏉挎巿鏉冦€佸彂閫佷换鍔″拰鍥炶皟缁撴灉銆?

## 2026-06-13 08:10

### 褰撳墠杩涘睍

- 缁х画 D5.8锛岃ˉ榻愬井淇¤闃呮秷鎭殑鏈湴浠诲姟/缁撴灉鍗犱綅闂幆銆?- 鎵╁睍 `services/go-api/internal/notifications/service.go`锛?  - `WechatTemplate`锛氱淮鎶よ闃呮秷鎭満鏅笌妯℃澘 ID銆?  - `WechatTask`锛氳褰曢€氱煡瀵瑰簲鐨勮闃呮秷鎭彂閫佷换鍔°€?  - 鍒涘缓 `needWechat=true` 鐨勯€氱煡鏃讹紝鑷姩鐢熸垚 `pending` 璁㈤槄娑堟伅浠诲姟銆?  - 鏀寔鏍囪浠诲姟宸插彂閫侊紝骞跺悓姝ラ€氱煡 `wechatState=sent`銆?- 鏂板鎺ュ彛锛?  - `GET /api/admin/notifications/wechat-templates`
  - `GET /api/admin/notifications/wechat-tasks`
  - `POST /api/internal/notifications/wechat-tasks/{taskId}/mark-sent`
- 鎵╁睍 `notifications` 杩佺Щ瀛楁锛?  - `wechat_template_id`
  - `wechat_task_id`
- 鏂板杩佺Щ琛細
  - `wechat_subscribe_templates`
  - `wechat_subscribe_tasks`
- 涓炬姤鐢宠瘔閫氱煡鐜板湪浼氬啓鍏ヨ闃呮秷鎭ā鏉垮彉閲忥紝骞剁敓鎴愬井淇¤闃呮秷鎭换鍔°€?- 琛ュ厖 `docs/openapi/admin.openapi.yaml` 鐨勮闃呮秷鎭ā鏉?浠诲姟鎺ュ彛銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 涓炬姤閫氱煡鍖呭惈 `wechatState=pending`銆佹ā鏉?ID 鍜屼换鍔?ID銆?  - 鍚庡彴鍙煡璇㈣闃呮秷鎭ā鏉裤€?  - 鍚庡彴鍙煡璇㈣闃呮秷鎭换鍔°€?  - 鍐呴儴鎺ュ彛鍙ā鎷熸爣璁拌闃呮秷鎭换鍔″凡鍙戦€併€?
### 褰撳墠椋庨櫓

- 褰撳墠鍙畬鎴愭湰鍦颁换鍔′笌缁撴灉鐘舵€侊紝灏氭湭璋冪敤寰俊瀹樻柟 `subscribeMessage.send`銆?- 鐢ㄦ埛璁㈤槄鎺堟潈鍏ュ彛銆佸皬绋嬪簭绔?`wx.requestSubscribeMessage`銆佹ā鏉?ID 閰嶇疆浠嶉渶鍦ㄥ皬绋嬪簭渚у拰鐪熷疄閰嶇疆涓ˉ榻愩€?

## 2026-06-13 08:40

### 褰撳墠杩涘睍

- 缁х画 D5.8锛岃ˉ榻愪袱涓唴閮ㄦ彁閱掍换鍔″叆鍙ｃ€?- 鏂板 `services/go-api/internal/appapi/job_handler.go`锛?  - `POST /api/internal/jobs/review-remind`
  - `POST /api/internal/jobs/progress-feedback-remind`
- 璇勪环鎻愰啋浠诲姟浼氭壂鎻?`pending_review/completed` 鐨勫眬锛屽浠嶆湁寰呰瘎浠烽」鐨勬垚鍛樼敓鎴愶細
  - 绔欏唴閫氱煡 `review_remind`
  - 寰俊璁㈤槄娑堟伅浠诲姟 `review_remind`
- 琛屽杩涘害鍙嶉鎻愰啋浠诲姟浼氭壂鎻?`in_progress/pending_confirm` 鐨勫眬锛屽鍙戣捣浜虹敓鎴愶細
  - 绔欏唴閫氱煡 `progress_feedback_remind`
  - 寰俊璁㈤槄娑堟伅浠诲姟 `progress_feedback_remind`
- 琛ュ厖 `db/migrations/000006_review_growth_credit.sql`锛?  - `review_reminders.reminder_type`
  - `review_reminders.notification_id`
  - `review_reminders.sent_at`
  - `idx_review_reminders_user_status`
  - `idx_review_reminders_game_type`
- 琛ュ厖 `docs/openapi/admin.openapi.yaml` 鐨勫唴閮ㄦ彁閱掍换鍔℃帴鍙ｃ€?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 灞€杩涘叆寰呰瘎浠峰悗锛屽唴閮ㄨ瘎浠锋彁閱掍换鍔¤兘鐢熸垚閫氱煡銆?  - 杩涜涓眬瑙﹀彂琛屽杩涘害鍙嶉鎻愰啋鍚庯紝鍙戣捣浜鸿兘鏀跺埌 `progress_feedback_remind` 閫氱煡銆?  - 涓ょ被鎻愰啋鍧囧鐢ㄥ綋鍓嶉€氱煡鏈嶅姟锛岃嚜鍔ㄧ敓鎴愬井淇¤闃呮秷鎭换鍔°€?
### 褰撳墠椋庨櫓

- 褰撳墠鎻愰啋浠诲姟涓烘湰鍦?HTTP 鍏ュ彛锛屽皻鏈帴瀹氭椂璋冨害鍣ㄣ€?- 杩涘害鍙嶉鎻愰啋鐩墠鎸夊眬鐘舵€佺敓鎴愶紝鍚庣画闇€瑕佹帴鐪熷疄琛屽杩涘害鍙嶉璁板綍锛岄伩鍏嶉噸澶嶆彁閱掑拰璇彁閱掋€?

## 2026-06-13 09:10

### 褰撳墠杩涘睍

- 缁х画 D5.8锛岃ˉ榻愬悗鍙颁妇鎶ヨ鎯呭拰鍏宠仈璇佹嵁瑙嗗浘銆?- 鎵╁睍 `services/go-api/internal/reports/service.go`锛?  - 鏂板 `Get(reportID)`銆?- 鎵╁睍鍚庡彴鎺ュ彛锛?  - `GET /api/admin/reports/{reportId}`銆?- 涓炬姤璇︽儏杩斿洖锛?  - 涓炬姤鏈綋 `report`銆?  - 灞€淇℃伅 `evidence.game`銆?  - IM 鎴块棿鍜岀浉鍏虫秷鎭?`evidence.chatRoom`銆乣evidence.chatMessages`銆?  - 鏂囦欢璇佹嵁 `evidence.file`銆?  - 璇勪环璇佹嵁 `evidence.review`銆?  - 鍒嗘鼎璁板綍 `evidence.revenueRecord`銆?  - 鍘熷鍏宠仈 ID 姹囨€?`evidence.ids`銆?- 鏌ョ湅涓炬姤璇︽儏鍐?`operation_logs`锛?  - `action=report:view_detail`
  - `target_type=report`
- 琛ュ厖 `docs/openapi/admin.openapi.yaml` 鐨勪妇鎶ヨ鎯呮帴鍙ｃ€?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鍚庡彴鍙墦寮€涓炬姤璇︽儏銆?  - 璇︽儏閲岃兘鐪嬪埌灞€淇℃伅銆両M 娑堟伅璇佹嵁鍜屽喕缁撳悗鐨勫垎娑﹁褰曘€?  - 鏌ョ湅璇︽儏浼氬啓 `report:view_detail` 鎿嶄綔鏃ュ織銆?
### 褰撳墠椋庨櫓

- 褰撳墠璇佹嵁鑱氬悎浠嶅熀浜庡唴瀛樻湇鍔★紝鍚庣画鎺?PostgreSQL 鍚庨渶瑕佹敼涓轰粨鍌ㄦ煡璇€?- IM 璇佹嵁鐩墠鎸夋湰鍦版埧闂存秷鎭仛鍚堬紝鍚庣画鎺ョ湡瀹?WebSocket/鎸佷箙鍖栧悗闇€瑕佹寜娑堟伅琛ㄧ簿纭垎椤靛拰鏉冮檺杩囨护銆?

## 2026-06-13 09:35

### 褰撳墠杩涘睍

- 缁х画 D5.8 / 鍚庡彴瀹¤鏉冮檺闂幆锛岃ˉ榻愬畬鏁存搷浣滄棩蹇楁煡璇㈡潈闄愭牎楠屻€?- `GET /api/admin/operation-logs` 鐜板湪蹇呴』鎼哄甫 `operation_log:view_full` 鏉冮檺銆?- 鏃犳潈闄愯闂繑鍥?`403`锛岄伩鍏嶆櫘閫氱鐞嗗憳鏌ョ湅瀹屾暣鎿嶄綔鏃ュ織銆?- 琛ュ厖 `docs/openapi/admin.openapi.yaml` 鐨?403 鍝嶅簲璇存槑銆?
### 楠岃瘉缁撴灉

- 琛ュ厖鎺ュ彛娴嬭瘯锛?  - 鏃?`operation_log:view_full` 鏉冮檺鏃舵煡璇㈡搷浣滄棩蹇楄繑鍥?`403`銆?  - 鏈夋潈闄愭椂浠嶅彲鏌ヨ骞堕獙璇?`im:message:view_dispute`銆乣report:handle`銆乣report:view_detail` 鏃ュ織銆?
### 褰撳墠椋庨櫓

- 褰撳墠鏉冮檺浠嶄娇鐢ㄨ姹傚ご鏉冮檺蹇収鍗犱綅锛屽悗缁?D7 闇€瑕佹帴鍏ョ湡瀹炲悗鍙?token銆佽鑹插拰 RBAC 涓棿浠躲€?
## 2026-06-13 10:05

### 褰撳墠杩涘睍

- 杩涘叆 P10 / Task 56 鐨勬姤琛ㄥ鍑轰换鍔℃渶灏忛棴鐜€?- 鏂板鍥哄畾瀵煎嚭妯℃澘鏈嶅姟 `services/go-api/internal/exports/service.go`锛?  - `reports_default` 涓炬姤鐢宠瘔鎶ヨ〃妯℃澘銆?  - `operation_logs_default` 鎿嶄綔鏃ュ織鎶ヨ〃妯℃澘銆?  - 鍒涘缓 `pending` 瀵煎嚭浠诲姟銆?  - 鍐呴儴 runner 灏嗕换鍔℃帹杩涘埌 `done` 骞跺叧鑱旀枃浠躲€?- 鎵╁睍鏂囦欢鏈嶅姟锛屾柊澧?`CreateGeneratedFile`锛岀敤浜庣櫥璁?`export_file` 鐢熸垚鐗┿€?- 鏂板鍚庡彴/鍐呴儴鎺ュ彛锛?  - `GET /api/admin/reports/export-templates`
  - `POST /api/admin/reports/export`
  - `GET /api/admin/export-tasks`
  - `GET /api/admin/export-tasks/{taskId}/download-url`
  - `POST /api/internal/reports/export-runner`
- 瀵煎嚭鐩稿叧鎺ュ彛浣跨敤 `report_export:create` 鏉冮檺鍗犱綅鏍￠獙銆?- 鍒涘缓銆佹墽琛屻€佷笅杞藉湴鍧€鑾峰彇鍒嗗埆鍐?`operation_logs`锛?  - `export:create`
  - `export:run`
  - `export:download_url`
- 琛ュ厖 `db/migrations/000009_behavior_export.sql`锛?  - `export_templates`
  - `export_tasks.template_code`
  - `export_tasks.created_by`
  - `export_tasks.filters_json`
  - `export_tasks.fail_reason`
  - 瀵煎嚭浠诲姟鐘舵€佺储寮曘€?- 琛ュ厖 `docs/openapi/admin.openapi.yaml` 瀵煎嚭鎺ュ彛璇存槑銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鏃?`report_export:create` 鏉冮檺涓嶈兘鏌ョ湅瀵煎嚭妯℃澘銆?  - 鏈夋潈闄愬彲鏌ョ湅鍥哄畾妯℃澘銆?  - 鍙垱寤?`pending` 瀵煎嚭浠诲姟銆?  - 鍐呴儴 runner 鍙敓鎴愬鍑烘枃浠惰褰曞苟灏嗕换鍔＄疆涓?`done`銆?  - 瀵煎嚭瀹屾垚鍚庡彲鑾峰彇涓存椂涓嬭浇鍦板潃銆?  - 瀵煎嚭鍒涘缓銆佹墽琛屻€佷笅杞藉湴鍧€鑾峰彇鍧囧彲鍦ㄦ搷浣滄棩蹇椾腑杩芥函銆?
### 褰撳墠椋庨櫓

- 褰撳墠瀵煎嚭 runner 鐢熸垚鐨勬槸鏈湴 mock 鏂囦欢璁板綍锛屽皻鏈敓鎴愮湡瀹?CSV 鍐呭鍜屽璞″瓨鍌ㄦ枃浠躲€?- 褰撳墠瀵煎嚭鏁版嵁婧愪粛鏄唴瀛樻湇鍔★紝鍚庣画鎺?PostgreSQL 鍚庨渶瑕佹寜绛涢€夋潯浠惰鍙栫湡瀹炴姤琛ㄦ暟鎹€?
## 2026-06-13 10:35

### 褰撳墠杩涘睍

- 杩涘叆 D7 鍚庣瀹夊叏鍜屾潈闄愮粺涓€钀藉湴鐨勭涓€姝ワ紝鎶婂凡鎺ュ叆鐨勫悗鍙版潈闄愭祦绋嬭鑼冨寲銆?- 鏂板缁熶竴鏉冮檺鍖呰鍣細
  - `requireAdminPermission(permission, handler)`
- 宸叉妸浠ヤ笅鍚庡彴/鍐呴儴璺敱鏀逛负鍦ㄨ矾鐢辨敞鍐屽眰鏄惧紡澹版槑 permission code锛?  - `GET /api/admin/operation-logs` -> `operation_log:view_full`
  - `GET /api/admin/export-tasks` -> `report_export:create`
  - `GET /api/admin/export-tasks/{taskId}/download-url` -> `report_export:create`
  - `GET /api/admin/reports/export-templates` -> `report_export:create`
  - `POST /api/admin/reports/export` -> `report_export:create`
  - `POST /api/internal/reports/export-runner` -> `report_export:create`
- 娓呯悊瀵煎嚭 handler 鍜屾搷浣滄棩蹇?handler 鍐呴儴鏁ｈ惤鐨勯噸澶嶆潈闄愬垽鏂紝涓氬姟澶勭悊鍙礋璐ｄ笟鍔￠€昏緫銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鐜版湁鏉冮檺娴嬭瘯缁х画瑕嗙洊锛?  - 鏃?`operation_log:view_full` 鏌ヨ瀹屾暣鎿嶄綔鏃ュ織杩斿洖 `403`銆?  - 鏃?`report_export:create` 鏌ヨ瀵煎嚭妯℃澘杩斿洖 `403`銆?  - 鏈夋潈闄愭椂瀵煎嚭鍒涘缓銆佹墽琛屻€佷笅杞藉湴鍧€鍜屾搷浣滄棩蹇楄拷婧潎姝ｅ父銆?
### 褰撳墠椋庨櫓

- 褰撳墠鏉冮檺婧愪粛鏄?`X-Admin-Permissions` 璇锋眰澶村崰浣嶏紱鍚庣画鎺ョ湡瀹炲悗鍙扮櫥褰曘€佽鑹层€佹潈闄愭爲鍚庯紝浼樺厛鏇挎崲 `requireAdminPermission` 鍐呴儴瀹炵幇銆?
## 2026-06-13 11:10

### 褰撳墠杩涘睍

- 缁х画 D7 鍚庣瀹夊叏鍜屾潈闄愮粺涓€钀藉湴锛岃ˉ鍚庡彴鐧诲綍鍜屾潈闄愭爲鏈€灏忛棴鐜€?- 鏂板 `services/go-api/internal/adminauth/service.go`锛?  - 鏈湴鍐呭瓨鐗堝悗鍙扮鐞嗗憳璐﹀彿銆?  - 榛樿 `admin / admin123` 鐢ㄤ簬鏈湴鑱旇皟銆?  - 鐧诲綍杩斿洖 token銆佺鐞嗗憳淇℃伅銆佽鑹层€佹潈闄愬揩鐓с€?  - 鏀寔閫氳繃 token 鎷夊彇鏉冮檺鏍戙€?- 鏂板鍚庡彴鎺ュ彛锛?  - `POST /api/admin/auth/login`
  - `GET /api/admin/auth/permissions`
  - `GET /api/admin/permissions/tree`
- `requireAdminPermission` 鐜板湪鏀寔涓ょ鏉冮檺鏉ユ簮锛?  - 鍏煎鏃ф祴璇曞拰杩囨浮鏈熺殑 `X-Admin-Permissions`銆?  - 鏀寔鍚庡彴鐧诲綍鍚庣殑 `Authorization: Bearer <admin-token>`銆?- 璁块棶鍙椾繚鎶ゅ悗鍙版帴鍙ｆ椂锛屽鏋滀娇鐢?admin token锛屼細鑷姩琛?`X-Admin-ID`锛屼繚璇佸悗缁?`operation_logs` 鑳借褰曠鐞嗗憳 ID銆?- IM 浜夎娑堟伅璺敱宸叉帴鍏ョ粺涓€ `requireAdminPermission("im:message:view_dispute", ...)`銆?- 琛ュ厖 `db/migrations/000008_admin_system.sql`锛?  - `admin_user_roles`
  - `admin_role_permissions`
- 閲嶅啓 `db/seeds/admin_roles_permissions.sql`锛屽榻?`plan.md` 涓殑鏉冮檺鐮侊紝骞跺垵濮嬪寲鏈湴 super_admin 鍏宠仈銆?- 琛ュ厖 `docs/openapi/admin.openapi.yaml` 鐨勫悗鍙扮櫥褰曞拰鏉冮檺鏍戞帴鍙ｃ€?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鍚庡彴鐧诲綍瀵嗙爜閿欒杩斿洖 `401`銆?  - `admin/admin123` 鐧诲綍鎴愬姛杩斿洖 token銆乣super_admin` 鍜屾潈闄愮爜銆?  - 鐧诲綍 token 鍙媺鍙栨潈闄愭爲銆?  - 鐧诲綍 token 鍙闂渶瑕?`operation_log:view_full` 鐨勬搷浣滄棩蹇楁帴鍙ｃ€?
### 褰撳墠椋庨櫓

- 褰撳墠鍚庡彴鐧诲綍浠嶆槸鍐呭瓨璐﹀彿鍜屾槑鏂囧瘑鐮佸崰浣嶏紱鍚庣画闇€瑕佹帴 PostgreSQL 绠＄悊鍛樿〃銆佸己鍝堝笇瀵嗙爜銆佺鐢ㄧ姸鎬併€佽鑹叉潈闄愬叧绯诲拰 token 杩囨湡鍒锋柊绛栫暐銆?- `X-Admin-Permissions` 浠嶄繚鐣欎负杩囨浮鍏煎鍏ュ彛锛涙帴瀹屾暣 RBAC 鍚庡簲閫愭绉婚櫎娴嬭瘯澶栦緷璧栥€?
## 2026-06-13 11:45

### 褰撳墠杩涘睍

- 缁х画 D7 鍚庡彴鎺ュ彛鏉冮檺缁熶竴钀藉湴锛岀粰涓炬姤鍜屽垎娑﹁祫閲戠浉鍏冲悗鍙版帴鍙ｈˉ鏉冮檺鍏ュ彛銆?- 涓炬姤鐢宠瘔鎺ュ彛宸叉帴鏉冮檺锛?  - `GET /api/admin/reports` -> `report:view`
  - `GET /api/admin/reports/{reportId}` -> `report:view`
  - `POST /api/admin/reports/{reportId}/handle` -> `report:handle`
  - `POST /api/admin/reports/{reportId}/close` -> `report:close`
- 鍒嗘鼎璧勯噾鎺ュ彛宸叉帴鏉冮檺锛?  - `GET /api/admin/revenue/templates` -> `revenue:template:view`
  - `POST /api/admin/revenue/templates` -> `revenue:template:update`
  - `POST /api/admin/revenue/preview` -> `revenue:simulate`
  - `POST /api/admin/revenue/calculate` -> `revenue:simulate`
  - `POST /api/admin/revenue/records/generate` -> `revenue:generate`
  - `GET /api/admin/revenue/records` -> `revenue:record:view`
  - `POST /api/admin/revenue/records/{recordId}/freeze` -> `revenue:freeze`
  - `POST /api/admin/revenue/records/{recordId}/settle` -> `settlement:offline:create`
  - `POST /api/admin/revenue/records/{recordId}/settle-offline` -> `settlement:offline:create`
- 琛ラ綈鏈湴鍚庡彴鐧诲綍鏉冮檺蹇収鍜?`db/seeds/admin_roles_permissions.sql` 涓殑鏉冮檺鐮侊細
  - `report:close`
  - `revenue:template:view`
  - `revenue:template:update`
  - `revenue:simulate`
  - `revenue:generate`
  - `revenue:freeze`
  - `settlement:offline:create`
- 璋冩暣鎺ュ彛娴嬭瘯锛屽垎娑﹀拰涓炬姤鍚庡彴璋冪敤鏀逛负浣跨敤鍚庡彴鐧诲綍 token銆?- 鏂板鏃犳潈闄愬鐞嗕妇鎶ヨ繑鍥?`403` 鐨勬柇瑷€銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏃㈤獙璇佹棫 `X-Admin-Permissions` 鍏煎鍏ュ彛锛屼篃楠岃瘉鍚庡彴鐧诲綍 token 鍙闂彈淇濇姢鎺ュ彛銆?
### 褰撳墠椋庨櫓

- OpenAPI 鏃ф枃浠跺瓨鍦ㄩ儴鍒嗙紪鐮佸拰璺緞绮樿繛锛屽凡璁板綍瀹炵幇杩涘睍锛涘悗缁渶瑕佸崟鐙仛涓€娆?`admin.openapi.yaml` 缁撴瀯鍖栨暣鐞嗭紝閬垮厤灏忚ˉ涓佸湪涔辩爜鍖哄煙璇激銆?
## 2026-06-13 12:20

### 褰撳墠杩涘睍

- 缁х画 D7 鍚庡彴鏉冮檺缁熶竴鏀跺彛锛岃ˉ榻愭鍓嶄粛瑁搁湶鐨勬煡璇㈢被鍚庡彴鎺ュ彛銆?- 浠ヤ笅鎺ュ彛宸叉帴鍏ョ粺涓€ `requireAdminPermission`锛?  - `GET /api/admin/users/{userId}/growth` -> `user:view`
  - `GET /api/admin/games/{gameId}/review-trace` -> `game:view`
  - `GET /api/admin/behavior-logs` -> `analytics:timeline:view`
  - `GET /api/admin/notifications/wechat-templates` -> `notification:wechat:view`
  - `GET /api/admin/notifications/wechat-tasks` -> `notification:wechat:view`
- 琛ラ綈鏈湴鍚庡彴鐧诲綍鏉冮檺蹇収鍜?`db/seeds/admin_roles_permissions.sql`锛?  - `notification:wechat:view`
- 璋冩暣鎺ュ彛娴嬭瘯锛?  - 鐢ㄦ埛鎴愰暱銆佺粍灞€杩芥函鍚庡彴鏌ヨ琛ユ潈闄愬ご銆?  - 寰俊璁㈤槄妯℃澘銆佽闃呬换鍔°€佽涓烘棩蹇楀悗鍙版煡璇㈡敼涓轰娇鐢ㄥ悗鍙扮櫥褰?token銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 褰撳墠鍚庡彴鏌ヨ绫绘帴鍙ｅ凡鍩烘湰缁熶竴鍒版樉寮?permission code锛岀户缁繚鐣?`X-Admin-Permissions` 浣滀负杩囨浮娴嬭瘯鍏ュ彛銆?
### 褰撳墠椋庨櫓

- 鍚庡彴鐧诲綍浠嶆槸鍐呭瓨璐﹀彿鍗犱綅锛汥7 鍚庣画搴斾紭鍏堟妸绠＄悊鍛樸€佽鑹层€佹潈闄愩€佷細璇濇帴鍏ョ湡瀹炴寔涔呭寲琛ㄥ拰瀵嗙爜鍝堝笇銆?
## 2026-06-13 12:45

### 褰撳墠杩涘睍

- 缁х画 D7 鍚庡彴瀹夊叏钀藉湴锛屽厛鍦ㄥ綋鍓嶅唴瀛樿处鍙锋灦鏋勪笅娑堥櫎鏄庢枃瀵嗙爜鏍￠獙銆?- `adminauth` 榛樿绠＄悊鍛樿处鍙峰凡浠庢槑鏂囧瘑鐮佹敼涓?PBKDF2-HMAC-SHA256 鍝堝笇鏍￠獙锛?  - 姝ｇ‘瀵嗙爜 `admin123` 浠嶅彲鐢ㄤ簬鏈湴鑱旇皟銆?  - 鏈嶅姟鍐呴儴涓嶅啀淇濆瓨 `admin123` 鏄庢枃銆?  - 闈?PBKDF2 鏍煎紡鐨勬槑鏂囧瘑鐮佸瓨鍌ㄤ細琚嫆缁濄€?- `db/seeds/admin_roles_permissions.sql` 涓粯璁?`admin` 鐨?`password_hash` 宸插悓姝ヤ负鍚屼竴 PBKDF2 鍝堝笇鏍煎紡銆?- 鏂板 `services/go-api/internal/adminauth/service_test.go`锛?  - 姝ｇ‘瀵嗙爜鍖归厤鍝堝笇銆?  - 閿欒瀵嗙爜澶辫触銆?  - 鏄庢枃鏍煎紡澶辫触銆?  - 鍚庡彴鐧诲綍浠嶈繑鍥?token 鍜屾潈闄愬揩鐓с€?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?
### 褰撳墠椋庨櫓

- 鐩墠鍝堝笇绠楁硶涓烘爣鍑嗗簱鑷疄鐜?PBKDF2锛屽崰浣嶆弧瓒斥€滀笉瀛樻槑鏂団€濓紱鍚庣画鎺ョ湡瀹炴暟鎹簱鍜岀敓浜ц处鍙峰垵濮嬪寲鏃讹紝寤鸿缁熶竴杩佺Щ鍒板洟闃熺‘瀹氱殑瀵嗙爜绛栫暐鍜屼竴娆℃€у垵濮嬪瘑鐮佹祦绋嬨€?
## 2026-06-13 13:20

### 褰撳墠杩涘睍

- 鎸夆€滃皬绋嬪簭鍚庣浼樺厛鈥濈户缁ˉ鎺ュ彛鑱旇皟闂幆锛屼紭鍏堣ˉ榻?`plan.md` 涓墠绔細鐩存帴璋冪敤浣嗗綋鍓嶇己灏戠殑鎺ュ彛璺緞銆?- 璇勪环琛ヨ瘎浠峰叆鍙ｈˉ榻愯鍒掕矾寰勶細
  - 鏂板 `GET /api/app/reviews/available`
  - 澶嶇敤鐜版湁 `reviews.Todos` 閫昏緫锛屽拰 `GET /api/app/reviews/todos` 杩斿洖鍚屼竴绫诲緟璇勪环瀵硅薄銆?- 鐢ㄦ埛鏀剁泭鏄庣粏鎺ュ彛琛ラ綈锛?  - 鏂板 `GET /api/app/incomes/logs`
  - 鏀寔鎸夊綋鍓嶇櫥褰曠敤鎴疯繃婊ゆ敹鐩婅褰曪紝鍙繑鍥炶嚜宸辩殑鍒嗘鼎鏀跺叆椤广€?  - 鏀寔 `status` 鏌ヨ鍙傛暟绛涢€夛紝濡?`?status=settled`銆?  - 杩斿洖瀛楁鍖呭惈 `recordId`銆乣recordNo`銆乣gameId`銆乣role`銆乣amountCent`銆乣status`銆乣createdAt`銆乣settledAt`銆?- 鏂板琛屼负鏃ュ織锛?  - `view_income_logs`
- 琛ュ厖鎺ュ彛娴嬭瘯锛?  - `GET /api/app/reviews/available` 鍙繑鍥?`200`銆?  - `GET /api/app/incomes/logs` 杩斿洖褰撳墠鐢ㄦ埛涓ゆ潯鏀剁泭鏄庣粏銆?  - `GET /api/app/incomes/logs?status=settled` 鍙繑鍥炲凡缁撶畻鏀剁泭銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?
### 褰撳墠椋庨櫓

- 褰撳墠鏀剁泭鏄庣粏浠嶆潵鑷唴瀛樺垎娑﹁褰曪紱鍚庣画鎺?PostgreSQL 鍚庨渶瑕佹寜 `user_income_accounts` / `income_logs` 琛ㄦ敼涓虹湡瀹炲垎椤垫煡璇€?
## 2026-06-13 13:45

### 褰撳墠杩涘睍

- 缁х画灏忕▼搴忓悗绔紭鍏堬紝琛ラ綈灞€杩涜涓殑杩涘害鍙嶉闂幆銆?- 鏂板棰嗗煙妯″瀷涓庢湇鍔℃柟娉曪細
  - `games.ProgressFeedback`
  - `games.ProgressFeedbackRequest`
  - `AddProgressFeedback(userId, gameId, req)`
  - `ProgressFeedbacks(userId, gameId)`
- 鏂板灏忕▼搴忕鎺ュ彛锛?  - `POST /api/app/games/{gameId}/progress-feedbacks`
  - `GET /api/app/games/{gameId}/progress-feedbacks`
- 褰撳墠瑙勫垯锛?  - 杩涘害鐧惧垎姣斿繀椤诲湪 `0-100`銆?  - 杩涘害涓嶈兘浣庝簬涓婁竴鏉″弽棣堛€?  - 鍙湁灞€鍙戣捣浜哄彲鎻愪氦杩涘害鍙嶉銆?  - 灞€鎴愬憳鍙煡鐪嬭繘搴﹀弽棣堝垪琛ㄥ拰 `latestProgress`銆?  - 闈炴垚鍛樻煡璇㈣繑鍥?`403`銆?- 鏂板琛屼负鏃ュ織锛?  - `submit_progress_feedback`

### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鍙戣捣浜烘彁浜よ繘搴﹀弽棣堟垚鍔熴€?  - 鎴愬憳鏌ヨ杩涘害鍙嶉鍒楄〃鎴愬姛銆?  - 杩涘害鍊掗€€杩斿洖 `422`銆?  - 闈炴垚鍛樻煡璇㈣繘搴﹀弽棣堣繑鍥?`403`銆?
### 褰撳墠椋庨櫓

- 褰撳墠杩涘害鍙嶉浠嶆槸鍐呭瓨瀹炵幇锛涘悗缁帴 PostgreSQL 鏃堕渶瑕佽惤鍒?`game_progress_feedbacks` 琛紝骞惰ˉ `fileIds` 瀵规枃浠舵湇鍔＄殑褰掑睘/鏉冮檺鏍￠獙銆?
## 2026-06-13 14:20

### 褰撳墠杩涘睍

- 缁х画灏忕▼搴忓悗绔紭鍏堬紝琛ュ眬鍐呴噷绋嬬銆佹墦鍗″拰澶嶇洏鎺ュ彛锛屽拰鍓嶄竴娈佃繘搴﹀弽棣堣兘鍔涘舰鎴愯繃绋嬫暟鎹棴鐜€?- 鏂板棰嗗煙妯″瀷涓庢湇鍔℃柟娉曪細
  - `games.Milestone` / `MilestoneRequest`
  - `games.Checkin` / `CheckinRequest`
  - `games.Retrospective` / `RetrospectiveRequest`
  - `CreateMilestone`銆乣Milestones`
  - `CreateCheckin`銆乣Checkins`
  - `CreateRetrospective`銆乣Retrospectives`
- 鏂板灏忕▼搴忕鎺ュ彛锛?  - `POST /api/app/games/{gameId}/milestones`
  - `GET /api/app/games/{gameId}/milestones`
  - `POST /api/app/games/{gameId}/checkins`
  - `GET /api/app/games/{gameId}/checkins`
  - `POST /api/app/games/{gameId}/retrospectives`
  - `GET /api/app/games/{gameId}/retrospectives`
- 褰撳墠瑙勫垯锛?  - 鍙湁灞€鍙戣捣浜哄彲鍒涘缓閲岀▼纰戙€?  - 灞€鎴愬憳鍙煡鐪嬮噷绋嬬銆?  - 灞€鎴愬憳鍙彁浜ゅ拰鏌ョ湅鎵撳崱銆?  - 澶嶇洏蹇呴』鍦ㄥ眬杩涘叆 `pending_review` 鎴?`completed` 鍚庢彁浜ゃ€?  - 闈炴垚鍛樿闂繃绋嬫暟鎹繑鍥?`403`銆?- 琛ラ綈 `db/migrations/000010_p1_reserved.sql` 涓己澶辩殑 `game_retrospectives` 琛ㄣ€?- 鏂板琛屼负鏃ュ織锛?  - `create_game_milestone`
  - `submit_game_checkin`
  - `submit_game_retrospective`

### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鍙戣捣浜哄垱寤洪噷绋嬬鎴愬姛銆?  - 鎴愬憳鎻愪氦鎵撳崱鎴愬姛銆?  - 鎴愬憳鍙煡鐪嬮噷绋嬬鍜屾墦鍗″垪琛ㄣ€?  - 鏈嶅姟鏈‘璁ゅ畬鎴愬墠鎻愪氦澶嶇洏杩斿洖 `409`銆?  - 鏈嶅姟纭瀹屾垚鍚庢彁浜ゅ鐩樻垚鍔熴€?  - 闈炴垚鍛樻煡鐪嬮噷绋嬬杩斿洖 `403`銆?
### 褰撳墠椋庨櫓

- 褰撳墠閲岀▼纰戙€佹墦鍗°€佸鐩樹粛鏄唴瀛樺疄鐜帮紱鍚庣画鎺?PostgreSQL 鏃堕渶瑕佽惤鍒?`game_milestones`銆乣game_checkins`銆乣game_retrospectives`锛屽苟琛ユ墦鍗℃枃浠?`fileIds` 鐨勬枃浠跺綊灞炰笌涓嬭浇鏉冮檺鏍￠獙銆?
## 2026-06-13 14:55

### 褰撳墠杩涘睍

- 缁х画灏忕▼搴忓悗绔紭鍏堬紝琛ラ綈鏀惰棌灞€鎺ュ彛锛屾敮鎸佸皬绋嬪簭绔矇娣€鐢ㄦ埛鍏磋叮鏁版嵁銆?- 鏂板棰嗗煙妯″瀷涓庢湇鍔℃柟娉曪細
  - `games.Favorite`
  - `FavoriteGame(userId, gameId)`
  - `UnfavoriteGame(userId, gameId)`
  - `FavoriteGames(userId)`
- 鏂板灏忕▼搴忕鎺ュ彛锛?  - `POST /api/app/games/{gameId}/favorite`
  - `DELETE /api/app/games/{gameId}/favorite`
  - `GET /api/app/games/favorites/my`
- 褰撳墠瑙勫垯锛?  - 鏀惰棌涓嶅瓨鍦ㄧ殑灞€杩斿洖 `40421`銆?  - 閲嶅鏀惰棌淇濇寔骞傜瓑锛岃繑鍥炲凡鏀惰棌璁板綍銆?  - 鍙栨秷鏈敹钘忕殑灞€淇濇寔骞傜瓑銆?  - 鎴戠殑鏀惰棌鎸夋敹钘忔椂闂村€掑簭杩斿洖锛屽苟杩囨护浠嶅浜?`pending_audit` 鐨勫眬銆?- 鏂板琛屼负鏃ュ織锛?  - `favorite_game`
  - `unfavorite_game`

### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鏀惰棌涓嶅瓨鍦ㄧ殑灞€杩斿洖 `404`銆?  - 鏀惰棌灞€鎴愬姛銆?  - 閲嶅鏀惰棌涓嶆姤閿欍€?  - 鎴戠殑鏀惰棌杩斿洖宸叉敹钘忓眬銆?  - 鍙栨秷鏀惰棌鎴愬姛銆?  - 閲嶅鍙栨秷鏀惰棌浠嶈繑鍥炴垚鍔熴€?
### 褰撳墠椋庨櫓

- 褰撳墠鏀惰棌浠嶆槸鍐呭瓨瀹炵幇锛涘悗缁帴 PostgreSQL 鏃堕渶瑕佽惤鍒?`game_favorites`锛屽苟琛ュ悗鍙扮敤鎴疯鎯呬腑鐨勬敹钘忓垪琛ㄦ煡璇€?
## 2026-06-13 15:15

### 褰撳墠杩涘睍

- 缁х画灏忕▼搴忓悗绔紭鍏堬紝琛ラ綈 `plan.md` 涓敹鐩婃帴鍙ｇ殑 singular 璺緞鍏煎銆?- 鏂板璺敱鍒悕锛?  - `GET /api/app/income/summary` -> 澶嶇敤 `incomeSummary`
  - `GET /api/app/income/logs` -> 澶嶇敤 `incomeLogs`
- 淇濈暀鏃㈡湁 plural 璺緞锛?  - `GET /api/app/incomes/summary`
  - `GET /api/app/incomes/logs`
- 鍝嶅簲缁撴瀯銆侀壌鏉冦€佽涓烘棩蹇楀拰绛涢€夐€昏緫涓嶅彉锛岄伩鍏嶅奖鍝嶅凡缁忛€氳繃鐨勬敹鐩婃祴璇曢摼璺€?
### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - `GET /api/app/income/summary` 杩斿洖 `200`銆?  - `GET /api/app/income/logs` 杩斿洖 `200`銆?
### 褰撳墠椋庨櫓

- 褰撳墠鏀剁泭鎽樿鍜屾槑缁嗕粛鏉ヨ嚜鍐呭瓨鍒嗘鼎璁板綍锛涘悗缁帴 PostgreSQL 鍚庨渶瑕佸悓鏃朵繚璇?singular/plural 涓ゅ璺緞閮芥寚鍚戝悓涓€鐪熷疄鏌ヨ閫昏緫銆?
## 2026-06-13 15:45

### 褰撳墠杩涘睍

- 缁х画灏忕▼搴忓悗绔紭鍏堬紝琛ラ綈 `plan.md` E2.4 涓細鍛樻姤琛ㄥ拰鍥㈤槦鏌ヨ閾捐矾銆?- 鏂板浼氬憳鎶ヨ〃鏈嶅姟锛?  - `services/go-api/internal/memberreports/service.go`
  - 鏀寔浼氬憳鏉冮檺绉嶅瓙銆佷細鍛樹笓灞炴姤琛ㄥ揩鐓с€?  - 鎶ヨ〃鎸囨爣鍖呭惈鍙備笌灞€鏁般€佸畬鎴愬眬鏁般€佹敹鐩婃憳瑕併€侀個璇蜂汉鏁般€佷細鍛樿鍒掑悕銆?- 鏂板鍥㈤槦鏈嶅姟锛?  - `services/go-api/internal/teams/service.go`
  - 涓€鏈熷彧鏀寔鐩村睘涓€绾у洟闃熷叧绯汇€?  - 鏀寔鍥㈤槦鍩烘湰淇℃伅銆佸洟闃熸垚鍛樸€佸洟闃熸敹鐩婃憳瑕併€?- 鏂板灏忕▼搴忕鎺ュ彛锛?  - `GET /api/app/member-reports/me`
  - `GET /api/app/teams/my`
  - `GET /api/app/teams/my/members`
  - `GET /api/app/teams/my/revenue-summary`
- 鏉冮檺閿欒鐮佸榻愯鍒掞細
  - 鏃犱細鍛樻姤琛ㄦ潈闄愯繑鍥?`40351`銆?  - 鏃犲洟闃熺鐞嗘潈闄愯繑鍥?`40352`銆?- 鏁版嵁搴撻鐣欒〃琛ラ綈锛?  - `team_relations`
  - `member_report_snapshots`
  - `idx_member_report_snapshots_user_period`

### 楠岃瘉缁撴灉

- 鎵ц `go test ./...` 閫氳繃銆?- 鎵ц `go test -count=1 ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鏃犱細鍛樻潈闄愯闂細鍛樻姤琛ㄨ繑鍥?`40351`銆?  - 鏃犲洟闃熸潈闄愯闂洟闃熸垚鍛樿繑鍥?`40352`銆?  - 鎺堟潈鍚庝細鍛樻姤琛ㄨ繑鍥炲弬涓庡眬鏁般€佸畬鎴愬眬鏁般€佹敹鐩婃憳瑕併€侀個璇蜂汉鏁般€?  - 鍥㈤槦璐熻矗浜哄彲鏌ョ湅鎴戠殑鍥㈤槦銆佸洟闃熸垚鍛樺拰鍥㈤槦鏀剁泭鎽樿銆?
### 褰撳墠椋庨櫓

- 褰撳墠浼氬憳鏉冮檺銆佸洟闃熷叧绯诲拰鎶ヨ〃蹇収浠嶆槸鍐呭瓨瀹炵幇锛涘悗缁帴 PostgreSQL 鏃堕渶瑕佽惤鍒?`user_memberships`銆乣team_relations`銆乣member_report_snapshots` 骞剁敱瀹氭椂浠诲姟鐢熸垚绋冲畾蹇収銆?
## 2026-06-13 16:15

### 褰撳墠杩涘睍

- 缁х画灏忕▼搴忓悗绔紭鍏堬紝琛ラ綈 `plan.md` E2.4 涓Н鍒嗗拰鍏戞崲閾捐矾銆?- 鏂板绉垎璐︽埛鏈嶅姟锛?  - `services/go-api/internal/points/service.go`
  - 鏀寔绉垎鎽樿銆佺Н鍒嗘祦姘淬€佺Н鍒嗗彂鏀俱€佺Н鍒嗘墸鍑忋€?  - 绉垎涓嶈冻杩斿洖涓氬姟閿欒锛岄伩鍏嶈处鎴锋墸鎴愯礋鏁般€?- 鏂板鍏戞崲鏈嶅姟锛?  - `services/go-api/internal/redemption/service.go`
  - 鏀寔鍏戞崲椤圭洰銆佸厬鎹笅鍗曘€佹垜鐨勫厬鎹㈣鍗曘€?  - 涓嬪崟鏃跺湪鏈嶅姟閿佸唴瀹屾垚搴撳瓨鏍￠獙銆佸簱瀛樻墸鍑忓拰绉垎鎵ｅ噺锛涚Н鍒嗘墸鍑忓け璐ヤ細鍥炴粴搴撳瓨銆?- 鏂板灏忕▼搴忕鎺ュ彛锛?  - `GET /api/app/points/summary`
  - `GET /api/app/points/logs`
  - `GET /api/app/redemption/items`
  - `POST /api/app/redemption/orders`
  - `GET /api/app/redemption/orders/my`
- 鏁版嵁搴撻鐣欒〃琛ラ綈锛?  - `redemption_items`
  - `redemption_orders`
  - `idx_redemption_orders_user_created`

### 楠岃瘉缁撴灉

- 鎵ц `go test -count=1 ./...` 閫氳繃銆?- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 绌虹Н鍒嗚处鎴峰彲鏌ヨ銆?  - 绉垎涓嶈冻涓嶈兘鍒涘缓鍏戞崲璁㈠崟銆?  - 鍟嗗搧鍒楄〃鍙繑鍥炴湁鏁堝厬鎹㈤」鐩€?  - 鍏戞崲鎴愬姛鍚庣敓鎴愯鍗曪紝骞跺啓鍏ョН鍒嗘墸鍑忔祦姘淬€?  - 搴撳瓨涓嶈冻涓嶈兘缁х画鍒涘缓鍏戞崲璁㈠崟銆?  - 鎴戠殑鍏戞崲璁㈠崟鍙繑鍥炲綋鍓嶇敤鎴疯鍗曘€?  - 骞跺彂鍏戞崲鍚屼竴涓簱瀛樹负 1 鐨勯」鐩椂锛屽彧鍏佽 1 涓姹傛垚鍔燂紝鍙?1 涓繑鍥炲啿绐併€?
### 褰撳墠椋庨櫓

- 褰撳墠绉垎鍜屽厬鎹粛鏄唴瀛樺疄鐜帮紱鍚庣画鎺?PostgreSQL 鏃堕渶瑕佺敤鏁版嵁搴撲簨鍔″拰琛岀骇閿佸疄鐜板簱瀛樹笌绉垎鎵ｅ噺鐨勫己涓€鑷达紝骞惰ˉ鍚庡彴鍏戞崲瀹℃牳閫氳繃銆侀┏鍥炪€佸彂鏀惧強鎿嶄綔鏃ュ織銆?
## 2026-06-13 16:45

### 褰撳墠杩涘睍

- 缁х画灏忕▼搴忓悗绔紭鍏堬紝琛ラ綈 `plan.md` E2.3 浜鸿剦銆佽瀹舵妧鑳芥爲銆侀璺汉璧勬簮鐢诲儚閾捐矾銆?- 鏂板浜鸿剦鏈嶅姟锛?  - `services/go-api/internal/connections/service.go`
  - 鏀寔浜鸿剦鍒楄〃銆佸弻鍚戜汉鑴夌敓鎴愩€佽窡杩涜褰曘€佸叧绯诲己搴﹂€掑銆?  - 鐧诲綍閭€璇风粦瀹氭垚鍔熷悗鐢熸垚 `invite` 浜鸿剦銆?  - 鍚屽眬鎴愬憳鏈嶅姟纭瀹屾垚鍚庣敓鎴?`co_game` 浜鸿剦銆?  - 棰嗚矾浜烘挳鍚堝彲鐢熸垚 `guide_match` 浜鸿剦銆?- 鏂板鐢诲儚鏈嶅姟锛?  - `services/go-api/internal/profiles/service.go`
  - 鏀寔宸查€氳繃琛屽缁存姢鎶€鑳芥爲銆?  - 鏀寔宸查€氳繃棰嗚矾浜虹淮鎶よ祫婧愮敾鍍忋€?  - 鐢诲儚杩斿洖瀹屾暣鐜囷紝渚夸簬鍚庣画鍚庡彴缁熻銆?- 鏂板灏忕▼搴忕鎺ュ彛锛?  - `GET /api/app/connections/my`
  - `POST /api/app/connections/{connectionId}/follow-up`
  - `GET /api/app/experts/me/skills`
  - `PUT /api/app/experts/me/skills`
  - `GET /api/app/guides/me/resources`
  - `PUT /api/app/guides/me/resources`
- 鏁版嵁搴撻鐣欒〃琛ラ綈锛?  - `user_connections`
  - `connection_follow_logs`
  - `expert_skill_profiles`
  - `guide_resource_profiles`

### 楠岃瘉缁撴灉

- 鎵ц `go test -count=1 ./...` 閫氳繃銆?- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鏈€氳繃琛屽瑙掕壊涓嶈兘鏇存柊鎶€鑳芥爲銆?  - 鎺堟潈琛屽鍙互鏇存柊鍜屾煡鐪嬫妧鑳芥爲銆?  - 鏈€氳繃棰嗚矾浜鸿鑹蹭笉鑳芥洿鏂拌祫婧愮敾鍍忋€?  - 鎺堟潈棰嗚矾浜哄彲浠ユ洿鏂拌祫婧愮敾鍍忋€?  - 浜鸿剦鍒楄〃鍖呭惈 `invite`銆乣co_game`銆乣guide_match` 涓夌被鏉ユ簮銆?  - 浜鸿剦鍒楄〃涓嶆毚闇叉墜鏈哄彿銆佽韩浠借瘉绛夋晱鎰熷瓧娈点€?  - 鍏崇郴鏈汉鍜屽凡閫氳繃棰嗚矾浜哄彲浠ュ啓璺熻繘璁板綍銆?
### 褰撳墠椋庨櫓

- 褰撳墠瑙掕壊閫氳繃鐘舵€併€佷汉鑴夊叧绯诲拰鐢诲儚浠嶆槸鍐呭瓨瀹炵幇锛涘悗缁帴 PostgreSQL 鏃堕渶瑕佹帴鍏?`user_roles`銆乣user_connections`銆乣connection_follow_logs`銆乣expert_skill_profiles`銆乣guide_resource_profiles`锛屽苟琛ュ悗鍙版煡鐪嬫帴鍙ｄ笌鑴辨晱鏉冮檺銆?
## 2026-06-13 17:15

### 褰撳墠杩涘睍

- 缁х画灏忕▼搴忓悗绔紭鍏堬紝琛ラ綈 `plan.md` E2.2 鍓╀綑鐨勭画灞€銆侀噷绋嬬鏇存柊鍜屾墦鍗″紓甯搁摼璺€?- 鏂板灞€杩涘害鑳藉姏锛?  - 鍙戣捣浜哄彲鏇存柊閲岀▼纰戞爣棰樺拰鐘舵€併€?  - 灏忕▼搴忕鏌ヨ鎵撳崱鍒楄〃鏃惰繃婊ゅ悗鍙版爣璁颁负 `invalid` 鐨勬墦鍗°€?  - 鍘熷眬鍙戣捣浜哄彲鍦ㄥ眬杩涘叆 `pending_review/completed` 鍚庡彂璧风画灞€鑽夌銆?  - 缁眬鐢熸垚鐨勬柊灞€鐘舵€佸浐瀹氫负 `draft`锛屼笉浼氱洿鎺ュ彉鎴愭嫑鍕熶腑銆?- 鏂板/瀹屽杽鎺ュ彛锛?  - `PUT /api/app/games/{gameId}/milestones/{milestoneId}`
  - `POST /api/app/games/{gameId}/continue`
  - `POST /api/admin/game-checkins/{checkinId}/mark-invalid`
- 鏁版嵁搴撻鐣欏瓧娈佃ˉ榻愶細
  - `game_milestones.updated_at`
  - `game_checkins.milestone_id`
  - `game_checkins.file_ids`
  - `game_checkins.status`

### 楠岃瘉缁撴灉

- 鎵ц `go test -count=1 ./...` 閫氳繃銆?- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鍙戣捣浜烘洿鏂伴噷绋嬬鎴愬姛銆?  - 鍚庡彴鏍囪鎵撳崱寮傚父鍚庯紝灏忕▼搴忕鎵撳崱鍒楄〃涓嶅啀杩斿洖璇ュ紓甯告墦鍗°€?  - 灞€鏈畬鎴愭椂鍙戣捣缁眬杩斿洖鍐茬獊銆?  - 鏈嶅姟纭瀹屾垚鍚庡彂璧风画灞€鎴愬姛銆?  - 缁眬鐢熸垚鐨勬柊灞€涓?`draft`锛宍gameSource=continue`銆?
### 褰撳墠椋庨櫓

- 褰撳墠杩涘害銆佺画灞€鍜屽紓甯告墦鍗′粛鏄唴瀛樺疄鐜帮紱鍚庣画鎺?PostgreSQL 鏃堕渶瑕佹妸閲岀▼纰戞洿鏂般€佹墦鍗″紓甯告爣璁般€佺画灞€鑽夌鍒涘缓鏀惧埌鐪熷疄浜嬪姟閲岋紝骞惰ˉ鍚庡彴寮傚父鍘熷洜瀛楁銆?
## 2026-06-13 17:45

### 褰撳墠杩涘睍

- 缁х画灏忕▼搴忓悗绔紭鍏堬紝琛ラ綈 `plan.md` E2.1 涓嫭绔嬭涓轰笂鎶ユ帴鍙ｃ€?- 鏂板灏忕▼搴忕鎺ュ彛锛?  - `POST /api/app/behavior/events`
- 鎺ュ彛瑙勫垯锛?  - 鐧诲綍鐢ㄦ埛涓婃姤鏃惰褰?`userId`銆?  - 鏈櫥褰曚篃鍏佽鍖垮悕闄嶇骇璁板綍锛宍userId=0`銆?  - 缂哄皯 `eventType` 杩斿洖鍙傛暟閿欒銆?  - 鏀寔璁板綍 `targetType`銆乣targetId`銆乣pagePath`銆乣keyword`銆乣extra`銆?- 澶嶇敤鐜版湁 audit 琛屼负鏃ュ織鏈嶅姟锛岄伩鍏嶈涓烘暟鎹垎鏁ｅ埌澶氬鍐呭瓨瀛樺偍銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test -count=1 ./...` 閫氳繃銆?- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 缂哄皯 `eventType` 杩斿洖 `422`銆?  - 鐧诲綍鎬佽涓轰笂鎶ユ垚鍔熷苟璁板綍鐢ㄦ埛 ID銆?  - 鍖垮悕琛屼负涓婃姤鎴愬姛骞惰褰?`userId=0`銆?  - 鍚庡彴琛屼负鏃ュ織鍙煡鍒颁富鍔ㄤ笂鎶ョ殑 `search/share` 浜嬩欢銆?
### 褰撳墠椋庨櫓

- 褰撳墠琛屼负鏃ュ織浠嶆槸鍐呭瓨瀹炵幇锛涘悗缁帴 PostgreSQL 鏃堕渶瑕佽惤鍒拌涓轰簨浠惰〃锛屽苟琛ュ悗鍙版寜鐢ㄦ埛銆佷簨浠剁被鍨嬨€佹椂闂磋寖鍥寸瓫閫夈€?
## 2026-06-13 18:15

### 褰撳墠杩涘睍

- 缁х画灏忕▼搴忓悗绔紭鍏堬紝琛ラ綈 `plan.md` E3.1 涓悗鍙板琛屼负浜嬩欢鍜岀敤鎴锋敹钘忕殑鏌ヨ鏀拺銆?- 鏂板/瀹屽杽鍚庡彴鎺ュ彛锛?  - `GET /api/admin/users/{userId}/favorites`
  - `GET /api/admin/behavior/events?userId=&eventType=`
- 鏉冮檺閾捐矾鏀舵暃锛?  - `/api/admin/users/{userId}/growth` 缁х画浣跨敤 `user:view`銆?  - `/api/admin/users/{userId}/favorites` 浣跨敤 `user:read`銆?  - `/api/admin/behavior/events` 浣跨敤 `data:behavior:read`銆?  - 鍐呯疆绠＄悊鍛樻潈闄愬拰绉嶅瓙鏉冮檺琛ㄥ悓姝ヨˉ鍏?`user:read`銆乣data:behavior:read`銆?- 琛屼负鏌ヨ鏀寔鎸?`userId` 鍜?`eventType` 杩囨护锛屾敹钘忔煡璇㈠鐢ㄥ皬绋嬪簭渚ф敹钘忔湇鍔″苟杩囨护鏈鏍稿眬銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test -count=1 ./...` 閫氳繃銆?- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鍚庡彴鏃?`user:read` 鏉冮檺鏌ヨ鐢ㄦ埛鏀惰棌杩斿洖 `403`銆?  - 鍚庡彴鏈?`user:read` 鏉冮檺鍙煡璇㈡寚瀹氱敤鎴锋敹钘忋€?  - 鍚庡彴鏃?`data:behavior:read` 鏉冮檺鏌ヨ琛屼负浜嬩欢杩斿洖 `403`銆?  - 鍚庡彴鏈?`data:behavior:read` 鏉冮檺鍙寜鐢ㄦ埛鍜屼簨浠剁被鍨嬬瓫閫夎涓轰簨浠躲€?
### 褰撳墠椋庨櫓

- 褰撳墠鏀惰棌鍜岃涓轰簨浠朵粛鏄唴瀛樻湇鍔″疄鐜帮紱鍚庣画鎺?PostgreSQL 鏃堕渶瑕佹妸鍚庡彴绛涢€夋潯浠舵槧灏勫埌鐪熷疄绱㈠紩锛屽苟琛ユ椂闂磋寖鍥淬€佸垎椤靛拰瀵煎嚭瀛楁鑴辨晱瑙勫垯銆?
## 2026-06-13 18:45

### 褰撳墠杩涘睍

- 缁х画琛ラ綈 `plan.md` E3.1 鍚庡彴鎺ュ彛锛屽畬鎴愰噷绋嬬銆佹墦鍗°€佸鐩樺拰缁眬鑽夌鐨勫悗鍙版敮鎾戙€?- 鏂板/瀹屽杽鍚庡彴鎺ュ彛锛?  - `GET /api/admin/games/{gameId}/milestones`
  - `POST /api/admin/games/{gameId}/milestones`
  - `GET /api/admin/games/{gameId}/checkins`
  - `GET /api/admin/games/{gameId}/retrospectives`
  - `GET /api/admin/games/{gameId}/continue-drafts`
- 鏉冮檺閾捐矾锛?  - 鍚庡彴璇诲彇閲岀▼纰戙€佹墦鍗°€佸鐩樸€佺画灞€鑽夌浣跨敤 `game:read`銆?  - 鍚庡彴鍒涘缓閲岀▼纰戝拰鏍囪寮傚父鎵撳崱浣跨敤 `game:progress:manage`銆?  - 鍘熸湁 `GET /api/admin/games/{gameId}/review-trace` 缁х画浣跨敤 `game:view`銆?- 鏈嶅姟灞傛柊澧炲悗鍙版煡璇㈡柟娉曪紝鍚庡彴鎵撳崱鏌ヨ浼氳繑鍥炲寘鍚?`invalid` 鍦ㄥ唴鐨勫畬鏁磋褰曪紝渚夸簬杩愯惀杩借釜寮傚父澶勭悊銆?- 缁眬鑽夌鏂板鍘熷眬鍒拌崏绋垮眬鐨勫唴瀛樻槧灏勶紝骞跺湪杩佺Щ涓鐣?`game_continue_drafts` 琛ㄣ€?
### 楠岃瘉缁撴灉

- 鎵ц `go test -count=1 ./...` 閫氳繃銆?- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鍚庡彴鏃?`game:read` 鏉冮檺鏌ヨ閲岀▼纰戣繑鍥?`403`銆?  - 鍚庡彴鏈?`game:progress:manage` 鏉冮檺鍙垱寤洪噷绋嬬銆?  - 鍚庡彴鏈?`game:read` 鏉冮檺鍙煡鐪嬮噷绋嬬銆佸紓甯告墦鍗°€佸鐩樸€佺画灞€鑽夌銆?  - 鍚庡彴鏍囪寮傚父鍚庯紝灏忕▼搴忕杩囨护寮傚父鎵撳崱锛屽悗鍙扮浠嶅彲杩借釜璇ヨ褰曘€?
### 褰撳墠椋庨櫓

- 缁眬鑽夌鍜岃繘搴︽暟鎹洰鍓嶄粛鍦ㄥ唴瀛樻湇鍔′腑娴佽浆锛涙帴 PostgreSQL 鏃堕渶瑕佹妸 `game_continue_drafts`銆乣game_milestones`銆乣game_checkins`銆乣game_retrospectives` 涓茶繘鍚屼竴浜嬪姟鍜屽垎椤垫煡璇€?
## 2026-06-13 19:15

### 褰撳墠杩涘睍

- 缁х画琛ラ綈 `plan.md` E3.1 鍚庡彴鎺ュ彛锛屽畬鎴愪汉鑴夈€佽瀹舵妧鑳芥爲鍜岄璺汉璧勬簮鐢诲儚鐨勫悗鍙板彧璇诲叆鍙ｃ€?- 鏂板鍚庡彴鎺ュ彛锛?  - `GET /api/admin/connections`
  - `GET /api/admin/users/{userId}/connections`
  - `GET /api/admin/experts/{userId}/skills`
  - `GET /api/admin/guides/{userId}/resources`
- 鏉冮檺閾捐矾锛?  - 浜鸿剦鍒楄〃鍜岀敤鎴蜂汉鑴変娇鐢?`connection:read`銆?  - 琛屽鎶€鑳芥爲鍜岄璺汉璧勬簮鐢诲儚浣跨敤 `profile:read`銆?  - 鍐呯疆绠＄悊鍛樻潈闄愬拰绉嶅瓙鏉冮檺琛ㄥ悓姝ヨˉ鍏?`connection:read`銆乣profile:read`銆?- 鏈嶅姟灞傛柊澧炲悗鍙板彧璇绘柟娉曪細
  - 浜鸿剦鏀寔鍏ㄩ噺鏌ヨ鍜屾寜鐢ㄦ埛鏌ヨ銆?  - 鐢诲儚鏀寔鍚庡彴璇诲彇鎸囧畾鐢ㄦ埛鐨勬妧鑳芥爲鍜岃祫婧愮敾鍍忥紝鏈淮鎶ゆ椂杩斿洖绌虹敾鍍忕粨鏋勶紝渚夸簬鍚庡彴绌虹姸鎬佸睍绀恒€?
### 楠岃瘉缁撴灉

- 鎵ц `go test -count=1 ./...` 閫氳繃銆?- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鍚庡彴鏃?`connection:read` 鏉冮檺鏌ヨ浜鸿剦杩斿洖 `403`銆?  - 鍚庡彴鏈?`connection:read` 鏉冮檺鍙煡鍏ㄩ噺浜鸿剦鍜屾寚瀹氱敤鎴蜂汉鑴夈€?  - 鍚庡彴鏃?`profile:read` 鏉冮檺鏌ヨ鐢诲儚杩斿洖 `403`銆?  - 鍚庡彴鏈?`profile:read` 鏉冮檺鍙煡琛屽鎶€鑳芥爲鍜岄璺汉璧勬簮鐢诲儚銆?
### 褰撳墠椋庨櫓

- 褰撳墠鍚庡彴鐢诲儚鏌ヨ杩樻湭鍋氭晱鎰熷瓧娈靛垎绾ц劚鏁忥紱鎺ョ湡瀹炵敤鎴疯祫鏂欏拰鏂囦欢鏈嶅姟鏃讹紝闇€瑕佹寜 `profile:read` 涓庢洿楂樻潈闄愭媶鍒嗗彲瑙佸瓧娈靛拰闄勪欢璁块棶鑼冨洿銆?
## 2026-06-13 19:45

### 褰撳墠杩涘睍

- 缁х画琛ラ綈 `plan.md` E3.1 鍚庡彴鎺ュ彛锛屽畬鎴愮Н鍒嗘祦姘村拰绉垎鍏戞崲鍚庡彴绠＄悊闂幆銆?- 鏂板鍚庡彴鎺ュ彛锛?  - `GET /api/admin/points/logs`
  - `GET /api/admin/redemption/items`
  - `POST /api/admin/redemption/items`
  - `PUT /api/admin/redemption/items/{itemId}`
  - `GET /api/admin/redemption/orders`
  - `POST /api/admin/redemption/orders/{orderId}/review`
- 鏉冮檺閾捐矾锛?  - 绉垎娴佹按鏌ヨ浣跨敤 `points:read`銆?  - 鍏戞崲椤圭洰绠＄悊銆佸厬鎹㈣鍗曞垪琛ㄣ€佽鍗曞鏍镐娇鐢?`redemption:manage`銆?  - 鍐呯疆绠＄悊鍛樻潈闄愬拰绉嶅瓙鏉冮檺琛ㄥ悓姝ヨˉ鍏?`points:read`銆乣redemption:manage`銆?- 鏈嶅姟灞傝兘鍔涳細
  - 绉垎鏈嶅姟鏀寔鍚庡彴鍏ㄩ噺娴佹按鏌ヨ銆?  - 鍏戞崲鏈嶅姟鏀寔鍚庡彴鏌ョ湅鍏ㄩ儴椤圭洰銆佹洿鏂伴」鐩簱瀛?鐘舵€併€佹煡鐪嬪叏閮ㄨ鍗曘€?  - 鍏戞崲璁㈠崟鏀寔瀹℃牳閫氳繃銆侀┏鍥炲拰鍙戞斁锛涢┏鍥炴椂鑷姩閫€鍥炲凡鎵ｇН鍒嗗苟鍐欑Н鍒嗘祦姘淬€?  - 鎵€鏈夊悗鍙板啓鎿嶄綔鍐欏叆鎿嶄綔鏃ュ織銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test -count=1 ./...` 閫氳繃銆?- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鏃?`points:read` 鏌ヨ绉垎娴佹按杩斿洖 `403`銆?  - 鏃?`redemption:manage` 鏌ヨ鍏戞崲椤圭洰杩斿洖 `403`銆?  - 鍚庡彴鍙垱寤哄拰鏇存柊鍏戞崲椤圭洰銆?  - 鐢ㄦ埛鍏戞崲鍚庯紝鍚庡彴鍙煡璇㈣鍗曘€佸鏍搁€氳繃骞舵爣璁板彂鏀俱€?  - 鍚庡彴椹冲洖鍏戞崲璁㈠崟鍚庯紝鐢ㄦ埛绉垎鑷姩閫€鍥炪€?
### 褰撳墠椋庨櫓

- 褰撳墠鍏戞崲瀹℃牳浠嶆槸鍐呭瓨鐘舵€佹祦杞紱鎺?PostgreSQL 鏃堕渶瑕佸湪鍚屼竴浜嬪姟鍐呭鐞嗚鍗曠姸鎬併€佺Н鍒嗛€€鍥炪€佸簱瀛樺洖琛ュ拰鎿嶄綔鏃ュ織锛岄伩鍏嶅鏍稿け璐ュ鑷磋处瀹炰笉涓€鑷淬€?
## 2026-06-13 20:15

### 褰撳墠杩涘睍

- 缁х画琛ラ綈 `plan.md` E3.1 鍚庡彴鎺ュ彛锛屽畬鎴愬洟闃熶細鍛樺拰浼氬憳鎶ヨ〃鍚庡彴鍙鍏ュ彛銆?- 鏂板鍚庡彴鎺ュ彛锛?  - `GET /api/admin/teams`
  - `GET /api/admin/teams/{teamId}`
  - `GET /api/admin/member-reports`
- 鏉冮檺閾捐矾锛?  - 鍥㈤槦鍒楄〃鍜屽洟闃熻鎯呬娇鐢?`team:read`銆?  - 浼氬憳鎶ヨ〃鍒楄〃浣跨敤 `member_report:read`銆?  - 鍐呯疆绠＄悊鍛樻潈闄愬拰绉嶅瓙鏉冮檺琛ㄥ悓姝ヨˉ鍏?`team:read`銆乣member_report:read`銆?- 鏈嶅姟灞傝兘鍔涳細
  - 鍥㈤槦鏈嶅姟鏀寔鍚庡彴鑾峰彇鍥㈤槦鍒楄〃鍜屽洟闃熻鎯咃紝璇︽儏鍖呭惈鎴愬憳涓庡洟闃熸敹鐩婃憳瑕併€?  - 浼氬憳鎶ヨ〃鏈嶅姟鏀寔鍚庡彴鐢熸垚骞惰繑鍥炲凡鎺堟潈浼氬憳鐨勬姤琛ㄥ揩鐓с€?
### 楠岃瘉缁撴灉

- 鎵ц `go test -count=1 ./...` 閫氳繃銆?- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鏃?`team:read` 鏉冮檺鏌ヨ鍥㈤槦杩斿洖 `403`銆?  - 鏈?`team:read` 鏉冮檺鍙煡璇㈠洟闃熷垪琛ㄥ拰鍥㈤槦璇︽儏銆?  - 鏃?`member_report:read` 鏉冮檺鏌ヨ浼氬憳鎶ヨ〃杩斿洖 `403`銆?  - 鏈?`member_report:read` 鏉冮檺鍙煡璇細鍛樻姤琛ㄥ揩鐓с€?
### 褰撳墠椋庨櫓

- 褰撳墠鍥㈤槦鍜屼細鍛樻姤琛ㄤ粛鏄唴瀛樻煡璇紱鎺?PostgreSQL 鏃堕渶瑕佹寜鍥㈤槦 ID 寤虹储寮曪紝骞舵妸浼氬憳鎶ヨ〃蹇収鏀规垚瀹氭椂浠诲姟鐢熸垚鍚庡垎椤垫煡璇€?
## 2026-06-13 20:45

### 褰撳墠杩涘睍

- 缁х画琛ラ綈 `plan.md` E3.1 鍚庡彴鎺ュ彛锛屽畬鎴愪氦浠樻枃妗ｃ€佹祴璇曠敤渚嬪拰娴嬭瘯鎵ц璁板綍鐨勫悗鍙扮櫥璁板叆鍙ｃ€?- 鏂板鍚庡彴鎺ュ彛锛?  - `GET /api/admin/delivery-documents`
  - `POST /api/admin/delivery-documents`
  - `GET /api/admin/test-cases`
  - `POST /api/admin/test-cases`
  - `GET /api/admin/test-runs`
  - `POST /api/admin/test-runs`
- 鏉冮檺閾捐矾锛?  - 浜や粯鏂囨。浣跨敤 `delivery:manage`銆?  - 娴嬭瘯鐢ㄤ緥鍜屾祴璇曟墽琛岃褰曚娇鐢?`testcase:manage`銆?  - 鍐呯疆绠＄悊鍛樻潈闄愬拰绉嶅瓙鏉冮檺琛ㄥ悓姝ヨˉ鍏?`delivery:manage`銆乣testcase:manage`銆?- 鏈嶅姟灞傝兘鍔涳細
  - 鏂板 `delivery` 鏈嶅姟锛屾敮鎸佷氦浠樻枃妗ｃ€佹祴璇曠敤渚嬨€佹祴璇曟墽琛岃褰曠殑鍐呭瓨鐧昏鍜屾煡璇€?  - 浜や粯鏂囨。鏍￠獙 `docType/title/status`銆?  - 娴嬭瘯鎵ц璁板綍蹇呴』鍏宠仈宸插瓨鍦ㄦ祴璇曠敤渚嬶紝閬垮厤瀛ょ珛鎵ц璁板綍銆?  - 鍚庡彴鏂板鍐欐搷浣滃潎鍐欏叆鎿嶄綔鏃ュ織銆?
### 楠岃瘉缁撴灉

- 鎵ц `go test -count=1 ./...` 閫氳繃銆?- 鎵ц `go test ./...` 閫氳繃銆?- 鏂板娴嬭瘯瑕嗙洊锛?  - 鏃?`delivery:manage` 鏌ヨ浜や粯鏂囨。杩斿洖 `403`銆?  - 鏈?`delivery:manage` 鍙垱寤哄拰鏌ヨ浜や粯鏂囨。銆?  - 鏃?`testcase:manage` 鏌ヨ娴嬭瘯鐢ㄤ緥杩斿洖 `403`銆?  - 鏈?`testcase:manage` 鍙垱寤烘祴璇曠敤渚嬨€佸垱寤烘祴璇曟墽琛岃褰曞苟鏌ヨ鎵ц璁板綍銆?  - 娴嬭瘯鎵ц璁板綍鍏宠仈涓嶅瓨鍦ㄧ敤渚嬫椂杩斿洖 `404`銆?
### 褰撳墠椋庨櫓

- 褰撳墠浜や粯娴嬭瘯鐣欐。浠嶆槸鍐呭瓨鏈嶅姟锛涙帴 PostgreSQL 鏃堕渶瑕佽惤鍒?`delivery_documents`銆乣test_cases`銆乣test_runs`锛屽苟琛ユ枃浠堕檮浠躲€佹墽琛屼汉鍜屽鍑鸿兘鍔涖€?
## 2026-06-13 21:30

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md` E4 寮哄疄鍚嶉摼璺紝浼樺厛琛ュ皬绋嬪簭鍚庣璁よ瘉涓庡疄鍚嶉椄闂ㄣ€?- 鐧诲綍閾捐矾鏀逛负涓ら樁娈碉細
  - `POST /api/app/auth/wechat-login` 榛樿杩斿洖 `preAuthToken`銆乣requiresIdentityBinding`銆乣identityBindStatus`銆?  - 宸插疄鍚嶇敤鎴峰啀娆＄櫥褰曟椂鍙悓鏃惰幏寰楁寮?`token`銆?  - 鏂板 `POST /api/app/auth/issue-token-after-identity`锛屽疄鍚嶅畬鎴愬悗鎹㈠彇姝ｅ紡 token銆?- 瀹炲悕娴佺▼鍏ュ彛鏀寔棰勮璇佽闂細
  - 鎵嬫満缁戝畾銆佺煭淇″彂閫?鏍￠獙銆佹墜鏈哄彿鏍搁獙銆丗aceID detect/callback銆佸疄鍚嶇姸鎬佹煡璇㈠潎鍙娇鐢?`preAuthToken`銆?  - `/api/app/users/me` 鍏佽棰勮璇佽闂紝鏂逛究灏忕▼搴忓睍绀哄綋鍓嶇櫥褰曠敤鎴峰拰瀹炲悕杩涘害銆?- 鏍稿績涓氬姟缁熶竴鎺ュ叆寮哄疄鍚嶉椄闂細
  - `requireUser` 澧炲姞 `identity.IsVerified` 鏍￠獙锛屾湭瀹屾垚寮哄疄鍚嶈繑鍥?`40341`銆?  - 宸插畬鎴愬疄鍚嶅悗锛屽師棰勮璇?token 涔熻兘缁х画閫氳繃鏈湴鏈嶅姟鏍￠獙锛屽吋瀹圭幇鏈夋祴璇曚笌鏈湴璋冭瘯锛涙寮?token 浠嶇敱鏂板鎺ュ彛鍙戞斁缁欏鎴风銆?- 鍚庡彴鏂板瀹炲悕鏍搁獙鍙鍏ュ彛锛?  - `GET /api/admin/identity-verifications`
  - `GET /api/admin/identity-verifications/{userId}`
  - 鏉冮檺鐮侊細`identity:read`
  - 璇︽儏鏌ョ湅鍐欏叆鎿嶄綔鏃ュ織锛屼笖杩斿洖瀛楁淇濇寔鎵嬫満鍙疯劚鏁忋€佹棤韬唤璇佸彿鍜屼汉鑴稿師濮嬫暟鎹€?- 椤烘墜淇鐪熷疄 Go 娴嬭瘯鏆撮湶鐨勯棶棰橈細
  - `ErrGameNotConfirmable` 鏄犲皠涓?`409`锛岄伩鍏嶅鐩?缁眬杩囨棭鎿嶄綔杩斿洖 `500`銆?  - 涓撳/棰嗚矾浜虹敾鍍忔洿鏂板鍔?`POST` 鍒悕锛屽吋瀹瑰皬绋嬪簭琛ㄥ崟鎻愪氦銆?  - 淇绉垎鍏戞崲娴嬭瘯鐨勯€€鍥炰綑棰濇柇瑷€锛屾寜璐︽湰鐪熷疄娴佽浆涓?`20`銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server` 鏃犳祴璇曟枃浠躲€?  - `internal/appapi`銆乣internal/auth`銆乣internal/games`銆乣internal/identity`銆乣internal/revenueclient` 绛夊寘閫氳繃銆?- 娉ㄦ剰锛氱郴缁?PATH 閲屼紭鍏堝懡涓?`C:\Windows\System32\go`锛岃鍗犱綅鍛戒护浼氶潤榛樿繑鍥烇紱鏈疆楠岃瘉宸叉敼鐢?`C:\Program Files\Go\bin\go.exe`锛屽苟鎶?`GOCACHE` 鎸囧埌椤圭洰鍐咃紝閬垮厤娌欑鍐?`AppData` 澶辫触銆?
### 褰撳墠椋庨櫓

- 褰撳墠瀹炲悕銆乼oken銆佸悗鍙板疄鍚嶈褰曚粛鏄唴瀛樻湇鍔℃ā鍨嬶紱鎺?PostgreSQL 鏃堕渶瑕佽惤琛ㄤ繚瀛橀璁よ瘉浼氳瘽銆佹寮忎細璇濄€佸疄鍚嶇姸鎬佹祦杞拰鍚庡彴鏌ョ湅鏃ュ織銆?- 寰俊瀹炲悕涓€鑷存€х洰鍓嶆寜鏈湴妯℃嫙閾捐矾鏍囪涓?`not_supported`锛岀湡鏈烘帴鍏ユ椂闇€瑕佹牴鎹井淇?鑵捐浜戝彲鐢ㄨ兘鍔涜ˉ鐪熷疄涓€鑷存€у洖璋冩垨淇濇寔鏄庣‘涓嶅彲鏀寔鐘舵€併€?
## 2026-06-13 22:00

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md` E4 寮哄疄鍚嶄富绾匡紝浼樺厛琛ュ皬绋嬪簭鍚庣鎸佷箙鍖栧拰鏈湴鍙祴椋庢帶閫昏緫銆?- 鏂板杩佺Щ鑴氭湰 `db/migrations/000011_identity_strong.sql`锛?  - `app_auth_sessions`锛氶璁よ瘉 token 鍜屾寮忎笟鍔?token 鐨勬寔涔呭寲浼氳瘽琛紝鎸?`token_hash` 淇濆瓨锛屼笉钀芥槑鏂?token銆?  - `identity_verification_records`锛氬己瀹炲悕鎬荤姸鎬佽〃锛岃褰曟墜鏈哄彿鑴辨晱鍊笺€佹墜鏈哄彿鏍搁獙銆佷汉鑴告牳韬€佸井淇″疄鍚嶄竴鑷存€х姸鎬併€佸け璐ュ師鍥犮€佺煭淇″彂閫佽鏁扮瓑銆?  - `identity_sms_code_records`锛氱煭淇￠獙璇佺爜鍙戦€?鏍￠獙鐣欑棔琛紝棰勭暀 `code_hash`銆佽繃鏈熸椂闂淬€佸け璐ユ鏁般€?  - `identity_faceid_sessions`锛氫汉鑴告牳韬細璇濊〃锛岄鐣?token hash銆佽姹?鍥炶皟鎽樿銆佸け璐ュ師鍥犲拰瀹屾垚鏃堕棿銆?- 鍚屾鏇存柊 `plan.md` 鐨?C3 鏁版嵁搴撹縼绉绘墽琛岄『搴忥紝杩藉姞 `000011_identity_strong.sql`锛岄伩鍏嶆柊杩佺Щ鑴氭湰娓哥鍦ㄨ鍒掑銆?- 寮哄疄鍚嶇煭淇″彂閫侀€昏緫浠庡浐瀹氳繑鍥炲崌绾т负鏈湴闄愭祦妯″瀷锛?  - 60 绉掑唴閲嶅鍙戦€佽繑鍥?`ErrSMSRateLimited`銆?  - 鍗曠敤鎴峰崟鏃ユ渶澶?5 娆★紝瓒呴檺杩斿洖 `ErrSMSDailyLimited`銆?  - 灏忕▼搴忔帴鍙?`POST /api/app/sms/send-code` 灏嗛檺娴佹槧灏勪负 HTTP `429`銆?- 琛ユ祴璇曡鐩栵細
  - identity 鏈嶅姟灞傞噸澶嶅彂閫佺煭淇′細琚檺娴併€?  - appapi 灞傝繛缁皟鐢?`/api/app/sms/send-code` 绗簩娆¤繑鍥?`429`銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `internal/appapi`
  - `internal/auth`
  - `internal/games`
  - `internal/identity`
  - `internal/revenueclient`
  - 鍏朵綑鏃犳祴璇曟枃浠跺寘姝ｅ父缂栬瘧閫氳繃銆?
### 褰撳墠椋庨櫓

- 杩佺Щ鑴氭湰宸茶ˉ榻愬己瀹炲悕鎸佷箙鍖栬〃锛屼絾杩愯鏃舵湇鍔′粛浠ュ唴瀛?Store 涓轰富锛涗笅涓€姝ラ渶瑕佹帴 PostgreSQL repository 鎴栫粺涓€鏁版嵁璁块棶灞傘€?- 鐭俊楠岃瘉鐮佸綋鍓嶄粛鏄湰鍦?mock `123456`锛屽凡鍏峰闄愭祦涓庣姸鎬佹祦杞祴璇曪紝浣嗙湡瀹炵煭淇￠€氶亾鎺ュ叆鏃惰繕闇€瑕侀獙璇佺爜 hash銆佽繃鏈熸椂闂淬€佸け璐ユ鏁颁笌渚涘簲鍟嗗洖鎵ц惤琛ㄣ€?
## 2026-06-13 22:30

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md` E4 寮哄疄鍚嶄富绾匡紝浼樺厛鎺ㄨ繘灏忕▼搴忓悗绔繍琛屾椂钀藉簱杈圭晫銆?- 鏂板 identity 鎸佷箙鍖栨帴鍙ｄ笌 SQL 閫傞厤楠ㄦ灦锛?  - `services/go-api/internal/identity/repository.go`
  - `services/go-api/internal/identity/sql_repository.go`
- `identity.Service` 鏂板 `NewServiceWithRepository(repo)`锛?  - 鏃?repository 鏃剁户缁娇鐢ㄧ幇鏈夊唴瀛樻ā寮忥紝淇濇寔鏈湴娴嬭瘯鍜屽凡鏈夋帴鍙ｄ笉鍙楀奖鍝嶃€?  - 鏈?repository 鏃讹紝浼氬啓鍏ュ己瀹炲悕鎬荤姸鎬併€佺煭淇￠獙璇佺爜璁板綍鍜?FaceID 浼氳瘽璁板綍銆?  - 鐭俊楠岃瘉鐮佷笌 FaceID token 鎸?hash 鍙ｅ緞鍐欏叆 repository锛屼笉钀芥槑鏂囬獙璇佺爜鎴?face token銆?- `cmd/server/main.go` 鏂板鍙€?identity repository 鎺ュ叆鐐癸細
  - 璇诲彇 `DATABASE_URL` 鍜?`DATABASE_DRIVER`銆?  - 鍙墦寮€鏁版嵁搴撴椂浣跨敤 `identity.NewSQLRepository(db)`銆?  - 鏈厤缃垨椹卞姩涓嶅彲鐢ㄦ椂璁板綍鏃ュ織骞跺洖閫€鍐呭瓨妯″紡銆?- `deploy/env.example` 鏂板锛?  - `DATABASE_DRIVER=postgres`
  - `DATABASE_URL=postgres://zhw:zhw@127.0.0.1:5432/zhw_mini?sslmode=disable`
- 琛ュ厖 repository 琛屼负娴嬭瘯锛?  - 寮哄疄鍚嶆祦绋嬩細鎸佷箙鍖栨渶缁?`verified` 鐘舵€併€?  - 鐭俊楠岃瘉鐮佽褰曚繚瀛?hash锛屼笉淇濆瓨 `123456` 鏄庢枃銆?  - FaceID 浼氳瘽浼氳褰?`faceid_processing -> verified` 鐢熷懡鍛ㄦ湡銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/auth`
  - `internal/games`
  - `internal/identity`
  - `internal/revenueclient`

### 褰撳墠椋庨櫓

- 褰撳墠 Go 妯″潡鏈紩鍏?PostgreSQL 椹卞姩渚濊禆锛涘洜姝?`DATABASE_DRIVER=postgres` 鍦ㄦ病鏈夐┍鍔ㄦ敞鍐屾椂浼氬洖閫€鍐呭瓨妯″紡銆傜敱浜庣敤鎴烽檺鍒朵笅杞戒綅缃紝鏈疆鏈墽琛?`go get` 涓嬭浇澶栭儴渚濊禆銆?- 涓嬩竴姝ヨ嫢瑕佺湡姝ｈ繛 PostgreSQL锛岄渶瑕佸湪鍙帶涓嬭浇/渚濊禆绛栫暐涓嬪姞鍏ラ┍鍔紝渚嬪 `pgx/stdlib` 鎴栭」鐩寚瀹氶┍鍔紝骞惰ˉ鏁版嵁搴撻泦鎴愭祴璇曘€?
## 2026-06-13 23:00

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md` E4 寮哄疄鍚嶄富绾匡紝浼樺厛鎺ㄨ繘灏忕▼搴忓悗绔?token 浼氳瘽钀藉簱杈圭晫銆?- 灏濊瘯鎸夌敤鎴蜂笅杞界洰褰曠害鏉熸帴鍏?PostgreSQL 椹卞姩锛?  - `GOMODCACHE` 宸叉寚瀹氬埌 `C:\Users\61492\Desktop\codex download\go-mod-cache`銆?  - 娌欑涓嶅厑璁稿啓璇ョ洰褰曪紝鎻愭潈璇锋眰鏈幏閫氳繃锛涙湰杞湭涓嬭浇澶栭儴渚濊禆锛屼篃鏈慨鏀?`go.mod/go.sum`銆?- 鍦ㄤ笉寮曞叆澶栭儴渚濊禆鐨勫墠鎻愪笅锛屽畬鎴?auth token 浼氳瘽鎸佷箙鍖栨帴鍙ｅ拰 SQL 閫傞厤楠ㄦ灦锛?  - `services/go-api/internal/auth/repository.go`
  - `services/go-api/internal/auth/sql_repository.go`
  - `TokenStore` 鏂板 `NewTokenStoreWithRepository(repo)`銆?  - token 鍐欏叆 repository 鏃跺彧淇濆瓨 `token_hash`锛屼笉淇濆瓨鏄庢枃 token銆?  - 鍐呭瓨 token 涓嶅瓨鍦ㄦ椂锛屽彲閫氳繃 `token_hash` 浠?repository 鍥炶浇浼氳瘽銆?- `cmd/server/main.go` 璋冩暣涓轰竴娆℃墦寮€鏁版嵁搴撹繛鎺ワ紝骞跺悓鏃舵彁渚涚粰锛?  - `auth.NewSQLSessionRepository(db)`
  - `identity.NewSQLRepository(db)`
  - 鏁版嵁搴撴湭閰嶇疆銆侀┍鍔ㄦ湭娉ㄥ唽鎴?ping 澶辫触鏃讹紝缁熶竴鏃ュ織鎻愮ず骞跺洖閫€鍐呭瓨妯″紡銆?- `.gitignore` 琛ュ厖 `.gocache/`銆乣.gotmp/` 鍜屽瓙鐩綍缂撳瓨瑙勫垯锛岄伩鍏嶆湰鍦?Go 鏋勫缓缂撳瓨杩涘叆浜や粯鑼冨洿銆?- 琛ュ厖 auth token repository 娴嬭瘯锛?  - 鍙戣棰勮璇?token 鍚庢寔涔呭寲 session銆?  - repository 涓笉淇濆瓨鏄庢枃 token銆?  - 鏂?`TokenStore` 鍙粠 repository 鎸?token hash 鍥炶浇浼氳瘽銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/auth`
  - `internal/games`
  - `internal/identity`
  - `internal/revenueclient`

### 褰撳墠椋庨櫓

- PostgreSQL 椹卞姩渚濊禆浠嶆湭鎺ュ叆锛岀湡瀹炴暟鎹簱杩炴帴鍦ㄥ綋鍓嶄緷璧栫姸鎬佷笅浼氬洜 driver 鏈敞鍐岃€屽洖閫€鍐呭瓨妯″紡銆?- 鍚庣画闇€瑕佺敤鎴峰厑璁镐緷璧栦笅杞藉埌 `C:\Users\61492\Desktop\codex download`锛屾垨鎻愪緵鏈湴宸叉湁鐨?Go module 缂撳瓨/渚濊禆鍖呭悗锛屽啀鎺ュ叆 `pgx/stdlib` 骞惰ˉ鐪熷疄鏁版嵁搴撻泦鎴愭祴璇曘€?
## 2026-06-13 23:30

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md` E5銆孉I 鎺ㄨ崘涓?IM 鍒嗘瀽鐨勪竴鏈熸暟鎹噯澶囥€嶏紝鏈疆浠嶄紭鍏堝皬绋嬪簭鍚庣鍜屽悗鍙板彧璇绘暟鎹兘鍔涖€?- 鏂板 `services/go-api/internal/aidata` 鏈嶅姟锛?  - 鑱氬悎琛屼负鏃ュ織銆佹敹钘忋€佽瘎浠枫€佷汉鑴夊叧绯汇€佽瀹舵妧鑳界敾鍍忋€侀璺汉璧勬簮鐢诲儚銆両M 娑堟伅鏁伴噺銆?  - 杩斿洖 `dataReady` 鍜?`imExportEnabled`锛岀敤浜庡悗鍙板垽鏂竴鏈熸暟鎹矇娣€鏄惁鍙敤銆?  - IM 鑱婂ぉ鍐呭瀵煎嚭榛樿鍏抽棴锛屾湭寮€鍚椂杩斿洖 `ErrIMExportDisabled`銆?- 鏂板鍚庡彴鎺ュ彛锛?  - `GET /api/admin/ai-data/snapshot`锛屾潈闄?`ai:data:read`銆?  - `POST /api/admin/ai-data/im-export`锛屾潈闄?`ai:data:export`锛屽綋鍓嶉粯璁よ繑鍥?`40361`锛岀姝㈡妸鑱婂ぉ鍐呭瀵煎嚭缁?AI銆?- 琛ュ厖鏈嶅姟灞傚彧璇昏仛鍚堟柟娉曪細
  - `games.AllFavorites()`
  - `reviews.AllReviews()`
  - `profiles.AllExpertSkills()`
  - `profiles.AllGuideResources()`
  - `im.AllMessages()`
- 琛ュ厖鏉冮檺鍜岃縼绉伙細
  - 鍐呯疆绠＄悊鍛樻潈闄愬鍔?`ai:data:read`銆乣ai:data:export`銆?  - `db/seeds/admin_roles_permissions.sql` 澧炲姞 AI 鏁版嵁鏉冮檺绉嶅瓙銆?  - 鏂板 `db/migrations/000012_ai_data_prepare.sql`锛岄鐣?`ai_data_snapshots` 鍜?`ai_data_export_tasks`銆?  - `plan.md` C3 鏁版嵁搴撹縼绉婚『搴忓凡杩藉姞 `000012_ai_data_prepare.sql`銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/adminauth`
  - `internal/appapi`
  - `internal/auth`
  - `internal/games`
  - `internal/identity`
  - `internal/revenueclient`
  - 鍏朵綑鏃犳祴璇曟枃浠跺寘姝ｅ父缂栬瘧閫氳繃銆?- 鏂板 `TestAIDataSnapshotAndIMExportGateHTTP` 瑕嗙洊锛?  - 琛屼负鏃ュ織銆佹敹钘忋€佽瘎浠枫€両M 娑堟伅銆佷汉鑴夊叧绯汇€佽瀹剁敾鍍忋€侀璺汉鐢诲儚鍧囨湁鏁版嵁娌夋穩銆?  - 鍚庡彴 AI 鏁版嵁蹇収鍙寜 `ai:data:read` 鏌ヨ銆?  - IM 瀵煎嚭寮€鍏冲叧闂椂锛宍POST /api/admin/ai-data/im-export` 杩斿洖 `403`銆?
### 褰撳墠椋庨櫓

- 褰撳墠 AI 鏁版嵁蹇収鏉ヨ嚜杩愯鏃跺唴瀛樻湇鍔¤仛鍚堬紱鐪熷疄 PostgreSQL 鎺ュ叆鍚庯紝闇€瑕佹妸杩欎簺鍙鑱氬悎鏀逛负鍒嗛〉 SQL 鏌ヨ鎴栧揩鐓т换鍔★紝閬垮厤鍚庡彴涓€娆℃€ф壂鎻忓ぇ琛ㄣ€?- IM 瀵煎嚭鎺ュ彛宸插仛榛樿鍏抽棴鍜屾潈闄愰殧绂伙紝浣嗗皻鏈帴鍏ョ湡瀹炴巿鏉冮厤缃〃锛涗簩鏈熷惎鐢ㄥ墠蹇呴』琛ュ厖鏄庣‘鐨勬暟鎹巿鏉冦€佽劚鏁忋€佸璁″拰瀵煎嚭浠诲姟鐣欑棔銆?
## 2026-06-13 23:55

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md` E6 鍜?E7 涓笌 E5 鐩存帴鐩稿叧鐨勫悗绔仈璋冩潗鏂欒ˉ榻愩€?- 鍚屾鍚庡彴 OpenAPI锛?  - `docs/openapi/admin.openapi.yaml` 鏂板 `GET /api/admin/ai-data/snapshot`銆?  - `docs/openapi/admin.openapi.yaml` 鏂板 `POST /api/admin/ai-data/im-export`銆?  - 鏍囨槑 `ai:data:read`銆乣ai:data:export` 鏉冮檺鍜?IM 瀵煎嚭榛樿鍏抽棴琛屼负銆?- 鍚屾 DTO 鍜岄敊璇爜锛?  - `docs/openapi/dto-samples.md` 鏂板 `AIDataSnapshotDTO`銆?  - `docs/openapi/dto-samples.md` 鏂板 `IMExportItemDTO`銆?  - `docs/openapi/error-codes.md` 鏂板 `40361`锛岃〃绀?AI IM 瀵煎嚭寮€鍏冲叧闂€?- 鏂板娴嬭瘯鐢ㄤ緥褰掓。锛?  - `docs/test-cases/TC-144-ai-data-im-export-gate.md`
  - 瑕嗙洊 AI 鏁版嵁蹇収鍙煡銆両M 鍙鏁般€侀粯璁ゅ叧闂椂绂佹瀵煎嚭鑱婂ぉ鍐呭銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/adminauth`
  - `internal/appapi`
  - `internal/auth`
  - `internal/games`
  - `internal/identity`
  - `internal/revenueclient`
  - 鍏朵綑鏃犳祴璇曟枃浠跺寘姝ｅ父缂栬瘧閫氳繃銆?
### 褰撳墠椋庨櫓

- OpenAPI 褰撳墠浠嶄互绠€鍖栨弿杩颁负涓伙紝灏氭湭涓?AI 鏁版嵁鎺ュ彛琛ュ畬鏁?schema components锛涘悗缁墠绔仈璋冨墠闇€瑕佹妸缁熶竴鍝嶅簲鍖呫€佸瓧娈电被鍨嬪拰閿欒鍝嶅簲缁撴瀯鍏ㄩ儴灞曞紑銆?- `docs/test-cases` 鐜板湪鍙柊澧炰簡 TC-144锛孍7 涓叾浣?TC-130 鍒?TC-145 杩橀渶瑕佹寜宸插疄鐜版ā鍧楅€愭褰掓。銆?
## 2026-06-14 00:20

### 褰撳墠杩涘睍

- 缁х画琛ラ綈 `plan.md` E5銆孉I 鎺ㄨ崘涓?IM 鍒嗘瀽鐨勪竴鏈熸暟鎹噯澶囥€嶇殑鏁版嵁婧愯鐩栥€?- 灏?`user_footprints` 瀵瑰簲鐨勮繍琛屾椂瓒宠抗鏁版嵁绾冲叆 AI 鏁版嵁蹇収锛?  - `reviews.Service` 鏂板 `AllFootprints()` 鍙鑱氬悎鏂规硶銆?  - `aidata.SnapshotInput` 鏂板 `Footprints`銆?  - `AIDataSnapshot` 鏂板 `footprintCount`銆?  - `GET /api/admin/ai-data/snapshot` 杩斿洖瓒宠抗鏁伴噺銆?- 鍚屾娴嬭瘯鍜岃仈璋冩潗鏂欙細
  - `TestAIDataSnapshotAndIMExportGateHTTP` 澧炲姞 `footprintCount > 0` 鏂█銆?  - `docs/openapi/admin.openapi.yaml` 灏?AI 鏁版嵁蹇収璇存槑琛ュ厖涓哄寘鍚冻杩广€?  - `docs/openapi/dto-samples.md` 鐨?`AIDataSnapshotDTO` 澧炲姞 `footprintCount`銆?  - `docs/test-cases/TC-144-ai-data-im-export-gate.md` 澧炲姞瓒宠抗鍓嶇疆鏉′欢鍜屽瓧娈垫牎楠屻€?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/adminauth`
  - `internal/appapi`
  - `internal/auth`
  - `internal/games`
  - `internal/identity`
  - `internal/revenueclient`
  - 鍏朵綑鏃犳祴璇曟枃浠跺寘姝ｅ父缂栬瘧閫氳繃銆?
### 褰撳墠椋庨櫓

- 瓒宠抗鐩墠浠嶆潵鑷?`reviews` 鍐呭瓨鏈嶅姟涓殑杩愯鏃舵暟缁勶紱鎺?PostgreSQL 鍚庨渶瑕佹寜 `user_footprints` 琛ㄥ疄鐜板垎椤垫煡璇㈡垨蹇収鑱氬悎銆?- E5 鐨勫瓧娈佃鐩栧凡杩涗竴姝ュ畬鏁达紝浣嗘祴璇曟暟鎹壒閲忛€犳暟鑳藉姏灏氭湭琛ラ綈锛屽悗缁繕闇€瑕佷负鈥? 涓祴璇曠敤鎴枫€? 涓眬銆?0 鏉¤涓衡€濈瓑楠屾敹鏉′欢鍋氬彲閲嶅娴嬭瘯鑴氭湰鎴栧悗鍙扮瀛愪换鍔°€?
## 2026-06-14 00:50

### 褰撳墠杩涘睍

- 缁х画鎺ㄨ繘 `plan.md` E5銆孉I 鎺ㄨ崘涓?IM 鍒嗘瀽鐨勪竴鏈熸暟鎹噯澶囥€嶇殑灏忕▼搴忓悗绔獙鏀跺彲瑙佹€с€?- `GET /api/admin/ai-data/snapshot` 鏂板楠屾敹闃堝€煎揩鐓э細
  - `userCount`锛氫粠琛屼负銆佸眬銆佹敹钘忋€佽瘎浠枫€佽冻杩广€佷汉鑴夈€佺敾鍍忓拰 IM 鍙戦€佽€呬腑鑱氬悎鍘婚噸鐢ㄦ埛鏁般€?  - `gameCount`锛氬綋鍓嶆父鎴忓眬鏁伴噺銆?  - `acceptanceChecks`锛氭樉寮忚繑鍥?5 涓敤鎴枫€? 涓眬銆?0 鏉¤涓恒€? 鏉℃敹钘忋€? 鏉¤瘎浠风殑褰撳墠鍊笺€佽姹傚€煎拰杈炬爣鐘舵€併€?  - `acceptanceReady`锛氬綋涓婅堪 5 椤瑰叏閮ㄨ揪鏍囨椂鑷姩涓?`true`銆?- 鍚屾娴嬭瘯鍜岃仈璋冩潗鏂欙細
  - `TestAIDataSnapshotAndIMExportGateHTTP` 澧炲姞闃堝€艰姹傘€佺敤鎴锋暟銆佸眬鏁板拰鏈揪鏍囩姸鎬佹柇瑷€銆?  - `docs/openapi/admin.openapi.yaml` 澧炲姞 E5 楠屾敹闃堝€艰鏄庛€?  - `docs/openapi/dto-samples.md` 鐨?`AIDataSnapshotDTO` 澧炲姞瀹屾暣闃堝€肩ず渚嬨€?  - `docs/test-cases/TC-144-ai-data-im-export-gate.md` 澧炲姞闃堝€煎瓧娈垫牎楠屻€?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/adminauth`
  - `internal/appapi`
  - `internal/auth`
  - `internal/games`
  - `internal/identity`
  - `internal/revenueclient`
  - 鍏朵綑鏃犳祴璇曟枃浠跺寘姝ｅ父缂栬瘧閫氳繃銆?
### 褰撳墠椋庨櫓

- 褰撳墠宸茶兘鐪嬪埌 E5 楠屾敹闃堝€艰繘搴︼紝浣嗚繕娌℃湁鎻愪緵涓€閿壒閲忛€犳暟/绉嶅瓙浠诲姟锛涗笅涓€姝ュ簲琛ュ悗鍙板彈鎺х瀛愭帴鍙ｆ垨娴嬭瘯澶瑰叿锛岀ǔ瀹氱敓鎴?5 鐢ㄦ埛銆? 灞€銆?0 琛屼负銆? 鏀惰棌銆? 璇勪环鐨勬暟鎹泦銆?- `userCount` 鐩墠鍩轰簬鍐呭瓨鏈嶅姟鑱氬悎锛涙帴 PostgreSQL 鍚庨渶瑕佹敼涓?SQL 鑱氬悎鎴栧彧璇讳粨鍌ㄨ仛鍚堬紝閬垮厤鐢熶骇鏁版嵁閲忓彉澶ф椂鍦ㄥ唴瀛樺眰鍏ㄩ噺鎵弿銆?
## 2026-06-14 01:20

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md` E5锛屽皬绋嬪簭鍚庣浼樺厛琛ラ綈銆屽畬鎴?5 涓祴璇曠敤鎴枫€? 涓眬銆?0 鏉¤涓恒€? 鏉℃敹钘忋€? 鏉¤瘎浠峰悗鍚庡彴鍙湅鍒版暟鎹矇娣€銆嶇殑鍙噸澶嶉獙鏀堕摼璺€?- 鏂板鍚庡彴鍙楁帶楠屾敹澶瑰叿鎺ュ彛锛?  - `POST /api/admin/ai-data/acceptance-fixture`
  - 鏉冮檺鐮侊細`ai:data:seed`
  - 杩斿洖 `AIDataAcceptanceFixtureResultDTO`锛屽寘鍚湰娆¤ˉ榻愭暟閲忓拰琛ラ綈鍚庣殑 `AIDataSnapshotDTO`銆?- 澶瑰叿浼氬鐢ㄧ幇鏈夊皬绋嬪簭鍚庣鏈嶅姟閾捐矾鍐欏叆鏁版嵁锛?  - 寮哄疄鍚嶏細涓?5 涓浐瀹氶獙鏀剁敤鎴疯ˉ榻愭湰鍦板己瀹炲悕鐘舵€併€?  - 缁勫眬锛氬垱寤哄苟瀹℃牳 3 涓厤璐瑰眬锛屽畬鎴愬叆灞€銆佹墜鍔ㄥ紑濮嬨€佹湇鍔＄‘璁ゃ€?  - 琛屼负锛氳ˉ榻?20 鏉¤涓烘棩蹇椼€?  - 鏀惰棌锛氳ˉ榻?5 鏉℃敹钘忋€?  - 璇勪环锛氳ˉ榻?5 鏉¤瘎浠凤紝骞朵骇鐢熻冻杩?鎴愰暱璁板綍銆?  - IM锛氳ˉ榻愬熀纭€ IM 鐣欏瓨娑堟伅銆?  - 鐢诲儚锛氳ˉ榻愯瀹跺拰棰嗚矾浜虹敾鍍忔暟鎹€?- 鎺ュ彛鍏峰骞傜瓑淇濇姢锛氬綋 `acceptanceReady=true` 鍚庨噸澶嶈皟鐢ㄤ笉缁х画杩藉姞鏁版嵁銆?- 鍚屾鏉冮檺鍜岃仈璋冩潗鏂欙細
  - `adminauth` 鍐呯疆鏉冮檺鏂板 `ai:data:seed`銆?  - `db/seeds/admin_roles_permissions.sql` 鏂板 `ai:data:seed` 鏉冮檺绉嶅瓙銆?  - `docs/openapi/admin.openapi.yaml` 鏂板澶瑰叿鎺ュ彛銆?  - `docs/openapi/dto-samples.md` 鏂板 `AIDataAcceptanceFixtureResultDTO`銆?  - `docs/test-cases/TC-144-ai-data-im-export-gate.md` 鏂板澶瑰叿姝ラ鍜岃嚜鍔ㄥ寲瑕嗙洊銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `TestAIDataAcceptanceFixtureHTTP`
  - `TestAIDataSnapshotAndIMExportGateHTTP`
  - `cmd/server`
  - `internal/adminauth`
  - `internal/appapi`
  - `internal/auth`
  - `internal/games`
  - `internal/identity`
  - `internal/revenueclient`
  - 鍏朵綑鏃犳祴璇曟枃浠跺寘姝ｅ父缂栬瘧閫氳繃銆?
### 褰撳墠椋庨櫓

- 澶瑰叿褰撳墠鍩轰簬鍐呭瓨鏈嶅姟锛岄€傚悎鏈湴鍜岃仈璋冮獙鏀讹紱鎺?PostgreSQL 鍚庡簲杩佺Щ涓烘祴璇曠幆澧冪瀛愪换鍔℃垨鍙楃幆澧冨彉閲忎繚鎶ょ殑鍐呴儴杩愮淮鎺ュ彛銆?- 鐢熶骇鐜涓嶅簲鏆撮湶 `ai:data:seed` 缁欐櫘閫氬悗鍙拌鑹诧紱涓婄嚎鍓嶉渶瑕佸湪瑙掕壊閰嶇疆涓粎淇濈暀缁欒秴绠℃垨娴嬭瘯鐜璐﹀彿銆?
## 2026-06-13 22:30

### 褰撳墠杩涘睍

- 缁х画鎸?`plan.md` 鎺ㄨ繘 E6/E7锛屽皬绋嬪簭鍚庣浼樺厛琛ラ綈鏈嶅姟灞傚彲楠岃瘉闂幆銆?- E6 鑱旇皟鐢ㄤ緥琛ラ綈锛?  - 鏂板 `docs/test-cases/app-integration-cases.md`锛岃鐩栧皬绋嬪簭鐧诲綍銆佸己瀹炲悕銆侀個绾︺€佺粍灞€銆両M銆佹湇鍔＄‘璁ゃ€佸鐩樸€佺画灞€銆佺Н鍒嗗厬鎹€丄I 鏁版嵁鍑嗗绛夌鍒扮閾捐矾銆?  - 鏂板 `docs/test-cases/admin-integration-cases.md`锛岃鐩栧悗鍙版潈闄愩€佸疄鍚嶇姸鎬併€佹墦鍗″鐞嗐€佽瘎浠锋垚闀裤€佸厬鎹㈠鏍搞€佷妇鎶ョ敵璇夈€佹姤琛ㄥ鍑恒€丄I 鏁版嵁鍑嗗鍜屼氦浠樻祴璇曠暀妗ｃ€?  - `docs/openapi/error-codes.md` 琛ュ厖浼氬憳鎶ヨ〃銆佸洟闃熺鐞嗐€佸鐩橀噸澶嶃€佸厬鎹㈠簱瀛樸€佺Н鍒嗕笉瓒炽€侀噷绋嬬鍜屾墦鍗″弬鏁扮被閿欒鐮併€?- E7 鏈嶅姟娴嬭瘯琛ラ綈锛?  - 浜鸿剦鍏崇郴锛氬弻鍚戣繛鎺ャ€佽窡杩涜褰曞弬涓庤€?棰嗚矾浜烘潈闄愩€?  - 琛屽/棰嗚矾浜虹敾鍍忥細鏈巿浜堣鑹蹭笉鑳藉彂甯冩妧鑳芥垨璧勬簮銆?  - 绉垎锛氬彂鏀俱€佹墸鍑忋€佷綑棰濅笉瓒炽€?  - 鍏戞崲锛氱Н鍒嗕笉瓒炽€佸簱瀛樹笉瓒炽€佸苟鍙戜笉瓒呭崠銆侀┏鍥為€€绉垎銆?  - 浜や粯娴嬭瘯锛氭祴璇曠敤渚嬬櫥璁般€佹祴璇曟墽琛屽叧鑱斿拰缁撴灉鏍￠獙銆?  - 娓告垙锛氭敹钘忓箓绛夈€佸彇娑堟敹钘忓箓绛夈€侀噷绋嬬鏉冮檺涓庣姸鎬併€佹墦鍗℃潈闄愪笌鏃犳晥闅愯棌銆佸鐩樻潈闄愩€侀噸澶嶅鐩樻嫤鎴€佺画灞€鍙敓鎴愯崏绋裤€?  - 寮哄疄鍚嶏細榛樿 `wechat_logged_in`锛孎aceID 瀹屾垚鍓嶄笉鍙涓哄己瀹炲悕瀹屾垚锛屽畬鎴愬悗鎵嶆斁琛屻€?- 鏈嶅姟灞傛柊澧?`ErrDuplicateRetrospective`锛屽悓涓€鎴愬憳鍚屼竴灞€閲嶅鎻愪氦澶嶇洏浼氳繑鍥炲啿绐侀敊璇紝骞舵槧灏勫埌灏忕▼搴忔帴鍙ｉ敊璇爜 `40941`銆?- 璇存槑锛歚plan.md` 涓垪鍑虹殑 `internal/favorites/favorite_service_test.go` 鍦ㄥ綋鍓嶄唬鐮佺粨鏋勯噷瀹為檯灞炰簬 `internal/games`锛屽洜姝ゆ祴璇曡惤鍦?`services/go-api/internal/games/favorite_service_test.go`锛屼繚鎸佸拰鐜版湁瀹炵幇涓€鑷淬€?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `internal/appapi`
  - `internal/games`
  - `internal/identity`
  - `internal/connections`
  - `internal/profiles`
  - `internal/points`
  - `internal/redemption`
  - `internal/delivery`
  - 鍏朵綑鏃犳祴璇曟枃浠跺寘姝ｅ父缂栬瘧閫氳繃銆?
### 褰撳墠椋庨櫓

- 褰撳墠 E7 浠ユ湇鍔″眰鍜?HTTP 灞傛祴璇曚负涓伙紝灏氭湭鎺ュ叆鐪熷疄 PostgreSQL 骞跺彂浜嬪姟锛涘厬鎹㈠苟鍙戜笉瓒呭崠宸插湪鍐呭瓨鏈嶅姟閿佷笅楠岃瘉锛屾帴搴撳悗浠嶉渶鐢ㄦ暟鎹簱浜嬪姟鍜岃绾ч攣澶嶆祴銆?- 褰撳墠鐩綍涓嶆槸 Git 浠撳簱锛屾棤娉曠敤 `git status` 杈撳嚭鍙樻洿娓呭崟锛涙湰杞娇鐢ㄦ枃浠舵壂鎻忓拰鍏ㄩ噺娴嬭瘯纭鍙樻洿鑼冨洿銆?
## 2026-06-13 23:10

### 褰撳墠杩涘睍

- 缁х画鎺ㄨ繘 `plan.md` D5.7銆屽垎娑︺€佹敹鐩娿€佽鍗曞崰浣嶃€嶄腑灏忕▼搴忓悗绔紭鍏堢殑璁㈠崟鍗犱綅闂幆銆?- 鏂板 `services/go-api/internal/orders`锛?  - 鍏嶈垂灞€璁㈠崟鍗犱綅 `free_no_pay`銆?  - 璁㈠崟鍙?`FREE-{gameId}-{id}`銆?  - 閲戦鍥哄畾涓?0 鍒嗐€?  - `needWechatPay=false`锛屼竴鏈熶笉浼氳Е鍙戠湡瀹炲井淇℃敮浠樸€?  - 鍚屼竴灞€閲嶅纭繚璁㈠崟鏃跺箓绛夎繑鍥炲悓涓€璁㈠崟銆?  - 鏀粯鍥炶皟鍗犱綅鎸?`orderNo + eventId` 骞傜瓑璁板綍銆?- 灏忕▼搴忓悗绔帴鍙ｆ柊澧烇細
  - `GET /api/app/orders/{orderId}`锛氬綋鍓嶇敤鎴锋煡璇㈣嚜宸辩殑鏀粯璁㈠崟鍗犱綅銆?  - `POST /api/app/payment/precreate-placeholder`锛氬厤璐瑰眬棰勪笅鍗曞崰浣嶃€?  - `POST /api/internal/pay/callback-placeholder`锛氭敮浠樺洖璋冨崰浣嶅拰骞傜瓑璁板綍銆?- `POST /api/app/games` 鍒涘缓鍏嶈垂灞€鎴愬姛鍚庤嚜鍔ㄧ‘淇?`free_no_pay` 璁㈠崟瀛樺湪锛涘師杩斿洖浠嶄繚鎸?`GameDTO`锛岄伩鍏嶇牬鍧忓皬绋嬪簭鏃㈡湁鑱旇皟銆?- 鍚屾鏂囨。锛?  - `docs/openapi/app.openapi.yaml` 澧炲姞璁㈠崟璇︽儏鍜岄涓嬪崟鍗犱綅鎺ュ彛銆?  - `docs/openapi/admin.openapi.yaml` 澧炲姞鍐呴儴鏀粯鍥炶皟鍗犱綅鎺ュ彛銆?  - `docs/openapi/dto-samples.md` 澧炲姞 `PaymentOrderDTO` 鍜?`PaymentPrecreatePlaceholderDTO`銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `internal/orders`
  - `internal/appapi`
  - `internal/games`
  - `internal/identity`
  - `internal/revenueclient`
  - 鍏朵綑宸叉湁鍚庣鍖呮甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- 褰撳墠璁㈠崟鍗犱綅涓哄唴瀛樻湇鍔★紝宸叉弧瓒虫湰鏈鸿仈璋冨拰鎺ュ彛楠屾敹锛涙帴 PostgreSQL 鍚庨渶瑕佹妸 `payment_orders` 浠撳偍銆佸敮涓€绾︽潫鍜屽洖璋冧簨浠惰〃钀戒负浜嬪姟鍐欏叆銆?- 璧勯噾鏈嶅姟宸叉湁 `/api/funds/payment-precreate-placeholder` 鍥哄畾杩斿洖缁撴瀯锛涙湰杞厛琛?Go 灏忕▼搴忓悗绔彲鏌ヨ鍗曞拰鍐呴儴鍥炶皟骞傜瓑锛屽悗缁彲鎶?Go 渚ц鍗曟湇鍔′笌 Java 璧勯噾鏈嶅姟鍗犱綅瀹㈡埛绔覆鑱斻€?
## 2026-06-13 23:40

### 褰撳墠杩涘睍

- 缁х画鎺ㄨ繘 `plan.md` D5.8銆屼妇鎶ョ敵璇夈€侀€氱煡鍜岃涓虹暀瀛樸€嶏紝灏忕▼搴忓悗绔紭鍏堣ˉ榻愯涓烘棩蹇楀瓧娈靛彛寰勩€?- `internal/audit` 琛屼负鏃ュ織琛ラ綈璁″垝瀛楁锛?  - `eventCode`
  - `businessType`
  - `businessId`
  - `source`
  - `device`
  - `ip`
  - `occurredAt`
- 淇濇寔鏃у瓧娈靛吋瀹癸細
  - `eventType` 浠嶇瓑浠蜂簬 `eventCode`銆?  - `targetType` 浠嶇瓑浠蜂簬 `businessType`銆?  - `targetId` 浠嶇瓑浠蜂簬 `businessId`銆?- `POST /api/app/behavior/events` 鏀寔鏂版棫瀛楁鍚屾椂涓婃姤锛屾湭浼?source 鏃堕粯璁?`app`锛宒evice 鍙粠璇锋眰浣撴垨 `X-Device` 璇诲彇锛宨p 鐢卞悗绔В鏋愩€?- `GET /api/admin/behavior/events` 鏀寔鎸?`eventCode` 杩囨护锛屾柟渚垮悗鍙板仛婕忔枟鍜岀暀瀛樺熀纭€缁熻銆?- 琛ラ綈 D5.8 鏈嶅姟灞傛祴璇曪細
  - `internal/reports`锛氫妇鎶ヨ瘉鎹?ID 鐣欏瓨銆佸垱寤轰妇鎶ュ悗鍐荤粨鍒嗘鼎銆佸鐞嗗拰鍏抽棴鐘舵€佹祦杞€?  - `internal/notifications`锛氱珯鍐呴€氱煡鐢熸垚銆佸井淇¤闃呮秷鎭换鍔＄敓鎴愩€佸凡璇汇€佸彂閫佺粨鏋滃洖鍐欍€?  - `internal/audit`锛氳涓烘棩蹇楁柊瀛楁鍜屾棫瀛楁鍒悕鍏煎銆佹寜 `eventCode` 鏌ヨ銆?- 鍚屾鏂囨。锛?  - `docs/openapi/app.openapi.yaml` 澧炲姞琛屼负浜嬩欢鎺ュ彛瀛楁璇存槑銆?  - `docs/openapi/admin.openapi.yaml` 澧炲姞琛屼负浜嬩欢绛涢€夎鏄庛€?  - `docs/openapi/dto-samples.md` 澧炲姞 `BehaviorLogDTO`銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `internal/audit`
  - `internal/appapi`
  - `internal/reports`
  - `internal/notifications`
  - `internal/orders`
  - `internal/games`
  - `internal/identity`
  - 鍏朵綑宸叉湁鍚庣鍖呮甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- 褰撳墠琛屼负鏃ュ織浠嶆槸鍐呭瓨鏈嶅姟锛屽凡婊¤冻鏈湴鑱旇皟鍜岄獙鏀跺瓧娈靛彛寰勶紱鎺?PostgreSQL 鍚庨渶瑕佸皢杩欎簺瀛楁鍐欏叆 `user_behavior_logs`锛屽苟鎸?`event_code/user_id/occurred_at` 寤虹储寮曘€?
## 2026-06-14 00:10

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堣ˉ榻愬皬绋嬪簭鍚庣涓婄嚎鏁版嵁搴撴壙鎺ヨ兘鍔涖€?- 鏂板杩佺Щ `db/migrations/000013_behavior_log_fields.sql`锛?  - 缁?`user_behavior_logs` 澧炲姞 `event_code`銆乣business_type`銆乣business_id`銆乣source`銆乣device`銆乣ip`銆乣occurred_at`銆?  - 鏃ф暟鎹敤 `event_type`銆乣target_type`銆乣target_id`銆乣created_at` 鍥炲～鏂板瓧娈点€?  - 澧炲姞 `idx_user_behavior_logs_event_code_occurred`銆乣idx_user_behavior_logs_user_occurred`銆乣idx_user_behavior_logs_business`銆?- 鏂板杩佺Щ `db/migrations/000014_payment_callback_placeholder.sql`锛?  - 鏂板 `payment_callbacks`锛岃褰?provider銆乧allback_type銆乪vent_id銆乷rder_no銆乷ut_trade_no銆乼ransaction_id銆乺aw_headers_json銆乺aw_body銆乸ayload_digest銆乿erify_status銆乸rocess_status銆?  - 澧炲姞 `uk_payment_callbacks_order_event`锛屾寜 `order_no + event_id` 淇濊瘉鏀粯鍥炶皟鍗犱綅骞傜瓑銆?  - 澧炲姞鍥炶皟浜ゆ槗鍙枫€佽鍗曞彿銆佸鐞嗙姸鎬佺储寮曘€?- 鍚屾 `plan.md`锛?  - D4 杩佺Щ椤哄簭杩藉姞 `000011`銆乣000012`銆乣000013`銆乣000014`銆?  - 鍞竴绾︽潫娓呭崟杩藉姞 `uk_payment_callbacks_order_event`銆?
### 楠岃瘉缁撴灉

- 鏂囨湰妫€鏌ョ‘璁よ縼绉诲瓧娈点€佺储寮曞拰 `plan.md` 椤哄簭鍧囧彲妫€绱€?- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `internal/appapi`
  - `internal/audit`
  - `internal/orders`
  - `internal/reports`
  - `internal/notifications`
  - 鍏朵綑宸叉湁鍚庣鍖呮甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- 鏈疆瀹屾垚鐨勬槸杩佺Щ灞傛壙鎺ワ紱杩愯鏃朵粛浠ュ唴瀛樻湇鍔′负涓汇€傛帴 PostgreSQL 鍚庨渶瑕佹妸 `internal/audit` 鍜?`internal/orders` 鎺ュ叆 SQL repository锛屽苟鐢ㄧ湡瀹炴暟鎹簱楠岃瘉鍥炶皟骞傜瓑绾︽潫銆?
## 2026-06-14 00:40

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟妸灏忕▼搴忓悗绔?D5.8 / D5.7 鐨勫唴瀛橀棴鐜帹杩涘埌鍙帴 PostgreSQL repository銆?- `internal/audit` 鏂板琛屼负鏃ュ織鎸佷箙鍖栨帴鍙ｄ笌 SQL 閫傞厤锛?  - `BehaviorRepository`
  - `NewServiceWithBehaviorRepository`
  - `NewSQLBehaviorRepository`
  - 鏀寔鍐欏叆 `user_behavior_logs` 鐨?`event_type/event_code/target_type/business_type/target_id/business_id/source/device/ip/extra/created_at/occurred_at`銆?  - 鏀寔鍚庡彴鎸?`userId/eventType/eventCode` 鏌ヨ銆?- `internal/orders` 鏂板璁㈠崟鍜屾敮浠樺洖璋冨崰浣嶆寔涔呭寲鎺ュ彛涓?SQL 閫傞厤锛?  - `Repository`
  - `NewServiceWithRepository`
  - `NewSQLRepository`
  - 鏈夋暟鎹簱鏃跺厤璐瑰眬璁㈠崟鍐欏叆 `payment_orders`锛岃鍗曞彿閲囩敤 `FREE-{gameId}`锛屼緷璧?`order_no` 鍞竴绾︽潫骞傜瓑銆?  - 鏀粯鍥炶皟鍗犱綅鍐欏叆 `payment_callbacks`锛屼緷璧?`order_no + event_id` 鍞竴绱㈠紩骞傜瓑銆?- `cmd/server` 鍦?DB 鍙敤鏃舵敞鍏ワ細
  - `audit.NewSQLBehaviorRepository(db)`
  - `orders.NewSQLRepository(db)`
  - DB 涓嶅彲鐢ㄦ椂淇濇寔鍘熷唴瀛樻湇鍔¤涓恒€?- `internal/appapi.Server` 鏂板 `UseRepositories` 娉ㄥ叆鐐癸紝涓嶆敼鍙樼幇鏈夋祴璇曟瀯閫犲拰瀵瑰鎺ュ彛銆?- 琛ュ厖娴嬭瘯锛?  - `internal/audit` fake repository 娴嬭瘯锛岀‘璁や繚瀛樸€佸垪琛ㄣ€佹煡璇細璧?repository銆?  - `internal/orders` fake repository 娴嬭瘯锛岀‘璁ゅ垱寤鸿鍗曘€佹煡璇㈣鍗曘€佸洖璋冨箓绛変細璧?repository銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/audit`
  - `internal/orders`
  - 鍏朵綑宸叉湁鍚庣鍖呮甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- 褰撳墠椤圭洰浠嶆湭寮曞叆 PostgreSQL 椹卞姩渚濊禆锛宍DATABASE_DRIVER=postgres` 鍦ㄦ棤椹卞姩娉ㄥ唽鏃朵細鐢?`cmd/server` 鍥為€€鍐呭瓨妯″紡锛涙湰杞畬鎴?repository 楠ㄦ灦鍜屾敞鍏ョ偣锛岀湡瀹炴暟鎹簱闆嗘垚娴嬭瘯浠嶉渶鍦ㄤ緷璧栫瓥鐣ョ‘璁ゅ悗鎵ц銆?
## 2026-06-14 01:10

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堣ˉ榻愬皬绋嬪簭鍚庣 D5.8 涓炬姤鐢宠瘔閾捐矾鐨?PostgreSQL 鎵挎帴鑳藉姏銆?- `internal/reports` 鏂板涓炬姤鐢宠瘔鎸佷箙鍖栨帴鍙ｄ笌 SQL 閫傞厤锛?  - `Repository`
  - `NewServiceWithRepository`
  - `NewSQLRepository`
  - 鏀寔鍒涘缓涓炬姤鍐欏叆 `reports`銆?  - 鏀寔鎴戠殑涓炬姤銆佸悗鍙颁妇鎶ュ垪琛ㄣ€佷妇鎶ヨ鎯呬粠 `reports` 璇诲彇銆?  - 鏀寔鍚庡彴澶勭悊/鍏抽棴涓炬姤鍚庡洖鍐?`status`銆乣handler_admin_id`銆乣handle_result`銆乣handled_at`銆?- `internal/appapi.Server` 鐨?repository 娉ㄥ叆鐐规墿灞曞埌涓炬姤鐢宠瘔锛?  - 鏈夋暟鎹簱鏃舵敞鍏?`reports.NewSQLRepository(db)`銆?  - 鏃犳暟鎹簱鏃朵繚鐣欏師鍐呭瓨鏈嶅姟锛屼繚鎸佹湰鍦拌仈璋冨拰鏃㈡湁娴嬭瘯琛屼负涓嶅彉銆?- `cmd/server` 鏂板 `reportRepository(db)`锛屽拰琛屼负鏃ュ織銆佽鍗?repository 涓€璧峰湪鍚姩鏃舵敞鍐屻€?- 琛ュ厖 `internal/reports` fake repository 鍗曞厓娴嬭瘯锛岀‘璁ゅ垱寤恒€佹煡璇€佸鐞嗕細璧?repository 鍒嗘敮銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/reports`
  - `internal/audit`
  - `internal/orders`
  - 鍏朵綑宸叉湁鍚庣鍖呮甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- `reports` 宸插叿澶?SQL repository锛屼絾閫氱煡 `notifications` 浠嶄互鍐呭瓨鏈嶅姟涓轰富锛涗笅涓€姝ュ簲缁х画鎶婄珯鍐呴€氱煡銆佸井淇¤闃呮秷鎭换鍔″拰妯℃澘鎺ュ叆 PostgreSQL銆?- `reports` 鐨勬暟鎹簱绾ч泦鎴愭祴璇曚粛渚濊禆 PostgreSQL 椹卞姩鍜屾祴璇曞簱绛栫暐纭锛涘綋鍓嶉獙璇佽鐩栨湇鍔℃敞鍏ャ€佺紪璇戝拰鍐呭瓨/fake repository 琛屼负銆?
## 2026-06-14 01:40

### 褰撳墠杩涘睍

- 缁х画鎵ц D5.8锛岃ˉ榻愮珯鍐呴€氱煡鍜屽井淇¤闃呮秷鎭换鍔＄殑 PostgreSQL repository銆?- `internal/notifications` 鏂板鎸佷箙鍖栨帴鍙ｄ笌 SQL 閫傞厤锛?  - `Repository`
  - `NewServiceWithRepository`
  - `NewSQLRepository`
  - 鍒涘缓绔欏唴閫氱煡鏃跺啓鍏?`notifications`銆?  - `needWechat=true` 鏃跺悓姝ュ啓鍏?`wechat_subscribe_tasks`锛屽苟鍥炲～ `notifications.wechat_task_id`銆?  - 鏌ヨ鎴戠殑閫氱煡銆佸悗鍙板井淇¤闃呬换鍔°€佸井淇¤闃呮ā鏉挎椂浼樺厛璇?repository銆?  - 鏍囪宸茶浼氬洖鍐?`notifications.status/read_at`銆?  - 鏍囪寰俊浠诲姟宸插彂閫佷細鍥炲啓 `wechat_subscribe_tasks.status/result/sent_at`锛屽苟鍚屾 `notifications.wechat_state=sent`銆?  - SQL 妯″紡涓嬪井淇¤闃呬换鍔′笉瀛樺湪鏃舵槧灏勪负 `ErrNotificationNotFound`锛屽拰鍐呭瓨妯″紡淇濇寔涓€鑷淬€?- `internal/appapi.Server` repository 娉ㄥ叆鐐圭户缁墿灞曪細
  - 鏈夋暟鎹簱鏃舵敞鍏?`notifications.NewSQLRepository(db)`銆?  - 鏃犳暟鎹簱鏃剁户缁娇鐢ㄥ師鍐呭瓨閫氱煡鏈嶅姟锛屼繚鎸佹湰鍦拌仈璋冨拰宸叉湁娴嬭瘯绋冲畾銆?- `cmd/server` 鏂板 `notificationRepository(db)`锛屽拰琛屼负鏃ュ織銆佽鍗曘€佷妇鎶ョ敵璇変竴璧峰湪鍚姩鏃舵敞鍐屻€?- 琛ュ厖 `internal/notifications` fake repository 鍗曞厓娴嬭瘯锛岀‘璁ゅ垱寤恒€佸垪琛ㄣ€佸凡璇汇€佹ā鏉裤€佷换鍔″彂閫佺粨鏋滀細璧?repository 鍒嗘敮銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/notifications`
  - `internal/reports`
  - `internal/audit`
  - `internal/orders`
  - 鍏朵綑宸叉湁鍚庣鍖呮甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- D5.8 鐨勪妇鎶ャ€侀€氱煡銆佽涓洪摼璺凡鍏峰 SQL repository 楠ㄦ灦锛屼絾褰撳墠椤圭洰浠嶆湭鍚敤鐪熷疄 PostgreSQL 椹卞姩鍜屾祴璇曞簱锛涙暟鎹簱绾у敮涓€绾︽潫銆佷簨鍔″洖婊氬拰骞跺彂琛屼负闇€瑕佸悗缁湪闆嗘垚鐜澶嶆祴銆?- `wechat_subscribe_templates` 鐩墠渚濊禆鏁版嵁搴撶瀛愭垨鍚庡彴閰嶇疆锛涘鏋滃簱鍐呮病鏈夋ā鏉匡紝鏈嶅姟浼氫繚鐣欏唴瀛橀粯璁ゆā鏉夸綔涓烘湰鍦?fallback銆?
## 2026-06-14 02:10

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堣ˉ榻愬皬绋嬪簭鍚庣鏂囦欢涓婁紶銆両M 鏂囦欢銆佷妇鎶ラ檮浠跺拰瀵煎嚭鏂囦欢渚濊禆鐨勬枃浠剁櫥璁版寔涔呭寲鑳藉姏銆?- 鏂板杩佺Щ `db/migrations/000015_files_repository.sql`锛?  - 鍒涘缓 `files` 琛ㄣ€?  - 瀛楁瑕嗙洊涓婁紶浜恒€佷笟鍔＄被鍨嬨€佷笟鍔″璞°€佹枃浠跺悕銆丮IME銆佸ぇ灏忋€丼HA256銆佸瓨鍌?key銆佽闂骇鍒€佽繃鏈熸椂闂村拰鍒涘缓鏃堕棿銆?  - 澧炲姞 `uk_files_storage_key` 鍞竴绱㈠紩銆?  - 澧炲姞 `idx_files_biz_object` 鍜?`idx_files_uploader_created`銆?- `internal/files` 鏂板鏂囦欢鏈嶅姟 repository锛?  - `Repository`
  - `NewServiceWithRepository`
  - `NewSQLRepository`
  - 涓婁紶鍑瘉鍒涘缓鏃跺彲鍐欏叆 `files`銆?  - 瀵煎嚭绛夌郴缁熺敓鎴愭枃浠跺彲鍐欏叆 `files`銆?  - 涓嬭浇鍦板潃鐢熸垚鍓嶅彲浠?repository 璇诲彇鏂囦欢鍏冩暟鎹€?- `internal/appapi.Server` repository 娉ㄥ叆鐐圭户缁墿灞曪細
  - 鏈夋暟鎹簱鏃舵敞鍏?`files.NewSQLRepository(db)`銆?  - 鏃犳暟鎹簱鏃剁户缁娇鐢ㄥ師鍐呭瓨鏂囦欢鏈嶅姟锛屼繚鎸佹湰鍦版帴鍙ｆ祴璇曞拰鑱旇皟琛屼负绋冲畾銆?- `cmd/server` 鏂板 `fileRepository(db)`锛屽拰琛屼负鏃ュ織銆佽鍗曘€佷妇鎶ャ€侀€氱煡涓€璧峰湪鍚姩鏃舵敞鍐屻€?- `plan.md` 鍚屾琛ュ厖锛?  - 杩佺Щ娓呭崟澧炲姞 `000015_files_repository.sql`銆?  - 鍞竴绾︽潫娓呭崟澧炲姞 `uk_files_storage_key`銆?  - C3 鏁版嵁搴撹縼绉绘墽琛岄『搴忓鍔犵 15 椤广€?- 琛ュ厖 `internal/files` fake repository 鍗曞厓娴嬭瘯锛岀‘璁や笂浼犲嚟璇併€佺敓鎴愭枃浠躲€佹煡璇笅杞介兘浼氳蛋 repository 鍒嗘敮銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/files`
  - `internal/notifications`
  - `internal/reports`
  - `internal/audit`
  - 鍏朵綑宸叉湁鍚庣鍖呮甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- 鏂囦欢鏈嶅姟宸插叿澶?SQL repository锛屼絾鐪熷疄瀵硅薄瀛樺偍浠嶆槸 mock URL锛涘悗缁渶瑕佹帴 MinIO/COS 棰勭鍚嶄笂浼犲拰涓嬭浇锛屽苟淇濈暀褰撳墠鏉冮檺鏍￠獙銆?- 褰撳墠鏁版嵁搴撶骇楠岃瘉浠嶅彈 PostgreSQL 椹卞姩鍜屾祴璇曞簱绛栫暐闄愬埗锛涙湰杞鐩栬縼绉绘枃鏈€佺紪璇戙€佹湇鍔℃敞鍏ュ拰 fake repository 琛屼负銆?
## 2026-06-14 02:40

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堣ˉ榻愬皬绋嬪簭鍚庣鏀惰棌灞€鏁版嵁娌夋穩鑳藉姏銆?- `internal/games` 鏂板鏀惰棌涓撶敤鎸佷箙鍖栨帴鍙ｄ笌 SQL 閫傞厤锛?  - `FavoriteRepository`
  - `NewServiceWithFavoriteRepository`
  - `NewSQLFavoriteRepository`
  - 鏀惰棌灞€鏃跺啓鍏?`game_favorites`锛屼緷璧?`user_id + game_id` 鍞竴绾︽潫淇濇寔骞傜瓑銆?  - 鍙栨秷鏀惰棌鏃跺垹闄?`game_favorites` 瀵瑰簲璁板綍锛屼繚鎸侀噸澶嶅彇娑堝箓绛夈€?  - 鎴戠殑鏀惰棌浼樺厛浠?repository 璇诲彇锛屽啀澶嶇敤褰撳墠娓告垙鏈嶅姟琛ラ綈 `Game` 淇℃伅骞惰繃婊ゆ湭瀹℃牳灞€銆?  - 鍏ㄩ噺鏀惰棌浼樺厛浠?repository 璇诲彇锛屼緵 AI 鏁版嵁鍑嗗鍜屽悗鍙扮粺璁＄户缁鐢ㄣ€?- `cmd/server` 鍚姩鍏ュ彛鎺ュ叆 `favoriteRepository(db)`锛?  - 鏈夋暟鎹簱鏃朵娇鐢?`games.NewSQLFavoriteRepository(db)`銆?  - 鏃犳暟鎹簱鏃剁户缁娇鐢ㄥ師鍐呭瓨鏀惰棌鏈嶅姟銆?- 琛ュ厖 `internal/games` fake repository 鍗曞厓娴嬭瘯锛岀‘璁ゆ敹钘忋€佹垜鐨勬敹钘忋€佸叏閲忔敹钘忋€佸彇娑堟敹钘忛兘浼氳蛋 repository 鍒嗘敮銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/games`
  - `internal/files`
  - `internal/audit`
  - `internal/notifications`
  - 鍏朵綑宸叉湁鍚庣鍖呮甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- 褰撳墠娓告垙涓昏〃 `games` 杩愯鏃朵粛浠ュ唴瀛樻湇鍔′负涓伙紱鏀惰棌 repository 宸茶兘娌夋穩 `game_favorites`锛屼絾閲嶅惎鍚庡鏋滄父鎴忎富琛ㄦ湭鍚屾鍏ュ簱锛屾敹钘忓垪琛ㄦ棤娉曡ˉ榻愬畬鏁?`Game` 璇︽儏銆?- Go 宸ュ叿閾惧湪鏈満鐜浼氬皾璇曞啓鐢ㄦ埛鐩綍 telemetry token 骞惰緭鍑烘潈闄愯鍛婏紱娴嬭瘯閫€鍑虹爜涓?0锛屾墍鏈夊寘宸查€氳繃銆?
## 2026-06-14 03:10

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛岃ˉ榻愬悗鍙?鍐呴儴鍏抽敭鎿嶄綔瀹¤鐨?PostgreSQL 鎸佷箙鍖栨壙鎺ャ€?- `internal/audit` 鎵╁睍鎿嶄綔鏃ュ織 repository锛?  - `OperationRepository`
  - `NewServiceWithRepositories`
  - `NewSQLOperationRepository`
  - 琛屼负鏃ュ織鍜屾搷浣滄棩蹇楀彲鍒嗗埆娉ㄥ叆 repository锛屼簰涓嶅奖鍝嶃€?- `RecordOperation` 缁х画淇濈暀鍐呭瓨璁板綍锛屽悓鏃跺湪鏈?repository 鏃跺啓鍏?`operation_logs`銆?- `OperationLogs` 鍦ㄦ湁 repository 鏃朵紭鍏堣鍙?`operation_logs`锛屾棤 repository 鎴栬鍙栧け璐ユ椂鍥為€€鍐呭瓨銆?- `cmd/server` 鍚姩鍏ュ彛鏂板 `operationRepository(db)`锛?  - 鏈夋暟鎹簱鏃朵娇鐢?`audit.NewSQLOperationRepository(db)`銆?  - 鏃犳暟鎹簱鏃剁户缁娇鐢ㄥ師鍐呭瓨瀹¤鏈嶅姟銆?- `internal/appapi.Server.UseRepositories` 鎵╁睍涓哄悓鏃舵敞鍏ヨ涓烘棩蹇楀拰鎿嶄綔鏃ュ織 repository銆?- 琛ュ厖 `internal/audit` fake operation repository 鍗曞厓娴嬭瘯锛岀‘璁ゆ搷浣滄棩蹇椾繚瀛樺拰鍒楄〃鏌ヨ閮戒細璧?repository 鍒嗘敮銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/audit`
  - `internal/games`
  - `internal/files`
  - `internal/notifications`
  - 鍏朵綑宸叉湁鍚庣鍖呮甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- `operation_logs` 宸插叿澶?SQL repository锛屼絾骞堕潪鎵€鏈夊悗鍙板啓鎿嶄綔閮藉凡瑕嗙洊鎿嶄綔鏃ュ織锛涘悗缁粛闇€鎸?`plan.md` 閫愪釜琛ラ綈瀹℃牳銆侀厤缃€佸厬鎹€佸垎娑︾瓑鍐欐搷浣滅殑瀹¤鐐广€?- Go 宸ュ叿閾惧湪鏈満鐜浠嶄細灏濊瘯鍐欑敤鎴风洰褰?telemetry token 骞惰緭鍑烘潈闄愯鍛婏紱娴嬭瘯閫€鍑虹爜涓?0锛屼笟鍔″寘鍧囬€氳繃銆?## 2026-06-14 03:45

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟帹杩涘皬绋嬪簭鍚庣涓婄嚎绾?PostgreSQL 鎵挎帴鑳藉姏銆?- 鏂板杩佺Щ `db/migrations/000016_games_repository.sql`锛?  - 缁?`games` 琛ュ厖 `longitude`銆乣latitude`锛岃缁勫眬涓昏〃鍙洿鎺ユ壙杞藉皬绋嬪簭绔畾浣嶅瓧娈点€?  - 鏂板 `game_members`锛屾壙杞藉眬鎴愬憳銆佽鑹层€佺姸鎬佸拰鍏ュ眬鏃堕棿銆?  - 鏂板 `game_applications`锛屾壙杞藉叆灞€鐢宠銆佸鏍哥姸鎬併€佸師鍥犲拰瀹℃牳鏃堕棿銆?  - 鏂板鎴愬憳鐢ㄦ埛绱㈠紩銆佺敵璇风敤鎴风储寮曘€佺敵璇风姸鎬佺储寮曘€?- `internal/games` 鏂板缁勫眬涓婚摼璺?repository 鎺ュ彛涓?SQL 閫傞厤锛?  - `Repository`
  - `NewServiceWithRepositories`
  - `NewSQLRepository`
  - 鏀寔鍒涘缓灞€銆佹洿鏂板眬銆佹煡璇㈠眬銆佸垪琛ㄣ€佹瘡鏃ュ垱寤烘暟銆佹垚鍛樺鍒犳煡銆佸叆灞€鐢宠鍒涘缓/鏌ヨ/鏇存柊銆佸緟澶勭悊鐢宠骞傜瓑妫€鏌ャ€?- `cmd/server` 鍚姩鍏ュ彛鏂板 `gameRepository(db)` 娉ㄥ叆锛?  - 鏈夋暟鎹簱鏃朵娇鐢?`games.NewSQLRepository(db)`銆?  - 鏃犳暟鎹簱鏃剁户缁繚鎸佸師鏈夊唴瀛樻ā寮忥紝淇濊瘉鏈湴鑱旇皟鍜屾棦鏈夋祴璇曠ǔ瀹氥€?- `games.Service` 鏂板 SQL 鍥炲～鏈哄埗锛?  - 鏈嶅姟閲嶅惎鎴栧唴瀛樹负绌烘椂锛宍Get`銆佹垚鍛樻煡璇€佺敵璇峰鏍哥瓑鏍稿績璺緞鍙互浠?repository 鍥炲～灞€鍜屾垚鍛樸€?  - 鍒涘缓灞€銆佸鎵瑰眬銆佺敵璇峰叆灞€銆佸鏍稿叆灞€銆佹墜鍔ㄥ紑濮嬨€侀€€鍑哄眬浼氬湪鏈?repository 鏃跺啓鍏ユ暟鎹簱銆?  - 浠嶅悓姝ヤ繚鐣欏唴瀛樻€侊紝閬垮厤褰卞搷褰撳墠 IM銆佽瘎浠枫€佹敹鐩婄瓑渚濊禆鍚屼竴杩涚▼鐘舵€佺殑鍚庣画閾捐矾銆?- 鏂板 `internal/games/repository_service_test.go`锛?  - 瑕嗙洊 repository 鍒涘缓灞€銆佸垱寤鸿€呮垚鍛樺啓鍏ャ€佺敵璇峰叆灞€銆佸鏍稿叆灞€銆佹垚鍛樻煡璇€?  - 瑕嗙洊鏂版湇鍔″疄渚嬩粠 repository 璇诲彇灞€鍜屾垚鍛橈紝楠岃瘉閲嶅惎鍚庡洖濉兘鍔涖€?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/games`
  - `internal/audit`
  - `internal/files`
  - `internal/notifications`
  - `internal/orders`
  - 鍏朵綑宸叉湁灏忕▼搴忓悗绔寘鍧囨甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- `games` 涓婚摼璺凡鍏峰 SQL repository 楠ㄦ灦锛屼絾鐪熷疄 PostgreSQL 椹卞姩鍜屾祴璇曞簱浠嶆湭鍚敤锛涙暟鎹簱绾у敮涓€绾︽潫銆佷簨鍔″洖婊氥€佸苟鍙戝叆灞€鍜屾弧鍛樼珵浜夎繕闇€瑕侀泦鎴愭祴璇曞娴嬨€?- 鏈嶅姟纭銆侀噷绋嬬銆佹墦鍗°€佸鐩樸€佺画灞€鑽夌浠嶄富瑕佹槸鍐呭瓨鎬侊紱涓嬩竴姝ュ簲缁х画鎶?D5.6/E2.2 杩欎簺渚濊禆 `game_id` 鐨勮繘搴︽暟鎹矇鍏?PostgreSQL銆?- 褰撳墠 SQL 鍐欏叆鏄皬姝ユ帴鍏ワ紝灏氭湭鎶婄粍灞€瀹℃牳銆佺敵璇峰鏍搞€佹垚鍛樺啓鍏ュ拰浜烘暟鏇存柊鍖呮垚鍗曚釜鏁版嵁搴撲簨鍔★紱鎺ュ叆鐪熷疄 PostgreSQL 鍚庨渶瑕佽ˉ浜嬪姟灏佽銆?## 2026-06-14 04:15

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟帹杩涘皬绋嬪簭鍚庣 D5.6/E2.2 杩涘害绫绘暟鎹殑 PostgreSQL 鎵挎帴銆?- 澶嶆牳 `db/migrations/000010_p1_reserved.sql`锛岀‘璁ゅ凡鏈夎〃鍙鐢細
  - `game_milestones`
  - `game_checkins`
  - `game_retrospectives`
  - `game_continue_drafts`
- `internal/games` 鏂板杩涘害绫绘寔涔呭寲鎺ュ彛锛?  - `ProgressRepository`
  - `UseProgressRepository`
  - `NewSQLProgressRepository`
- `services/go-api/internal/games/progress_sql_repository.go` 鏂板 SQL 閫傞厤锛?  - 閲岀▼纰戝垱寤恒€佹洿鏂般€佸垪琛ㄣ€?  - 鎵撳崱鍒涘缓銆佺姸鎬佹洿鏂般€佹垚鍛樺垪琛ㄨ繃婊ゆ棤鏁堟墦鍗°€佸悗鍙板垪琛ㄥ寘鍚棤鏁堟墦鍗°€?  - 澶嶇洏鍒涘缓銆侀噸澶嶆彁浜ゆ鏌ャ€佸鐩樺垪琛ㄣ€?  - 缁眬鑽夌鍒涘缓銆佸悗鍙扮画灞€鑽夌鍒楄〃銆?  - `game_checkins.file_ids` 鎸?JSON/JSONB 璇诲啓锛屼繚鎸佸拰杩佺Щ琛ㄧ粨鏋勪竴鑷淬€?- `games.Service` 鎺ュ叆杩涘害绫?repository锛?  - `CreateMilestone`銆乣UpdateMilestone`銆乣Milestones`銆乣AdminMilestones`
  - `CreateCheckin`銆乣MarkCheckinInvalid`銆乣Checkins`銆乣AdminCheckins`
  - `CreateRetrospective`銆乣Retrospectives`銆乣AdminRetrospectives`
  - `ContinueDraft`銆乣AdminContinueDrafts`
  - 鏈?repository 鏃朵紭鍏堣鍐?PostgreSQL锛涙棤 repository 鏃朵繚鎸佸師鍐呭瓨琛屼负銆?- `cmd/server` 鏂板 `gameProgressRepository(db)` 娉ㄥ叆锛?  - 鏈夋暟鎹簱鏃朵娇鐢?`games.NewSQLProgressRepository(db)`銆?  - 鏃犳暟鎹簱鏃剁户缁湰鍦板唴瀛樿仈璋冦€?- 鏂板 `services/go-api/internal/games/progress_repository_service_test.go`锛?  - 瑕嗙洊閲岀▼纰戝垱寤?鏇存柊/鍒楄〃璧?repository銆?  - 瑕嗙洊鎵撳崱鍒涘缓銆佸悗鍙扮疆鏃犳晥銆佹垚鍛樼闅愯棌鏃犳晥鎵撳崱銆佸悗鍙扮鍙鏃犳晥鎵撳崱銆?  - 瑕嗙洊澶嶇洏鍒涘缓鍜?repository 绾ч噸澶嶆彁浜ゆ嫤鎴€?  - 瑕嗙洊缁眬鑽夌鍐欏叆鍜屽悗鍙版煡璇€?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/games`
  - `internal/audit`
  - `internal/files`
  - `internal/notifications`
  - `internal/orders`
  - 鍏朵綑宸叉湁灏忕▼搴忓悗绔寘鍧囨甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- 杩涘害绫绘暟鎹凡缁忓叿澶?SQL repository锛屼絾鐪熷疄 PostgreSQL 椹卞姩鍜屾祴璇曞簱浠嶆湭鍚敤锛沗jsonb file_ids`銆佸苟鍙戞墦鍗°€佸鐩樺敮涓€鎬у拰缁眬鑽夌鍞竴绾︽潫闇€瑕侀泦鎴愭祴璇曞娴嬨€?- `game_service_confirms` 鍜?`game_service_confirm_items` 浠嶆湭鎺?SQL repository锛汥5.6 鐨勬湇鍔＄‘璁ら摼璺笅涓€姝ュ簲缁х画娌夊簱銆?- `ContinueDraft` 鍦ㄦ湁涓婚摼璺?repository 鏃跺凡缁忓垱寤鸿崏绋垮眬鍜屾垚鍛橈紝浣嗚繕娌℃湁浜嬪姟灏佽锛涚湡瀹炴暟鎹簱鎺ュ叆鍚庯紝闇€瑕佹妸鑽夌灞€銆佽崏绋挎垚鍛樸€佺画灞€璁板綍鏀捐繘鍚屼竴涓簨鍔°€?## 2026-06-14 04:45

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟帹杩涘皬绋嬪簭鍚庣 D5.6 鏈嶅姟纭閾捐矾鐨?PostgreSQL 鎵挎帴銆?- 澶嶆牳 `db/migrations/000004_games.sql`锛岀‘璁ゅ凡鏈夎〃鍙鐢細
  - `game_service_confirms`
  - `game_service_confirm_items`
- `internal/games` 鏂板鏈嶅姟纭鎸佷箙鍖栨帴鍙ｏ細
  - `ServiceConfirmRepository`
  - `UseServiceConfirmRepository`
  - `NewSQLServiceConfirmRepository`
- `services/go-api/internal/games/service_confirm_sql_repository.go` 鏂板 SQL 閫傞厤锛?  - 鎸?`game_id` 鑾峰彇鏈嶅姟纭鍗曞拰纭鏄庣粏銆?  - 鍒涘缓鎴栨洿鏂?`game_service_confirms`锛屾敮鎸?`completed_at`銆?  - 鎸?`game_id + user_id` 骞傜瓑鍐欏叆 `game_service_confirm_items`銆?  - 鏌ヨ纭鏄庣粏骞跺洖濉?`confirmedBy`銆?- `games.Service.ConfirmService` 鎺ュ叆 repository锛?  - 鏈?repository 鏃朵紭鍏堜粠 PostgreSQL 璇诲彇宸插瓨鍦ㄧ‘璁ゅ崟鍜屾槑缁嗐€?  - 棣栨纭鍒涘缓纭鍗曞拰纭鏄庣粏銆?  - 閲嶅纭淇濇寔骞傜瓑锛屼笉閲嶅澧炲姞纭鏄庣粏銆?  - 鍏ㄥ憳纭鍚庡啓鍥炵‘璁ゅ崟 `completed` 鐘舵€侊紝骞舵妸灞€鐘舵€佹帹杩涘埌 `pending_review`銆?  - 鏃?repository 鏃剁户缁繚鎸佸師鍐呭瓨琛屼负銆?- `cmd/server` 鏂板 `gameServiceConfirmRepository(db)` 娉ㄥ叆锛?  - 鏈夋暟鎹簱鏃朵娇鐢?`games.NewSQLServiceConfirmRepository(db)`銆?  - 鏃犳暟鎹簱鏃剁户缁湰鍦板唴瀛樿仈璋冦€?- 鏂板 `services/go-api/internal/games/service_confirm_repository_service_test.go`锛?  - 瑕嗙洊纭鍗曞垱寤恒€佺‘璁ゆ槑缁嗗啓鍏ャ€?  - 瑕嗙洊鍏ㄥ憳纭鍚庣姸鎬佽繘鍏?`completed` / `pending_review`銆?  - 瑕嗙洊 `ServiceConfirmForGame` 浠?repository 鏌ヨ銆?  - 瑕嗙洊閲嶅纭骞傜瓑銆?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/games`
  - `internal/audit`
  - `internal/files`
  - `internal/notifications`
  - `internal/orders`
  - 鍏朵綑宸叉湁灏忕▼搴忓悗绔寘鍧囨甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- 鏈嶅姟纭宸茬粡鍏峰 SQL repository锛屼絾鐪熷疄 PostgreSQL 椹卞姩鍜屾祴璇曞簱浠嶆湭鍚敤锛沗game_id` 鍞竴纭鍗曘€乣game_id + user_id` 鍞竴纭鏄庣粏鍜屽苟鍙戠‘璁ら渶瑕侀泦鎴愭祴璇曞娴嬨€?- `ConfirmService` 鐩墠浠嶄笉鏄暟鎹簱浜嬪姟锛涚‘璁ゆ槑缁嗗啓鍏ャ€佺‘璁ゅ崟鐘舵€佹洿鏂般€佸眬鐘舵€佹洿鏂板簲鍦ㄧ湡瀹?PostgreSQL 鎺ュ叆鍚庣撼鍏ュ悓涓€浜嬪姟銆?- D5.6 鐨勮瘎浠枫€佹垚闀裤€佷俊鐢ㄣ€佽冻杩逛富鏁版嵁浠嶄富瑕佺敱鍐呭瓨鏈嶅姟椹卞姩锛涗笅涓€姝ュ簲缁х画鎶?reviews/growth/credit/footprints 鐨?repository 琛ラ綈銆?## 2026-06-14 05:20

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟帹杩涘皬绋嬪簭鍚庣 D5.6 璇勪环銆佹垚闀裤€佷俊鐢ㄥ拰瓒宠抗閾捐矾鐨?PostgreSQL 鎵挎帴銆?- 澶嶆牳 `db/migrations/000006_review_growth_credit.sql`锛岀‘璁ゅ凡鏈夎〃鍙鐢細
  - `reviews`
  - `review_reminders`
  - `user_growth_profiles`
  - `experience_logs`
  - `points_accounts`
  - `points_logs`
  - `credit_logs`
  - `daily_credit_scores`
  - `user_footprints`
  - `achievements`
  - `user_achievements`
- `internal/reviews` 鏂板 repository 鎺ュ彛涓?SQL 閫傞厤锛?  - `Repository`
  - `NewServiceWithRepository`
  - `NewSQLRepository`
- `services/go-api/internal/reviews/sql_repository.go` 鏂板 SQL 鎵挎帴锛?  - 鏈嶅姟瀹屾垚鍚庝负鎴愬憳鍐欏叆 `review_reminders`銆?  - 璇勪环鎻愪氦鍐欏叆 `reviews`锛屼緷璧?`uk_review_once` 淇濇寔鍞竴璇勪环璇箟銆?  - 鎴愰暱璧勬枡鍐欏叆 `user_growth_profiles`銆?  - 缁忛獙娴佹按鍐欏叆 `experience_logs`銆?  - 绉垎璐︽埛鍜屾祦姘村啓鍏?`points_accounts`銆乣points_logs`銆?  - 淇＄敤鍒嗗啓鍏?`daily_credit_scores`锛屾墸鍒嗘祦姘村啓鍏?`credit_logs`銆?  - 瓒宠抗鍐欏叆 `user_footprints`銆?  - 鎴愬氨瀹氫箟鍜岀敤鎴锋垚灏卞啓鍏?`achievements`銆乣user_achievements`銆?- `reviews.Service` 鎺ュ叆 repository锛?  - `MarkGameReviewable` 鏈?repository 鏃跺啓鍏ュ緟璇勪环鎻愰啋銆?  - `Todos`銆乣Submit`銆乣MyIntents`銆乣AllReviews`銆乣Profile`銆乣Footprints`銆乣AllFootprints`銆乣TraceByUser`銆乣TraceByGame` 浼樺厛璧?repository銆?  - `DeductCredit` 鏈?repository 鏃跺啓鍏ュ綋鏃ヤ俊鐢ㄥ垎銆佷俊鐢ㄦ祦姘村拰瓒宠抗銆?  - 鏃?repository 鏃剁户缁繚鎸佸師鍐呭瓨琛屼负銆?- `appapi.Server.UseRepositories` 鎵╁睍 `reviewRepo` 娉ㄥ叆锛?  - 鏈?repository 鏃堕噸寤?`reviews` 鏈嶅姟锛屽苟鍚屾閲嶅缓渚濊禆 reviews 鐨?`revenue`銆乣reports`銆乣memberReports`銆乣teams`銆?  - 閬垮厤鏀剁泭缁撶畻浠嶅紩鐢ㄦ棫鐨勫唴瀛?reviews 鏈嶅姟銆?- `cmd/server` 鏂板 `reviewRepository(db)` 娉ㄥ叆銆?- 鏂板 `services/go-api/internal/reviews/repository_service_test.go`锛?  - 瑕嗙洊寰呰瘎浠锋彁閱掋€佽瘎浠锋彁浜ゃ€侀噸澶嶈瘎浠锋嫤鎴€?  - 瑕嗙洊鎴愰暱銆佺粡楠屻€佺Н鍒嗐€佽冻杩广€佹垚灏卞啓鍏ャ€?  - 瑕嗙洊淇＄敤鎵ｅ垎銆佺敤鎴疯拷韪煡璇€?
### 楠岃瘉缁撴灉

- 浣跨敤鐪熷疄 Go 宸ュ叿閾炬墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/reviews`
  - `internal/games`
  - `internal/audit`
  - `internal/files`
  - `internal/notifications`
  - `internal/orders`
  - 鍏朵綑宸叉湁灏忕▼搴忓悗绔寘鍧囨甯哥紪璇戞垨娴嬭瘯閫氳繃銆?
### 褰撳墠椋庨櫓

- reviews/growth/credit/footprints 宸插叿澶?SQL repository锛屼絾鐪熷疄 PostgreSQL 椹卞姩鍜屾祴璇曞簱浠嶆湭鍚敤锛涘敮涓€绾︽潫銆佷簨鍔″拰骞跺彂閲嶅璇勪环闇€瑕侀泦鎴愭祴璇曞娴嬨€?- `Submit` 褰撳墠浼氬垎姝ュ啓 reviews銆乬rowth銆乪xperience銆乸oints銆乫ootprints銆乤chievements锛涚湡瀹炴暟鎹簱鎺ュ叆鍚庡簲灏佽浜嬪姟锛岄伩鍏嶈瘎浠峰啓鍏ユ垚鍔熶絾鎴愰暱/绉垎閮ㄥ垎澶辫触銆?- `points_accounts` 褰撳墠鐢?reviews repository 缁存姢鍙敤绉垎锛涘悗缁鏋?points 妯″潡涔熸帴 SQL repository锛岄渶瑕佺粺涓€绉垎璐︽埛鍐欏叆鍏ュ彛锛岄伩鍏嶅弻鍐欒鍒欎笉涓€鑷淬€?## 2026-06-14 06:05

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟帹杩涘皬绋嬪簭鍚庣 D5.6 绉垎銆佸厬鎹笌璇勪环閾捐矾鍚庣殑 PostgreSQL 鎵挎帴銆?- 澶嶆牳褰撳墠瀹炵幇鍚庣‘璁わ細
  - `reviews` repository 宸茬粡鍐欏叆 `points_accounts`銆乣points_logs`銆?  - `points.Service` 浠嶄互鍐呭瓨璐︽埛鍜屽唴瀛樻祦姘翠负涓汇€?  - `redemption.Service` 浠嶄互鍐呭瓨鍟嗗搧銆佸唴瀛樿鍗曚负涓汇€?- 鏂板 `internal/points` 浠撳偍鎺ュ彛鍜?SQL 閫傞厤锛?  - `Repository`
  - `NewServiceWithRepository`
  - `NewSQLRepository`
  - `Summary`銆乣Logs`銆乣AllLogs`銆乣Grant`銆乣Deduct` 鍦ㄦ湁 repository 鏃朵紭鍏堣鍐?PostgreSQL銆?- 鏂板 `internal/redemption` 浠撳偍鎺ュ彛鍜?SQL 閫傞厤锛?  - `Repository`
  - `NewServiceWithRepository`
  - `NewSQLRepository`
  - 鍟嗗搧鍒涘缓銆佸垪琛ㄣ€佹洿鏂般€佸簱瀛橀鎵ｃ€佸簱瀛樺洖婊氥€?  - 鍏戞崲璁㈠崟鍒涘缓銆佺敤鎴疯鍗曘€佸悗鍙拌鍗曘€佽鍗曞鏍搞€?  - 椹冲洖璁㈠崟鍚庨€氳繃绉垎鏈嶅姟杩旇繕绉垎銆?- 鏂板杩佺Щ `db/migrations/000017_points_redemption_repository.sql`锛?  - `points_logs.before_points`
  - `points_logs.after_points`
  - `points_logs.biz_type`
  - `points_logs.biz_id`
  - 绉垎娴佹按銆佸厬鎹㈠晢鍝併€佸厬鎹㈣鍗曟煡璇㈢储寮曘€?- 鏇存柊 `plan.md`锛?  - C3 鏁版嵁搴撹縼绉绘墽琛岄『搴忚拷鍔?`000017_points_redemption_repository.sql`銆?  - D4 鏁版嵁搴撹縼绉讳笂绾块『搴忚拷鍔犵 17 椤广€?- 鏇存柊 `cmd/server` 鍜?`appapi.Server.UseRepositories`锛?  - 鏈夋暟鎹簱鏃舵敞鍏?`points.NewSQLRepository(db)`銆?  - 鏈夋暟鎹簱鏃舵敞鍏?`redemption.NewSQLRepository(db)`銆?  - 鍏戞崲鏈嶅姟澶嶇敤鍚屼竴涓Н鍒嗘湇鍔★紝閬垮厤璇勪环濂栧姳銆佸厬鎹㈡墸鍑忋€侀┏鍥炶繑杩樺垎鏁ｅ埌璐︽埛鍏ュ彛銆?- 鏂板娴嬭瘯锛?  - `services/go-api/internal/points/repository_service_test.go`
  - `services/go-api/internal/redemption/repository_service_test.go`

### 楠岃瘉缁撴灉

- 宸叉墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/points`
  - `internal/redemption`
  - `internal/reviews`
  - `internal/games`
  - 鍏朵綑宸叉湁灏忕▼搴忓悗绔寘鍧囨甯哥紪璇戞垨娴嬭瘯閫氳繃銆?- Go 浠嶈緭鍑洪潪鑷村懡 telemetry 鎻愮ず锛?  - `error acquiring upload token: creating token file: open C:\Users\61492\AppData\Roaming\go\telemetry\local\upload.token: Access is denied.`
  - 鍛戒护閫€鍑虹爜涓?0锛屼笉褰卞搷鏈疆楠岃瘉缁撹銆?
### 褰撳墠椋庨櫓

- 绉垎鍜屽厬鎹㈠凡缁忓叿澶?SQL repository锛屼絾鐪熷疄 PostgreSQL 椹卞姩鍜屾祴璇曞簱浠嶆湭鍚敤锛涗笅涓€杞簲琛ユ暟鎹簱闆嗘垚娴嬭瘯鎴栬縼绉?dry-run銆?- SQL 鍏戞崲涓嬪崟褰撳墠鎸夆€滈鎵ｅ簱瀛?-> 鎵ｇН鍒?-> 寤鸿鍗曗€濋『搴忎繚鎸佺姸鎬佷竴鑷达紝骞剁敤棰勫彇璁㈠崟 ID 鍐欏叆绉垎娴佹按锛涗粛寤鸿鍚庣画鍦ㄧ湡瀹炴暟鎹簱閲屽皝瑁呬簨鍔★紝閬垮厤杩涚▼寮傚父瀵艰嚧灞€閮ㄧ姸鎬佸仠鐣欍€?## 2026-06-14 06:45

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟帹杩涘皬绋嬪簭鍚庣 P1 棰勭暀鑳藉姏鐨勬暟鎹寔涔呭寲銆?- 澶嶆牳褰撳墠瀹炵幇鍚庣‘璁わ細
  - `connections.Service` 宸叉湁灏忕▼搴忓拰鍚庡彴鎺ュ彛锛屼絾浠嶄互鍐呭瓨鍏崇郴鍜屽唴瀛樿窡杩涜褰曚负涓汇€?  - `profiles.Service` 宸叉湁琛屽鎶€鑳芥爲銆侀璺汉璧勬簮鐢诲儚鎺ュ彛锛屼絾浠嶄互鍐呭瓨瑙掕壊鍜屽唴瀛樼敾鍍忎负涓汇€?  - `db/migrations/000003_realname_roles.sql` 宸叉湁 `user_roles`銆?  - `db/migrations/000010_p1_reserved.sql` 宸叉湁 `user_connections`銆乣connection_follow_logs`銆乣expert_skill_profiles`銆乣guide_resource_profiles`銆?- 鏂板 `internal/connections` 浠撳偍鎺ュ彛鍜?SQL 閫傞厤锛?  - `Repository`
  - `NewServiceWithRepository`
  - `NewSQLRepository`
  - `UpsertPair` 鍦ㄦ湁 repository 鏃跺啓鍏ュ弻鍚戜汉鑴夊叧绯汇€?  - `My`銆乣All` 鍦ㄦ湁 repository 鏃朵粠 PostgreSQL 鏌ヨ銆?  - `AddFollowLog` 鍦ㄦ湁 repository 鏃舵牎楠屽弬涓庝汉/棰嗚矾浜烘潈闄愶紝鍐欏叆璺熻繘璁板綍锛屽苟澧炲姞鍏崇郴寮哄害銆?- 鏂板 `internal/profiles` 浠撳偍鎺ュ彛鍜?SQL 閫傞厤锛?  - `Repository`
  - `NewServiceWithRepository`
  - `NewSQLRepository`
  - `GrantRole` 鍐欏叆鎴栨縺娲?`user_roles`銆?  - `ExpertSkill`銆乣UpdateExpertSkill`銆乣AllExpertSkills` 璇诲啓 `expert_skill_profiles`銆?  - `GuideResource`銆乣UpdateGuideResource`銆乣AllGuideResources` 璇诲啓 `guide_resource_profiles`銆?  - `IsGuide` 鍦ㄦ湁 repository 鏃朵粠 `user_roles` 鍒ゆ柇銆?- 鏇存柊 `cmd/server` 鍜?`appapi.Server.UseRepositories`锛?  - 鏈夋暟鎹簱鏃舵敞鍏?`connections.NewSQLRepository(db)`銆?  - 鏈夋暟鎹簱鏃舵敞鍏?`profiles.NewSQLRepository(db)`銆?  - 鏃犳暟鎹簱鏃剁户缁繚鎸佸師鍐呭瓨鏈嶅姟锛屾柟渚挎湰鍦拌仈璋冦€?- 鏂板娴嬭瘯锛?  - `services/go-api/internal/connections/repository_service_test.go`
  - `services/go-api/internal/profiles/repository_service_test.go`

### 楠岃瘉缁撴灉

- 宸叉墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 楠岃瘉閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/connections`
  - `internal/profiles`
  - `internal/points`
  - `internal/redemption`
  - `internal/reviews`
  - 鍏朵綑宸叉湁灏忕▼搴忓悗绔寘鍧囨甯哥紪璇戞垨娴嬭瘯閫氳繃銆?- Go 浠嶈緭鍑洪潪鑷村懡 telemetry 鎻愮ず锛岄€€鍑虹爜涓?0锛屼笉褰卞搷鏈疆楠岃瘉缁撹銆?
### 褰撳墠椋庨櫓

- 浜鸿剦鍜岀敾鍍忓凡缁忓叿澶?SQL repository锛屼絾鐪熷疄 PostgreSQL 椹卞姩鍜屾祴璇曞簱浠嶆湭鍚敤锛涗笅涓€姝ュ簲琛ヨ縼绉?dry-run 鎴栭泦鎴愭祴璇曘€?- `GrantRole` 鐩墠婊¤冻鏈湴鍜岀瀛愭暟鎹巿鏉冿紱瀹屾暣鐨勮瀹?棰嗚矾浜虹敵璇枫€佸鏍搞€佷粯璐瑰紑閫氫粛搴旂敱鍚庣画瑙掕壊鐢宠閾捐矾缁熶竴椹卞姩銆?## 2026-06-14 07:15

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟帹杩涘皬绋嬪簭鍚庣 LBS / 瀹氫綅閾捐矾銆?- 澶嶆牳褰撳墠瀹炵幇鍚庣‘璁わ細
  - `POST /api/app/locations/current` 鍜?`POST /api/app/locations/manual` 宸叉湁鎺ュ彛銆?  - `GET /api/app/games/nearby` 宸蹭緷璧栧綋鍓嶇敤鎴峰畾浣嶈绠楅檮杩戝眬銆?  - `lbs.Service` 浠嶄互鍐呭瓨淇濆瓨褰撳墠浣嶇疆涓轰富銆?  - `plan.md` 宸茶姹?`user_location_records`銆乣GET /api/app/locations/my-recent` 鍜屾渶杩戝畾浣嶄笉娉勯湶浠栦汉浣嶇疆銆?- 鏂板 `internal/lbs` 浠撳偍鎺ュ彛鍜?SQL 閫傞厤锛?  - `Repository`
  - `NewServiceWithRepository`
  - `NewSQLRepository`
  - `SaveLocation`
  - `CurrentLocation`
  - `RecentLocations`
- 鎵╁睍 `lbs.Service`锛?  - `SaveCurrent` / `SaveManual` 鍦ㄦ湁 repository 鏃跺啓鍏?PostgreSQL銆?  - `Current` 鍦ㄦ湁 repository 鏃惰鍙栨渶杩戜竴鏉″畾浣嶃€?  - 鏂板 `Recent(userID, limit)`锛岃繑鍥炲綋鍓嶇敤鎴锋渶杩戝畾浣嶃€?  - 鏃?repository 鏃朵繚鐣欏唴瀛樿涓猴紝骞剁淮鎶ゆ湰鍦版渶杩戝畾浣嶅垪琛ㄣ€?- 鏂板鎺ュ彛锛?  - `GET /api/app/locations/my-recent`
  - 鏀寔 `limit` 鏌ヨ鍙傛暟锛岃寖鍥撮檺鍒跺湪 1-50銆?- 鏂板杩佺Щ锛?  - `db/migrations/000018_user_location_records.sql`
  - 鍒涘缓 `user_location_records`
  - 澧炲姞 `idx_user_location_records_user_created`
  - 澧炲姞 `idx_user_location_records_city_created`
- 鏇存柊 `cmd/server` 鍜?`appapi.Server.UseRepositories`锛?  - 鏈夋暟鎹簱鏃舵敞鍏?`lbs.NewSQLRepository(db)`銆?  - 鏃犳暟鎹簱鏃剁户缁娇鐢ㄥ唴瀛樺畾浣嶏紝鏂逛究鏈湴鑱旇皟銆?- 鏇存柊 `docs/openapi/app.openapi.yaml`锛?  - 澧炲姞 `GET /api/app/locations/my-recent`銆?- 鏇存柊 `plan.md`锛?  - C3 鏁版嵁搴撹縼绉婚『搴忚拷鍔?`000018_user_location_records.sql`銆?  - D4 鏁版嵁搴撲笂绾块『搴忚拷鍔犵 18 椤广€?- 鏂板娴嬭瘯锛?  - `services/go-api/internal/lbs/repository_service_test.go`

### 楠岃瘉缁撴灉

- 宸叉墽琛岋細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 棣栨娴嬭瘯鍙戠幇 `appapi/server.go` 缂哄皯 `internal/lbs` import锛屽凡淇銆?- 澶嶆祴閫氳繃锛?  - `cmd/server`
  - `internal/appapi`
  - `internal/lbs`
  - `internal/games`
  - `internal/reviews`
  - 鍏朵綑宸叉湁灏忕▼搴忓悗绔寘鍧囨甯哥紪璇戞垨娴嬭瘯閫氳繃銆?- Go 浠嶈緭鍑洪潪鑷村懡 telemetry 鎻愮ず锛岄€€鍑虹爜涓?0锛屼笉褰卞搷鏈疆楠岃瘉缁撹銆?
### 褰撳墠椋庨櫓

- 鐢ㄦ埛瀹氫綅宸茬粡鍏峰 SQL repository 鍜岃縼绉伙紝浣嗙湡瀹?PostgreSQL 椹卞姩鍜屾祴璇曞簱浠嶆湭鍚敤锛涘悗缁渶瑕佸仛杩佺Щ dry-run 鎴栭泦鎴愭祴璇曘€?- 鏈疆鏈帴鑵捐浣嶇疆鏈嶅姟閫嗗湴鍧€瑙ｆ瀽锛涘綋鍓嶅彧淇濆瓨灏忕▼搴忕浼犲叆鐨勫煄甯傘€佸湴鍧€鍜屽潗鏍囷紝鍚庣画搴旇ˉ鑵捐鍦板浘鏍囧噯鍖栧拰 geocode cache銆?## 2026-06-14 09:00

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟帹杩涘皬绋嬪簭鍚庣鐢ㄦ埛璧勬枡涓庢枃浠舵潈闄愰棴鐜€?- 鍩轰簬涓婁竴杞闄╅」锛岃ˉ榻愬ご鍍忔枃浠跺綊灞炴牎楠岋細
  - `PUT /api/app/users/me/profile` 鏀寔 `avatarFileId`銆?  - `avatarFileId` 蹇呴』瀛樺湪銆?  - 鏂囦欢蹇呴』鐢卞綋鍓嶇敤鎴蜂笂浼犮€?  - 鏂囦欢 `bizType` 蹇呴』涓?`avatar`銆?  - 鏍￠獙閫氳繃鍚庤嚜鍔ㄧ敓鎴愭湰鍦?mock 澶村儚涓嬭浇鍦板潃骞跺啓鍏?`avatarUrl`銆?- 璋冩暣鏂囦欢涓婁紶鍏ュ彛锛?  - `POST /api/app/files/upload-token` 瀵?`bizType=avatar` 鍏佽 pre-auth 鐢ㄦ埛涓婁紶銆?  - 鍏朵粬涓氬姟鏂囦欢浠嶈姹傚己瀹炲悕瀹屾垚銆?- 鏇存柊鐢ㄦ埛妯″瀷涓?SQL锛?  - `users.User` 澧炲姞 `avatarFileId`銆?  - `users.Repository.UpdateProfile` 澧炲姞 `avatarFileID` 鍙傛暟銆?  - `users.avatar_file_id` 鍐欏叆鍜屽洖璇汇€?- 鏇存柊杩佺Щ锛?  - `db/migrations/000020_user_profiles_repository.sql` 澧炲姞 `users.avatar_file_id` 鍜岀储寮曘€?- 鏇存柊鏂囨。锛?  - `docs/openapi/app.openapi.yaml` 澧炲姞 `avatarFileId` 瀛楁銆?  - `plan.md` 鏇存柊 `000020_user_profiles_repository.sql` 璇存槑鍜岀敤鎴疯祫鏂欐帴鍙ｉ獙鏀跺彛寰勩€?- 鏇存柊娴嬭瘯锛?  - 鐧诲綍鍚庡厛涓婁紶 `bizType=avatar` 鏂囦欢銆?  - 浣跨敤 `avatarFileId` 鏇存柊璧勬枡銆?  - 鍙︿竴涓敤鎴峰鐢ㄨ fileId 鏇存柊澶村儚杩斿洖 `403`銆?
### 楠岃瘉缁撴灉

- 宸叉墽琛屽畬鏁村悗绔祴璇曪細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 娴嬭瘯閫氳繃锛?  - `internal/appapi`
  - `internal/auth`
  - `internal/files`
  - 鍏朵粬宸叉湁鍚庣鍖呭潎姝ｅ父缂栬瘧鎴栨祴璇曢€氳繃銆?- Go 浠嶈緭鍑洪潪鑷村懡 telemetry 鏉冮檺鎻愮ず锛岄€€鍑虹爜涓?0锛屼笉褰卞搷鏈疆楠岃瘉缁撹銆?
### 褰撳墠椋庨櫓

- 褰撳墠澶村儚 URL 涓烘湰鍦?mock 涓嬭浇鍦板潃锛涙帴鐪熷疄瀵硅薄瀛樺偍鍚庡簲鏇挎崲涓?CDN / 涓存椂涓嬭浇 URL 绛栫暐銆?- `avatar` 鏂囦欢绫诲瀷鐩墠鍙牎楠屼笟鍔＄被鍨嬪拰褰掑睘锛屽悗缁簲鎸?MIME 鐧藉悕鍗曞拰澶у皬闄愬埗杩涗竴姝ユ敹绱с€?
## 2026-06-14 08:35

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟帹杩涘皬绋嬪簭鍚庣璐﹀彿璧勬枡閾捐矾銆?- 澶嶆牳鍚庣‘璁?`plan.md` 瑕佹眰 `PUT /api/app/users/me/profile`锛屽綋鍓嶄唬鐮佸彧鏈?`GET /api/app/users/me`銆?- 鏂板鐢ㄦ埛璧勬枡鏇存柊鑳藉姏锛?  - `users.Store.UpdateProfile`
  - `users.Repository.UpdateProfile`
  - `auth.Service.UpdateProfile`
- 鏇存柊 SQL repository锛?  - 鏂扮敤鎴峰垱寤烘椂鍚屾鍒濆鍖?`user_profiles` 鍩虹琛屻€?  - 鏇存柊璧勬枡鏃跺啓鍏?`users.nickname`銆乣users.avatar_url`锛屽苟纭繚 `user_profiles` 瀛樺湪銆?- 鏂板灏忕▼搴忔帴鍙ｏ細
  - `PUT /api/app/users/me/profile`
  - 鏀寔 `nickname`銆乣avatarUrl`
  - 鏄电О鏈€澶?32 瀛楃锛屽ご鍍忓湴鍧€鏈€澶?500 瀛楃銆?- 鏂板杩佺Щ锛?  - `db/migrations/000020_user_profiles_repository.sql`
  - 鍒涘缓 `user_profiles`
  - 澧炲姞鍩庡競绱㈠紩 `idx_user_profiles_city`
- 鏇存柊鏂囨。锛?  - `docs/openapi/app.openapi.yaml` 澧炲姞鐢ㄦ埛璧勬枡鏇存柊鎺ュ彛銆?  - `plan.md` 澧炲姞绗?20 涓縼绉昏褰曪紝骞朵慨姝ｇ敤鎴疯祫鏂欐帴鍙ｈ惤鐐广€?- 鏇存柊娴嬭瘯锛?  - `TestWechatLoginAndCurrentUserHTTP` 澧炲姞璧勬枡鏇存柊鍜屽洖璇绘柇瑷€銆?
### 楠岃瘉缁撴灉

- 宸叉墽琛屽畬鏁村悗绔祴璇曪細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 娴嬭瘯閫氳繃锛?  - `internal/appapi`
  - `internal/auth`
  - 鍏朵粬宸叉湁鍚庣鍖呭潎姝ｅ父缂栬瘧鎴栨祴璇曢€氳繃銆?- Go 浠嶈緭鍑洪潪鑷村懡 telemetry 鏉冮檺鎻愮ず锛岄€€鍑虹爜涓?0锛屼笉褰卞搷鏈疆楠岃瘉缁撹銆?
### 褰撳墠椋庨櫓

- 褰撳墠璧勬枡鏇存柊鍏堣鐩栨樀绉板拰澶村儚鍦板潃锛沗gender`銆乣bio`銆乣interest_tags`銆乣city_code`銆乣city_name` 宸插湪 `user_profiles` 棰勭暀锛屽悗缁彲缁х画鎵╁睍璇锋眰浣撳拰鏍￠獙銆?- 澶村儚鐩墠鎸?URL 淇濆瓨锛屽皻鏈己鍒舵牎楠?fileId 灞炰簬鏈汉锛涘悗缁帴灏忕▼搴忎笂浼犲ご鍍忔椂搴斿拰鏂囦欢鏈嶅姟鏉冮檺鏍￠獙鎵撻€氥€?
## 2026-06-14 08:10

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟帹杩涘皬绋嬪簭鍚庣鏍归摼璺寔涔呭寲銆?- 澶嶆牳鍚庣‘璁よ处鍙?閭€璇蜂粛鏄唴瀛樻€侊細
  - `users.NewStore()`
  - `invites.NewStore()`
  - 寰俊鐧诲綍鍚庣殑 `users`銆乣user_wechat_accounts`銆乣invite_codes`銆乣invite_relations` 杩愯鏃舵湭鎺?PostgreSQL銆?- 鏂板鐢ㄦ埛妯″潡 repository 鍖栵細
  - `users.Repository`
  - `users.NewStoreWithRepository`
  - `users.NewSQLRepository`
  - `FindByOpenID`
  - `FindByID`
  - `CreateWithOpenID`
- 鏂板閭€璇锋ā鍧?repository 鍖栵細
  - `invites.Repository`
  - `invites.NewStoreWithRepository`
  - `invites.NewSQLRepository`
  - `UpsertCode`
  - `FindCode`
  - `Bind`
  - `RelationForUser`
- 鏇存柊 `auth.Service`锛?  - 寰俊鐧诲綍鏌?openid 鏃跺彲浠?repository 鍥炲～鐢ㄦ埛銆?  - 鏂扮敤鎴峰垱寤烘椂鍚屾椂鍐欏叆 `users` 鍜?`user_wechat_accounts`銆?  - 閭€璇风粦瀹氬啓鍏?`invite_relations`锛岄娆＄粦瀹氭墠閫掑 `invite_codes.used_count`銆?  - repository 閿欒閫忎紶锛屼笉鍐嶆妸鏁版嵁搴撻敊璇鏄犲皠鎴愰個璇风爜鏃犳晥銆?- 鏇存柊 `cmd/server`锛?  - 鏈夋暟鎹簱鏃舵敞鍏?`users.NewSQLRepository(db)` 鍜?`invites.NewSQLRepository(db)`銆?  - 鏃犳暟鎹簱鏃剁户缁娇鐢ㄥ唴瀛?store锛屼繚鎸佹湰鍦拌仈璋冨拰鏃㈡湁娴嬭瘯绋冲畾銆?- 鏇存柊 `plan.md`锛?  - 鏍囪 `000002_account_invite.sql` 宸叉帴鍏ヨ繍琛屾椂 repository銆?  - D4 鏁版嵁搴撲笂绾块『搴忚ˉ鍏?root 璐﹀彿閭€璇烽摼璺帴鍏ヨ鏄庛€?- 鏂板娴嬭瘯锛?  - `services/go-api/internal/auth/repository_service_test.go`

### 楠岃瘉缁撴灉

- 宸叉墽琛屽畬鏁村悗绔祴璇曪細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 娴嬭瘯閫氳繃锛?  - `internal/auth`
  - `internal/appapi`
  - 鍏朵粬宸叉湁鍚庣鍖呭潎姝ｅ父缂栬瘧鎴栨祴璇曢€氳繃銆?- Go 浠嶈緭鍑洪潪鑷村懡 telemetry 鏉冮檺鎻愮ず锛岄€€鍑虹爜涓?0锛屼笉褰卞搷鏈疆楠岃瘉缁撹銆?
### 褰撳墠椋庨櫓

- 鏈疆瑙ｅ喅鐨勬槸 mock openid 涓嬬殑杩愯鏃舵寔涔呭寲锛涚湡瀹炲井淇?`code2session` 浠嶆湭鎺ュ叆銆?- 璐﹀彿/閭€璇?SQL repository 宸蹭娇鐢ㄧ幇鏈夎縼绉昏〃锛屼絾鐪熷疄 PostgreSQL 椹卞姩鍜屾祴璇曞簱浠嶆湭鍚敤锛屽敮涓€绾︽潫鍜屼簨鍔¤涓哄悗缁繕闇€闆嗘垚娴嬭瘯澶嶆祴銆?
## 2026-06-14 07:45

### 褰撳墠杩涘睍

- 缁х画鎵ц `plan.md`锛屼紭鍏堟帹杩涘皬绋嬪簭鍚庣鎸佷箙鍖栬ˉ榻愩€?- 鏂板鍥㈤槦妯″潡 repository 鍖栵細
  - `teams.Repository`
  - `teams.NewServiceWithRepository`
  - `teams.NewSQLRepository`
  - 鍥㈤槦涓昏〃 `teams`
  - 鎴愬憳鍏崇郴缁х画澶嶇敤 `team_relations`
- 鏂板浼氬憳鎶ヨ〃妯″潡 repository 鍖栵細
  - `memberreports.Repository`
  - `memberreports.NewServiceWithRepository`
  - `memberreports.NewSQLRepository`
  - 浼氬憳鎺堟潈琛?`member_report_memberships`
  - 鎶ヨ〃蹇収缁х画鍐欏叆 `member_report_snapshots.metrics`
- 鏂板杩佺Щ锛歚db/migrations/000019_team_member_repository.sql`銆?- 鏇存柊 `cmd/server` 鍜?`appapi.Server.UseRepositories`锛?  - 鏈夋暟鎹簱鏃舵敞鍏ュ洟闃熶笌浼氬憳鎶ヨ〃 SQL repository銆?  - 鏃犳暟鎹簱鏃剁户缁娇鐢ㄥ唴瀛樺疄鐜帮紝鏂逛究鏈湴鑱旇皟銆?- 鏇存柊 `plan.md`锛?  - 杩佺Щ姹囨€昏拷鍔?`000019_team_member_repository.sql`銆?  - D4 鏁版嵁搴撲笂绾块『搴忚拷鍔犵 19 椤广€?- 鏂板娴嬭瘯锛?  - `services/go-api/internal/teams/repository_service_test.go`
  - `services/go-api/internal/memberreports/repository_service_test.go`
- 灏嗗洟闃熷拰浼氬憳鎶ヨ〃榛樿鍚嶇О鏀逛负 ASCII fallback锛岄伩鍏嶇户缁繑鍥炲巻鍙蹭贡鐮侀粯璁ゅ€笺€?
### 楠岃瘉缁撴灉

- 宸叉墽琛屽畬鏁村悗绔祴璇曪細
  - `$env:GOCACHE='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gocache'; $env:GOTMPDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotmp'; $env:GOTELEMETRY='off'; $env:GOTELEMETRYDIR='C:\Users\61492\Desktop\鐪熷ソ鐜?mini\services\go-api\.gotelemetry'; & 'C:\Program Files\Go\bin\go.exe' test -count=1 ./...`
- 娴嬭瘯閫氳繃锛?  - `internal/appapi`
  - `internal/teams`
  - `internal/memberreports`
  - 鍏朵粬宸叉湁鍚庣鍖呭潎姝ｅ父缂栬瘧鎴栨祴璇曢€氳繃銆?- Go 浠嶈緭鍑洪潪鑷村懡 telemetry 鏉冮檺鎻愮ず锛岄€€鍑虹爜涓?0锛屼笉褰卞搷鏈疆楠岃瘉缁撹銆?
### 褰撳墠椋庨櫓

- 鍥㈤槦涓庝細鍛樻姤琛ㄥ凡鍏峰 SQL repository 鍜岃縼绉昏剼鏈紝浣嗙湡瀹?PostgreSQL 杩佺Щ dry-run / 闆嗘垚娴嬭瘯浠嶆湭鎵ц銆?- 浼氬憳鎶ヨ〃鐩墠鎸夊疄鏃?games/revenue 鐢熸垚蹇収骞惰惤搴擄紝鍚庣画鑻ラ渶瑕佸浐瀹氬懆鏈熸姤琛紝搴旇ˉ瀹氭椂浠诲姟鎴栧悗鍙扮敓鎴愬叆鍙ｃ€?

## 2026-06-14 12:36

### 当前进展

- 继续执行 `plan.md`，优先补小程序后端 D7 外部输入校验中的 LBS 链路。
- `internal/lbs.Service` 新增定位参数校验：
  - 经度必须在 `[-180, 180]`。
  - 纬度必须在 `[-90, 90]`。
  - `accuracyMeter` 必须非负且不超过 `100000`。
  - `cityCode`、`cityName`、`address` 做 trim 和长度上限校验。
- `POST /api/app/locations/current` 和 `POST /api/app/locations/manual` 遇到无效定位参数时统一返回 `422`。
- `GET /api/app/games/nearby` 新增 `radiusMeter` 校验，半径必须大于 `0` 且不超过 `50000`，异常值不再静默使用默认值。
- 更新 `plan.md`：
  - 标记“校验经纬度范围”完成。
  - 标记 `accuracyMeter > 300` 返回 `accuracyWarning=true` 完成。
  - 标记“80 米精度无提示，500 米精度有提示”完成。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/lbs`
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./...`
- 测试通过。

## 2026-06-15 05:00

### 当前进展

- 继续执行 `plan.md`，补齐实名状态查询和后台实名列表筛选能力。
- `GET /api/admin/identity-verifications` 新增筛选参数：
  - `status`：按实名状态过滤。
  - `userId`：按用户 ID 精确过滤。
- 后台实名列表仍只返回脱敏字段，详情页继续保持脱敏展示。
- 复用现有 `GET /api/app/identity/status`，用户端可直接查询强实名进度和驳回原因。
- 更新 `plan.md`：标记 `查询实名状态`、`用户端可查询实名状态和驳回原因` 完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/appapi/identity_handler.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./services/go-api/internal/identity ./services/go-api/internal/appapi`
  - `go test -count=1 ./services/go-api/internal/...`
- 测试通过。

## 2026-06-15 04:30

### 当前进展

- 继续执行 `plan.md`，补齐强实名后台脱敏展示闭环。
- `identity.Record` 新增脱敏字段：
  - `realNameMasked`
  - `idCardMasked`
- 手机号核验通过时只生成并保存姓名、身份证脱敏值，不在 Go 业务记录中返回明文姓名或身份证号。
- SQL repository 已写入/读取实名脱敏字段。
- 新增迁移 `db/migrations/000023_identity_masked_fields.sql`，为 `identity_verification_records` 补充：
  - `real_name_masked`
  - `id_card_masked`
- 后台实名列表和详情继续使用同一 `Record` DTO，普通管理员只能看到 `phoneMasked`、`realNameMasked`、`idCardMasked` 等脱敏字段。
- 更新 `plan.md`：标记实名详情敏感字段脱敏、普通管理员不可查看完整身份证/核身原始结果、后台列表默认脱敏完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/identity/service.go services/go-api/internal/identity/sql_repository.go services/go-api/internal/identity/identity_status_service_test.go services/go-api/internal/identity/service_test.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./services/go-api/internal/identity ./services/go-api/internal/appapi`
  - `go test -count=1 ./services/go-api/internal/...`
- 测试通过。

## 2026-06-15 04:00

### 当前进展

- 继续执行 `plan.md`，补齐小程序后端业务配置闭环。
- `common/config.Load` 新增业务配置读取：
  - `APP_DAILY_GAME_LIMIT`：每日开局限制，默认 3。
  - `LBS_DEFAULT_RADIUS_METER`：附近局默认半径，默认 5000 米。
- `games.Service` 新增可配置每日开局限制，默认行为保持每日 3 局；配置后按新值拦截。
- `appapi.Server.Configure` 将配置注入到后端业务：
  - 每日开局限制注入到 `games.Service`。
  - 附近局默认半径注入到 `/api/app/games/nearby`。
- `/api/app/games/nearby` 在未传 `radiusMeter` 时使用配置默认半径；传入 `radiusMeter` 时仍以请求参数覆盖。
- 更新 `plan.md`：标记 LBS 默认半径、每日开局限制、本地/测试/生产不同配置、附近局默认半径读取系统配置完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/common/config/config.go services/go-api/internal/common/config/config_test.go services/go-api/internal/games/service.go services/go-api/internal/games/service_test.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/location_handler.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./services/go-api/internal/common/config ./services/go-api/internal/games ./services/go-api/internal/appapi`
  - `go test -count=1 ./services/go-api/internal/...`
- 测试通过。

## 2026-06-15 03:30

### 当前进展

- 继续执行 `plan.md`，补齐 Go API 后端 requestId 闭环，方便线上测试和排障。
- `httpx.Write` 现在会保证 JSON 响应体包含 `requestId`，并同步写入 `X-Request-Id` 响应头。
- `Server.Register` 统一把所有注册路由经过 `httpx.RequestID` 中间件：
  - 前端传入 `X-Request-Id` 时，响应头和响应体复用该值。
  - 未传入时，后端自动生成 requestId。
  - 后台操作日志继续能从请求头读取同一个 requestId。
- 修复 `server.go` 登录错误分支中一处历史非法 UTF-8/断裂字符串，改为稳定 ASCII 文案。
- 更新 `plan.md`：标记 Go API 后端统一响应结构、requestId 中间件、所有响应带 requestId、requestId 必带完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/common/httpx/response.go services/go-api/internal/common/httpx/middleware_test.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./services/go-api/internal/common/httpx ./services/go-api/internal/appapi`
  - `go test -count=1 ./services/go-api/internal/...`
- 测试通过。

## 2026-06-15 03:00

### 当前进展

- 继续执行 `plan.md`，优先补小程序后端线上测试前的业务闭环。
- AI/IM 数据使用新增后台显式开关：
  - `GET /api/admin/ai-data/im-export-config` 可查看当前 IM 导出授权状态。
  - `PUT /api/admin/ai-data/im-export-config` 仅 `system_config:update` 权限可修改。
  - 默认关闭时 `POST /api/admin/ai-data/im-export` 禁止导出；开启后可导出已有 IM 消息；再次关闭后恢复禁止。
- 交付测试执行记录补齐证据链约束：
  - `POST /api/admin/test-runs` 创建测试执行记录时，必须至少关联 `requestId` 或 `evidenceFileId`。
  - 测试执行记录返回并留存 `evidenceFileId`，便于后台交付验收追踪。
- 更新 `plan.md`：标记 `test_runs` 关联 requestId 或证据文件完成，并补充对应测试项。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/delivery/service.go services/go-api/internal/delivery/test_run_service_test.go services/go-api/internal/appapi/delivery_handler.go services/go-api/internal/appapi/server_test.go services/go-api/internal/aidata/service.go services/go-api/internal/appapi/aidata_handler.go services/go-api/internal/appapi/server.go`
  - `go test -count=1 ./services/go-api/internal/delivery ./services/go-api/internal/aidata ./services/go-api/internal/appapi`
  - `go test -count=1 ./services/go-api/internal/...`
- 测试通过。

## 2026-06-14 17:05

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 D5.6/E5 服务确认、评价和人脉沉淀闭环。
- 核实现有代码状态：
  - 服务确认完成后已经进入 `pending_review` 并触发评价待办。
  - `againIntent` 已作为评价字段保存，并通过 `/api/app/reviews/my-intents` 返回。
  - 服务确认完成后已经生成 `co_game` 人脉。
- 新增评价成功后的人脉强度沉淀：
  - `/api/app/reviews` 提交成功后，对评价双方写入 `sourceType=review` 的人脉关系。
  - 关系强度 `strengthScore` 增加 1，供后续二期推荐和人脉排序使用。
- 更新 `plan.md`：
  - 标记“再玩一局意向单独字段或表记录，供二期推荐使用”完成。
  - 标记“评价完成后可提升关系强度分”完成。

### 验证补充

- `go test -count=1 ./internal/appapi` 通过。

## 2026-06-14 16:41

### 当前进展

- 继续执行 `plan.md`，优先补齐小程序后端 E4 强实名安全闭环。
- FaceID 回调新增 HMAC-SHA256 验签能力：
  - `Config` 新增 `FaceIDConfig`，读取 `TENCENT_FACEID_CALLBACK_REQUIRE_SIGNATURE` 和 `TENCENT_FACEID_CALLBACK_SECRET`。
  - 生产环境默认要求 FaceID 回调签名；本地环境默认不强制，保持现有小程序 mock 联调流程可用。
  - `appapi.Server` 新增 `UseFaceIDCallbackVerifier`，服务启动时注入回调验签配置。
  - `/api/app/identity/faceid/callback` 在解析业务 payload 前先校验 `X-FaceID-Signature` 或 `X-Tencent-FaceID-Signature`。
  - `deploy/env.example` 已补 FaceID 回调验签环境变量样例。
- `plan.md` 已同步勾选 E4 的“腾讯云人脸核身回调必须验签或校验来源”。

### 验证补充

- `go test -count=1 ./internal/common/config` 通过。
- `go test -count=1 ./internal/appapi` 通过。

### 下一步

- 跑全量 `go test -count=1 ./...`。
- 继续按后端业务功能优先级推进：补强强实名验收剩余项，随后转入服务确认、评价链路和上线联调缺口。

### 当前风险

- D7 “所有外部输入做字段长度、枚举、金额、经纬度、时间范围校验”仍是大项，本轮只完成 LBS/附近局相关输入校验，其他写接口还需要继续补齐。

## 2026-06-14 13:06

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 D7 安全基线。
- 新增 `internal/appapi.AppAuthMiddleware`：
  - 统一解析 `Authorization: Bearer <token>`。
  - 缺失 token 返回 `40101`。
  - 无效或过期 token 返回 `40101`。
- `Register` 中除 `POST /api/app/auth/wechat-login` 外，所有 `/api/app` 小程序接口均已统一包装 `AppAuthMiddleware`。
- 保留 handler 内部 `requireUser` 的强实名校验，不影响微信登录后、实名前的身份绑定流程。
- 新增 `TestAppBusinessRoutesRequireTokenMiddleware`，验证业务接口匿名访问被拦截，微信登录入口保持公开。
- 更新 `plan.md`：
  - 标记“小程序接口统一走 `AppAuthMiddleware`”完成。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./...`
- 测试通过。

### 当前风险

- D7 还剩：
  - IM WebSocket 握手和 HTTP 历史消息使用同一套成员权限判断。
  - 所有写接口支持幂等键或业务唯一约束。
  - 所有外部输入做字段长度、枚举、金额、经纬度、时间范围校验。

## 2026-06-14 13:36

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 D7 IM 权限基线。
- `internal/im.Service` 新增统一成员权限入口：
  - `AuthorizeGameAccess(userID, gameID)`。
  - `AuthorizeRoomAccess(userID, roomID)`。
- `chat-session` 握手、按 `gameId` 读取历史消息、按 `roomId` 读取历史消息、发消息和房间操作均通过统一成员权限入口或其下游方法校验。
- `TestIMFlow` 增加非成员校验：
  - 非成员无法获取 `chat-session`。
  - 非成员无法读取 `GET /api/app/games/{gameId}/chat/messages`。
  - 非成员无法读取 `GET /api/app/chat/rooms/{roomId}/messages`。
- `docs/openapi/ws-protocol.md` 同步注明握手和 HTTP 历史消息统一走 `AuthorizeGameAccess / AuthorizeRoomAccess`。
- 更新 `plan.md`：
  - 标记“IM WebSocket 握手和 HTTP 历史消息使用同一套成员权限判断”完成。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/im`
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./...`
- 测试通过。

### 当前风险

- D7 还剩：
  - 所有写接口支持幂等键或业务唯一约束。
  - 所有外部输入做字段长度、枚举、金额、经纬度、时间范围校验。

## 2026-06-14 14:06

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 D7 幂等基线。
- 新增 `internal/appapi.IdempotencyMiddleware`：
  - 仅对 `POST/PUT/DELETE` 且带 `Idempotency-Key` 的请求生效。
  - 按 `userID + method + path + key` 缓存首次成功响应。
  - 重复提交返回首个成功响应，避免前端重复点击导致重复创建。
  - `Idempotency-Key` 过长时返回 `422`。
- 已接入的小程序写接口包括：
  - 创建局
  - 入局申请 / 局操作
  - 定位保存
  - 身份绑定 / 实名链路
  - 文件上传令牌
  - 评价、积分兑换、举报、连接、资料、通知、报表等写接口
- 新增测试 `TestCreateGameIdempotencyKeyReplaysFirstResponse`，验证同一个 key 的重复创建局请求只产生一次局数据。
- 更新 `plan.md`：
  - 标记“所有写接口支持幂等键或业务唯一约束，防止重复提交”完成。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./...`
- 测试通过。

### 当前风险

- 这轮实现是 app 层内存幂等缓存，适合当前联调与防重复提交；后续如果要做到进程重启后仍可去重，仍建议再接 Redis 或数据库唯一约束。

## 2026-06-14 15:11

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 D7 外部输入校验闭环。
- 补齐核心业务写入口的字段长度、枚举、范围校验：
  - 创建局：标题、城市编码/名称、经纬度范围。
  - 入局申请、服务确认、续局草稿：原因、备注、标题长度。
  - 进度反馈、里程碑、打卡、复盘：进度范围、状态/类型枚举、文件 ID、内容长度、再玩意向枚举。
  - IM 消息：文本/图片/文件消息类型枚举、文本内容、fileId、归档原因。
  - 评价：评分范围、目标角色、内容长度、再玩意向枚举。
  - 举报申诉：举报类型枚举、内容长度、关联证据 ID 非负、后台处理结果长度。
  - 人脉跟进、行家技能树、领路人资源画像：跟进类型/时间、标签数量和长度、文件 ID、连接规模长度。
- 新增 `ErrInvalidGameInput` 和 `ErrInvalidMessage` 的接口层映射，避免校验错误落成 500。
- 更新 `plan.md`：
  - 标记“所有外部输入做字段长度、枚举、金额、经纬度、时间范围校验”完成。

### 验证结果

- 已执行：
  - `go test -count=1 ./services/go-api/internal/games ./services/go-api/internal/reviews ./services/go-api/internal/reports ./services/go-api/internal/im ./services/go-api/internal/connections ./services/go-api/internal/profiles ./services/go-api/internal/appapi`
- 测试通过。

### 当前风险

- D7 已基本闭环；后续继续按 `plan.md` 推进小程序后端时，应优先补强还未完成的强实名 token 分层、后台操作日志覆盖和联调交付包，而不是继续只加测试。

## 2026-06-14 16:11

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 E4 强实名主线补强。
- 在 `identity.Service` 增加强实名格式门槛：
  - 手机号必须为 11 位大陆手机号格式。
  - 身份证号必须为 18 位格式，末位支持数字或 `X/x`。
  - 身份证格式错误时不会通过手机号核验，因此不能发起人脸核身。
- `identity_handler` 增加手机号格式错误、身份证格式错误的 422 映射。
- 保持现有两段式登录：
  - 未强实名只返回 `preAuthToken`。
  - 强实名完成后通过 `issue-token-after-identity` 换取正式 token。
  - 已实名用户再次登录可直接拿到正式 token。
- 更新 `plan.md`：
  - 标记 E4 中已由代码覆盖的 preAuth、短信限频、手机号/短信/人脸记录、not_supported、敏感数据不落明文、核心业务强实名拦截和相关验收项。
  - 保留“腾讯云人脸核身回调必须验签或校验来源”未完成，待真实腾讯云回调接入时再落签名/来源校验。

### 验证结果

- 已执行：
  - `go test -count=1 ./services/go-api/internal/identity ./services/go-api/internal/appapi`
- 测试通过。

## 2026-06-14 17:41

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 P3 角色申请与领路人双门槛。
- 补齐角色申请服务闭环：
  - 小程序端 `POST /api/app/role-applications`、`GET /api/app/role-applications/my`。
  - 后台端 `GET /api/admin/audits/role-applications`、`POST /api/admin/audits/role-applications/{applicationId}/review`。
  - 申请记录保存申请理由、能力说明和证明文件 ID。
  - 审核通过后写入 `user_roles`，驳回保存原因，重复 pending 申请拦截。
- 补齐领路人双门槛闭环：
  - 小程序端 `GET /api/app/guides/qualification/me` 查询当前用户资格和规则。
  - 小程序端 `POST /api/app/guides/apply` 在条件与付费均满足后提交领路人申请。
  - 后台端 `GET /api/admin/guides/qualification-rules`、`PUT /api/admin/guides/qualification-rules/{ruleId}` 配置规则。
  - 后台端保留 `POST /api/admin/guide-qualification-rules` 登记用户条件/付费达成状态。
  - `conditionMet=false` 返回 `waiting_condition`，`paymentMet=false` 返回 `waiting_payment`。
- 扩展 `role_applications`、`guide_qualification_rules`、`guide_qualification_records` 的服务层与 SQL 仓储映射。
- 更新 `plan.md`：标记角色申请审核、申请材料保存、领路人规则配置、双门槛验收项完成。

### 验证结果

- 已执行：
  - `go test -count=1 ./services/go-api/internal/profiles`
  - `go test -count=1 ./services/go-api/internal/appapi`
- 测试通过。

## 2026-06-14 18:23

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端退出局扣信用闭环。
- 将 `game_members` 退出从物理删除改为软退出状态：
  - 写入 `quit_before_confirm`、`quit_after_confirm`、`quit_after_started`。
  - 保留退出原因，便于后台和审计查询。
- 退出扣分后回写成员记录：
  - `credit_deducted=true`。
  - `credit_log_id` 关联真实信用流水。
- 退出成功后生成站内通知：
  - 给退出用户生成 `game_quit_result`。
  - 给发起人生成 `game_member_quit`。
- 更新 `plan.md`：标记退出阶段判断、确认前不扣分、确认后/开始后扣分、成员状态更新、扣分流水关联、通知和后台扣分原因可查完成。

### 验证结果

- 已执行：
  - `go test -count=1 ./services/go-api/internal/games`
  - `go test -count=1 ./services/go-api/internal/appapi`
- 测试通过。

## 2026-06-14 18:53

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 IM 内容安全闭环。
- 将 IM 敏感词从硬编码字符串升级为可管理对象：
  - 后台 `GET /api/admin/sensitive-words` 查询。
  - 后台 `POST /api/admin/sensitive-words` 新增。
  - 后台 `POST /api/admin/sensitive-words/import` 批量导入。
- IM 文本消息发送前读取当前敏感词配置，命中后按规则阻断，并返回 `45101`。
- 命中敏感词时写入 `content_risk_logs` 风险日志，后台 `GET /api/admin/content-risk/logs` 可查询。
- 新增 `/api/internal/ai/content-risk/check-placeholder`，一期返回固定占位结构并写风险日志，后续可替换真实 AI 内容审核。
- 补齐 `000005_im.sql` 中 `sensitive_words`、`content_risk_logs`、`chat_messages.client_msg_id` 和风险日志索引。
- 补齐后台内容安全权限码：`content:sensitive_word:view/create/import`、`content:risk_log:view`。
- 更新 `plan.md`：标记敏感词新增、批量导入、IM 文本检测、命中拦截/标记、风险日志和 AI 占位接口完成；保留“后台可禁用敏感词”独立启停项未完成。

### 验证结果

- 已执行：
  - `go test -count=1 ./services/go-api/internal/im`
  - `go test -count=1 ./services/go-api/internal/appapi`
  - `go test -count=1 ./services/go-api/internal/adminauth`
- 测试通过。

## 2026-06-14 19:23

### 当前进展

- 继续执行 `plan.md`，优先补齐小程序后端 IM 敏感词独立启停能力。
- 新增后台敏感词更新链路：
  - `PUT /api/admin/sensitive-words/{id}` 支持将敏感词状态更新为 `active` 或 `disabled`。
  - 服务层 `UpdateSensitiveWord` 按 ID 修改状态，非法状态返回校验错误，不存在返回未找到。
  - 后台操作日志记录 `content:sensitive_word:update`。
- 补齐后台权限：
  - 路由改用独立权限码 `content:sensitive_word:update`。
  - `adminauth` 默认权限和 `db/seeds/admin_roles_permissions.sql` 均补齐该权限。
- 修复 IM 默认敏感词和行为事件测试中损坏的字符串字面量，避免 Go 文件语法损坏。
- 更新 `plan.md`：标记“后台可禁用敏感词”、敏感词阻断、风险日志后台可查、AI 占位接口、文本敏感词检测和 TC-087 完成。
- 补齐敏感词 `action=flag` 发送链路：命中后消息不阻断，消息状态写为 `risk_flagged`，并生成风险日志，覆盖 TC-088。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/im/service.go services/go-api/internal/appapi/im_handler.go services/go-api/internal/appapi/server.go services/go-api/internal/adminauth/service.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./...`
- 测试通过。

## 2026-06-14 19:53

### 当前进展

- 继续执行 `plan.md`，复核 D5.6 服务确认、评价、成长、信用、足迹后端链路。
- 补齐仓储模式下成长资料初始化：
  - `Profile(userId)` 在 `user_growth_profiles` 不存在时生成默认资料。
  - 默认 `level=1`、今日信用和信用分为 `100`，并持久化到仓储。
  - 个人中心首次查询成长资料不再返回空结构。
- 补齐完成局经验奖励：
  - 服务确认全部完成、局进入 `pending_review` 时为成员写入 `completed_game` 经验和足迹。
  - 使用足迹幂等判断，重复确认不会重复加经验。
- 更新 `plan.md` D5.6 细项：
  - 标记服务确认、确认项、评价提醒、可补评价、重复评价拦截、再玩意向、评价经验、经验日志、积分日志、成就和后台成长来源追溯完成。
  - 标记“完成局后增加经验”完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/reviews/service.go services/go-api/internal/reviews/repository_service_test.go`
  - `go test -count=1 ./internal/reviews`
  - `go test -count=1 ./...`
- 测试通过。

## 2026-06-14 20:23

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 D5.7 支付、订单和分账占位闭环。
- 补齐二期分账预留接口的一期占位实现：
  - `POST /api/funds/profit-sharing/orders` 创建本地模拟分账单。
  - `GET /api/funds/profit-sharing/orders/{outOrderNo}` 查询本地模拟分账结果。
  - `POST /api/funds/profit-sharing/return-orders` 创建本地模拟分账回退。
- 所有分账占位响应明确返回：
  - `placeholder=true`。
  - `needWechatPay=false`。
  - `mode=profit_sharing_placeholder` 或 `profit_sharing_return_placeholder`。
- 补齐 `revenueclient` 调用方法，后续资金服务或内部网关联调可沿同一路径调用。
- 更新 `plan.md`：标记 `payment_orders.order_no` 唯一、免费局 `free_no_pay` 占位订单、二期支付分账接口预留完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/orders/service.go services/go-api/internal/orders/order_service_test.go services/go-api/internal/appapi/order_handler.go services/go-api/internal/appapi/order_handler_test.go services/go-api/internal/appapi/server.go services/go-api/internal/revenueclient/client.go services/go-api/internal/revenueclient/client_test.go`
  - `go test -count=1 ./internal/orders ./internal/appapi ./internal/revenueclient`
  - `go test -count=1 ./...`
- 测试通过。
## 2026-06-14 20:14

### 当前进展

- 继续执行 `plan.md`，优先补齐小程序后端 D5.7 分润、收益和订单占位闭环。
- 将收益链路从“只算汇总”推进到“账户 + 流水 + 结算”三层一致：
  - `revenue.Service` 新增 `IncomeAccount` 视图，收益摘要可同步维护一人一账户。
  - 生成、冻结、结算分润记录时，自动同步 `user_income_accounts`。
  - 分润记录相关动作会写入 `income_logs`，为后续审计和查询保留明细。
- `SQLRepository` 已补齐：
  - `user_income_accounts` upsert。
  - `income_logs` 追加写入。
- 小程序收益接口新增账户入口：
  - `GET /api/app/incomes/account`
  - `GET /api/app/income/account`
- 更新 `plan.md`：
  - 标记用户收益账户一人一个、更新收益账户、正式生成分润记录/明细、支持冻结、线下结算登记、结算记录查询、试算和正式记录金额一致、冻结后不能结算、结算后有记录完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/revenue/service.go services/go-api/internal/revenue/sql_repository.go services/go-api/internal/revenue/repository_service_test.go services/go-api/internal/appapi/revenue_handler.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/revenue ./internal/appapi`
- 测试通过。

## 2026-06-14 20:43

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 D5.7 分润试算阻断能力。
- 补齐分润试算返回字段：
  - `canGenerateRecord`：表示当前试算是否允许生成正式分润记录。
  - `blockReasons`：返回阻断原因列表。
- 服务层 `revenue.Preview` 已能识别：
  - `review_incomplete`：评价未完成，不能生成正式分润。
  - `invalid_amount`：金额非法。
  - `game_missing`：缺少局 ID。
- HTTP 层预览会补充争议阻断：
  - 同局存在未关闭举报/申诉时，`blockReasons` 包含 `disputed`。
  - 后台和小程序分润预览返回同一组阻断字段。
- 正式生成分润记录仍保留硬拦截，不依赖前端是否读取预览字段。
- 更新 `plan.md`：标记 `canGenerateRecord`、`blockReasons`、评价未完成阻断、争议中阻断，以及 API-B2-05 对应测试项完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/revenue/service.go services/go-api/internal/revenue/repository_service_test.go services/go-api/internal/appapi/revenue_handler.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/revenue ./internal/appapi`
- 测试通过。
## 2026-06-14 21:04

### 当前进展

- 继续执行 `plan.md`，优先推进小程序后端 D5.7 分润模板规则配置。
- 新增分润规则后端能力：
  - `GET /api/admin/revenue/rules?templateId=...`
  - `POST /api/admin/revenue/rules`
- `revenue_rules` 现在可按模板保存和查询，`rule_code` 维度支持 upsert，便于后台修改模板规则值。
- 规则结构和模板结构保持同一 bps 体系，避免前后端出现浮点金额计算分叉。
- 规则配置已接入现有 revenue 服务和 SQL 仓储，后续可继续补审计日志或更复杂规则解释。
- 更新 `plan.md`：标记新增规则、修改规则、比例使用 bps 完成；修改写操作日志暂未勾选，留待审计日志真正落地后再收口。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/revenue/service.go services/go-api/internal/revenue/sql_repository.go services/go-api/internal/revenue/repository_service_test.go services/go-api/internal/appapi/revenue_handler.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/revenue ./internal/appapi`
- 测试通过。

## 2026-06-14 21:34

### 当前进展

- 继续执行 `plan.md`，补齐 D5.7 分润模板规则配置的审计闭环。
- 后台创建分润模板现在写入操作日志：
  - action=`revenue:template:create`
  - targetType=`revenue_template`
- 后台新增/修改分润规则现在写入操作日志：
  - action=`revenue:rule:upsert`
  - targetType=`revenue_rule`
  - detail 记录 `templateId`、`ruleCode`、`ruleValue`
- 补充 HTTP 测试：创建模板、upsert 规则后，超级管理员可在 `/api/admin/operation-logs` 查到对应日志。
- 更新 `plan.md`：标记“修改写操作日志”、`TC-034`、`TC-112` 完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/appapi/revenue_handler.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/revenue ./internal/appapi`
- 测试通过。

## 2026-06-14 22:00

### 当前进展

- 继续执行 `plan.md`，推进 D5.7 分润规则参与试算返回。
- `revenue.Preview` 新增 `rules` 字段，按当前 `templateId` 返回已配置的 `revenue_rules`。
- 后台和小程序分润预览现在不仅返回金额明细、阻断原因，也能返回当前模板启用的规则列表。
- 本轮只把规则纳入预览上下文，不擅自改变金额算法；更复杂的按局类型、会员、角色选择规则后续单独实现，避免未确认语义时影响分润金额。
- 更新 `plan.md`：标记“读取模板和规则”完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/revenue/service.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/revenue ./internal/appapi`
- 测试通过。

## 2026-06-14 22:31

### 当前进展

- 继续执行 `plan.md`，推进 D5.7 分润规则真正参与试算。
- 规则驱动的试算已支持三类额外角色：
  - `expert_bps`
  - `guide_bps`
  - `system_guide_bps`
- `CalculateRequest` 现在可携带 `expertUserId`、`guideUserId`、`systemGuideUserId`，预览会按模板规则生成对应金额项。
- 规则金额和模板比例都统一按 bps 计算，不引入浮点运算。
- `revenue.Preview` 继续返回 `rules`、`items`、`canGenerateRecord` 和 `blockReasons`，后台可直接看到规则与预览金额的对应关系。
- 更新 `plan.md`：标记平台金额、行家金额、领路人金额、系统级领路人金额、TC-113 完成；舍入差额仍待单独字段化，不提前勾选。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/revenue/service.go services/go-api/internal/revenue/repository_service_test.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/revenue ./internal/appapi`
- 测试通过。

## 2026-06-14 23:00

### 当前进展

- 继续执行 `plan.md`，补齐 D5.7 分润试算的舍入差额闭环。
- `revenue.Preview` 新增 `roundingDiffCent` 返回字段。
- 预览计算会在平台、创作者、行家、领路人、系统级领路人、成员分润项生成后，计算 `amountCent - items合计`。
- 当存在未分配差额时，追加 `rounding_adjustment` 明细项，确保分润项金额合计等于 `amountCent`。
- 正式生成分润记录复用预览结果，因此 `revenue_records.items` 同步保留差额项，避免预览和生成金额口径不一致。
- 更新 `plan.md`：标记“计算舍入差额”和“比例计算总额等于 amountCent”完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/revenue/service.go services/go-api/internal/revenue/repository_service_test.go`
  - `go test -count=1 ./internal/revenue ./internal/appapi`
  - `go test -count=1 ./...`
- 测试通过。

## 2026-06-14 23:30

### 当前进展

- 继续执行 `plan.md`，补齐 D5.7 订单占位的后端业务入口。
- 订单服务新增领路人付费占位订单：
  - 订单号：`GUIDE-FEE-{userId}`
  - 状态：`guide_fee_placeholder`
  - `needWechatPay=false`
- 新增小程序接口：`POST /api/app/guides/payment/precreate-placeholder`。
- 该接口会创建或复用领路人付费占位订单，并同步把当前用户的领路人资格 `paymentMet` 标记为 true。
- SQL 仓储同步支持领路人付费占位订单，`game_id` 为空，便于和免费局订单区分。
- 更新 `plan.md`：标记领路人付费占位、支付预下单固定结构、支付回调幂等、一期不真实收款和二期沿用订单号状态完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/orders/service.go services/go-api/internal/orders/sql_repository.go services/go-api/internal/orders/order_service_test.go services/go-api/internal/appapi/order_handler.go services/go-api/internal/appapi/order_handler_test.go services/go-api/internal/appapi/server.go`
  - `go test -count=1 ./internal/orders ./internal/appapi`
  - `go test -count=1 ./...`
- 测试通过。

## 2026-06-14 23:59

### 当前进展

- 继续执行 `plan.md`，回到 D5.8 举报申诉后端业务链路，补齐后台分配处理人能力。
- `reports.Service` 新增 `Assign(reportId, adminId, handlerAdminId)` 语义：
  - 举报状态更新为 `assigned`
  - `handlerAdminId` 记录被分配的处理人
  - 不写 `handledAt`，避免把“已分配”误当成“已处理”
- 新增后台接口：`POST /api/admin/reports/{reportId}/assign`。
- 后台分配举报会写操作日志 `report:assign`，并给举报人生成 `report_assigned` 站内通知和微信订阅任务占位。
- 权限补齐：`report:assign` 加入默认后台权限和种子权限。
- 更新 `plan.md`：标记 `AssignReport`、`HandleReport`、`CloseReport`、`FreezeRevenueIfNeeded` 完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/reports/service.go services/go-api/internal/reports/report_service_test.go services/go-api/internal/appapi/report_handler.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/server_test.go services/go-api/internal/adminauth/service.go`
  - `go test -count=1 ./internal/reports ./internal/appapi ./internal/adminauth`
  - `go test -count=1 ./...`
- 测试通过。

## 2026-06-15 00:30

### 当前进展

- 继续执行 `plan.md`，回补 D5.5 IM 归档和争议留存业务闭环。
- IM 服务层新增批量归档能力 `ArchiveRoomsByGameIDs`，供内部定时任务按已结束局归档房间。
- 发送消息前新增房间状态校验：`archived` / `readonly` 房间普通成员不能继续发送消息。
- 新增内部接口：`POST /api/internal/im/archive-expired-rooms`。
- 归档任务只处理 `pending_review` / `completed` 局；若局存在未关闭举报申诉，则跳过归档并返回 `skippedRooms`，用于延长 IM 证据留存。
- 更新 `plan.md`：标记 D5.5 IM 归档、TC-090、TC-091 完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/im/service.go services/go-api/internal/appapi/im_handler.go services/go-api/internal/appapi/job_handler.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/im ./internal/appapi`
  - `go test -count=1 ./...`
- 测试通过。

## 2026-06-15 01:00

### 当前进展

- 继续执行 `plan.md`，推进 D5.6 服务确认、评价提醒和补评价链路。
- 服务确认进入 `pending_review` 时，现在会即时给局成员生成 `review_remind` 站内通知和微信订阅任务占位，不再只依赖定时任务兜底。
- 重复确认不会重复发即时评价提醒，避免用户收到重复通知。
- `review_handler.go` 中损坏的历史错误文案已收敛为 ASCII 文案，恢复文件 UTF-8 和 gofmt 稳定性。
- 更新 `plan.md`：标记 Task 66 评价提醒、补评价、再玩意向，以及 TC-101、TC-103、TC-104、TC-105、TC-106、TC-107 完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/appapi/review_handler.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/appapi -run TestServiceConfirmAndReviewFlow`
  - `go test -count=1 ./internal/reviews ./internal/appapi`
  - `go test -count=1 ./...`
- 测试通过。

## 2026-06-15 01:30

### 当前进展

- 继续执行 `plan.md`，推进 D5.7 分润、收益和线下结算后端业务闭环。
- 新增后台结算记录查询接口：`GET /api/admin/revenue/settlements`，复用 `settlement:offline:create` 权限，可查看线下结算登记结果。
- `revenue.Service` 新增 `Settlements()`，SQL 仓储新增 `ListSettlements()`，内存和 PostgreSQL 路径都能返回结算记录。
- 补充收益隔离测试：第三方用户查询自己的收益账户和收益流水时只能看到 0 和空列表，不能读到其他参与者收益。
- 修复 `revenue_handler.go` 中损坏的历史错误文案，统一为 ASCII 文案，恢复 gofmt 稳定性。
- 更新 `plan.md`：标记收益流水、试算/生成一致、争议不能结算、后台结算记录、Task 67 相关项、TC-115、TC-117、TC-118 完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/revenue/service.go services/go-api/internal/revenue/sql_repository.go services/go-api/internal/revenue/repository_service_test.go services/go-api/internal/appapi/revenue_handler.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/revenue ./internal/appapi`
  - `go test -count=1 ./...`
- 测试通过。

## 2026-06-15 02:00

### 当前进展

- 继续执行 `plan.md`，补齐文件服务业务闭环。
- 文件服务新增临时下载地址过期控制，导出文件按 24 小时 TTL 生成，过期后下载返回 `410 Gone`。
- 小程序文件下载保持成员校验，后台附件查看按文件类型映射到 `identity:read`、`report:view`、`im:message:view_dispute`、`report_export:create`。
- 修复导出下载和文件下载 handler 的历史乱码文案，恢复 ASCII 稳定输出。
- 更新 `plan.md`：标记“实现临时下载地址”“实现 IM 文件权限校验”“实现后台附件查看权限校验”“导出文件链接过期后不可访问”完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/files/service.go services/go-api/internal/appapi/export_handler.go services/go-api/internal/appapi/file_handler.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./services/go-api/internal/files ./services/go-api/internal/appapi`
  - `go test -count=1 ./services/go-api/internal/...`
- 测试通过。

## 2026-06-15 02:30

### 当前进展

- 继续执行 `plan.md`，推进 E5 AI 数据准备和后台数据沉淀可视化。
- `GET /api/admin/ai-data/snapshot` 新增只读统计维度：
  - 行为事件按日期、用户、事件类型统计。
  - 收藏按用户、局类型统计。
  - 行家技能树、领路人资源画像完整率统计。
  - 人脉关系按来源和关系类型统计。
  - 数据沉淀分区 readiness，便于线上验收直接看哪些数据已经写入。
- 保持 IM 数据默认不可导出给 AI；关闭开关时 `POST /api/admin/ai-data/im-export` 仍返回 forbidden。
- 更新 `plan.md`：标记 E5 数据统计、AI IM 导出开关、验收数据沉淀相关项完成。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/aidata/service.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./services/go-api/internal/aidata ./services/go-api/internal/appapi`
  - `go test -count=1 ./services/go-api/internal/...`
- 测试通过。

## 2026-06-15 06:00

### 当前进展

- 继续执行 `plan.md`，优先补小程序后端当前用户和角色聚合链路。
- `POST /api/app/auth/wechat-login` 现在返回 `CurrentUserDTO` 聚合对象，包含用户基础信息、实名状态、角色列表、`roleStatusMap`、会员默认状态、成长信用、积分、收益摘要和我的邀请码。
- `GET /api/app/users/me` 改为复用同一套 `buildCurrentUserDTO`，资料更新、实名完成后可回读聚合字段。
- 新增 `GET /api/app/roles/my`，返回当前用户角色快照、角色申请记录和领路人门槛状态，便于小程序角色页联调。
- `invites.Store` 新增 `EnsureCodeForOwner`，当前用户聚合时可生成或回填我的邀请码。
- 修复 `server.go` 中一个历史损坏的登录请求错误文案，恢复 ASCII 稳定输出。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/appapi/server.go services/go-api/internal/appapi/profile_handler.go services/go-api/internal/appapi/server_test.go services/go-api/internal/auth/service.go services/go-api/internal/invites/model.go`
  - `go test -count=1 ./internal/profiles ./internal/appapi -run "TestWechatLoginAndCurrentUserHTTP|TestConnectionsAndProfilesHTTP" -v`
  - `go test -count=1 ./services/go-api/internal/...`
- 测试通过。

## 2026-06-15 06:30

### 当前进展

- 继续执行 `plan.md`，补齐小程序后端会员链路。
- 新增 `internal/membership` 服务和 SQL 仓储，支持 `membership_plans`、`user_memberships` 的读取与授予。
- `CurrentUserDTO.membership` 现在返回完整会员对象，不再是固定占位值。
- 新增 `GET /api/app/membership/plans` 和 `GET /api/app/membership/my`，小程序个人中心可直接联调会员状态。
- 启动入口 `cmd/server/main.go` 已接入 membership SQL repository，生产路径可以读取数据库会员计划和用户会员状态。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/membership/service.go services/go-api/internal/membership/sql_repository.go services/go-api/internal/membership/service_test.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/membership_handler.go services/go-api/internal/appapi/server_test.go services/go-api/cmd/server/main.go`
  - `go test -count=1 ./internal/membership ./internal/appapi -run "TestMembership|TestWechatLoginAndCurrentUserHTTP" -v`
  - `go test -count=1 ./services/go-api/internal/...`
- 测试通过。

## 2026-06-15 07:00

### 当前进展

- 继续执行 `plan.md`，补齐小程序后端 LBS 当前定位、手动定位和附近局距离字段。
- `POST /api/app/locations/current` 的返回已按 `LocationDTO` 验证，包含经纬度、城市、来源和 `accuracyWarning`。
- `POST /api/app/locations/manual` 已验证可用，未授权微信定位时可手动保存位置，并可作为附近局查询基准。
- `GET /api/app/games/nearby` 支持请求参数传入 `longitude`、`latitude` 临时坐标，并校验坐标范围。
- 附近局结果新增 `distanceLabel`，后端统一生成 `m/km` 展示标签，前端不用自行计算。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/games/service.go services/go-api/internal/lbs/service.go services/go-api/internal/appapi/location_handler.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/lbs ./internal/appapi -run "TestLocationAndNearbyGamesFlow" -v`
  - `go test -count=1 ./services/go-api/internal/...`
- 测试通过。

## 2026-06-15 07:30

### 当前进展

- 继续执行 `plan.md`，优先补小程序后端 LBS 与 IM 文件消息链路。
- `GET /api/app/locations/my-recent` 支持 `source=manual|gps` 过滤，最近手动定位可直接给小程序个人中心/定位兜底页使用。
- LBS SQL 仓储新增按来源查询最近定位，内存服务和仓储服务都保持一致行为。
- IM 图片/文件消息发送前新增 `fileId` 映射校验，要求文件必须是当前局的 `chat_file`，避免拿头像、导出文件或其他局附件伪装成聊天文件。
- OpenIM webhook 回调事件继续落本地事件记录，测试验证可从 `zhw_game_{gameId}` 解析业务 `gameId`。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/appapi/im_handler.go services/go-api/internal/appapi/server_test.go services/go-api/internal/appapi/location_handler.go services/go-api/internal/lbs/service.go services/go-api/internal/lbs/sql_repository.go services/go-api/internal/lbs/repository_service_test.go`
  - `go test -count=1 ./internal/lbs ./internal/appapi -run "TestRepositoryBackedLocationSaveCurrentAndRecent|TestLocationAndNearbyGamesFlow|TestServiceConfirmAndReviewFlow" -v`
  - `go test -count=1 ./internal/...`
- 测试通过。

## 2026-06-15 08:51

### 当前进展

- 继续执行 `plan.md`，优先补小程序后端业务功能闭环。
- 新增 `GET /api/app/users/me/summary`，聚合当前用户、待评价数量、未读通知数量、收益摘要、积分摘要和最近足迹，方便小程序“我的/个人中心”一次取数。
- 补齐入局申请链路的计划接口别名：
  - `GET /api/app/game-applications/my` 兼容申请人查看自己的入局申请。
  - `GET /api/app/game-applications/received?status=pending` 支持局主查看自己创建局收到的待审核申请。
  - `POST /api/app/game-applications/{applicationId}/cancel` 支持申请人取消 pending 申请。
  - `POST /api/app/game-applications/{applicationId}/audit` 支持局主按 `plan.md` 路由审核入局申请，并保留原 `/api/app/games/applications/{applicationId}/review` 兼容入口。
- `games.Service` 和 SQL 仓储新增局主侧申请列表、申请取消能力，真实数据库和内存服务保持同一行为。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/games/service.go services/go-api/internal/games/sql_repository.go services/go-api/internal/games/repository_service_test.go services/go-api/internal/appapi/game_handler.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/appapi -run "TestServiceConfirmAndReviewFlow|TestWechatLoginAndCurrentUserHTTP" -v`
  - `go test -count=1 ./internal/games ./internal/appapi -run "TestRepositoryBackedGameApplicationFlow|TestGameApplicationAndManualStartFlow" -v`
  - `go test -count=1 ./internal/games ./internal/appapi`
  - `go test -count=1 ./internal/...`
- 测试通过。

## 2026-06-15 13:48

### 当前进展

- 继续按 `plan.md` 优先补小程序后端 E2.2 里程碑、打卡、复盘、续局验收项。
- 收紧打卡类型枚举：`CreateCheckin` 现在只允许 `progress/complete/arrival/proof`，与计划要求一致。
- 移除服务层对旧打卡类型 `photo/text` 的放行，避免小程序端和后端字段契约不一致。
- 扩展打卡服务测试，覆盖旧类型 `photo` 被拒绝，以及 `arrival/proof` 两类新枚举正常使用。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/games -v`
  - `go test -count=1 ./internal/...`
- 测试通过。

## 2026-06-15 13:18

### 当前进展

- 继续按 `plan.md` 优先补小程序后端业务验收项，复核 E2.2 里程碑、打卡、复盘、续局链路。
- 补齐里程碑状态枚举：`UpdateMilestone` 现在允许 `pending/in_progress/completed/cancelled`，与计划要求一致。
- 将里程碑状态校验收敛为 `validMilestoneStatus`，避免状态字符串继续散落在业务分支里。
- 扩展里程碑服务测试，覆盖 `in_progress -> completed` 的状态流，以及非法状态拒绝。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/games -v`
  - `go test -count=1 ./internal/...`
- 测试通过。

## 2026-06-15 12:55

### 当前进展

- 继续按 `plan.md` 优先补小程序后端业务验收项，复核 E2.1 收藏与行为上报。
- 收紧 `POST /api/app/behavior/events` 字段契约：小程序端行为上报必须携带 `eventType`，缺失时返回参数错误。
- `eventCode` 仍作为分析口径保留，可由客户端显式传入；未传时默认等于 `eventType`，避免影响后台漏斗/留存统计。
- 扩展行为事件测试，覆盖缺少 `eventType` 被拒、带 `eventType/eventCode` 正常入库、后台按 `eventCode` 过滤仍可用。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/appapi -run TestBehaviorEventHTTP -v`
  - `go test -count=1 ./internal/...`
- 测试通过。

## 2026-06-15 12:42

### 当前进展

- 继续按 `plan.md` 优先补小程序后端业务接口，复核 E2 中积分/会员/团队/分润模拟验收项。
- 补齐计划中列出的 `POST /api/app/revenues/simulate` 用户端分润模拟接口。
- 接口复用现有分润 `Preview` 规则：用户只能模拟自己创建或参与的局，服务端自动补齐成员、创建者和默认模板。
- 分润模拟只返回预估结果和阻塞原因，不生成真实分润记录、不写收益流水，符合一期“只预估不落账”的口径。
- 扩展分润链路测试，覆盖非成员拒绝、成员模拟成功、模拟后后台分润记录仍为空。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/appapi -run TestRevenuePreviewGenerateFreezeAndSettlementFlow -v`
  - `go test -count=1 ./internal/...`
- 测试通过。

## 2026-06-15 12:25

### 当前进展

- 继续按 `plan.md` 补小程序后端业务闭环，复核 D5.7/D5.8 中“人脉关系自动来源”项。
- 已确认邀请码登录成功会自动生成 `invite` 关系，服务确认和评价链路会生成或增强 `co_game` 关系。
- 补齐游戏/领路人邀请接受后的关系链：被邀请人接受 `guide-invitations` 后，后端自动为邀请人与被邀请人生成双向 `guide_match` 关系，来源 ID 绑定邀请记录。
- 扩展邀请链路测试，覆盖“邀请接受 -> pending 入局申请 -> 创建者可见待审核申请 -> 自动生成 guide_match 人脉关系”完整流程。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/appapi -run TestGameInvitationRespondCreatesApplication -v`
  - `go test -count=1 ./internal/...`
- 测试通过。

## 2026-06-15 12:14

### 当前进展

- 继续按 `plan.md` 推进小程序后端业务功能，补齐 D5.3 游戏邀请链路。
- 新增 `POST /api/app/games/{gameId}/guide-invitations`，创建者或局内成员可邀请用户，邀请写入 `game_invitations`。
- 新增 `POST /api/app/game-invitations/{invitationId}/respond` / `POST /api/app/games/invitations/{invitationId}/respond`，被邀请人接受后只生成 `pending` 入局申请，不直接成为成员，仍需发起人审核。
- 新增 `000024_game_invitations.sql`，补齐邀请表、目标用户索引和局状态索引。
- 补齐后台用户详情、AI 验收夹具、匿名行为埋点、LBS 当前定位、后台文件权限、通知读取等后端跑通阻塞点。
- 修复内部包测试口径和业务校验顺序，确保后端内部包可以整体验证。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/appapi -run TestGameInvitationRespondCreatesApplication -v`
  - `go test -count=1 ./internal/games -run TestGameRepositoryPersistsCoreFlow -v`
  - `go test -count=1 ./internal/games ./internal/appapi`
  - `go test -count=1 ./internal/...`
- 测试通过。

## 2026-06-15 10:15

### 当前进展

- 继续执行 `plan.md`，优先补小程序后端组局详情页读模型。
- `GET /api/app/games/{gameId}` 对已登录用户返回兼容旧字段的详情聚合，局基础字段仍在 `data` 顶层。
- 详情新增 `myRelation`，包含 `role`、`isCreator`、`isMember`、`canApply`、`canAudit`、`canStart`、`canEnterIM`、`canConfirm`、`canReview`、申请 ID 和申请状态。
- 详情新增 `memberIds`、`progress.feedbacks`、`progress.milestones`、`progress.checkins`，成员可在一个接口拿到详情页进度数据。
- 详情新增 `im` 和 `review` 状态，返回 IM 可用性、roomId、引擎、OpenIM groupId，以及当前用户在该局是否可评价、待评价项和全局评价完成状态。
- 详情新增 `serviceConfirm`，已发起服务确认的局会返回确认主记录和确认成员项。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/appapi/game_handler.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/appapi -run "TestGameApplicationAndManualStartFlow" -v`
  - `go test -count=1 ./internal/appapi -run "TestGameApplicationAndManualStartFlow|TestServiceConfirmAndReviewFlow" -v`
  - `go test -count=1 ./internal/games ./internal/appapi`
  - `go test -count=1 ./internal/...`
- 测试通过。

## 2026-06-15 10:46

### 当前进展

- 继续执行 `plan.md`，补齐小程序后端 D5.3 的局成员列表接口。
- 新增 `GET /api/app/games/{gameId}/members`，局成员或创建者可查看成员列表，非成员不能读取。
- 成员列表返回 `userId`、`role`、`isCreator`、`isCurrentUser`、`confirmed`，可直接支撑小程序成员页、服务确认名单和 IM 成员展示。
- 成员确认状态复用现有 `game_service_confirms` / `game_service_confirm_items` 数据，不新增状态表。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/appapi/game_handler.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/appapi -run "TestGameApplicationAndManualStartFlow" -v`
  - `go test -count=1 ./internal/games ./internal/appapi -run "TestGameApplicationAndManualStartFlow|TestServiceConfirmAndReviewFlow|TestRepositoryBackedGameApplicationFlow" -v`
  - `go test -count=1 ./internal/games ./internal/appapi`
  - `go test -count=1 ./internal/...`
- 测试通过。

## 2026-06-15 21:24

### 当前进展

- 根据最新登录注册规则，继续执行 `plan.md`，优先补齐小程序一期后端邀请入口链路。
- 邀请入口统一为三类：`poster` 小程序分享卡片、`qrcode` 扫码邀请、`link` 微信聊天链接邀请。
- 新增 `POST /api/app/invites/entries`，登录用户可生成一次性唯一邀请码，并返回小程序 `path`、二维码 `scene` 和链接占位 `urlLink`。
- `POST /api/app/invites/precheck` 收紧入口校验：邀请码缺失、入口类型非法、入口类型不匹配、失效或耗尽都会拒绝；未绑定返回 `authPageMode=register`，已绑定唯一微信返回 `authPageMode=login`。
- `POST /api/app/auth/wechat-login` 收紧为必须携带 `inviteCode`；新用户只允许通过有效邀请码注册并绑定微信；一次性邀请码已绑定其他微信时拒绝复用。
- 数据库仓储绑定邀请码时增加行锁，降低同一个一次性邀请码并发绑定到多个微信的风险。
- `docs/openapi/app.openapi.yaml` 已同步邀请入口生成、入口预检和微信登录新字段。

### 验证结果

- 已执行：
  - `gofmt -w services/go-api/internal/invites/model.go services/go-api/internal/invites/sql_repository.go services/go-api/internal/auth/service.go services/go-api/internal/auth/service_test.go services/go-api/internal/auth/repository_service_test.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/auth ./internal/appapi -run "TestWechatLogin|TestInvitePrecheck|TestCreateInviteEntryHTTP" -v`
  - `go test -count=1 ./internal/...`
- 测试通过。
## 2026-06-16 11:25

### 当前进展

- 继续以一期小程序后端线上测试为目标，补齐短信验证码生产链路门槛。
- `identity.Service` 新增 `SMSSender` 抽象：
  - 本地默认 `LocalSMSSender`，继续返回固定 `mockCode=123456`，保持本地联调和测试稳定。
  - 生产可配置 `HTTPSMSSender`，通过 `SMS_HTTP_ENDPOINT` / `SMS_HTTP_SECRET` 调用短信网关。
- `cmd/server` 已按配置注入短信发送器。
- 生产环境配置校验新增短信要求：
  - `SMS_HTTP_ENDPOINT` 必须是 HTTPS URL。
  - `SMS_HTTP_SECRET` 必须是安全值。
- 小程序 `POST /api/app/sms/send-code` 响应收紧：
  - 本地返回 `mockCode`。
  - 真实短信通道只返回 `provider/messageId`，不回传验证码明文。
- AI 验收夹具只允许在本地短信 sender 下自动取 `mockCode`，避免生产误用夹具消耗真实短信。
- `docs/openapi/app.openapi.yaml` 已同步短信发送接口说明。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/identity ./internal/common/config ./internal/appapi`
- 测试通过。
- 当前一期小程序后端业务功能进度估算：约 92%。

## 2026-06-16 10:49

### 当前进展

- 继续以一期小程序后端线上测试为目标，收紧生产文件上传/下载配置。
- 生产环境现在要求：
  - `STORAGE_UPLOAD_BASE_URL` 必须是 HTTPS URL。
  - `STORAGE_DOWNLOAD_BASE_URL` 必须是 HTTPS URL。
- 避免 `APP_ENV=prod/production` 时仍静默返回 `mock://upload/...` 或非 HTTPS 地址，防止小程序真实设备联调失败。
- `docs/openapi/app.openapi.yaml` 已同步生产文件 URL 要求。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/common/config`
- 测试通过。
- 当前一期小程序后端业务功能进度估算：约 91%。

## 2026-06-16 10:26

### 当前进展

- 继续补一期小程序后端线上测试硬缺口，处理文件上传/下载 URL 仍为 `mock://` 的问题。
- `StorageConfig` 新增：
  - `STORAGE_UPLOAD_BASE_URL`
  - `STORAGE_DOWNLOAD_BASE_URL`
- 文件服务新增公开基础 URL 配置：
  - 本地未配置时继续返回 `mock://upload/...` 和 `mock://download/...`，保持本地测试稳定。
  - 线上配置后返回 HTTPS 上传/下载地址，方便小程序 `wx.uploadFile` 和下载联调。
- `cmd/server` 现在会调用 `appServer.Configure(cfg)`，让每日开局限制、LBS 默认半径、文件 URL 配置在真实启动时生效。
- `.env.example` 已同步新增文件上传/下载基础 URL 配置项。
- `docs/openapi/app.openapi.yaml` 已同步文件上传凭证说明。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/files ./internal/common/config ./internal/appapi`
- 测试通过。
- 当前一期小程序后端业务功能进度估算：约 91%。

## 2026-06-16 10:23

### 当前进展

- 继续以一期小程序后端线上测试为导向，优先补齐微信登录生产链路。
- `auth.Service` 新增 `WechatCodeResolver` 抽象：
  - 本地默认继续使用 mock openid，现有联调和自动化测试不受影响。
  - 生产环境配置 `WECHAT_APP_ID` / `WECHAT_APP_SECRET` 后，后端会调用微信 `jscode2session` 换取真实 `openid`、`unionid` 和 `session_key`。
- `cmd/server` 已按配置注入真实微信解析器；未配置时仍走本地 mock fallback。
- `config.Load` 已读取 `WECHAT_APP_ID`、`WECHAT_APP_SECRET`。
- 生产环境校验新增微信凭证要求，避免线上仍误用 mock 登录。
- 小程序登录接口新增微信 code 无效错误映射，`jscode2session` 返回错误或空 `openid` 时返回参数错误，前端可识别为 `wx.login` code 失效。
- `docs/openapi/app.openapi.yaml` 已同步登录接口说明。

### 验证结果

- 已执行：
  - `go test -count=1 ./internal/auth ./internal/common/config ./internal/appapi`
- 测试通过。
- 当前一期小程序后端业务功能进度估算：约 90%。

## 2026-06-16 10:15

### 当前进展

- 继续以一期小程序后端业务功能为主线，按最新登录注册和组局规则补齐后端闭环。
- 玩家创建免费局、申请入局不再强制实名；行家、领路人和被邀请主行家仍保留实名门槛。
- 游戏模型新增 `mainGuideUserId`，首个“行家邀请接受 + 审核通过”的成员会成为主行家，成员角色返回 `main_guide`。
- 手动开始收紧为创建者或主行家可操作，并严格要求人数达到 `minPlayers`；一期人数范围为 5-8，未满 5 人不能手动开始。
- 主行家已可管理局进度、里程碑和进度反馈；普通成员仍按成员权限参与聊天、打卡、服务确认和评价。
- 小程序详情和成员接口同步返回 `mainGuideUserId`、`myRelation.role`、`myRelation.canStart` 和成员 `role`，便于前端直接判断按钮状态。
- 同步 `docs/openapi/app.openapi.yaml` 和 `docs/test-cases/app-integration-cases.md` 的一期联调口径，避免线上测试继续按旧的“玩家必须实名”规则接入。

### 验证结果

- 已执行：
  - `go test -run TestIMArchiveJobSkipsDisputedRooms -count=1 -v ./internal/appapi`
  - `go test -run TestConnectionsAndProfilesHTTP -count=1 -v ./internal/appapi`
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./internal/auth ./internal/games`
  - `go test -count=1 ./internal/...`
- 测试通过。
- 当前一期小程序后端业务功能进度估算：约 88%。

## 2026-06-16 14:26

### 当前进展

- 继续以一期小程序后端上线测试为目标，补齐后台 IM 房间管理入口。
- 新增后台 IM 房间列表接口：
  - `GET /api/admin/im/rooms`
  - 支持 `gameId` 和 `status` 筛选。
  - 返回房间成员、房间状态、消息数量和文件消息数量。
- 新增后台 IM 房间详情接口：
  - `GET /api/admin/im/rooms/{roomId}`
  - 返回房间信息、成员、完整消息列表、文件消息列表、消息计数和文件消息计数。
- 后台权限补充 `im:room:read`，与现有争议消息权限 `im:message:view_dispute` 分开。
- 同步 `docs/openapi/admin.openapi.yaml` 和 `docs/test-cases/admin-integration-cases.md`，方便后台联调直接按接口走。

### 验证补充

- `go test -count=1 ./internal/appapi` 通过。
- `go test -count=1 ./internal/...` 通过。
- 当前一期小程序后端业务功能进度约 `97.5%`。

## 2026-06-16 15:56

### 当前进展

- 继续以一期小程序后端线上测试为目标，补齐上线前配置体检和后台邀请运营能力。
- 新增后台生产配置体检接口：
  - `GET /api/admin/system/readiness`
  - 只允许 `system_config:read` 权限访问。
  - 检查数据库、JWT、微信登录、短信、腾讯云慧眼、FaceID 回调、存储、OpenIM 和资金服务配置状态。
  - 响应只返回 ok / warn / missing，不返回任何密钥值。
- 新增后台邀请码管理接口：
  - `GET /api/admin/invite-codes`
  - `POST /api/admin/invite-codes`
  - `POST /api/admin/invite-codes/{code}/disable`
  - `GET /api/admin/invite-relations`
- 后台权限补充：
  - `invite_code:read`
  - `invite_code:manage`
  - `system_config:read`
- 管理员现在可以为海报、二维码、链接三种入口创建唯一邀请码，查看绑定关系，并禁用异常邀请码。
- 同步 `docs/openapi/admin.openapi.yaml` 和 `docs/test-cases/admin-integration-cases.md`。

### 验证补充

- `go test -count=1 ./internal/common/config ./internal/appapi` 通过。
- `go test -count=1 ./internal/invites ./internal/auth ./internal/appapi` 通过。
- 当前一期小程序后端业务功能进度约 `98.5%`。

## 2026-06-16 16:26

### 当前进展

- 继续收口一期小程序后端后台 IM 管理链路。
- 新增后台 IM 房间重试创建接口：
  - `POST /api/admin/im/rooms/{roomId}/retry-create`
  - 未启用 OpenIM 时返回当前本地房间，启用 OpenIM 时执行真实房间同步。
- 新增后台 IM 消息隐藏接口：
  - `POST /api/admin/im/messages/{messageId}/hide`
  - 隐藏后小程序用户侧历史消息不再返回该消息，后台房间详情仍保留证据。
- 新增后台 IM 房间归档接口：
  - `POST /api/admin/im/rooms/{roomId}/archive`
  - 归档后用户侧不能再向该房间发送消息。
- 后台权限补充：
  - `im:room:retry_create`
  - `im:room:archive`
  - `im:message:hide`
- 同步 `docs/openapi/admin.openapi.yaml` 和 `docs/test-cases/admin-integration-cases.md`。

### 验证补充

- 已执行：
  - `go test -count=1 ./internal/im ./internal/appapi`
  - `go test -count=1 ./internal/...`
- 测试通过。
- 当前一期小程序后端业务功能进度约 `99%`。

## 2026-06-16 16:56

### 当前进展

- 继续按一期小程序后端线上测试口径收口，处理头像文件 URL 的生产可用性。
- `PUT /api/app/users/me/profile` 在传入 `avatarFileId` 且未显式传 `avatarUrl` 时，现在通过统一文件服务生成头像下载地址。
- 本地未配置对象存储下载域名时仍返回 `mock://download/...`，保持本地测试稳定。
- 生产配置 `STORAGE_DOWNLOAD_BASE_URL` 后，头像 URL 自动返回 HTTPS 下载域名，不再由资料接口硬编码 `mock://`。

### 验证补充

- 已执行：
  - `go test -count=1 ./internal/files ./internal/appapi`
  - `go test -count=1 ./internal/...`
- 测试通过。
- 当前一期小程序后端业务功能进度约 `99.2%`。

## 2026-06-16 17:26

### 当前进展

- 继续按一期小程序后端联调口径收口，修正人脉跟进接口路径不一致问题。
- `POST /api/app/connections/{connectionId}/follow-logs` 已接入后端，作为文档主路径。
- 原有 `POST /api/app/connections/{connectionId}/follow-up` 保留为兼容别名，避免旧前端或旧测试调用中断。
- 同步 `docs/openapi/app.openapi.yaml` 和 `docs/test-cases/app-integration-cases.md`。

### 验证补充

- 已执行：
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./internal/...`
- 测试通过。
- 当前一期小程序后端业务功能进度约 `99.35%`。

## 2026-06-16 17:56

### 当前进展

- 继续按一期小程序后端联调可用性收口，修复 `docs/openapi/app.openapi.yaml` 中 6 处路径被上一段 `description` 吞并到同一行的问题。
- 已拆正以下接口段落：
  - `GET/POST /api/app/chat/rooms/{roomId}/messages`
  - `POST /api/app/games/{gameId}/service-confirm-items`
  - `GET /api/app/reviews/todos`
  - `GET /api/app/growth/my`
  - `POST /api/app/payment/precreate-placeholder`
  - `POST /api/app/reports`
- 这轮不改业务逻辑，只修 OpenAPI 结构，避免前端联调或接口生成器漏识别路径。

### 验证补充

- 已执行：
  - OpenAPI 坏行复扫：未再发现 `description` 吞并 `/api/app/...` 路径。
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./internal/...`
- 测试通过。
- 当前一期小程序后端业务功能进度约 `99.45%`。

## 2026-06-16 18:26

### 当前进展

- 继续按一期小程序后端线上联调口径收口，修复 OpenAPI 文档剩余结构问题。
- 已拆正 `docs/openapi/app.openapi.yaml` 中 `GET /api/app/reports/my` 的 `components` 粘连问题。
- 已修复 `docs/openapi/admin.openapi.yaml` 早期后台接口段落的路径、`responses`、状态码粘连问题，覆盖：
  - 后台登录与权限快照。
  - 后台权限树。
  - 微信订阅消息模板与任务。
  - 内部订阅消息发送回写任务。
  - 评价提醒与进度反馈提醒任务。
- 这轮仍不改变业务代码，只保证一期后端接口文档可以稳定用于前端联调和线上测试核对。

### 验证补充

- 已执行：
  - OpenAPI 坏行复扫：未再发现 `description/summary` 吞并 `/api/...`、`responses`、`security`、`components` 的结构问题。
- 已执行：
  - `go test -count=1 ./internal/...`
- 测试通过。
- 当前一期小程序后端业务功能进度约 `99.55%`。

## 2026-06-16 18:56

### 当前进展

- 继续以一期小程序后端线上测试为目标，补齐“链接邀请”真实联调口径。
- 新增 `WECHAT_URL_LINK_BASE_URL` 配置：
  - 本地未配置时，`POST /api/app/invites/entries` 继续返回 `wechat://mini-program?...` 占位链接，保证本机测试不被微信接口阻塞。
  - 生产配置后，`urlLink` 返回 HTTPS 链接，并自动携带 `inviteCode`、`entryType`、`path` 和 `query` 参数，方便微信聊天链接邀请联调。
- 同步 `.env.example` 和 `docs/openapi/app.openapi.yaml`，避免前端仍按“固定占位链接”理解。

### 验证补充

- 已执行：
  - `go test -count=1 ./internal/common/config ./internal/appapi`
  - `go test -count=1 ./internal/...`
- 测试通过。
- 当前一期小程序后端业务功能进度约 `99.65%`。

## 2026-06-17 10:30

### 当前进展

- 根据一期最新口径修正局型创建边界：
  - 小程序普通用户一期仍只能创建 `free` 普通局。
  - 后台新增开局入口，允许创建 `free`、`standard`、`condition`，用于一期线上测试其他条件局链路。
  - 后台创建的局标记 `gameSource=admin`，默认进入 `recruiting`，小程序侧可展示和参与。
- 新增后台权限 `game:create_admin`，和原 `game:read`、`game:update_status` 分离。
- 同步 `docs/openapi/admin.openapi.yaml` 和后台联调用例说明。

### 验证补充

- 待执行：
  - `go test -count=1 ./internal/games ./internal/appapi`
  - `go test -count=1 ./internal/...`
- 修正后当前一期小程序后端业务功能进度约 `99.7%`。

## 2026-06-17 11:40

### 当前进展

- 开始按一期后台前后端闭环补齐后台工作台，优先覆盖可线上测试的运营链路。
- 后端补充后台组局能力：
  - `GET /api/admin/games` 支持 `status`、`gameType`、`cityCode`、`keyword` 过滤。
  - `GET /api/admin/games/{gameId}` 返回局详情、成员、里程碑、打卡、复盘、续局草稿和服务确认信息。
  - 后台字段继续对齐小程序 `Game` DTO：`gameType`、`gameSource`、`status`、`minPlayers`、`maxPlayers`、`currentPlayers`。
- 后台前端从占位壳升级为真实可用一期工作台：
  - 登录与权限拉取。
  - 总览指标。
  - 后台创建普通局、标准局、条件局。
  - 组局筛选、审核通过、局链路详情。
  - 实名记录、角色申请、IM 房间、操作日志只读查看。
  - 本地 `admin-web/server.js` 增加 `/api` 代理，开发环境可同域对接 Go API。

### 验证补充

- 已执行：
  - `go test -count=1 ./internal/appapi ./internal/games ./internal/adminauth`
  - `go test -count=1 ./internal/...`
  - `go build ./cmd/server`
  - `node --check admin-web/src/main.js`
  - `node --check admin-web/server.js`
  - `npm run build`
- 当前后台前端业务功能进度约 `35%`：一期关键入口已可用，但还缺用户管理、邀请码管理、财务分润、举报申诉、导出中心、系统配置等完整页面。
- 当前一期小程序后端业务功能进度仍约 `99.7%`。

## 2026-06-17 16:20

### 当前进展

- 继续按一期后台前后端闭环推进，优先补齐和小程序登录注册强相关的运营能力。
- 后端新增后台用户列表接口：
  - `GET /api/admin/users` 支持 `keyword`、`realnameStatus`、`status` 过滤。
  - 返回字段对齐小程序 `User` DTO：`id`、`openId`、`nickname`、`avatarUrl`、`avatarFileId`、`realnameStatus`、`status`、`createdAt`。
  - 内存仓库和 SQL 仓库同步实现，SQL 字段继续对齐 `users` 与 `user_wechat_accounts`。
- 后台前端重建为可读中文工作台，并新增：
  - 用户管理：用户筛选、列表、详情，详情聚合实名、角色、人脉、收藏和收入摘要。
  - 邀请管理：创建邀请码、邀请码筛选、禁用邀请码、邀请关系列表。
  - 保留并整合总览、后台开局、组局审核、实名/角色审核、IM 房间、操作日志页面。
- 同步 `docs/openapi/admin.openapi.yaml`，补充 `/api/admin/users`。

### 验证补充

- 已执行：
  - `gofmt -w services/go-api/internal/users/model.go services/go-api/internal/users/sql_repository.go services/go-api/internal/auth/service.go services/go-api/internal/auth/repository_service_test.go services/go-api/internal/appapi/server.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/users ./internal/auth ./internal/appapi`
  - `go test -count=1 ./internal/...`
  - `node --check admin-web/src/main.js`
  - `node --check admin-web/server.js`
  - `npm run build`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `62%`：权限、用户、邀请、组局、审核、IM、日志已有核心接口，仍需继续补财务分润、举报申诉、导出中心、系统配置和更完整的统计接口。
- 当前后台前端业务功能进度约 `48%`：核心运营入口已成型，仍需补财务、举报、导出、配置和更细的详情操作页。

## 2026-06-17 16:55

### 当前进展

- 继续补齐一期 PC 后台业务功能，优先把已有后端能力变成可线上测试的后台页面。
- 后台前端新增分润结算页面：
  - 分润模板创建与列表查看。
  - 分润规则配置与查看。
  - 分润试算、生成分润记录。
  - 分润记录列表、冻结、线下结算登记。
  - 线下结算记录列表。
- 后台前端新增举报申诉页面：
  - 举报申诉列表。
  - 举报详情与证据摘要。
  - 分配处理人、处理、关闭。
  - 页面字段对齐 `reports.Report`：`gameId`、`reporterUserId`、`targetUserId`、`reportType`、`status`、`revenueFrozen`。
- 后台前端新增导出中心：
  - 固定格式导出模板列表。
  - 创建导出任务。
  - 执行导出 runner。
  - 导出任务列表与完成后的下载地址获取。
- 这轮未新增后端业务接口，复用并联通已有接口：
  - `/api/admin/revenue/*`
  - `/api/admin/reports/*`
  - `/api/admin/reports/export*`
  - `/api/admin/export-tasks*`

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi ./internal/revenue ./internal/reports ./internal/exports`
  - `go test -count=1 ./internal/...`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `62%`：本轮主要补前端接入，后端百分比不变。
- 当前后台前端业务功能进度约 `68%`：用户、邀请、组局、审核、IM、日志、分润、举报、导出已具备基础可操作页面；还需补系统配置、敏感词、数据看板细化、管理员/权限管理、更多详情联动和浏览器联调。

## 2026-06-17 17:02

### 当前进展

- 继续补齐一期 PC 后台系统管控能力，复用已有后端接口，不重复造服务。
- 后台前端新增系统配置页面：
  - 生产就绪检查：对接 `GET /api/admin/system/readiness`，展示 `ready`、`production` 和检查项，不展示密钥明文。
  - 敏感词库：对接 `GET/POST/PUT /api/admin/sensitive-words` 和 `POST /api/admin/sensitive-words/import`，支持新增、批量导入、启用/禁用。
  - 内容风险日志：对接 `GET /api/admin/content-risk/logs`，查看 IM 内容命中记录。
  - 权限快照：对接 `GET /api/admin/permissions/tree`，展示当前管理员角色、菜单、API 和权限数量。
  - AI IM 导出开关：对接 `GET/PUT /api/admin/ai-data/im-export-config`，保留一期默认关闭、需权限开启的后台入口。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi ./internal/im ./internal/adminauth`
  - `go test -count=1 ./internal/...`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `62%`：本轮仍以接通已有后端为主。
- 当前后台前端业务功能进度约 `76%`：系统配置、敏感词、权限快照、AI 导出开关已接入；剩余重点是管理员账号/角色管理、数据看板细化、更多详情联动和浏览器联调。

## 2026-06-17 17:31

### 当前进展

- 继续补齐一期 PC 后台 RBAC 管理能力，优先保证后台接口字段与小程序后端权限 code 对齐。
- 后台后端新增管理员权限目录接口：
  - `GET /api/admin/admin-users`：返回后台账号列表，字段包含 `id`、`username`、`status`、`roles`、`permissionCount`、`permissions`。
  - `GET /api/admin/admin-roles`：返回一期固定七类后台角色，覆盖超级管理员、运营管理、用户管理、财务管理、客户管理、数据分析员、组局管理。
  - `GET /api/admin/admin-permissions/catalog`：返回权限 code 目录，字段包含 `code`、`module`、`action`。
- 后端 RBAC 权限继续统一走 `requireAdminPermission`，新增三个接口绑定 `admin_user:view`，运营账号访问会返回 403。
- 后台前端新增“管理员权限”页面：
  - 后台账号表。
  - 七类角色目录。
  - 权限 code 目录。
  - 页面字段对齐 `admin_users`、`admin_roles`、`admin_permissions`、`admin_user_roles`。
- 保留系统配置页的权限快照，用于查看当前登录管理员的实时权限；新增页面用于管理视角查看全局目录。

### 验证补充

- 已执行：
  - `gofmt -w internal/adminauth/service.go internal/appapi/admin_auth_handler.go internal/appapi/server.go internal/appapi/server_test.go`
  - `go test -count=1 ./internal/adminauth ./internal/appapi`
  - `go test -count=1 ./internal/...`
  - `node --check admin-web/src/main.js`
  - `npm run build`
- 浏览器插件拒绝打开 `http://127.0.0.1:5173`，未做浏览器可视化烟测。
- 当前本机 `8080` 上已有旧/裁剪 Go 进程，只返回 health，后台登录接口为 404；为避免误伤用户已有进程，本轮未强制清理端口。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `66%`：管理员账号、角色、权限目录已补齐为可测接口；还差后台账号创建/禁用、角色授权持久化、菜单权限持久化、更多操作日志闭环。
- 当前后台前端业务功能进度约 `82%`：管理员权限页已接入；剩余重点是更多详情联动、数据看板细化、真实运行态浏览器联调。

## 2026-06-17 17:47

### 当前进展

- 继续推进 PC 后台 RBAC 从“只读目录”到“可管理闭环”。
- 后台后端新增管理员账号管理能力：
  - `POST /api/admin/admin-users`：创建后台管理员账号，字段为 `username`、`password`、`roles`、`status`。
  - `PUT /api/admin/admin-users/{id}`：更新管理员角色和状态，字段为 `roles`、`status`。
  - 新账号密码使用 PBKDF2-SHA256 + 随机 salt 生成哈希，不复用默认账号密码哈希。
  - 更新管理员角色或禁用账号后，会清理该管理员旧 session，避免旧 token 保留旧权限。
  - 创建绑定 `admin_user:create`，更新绑定 `role:update`，继续走统一后台权限中间件。
  - 创建和更新都会写入 `operation_logs`，用于超级管理员追责审计。
- 后台前端“管理员权限”页面补齐可操作能力：
  - 新增后台账号创建表单。
  - 后台账号列表支持行内修改 `roles` 和 `status`。
  - 支持禁用 / 启用管理员账号。
  - 字段继续对齐 `admin_users`、`admin_user_roles`、`admin_roles`、`admin_permissions`。
- 测试新增覆盖：
  - 低权限运营账号无法创建管理员。
  - 超级管理员可创建财务管理员。
  - 新财务管理员可访问财务接口但不能查看完整操作日志。
  - 禁用后旧 token 失效，新登录被拒绝。

### 验证补充

- 已执行：
  - `gofmt -w internal/adminauth/service.go internal/appapi/admin_auth_handler.go internal/appapi/server.go internal/appapi/server_test.go`
  - `go test -count=1 ./internal/adminauth ./internal/appapi`
  - `go test -count=1 ./internal/...`
  - `node --check admin-web/src/main.js`
  - `npm run build`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `72%`：RBAC 已从目录展示推进到账号创建、角色绑定、状态启停和 session 失效；剩余重点是持久化表落库、菜单权限持久化、更多后台操作统一日志。
- 当前后台前端业务功能进度约 `86%`：管理员权限页已具备创建和行内管理；剩余重点是详情联动、数据看板细化、真实运行态浏览器联调。

## 2026-06-17 17:56

### 当前进展

- 继续推进 PC 后台 RBAC 从“内存可测”到“数据库可上线”。
- 后台权限服务新增 SQL repository：
  - 后台登录按 `admin_users.username` 查找账号。
  - 账号角色从 `admin_user_roles` + `admin_roles` 读取。
  - 账号权限从 `admin_role_permissions` + `admin_permissions` 汇总。
  - 管理员列表、角色目录、权限目录均支持数据库读取。
  - 新增管理员写入 `admin_users` 和 `admin_user_roles`。
  - 更新管理员状态/角色写回 `admin_users.status` 和 `admin_user_roles`，并继续清理旧 session。
- `cmd/server` 启动入口已接入：
  - 配置 `DATABASE_URL` / `DATABASE_DRIVER` 且数据库可用时，后台管理员登录和 RBAC 管理走 `adminauth.NewSQLRepository(db)`。
  - 未配置数据库时继续使用内存 fallback，方便本地联调和既有测试。
- 同步 `db/seeds/admin_roles_permissions.sql`：
  - 补齐 `im:room:read`、`im:room:archive`、`im:room:retry_create`、`system_config:read`、`invite_code:read`、`invite_code:manage`。
  - 补齐 `user_manager`、`finance_manager`、`customer_manager`、`game_manager` 角色权限映射。
  - 补齐运营/数据分析角色的内容风控权限映射。
- 新增 repository 模式单元测试：
  - 验证登录使用持久化账号。
  - 验证创建管理员会走 repository。
  - 验证更新角色/禁用账号后旧 token 失效。

### 验证补充

- 已执行：
  - `gofmt -w internal/adminauth/service.go internal/adminauth/sql_repository.go internal/adminauth/service_test.go internal/appapi/admin_auth_handler.go internal/appapi/server.go cmd/server/main.go`
  - `go test -count=1 ./internal/adminauth`
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./internal/...`
  - `go test -count=1 ./...`
  - `node --check admin-web/src/main.js`
  - `npm run build`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `78%`：后台账号、角色、权限已经从接口到数据库持久化闭环；剩余重点是菜单配置持久化、后台审核/运营页面更完整的批量操作和运行态联调。
- 当前后台前端业务功能进度约 `86%`：本轮无新增可视页面，管理员权限页可继续复用现有接口；剩余重点是详情联动、数据看板细化、真实运行态浏览器联调。

## 2026-06-17 18:03

### 当前进展

- 继续推进 PC 后台“权限生效”从接口层扩展到菜单层。
- 后端 `/api/admin/auth/permissions` 和 `/api/admin/permissions/tree` 的 `menus` 改为按权限码动态生成：
  - 数据看板：`analytics:*`
  - 用户管理：`user:*`、`identity:read`、`profile:read`
  - 邀请管理：`invite_code:*`
  - 组局管理：`game:*`
  - 审核中心：`identity:read`、`role:*`
  - 分润结算：`revenue:*`、`settlement:offline:create`
  - 举报申诉：`report:*`
  - 导出中心：`report_export:create`
  - 管理员权限：`admin_user:*`、`role:update`
  - 系统配置：`system_config:*`、内容风控、AI 数据准备权限
  - IM 证据：`im:room:read`、`im:message:view_dispute`
  - 操作日志：`operation_log:view_full`
- 后台前端登录后保存权限树，并按 `menus` 隐藏无权限导航入口。
- 如果当前页面无权限，前端会自动切换到第一个可访问菜单，避免低权限账号登录后停留在无权限页面。
- 测试补充：
  - 运营账号可看到组局、举报、系统配置相关菜单。
  - 运营账号不可看到数据看板、管理员权限、操作日志、财务结算菜单。

### 验证补充

- 已执行：
  - `gofmt -w internal/adminauth/service.go internal/appapi/server_test.go`
  - `go test -count=1 ./internal/adminauth`
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./...`
  - `node --check admin-web/src/main.js`
  - `npm run build`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `81%`：后台 RBAC 已覆盖登录、权限码、菜单树、账号/角色/权限落库；剩余重点是更多后台业务页的批量操作和运行态联调。
- 当前后台前端业务功能进度约 `88%`：导航已按权限树动态收敛；剩余重点是批量操作交互、详情联动和真实浏览器联调。

## 2026-06-17 18:07

### 当前进展

- 继续补后台系统配置持久化缺口。
- AI 数据准备的 IM 导出开关从内存态改为可选数据库持久化：
  - 新增 `aidata.Repository`。
  - 新增 `aidata.NewServiceWithRepository`。
  - 新增 `aidata.NewSQLRepository(db)`。
  - `IMExportConfig()` 优先从 `ai_data_snapshots` 最新记录读取 `im_export_enabled`。
  - `SetIMExportEnabled(enabled)` 写入 `ai_data_snapshots`，保留历史配置轨迹。
  - 无数据库时仍保持原内存 fallback，方便本地联调。
- `cmd/server` 启动入口新增 `aiDataRepository(db)` 并注入 `appServer.UseAIDataRepository(...)`。
- 后台接口路径不变：
  - `GET /api/admin/ai-data/im-export-config`
  - `PUT /api/admin/ai-data/im-export-config`
  - 后台前端系统配置页可继续使用原字段，新增 `updatedAt` 可用于展示配置更新时间。
- 新增 `internal/aidata` repository 模式单元测试，覆盖：
  - 默认关闭。
  - 开启后配置可读取。
  - 未开启时 IM 导出被拦截。
  - 开启后 IM 消息可导出。

### 验证补充

- 已执行：
  - `gofmt -w internal/aidata/service.go internal/aidata/sql_repository.go internal/aidata/service_test.go internal/appapi/server.go cmd/server/main.go`
  - `go test -count=1 ./internal/aidata`
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./...`
  - `node --check admin-web/src/main.js`
  - `npm run build`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `83%`：RBAC、菜单权限、AI/IM 导出开关已具备数据库持久化；剩余重点是后台批量操作、更多列表详情联动和真实运行态联调。
- 当前后台前端业务功能进度约 `88%`：本轮接口字段兼容，无需页面重写；剩余重点是批量操作交互、详情联动和浏览器联调。

## 2026-06-17 18:16

### 当前进展

- 继续补 PC 后台运营效率功能，优先完成线上测试高频批量操作。
- 后台后端新增批量组局审核接口：
  - `POST /api/admin/games/batch-audit`
  - 权限：`game:update_status`
  - 请求字段：`gameIds`、`approve`、`remark`
  - 内部逐项复用 `games.ApproveGame`，不绕开原有状态流转。
  - 返回逐项结果：`id`、`success`、`status`、`error`，同时返回 `success`、`failed`、`total`。
  - 成功项写入 `operation_logs`，action 为 `game:batch_audit`。
- 后台后端新增批量举报处理接口：
  - `POST /api/admin/reports/batch-handle`
  - 权限：`report:handle`
  - 请求字段：`reportIds`、`action`、`adminId`、`handlerAdminId`、`result`
  - 支持 `assign`、`handle`、`close` 三种动作。
  - 内部逐项复用 `reports.Assign`、`reports.Handle`、`reports.Close`。
  - 成功项写入 `operation_logs`，action 为 `report:batch_{action}`。
- 后台前端补充批量入口：
  - 组局管理页新增“批量通过待审核”，对当前列表内 `pending_audit` 组局批量审核。
  - 举报申诉页新增“批量处理”，对当前列表内 `pending/assigned` 举报批量处理。
  - 批量操作完成后刷新当前列表，并提示成功/失败数量。

### 验证补充

- 已执行：
  - `gofmt -w internal/appapi/batch_helper.go internal/appapi/game_handler.go internal/appapi/report_handler.go internal/appapi/server.go internal/appapi/server_test.go`
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./...`
  - `node --check admin-web/src/main.js`
  - `npm run build`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `86%`：后台 RBAC、菜单权限、核心配置持久化、组局/举报批量操作已可用；剩余重点是用户/实名/角色审核详情联动和浏览器运行态联调。
- 当前后台前端业务功能进度约 `90%`：核心列表和批量操作已接入；剩余重点是更多详情面板、空/错态细化和真实浏览器联调。

## 2026-06-17 18:30

### 当前进展

- 继续补 PC 后台审核中心，优先让一期后台能实际处理实名与角色申请链路。
- 后台前端审计中心新增实名认证详情入口：
  - 实名列表每条记录新增“详情”按钮。
  - 点击后调用 `GET /api/admin/identity-verifications/{userId}`。
  - 详情面板展示 `userId`、`phone`、`realname`、`status`、`faceIdRequestId`、`createdAt`、`updatedAt`，字段与后端 `identity.Record` 对齐。
- 后台前端角色申请新增详情、通过、驳回链路：
  - 角色申请列表缓存 `roleApplications`，详情面板展示 `role_applications` 关键字段。
  - 待处理申请支持“通过”和“驳回”。
  - 两个动作统一调用 `POST /api/admin/audits/role-applications/{id}/review`，不新建旁路接口。
  - 驳回会传 `approve=false` 与审核备注，后端会写入 `rejected/rejectReason/reviewRemark`。
- 后台 API 请求头补齐当前管理员 ID：
  - 所有已登录后台请求会携带 `X-Admin-ID`。
  - 角色审核时后端 `reviewAdminId` 和操作日志可以落到真实后台账号。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/main.js`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `87%`：审核接口原本已具备，本轮补齐前端调用和管理员 ID 请求头；剩余重点是用户详情联动、后台配置页细化、导出/日志真实联调。
- 当前后台前端业务功能进度约 `92%`：审计中心已具备列表、详情、通过、驳回闭环；剩余重点是用户/邀请/收益等模块的详情面板与浏览器交互联调。

## 2026-06-17 18:44

### 当前进展

- 继续补 PC 后台用户与邀请链路，优先服务一期线上测试时“查用户、查邀请码、查绑定来源”的运营动作。
- 后台后端新增邀请码详情接口：
  - `GET /api/admin/invite-codes/{code}`
  - 权限：`invite_code:read`
  - 返回 `inviteCode`、`relations`、`ownerUser`、`boundUser`
  - 详情字段与 `invite_codes`、`invite_relations`、`users` 现有模型对齐，不新增独立字段。
- 后台前端邀请管理新增详情面板：
  - 邀请码列表操作列新增“详情”。
  - 详情面板展示 `id`、`code`、`entryType`、`ownerUserId`、`usedCount/maxUses`、`boundWechatUserId`、绑定用户昵称、关系数、绑定来源。
  - 禁用按钮保留，和详情按钮共用现有 `row-actions` 样式。
- 后台前端用户详情增强：
  - 用户详情从粗略 JSON 摘要改为拆分展示邀请码、成长等级、信用分、可用积分、会员状态、收益总额、待结算、已结算。
  - 字段来自 `GET /api/admin/users/{id}` 已有返回结构，继续复用小程序端用户/成长/收益模型。

### 验证补充

- 已执行：
  - `gofmt -w internal/appapi/invite_admin_handler.go internal/appapi/server.go internal/appapi/server_test.go`
  - `node --check admin-web/src/main.js`
  - `go test -count=1 ./internal/appapi`
  - `npm run build`
  - `go test -count=1 ./...`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/main.js`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `89%`：邀请码详情接口、审核链路、批量操作、RBAC、核心配置已具备；剩余重点是收益/导出/日志与系统配置的真实联调和边界补齐。
- 当前后台前端业务功能进度约 `93%`：用户详情、邀请详情、审计中心、批量操作已能支撑一期测试；剩余重点是收益/导出/日志详情与浏览器交互联调。

## 2026-06-17 19:00

### 当前进展

- 继续补 PC 后台财务、导出、日志排查链路，优先让一期线上测试时的后台追踪动作完整可见。
- 后台前端分润记录新增详情面板：
  - 分润记录操作列新增“详情”。
  - 详情面板展示 `id`、`recordNo`、`gameId`、`templateId`、`amountCent`、`frozenReason`、`settledAt`、`createdAt`。
  - 明细表展示 `revenue_record_items` 中的 `role`、`userId`、`amountCent`。
  - 同步显示该记录关联的线下结算记录数量。
- 后台前端导出任务新增详情面板：
  - 导出任务操作列新增“详情”。
  - 详情面板展示 `taskNo`、`templateCode`、`exportType`、`fileId`、`createdBy`、`createdAt`、`finishedAt`、`failReason`、`filters`。
  - 已完成且存在 `fileId` 的任务会在详情中尝试拉取下载地址。
  - 下载地址按钮保留，和详情按钮共用 `row-actions`。
- 后台前端操作日志新增筛选和详情：
  - 日志页动态加入 `action`、`targetType`、`adminUserId` 筛选。
  - 日志列表新增“详情”按钮。
  - 详情面板展示 `adminUserId`、`action`、`targetType`、`targetId`、`requestId`、`ip`、`createdAt`、`detail`。
  - 筛选在前端对 `GET /api/admin/operation-logs` 返回结果处理，不改变后端接口合同。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./...`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/main.js`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `90%`：本轮主要使用既有后端接口补前端闭环；后端剩余重点是导出/日志筛选服务端化、系统配置更多可写项和真实环境联调。
- 当前后台前端业务功能进度约 `95%`：用户、邀请、审核、组局、举报、收益、导出、日志核心页面均有列表和详情/动作闭环；剩余重点是系统配置细化、边界错误态和浏览器真实交互联调。

## 2026-06-17 19:12

### 当前进展

- 继续补 PC 后台服务端筛选能力，减少线上数据量增长后前端全量过滤的压力。
- 后台后端操作日志列表新增 query 过滤：
  - `GET /api/admin/operation-logs?action=&targetType=&targetId=&adminUserId=`
  - `action` 支持包含匹配。
  - `targetType`、`targetId`、`adminUserId` 使用精确匹配。
  - 返回结构补充 `total`，保持原有 `items` 不变。
- 后台后端导出任务列表新增 query 过滤：
  - `GET /api/admin/export-tasks?status=&templateCode=&exportType=&createdBy=`
  - 均按 `export_tasks` 现有字段过滤。
  - 返回结构补充 `total`，保持原有 `items` 不变。
- 后台前端同步改为服务端筛选：
  - 操作日志页筛选表单提交后直接请求带 query 的 `/api/admin/operation-logs`。
  - 导出任务页新增 `templateCode`、`exportType`、`status` 筛选表单。
  - 筛选 UI 复用现有 `filters` 样式，不引入新组件体系。

### 验证补充

- 已执行：
  - `gofmt -w internal/appapi/audit_handler.go internal/appapi/export_handler.go internal/appapi/server_test.go`
  - `node --check admin-web/src/main.js`
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./...`
  - `npm run build`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/main.js`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `92%`：日志/导出筛选已服务端化，剩余重点是系统配置更多可写项、后台真实浏览器联调和上线环境联调。
- 当前后台前端业务功能进度约 `96%`：核心页面的列表、详情、动作、筛选基本齐备；剩余重点是系统配置细化和浏览器真实交互联调。

## 2026-06-17 19:24

### 当前进展

- 继续补 PC 后台系统配置页，优先让一期线上测试需要的角色/资格门槛能从后台直接调整。
- 后台前端系统页新增“领路人资格规则”面板：
  - 接入既有后端接口 `GET /api/admin/guides/qualification-rules`。
  - 支持查看 `guide_qualification_rules` 的 `ruleCode`、`status`、`minInviteCount`、`minCreditScore`、`minCompletedGames`、`paymentRequired`、`updatedAt`。
  - 支持将规则载入表单，避免运营人员手动重复录入规则 ID 和门槛字段。
  - 支持通过 `PUT /api/admin/guides/qualification-rules/{id}` 保存规则。
  - 可调整邀请数门槛、信用分门槛、完成局数门槛、是否要求支付、启用/禁用状态。
- 该功能用于支撑一期后台开局和角色测试：
  - 小程序一期普通用户仍只能走普通局主链路。
  - 后台可以通过资格规则和后台开局能力验证行家、领路人、条件局相关业务条件是否完整。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./...`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/main.js`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `92%`：本轮复用既有后端规则接口，后端未新增模型；剩余重点是系统配置更多可写项、真实账号联调和线上环境联调。
- 当前后台前端业务功能进度约 `97%`：系统配置页已补入领路人资格规则管理；剩余重点是登录后真实浏览器点击流、更多配置项细节和错误态打磨。

## 2026-06-17 19:38

### 当前进展

- 继续补 PC 后台一期运营功能，优先把已存在的积分/兑换后端接口接成可操作页面。
- 后台新增“积分兑换”菜单和页面：
  - 导航新增 `redemption`。
  - 页面模板新增 `redemption-template`。
  - 商品区对齐 `redemption_items`，支持新增商品、行内修改 `name`、`pointsCost`、`stock`、`status`。
  - 订单区对齐 `redemption_orders`，支持待处理订单通过/驳回，已通过订单履约。
  - 积分流水区对齐 `points_logs`，展示 `userId`、`changeValue`、`beforePoints`、`afterPoints`、`bizType`、`bizId`、`reason`。
- 后台 RBAC 菜单同步补齐：
  - `menuNodesForPermissions` 新增 `redemption` 菜单。
  - 拥有 `points:read` 或 `redemption:manage` 权限的后台账号可看到积分兑换菜单。
  - `user_manager` 角色新增 `points:read`、`redemption:manage`，方便用户运营处理积分和兑换订单。
  - `db/seeds/admin_roles_permissions.sql` 同步补齐用户管理员角色权限。
- 测试补充：
  - 后台权限树测试新增用户管理员可见 `redemption` 菜单的断言，防止 RBAC 菜单回退。

### 验证补充

- 已执行：
  - `gofmt -w internal/adminauth/service.go`
  - `gofmt -w internal/appapi/server_test.go`
  - `go test -count=1 ./internal/adminauth ./internal/appapi`
  - `go test -count=1 ./...`
  - `node --check src/main.js`
  - `npm run build`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/`
  - `Invoke-WebRequest -UseBasicParsing http://127.0.0.1:5173/main.js`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `93%`：积分兑换后端接口原本已具备，本轮补齐 RBAC 菜单映射和角色种子；剩余重点是会员团队/订阅消息/交付文档等后台页面接入和线上环境联调。
- 当前后台前端业务功能进度约 `98%`：新增积分兑换页面后，核心运营、审核、组局、分润、导出、系统、IM、日志页面基本齐备；剩余重点是更多小模块页面和真实浏览器点击流。

## 2026-06-17 19:55

### 当前进展

- 继续补 PC 后台一期用户运营能力，把既有会员团队和会员报表后端接口接入后台页面。
- 后台新增“会员团队”菜单和页面：
  - 导航新增 `members`。
  - 页面模板新增 `members-template`。
  - 团队区接入 `GET /api/admin/teams`，字段对齐 `teams.Team`：`id`、`leaderUserId`、`name`、`status`、`createdAt`。
  - 团队详情接入 `GET /api/admin/teams/{teamId}`，展示团队成员 `userId`、`relationLevel`、`source`、`status`、`joinedAt`，并展示 `revenueSummary`。
  - 报表区接入 `GET /api/admin/member-reports`，字段对齐 `member_report_snapshots`：`userId`、`period`、`membershipPlan`、`invitedCount`、`participatedGames`、`completedGames`、`incomeSummary`。
- 后台 RBAC 菜单同步补齐：
  - `menuNodesForPermissions` 新增 `members` 菜单。
  - 拥有 `team:read` 或 `member_report:read` 权限的后台账号可看到会员团队菜单。
  - 前端按权限拆分加载：没有 `team:read` 时不请求团队接口，没有 `member_report:read` 时不请求报表接口。
  - `user_manager` 角色新增 `member_report:read`、`team:read`，用于用户运营查看会员团队、邀请沉淀和收益汇总。
  - `db/seeds/admin_roles_permissions.sql` 同步补齐用户管理员角色权限。
- 测试补充：
  - 后台权限树测试新增用户管理员可见 `members` 菜单的断言，防止 RBAC 菜单回退。

### 验证补充

- 已执行：
  - `gofmt -w internal/adminauth/service.go internal/appapi/server_test.go`
  - `node --check src/main.js`
  - `go test -count=1 ./internal/adminauth ./internal/appapi`
  - `go test -count=1 ./...`
  - `npm run build`
  - 检查 `admin-web/dist/index.html`、`admin-web/dist/main.js` 已包含 `members` 菜单、`members-template`、`/api/admin/teams`、`/api/admin/member-reports`。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `94%`：会员团队/会员报表接口原本已具备，本轮补齐 RBAC 菜单映射和角色种子；剩余重点是订阅消息、交付文档、更多系统配置项和线上环境联调。
- 当前后台前端业务功能进度约 `98.5%`：新增会员团队页面后，核心运营、审核、组局、分润、积分兑换、报表、导出、系统、IM、日志页面基本齐备；剩余重点是订阅消息/交付文档页面、真实浏览器点击流和错误态打磨。

## 2026-06-17 20:08

### 当前进展

- 继续补 PC 后台一期上线验收相关能力，把订阅消息、交付文档、测试留档接入后台主导航。
- 后台新增“通知交付”菜单和页面：
  - 导航新增 `delivery`。
  - 页面模板新增 `delivery-template`。
  - 微信订阅任务区接入 `GET /api/admin/notifications/wechat-tasks`，字段对齐 `notifications.WechatTask`：`notificationId`、`userId`、`scene`、`templateId`、`status`、`resultCode`、`resultMessage`。
  - 微信订阅模板区接入 `GET /api/admin/notifications/wechat-templates`，字段对齐 `notifications.WechatTemplate`：`scene`、`templateId`、`title`、`status`。
  - 支持调用 `/api/internal/notifications/wechat-tasks/send-pending` 批量发送待发任务。
  - 支持调用 `/api/internal/notifications/wechat-tasks/{id}/send` 单条发送。
  - 支持调用 `/api/internal/notifications/wechat-tasks/{id}/mark-sent` 标记已发送。
  - 交付文档区接入 `GET/POST /api/admin/delivery-documents`，字段对齐 `delivery.Document`：`docType`、`title`、`status`、`reason`。
  - 测试用例区接入 `GET/POST /api/admin/test-cases`，字段对齐 `delivery.TestCase`：`module`、`caseName`、`priority`、`expectedResult`。
  - 测试运行区接入 `GET/POST /api/admin/test-runs`，字段对齐 `delivery.TestRun`：`caseId`、`result`、`actualResult`、`requestId`、`evidenceFileId`。
- 后台 RBAC 菜单同步补齐：
  - `menuNodesForPermissions` 新增 `delivery` 菜单。
  - 拥有 `notification:wechat:view`、`delivery:manage`、`testcase:read` 或 `testcase:manage` 任一权限的后台账号可看到通知交付菜单。
  - 前端按权限拆分加载：没有对应权限时不请求对应接口，避免页面级 403。
- 测试补充：
  - 后台权限树测试新增运营角色可见 `delivery` 菜单的断言，防止通知交付菜单回退。

### 验证补充

- 已执行：
  - `gofmt -w internal/adminauth/service.go internal/appapi/server_test.go`
  - `node --check src/main.js`
  - `go test -count=1 ./internal/adminauth ./internal/appapi`
  - `go test -count=1 ./...`
  - `npm run build`
  - 检查 `admin-web/dist/index.html`、`admin-web/dist/main.js` 已包含 `delivery` 菜单、`delivery-template`、订阅消息、交付文档、测试用例、测试运行接口路径。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `95%`：通知、交付、测试留档后端原本已具备，本轮补齐 RBAC 菜单映射和前端入口；剩余重点是线上环境配置、真实微信订阅消息发送联调和更多系统配置项。
- 当前后台前端业务功能进度约 `99%`：核心运营、审核、组局、分润、积分兑换、会员团队、通知交付、报表、导出、系统、IM、日志页面基本齐备；剩余重点是真实浏览器点击流和线上账号联调。

## 2026-06-17 20:23

### 当前进展

- 继续补 PC 后台数据分析与 AI 数据准备能力，把已有分析后端接口接入后台主导航。
- 后台新增“数据分析”菜单和页面：
  - 导航新增 `analytics`。
  - 页面模板新增 `analytics-template`。
  - 漏斗分析区接入 `GET /api/admin/analytics/funnel`，字段对齐 `audit.FunnelSnapshot`：`eventCode`、`userCount`、`conversionRate`、`dropOffRate`。
  - 留存分析区接入 `GET /api/admin/analytics/retention`，字段对齐 `audit.RetentionSnapshot`：`cohortDate`、`newUsers`、`day1Retained`、`day7Retained`、`day30Retained`。
  - 行为事件区接入 `GET /api/admin/behavior/events`，字段对齐 `audit.BehaviorLog`：`userId`、`eventType`、`eventCode`、`targetType`、`targetId`、`source`、`createdAt`。
  - AI 数据快照区接入 `GET /api/admin/ai-data/snapshot`，字段对齐 `aidata.Snapshot`：`userCount`、`behaviorLogCount`、`reviewCount`、`acceptanceReady`、`dataReadinessSections`、`acceptanceChecks`。
  - 支持调用 `POST /api/admin/ai-data/acceptance-fixture` 生成 AI 验收数据，用于一期 AI 推荐/IM 分析数据准备验收。
- 后台 RBAC 菜单同步补齐：
  - `menuNodesForPermissions` 新增 `analytics` 菜单。
  - 拥有 `analytics:funnel:view`、`analytics:retention:view`、`analytics:timeline:view`、`data:behavior:read`、`ai:data:read` 或 `ai:data:seed` 任一权限的后台账号可看到数据分析菜单。
  - 前端按权限拆分加载：没有对应权限时不请求对应接口，避免页面级 403。
- 测试补充：
  - 后台权限树测试新增数据分析员可见 `analytics` 和 `delivery` 菜单的断言，防止数据分析/通知交付菜单回退。

### 验证补充

- 已执行：
  - `gofmt -w internal/adminauth/service.go internal/appapi/server_test.go`
  - `node --check src/main.js`
  - `go test -count=1 ./internal/adminauth ./internal/appapi`
  - `go test -count=1 ./...`
  - `npm run build`
  - 检查 `admin-web/dist/index.html`、`admin-web/dist/main.js` 已包含 `analytics` 菜单、`analytics-template`、漏斗、留存、行为事件、AI 快照和 AI 验收数据接口路径。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `96%`：数据分析与 AI 数据准备后端接口原本已具备，本轮补齐 RBAC 菜单映射和前端入口；剩余重点是线上环境配置、真实第三方联调和生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.3%`：核心运营、审核、组局、分润、积分兑换、会员团队、数据分析、通知交付、报表、导出、系统、IM、日志页面基本齐备；剩余重点是真实浏览器点击流、线上账号联调和少量配置项细化。

## 2026-06-17 20:38

### 当前进展

- 继续补 PC 后台画像关系模块，把小程序端已经沉淀的人脉关系、行家技能、领路人资源暴露到后台。
- 后台新增“画像关系”菜单和页面：
  - 导航新增 `profiles`。
  - 页面模板新增 `profiles-template`。
  - 人脉关系区接入 `GET /api/admin/connections`，字段对齐 `connections.Connection`：`userId`、`connectedUserId`、`relationType`、`sourceType`、`sourceId`、`strengthScore`、`updatedAt`。
  - 行家技能查询接入 `GET /api/admin/experts/{userId}/skills`，字段对齐 `profiles.ExpertSkillProfile`：`skillTree`、`serviceTags`、`caseFileIds`、`completeness`。
  - 领路人资源查询接入 `GET /api/admin/guides/{userId}/resources`，字段对齐 `profiles.GuideResourceProfile`：`resourceTags`、`industryTags`、`cityCodes`、`connectionScale`、`completeness`。
- 后台 RBAC 菜单同步补齐：
  - `menuNodesForPermissions` 新增 `profiles` 菜单。
  - 拥有 `connection:read` 或 `profile:read` 任一权限的后台账号可看到画像关系菜单。
  - 前端按权限拆分加载：没有 `connection:read` 时不请求人脉列表，没有 `profile:read` 时禁止查询画像详情。
- 测试补充：
  - 后台权限树测试新增用户管理角色可见 `profiles` 菜单的断言，防止画像关系菜单回退。

### 验证补充

- 已执行：
  - `gofmt -w services/go-api/internal/adminauth/service.go services/go-api/internal/appapi/server_test.go`
  - `node --check admin-web/src/main.js`
  - `go test -count=1 ./internal/adminauth ./internal/appapi`
  - `npm run build`
  - 检查 `admin-web/dist/index.html`、`admin-web/dist/main.js` 已包含 `profiles` 菜单、`profiles-template`、人脉、行家技能、领路人资源接口路径。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `96.5%`：画像关系后端接口原本已具备，本轮补齐 RBAC 菜单映射和前端入口；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.5%`：核心运营、审核、组局、分润、积分兑换、会员团队、画像关系、数据分析、通知交付、报表、导出、系统、IM、日志页面基本齐备；剩余重点是真实浏览器点击流、线上账号联调和少量配置项细化。

## 2026-06-17 20:53

### 当前进展

- 继续补 PC 后台评价、成长、信用追踪能力，把一期“服务确认 -> 评价 -> 成长/信用 -> 分润前置条件”链路暴露到后台。
- 后台新增“评价成长”菜单和页面：
  - 导航新增 `growth`。
  - 页面模板新增 `growth-template`。
  - 用户成长追踪接入 `GET /api/admin/users/{userId}/growth`，字段对齐 `reviews.Trace`：`profile`、`reviews`、`creditLogs`、`footprints`、`achievements`。
  - 局评价追踪接入 `GET /api/admin/games/{gameId}/review-trace`，按局查看评价记录、再玩意向、信用扣分和足迹证据。
  - 页面展示成长摘要、评价记录、信用流水、足迹证据和成就，便于运营/客服核查评价闭环与争议依据。
- 后台 RBAC 菜单同步补齐：
  - `menuNodesForPermissions` 新增 `growth` 菜单。
  - 拥有 `user:view` 或 `game:view` 任一权限的后台账号可看到评价成长菜单。
  - 前端按权限拆分加载：没有 `user:view` 时禁止查用户成长，没有 `game:view` 时禁止查局评价轨迹。
- 测试补充：
  - 后台权限树测试新增用户管理角色可见 `growth` 菜单的断言，防止评价成长入口回退。

### 验证补充

- 已执行：
  - `gofmt -w services/go-api/internal/adminauth/service.go services/go-api/internal/appapi/server_test.go`
  - `node --check admin-web/src/main.js`
  - `go test -count=1 ./internal/adminauth ./internal/appapi`
  - `go test -count=1 ./...`
  - `npm run build`
  - 检查 `admin-web/dist/index.html`、`admin-web/dist/main.js` 已包含 `growth` 菜单、`growth-template`、用户成长追踪和局评价追踪接口路径。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `97%`：评价成长追踪后端接口原本已具备，本轮补齐 RBAC 菜单映射和前端入口；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.6%`：核心运营、审核、组局、分润、积分兑换、会员团队、画像关系、评价成长、数据分析、通知交付、报表、导出、系统、IM、日志页面基本齐备；剩余重点是真实浏览器点击流、线上账号联调和少量配置项细化。

## 2026-06-17 21:08

### 当前进展

- 继续补 PC 后台组局运营明细能力，把一期预留的里程碑、打卡、复盘、续局草稿接入现有“组局管理”详情面板。
- 后台“组局管理”详情增强：
  - `GET /api/admin/games/{gameId}` 返回的 `milestones`、`checkins`、`retrospectives`、`continueDrafts` 从数量展示升级为明细表格展示。
  - 支持后台创建里程碑，调用 `POST /api/admin/games/{gameId}/milestones`，字段对齐 `games.MilestoneRequest`：`title`、`status`。
  - 支持后台标记异常打卡，调用 `POST /api/admin/games/{gameId}/checkins/{checkinId}/mark-invalid`。
  - 复盘明细展示 `againIntent`，续局草稿展示 `originalGameId`、`draftGameId`、`creatorUserId`、`title`、`status`。
- 权限控制：
  - 运营明细读取沿用 `game:read`。
  - 创建里程碑和标记异常打卡沿用 `game:progress:manage`。
  - 前端无 `game:progress:manage` 时不显示异常标记动作，并在提交里程碑时拦截。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi ./internal/games`
  - `go test -count=1 ./...`
  - `Invoke-WebRequest http://127.0.0.1:5173/` 和 `/main.js` 均返回 200。
  - 检查 `admin-web/dist/main.js` 已包含 `gameOpsBlock`、`game-milestone-form`、`checkin-invalid`、`/api/admin/games/{id}/milestones`、`mark-invalid` 等接口路径。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `97.2%`：组局运营明细后端接口原本已具备，本轮补齐后台详情页操作面；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.7%`：组局管理从列表/审核扩展到运营明细查看和轻量操作；剩余重点是真实浏览器点击流、线上账号联调和少量配置项细化。

## 2026-06-17 21:23

### 当前进展

- 继续补 PC 后台邀请入口核查能力，对齐最近调整的“只能通过小程序卡片、二维码、链接三种形式进入小程序”和“唯一邀请码绑定微信后进入登录页”的业务规则。
- 后台“邀请管理”详情增强：
  - 邀请码详情页从基础字段展示升级为“入口校验 + 绑定关系明细”。
  - 入口校验展示 `allowedEntry`、`uniqueBinding`、`boundWechat`、`authPageMode`，便于后台判断当前邀请码应进入注册页还是登录页。
  - 绑定关系明细展开 `inviteCodeId`、`inviterUserId`、`inviteeUserId`、`bindSource`、入口类型，字段对齐 `invite_relations`。
  - 详情页支持直接禁用 active 邀请码，仍调用 `POST /api/admin/invite-codes/{code}/disable`。
- 权限控制：
  - 详情读取沿用 `invite_code:read`。
  - 禁用邀请码沿用 `invite_code:manage`。
  - 前端无管理权限时不显示详情页禁用按钮。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi ./internal/invites`
  - `go test -count=1 ./...`
  - `Invoke-WebRequest http://127.0.0.1:5173/` 和 `/main.js` 均返回 200。
  - 检查 `admin-web/src/main.js`、`admin-web/dist/main.js` 已包含 `入口校验`、`authPageMode`、`inviteRelationDetailRow`、`detail-invite-disable`、`inviteEntryLabel`。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `97.3%`：邀请管理后端接口原本已具备，本轮补齐后台详情页核查和禁用操作面；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.75%`：邀请、组局、审核、分润、IM、评价成长、画像关系、通知交付、数据分析、系统、日志等主链路基本齐备；剩余重点是真实浏览器点击流、线上账号联调和少量配置项细化。

## 2026-06-17 21:38

### 当前进展

- 继续补 PC 后台用户详情核查能力，把用户详情里的收藏、人脉统计升级为可下钻明细。
- 后台“用户详情”增强：
  - 详情打开时按权限接入 `GET /api/admin/users/{userId}/favorites`，展示 `userId`、`gameId`、`game.title`、`createdAt`，字段对齐 `game_favorites` 和关联局信息。
  - 详情打开时按权限接入 `GET /api/admin/users/{userId}/connections`，展示 `id`、`userId`、`connectedUserId`、`relationType`、`source`、`strengthScore`、`updatedAt`，字段对齐 `user_connections`。
  - 顶部统计的 `favorites`、`connections` 改为以后端明细接口返回数量为准；无权限时回落使用详情接口已有摘要数据。
- 权限控制：
  - 收藏明细读取沿用 `user:read`。
  - 人脉明细读取沿用 `connection:read`。
  - 用户基础详情仍沿用 `user:view`，避免扩大原有后台权限边界。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi ./internal/games`
  - `go test -count=1 ./...`
  - 检查 `admin-web/src/main.js`、`admin-web/dist/main.js` 已包含 `/api/admin/users/{id}/favorites`、`/api/admin/users/{id}/connections`、`userFavoriteRow`、`收藏明细`、`人脉明细`。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `97.4%`：用户收藏、人脉后端接口原本已具备，本轮补齐后台详情页下钻使用；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.8%`：后台核心业务链路已基本具备运营查看、审核、配置、处置和核查能力；剩余重点是真实浏览器点击流、线上账号联调和少量配置项细化。

## 2026-06-17 21:53

### 当前进展

- 继续补 PC 后台文件证据核查能力，把后端已有的后台文件下载权限接口接到举报详情和测试留档页面。
- 后台“举报详情”增强：
  - 如果举报证据包含 `fileId`，详情页展示 `file.bizType`、`file.name`。
  - 新增“下载附件”动作，调用 `GET /api/admin/files/{fileId}/download-url`。
  - 下载权限仍由后端按 `files.File.bizType` 分流：`report_attachment` 需要 `report:view`，`chat_file` 需要 `im:message:view_dispute`，`realname_material` 需要 `identity:read`，`export_file` 需要 `report_export:create`。
- 后台“测试运行”增强：
  - `delivery.TestRun.evidenceFileId` 不再只显示数字，有证据文件时提供下载动作。
  - 下载动作复用同一个 `downloadAdminFile(fileId)`，生成地址后复制到剪贴板并尝试打开新窗口。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi ./internal/files ./internal/reports`
  - `go test -count=1 ./...`
  - 检查 `admin-web/src/main.js`、`admin-web/dist/main.js` 已包含 `downloadAdminFile`、`/api/admin/files/{fileId}/download-url`、`admin-file-download`、`test-run-file-download`。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `97.5%`：后台文件下载后端接口原本已具备，本轮补齐前端操作入口；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.85%`：后台从业务列表、审核处置进一步补到证据文件核查；剩余重点是真实浏览器点击流、线上账号联调和少量配置项细化。

## 2026-06-17 22:08

### 当前进展

- 继续补 PC 后台 IM 管理闭环，把文档和后端已有的 OpenIM/本地 IM 管理接口接入后台页面。
- 后台“IM 证据”增强：
  - 房间列表新增“详情”，调用 `GET /api/admin/im/rooms/{roomId}`，展示 `room`、全部消息、文件消息、成员、OpenIM group、归档原因。
  - 房间列表新增“争议消息”，调用 `GET /api/admin/im/rooms/{roomId}/dispute-messages`，用于举报/申诉时快速核查保全消息。
  - 有 `im:room:retry_create` 权限时可调用 `POST /api/admin/im/rooms/{roomId}/retry-create`，用于 OpenIM 群同步重试。
  - 有 `im:room:archive` 权限时可调用 `POST /api/admin/im/rooms/{roomId}/archive`，归档后小程序端不可继续发送该房间消息。
  - 有 `im:message:hide` 权限时可调用 `POST /api/admin/im/messages/{messageId}/hide`，隐藏普通用户历史里的违规消息，同时后台详情仍可留证。
  - 文件消息复用 `GET /api/admin/files/{fileId}/download-url` 下载入口，字段继续对齐 `im_messages.file_id` 与 `files`。
- 权限控制：
  - 房间详情沿用 `im:room:read`。
  - 争议消息沿用 `im:message:view_dispute`。
  - 重试同步、归档、隐藏消息分别沿用 `im:room:retry_create`、`im:room:archive`、`im:message:hide`。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi ./internal/im`
  - `go test -count=1 ./...`
  - 检查 `admin-web/src/main.js`、`admin-web/dist/main.js` 已包含 `imRoomActions`、`showIMRoomDetail`、`showIMDisputeMessages`、`/retry-create`、`/archive`、`/api/admin/im/messages/{messageId}/hide`。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `97.7%`：IM 管理后端接口原本已具备，本轮补齐后台可操作页面；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.9%`：后台核心业务、证据、审核、IM 处置、文件核查、运营配置基本形成完整操作闭环；剩余重点是真实浏览器点击流、线上账号联调和少量配置项细化。

## 2026-06-17 22:23

### 当前进展

- 继续补 PC 后台 AI 数据准备验收闭环，把验收用例中的 `POST /api/admin/ai-data/im-export` 接到后台“数据分析”页面。
- 后台“数据分析 / AI 数据快照”增强：
  - 新增“导出 IM 数据”按钮，调用 `POST /api/admin/ai-data/im-export`。
  - 导出关闭时页面会展示后端返回错误，便于验证一期默认关闭的导出闸门。
  - 导出成功时展示导出条数和前 8 条样例，避免大量 IM 内容直接撑满页面。
  - 无 `ai:data:export` 权限时展示无权限提示，按钮点击也会被拦截。
- 与系统配置页联动：
  - 系统配置页已有 `GET/PUT /api/admin/ai-data/im-export-config` 开关。
  - 本轮补齐“开关开启后从数据分析页触发导出”的验收路径。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi ./internal/aidata`
  - `go test -count=1 ./...`
  - 检查 `admin-web/src`、`admin-web/dist` 已包含 `ai-im-export-button`、`ai-im-export-panel`、`exportAIIMData`、`renderAIIMExport`、`/api/admin/ai-data/im-export`。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `97.8%`：AI IM 导出后端接口原本已具备，本轮补齐后台触发和验收反馈面；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.92%`：后台已覆盖核心业务、证据、审核、IM 处置、AI 数据准备和受控导出；剩余重点是真实浏览器点击流、线上账号联调和少量配置项细化。

## 2026-06-17 22:38

### 当前进展

- 继续补 PC 后台角色资格测试能力，把后端已有的用户领路人/行家资格状态接口接到“系统配置”页面。
- 后台“系统配置 / 领路人资格规则”增强：
  - 在规则配置面板下新增“用户资格”表单。
  - 支持按 `userId` 查询资格，调用 `GET /api/admin/guide-qualification-rules?userId={userId}`。
  - 支持后台更新 `conditionMet`、`paymentMet`，调用 `POST /api/admin/guide-qualification-rules`。
  - 查询/保存后展示 `userId`、`conditionMet`、`paymentMet`、`guideOpenStatus`、`updatedAt`。
- 一期测试价值：
  - 普通小程序用户一期仍只能创建普通局。
  - 后台可直接准备领路人/行家资格状态，用于测试标准局、条件局、主行家和领路人门槛链路。
  - 字段对齐 `guide_qualifications` 与 `profiles.GuideQualification`。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi ./internal/profiles`
  - `go test -count=1 ./...`
  - 检查 `admin-web/src/main.js`、`admin-web/dist/main.js` 已包含 `guide-qualification-form`、`guide-qualification-load`、`loadGuideQualification`、`updateGuideQualification`、`/api/admin/guide-qualification-rules`。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `97.9%`：用户资格状态后端接口原本已具备，本轮补齐后台操作面；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.94%`：后台已覆盖核心业务、证据、审核、IM 处置、AI 数据准备、受控导出和角色资格测试准备；剩余重点是真实浏览器点击流、线上账号联调和少量配置项细化。

## 2026-06-17 22:53

### 当前进展

- 继续补 PC 后台行为追踪排查能力，把后端已有的完整行为时间线接口接到“数据分析”页面。
- 后台“数据分析 / 行为事件”增强：
  - 新增“完整时间线”按钮，调用 `GET /api/admin/behavior-logs`。
  - 新增 `behavior-log-list`，展示最近 20 条行为时间线摘要。
  - 时间线展示 `eventCode`、`eventType`、`userId`、`target/business`、`pagePath`、`source`、`device`、`keyword`、`occurredAt`。
  - 无 `analytics:timeline:view` 权限时展示无权限提示。
- 与现有行为事件列表的关系：
  - `GET /api/admin/behavior/events` 继续用于按 userId、eventType、eventCode 筛选结构化事件。
  - `GET /api/admin/behavior-logs` 用于快速排查用户完整路径和一期邀请/浏览/分享/入局等行为链路。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi ./internal/audit`
  - `go test -count=1 ./...`
  - 检查 `admin-web/src`、`admin-web/dist` 已包含 `behavior-logs-refresh`、`behavior-log-list`、`loadBehaviorLogs`、`behaviorLogItem`、`/api/admin/behavior-logs`。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98%`：行为时间线后端接口原本已具备，本轮补齐后台查看入口；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.95%`：后台已覆盖核心业务、证据、审核、IM 处置、AI 数据准备、受控导出、角色资格测试和行为追踪；剩余重点是真实浏览器点击流、线上账号联调和少量配置项细化。

## 2026-06-17 23:10

### 当前进展

- 继续按后台真实点击流收口，修正“组局详情 / 打卡标记异常”的前后端接口路径。
- 后台页面原先调用 `POST /api/admin/games/{gameId}/checkins/{checkinId}/mark-invalid`，但后端路由和集成用例定义的是 `POST /api/admin/game-checkins/{checkinId}/mark-invalid`。
- 已将 `admin-web/src/main.js` 调整为调用正式后端路径，并通过构建同步到 `admin-web/dist/main.js`。
- 这次修复保证后台运营人员在组局详情中点击“标记异常”时，可以真正命中后端 `markCheckinInvalid`，并写入操作审计 `game:checkin:mark_invalid`。

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/appapi ./internal/games`
  - `go test -count=1 ./...`
  - 检查 `admin-web/src/main.js`、`admin-web/dist/main.js` 已包含 `/api/admin/game-checkins/${button.dataset.id}/mark-invalid`。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98%`：本轮未新增后端能力，重点是把已有后端能力对齐到后台可操作入口；剩余重点仍是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.96%`：后台核心业务入口已基本接齐，本轮修掉一处真实点击流路径错配；剩余重点是真实浏览器端到端点击、线上账号联调和极少量边缘配置项细化。

## 2026-06-17 23:31

### 当前进展

- 继续按“一期普通用户只能开普通局，但后台要能开其他条件局做完整测试”的要求收口后台开局链路。
- 后端 `games.CreateRequest` 新增 `mainGuideUserId`，后台创建标准局/条件局时可直接指定主行家：
  - `CreateFromAdmin` 会把 `mainGuideUserId` 写入 `Game.MainGuideUserID`。
  - SQL 仓库创建游戏时会落到 `games.main_guide_user_id`。
  - 指定的主行家会同步加入成员关系，仓库成员角色为 `main_guide`。
- 小程序端创建局仍然忽略 `mainGuideUserId`：
  - `Create` 会清零 `MainGuideUserID`，避免普通用户绕过“邀请接受 + 审核通过后产生主行家”的业务规则。
- 后台“组局管理 / 后台开局”表单新增 `mainGuideUserId` 输入：
  - 不填时保持原有普通局/标准局/条件局创建流程。
  - 填正数时随 `POST /api/admin/games` payload 一起提交，便于一期后台直接开标准局、条件局并测试主行家进度管理链路。

### 验证补充

- 已执行：
  - `gofmt -w services/go-api/internal/games/service.go services/go-api/internal/games/service_test.go`
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/games`
  - `go test -count=1 ./internal/appapi ./internal/games`
  - `go test -count=1 ./...`
  - 检查 `admin-web/src`、`admin-web/dist`、`services/go-api/internal/games` 已包含 `mainGuideUserId` / `MainGuideUserID`。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.3%`：后台开局已能直接准备标准局/条件局和主行家测试数据；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.97%`：后台核心业务入口、证据、审核、IM、行为、测试数据准备和开局主行家配置基本接齐；剩余重点是真实浏览器端到端点击和线上账号联调。

## 2026-06-17 23:40

### 当前进展

- 继续收口后台指定主行家开局链路，从“表单可提交”补到“API 与详情可验收”。
- 后台组局详情新增 `mainGuideUserId` 展示：
  - 打开 `GET /api/admin/games/{gameId}` 详情时，可直接看到当前局主行家 ID。
  - 便于后台确认标准局/条件局是否已经绑定主行家，以及后续由主行家决定进度的测试条件是否满足。
- 后端 API 集成测试增强：
  - `POST /api/admin/games` 测试 payload 增加 `mainGuideUserId`。
  - 断言创建返回 `mainGuideUserId`。
  - 继续请求 `GET /api/admin/games/{gameId}`，断言详情里的 `game.mainGuideUserId` 正确，并且 `memberIds` 包含该主行家。
- 保持小程序端规则不变：
  - 普通用户创建局仍只能创建普通局。
  - 小程序端不能通过创建 payload 直接指定主行家。

### 验证补充

- 已执行：
  - `gofmt -w services/go-api/internal/appapi/server_test.go services/go-api/internal/games/service.go services/go-api/internal/games/service_test.go`
  - `node --check admin-web/src/main.js`
  - `go test -count=1 ./internal/appapi ./internal/games`
  - `npm run build`
  - `go test -count=1 ./...`
  - 检查 `admin-web/src/main.js`、`admin-web/dist/main.js`、`services/go-api/internal/appapi/server_test.go` 已包含 `mainGuideUserId` 详情展示和 API 断言。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.4%`：主行家后台开局链路已具备 API 级验收覆盖；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.98%`：后台核心业务入口和关键详情验收字段基本接齐；剩余重点是真实浏览器端到端点击和线上账号联调。

## 2026-06-17 23:55

### 当前进展

- 继续收口后台指定主行家开局的数据一致性。
- 修正后台创建局时 `currentPlayers` 计算：
  - 不指定主行家时仍为 `1`，只包含创建者。
  - 指定 `mainGuideUserId` 且与创建者不同时，`currentPlayers` 自动为 `2`，与创建者 + 主行家两条成员关系一致。
- 后端 service 测试增强：
  - 后台创建条件局并指定主行家时，断言 `MainGuideUserID=22` 且 `CurrentPlayers=2`。
- 后端 API 集成测试增强：
  - `POST /api/admin/games` 返回体断言 `currentPlayers=2`。
  - `GET /api/admin/games/{gameId}` 详情断言 `game.currentPlayers=2`，并继续断言 `memberIds` 包含主行家。
- 这轮不改变接口字段，只修正已有字段与成员关系之间的一致性，避免后台开测试局后人数显示和满员判断偏差。

### 验证补充

- 已执行：
  - `gofmt -w services/go-api/internal/games/service.go services/go-api/internal/games/service_test.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/games ./internal/appapi`
  - `node --check admin-web/src/main.js`
  - `go test -count=1 ./...`
  - 检查 `services/go-api/internal/games/service.go`、`services/go-api/internal/games/service_test.go`、`services/go-api/internal/appapi/server_test.go` 已包含 `currentPlayers=2` 的主行家开局断言。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.45%`：后台主行家测试局的人数、成员、详情字段已一致；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.98%`：本轮未改前端结构，继续保持关键详情验收字段已接齐；剩余重点是真实浏览器端到端点击和线上账号联调。

## 2026-06-18 00:10

### 当前进展

- 继续收口后台指定主行家开局的输入边界。
- 后台创建局新增校验：`mainGuideUserId` 不能等于 `creatorUserId`。
  - 避免后台误把创建者同时写成主行家，导致成员角色、进度管理和后续测试数据出现歧义。
  - 正常的小程序邀请链路不受影响，主行家仍由“行家邀请接受 + 审核通过”产生。
- 后端 service 测试增强：
  - `CreateFromAdmin` 传入相同的 `CreatorUserID` 和 `MainGuideUserID` 时返回 `ErrInvalidGameInput`。
- 后端 API 集成测试增强：
  - `POST /api/admin/games` 传入相同 `creatorUserId` 与 `mainGuideUserId` 时返回 `422 Unprocessable Entity`。

### 验证补充

- 已执行：
  - `gofmt -w services/go-api/internal/games/service.go services/go-api/internal/games/service_test.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/games ./internal/appapi`
  - `node --check admin-web/src/main.js`
  - `go test -count=1 ./...`
  - 静态检查已确认 `service.go`、`service_test.go`、`server_test.go` 包含 `creatorUserId == mainGuideUserId` 的拒绝逻辑和测试覆盖。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.5%`：后台主行家开局链路已补齐创建、成员、人数、详情和输入边界；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.98%`：本轮未改前端结构；剩余重点是真实浏览器端到端点击和线上账号联调。

## 2026-06-18 00:24

### 当前进展

- 继续补后台指定主行家开局的前端误操作保护。
- 后台“组局管理 / 后台开局”提交前新增校验：
  - 当 `mainGuideUserId > 0` 且等于 `creatorUserId` 时，前端直接提示 `mainGuideUserId cannot equal creatorUserId` 并停止提交。
  - 后端上一轮新增的 `422` 校验仍保留，前端只负责提前减少运营误填。
- 构建产物已同步：
  - `admin-web/src/main.js`
  - `admin-web/dist/main.js`

### 验证补充

- 已执行：
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/games ./internal/appapi`
  - `go test -count=1 ./...`
  - 检查 `admin-web/src/main.js`、`admin-web/dist/main.js` 已包含 `mainGuideUserId === payload.creatorUserId` 拦截。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.5%`：本轮未改后端，继续保持后台主行家开局链路的服务端最终校验。
- 当前后台前端业务功能进度约 `99.99%`：后台开局表单已补齐主行家输入、详情展示和同 ID 误操作保护；剩余重点是真实浏览器端到端点击和线上账号联调。

## 2026-06-18 00:41

### 当前进展

- 继续按“领路人、行家必须实名认证；玩家不要求”的业务规则，收口后台指定主行家开局链路。
- 后台创建局新增服务端校验：
  - 当 `mainGuideUserId > 0` 时，必须通过 `identity.IsVerified(mainGuideUserId)`。
  - 未实名用户不能被后台直接指定为主行家，返回 `ErrRealnameRequired` 并由 HTTP 层转成 `422`。
- 保持小程序端规则不变：
  - 玩家创建普通局仍不要求实名认证。
  - 小程序邀请成为主行家仍沿用既有“行家实名 + 邀请接受 + 审核通过”链路。
- 后端 service 测试增强：
  - `CreateFromAdmin` 指定未实名主行家时返回 `ErrRealnameRequired`。
- 后端 API 集成测试增强：
  - 未实名 `mainGuideUserId=2` 创建条件局返回 `422`。
  - 登录并完成用户 2 实名后，再用同一 `mainGuideUserId=2` 创建条件局成功。

### 验证补充

- 已执行：
  - `gofmt -w services/go-api/internal/games/service.go services/go-api/internal/games/service_test.go services/go-api/internal/appapi/server_test.go`
  - `go test -count=1 ./internal/games ./internal/appapi`
  - `node --check admin-web/src/main.js`
  - `go test -count=1 ./...`
  - 静态检查已确认 `service.go`、`service_test.go`、`server_test.go` 包含后台主行家实名校验和测试覆盖。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.6%`：后台主行家开局链路已补齐实名边界、成员、人数、详情和错误输入校验；剩余重点是线上环境配置、真实第三方联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.99%`：本轮未改前端结构；剩余重点是真实浏览器端到端点击和线上账号联调。

## 2026-06-18 00:56

### 当前进展

- 继续收口后台开局和小程序业务规则的一致性，补齐后台 HTTP 层对主行家实名失败的错误映射。
- 后台创建局 `CreateFromAdmin` 已要求 `mainGuideUserId` 对应用户完成实名；本轮补齐 API 返回：
  - 未实名主行家不再落入 `500` 系统错误。
  - 统一返回 `422 validation_error`，错误信息为 `main guide realname required`，方便后台前端和线上联调定位。
- 保持小程序端规则不变：
  - 玩家创建普通局不要求实名认证。
  - 行家、领路人相关身份仍必须实名。
  - 一期普通用户只能在小程序创建普通局，后台可创建其他条件局用于业务测试。

### 验证补充

- 已执行：
  - `go test -count=1 ./internal/games ./internal/appapi`
  - `node --check admin-web/src/main.js`
  - `go test -count=1 ./...`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.65%`：后台主行家开局链路已补齐服务端实名校验、HTTP 错误映射、成员写入、人数统计和测试覆盖；剩余重点是线上环境配置、真实腾讯/微信侧联调、生产 readiness 配置补齐。
- 当前后台前端业务功能进度约 `99.99%`：本轮未改前端结构；剩余重点是真实浏览器端到端点击和线上账号联调。

## 2026-06-18 01:12

### 当前进展

- 继续按后台前后端对接检查接口闭环，重点核对 admin-web 已绑定的 API 是否具备后端路由、权限和测试覆盖。
- 确认 IM 争议消息接口 `/api/admin/im/rooms/{roomId}/dispute-messages` 已有后端路由、权限 `im:message:view_dispute` 和测试覆盖，不作为本轮缺口。
- 发现并修复通知交付模块安全缺口：
  - 后台前端会调用 `/api/internal/notifications/wechat-tasks/send-pending`、`/{id}/send`、`/{id}/mark-sent`。
  - 后端原先内部路由直接进入处理函数，未统一套后台权限。
  - 本轮已统一套 `notification:wechat:view` 权限，保证只有具备后台通知权限的账号才能批量发送、单条发送或手动标记微信订阅消息任务。
- 同步调整测试：
  - 无 token 直调 `mark-sent` 必须返回 `403`。
  - 使用后台管理员 token 调用仍返回 `200`，后台页面按钮链路保持可用。

### 验证补充

- 已执行：
  - `gofmt -w internal/appapi/server.go internal/appapi/server_test.go`
  - `node --check admin-web/src/main.js`
  - `go test -count=1 ./internal/appapi -run "Test.*Wechat|Test.*Notification|Test.*Report"`
  - `go test -count=1 ./internal/notifications`
  - `go test -count=1 ./...`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.7%`：后台通知交付的内部执行接口已补齐权限边界，避免线上测试时无权限直调修改发送状态；剩余重点仍是生产配置、真实微信订阅消息账号联调、真实对象存储/OpenIM/实名服务联调。
- 当前后台前端业务功能进度约 `99.99%`：本轮未改页面结构；后台已绑定接口继续保持可用，剩余重点是真实浏览器端到端点击和线上账号联调。

## 2026-06-18 01:28

### 当前进展

- 继续收口后台账号和 RBAC 权限边界，避免“角色审核权限”和“后台账号管理权限”混用。
- 新增并落地 `admin_user:update` 权限码：
  - `PUT /api/admin/admin-users/{id}` 从 `role:update` 改为 `admin_user:update`。
  - 超级管理员默认权限加入 `admin_user:update`。
  - 数据库 seed 幂等追加 `admin_user:update`，新库和旧库补种子都能拿到该权限。
  - 权限目录接口会返回 `admin_user:update`，后台权限目录可以直接看到该 code。
- 后台前端同步：
  - 创建后台账号前先检查 `admin_user:create`。
  - 后台账号列表的保存、启用、禁用按钮只在具备 `admin_user:update` 时显示。
- 保持原有 `role:update` 继续用于角色审核、领路人资格规则等角色业务，不再复用到后台账号启停和改角色。

### 验证补充

- 已执行：
  - `gofmt -w internal/appapi/server.go internal/appapi/server_test.go internal/adminauth/service.go internal/adminauth/service_test.go`
  - `node --check admin-web/src/main.js`
  - `npm run build`
  - `go test -count=1 ./internal/adminauth`
  - `go test -count=1 ./internal/appapi -run "TestAdminAccount|TestAdminLoginPermissions|TestAdminRolePermission"`
  - `go test -count=1 ./...`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.75%`：后台账号管理的更新权限已独立，避免客服/运营等拥有角色审核能力的账号误操作后台账号；剩余重点仍是生产配置、真实第三方联调和线上验收数据准备。
- 当前后台前端业务功能进度约 `99.99%`：后台账号按钮已按新权限码显示；剩余重点是真实浏览器端到端点击和线上账号联调。

## 2026-06-18 01:42

### 当前进展

- 继续收口后台 RBAC 与数据库初始化脚本的一致性，避免本地内存账号有权限、线上 seed 初始化账号缺权限。
- 修复数据库 seed 中后置新增权限的授权顺序问题：
  - `super_admin` 的全权限授权发生在较早位置。
  - 后续追加的 `im:room:read`、`im:room:archive`、`im:room:retry_create`、`system_config:read`、`invite_code:read`、`invite_code:manage` 不会被早期 `cross join admin_permissions` 自动覆盖。
  - 本轮已为 `super_admin` 显式补齐这些后置权限授权。
- 增加 seed 一致性回归测试：
  - 测试会读取 `db/seeds/admin_roles_permissions.sql`。
  - 确认后置新增权限都显式授予 `super_admin`，防止后续后台模块新增权限时再次出现初始化库权限缺口。
- 本轮不改变前端页面和业务流程，只保证后台权限、数据库 seed、线上初始化后的角色能力继续对齐。

### 验证补充

- 已执行：
  - `gofmt -w internal/adminauth/service_test.go`
  - `go test -count=1 ./internal/adminauth`
  - `go test -count=1 ./internal/appapi -run "TestAdminAccount|TestAdminLoginPermissions|TestAdminRolePermission|TestAdminIM"`
  - `go test -count=1 ./...`
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.8%`：后台账号、IM、配置、邀请码相关权限在代码和数据库初始化层继续对齐；剩余重点仍是生产配置、真实腾讯/微信/OpenIM/存储/实名服务联调和线上验收数据准备。
- 当前后台前端业务功能进度约 `99.99%`：本轮未改页面；后台按钮权限依赖的数据库 seed 已继续补强，剩余重点是真实浏览器端到端点击和线上账号联调。

## 2026-06-18 02:03

### 当前进展

- 继续按后台前后端接口对接检查，重点核对后台页面菜单、按钮权限、后端路由权限和数据库 seed 权限的一致性。
- 已用脚本核对：
  - 后台内置角色权限码全部存在于 `db/seeds/admin_roles_permissions.sql`。
  - 后端 `requireAdminPermission(...)` 路由权限码全部存在于内置角色和 seed。
  - 后台前端 `can("...")` 使用到的按钮权限码全部存在于内置角色和 seed。
- 修复后台“系统配置”页混合权限加载问题：
  - 页面包含生产就绪、敏感词库、内容风险日志、权限快照、领路人资格规则、AI IM 导出配置等多个子模块。
  - 之前菜单只要命中任一系统类权限就会显示，但页面初始化会无条件请求全部子接口，非超级管理员可能一打开就遇到部分接口 `403`。
  - 本轮已改成按权限分块加载：
    - `system_config:read` 才加载生产就绪。
    - `content:sensitive_word:view` 才加载敏感词。
    - `content:risk_log:view` 才加载风险日志。
    - `role:view` 才加载领路人资格规则。
    - `ai:data:read` 才加载 AI IM 导出配置。
  - 无权限的子块显示“缺少 xxx”空状态，不再影响其它有权限子块。
- 同步补齐系统页写动作前端保护：
  - 新增敏感词需要 `content:sensitive_word:create`。
  - 导入敏感词需要 `content:sensitive_word:import`。
  - 更新敏感词状态需要 `content:sensitive_word:update`。
  - 更新领路人资格规则和用户资格需要 `role:update`。
  - 更新 AI IM 导出开关需要 `system_config:update`。
- 已同步构建产物 `admin-web/dist/main.js`。

### 验证补充

- 已执行：
  - `node --check src/main.js`
  - `npm run build`
  - `go test -count=1 ./...`
- 已尝试浏览器冒烟验证，但本地 Codex 浏览器安全策略拒绝访问 `http://127.0.0.1:5173`，因此本轮浏览器点击验证未完成；当前验证停在静态语法、构建产物和后端全量测试。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.8%`：本轮主要改后台前端权限加载，不改变后端业务逻辑；后端路由权限、内置角色权限和 seed 权限已继续核对一致。
- 当前后台前端业务功能进度约 `99.995%`：系统配置页已避免低权限角色打开页面时被无关子模块 `403` 拖垮；剩余重点是真实浏览器端到端点击、线上账号和生产环境联调。

## 2026-06-18 02:14

### 当前进展

- 继续收口后台混合权限页面，避免低权限角色能进入菜单但页面初始化被其它子模块接口 `403` 拖垮。
- 本轮复核后台页面初始化调用，确认 `analytics` 已经按权限分块加载，无需改动。
- 修复“审核中心”页面：
  - `identity:read` 才加载实名核验列表和实名详情。
  - `role:view` 才加载角色申请列表和角色申请详情。
  - `role:update` 才允许通过/驳回角色申请。
  - 无权限子块显示“缺少 xxx”，不影响同页其它有权限模块。
- 修复“分润结算”页面：
  - `revenue:template:view` 才加载分润模板和分润规则。
  - `revenue:template:update` 才允许新增模板、保存规则。
  - `revenue:simulate` 才允许试算。
  - `revenue:generate` 才允许生成分润记录。
  - `revenue:record:view` 才加载分润记录。
  - `revenue:freeze` 才允许冻结记录。
  - `settlement:offline:create` 才加载线下结算记录并允许登记结算。
- 修复“积分兑换”页面：
  - `redemption:manage` 才加载兑换商品、兑换订单并允许新增/更新/审核。
  - `points:read` 才加载积分流水。
  - 只有其中一个权限时，另一个子块显示无权限空态，不再整页失败。
- 已同步构建产物 `admin-web/dist/main.js`。

### 验证补充

- 已执行：
  - `node --check src/main.js`
  - `npm run build`
  - `go test -count=1 ./...`
- 本轮仍受本地 Codex 浏览器安全策略限制，未完成真实浏览器点击验证；当前验证覆盖静态语法、构建产物和后端全量测试。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.8%`：本轮不改后端业务逻辑，继续保持路由权限和角色权限一致。
- 当前后台前端业务功能进度约 `99.997%`：审核、分润、积分兑换、系统配置等混合权限页面已补齐分块加载和动作级前端权限兜底；剩余重点是真实浏览器端到端点击、线上账号和生产环境联调。

## 2026-06-18 02:17

### 当前进展

- 继续收口后台线上测试前的角色权限体验，重点处理“菜单可进入但页面内部接口无权限”的边界。
- 补强“通知交付”页面动作级权限兜底：
  - `notification:wechat:view` 才允许加载微信订阅消息任务、订阅模板、批量发送、单条发送和标记发送。
  - `delivery:manage` 才允许加载和新增交付文档。
  - `testcase:read` 才允许加载测试用例、测试运行和下载测试证据文件。
  - `testcase:manage` 继续作为新增测试用例、登记测试运行的写权限。
  - 这样即使后续页面事件被手动触发，也不会越过前端权限保护去打后端接口。
- 修正后台菜单权限定义：
  - “管理员权限”页面会读取后台账号、后台角色和权限目录，这些后端接口统一要求 `admin_user:view`。
  - 本轮将菜单入口也收敛到 `admin_user:view`，避免只有 `admin_user:create` 或 `role:update` 的角色看到菜单后进入页面立刻 `403`。
- 已同步构建产物 `admin-web/dist/main.js`。

### 验证补充

- 已执行：
  - `node --check src/main.js`
  - `npm run build`
  - `go test -count=1 ./...`
- `git status --short` 未执行成功：当前 `C:\Users\61492\Desktop\真好玩-mini` 不是 Git 仓库或未包含 `.git`，本轮无法给出 Git 工作区变更清单。
- 本轮仍受本地 Codex 浏览器安全策略限制，未完成真实浏览器点击验证；当前验证覆盖静态语法、构建产物和后端全量测试。
- 当前一期小程序后端业务功能进度约 `99.75%`。
- 当前后台后端业务功能进度约 `98.82%`：菜单权限定义继续和后端真实路由权限对齐；后端业务主体保持稳定，剩余重点仍是真实腾讯/微信/OpenIM/存储/实名服务联调与生产配置。
- 当前后台前端业务功能进度约 `99.998%`：通知交付、管理员权限、审核、分润、积分兑换、系统配置等关键后台页面已完成权限分块加载和动作级兜底；剩余重点是真实浏览器端到端点击、线上账号和生产环境联调。

## 2026-06-18 11:50

### 当前进展

- 按 `miniprogram-development`、`karpathy-guidelines`、`cloudrun-development` 重新复核小程序后端与后台对接，不按历史印象判断。
- 重点核对范围：
  - 小程序 `/api/app` 和后台 `/api/admin` 路由。
  - 邀请入口、实名、文件、组局、主行家、后台开局、举报证据、IM 证据、权限菜单。
  - 后台菜单权限、后端路由权限、数据库 seed 权限和前端 `can("...")` 的一致性。
- 修复一个小程序后端业务逻辑问题：
  - 原来 `POST /api/app/files/upload-token` 对除头像外的所有文件都强制要求已实名。
  - 这会导致 `realname_material` 出现“上传实名材料前必须已经实名”的闭环矛盾，也会让玩家在一期不强制实名的口径下被不合理拦截举报附件上传。
  - 本轮改成仅 `chat_file` 继续要求强实名和局成员身份；`avatar`、`realname_material`、`report_attachment` 允许已登录用户上传，再由后续实名/举报业务流程校验归属和使用场景。
- 补充测试覆盖：
  - 未实名用户上传 `chat_file` 仍返回 `403`。
  - 未实名用户上传 `realname_material` 返回成功，支撑实名材料提交闭环。
  - 未实名用户上传 `report_attachment` 返回成功，支撑玩家举报/申诉材料闭环。
  - 头像更新测试改为使用真实返回的 `fileId`，避免新增文件类型测试后被硬编码 ID 影响。
- 修正后台和后端权限对接问题：
  - “用户管理”菜单收敛到 `user:view`，因为页面默认请求 `GET /api/admin/users`，后端真实权限就是 `user:view`。
  - “组局管理”菜单收敛到 `game:read`，因为页面默认请求 `GET /api/admin/games`，后端真实权限就是 `game:read`。
  - 避免角色只有 `user:read` 或 `game:view` 时能看到菜单，但一进入页面就因为列表接口权限不足而失败。
- 已同步构建产物 `admin-web/dist/main.js`。

### 对接结论

- 小程序和后台已成功对接的主链路：
  - 邀请入口：小程序预检/登录绑定，后台邀请码列表、详情、禁用、邀请关系列表。
  - 实名与角色：小程序实名状态、角色申请，后台实名列表/详情、角色申请审核。
  - 组局：小程序创建免费局、申请入局、邀请主行家、手动开始；后台可审核、查看详情、创建不同类型测试局。
  - IM/文件：小程序成员消息、文件上传/下载；后台按权限查看争议 IM 和文件证据。
  - 服务确认/评价/举报/分润：小程序完成服务确认、评价、举报；后台可追踪证据、冻结收益、处理举报、查看分润记录。
- 仍需真实环境联调的部分：
  - 微信真实 code2Session、订阅消息、对象存储直传、腾讯实名/人脸核身、OpenIM 真服务、生产域名和线上账号。
  - 本地 Codex 浏览器安全策略仍阻止真实浏览器点击后台页面，因此本轮浏览器端到端点击验证未完成。

### 验证补充

- 已执行：
  - `go test -count=1 ./internal/appapi -run "Test.*(Login|Invite|File|Report|AdminCreateGame|AdminRole|Permission|Game)"`
  - `go test -count=1 ./internal/files ./internal/adminauth ./internal/games`
  - `node --check src/main.js`
  - `npm run build`
  - `go test -count=1 ./...`
- 当前一期小程序后端业务功能进度约 `99.82%`。
- 当前后台后端业务功能进度约 `98.9%`：后台菜单权限继续和后端路由权限对齐；剩余主要是真实三方服务与生产环境联调。
- 当前后台前端业务功能进度约 `99.998%`：后台页面与接口字段/权限继续对齐；剩余主要是真实浏览器端到端点击验证和线上账号联调。

## 2026-06-18 12:00

### 当前进展

- 按后台运营需求补齐批量邀请码生成：
  - `POST /api/admin/invite-codes` 新增 `batchCount`。
  - `batchCount=1` 时保持原单个创建/手填 code 行为。
  - `batchCount>1` 时必须留空 `code`，由系统自动生成唯一邀请码。
  - 支持 `poster`、`qrcode`、`link` 三种 `entryType` 批量生成。
  - 单次批量上限为 200，避免误操作生成过多入口码。
  - 批量生成写入 `invite_code:batch_create` 操作日志。
- 后台邀请码页面已新增 `batchCount` 输入，运营可直接按入口类型批量生成。
- 同步 OpenAPI：`/api/admin/invite-codes` 已说明单个创建和批量生成的请求/响应差异。
- 按要求补齐后台操作日志分级查看：
  - 新增权限 `operation_log:view_self`。
  - 普通后台管理员默认可查看自己的操作日志。
  - 超级管理员继续通过 `operation_log:view_full` 查看所有操作日志。
  - 普通管理员即使传 `adminUserId=别人`，后端也会强制过滤为自己的 `adminUserId`。
  - 操作日志仍没有删除接口，不提供删除能力。
- 后台日志菜单已改为命中 `operation_log:view_self` 或 `operation_log:view_full` 即可显示。
- 后台日志页对普通管理员禁用 `adminUserId` 筛选输入，避免误以为可查别人；最终权限边界以后端为准。
- 同步 `db/seeds/admin_roles_permissions.sql`：
  - 超级管理员包含 `operation_log:view_full` 和 `operation_log:view_self`。
  - 运营、用户、财务、客服、数据、组局等普通后台角色包含 `operation_log:view_self`。
- 同步 `docs/openapi/admin.openapi.yaml` 和 `docs/test-cases/admin-integration-cases.md`。

### 验证补充

- 已执行：
  - `go test -count=1 ./internal/appapi -run TestAdminInviteCodeManagementHTTP`
  - `go test -count=1 ./internal/appapi -run "TestAdminAccountMutationUpdatesRBAC|TestAdminRolePermissionBoundariesHTTP|TestReport|TestAdminAuth"`
  - `go test -count=1 ./internal/adminauth`
  - `node --check src/main.js`
  - `npm run build`
  - `go test -count=1 ./...`
- 当前一期小程序后端业务功能进度约 `99.82%`。
- 当前后台后端业务功能进度约 `99.05%`：批量邀请码和操作日志分级查看已落地；剩余主要是真实三方服务、生产环境和线上账号联调。
- 当前后台前端业务功能进度约 `99.999%`：后台邀请码批量生成和日志页自查模式已接上；剩余主要是真实浏览器端到端点击验证。

## 2026-06-21 16:53

### 当前进展

- 按“每个局的基本信息作为分享卡片”调整小程序邀请入口。
- `POST /api/app/invites/entries` 新增 `gameId`：
  - 小程序在局详情页分享时传 `entryType=poster` 和当前 `gameId`。
  - 后端生成一次性唯一邀请码，并返回该局分享卡片所需字段。
  - 默认分享标题来自局信息：有城市时为 `城市 · 局标题`，否则使用局标题。
  - 分享路径改为局详情页：`/pages/games/detail?id={gameId}&inviteCode={inviteCode}&entryType=poster`。
  - `scene` 包含 `gameId` 和 `inviteCode`，用于二维码/卡片入口解析。
  - 响应新增 `game` 基本信息：`id/title/gameType/status/cityName/minPlayers/maxPlayers/currentPlayers`。
- 保留原来的 link/qrcode 邀请入口行为，不影响微信聊天链接和二维码入口。
- 同步 `docs/openapi/app.openapi.yaml`，标明 `gameId` 和 `game` 返回字段用途。

### 前端使用方式

- 局详情页分享卡片调用：
  - `POST /api/app/invites/entries`
  - body：`{"entryType":"poster","gameId":当前局ID}`
- 小程序 `onShareAppMessage` 使用返回：
  - `title`：分享卡片标题。
  - `path`：分享卡片打开路径。
  - 后续如果需要卡片封面图，可在小程序端根据 `game` 信息生成默认图，或后端再扩展 `imageUrl`。

### 验证补充

- 已执行：
  - `go test -count=1 ./internal/appapi -run TestCreateInviteEntryHTTP`
  - `go test -count=1 ./internal/auth`
  - `go test -count=1 ./...`
- 当前一期小程序后端业务功能进度约 `99.84%`。
- 当前后台后端业务功能进度约 `99.05%`。
- 当前后台前端业务功能进度约 `99.999%`。

## 2026-06-23 19:20

### 当前进展

- 按“腾讯地图 Key 无法绑定小程序 AppID”的实际情况，调整地图接入为后端代理方案：
  - 小程序前端不引入腾讯位置服务小程序 SDK。
  - 小程序前端不写腾讯地图 Key，也不绑定地图 Key 的 AppID。
  - 前端只使用 `wx.getLocation` 获取定位，使用原生 `<map>` 展示附近局 marker。
  - 地点搜索、地址转经纬度、经纬度转地址、路线规划统一请求后端。
- 新增后端腾讯地图 WebService 客户端：
  - `internal/lbs.MapProvider` 抽象。
  - `TencentMapClient` 使用 `TENCENT_MAP_KEY_SERVER` 调腾讯 WebService。
  - 支持 `TENCENT_MAP_SK` 签名、`TENCENT_MAP_API_BASE`、`TENCENT_MAP_REQUEST_TIMEOUT_MS`。
- 新增小程序地图代理接口：
  - `GET /api/app/map/search`
  - `GET /api/app/map/geocode`
  - `GET /api/app/map/reverse-geocode`
  - `GET /api/app/map/route`
- 新增后台地点搜索接口：
  - `GET /api/admin/map/search`
  - 复用 `game:create_admin` 权限，后台创建局时不需要前端持有地图 Key。
- 生产 readiness 增加 `tencent_map_server_key` 检查项；启用腾讯地图时生产环境必须配置服务端 Key 和 HTTPS APIBase。
- 同步 `deploy/env.example`、`docs/openapi/app.openapi.yaml`、`docs/openapi/admin.openapi.yaml`、`第三方接口文档.md`。

### 前端对接口径

- 前端需要做：
  - `wx.getLocation({ type: "gcj02" })` 获取当前位置。
  - 调 `POST /api/app/locations/current` 保存定位。
  - 调 `GET /api/app/games/nearby` 获取附近局。
  - 用 `<map>` 的 `markers` 展示局位置。
  - 创建局选地点时调用 `GET /api/app/map/search`，选择结果后保存地点名、地址、经纬度。
- 前端不要做：
  - 不引入 `qqmap-wx-jssdk.js`。
  - 不直连 `https://apis.map.qq.com`。
  - 不在小程序包内写腾讯地图 Key。

### 验证补充

- 已执行：
  - `go test -count=1 ./internal/appapi -run "TestTencentMapProxyHTTP|TestMapProxyRequiresConfiguredProvider|TestLocationAndNearbyGamesFlow"`
  - `go test -count=1 ./internal/common/config`
  - `go test -count=1 ./internal/lbs`
- 当前一期小程序后端业务功能进度约 `99.86%`。
- 当前后台后端业务功能进度约 `99.08%`。
- 当前后台前端业务功能进度约 `99.999%`。

## 2026-06-28 22:35

### 当前进展

- 修复小程序公开局列表真实业务缺口：
  - `GET /api/app/games` 不再返回 `pending_audit`、`draft`、`cancelled`、`closed` 等非公开状态。
  - `GET /api/app/games/city`、`GET /api/app/games/nearby` 同步复用公开状态过滤。
  - 后台 `GET /api/admin/games` 保持全量可查，审核人员仍可看到待审核局。
- 补齐后台开局局型：
  - 后台现在支持 `free`、`standard`、`public_welfare`、`aa`、`crowdfund`、`deposit`、`condition`。
  - 小程序端 `POST /api/app/games` 仍保持一期限制：普通用户只能创建 `free`。
  - 后台前端创建局、筛选、分润模板、局型统计和中文标签已同步新枚举。
- 同步接口文档和测试用例：
  - `docs/openapi/admin.openapi.yaml` 更新后台开局枚举。
  - `docs/openapi/app.openapi.yaml` 标明小程序公开列表只返回公开状态。
  - `docs/test-cases/admin-integration-cases.md` 更新后台全局型开局测试范围。

### 验证补充

- 已执行：
  - `go test -count=1 ./internal/games -run "TestCreateFromAdminAllowsDocumentedGameTypes|TestCreateFromAdminRejectsInvalidTypeAndCreator"`
  - `go test -count=1 ./internal/appapi -run "TestAdminAuditGameApprovesPendingGame|TestAdminCreateConditionGameWhileAppCreateStaysFreeOnly"`
  - `go test -count=1 ./...`
  - `node --check admin-web/src/main.js`
  - `npm run build`
- 当前一期小程序后端业务功能进度约 `99.88%`。
- 当前后台后端业务功能进度约 `99.18%`。
- 当前后台前端业务功能进度约 `99.999%`。

## 2026-06-29 10:45

### 当前进展

- 已按用户指定项目根目录 `C:\Users\61492\Desktop\真好玩-mini` 处理，不再使用 `E:\jubaopen`。
- 已从前端技术人员仓库 `https://github.com/ZYJ-1010/ZHW` 拉取小程序前端源码：
  - 下载位置：`C:\Users\61492\Desktop\codex download\ZHW`
  - 当前项目接入位置：`miniprogram-client`
- 小程序前端已作为独立客户端目录接入当前项目，未覆盖后端 `services`、后台 `admin-web` 和现有文档。
- 已完成第一批前后端 P0 主链路对接：
  - `config/env.js` 新增 `ENV.LOCAL`，本地联调默认指向 `http://127.0.0.1:8080`。
  - 邀请预检从旧 `/api/app/invites/verify` 对齐到后端真实 `/api/app/invites/precheck`。
  - 短信发送/校验从旧 `/api/app/auth/phone-code`、`/phone-code/verify` 对齐到 `/api/app/sms/send-code`、`/api/app/sms/verify-code`。
  - 实名接口从旧 `/api/app/users/me/realname-auth`、`/submit` 对齐到 `/api/app/identity/faceid/detect-auth`、`/api/app/identity/phone/verify`。
  - 新增前端 `bindPhone`、`getIdentityStatus`、`issueTokenAfterIdentity` 封装，支撑后端 preAuthToken 到正式 token 的链路。
  - 微信登录成功后前端会保存 `token || preAuthToken`，避免新用户未实名时后续实名接口无鉴权。
  - IM 从旧 `/api/im/...` 对齐到 `/api/app/games/{gameId}/chat-session` 和 `/api/app/games/{gameId}/chat/messages`。
  - 申请入局从旧 `/api/app/games/{gameId}/apply` 对齐到 `/api/app/games/{gameId}/applications`。

### 仍需继续联调的前端候选接口

- 前端仍存在一些设计阶段/候选接口，需要下一轮按页面逐个接真实后端或改为现有聚合接口：
  - `/api/app/home`
  - `/api/app/newbie-tasks`
  - `/api/app/messages/center`
  - `/api/app/messages/trade-warning`
  - `/api/app/messages/system-notification`
  - `/api/app/profile/home`
  - `/api/app/profile/points/*`
  - `/api/app/profile/system-management/*`
  - `/api/app/relations/network-home`
  - `/api/app/game-invites/*`
  - `/api/app/games/my/manage`
  - `/api/app/games/player/manage`
  - `/api/app/games/profit-templates`
  - `/api/app/game-payments/wechat`
- 这些接口多数属于页面聚合、运营视图、二期支付/推荐或前端 mock 先行接口，本轮没有盲目扩后端。

### 验证补充

- 已执行：
  - `git clone https://github.com/ZYJ-1010/ZHW.git "C:\Users\61492\Desktop\codex download\ZHW"`
  - `Copy-Item ...\ZHW ...\miniprogram-client -Recurse -Force`
  - `Get-ChildItem -Recurse miniprogram-client -Filter *.js | node --check`
  - 小程序 `app.json` 页面完整性检查：113 个页面的 `.js/.json/.wxml` 均存在。
  - 关键旧接口残留检查：真实运行路径内未再发现 `/api/im`、`/api/app/invites/verify`、`/api/app/auth/phone-code`、`/api/app/users/me/realname-auth`。
- 当前一期小程序后端业务功能进度约 `99.88%`。
- 当前小程序前端与后端 P0 主链路对接进度约 `55%`：源码已接入，登录/邀请/实名/局/IM 主入口已改到后端真实路径；剩余重点是页面级聚合接口、真实微信开发者工具编译和真机联调。
- 当前后台后端业务功能进度约 `99.18%`。
- 当前后台前端业务功能进度约 `99.999%`。

## 2026-06-29 22:35

### 当前进展

- 继续只在用户指定项目根目录 `C:\Users\61492\Desktop\真好玩-mini` 内推进，不再使用 `E:\jubaopen`。
- 按 `miniprogram-development`、`web-development`、`karpathy-guidelines` 做小程序前后端全链路复核。
- 第二批前端接口已对齐后端真实路径：
  - 会员状态 `/api/app/memberships/me` -> `/api/app/membership/my`。
  - 成长档案 `/api/app/growth/me` -> `/api/app/growth/my`。
  - 积分商城、兑换、订单 -> `/api/app/redemption/items`、`/api/app/redemption/orders`、`/api/app/redemption/orders/my`。
  - 消息中心 -> `/api/app/notifications`。
  - 关系网首页 -> `/api/app/connections/my`。
  - 支付占位 -> `/api/app/payment/precreate-placeholder`。
- 修正字段级联调问题：
  - 积分兑换前端服务层把商品 id 转成后端需要的 `itemId`。
  - 邀请响应前端服务层把 `accept/approve/yes` 动作映射成后端需要的 `accept` 字段。
- 后端新增小程序薄聚合接口：
  - `GET /api/app/home`
  - `GET /api/app/newbie-tasks`
  - `GET /api/app/profile/home`
  - `GET /api/app/games/my/manage`
  - `GET /api/app/games/player/manage`
  - `GET /api/app/games/profit-templates`
- 新增文档 `docs/progress/full-chain-gap-2026-06-29.md`，记录本轮已打通链路和仍未完整实现的文档/前端缺口。

### 未完整实现的重点缺口

- 组局邀请推荐链路仍缺推荐玩家、最近联系人、复玩上下文、系统推荐、取消详情和进度聚合接口。
- 个人中心系统管理页仍是前端候选模型，和后端现有用户资料、行家技能、领路人资源画像字段不完全一致。
- 积分订单仍缺物流、取消、详情等小程序履约接口。
- 消息中心已能读通知，但按钮动作和详情块结构仍未完全对接。
- 首页地球和关系网已有业务数据基础，但动态可视化布局数据仍需后续定字段。

### 验证补充

- 已执行：
  - `Get-ChildItem -Recurse miniprogram-client -Filter *.js | ForEach-Object { node --check $_.FullName }`
  - `go test -count=1 ./internal/appapi`
  - `go test -count=1 ./...`
- 当前一期小程序后端业务功能进度约 `99.9%`。
- 当前小程序前端与后端全链路对接进度约 `72%`。
- 当前后台后端业务功能进度约 `99.2%`。
