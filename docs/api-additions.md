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

## 8. 玩家首页组局卡列表接口字段

记录日期：2026-06-17

更新日期：2026-06-18

模块：首页 / 玩家首页 / 组局卡片

页面：`pages/home/player/index`、`components/game-card/index.*`

功能：玩家首页 `附近正在发生` 和 `朋友在玩` 都使用公共组局卡。区块标题、`全部 / 附近` tab、`查看全部`、朋友数量和卡片列表都需要由后台返回。卡片需要由后台返回不同组局的数据，包括局类型、封面图、标题、地点距离、人数、时间、价格、参与玩家头像和当前用户可执行动作。

状态：待后端确认。当前页面仍使用本地静态测试数据；公共组件已支持 `coverSrc`、`avatarUrls`、`joinedText`、`actions` 等字段。用户指定的 HTML 目录和对外资料包中未找到截图里的三枚卡通头像原图，正式阶段建议后台返回真实参与者头像 URL。

推荐接口方案：

```text
GET /api/app/home
```

用于首页首屏聚合，直接返回玩家首页需要的 `nearbyGames` 和 `friendGames`。如果后续附近列表需要分页、筛选或地图联动，再补独立列表接口：

```text
GET /api/app/games/nearby?scope=all|nearby&page=1&pageSize=10
GET /api/app/games/friends?page=1&pageSize=10
```

推荐响应结构：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "nearbySection": {
      "title": "附近正在发生",
      "tabs": [
        { "key": "all", "name": "全部" },
        { "key": "nearby", "name": "附近" }
      ]
    },
    "nearbyGames": [
      {
        "id": "game_001",
        "scope": "nearby",
        "title": "苏州河“记忆碎片”采集",
        "type": "explore",
        "typeText": "探索局",
        "coverUrl": "https://cdn.example.com/games/game_001-cover.jpg",
        "priceText": "￥0/人",
        "cityName": "静安区",
        "distanceText": "3.2km",
        "memberText": "5/8人",
        "timeText": "2026年5月1日 20:00--22:00",
        "joinedCount": 5,
        "joinedText": "+5位玩家已入局",
        "participantAvatars": [
          "https://cdn.example.com/users/u001-avatar.jpg",
          "https://cdn.example.com/users/u002-avatar.jpg",
          "https://cdn.example.com/users/u003-avatar.jpg"
        ],
        "actionText": "加入",
        "actions": ["share", "follow", "refer", "greet"],
        "route": "pages/game/detail/index?id=game_001"
      }
    ],
    "friendSection": {
      "icon": "🎲",
      "title": "朋友在玩",
      "count": 2,
      "moreText": "查看全部"
    },
    "friendGames": [
      {
        "id": "game_friend_001",
        "title": "盲盒路线：3小时点亮天际线",
        "type": "explore",
        "typeText": "探索局",
        "coverUrl": "https://cdn.example.com/games/game_friend_001-cover.jpg",
        "priceText": "￥29/人",
        "cityName": "梧桐山",
        "distanceText": "1.5km",
        "memberText": "3/8人",
        "timeText": "2026年5月1日 14:00--16:00",
        "joinedCount": 3,
        "participantAvatars": [
          "https://cdn.example.com/users/u101-avatar.jpg",
          "https://cdn.example.com/users/u102-avatar.jpg",
          "https://cdn.example.com/users/u103-avatar.jpg"
        ],
        "friendContextText": "好友正在玩",
        "actionText": "加入",
        "actions": ["share", "follow", "refer", "greet"],
        "route": "pages/game/detail/index?id=game_friend_001"
      }
    ]
  },
  "requestId": "req_xxx"
}
```

字段分工：

| 字段 | 来源 | 用途 | 说明 |
| --- | --- | --- | --- |
| `nearbySection.title` | 后台 | 附近区块标题 | 当前文案为 `附近正在发生`。 |
| `nearbySection.tabs` | 后台 | 附近区块筛选 tab | 当前为 `全部 / 附近`；前端按 `key` 过滤。 |
| `friendSection.icon` | 后台 | 朋友在玩标题图标 | 当前为骰子图标。 |
| `friendSection.title` | 后台 | 朋友区块标题 | 当前文案为 `朋友在玩`。 |
| `friendSection.count` | 后台 | 朋友在玩数量角标 | 当前 HTML 显示 `2`。 |
| `friendSection.moreText` | 后台 | 查看全部文案 | 当前文案为 `查看全部`。 |
| `id` | 后台 | 跳转详情、埋点 | 必传。 |
| `scope` | 后台 | `全部 / 附近` 筛选 | `nearbyGames` 中建议返回，取值如 `nearby`、`city`。 |
| `title` | 后台 | 卡片标题 | 必传。 |
| `type`、`typeText` | 后台 | 局类型标签 | 前端按 `type` 或 `typeText` 判断颜色；当前 `task=任务局` 黄橙底白字，`explore=探索局` 绿底白字。 |
| `coverUrl` | 后台 / CDN | 卡片左侧封面 | 当前前端测试字段为 `coverSrc`，正式建议统一为 `coverUrl`。 |
| `priceText` | 后台 | 价格展示 | 直接返回展示文案，避免前端拼货币和免费规则。 |
| `cityName`、`distanceText`、`memberText` | 后台 | 地址栏 | 前端可拼成 `📍静安区 · 3.2km · 5/8人`。 |
| `timeText` | 后台 | 时间栏 | 建议后台返回已格式化文案，前端只加图标。 |
| `joinedCount`、`joinedText` | 后台 | 已入局说明 | 后台可返回完整文案；前端卡片窄位可压缩显示为 `+5已入局`。 |
| `participantAvatars` | 后台 / CDN | 参与玩家头像 | 每个组局返回自己的头像列表，避免不同卡片头像相同。前端最多展示 3 个。 |
| `actionText` | 后台 | 主按钮 | 当前为 `加入`。 |
| `actions` | 前端固定或后台返回 | 分享、关注、引荐、打招呼 | 如果不同卡片能力一致，可前端固定；如果权限不同，后台返回 action key。 |
| `friendContextText` | 后台 | 朋友在玩上下文 | 可选，例如 `好友正在玩`、`3位好友参与`。 |
| `route` | 后台或前端拼接 | 跳转详情 | 建议前端基于 `id` 拼详情路径，后台可只返回 `id`。 |

前端派生规则：

1. `nearbySection.tabs` 由后台返回，前端只按 `key` 更新当前筛选态。
2. `friendSection.title`、`friendSection.count`、`friendSection.moreText` 由后台返回，前端不写死文案。
3. 标签颜色由 `type` 或 `typeText` 决定，不建议后台返回颜色值；除非运营需要后台配置主题色。
4. 地址栏可由 `cityName`、`distanceText`、`memberText` 拼接，也可后端直接返回 `locationText`。
5. 时间栏可由 `timeText` 直接展示；如果后端返回开始/结束时间戳，前端需要统一格式化。
6. `joinedText` 在后台可保持完整文案，卡片内前端按窄位压缩成 `+N已入局`。
7. 分享、关注、引荐、打招呼图标由前端固定，后台只需要返回 action key 或权限状态。

待确认：

1. 首页首屏是否定稿为 `GET /api/app/home` 聚合返回 `nearbySection`、`nearbyGames`、`friendSection`、`friendGames`。
2. 附近正在发生是否需要独立分页接口 `GET /api/app/games/nearby`。
3. 朋友在玩是否需要独立分页接口 `GET /api/app/games/friends`，以及“朋友”关系由哪个服务判定。
4. 封面字段命名使用 `coverUrl` 还是沿用当前前端测试字段 `coverSrc`。
5. 参与者头像字段命名使用 `participantAvatars`、`avatarUrls` 还是参与者对象数组。
6. 右侧操作项是否所有卡片固定一致，还是由后台按权限返回可用 action。

## 9. 玩家首页本周玩霸榜接口字段

记录日期：2026-06-18

模块：首页 / 玩家首页 / 本周玩霸榜

页面：`pages/home/player/index`

功能：玩家首页 `本周玩霸榜` 的榜单数据由后台接口返回，包括标题、角色 tab、当前默认榜单、当前用户排名、头像、昵称、榜单说明、经验 XP 和查看全部榜单文案。`玩家 / 行家 / 领路人` 的 key、展示文案、顺序和是否展示都由后台返回，前端只按接口结果渲染和点亮。当前 mock 仅用于本地占位，默认返回当前用户所属角色的榜单；没有对应角色榜单数据时，tab 点击只切换 active 样式，不硬造当前用户排名。

状态：候选字段，待后端确认。

推荐接口：

```text
GET /api/app/home
```

推荐响应结构：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "rankingSection": {
      "icon": "🏆",
      "title": "本周玩霸榜",
      "moreText": "查看全部榜单",
      "defaultTab": "player",
      "tabs": [
        { "key": "player", "name": "玩家" },
        { "key": "expert", "name": "行家" },
        { "key": "guide", "name": "领路人" }
      ]
    },
    "rankingBoards": {
      "player": {
        "list": [
          {
            "id": "rank_player_001",
            "userId": "user_001",
            "rank": 1,
            "nickname": "领域专家 PRO",
            "avatarUrl": "https://cdn.example.com/users/user_001-avatar.jpg",
            "desc": "本周组局 12 · MVP 5次",
            "xpText": "2,450 XP"
          }
        ],
        "myRank": {
          "rank": 52,
          "nickname": "我（Alex）",
          "avatarUrl": "https://cdn.example.com/users/me-avatar.jpg",
          "desc": "上周排名 65",
          "xpText": "520 XP"
        }
      }
    }
  },
  "requestId": "req_xxx"
}
```

字段分工：

| 字段 | 来源 | 用途 | 说明 |
| --- | --- | --- | --- |
| `rankingSection.icon` | 后台或前端固定 | 榜单标题图标 | 当前为奖杯图标。 |
| `rankingSection.title` | 后台 | 榜单标题 | 当前文案为 `本周玩霸榜`。 |
| `rankingSection.moreText` | 后台 | 查看全部榜单文案 | 当前文案为 `查看全部榜单`。 |
| `rankingSection.defaultTab` | 后台 | 默认点亮角色 tab | 取值应匹配 `rankingSection.tabs[].key`，通常为当前用户所属角色。 |
| `rankingSection.tabs` | 后台 | 榜单角色 tab | 后台返回 key、文案、顺序和是否展示；前端不固定写死 `玩家/行家/领路人`。 |
| `rankingSection.tabs[].key` | 后台 | 角色 tab key | 用于 active 状态、请求榜单和匹配 `rankingBoards`。 |
| `rankingSection.tabs[].name` | 后台 | 角色 tab 文案 | 例如 `玩家`、`行家`、`领路人`，以前端实际收到为准。 |
| `rankingBoards` | 后台 | 分角色排行榜数据 | 可返回默认角色榜单，也可预载多个角色榜单；前端按角色 key 读取。 |
| `rankingBoards.{role}.list` | 后台 | 当前角色榜单前三名 | 当前 UI 只展示前三名；后续完整榜单页可分页。 |
| `rankingBoards.{role}.myRank` | 后台 | 当前用户在该角色榜单中的排名 | 用户不属于该角色或没有排名时可返回 `null`；前端不能为其它角色硬造“我的排名”。 |
| `id` | 后台 | 榜单记录 ID | 用于埋点、跳转或后续详情。 |
| `userId` | 后台 | 用户 ID | 用于跳转用户主页或埋点。 |
| `rank` | 后台 | 排名 | 数字或 `01` 字符串均可；前端会把 1-9 格式化为两位。 |
| `nickname` / `displayName` | 后台 | 昵称 | 前端优先展示昵称字段。 |
| `avatarUrl` | 后台 / CDN | 玩家头像 | 前端优先展示图片；没有图片时用 `avatarFallback` 或昵称前两字兜底。 |
| `avatarFallback` | 后台或前端兜底 | 头像占位文字 | 可选，用于 mock 或头像为空场景。 |
| `desc` | 后台 | 榜单说明 | 建议后台直接返回展示文案，例如 `本周组局 12 · MVP 5次`。 |
| `weeklyGameCount`、`weeklyMvpCount` | 后台 | 榜单说明派生字段 | 可选；如果不返回 `desc`，前端可临时拼说明。 |
| `xpText` | 后台 | 经验展示 | 推荐直接返回 `2,450 XP`；前端会拆出数字和单位展示。 |
| `xp`、`xpUnit` | 后台 | 经验展示备用字段 | 如果不用 `xpText`，可返回数字和单位。 |

待确认：

1. `本周玩霸榜` 默认榜单是否继续放在 `GET /api/app/home` 聚合接口中，还是拆成独立榜单接口。
2. 如果拆独立接口，是否使用 `GET /api/app/rankings/weekly?roleType={rankingSection.tabs[].key}`；tab 的 key 必须来自后台返回。
3. 首页是否只需要返回前三名和当前用户排名，完整榜单是否另走分页接口。
4. 榜单说明由后台直接返回 `desc`，还是返回 `weeklyGameCount/weeklyMvpCount` 等结构字段由前端拼接。
5. 榜单 tab 是否需要支持后台调整顺序、改文案、隐藏某个 tab，或返回更多角色 tab。

## 10. 玩家首页我的成就字段

记录日期：2026-06-18

模块：首页 / 玩家首页 / 我的成就

页面：`pages/home/player/index`

功能：玩家首页进入时只请求首页聚合接口，不为 `我的成就` 额外重复请求。后台在首页聚合响应中返回成就标题、图标和成就列表；前端按 React 参考样式固定渲染，不从后台获取颜色、尺寸、圆角等样式。

状态：候选字段，待后端确认是否并入首页聚合接口。后台文档中已有 `GET /api/app/growth/me` 返回成长和成就的说明，但玩家首页为了减少请求次数，建议首页聚合接口同步返回用于当前模块展示的精简成就列表。

推荐接口：

```text
GET /api/app/home
```

推荐响应片段：

```json
{
  "achievementSection": {
    "icon": "💎",
    "title": "我的成就"
  },
  "achievementList": [
    {
      "id": "hundred",
      "code": "hundred_king",
      "title": "百场王者",
      "icon": "🏆",
      "statusText": "等级",
      "unlocked": true
    },
    {
      "id": "earth",
      "code": "earth_roamer",
      "title": "地球漫游者",
      "icon": "🌍",
      "statusText": "进度20%",
      "progressPercent": 20,
      "unlocked": true
    }
  ]
}
```

字段分工：

| 字段 | 来源 | 用途 | 说明 |
| --- | --- | --- | --- |
| `achievementSection.icon` | 后台或前端固定 | 成就标题图标 | 当前视觉为 `💎`。 |
| `achievementSection.title` | 后台 | 成就标题 | 当前文案为 `我的成就`。 |
| `achievementList` / `achievements` | 后台 | 首页展示的成就列表 | 当前 UI 展示 4 个；排序由后台返回。 |
| `id` | 后台 | 成就记录 ID | 用于跳转详情或埋点。 |
| `code` | 后台 | 成就编码 | 前端可用作业务类型识别，但不由后台直接下发样式。 |
| `title` / `name` | 后台 | 成就名称 | 例如 `百场王者`。 |
| `icon` | 后台或前端配置 | 成就图标 | 可返回 emoji 或后续扩展为图标 URL。 |
| `statusText` | 后台 | 成就状态文案 | 例如 `等级`、`未解锁`、`进度20%`，前端统一补 `▲` 展示。 |
| `progressPercent` | 后台 | 成就进度 | 可选；用于进度型成就。 |
| `unlocked` | 后台 | 是否已解锁 | 未解锁时前端按固定灰/虚线视觉展示。 |

注意：成就卡颜色、徽章边框、阴影、字号、圆角和横向排布由前端 WXSS 按 React 参考固定实现；后台不要返回颜色值、尺寸值或 CSS 类名。

## 11. 玩家首页元宇宙入口字段

记录日期：2026-06-18

模块：首页 / 玩家首页 / 元宇宙入口

页面：`pages/home/player/index`

功能：玩家首页进入时通过首页聚合接口返回元宇宙入口展示数据；前端按 React 参考样式固定渲染入口卡片，不为该模块额外请求接口。

状态：候选字段，待后端确认是否并入首页聚合接口。

推荐接口：

```text
GET /api/app/home
```

推荐响应片段：

```json
{
  "metaverseEntry": {
    "title": "进入元宇宙",
    "desc": "共创数字街区｜全球联机互动",
    "tags": ["3D空间", "NFT徽章"],
    "avatars": [
      { "avatarUrl": "https://cdn.example.com/users/a.png", "avatarFallback": "A" },
      { "avatarUrl": "https://cdn.example.com/users/l.png", "avatarFallback": "L" },
      { "avatarUrl": "https://cdn.example.com/users/m.png", "avatarFallback": "M" }
    ],
    "joinedCount": 99,
    "route": "pages/placeholder/metaverse/index"
  }
}
```

字段分工：

| 字段 | 来源 | 用途 | 说明 |
| --- | --- | --- | --- |
| `metaverseEntry.title` / `actionText` | 后台 | 入口主标题 | 当前文案为 `进入元宇宙`。 |
| `metaverseEntry.desc` | 后台 | 入口说明 | 当前文案为 `共创数字街区｜全球联机互动`。 |
| `metaverseEntry.tags` | 后台 | 标签列表 | 当前展示前 2 个，例如 `3D空间`、`NFT徽章`。 |
| `metaverseEntry.avatars` | 后台 | 参与用户头像 | 当前展示前 3 个；支持 `avatarUrl` 和 `avatarFallback`，mock 复用排行榜头像，正式由后台/CDN 返回。 |
| `metaverseEntry.joinedCount` / `onlineCount` / `participantCount` | 后台 | 头像组角标数字 | 前端格式化为 `+99`；当前 mock 为 `99`。 |
| `metaverseEntry.badgeText` / `badge` / `onlineText` | 后台可选 | 头像组角标兜底文案 | 若后台只能返回完整文案，前端可直接展示。 |
| `metaverseEntry.route` | 后台 | 点击跳转目标 | 当前仍是预留页，后续按真实元宇宙入口调整。 |

注意：卡片尺寸、蓝色渐变背景、标签颜色、头像叠放、右侧地球轨道图案等样式由前端固定实现；后台不要返回颜色、圆角、阴影或 CSS 类名。

## 12. 行家 / 领路人首页工作台字段

记录日期：2026-06-18

模块：首页 / 行家首页 / 领路人首页

页面：`pages/home/expert/index`、`pages/home/guide/index`、`components/role-dashboard-home/index.*`

功能：行家和领路人首页复用同一个工作台组件，通过 `roleType=expert|guide` 从首页聚合接口返回不同角色的标题、统计、快捷动作、待办卡和洞察面板。样式由前端固定，后台只返回内容、数字、排序和跳转目标。

状态：候选字段，待后端确认是否并入首页聚合接口。当前 mock 已按 `GET /api/app/home?roleType=expert` 和 `GET /api/app/home?roleType=guide` 返回不同工作台数据。

推荐接口：

```text
GET /api/app/home?roleType=expert
GET /api/app/home?roleType=guide
```

推荐响应片段：

```json
{
  "roleDashboard": {
    "roleName": "行家",
    "badge": "行家 Lv.20",
    "stateText": "今日待处理 7",
    "title": "行家工作台",
    "subtitle": "管理邀约与审核",
    "desc": "处理玩家入局申请，维护高质量局内体验",
    "icon": "🎯",
    "actionTitle": "快捷处理",
    "cardTitle": "待办提醒",
    "moreText": "查看全部",
    "stats": [
      { "value": "5", "label": "待审核" },
      { "value": "18", "label": "已交付" },
      { "value": "4.9", "label": "评分" }
    ],
    "actions": [
      { "id": "audit", "icon": "✅", "title": "审核列表", "desc": "处理玩家入局申请", "route": "pages/game/applications/index" }
    ],
    "cards": [
      { "id": "expert-card-1", "tag": "审核", "title": "玩家入局申请", "meta": "3 条新申请等待处理", "count": "3", "route": "pages/game/applications/index" }
    ],
    "insight": {
      "title": "本周服务质量",
      "desc": "交付及时率保持稳定，继续关注待审核申请。",
      "progress": 76,
      "metrics": ["及时率 96%", "好评 18", "待反馈 2"]
    }
  }
}
```

字段分工：

