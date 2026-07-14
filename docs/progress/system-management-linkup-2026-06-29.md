# 2026-06-29 系统管理与邀请链路推进

## 本轮已完成

- 组局邀请聚合接口已接入后端真实路由：
  - `GET /api/app/game-invites/player-config`
  - `GET /api/app/game-invites/recent-players`
  - `GET /api/app/game-invites/players`
  - `GET /api/app/game-invites/replay-context`
  - `GET /api/app/game-invites/system-recommendations`
  - `POST /api/app/game-invites/replay`
  - `GET /api/app/game-invites/guide-progress`
  - `GET /api/app/game-invites/guide-cancel-detail`
- 系统管理“我的资料”已补齐后端读写闭环：
  - `GET /api/app/profile/system-management/profile-info`
  - `PUT /api/app/profile/system-management/profile-info`
  - 保存姓名时同步更新 `users.nickname`，页面再次进入可以读取上次保存的 `personalInfo / enterpriseInfo / certifications`。
- 系统管理“技能配置”已补齐后端读写闭环：
  - `GET /api/app/profile/system-management/skill-config`
  - `PUT /api/app/profile/system-management/skill-config`
  - 默认配置会从行家技能档案、用户局统计、信用档案聚合生成；保存后按页面结构返回 `roleSummary / skillSlots / skillGroups / unlockSuggestion`。
- 小程序前端已新增 `getSystemProfileInfo`，`pages/profile/system-management/profile-info/index` 进入页面时会从后端读取资料并刷新已有展示行。
- 删除未使用的 `AllInvitations` 方法，避免扩大不必要接口面。
- 消息中心已从纯列表推进到页面聚合和动作闭环：
  - `GET /api/app/notifications` 继续返回 `items`，同时返回小程序页面直接消费的 `quickActions / tabs / sections`。
  - 创建组局邀请时自动给被邀请人生成 `notifyType=game_invitation`、`bizType=game_invitation` 的通知。
  - `POST /api/app/notifications/{id}/actions` 已支持邀请通知的 `accept / reject` 等动作，并复用已有邀请响应服务。
  - 小程序 `pages/message/index` 的消息按钮已接入 `messageService.handleNotificationAction`。
- 系统管理“反馈”和“屏蔽设置”已补齐真实读写闭环：
  - `GET /api/app/profile/system-management/feedback`
  - `POST /api/app/profile/system-management/feedback`
  - `GET /api/app/profile/system-management/feedback-records`
  - `GET /api/app/profile/system-management/block-settings`
  - `PUT /api/app/profile/system-management/block-settings`
  - 小程序反馈页、反馈记录页、屏蔽设置页、关键词页、分场景配置页、白名单页均已接到同一份后端配置。
- 消息详情页已补齐真实详情接口：
  - `GET /api/app/messages/trade-warning`
  - `GET /api/app/messages/system-notification`
  - 小程序交易预警页、系统通知页已从详情专用接口读取，不再借用消息列表结构。
- 组局进度详情页已把关键动作接到现有真实页面：
  - 地图查看跳转到 `pages/map/index`
  - 取消组局跳转到 `pages/game/guide-cancel/index`
  - 提醒行家跳转到现成的 `pages/game/guide-chat/index`

## 验证结果

- `go test -count=1 ./...`
- `go test -count=1 ./...`
- `node --check` 遍历 `miniprogram-client/**/*.js`
- `rg -n "ponytail:" services/go-api miniprogram-client` 未发现新增债务标记

## 当前进度估算

- 小程序后端业务功能：约 `99.9%`。主要剩余不是本地业务代码，而是真实微信、地图、对象存储、实名/人脸、OpenIM 等外部环境联调。
- 小程序前后端全链路对接：约 `92%`。本轮从约 `91%` 推进到约 `92%`，核心提升来自组局进度详情页动作闭环。
- 后台后端业务功能：约 `99.2%`。核心管理、权限、审计、邀请码、组局、审核、分润、导出等已经具备，剩余集中在浏览器 E2E、线上环境联调和少数页面级细分接口。

## 下一步优先级

1. 继续补消息中心非邀请动作，包括查看路线、联系发起人、立即处理、举报/申诉详情跳转。
2. 补首页动态地球/关系网所需的可视化聚合数据，包括节点、边、热力、在线状态、附近局摘要。
3. 补系统管理剩余细分页的后端聚合接口：举报、协议签署、信用申诉。
