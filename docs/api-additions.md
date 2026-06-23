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
## 20. 组局发起邀请页预算上限与分成比例字段
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
## 25. 组局成功行家提示页详情接口

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

## 26. 组局成功领路人提示页后续跟进接口

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
      "serviceOrderId": "SO-20260320-001",
      "fundAmount": 800,
      "expert": { "id": "expert-zhang", "name": "张专家", "avatarText": "ZH" },
      "guide": { "id": "guide-wang", "name": "王引荐" },
      "serviceTitle": "产品架构咨询",
      "startedAt": "2026-03-10T14:00:00+08:00",
      "expectedDeliveryAt": "2026-03-25T14:00:00+08:00",
      "noticeText": "取消需赔付一定比例金额给行家",
      "canContactExpert": true,
      "canCancel": true
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

功能：领路人查看已取消组局结果。取消方可能是玩家，也可能是行家，因此系统设定理由、取消方解释话语、提出取消的角色信息、取消时间和取消历程均需要由后台返回。

候选接口：`GET /api/app/game-invites/guide-cancel-detail`

建议入参：`id`。

建议返回字段：`statusTitle`、`statusDesc`、`canceledBy`、`reason`、`message`、`messageTimeText`、`timeline`。其中 `canceledBy.roleType` 建议使用 `player/expert/guide`，`reason.title/reason.desc` 为系统设定理由，`message` 为取消方解释话语。

待后端 / 产品确认：

1. 接口路径是否使用 `/api/app/game-invites/guide-cancel-detail`，还是复用进度详情接口返回取消态字段。
2. 取消方角色枚举是否固定为 `player/expert/guide`。
3. `reason.title`、`reason.desc`、`message` 是否均由后台返回展示文案。
4. 取消历程是否由后台完整返回，还是前端按取消状态本地生成。
## 32. 局内协作页数据与操作接口

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

## 33. 服务交付确认页操作接口

记录日期：2026-06-23
模块：组局 / 服务交付 - 免费
页面：`pages/game/delivery/index`
功能：服务交付免费页当前为静态 UI。快捷操作中的 `联系玩家`、`联系领路人` 后续需要进入对应人员的消息界面；确认状态里的 `提醒确认` 后续需要给玩家发消息 / 提醒玩家确认服务完成；底部 `确认服务完成` 后续需要提交服务完成确认，并触发玩家确认、状态刷新和免费局归档流程。

建议接口 / 能力：

```text
GET /api/app/services/{serviceId}/delivery-detail
POST /api/app/services/{serviceId}/remind-confirmation
POST /api/app/services/{serviceId}/complete-confirmation
GET /api/app/messages/session?targetUserId={userId}&serviceId={serviceId}
```

待确认：

1. 服务交付确认页是否使用独立详情接口，还是复用业务管理 / 玩家管理的服务订单详情。
2. 点击 `联系玩家`、`联系领路人` 是先进入聊天页由用户手动发送，还是直接发送默认话术。
3. `提醒确认` 是进入玩家聊天页预填话术，还是直接调用提醒接口给玩家发送确认消息。
4. `确认服务完成` 是由行家单方提交后等待玩家确认，还是玩家 / 行家任一方都可确认。
