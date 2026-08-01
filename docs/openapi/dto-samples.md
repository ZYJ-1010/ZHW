# DTO 示例

## HealthDTO

```json
{
  "status": "ok",
  "service": "go-api"
}
```

## WechatLoginResponseDTO

```json
{
  "token": "mock-token",
  "expiresAt": "2026-06-14T01:00:00+08:00",
  "needProfile": true,
  "user": {
    "id": 1,
    "openId": "mock_openid_u1",
    "nickname": "",
    "avatarUrl": "",
    "realnameStatus": "pending",
    "status": "active"
  },
  "inviteRelation": {
    "inviteCodeId": 1,
    "inviterUserId": 0,
    "inviteeUserId": 1,
    "bindSource": "invite_code"
  }
}
```

## IdentityStatusDTO

```json
{
  "userId": 1,
  "status": "verified",
  "phoneMasked": "138****8000",
  "smsVerified": true,
  "phoneVerified": true,
  "faceVerified": true,
  "wechatRealnameConsistency": "not_supported",
  "updatedAt": "2026-06-13T01:55:00+08:00"
}
```

## GameDTO

```json
{
  "id": 1,
  "creatorUserId": 1,
  "title": "周末桌游局",
  "gameType": "free",
  "gameSource": "app",
  "status": "pending_audit",
  "minPlayers": 5,
  "maxPlayers": 8,
  "currentPlayers": 1,
  "cityCode": "110100",
  "cityName": "北京",
  "createdAt": "2026-06-13T02:10:00+08:00"
}
```

## PaymentOrderDTO

```json
{
  "id": 1,
  "orderNo": "FREE-1-1",
  "userId": 1,
  "gameId": 1,
  "amountCent": 0,
  "payStatus": "free_no_pay",
  "payChannel": "none",
  "needWechatPay": false,
  "createdAt": "2026-06-13T22:30:00+08:00",
  "updatedAt": "2026-06-13T22:30:00+08:00"
}
```

## PaymentPrecreatePlaceholderDTO

```json
{
  "orderNo": "FREE-1-1",
  "payStatus": "free_no_pay",
  "amountCent": 0,
  "needWechatPay": false,
  "mode": "free_no_pay_placeholder",
  "order": {
    "id": 1,
    "orderNo": "FREE-1-1",
    "userId": 1,
    "gameId": 1,
    "amountCent": 0,
    "payStatus": "free_no_pay",
    "payChannel": "none",
    "needWechatPay": false
  }
}
```

## BehaviorLogDTO

```json
{
  "id": 1,
  "userId": 1,
  "eventType": "search",
  "eventCode": "search",
  "targetType": "game",
  "businessType": "game",
  "targetId": 1,
  "businessId": 1,
  "pagePath": "/pages/home/index",
  "keyword": "桌游",
  "source": "app",
  "device": "ios",
  "ip": "127.0.0.1",
  "extra": {
    "cityCode": "110100"
  },
  "createdAt": "2026-06-13T23:40:00+08:00",
  "occurredAt": "2026-06-13T23:40:00+08:00"
}
```

## FunnelSnapshotDTO

```json
{
  "steps": [
    {
      "eventCode": "login",
      "userCount": 100,
      "conversionRate": 1,
      "dropOffRate": 0
    },
    {
      "eventCode": "browse_games",
      "userCount": 72,
      "conversionRate": 0.72,
      "dropOffRate": 0.28
    }
  ],
  "updatedAt": "2026-06-14T11:58:00+08:00"
}
```

## RetentionSnapshotDTO

```json
{
  "buckets": [
    {
      "cohortDate": "2026-06-14",
      "newUsers": 100,
      "day1Retained": 42,
      "day7Retained": 18,
      "day30Retained": 6,
      "day1Rate": 0.42,
      "day7Rate": 0.18,
      "day30Rate": 0.06
    }
  ],
  "updatedAt": "2026-06-14T11:58:00+08:00"
}
```

