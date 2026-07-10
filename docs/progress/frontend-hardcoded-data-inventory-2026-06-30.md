# 前端硬编码数据盘点 - 2026-06-30

## 判定规则

- 应进入数据库或后端接口：业务标题、用户/行家/玩家、金额、局类型、活动类型、订单/局状态、推荐分类、评价模板、协议/规则文案、榜单/成就/商品等可运营内容。
- 可留在前端：尺寸、颜色、动画、文件扩展名、角色样式映射、按钮图标映射、页面路由映射、空数组初始态。

## 已处理

| 模块 | 前端位置 | 后端/数据库落点 |
| --- | --- | --- |
| 我的局卡片 | `pages/profile/service-center/my-games/index.js` | `GET /api/app/games/player/manage`、`GET /api/app/games/my/manage`、`GET /api/app/games/favorites/my` |
| 创建局类型 | `pages/game/create/index.js` | `system_configs.game.category_config` + `GET /api/app/games/category-config` |
| 创建局收益模板 | `pages/game/create/index.js` | 后台收益模板/收益配置接口 + `GET /api/app/games/profit-templates` |
| 邀请默认标题/详情/预算/行家/活动类型/引荐语 | `pages/game/invite/index.js` | `system_configs.game.invite_config` + `GET /api/app/game-invites/player-config` |
| 系统推荐分类 | `pages/game/system-recommend/index.js` | `GET /api/app/game-invites/system-recommendations` |

## 仍需迁移

| 优先级 | 前端位置 | 当前硬编码 | 建议落点 |
| --- | --- | --- | --- |
| P0 | `pages/game/confirm/index.js` | `DEFAULT_REPLAY_CONTEXT`、`QUICK_MESSAGES` | `GET /api/app/game-invites/replay-context` + `system_configs.game.replay_quick_messages` |
| P0 | `pages/game/payment/index.js` | 默认支付金额、拆分说明 | 订单/支付预创建接口返回金额与拆分 |
| P0 | `pages/game/apply/index.js` | 协议入口、申请提示与部分待接入动作 | 后端协议/规则接口或系统配置 |
| P1 | `pages/game/manage/index.js`、`pages/game/player-manage/index.js` | 管理/玩家订单默认摘要与时间线兜底 | 管理接口返回 summary、orders、timeline，失败只空态 |
| P1 | `pages/game/share/index.js` | `DEFAULT_GAME_INFO` 分享卡片兜底 | 游戏详情接口返回 shareCard |
| P1 | `pages/game/success-expert/index.js`、`pages/game/success-guide/index.js` | 成功页默认详情 | 成功详情接口返回完整 DTO，失败只空态 |
| P1 | `pages/game/audit-detail/index.js` | 玩家、领路人、结算示例 | 审核详情接口返回真实申请/结算数据 |
| P1 | `pages/profile/asset-center/mall/index.js` | 默认商品 | 积分商城接口返回商品列表；测试商品写 seed |
| P1 | `pages/profile/footprint/achievements/index.js` | 成就与锁定项 | 成就配置表或 `system_configs.profile.achievements` |
| P2 | `pages/profile/settings/index.js`、`pages/profile/system-management/agreement-detail/index.js` | 设置分组、协议兜底 | 后台协议/系统配置接口 |

## 本轮新增数据库测试数据

- 新增 `db/seeds/system_configs.sql`
  - `game.invite_config`
  - 用于邀请页默认标题、详情、预算、活动类型、奖励比例和引荐语模板。

## 执行原则

1. 删除前端业务硬编码时，必须同时确认后端接口和数据库/seed 落点。
2. 开发 mock 可以保留，但只能作为无后端开发模式，不作为真实联调数据源。
3. 线上/联调环境前端必须关闭 mock，所有业务展示来自接口。
