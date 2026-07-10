# 2026-06-29 全链路补齐追加记录

## 本轮已补齐

- 积分订单列表接口已补齐小程序订单页契约：
  - `GET /api/app/redemption/orders/my` 现在同时返回 `items` 和 `orders`。
  - 返回 `tabs`、`statusKey`、`statusText`、`statusTone`、`actions`，前端订单页可以直接渲染状态 tab 和“查看物流”动作。
  - 支持 `status` 查询筛选，并兼容旧前端的 `pending_ship`、`shipping`、`completed` 状态别名。
- 积分订单物流接口已落地：
  - `GET /api/app/profile/points/orders/{orderId}/logistics`
  - 后端会校验订单属于当前用户，再返回发放轨迹、订单号、状态和平台发放信息。
- 登录注册口径继续收紧：
  - 小程序服务层和底层请求层均已禁用旧手机号登录、密码登录、找回密码入口。
  - 当前业务只保留“唯一邀请入口 + 微信登录 + 必要实名/手机号绑定”的主链路。
- 邀请响应和新手任务修正：
  - 小程序邀请响应已兼容 `确认`、`确认参加`、`接受`、`同意`、`加入` 等中文动作值。
  - 新手任务“完成评价”改为读取真实评价记录，不再用“完成局且没有待评价”误判。

## 审核和验证

- 使用 `miniprogram-development`、`web-development`、`karpathy-guidelines` 的流程口径做小程序/后端对接复核。
- 按 `ponytail-audit` 口径优先删除/禁用不符合现业务的旧登录入口，而不是继续维护一套无后端支持的账号密码登录。
- 已执行：
  - `Get-ChildItem -Recurse miniprogram-client -Filter *.js | ForEach-Object { node --check $_.FullName }`
  - `go test -count=1 ./internal/redemption ./internal/appapi`
  - `go test -count=1 ./...`

## 当前进度判断

- 当前前后端全链路对接进度由约 `72%` 提升到约 `76%`。
- 已打通积分商城、兑换订单、订单状态筛选、物流弹层的主要页面链路。
- 剩余较大的真实业务缺口仍是：
  - 组局邀请推荐聚合。
  - 个人中心系统管理字段模型。
  - 消息卡片动作和详情结构。
  - 首页地球/关系网可视化数据结构。