| 字段 | 来源 | 用途 | 说明 |
| --- | --- | --- | --- |
| `roleDashboard.roleName` / `badge` / `stateText` | 后台 | 角色身份和今日状态 | 行家、领路人展示不同文案。 |
| `roleDashboard.title` / `subtitle` / `desc` / `icon` | 后台 | 工作台主卡内容 | 不返回颜色、圆角、阴影等样式。 |
| `roleDashboard.stats` | 后台 | 三个核心指标 | 行家示例：待审核、已交付、评分；领路人示例：待接收、成功引荐、活跃城市。 |
| `roleDashboard.actions` | 后台 | 快捷入口 | 前端按返回顺序展示；`route` 目前只记录目标，正式跳转后再接。 |
| `roleDashboard.cards` | 后台 | 待办提醒卡 | 包含标签、标题、说明和数量角标。 |
| `roleDashboard.insight` | 后台 | 周维度状态面板 | `progress` 为 0-100，`metrics` 为标签列表。 |

注意：当前阶段只实现行家 / 领路人首页工作台骨架；右侧底部按钮除 `首页` 外暂不接真实跳转。

## 13. 首页入口角色分流字段

记录日期：2026-06-18

模块：首页 / 角色首页入口

页面：`pages/home/player/index`、`pages/home/expert/index`、`pages/home/guide/index`

功能：后续真实入口不应由前端固定猜测进入哪个角色首页，应由后台返回当前默认角色或已开通角色后决定进入玩家、行家或领路人首页。当前阶段只完成三个首页页面和对应 mock 数据，不在登录页或其它入口里扩展真实跳转。

状态：待后端确认。对外后台资料中 `users` 表已包含 `default_role` 字段，`GET /api/app/users/me` 已声明返回用户角色信息，但 `CurrentUserDTO` 中首页默认角色字段名仍待确认；本次没有改登录页跳转逻辑。

当前前端兼容字段：

```text
defaultRole / default_role / homeRole / home_role / currentRole / current_role / roleType / role_type
```

建议分流规则：

1. 优先使用后台明确返回的默认角色字段。
2. 若默认角色为空或未通过审核，则从 `roles` + `roleStatusMap` 中选择已通过角色。
3. 若后台未返回可用角色，则兜底进入玩家首页。
4. 角色枚举统一兼容 `player`、`expert`、`guide`；历史 `leader` 会归一为 `guide`。

待后端确认：

1. `GET /api/app/users/me` 是否正式返回 `defaultRole`，或沿用数据库字段 `default_role`。
2. `roles` 是字符串数组，还是 `{ roleType, status }` 对象数组。
3. `roleStatusMap` 的通过状态枚举是否固定为 `approved`，还是还有 `passed/active/enabled` 等值。
4. 用户同时拥有多个角色时，是否完全以后端 `defaultRole/default_role` 为准。

## 14. 行家首页最新评价数量字段

记录日期：2026-06-19

模块：首页 / 行家首页 / 最新评价

页面：`pages/home/expert/index`、`components/role-dashboard-home/index.*`

功能：行家首页“最新评价”模块只在首页展示最新一条评价，右侧展示全部评价数量，点击后进入全部评价列表页。

接口：

```text
GET /api/app/home?roleType=expert
```

建议响应片段：

```json
{
  "roleDashboard": {
    "review": {
      "title": "最新评价",
      "count": 156,
      "avatarUrl": "https://example.com/avatar.png",
      "playerLevel": "萌新玩家",
      "timeText": "2小时前",
      "rating": 4,
      "content": "DM非常专业，带本节奏很好，气氛拉满！"
    }
  }
}
```

状态：待后端确认。当前前端优先读取 `roleDashboard.review.count`，并兼容 `roleDashboard.review.total`、`roleDashboard.review.totalCount`、`roleDashboard.review.reviewCount`、`roleDashboard.reviewCount`、`roleDashboard.reviewTotal`、`roleDashboard.reviewTotalCount`。评价卡片字段来自后台：头像优先读取 `avatarUrl/avatar/userAvatar/playerAvatar`，玩家等级读取 `playerLevel/levelText/userLevel/authorLevel`，时间读取 `timeText/createdAtText/reviewTimeText/createdAt`，星级读取 `rating/score/star/stars`，内容读取 `content/comment/text/reviewContent`。首页只渲染最新一条评价；如后端返回 `review.items`、`review.list`、`review.reviews`、`review.latest` 或 `review.latestReview`，前端会取第一条/最新条展示。

## 15. 行家首页后半段复用玩家首页模块

记录日期：2026-06-19

模块：首页 / 行家首页 / 本周玩霸榜、我的成就、进入元宇宙

页面：`pages/home/expert/index`、`components/role-dashboard-home/index.*`

接口：

```text
GET /api/app/home?roleType=expert
```

状态：前端已复用玩家首页的榜单、成就和元宇宙模块结构。行家首页榜单默认高亮 `expert`，建议后台返回 `rankingSection.defaultTab: "expert"`，并在 `rankingBoards.expert` 下返回行家榜单列表和当前用户排名。`achievementSection`、`achievementList`、`metaverseEntry` 沿用玩家首页字段结构即可。

## 16. 领路人首页关系网络字段

记录日期：2026-06-19

模块：首页 / 领路人首页 / 我的关系网络

页面：`pages/home/guide/index`、`components/role-dashboard-home/index.*`

接口：

```text
GET /api/app/home?roleType=guide
```

功能：领路人首页“我的关系网络”模块由首页聚合接口返回展示文案、中心节点、连接数量、收益、区域、指标卡和操作按钮。关系图连线、节点位置、颜色、圆角和卡片样式由前端固定实现，后台只返回内容、数字、排序和跳转目标。

建议响应片段：

```json
{
  "roleDashboard": {
    "network": {
      "title": "我的关系网络",
      "status": "实时连接中",
      "hubTitle": "萧飒",
      "hubDesc": "领路人",
      "connectedCount": 156,
      "summary": "● 已连接 156 位玩家",
      "income": "本周收益 ¥1,240",
      "location": "📍 镇海区",
      "items": [
        { "id": "players", "icon": "●", "name": "已连接", "desc": "156 位玩家" },
        { "id": "income", "icon": "¥", "name": "本周收益", "desc": "¥1,240" },
        { "id": "area", "icon": "📍", "name": "核心区域", "desc": "镇海区" }
      ],
      "buttons": [
        { "text": "查看全部", "primary": true, "route": "pages/profile/index" },
        { "text": "管理连接", "route": "pages/message/index" }
      ]
    }
  }
}
```

状态：待后端确认。当前前端已兼容 `summary/summaryText`、`connectedCount/playerCount`、`hubTitle/centerTitle`、`hubDesc/centerDesc`、`actionText/moreText`、`secondaryText` 等字段；若后端暂未返回 `items/buttons`，前端会用连接数、收益和区域生成兜底指标卡与按钮。

## 17. 申请行家条件预检字段

记录日期：2026-06-20

模块：首页 / 申请行家

页面：`pages/home/index` 的 `申请行家浏览页`

功能：申请行家浏览页的申请条件列表需要由后台返回逐项状态。左侧圆圈状态由后台字段控制：满足时显示勾选圆，不满足、未提交或未完成时显示空圆。当前前端静态走查先写死 `checked` 字段，生产联调时替换为后台返回。

现有资料核对：项目已有 `GET /api/app/role-applications/my`、`POST /api/app/role-applications` 和 `GET /api/app/roles/my`，旧角色申请页只使用 `qualification.conditionMet`、`paymentMet` 这类粗粒度字段；未找到行家申请浏览页“条件逐项状态 + 行家计划书状态”的正式接口。

候选接口：

```text
GET /api/app/role-applications/expert/precheck
```

也可并入：

```text
GET /api/app/role-applications/my
```

建议响应片段：

```json
{
  "roleType": "expert",
  "requirements": [
    { "key": "level", "title": "玩家等级达到 Lv.20", "status": "当前等级: Lv.21 / 已满足", "checked": true },
    { "key": "realname", "title": "完成实名认证", "status": "认证状态: 已通过 / 已满足", "checked": true },
    { "key": "enterprise", "title": "完成企业认证", "status": "认证状态: 已通过 / 已满足", "checked": true },
    { "key": "games", "title": "发起过 5次以上组局", "status": "当前: 5 次 / 已满足", "checked": true },
    { "key": "credit", "title": "信用分 ≥ 90 分", "status": "当前: 92 分 / 已满足", "checked": true },
    { "key": "member", "title": "会员等级≥ 高级会员", "status": "当前: 高级会员 / 已满足", "checked": true }
  ],
  "planTask": {
    "key": "expertPlan",
    "title": "提交行家计划书",
    "desc": "需描述你的资源、能力和项目说明书",
    "action": "去填写 ›",
    "checked": false,
    "route": "pages/role/apply/index?roleType=expert&step=plan"
  }
}
```

待确认：

1. 条件预检是新增独立接口，还是扩展 `GET /api/app/role-applications/my`。
2. 逐项状态字段是否统一使用 `checked`，或使用 `met/status/completed`。
3. 行家计划书是否作为第 7 个 requirement 返回，还是单独返回 `planTask`。
4. `去填写` 当前前端按用户要求不接跳转；生产需要后台确认目标 `route`。

## 18. 申请行家表单配置接口

记录日期：2026-06-20

模块：首页 / 申请行家

页面：`pages/home/index` 的 `申请行家内页`

接口：

```text
GET /api/app/role-applications/expert/config
```

功能：申请行家内页的限制条件、已有技能领域、字段提示内容、技能标签规则、作品类型和作品数量由后台配置返回；如果后台暂未提供或请求失败，前端使用当前静态默认配置兜底。

建议响应片段：

```json
{
  "skillOptions": [
    { "name": "摄影", "active": true },
    { "name": "户外" },
    { "name": "美食" },
    { "name": "文化" },
    { "name": "手工" },
    { "name": "运动" },
    { "name": "音乐" },
    { "name": "+ 自定义", "custom": true }
  ],
  "fields": [
    { "type": "chips", "key": "skillDomain", "label": "选择技能领域", "required": true },
    {
      "type": "input",
      "key": "skillTags",
      "label": "技能标签",
      "required": true,
      "placeholder": "如：人像摄影、风光摄影、夜景拍摄",
      "helper": "添加具体标签，让用户更容易找到你",
      "maxlength": 30
    },
    { "type": "select", "key": "experienceYears", "label": "从业年限", "required": true, "placeholder": "请选择从业年限" },
    {
      "type": "textarea",
      "key": "intro",
      "label": "个人简介",
      "required": true,
      "placeholder": "介绍你的专业背景、服务风格、擅长领域...",
      "helper": "不少于 50 字，突出你的专业优势",
      "maxlength": 300
    }
  ],
  "uploadField": {
    "label": "资质证明",
    "required": true,
    "icon": "📎",
    "title": "点击上传作品集及凭证",
    "acceptTypes": ["JPG", "PNG", "PDF"],
    "maxCount": 5
  },
  "validationRules": {
    "skillTags": { "minLength": 2, "maxLength": 30 },
    "intro": { "minLength": 50, "maxLength": 300 },
    "serviceName": { "minLength": 2, "maxLength": 20 },
    "customSkill": { "minLength": 2, "maxLength": 8 },
    "money": { "integerMaxLength": 8, "decimalMaxLength": 2 }
  },
  "yearOptions": ["1年", "2年", "3年", "30年以上"],
  "serviceCount": 3,
  "priceHint": "平台将收取 10% 服务费"
}
```

状态：待后端确认。当前前端已接入该候选接口，并在 mock 和接口失败时使用本地默认配置。

## 19. 申请行家表单提交字段

记录日期：2026-06-20

模块：首页 / 申请行家

页面：`pages/home/index` 的 `申请行家内页`

现有接口：

```text
POST /api/app/role-applications
```

现有资料核对：项目已有该接口，旧角色申请页只明确传入 `{ "roleType": "expert" }` 或 `{ "roleType": "guide" }`。未在当前项目文档中找到“行家申请内页”完整字段表。当前前端先按配置 key 和业务语义提交完整表单，后端字段名需要确认。

当前前端提交 payload：

```json
{
  "roleType": "expert",
  "skillDomain": "摄影",
  "skillDomains": ["摄影"],
  "skillTags": "人像摄影、风光摄影、夜景拍摄",
  "experienceYears": "3年",
  "intro": "不少于 50 字的个人简介内容",
  "personalIntro": "不少于 50 字的个人简介内容",
  "uploadFiles": [
    {
      "name": "作品集.pdf",
      "fileName": "作品集.pdf",
      "path": "wxfile://tmp_xxx",
      "tempFilePath": "wxfile://tmp_xxx",
      "size": 102400,
      "type": "file",
      "extension": "PDF"
    }
  ],
  "qualifications": [
    {
      "name": "作品集.pdf",
      "fileName": "作品集.pdf",
      "path": "wxfile://tmp_xxx",
      "tempFilePath": "wxfile://tmp_xxx",
      "size": 102400,
      "type": "file",
      "extension": "PDF"
    }
  ],
  "services": [
    {
      "name": "摄影陪拍",
      "serviceName": "摄影陪拍",
      "price": "299.00",
      "hourlyPrice": "299.00",
      "cost": "80.00"
    }
  ]
}
```

待确认：

1. `POST /api/app/role-applications` 是否直接接收文件临时路径/文件元信息，还是需要先上传文件并提交文件 ID/URL。
2. 技能领域字段使用单值 `skillDomain` 还是数组 `skillDomains`。
3. 个人简介字段使用 `intro` 还是 `personalIntro`。
4. 服务定价字段使用 `price` 还是 `hourlyPrice`，金额类型是字符串还是 number。
5. 资质文件字段使用 `uploadFiles` 还是 `qualifications`。


## 20. 组局申请加入页微信授权信息字段

记录日期：2026-06-21

模块：组局 / 入局申请

页面：`pages/game/apply/index`

功能：申请加入页顶部“微信信息”卡片需要展示用户的微信授权头像和微信授权昵称。该卡片文案当前为“微信信息 / 昵称：张伟 · 头像已同步”，生产环境不应简单假设等同于小程序平台个人资料昵称和头像。

现有资料核对：

- `GET /api/app/users/me` 已返回 `CurrentUserDTO`，包含用户资料类字段。
- 登录/注册接口中有 `nickname`、`avatarUrl` 入参和 `user.nickname`、`user.avatarUrl` 返回字段。
- 当前文档未明确 `CurrentUserDTO.nickname/avatarUrl` 是“微信授权信息”，还是“平台个人资料信息”。若用户在小程序内修改昵称/头像，这两个来源可能不一致。
- 入局申请最终口径为 `POST /api/app/games/{gameId}/apply`，入参目前为 `applyReason`、`fromGuideId`，与微信信息卡片预填数据无直接字段关系。

建议后端明确一种方案：

```text
GET /api/app/users/me
```

在 `CurrentUserDTO` 中补充分离字段：

```json
{
  "nickname": "平台昵称",
  "avatarUrl": "平台头像",
  "wechatProfile": {
    "nickname": "微信授权昵称",
    "avatarUrl": "微信授权头像",
    "synced": true,
    "authorizedAt": "2026-06-21T12:00:00+08:00"
  }
}
```

或新增独立接口：

```text
GET /api/app/users/me/wechat-profile
```

建议返回：

```json
{
  "nickname": "微信授权昵称",
  "avatarUrl": "微信授权头像",
  "synced": true,
  "authorizedAt": "2026-06-21T12:00:00+08:00"
}
```

前端接入建议：

- 申请加入页“微信信息”卡片优先展示微信授权信息：`wechatProfile.nickname`、`wechatProfile.avatarUrl`。
- 如果微信授权信息为空，展示授权缺失态或引导同步，不自动用平台资料冒充微信授权信息，除非产品明确允许兜底。
- 当前静态开发阶段仍使用写死默认值，后续联调时替换为后端返回字段。

待后端 / 产品确认：

1. `/api/app/users/me` 中的 `nickname/avatarUrl` 是否表示微信授权信息，还是平台个人资料信息。
2. 是否需要把微信授权信息和平台个人资料信息分开返回。
3. 微信头像昵称同步状态字段命名使用 `synced`、`wechatSynced` 还是 `avatarSynced`。
4. 若用户未授权或授权过期，申请加入页应展示默认头像、平台资料兜底，还是弹出微信授权引导。

## 21. 组局发起邀请页预算上限字段
记录日期：2026-06-21

模块：组局 / 发起邀请

页面：`pages/game/invite/index`

功能：发起邀请页“预算与奖励”中的费用预算为人民币元整数金额。前端当前通过页面变量 `budgetMaxAmount` 控制最大可输入金额，默认值为 `99999999` 元；后续如果后台返回预算上限，则输入值必须小于等于后台返回的上限。平台服务费、系统级领路人奖励、邀请奖励的分成比例后续也应由后台返回，前端根据比例计算行家实收金额。

建议后端在发起邀请页配置或局邀请初始化接口中返回：

```json
{
  "budgetMaxAmount": 99999999,
  "rewardRateConfig": {
    "platformServiceRate": 10,
    "systemGuideRewardRate": 10,
    "inviteRewardRate": 40
  }
}
```

前端接入口径：

- 若后台返回 `budgetMaxAmount` 且为正整数，前端使用该值作为预算输入上限。
- 若后台未返回、返回为空或格式非法，前端使用默认上限 `99999999`。
- 预算输入只允许整数人民币元，不支持小数。
- 若后台返回 `rewardRateConfig`，前端按比例计算平台服务费、系统级领路人奖励、邀请奖励和行家实收金额。
- 比例默认值为：平台服务费 10%、系统级领路人奖励 10%、邀请奖励 40%；比例总和不能超过 100%。
- 行家实收金额 = 费用预算 - 平台服务费 - 系统级领路人奖励 - 邀请奖励。

待后端 / 产品确认：

1. 预算上限字段是否命名为 `budgetMaxAmount`。
2. 该字段由发起邀请页配置接口返回，还是由局详情 / 邀请初始化接口返回。
3. 正式业务预算最大值是否仍为 `99999999` 元，还是需要按局类型、用户身份或会员等级动态返回。
4. 分成比例字段是否命名为 `rewardRateConfig`，以及比例值使用百分数 `10` 还是小数 `0.1`。
5. 各项金额是否按当前前端的四舍五入口径取整，还是由后台直接返回最终金额。

## 22. 组局发起邀请页玩家选择规则与列表接口

记录日期：2026-06-22

模块：组局 / 发起邀请

页面：`pages/game/invite/index` 第 2 步 `选择玩家`

功能：进入发起邀请第 2 步时，页面需要从后台获取可邀请玩家数量规则，并加载最近联系玩家列表；点击“添加新的玩家”时打开玩家列表，展示形式与最近联系一致。当前按用户要求先以最多 1 位玩家处理，最少 1 位固定生效，后续最多可添加人数待产品确认后由后台配置返回。

候选接口：

```text
GET /api/app/game-invites/player-config
GET /api/app/game-invites/recent-players
GET /api/app/game-invites/players
```

建议 `GET /api/app/game-invites/player-config` 返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "minPlayerCount": 1,
    "maxPlayerCount": 1
  },
  "requestId": "req_xxx"
}
```

建议玩家列表返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "list": [
      {
        "id": "user_001",
        "name": "李明",
        "avatarUrl": "https://cdn.example.com/avatar.png",
        "avatarText": "LM",
        "tag": "需求匹配",
        "desc": "某互联网公司 · 产品总监",
        "meta": "预算: ¥500-1000 | 时间: 本周"
      }
    ],
    "total": 1
  },
  "requestId": "req_xxx"
}
```

当前前端接入口径：

- `minPlayerCount` 兜底为 `1`，确认发起前必须至少选中 1 位玩家。
- `maxPlayerCount` 当前 mock 与页面兜底均为 `1`，因此选择新玩家时会替换当前玩家。
- 页面展示 `已添加 X/N 位玩家`，其中 `N` 来自 `maxPlayerCount`。
- `GET /api/app/game-invites/recent-players` 用于第 2 步最近联系列表，可接受 `keyword` 搜索参数。
- `GET /api/app/game-invites/players` 用于“添加新的玩家”弹层，可接受 `keyword` 搜索参数。

待后端 / 产品确认：

1. 最多可添加玩家数是否仍为 1，还是后续按局类型、角色、会员等级或后台配置动态返回。
2. 选择玩家是否只允许单选；如果 `maxPlayerCount > 1`，前端需要改为多选和批量确认。
3. 玩家列表字段命名使用 `id/name/avatarUrl/tag/desc/meta`，还是沿用用户资料 DTO 字段。
4. 最近联系列表排序依据是最近沟通时间、合作次数，还是后台推荐分。
5. “添加新的玩家”列表是否需要分页字段 `page/pageSize/hasMore`。

## 23. 领路人组局进度查询页列表接口

