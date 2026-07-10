# Follow-ups

## 2026-07-05 测试分支：组局创建页自动填表

### 功能模块：创建组局

- 当前分支：`codex/game-create-autofill-test`
- 用途：本地测试时打开组局创建页自动填入可发布的测试数据，方便验证创建、后台审核、玩家报名、组局者审核链路。
- 规则：该自动填表仅用于测试分支；合入正式分支前需要关闭或移除 `pages/game/create/index.js` 中的 `CREATE_TEST_AUTOFILL_ENABLED`。
- 状态：已实现非 release 环境自动填表；如果启动 query 带 `autofill=0`，可临时关闭自动填。

## 2026-07-05 组局详情：业务展示字段由后台和数据库兜底

### 功能模块：组局详情 / 创建组局 / 报名审核

- 规则：前端兜底只能用于样式、布局结构、加载态、空态和接口异常处理；后台已有字段但返回空值时，不允许前端自行补业务文案、地址、状态、标题、按钮、人数、规则等内容。
- 当前边界：组局详情中的标题、地址、状态文案、按钮文案、报名审核入口、参与人数、组局者信息等应以后端 `detailDisplay` / `game` / `myRelation` 返回为准。
- 地址处理：如果创建组局时只有经纬度、地址为空，应由后台地图反查或数据库补齐 `address` / `cityName` / `cityCode` 后返回；前端不要把“当前位置”或旧预览地址当作业务地址展示。
- 状态处理：前端不要按用户随口列举的状态写死完整流程，应匹配后台 `game.status`、`detailDisplay.event.time`、`detailDisplay.stats`、`myRelation` 等字段；若新增状态或状态文案，应后台统一返回。
- 报名审核入口：玩家报名后进入 `game_applications` 待审核；组局者入口应由后台关系字段和待审数量决定，前端只按后台返回的 action/route 跳转本局审核列表。
- 当前状态：已记录，后续组局详情和创建组局联调必须按“后台定业务字段，前端只消费和渲染”处理。

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

## 2026-07-06 领路人引荐：选择行家卡片显示测试账号资料

### 功能模块：组局 / 领路人引荐 / 选择行家

- 证据截图：用户提供的调试截图 `d:/temp/codex-clipboard-5ec535b0-f0a0-42d8-81d0-57fd19268ee4.png`。
- 证据现象：领路人引荐流程第 1 步“选择行家”区域中，候选行家卡片显示：
  - 名称：`Test Expert`
  - 头像文字：`EX`
  - 角色标识：`行家`
  - 描述：`strategy / local-play / test-account`
  - 标签：`strategy`、`local-play`、`test-account`
- 用户疑问：选择行家时不应展示这些明显的测试字段，当前信息会影响判断实际组局消息发给了谁。
- 当前判断：疑似本地测试账号资料或候选行家接口返回的测试画像直接透出；需要后续确认候选行家列表接口返回字段来源，前端不应自行把内部标签映射成业务展示内容。
- 当前状态：已记录，暂不修复；后续处理选择行家候选数据、真实显示名和调试账号命名时一起确认。

## 2026-07-06 组局动态：查看全部后的组局消息页需按原型补齐

### 功能模块：消息中心 / 组局动态 / 领路人组局消息

- 证据截图：用户提供的参考图 `d:/temp/codex-clipboard-f3be3199-29d8-4cde-aea3-15e86dfc32ec.png`。
- 用户预期：消息中心“组局动态”右侧查看全部进入 `pages/game/guide-progress/index` 后，应展示类似参考图的领路人组局消息页，包括“进行中的组局”数量、进行中/待双方确认卡片、双方头像与确认状态、剩余时间、提醒条，以及最近完成的成功/取消摘要。
- 当前状态：本次先继续处理消息中心组局动态卡片字段和跳审核详情；用户已明确该组局消息页后面再补。
- 后续建议：补齐后端 `GET /api/app/game-invites/guide-progress` 返回字段，再让 `components/guide-progress-card` 和 `components/guide-complete-card` 只按后台字段渲染，不在前端自行拼业务状态文案。

## 2026-07-07 行家审核详情：玩家信息卡后台数据联调

### 功能模块：组局 / 行家审核详情 / 玩家信息