## DashboardSnapshotDTO

```json
{
  "funnel": {
    "steps": [
      {
        "eventCode": "login",
        "userCount": 100,
        "conversionRate": 1,
        "dropOffRate": 0
      }
    ],
    "updatedAt": "2026-06-14T11:58:00+08:00"
  },
  "retention": {
    "buckets": [
      {
        "cohortDate": "2026-06-14",
        "newUsers": 100,
        "day1Retained": 42,
        "day7Retained": 18,
        "day30Retained": 6,
        "day1Rate": 0.42,
        "day7Rate": 0.18,
        "day30Rate": 0.06
      }
    ],
    "updatedAt": "2026-06-14T11:58:00+08:00"
  },
  "summary": {
    "registeredUserCount": 100,
    "verifiedUserCount": 80,
    "activeGameCount": 12,
    "imRoomCount": 8
  },
  "distributions": {
    "userStatuses": { "verified": 80, "pending": 5 },
    "gameStatuses": { "pending_audit": 2, "recruiting": 7, "in_progress": 5 },
    "gameTypes": { "free": 14 }
  },
  "updatedAt": "2026-06-14T11:58:00+08:00"
}
```

## OpenIMSessionDTO

```json
{
  "engine": "openim",
  "imUserId": "zhw_user_1",
  "openIMGroupId": "zhw_game_1",
  "openIMToken": "openim-user-token"
}
```

## ChatRoomDTO

```json
{
  "id": 1,
  "gameId": 1,
  "status": "active",
  "memberIds": [1, 2],
  "engine": "openim",
  "openIMGroupId": "zhw_game_1"
}
```

## UploadTokenDTO

```json
{
  "upload": {
    "fileId": 1,
    "uploadUrl": "mock://upload/chat_file/1/1-a.png",
    "storageKey": "chat_file/1/1-a.png",
    "headers": {
      "x-zhw-file-id": "1"
    },
    "expiresAt": "2026-06-13T03:30:00+08:00"
  }
}
```

## AIDataSnapshotDTO

```json
{
  "userCount": 5,
  "gameCount": 3,
  "behaviorLogCount": 20,
  "favoriteCount": 5,
  "reviewCount": 5,
  "footprintCount": 10,
  "connectionCount": 8,
  "expertProfileCount": 2,
  "guideProfileCount": 1,
  "imMessageCount": 12,
  "dataReady": true,
  "acceptanceReady": true,
  "acceptanceChecks": {
    "users": {
      "current": 5,
      "required": 5,
      "ready": true
    },
    "games": {
      "current": 3,
      "required": 3,
      "ready": true
    },
    "behaviorLogs": {
      "current": 20,
      "required": 20,
      "ready": true
    },
    "favorites": {
      "current": 5,
      "required": 5,
      "ready": true
    },
    "reviews": {
      "current": 5,
      "required": 5,
      "ready": true
    }
  },
  "imExportEnabled": false,
  "updatedAt": "2026-06-13T23:30:00+08:00"
}
```

## AIDataAcceptanceFixtureResultDTO

```json
{
  "usersVerified": 5,
  "gamesCreated": 3,
  "behaviorLogsAdded": 20,
  "favoritesAdded": 5,
  "reviewsAdded": 5,
  "imMessagesAdded": 5,
  "acceptanceSnapshot": {
    "userCount": 5,
    "gameCount": 3,
    "behaviorLogCount": 20,
    "favoriteCount": 5,
    "reviewCount": 5,
    "acceptanceReady": true,
    "imExportEnabled": false
  }
}
```

## IMExportItemDTO

```json
{
  "messageId": 1,
  "roomId": 1,
  "gameId": 1,
  "senderUserId": 2,
  "messageType": "text",
  "content": "hello",
  "createdAt": "2026-06-13T23:30:00+08:00"
}
```
