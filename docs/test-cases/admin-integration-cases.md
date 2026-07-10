# PC 后台联调用例包

## 使用口径

- 本文件列后台页面、后台接口和内部任务的联调顺序。
- 后台接口必须携带 `Authorization: Bearer <adminToken>`，或在接口测试中使用对应 `X-Admin-Permissions`。
- 无权限时统一返回 `40301`。

## 菜单和权限码

| 模块 | 主要接口 | 权限码 | 验收级别 |
| --- | --- | --- | --- |
| 后台登录 | `POST /api/admin/auth/login`、`GET /api/admin/permissions/tree` | 登录态 | 验收必测 |
| 用户实名 | `GET /api/admin/identity-verifications`、`GET /api/admin/identity-verifications/{userId}` | `identity:read` | 验收必测 |
| 组局管理 | `GET /api/admin/games?status=pending_audit`、`POST /api/admin/games/{gameId}/audit`、`GET /api/admin/games/{gameId}/milestones`、`GET /api/admin/games/{gameId}/checkins`、`POST /api/admin/game-checkins/{checkinId}/mark-invalid` | `game:read`、`game:update_status`、`game:progress:manage` | 验收必测 |
| 评价成长追溯 | `GET /api/admin/users/{userId}/growth`、`GET /api/admin/games/{gameId}/review-trace` | `user:view`、`game:view` | 验收必测 |
| 人脉画像 | `GET /api/admin/connections`、`GET /api/admin/users/{userId}/connections`、`GET /api/admin/experts/{userId}/skills`、`GET /api/admin/guides/{userId}/resources` | `connection:read`、`profile:read` | 验收必测 |
| 积分兑换 | `GET /api/admin/points/logs`、`GET/POST/PUT /api/admin/redemption/items`、`GET /api/admin/redemption/orders`、`POST /api/admin/redemption/orders/{orderId}/review` | `points:read`、`redemption:manage` | 验收必测 |
| 会员团队 | `GET /api/admin/member-reports`、`GET /api/admin/teams`、`GET /api/admin/teams/{teamId}` | `member_report:read`、`team:read` | 可选入口 |
| 举报申诉 | `GET /api/admin/reports`、`GET /api/admin/reports/{reportId}`、`POST /api/admin/reports/{reportId}/handle`、`POST /api/admin/reports/{reportId}/close` | `report:view`、`report:handle`、`report:close` | 验收必测 |
| 争议 IM | `GET /api/admin/im/rooms/{roomId}/dispute-messages` | `im:message:view_dispute` | 验收必测 |
| 数据导出 | `GET /api/admin/reports/export-templates`、`POST /api/admin/reports/export`、`GET /api/admin/export-tasks`、`GET /api/admin/export-tasks/{taskId}/download-url` | `report_export:create` | 验收必测 |
| 操作日志 | `GET /api/admin/operation-logs` | `operation_log:view_self` / `operation_log:view_full` | 验收必测；普通管理员仅本人，超级管理员全量 |
| AI 数据准备 | `GET /api/admin/ai-data/snapshot`、`POST /api/admin/ai-data/acceptance-fixture`、`POST /api/admin/ai-data/im-export` | `ai:data:read`、`ai:data:seed`、`ai:data:export` | 验收必测 |
| 交付测试留档 | `GET/POST /api/admin/delivery-documents`、`GET/POST /api/admin/test-cases`、`GET/POST /api/admin/test-runs` | `delivery:manage`、`testcase:manage` | 验收必测 |

## 后台验收路径

| 编号 | 场景 | 步骤 | 关键断言 |
| --- | --- | --- | --- |
| ADM-001 | 登录和权限树 | 登录后访问权限树 | 返回角色、菜单、按钮和接口权限 |
| ADM-002 | 强实名审核只读 | 访问实名列表和详情 | 可看到手机号、短信、人脸核身状态，不暴露明文证件 |
| ADM-003 | 异常打卡处理 | 创建打卡后后台标记无效 | 打卡状态变为 `invalid`，无权限返回 `40301` |
| ADM-004 | 成长和评价追溯 | 按用户和按局查询追溯接口 | 返回评价、信用、足迹、成就来源 |
| ADM-005 | 积分兑换审核 | 创建兑换订单后后台审核 | 审核状态流转，驳回时积分退回 |
| ADM-006 | 举报申诉冻结 | 创建举报后后台处理 | 可查看 IM/文件/评价证据，处理动作写入操作日志 |
| ADM-007 | 报表导出 | 创建导出任务并执行 runner | 任务从 `pending` 变为完成，下载地址只在完成后返回 |
| ADM-008 | AI 数据准备 | 调用验收夹具、读取快照、尝试导出 IM | `acceptanceReady=true`，默认导出 IM 返回 `40361` |
| ADM-009 | 交付测试留档 | 登记交付文档、测试用例和测试执行 | 列表可查，测试执行关联已存在测试用例 |
| ADM-010 | 权限拦截 | 每个后台页面用缺权限账号访问 | 接口返回 `40301`，页面不展示无权限操作 |