- 当前处理：已移除 `pages/game/audit-detail/index.js` 中玩家信息卡的临时固定预览数据，页面改为读取后台 `detailDisplay.player` 与本次邀请填写字段。
- 当前状态：待用后台真实数据联调；前端只保留玩家信息卡结构、颜色和期望时间/玩家备注图标样式。
- 注意事项：公司、职位、标签、行业等来自个人资料；需求、预计时长、期望时间、玩家备注来自本次组局邀请填写，不应写在前端。免费局不应返回或展示预算金额。

## 2026-07-07 局 IM 群聊页：静态群聊结构已落地，待补真实群资料与高级消息

### 功能模块：组局 / 局 IM 群聊

- 页面：`pages/im/room/index`
- 当前处理：已把原简陋 IM 页面改为多对多局群聊界面，包含局聊顶部导航、在线/成员入口、系统卡片、自己/他人气泡、图片/文件消息样式、底部语音/输入/表情/加号栏和加号面板。
- 当前边界：保留现有 `GET /chat-room`、`GET /chat-session`、`GET /chat/messages`、`POST /chat/messages` 接口逻辑；文本发送、图片/文件上传发送仍走现有接口；语音、位置、表情面板、文件下载/预览暂未做真实业务。
- 权限规则：局外成员不可进入、不可查看；局结束后成员只能查看历史消息，不能继续发送消息；后台管理端可查看 IM 消息与文件记录，并可下载文档。
- 待完善：
  - 后端消息列表需要返回发送者昵称、头像、角色标签，否则群聊里只能按 `senderUserId` 做临时展示。
  - 后端需要返回系统消息结构化 payload，用于组局邀请、组局成功、成员加入、时间地点变更等卡片，而不是前端推断卡片内容。
  - 图片/文件消息需要补下载 URL 或预览 URL，前端再接入图片预览、文件打开/下载。
  - 文件传输链路需要明确：请求上传凭证、上传文件、落库 `fileId`、详情页与 IM 均可查看；后台管理端需要按 `fileId` 获取下载地址并下载文档。
  - 语音消息、位置消息、表情消息需要确认消息类型、上传/存储字段和展示规则。
- 状态：已记录，后续接口联调阶段处理。

## 2026-07-07 局 IM 群聊页：真实后台 IM DTO 与文件发送联调

### 功能模块：组局 / 局 IM 群聊

- 页面：`pages/im/room/index`
- 后端：`services/go-api/internal/appapi/im_handler.go`
- 当前处理：App 侧 IM 房间、会话和消息接口已改为返回页面消费所需 DTO，包含局标题、当前用户、成员资料、发送者姓名、角色、头像文字/头像地址和消息时间等字段；发送消息接口也返回同一消息 DTO。
- 前端处理：IM 页面已移除 `devUserId` 测试用户透传，真实联调按登录 token 识别当前用户；图片/文件发送仍使用现有 `chat_file` 上传凭证和 `fileId` 发送消息链路。
- 验证结果：`go test ./internal/appapi ./internal/im ./internal/files` 已通过；`pages/im/room/index.js`、`services/file.js`、`services/im.js` 已通过基础 JS 语法检查。
- 待验证：需要在真实登录态下手动验证局成员进入、局外用户 40331、已结束局只读、文本发送、图片/文件上传发送、后台按 `fileId` 下载文件。
- 已知限制：整服务 `go test ./...` 跑到 `cmd/server` 时被本机 Go 模块缓存目录权限阻断，未能完成全量测试；当前不是 IM 相关包编译失败。
- 状态：已记录，下一步建议用本地 Go 后台和小程序真实 token 做端到端联调。

## 2026-07-07 局 IM 群聊页：图片/文件/后台/归档后禁发接口实测

### 功能模块：组局 / 局 IM 群聊 / 后台 IM 证据

- 新增测试：`services/go-api/internal/appapi/im_integration_test.go`
- 覆盖场景：
  - 单独创建 5 个局内测试成员和 1 个局外测试用户，不复用现有角色确认或服务确认数据。
  - 局内成员进入 `GET /api/app/games/{gameId}/chat-room`，返回局标题、成员列表、角色、昵称和头像字段。
  - 局外用户访问 `chat-room`、`chat/messages`、图片/文件下载地址均返回 `40331`。
  - 上传图片和文件均走 `POST /api/app/files/upload-token`，业务类型为 `chat_file`。
  - 发送图片和文件均走 `POST /api/app/games/{gameId}/chat/messages`，消息返回 `fileId`、发送者昵称、角色文案和头像字段。
  - 所有局内成员均可查看消息，并可通过 `GET /api/app/files/{fileId}/download-url` 获取图片/文件下载地址。
  - 后台使用测试后台账号登录后，通过 `GET /api/admin/im/rooms`、`GET /api/admin/im/rooms/{roomId}` 查看 IM 房间、消息数量、文件消息数量和文件消息记录。
  - 后台可通过 `GET /api/admin/files/{fileId}/download-url` 获取图片/文件下载地址。
  - 后台归档 IM 房间后，所有局内成员继续发送消息均返回 `42200`，不能再发。
