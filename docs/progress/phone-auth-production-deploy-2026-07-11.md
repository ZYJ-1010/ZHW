# 手机号注册登录生产发布记录 - 2026-07-11

## 发布范围

- 新增公开接口：`POST /api/app/auth/phone-login`。
- 新用户必须携带有效邀请码；已有手机号账号可直接登录原用户。
- 当前临时验证码按产品确认使用 `000000`。
- 新增 `users.mobile_hash` 和唯一部分索引 `uk_users_mobile_hash`。
- 手机号查重使用生产 `JWT_SECRET` 参与计算的 HMAC-SHA256；数据库不保存手机号明文。
- 原微信登录、微信账号绑定和邀请码入口预检继续保留。

## 发布方式

- 只替换手机号认证涉及的 6 个 Go 生产源码文件。
- 只新增迁移 `db/migrations/000036_user_phone_auth.sql`。
- 未覆盖线上已有 `000044_redemption_item_image.sql`。
- 未修改 `deploy/env.prod`、Nginx、Redis、PostgreSQL 容器或后台网页。
- 发布前已建立数据库备份、Go API 源码备份和隔离构建目录。

## 验证结果

- 本地定向 Go 测试通过：`internal/auth`、`internal/users`、`internal/appapi`。
- 服务器隔离 Docker 镜像构建通过。
- 生产迁移执行成功，`mobile_hash` 字段和唯一索引存在。
- `https://api.haowan.net.cn/api/app/health` 返回成功。
- 无效验证码请求返回 HTTP 422 / `42200`。
- 正确临时验证码但缺少邀请码返回 HTTP 403 / `40321`。
- 验证前后生产用户数保持不变，未创建测试账号。
- 从生产库选择未绑定的有效邀请码只执行预校验，返回 `valid=true`、`authPageMode=register`，未绑定、未消耗。
- Go API 容器重建后持续运行，启动日志和上述请求均无异常。

## 后续人工验证

- 在微信开发者工具或真机从真实二维码/链接进入。
- 确认启动 Brand 页预校验后自动进入手机号注册页。
- 输入手机号、`000000` 和自动带入的邀请码完成注册。
- 确认正式 token 写入本地存储并进入后续业务页面。
- 首次完整注册会真实创建用户并绑定邀请码，因此应使用计划用于测试的新邀请码。
