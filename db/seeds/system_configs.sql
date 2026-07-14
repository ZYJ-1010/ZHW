insert into system_configs(config_key, config_value, status, updated_at)
values
  (
    'game.category_config',
    '{
      "primaryCategories": [
        { "key": "task", "name": "任务局", "icon": "category-task", "visible": true, "order": 10, "children": [
          { "key": "project", "name": "做项目", "visible": true, "order": 10, "selectable": true },
          { "key": "brainstorm", "name": "头脑风暴", "visible": true, "order": 20, "selectable": true }
        ] },
        { "key": "explore", "name": "探索局", "icon": "category-income", "visible": true, "order": 20, "children": [
          { "key": "city_explore", "name": "城市探索", "visible": true, "order": 10, "selectable": true },
          { "key": "route_blind_box", "name": "路线盲盒", "visible": true, "order": 20, "selectable": true }
        ] }
      ],
      "typeFilters": [
        { "key": "free", "name": "普通局", "visible": true, "order": 10, "selectable": true },
        { "key": "standard", "name": "标准局", "visible": true, "order": 20, "selectable": true },
        { "key": "public_welfare", "name": "公益局", "visible": true, "order": 30, "selectable": true },
        { "key": "aa", "name": "均摊局", "visible": true, "order": 40, "selectable": true },
        { "key": "crowdfund", "name": "众筹局", "visible": true, "order": 50, "selectable": true },
        { "key": "deposit", "name": "押金局", "visible": true, "order": 60, "selectable": true },
        { "key": "condition", "name": "条件局", "visible": true, "order": 70, "selectable": true }
      ],
      "locationFilters": [
        { "key": "all", "name": "全国", "visible": true, "order": 0, "selectable": true },
        { "key": "nearby", "name": "附近(50km)", "visible": true, "order": 10, "selectable": true }
      ],
      "sortOptions": [
        { "key": "comprehensive", "name": "综合排序", "sortKey": "", "sortOrder": "asc" },
        { "key": "latest", "name": "最新发布", "sortKey": "time", "sortOrder": "desc" },
        { "key": "hot", "name": "热度最高", "sortKey": "hot", "sortOrder": "desc" },
        { "key": "distance", "name": "距离最近", "sortKey": "distance", "sortOrder": "asc" },
        { "key": "credit", "name": "信用优先", "sortKey": "credit", "sortOrder": "desc" }
      ],
      "eventActions": ["分享", "关注", "引荐", "打招呼"],
      "defaultPrimaryCategory": "task",
      "defaultSecondaryCategory": "project",
      "defaultType": "free",
      "createForm": {
        "capacity": { "min": 5, "max": 8 },
        "currentLocationText": "当前位置",
        "participationModes": [
          { "key": "online", "name": "线上" },
          { "key": "offline", "name": "线下" },
          { "key": "hybrid", "name": "混合" }
        ],
        "tags": [
          { "key": "product", "name": "产品研发" },
          { "key": "startup", "name": "创业" },
          { "key": "city_explore", "name": "城市探索" },
          { "key": "cocreation", "name": "共创" }
        ],
        "completionRules": [
          { "key": "goal", "name": "目标达成", "active": true },
          { "key": "manual", "name": "手动结束", "active": true }
        ],
        "feeTypes": [
          { "key": "free", "name": "免费局" },
          { "key": "paid", "name": "收费局" }
        ]
      },
      "version": "2026-07-01"
    }'::jsonb,
    'active',
    now()
  ),
  (
    'game.profit_template_config',
    '{
      "depositRuleText": "连续打卡 7 天即完成。完成者拿回押金池金额，未完成者押金由完成者平分。",
      "depositNoticeText": "支付金额：100元 = 服务费10元 + 押金池90元。服务费不退，押金池按完成情况结算。",
      "version": "2026-07-01"
    }'::jsonb,
    'active',
    now()
  ),
  (
    'game.invite_config',
    '{
      "minPlayerCount": 1,
      "maxPlayerCount": 1,
      "budgetMaxAmount": 99999999,
      "defaultBudget": "800",
      "defaultTitle": "产品架构梳理咨询",
      "defaultDetail": "需要资深产品经理帮忙梳理B端产品架构，预计咨询时长2小时，涉及模块划分和数据流转设计。",
      "playerIntroTemplate": "我帮你邀请了{expertName}，可以一起确认需求、预算和服务节奏。",
      "activityTypes": [
        { "key": "product", "name": "产品咨询" },
        { "key": "design", "name": "设计服务" },
        { "key": "tech", "name": "技术开发" }
      ],
      "rewardRateConfig": {
        "platformServiceRate": 10,
        "systemGuideRewardRate": 10,
        "inviteRewardRate": 40
      }
    }'::jsonb,
    'active',
    now()
  ),
  (
    'game.cancel_config',
    '{
      "player": {
        "reasonOptions": [
          { "key": "need_changed", "text": "需求变更，不再需要服务" },
          { "key": "other_solution", "text": "找到其他解决方案" },
          { "key": "service_unexpected", "text": "业务主服务不符合预期" },
          { "key": "budget", "text": "预算问题/资金紧张" }
        ],
        "defaultReason": "other_solution",
        "agreementText": "我已阅读并同意上述赔付协议，理解主动取消需承担行家的时间成本损失，并同意按设置比例从托管资金中赔付行家。",
        "agreementItems": [
          "我理解主动取消需承担行家的时间成本损失",
          "我同意按设置比例赔付行家，金额从托管资金扣除",
          "剩余金额将在3个工作日内原路退回",
          "此取消记录将影响信用分（-3分）"
        ]
      },
      "expert": {
        "reasonOptions": [
          { "key": "schedule_conflict", "text": "个人时间冲突，无法交付" },
          { "key": "requirement_mismatch", "text": "需求与描述不符，无法完成" },
          { "key": "emergency", "text": "身体原因/突发状况" },
          { "key": "other", "text": "其他原因" }
        ],
        "defaultReason": "schedule_conflict",
        "agreementText": "我已阅读并同意《服务取消协议》，理解主动取消将对我的信用分产生影响（-5分），并同意按设置比例赔付玩家损失。"
      },
      "version": "2026-07-01"
    }'::jsonb,
    'active',
    now()
  ),
  (
    'points.page_config',
    '{
      "stats": [
        { "key": "total", "label": "累计积分" },
        { "key": "redeemed", "label": "已兑换" },
        { "key": "expired", "label": "过期积分" }
      ],
      "rules": [
        {
          "text": "服务完成、举报核实、平台活动等行为可产生积分，具体比例以后台配置为准",
          "strong": "后台规则",
          "suffix": ""
        },
        {
          "text": "积分有效期按平台规则执行，到期后由后台任务处理",
          "strong": "有效期规则",
          "suffix": ""
        },
        {
          "text": "积分仅可兑换",
          "strong": "平台限定商品",
          "suffix": "，不可提现或抵扣付费局"
        }
      ],
      "earnExample": {
        "title": "可获得积分的行为",
        "subtitle": "服务分润、举报核实、活动奖励",
        "points": "+20",
        "rows": [
          { "label": "举报核实奖励", "value": "后台确认后发放" },
          { "label": "服务分润积分", "value": "按后台比例生成" }
        ],
        "result": "积分以后台流水为准"
      },
      "roleExamples": [
        { "key": "expert", "role": "行家服务完成", "amount": "按分润金额", "points": "+积分", "iconText": "行" },
        { "key": "guide", "role": "领路人引荐成功", "amount": "按引荐收益", "points": "+积分", "iconText": "领" },
        { "key": "platform", "role": "平台核实奖励", "amount": "后台配置", "points": "+积分", "iconText": "奖" }
      ],
      "filters": [
        { "key": "all", "label": "全部", "tone": "all" },
        { "key": "income", "label": "收入", "tone": "income" },
        { "key": "expense", "label": "支出", "tone": "expense" }
      ],
      "noteText": "积分规则、比例、有效期和兑换限制均以后端后台配置为准。积分不可提现，不可支付付费局，仅可兑换平台限定商品。",
      "version": "2026-07-01"
    }'::jsonb,
    'active',
    now()
  ),
  (
    'redemption.order_page_config',
    '{
      "tabs": [
        { "key": "all", "label": "\u5168\u90e8" },
        { "key": "pending", "label": "\u5f85\u5ba1\u6838" },
        { "key": "approved", "label": "\u5f85\u53d1\u653e" },
        { "key": "fulfilled", "label": "\u5df2\u5b8c\u6210" },
        { "key": "rejected", "label": "\u5df2\u9a73\u56de" },
        { "key": "canceled", "label": "\u5df2\u53d6\u6d88" }
      ],
      "emptyText": "\u6682\u65e0\u5151\u6362\u8ba2\u5355",
      "logisticsEmptyText": "\u6682\u65e0\u7269\u6d41\u4fe1\u606f",
      "detailEmptyText": "\u6682\u65e0\u8ba2\u5355\u8be6\u60c5",
      "cancelConfirm": {
        "title": "\u53d6\u6d88\u8ba2\u5355",
        "content": "\u53d6\u6d88\u540e\u79ef\u5206\u5c06\u9000\u56de\u5230\u8d26\u6237\uff0c\u786e\u8ba4\u53d6\u6d88\u8fd9\u4e2a\u5151\u6362\u8ba2\u5355\u5417\uff1f",
        "confirmText": "\u786e\u8ba4\u53d6\u6d88",
        "cancelText": "\u518d\u60f3\u60f3",
        "reason": "\u7528\u6237\u4e3b\u52a8\u53d6\u6d88"
      },
      "actions": {
        "detail": "\u67e5\u770b\u8be6\u60c5",
        "cancel": "\u53d6\u6d88\u8ba2\u5355",
        "logistics": "\u67e5\u770b\u7269\u6d41",
        "again": "\u518d\u6b21\u5151\u6362"
      },
      "version": "2026-07-01"
    }'::jsonb,
    'active',
    now()
  ),
  (
    'profile.asset_manage_config',
    '{
      "overviewLabel": "总资产（元）",
      "overviewUpdatedText": "实时同步分润与积分账户",
      "assetStats": [
        { "key": "totalDealAmount", "label": "总成交额" },
        { "key": "withdrawable", "label": "可提现", "tone": "green" },
        { "key": "pendingSettlement", "label": "待结算", "tone": "yellow" }
      ],
      "quickActions": [
        { "key": "withdraw", "label": "提现", "tone": "green", "iconSrc": "/pages/profile/asset-center/manage/assets/fa/download.svg" },
        { "key": "recharge", "label": "充值", "tone": "blue", "iconSrc": "/pages/profile/asset-center/manage/assets/fa/plus.svg", "enabled": false, "disabledReason": "一期未接真实支付充值" }
      ],
      "menuItems": [
        { "key": "balance", "title": "余额明细", "desc": "收入支出记录", "iconSrc": "/pages/profile/asset-center/manage/assets/fa/list-ul.svg", "tone": "blue" },
        { "key": "bankCards", "title": "银行卡", "desc": "管理收款账户", "iconSrc": "/pages/profile/asset-center/manage/assets/fa/credit-card.svg", "tone": "green" },
        { "key": "orders", "title": "我的订单", "desc": "查看全部订单", "iconSrc": "/pages/profile/asset-center/manage/assets/fa/bag-shopping.svg", "tone": "purple", "route": "/pages/profile/asset-center/orders/index" }
      ],
      "orderStatuses": [
        { "key": "pendingPay", "label": "待付款", "iconSrc": "/pages/profile/asset-center/manage/assets/fa/hourglass-half.svg", "tone": "blue", "route": "/pages/profile/asset-center/orders/index?status=pending_pay" },
        { "key": "processing", "label": "进行中", "iconSrc": "/pages/profile/asset-center/manage/assets/fa/spinner.svg", "tone": "orange", "route": "/pages/profile/asset-center/orders/index?status=pending" },
        { "key": "completed", "label": "已完成", "iconSrc": "/pages/profile/asset-center/manage/assets/fa/check.svg", "tone": "green", "route": "/pages/profile/asset-center/orders/index?status=fulfilled" },
        { "key": "refund", "label": "退款/售后", "iconSrc": "/pages/profile/asset-center/manage/assets/fa/rotate-left.svg", "tone": "red", "route": "/pages/profile/asset-center/orders/index?status=canceled" },
        { "key": "review", "label": "待评价", "iconSrc": "/pages/profile/asset-center/manage/assets/fa/star.svg", "tone": "gray", "route": "/pages/profile/service-center/manage/review-manage/index" }
      ],
      "bankCards": { "unboundText": "未绑定", "boundSuffix": "张", "canBind": true },
      "faqLinks": [
        { "key": "withdrawArrival", "label": "提现多久到账？", "answer": "提现需在后台财务审核后处理，具体到账时间以后续支付通道规则为准。" },
        { "key": "bindBankCard", "label": "如何绑定银行卡？", "answer": "银行卡绑定入口已预留，正式资金通道接入后开放。" }
      ],
      "version": "2026-07-01"
    }'::jsonb,
    'active',
    now()
  ),
  (
    'role.expert_apply_config',
    '{
      "skillOptions": [
        { "name": "摄影", "active": true },
        { "name": "户外", "active": false },
        { "name": "美食", "active": false },
        { "name": "文化", "active": false },
        { "name": "手工", "active": false },
        { "name": "运动", "active": false },
        { "name": "音乐", "active": false },
        { "name": "+自定义", "custom": true }
      ],
      "fields": [
        { "type": "chips", "key": "skillDomain", "label": "选择技能领域", "required": true },
        {
          "type": "input",
          "key": "skillTags",
          "label": "技能标签",
          "required": true,
          "placeholder": "如：人像摄影、风光摄影、夜景拍摄",
          "helper": "添加具体标签，让用户更容易找到你",
          "maxlength": 30
        },
        { "type": "select", "key": "experienceYears", "label": "从业年限", "required": true, "placeholder": "请选择从业年限" },
        {
          "type": "textarea",
          "key": "intro",
          "label": "个人简介",
          "required": true,
          "placeholder": "介绍你的专业背景、服务风格、擅长领域...",
          "helper": "不少于 50 字，突出你的专业优势",
          "maxlength": 300
        }
      ],
      "uploadField": {
        "label": "资质证明",
        "required": true,
        "icon": "📎",
        "title": "点击上传作品集及凭证",
        "acceptTypes": ["JPG", "PNG", "PDF"],
        "maxCount": 5
      },
      "validationRules": {
        "skillTags": { "minLength": 2, "maxLength": 30 },
        "intro": { "minLength": 50, "maxLength": 300 },
        "serviceName": { "minLength": 2, "maxLength": 20 },
        "customSkill": { "minLength": 2, "maxLength": 8 },
        "money": { "integerMaxLength": 8, "decimalMaxLength": 2 }
      },
      "yearOptions": [
        "1年", "2年", "3年", "4年", "5年", "6年", "7年", "8年", "9年", "10年",
        "11年", "12年", "13年", "14年", "15年", "16年", "17年", "18年", "19年", "20年",
        "21年", "22年", "23年", "24年", "25年", "26年", "27年", "28年", "29年", "30年",
        "30年以上"
      ],
      "serviceCount": 3,
      "priceHint": "平台将收取 10% 服务费"
    }'::jsonb,
    'active',
    now()
  ),
  (
    'role.application_page_config',
    '{
      "pageTitle": "选择你的身份",
      "pageDesc": "玩家为默认身份。行家和领路人需提交申请，审核通过后开放对应能力。",
      "texts": {
        "loadingText": "加载中...",
        "loadFailedText": "角色申请加载失败",
        "conditionLabel": "条件达成",
        "paymentLabel": "付费状态",
        "pendingButtonText": "已进入审核",
        "submittingText": "提交中...",
        "submitButtonText": "提交申请",
        "submitFailedText": "提交失败",
        "requirementPrefix": "• ",
        "defaultSuccessText": "申请已提交"
      }
    }'::jsonb,
    'active',
    now()
  ),
  (
    'role.guide_apply_config',
    '{
      "applyRoleType": "guide",
      "applyRoleName": "领路人",
      "requirements": [
        { "title": "玩家等级达到 Lv.5", "text": "以后台资格规则为准", "done": false },
        { "title": "完成实名认证", "text": "领路人必须实名", "done": false },
        { "title": "信用分 ≥ 80 分", "text": "以信用记录为准", "done": false }
      ],
      "planTask": { "title": "提交领路计划书", "text": "描述你的带队风格、战绩、资源和规划", "done": false, "action": "去填写 ›" },
      "perks": [
        { "icon": "¥", "text": "有权益的领路人引荐玩家组局可获得相应收入" },
        { "icon": "★", "text": "专属领路人标识与优先推荐位" },
        { "icon": "D", "text": "数据看板：查看邀约数据与关系网络" }
      ],
      "fields": [
        { "key": "city", "label": "所在城市", "type": "input", "required": true, "placeholder": "请输入常驻城市", "maxlength": 20, "helper": "用于匹配同城玩家与组局推荐" },
        {
          "key": "audience",
          "label": "可推荐人群",
          "type": "chips",
          "required": true,
          "options": [
            { "name": "朋友", "active": true },
            { "name": "同事", "active": true },
            { "name": "同城玩家", "active": true },
            { "name": "社群成员", "active": false }
          ],
          "helper": "可多选，后续将用于关系网推荐"
        },
        { "key": "contact", "label": "常用联系方式", "type": "input", "required": true, "placeholder": "请输入微信号或手机号", "maxlength": 30 },
        { "key": "guidePlan", "label": "领路计划书", "type": "textarea", "required": true, "placeholder": "请描述你的带队风格、战绩、资源和规划", "maxlength": 300, "helper": "不少于 50 字，说明你能帮助玩家完成组局的方式" }
      ],
      "uploadField": {
        "label": "资质证明",
        "required": true,
        "icon": "📎",
        "title": "点击上传作品集及凭证",
        "helper": "支持 JPG、PNG、PDF，最多 5 张",
        "acceptTypes": ["JPG", "PNG", "PDF"],
        "maxCount": 5
      },
      "serviceCount": 3,
      "serviceBlocks": [
        { "id": "guide-service-1", "title": "业务" },
        { "id": "guide-service-2", "title": "业务" },
        { "id": "guide-service-3", "title": "业务" }
      ],
      "validationRules": {
        "guidePlan": { "minLength": 50, "maxLength": 300 },
        "serviceName": { "minLength": 2, "maxLength": 20 },
        "money": { "integerMaxLength": 8, "decimalMaxLength": 2 }
      },
      "priceHint": "平台将收取 10% 服务费",
      "primaryText": "提交领路人申请",
      "helperText": "审核预计 1-3 个工作日"
    }'::jsonb,
    'active',
    now()
  ),
  (
    'role.status_page_config',
    '{
      "roleAliases": { "player": "player", "expert": "expert", "master": "expert", "guide": "guide", "leader": "guide", "玩家": "player", "行家": "expert", "领路人": "guide" },
      "statusMap": { "active": "approved", "enabled": "approved", "passed": "approved", "success": "approved", "waiting": "pending", "reviewing": "pending", "auditing": "pending", "pending_audit": "pending", "rejected_audit": "rejected", "reject": "rejected", "disabled": "disabled", "available": "none", "locked": "none", "unavailable": "none" },
      "roleMeta": {
        "expert": { "roleName": "行家", "applyTitle": "行家申请", "successAccent": "cyan", "approvedCopy": "你已获得行家身份，可在平台内使用对应能力", "primaryText": "开启行家之旅" },
        "guide": { "roleName": "领路人", "applyTitle": "领路人申请", "successAccent": "orange", "approvedCopy": "你已获得领路人身份，可在平台内使用对应能力", "primaryText": "开启领路人之旅" }
      },
      "pendingTimeline": [
        { "title": "提交申请", "descTemplate": "已成功提交{roleName}申请资料", "timeField": "submittedAt", "fallbackTime": "已提交", "state": "done" },
        { "title": "资料初审", "desc": "平台审核团队已接收并开始初审", "timeWhenSubmitted": "已接收", "fallbackTime": "待系统同步", "state": "done" },
        { "title": "深度审核", "descByRole": { "guide": "正在评估你的组局记录、信用分及领路计划书", "expert": "正在评估你的专业能力、资质材料及服务说明" }, "time": "进行中...", "state": "active" },
        { "title": "结果通知", "desc": "审核结果将通过消息推送通知你", "time": "待完成", "state": "pending" }
      ],
      "approvedActions": [
        { "iconKey": "network", "text": "关系网开启", "routeKey": "relationNetwork" },
        { "iconKey": "invite", "text": "邀请玩家", "routeKey": "gameInvite" },
        { "iconKey": "profile", "text": "完善资料", "routeKey": "profileSystemProfileInfo" }
      ],
      "texts": {
        "loadingText": "加载中...", "errorTitle": "审核状态加载失败", "backHomeText": "返回首页", "retryText": "重试", "pendingPageTitle": "审核进度", "resultPageTitle": "审核结果", "pendingTitle": "审核中", "approvedTitle": "恭喜审核通过！", "rejectedTitle": "审核未通过", "pendingSubtitleTemplate": "{roleName}申请正在审核", "approvedSubtitleTemplate": "你已成为「{roleName}」", "rejectedSubtitle": "查看原因并完善后可再次申请", "pendingDesc": "平台正在评估你的申请资料，请耐心等待", "rejectedDesc": "感谢你的申请，但本次审核未通过", "approvedAuditDesc": "你的申请已通过平台审核", "expectedLabel": "预计完成时间", "expectedTemplate": "预计 {expectedReviewAt} 前完成审核，届时将通过站内消息通知你审核结果。", "expectedFallback": "审核预计 1-3 个工作日，结果将通过站内消息通知你。", "detailTitle": "申请详情", "pendingHelper": "审核期间你可以继续使用玩家身份", "certNoLabel": "认证编号", "certTimePrefix": "认证时间: ", "giftTitle": "新手礼包", "reasonTitle": "驳回原因", "suggestionTitle": "改进建议", "reapplyTitle": "重新申请", "reapplyDesc": "完善资料后可再次提交申请。建议根据驳回原因逐项改进，提高通过率。", "recordTitle": "申请记录", "pendingFooterHomeText": "返回玩家首页", "pendingFooterBenefitsText": "查看权益对比", "rejectedHelpText": "查看帮助", "rejectedImproveText": "完善资料", "routeMissingText": "请选择可用入口", "fieldRoleLabel": "申请角色", "fieldApplyTimeLabel": "申请时间", "fieldApplicationNoLabel": "申请编号", "fieldCurrentStatusLabel": "当前状态", "fieldExpectedLabel": "预计完成", "fieldRejectTimeLabel": "驳回时间", "fieldReapplyLabel": "可重新申请", "submittedFallback": "已提交", "backendRecordFallback": "以后台记录为准", "applicationNoFallback": "审核中生成", "pendingStatusText": "深度审核中", "statusFallback": "待确认", "expectedDoneFallback": "预计 1-3 个工作日", "reapplySuffix": " 后", "reapplyNotifyFallback": "请关注后台通知", "loadFailedText": "审核状态加载失败"
      },
      "defaultRejectReasons": ["申请资料暂未达到当前角色审核要求", "部分证明材料或计划说明仍需补充完善"],
      "suggestionTemplates": ["多参与平台组局活动，积累带队经验", "完善个人资料，提升信用评分", "{improvePlanText}，详细描述你的服务优势", "获得同伴推荐背书可提升审核通过率"],
      "improvePlanTextByRole": { "guide": "重新撰写领路计划书", "expert": "补充服务说明" },
      "reapplyDays": 7
    }'::jsonb,
    'active',
    now()
  ),
  (
    'home.display_config',
    '{
      "onlineBaseCount": 0,
      "onlineSuffix": "人在线"
    }'::jsonb,
    'active',
    now()
  ),
  (
    'game.application_config',
    '{
      "agreementTitle": "入局申请须知",
      "agreementText": "申请入局前请确认本人已完成实名，了解局的主题、地点、时间和成员规则。申请通过后请按约参与，临时退出可能影响信用分。",
      "requireRealname": true,
      "requireIntro": true,
      "requireAgreement": true,
      "allowDuplicateApply": false,
      "uploadRequired": false,
      "maxUploadCount": 3,
      "allowedUploadTypes": ["jpg", "png", "pdf"],
      "minIntroLength": 5,
      "maxIntroLength": 200,
      "maxMessageLength": 120,
      "searchEnabled": false,
      "recommendationHint": "一期优先展示审核通过且人数未满的局。",
      "texts": {
        "subtitle": "你的信息将展示给发起人",
        "wechatTitle": "微信信息",
        "nicknameLabel": "昵称",
        "introLabel": "自我介绍",
        "introPlaceholder": "介绍你的背景、能力和参与动机",
        "portfolioLabel": "相关经历/作品",
        "messageLabel": "申请留言",
        "messagePlaceholder": "给发起人留一句话",
        "agreementPrefix": "我已阅读并同意",
        "cancelText": "取消",
        "submitText": "提交申请",
        "profileSyncedText": "资料已同步",
        "profilePendingText": "资料待同步",
        "profileNameFallback": "待同步",
        "loadFailedText": "入局申请配置加载失败",
        "mediaUnsupportedText": "当前微信版本不支持选择图片",
        "fileUnsupportedText": "当前微信版本不支持选择文件",
        "imageTypeErrorText": "仅支持 JPG、PNG、GIF、WEBP 图片",
        "fileTypeErrorText": "仅支持 PDF 文件",
        "imageSelectedText": "图片已选择",
        "fileSelectedText": "文件已选择",
        "chooseFailedText": "选择失败，请重试",
        "introRequiredText": "请先填写自我介绍",
        "introMinTemplate": "自我介绍不少于{min}字",
        "agreementRequiredText": "请先勾选平台协议",
        "uploadRequiredText": "请先上传相关经历/作品",
        "gameMissingText": "缺少局信息",
        "submittingText": "提交中",
        "submitSuccessText": "申请已提交",
        "submitFailedText": "提交失败，请重试",
        "maxUploadTemplate": "最多上传{max}个文件",
        "navUnavailableText": "当前页暂无左右切换"
      },
      "auditPage": {
        "pageTitle": "审核申请列表",
        "filters": [
          { "key": "all", "name": "全部" },
          { "key": "pending", "name": "待审核" },
          { "key": "approved", "name": "已通过" },
          { "key": "rejected", "name": "已拒绝" }
        ],
        "statusTexts": {
          "pending": "待审核",
          "approved": "已通过",
          "rejected": "已拒绝"
        },
        "roleNames": {
          "expert": "行家",
          "guide": "领路人",
          "main_guide": "主行家",
          "player": "玩家",
          "member": "玩家"
        },
        "texts": {
          "userFallbackTemplate": "用户{userId}",
          "avatarFallback": "玩",
          "applyTimeLabel": "申请时间",
          "approveText": "通过申请",
          "rejectText": "拒绝",
          "reviewedText": "已完成审核",
          "detailText": "详情",
          "selectAllText": "全选",
          "batchRejectText": "批量拒绝",
          "batchApproveText": "批量通过",
          "loadFailedText": "申请列表加载失败",
          "approvingText": "通过中",
          "rejectingText": "拒绝中",
          "approveSuccessText": "已通过申请",
          "rejectSuccessText": "已拒绝申请",
          "reviewFailedText": "审核失败",
          "emptyPendingText": "暂无待审核申请",
          "batchApprovingText": "批量通过中",
          "batchRejectingText": "批量拒绝中",
          "batchApproveSuccess": "已批量通过",
          "batchRejectSuccess": "已批量拒绝",
          "batchReviewFailedText": "批量审核失败"
        },
        "detail": {
          "pageTitle": "审核组局",
          "referralText": "已撮合双方意向",
          "statusTitles": { "pending": "等待你审核", "approved": "已通过申请", "rejected": "已拒绝申请" },
          "countdownTexts": { "pending": "待处理", "approved": "已处理", "rejected": "已处理" },
          "playerStatusTexts": {
            "pendingRequirement": "待确认需求",
            "reviewedRequirement": "已完成审核",
            "pending": "等待审核",
            "approved": "申请已通过",
            "rejected": "申请已拒绝"
          },
          "texts": {
            "playerTitle": "玩家信息",
            "portfolioTitle": "相关经历/作品",
            "confirmTitle": "局信息确认",
            "optionTitle": "可选操作",
            "noticeTitle": "确认须知",
            "relationTitle": "组局关系图",
            "expertName": "我",
            "expertRoleText": "审核方",
            "expertAvatarText": "我",
            "guideAvatarFallback": "领",
            "playerAvatarFallback": "玩",
            "needPrefix": "申请说明：",
            "remarkPrefix": "申请时间：",
            "detailMissingText": "申请详情不存在",
            "loadFailedText": "申请详情加载失败",
            "emptyTitle": "申请详情未加载",
            "emptyText": "请确认审核入口携带的申请 ID 是否有效，或返回申请列表重新打开。",
            "mediaUnsupportedText": "当前微信版本不支持选择图片",
            "fileUnsupportedText": "当前微信版本不支持选择文件",
            "imageTypeErrorText": "仅支持 JPG、PNG、GIF、WEBP 图片",
            "fileTypeErrorText": "仅支持 PDF 文件",
            "filePathInvalidText": "文件路径无效",
            "uploadingText": "上传中",
            "imageUploadedText": "图片已上传",
            "fileUploadedText": "文件已上传",
            "uploadFailedText": "上传失败，请重试",
            "chooseFailedText": "选择失败，请重试",
            "detailRequiredActionText": "申请详情加载后才可以沟通",
            "detailRequiredReviewText": "申请详情加载后才可以审核",
            "chatPrefill": "你好，我想进一步确认本次组局申请。",
            "timePrefill": "我建议进一步确认本次组局的具体时间，请看是否方便。",
            "unavailableActionText": "请选择可用操作",
            "approvingText": "通过中",
            "rejectingText": "拒绝中",
            "approveSuccessText": "已确认通过",
            "rejectSuccessText": "已拒绝申请",
            "reviewFailedText": "审核失败",
            "actionTip": "确认后将建立三方连接群并冻结资金",
            "actionLoadingText": "处理中...",
            "confirmText": "确认通过"
          },
          "sessionItems": [
            { "key": "topic", "label": "组局主题", "iconText": "H", "iconClass": "topic" },
            { "key": "time", "label": "时间", "iconSrc": "/pages/game/detail/assets/icon-clock.png", "iconClass": "time" },
            { "key": "location", "label": "地点", "actionText": "地图位置", "iconSrc": "/pages/game/detail/assets/icon-location.png", "iconClass": "place" }
          ],
          "confirmRows": [
            { "key": "activityType", "label": "活动类型" },
            { "key": "serviceDuration", "label": "服务时长" },
            { "key": "clientBudget", "label": "客户预算" },
            { "key": "platformFee", "label": "平台" },
            { "key": "guideReward", "label": "领路人" },
            { "key": "partnerReward", "label": "生态合伙人" },
            { "key": "expertIncome", "label": "你的收益" }
          ],
          "optionalActions": [
            { "key": "time", "name": "提议具体时间", "iconSrc": "/pages/game/audit-detail/assets/option-time.png" },
            { "key": "chat", "name": "与玩家沟通", "iconSrc": "/pages/game/audit-detail/assets/option-chat.png" }
          ],
          "noticeBullets": [
            "确认后请准时参加，如需取消请提前通知",
            "双方确认后组局正式生效，领路人将获得积分奖励",
            "请保持专业态度，维护平台信誉"
          ]
        }
      },
      "version": "2026-06-30"
    }'::jsonb,
    'active',
    now()
  ),
  (
    'game.audit_config',
    '{
      "autoApproveFreeGames": false,
      "requireManualAuditTypes": ["free", "standard", "public_welfare", "aa", "crowdfund", "deposit", "condition"],
      "requiredRejectReason": true,
      "allowUserResubmitAfterReject": true,
      "batchAuditMaxCount": 50,
      "applicationAuditMode": "creator_or_main_guide",
      "reviewerRoles": ["super_admin", "audit_admin"],
      "version": "2026-06-30"
    }'::jsonb,
    'active',
    now()
  ),
  (
    'game.condition_rule_config',
    '{
      "enabled": true,
      "visibleInMiniProgram": true,
      "adminOnlyCreate": true,
      "ruleItems": [
        { "key": "realname_verified", "name": "完成实名认证", "description": "玩家必须完成实名后才能申请条件局", "required": true, "order": 10 },
        { "key": "credit_min_80", "name": "信用分不低于 80", "description": "用于测试条件局的信用门槛", "required": true, "order": 20 },
        { "key": "profile_complete", "name": "资料完整", "description": "昵称、头像、城市等资料达到基础完整度", "required": false, "order": 30 }
      ],
      "defaultVisibility": "approved_users",
      "reviewRequired": true,
      "paymentRequired": false,
      "version": "2026-06-30"
    }'::jsonb,
    'active',
    now()
  ),
  (
    'game.replay_quick_actions',
    '[
      {
        "id": "same-friends",
        "theme": "green",
        "iconText": "👫",
        "title": "同局好友再玩一局",
        "desc": "立即邀请上一局成员",
        "route": "confirm",
        "order": 10,
        "visible": true
      },
      {
        "id": "smart-match",
        "theme": "blue",
        "iconText": "🤖",
        "title": "系统推荐适配组局",
        "desc": "基于资料和关系数据返回适配候选",
        "route": "system_recommend",
        "order": 20,
        "visible": true
      },
      {
        "id": "create-new",
        "theme": "pink",
        "iconType": "plus",
        "title": "玩家创建新局",
        "desc": "自定义需求，开启全新组局",
        "route": "create",
        "order": 30,
        "visible": true
      }
    ]'::jsonb,
    'active',
    now()
  ),
  (
    'game.replay_quick_messages',
    '[
      { "text": "再来一局？", "order": 10, "visible": true },
      { "text": "上次合作很愉快，继续！", "order": 20, "visible": true },
      { "text": "有个新需求想聊聊", "order": 30, "visible": true },
      { "text": "有空再约一局", "order": 40, "visible": true }
    ]'::jsonb,
    'active',
    now()
  ),
  (
    'review.page_config',
    '{
      "navTitle": "服务评价",
      "skipText": "跳过",
      "statusTitle": "服务已完成！",
      "statusDesc": "请对本次服务进行评价",
      "satisfactionQuestion": "这一局好玩吗？",
      "satisfactionOptions": [
        { "id": "great", "emoji": "😀", "title": "很好玩", "desc": "五星体验" },
        { "id": "ok", "emoji": "🙂", "title": "还行", "desc": "基本合格" },
        { "id": "bad", "emoji": "😕", "title": "不好玩", "desc": "有待改进" }
      ],
      "storyTitle": "发生了什么有趣的事？",
      "aiTip": "AI小助手提示：可以从收获、惊喜、合作感受等方面描述哦",
      "storyPlaceholder": "我们碰撞出了新的思路，对方的经验帮了大忙！",
      "storyMaxLength": 100,
      "aiSummaryText": "AI帮我总结",
      "ratingHint": "点击星星评分",
      "npsHeadTitle": "发起人专属",
      "npsQuestion": "你会推荐“真好玩”给朋友吗？ (NPS)",
      "npsLowLabel": "不可能",
      "npsHighLabel": "极有可能",
      "submitText": "提交评价",
      "submitNote": "评价内容仅双方可见，请客观公正",
      "skipToast": "已跳过评价",
      "submitSuccessText": "评价已提交",
      "noReviewTargetText": "暂无可评价对象",
      "missingTargetText": "缺少评价对象，无法提交",
      "missingScoreText": "请先为每个评价对象打分",
      "submitFailedText": "提交评价失败",
      "defaultSummary": "本次合作沟通顺畅，交付清晰，整体体验不错。",
      "againIntentBySatisfaction": { "great": "yes", "ok": "maybe", "bad": "no" },
      "roleConfigs": {
        "expert": {
          "id": "expert",
          "avatarText": "ZH",
          "avatarTheme": "blue",
          "title": "评价行家",
          "desc": "本次服务已完成",
          "ratingTitle": "服务质量",
          "tagTitle": "行家标签（多选）",
          "tags": ["专业能力强", "交付及时", "沟通顺畅", "超出预期", "性价比高", "推荐再合作"],
          "placeholder": "分享你对本次服务的评价..."
        },
        "player": {
          "id": "player",
          "avatarText": "WA",
          "avatarTheme": "pink",
          "title": "评价玩家",
          "desc": "需求已确认，开始反馈",
          "ratingTitle": "合作满意度",
          "tagTitle": "玩家标签（多选）",
          "tags": ["需求明确", "配合度高", "付款及时", "沟通友好", "长期合作潜力"],
          "placeholder": "写下你对需求方的评价..."
        },
        "guide": {
          "id": "guide",
          "avatarText": "WA",
          "avatarTheme": "orange",
          "title": "评价领路人",
          "desc": "撮合已完成，协助交付",
          "ratingTitle": "引荐满意度",
          "tagTitle": "邀约标签（多选）",
          "tags": ["匹配精准", "响应及时", "协助积极", "沟通高效", "值得信赖"],
          "placeholder": "写下你对引荐人的服务评价..."
        }
      },
      "version": "2026-07-01"
    }'::jsonb,
    'active',
    now()
  ),
  (
    'review.complete_config',
    '{
      "reward": { "show": false },
      "benefits": [
        { "iconText": "⭐", "theme": "blue", "title": "优先推荐权益", "desc": "下次组局时优先展示给行家", "order": 10, "visible": true },
        { "iconText": "👑", "theme": "purple", "title": "评价达人徽章", "desc": "累计评价3次可获得专属标识", "order": 20, "visible": true },
        { "iconText": "📈", "theme": "red", "title": "信用分提升", "desc": "活跃评价有助于提升账号权重", "order": 30, "visible": true }
      ],
      "playOptions": [
        { "id": "again", "theme": "green", "title": "再玩一局", "desc": "随时可约", "intent": "yes", "route": "play_again", "order": 10, "visible": true },
        { "id": "pause", "theme": "yellow", "title": "暂停", "desc": "想歇歇", "intent": "maybe", "route": "game_hall", "order": 20, "visible": true },
        { "id": "stop", "theme": "red", "title": "不玩了", "desc": "不再参与", "intent": "no", "route": "game_hall", "order": 30, "visible": true }
      ],
      "version": "2026-07-01"
    }'::jsonb,
    'active',
    now()
  ),
  (
    'growth.achievement_config',
    '{
      "filters": [
        { "key": "all", "label": "全部", "order": 10 },
        { "key": "city", "label": "点亮城市", "order": 20 },
        { "key": "streak", "label": "连续打卡", "order": 30 }
      ],
      "catalog": [
        { "id": "first_review", "code": "first_review", "title": "首次评价", "desc": "完成第一次服务评价", "icon": "/pages/profile/footprint/achievements/assets/icon-star.png", "tone": "gold", "category": "city", "statusText": "已解锁", "order": 10, "visible": true },
        { "id": "first_received_review", "code": "first_received_review", "title": "首次收到评价", "desc": "收到第一条服务评价", "icon": "/pages/profile/footprint/achievements/assets/icon-star.png", "tone": "gold", "category": "city", "statusText": "已解锁", "order": 20, "visible": true },
        { "id": "completed_game", "code": "completed_game", "title": "完成一局", "desc": "完成一次组局服务确认", "icon": "/pages/profile/footprint/achievements/assets/icon-star.png", "tone": "green", "category": "city", "statusText": "已解锁", "order": 30, "visible": true }
      ],
      "locked": [
        { "id": "credit_keeper", "code": "credit_keeper", "title": "信用守护", "desc": "信用分保持 80 分以上", "icon": "/pages/profile/footprint/achievements/assets/icon-star.png", "tone": "blue", "category": "streak", "statusText": "进行中", "progressPercent": 80, "order": 10, "visible": true },
        { "id": "review_master", "code": "review_master", "title": "评价达人", "desc": "累计完成 3 次服务评价", "icon": "/pages/profile/footprint/achievements/assets/icon-star.png", "tone": "gold", "category": "streak", "statusText": "进行中", "progressPercent": 30, "order": 20, "visible": true }
      ],
      "season": {
        "title": "成长赛季",
        "status": "进行中",
        "remainTpl": "已沉淀 {footprintCount} 条足迹"
      },
      "onlineSuffix": "人成长中",
      "levelTitlePrefix": "探索行家 Lv.",
      "version": "2026-07-01"
    }'::jsonb,
    'active',
    now()
  ),
  (
    'role.benefit_config',
    '{
      "version": "2026-07-01",
      "permissionPrompts": {
        "expert": {
          "roleType": "expert",
          "title": "我懂玩家需要什么！我申请成为行家",
          "primary": "申请成为行家",
          "secondary": "查看权益对比"
        },
        "guide": {
          "roleType": "guide",
          "title": "我愿意带领更多人一起玩！我申请成为领路人",
          "primary": "申请成为领路人",
          "secondary": "查看权益对比"
        }
      },
      "roleComparison": {
        "name": "权益对比页",
        "mode": "roleComparison",
        "roleBadge": "权益",
        "title": "角色权益对比",
        "subtitle": "选择适合你的角色，开启不同玩法",
        "roles": [
          { "key": "player", "name": "玩家", "level": "Lv.1+", "active": true },
          { "key": "guide", "name": "领路人", "level": "Lv.5+", "active": false },
          { "key": "expert", "name": "行家", "level": "Lv.20+", "active": false }
        ],
        "benefits": [
          { "name": "发起组局", "player": "✓", "leader": "—", "expert": "✓" },
          { "name": "加入组局", "player": "✓", "leader": "✓", "expert": "✓" },
          { "name": "创建路线", "player": "✓", "leader": "—", "expert": "✓" },
          { "name": "分润收益", "player": "—", "leader": "基础会员40%", "expert": "高级会员40%" },
          { "name": "服务交易", "player": "—", "leader": "—", "expert": "✓" },
          { "name": "数据看板", "player": "—", "leader": "✓", "expert": "✓" },
          { "name": "信用背书", "player": "—", "leader": "✓", "expert": "✓" }
        ],
        "primary": "立即申请角色"
      }
    }'::jsonb,
    'active',
    now()
  ),
  (
    'membership.page_config',
    '{
      "items": [
        {
          "key": "basic",
          "code": "basic",
          "memberLevel": "基础会员",
          "name": "基础会员",
          "cardClass": "basic",
          "benefitClass": "purple",
          "primaryBenefitTitle": "业务引荐权益",
          "price": "515",
          "profitRate": "40%",
          "noticeLevel": "高级会员",
          "noticeAvatar": "林",
          "noticeName": "林丽 总",
          "agreementKey": "user-service",
          "agreementTitle": "服务协议",
          "referralBenefits": ["可引荐平台业务", "享受引荐收益"],
          "audience": ["有人脉、善对接的社交达人、资源型人才；", "希望不做销售、不投重金，只靠人脉赚钱；", "有高客单价产品/资源，想初步了解；", "连接供需，促成交易"],
          "openRules": {
            "step": "1. 选择会员等级 → 2. 在线支付 → 3. 即时生效",
            "notes": ["支持微信支付、支持银行卡支付", "升级后原有权益自动叠加，不重复收费", "如需帮助，请在客服中心提交咨询"]
          }
        },
        {
          "key": "advanced",
          "code": "advanced",
          "memberLevel": "高级会员",
          "name": "高级会员",
          "cardClass": "advanced",
          "benefitClass": "advanced",
          "primaryBenefitTitle": "业务被引荐权益",
          "price": "10000",
          "profitRate": "40%",
          "noticeLevel": "高级会员",
          "noticeAvatar": "林",
          "noticeName": "林丽 总",
          "agreementKey": "user-service",
          "agreementTitle": "服务协议",
          "referralBenefits": ["您的产品或服务进入引荐池", "平台引荐人主动为您引荐", "享受被引荐带来的40%收益分成", "提供推广数据报表和分析"],
          "audience": ["有高客单价产品/资源、有技术的专业人士", "希望获得额外收入来源"],
          "openRules": {
            "step": "1. 选择会员等级 → 2. 在线支付 → 3. 即时生效",
            "notes": ["支持微信支付、支持银行卡支付", "升级后原有权益自动叠加，不重复收费", "如需帮助，请在客服中心提交咨询"]
          }
        },
        {
          "key": "premium",
          "code": "premium",
          "memberLevel": "尊享会员",
          "name": "尊享会员",
          "cardClass": "premium",
          "benefitClass": "premium",
          "primaryBenefitTitle": "渠道引荐权益",
          "price": "39800",
          "profitRate": "10%",
          "noticeLevel": "高级会员",
          "noticeAvatar": "林",
          "noticeName": "林丽 总",
          "agreementKey": "user-service",
          "agreementTitle": "服务协议",
          "referralBenefits": ["可发展和管理下级推广团队", "团队引荐收益的10%作为管理奖励", "提供团队管理工具和数据看板"],
          "audience": ["有人脉、爱分享、想裂变的推广能手", "团队管理者、团长、行业主理人", "拥有推广资源的个人/机构", "希望建立推广体系的创业者"],
          "openRules": {
            "step": "1. 选择会员等级 → 2. 在线支付 → 3. 即时生效",
            "notes": ["支持微信支付、支持银行卡支付", "升级后原有权益自动叠加，不重复收费", "如需帮助，请在客服中心提交咨询"]
          }
        }
      ]
    }'::jsonb,
    'active',
    now()
  ),
  (
    'report.center_config',
    '{
      "types": [
        { "key": "private-guide", "label": "诱导私下交易", "reportType": "revenue_dispute", "order": 10, "visible": true },
        { "key": "private-done", "label": "私下交易已完成", "reportType": "revenue_dispute", "order": 20, "visible": true },
        { "key": "harassment", "label": "言语骚扰", "reportType": "user_complaint", "order": 30, "visible": true },
        { "key": "fake", "label": "虚假信息", "reportType": "user_complaint", "order": 40, "visible": true },
        { "key": "cancel", "label": "恶意取消", "reportType": "service_dispute", "order": 50, "visible": true },
        { "key": "other", "label": "其他违规", "reportType": "other", "order": 60, "visible": true }
      ],
      "defaultType": "private-guide",
      "maxEvidenceCount": 9,
      "allowedUploadTypes": ["jpg", "png", "pdf"],
      "tips": [
        "举报属实且能核实金额：罚款20% (50%奖励举报人)",
        "属实但无法核实：按后台处理规则发放奖励并记录信用变化",
        "不属实扣除举报人信用分2分，多次恶意举报封号"
      ],
      "appealReasons": [
        { "key": "misjudge", "label": "误判扣分", "order": 10, "visible": true },
        { "key": "system", "label": "系统错误", "order": 20, "visible": true },
        { "key": "special", "label": "特殊情况", "order": 30, "visible": true },
        { "key": "other", "label": "其他", "order": 40, "visible": true }
      ],
      "appealPlaceholder": "请详细说明申诉原因，包括但不限于事件经过、时间、涉及人员等信息...",
      "appealUploadNote": "支持 JPG、PNG 格式，单张不超过 5MB，最多 4 张证明材料",
      "appealFileMaxCount": 4,
      "appealUploadFullText": "最多上传 4 张证明材料",
      "appealUploadSelectedTemplate": "已选择 {selected}/{max} 张证明材料",
      "appealReviewTitle": "处理时效",
      "appealReviewRules": [
        "提交后24小时内初审",
        "复杂情况48小时内复核",
        "结果将通过站内消息通知"
      ],
      "version": "2026-07-01"
    }'::jsonb,
    'active',
    now()
  ),
  (
    'membership.radar_config',
    '{
      "pages": {
        "radar": {
          "key": "radar",
          "title": "组局雷达",
          "skin": "dark",
          "estimate": "预计可匹配7461位商界决策者",
          "primaryAction": "开启适配人脉",
          "tip": "信息填写越完整，人脉匹配越精准",
          "linkText": "填写适配信息 >"
        },
        "matching": {
          "key": "matching",
          "title": "组局雷达",
          "skin": "dark",
          "statusText": "人脉雷达正在寻找与您适配的企业家…"
        },
        "info": {
          "key": "info",
          "title": "适配信息",
          "skin": "light"
        },
        "query": {
          "key": "query",
          "title": "组局雷达",
          "skin": "dark result",
          "foundPrefix": "为您找到",
          "foundSuffix": "位适配您的优质行家信息"
        },
        "result": {
          "key": "result",
          "title": "组局雷达",
          "skin": "dark result",
          "loadingText": "正在寻找与您适配的优质业务主.....",
          "resultTitle": "本轮匹配组局已完成推荐",
          "resultDescPrefix": "共推荐了",
          "resultDescSuffix": "位优质行家",
          "resultLink": "重新查看 >",
          "actionText": "再次重新匹配"
        }
      },
      "formRows": [
        { "key": "location", "label": "地址定位", "value": "", "placeholder": "选择" },
        { "key": "industry", "label": "所在行业", "value": "", "placeholder": "选择" },
        { "key": "revenueScale", "label": "营收规模", "value": "", "placeholder": "选填" },
        { "key": "interestedGames", "label": "感兴趣组局", "value": "", "placeholder": "选择" },
        { "key": "resources", "label": "我的资源", "value": "", "placeholder": "前往个人主页填写" },
        { "key": "needs", "label": "我的需求", "value": "", "placeholder": "前往个人主页填写" },
        { "key": "recentDemand", "label": "近期诉求", "value": "", "placeholder": "自定义填写" }
      ],
      "profile": {
        "id": "lu-yi",
        "userId": 201,
        "name": "陆毅",
        "title": "总经理｜上海创世界科技有限公司",
        "tag": "第一标签：上海TMT投资领军者，数字化内容服务",
        "need": "我的需求：AI赋能与市场运营助力企业IP打造",
        "resource": "我的资源：10年TMT投资经验",
        "address": "上海市浦东新区沙新镇黄赵路310号",
        "distance": "231 km",
        "avatar": "/pages/profile/member/assets/radar-avatar.png"
      },
      "radarNodes": [
        { "id": "hu-fang", "userId": 200, "className": "node-leader", "name": "胡芳", "title": "董事长、创始人｜千浪化研新材料（上海…", "avatar": "/pages/profile/member/assets/radar-avatar.png" },
        { "id": "lu-yi", "userId": 201, "className": "node-maker", "name": "陆毅", "title": "总经理｜上海创世界科技有限公司", "avatar": "/pages/profile/member/assets/radar-avatar.png" },
        { "id": "chen-zong", "userId": 202, "className": "node-owner", "name": "陈总", "title": "企业服务资源方", "avatar": "", "shortName": "陈" },
        { "id": "wang-zong", "userId": 203, "className": "node-investor", "name": "王总", "title": "产业投资合伙人", "avatar": "", "shortName": "王" },
        { "id": "li-zong", "userId": 204, "className": "node-expert", "name": "李总", "title": "品牌增长顾问", "avatar": "", "shortName": "李" },
        { "id": "zhao-zong", "userId": 205, "className": "node-partner", "name": "赵总", "title": "渠道合作伙伴", "avatar": "", "shortName": "赵" },
        { "id": "sun-zong", "userId": 206, "className": "node-small", "name": "孙总", "title": "本地服务主理人", "avatar": "", "shortName": "孙" }
      ],
      "result": {
        "total": 10
      },
      "texts": {
        "criteriaMatchLabel": "符合条件的企业家",
        "allMatchLabel": "适配企业家",
        "criteriaScanningText": "人脉雷达正在按您的适配信息寻找企业家…",
        "allScanningText": "人脉雷达正在为您匹配全部适配企业家…",
        "scanDoneTemplate": "已扫描到 {count} 位{label}",
        "scanProgressTemplate": "正在扫描，已发现 {count} 位{label}",
        "actionFailedText": "人脉雷达操作失败",
        "entryMissingText": "请选择可用入口"
      },
      "actionMessages": {
        "save": "已保存匹配偏好",
        "next": "已为你刷新下一位",
        "follow": "已关注该成员",
        "profile": "暂无成员主页",
        "share": "请使用右上角分享"
      }
    }'::jsonb,
    'active',
    now()
  ),
  (
    'game.my_games_page_config',
    '{
      "pageTitle": "我的局",
      "emptyText": "暂无相关局",
      "detailMissing": "暂无组局详情",
      "actionMissing": "暂无可执行操作",
      "categoryTabs": [
        { "key": "joined", "text": "我参与的" },
        { "key": "invited", "text": "我受邀的" },
        { "key": "favorite", "text": "我收藏的" }
      ],
      "statusTabs": [
        { "key": "all", "text": "全部" },
        { "key": "active", "text": "进行中" },
        { "key": "complete", "text": "已完成" },
        { "key": "overdue", "text": "超时" },
        { "key": "canceled", "text": "已取消" }
      ],
      "version": "2026-07-01"
    }'::jsonb,
    'active',
    now()
  ),
  (
    'game.referral_records_config',
    '{
      "pageTitle": "我的引荐记录",
      "summary": {
        "label": "本月引荐收益",
        "background": "linear-gradient(135deg, #ffb347 0%, #ff7b00 100%)",
        "iconSrc": "/pages/game/referral-record/assets/wallet.png",
        "statTemplates": {
          "success": "成功 {count}单",
          "processing": "进行中 {count}单",
          "review": "待评价 {count}单"
        }
      },
      "tabs": [
        { "key": "processing", "label": "进行中" },
        { "key": "completed", "label": "已完成" },
        { "key": "canceled", "label": "已取消" }
      ],
      "texts": {
        "emptyText": "暂无引荐记录",
        "loadFailedText": "引荐记录加载失败",
        "expertRoleText": "行家",
        "playerRoleText": "玩家",
        "selfLabel": "我",
        "reviewedTagText": "已评价",
        "pendingReviewTagText": "待评价",
        "reviewedActionText": "已评价",
        "reviewActionText": "评价双方",
        "remindActionText": "提醒交付",
        "chatActionText": "查看群聊",
        "chatPrefill": "你好，我想查看本次引荐服务的群聊进度。",
        "unavailableText": "该操作暂不可用",
        "remindMessageTemplate": "请及时确认交付：{serviceTitle}",
        "remindServiceFallback": "引荐服务",
        "remindSuccessText": "已提醒交付",
        "remindFailedText": "提醒交付失败",
        "rewardPrefix": "¥"
      }
    }'::jsonb,
    'active',
    now()
  ),
  (
    'message.trade_warning_config',
    '{
      "pageTitle": "交易预警",
      "onlineText": "在线",
      "warning": { "title": "交易预警", "prefixText": "", "highlightText": "", "suffixText": "" },
      "countdown": [],
      "order": { "orderNo": "", "statusText": "", "customerAvatarText": "", "customerTitle": "", "customerDesc": "", "detailRows": [] },
      "deliveryMethods": [
        { "id": "online", "title": "线上确认", "desc": "双方在线确认服务完成", "active": true },
        { "id": "upload", "title": "上传凭证", "desc": "上传服务完成截图或文件", "active": false }
      ],
      "actions": { "delayText": "申请延期", "deliverText": "立即交付" },
      "texts": {
        "loadingText": "加载中...",
        "loadFailedText": "获取交易预警失败",
        "invalidActionText": "交易预警操作无效",
        "actionFailedText": "交易预警处理失败",
        "delaySuccessText": "延期申请已提交",
        "delayStatusText": "已申请延期",
        "deliverSuccessText": "已进入交付确认",
        "countdownTitle": "剩余交付时间",
        "orderNoLabel": "订单编号",
        "deliveryTitle": "交付方式"
      }
    }'::jsonb,
    'active',
    now()
  ),
  (
    'message.system_notification_config',
    '{
      "pageTitle": "系统通知",
      "onlineText": "在线",
      "article": {
        "tagText": "系统通知",
        "title": "",
        "author": "",
        "publishedAtText": "",
        "readText": "",
        "blocks": []
      },
      "feedback": {
        "question": "这篇通知对你有帮助吗？",
        "useful": { "icon": "赞", "label": "有用", "countText": "0" },
        "useless": { "icon": "踩", "label": "没用", "countText": "0" }
      },
      "texts": {
        "loadFailedText": "获取系统通知失败",
        "feedbackFailedText": "反馈提交失败",
        "feedbackSuccessText": "已记录{label}反馈"
      }
    }'::jsonb,
    'active',
    now()
  ),
  (
    'message.center_config',
    '{
      "pageTitle": "消息中心",
      "onlineText": "在线",
      "quickActions": [
        { "key": "join", "label": "组局加入", "iconSrc": "/pages/message/assets/i53@3x.png", "tone": "blue", "bucket": "group" },
        { "key": "system", "label": "系统通知", "iconSrc": "/pages/message/assets/i54@3x.png", "tone": "green", "bucket": "system" },
        { "key": "achievement", "label": "成就解锁", "iconSrc": "/pages/message/assets/i55@3x.png", "tone": "yellow", "bucket": "achievement" },
        { "key": "warning", "label": "预警通知", "iconSrc": "/pages/message/assets/i56@3x.png", "tone": "red", "bucket": "warning" },
        { "key": "friend", "label": "好友", "iconSrc": "/pages/message/assets/i57@3x.png", "tone": "cyan", "bucket": "friend" }
      ],
      "tabs": [
        { "key": "all", "label": "全部消息", "countSource": "unread" },
        { "key": "unread", "label": "未读", "countSource": "unread" },
        { "key": "trade", "label": "交易通知", "countSource": "trade" }
      ],
      "sections": [
        { "key": "group", "title": "组局动态", "bucket": "group" },
        { "key": "system", "title": "系统通知", "bucket": "system" },
        { "key": "warning", "title": "预警提醒", "bucket": "warning" }
      ],
      "actionTexts": {
        "accept": "确认参加",
        "reject": "婉拒",
        "process": "立即处理",
        "review": "立即评价",
        "detail": "查看详情",
        "game": "查看组局",
        "contact": "联系发起人"
      },
      "texts": {
        "loadFailedText": "消息中心加载失败",
        "entryMissingText": "暂无可打开的消息入口",
        "openFailedText": "消息打开失败",
        "actionMissingText": "操作信息不完整",
        "actionSuccessText": "操作成功",
        "actionHandledText": "已标记处理",
        "actionFailedText": "消息操作失败",
        "serviceMissingText": "缺少消息操作信息",
        "serviceFailedText": "消息操作失败"
      }
    }'::jsonb,
    'active',
    now()
  ),
  (
    'message.my_config',
    '{
      "pageTitle": "好友消息",
      "onlineText": "在线",
      "friend": {
        "defaultInitials": "IM",
        "defaultName": "局内会话",
        "defaultStatus": "在线",
        "nameTemplate": "成员 {userId}"
      },
      "quickActions": [
        { "key": "friend", "label": "加好友" },
        { "key": "greet", "label": "打招呼", "messageText": "你好，我看到你的消息了。" },
        { "key": "card", "label": "发名片", "messageText": "这是我的名片，后续可以在局内继续沟通。" },
        { "key": "location", "label": "发定位" }
      ],
      "texts": {
        "loadFailedText": "加载会话失败",
        "actionMissingText": "操作信息不完整",
        "sendFailedText": "发送失败",
        "recordStartText": "开始录音",
        "recordStopText": "当前支持文字、图片和文件消息",
        "recordErrorText": "录音失败",
        "fileEntryMissingText": "请从局内消息入口发送文件",
        "justNowText": "刚刚"
      }
    }'::jsonb,
    'active',
    now()
  ),
  (
    'game.delivery_page_config',
    '{
      "paid": {
        "pageTitle": "确认服务完成",
        "status": { "theme": "paid", "title": "服务已完成!", "desc": "双方确认后，资金将全额结算" },
        "statePill": { "theme": "green", "text": "待确认完成" },
        "notice": {},
        "confirmItems": [
          { "id": "completed", "title": "服务已全部完成", "desc": "约定的2小时咨询服务已完整交付", "checked": false },
          { "id": "qualified", "title": "服务质量达标", "desc": "需求方对服务内容和质量无异议", "checked": false },
          { "id": "communicated", "title": "双方已沟通确认", "desc": "已与需求方确认服务完成，对方同意结算", "checked": false }
        ],
        "confirmNote": "正常交付无需扣减任何费用，只需双方确认服务已完成，资金将按全额结算。如服务未完全达标，请与玩家沟通后再确认。",
        "security": { "title": "", "desc": "" },
        "submitHints": { "ready": "确认后将通知玩家进行最终确认", "pending": "需勾选上方确认项后方可提交" },
        "submitToast": "服务完成确认已提交",
        "submitLoadingText": "提交中",
        "amountRowLabel": "合同金额"
      },
      "free": {
        "pageTitle": "确认服务完成",
        "status": { "theme": "free", "title": "服务已完成!", "desc": "双方确认后，服务正式结束" },
        "statePill": { "theme": "blue", "text": "待确认完成" },
        "notice": {
          "iconText": "🎁",
          "title": "免费局说明",
          "parts": [
            { "text": "本局为" },
            { "text": "免费体验局", "strong": true },
            { "text": "不涉及资金结算。双方确认完成后，行家将获得" },
            { "text": "信用积分+5和免费局贡献徽章", "strong": true },
            { "text": "，玩家" },
            { "text": "优先推荐权益", "strong": true }
          ]
        },
        "confirmItems": [
          { "id": "completed", "title": "服务已全部完成", "desc": "约定的2小时咨询服务已完整交付", "checked": true, "locked": true },
          { "id": "qualified", "title": "服务质量达标", "desc": "需求方对服务内容和质量无异议", "checked": true, "locked": true },
          { "id": "communicated", "title": "双方已沟通确认", "desc": "已与需求方确认服务完成，对方同意归档", "checked": false }
        ],
        "confirmNote": "免费局无需扣除任何费用，只需双方确认服务已完成，系统将自动归档。如服务未完全达标，请与玩家沟通后再次确认。",
        "security": { "title": "服务保障", "desc": "免费局同样享受平台服务保障，评价真实有效" },
        "submitHints": { "ready": "确认后将通知玩家进行最终确认", "pending": "需勾选上方确认项后方可提交" },
        "submitToast": "免费局服务完成确认已提交",
        "submitLoadingText": "提交中",
        "amountRowLabel": "服务类型"
      },
      "quickActions": [
        { "key": "upload", "title": "上传凭证", "theme": "blue", "iconText": "📎" },
        { "key": "contact_player", "title": "联系玩家", "theme": "blue", "iconSrc": "/pages/game/delivery/assets/i18@3x.png" },
        { "key": "contact_guide", "title": "联系领路人", "theme": "orange", "iconText": "👬" }
      ],
      "version": "2026-07-01"
    }'::jsonb,
    'active',
    now()
  ),
  (
    'game.system_recommendations_config',
    '{
      "title": "系统推荐适配局",
      "desc": "暂无推荐结果，请完善资料或稍后重试",
      "loadingText": "推荐数据加载中",
      "emptyText": "暂无匹配行家",
      "summaryTemplate": "已选择 {count} 位行家",
      "summaryDesc": "还可以选择多位行家组成顾问团，或搭配玩家共同组局",
      "cancelText": "取消",
      "confirmText": "确认组局",
      "minSelectToast": "请选择至少一位行家",
      "confirmingText": "正在进入组局",
      "defaultCategory": "all",
      "categories": [
        { "key": "all", "name": "全部" },
        { "key": "product", "name": "产品架构" },
        { "key": "tech", "name": "技术咨询" },
        { "key": "operation", "name": "运营策略" }
      ],
      "version": "2026-07-01"
    }'::jsonb,
    'active',
    now()
  ),
  (
    'map.index_config',
    '{
      "onlineText": "在线",
      "defaultLocation": { "latitude": 31.2304, "longitude": 121.4737 },
      "defaultRadiusMeters": 3000,
      "radiusOptions": [1000, 3000, 5000],
      "mapFilters": ["附近组局", "组局路线", "热力图", "好友分布", "解锁图鉴", "AR"],
      "texts": {
        "searchPlaceholder": "搜局、搜人、搜地块...",
        "loadingNearbyText": "正在获取附近数据...",
        "locateToolText": "定",
        "refreshToolText": "刷",
        "dateRangeText": "2025.09.23 - 2026.03.30",
        "detailActionText": "查看详情",
        "joinActionText": "去组队",
        "playerDetailActionText": "查看资料",
        "currentInteractText": "当前可互动",
        "offlineInteractText": "暂未在线，可查看轨迹",
        "nearbySectionTitle": "附近玩法点",
        "nearbyEmptyText": "当前位置附近暂无玩法点"
      },
      "version": "2026-07-01"
    }'::jsonb,
    'active',
    now()
  ),
  (
    'map.my_city_config',
    '{
      "onlineText": "在线",
      "pageTitle": "我的城市故事",
      "profileName": "我的信息",
      "routeTip": "收集8条更早行程，航线图更完整",
      "switchMapText": "切换为火车",
      "journeyTitle": "我的局迹",
      "journeyDesc": "通过 12 个局，认识了 28 位朋友",
      "participantLabel": "参与者：",
      "endingTitle": "我是有底线的",
      "endingDesc": "继续探索，创造更多故事",
      "summaryStats": [
        { "value": "28", "label": "故事总数", "tone": "blue" },
        { "value": "6", "label": "覆盖城市", "tone": "violet" },
        { "value": "52", "label": "参与组局数", "tone": "pink" }
      ],
      "mapLegends": [
        { "label": "第一次", "tone": "pink" },
        { "label": "夜游", "tone": "violet" },
        { "label": "社交局", "tone": "blue" },
        { "label": "最难忘", "tone": "gold" }
      ],
      "mapStats": [
        { "label": "里程", "value": "40244", "unit": "公里" },
        { "label": "次数", "value": "33", "unit": "次" },
        { "label": "国家/地区", "value": "1", "unit": "个" },
        { "label": "城市", "value": "10", "unit": "个" }
      ],
      "tagEmojis": {
        "最难忘": "👑",
        "桌游局": "🎲",
        "微醺局": "🍷",
        "脑暴局": "💡",
        "篮球局": "🏀",
        "第一次": "🌱",
        "起点": "🌱",
        "摄影局": "📷"
      },
      "badgeEmojis": {
        "创业伙伴": "🤝",
        "深度密友": "💬",
        "合伙人": "🤝",
        "室友": "🏠",
        "固定局友": "📌"
      },
      "storyGroups": [
        {
          "year": 2024,
          "stories": [
            { "id": "countdown-night", "tag": "最难忘", "tagTone": "gold", "title": "跨年夜的倒计时", "date": "12.31", "sortDate": "2024-12-31", "location": "", "cover": "/pages/map/my-city/assets/story-river-cover.png", "coverLocation": "上海 · 外滩", "desc": "和刚认识的摄影局朋友们一起在外滩等待新年钟声。江风吹得发抖，但倒数的那一刻，所有的陌生人都变成了朋友。", "participants": ["a", "b", "c"], "badgeTitle": "", "badgeDesc": "", "actionText": "", "actionTone": "" },
            { "id": "script-rain", "tag": "桌游局", "subTag": "新手场", "tagTone": "violet", "title": "暴雨中的剧本杀", "date": "2024.09.20", "sortDate": "2024-09-20", "location": "杭州 · 西湖区 · 14:00-22:00", "desc": "原定5人的局因为暴雨只来了3人，却因此有了最深入的交谈。认识了做AI的@阿杰，现在我们是创业合伙人。", "participants": ["a", "b", "c", "d"], "badgeTitle": "创业伙伴", "badgeDesc": "已共同发起 3 个项目", "actionText": "查看项目", "actionTone": "violet" },
            { "id": "truth-night", "tag": "微醺局", "subTag": "深夜场", "tagTone": "orange", "title": "周五晚上的坦白局", "date": "2024.11.03", "sortDate": "2024-11-03", "location": "北京 · 三里屯 · 21:00", "desc": "\"你最后悔的事是什么？\"那个问题让陌生人变成了知己。和@Lucy约定每月一次深度对话。", "participants": ["a", "b"], "badgeTitle": "深度密友", "badgeDesc": "", "actionText": "约下次", "actionTone": "orange" },
            { "id": "business-canvas", "tag": "脑暴局", "subTag": "创始人专场", "tagTone": "gold", "title": "凌晨的商业模式画布", "date": "2024.12.15", "sortDate": "2024-12-15", "location": "深圳 · 科技园 · 通宵", "desc": "从晚上8点到早上6点，8个人在黑板上画满了想法。这个局让我找到了技术合伙人@老K。", "participants": ["a", "b", "c", "d"], "badgeTitle": "合伙人", "badgeDesc": "公司估值 500w", "actionText": "查看公司", "actionTone": "gold" },
            { "id": "court-weekly", "tag": "篮球局", "subTag": "每周固定", "tagTone": "cyan", "title": "东华球场的汗水", "date": "每周六", "sortDate": "2024-06-15", "location": "上海 · 东华大学 · 16:00", "desc": "最纯粹的快乐。这里没有身份，只有队友。通过球局认识了现在的室友@阿强。", "participants": ["a", "b", "c"], "badgeTitle": "室友", "badgeDesc": "合租 6 个月", "actionText": "加入球局", "actionTone": "cyan" },
            { "id": "first-use", "type": "first", "tag": "起点", "subTag": "", "tagTone": "green", "highlightTag": "第一次", "title": "第一次使用真好玩", "date": "03.12", "sortDate": "2024-03-12", "location": "广州 · 天河公园", "desc": "抱着试试看的心态参加了第一次飞盘局，从此打开了城市探索的新世界。", "participants": [], "badgeTitle": "", "badgeDesc": "", "actionText": "", "actionTone": "" }
          ]
        },
        {
          "year": 2023,
          "stories": [
            { "id": "sunrise-photo", "tag": "摄影局", "subTag": "第1次参与", "tagTone": "pink", "title": "外滩 sunrise 拍摄", "date": "2023.03.12", "sortDate": "2023-03-12", "location": "上海 · 外滩观景台 · 06:00", "desc": "为了拍日出早上5点起床，认识了同样疯狂的@小林和@大为。后来我们组成了固定摄影小队，每周六早扫街。", "participants": ["a", "b", "c"], "badgeTitle": "固定局友", "badgeDesc": "已持续组队 8 个月", "actionText": "再组一局", "actionTone": "blue" }
          ]
        }
      ],
      "version": "2026-06-30"
    }'::jsonb,
    'active',
    now()
  ),
  (
    'map.play_pages_config',
    '{
      "version": "2026-07-01",
      "pages": {
        "blind-route": {
          "title": "组局盲盒",
          "description": "不知道去哪局？让命运决定你的下一次城市冒险",
          "sectionTitle": "最近开启",
          "selectToast": "已选择 {title}",
          "cards": [
            { "id": "walk", "title": "城市漫步盲盒", "desc": "30 分钟内出发，随机匹配附近轻量局", "tone": "blue", "icon": "/pages/map/blind-route/assets/i50.png", "tags": ["轻松", "附近", "低门槛"] },
            { "id": "food", "title": "深夜食堂盲盒", "desc": "匹配同城饭搭子和夜宵路线", "tone": "orange", "icon": "/pages/map/blind-route/assets/i52.png", "tags": ["饭局", "夜间", "社交"] },
            { "id": "photo", "title": "拍照路线盲盒", "desc": "用一个主题串起三处城市机位", "tone": "purple", "icon": "/pages/map/blind-route/assets/i54.png", "tags": ["拍照", "路线", "打卡"] }
          ],
          "recentRoutes": [
            { "title": "徐汇夜风路线", "timeText": "20 分钟前", "statusText": "已成局" },
            { "title": "周末咖啡搭子", "timeText": "1 小时前", "statusText": "招募中" }
          ]
        },
        "city-atlas": {
          "title": "上海探索图鉴",
          "progressLabel": "探索进度",
          "lockedText": "未解锁",
          "hiddenBadge": "隐藏",
          "routeSectionTitle": "主题路线",
          "filters": ["全部", "已解锁", "未解锁", "隐藏点"],
          "unlockInfo": { "unlockedCount": 1, "totalCount": 36, "progressPercent": 33 },
          "unlockToast": "{name}已解锁",
          "lockedToast": "{name}待解锁",
          "unlockPoints": [
            { "id": "bund-night", "name": "外滩夜景", "statusType": "unlocked", "unlockText": "2024.01.15 解锁", "footprintValue": "+50 足迹值", "tone": "blue", "iconType": "building", "checked": true },
            { "id": "tianzifang", "name": "田子坊", "statusType": "unlocked", "unlockText": "2024.02.03 解锁", "footprintValue": "+30 足迹值", "tone": "green", "iconType": "lantern", "checked": true },
            { "id": "wukang-road", "name": "武康路街角", "statusType": "locked", "unlockText": "完成 2 次附近打卡后解锁", "footprintValue": "+40 足迹值", "tone": "locked", "iconType": "lock", "checked": false },
            { "id": "hidden-rooftop", "name": "城市天台", "statusType": "hidden", "unlockText": "隐藏点待发现", "footprintValue": "+80 足迹值", "tone": "purple", "iconType": "hidden", "checked": false }
          ],
          "themeRoutes": [
            { "id": "couple-walk", "name": "情侣漫步", "meta": "6个地点 · 预计3小", "progressText": "已解锁 2/6", "tone": "sunset", "iconText": "💕" },
            { "id": "coffee-shop", "name": "咖啡探店", "meta": "8个地点 · 预计4小", "progressText": "已解锁 0/8", "tone": "cyan", "iconText": "☕" }
          ]
        },
        "footprint-heatmap": {
          "title": "足迹热力图",
          "heatTitle": "全国城市打卡热力",
          "legendLabel": "城市打卡热度",
          "friendTitle": "好友也在打卡",
          "friendMoreText": "查看全部",
          "hotTitle": "城市热点排行",
          "rangeTabs": ["今日", "本周", "本月", "全部"],
          "rangeStats": {
            "今日": [{ "value": "12", "label": "打卡城市" }, { "value": "1.8k", "label": "玩家足迹" }, { "value": "3", "label": "热门城市" }],
            "本周": [{ "value": "38", "label": "打卡城市" }, { "value": "8.5k", "label": "玩家足迹" }, { "value": "9", "label": "热门城市" }],
            "本月": [{ "value": "76", "label": "打卡城市" }, { "value": "26k", "label": "玩家足迹" }, { "value": "18", "label": "热门城市" }],
            "全部": [{ "value": "126", "label": "打卡城市" }, { "value": "92k", "label": "玩家足迹" }, { "value": "31", "label": "热门城市" }]
          },
          "cityHeatPoints": [
            { "id": "beijing", "city": "北京", "level": "mid", "className": "footprint-city-point beijing level-mid" },
            { "id": "shanghai", "city": "上海", "level": "hot", "className": "footprint-city-point shanghai level-hot" },
            { "id": "chengdu", "city": "成都", "level": "hot", "className": "footprint-city-point chengdu level-hot" },
            { "id": "guangzhou", "city": "广州", "level": "mid", "className": "footprint-city-point guangzhou level-mid" },
            { "id": "shenzhen", "city": "深圳", "level": "high", "className": "footprint-city-point shenzhen level-high" },
            { "id": "xian", "city": "西安", "level": "low", "className": "footprint-city-point xian level-low" },
            { "id": "hangzhou", "city": "杭州", "level": "high", "className": "footprint-city-point hangzhou level-high" }
          ],
          "friendUpdates": [
            { "id": "alex", "avatarText": "AL", "name": "Alex", "desc": "刚刚在成都宽窄巷子打卡", "online": true },
            { "id": "sarah", "avatarText": "SA", "name": "Sarah", "desc": "25分钟前在西安城墙打卡", "online": false }
          ],
          "hotCities": [
            { "id": "shanghai", "rank": 1, "city": "上海市中心", "desc": "2456人在这里打卡", "progress": 86, "level": "hot" },
            { "id": "chengdu", "rank": 2, "city": "成都市", "desc": "1892人在这里打卡", "progress": 72, "level": "warm" },
            { "id": "shenzhen", "rank": 3, "city": "深圳湾", "desc": "1567人在这里打卡", "progress": 58, "level": "active" }
          ]
        },
        "friend-city": {
          "title": "好友",
          "challengeTitle": "进行中的挑战",
          "rankingTitle": "城　市　榜",
          "nationalRankText": "查看全国榜",
          "challengeToast": "{title}进行中",
          "emptyChallengeText": "暂无挑战详情",
          "rankingToast": "第{rank}名城市榜",
          "emptyRankingText": "暂无城市榜详情",
          "nationalToast": "已展示当前城市榜",
          "duel": { "selfName": "我", "selfCount": "12区已点亮", "rivalName": "Sarah", "rivalCount": "10区已点亮", "vsText": "VS", "subtitle": "友谊赛", "startButtonText": "发起挑战", "recordButtonText": "查看记录" },
          "challenges": [
            { "id": "jingan-first", "title": "率先点亮静安区", "timeLeft": "2天", "statusText": "进行中", "statusTone": "pending", "selfValue": 3, "rivalValue": 2, "total": 5, "selfPercent": 60, "rivalPercent": 40 },
            { "id": "landmark-speed", "title": "10个地标速通", "timeLeft": "5天", "statusText": "领先中", "statusTone": "leading", "selfValue": 7, "rivalValue": 4, "total": 10, "selfPercent": 70, "rivalPercent": 40 }
          ],
          "rankings": [
            { "rank": 1, "name": "Mike", "desc": "已点亮 28 区", "score": "2,450", "tone": "gold" },
            { "rank": 2, "name": "Sarah", "desc": "已点亮 24 区", "score": "2,180", "tone": "silver" },
            { "rank": 3, "name": "David", "desc": "已点亮 22 区", "score": "1,950", "tone": "bronze" },
            { "rank": 4, "name": "我", "desc": "已点亮 12 区", "score": "1,240", "tone": "normal" }
          ]
        },
        "real-checkin": {
          "taskSectionTitle": "选择打卡任务",
          "generateButtonText": "生成足迹碎片",
          "texts": { "unsupportedCamera": "当前基础库不支持拍照", "photoSelected": "打卡照片已选择", "storySaved": "打卡文字已记录", "storyRequired": "请先填写打卡文字", "taskSelected": "{title}已选中", "taskRequired": "请选择打卡任务", "fragmentPending": "足迹碎片待生成" },
          "checkinDetail": { "distanceText": "距离目标 15米", "spotName": "外滩观景台", "statusTitle": "地点已解锁", "statusDesc": "完成打卡任务获得足迹值", "storyTitle": "留下你的故事", "storyPlaceholder": "用20个字记录此刻的心情...", "storyMinLength": 20, "storyMaxLength": 120, "rewards": ["+20 足迹值", "+1 成就点"] },
          "checkinTasks": [
            { "id": "photo", "title": "拍摄地标合影", "desc": "与标志性建筑合影", "scoreText": "+10分" },
            { "id": "angle", "title": "发现隐藏角度", "desc": "拍摄独特的视角", "scoreText": "+20分" }
          ]
        }
      }
    }'::jsonb,
    'active',
    now()
  )
on conflict (config_key) do update set
  config_value = excluded.config_value,
  status = excluded.status,
  updated_at = now();
