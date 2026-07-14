insert into user_system_management_configs(user_id, config_key, config_value, updated_at)
values
  (
    1,
    'profile-info',
    '{
      "personalInfo": {
        "name": "测试用户",
        "avatarText": "CE",
        "phoneMasked": "138****0001",
        "contactVisibility": "member",
        "hobby": "城市探索、组局共创"
      },
      "enterpriseInfo": {
        "company": "真好玩测试团队",
        "jobTitle": "产品负责人",
        "businessCountText": "2",
        "resources": "活动空间、行业资源",
        "publicBusinessInfo": true
      },
      "visibilityOptions": [
        { "key": "private", "name": "仅自己可见" },
        { "key": "member", "name": "同局成员可见" },
        { "key": "public", "name": "公开展示" }
      ],
      "certifications": [
        { "key": "personal", "status": "verified", "statusText": "已实名" }
      ]
    }'::jsonb,
    now()
  ),
  (
    1,
    'skill-config',
    '{
      "activeTab": "visible",
      "roleSummary": {
        "roleName": "expert",
        "maxSkillCount": 3,
        "monthlyLimit": 3,
        "usedCount": 1,
        "remainingCount": 2,
        "configuredCount": 1
      },
      "skillSlots": [
        { "id": "product", "title": "产品梳理", "empty": false },
        { "id": "city", "title": "城市探索", "empty": true }
      ],
      "skillGroups": {
        "visible": [
          { "id": "product", "title": "产品梳理", "desc": "B 端产品架构和需求拆解" }
        ],
        "hidden": [],
        "cases": [
          { "id": "case-product", "title": "产品架构共创局", "caseTitle": "真实服务案例", "caseDate": "2026-06-30", "casePlayers": "5人局", "rating": 4.8, "iconText": "P", "tone": "blue", "sourceText": "历史组局服务", "badge": "已绑定", "tags": ["交付稳定"] }
        ]
      },
      "unlockSuggestion": {
        "title": "继续完成服务可解锁更多技能槽"
      }
    }'::jsonb,
    now()
  ),
  (
    1,
    'feedback-home',
    '{
      "activeType": "feature",
      "activeSession": "general",
      "feedbackTypes": [
        { "key": "feature", "label": "功能建议", "iconKey": "pencil" },
        { "key": "problem", "label": "问题反馈", "iconKey": "alert" },
        { "key": "experience", "label": "体验优化", "iconKey": "star" },
        { "key": "game", "label": "组局相关", "iconKey": "problem" },
        { "key": "expert", "label": "行家相关", "iconKey": "problem" },
        { "key": "points", "label": "积分/提现", "iconKey": "notify" },
        { "key": "other", "label": "其他", "iconKey": "alert" }
      ],
      "sessions": [
        { "key": "general", "title": "通用反馈", "meta": "不关联具体组局" },
        { "key": "latest_game", "title": "最近组局", "meta": "可在提交时补充说明" }
      ],
      "quickTypes": [
        { "key": "problem", "label": "遇到问题", "iconKey": "problem", "tone": "red" },
        { "key": "feature", "label": "功能建议", "iconKey": "pencil", "tone": "cyan" },
        { "key": "experience", "label": "体验优化", "iconKey": "star", "tone": "gold" },
        { "key": "other", "label": "其他", "iconKey": "alert", "tone": "gray" }
      ],
      "quickActions": [
        { "key": "feature", "label": "功能建议", "iconKey": "pencil" },
        { "key": "problem", "label": "问题反馈", "iconKey": "alert" },
        { "key": "screenshot", "label": "截图反馈", "iconKey": "image" }
      ],
      "limits": {
        "contentMaxLength": 500,
        "fileMaxCount": 9,
        "uploadNote": "支持 JPG/PNG 图片和语音文件，单个附件不超过 5MB，最多 9 个附件",
        "uploadFullText": "最多上传 9 个附件",
        "uploadSelectedTemplate": "已选择 {selected}/{max} 个附件"
      }
    }'::jsonb,
    now()
  ),
  (
    1,
    'feedback-records',
    '{
      "items": [
        {
          "id": "feature-address-copy",
          "type": "功能建议",
          "typeKey": "feature",
          "typeTone": "blue",
          "status": "处理中",
          "statusClass": "processing",
          "content": "建议在组局详情页增加「一键复制地址」功能，目前每次都要手动输入地址比较麻烦，希望能优化一下。",
          "time": "2026-06-14 18:32",
          "createdAt": "2026-06-14T18:32:00+08:00",
          "replyText": "客服已收到，正在评估",
          "replyClass": "processing",
          "images": [],
          "messages": [
            { "id": "initial", "role": "me", "avatar": "我", "time": "18:32", "content": "建议在组局详情页增加「一键复制地址」功能，目前每次都要手动输入地址比较麻烦，希望能优化一下。" },
            { "id": "service-reply", "role": "service", "avatar": "客", "time": "18:45", "content": "客服已收到，正在评估" }
          ]
        },
        {
          "id": "cashout-busy",
          "type": "问题反馈",
          "typeKey": "problem",
          "typeTone": "red",
          "status": "待处理",
          "statusClass": "pending",
          "content": "积分提现时提示「系统繁忙」，已经持续两天了，请尽快修复。",
          "time": "2026-06-13 09:15",
          "createdAt": "2026-06-13T09:15:00+08:00",
          "images": [],
          "messages": [
            { "id": "initial", "role": "me", "avatar": "我", "time": "09:15", "content": "积分提现时提示「系统繁忙」，已经持续两天了，请尽快修复。" }
          ]
        },
        {
          "id": "expert-home-speed",
          "type": "体验优化",
          "typeKey": "experience",
          "typeTone": "gold",
          "status": "已解决",
          "statusClass": "resolved",
          "content": "行家主页加载速度有点慢，图片可以优化一下压缩策略。",
          "time": "2026-06-10 14:22",
          "createdAt": "2026-06-10T14:22:00+08:00",
          "replyText": "已解决",
          "replyClass": "resolved",
          "images": [],
          "messages": [
            { "id": "initial", "role": "me", "avatar": "我", "time": "14:22", "content": "行家主页加载速度有点慢，图片可以优化一下压缩策略。" },
            { "id": "service-reply", "role": "service", "avatar": "客", "time": "15:10", "content": "已解决" }
          ]
        },
        {
          "id": "game-reminder",
          "type": "组局相关",
          "typeKey": "game",
          "typeTone": "gray",
          "status": "已解决",
          "statusClass": "resolved",
          "content": "组局开始前30分钟没有收到提醒，希望能增加推送通知功能。",
          "time": "2026-06-08 11:05",
          "createdAt": "2026-06-08T11:05:00+08:00",
          "images": [],
          "messages": [
            { "id": "initial", "role": "me", "avatar": "我", "time": "11:05", "content": "组局开始前30分钟没有收到提醒，希望能增加推送通知功能。" }
          ]
        }
      ]
    }'::jsonb,
    now()
  ),
  (
    1,
    'block-settings',
    '{
      "enabled": true,
      "protectionMode": "hard",
      "renewalDays": 30,
      "summary": {
        "protectedUserText": "0位用户",
        "blockedExpertText": "5个关键词",
        "renewalDaysText": "30天"
      },
      "stats": [
        { "value": 0, "label": "保护用户" },
        { "value": 5, "label": "关键词" },
        { "value": "100%", "label": "生效率" }
      ],
      "configRows": [
        { "key": "protection", "title": "保护模式", "desc": "当前：硬保护", "badge": "已开启", "iconText": "盾", "tone": "" },
        { "key": "scene", "title": "分场景配置", "desc": "推荐/附近/列表", "badge": "4开", "iconText": "场", "tone": "green" },
        { "key": "whitelist", "title": "白名单", "desc": "1位不受保护", "badge": "1/20", "iconText": "名", "tone": "gold" }
      ],
      "manageRows": [
        { "key": "users", "title": "用户屏蔽", "desc": "双方互不可见", "badge": "0人", "badgeClass": "red", "iconText": "禁", "tone": "red" },
        { "key": "keywords", "title": "关键词屏蔽", "desc": "自动过滤内容", "badge": "5/20", "iconText": "词", "tone": "" }
      ],
      "rules": [
        { "prefix": "保护期默认", "strong": "30天", "suffix": "" },
        { "prefix": "配置变更", "strong": "5秒内", "suffix": "对新请求生效" }
      ],
      "protectionModes": [
        {
          "key": "hard",
          "title": "硬保护",
          "desc": "完全过滤，用户无感知",
          "rules": [
            { "prefix": "", "strong": "完全过滤", "suffix": "，同类行家内容不展示" },
            { "prefix": "被保护用户", "strong": "零曝光", "suffix": "" }
          ]
        },
        {
          "key": "soft",
          "title": "软保护",
          "desc": "排序降权，推至第5页后",
          "rules": [
            { "prefix": "排序权重", "strong": "降低", "suffix": "，减少曝光" },
            { "prefix": "翻页至局列表第5页后", "strong": "不受保护", "suffix": "" }
          ]
        }
      ],
      "renewalOptions": [
        { "days": 30, "label": "标准周期" },
        { "days": 60, "label": "双倍保护" },
        { "days": 90, "label": "季度保护" }
      ],
      "renewalRules": [
        { "prefix": "保护期最长", "strong": "90天", "suffix": "，到期需重新续期" },
        { "prefix": "到期前", "strong": "3天", "suffix": "将发送提醒通知" }
      ],
      "keywords": ["培训", "课程", "收费教学", "加微信", "私下"],
      "suggestions": ["微商", "直销", "刷单", "贷款", "兼职", "代理", "拉群", "推广"],
      "keywordMaxCount": 20,
      "keywordStats": [
        { "value": 5, "label": "已设置" },
        { "value": 20, "label": "上限", "color": "gold" },
        { "value": "模糊", "label": "匹配模式", "color": "green" }
      ],
      "keywordRules": [
        { "prefix": "支持", "strong": "模糊匹配", "suffix": "，如“培训”匹配“培训机构”“培训课程”" },
        { "prefix": "最多可设", "strong": "20个", "suffix": "关键词" }
      ],
      "whitelist": [
        { "id": "guide-1", "name": "测试用户", "role": "我", "reason": "默认不受屏蔽" }
      ],
      "whitelistCandidates": [],
      "whitelistRules": [
        { "prefix": "白名单行家", "strong": "不受保护", "suffix": "影响" },
        { "prefix": "最多添加", "strong": "20位", "suffix": "" }
      ],
      "blockedUsers": [],
      "blockCandidates": [],
      "blockRules": [
        { "prefix": "屏蔽后双方", "strong": "互不可见", "suffix": "，历史互动记录保留" },
        { "prefix": "屏蔽人数上限", "strong": "100人", "suffix": "" }
      ],
      "scenes": [
        { "key": "recommend", "title": "推荐场景", "desc": "首页/发现页推荐", "mode": "hard", "enabled": true },
        { "key": "nearby", "title": "附近场景", "desc": "LBS地理位置推荐", "mode": "soft", "enabled": true },
        { "key": "message", "title": "私信场景", "desc": "局内与私信内容", "mode": "hard", "enabled": true },
        { "key": "list", "title": "列表浏览", "desc": "局列表/行家列表", "mode": "none", "enabled": true }
      ],
      "sceneRules": [
        { "strong": "硬保护", "suffix": " = 完全过滤" },
        { "strong": "软保护", "suffix": " = 降权至第5页后" },
        { "strong": "不过滤", "suffix": " = 正常展示" }
      ]
    }'::jsonb,
    now()
  )
on conflict (user_id, config_key) do update set
  config_value = excluded.config_value,
  updated_at = now();
