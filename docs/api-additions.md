# 接口补充记录

本文只记录当前前端开发中发现的候选补充接口，不直接修改正式 PRD、功能清单或接口数据库文档。

正式开发仍以 `00.文档/接口、数据库初稿.md` 第 20 章为准。

## 1. 邀请码预校验接口

### 背景

当前普通登录页不再展示邀请码。后续如果补“邀请注册”链路，用户可能会先输入邀请码并看到“已确认 / 无效”的即时反馈。

正式文档第 20.4.1 已定稿微信登录接口：

```text
POST /api/app/auth/wechat-login
```

该接口可以在登录时携带 `inviteCode`，但没有单独定义“未登录状态下的邀请码预校验接口”。

### 候选前端路径

```text
POST /api/app/invites/verify
```

### 建议入参

```json
{
  "code": "ENJOY2026"
}
```

### 建议返回

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": "invite-code-id",
    "code": "ENJOY2026",
    "ownerUserId": "10003",
    "inviterName": "邀请人昵称",
    "city": "城市",
    "status": "active"
  },
  "requestId": "req_xxx"
}
```

### 风险和待确认

1. 该接口是否允许未登录访问，需要后端确认。
2. 返回字段是否暴露邀请人信息，需要产品和隐私口径确认。
3. 如果不新增该接口，邀请注册页可以只在最终注册 / 微信登录时把 `inviteCode` 一并提交。

## 2. 手机验证码与账号密码登录接口

### 背景

墨刀已有手机号验证码登录、账号密码登录、找回密码流程，但正式 `接口、数据库初稿.md` 第 5.1 目前只明确列出了微信登录、当前用户、个人资料、邀请码相关接口。

当前工程为了让登录流程可测试，先在 `api/request.js`、`services/auth.js` 和 `api/mock.js` 中实现了以下候选接口。它们不是正式文档已定稿内容，需要后端确认。

### 2.1 获取手机验证码

候选路径：

```text
POST /api/app/auth/phone-code
```

建议入参：

```json
{
  "phone": "13888888888",
  "scene": "login"
}
```

建议返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "phone": "13888888888",
    "expiresIn": 60
  },
  "requestId": "req_xxx"
}
```

说明：真实环境不能返回验证码明文；当前 mock 返回 `123456` 只用于测试。

### 2.2 手机验证码预校验

候选路径：

```text
POST /api/app/auth/phone-code/verify
```

建议入参：

```json
{
  "phone": "13888888888",
  "code": "123456",
  "scene": "password_reset"
}
```

建议返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "phone": "13888888888",
    "verified": true
  },
  "requestId": "req_xxx"
}
```

用途：找回密码第一步点击 `下一步` 时由后台校验验证码是否正确，前端不写死判断。

### 2.3 手机验证码登录

候选路径：

```text
POST /api/app/auth/phone-login
```

建议入参：

```json
{
  "phone": "13888888888",
  "code": "123456",
  "inviteCode": "ENJOY2026"
}
```

建议返回：同微信登录，包含 `token` 和 `user`。

### 2.4 账号密码登录

候选路径：

```text
POST /api/app/auth/password-login
```

建议入参：

```json
{
  "account": "13888888888",
  "password": "Test123456"
}
```

建议返回：同微信登录，包含 `token` 和 `user`。

### 2.5 找回密码 / 重置密码

候选路径：

```text
POST /api/app/auth/password/reset
```

建议入参：

```json
{
  "phone": "13888888888",
  "code": "123456",
  "password": "Test123456"
}
```

建议返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "reset": true
  },
  "requestId": "req_xxx"
}
```

### 待确认

1. 是否允许手机号验证码登录，还是一期只保留微信登录。
2. 是否允许账号密码登录，账号字段使用 `phone`、`account` 还是用户名。
3. 找回密码是否需要先返回一次性 `resetToken`，再提交新密码。
4. 手机验证码是否按 `scene` 区分登录、找回密码、绑定手机号。

## 3. 实名认证页面入口接口

记录日期：2026-06-14

模块：邀请注册 / 实名认证

页面：`pages/login/index` 邀请注册 5 页走查中的 `实名认证弹窗`、`实名认证引导`

功能：未实名用户点击 `开始认证` 时，由后台返回实名认证页面地址；已实名用户通过 `GET /api/app/users/me` 判断后不弹窗，直接进入个人首页。

状态：候选接口，待后端确认。

候选路径：

```text
POST /api/app/users/me/realname-auth
```

建议入参：无。

建议返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "url": "/pages/login/realname/index",
    "provider": "mock"
  },
  "requestId": "req_xxx"
}
```

待确认：

1. 后台返回的是小程序内页面路径、H5 链接，还是第三方小程序跳转参数。
2. 如果返回 H5 链接，前端需要新增或复用 `web-view` 承载页，并确认业务域名白名单。

## 4. 新手任务状态接口

记录日期：2026-06-14

模块：邀请注册 / 新手任务

页面：`pages/login/index` 邀请注册 5 页走查中的 `完成新手任务` 弹窗

功能：实名认证完成后，由后台返回新手任务列表、任务类型、完成数量、总数量和进度；前端根据每条任务的完成状态把右侧文案显示为 `已完成` 或 `去完成`。

状态：候选接口，正式接口文档中暂未找到已定稿的新手任务列表接口，待后端确认。

候选路径：

```text
GET /api/app/newbie-tasks
```

建议返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "completedCount": 1,
    "totalCount": 3,
    "progressPercent": 33,
    "tasks": [
      {
        "id": "newbie-realname",
        "type": "realname",
        "title": "完成实名认证",
        "rewardText": "+50 经验值",
        "completed": true,
        "status": "completed",
        "statusText": "已完成",
        "actionText": "去完成",
        "route": "pages/profile/index"
      }
    ]
  },
  "requestId": "req_xxx"
}
```

待确认：

1. 是否需要独立新手任务接口，还是由 `GET /api/app/dashboard/me` 或 `GET /api/app/growth/me` 扩展返回。
2. 任务类型枚举是否固定为 `realname`、`profile`、`first_game`，以及奖励文案是否由后台直接返回。
3. 未完成任务点击 `去完成` 时是否由后台返回 `route`，当前前端兜底为：`profile` -> `pages/profile/index`，`first_game` -> `pages/game/create/index`，`realname` -> `pages/login/index?ui=1&mode=realnameGuide`。
