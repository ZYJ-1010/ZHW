# Follow-ups

## 2026-07-02 邀请注册链路：邀请码跳转后显示 `UNDEFINED`

### 功能模块：邀请注册

- 证据截图：`docs/evidence/2026-07-02-invite-login-undefined.png`
- 证据现象：登录/注册页的邀请码输入框被红框标注，显示为 `UNDEFINED`，没有显示后台有效邀请码 `PLAYER_LINK`。
- 业务影响：用户从“别人发来的邀请链接”进入后，邀请码无法自动延续到注册登录页，邀请关系绑定链路被打断。
- 当前状态：已记录，暂不修复；后续集中测试邀请注册链路时处理。

### 页面：`pages/login/invite/index`

- 问题：从邀请落地页进入时，`PLAYER_LINK` 可以被后端识别为有效邀请码，但点击“继续注册”跳到登录页后，登录页的邀请码输入框显示 `UNDEFINED`。
- 复现路径：
  1. 微信开发者工具启动页配置为 `pages/login/invite/index`
  2. 启动参数配置为 `inviteCode=PLAYER_LINK&entryType=link`
  3. 进入邀请注册页，确认邀请码
  4. 点击“继续注册”
  5. 登录页邀请码变为 `UNDEFINED`
- 原因：邀请页保存和跳转时使用了 `this.data.invite.code`，但后端预检返回结构里真实邀请码在 `this.data.invite.inviteCode.code`。
- 影响：邀请页到登录页的自动带码失败；用户无法按“别人发来的邀请链接”完整走通注册链路。
- 建议修复：
  - 在 `services/invite.js` 里把后端返回值统一归一化，让页面拿到的 `invite` 直接包含 `code`。
  - 或在 `pages/login/invite/index.js` 的 `continueToLogin` 中改为从 `this.data.invite.inviteCode.code` 取值，并保留 `entryType`。

### 功能模块：登录/注册

- 证据截图：`docs/evidence/2026-07-02-invite-login-undefined.png`
- 证据现象：登录/注册页输入框展示 `UNDEFINED`，说明页面接收了无效的邀请码参数并直接渲染。
- 业务影响：用户可能以错误邀请码继续点击“登录/注册”，导致注册失败或邀请关系无法绑定。
- 当前状态：已记录，暂不修复；后续测试登录注册链路时验证参数兜底。

### 页面：`pages/login/index`

- 问题：登录页只接收跳转参数里的 `inviteCode`，当上游传入 `undefined` 字符串时会直接显示 `UNDEFINED`。
- 影响：用户可能继续以错误邀请码尝试登录/注册，导致邀请关系无法绑定。
- 建议修复：
  - 对 `options.inviteCode` / `options.code` 做防御处理，过滤空值、`undefined`、`null` 字符串。
  - 当 URL 参数无有效邀请码时，优先读取 `enjoy_invite_context` 中保存的邀请码。
  - 登录提交前校验邀请码必须是已确认的有效码。

### 功能模块：接口字段联调

- 证据截图：`docs/evidence/2026-07-02-invite-login-undefined.png`
- 证据现象：后端预检接口能返回有效邀请码，但前端展示错误，属于前后端字段结构理解不一致。
- 业务影响：接口可用但页面消费字段失败，容易造成“后台配置正确、前端仍显示异常”的误判。
- 当前状态：已记录，暂不修复；后续联调时统一接口返回结构和前端归一化逻辑。

### 服务层：`services/invite.js`

- 问题：`verifyInviteCode` 当前返回 `invite: result.data`，而后端 `result.data` 是完整预检对象，不是邀请码对象本身。
- 影响：页面层误以为 `invite.code` 存在，导致保存邀请上下文和跳转参数都取不到真实邀请码。
- 建议修复：
  - 将返回值调整为：
    - `invite: result.data.inviteCode`
    - 同时保留 `entryType`、`authPageMode`、`boundWechat` 等预检字段。
  - `saveInviteContext` 可兼容完整预检对象和邀请码对象，避免同类字段结构问题再次出现。

