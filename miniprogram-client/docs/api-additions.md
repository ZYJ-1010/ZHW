# API Additions

## 2026-07-07 局内协作 / 交付确认

- 模块：组局 / 局内协作
- 页面：`pages/game/collaboration/index`、`pages/im/room/index`
- 接口：`POST /api/app/games/{gameId}/collaboration/end`
- 用途：创建者在局内协作页结束本局，将局状态推进到 `pending_confirm`，并通知行家和玩家进入交付确认；全部成员通过 `POST /api/app/games/{gameId}/service-confirm` 确认后，再开启互评。
- 权限：仅 `creator_user_id` 可调用；局成员可通过 `GET /api/app/games/{gameId}/collaboration` 查看协作页。
- 状态：已在本地 Go API 增加；通知类型为 `service_confirm_remind`，处理入口跳转 `pages/game/delivery/index?gameId={gameId}`。

## 2026-07-07 局 IM 群聊

- 模块：组局 / 局 IM 群聊
- 页面：`pages/im/room/index`
- 接口：`GET /api/app/games/{gameId}/chat-room`
- 用途：建议在现有房间信息基础上补充局名称、在线人数、成员入口所需统计、当前用户在群内角色，以及成员资料摘要；前端顶部当前只能使用 `memberIds/currentUserId/openIMGroupId/status` 做基础展示。
- 权限：局外成员必须返回无权访问，不可进入也不可查看；局结束后建议返回 `status=readonly` 或 `gameStatus=ended`，前端据此禁用发送，仅保留历史查看。
- 状态：待后端确认字段与状态枚举。
- 接口：`GET /api/app/games/{gameId}/chat/messages`
- 用途：建议消息列表返回 `senderName`、`senderAvatar`、`senderRoleText`、`messageType`、`content`、`fileId`、`fileName`、`fileUrl`、`systemCard` 等字段；其中 `systemCard` 用于组局邀请、组局成功、成员加入、时间地点变更等局内系统卡片。
- 状态：待后端确认字段；当前前端只按已有 `senderUserId/messageType/content/fileId/status/createdAt` 渲染。
- 接口：`POST /api/app/games/{gameId}/chat/messages`
- 用途：在现有 `text/image/file` 基础上，后续需要确认 `voice/location/emoji/system_ack` 等消息类型的请求体、文件上传方式和校验规则。
- 权限：局外成员不可发送；局结束或房间只读后成员不可发送，后台可保留只读审计查看。
- 状态：待后端确认字段。
- 接口：`POST /api/app/files/upload-token`、`GET /api/app/files/{fileId}/download-url`
- 用途：局 IM 文件传输按“请求上传凭证 -> 上传文件 -> 落库 `fileId` -> 详情页和 IM 可查看”处理；后台管理端查看 IM 时需要能按 `fileId` 获取下载地址并下载文档。
- 状态：上传凭证接口已存在；IM 消息文件详情、后台下载入口和权限规则待联调确认。

## 2026-07-05 组局详情 / 报名审核

- 模块：组局详情 / 创建组局 / 报名审核
- 接口：`GET /api/app/games/{id}`
- 用途：详情页业务展示由后端返回 `detailDisplay`，包含标题、状态文案、地址、统计项、组局者、参与者、底部主按钮文案/禁用态/action/route、待审核报名数量等字段。
- 状态：已在本地 Go API 增加；前端存在旧接口兼容路径，但新详情优先消费 `detailDisplay`。

- 模块：报名审核
- 接口：`GET /api/app/game-applications/received?gameId={gameId}`
- 用途：组局者从某个组局详情进入报名审核时，只查看当前局的报名申请。
- 状态：已在本地 Go API 增加 `gameId` 过滤；前端 `pages/game/audit/index` 已带入该参数。

- 模块：创建组局 / 组局详情地址
- 接口：`GET /api/app/map/reverse-geocode`
- 用途：创建组局拿到当前位置经纬度后，由后端地图反查地址；若本地非生产环境地图服务不可用，后端临时写入测试地址，后续替换为真实反查结果。
- 状态：接口已存在；本次补齐创建和详情读取时的后端地址补齐逻辑。

## 2026-07-05 组局状态：两类未成局

- 模块：组局详情 / 首页组局列表 / 报名审核 / 开局
- 接口：`GET /api/app/games/{id}`、首页组局列表及所有读取 `game.status` 的组局接口
- 用途：由后台根据组局时间和人数刷新未开局状态，前端只消费后台状态和文案。
- 新增状态：`recruit_failed` 表示报名截止、开始时间或结束时间已过但人数未达最低要求；`start_expired` 表示人数已达最低要求但结束时间已过仍未手动开局。
- 状态：已在本地 Go API 增加状态刷新和详情/首页文案映射；前端和 Mock 已补充同名状态文案。

## 2026-07-05 领路人邀请角色

- 模块：组局详情 / 领路人邀请 / 报名审核
- 接口：`POST /api/app/games/{id}/guide-invitations`
- 用途：邀请时支持传 `role` 或 `targetRole`，当前支持 `member` / `expert`。后台将角色落到 `game_invitations.role`，被邀请人接受后同步写入 `game_applications.role`。
- 权限：邀请 `expert` 时，目标用户必须已通过行家角色审批；领路人可审核自己发出的邀请所产生的申请，创建者仍可审核全部申请。
- 状态：已在本地 Go API 增加字段、迁移和服务层封装；前端页面入口待具体交互接入。

## 2026-07-06 组局分享 / 领路人引荐

- 模块：组局分享
- 接口：`POST /api/app/invites/entries`
- 用途：带 `gameId` 创建分享入口时，只生成分享/注册来源码，返回路径改为 `pages/game/share/index?id={gameId}&inviteCode=...&entryType=...`；不创建 `game_invitations` 组局引荐记录。
- 状态：已在本地 Go API 调整；前端分享页会保存 `inviteCode` 上下文用于来源追踪。

- 模块：领路人引荐 / 发起邀请
- 接口：`POST /api/app/game-invites/replay`
- 用途：领路人发起引荐时接收 `sourceGameId`、`expertUserId`、`invitees[]`、`message`，后台写入 `game_invitations.player_user_id` 和 `game_invitations.expert_user_id`，并给玩家和行家分别生成 `game_invitation` 通知。
- 详情字段：新邀请提交时额外接收 `serviceType`、`serviceDuration`、`demandDetail`、`budgetAmountCent`、`expectedTime`，后台写入 `game_invitations`；免费局 `budgetAmountCent` 为 `0`，其他详情字段以前端提交和后台存储为准，不在详情页前端补业务兜底。
- 状态：已在本地 Go API 增加字段、迁移和通知发送；前端发起邀请页提交已带 `expertUserId` 和新邀请详情字段。

- 模块：领路人引荐进度 / 参与者展示
- 接口：`GET /api/app/game-invites/guide-progress`、`GET /api/app/games/{id}/members`
- 用途：进度页由后台返回玩家、行家、组局信息、进度步骤和提醒目标；参与者列表由后台返回角色展示字段，前端不再根据角色码自行映射展示文案。
- 详情展示：`GET /api/app/game-invites/guide-progress` 的 `detailDisplay.confirmRows` 按 `game_invitations` 中的新邀请详情字段返回服务类型、咨询时长、预算金额、预计时间；前端只渲染后台返回的非空行。
- 状态：已在本地 Go API 补齐展示字段；前端优先消费后台字段。