记录日期：2026-06-22

模块：组局 / 领路人组局进度查询

页面：`pages/game/guide-progress/index`

功能：进入“组局消息”页时从后台获取进行中组局数量、进行中卡片和最近完成卡片。当前页面只按 1 个玩家对 1 个行家展示，卡片左侧为玩家，右侧为行家；卡片外框颜色按确认状态渲染：已有任一方确认时使用橙色外框，双方均未确认时使用蓝色外框。多玩家 / 多行家的展示形式另记在 `docs/follow-ups.md`，待甲方补充。

候选接口：

```text
GET /api/app/game-invites/guide-progress
```

建议返回：

```json
{
  "activeCount": 2,
  "activeParties": [
    {
      "id": "invite-progress-001",
      "status": "waiting_expert",
      "statusText": "进行中",
      "timeText": "剩余23小时",
      "progressText": "等待行家确认",
      "progressPercent": 49,
      "noticeText": "行家尚未查看邀请，可发送提醒",
      "players": [
        {
          "id": "player_001",
          "name": "李娜",
          "avatarUrl": "",
          "avatarText": "LN",
          "confirmStatus": "confirmed",
          "statusText": "已确认"
        }
      ],
      "experts": [
        {
          "id": "expert_001",
          "name": "王强",
          "avatarUrl": "",
          "avatarText": "WQ",
          "confirmStatus": "pending",
          "statusText": "待确认"
        }
      ]
    }
  ],
  "completedParties": [
    {
      "id": "invite-complete-001",
      "resultStatus": "success",
      "title": "组局成功",
      "timeText": "昨天",
      "summaryPrefix": "你引荐的",
      "completedMemberText": "张伟 与 李娜",
      "players": [{ "id": "player_001", "name": "张伟" }],
      "experts": [{ "id": "expert_001", "name": "李娜" }],
      "summarySuffix": "已成功组局",
      "rewardText": "+50积分",
      "gameTitle": "产品经理交流会"
    },
    {
      "id": "invite-complete-002",
      "resultStatus": "canceled",
      "title": "组局已取消",
      "timeText": "3天前",
      "players": [{ "id": "player_002", "name": "王芳" }],
      "rejectName": "王芳",
      "rejectRoleText": "玩家",
      "rejectText": "婉拒了组局邀请",
      "reasonText": "原因：时间冲突"
    }
  ]
}
```

当前前端接入口径：

- `activeCount` 用于标题 `进行中的组局 X`，进入页面时请求接口获取。
- `activeParties[].players[0]` 展示在卡片左侧，`experts[0]` 展示在右侧。
- `confirmStatus/statusText` 用于每个人的确认状态展示。
- `progressPercent/progressText` 由后台返回；后台未返回百分比时前端按已确认人数做兜底推算。
- `noticeText` 当前只展示提示，不在列表页发送提醒；提醒功能后续放到领路人查看组局详情页处理。
- `completedParties` 作为最近完成卡片列表渲染，页面只展示后台给出的卡片信息。
- 成功卡展示 `title`、`completedMemberText`、`summaryPrefix`、`summarySuffix`、`rewardText`、`gameTitle` 和 `timeText`。
- 取消卡展示 `title`、`rejectName`、`rejectRoleText`、`rejectText`、`reasonText` 和 `timeText`。

待后端 / 产品确认：

1. 接口路径是否使用 `/api/app/game-invites/guide-progress`。
2. 状态枚举是否使用 `waiting_expert`、`waiting_all`、`success`、`canceled`。
3. 进行中数量是否由 `activeCount` 返回，还是以前端按 `activeParties.length` 兜底即可。
4. 进入详情页需要的详情页 ID 使用 `id`、`invitationId` 还是独立 `gameInviteId`。
5. 列表页是否需要分页；如果需要，补充 `page/pageSize/hasMore` 字段。

## 24. 组局支付页微信支付下单接口

记录日期：2026-06-22

模块：组局 / 支付

页面：`pages/game/payment/index`

功能：用户在组局支付页点击 `微信支付` 时，前端需要先校验是否勾选 `我已了解并同意押金局规则`。未勾选时前端弹窗提示勾选；已勾选时，前端把支付金额和分成明细提交后台，由后台创建微信支付订单并返回 `wx.requestPayment` 所需参数。

现有资料核对：

- 当前项目代码和工程文档中未发现已定稿的组局微信支付接口。
- `docs/architecture.md` 目前记录“一期免费局为主，支付相关页面只做状态和二期预留”。
- 因此本接口为候选接口，待后端确认正式路径、字段、金额口径、分成口径和支付完成后的跳转规则。

候选接口：

```text
POST /api/app/game-payments/wechat
```

当前前端提交 payload：

```json
{
  "gameId": "game_001",
  "scene": "deposit_game",
  "payChannel": "wechat",
  "amount": 100,
  "currency": "CNY",
  "agreementChecked": true,
  "splits": [
    { "key": "serviceFee", "label": "服务费", "amount": 10, "desc": "" },
    { "key": "platformServiceFee", "label": "平台服务费", "amount": 2.5, "desc": "" },
    { "key": "inviterReward", "label": "邀请人奖励", "amount": 5, "desc": "" },
    { "key": "partnerReward", "label": "合伙人奖励", "amount": 2.5, "desc": "" },
    { "key": "depositPool", "label": "押金池", "amount": 90, "desc": "完成任务后返还" }
  ]
}
```

建议返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "paymentOrderId": "pay_001",
    "gameId": "game_001",
    "amount": 100,
    "currency": "CNY",
    "paymentParams": {
      "timeStamp": "1719000000",
      "nonceStr": "nonce_xxx",
      "package": "prepay_id=wx_xxx",
      "signType": "RSA",
      "paySign": "sign_xxx"
    }
  },
  "requestId": "req_xxx"
}
```

当前前端接入口径：

- 未勾选同意规则时，不请求后台，直接弹窗提示用户勾选。
- 已勾选时，请求 `POST /api/app/game-payments/wechat`。
- 后台返回 `paymentParams` 后，前端调用 `wx.requestPayment`。
- 微信支付面板中用户取消时，前端保留在当前支付页并提示 `已取消支付`，方便重新发起支付。
- 页面底部 `取消` 按钮：若有上一页则 `navigateBack`；若编译模式直开且有 `gameId`，则跳转到 `pages/game/detail/index?gameId=...`；若没有 `gameId`，则跳转到 `pages/game/hall/index`。
- 当前 mock 环境返回 `mockPayment: true`，只表示支付请求已提交，不调用真实 `wx.requestPayment`。

待后端 / 产品确认：

1. 接口路径是否使用 `/api/app/game-payments/wechat`，还是放在 `/api/app/games/{gameId}/payments`。
2. 金额单位使用元、分还是 decimal 字符串；当前前端展示和提交为人民币元。
3. 分成明细是否允许前端提交，还是应由后台根据局模板、角色关系和业务规则重新计算并返回。
4. `splits` 中 `serviceFee/platformServiceFee/inviterReward/partnerReward/depositPool` 的 key 是否符合后端命名。
5. 支付成功后应跳转到组局成功页、组局详情页，还是保留当前页等待订单状态轮询。
6. 是否需要支付结果查询接口，例如 `GET /api/app/game-payments/{paymentOrderId}`。
7. 微信支付参数字段是否统一包在 `paymentParams` 下，签名算法 `signType` 使用 `RSA` 还是 `MD5`。

## 25. 我的组局管理页统计与卡片列表接口

记录日期：2026-06-22

模块：组局 / 我的组局管理

页面：`pages/game/manage/index`

功能：页面顶部紫色统计卡中的 `本月服务支出`、`进行中 N单`、`已完成 N单`、`已取消 N单`，下方 tab 数量，以及 tab 下的组局卡片列表需要由后台返回，不能由前端写死。

候选接口：

```text
GET /api/app/games/my/manage
```

建议返回：

```json
{
  "summary": {
    "label": "本月服务支出",
    "amount": 3200,
    "amountText": "¥3,200",
    "activeCount": 1,
    "completedCount": 4,
    "canceledCount": 1
  },
  "orders": [
    {
      "id": "manage-active-001",
      "statusType": "active",
      "statusText": "服务进行中",
      "ref": "REF-20260320-001",
      "avatarText": "ZH",
      "expertName": "张专家",
      "serviceText": "产品架构咨询 · ¥800",
      "guideName": "王引荐",
      "progressPercent": 60,
      "progressText": "60%",
      "deliveryText": "预计交付：2026-03-25 14:00",
      "elapsedText": "已进行 6/15 天",
      "remainingText": "剩余 9 天",
      "noticeText": "取消需赔付一定比例金额给行家"
    },
    {
      "id": "manage-complete-001",
      "statusType": "complete",
      "statusText": "已完成",
      "ref": "REF-20260318-002",
      "avatarText": "WM",
      "expertName": "王导师",
      "serviceText": "品牌定位咨询 · ¥1,200",
      "completeSummary": "服务已完成",
      "amountText": "¥1,200",
      "completedAtText": "完成时间：2026-03-19 18:30",
      "resultText": "已完成验收，可查看服务记录"
    },
    {
      "id": "manage-canceled-001",
      "statusType": "canceled",
      "statusText": "已取消（已赔付）",
      "ref": "REF-20260310-003",
      "avatarText": "LI",
      "expertName": "刘设计师",
      "serviceText": "UI设计服务",
      "reasonSummary": "我主动取消 · 赔付15%",
      "compensationAmountText": "¥120",
      "reasonText": "取消原因：需求变更，不再需要服务"
    }
  ]
}
```

当前前端接入口径：

- 页面进入时请求 `GET /api/app/games/my/manage`。
- `summary.label` 展示统计标题，默认兜底为 `本月服务支出`。
- 金额优先使用 `amountText`；若只返回 `amount/serviceExpense/monthlyServiceExpense` 数字，前端格式化为 `¥3,200`。
- 数量优先使用 `activeCount/completedCount/canceledCount`；前端也兼容 `ongoingCount/processingCount`、`completeCount/doneCount`、`cancelledCount/refundCount/refundCancelCount`。
- 统计卡底部和 tab 数量共用同一组后台计数。
- `orders` 用于渲染下方组局卡片列表，当前支持三种 `statusType`：`active`（服务进行中）、`complete`（已完成）、`canceled`（取消 / 退款）。
- tab 切换时前端按 `statusType` 过滤：`active` 进入 `进行中`，`complete` 进入 `已完成`，`canceled` 进入 `退款/取消`。
- 当前 mock 已按上述候选接口返回三种状态静态卡片，正式字段待后端确认后收敛。

待后端 / 产品确认：

1. 接口路径是否使用 `/api/app/games/my/manage`。
2. 金额单位使用元、分还是 decimal 字符串。
3. 金额字段以后端直接返回 `amountText` 为准，还是前端按数值格式化。
4. `已取消` 是否包含所有退款 / 取消单，还是仅包含已赔付取消单。
5. 页面卡片列表是否由同一接口返回，还是拆分为统计接口和列表接口。
6. 卡片状态枚举是否使用 `active/complete/canceled`，还是使用后端现有订单状态枚举。
7. 已完成卡片是否需要展示评价入口、服务记录入口或再次邀约入口。

## 26. 组局成功行家提示页详情接口

记录日期：2026-06-22

模块：组局 / 成功行家提示

页面：`pages/game/success-expert/index`

功能：组局成功后展示三方连接群、活动信息、资金状态和下一步。当前页面仍按静态设计稿还原，正式联调时这些数据必须由后台返回，不能由前端写死。

候选接口：

```text
GET /api/app/games/{gameId}/success-expert
```

建议返回：

```json
{
  "group": {
    "id": "group_001",
    "title": "产品架构咨询 - 三方群",
    "rolesText": "行家、领路人、玩家",
    "online": true,
    "memberCount": 3,
    "hintText": "领路人王引荐将持续跟进活动进度，确保双方顺利对接"
  },
  "viewer": {
    "role": "expert",
    "roleText": "行家"
  },
  "activity": {
    "ref": "REF-20260323-001",
    "createdAtText": "2026-03-23 10:23",
    "successAtText": "2026-03-23 14:30",
    "stageText": "待交付服务"
  },
  "fund": {
    "title": "资金已托管",
    "desc": "服务完成后自动结算",
    "amountText": "¥800",
    "status": "escrowed",
    "progressPercent": 33,
    "steps": ["已托管", "服务中", "已完成"]
  },
  "nextSteps": [
    { "index": 1, "title": "联系玩家确认具体时间", "desc": "建议24小时内完成", "active": true },
    { "index": 2, "title": "按时交付服务", "desc": "等待确认时间" },
    { "index": 3, "title": "确认完成并收款", "desc": "等待服务完成" }
  ]
}
```

当前前端接入口径：

- `group` 用于三方连接群卡片，群聊入口后续需要携带 `group.id`、`gameId` 和 `viewer.role`。
- `viewer.role` 表示当前点击者在此组局中的角色，例如 `expert/guide/player`，进入群聊页时需要传递。
- `activity` 用于活动信息卡片，包含活动编号、创建时间、组局成功时间和当前阶段。
- `fund` 用于资金状态卡片，包含托管状态、金额、进度和步骤文案。
- `nextSteps` 用于下一步列表，是否高亮由后台返回。
- 当前静态页仍用本地数据占位，后续联调时应替换为接口数据。
- 前端实现建议先保持页面级数据驱动，不优先抽组件：新增 `successDetail` 和 `normalizeSuccessDetail(data)`，把接口数据归一化为当前页面结构；待页面稳定或其他页面复用后，再考虑抽 `fund-status-card`、`next-steps-card` 等组件。

待后端 / 产品确认：

1. 接口路径是否使用 `/api/app/games/{gameId}/success-expert`，还是复用组局详情接口扩展成功态字段。
2. 活动时间字段是否返回格式化文案，还是返回时间戳由前端格式化。
3. 资金金额单位使用元、分还是后端直接返回 `amountText`。
4. 资金状态枚举是否使用 `escrowed/serving/completed`。
5. 三方群是否已有独立群 ID，进入局内群聊时需要传哪些参数。
6. 当前用户在组局中的角色字段使用 `viewer.role`、`currentRole` 还是成员列表中的角色推导。
7. 下一步列表是否由后台返回，还是前端按状态枚举本地生成。

## 27. 组局成功领路人提示页后续跟进接口

记录日期：2026-06-22

模块：组局 / 成功领路人提示

页面：`pages/game/success-guide/index`

功能：领路人在组局成功后可进行三类后续跟进：查看局日程、联系玩家和行家询问双方反馈、记录或促成合作交易。当前页面仍为静态 UI，三个入口的真实接口和跳转逻辑待后端 / 产品确认。

候选接口：

```text
GET /api/app/games/{gameId}/schedule
POST /api/app/games/{gameId}/feedback-contact
POST /api/app/games/{gameId}/deal-conversions
```

当前前端接入口径：

- `查看组局日程`：后续点击后应查看该局日程，需确认是跳转到已有局详情 / 日程页，还是调用独立日程接口后在当前页弹层展示。
- `询问双方反馈`：后续点击后应弹出联系框，用于联系玩家和行家聊天咨询反馈；需确认联系框样式、是否一次展示双方、是否进入一对一聊天或群聊。
- `促成交易`：后续点击后用于记录或推进双方达成合作交易；需确认是提交交易达成状态、进入交易表单，还是跳转到合作 / 订单页面。
- 三个入口都需要携带 `gameId`、领路人身份、玩家 ID、行家 ID 等参数，具体字段待接口确认。
- 当前页面只保留静态按钮和待接入提示，不接真实业务逻辑。

待后端 / 产品确认：

1. 查看局日程是否已有正式页面或接口，接口路径是否使用 `/api/app/games/{gameId}/schedule`。
2. 询问双方反馈的联系框由前端本地弹出，还是由后端返回可联系对象、IM 会话 ID 和默认话术。
3. 玩家与行家的反馈是否需要分别记录，是否需要回传反馈状态或积分奖励状态。
4. 促成交易是否创建交易记录 / 合作记录，接口路径、请求字段和状态枚举待确认。
5. 三个入口是否需要积分奖励发放规则，以及奖励是否由接口返回。

## 28. 专家业务管理页进行中服务操作接口

记录日期：2026-06-22

模块：组局 / 我的业务管理 - 专家

页面：`pages/game/manage/index`

功能：专家角色进入“我的业务管理”后，进行中服务卡片的 4 个按钮后续需要接真实业务逻辑。当前页面只做静态 UI 和待接入提示，不实现真实状态变更、赔付取消或发消息。

涉及按钮：

```text
提前结束交付
取消并赔付
联系玩家
联系领路人
```

候选接口：

```text
POST /api/app/game-services/{serviceOrderId}/finish-delivery
POST /api/app/game-services/{serviceOrderId}/cancel-with-compensation
POST /api/app/im/conversations
POST /api/app/im/messages
```

建议 `GET /api/app/games/my/manage` 中补充的进行中服务字段：

```json
{
  "orders": [
    {
      "id": "business-active-001",
      "serviceOrderId": "service_order_001",
      "gameId": "game_001",
      "statusType": "active",
      "statusText": "服务进行中",
      "player": {
        "id": "player_001",
        "name": "李明",
        "avatarUrl": "",
        "avatarText": "LI"
      },
      "guide": {
        "id": "guide_001",
        "name": "王引荐",
        "avatarUrl": ""
      },
      "service": {
        "title": "产品架构咨询",
        "amountText": "¥800",
        "expectedDeliveryText": "预计交付：03-25"
      },
      "timeline": [
        { "id": "group-success", "title": "组局成功", "time": "03-20 14:30", "state": "done" },
        { "id": "service-active", "title": "服务进行中", "time": "预计交付：03-25", "state": "current" },
        { "id": "waiting-confirm", "title": "等待确认完成", "state": "future" }
      ],
      "actions": {
        "canFinishDelivery": true,
        "canCancelWithCompensation": true,
        "canContactPlayer": true,
        "canContactGuide": true
      }
    }
  ]
}
```

### 提前结束交付

交互口径：

- 点击 `提前结束交付` 后，前端弹出确认交付按钮 / 确认弹窗。
- 用户确认后调用接口改变服务状态。
- 状态变化后刷新当前卡片，或由接口直接返回新的 `statusType/statusText/timeline/actions`。

候选路径：

```text
POST /api/app/game-services/{serviceOrderId}/finish-delivery
```

建议入参：

```json
{
  "gameId": "game_001",
  "confirm": true
}
```

建议返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "serviceOrderId": "service_order_001",
    "statusType": "complete",
    "statusText": "已提前交付",
    "timeline": [
      { "id": "group-success", "title": "组局成功", "time": "03-20 14:30", "state": "done" },
      { "id": "service-finished", "title": "服务已交付", "time": "03-23 16:00", "state": "done" },
      { "id": "waiting-confirm", "title": "等待确认完成", "state": "current" }
    ]
  },
  "requestId": "req_xxx"
}
```

待确认：

1. 提前结束交付后状态应进入 `已提前交付`、`等待玩家确认`，还是直接 `已完成`。
2. 是否需要玩家确认后才结算。
3. 接口是否需要提交交付说明、附件或实际服务时长。

### 取消并赔付

交互口径：

- 点击 `取消并赔付` 后，前端弹出专家赔付取消界面。
- 取消界面需要展示赔付规则、预计赔付金额、取消原因输入 / 选择和确认按钮。
- 合同金额、赔付金额、平台手续费、实际扣款需要由后台返回或以后端规则计算；前端当前仅支持从页面参数读取金额并做静态预览兜底，不能作为正式结算依据。
- 取消原因由后台返回可编辑的默认原因列表，前端单选，至少需要有 1 个可选原因并默认选中 1 个；用户填写的详细说明最多 50 字。
- `pages/game/expert-cancel/index` 当前确认按钮只做必填项校验和静态提示，真实取消赔付提交接口后续添加。
- 用户确认后应调用取消赔付接口，并刷新当前服务状态。

候选路径：

```text
POST /api/app/game-services/{serviceOrderId}/cancel-with-compensation
```

建议入参：

```json
{
  "gameId": "game_001",
  "reasonCode": "schedule_conflict",
  "reasonText": "时间冲突，无法继续服务",
  "reasonRemark": "本周临时出差，无法继续交付",
  "compensationRate": 20,
  "compensationAmount": 160,
  "platformFee": 16,
  "totalDebitAmount": 176
}
```

