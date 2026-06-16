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

## 4. 实名认证信息提交接口

记录日期：2026-06-16

模块：邀请注册 / 实名认证

页面：`pages/login/realname/index`

功能：用户在实名认证页填写姓名和身份证号后提交给后台校验；校验通过后前端进入新手任务奖励页，校验失败时提示 `认证失败，请重新核对后填写`。

状态：候选接口，待后端确认正式路径、字段名、失败码和是否接第三方实名服务。

候选路径：

```text
POST /api/app/users/me/realname-auth/submit
```

建议入参：

```json
{
  "realname": "测试用户",
  "idNumber": "110101199003070011"
}
```

建议返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "realnameStatus": "verified",
    "verified": true
  },
  "requestId": "req_xxx"
}
```

待确认：

1. 正式接口路径是否使用 `/api/app/users/me/realname-auth/submit`。
2. 字段名使用 `realname/idNumber`，还是后端已有的 `name/idCard`、`realName/idNo`。
3. 后台失败时返回码和文案是否统一为 `认证失败，请重新核对后填写`。
4. 身份证号、姓名属于敏感信息，请求日志、错误日志和埋点必须脱敏或不记录明文。

## 5. 新手任务状态接口

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

## 6. 玩家首页玩家信息卡聚合接口

记录日期：2026-06-16

模块：首页 / 玩家首页

页面：`pages/home/player/index`

功能：玩家首页黑色内容区顶部的玩家信息卡需要由后台返回当前角色、昵称、等级、经验、升级差值、参与局数、本月 MVP 和参与率，并根据当前角色点亮 `玩家 / 行家 / 领路人` 按钮。

状态：候选接口，项目内前端已先通过 `GET /api/app/home` 和 mock 数据接入；正式接口文档中暂未找到已定稿的 `/api/app/home` 首页聚合接口，也未找到本月 MVP、参与率、距离下一级还差多少经验的明确返回字段，待后端确认。

候选路径：

```text
GET /api/app/home
```

建议返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "hero": {
      "roleName": "玩家",
      "dateLabel": "2026.03.30",
      "subtitle": "开启你的今日副本",
      "onlineText": "3999人在线"
    },
    "playerSummary": {
      "currentRole": "player",
      "roleType": "player",
      "displayName": "Alex Chen",
      "nickname": "Alex Chen",
      "roleLabel": "玩家 Lv.5",
      "title": "活跃达人",
      "level": 5,
      "nextLevel": 6,
      "experience": 580,
      "nextLevelExperience": 1000,
      "expToNextLevel": 420,
      "xpText": "580/1000 XP",
      "progressPercent": 58,
      "nextLevelText": "距离下一等级还需 420 经验值",
      "joinCount": 12,
      "monthlyMvpCount": 3,
      "participationRate": "98%",
      "stats": [
        { "label": "参与局数", "value": "12" },
        { "label": "本月MVP", "value": "3" },
        { "label": "参与率", "value": "98%" }
      ]
    }
  },
  "requestId": "req_xxx"
}
```

待确认：

1. 玩家首页是否定稿为独立 `GET /api/app/home` 聚合接口，还是由 `GET /api/app/users/me`、`GET /api/app/growth/me` 和 `GET /api/app/dashboard/me` 拼装。
2. 当前角色字段使用 `currentRole`、`roleType` 还是复用 `CurrentUserDTO.roles` 中的主角色，枚举是否固定为 `player/expert/guide`。
3. `level/experience/nextLevelExperience/expToNextLevel/progressPercent` 是否由后台直接计算并返回，还是只返回经验和等级由前端计算。
4. `joinCount/monthlyMvpCount/participationRate` 是否作为独立字段返回，还是统一放在 `stats` 数组中由后台返回展示文案。

## 7. 玩家首页地球 online 计数接口

记录日期：2026-06-16

模块：首页 / 玩家首页

页面：`pages/home/player/index`

功能：玩家首页 `地球online` 卡片中的 `附近 N 个组局` 和 `已打卡 N 处` 需要由后台返回，不再由页面写死。

状态：待后端确认。项目内当前只有 `GET /api/app/home` mock 中的 `nearbySummary.count`，可覆盖附近组局数量；正式接口文档中暂未找到已定稿的首页地球 online 聚合字段，也未找到已打卡数量字段。项目内 `GET /api/app/games/nearby` 仅作为附近组局列表候选接口声明，mock 未配置，不能直接覆盖首页聚合计数。

候选路径：

```text
GET /api/app/home
```

建议返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "nearbySummary": {
      "nearbyGameCount": 12,
      "checkedInCount": 8,
      "cityName": "上海",
      "accuracyText": "定位精度 300m 内"
    }
  },
  "requestId": "req_xxx"
}
```

待确认：

1. 首页 `地球online` 是否继续复用 `GET /api/app/home` 的 `nearbySummary`，还是新增独立地球 online / 地图摘要接口。
2. 附近组局数量字段使用现有 `count`，还是改为更明确的 `nearbyGameCount`。
3. 已打卡数量字段命名使用 `checkedInCount`、`checkinCount` 还是由后台直接返回展示文案。

## 8. 玩家首页组局卡封面和参与者头像字段

记录日期：2026-06-17

模块：首页 / 玩家首页 / 组局卡片

页面：`pages/home/player/index`、`components/game-card/index.*`

功能：`附近正在发生`、`朋友都在玩` 等组局卡片需要展示封面图、参与者头像、入局人数文案和右侧操作项。当前公共组件已支持通过数据填充，测试阶段使用本地默认封面和头像占位。

状态：待后端确认。用户指定的 HTML 目录和对外资料包中未找到截图里的三枚卡通头像原图；正式阶段建议后台直接返回参与者头像 URL 列表。

候选路径：

```text
GET /api/app/home
GET /api/app/games/nearby
```

建议字段：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "recommendedGames": [
      {
        "id": "game_001",
        "title": "城市探索夜跑",
        "coverUrl": "https://cdn.example.com/games/game_001-cover.jpg",
        "avatarUrls": [
          "https://cdn.example.com/users/u001-avatar.jpg",
          "https://cdn.example.com/users/u002-avatar.jpg",
          "https://cdn.example.com/users/u003-avatar.jpg"
        ],
        "joinedText": "+5位玩家已入局",
        "actions": ["分享", "关注", "引荐", "打招呼"]
      }
    ]
  },
  "requestId": "req_xxx"
}
```

待确认：

1. 首页卡片列表是否继续放在 `GET /api/app/home`，还是附近组局改用独立 `GET /api/app/games/nearby`。
2. 封面字段命名使用 `coverUrl`、`coverSrc` 还是后端已有字段。
3. 参与者头像字段命名使用 `avatarUrls`、`participantAvatarUrls` 还是跟随参与者对象数组返回。
4. 右侧操作项是由前端固定，还是由后台返回 `actions` 以支持不同卡片能力。
