# Modao Observation 2026-06-13

## Access Status

- Prototype: `小程序开发（对外）`
- Current URL as of 2026-06-16: `https://modao.cc/proto/BJBH1Uptgioo7R81pGx2/sharing?view_mode=read_only&screen=rbpVMU6bjYrmWMsnN`
- Current password as of 2026-06-16: `vhni5m`
- Historical URL observed on 2026-06-13: `https://modao.cc/proto/BJBH1Uptgioo7R81pGx2/sharing?view_mode=read_only&screen=rbpVFDJKkOR1Pi9Cf`
- Historical password used earlier: `9o04o7` (expired)
- 2026-06-13 11:05 re-entered successfully after cache/session recovery.
- Current canvas observed: `组局分布地球网`
- Modao side tree shows: `画布（50）`

## Page Tree Observed

### 邀请注册

- 新用户邀请注册登录
- 已注册用户登录
- 找回密码

### 首页

- 玩家首页
- 行家首页
- 领路人首页
- 无行家权限提示页
- 无领路人权限提示页
- 权益对比页
- 申请行家操作页
- 申请领路人操作页
- 审核进度页
- 审核通过页
- 审核驳回页
- 申请状态提示页
- 申请进度查看页
- 首页调色说明页

### 组局主流程

- 局前大厅
- 玩家自申请入局
- 领路人发起引荐（邀请）页
- 玩家被邀约确认页
- 行家审核列表页
- 行家审核详情页
- 领路人接收页
- 组局支付页
- 组局成功页
- 组局取消页
- 交付操作页
- 交付确认页
- 评价页面

### 发起组局

- 发起组局

### 地图首页

- 组局分布地球网
- 组局盲盒
- 城市图鉴页
- 足迹热力图页
- 实景打卡
- 好友点亮城市页
- 我的城市故事页

### 关系与消息

- 关系网
- 消息

### 元宇宙

- 元宇宙（预留）
- 元宇宙内页及管理中心

### 我的

- 我的
- 会员中心
- 母版
- 成就页

## Current Canvas: 组局分布地球网

### Visible UI Content

- Top status shows `3999人在线`.
- Map filters or tabs:
  - 在线（23）
  - 离线（8）
  - 全部组局
  - 附近组局
  - 组局路线
  - 热力图
  - 好友分布
  - 解锁图鉴
- Nearby play card:
  - `鱼尾狮夜景打卡点`
  - `距离 420m · 已有 123 条城市故事 · 成就：夜游新手`
  - Actions: `立即打卡`, `看故事`
- Route card:
  - `3点路线盲盒：港湾微风版`
  - `预计 90 分钟 · 适合 2-4 人 · 可解锁 1 枚路线勋章`
  - Actions: `开始路线`, `加入组队`
- Search:
  - `搜局、搜人、搜地块...`
  - Example text: `"金牌领路人: 阿亮"`
- Floating profile card:
  - `剧本杀小王`
  - `距你 280m · 2分钟前活跃`
  - `行家 Lv.3`
  - `评分 4.9`
  - `共同好友 3`
  - Action: `发起连接`
- Bottom tabs:
  - 我的
  - 元宇宙
  - 地图
  - 消息
  - 首页
  - 智能寻局

### Annotations

1. 点击或者鼠标移动到组局点，自动弹出组局基本信息。
2. 参考灵敢足迹 app，不同组局点设置闪烁呼吸点。通过双指展开或者收缩展开地图，收缩成地球，地球显示已组局或挂牌区域的颜色分布。
3. 交互说明：
   - 入口：底部地图 Tab、首页/元宇宙、我的/快捷入局的快捷入口。
   - 核心动作：看点位、看路线、发起组队、进入图鉴、打卡。
   - 状态：地图点位区分普通打卡点、路线点、商家点、成就点。
   - 支持热力与好友分布视图。

## Development Notes

- Current implemented `pages/map/index` is only a placeholder/simple entry page. It should later align with this Modao canvas.
- First implementation can keep list-first layout, then reserve advanced map/earth/heatmap behavior.
- Need add data fields later for point type, distance, online status, route points, achievement, story count, and friend relation.

## Canvas: 邀请注册

### Sub Screens Observed

- 登录主页
- 验证码登录
- 账号密码登录
- 微信授权
- 进入启动页面

### Login Content

- Brand wording includes: `欢迎进入真好玩~`
- Login options:
  - 手机号登录
  - 账号密码登录
  - 微信登录
  - 验证码登录
- Agreement:
  - `我已阅读并同意`
  - `用户协议`
  - `隐私协议`
- WeChat auth page:
  - `微信登录`
  - `真好玩 申请获取你的微信头像、昵称`
  - `用于完善个人资料`
  - Actions: `允许`, `拒绝`