建议返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "serviceOrderId": "service_order_001",
    "statusType": "canceled",
    "statusText": "已取消（已赔付）",
    "compensationAmountText": "¥120",
    "reasonText": "时间冲突，无法继续服务"
  },
  "requestId": "req_xxx"
}
```

待确认：

1. 赔付金额由后台计算并返回，还是前端根据规则展示预估。
2. 取消原因使用固定枚举还是自由文本。
3. 取消后是否进入争议、已取消，还是单独状态 `canceled_with_compensation`。
4. 是否需要二次确认、风控校验或客服介入。
5. 专家取消页中的合同金额、平台手续费、实际扣款是否由预览接口完整返回；若仅返回费率，字段名使用 `platformFeeRate` 还是 `serviceFeeRate`，金额单位使用元还是分。
6. 专家取消原因列表是否由后台运营配置返回，字段名使用 `reasonOptions` 还是 `cancelReasons`；详细说明最大长度当前按 50 字处理。

### 联系玩家 / 联系领路人

交互口径：

- 点击 `联系玩家` 后，应进入或创建当前专家与玩家的一对一消息会话，并支持给玩家发消息。
- 点击 `联系领路人` 后，应进入或创建当前专家与领路人的一对一消息会话，并支持给领路人发消息。
- 当前先不实现真实 IM，只记录需要补充会话创建 / 消息发送能力。

候选路径：

```text
POST /api/app/im/conversations
POST /api/app/im/messages
```

建议创建会话入参：

```json
{
  "scene": "game_service",
  "gameId": "game_001",
  "serviceOrderId": "service_order_001",
  "targetUserId": "player_001",
  "targetRole": "player"
}
```

建议创建会话返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "conversationId": "conversation_001",
    "targetUserId": "player_001",
    "targetRole": "player",
    "route": "pages/im/room/index?conversationId=conversation_001"
  },
  "requestId": "req_xxx"
}
```

建议发送消息入参：

```json
{
  "conversationId": "conversation_001",
  "contentType": "text",
  "content": "你好，我来和你确认本次服务安排。"
}
```

待确认：

1. 联系玩家 / 领路人是先进入聊天页由用户手动发送，还是直接发送默认话术。
2. IM 会话是否已有正式接口和会话 ID 规则，当前候选路径需后端确认。
3. `GET /api/app/games/my/manage` 是否返回 `player.id`、`guide.id` 和当前专家 `viewer.role`。
4. 是否允许在服务取消、完成后继续联系玩家 / 领路人，按钮权限应由后台 `actions` 返回。

## 29. 我的引荐记录页列表与操作接口

记录日期：2026-06-23

模块：组局 / 我的引荐记录 - 领路人

页面：`pages/game/referral-record/index`

功能：领路人查看自己的引荐收益统计和引荐服务记录。页面中的收益统计、tab 数量、玩家信息、行家信息、引荐编号、服务标题、引荐奖励、预计交付时间、取消原因、评价状态等都应来自后台，当前前端只做静态走查占位。

候选接口：

```text
GET /api/app/game-referrals/my-records
POST /api/app/game-referrals/{referralId}/remind-delivery
POST /api/app/im/conversations
```

建议列表入参：

```json
{
  "status": "processing",
  "page": 1,
  "pageSize": 20
}
```

说明：`status` 对应页面 tab，取值建议为 `processing`、`completed`、`canceled`。如果后台希望一次返回三类记录，也可以返回 `records` 全量列表，由前端按 `state/statusType` 筛选。

建议列表返回：

```json
{
  "summary": {
    "label": "本月引荐收益",
    "amountText": "¥2,450",
    "successCount": 12,
    "processingCount": 3,
    "reviewCount": 2
  },
  "tabs": [
    { "key": "processing", "label": "进行中", "count": 3 },
    { "key": "completed", "label": "已完成", "count": 12 },
    { "key": "canceled", "label": "已取消", "count": 2 }
  ],
  "records": [
    {
      "id": "referral_001",
      "referralId": "referral_001",
      "ref": "REF-20260320-001",
      "status": "processing",
      "statusText": "服务进行中",
      "timeText": "3天前",
      "expert": {
        "id": "expert_001",
        "name": "张专家",
        "avatarUrl": "",
        "avatarText": "ZH"
      },
      "player": {
        "id": "player_001",
        "name": "李明",
        "avatarUrl": "",
        "avatarText": "LI"
      },
      "service": {
        "title": "产品架构咨询",
        "amountText": "¥800",
        "expectedDeliveryText": "预计交付时间：2026-03-25（剩余2天）"
      },
      "rewardText": "¥80",
      "actions": {
        "canRemindDelivery": true,
        "canViewGroupChat": true,
        "canReviewBoth": false
      },
      "groupChat": {
        "groupId": "group_001",
        "conversationId": "conversation_001"
      }
    }
  ],
  "pagination": {
    "page": 1,
    "pageSize": 20,
    "hasMore": false
  }
}
```

当前前端接入口径：

- 页面顶部收益卡复用 `components/business-summary-card`，但引荐页传入自定义统计项：`成功 / 进行中 / 待评价`。
- 页面 tab 与服务记录状态一一对应：`进行中` 只展示 `processing`，`已完成` 只展示 `completed`，`已取消` 只展示 `canceled`。
- 服务记录卡中的玩家、行家、编号、预计交付时间等均应以后端返回为准。
- 已完成服务的评价状态以后端返回为准；`reviewed/hasReviewed/hasEvaluated` 为真，或 `canReviewBoth/canReview` 为假时，前端显示 `已评价` 并禁用评价按钮，不能再次评价。
- 当前 `提醒交付` 和 `查看群聊` 只保留按钮和待接入提示；正式实现需要调用接口或进入 IM 会话。
- 当前 `评价双方` 只保留按钮和待接入提示；正式实现需要进入评价页或弹出评价表单，提交成功后刷新当前服务的评价状态。
- 列表需要支持按 tab 下拉刷新和分页 / 加载更多；分页字段、每个 tab 是否独立请求，以及刷新后是否同步更新顶部收益统计和 tab 数量均待后端确认。

### 提醒交付

候选路径：

```text
POST /api/app/game-referrals/{referralId}/remind-delivery
```

建议入参：

```json
{
  "gameId": "game_001",
  "serviceOrderId": "service_order_001",
  "targetRole": "expert"
}
```

待确认：

1. 提醒交付是直接发送提醒，还是进入聊天页并预填提醒话术。
2. 提醒对象是行家、玩家，还是按服务状态由后台决定。
3. 是否需要限制提醒频率，后台是否返回下次可提醒时间。

### 查看群聊

候选路径：

```text
POST /api/app/im/conversations
```

建议入参：

```json
{
  "scene": "game_referral_group",
  "gameId": "game_001",
  "referralId": "referral_001",
  "conversationId": "conversation_001"
}
```

待确认：

1. 查看群聊是直接使用列表返回的 `conversationId`，还是每次调用接口获取 / 创建会话。
2. 进入群聊时需要携带当前用户角色 `guide`，以及玩家 ID、行家 ID、组局 ID。
3. 群聊页是否复用 `pages/im/room/index`，还是需要单独的三方群聊页。

### 评价双方

候选路径：

```text
POST /api/app/game-referrals/{referralId}/reviews
```

备选路径：

```text
POST /api/app/game-services/{serviceOrderId}/reviews
```

建议入参：

```json
{
  "gameId": "game_001",
  "serviceOrderId": "service_order_001",
  "expertRating": 5,
  "playerRating": 5,
  "expertComment": "行家交付及时，服务专业",
  "playerComment": "玩家沟通顺畅，需求明确"
}
```

建议返回：

```json
{
  "reviewStatus": "reviewed",
  "reviewed": true,
  "canReviewBoth": false,
  "reviewedAt": "2026-03-20T12:00:00+08:00"
}
```

待确认：

1. 评价入口是跳转独立评价页，还是当前页弹出评价表单。
2. 领路人是否需要分别评价玩家和行家，还是只提交一次综合评价。
3. 评分、标签、文字评价、匿名评价等字段是否必填。
4. 提交成功后是否由接口返回新的记录项，还是前端重新请求列表。
5. 已评价服务是否需要展示评价详情入口，还是只置灰显示 `已评价`。

待后端 / 产品确认：

1. 接口路径是否使用 `/api/app/game-referrals/my-records`。
2. 列表是否分页；如果分页，tab 数量由 summary 返回还是每个 tab 分别请求。
3. 金额单位使用元、分还是后端直接返回 `amountText/rewardText`。
4. 状态枚举是否固定为 `processing/completed/canceled`。
5. 已完成卡片的 `待评价` 状态是否由 `reviewStatus` 或 `canReviewBoth` 返回。
6. 已取消卡片是否展示取消原因、取消方、赔付 / 退款金额，以及这些字段命名。

## 30. 玩家组局管理页列表与服务进度接口

记录日期：2026-06-23

模块：组局 / 组局管理 - 玩家

页面：`pages/game/player-manage/index`

功能：玩家侧“组局-组局管理”页面的顶部服务支出统计、tab 数量、服务卡片、服务进度和评价状态需要由后台返回。前端只负责展示和按时间计算进度百分比。

候选接口：

```text
GET /api/app/games/player/manage
```

建议返回：

```json
{
  "currentTime": "2026-03-19T14:00:00+08:00",
  "summary": {
    "label": "本月服务支出",
    "amount": 3200,
    "amountText": "¥3,200",
    "activeCount": 1,
    "completedCount": 4,
    "canceledCount": 1
  },
  "orders": [
    {
      "id": "player-manage-active-001",
      "statusType": "active",
      "statusText": "服务进行中",
      "ref": "REF-20260320-001",
      "fundAmount": 800,
      "expert": { "id": "expert-zhang", "name": "张专家", "avatarText": "ZH" },
      "guide": { "id": "guide-wang", "name": "王引荐" },
      "serviceTitle": "产品架构咨询",
      "startedAt": "2026-03-10T14:00:00+08:00",
      "expectedDeliveryAt": "2026-03-25T14:00:00+08:00",
      "remainingDays": 6,
      "noticeText": "取消需赔付一定比例金额给行家"
    },
    {
      "id": "player-manage-complete-001",
      "statusType": "complete",
      "statusText": "已完成",
      "ref": "REF-20260318-002",
      "fundAmount": 1200,
      "expert": { "id": "expert-wang", "name": "王导师", "avatarText": "WM" },
      "guide": { "id": "guide-chen", "name": "陈引荐" },
      "serviceTitle": "品牌定位咨询",
      "completedAtText": "完成时间：2026-03-19 18:30",
      "reviewStatus": "pending",
      "canReview": true,
      "reviewActionText": "评价双方"
    }
  ]
}
```

当前前端接入口径：

- 页面进入时请求 `GET /api/app/games/player/manage`。
- 统计卡和 tab 数量读取 `summary`。
- 卡片编号读取 `ref/refNo/orderNo/serviceNo/gameNo/groupNo`。
- 资金金额优先读取 `amountText/serviceAmountText/fundAmountText/expertAmountText/serviceFeeText/priceText`；若只返回 `amount/fundAmount/expertAmount/serviceFee` 数字，前端格式化为人民币。
- 专家信息优先读取 `expert.name/expert.nickname/expert.avatarText`，也兼容平铺的 `expertName/avatarText`。
- 领路人信息优先读取 `guide.name/guide.nickname`，也兼容平铺的 `guideName`。
- 预计交付时间读取 `expectedDeliveryAt/deliveryDeadlineAt/deliveryAt/dueAt`，显示为 `预计交付：YYYY-MM-DD HH:mm`。
- 服务进度百分比和进度条由前端根据 `startedAt`、`expectedDeliveryAt` 与 `currentTime/serverTime` 计算；如果后台不返回 `currentTime/serverTime`，前端用本机当前时间。后台也可返回 `remainingDays/leftDays` 覆盖剩余天数显示。
- 已完成服务的评价按钮由后台字段控制：`reviewStatus/evaluateStatus/commentStatus`、`reviewed/hasReviewed/hasEvaluated`、`canReview/canReviewBoth`。未评价显示 `评价双方`，已评价显示 `已评价` 并禁用。

待后端 / 产品确认：

1. 接口路径是否使用 `/api/app/games/player/manage`。
2. 进度计算的起点应使用 `startedAt`、组局成功时间，还是行家确认服务开始时间。
3. 剩余天数以后端 `remainingDays` 为准，还是前端按交付时间实时计算。
4. 评价状态枚举使用 `pending/reviewed`，还是使用现有评价模块状态。
5. 点击 `联系行家` 后应进入已有会话还是调用创建会话接口。
6. 点击 `申请取消` 后取消窗口需要展示哪些赔付规则、取消原因和确认接口字段。

## 31. 组局取消领路人详情接口

记录日期：2026-06-23

模块：组局 / 组局取消 - 领路人

页面：`pages/game/guide-cancel/index`

功能：领路人查看某次组局取消结果。取消方可能是玩家，也可能是行家，因此取消方、角色、原因、说明话语、取消时间和取消历程均需要由后台返回，前端不写死“玩家婉拒”。

候选接口：

```text
GET /api/app/game-invites/guide-cancel-detail
```

建议入参：

```json
{
  "id": "invite-complete-002"
}
```

建议返回：

```json
{
  "id": "invite-complete-002",
  "statusTitle": "组局已取消",
  "statusDesc": "行家取消了此次组局邀请",
  "canceledBy": {
    "id": "expert_001",
    "name": "王强",
    "roleType": "expert",
    "roleLabel": "行家",
    "avatarUrl": "",
    "avatarText": "WQ"
  },
  "reason": {
    "title": "档期冲突",
    "desc": "行家临时档期调整，无法按时参加"
  },
  "message": "抱歉，临时档期有冲突，本次无法继续参加。",
  "messageTimeText": "2小时前",
  "timeline": [
    {
      "key": "invite",
      "title": "发起邀请",
      "desc": "你向双方发送了组局邀请",
      "timeText": "03-21 10:23",
      "state": "active"
    },
    {
      "key": "cancel",
      "title": "行家取消",
      "desc": "王强因档期冲突取消本次组局",
      "timeText": "03-21 16:45",
      "state": "error"
    },
    {
      "key": "canceled",
      "title": "组局取消",
      "desc": "因一方取消，组局自动取消",
      "state": "pending"
    }
  ]
}
```

待确认：

1. 接口路径是否使用 `/api/app/game-invites/guide-cancel-detail`，还是复用进度详情接口返回取消态字段。
2. 取消方角色枚举是否固定为 `player/expert/guide`。
3. `reason.title`、`reason.desc`、`message` 是否均由后台返回展示文案。
4. 取消历程是否由后台完整返回，还是前端按取消状态本地生成。

## 32. 玩家申请取消服务确认接口

记录日期：2026-06-23

模块：组局 / 取消申请 - 玩家

页面：`pages/game/player-cancel/index`

功能：玩家从“组局管理”点击 `申请取消` 后进入取消服务确认页，展示活动信息、建议赔付比例、赔付金额、平台手续费、实际支付、剩余可退金额、取消原因和赔付协议。取消原因由后台返回可编辑的默认原因列表，前端单选，至少需要有 1 个可选原因并默认选中 1 个；用户填写的详细说明最多 50 字。当前页面按用户提供的 `申请取消服务.txt` 静态设计稿落地白色母版走查，提交按钮只做静态提示，未接真实赔付取消接口。

数据口径：活动信息中的合同金额、服务时长必须在进入 `pages/game/player-cancel/index` 前明确给出。当前入口由 `pages/game/player-manage/index` 点击 `申请取消` 时携带订单金额和服务时长参数；正式联调时也可由 `GET /api/app/game-services/{serviceOrderId}/player-cancel-preview` 返回 `contractAmount` / `contractAmountText`、`servedDurationText`、`totalDurationText` 或组合后的 `servedText`。服务时长展示中 `/` 左边为已经服务时长，右边为活动总时长。玩家取消页只展示和消费这些数据，静态兜底仅用于走查，不能在进入页面后再临时猜测或推算合同金额、已服务时长和活动总时长。

交互口径：`自定义赔付比例` 标题行中的 `可协商` 标签需要放在右侧；后续点击该标签应进入与行家 / 专家的聊天界面，让玩家先沟通赔付比例。当前先记录交互要求，不接真实聊天跳转、会话创建或消息预填逻辑。

赔付比例口径：建议赔付比例当前暂定为页面展示值 `15%`，比例滑块当前静态范围为最低 `5%`、最高 `30%`。正式联调时最低比例、最高比例、默认建议比例应支持由后台返回；玩家调整比例后可调用赔付预览 / 调整接口重新计算赔付金额、平台手续费、实际支付和剩余可退金额。

候选接口：

```text
GET /api/app/game-services/{serviceOrderId}/player-cancel-preview
POST /api/app/game-services/{serviceOrderId}/player-cancel-with-compensation
```

建议预览返回：

```json
{
  "serviceOrderId": "SO-20260320-001",
  "gameId": "game_001",
  "ref": "REF-20260320-001",
  "expert": { "id": "expert_001", "name": "张专家", "avatarText": "ZH" },
  "serviceTitle": "产品架构咨询",
  "contractAmount": 800,
  "contractAmountText": "¥800",
  "servedDurationText": "1.5小时",
  "totalDurationText": "2小时",
  "servedText": "1.5小时 / 2小时",
  "minRate": 5,
  "maxRate": 30,
  "suggestedRate": 15,
  "platformFeeRate": 10,
  "reasonOptions": [
    { "key": "need_changed", "text": "需求变更，不再需要服务" },
    { "key": "other_solution", "text": "找到其他解决方案" }
  ],
  "agreements": [
    "我理解主动取消需承担行家的时间成本损失",
    "我同意按设置比例赔付行家，金额从托管资金扣除"
  ]
}
```

建议提交入参：

```json
{
  "gameId": "game_001",
  "reasonCode": "other_solution",
  "reasonText": "已找到其他解决方案",
  "compensationRate": 15,
  "compensationAmount": 120,
  "platformFee": 12,
  "payAmount": 132
}
```

待确认：

1. 赔付预览是否由后台返回完整金额，还是前端仅根据后台规则估算展示。
2. 赔付比例是否允许玩家拖动调整，调整范围是否固定为 `5% - 30%`。
3. 平台手续费是否固定为赔付金额的 `10%`，以及金额单位使用元还是分。
4. 提交后是否需要发起微信支付、直接从托管资金扣除，还是进入客服 / 行家确认流程。
5. 取消原因枚举、协议文案、信用分影响是否全部由后台返回。
6. 玩家取消原因列表是否由后台运营配置返回，字段名使用 `reasonOptions` 还是 `cancelReasons`；详细说明最大长度当前按 50 字处理。
7. 上一页订单列表或取消预览接口必须在进入玩家取消页前提供合同金额；若订单列表没有金额字段，需要先调用预览接口拿到金额后再跳转。
8. 上一页订单列表或取消预览接口必须在进入玩家取消页前提供已服务时长和活动总时长；展示格式为 `已服务时长 / 活动总时长`，若订单列表没有完整时长字段，需要先调用预览接口拿到时长后再跳转。
9. `可协商` 标签点击后的聊天入口需要确认使用已有 `conversationId` 进入会话，还是由接口按 `serviceOrderId/gameId/expertId` 创建或获取与行家 / 专家的会话。
10. 赔付比例是否需要独立调整接口，或复用取消预览接口传入 `compensationRate` 后返回重算结果；最低比例、最高比例和默认建议比例应以后端返回为准，当前 `5% - 30%`、`15%` 仅为静态走查口径。

## 33. 局内协作页数据与操作接口

记录日期：2026-06-23

模块：组局 / 局内协作

页面：`pages/game/collaboration/index`

功能：局内协作页已按白色母版加深色内容区静态落地。页面主体 3 个区域分别为 `进度情况`、`成员`、`聊天区`，三块都需要保持圆角卡片样式。正式联调时进度情况、成员列表、聊天区消息都必须从后台获取，不能继续写死静态数据。底部 `成员管理` 点击后应列出当前局成员；`结束本局` 点击后应跳转到结束确认页面。

候选接口：

```text
GET /api/app/games/{gameId}/collaboration
GET /api/app/games/{gameId}/members
GET /api/app/games/{gameId}/messages
```

建议协作页聚合返回：

