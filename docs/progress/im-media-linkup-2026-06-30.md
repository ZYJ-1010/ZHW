# 2026-06-30 IM Media Linkup

## Done
- `message/my`, `game/guide-chat`, and `game/greet` now send real IM text messages when opened with `gameId`.
- Image/file actions now reuse `chat-input-bar` selected files, upload them as `chat_file`, and send `image` / `file` IM messages.
- `chat_file` MIME guessing now includes `txt` and `zip`, matching backend file policy.
- Mock API now supports file download URLs and `/api/app/chat/rooms/{roomId}/messages` for local linkup.
- Voice remains a visible placeholder because backend IM currently supports only `text`, `image`, and `file`.

## Verification
- `node --check` on chat media service, file service, mock API, chat input component, and three chat pages.
- `go test ./internal/appapi -run 'Test.*IM|TestAdminIM|TestAdminArchive' -count=1`
- `go test ./internal/files ./internal/im -count=1`
- `go test ./...`
- `npm run build` in `admin-web`

## Progress
- Backend + admin main business linkup: about `84%`.
- Full mini program/admin/backend linkup: about `80%`.
- Next best slice: map team action/player detail, then home earth/network real relationship data.
