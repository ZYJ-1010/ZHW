# 确认与交付页面断点收口

日期：2026-06-30

## 本轮完成

- 再次组局确认页取消动作改为返回上一页或进度页。
- 再次组局提交成功后优先跳组局进度页，`navigateTo` 失败时自动 `redirectTo`，不再停留在“进度页待接入”。
- 交付页时间线未知动作改为当前状态提示。
- 交付页快捷动作未知 fallback 改为暂不可用提示；上传凭证、联系玩家/领路人、申请取消服务仍走已有真实链路。

## 验证

- `node --check miniprogram-client/pages/game/confirm/index.js`
- `node --check miniprogram-client/pages/game/delivery/index.js`
- `go test -count=1 ./internal/appapi -run "Test.*Replay|Test.*Delivery|Test.*ServiceConfirm|Test.*Confirm"`