```json
{
  "gameId": "game_001",
  "statusText": "进行中",
  "dayText": "第 6 天 / 共 15 天",
  "progress": {
    "percent": 46,
    "title": "进度 46%",
    "tasks": [
      {
        "key": "done",
        "title": "已完成",
        "desc": "确认产品方向",
        "state": "completed"
      },
      {
        "key": "doing",
        "title": "进行中",
        "desc": "完成用户旅程与页面结构",
        "state": "active"
      },
      {
        "key": "todo",
        "title": "待完成",
        "desc": "输出 PRD 与字段清单",
        "state": "pending"
      }
    ]
  },
  "members": [
    {
      "id": "user_001",
      "name": "佩娜",
      "roleText": "发起人",
      "avatarUrl": "",
      "avatarText": "PN"
    }
  ],
  "messages": [
    {
      "id": "msg_001",
      "senderId": "user_001",
      "senderName": "佩娜",
      "content": "欢迎大家加入，我们先把页面结构和MVP优先级收住。",
      "createdAt": "2026-06-23T16:00:00+08:00"
    }
  ],
  "actions": {
    "canManageMembers": true,
    "canEndGame": true,
    "endConfirmRoute": "pages/game/end-confirm/index?gameId=game_001"
  }
}
```

当前前端接入口径：

- 页面顶部标题为 `局内协作`，内容区为深色背景，`进度情况 / 成员 / 聊天区` 三个区域均保持圆角卡片。
- `progress.percent` 控制进度条和进度标题；任务列表由 `progress.tasks` 渲染。
- `members` 用于成员区域摘要，也用于点击 `成员管理` 后的成员列表页面 / 弹层。
- `messages` 用于聊天区消息列表，消息发送、分页和实时刷新规则后续确认。
- `成员管理` 不只是静态提示，后续应进入成员列表或弹出成员管理面板，并展示当前局所有成员。
- `结束本局` 后续应跳转结束确认页面，确认页路径和参数以后端返回或前端路由约定为准。

待确认：

1. 协作页是否使用一个聚合接口 `/api/app/games/{gameId}/collaboration`，还是进度、成员、聊天分别请求。
2. `progress.tasks` 是否固定三类状态，还是由后台按当前局阶段动态返回。
3. 成员管理使用独立页面、底部弹层还是弹窗；成员列表是否支持移除成员、变更角色、转让发起人等操作。
4. 聊天区是否复用正式 IM 会话接口，还是只展示当前局内留言摘要。
5. 结束确认页面路径是否新增 `pages/game/end-confirm/index`，以及结束本局需要哪些确认信息、权限校验和提交接口。

## 34. 服务交付确认页操作接口

记录日期：2026-06-23

模块：组局 / 服务交付 - 有偿 / 免费

页面：`pages/game/delivery/index`

功能：服务交付确认页当前为静态 UI。快捷操作中的 `联系玩家`、`联系领路人` 后续需要进入对应人员的消息界面；确认状态里的 `提醒确认` 后续需要给玩家发消息 / 提醒玩家确认服务完成；底部 `确认服务完成` 后续需要提交服务完成确认，并触发玩家确认、状态刷新，以及有偿局结算流程或免费局归档流程。

候选接口：

```text
GET /api/app/game-services/{serviceOrderId}/delivery-detail
POST /api/app/im/conversations
POST /api/app/game-services/{serviceOrderId}/remind-player-confirm
POST /api/app/game-services/{serviceOrderId}/confirm-delivery
```

建议详情返回：

```json
{
  "serviceOrderId": "service_order_001",
  "gameId": "game_001",
  "orderNo": "REF-20260320-001",
  "status": "waiting_player_confirm",
  "expert": {
    "id": "expert_001",
    "name": "李明",
    "avatarText": "LI"
  },
  "player": {
    "id": "player_001",
    "name": "玩家A"
  },
  "guide": {
    "id": "guide_001",
    "name": "领路人A"
  },
  "service": {
    "title": "产品架构咨询",
    "durationText": "2小时 (已完成)",
    "startedAtText": "03-20 14:30",
    "completedAtText": "03-20 16:30",
    "contractAmountText": "¥800.00"
  },
  "serviceType": "paid",
  "settlement": {
    "contractAmountText": "¥800.00",
    "settlementRatioText": "60% (扣除成本40%)",
    "settlementAmountText": "¥480.00",
    "platformFeeText": "-¥48.00",
    "guideRewardText": "-¥192.00",
    "systemGuideRewardText": "-¥48.00",
    "actualAmountText": "¥192.00"
  },
  "actions": {
    "canContactPlayer": true,
    "canContactGuide": true,
    "canConfirmDelivery": true
  }
}
```

快捷操作口径：

- 点击 `联系玩家` 后，应进入或创建与玩家的一对一消息会话，并给玩家发消息。
- 点击 `联系领路人` 后，应进入或创建与领路人的一对一消息会话，并给领路人发消息。
- 会话创建可复用 IM 会话接口，返回 `conversationId` 或可跳转的 `route`。

提醒确认口径：

- 确认状态里的 `提醒确认` 按钮当前只做静态提示，后续应给玩家发消息 / 提醒玩家进行服务完成确认。
- 产品需确认该动作是进入玩家聊天页并预填确认话术，还是直接调用提醒接口发送系统提醒。
- 若直接发送提醒，建议后端返回提醒发送时间、冷却时间和按钮文案，避免重复频繁提醒。

候选提醒接口：

```text
POST /api/app/game-services/{serviceOrderId}/remind-player-confirm
```

建议入参：

```json
{
  "gameId": "game_001",
  "targetUserId": "player_001",
  "message": "服务已完成，请确认服务完成状态。"
}
```

建议返回：

```json
{
  "serviceOrderId": "service_order_001",
  "reminded": true,
  "remindedAtText": "刚刚",
  "nextAllowedAt": "2026-06-23T18:30:00+08:00"
}
```

建议创建会话入参：

```json
{
  "scene": "game_service_delivery",
  "gameId": "game_001",
  "serviceOrderId": "service_order_001",
  "targetUserId": "player_001",
  "targetRole": "player"
}
```

确认完成口径：

- 底部 `确认服务完成` 按钮当前只做静态提示，后续应调用服务完成确认接口。
- 提交前需要校验服务确认勾选项，确认后由接口返回最新服务状态、确认状态和结算状态。
- 如果仍需玩家二次确认，状态应进入 `waiting_player_confirm`；若双方均已确认，有偿局进入结算 / 已完成状态，免费局进入归档 / 已完成状态。

候选确认接口：

```text
POST /api/app/game-services/{serviceOrderId}/confirm-delivery
```

建议入参：

```json
{
  "gameId": "game_001",
  "confirmedItems": ["completed", "qualified", "communicated"]
}
```

建议返回：

```json
{
  "serviceOrderId": "service_order_001",
  "status": "waiting_player_confirm",
  "statusText": "等待玩家确认",
  "timeline": [
    { "key": "completed", "title": "服务已完成", "state": "done", "timeText": "03-20 16:35" },
    { "key": "waiting-player", "title": "等待玩家确认", "state": "active" },
    { "key": "settlement", "title": "资金结算", "state": "pending" }
  ]
}
```

待确认：

1. 服务交付确认页是否使用独立详情接口，还是复用业务管理 / 玩家管理的服务订单详情。
2. 点击 `联系玩家`、`联系领路人` 是先进入聊天页由用户手动发送，还是直接发送默认话术。
3. `提醒确认` 是进入玩家聊天页预填话术，还是直接调用提醒接口给玩家发送确认消息。
4. `提醒确认` 是否需要冷却时间、次数限制和重复提醒文案。
5. `确认服务完成` 是由行家单方提交后等待玩家确认，还是玩家 / 行家任一方都可确认。
6. 结算明细中的平台服务费、领路人奖励、系统级领路人奖励是否全部由后台返回，金额单位使用元还是分；免费局是否只返回免费局说明、积分 / 徽章 / 推荐权益和归档状态。
7. 提交确认后是否立即刷新当前页，还是跳转到进度 / 管理页。

## 35. 服务评价页提交接口

记录日期：2026-06-23

模块：组局 / 服务评价 - 行家

页面：`pages/game/review/index`

功能：服务评价页当前只完成静态表单和前端选中态。后续点击 `提交评价` 时，需要把页面中的满意度、故事文本、玩家评价、领路人评价、标签、NPS、当前用户角色、组局 ID / 服务订单 ID / 引荐 ID 等发送给后台；提交成功后跳转到哪里仍需产品确认，当前不在前端写死。

候选接口：

```text
POST /api/app/reviews
```

建议入参：

```json
{
  "source": "game-service-review",
  "role": "expert",
  "gameId": "game_001",
  "serviceOrderId": "service_order_001",
  "orderId": "order_001",
  "referralId": "referral_001",
  "satisfaction": {
    "value": "great",
    "label": "真好玩"
  },
  "story": "本次合作沟通顺畅，需求清晰，过程里有不少新启发。",
  "evaluations": [
    {
      "targetType": "player",
      "targetUserId": "player_001",
      "score": 5,
      "tags": ["需求明确", "配合度高"],
      "comment": "写下对需求方的评价..."
    },
    {
      "targetType": "guide",
      "targetUserId": "guide_001",
      "score": 5,
      "tags": ["匹配精准", "响应及时"],
      "comment": "评价引荐人的服务质量..."
    }
  ],
  "npsScore": 7
}
```

建议返回：

```json
{
  "reviewId": "review_001",
  "serviceOrderId": "service_order_001",
  "reviewStatus": "reviewed",
  "reviewed": true,
  "reviewedAt": "2026-06-23T23:30:00+08:00",
  "nextAction": {
    "type": "navigate",
    "route": "pages/game/manage/index",
    "query": "tab=completed"
  }
}
```

提交后跳转待确认：

1. 是否返回 `pages/game/delivery/index` 服务交付页，并刷新服务状态。
2. 是否跳转 `pages/game/manage/index` 我的业务管理已完成列表。
3. 是否跳转服务详情 / 评价成功页，并展示奖励发放结果。
4. 后端是否直接返回 `nextAction.route` 和 `nextAction.query`，由前端按返回值跳转。

待确认：

1. 本页面是否只面向行家，还是玩家 / 领路人也复用同一提交接口。
2. 玩家、领路人的 `targetUserId` 是否由进入页面时的订单详情接口返回，还是由提交接口根据 `serviceOrderId` 自动推断。
3. 评分、标签、文字评价、NPS 是否必填；未填写时后端是否允许提交。
4. `评价奖励 50PX + 优先推荐权益` 是否由提交接口同步返回发放状态。
5. 已评价后再次进入页面时，是展示评价详情、禁用提交，还是直接跳转已完成状态页。

## 36. 再次组局确认页接口

记录日期：2026-06-24

模块：组局 / 组局确认

页面：`pages/game/confirm/index`

功能：再次组局确认页需要先根据上一局信息展示服务类型、完成时间、参与人数，并展示当前发起人需要邀请的另外两方。当前测试默认按“领路人发起，再邀请行家和玩家”处理；正式逻辑需要后端返回当前发起人、发起人角色、可邀请成员列表和上一局上下文，前端不能固定展示 3 个上一局成员。

候选接口：

```text
GET /api/app/game-invites/replay-context
POST /api/app/game-invites/replay
```

`GET /api/app/game-invites/replay-context` 建议入参：

```json
{
  "sourceGameId": "game_001",
  "serviceOrderId": "service_order_001"
}
```

建议返回：

```json
{
  "sourceGameId": "game_001",
  "serviceOrderId": "service_order_001",
  "inviter": {
    "id": "guide_001",
    "name": "王引荐",
    "roleType": "guide",
    "roleLabel": "领路人"
  },
  "previousSession": {
    "serviceType": "产品架构咨询",
    "completedAtText": "2026-06-13 14:30",
    "participantText": "3人（行家+玩家+领路人）"
  },
  "invitees": [
    { "id": "expert_001", "name": "张专家", "roleType": "expert", "roleLabel": "行家", "desc": "产品架构咨询" },
    { "id": "player_001", "name": "王总", "roleType": "player", "roleLabel": "玩家", "desc": "需求方" }
  ]
}
```

`POST /api/app/game-invites/replay` 建议入参：

```json
{
  "sourceGameId": "game_001",
  "serviceOrderId": "service_order_001",
  "inviter": { "id": "guide_001", "roleType": "guide" },
  "invitees": [
    { "id": "expert_001", "roleType": "expert" },
    { "id": "player_001", "roleType": "player" }
  ],
  "message": "再来一局？"
}
```

建议返回：

```json
{
  "replayInvitationId": "replay_invite_001",
  "status": "pending",
  "statusText": "等待双方确认"
}
```

待确认：

1. 再次组局的发起人是否固定为领路人，还是玩家 / 行家也可以从评价完成页或组局记录发起。
2. 如果当前发起人是上一局参与方，是否只邀请另外两方；若发起人不是上一局参与方，是否允许邀请上一局三方。
3. 确认发起后应进入哪个流程页：当前先跳 `pages/game/guide-progress/index`，后续可改为再次组局专属进度页或后端返回 `nextAction`。
4. 再次组局是否沿用上一局服务类型、预算、时间、分润规则，还是只复用成员关系并重新填写组局需求。

## 37. 系统推荐适配局数据与确认组局接口

记录日期：2026-06-24

模块：组局 / 系统推荐适配局

页面：`pages/game/system-recommend/index`

功能：用户进入系统推荐适配局页后，应从后台获取推荐上下文、筛选标签、行家列表、默认选中行家和匹配度等数据，前端只负责展示、筛选和选择态。用户点击 `确认组局` 后，当前先跳转到 `pages/game/create/index` 并携带推荐 ID、已选行家 ID、上一局 / 服务订单上下文；正式联调时需要确认是先进入创建组局页预填，还是直接调用创建组局 / 创建邀请接口。

候选接口：

```text
GET /api/app/game-invites/system-recommendations
POST /api/app/games
```

`GET /api/app/game-invites/system-recommendations` 建议入参：

```json
{
  "sourceGameId": "game_001",
  "serviceOrderId": "service_order_001",
  "recommendationId": "system-rec-001",
  "category": "all"
}
```

建议返回：

```json
{
  "recommendationId": "system-rec-001",
  "title": "系统推荐适配局",
  "desc": "基于你的偏好，已找到5个高匹配度行家",
  "defaultSelectedExpertIds": ["expert_001"],
  "categories": [
    { "key": "all", "name": "全部" },
    { "key": "product", "name": "产品架构" },
    { "key": "tech", "name": "技术咨询" },
    { "key": "operation", "name": "运营策略" }
  ],
  "experts": [
    {
      "id": "expert_001",
      "name": "李资深",
      "role": "前阿里P8 · 产品架构专家",
      "avatarUrl": "",
      "avatarText": "LI",
      "avatarClass": "purple",
      "rating": "5.0",
      "stars": "★★★★★",
      "reviewCount": 128,
      "match": 98,
      "price": 800,
      "category": "product",
      "tags": ["产品架构", "技术方案", "团队管理", "响应及时"]
    }
  ]
}
```

确认组局当前前端跳转：

```text
pages/game/create/index?source=systemRecommend&recommendationId=system-rec-001&selectedExpertIds=expert_001&sourceGameId=game_001&serviceOrderId=service_order_001
```

待确认：

1. 推荐数据接口路径是否使用 `/api/app/game-invites/system-recommendations`，还是归入组局推荐 / 匹配服务接口。
2. 筛选标签是否由后台返回并支持后台筛选，还是前端拿全量列表后本地筛选。
3. `avatarText` 是否由后台返回；若未返回，前端会按统一头像规则用姓名姓氏拼音前两个字母生成。
4. `price` 金额单位使用元还是分；当前页面按元展示 `¥800/小时`。
5. 点击 `确认组局` 后，是进入创建组局页预填已选行家，还是直接创建系统推荐适配局并进入组局进度 / 支付 / 邀请流程。
6. 若进入创建组局页，`pages/game/create/index` 需要补充接收 `source=recommendationId/selectedExpertIds/sourceGameId/serviceOrderId` 并预填推荐上下文的能力。

## 38. 发起组局创建 / 草稿接口人数校验

记录日期：2026-06-24

模块：组局 / 发起组局

页面：`pages/game/create/index`

功能：用户在发起组局页填写局人数总人数时，前端当前按用户反馈限制最少 3 人、最多 10 人。正式保存草稿和发布创建组局接口必须在后端按同一范围强校验，不能只依赖前端滑块或输入框归一化。

候选接口：

```text
POST /api/app/games
POST /api/app/games/drafts
```

建议入参片段：

```json
{
  "capacity": 5
}
```

校验规则：

```text
3 <= capacity <= 10
```

待确认：

1. 正式创建组局接口是否使用 `POST /api/app/games`，草稿保存是否使用 `POST /api/app/games/drafts`。
2. `capacity` 字段是否表示总人数上限；若未来区分发起人、行家、领路人或报名人数，需要明确是否包含发起人在内。
3. 后端非法人数错误码和文案需要统一，建议前端展示为 `局人数需为 3-10 人`。
4. 若不同局类型未来允许不同人数范围，需要由后端返回局类型对应范围，前端再替换当前固定 `3-10` 规则。

## 39. 发起组局分润模板配置与公益局权限强校验

记录日期：2026-06-24

模块：组局 / 发起组局

页面：`pages/game/create/index`

功能：发起组局页的分润模板、押金局规则和押金局提示由后台配置下发。前端当前新增候选接口读取配置；接口失败或未返回对应字段时使用页面默认模板和默认押金局文案兜底。

候选接口：

```text
GET /api/app/games/profit-templates
POST /api/app/games
POST /api/app/games/drafts
```

`GET /api/app/games/profit-templates` 建议入参：

```json
{
  "feeType": "paid"
}
```

建议返回：

```json
{
  "currentAccountType": "player",
  "depositRuleText": "连续打卡 7 天即完成。完成者拿回押金池金额，未完成者押金由完成者平分。",
  "depositNoticeText": "支付金额：100元 = 服务费10元 + 押金池90元。服务费不退，押金池按完成情况结算。",
  "templates": [
    {
      "key": "deposit",
      "name": "押金局",
      "desc": "平台2.5% · 交付方0% · 流量方5% · 推荐上级2.5% + 押金池90%，完成返还，未完成瓜分",
      "selectable": true,
      "allowedAccountTypes": ["player", "expert", "guide", "platform"]
    },
    {
      "key": "publicBenefit",
      "name": "公益局",
      "desc": "平台0% · 交付方100% · 流量方0% · 推荐上级0%",
      "selectable": false,
      "disabledReason": "仅平台账户可发起",
      "allowedAccountTypes": ["platform"]
    }
  ]
}
```

强校验规则：

```text
公益局只能平台账户发起。
普通用户、玩家、行家、领路人都不能选择或创建公益局。
POST /api/app/games 和 POST /api/app/games/drafts 必须校验 profitTemplate/publicBenefit 与当前登录账号类型，不允许绕过前端提交。
```

待确认：

1. 分润模板配置正式接口是否使用 `GET /api/app/games/profit-templates`。
2. 后端是否按当前登录账号类型直接过滤不可用模板，还是返回 `selectable=false` 和 `disabledReason` 给前端展示。
3. `profitTemplate` 字段是否使用 `standard`、`aa`、`deposit`、`crowdfunding`、`publicBenefit` 这些 key，还是使用后端模板 ID。
4. 押金局规则和提示是否随分润模板接口一起返回，字段是否采用 `depositRuleText`、`depositNoticeText`。
5. 保存草稿和发布创建组局接口的非法模板 / 权限错误码与前端展示文案需要统一。

## 40. 发起组局底部操作按钮接口与预览页

记录日期：2026-06-24

模块：组局 / 发起组局

页面：`pages/game/create/index`

功能：发起组局页底部三个按钮的功能口径已确认，但正式接口和预览提交页仍待补充。

按钮口径：

```text
保存草稿：保存到草稿箱，待实现。
浏览提交：弹出预览页面，把用户填写和选择的内容按字段展示，底部提供提交按钮，待补充页面。
发布并同步到地球网：正式发布并同步到地球网，待实现。
```

候选接口：

```text
POST /api/app/games/drafts
POST /api/app/games
POST /api/app/games/{gameId}/sync-earth
```

待确认：

1. 草稿箱正式接口路径、草稿列表入口、草稿状态字段和是否支持覆盖保存。
2. 浏览提交预览页是新增页面、弹窗组件，还是由当前页内弹层承载。
3. 预览页展示字段顺序需要和发布入参字段保持一致，包含封面、局属性、时间地点、人数、参与方式、标签、描述、费用与分润、规则提示等。
4. `发布并同步到地球网` 是一个创建接口内完成同步，还是创建成功后再调用同步接口。
5. 发布同步失败时是否允许组局创建成功但地球网同步失败，以及前端应该展示的状态和重试入口。

## 41. 组局大厅分类配置接口与前端兜底

记录日期：2026-06-24

模块：组局 / 组局大厅

页面：`pages/game/hall/index`