- 本次修复：
  - `chat-room` 在成员可访问但房间尚未创建时会先 `EnsureRoom`，避免前端必须先调用 `chat-session` 才能进入房间。
  - IM 成员和消息发送者展示名改为优先使用用户昵称，实名仅作为兜底，避免群聊中把所有测试账号显示成同一个实名。
- 验证结果：`go test ./internal/appapi -run TestIMChatFilesDownloadsAdminRecordsAndReadonlyAfterArchive -count=1 -v` 已通过；`go test ./internal/appapi ./internal/im ./internal/files` 已通过。
- 状态：接口自动化实测已通过；后续还需在小程序开发者工具中用真实页面点选图片/文件做一次人工冒烟。

## 2026-07-07 免费局服务交付确认页：行家/玩家视觉对齐

### 功能模块：组局 / 服务交付确认 / 免费局

- 参考来源：用户提供截图 `C:/Users/49637/Downloads/免费局行家服务交付确认@3x.png`。
- 本次处理：`pages/game/delivery/index` 免费局模式按截图调整为蓝色完成状态卡、白色活动信息卡、绿色免费局说明、确认状态时间线、蓝色选中服务内容确认、快捷操作和底部确认按钮；玩家侧复用同页相似结构，通过 `role=player` 或接口 `viewer.role` 自动切换联系对象与角色标签。
- 当前边界：未新增接口；页面仍读取后端 `deliveryPage`、`deliveryProof`、`participants`、`activityRows`、`nextSteps` 等字段。免费局隐藏交付凭证上传卡，避免和截图不一致。
- 验证结果：`node --check pages/game/delivery/index.js` 通过；本页 WXSS 未命中已知 Skyline 不支持项；`git diff --check` 通过。
- 待验证：需要在微信开发者工具中分别以行家和玩家身份打开 `pages/game/delivery/index?gameId=...&mode=free&role=expert/player`，确认顶部状态卡高度、底部按钮、安全区、玩家侧联系人文案和真实接口数据展示是否与原型一致。
- 状态：静态视觉和基础检查已完成，待小程序开发者工具视觉验收。

## 2026-07-07 组局邀请：玩家确认页待重启后验证

### 功能模块：组局 / 领路人邀请行家 / 玩家确认

- 本次处理：新增 `pages/game/player-confirm/index`，玩家收到领路人邀请行家的组局消息后进入玩家确认页，不复用行家审核页的操作身份。
- 当前边界：玩家确认页只处理玩家对本次邀请的同意/拒绝；行家未确认前不跳成功页，全部确认后仍由组局成功动态分别跳到领路人/行家成功页。
- 待验证：后端重启后，用玩家C账号检查组局动态通知是否跳转到 `pages/game/player-confirm/index?invitationId=25&gameId=37`。
- 状态：代码与基础检查已完成，等待后端重启后做接口实测。

## 2026-07-07 免费局服务确认页：独立本地测试数据

### 功能模块：组局 / 服务确认 / 免费局

- 本次处理：新增本地夹具脚本 `scripts/seed-local-service-confirm-fixture.ps1`，写入隔离测试用户、固定 devToken、本地免费局、成员角色和待确认服务确认状态；不预置任何一方确认记录，不复用已有 IM 局、已有组局或 `gameId=38`。
- 固定数据：`gameId=86002`；行家 `userId=86021` / `devToken=zhw-delivery-free-expert-86021`；玩家 `userId=86022` / `devToken=zhw-delivery-free-player-86022`。
- 编译模式：`project.private.config.json` 中“服务确认-免费局-行家”和“服务确认-免费局-玩家”已指向上述测试数据。
- 验证结果：本地数据库记录已写入；`GET /api/app/games/86002/success-detail?role=expert` 和 `role=player` 均返回成功，包含 2 个参与人、4 条活动信息、3 个下一步、3 个免费局确认项；`node --check pages/game/delivery/index.js` 通过；`git diff --check` 通过。
- 待验证：仍需在微信开发者工具中分别打开两个编译模式，人工确认视觉展示、底部按钮、安全区和玩家侧联系人文案。