## 自动化覆盖索引

- `services/go-api/internal/appapi/server_test.go`
- `services/go-api/internal/adminauth/service_test.go`
- `services/go-api/internal/delivery/test_run_service_test.go`
- `services/go-api/internal/redemption/redemption_service_test.go`
- `services/go-api/internal/connections/connection_service_test.go`
- `services/go-api/internal/profiles/expert_skill_service_test.go`

## Phase one admin game creation coverage

- `POST /api/admin/games` requires `game:create_admin`.
- Admin can create `gameType=free`, `gameType=standard`, `gameType=public_welfare`, `gameType=aa`, `gameType=crowdfund`, `gameType=deposit` and `gameType=condition` for phase-one online verification.
- Admin-created games return `gameSource=admin` and `status=recruiting`, so mini program users can see and join them without real payment.
- Mini program `POST /api/app/games` still only allows `gameType=free`; non-free types must return validation error.

## 2026-06-16 IM room management addendum

| Case | Endpoint | Permission | Expected result |
| --- | --- | --- | --- |
| ADM-IM-001 | `GET /api/admin/im/rooms?status=active` | `im:room:read` | Returns active room summaries with `memberIds`, `messageCount`, and `fileMessageCount`. |
| ADM-IM-002 | `GET /api/admin/im/rooms/{roomId}` | `im:room:read` | Returns room detail, all messages, file messages, and message counters. |
| ADM-IM-003 | `GET /api/admin/im/rooms/{roomId}/dispute-messages` | `im:message:view_dispute` | Returns dispute evidence messages and writes an operation log. |

Automated coverage:

- `TestAdminIMRoomsListAndDetail`
- `TestReportEvidenceAndAdminOperationLogs`

## 2026-06-16 production readiness addendum

| Case | Endpoint | Permission | Expected result |
| --- | --- | --- | --- |
| ADM-SYS-001 | `GET /api/admin/system/readiness` | `system_config:read` | Returns redacted readiness status for database, JWT, WeChat, SMS, FaceID, storage, OpenIM and funds-service config. |
| ADM-SYS-002 | `GET /api/admin/system/readiness` with operator account | `system_config:read` | Returns `40301`; ordinary operation account cannot inspect system config readiness. |

Automated coverage:

- `TestAdminSystemReadiness`
- `TestReadinessReportMarksMissingProductionDependencies`
- `TestReadinessReportAllowsCompleteProductionConfig`

## 2026-06-16 invite admin addendum

| Case | Endpoint | Permission | Expected result |
| --- | --- | --- | --- |
| ADM-INV-001 | `POST /api/admin/invite-codes` | `invite_code:manage` | Admin creates poster, qrcode, or link invite code for a user. |
| ADM-INV-002 | `GET /api/admin/invite-codes?entryType=qrcode&status=active` | `invite_code:read` | Returns invite code list with usage and binding summary. |
| ADM-INV-003 | `GET /api/admin/invite-relations?inviterUserId={userId}` | `invite_code:read` | Returns invite relation list for operation review. |
| ADM-INV-004 | `POST /api/admin/invite-codes/{code}/disable` | `invite_code:manage` | Disabled code can no longer pass app invite precheck. |

Automated coverage:

- `TestAdminInviteCodeManagementHTTP`

## 2026-06-16 IM admin action addendum

| Case | Endpoint | Permission | Expected result |
| --- | --- | --- | --- |
| ADM-IM-004 | `POST /api/admin/im/rooms/{roomId}/retry-create` | `im:room:retry_create` | Retries OpenIM room sync when OpenIM is configured; local mode returns the current room without external network. |
| ADM-IM-005 | `POST /api/admin/im/messages/{messageId}/hide` | `im:message:hide` | Message status becomes `hidden`; app room history filters it while admin room detail keeps it as evidence. |
| ADM-IM-006 | `POST /api/admin/im/rooms/{roomId}/archive` | `im:room:archive` | Room status becomes `archived`; app users can no longer send messages to that room. |

Automated coverage:

- `TestAdminIMRoomArchiveRetryAndHideMessageHTTP`
