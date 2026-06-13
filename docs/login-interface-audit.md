# 登录相关接口与硬编码审计

本文是工程侧审计记录，只读取 `E:\项目\03周总\00.文档`，不修改正式需求文档，也不把外部文档放进项目仓库。

## 依据

- 正式接口文档：`E:\项目\03周总\00.文档\接口、数据库初稿.md`
- 正式功能清单：`E:\项目\03周总\00.文档\功能清单.md`
- 当前工程代码：`E:\项目\03周总\01enjoy`

## 结论先说

当前正式文档 5.1 里明确提供的小程序登录账号接口只有：

| 接口 | 正式文档是否已有 | 当前代码是否使用 | 说明 |
| --- | --- | --- | --- |
| `POST /api/app/auth/wechat-login` | 已有 | 已使用 | 微信登录主接口 |
| `GET /api/app/users/me` | 已有 | 已使用 | 获取当前用户 |
| `PUT /api/app/users/me/profile` | 已有 | 暂未使用 | 个人资料维护 |
| `POST /api/app/invites/codes` | 已有 | 暂未使用 | 生成我的邀请码 |
| `GET /api/app/invites/my-relations` | 已有 | 暂未使用 | 我的邀请关系 |

当前页面里为了跑通“手机号登录 / 账号密码登录 / 找回密码 / 邀请码预校验”，我补了候选接口和 mock。这些还不是正式文档 5.1 已确认接口，需要你和后端确认后再定稿。

## 当前页面与接口归属

### 1. 登录主页 `pages/login/index`, `loginMode = home`

| 页面元素 / 行为 | 数据来源 | 接口状态 | 当前问题 |
| --- | --- | --- | --- |
| `手机号登录` 按钮 | 静态 UI | 不需要接口 | 正常 |
| `账号密码登录` 按钮 | 静态 UI | 不需要接口 | 正常 |
| 协议勾选 | 本地 UI 状态 | 不需要接口 | 正常 |
| 点击手机号登录后默认手机号 | 页面硬编码 | 不应上线 | 当前 `pages/login/index.js` 里默认用了 `13888888888`，只是为了测试，应改为用户输入或正式手机号授权流程 |

### 2. 验证码登录 `pages/login/index`, `loginMode = codeVerify`

| 行为 | 当前接口 | 正式文档是否已有 | 当前数据 |
| --- | --- | --- | --- |
| 获取验证码 | `POST /api/app/auth/phone-code` | 正式 5.1 未列出 | mock 返回验证码 `123456` |
| 手机号验证码登录 | `POST /api/app/auth/phone-login` | 正式 5.1 未列出 | mock 要求手机号 `13888888888`、验证码 `123456` |

说明：手机号验证码登录是墨刀页面需要，但正式接口文档 5.1 当前没有列出，需要作为候选接口确认。

### 3. 账号密码登录 `pages/login/index`, `loginMode = account`

| 行为 | 当前接口 | 正式文档是否已有 | 当前数据 |
| --- | --- | --- | --- |
| 账号密码登录 | `POST /api/app/auth/password-login` | 正式 5.1 未列出 | mock 要求账号 `13888888888`、密码 `Test123456` |
| 找回密码入口 | 跳转本地页面 | 不需要接口 | 正常 |
| 微信登录入口 | 进入 `wechatAuth` 状态 | 微信登录接口已有 | 正常 |

说明：账号密码登录是墨刀页面需要，但正式接口文档 5.1 当前没有列出，需要作为候选接口确认。

### 4. 微信授权 `pages/login/index`, `loginMode = wechatAuth`

| 行为 | 当前接口 | 正式文档是否已有 | 当前对齐情况 |
| --- | --- | --- | --- |
| 微信登录 | `POST /api/app/auth/wechat-login` | 已有 | 路径已对齐 |

正式文档请求字段：

| 字段 | 当前代码情况 | 备注 |
| --- | --- | --- |
| `code` | 已传 | 来自 `wx.login` |
| `inviteCode` | 已预留 | 有邀请上下文时传 |
| `scene` | 暂未传 | 后续扫码 / 分享进入时补 |
| `nickname` | 暂未传 | 后续接微信头像昵称授权时补 |
| `avatarUrl` | 暂未传 | 后续接微信头像昵称授权时补 |

当前偏差：