## 2026-07-07 免费局服务确认页：服务口径活动信息与确认流程

### 功能模块：组局 / 服务确认 / 免费局

- 本次处理：服务确认页活动信息改为服务口径字段：`服务类型=免费局`、`服务时长=...（已完成）`、`开始时间`、`完成时间`；移除本卡片里的活动编号、创建时间、组局时间和“待交付服务”阶段文案。
- 本次处理：确认状态流程按“行家已确认完成 → 玩家确认完成 → 服务归档”展示；行家确认后第一步描述为“行家标记服务已完成 + 时间”，第二步为“需玩家确认服务已达标”。
- 本次处理：服务保障盾牌改为页面内 `shield-service.svg`，快捷操作联系人图标改为页面内深色 SVG，避免浅灰底白色 PNG 看不清。
- 待商榷：最终业务流程仍需确认“行家确认服务完成”和“玩家确认服务达标”的按钮文案、可重复提醒规则、归档触发条件，以及玩家侧已确认后的按钮状态。
- 验证结果：`node --check pages/game/delivery/index.js` 通过；`go test ./internal/appapi` 通过；`git diff --check` 通过。

## 2026-07-07 组局邀请：行家确认页 audit-detail 收尾验证

### 功能模块：组局 / 领路人邀请玩家与行家 / 行家确认

- 本次处理：`pages/game/audit-detail/index` 的 shell 左键显式走返回逻辑；“通过/婉拒”操作区固定到页面内容底部；行家信息卡恢复展示后端返回的角色与标签信息；`components/game-detail/party-card` 中 audit-detail 的连接样式恢复为居中虚线；后端 `invitationAuditRelationDisplay` 的确认进度总人数改为优先取 `game.max_players`。
- 本次补充：倒计时文案为空时前端和状态卡组件兜底显示 `00:00:00`，避免出现空胶囊；行家信息卡按“专业领域 / 服务介绍”分块展示后端行家技能、服务标签和申请资料；玩家侧也返回联系行家的 IM 入口，行家侧保留联系玩家/建议时间入口；底部操作区继续固定展示。
- 数据结论：本地 `gameId=38` 为 `current_players=3`、`max_players=5`，成局双方确认进度应显示为 `3/5人已确认`，不应显示 `3/3人已确认`。
- 待验证：需要后端重启后，用行家A `token=zhw-test-expert-a-80` 打开 `pages/game/audit-detail/index?invitationId=28&gameId=38`，确认行家信息卡、倒计时、成局双方连接线、联系入口、底部按钮、左键返回和 `3/5人已确认` 均生效。
- 已知限制：当前 8080 旧后端进程可能仍在运行，新 Go 代码可能未生效；如仍看到旧数据，需要先手动停止旧进程并运行 `scripts/restart-im02-backend.ps1`。

## 2026-07-07 免费局服务确认页：前后台实测与缺口核实

### 功能模块：组局 / 服务确认 / 免费局

