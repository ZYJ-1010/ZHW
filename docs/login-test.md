# 登录测试说明

## 当前模式

当前项目默认使用测试模式，不需要真实服务器也能测试登录流程。mock 返回结构已经按正式接口统一为：

```json
{
  "code": 0,
  "message": "ok",
  "data": {},
  "requestId": "mock_xxx"
}
```

配置位置：

```js
// config/env.js
const currentEnv = ENV.MOCK
```

## 测试数据

有效邀请码：

```text
ENJOY2026
```

无效邀请码：

```text
随便输入其他内容
```

测试用户数据在：

```text
api/mock-data.js
```

登录成功后会写入本地缓存：

```text
enjoy_token
enjoy_user
```

## 怎么跑

1. 打开微信开发者工具。
2. 导入项目目录：`E:\项目\03周总\01enjoy`。
3. 点击编译。
4. 勾选“我已了解并同意使用微信授权登录”。
5. 点击“同意并微信登录”。
6. 看到“授权成功”，随后自动进入 `pages/home/index`，即表示 mock 登录流程通过。

邀请码测试：

1. 输入 `ENJOY2026`。
2. 点击“确认”。
3. 看到“邀请码已确认”即表示 mock 邀请码校验通过。

## 切换真实服务器

修改：

```js
// config/env.js
const currentEnv = ENV.PROD
```

同时把真实接口域名改成你的服务器地址：

```js
const serverMap = {
  [ENV.MOCK]: '',
  [ENV.PROD]: 'https://你的真实接口域名'
}
```

真实服务器至少需要提供文档定稿里的登录接口：

```text
POST /api/app/auth/wechat-login
```

`/api/app/auth/wechat-login` 入参：

```json
{
  "code": "微信 wx.login 返回的 code",
  "inviteCode": "用户输入的邀请码，可为空",
  "encryptedData": "",
  "iv": ""
}
```

成功返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "token": "服务器签发的 token",
    "user": {
      "id": "用户 id",
      "nickname": "微信昵称",
      "avatarUrl": "头像地址",
      "authStatus": "pending",
      "roles": ["player"],
      "needInvite": false,
      "needRealname": true
    },
    "isNewUser": true
  },
  "requestId": "req_xxx"
}
```