### 后端/联调备注

- 当前本地后端接口 `POST /api/app/invites/precheck` 对 `PLAYER_LINK` 返回正常，数据结构中有效邀请码字段为 `data.inviteCode.code`。
- 这次问题不是后台邀请码配置错误，也不是数据库缺数据，而是前端页面和服务层对接口返回字段的理解不一致。
- 本地临时绕过方式：
  - 直接启动 `pages/login/index?inviteCode=PLAYER_LINK&entryType=link`
  - 或在登录页手动把 `UNDEFINED` 改成 `PLAYER_LINK`

## 2026-07-02 登录/注册链路：获取验证码提示“未登录”

### 功能模块：登录/注册

- 证据截图：`docs/evidence/2026-07-02-sms-code-unauthorized.png`
- 证据现象：用户在登录/注册页输入手机号 `15008981096` 和邀请码 `PLAYER_LINK` 后，点击“获取验证码”提示未登录。
- 复现路径：
  1. 进入 `pages/login/index?inviteCode=PLAYER_LINK&entryType=link`
  2. 输入手机号
  3. 点击“获取验证码”
  4. 页面提示未登录
- 原因：前端点击“获取验证码”调用 `POST /api/app/sms/send-code`，但后端当前把该接口注册为 `s.AppAuthMiddleware(s.IdempotencyMiddleware(s.sendSMSCode))`，注册/登录前没有 `enjoy_token`，所以接口返回 `401`。
- 影响：邀请注册和手机号验证码登录无法从未登录状态发起，用户卡在注册页。
- 当前状态：已记录，暂不修复；后续集中处理登录注册链路。
- 建议修复：
  - 将注册/登录前使用的短信验证码接口改为公开接口，或新增公开接口，例如 `POST /api/app/auth/sms/send-code`。
  - 保留已登录后的手机号绑定验证码接口继续走登录校验，例如 `/api/app/identity/phone/bind` 或绑定专用短信接口。
  - 前端根据场景区分 `invite_register` / `login` / `bind_phone`，避免所有短信验证码都走同一个需要登录的接口。

### 页面：`pages/login/index`

- 问题：登录/注册页的“获取验证码”按钮绑定 `sendUiPreviewInviteCode`，内部调用 `authService.sendPhoneCode({ phone, scene: 'invite_register' })`，但该请求没有登录态。
- 影响：用户未登录时无法获取验证码，也无法继续“登录/注册”按钮流程。
- 建议修复：
  - 在登录/注册页调用公开短信接口。
  - 如果后端暂未提供公开接口，页面应提示“验证码接口未开放”而不是泛化为“未登录”。

### 后端/联调备注

- 已用本地接口复现：未带 Authorization 调用 `POST /api/app/sms/send-code` 返回 `HTTP 401`。
- 当前问题不是手机号错误，也不是邀请码错误，而是短信验证码接口权限设置不适合注册前场景。

## 2026-07-02 邀请注册链路：邀请页“普通登录”点击后无明显跳转

### 功能模块：邀请注册

- 证据截图：`docs/evidence/2026-07-02-invite-normal-login-no-navigation.png`
- 证据现象：邀请注册页中邀请码 `PLAYER_LINK` 已确认，下方“普通登录”按钮被标注；点击后用户反馈没有跳转。
- 复现路径：
  1. 进入 `pages/login/invite/index?inviteCode=PLAYER_LINK&entryType=link`
  2. 确认邀请码
  3. 点击页面底部“普通登录”
  4. 页面没有明显切换到普通登录流程
- 当前状态：已记录，暂不修复；后续集中处理邀请注册链路。

### 页面：`pages/login/invite/index`

- 问题：页面底部“普通登录”绑定 `goNormalLogin()`，函数内执行 `inviteService.clearInviteContext()` 后调用 `navigateShellRoute(ROUTES.login)`。
- 可能原因：
  - 当前开发者工具页面栈/路由状态导致 `navigateShellRoute` 没有明显页面切换。
  - 跳到 `pages/login/index` 后仍停留在相似的登录视觉布局，用户感知为没有跳转。
  - 跳转没有传入明确的普通登录模式，例如 `mode=account` 或清晰的登录页状态。