- 本次实测：重置本地夹具 `gameId=86002` 后，分别用行家 `devToken=zhw-delivery-free-expert-86021` 和玩家 `devToken=zhw-delivery-free-player-86022` 请求 `GET /api/app/games/86002/success-detail?role=expert/player`，接口均返回成功；活动信息包含服务类型、服务时长、开始时间、完成时间；参与人包含行家和玩家；时间线初始为等待行家确认。
- 本次实测：行家提交 `POST /api/app/games/86002/service-confirm` 后，确认单状态为 `pending`，局状态为 `pending_confirm`，玩家侧详情显示“行家已确认完成 / 等待玩家确认”；玩家提交后，确认单状态为 `completed`，局状态为 `pending_review`，详情显示“行家已确认完成 / 玩家已确认完成 / 服务归档”。
- 验证结果：`node --check pages/game/delivery/index.js` 通过；`go test ./internal/games ./internal/appapi ./internal/delivery` 在设置项目内 `GOCACHE` 后通过；测试结束后已再次运行 `scripts/seed-local-service-confirm-fixture.ps1`，将 `gameId=86002` 重置回未确认状态，方便小程序端继续手动点测。
- 已核实缺口：前端提交后重新加载详情，但确认勾选项仍来自后端静态 `deliveryPage.free.confirmItems`，没有结合当前用户是否已确认来锁定或切换按钮状态；用户已确认后页面仍可能显示可再次勾选提交。后端重复提交按用户幂等处理，不会新增确认项，但前端状态不够清晰。
- 已核实缺口：后端完成条件使用本局全部成员数判断，而页面时间线只表达行家和玩家两步。如果真实三方局包含领路人成员，可能出现行家和玩家都确认后仍未完成，需要明确“服务确认参与人”到底是行家+玩家，还是包含领路人，并据此调整后端完成条件或页面流程。
- 已核实缺口：玩家侧联系对象依赖前端 `role=player` 自行切换；后端 `deliveryProof` 目前只返回 `contactPlayerText/contactPlayerPrefill/contactGuideText`，没有返回 `contactExpertText/contactExpertPrefill`。快捷操作有前端兜底，但更多菜单仍可能显示“联系玩家”而不是“联系行家”，不符合后台字段为准的规则。
- 已核实缺口：前端只向 `service-confirm` 提交 `note` 和 `fileIds`，没有提交用户实际勾选的确认项明细；后端确认记录也只保存用户、备注和文件。若后续需要审计“服务已全部完成 / 服务质量达标 / 双方已沟通确认”具体勾选结果，需要补字段或接口。
- 待验证：还需要在微信开发者工具中分别打开“服务确认-免费局-行家”和“服务确认-免费局-玩家”两个编译模式，人工检查视觉展示、底部安全区、更多菜单、重复确认后的按钮状态，以及真实触摸交互。

## 2026-07-07 组局支付确认页：页面存在与占位支付实测

### 功能模块：组局 / 支付确认 / 免费局占位订单

- 页面确认：项目中已存在 `pages/game/payment/index`，并已配置到 `app.json` 与 `config/routes.js`，页面文件包含 `index.js`、`index.wxml`、`index.wxss`、`index.json`。
- 本次修复：页面从 URL 读取的 `gameId` 是字符串，提交到后端 `POST /api/app/payment/precreate-placeholder` 时会导致后端数字字段解析失败；已在 `pages/game/payment/index.js` 的提交 payload 中转换为数字。
- 本次实测：用本地免费局夹具 `gameId=86002` 和创建者 token `zhw-delivery-free-player-86022` 调用 `POST /api/app/payment/precreate-placeholder`，接口返回成功，`orderNo=FREE-86002`，`payStatus=free_no_pay`，`needWechatPay=false`。
- 验证结果：`node --check pages/game/payment/index.js` 通过；`go test ./internal/orders ./internal/appapi` 通过；修正后数字 payload 本地接口实测通过。
- 待验证：仍需在微信开发者工具中打开 `pages/game/payment/index?gameId=86002&amount=0`，勾选协议后点击“确认订单”，确认弹窗、loading、跳转回组局详情、底部安全区和页面视觉是否符合预期。

## 2026-07-07 局内协作页：交付操作页实测

### 功能模块：组局 / 局内协作 / 交付操作

- 页面确认：截图对应 `pages/game/collaboration/index`，页面已注册到 `app.json`，路由键为 `ROUTES.gameCollaboration`。
- 本次实测：用本地免费局夹具 `gameId=86002` 请求 `GET /api/app/games/86002/collaboration`，玩家创建者 `zhw-delivery-free-player-86022` 可看到进度、成员、成员管理和结束本局入口；行家 `zhw-delivery-free-expert-86021` 可查看页面数据，但成员管理和结束本局权限为 false。
- 本次修复：后端协作页 `endConfirmRoute` 原先只返回 `pages/game/delivery/index?gameId=...`，从免费局结束本局进入交付确认页时会丢失 `mode=free&role=player`；已改为按局类型返回 `mode=free/paid`，并补 `role=player`。
- 验证结果：`node --check pages/game/collaboration/index.js` 通过；`go test ./internal/appapi ./internal/games` 通过；`project.private.config.json` 已新增“局内协作-免费局-玩家”编译模式。
- 运行阻塞：当前本地 8080 仍由旧 `server` 进程 `pid=15736` 占用，且当前权限无法停止该进程；因此 8080 接口实测仍返回旧 `endConfirmRoute`。需要手动停止旧进程或以可停止权限重启后端后，再复测接口返回是否包含 `mode=free&role=player`。
- 待验证：在微信开发者工具中打开“局内协作-免费局-玩家”，检查进度卡、成员卡、聊天区空态、成员管理按钮、结束本局跳转交付确认页是否符合截图预期。
