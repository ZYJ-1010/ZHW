# 小程序前端架构分层

本文是工程实现文档，不修改 `00.文档` 下的正式 PRD、功能清单和接口数据库要求。

## 目录职责

```text
pages/        页面层，只做展示、页面状态和事件转发
services/     业务层，处理页面用例、错误转换和本地缓存
api/          接口层，封装请求、mock / prod 切换、接口模块
config/       环境、路由、页面树、常量
utils/        通用工具
docs/         工程侧记录文档
00.文档/      正式需求和接口数据库文档，只读参考
```

## 调用方向

```text
pages -> services -> api/modules -> api/request -> mock 或真实服务器
```

页面不要直接调用 `wx.request`。业务判断优先放在 `services/`，接口路径集中放在 `api/modules/`。

## 当前模块

```text
auth        微信登录、token、本地用户缓存
invite      邀请码校验和邀请关系入口
user        当前用户、资料维护
role        行家 / 领路人申请和状态
game        局列表、局详情、创建、申请
location    定位、手动定位、附近局
im          局内消息和房间
profile     我的、会员、成长聚合
review      可评价局和提交评价
revenue     收益摘要和明细
```

## 接口规则

正式接口以 `00.文档/接口、数据库初稿.md` 第 20 章为准。

统一返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {},
  "requestId": "req_xxx"
}
```

接口请求头默认包含：

```text
X-Client-Type: mp-wechat
X-App-Version: 0.1.0
Authorization: Bearer <token>
```

## 开发约定

1. 新页面先注册到 `app.json`。
2. 新页面路径同步记录到 `config/routes.js`。
3. 墨刀页面树和开发批次同步记录到 `config/page-map.js` 与 `docs/page-map.md`。
4. 正式文档缺失但前端确实需要的接口，记录到 `docs/api-additions.md`，不要直接改 `00.文档`。
5. 一期免费局为主，支付相关页面只做状态和二期预留。
