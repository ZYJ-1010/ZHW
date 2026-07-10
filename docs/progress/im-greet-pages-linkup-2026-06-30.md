# IM 打招呼页面对接

日期：2026-06-30

## 本轮完成

- `game/greet`、`game/guide-chat` 清理图片/文件发送后的不可达占位提示，继续复用现有 IM 文件消息链路。
- `message/my` 录音停止提示调整为当前支持范围，避免误判为接口未接。
- `game/hall-greet` 支持带 `gameId` 时调用 IM 文本消息接口。
- `game/hall-greet` 快捷动作接入现有能力：
  - 打招呼、发名片：发送文本消息。
  - 发定位：跳地图页并携带 `gameId`。
  - 加好友：跳个人资料页。

## 验证

- `node --check`：`game/greet`、`game/guide-chat`、`game/hall-greet`、`message/index`、`message/my`
- `go test -count=1 ./internal/appapi -run "Test.*IM|Test.*Chat|Test.*Message"`
- `go test -count=1 ./...`
- `npm run build`