- 影响：用户无法从邀请注册页稳定切换到已注册用户登录路径。
- 建议修复：
  - “普通登录”使用更明确的跳转方式，例如 `wx.reLaunch({ url: '/pages/login/index?mode=account' })`。
  - 登录页根据 `mode=account` 直接展示账号密码或微信登录入口。
  - 增加跳转失败日志或 toast，避免点击无反馈。

## 2026-07-02 行家申请：服务成本不支持 0 元

### 功能模块：角色申请 / 行家申请

- 证据截图：`docs/evidence/2026-07-02-expert-apply-zero-cost-invalid.png`
- 证据现象：行家申请业务表单中，“服务成本”输入 `0` 后被红框标记，并提示“金额需为最多 8 位整数和 2 位小数”。
- 复现路径：
  1. 使用玩家账号进入行家申请/行家资料填写流程
  2. 在业务服务项中填写业务名称和服务定价
  3. 在“服务成本”输入 `0`
  4. 页面提示金额格式错误
- 当前状态：已记录，暂不修复；后续集中处理行家申请链路。

### 页面：`pages/home/index`

- 问题：行家申请弹层里的服务成本字段不接受 `0`。
- 代码定位：
  - 表单字段：`pages/home/index.wxml` 中的“服务成本”
  - 校验逻辑：`pages/home/index.js` 的 `isValidMoney`
- 原因：`isValidMoney` 的正则允许 `0`，但最终返回条件包含 `Number(text) > 0`，导致 `0` 被判定为非法金额。
- 影响：行家填写免费服务、无成本服务、平台补贴服务或暂不计算成本时无法提交。
- 建议修复：
  - 区分“服务定价”和“服务成本”的校验规则：服务定价可要求 `> 0`，服务成本应允许 `>= 0`。
  - 将报错文案调整为更准确的提示，例如“服务成本不能小于 0，最多 8 位整数和 2 位小数”。
  - 如果产品规则确实不允许 0 元成本，需要在表单说明中明确“不支持 0 元成本”，避免用户误解为格式错误。

## 2026-07-02 行家申请：提交/审核通过后首页卡片未更新

### 功能模块：角色申请 / 行家申请

- 证据截图：`docs/evidence/2026-07-02-expert-application-submitted-home-not-updated.png`
- 证据现象：玩家提交成为行家申请后，回到首页仍显示“我懂玩家需要什么！我申请成为行家”和“申请成为行家”按钮，没有切换为“已提交/审核中/已成为行家”等状态。
- 复现路径：
  1. 使用玩家账号 `Test Player` 登录小程序。
  2. 提交成为行家的申请。
  3. 后台审核通过该申请。
  4. 回到首页角色卡片，仍展示“申请成为行家”入口。
- 当前本地数据确认：
  - `role_applications` 中用户 `1` 的 `expert` 申请状态为 `approved`。
  - `user_roles` 中用户 `1` 已有 `expert active` 角色。
- 影响：用户可能以为申请未提交或审核未通过，也可能重复进入申请流程；通过审核后的下一步测试路径不清晰。
- 当前状态：已记录，暂不修复；后续集中处理角色申请状态同步和首页状态展示。

### 页面：`pages/home/index`

- 问题：首页/角色卡片仍按静态申请入口展示，没有根据当前用户的角色申请状态或已拥有角色刷新按钮文案和跳转。
- 建议修复：
  - 首页接口或前端状态中合并当前用户角色申请状态。
  - `pending` 时显示“审核中/已提交”，按钮置灰或进入状态页。
  - `approved` 或 `user_roles.expert=active` 时显示“已成为行家/进入行家首页”。
  - `rejected` 时显示“申请未通过/重新申请/查看原因”。

## 2026-07-02 行家首页：页面展示不像正式行家业务首页

### 功能模块：首页 / 行家首页