功能：组局大厅顶部圆形分类入口和二级细分应优先由后台配置接口下发，前端按后台返回的一级分类、二级细分、排序、可见状态和筛选 key 渲染；若后台未返回分类配置、接口失败或字段为空，则前端使用用户截图中的二级细分作为静态兜底，保证大厅可用。

候选接口：

```text
GET /api/app/games/category-config
GET /api/app/games/hall
```

建议返回：

```json
{
  "cityFilters": [
    { "key": "shanghai", "name": "上海", "visible": true, "sort": 10 }
  ],
  "primaryCategories": [
    {
      "key": "social",
      "name": "社交局",
      "icon": "category-social",
      "visible": true,
      "sort": 10,
      "children": [
        { "key": "meal", "name": "饭局", "visible": true, "sort": 10 },
        { "key": "honest-talk", "name": "坦白局", "visible": true, "sort": 20 },
        { "key": "board-game", "name": "桌游局", "visible": true, "sort": 30 },
        { "key": "friend-making", "name": "交友局", "visible": true, "sort": 40 },
        { "key": "city-walk", "name": "同城散步局", "visible": true, "sort": 50 }
      ]
    }
  ],
  "typeFilters": [
    { "key": "standard", "name": "标准局", "visible": true, "sort": 10 },
    { "key": "aa", "name": "AA局", "visible": true, "sort": 20 },
    { "key": "crowdfunding", "name": "众筹局", "visible": true, "sort": 30 },
    { "key": "deposit", "name": "押金局", "visible": true, "sort": 40 },
    { "key": "publicBenefit", "name": "公益局", "visible": true, "sort": 50 }
  ]
}
```

前端静态兜底分类：

| 一级分类 | 二级细分 |
| --- | --- |
| 社交局 | 饭局、坦白局、桌游局、交友局、同城散步局 |
| 任务局 | 找合伙人、做项目、头脑风暴、组队共创、打磨方案 |
| 探索局 | 城市探索、路线盲盒、打卡挑战、城市故事采集、夜游/徒步/骑行 |
| 成长局 | 读书局、健身局、打卡局、押金局、学习共修局 |

玩法 / 结算类型兜底：

```text
标准局、AA局、众筹局、押金局、公益局
```

字段口径：

| 字段 | 来源 | 用途 |
| --- | --- | --- |
| `primaryCategories` | 后台 | 顶部一级分类入口，控制显示顺序和可见状态。 |
| `primaryCategories.children` | 后台 | 点击一级分类后展示的二级细分。 |
| `cityFilters` | 后台 | 高级筛选城市选择，不应长期固定为上海。后台未返回时前端可从大厅列表数据的城市字段去重兜底。 |
| `typeFilters` | 后台 | 筛选栏 `类型` 选项，表示玩法、收费或结算方式。 |
| `key` | 后台 | 前端筛选和列表查询参数，不直接展示。 |
| `name` | 后台 | 页面展示文案。 |
| `icon` | 后台 / 前端映射 | 图标资源 key，前端映射到项目内静态图标，避免后台直接控制本地路径。 |
| `visible` | 后台 | 是否展示该分类或筛选项。 |
| `sort` | 后台 | 展示顺序。 |

待确认：

1. 分类配置是独立接口 `GET /api/app/games/category-config`，还是并入大厅列表聚合接口 `GET /api/app/games/hall`。
2. 后台是否返回完整一级分类和二级细分；若仅返回列表数据中的分类枚举，前端是否仍需要本地配置兜底。
3. `宝妈局` 是否作为独立一级分类，还是归入圈层局 / 社交局的二级细分，需要产品确认。
4. `押金局` 当前既可能出现在成长局二级细分，也属于玩法 / 结算类型；后续建议后台明确主分类和玩法类型两个字段，避免筛选口径混淆。
5. 列表接口筛选参数建议区分 `primaryCategory`、`secondaryCategory` 和 `type`，分别对应一级目的分类、二级玩法细分、玩法 / 结算类型。
6. 高级筛选城市池是否由分类配置接口返回 `cityFilters`，还是由大厅列表聚合接口返回；前端当前只有上海静态兜底，正式联调时需要替换为后台城市池或从列表城市字段去重生成。

## 42. 发起组局分类标签字段

记录日期：2026-06-24

模块：组局 / 发起组局

页面：`pages/game/create/index`

功能：组局大厅已确认按一级分类和二级细分筛选，因此发起组局页也必须让发起人选择同一套分类标签，并在保存草稿 / 发布创建组局时提交给后台。当前创建页只有 `局类型` 一级选项和普通内容标签，尚未和大厅分类配置打通。

配置来源：

```text
GET /api/app/games/category-config
```

保存 / 发布候选接口：

```text
POST /api/app/games/drafts
POST /api/app/games
```

建议入参片段：

```json
{
  "primaryCategory": "growth",
  "primaryCategoryText": "成长局",
  "secondaryCategory": "reading",
  "secondaryCategoryText": "读书局",
  "type": "deposit",
  "typeText": "押金局",
  "tags": ["自律", "共修", "读书"]
}
```

字段口径：

| 字段 | 是否必填 | 用途 |
| --- | --- | --- |
| `primaryCategory` | 是 | 大厅一级分类筛选，例如 `social/task/explore/growth`。 |
| `primaryCategoryText` | 可选 | 后台可返回展示文案；前端也可由配置映射得到。 |
| `secondaryCategory` | 是 | 大厅二级细分筛选，例如 `meal/board-game/reading/checkin`。 |
| `secondaryCategoryText` | 可选 | 二级细分展示文案。 |
| `type` | 是 | 玩法 / 收费 / 结算类型，例如 `standard/aa/crowdfunding/deposit/publicBenefit`。 |
| `typeText` | 可选 | 类型展示文案。 |
| `tags` | 可选 | 普通内容标签，例如产品研发、创业、共创；不承担大厅分类筛选主逻辑。 |

前端交互口径：

- 创建页分类选择应和组局大厅使用同一份后台配置。
- 至少要求选择一个一级分类和一个二级细分。
- `类型` 使用玩法 / 收费 / 结算类型，不和一级分类混用。
- 后台没有返回分类配置时，前端按组局大厅静态兜底细分展示。
- 普通内容标签可以继续保留，但不能替代 `primaryCategory`、`secondaryCategory` 和 `type`。

大厅列表返回建议：

```json
{
  "id": "game_001",
  "title": "押金读书共修局",
  "primaryCategory": "growth",
  "primaryCategoryText": "成长局",
  "secondaryCategory": "reading",
  "secondaryCategoryText": "读书局",
  "type": "deposit",
  "typeText": "押金局",
  "tags": ["自律", "共修"]
}
```

待确认：

1. 分类字段命名是否采用 `primaryCategory`、`secondaryCategory`、`type`，还是使用后端已有枚举字段。
2. 创建时是否允许多选二级细分；当前建议先单选，避免大厅卡片归属和筛选结果混乱。
3. `押金局` 是否允许作为成长局二级细分，同时也作为 `type=deposit`；若产品保留双重含义，后台需要同时存储两个字段。
4. `tags` 是否继续由前端固定推荐标签，还是也由后台按分类配置联动下发。

## 43. 地图附近信息点接口

记录日期：2026-06-25

模块：地图 / 组局分布

页面：`pages/map/index`

功能：地图页进入后获取当前用户位置，并按当前位置或地图当前视野半径向后台查询附近组局 / 地点信息点；前端把接口返回的数据转换为微信 `map` 组件 markers，点击 marker 后弹出基本信息卡。

候选接口：

```text
GET /api/app/games/nearby
```

建议入参：

```json
{
  "latitude": 31.2304,
  "longitude": 121.4737,
  "radiusMeters": 3000,
  "pageSize": 50
}
```

建议返回：

```json
{
  "center": {
    "latitude": 31.2304,
    "longitude": 121.4737
  },
  "radiusMeters": 3000,
  "nearestDistanceText": "157m",
  "onlinePlayerCount": 23,
  "offlinePlayerCount": 8,
  "total": 3,
  "list": [
    {
      "id": "game_001",
      "title": "鱼尾狮夜景打卡点",
      "cityName": "海尚广场",
      "latitude": 31.2326,
      "longitude": 121.4753,
      "distanceText": "420m",
      "memberText": "3/6人",
      "timeText": "今晚 20:00",
      "statusText": "探索局",
      "priceText": "¥0/人",
      "route": "pages/game/detail/index?id=game_001"
    }
  ],
  "onlinePlayers": [
    {
      "id": "player_online_001",
      "latitude": 31.2318,
      "longitude": 121.4745,
      "statusText": "在线玩家"
    }
  ],
  "offlinePlayers": [
    {
      "id": "player_offline_001",
      "latitude": 31.2282,
      "longitude": 121.4716,
      "statusText": "离线玩家"
    }
  ]
}
```

待确认：

1. 附近地图接口是否使用 `GET /api/app/games/nearby`，还是需要独立 `GET /api/app/map/nearby-points`。
2. 坐标字段是否统一为 GCJ-02 坐标系；微信小程序 `map` 与 `wx.getLocation({ type: 'gcj02' })` 当前按 GCJ-02 处理。
3. `radiusMeters` 最大值、默认值、分页策略和是否支持按地图视野 bounding box 查询需要后端确认。
4. 信息点类型是否只包含组局，还是还会包含打卡点、好友、城市图鉴；若包含多类型，需要返回 `pointType` 和对应详情跳转规则。
5. 地图图例里的在线 / 离线玩家当前按 `onlinePlayerCount`、`offlinePlayerCount` 展示总数，点位可由 `onlinePlayers`、`offlinePlayers` 提供抽样或全量，后端需确认返回策略。

## 44. 地图盲盒路线最近开启接口

记录日期：2026-06-25

模块：地图 / 盲盒路线

页面：`pages/map/blind-route/index`

功能：盲盒路线页“最近开启”列表由后台返回，前端不写死最终数据。页面展示最近开启过的盲盒路线、开启时间和当前状态，后续路线点击 / 完成流程接入后，需要根据后台返回状态刷新该区域。

候选接口：

```text
GET /api/app/map/blind-routes/recent
```

建议入参：

```json
{
  "pageSize": 5
}
```

建议返回：

```json
{
  "list": [
    {
      "id": "blind_route_001",
      "routeId": "tonight",
      "title": "今晚去哪局",
      "openedAt": "2026-06-24T18:30:00+08:00",
      "timeText": "昨天 18:30",
      "status": "completed",
      "statusText": "已完成"
    }
  ]
}
```

待确认：

1. 最近开启接口是否使用独立 `GET /api/app/map/blind-routes/recent`，还是合并到盲盒路线配置接口里一起返回。
2. `status` 枚举需要后端确认，例如 `opened/completed/cancelled/expired`。
3. 路线卡片点击后的开启、完成、状态回写接口后续再补充；当前页面只记录待实现，不接真实业务流程。

## 45. 关系网首页抽象层接口

记录日期：2026-06-25

模块：关系 / 关系网首页

页面：`pages/relation/network/index`

功能：关系网首页顶部浮层的地点标题、营业状态、地址、tab 文案和在线人数由抽象层接口返回，页面只负责渲染；左侧返回按钮和右侧刷新按钮是前端固定交互。

候选接口：

```text
GET /api/app/relations/network-home
```

建议返回：

```json
{
  "onlineText": "3999人在线",
  "header": {
    "titleIcon": "📍",
    "title": "星巴克(镇海万科店)",
    "statusText": "营业中",
    "address": "宁波市镇海区庄市大道1088号万科广场1F"
  },
  "tabs": [
    { "key": "network", "text": "人脉网络" },
    { "key": "nearby", "text": "附近玩家" }
  ],
  "activeTab": "network"
}
```

待确认：

1. 正式接口路径是否使用 `GET /api/app/relations/network-home`，还是合并到地图 / 附近信息接口。
2. 地点标题、地址和营业状态是否来自当前定位 POI、用户手动选择地点，还是后台推荐关系网中心点。
3. `tabs` 是否固定为 `人脉网络 / 附近玩家`，还是允许后台配置文案和默认选中项。

## 46. 消息交易预警详情接口

记录日期：2026-06-26

模块：消息 / 交易预警

页面：`pages/message/trade-warning/index`

功能：交易预警页内容从后台返回，包含预警文案、剩余交付时间、订单信息、交付方式和底部操作按钮文案。页面本身只负责渲染和本地交付方式选中态切换。

候选接口：

```text
GET /api/app/messages/trade-warning
```

建议入参：

```json
{
  "warningId": "trade-warning-001",
  "orderId": "GD2024032201"
}
```

建议返回：

```json
{
  "id": "trade-warning-001",
  "warningId": "trade-warning-001",
  "pageTitle": "交易预警",
  "onlineText": "3999人在线",
  "warning": {
    "title": "即将超时",
    "prefixText": "该订单将于",
    "highlightText": "1小时30分钟",
    "suffixText": "后自动标记为逾期，请立即处理"
  },
  "countdown": [
    { "value": "01", "label": "小时" },
    { "value": "30", "label": "分钟" },
    { "value": "45", "label": "秒" }
  ],
  "order": {
    "orderNo": "GD2024032201",
    "statusText": "待交付",
    "customerAvatarText": "CL",
    "customerTitle": "客户需求",
    "customerDesc": "寻找资深产品经理进行业务咨询",
    "detailRows": [
      { "label": "约定交付时间", "value": "今天 16:00" },
      { "label": "服务费用", "value": "¥500", "strong": true }
    ]
  },
  "deliveryMethods": [
    {
      "id": "online",
      "title": "线上确认",
      "desc": "双方在线确认服务完成",
      "active": true
    },
    {
      "id": "upload",
      "title": "上传凭证",
      "desc": "上传服务完成截图或文件",
      "active": false
    }
  ],
  "actions": {
    "delayText": "申请延期",
    "deliverText": "立即交付"
  }
}
```

待确认：

1. 正式接口路径是否使用 `GET /api/app/messages/trade-warning`，还是归到订单接口，例如 `GET /api/app/orders/{orderId}/trade-warning`。
2. 倒计时由后端直接返回展示文案 / 数字，还是返回 `serverTime` 与 `expectedDeliveryAt` 后由前端计算。
3. 交付方式枚举、默认选中项以及“立即交付 / 申请延期”的真实提交接口需要后端补充。
4. `warningId` 与 `orderId` 是否都需要传；若只用订单号即可定位预警，前端后续可简化入参。

## 47. 消息系统通知详情接口

记录日期：2026-06-26

模块：消息 / 系统通知

页面：`pages/message/system-detail/index`

功能：系统通知详情页文章由后台推送 / 返回，前端按后台返回的文章块渲染标题、作者、发布时间、阅读统计、正文段落、更新内容、封面、署名和反馈统计。

候选接口：

```text
GET /api/app/messages/system-notification
```

建议入参：

```json
{
  "messageId": "system-notification-001"
}
```

建议返回：

```json
{
  "id": "system-notification-001",
  "messageId": "system-notification-001",
  "pageTitle": "系统通知",
  "onlineText": "3999人在线",
  "article": {
    "tagText": "重要更新",
    "title": "组局功能全新升级：智能匹配系统上线",
    "author": "官方运营团队",
    "publishedAtText": "2026-03-20",
    "readText": "阅读 1.2k",
    "blocks": [
      { "id": "lead", "type": "paragraph", "text": "亲爱的用户：", "lead": true },
      { "id": "intro", "type": "paragraph", "text": "为了提升组局效率和匹配精准度，我们于今日正式上新智能匹配功能..." },
      {
        "id": "update-content",
        "type": "updateBox",
        "icon": "★",
        "title": "主要更新内容",
        "points": [
          "AI智能推荐：基于行为分析的个性化推荐",
          "匹配度评分：直观展示双方契合程度",
          "一键邀约：简化组局发起流程"
        ]
      },
      {
        "id": "cover",
        "type": "cover",
        "imageUrl": "/pages/message/system-detail/assets/system-update-cover.png",
        "caption": "智能匹配界面示意图"
      },
      {
        "id": "signature",
        "type": "signature",
        "teamText": "产品团队",
        "dateText": "2026年3月20日"
      }
    ]
  },
  "feedback": {
    "question": "这篇文章对你有帮助吗？",
    "useful": { "icon": "👍", "label": "有用", "count": 128, "countText": "128" },
    "useless": { "icon": "👎", "label": "没用", "count": 10, "countText": "10" }
  }
}
```

待确认：

1. 正式接口路径是否使用 `GET /api/app/messages/system-notification`，还是按消息 ID 使用 `GET /api/app/messages/{messageId}`。
2. 文章内容是否由后台直接下发块结构，还是返回富文本 / Markdown；当前页面先按块结构渲染，避免前端写死文章内容。
3. 阅读量、有用数、没用数是否由同一个详情接口返回；如果点击反馈需要实时回写，还需补充反馈提交接口。
4. 封面图如果来自后台 CDN，需确认小程序域名白名单和图片裁剪比例。

## 48. 消息中心列表接口

记录日期：2026-06-26

模块：消息 / 消息中心

页面：`pages/message/index`

功能：消息中心进入页面后从后台获取顶部消息分类、未读状态、tab 配置和消息列表。顶部分类右侧红点由后台未读状态控制：`unread=true` 或 `unreadCount > 0` 时显示红点；没有未读时不显示红点。点击 `全部消息 / 未读 / 交易通知` 时前端携带 tab key 重新请求列表。

候选接口：

```text
GET /api/app/messages/center
```

建议入参：

```json
{
  "tab": "all"
}
```

建议返回：

```json
{
  "pageTitle": "消息中心",
  "onlineText": "3999人在线",
  "activeTab": "all",
  "quickActions": [
    { "key": "join", "label": "组局加入", "iconSrc": "/pages/message/assets/i53@3x.png", "unreadCount": 1 },
    { "key": "system", "label": "系统通知", "iconSrc": "/pages/message/assets/i54@3x.png", "unreadCount": 0 },
    { "key": "achievement", "label": "成就解锁", "iconSrc": "/pages/message/assets/i55@3x.png", "unreadCount": 0 },
    { "key": "warning", "label": "预警通知", "iconSrc": "/pages/message/assets/i56@3x.png", "unreadCount": 0 },
    { "key": "friend", "label": "好友", "iconSrc": "/pages/message/assets/i57@3x.png", "unreadCount": 3 }
  ],
  "tabs": [
    { "key": "all", "label": "全部消息" },
    { "key": "unread", "label": "未读 (3)", "unreadCount": 3 },
    { "key": "trade", "label": "交易通知" }
  ],
  "sections": [
    {
      "key": "system",
      "title": "系统通知",
      "items": [
        {
          "id": "platform-notice",
          "routeKey": "system",
          "iconSrc": "/pages/message/assets/i58@3x.png",
          "title": "平台公告",
          "timeText": "2小时前",
          "desc": "关于组局功能升级的通知..."
        }
      ]
    }
  ]
}
```

待确认：

1. 正式接口路径是否使用 `GET /api/app/messages/center`，还是并入统一消息列表接口。
2. `quickActions` 的分类 key 是否固定为 `join/system/achievement/warning/friend`，以及每个分类点击后的正式跳转规则。
3. 未读统计是否只返回 `unreadCount`，还是同时返回 `unread`；前端当前两者都兼容。
4. 消息卡片按钮如 `确认参加 / 婉拒 / 立即处理 / 查看路线 / 联系发起人` 的真实操作接口仍需补充。

## 49. 我的个人中心子页面接口

记录日期：2026-06-26

模块：我的 / 个人中心

页面：`pages/profile/index`、`pages/profile/service-center/manage/review-manage/index`、`pages/profile/service-center/manage/review-reply/index`、`pages/profile/match-info/index`、`pages/profile/service-center/my-games/index`、`pages/profile/asset-center/manage/index`、`pages/profile/asset-center/points/index`、`pages/profile/settings/index`、`pages/profile/credit-center/index`、`pages/profile/service-center/invite-records/index`

功能：本次先按用户粘贴的个人中心 HTML 和 `E:\项目\03周总\个人中心` 下 6 个 HTML 导出资料落地静态走查页。`pages/profile/index` 已套用 `母版-申请加入` 对应的 `home-shell` `joinApply` 变体，并隐藏母版标题和右侧头像；其余 6 个子页使用白色母版。正式联调时，这些页面的用户信息、资产、列表、统计、开关、保存和操作按钮需要改为接口驱动。

候选接口：

