# 小程序端联调用例包

## 使用口径

- 本文件只列小程序前端需要直接调用或配合验证的后端接口。
- `验收必测` 表示一期上线验收必须走通；`可选入口` 表示页面可延后，但接口已可联调。
- 所有受保护接口都使用 `Authorization: Bearer <token>`。

## 验收必测

| 编号 | 场景 | 接口 | 关键断言 |
| --- | --- | --- | --- |
| APP-001 | 微信登录和邀请码注册 | `POST /api/app/auth/wechat-login` | 新用户无邀请码返回 `40321`，有效邀请码返回预授权登录态 |
| APP-002 | 强实名闭环 | `POST /api/app/identity/phone/bind`、`POST /api/app/sms/send-code`、`POST /api/app/sms/verify-code`、`POST /api/app/identity/phone/verify`、`POST /api/app/identity/faceid/detect-auth`、`POST /api/app/identity/faceid/callback`、`GET /api/app/identity/status` | 最终 `status=verified`，未实名时核心业务返回 `40341` |
| APP-003 | 创建免费局 | `POST /api/app/games` | 只允许 `gameType=free`，人数 5-8，超每日 3 局返回 `42921` |
| APP-004 | LBS 局列表 | `POST /api/app/locations/current`、`POST /api/app/locations/manual`、`GET /api/app/games/city`、`GET /api/app/games/nearby` | 当前定位和手动定位均可驱动附近局查询 |
| APP-005 | 入局和手动开始 | `POST /api/app/games/{gameId}/applications`、`POST /api/app/games/applications/{applicationId}/review`、`POST /api/app/games/{gameId}/manual-start` | 申请审批后成为成员，局状态可进入 `in_progress` |
| APP-006 | IM 房间和消息 | `GET /api/app/games/{gameId}/chat-session`、`POST /api/app/games/{gameId}/chat/messages`、`GET /api/app/games/{gameId}/chat/messages` | 非成员不能访问；敏感词返回 `45101` |
| APP-007 | 文件上传和 IM 文件消息 | `POST /api/app/files/upload-token`、`GET /api/app/files/{fileId}/download-url`、`POST /api/app/games/{gameId}/chat/messages` | 成员可上传并发送 `messageType=file`，非成员不能下载 |
| APP-008 | 服务确认和评价 | `POST /api/app/games/{gameId}/service-confirm`、`POST /api/app/games/{gameId}/service-confirm-items`、`POST /api/app/reviews` | 双方确认后进入待评价，重复评价被拦截 |
| APP-009 | 收藏局 | `POST /api/app/games/{gameId}/favorite`、`DELETE /api/app/games/{gameId}/favorite`、`GET /api/app/games/favorites/my` | 收藏幂等，取消后我的收藏不再返回 |
| APP-010 | 进度里程碑和打卡 | `POST /api/app/games/{gameId}/milestones`、`POST /api/app/games/{gameId}/checkins`、`GET /api/app/games/{gameId}/checkins` | 非成员不能打卡，非法状态/类型返回 `42241`/`42242` |
| APP-011 | 复盘和续局草稿 | `POST /api/app/games/{gameId}/retrospectives`、`POST /api/app/games/{gameId}/continue` | 复盘不能重复提交，续局只生成草稿 |
| APP-012 | 积分和兑换 | `GET /api/app/points/summary`、`GET /api/app/points/logs`、`GET /api/app/redemption/items`、`POST /api/app/redemption/orders`、`GET /api/app/redemption/orders/my` | 积分不足返回 `40943`，库存不足返回 `40942` |
| APP-013 | 通知和举报 | `GET /api/app/notifications`、`POST /api/app/notifications/{notificationId}/read`、`POST /api/app/reports`、`GET /api/app/reports/my` | 处理结果可进入通知，用户可查看自己的举报 |

## 可选入口

| 编号 | 场景 | 接口 | 说明 |
| --- | --- | --- | --- |
| APP-O01 | 行家技能树 | `GET /api/app/experts/me/skills`、`PUT /api/app/experts/me/skills` | 需要后端已授予行家角色，否则返回无权限 |
| APP-O02 | 领路人资源画像 | `GET /api/app/guides/me/resources`、`PUT /api/app/guides/me/resources` | 需要后端已授予领路人角色，否则返回无权限 |
| APP-O03 | 人脉关系 | `GET /api/app/connections/my`、`POST /api/app/connections/{connectionId}/follow-logs` | 同局或领路人撮合后可查看和登记跟进；`/follow-up` 保留为兼容别名 |
| APP-O04 | 会员和团队 | `GET /api/app/member-reports/me`、`GET /api/app/teams/my`、`GET /api/app/teams/my/revenue-summary` | 一期可作为会员权益预留入口联调 |

## 自动化覆盖索引

- `services/go-api/internal/appapi/server_test.go`
- `services/go-api/internal/games/service_test.go`
- `services/go-api/internal/identity/service_test.go`
- `services/go-api/internal/connections/connection_service_test.go`
- `services/go-api/internal/profiles/expert_skill_service_test.go`
- `services/go-api/internal/points/points_service_test.go`
- `services/go-api/internal/redemption/redemption_service_test.go`
# 小程序端联调用例包

## 最新一期联调口径

- 邀请登录：小程序只能通过邀请入口进入；poster 小程序分享卡片、qrcode 扫码、link 微信聊天链接都必须携带唯一 `inviteCode`。前端先调 `POST /api/app/invites/precheck`，再按 `authPageMode=register/login` 跳注册页或登录页；已注册微信通过新的唯一入口码进入时，后端会绑定该入口码并返回登录模式。
- 玩家实名：玩家创建免费局、申请入局不强制实名；接口只要求已登录并通过有效邀请码进入。
- 行家/领路人实名：行家、领路人、被邀请为主行家的用户必须完成实名；未实名接受行家邀请或操作角色能力返回 `40341`。
- 组局人数：一期每局最少 5 人、最多 8 人；未满 5 人不能手动开始。
- 主行家：玩家可邀请一个行家作为主行家；主行家由首个“邀请接受 + 审核通过”的行家产生，局详情返回 `mainGuideUserId`，成员列表 role 返回 `main_guide`。
- 局进度：创建者或主行家可手动开始、维护里程碑和进度；普通成员只能按成员权限申请、聊天、打卡、确认服务和评价。