- 证据截图：`docs/evidence/2026-07-02-expert-home-display-wrong.png`
- 证据现象：进入行家首页后，页面显示为“HELLO，行家！”、“今日推荐”、“快捷入口”、“本周玩霸榜”、“我的成就”、“朋友在玩”、“进入元宇宙”等通用首页内容，整体不像独立的行家业务首页。
- 用户预期：不同角色进入后页面可以有少量差异，但应该保持同一产品体系和对应角色业务逻辑；当前行家页不是“细节差异”，而是整体页面结构和预期完全不一致。
- 页面来源：
  - 路由入口：`pages/home/expert/index`
  - 页面内容：`pages/home/expert/index.wxml` 只渲染 `<role-dashboard-home role-type="expert"></role-dashboard-home>`
  - 实际组件：`components/role-dashboard-home/index`
  - 前端数据接口：`services/home.js` -> `api/modules/home.js` -> `GET /api/app/home`
  - 后端接口：`services/go-api/internal/appapi/game_handler.go` 的 `appHome`，返回 `buildAppHomePayload`
- 当前判断：这不是后台管理页，也不是单独的行家业务详情页，而是通用角色首页组件按 `expert` 参数渲染出来的页面；玩家首页有独立完整页面结构，但行家/领路人入口目前只是通用组件包装。
- 影响：审核通过后用户进入行家页，看到的不是预期的行家工作台/接单/服务/审核/收益等核心业务入口，容易误判为跳转错误或行家权限未生效。
- 当前状态：已记录，后续修改；需要按角色重新确认首页入口和页面结构。

### 页面：`pages/home/expert/index`

- 问题：行家首页入口本身没有独立页面结构，仅包装通用组件。
- 对比：`pages/home/player/index` 是完整玩家首页实现；`pages/home/expert/index` 和 `pages/home/guide/index` 都是 `Page({})` 加通用组件入口。
- 建议确认：
  - 行家首页是否应继续复用 `role-dashboard-home`。
  - 如果应复用，需要后端返回独立 `roleDashboard` 数据，包含行家工作台、可接单/服务配置、收益、审核提醒等。
  - 如果不应复用，需要按原型新增/调整独立行家首页内容。
- 后续修改方向：不同角色进入后应展示各自对应的首页/工作台，允许局部通用，但不能只改角色名称后复用一套不匹配的通用页面。

## 2026-07-02 角色申请：申请条件展示为前端静态假数据

### 功能模块：角色申请 / 申请条件

- 证据截图：`docs/evidence/2026-07-02-role-apply-requirements-static.png`
- 证据现象：申请条件区域显示“玩家等级达到 Lv.20”“当前等级: Lv.21 / 已满足”“信用分 ≥ 90 分”“当前: 92 分 / 已满足”等固定内容。
- 当前来源：这组内容写在 `pages/home/index.js` 的行家申请操作页默认配置中，属于前端静态内容，不是后台按当前用户实时计算后返回。
- 影响：不同测试账号看到同一组条件和同一组“已满足”结果，无法真实判断用户是否具备申请资格，也无法验证后台资格规则。
- 当前状态：已记录，后续修改；申请条件需要从后台获取并按当前用户动态计算。

### 页面：`pages/home/index`

- 问题：行家申请页的 `requirements` 固定写死，包括等级、实名、企业认证、组局次数、信用分、会员等级和已满足状态。
- 建议修复：
  - 后端返回角色申请条件列表和当前用户满足状态。
  - 每个条件至少包含：条件标题、规则值、当前值、是否满足、未满足原因、后台规则版本。
  - 前端只负责渲染后台返回结果，不再写死 `Lv.21`、`92 分`、`已满足` 等用户状态。
  - 行家和领路人应使用各自的后台资格配置，避免行家条件误显示到领路人流程。

### 后端/联调备注

- 现有领路人配置接口：`GET /api/app/role-applications/guide/config`，目前偏页面配置/默认配置。
- 需要新增或扩展资格校验接口，例如：
  - `GET /api/app/role-applications/{roleType}/eligibility`
  - 返回当前用户针对该角色的申请资格计算结果。
