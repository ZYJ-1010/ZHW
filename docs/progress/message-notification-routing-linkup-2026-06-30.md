# 消息通知前后端路由对接

日期：2026-06-30

## 本轮完成

- 小程序消息中心卡片点击改为调用 `/api/app/notifications/{id}/actions` 的 `detail` 动作，由后端返回真实业务目标页。
- 消息中心动作按钮继续使用后端 `target.route` 跳转，接受/拒绝这类无跳转动作完成后刷新消息列表。
- “查看全部”不再误走通知处理接口，改为消息页内部筛选刷新。
- 消息中心快捷入口补齐：组局加入刷新消息流，成就跳成就页。
- 好友消息页快捷动作补齐：打招呼/发名片复用 IM 文本发送，发定位跳地图页，加好友跳个人资料页。
- 删除图片/文件发送后的不可达占位提示代码。

## 验证

- `node --check miniprogram-client/pages/message/index.js`
- `node --check miniprogram-client/pages/message/my/index.js`
- `go test -count=1 ./...`
- `npm run build`