- Startup/home transition page includes:
  - `3999人在线`
  - `GO`
  - `ZHEN HAO WAN`

### Development Notes

- Login implementation should make WeChat authorization the central visual focus.
- Secondary options can be lighter: verification code login, account password login, invitation code hint.
- Agreement checkbox must gate the login button.

## Canvas: 玩家首页

### Visible UI Content

- Header:
  - `HELLO, 玩家!`
  - `2026.03.30 | 开启你的今日副本`
  - `3999人在线`
- Main game card:
  - `AI赋能系统搭建交流局`
  - Tags/status: `任务局`, `￥0/人`
  - Location/time examples: `黄浦区 · 8.2km · 3/8人`, `2026年5月1日 14:00--16:00`
  - Actions: `加入`, `分享`, `关注`, `引荐`, `打招呼`
- Nearby section:
  - `附近正在发生`
  - Filters: `全部`, `附近`
  - Example game: `苏州河“记忆碎片”采集`
- Ranking/growth:
  - `本周玩霸榜`
  - Roles: `玩家`, `行家`, `领路人`
  - Ranking entries include XP values.
- My level section:
  - `玩家 Lv.5`
  - `探险家`
  - `Alex Chen`
  - `580/1000 XP`
  - Stats: `参与局数`, `本月MVP`, `参与率`
  - Prompt: `距离下一等级还需 420 经验值`
- Quick actions:
  - `发起组局`
  - `局前大厅`
- Friends/game section:
  - `朋友在玩`
  - `查看全部`
  - Example: `盲盒路线：3小时点亮天际线`
- Metaverse entry:
  - `共创数字街区｜全球联机互动`
  - `3D空间`
  - `NFT徽章`
  - `进入元宇宙`
- Bottom tabs:
  - 我的
  - 元宇宙
  - 地图
  - 消息
  - 首页

### Development Notes

- Current `pages/home/index` should later split content into role-aware variants.
- First production-friendly version can keep:
  - identity/status header
  - recommended game cards
  - nearby games
  - ranking/growth entry
  - quick actions
  - bottom navigation
- Full Modao visual has a game/metaverse style, but first implementation should remain simple until screenshots and dimensions are stable.

## Canvas: 我的

### Major Modules Observed

- 个人中心
- 我的局
- 我的资产
- 我的订单
- 积分中心
- 引荐记录
- 信用中心
- 资料设置
- 系统设置
- 评价管理

### 我的局

- Filters:
  - 全部
  - 进行中（2）
  - 已完成（5）
  - 超时（0）
  - 已取消（1）
- Example active record:
  - `ACT-20260320-001`
  - `产品架构咨询`
  - `行家：张专家`
  - `领路人：王引荐`
  - `¥800`
  - `已托管`
  - `预计交付：03-25 14:00`
- Timeline:
  - 需求确认
  - 组局成功
  - 服务进行中
- Actions:
  - 联系业务主
  - 查看群聊
  - 等待完成确认
  - 评价
  - 再次购买

### 我的资产

- `总资产 (元)`
- `¥16,580.00`
- `总成交额`
- `可提现`
- `待结算`
- Actions:
  - 提现
  - 充值
  - 余额明细
  - 银行卡
- Orders:
  - 待付款
  - 进行中
  - 已完成
  - 退款/售后
  - 待评价

### 积分中心

- `可用积分`
- `累计积分`
- `消费积分`
- `过期积分`
- Earn points:
  - 产生交易
  - 邀请玩家
  - 邀请升级高级会员
  - 邀请升级尊享会员
- Note: 积分有效期为 12 个月，过期自动清零。

### 引荐记录

- Tabs:
  - 我引荐的
  - 我发起的
- Status:
  - 全部
  - 进行中（3）
  - 已完成（12）
  - 超时（1）
  - 已取消（2）
- Includes timeout warning.
- Example record shows expert, referrer, player, budget, reward, delivery status.

### 信用中心

- `信用分（默认100分）`
- Example score: `98`
- Status: `优秀`
- Includes:
  - 信用等级
  - 奖励中心
  - 惩罚中心
  - 信用明细
- Note: 信用分低于 80 分限制部分功能，低于 60 分暂停服务资格。

### 资料设置

- Personal info:
  - 头像
  - 姓名
  - 联系方式
  - 可见性设置
  - 兴趣爱好
- Enterprise info:
  - 公司名称
  - 职务
  - 主营业务
  - 可提供资源
  - 公开注册业务信息
- Certification:
  - 个人身份认证
  - 企业认证

### Development Notes

- Current `pages/profile/index` should first recover from blank rendering.
- First useful version should include:
  - user/profile card
  - role and certification status
  - my games summary
  - assets/order entry
  - points/credit entry
  - invitation/referral entry
  - settings entry
- The full Modao canvas is too broad for one small page; split detailed modules into later child pages.
