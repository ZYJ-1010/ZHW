# 2026-06-29 首页动态地球与关系网对接进度

## 本轮完成

- `GET /api/app/home` 已从基础 `games / stats` 扩展为小程序首页可直接消费的聚合数据：
  - `hero`
  - `playerSummary`
  - `nearbySummary`
  - `nearbySection`
  - `nearbyGames`
  - `recommendedGames`
  - `friendSection`
  - `friendGames`
  - `rankingSection`
  - `rankingBoards`
  - `achievementSection`
  - `achievements`
  - `metaverseEntry`
  - `earth`
  - `visualization`
- 首页动态地球数据已包含：
  - 当前用户节点
  - 组局节点
  - 关系人节点
  - 关系边
  - 城市热力点
  - 附近局数量
- 附近局卡片已按用户当前定位计算距离，并返回 `distanceText / memberText / joinedText / actionText` 等前端已兼容字段。
- 好友在玩数据已基于关系链路和组局成员生成，前端 `friendGames` 可直接读取。
- `GET /api/app/connections/my` 已保持原 `items` 列表兼容，同时新增：
  - `onlineText`
  - `header`
  - `tabs`
  - `activeTab`
  - `network.nodes`
  - `network.edges`
  - `network.stats`
- 新增 HTTP 测试 `TestAppHomeAndConnectionNetworkPayloadHTTP`，覆盖：
  - 保存定位
  - 创建并审核组局
  - 建立用户关系
  - 首页返回动态附近局、地球节点、热力点、关系网
  - 关系网接口返回原连接列表与 nodes/edges/stats

## 验证结果

- `go test -count=1 ./...` 通过。
- `rg -n "buildAppHomePayload|homeEarth|connectionNetworkPayload|TestAppHomeAndConnectionNetworkPayloadHTTP|nearbySummary|visualization" services/go-api/internal/appapi` 已确认新增代码和测试在位。
- `rg -n "ponytail:" services miniprogram-client docs -g "!*node_modules*" -g "!*dist*" -g "!*.map"` 未发现本轮新增债务标记。

## 当前进度估算

- 小程序后端业务功能：约 `99.9%`。剩余主要是线上第三方环境联调，不是本地业务代码缺口。
- 小程序前后端全链路对接：约 `90%`。本轮从约 `87%` 推进到约 `90%`，核心提升来自首页动态地球、附近局、好友局、关系网聚合数据闭环。
- 后台后端业务功能：约 `99.2%`。仍以浏览器 E2E、线上联调和少量页面细分接口核对为主。

## 下一步建议

1. 补消息中心非邀请动作链路：路线、联系发起人、立即处理、举报/申诉详情跳转。
2. 补系统管理剩余细分页接口：屏蔽设置、反馈、举报、协议签署。
3. 对后台 Web 页面做一次接口字段核对，确认列表、详情、审核、导出和审计记录字段与后端数据库命名一致。
