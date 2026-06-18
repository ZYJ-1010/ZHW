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
