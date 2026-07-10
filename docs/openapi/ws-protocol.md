# OpenIM 二次开发 IM 协议

## 选型结论

一期 IM 不再从 0 实现完整 WebSocket 服务，采用 OpenIM 作为开源 IM 底座二次开发。

- 开源项目：`openimsdk/open-im-server`
- 本机源码：`C:\Users\61492\Desktop\codex download\open-im-server`
- 许可证：Apache-2.0
- 采用方式：真好玩 Go API 负责业务鉴权、局成员权限、敏感词策略和业务 DTO；OpenIM 负责 WebSocket 长连接、群组、消息投递、历史消息、文件对象能力。

## 小程序接入口径

小程序不直接创建 OpenIM 群，也不直接决定谁能进群。小程序先调用真好玩业务接口：

- `GET /api/app/games/{gameId}/chat-room`
- `POST /api/app/games/{gameId}/chat/messages`
- `GET /api/app/games/{gameId}/chat/messages`

真好玩后端校验：

- 必须登录。
- 必须是局成员；`chat-session` 握手、`GET /api/app/games/{gameId}/chat/messages` 和 `GET /api/app/chat/rooms/{roomId}/messages` 统一走后端 `AuthorizeGameAccess / AuthorizeRoomAccess` 成员权限判断。
- 局开始后成员锁定，非成员不能进 IM。
- 文本消息先过真好玩敏感词库，命中后返回 `45101`。

校验通过后，后端同步 OpenIM：

- 用户 ID 映射为 `zhw_user_{userId}`。
- 局群 ID 映射为 `zhw_game_{gameId}`。
- 后端通过 OpenIM `/user/user_register` 注册 IM 用户。
- 后端通过 OpenIM `/group/create_group` 创建群。
- 后端通过 OpenIM `/msg/send_msg` 写入群消息。

## OpenIM 配置

```text
OPENIM_API_ADDR=http://127.0.0.1:10002
OPENIM_SECRET=openIM123
OPENIM_ADMIN_USER_ID=imAdmin
```

## D5.5 当前落地映射

- 文本消息：小程序调用 `POST /api/app/games/{gameId}/chat/messages` 或 `POST /api/app/chat/rooms/{roomId}/messages`；后端先做成员权限和敏感词校验，再写入本地消息记录，OpenIM 可用时同步到 OpenIM 群消息。
- 图片/文件消息：小程序先调用 `POST /api/app/files/upload-token` 获取 `fileId`、`uploadUrl`、`storageKey`，上传完成后发送 `messageType=image/file` 并携带 `fileId`；下载时调用 `GET /api/app/files/{fileId}/download-url`，后端按局成员权限放行。
- ACK：`POST /api/app/chat/rooms/{roomId}/messages/{messageId}/ack`，服务端记录 `ackedBy`。
- 已读：`POST /api/app/chat/rooms/{roomId}/messages/{messageId}/read`，服务端记录 `ackedBy` 和 `readBy`。
- 归档：`POST /api/app/chat/rooms/{roomId}/archive`，局结束且无争议时可归档；有争议时后续应由举报申诉链路冻结证据。
- OpenIM Webhook：OpenIM 配置回调地址为 `POST /api/internal/openim/webhooks`；当前后端会记录 `callbackCommand`、`groupID` 原始 payload，并从 `zhw_game_{gameId}` 反解业务局 ID，后续可在该入口补充敏感词、成员权限和证据冻结规则。

未配置 `OPENIM_API_ADDR` 或 `OPENIM_SECRET` 时，Go API 使用本地内存 IM fallback，仅用于单元测试和早期联调。

## 后续二开点

- 在 OpenIM Webhooks 中接入真好玩敏感词和成员权限二次校验。
- 开启 OpenIM 对象存储能力承接图片、文件消息。
- 将后台 IM 管理页接 OpenIM 消息检索与归档能力。
- 生产 Nginx 需要代理 OpenIM WebSocket，并配置小程序合法 socket 域名。
