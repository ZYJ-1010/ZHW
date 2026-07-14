# 2026-06-29 全链路对接与未实现清单

## 本轮已打通

### 小程序前端 API 映射

- 会员状态：`/api/app/memberships/me` 已改为后端真实 `/api/app/membership/my`。
- 成长档案：`/api/app/growth/me` 已改为 `/api/app/growth/my`。
- 积分商城：`/api/app/profile/points/mall` 已改为 `/api/app/redemption/items`。
- 积分兑换：`/api/app/profile/points/mall/exchange` 已改为 `/api/app/redemption/orders`，并把前端商品 id 映射为后端 `itemId`。
- 积分订单：`/api/app/profile/points/orders` 已改为 `/api/app/redemption/orders/my`。
- 消息中心：`/api/app/messages/center` 已改为统一通知 `/api/app/notifications`。
- 交易预警/系统通知详情入口：先复用 `/api/app/notifications`，按 `type` 透传筛选条件。
- 关系网首页：`/api/app/relations/network-home` 已改为 `/api/app/connections/my`。
- 支付占位：`/api/app/game-payments/wechat` 已改为一期真实占位接口 `/api/app/payment/precreate-placeholder`。
- 邀请响应：前端服务层已兼容把 `accept/approve/yes` 动作映射为后端需要的 `accept` 字段。

### 后端新增薄聚合接口

- `GET /api/app/home`
  - 返回公开局列表前 10 条、用户局统计、积分概要、成长档案、未读通知数。
  - 用于小程序首页先跑通真实后端数据，不再依赖纯 mock。
- `GET /api/app/newbie-tasks`
  - 按现有身份、角色申请、局参与、局完成、评价状态即时计算新手任务。
  - 一期不新增任务领取/奖励表，仅保证页面可从后端拿到真实任务状态。
- `GET /api/app/profile/home`
  - 复用 `GET /api/app/users/me/summary`，作为个人中心首页聚合入口。
- `GET /api/app/games/my/manage`
  - 返回当前用户创建或作为主行家管理的局。
- `GET /api/app/games/player/manage`
  - 返回当前用户以玩家成员身份参与的局。
- `GET /api/app/games/profit-templates`
  - 返回后端已有分润模板，供小程序创建局/费用预览页读取。

## 按文档和前端代码对比后的剩余缺口

### 1. 组局邀请推荐链路未完整实现

前端仍在调用：

- `GET /api/app/game-invites/player-config`
- `GET /api/app/game-invites/recent-players`
- `GET /api/app/game-invites/players`
- `GET /api/app/game-invites/replay-context`
- `GET /api/app/game-invites/system-recommendations`
- `POST /api/app/game-invites/replay`
- `GET /api/app/game-invites/guide-progress`
- `GET /api/app/game-invites/guide-cancel-detail`

当前后端已具备“主行家邀请”和“邀请响应”基础能力：

- `POST /api/app/games/{gameId}/guide-invitations`
- `POST /api/app/game-invitations/{invitationId}/respond`

但还没有“推荐玩家列表、最近联系人、复玩上下文、系统推荐、取消详情、邀请进度聚合”这些页面级业务接口。该模块涉及推荐/复玩/进度聚合，不建议简单改成现有邀请接口，需要按字段重新定接口。

### 2. 个人中心系统管理页仍是前端候选模型

前端仍在调用：

- `PUT /api/app/profile/system-management/profile-info`
- `GET /api/app/profile/system-management/skill-config`
- `PUT /api/app/profile/system-management/skill-config`

后端已有较接近的能力：

- `PUT /api/app/users/me/profile`
- `GET/POST/PUT /api/app/experts/me/skills`
- `GET/POST/PUT /api/app/guides/me/resources`

但前端页面保存的是 `personalInfo`、`enterpriseInfo`、`certifications`、`skillSlots`、`skillGroups`、`unlockSuggestion` 等静态页结构；后端现有模型是用户基础资料、行家技能树、领路人资源画像。字段语义不一致，不能直接硬接，否则会丢资料、企业信息和技能槽位状态。

### 3. 积分订单物流/取消/详情未完整实现

已打通：

- 商城列表
- 兑换下单
- 我的兑换订单

仍缺：

- `GET /api/app/profile/points/orders/{orderId}/logistics`
- 订单详情页接口
- 取消兑换订单接口

后端当前兑换订单状态流转主要由后台审核、驳回、发放完成；小程序侧物流和取消属于订单履约细节，数据库和状态机还需要补字段。

### 4. 消息中心操作按钮未完整实现

已打通：

- 消息列表读取 `/api/app/notifications`
- 标记已读 `/api/app/notifications/{id}/read`

仍缺：

- 消息卡片上的 `确认参加`、`婉拒`、`立即处理`、`查看路线`、`联系发起人` 等动作和业务对象的精确映射。
- 消息详情富文本/块结构接口。

### 5. 首页/地球/关系网动态数据仍是基础版

已打通：

- `/api/app/home` 返回公开局、用户统计、积分、成长、未读数。
- `/api/app/connections/my` 返回真实人脉关系。
- `/api/app/games/nearby`、地图代理、定位记录已在后端存在。

仍缺：

- 前端“动态地球”和“关系网”需要的可视化节点、边、实时在线、地理热力、动画摘要字段。
- 当前后端只提供业务数据基础，不提供完整可视化布局数据。

### 6. 小程序真机全链路仍需外部环境验证

本地代码已通过语法和 Go 测试，但以下不是本地单元测试能替代的：

- 微信真实 `code2Session`。
- 微信订阅消息真实发送。
- 真实对象存储直传和下载域名。
- 真实人脸核身/实名网关。
- 真实地图 WebService Key 和线上域名。
- 真机小程序编译、页面跳转、分享卡片、二维码入口。

## 当前进度判断

- 一期小程序后端业务功能：约 `99.9%`。主体业务已完整，剩余主要是页面候选接口、真实三方和线上环境联调。
- 小程序前端与后端全链路对接：约 `72%`。登录、邀请、实名、组局、IM、会员、成长、积分、通知、首页基础聚合已经对上；推荐邀请、系统管理、积分履约、消息动作仍未完成。
- 后台后端业务功能：约 `99.2%`。后台核心管理、权限、日志、邀请码、审核、分润、导出等已基本齐；剩余主要是真实浏览器 E2E 和线上环境联调。
