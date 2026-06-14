# 小程序代码导航

本文用于记录“想改某个功能时先去哪找”。后续每新增页面、模块、接口或重要工具，都同步补到这里。

## 总体规则

1. 项目代码只改 `E:\项目\03周总\01enjoy`。
2. 页面优先按 `pages/业务/页面/index.*` 查找。
3. 页面只做展示、状态和事件转发；业务判断优先放 `services/`。
4. 接口路径优先放 `api/modules/` 或 `api/request.js`，页面不要直接调 `wx.request`。
5. 页面跳转优先使用 `config/routes.js`，页面归属优先查 `config/page-map.js` 和 `docs/page-map.md`。

## 常见修改入口

| 想改什么 | 优先位置 |
| --- | --- |
| 页面结构 / 文案 | `pages/**/index.wxml` |
| 页面样式 | `pages/**/index.wxss` |
| 点击事件 / 页面状态 | `pages/**/index.js` |
| 页面配置 / 标题 | `pages/**/index.json` |
| 新增页面注册 | `app.json` |
| 页面跳转路径 | `config/routes.js` |
| 页面与业务模块关系 | `config/page-map.js`、`docs/page-map.md` |
| 业务逻辑 | `services/*.js` |
| 接口调用 | `api/modules/*.js`、`api/request.js` |
| mock 数据 | `api/mock-data.js` |
| mock 接口行为 | `api/mock.js` |
| 环境切换 | `config/env.js` |
| 通用工具 | `utils/*.js` |
| 后续问题 | `docs/follow-ups.md` |
| 待补接口 | `docs/api-additions.md` |

## 当前主要模块

| 模块 | 页面 | 业务层 | 接口层 |
| --- | --- | --- | --- |
| 启动 / 登录前首页 | `pages/entry/index`、`pages/home/guest/index` | - | - |
| 登录 / 邀请注册 / 找回密码 | `pages/login/index`、`pages/login/invite/index`、`pages/login/forgot/index`、`pages/login/realname/index` | `services/auth.js`、`services/invite.js`、`services/user.js`、`services/newbie.js` | `api/request.js`、`api/modules/user.js`、`api/modules/newbie.js` |
| 首页 | `pages/home/index` | `services/home.js` | `api/modules/home.js` |
| 角色申请 | `pages/role/apply/index`、`pages/role/status/index` | `services/role.js` | `api/modules/role.js` |
| 组局 | `pages/game/hall/index`、`pages/game/create/index`、`pages/game/detail/index`、`pages/game/applications/index`、`pages/game/delivery/index`、`pages/game/review/index` | `services/game.js`、`services/review.js` | `api/modules/game.js`、`api/modules/review.js` |
| 地图 | `pages/map/index` | `services/location.js` | `api/modules/location.js` |
| 消息 / IM | `pages/message/index`、`pages/im/room/index` | `services/im.js` | `api/modules/im.js` |
| 我的 / 会员 / 成就 | `pages/profile/index`、`pages/profile/member/index`、`pages/profile/achievements/index` | `services/profile.js`、`services/revenue.js` | `api/modules/profile.js`、`api/modules/revenue.js` |
| 预留能力 | `pages/placeholder/metaverse/index` | - | - |

## 日志和提示

1. 用户可见提示用 `utils/toast.js`。
2. 开发排查日志用 `utils/logger.js`。
3. 请求日志统一在 `api/request.js`。
4. 全局异常在 `app.js`。
5. `wx.login` 这类不走接口层的兜底在对应 `services/`。

## 注释规则

1. 注释只写“为什么这样做”和“这里有什么约定”，不要复述代码。
2. 不在注释里写 token、手机号、密码、验证码、身份证、邀请码、真实接口密钥或内部账号。
3. 当前 `project.config.json` 中 `minified` 为 `false`，不要假设上传生产包时注释一定会被剥离。
4. 如果生产包需要强制去注释，应在正式发布前开启压缩或接入构建流程；在此之前，代码注释按“会进入生产包”来写。
5. 临时兜底用 `TODO(日期/原因)` 标记，并同步到 `docs/follow-ups.md`，不要长期只留在代码注释里。

## 新增内容时同步更新

1. 新增页面：更新 `app.json`、`config/routes.js`、`config/page-map.js`、`docs/page-map.md` 和本文。
2. 新增接口：更新 `api/modules/` 或 `api/request.js`，必要时补 `docs/api-additions.md`。
3. 新增后续待办：更新 `docs/follow-ups.md`。
4. 新增通用工具：放到 `utils/`，并在本文“常见修改入口”或相关模块中补说明。