1. `services/auth.js` 还保留了旧字段 `encryptedData`、`iv`，正式 5.1 没写这两个字段。
2. mock 暂不强制“新用户无邀请不能注册”，因为当前先按你说的保留普通登录入口，后续再隐藏或收紧。

### 5. 找回密码 `pages/login/forgot/index`

| 步骤 | 当前接口 | 正式文档是否已有 | 当前数据 |
| --- | --- | --- | --- |
| 获取验证码 | `POST /api/app/auth/phone-code` | 正式 5.1 未列出 | mock 返回 `123456` |
| 下一步校验验证码 | `POST /api/app/auth/phone-code/verify` | 正式 5.1 未列出，是我补的候选接口 | mock 校验手机号和验证码 |
| 确认修改密码 | `POST /api/app/auth/password/reset` | 正式 5.1 未列出，是我补的候选接口 | mock 校验手机号、验证码、密码长度 |

说明：你说“验证码校验要后台做”是对的，所以当前页面已改为通过 service 调候选接口。只是这些接口还没在正式文档 5.1 定稿。

### 6. 邀请注册 `pages/login/invite/index`

| 行为 | 当前接口 | 正式文档是否已有 | 当前数据 |
| --- | --- | --- | --- |
| 邀请码预校验 | `POST /api/app/invites/verify` | 正式 5.1 未列出，是我补的候选接口 | mock 有效邀请码 `ENJOY2026` |
| 登录时携带邀请码 | `POST /api/app/auth/wechat-login` / `phone-login` | 微信登录接口已有 `inviteCode` 字段 | 已预留 |

正式文档已有“邀请码识别 / 邀请关系永久绑定”的需求和数据库表，但小程序端未登录状态下的“邀请码预校验接口”没有在 5.1 明确列出。

## 当前硬编码清单

### 页面代码里不应长期存在

| 位置 | 写死内容 | 用途 | 建议 |
| --- | --- | --- | --- |
| `pages/login/index.js` | `13888888888` | 手机号登录测试默认值 | 改成必须输入手机号，或接微信手机号授权 |
| `pages/login/index.js` | toast `验证码已发送：123456` | 方便测试 | 正式环境只提示“验证码已发送” |
| `pages/login/forgot/index.js` | toast `验证码已发送：123456` | 方便测试 | 正式环境只提示“验证码已发送” |
| `pages/entry/index.wxml` | `3009人在线` | 入口页展示 | 应来自首页/公开配置接口或 mock 数据 |

### mock 层可以存在，但要明确是测试数据

| 位置 | mock 数据 | 说明 |
| --- | --- | --- |
| `api/mock.js` | `13888888888` | 测试手机号 |
| `api/mock.js` | `123456` | 测试验证码 |
| `api/mock.js` | `Test123456` | 测试密码 |
| `api/mock-data.js` | `ENJOY2026` | 测试邀请码 |
| `api/mock-data.js` | `微信用户` | mock 用户昵称 |
| `api/mock-data.js` | `3999人在线` | mock 首页在线人数 |
| `api/mock-data.js` | `Alex Chen` | mock 玩家成长展示名 |

## 建议新增 / 待确认接口

这些不是正式 5.1 已明确接口，建议继续放在 `docs/api-additions.md` 里，不直接改正式文档。

| 候选接口 | 用途 | 是否已有代码 | 是否需后端确认 |
| --- | --- | --- | --- |
| `POST /api/app/auth/phone-code` | 获取手机验证码 | 已有 | 是 |
| `POST /api/app/auth/phone-code/verify` | 只校验验证码，供找回密码下一步使用 | 已有 | 是 |
| `POST /api/app/auth/phone-login` | 手机验证码登录 | 已有 | 是 |
| `POST /api/app/auth/password-login` | 账号密码登录 | 已有 | 是 |
| `POST /api/app/auth/password/reset` | 找回密码 / 重置密码 | 已有 | 是 |
| `POST /api/app/invites/verify` | 邀请码预校验 | 已有 | 是 |

## 需要马上修正的代码项

1. 页面层不再显示 `验证码已发送：123456`，只在测试文档里告诉你验证码是 `123456`。
2. `pages/login/index.js` 不应默认塞入手机号 `13888888888`，否则测试会绕过真实输入。
3. `services/auth.js` 的微信登录请求字段应按正式 5.1 收敛为 `code / inviteCode / scene / nickname / avatarUrl`。
4. `docs/api-additions.md` 需要补齐手机验证码、密码登录、重置密码等候选接口。