```text
GET /api/app/profile/home
GET /api/app/profile/published-services
POST /api/app/profile/published-services/batch-offline
POST /api/app/profile/published-services/batch-delete
POST /api/app/profile/published-services/{serviceId}/offline
POST /api/app/profile/published-services/{serviceId}/online
POST /api/app/profile/published-services/{serviceId}/pin
GET /api/app/profile/match-info
PUT /api/app/profile/match-info
GET /api/app/profile/games
GET /api/app/profile/assets
POST /api/app/profile/assets/withdraw
POST /api/app/profile/assets/recharge
GET /api/app/profile/assets/balance-records
GET /api/app/profile/assets/bank-cards
GET /api/app/profile/assets/orders
GET /api/app/profile/points
GET /api/app/profile/points/records
GET /api/app/profile/settings
PUT /api/app/profile/settings
GET /api/app/profile/credit
GET /api/app/profile/invite-records
POST /api/app/profile/invite-records/{recordId}/remind-delivery
```

待确认：

1. 发布管理是否复用行家服务 / 商品接口，还是使用独立个人中心接口。
2. 适配信息字段是否来自个人主页、匹配画像，还是独立表单；地址定位是否需要 `chooseLocation`。
3. 我的组局和邀约记录是否复用组局模块已有接口，还是按个人中心聚合返回。
4. 系统设置里的支付密码、手机号、推送通知、邮件通知、隐私清单和缓存清理分别对应哪些正式接口。
5. 信用中心的信用分、奖励、惩罚和明细是否由一个聚合接口返回，奖励领取是否需要独立提交接口。

### 49.1 个人中心首页聚合接口

页面：`pages/profile/index`

候选接口：

```text
GET /api/app/profile/home
```

建议返回：

```json
{
  "user": {
    "nickname": "小明",
    "avatarUrl": "https://cdn.example.com/avatar.png",
    "avatarText": "小",
    "memberLevel": "基础会员",
    "growthLevel": "V5 探险家",
    "role": "玩家"
  },
  "stats": [
    { "key": "referrals", "label": "引荐数", "value": "128" },
    { "key": "successes", "label": "成功数", "value": "86" },
    { "key": "dealAmount", "label": "成交总额", "value": "¥45K" },
    { "key": "credit", "label": "信用度", "value": "98" }
  ],
  "assets": {
    "summary": [
      { "key": "totalDealAmount", "label": "总成交额", "value": "¥12,580" },
      { "key": "withdrawable", "label": "可提现", "value": "¥3,200", "tone": "green" },
      { "key": "pendingSettlement", "label": "待结算", "value": "¥800", "tone": "orange" }
    ],
    "vipBanner": {
      "text": "升级会员，认证您的角色",
      "actionText": "增购会员 >",
      "route": "pages/profile/member/index"
    }
  },
  "sections": [
    {
      "key": "service",
      "title": "服务中心",
      "items": [
        {
          "key": "myGames",
          "title": "我的局",
          "icon": "i66",
          "badgeText": "2进行中",
          "badgeTone": "pink",
          "route": "pages/profile/service-center/my-games/index",
          "enabled": true
        }
      ]
    }
  ]
}
```

当前首页入口路由口径：

| 分组 | 入口 | 当前路由 / 状态 |
| --- | --- | --- |
| 服务中心 | 我的局 | `pages/profile/service-center/my-games/index` |
| 服务中心 / 组局管理 | 评价管理 | `pages/profile/service-center/manage/review-manage/index` |
| 服务中心 / 组局管理 | 回复评价 | `pages/profile/service-center/manage/review-reply/index` |
| 服务中心 | 我的邀请 | `pages/profile/service-center/invite-records/index` |
| 资产中心 | 我的资产 | `pages/profile/asset-center/manage/index` |
| 资产中心 | 我的押金 | 待确认押金页 / 路由 |
| 资产中心 | 积分商城 | `pages/profile/asset-center/mall/index` |
| 资产中心 | 我的订单 | `pages/profile/asset-center/orders/index` |
| 资产中心 | 我的积分 | `pages/profile/asset-center/points/index` |
| 资产中心 | 开票中心 | 待确认开票页 / 路由 |
| 足迹中心 | 我的足迹 | 待确认足迹页 / 路由 |
| 足迹中心 | 我的城市故事 | 待确认城市故事页 / 路由 |
| 足迹中心 | 我的成就墙 | `pages/profile/footprint/achievements/index` |
| 账户管理 | 我的资料 | `pages/profile/match-info/index` |
| 账户管理 | 技能配置 | 待确认技能配置页 / 路由 |
| 账户管理 | 屏蔽设置 | 待确认屏蔽设置页 / 路由 |
| 账户管理 | 信用中心 | `pages/profile/credit-center/index` |
| 账户管理 | 举报中心 | 待确认举报中心页 / 路由 |
| 账户管理 | 签署协议 | 待确认协议页 / 路由 |
| 账户管理 | 建议反馈 | 待确认反馈页 / 路由 |
| 账户管理 | 系统设置 | `pages/profile/settings/index` |

待确认：

1. 首页菜单入口是否全部由后端返回，还是前端固定入口、后端只返回角标和可见状态。
2. `badgeText`、`badgeTone` 是否由后端直接返回；如果后端只返回数量和状态枚举，前端需要统一映射文案和颜色。
3. 资产中心当前已新增 `pages/profile/asset-center/manage/index`、`pages/profile/asset-center/mall/index`、`pages/profile/asset-center/orders/index` 和 `pages/profile/asset-center/points/index` 静态走查页；我的押金、开票中心是否继续放在该目录下仍需确认。
4. 足迹中心是否复用地图模块已有页面，例如 `pages/map/my-city/index`、`pages/map/footprint-heatmap/index`，需要产品确认。
5. 头像使用 `avatarUrl` 真实图片还是继续允许 `avatarText` 兜底。

### 49.2 资产中心资产管理与积分中心接口

页面：`pages/profile/asset-center/manage/index`、`pages/profile/asset-center/mall/index`、`pages/profile/asset-center/orders/index`、`pages/profile/asset-center/points/index`

候选接口：

```text
GET /api/app/profile/assets
POST /api/app/profile/assets/withdraw
POST /api/app/profile/assets/recharge
GET /api/app/profile/assets/balance-records
GET /api/app/profile/assets/bank-cards
GET /api/app/profile/assets/orders
GET /api/app/profile/points
GET /api/app/profile/points/records
GET /api/app/profile/points/mall
POST /api/app/profile/points/mall/exchange
GET /api/app/profile/points/orders
GET /api/app/profile/points/orders/{orderId}/logistics
POST /api/app/profile/points/orders/{orderId}/cancel
```

需要后台返回 / 确认：

1. 资产管理页全部展示数据都需由后台提供，前端当前只保留静态走查兜底：总资产、总成交额、可提现、待结算、银行卡绑定数量、银行卡列表摘要、订单状态数量、各订单状态对应的订单列表、最近订单、FAQ 文案和是否展示充值入口。
2. 订单状态区的 `待付款 / 进行中 / 已完成 / 退款/售后 / 待评价` 不应由前端写死最终数量；需要后台返回状态枚举、状态文案、数量和点击后对应筛选条件，进入列表时按状态查询对应订单。
3. 银行卡入口需要后台返回当前绑定数量、是否已实名 / 可绑卡状态、银行卡列表、默认收款账户和绑卡 / 解绑银行卡操作接口。
4. 提现 / 充值 / 余额明细 / 银行卡 / 我的订单分别跳转页面还是当前页弹层，真实接口路径、提交字段、失败提示和结果页仍需确认。
5. 积分中心页全部展示数据都需由后台提供，前端当前只保留静态走查兜底：可用积分、累计积分、已兑换、过期积分、积分规则、唯一积分来源、积分比例、计算规则、角色分润示例、积分明细记录、收入 / 支出筛选和分页。
6. 角色分润积分示例不由前端计算；后台需要直接返回各角色分成比例、分润金额、积分比例、积分结果和展示顺序，前端只按返回数据渲染红框中的示例列表。
7. 积分明细列表不由前端拼接；后台需要返回每条明细的类型、标题、说明、时间、订单号 / 关联单号、积分增减值、收入 / 支出分类、图标类型和分页信息。
8. 积分明细筛选 `全部 / 收入 / 支出` 的枚举 key、分页参数、过期 / 兑换 / 收入记录的字段与金额格式需后端确认。
9. 积分商城页全部展示数据都需由后台提供，前端当前只保留静态走查兜底：可用积分、商品列表、商品图标 / 图片、商品名称、兑换积分、库存、是否可兑换、兑换按钮状态和积分有效期提示。
10. 积分商城点击商品后在当前页打开确认兑换弹层；点击 `确认兑换` 调用 `POST /api/app/profile/points/mall/exchange`，前端根据后台 `code/message/data` 判断兑换成功或失败，并在当前页弹出结果提示框，不跳转到新页面。
11. `POST /api/app/profile/points/mall/exchange` 建议入参为 `{ "goodId": "mall-shirt" }`，成功返回建议包含 `orderId`、`orderStatus` 和最新商城快照 `mall`；业务失败时建议返回非 0 `code` 和可直接展示的 `message`，例如库存不足、积分不足、商品下架。
12. 无论兑换成功或失败，前端都会再次调用 `GET /api/app/profile/points/mall` 查询最新积分余额和商品剩余数量，避免使用本地扣减结果作为最终库存。后续正式联调时需确认该查询接口是否也返回最新积分明细入口数据。
13. 我的订单页全部展示数据都需由后台提供；当前前端已在进入页面和切换状态 tab 时调用 `GET /api/app/profile/points/orders`，并携带 `status` 查询条件（`all` 不传状态），接口返回 `tabs`、`orders/list`、`emptyText` 后渲染。订单状态 tab、订单号、商品信息、兑换积分、兑换时间、状态、可执行动作都应以后端返回为准。订单动作里 `再次兑换` 当前直接跳转 `pages/profile/asset-center/mall/index`；`查看物流` 当前在我的订单当前页打开物流详情弹层，并调用 `GET /api/app/profile/points/orders/{orderId}/logistics` 获取物流详情；`查看详情` 和 `取消订单` 都是跳转独立页面的需求，但具体页面、路由、展示字段、取消流程和结果页需业主确认后再实现；订单详情和取消订单仍需补充真实接口。
14. 物流详情当前按用户提供设计落地为我的订单页内紧凑深色弹层，单独物流详情页已删除；接口建议返回 `courier.name`、`courier.trackingNo`、`timeline[]`，其中轨迹节点建议包含 `id`、`desc`、`time`、`active/current`；前端复制按钮复制 `trackingNo`。正式联调需确认快递公司图标来源、轨迹排序方向、无物流 / 物流异常 / 已签收状态、弹层刷新策略和接口失败提示。
15. 积分兑换成功后如何同步刷新积分明细和订单列表，以及取消订单是否退回积分、退回后积分有效期如何处理，仍需后端 / 产品确认。

### 49.3 服务中心我的邀请子页接口

页面：`pages/profile/service-center/invite/overview/index`、`pages/profile/service-center/invite/network/index`、`pages/profile/service-center/invite/records/index`、`pages/profile/service-center/invite/ranking/index`、`pages/profile/service-center/invite/income/index`、`pages/profile/service-center/invite/member-detail/index`

候选接口：

```text
GET /api/app/profile/service-center/invite/overview
GET /api/app/profile/service-center/invite/network
GET /api/app/profile/service-center/invite/records
GET /api/app/profile/service-center/invite/ranking
GET /api/app/profile/service-center/invite/income
GET /api/app/profile/service-center/invite/members/{memberId}
POST /api/app/profile/service-center/invite/share
POST /api/app/profile/service-center/invite/records/{recordId}/remind-delivery
```

待确认：

1. 数据概览页需要返回用户等级、昵称、唯一 ID、加入天数、核心指标、快捷操作可用状态、趋势速览数据。
2. 数据概览页的三个快捷操作均需后台提供数据：`分享邀请码` 使用后台返回的 `inviteCode` 和分享文案 / 分享卡片配置；中间操作按 `二维码` 处理，不再使用 `复制链接` 口径，后台需返回可展示 / 保存的二维码或小程序码字段，例如 `qrCodeUrl` 或 `miniProgramCodeUrl`；`生成海报` 使用后台生成或返回的带邀请码海报图，字段建议包含 `posterImageUrl`，海报内也应包含二维码或小程序码。前端不写死邀请码、链接或二维码内容。
3. 关系网络页需要返回已服务玩家数、本周收益、一级成员关系图节点、一级成员列表、成员直接贡献和团队贡献。
4. 邀约记录页需要返回前端当前展示的所有数据：`我引荐的 / 我发起的` tab 数量与选中态、状态筛选数量、超时预警文案、记录列表、状态、记录编号、时间、行家 / 玩家头像昵称与角色、关系节点、服务标题、预算、你的奖励、已到账状态、超时 / 取消提示、提醒交付和查看组局入口可用状态。前端当前仅静态兜底，不写死最终业务数据；查看组局需返回可跳转的组局 / 订单 / 服务记录 ID。
5. 贡献排行页需要返回周期筛选、排行类型、成员排名、头像、活跃天数、邀约数和分润贡献数据。
6. 收益明细页需要返回收益趋势、本月分润、累计分润、活跃成员、产生分润局数和分润流水。
7. 成员详情页需要按 `memberId` 返回成员基础信息、邀约 / 转化 / 转化率、直接贡献收益、团队贡献收益、合计贡献和最近动态。
8. 当前 6 页均为静态走查；正式联调时需要确认页面之间的跳转关系，例如关系网络成员点击进入成员详情、查看分润进入收益明细、邀约记录查看详情进入对应组局详情。

### 49.4 服务中心评价管理接口

页面：`pages/profile/service-center/manage/review-manage/index`、`pages/profile/service-center/manage/review-reply/index`

候选接口：

```text
GET /api/app/profile/service-center/reviews
GET /api/app/profile/service-center/reviews/{reviewId}
GET /api/app/profile/service-center/reply-templates
GET /api/app/profile/service-center/reviews/{reviewId}/messages
POST /api/app/profile/service-center/reviews/{reviewId}/reply
POST /api/app/profile/service-center/reviews/{reviewId}/like
```

待确认：

1. 评价管理页需要返回评分统计、待回复数量、评价列表、标签、平台介入状态、行家回复内容、点赞状态和点赞数量；字段建议包含 `scoreSummary`、`pendingCount`、`reviews[]`、`reviews[].reply`、`reviews[].statusType`、`reviews[].liked`、`reviews[].likeCount`。
2. 回复评价页需要按 `reviewId` 获取单条评价详情，字段至少包含玩家头像 / 昵称、评价时间、评分、服务标题、评价内容、标签、订单号、服务金额和已有行家回复。
3. 快捷回复模板当前为前端静态兜底，后续应由后台返回模板列表，字段建议包含 `id`、`content`、`sort`、`enabled`。
4. 历史对话当前为前端静态兜底，后续应按 `reviewId` 获取，字段建议包含 `id`、`role`、`name`、`avatarUrl`、`content`、`createdAt`。
5. 回复评价页提交时，回复内容最大长度按 200 字处理；正式接口需要校验 `content` 非空且不超过 200 字。响应建议返回 `reviewId`、`reply.content`、`reply.repliedAt`、`statusType`，用于刷新评价管理页的“行家回复”内容。
6. 点赞按钮当前是图标按钮，后续需确认是否允许行家给玩家评价点赞、是否需要取消点赞，以及接口返回字段命名。

## 50. 会员中心三档会员配置接口

记录日期：2026-06-26

模块：我的 / 会员中心

页面：`pages/profile/member/index`、`pages/profile/member/advanced/index`、`pages/profile/member/premium/index`

功能：会员中心基础会员、高级会员、尊享会员 3 个页面的页面内容由后台按会员等级配置返回，便于运营或后台自由修改。当前页面静态展示三档会员的标题、权益、分润比例、适配人群、开通权益说明、最近开通会员好友提示和开通价格；正式实现时这些文案、价格、权益配置和按钮提示都不应写死在前端。

候选接口：

```text
GET /api/app/profile/member-center
```

建议入参：

```json
{
  "level": "basic"
}
```

建议返回：

```json
{
  "pageTitle": "会员中心",
  "level": {
    "key": "basic",
    "name": "基础会员",
    "cardWatermark": "VIP",
    "activeIndex": 0
  },
  "benefitsSection": {
    "title": "会员权益",
    "items": [
      {
        "key": "referral",
        "title": "业务引荐权益",
        "iconKey": "i86",
        "theme": "purple",
        "points": ["可引荐平台业务", "享受引荐收益"]
      },
      {
        "key": "profit",
        "title": "利润分成",
        "iconKey": "i87",
        "theme": "gold",
        "value": "40%"
      }
    ]
  },
  "radarSection": {
    "title": "组局雷达",
    "subtitle": "精准匹配附近组局，可平级参与组局",
    "buttonText": "开始适配"
  },
  "audienceSection": {
    "title": "适配人群",
    "items": [
      "有人脉、善对接的社交达人、资源型人才；",
      "希望不做销售、不投重金，只靠人脉赚钱；",
      "有高客单价产品/资源，想初步了解；",
      "连接供需，促成交易"
    ]
  },
  "openRulesSection": {
    "title": "开通权益说明",
    "stepText": "1. 选择会员等级 → 2. 在线支付 → 3. 即时生效",
    "notes": [
      "支持微信支付、支持银行卡支付",
      "升级后原有权益自动叠加，不重复收费",
      "如需帮助，请联系客服：400-XXX-XXXX"
    ]
  },
  "roleLink": {
    "text": "已是平台会员，前去解锁角色",
    "route": "pages/role/apply/index"
  },
  "latestNotice": {
    "avatarText": "林",
    "name": "林丽 总",
    "text": "刚开通了高级会员"
  },
  "purchase": {
    "price": 515,
    "priceText": "¥ 515 /年",
    "buttonText": "立即开通",
    "agreementPrefix": "请阅读",
    "agreementName": "《服务协议》",
    "agreementRoute": "pages/agreement/member-service/index",
    "agreementSuffix": "，购买视为确认协议。",
    "highlightText": "开通后需完成身份认证、解锁角色权益"
  }
}
```

字段口径：

| 字段 | 来源 | 用途 |
| --- | --- | --- |
| `level.name` | 后台 | 会员卡标题，例如 `基础会员`、`高级会员`、`尊享会员`。 |
| `benefitsSection.items` | 后台 | 权益卡标题、图标、说明和分润比例。 |
| `radarSection` | 后台 | 雷达区域标题、副标题和按钮文案。 |
| `audienceSection.items` | 后台 | 适配人群 4 行文案，可由后台调整数量和内容。 |
| `openRulesSection` | 后台 | 开通流程、支付方式、自动叠加说明和客服电话。 |
| `latestNotice` | 后台 | 最近开通会员好友提示，若无数据可返回 `null`，前端隐藏该条。 |
| `purchase.priceText` | 后台 | 开通价格展示文案，价格可后台修改，前端不拼死 `515`。 |
| `purchase.buttonText` | 后台 | 底部开通按钮文案。 |
| `purchase.agreement*` | 后台 / 配置 | 服务协议展示文案和点击路径。 |

待确认：

1. 正式接口路径是否使用 `GET /api/app/profile/member-center`，还是并入个人中心聚合接口 `GET /api/app/profile/home`。
2. 价格字段使用元还是分；展示建议由后台直接返回 `priceText`，前端仅展示。
3. 后台需要支持 `basic`、`advanced`、`premium` 三档会员分别配置，至少包含基础会员、高级会员、尊享会员的权益卡、价格、分润比例、适配人群和卡片样式。
4. 最近开通会员提示是否展示真实好友、平台会员动态，还是运营配置文案；涉及用户昵称时需确认隐私口径。
5. `立即开通` 点击后的支付预下单接口、支付成功后会员生效和身份认证 / 解锁角色流程仍需补充正式接口。

## 51. 会员中心人脉雷达流程接口

记录日期：2026-06-26

模块：我的 / 会员中心 / 人脉雷达

页面：`pages/profile/member/radar/index`、`pages/profile/member/match/index`、`pages/profile/member/match-info/index`、`pages/profile/member/match-query/index`、`pages/profile/member/match-result/index`

功能：会员中心人脉雷达 5 页目前为静态走查页，正式实现需要后端支持雷达概览、适配信息表单、发起匹配、匹配进度、推荐结果列表、下一位、关注、查看个人主页和重新匹配。

候选接口：

```text
GET /api/app/profile/member-radar/overview
GET /api/app/profile/member-radar/profile
POST /api/app/profile/member-radar/profile
POST /api/app/profile/member-radar/match
GET /api/app/profile/member-radar/match/{matchId}
GET /api/app/profile/member-radar/match/{matchId}/results
POST /api/app/profile/member-radar/results/{resultId}/follow
```

`POST /api/app/profile/member-radar/match` 发起匹配规则：

- 点击 `开启适配人脉` 后进入 `pages/profile/member/match/index` 搜索页。
- 前端提交 `criteria` 参数；若 `criteria` 为空对象或所有字段为空，后端按全部可匹配人脉池进行匹配。
- 若 `criteria` 中存在有效字段，则后端按地址、行业、营收规模、感兴趣组局、我的资源、我的需求、近期诉求等条件筛选后匹配。
- 搜索页只展示匹配中状态和返回按钮，不展示右侧搜索 / 分享按钮；匹配完成后再进入结果查询或最终结果页。

建议发起匹配请求体：

```json
{
  "criteria": {
    "location": "",
    "industry": "",
    "revenueScale": "",
    "interestedGames": "",
    "resources": "",
    "needs": "",
    "recentDemand": ""
  }
}
```

`pages/profile/member/match-query/index` 匹配结果查询页规则：

- 页面顶部的适配行家数量、当前行家卡片信息均从前面匹配查询接口获取，不在前端写死。
- 进入查询页时，优先携带 `matchId` 调用 `GET /api/app/profile/member-radar/match/{matchId}/results` 获取 `total`、当前展示行家和结果列表。
- 行家卡片字段包括头像、姓名、职位、第一标签、我的需求、我的资源、地址、距离、个人主页跳转路径、是否已关注。
- `下一个` 优先在已返回的 `results` 列表中切换下一位；若列表不足或需要服务端排序，携带 `matchId`、`cursor`、`excludeResultIds` 继续请求下一页 / 下一条。
- `关注` 点击调用 `POST /api/app/profile/member-radar/results/{resultId}/follow`，成功后更新当前卡片关注状态，避免重复关注；失败时提示失败原因。
- `pages/profile/member/match-result/index` 最终页展示的推荐行家数量复用前面匹配结果的 `total`，不在前端写死。
- `重新查看` 点击回到 `pages/profile/member/match-query/index` 查看已经扫描出的行家结果，复用当前 `matchId` 和已获取的结果列表，不重新调用发起匹配接口，不进入搜索扫描页。
- `再次重新匹配` 点击重新调用 `POST /api/app/profile/member-radar/match`，仍按适配信息是否为空决定全量匹配或条件匹配，并进入搜索页展示新的匹配进度。

建议匹配结果查询返回：

```json
{
  "matchId": "radar-20260626-001",
  "total": 10,
  "cursor": "page-2",
  "currentIndex": 0,
  "results": [
    {
      "id": "result-001",
      "expertId": "expert-001",
      "name": "陆毅",
      "avatarUrl": "https://cdn.example.com/avatar.png",
      "title": "总经理｜上海创世界科技有限公司",
      "firstTag": "第一标签：上海TMT投资领军者，数字化内容服务",
      "need": "AI赋能与市场运营助力企业IP打造",
      "resource": "10年TMT投资经验",
      "address": "上海市浦东新区沙新镇黄赵路310号",
      "distanceText": "231 km",
      "profileRoute": "/pages/home-other/index?userId=expert-001",
      "followed": false
    }
  ]
}
```

建议返回：

```json
{
  "overview": {
    "title": "人脉雷达",
    "estimateText": "预计可匹配7461位商界决策者",
    "primaryActionText": "开启适配人脉",
    "tip": "信息填写越完整，人脉匹配越精准"
  },
  "profileForm": {
    "fields": [
      { "key": "location", "label": "地址定位", "value": "", "placeholder": "选择", "type": "location" },
      { "key": "industry", "label": "所在行业", "value": "", "placeholder": "选择", "type": "picker" },
      { "key": "revenue", "label": "营收规模", "value": "", "placeholder": "选填", "type": "picker" },
      { "key": "interest", "label": "感兴趣组局", "value": "", "placeholder": "选择", "type": "picker" },
      { "key": "resources", "label": "我的资源", "value": "", "placeholder": "前往个人主页填写", "type": "profileLink" },
      { "key": "needs", "label": "我的需求", "value": "", "placeholder": "前往个人主页填写", "type": "profileLink" },
      { "key": "recentDemand", "label": "近期诉求", "value": "", "placeholder": "自定义填写", "type": "textarea" }
    ]
  },
  "match": {
    "matchId": "radar-20260626-001",
    "status": "completed",
    "loadingText": "人脉雷达正在寻找与您适配的企业家…",
    "foundText": "为您找到 10 位适配您的商界决策者",
    "total": 10
  },
  "results": [
    {
      "id": "result-001",
      "name": "胡芳",
      "avatarUrl": "https://cdn.example.com/avatar.png",
      "title": "董事长、创始人｜千浪化研新材料（上海…",
      "firstTag": "第一标签：手机漆，汽车漆深耕者",
      "industry": "化学原料和化学制品制造业",
      "supply": "化工涂料的生产销售,专业的塑胶工业漆及手机漆生产者",
      "need": "期待与更多需要油漆涂料的岛亲链接交流",
      "city": "上海",
      "distanceText": "231 km",
      "profileRoute": "/pages/home-other/index?userId=result-001"
    }
  ]
}
```

待确认：

1. 人脉雷达和现有组局推荐是否共用匹配算法与接口，还是会员中心独立接口。
2. `适配信息` 表单字段、选择项、必填校验、地址定位权限和保存失败提示需产品 / 后端确认。
3. 匹配结果卡片展示字段、头像来源、个人主页跳转路径和隐私口径需确认。
4. `下一位` 是前端翻页还是后端重新取下一条；`关注` 是否走行家关注接口，是否需要兼容好友申请 / IM 关系链。
5. 匹配完成页的 `重新查看`、`再次重新匹配` 是否复用同一 `matchId`，重新匹配是否扣次数或有会员等级限制。

## 52. 会员中心适配信息填写接口

记录日期：2026-06-26

模块：我的 / 会员中心 / 适配信息

页面：`pages/profile/member/match-info/index`

功能：适配信息页面需要从后台获取待填写字段、字段当前值、选择项、必填状态和保存规则，并在用户填写后提交保存。当前静态字段包括：地址定位、所在行业、营收规模、感兴趣组局、我的资源、我的需求、近期诉求。

候选接口：

```text
GET /api/app/profile/member-radar/profile-form
POST /api/app/profile/member-radar/profile-form
GET /api/app/profile/member-radar/profile-form/options
```

建议字段：

| 字段 | 类型 | 当前文案 | 说明 |
| --- | --- | --- | --- |
| `location` | 地址 / 定位 | 地址定位 | 需确认是否调用 `chooseLocation`，以及保存经纬度、城市、详细地址。 |
| `industry` | 选择项 | 所在行业 | 需后台返回行业树或行业列表。 |
| `revenueScale` | 选择项 / 选填 | 营收规模 | 需后台返回营收规模选项，并确认是否必填。 |
| `interestedGames` | 多选 | 感兴趣组局 | 需后台返回可选组局类型 / 标签。 |
| `resources` | 文本 / 个人主页同步 | 我的资源 | 当前提示前往个人主页填写，需确认是否在本页直接编辑或跳转个人主页。 |
| `needs` | 文本 / 个人主页同步 | 我的需求 | 当前提示前往个人主页填写，需确认是否在本页直接编辑或跳转个人主页。 |
| `recentDemand` | 文本输入 | 近期诉求 | 自定义填写，需确认最大字数、敏感词和审核规则。 |

待确认：

1. 适配信息是否与个人中心 `pages/profile/match-info/index` 共用接口和数据，还是会员中心单独保存。
2. 每个字段是否必填，以及保存前是否允许只保存部分字段。
3. 地址定位是否必须授权 `scope.userLocation`，拒绝授权时的兜底填写方式。
4. `我的资源`、`我的需求` 是读取个人主页资料还是会员雷达独立资料。
5. 保存成功后跳转回 `组局雷达` 还是进入 `适配组局` 流程。

## 54. 服务中心我的邀请 - 贡献排行页接口

记录日期：2026-06-27

模块：我的 / 服务中心 / 我的邀请

页面：`pages/profile/service-center/invite/ranking/index`

功能：贡献排行页当前为静态走查，正式联调时展示当前用户的一级成员贡献榜单。页面数据来自后台，前端不写死最终业务数据。

候选接口：

```text
GET /api/app/profile/service-center/invite/ranking
```

建议入参：

```json
{
  "period": "week",
  "rankType": "inviteCount",
  "page": 1,
  "pageSize": 20
}
```

需要后台返回 / 确认：

1. 周期筛选：本周、本月、本季、本年、全部的 key、文案、默认选中项和可见性。
2. 排行类型：邀约数排行、分润贡献排行的 key、文案、默认选中项和排序规则。
3. 榜单范围：仅统计当前用户的一级成员数据，不混入二级 / 团队汇总成员。
4. 成员字段：成员 ID、排名、头像、昵称、等级文案、活跃天数、邀约数、分润贡献金额。
5. 分页、空态、错误态、金额格式、同分排序和数据更新时间。

## 55. 服务中心我的邀请 - 收益明细页接口

记录日期：2026-06-27

模块：我的 / 服务中心 / 我的邀请

页面：`pages/profile/service-center/invite/income/index`

功能：收益明细页当前为静态走查，正式联调时收益趋势、核心指标和分润流水均由后台返回，前端不写死 10 个月或固定流水数量。

候选接口：

```text
GET /api/app/profile/service-center/invite/income
GET /api/app/profile/service-center/invite/income/flows
```

建议入参：

```json
{
  "year": 2026,
  "flowPage": 1,
  "flowPageSize": 20
}
```

需要后台返回 / 确认：

1. 收益趋势：返回一年内实际有展示意义的月份数组，例如 `trendSeries[].month`、`trendSeries[].amount`；前端按返回数量渲染月份和趋势点，并按每个月 `amount` 在本组数据中的比例换算趋势点纵坐标，不固定 10 个月。
2. 核心指标：本月分润、累计分润、活跃成员、产生分润局数、环比文案和累计笔数，字段建议包含 `metrics[]` 的 label、value、desc。
3. 分润流水：流水 ID、成员 ID、成员昵称、组局 / 服务标题、发生时间、分润金额、图标 / 类型、状态。
4. 流水数量多时需要分页或游标加载，字段建议包含 `page`、`pageSize`、`hasMore`、`nextCursor`。
5. 金额格式、空态、错误态、时间格式、退款 / 冲正 / 已结算状态是否展示需后端确认。

## 56. 服务中心我的邀请 - 成员详情页接口

记录日期：2026-06-27

模块：我的 / 服务中心 / 我的邀请

页面：`pages/profile/service-center/invite/member-detail/index`

功能：成员详情查看页当前为静态走查，从关系网络一级成员点击进入并携带 `memberId`。正式联调时成员资料、统计、收益贡献和最近动态均由后台返回。

候选接口：

```text
GET /api/app/profile/service-center/invite/members/{memberId}
```

需要后台返回 / 确认：

1. 成员基础信息：成员 ID、头像 URL 或兜底头像、昵称、等级文案、关系层级。
2. 成员统计：总邀约、成功转化、转化率。
3. 收益贡献：直接贡献收益、团队贡献收益、合计贡献金额。
4. 最近动态：动态 ID、类型图标 / 类型 key、标题、发生时间、金额或状态符号。
5. 空态、错误态、成员不存在 / 无权限访问、金额格式和时间格式。

## 59. 资产中心我的订单与物流弹层接口

记录日期：2026-06-27

模块：我的 / 资产中心 / 我的订单

页面：`pages/profile/asset-center/orders/index`

功能：我的订单页按后台订单状态查询订单列表，查看物流在当前订单页打开紧凑弹层并从后台获取物流详情；单独物流详情页已删除，不再作为当前查看物流入口。

候选接口：

```text
GET /api/app/profile/points/orders
GET /api/app/profile/points/orders/{orderId}/logistics
POST /api/app/profile/points/orders/{orderId}/cancel
```

需要后台返回 / 确认：

1. 订单列表字段：订单 ID / 订单号、商品名、商品图片或图标、积分数量、兑换时间、状态 key、状态文案、按钮列表和可操作状态。
2. 状态 tab 枚举：全部、待发货、配送中、已完成是否固定，以及各状态对应接口参数、数量、空态和分页规则。
3. 物流详情字段：快递公司、运输单号、物流轨迹、轨迹时间、当前节点、无物流 / 物流异常 / 已签收状态和错误文案。
4. 查看物流当前用订单页弹层展示，正式联调需确认弹层刷新策略、复制失败提示、轨迹排序和接口失败提示。
5. `查看详情` 和 `取消订单` 暂未实现；需要业主确认页面形态、路由、展示字段、取消流程、取消结果页和取消接口。

状态：前端已接订单列表和物流详情 mock，正式接口待后端确认。

## 60. 我的足迹 - 我的成就墙接口

记录日期：2026-06-27

模块：我的 / 我的足迹

页面：`pages/profile/footprint/achievements/index`

功能：我的成就墙当前为蓝色游戏母版 + 黑色内容区静态走查页，展示成就等级、升级进度、成就分类、已获得 / 未解锁成就和赛季信息。正式联调时页面数据和成就状态应由后台返回。

候选接口：

```text
GET /api/app/profile/footprint/achievements
GET /api/app/profile/footprint/achievements/{achievementId}
POST /api/app/profile/footprint/achievements/{achievementId}/claim
```

需要后台返回 / 确认：

1. 成就等级：等级标题、等级数值、当前经验 / 足迹值、下一级所需值、升级提示和进度百分比。
2. 分类标签：全部、点亮城市、连续打卡、隐藏成就等分类是否固定，分类 key、展示文案和每类数量。
3. 成就列表：成就 ID、标题、描述、图标 URL 或图标 key、分类、已获得 / 未解锁状态、完成进度、是否隐藏、是否可领取奖励。
4. 赛季信息：赛季名称、状态、剩余时间、已获得数量、未解锁数量和赛季奖励。
5. 交互规则：点击成就是进入详情、弹层展示还是领取奖励；领取奖励的成功 / 失败提示、幂等处理和刷新策略。
6. 空态、错误态、分页 / 展示数量、图标资源来源、时间格式和多语言 / 文案后台配置。

状态：前端已新增静态页和编译模式，正式接口待后端确认。

## 61. 系统管理我的资料接口

记录日期：2026-06-27

模块：我的 / 系统管理 / 我的资料

页面：`pages/profile/system-management/profile-info/index`

功能：我的资料页当前为静态走查；点击右上角保存时，前端已组装当前个人资料、企业资料、联系方式可见性、公开注册业务信息开关和认证状态，通过 `PUT /api/app/profile/system-management/profile-info` 提交到后台 / mock。正式联调时还需要从后台获取个人资料、企业资料、联系方式可见性、公开注册业务信息开关、认证状态，并支持头像更新、实名认证 / 企业认证入口。

候选接口：

```text
GET /api/app/profile/system-management/profile-info
PUT /api/app/profile/system-management/profile-info
POST /api/app/profile/system-management/profile-info/avatar
GET /api/app/profile/system-management/certifications
POST /api/app/profile/system-management/certifications/personal
POST /api/app/profile/system-management/certifications/enterprise
```

需要后台返回 / 确认：

1. 个人信息字段：头像 URL / 头像文字、姓名、脱敏联系方式、联系方式是否隐藏、可见性枚举、兴趣爱好。
2. 企业信息字段：公司名称、职务、主营业务数量或列表、可提供资源、公开注册业务信息开关。
3. 认证中心字段：个人身份认证状态、企业认证状态、状态文案、认证说明、认证入口是否可点。
4. 保存接口字段、校验规则、敏感词 / 审核规则、部分保存还是整体保存、保存失败码和提示文案；当前前端保存 payload 为 `personalInfo`、`enterpriseInfo`、`certifications` 三段。
5. 头像上传使用文件上传还是后台返回上传凭证；联系方式展示和可见性涉及隐私口径，需要产品 / 后端确认。
6. 认证入口跳转小程序内页、H5、第三方小程序还是后台返回跳转参数。

状态：前端已新增静态页和编译模式，并已接 `PUT /api/app/profile/system-management/profile-info` service / mock 保存链路；正式接口字段、校验规则和获取接口仍待后端确认。

## 62. 系统管理技能配置接口

记录日期：2026-06-27

模块：我的 / 系统管理 / 技能配置

页面：`pages/profile/system-management/skill-config/index`

功能：技能配置页当前为深色参考图样式的静态走查；进入页面时前端尝试读取技能配置，点击右上角保存时组装当前身份限制、技能槽位、显性技能、隐形技能、服务案例和解锁提示，通过 `PUT /api/app/profile/system-management/skill-config` 提交到后台 / mock。当前前端已补充本地响应交互：移除技能确认、修改次数不足提示、编辑技能底部面板、解锁新技能确认和添加技能底部面板；这些操作先按页面本地状态更新并扣减剩余修改次数。正式联调时技能槽位、技能列表、修改次数、AI 解锁来源、案例绑定和可操作状态均应由后台返回。

候选接口：

```text
GET /api/app/profile/system-management/skill-config
PUT /api/app/profile/system-management/skill-config
GET /api/app/profile/system-management/skill-config/cases/{caseId}
POST /api/app/profile/system-management/skill-config/skills
PATCH /api/app/profile/system-management/skill-config/skills/{skillId}
DELETE /api/app/profile/system-management/skill-config/skills/{skillId}
POST /api/app/profile/system-management/skill-config/skills/{skillId}/unlock
POST /api/app/profile/system-management/skill-config/cases/{caseId}/bind
```

需要后台返回 / 确认：

1. 身份与限制：当前身份、最多可配置技能数、每月可修改次数、已用次数、剩余次数、超限后的提示文案和是否允许保存。
2. 技能槽位：槽位 ID、标题、图标 key 或图标 URL、是否已配置、是否当前选中、是否可添加 / 编辑 / 移除。
3. 技能列表：技能 ID、名称、显性 / 隐形 / 服务案例分类、玩家是否可见、来源类型、来源文案、锁定时间、手动解锁 / AI 解锁状态、是否可编辑 / 移除。
4. 服务案例：案例 ID、标题、摘要、日期、人数 / 局类型、评分、是否可展示、是否已绑定技能、绑定规则和审核状态；服务案例列表进入详情页时还需要返回或透传 `iconText` / `iconKey`、`tone`、`linkedSkillTitle`，详情页图标应与列表点击项一致，不在详情页固定写死。
5. 解锁建议：潜力技能标题、提示文案、操作按钮文案、解锁接口成功 / 失败返回结构、是否扣减修改次数。
6. 保存规则：保存是整体覆盖还是增量更新，修改次数何时扣减，添加 / 编辑 / 移除 / 解锁是否需要二次确认，失败码和提示文案。
7. 图标资源：后台返回 `iconKey` 由前端映射项目内图标，还是直接返回已入库的图标 URL；若使用外部图标，需先下载到项目内后引用。
8. 隐形技能空态：当前前端已按产品反馈关闭默认隐形技能候选；当 `skillGroups.hidden` 为空时只展示锁定说明，不展示默认卡片或解锁卡。若后台需要开启隐形技能展示，需要返回非空 `skillGroups.hidden` 和对应解锁建议。
9. 第三个技能槽：当前第三槽未解锁时仍显示灰色 `+ / 添加技能`，但需要 `skillSlots[].locked=true` 控制不可添加；点击下方“解锁第三个技能”只把该槽位解锁为 `locked=false`，随后用户再点击槽位进入添加技能面板。正式接口需确认解锁槽位是否扣减修改次数，以及保存 / 单独解锁接口返回结构。
10. 服务案例详情：点击服务案例后当前前端使用 `caseId` 切换本地兜底详情；正式联调建议由 `GET /api/app/profile/system-management/skill-config/cases/{caseId}` 返回详情页所需字段，包括 `title`、`date`、`playersText`、`totalPlayers`、`ratingText`、`score`、`tags[]`、`detailSections[]`、`players[]`、`iconText` / `iconKey`、`tone`。评分星级可由后台直接返回或由前端根据 `score` 派生，但最终评分、标签和玩家列表都应以后台为准。

状态：前端已新增深色静态页、服务案例详情页和对应编译模式，并已接 `GET /api/app/profile/system-management/skill-config` 与 `PUT /api/app/profile/system-management/skill-config` service / mock 链路；添加 / 编辑 / 移除 / 解锁已先补本地响应 UI 和本地状态更新，服务案例详情当前仍为前端兜底数据。正式接口字段、操作权限、修改次数扣减、失败码、案例绑定流程和案例详情接口仍待后端和产品确认。
