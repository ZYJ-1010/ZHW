const validInvites = {}

const mockRoleStatusPageConfig = {
  roleAliases: {
    player: 'player',
    expert: 'expert',
    master: 'expert',
    guide: 'guide',
    leader: 'guide',
    玩家: 'player',
    行家: 'expert',
    领路人: 'guide'
  },
  statusMap: {
    active: 'approved',
    enabled: 'approved',
    passed: 'approved',
    success: 'approved',
    waiting: 'pending',
    reviewing: 'pending',
    auditing: 'pending',
    pending_audit: 'pending',
    rejected_audit: 'rejected',
    reject: 'rejected',
    disabled: 'disabled',
    available: 'none',
    locked: 'none',
    unavailable: 'none'
  },
  roleMeta: {
    expert: {
      roleName: '行家',
      applyTitle: '行家申请',
      successAccent: 'cyan',
      approvedCopy: '你已获得行家身份，可在平台内使用对应能力',
      primaryText: '开启行家之旅'
    },
    guide: {
      roleName: '领路人',
      applyTitle: '领路人申请',
      successAccent: 'orange',
      approvedCopy: '你已获得领路人身份，可在平台内使用对应能力',
      primaryText: '开启领路人之旅'
    }
  },
  pendingTimeline: [
    { title: '提交申请', descTemplate: '已成功提交{roleName}申请资料', timeField: 'submittedAt', fallbackTime: '已提交', state: 'done' },
    { title: '资料初审', desc: '平台审核团队已接收并开始初审', timeWhenSubmitted: '已接收', fallbackTime: '待系统同步', state: 'done' },
    {
      title: '深度审核',
      descByRole: {
        guide: '正在评估你的组局记录、信用分及领路计划书',
        expert: '正在评估你的专业能力、资质材料及服务说明'
      },
      time: '进行中...',
      state: 'active'
    },
    { title: '结果通知', desc: '审核结果将通过消息推送通知你', time: '待完成', state: 'pending' }
  ],
  approvedActions: [
    { iconKey: 'network', text: '关系网开启', routeKey: 'relationNetwork' },
    { iconKey: 'invite', text: '邀请玩家', routeKey: 'gameInvite' },
    { iconKey: 'profile', text: '完善资料', routeKey: 'profileSystemProfileInfo' }
  ],
  texts: {
    loadingText: '加载中...',
    errorTitle: '审核状态加载失败',
    backHomeText: '返回首页',
    retryText: '重试',
    pendingPageTitle: '审核进度',
    resultPageTitle: '审核结果',
    pendingTitle: '审核中',
    approvedTitle: '恭喜审核通过！',
    rejectedTitle: '审核未通过',
    pendingSubtitleTemplate: '{roleName}申请正在审核',
    approvedSubtitleTemplate: '你已成为「{roleName}」',
    rejectedSubtitle: '查看原因并完善后可再次申请',
    pendingDesc: '平台正在评估你的申请资料，请耐心等待',
    rejectedDesc: '感谢你的申请，但本次审核未通过',
    approvedAuditDesc: '你的申请已通过平台审核',
    expectedLabel: '预计完成时间',
    expectedTemplate: '预计 {expectedReviewAt} 前完成审核，届时将通过站内消息通知你审核结果。',
    expectedFallback: '审核预计 1-3 个工作日，结果将通过站内消息通知你。',
    detailTitle: '申请详情',
    pendingHelper: '审核期间你可以继续使用玩家身份',
    certNoLabel: '认证编号',
    certTimePrefix: '认证时间: ',
    giftTitle: '新手礼包',
    reasonTitle: '驳回原因',
    suggestionTitle: '改进建议',
    reapplyTitle: '重新申请',
    reapplyDesc: '完善资料后可再次提交申请。建议根据驳回原因逐项改进，提高通过率。',
    recordTitle: '申请记录',
    pendingFooterHomeText: '返回玩家首页',
    pendingFooterBenefitsText: '查看权益对比',
    rejectedHelpText: '查看帮助',
    rejectedImproveText: '完善资料',
    routeMissingText: '请选择可用入口',
    fieldRoleLabel: '申请角色',
    fieldApplyTimeLabel: '申请时间',
    fieldApplicationNoLabel: '申请编号',
    fieldCurrentStatusLabel: '当前状态',
    fieldExpectedLabel: '预计完成',
    fieldRejectTimeLabel: '驳回时间',
    fieldReapplyLabel: '可重新申请',
    submittedFallback: '已提交',
    backendRecordFallback: '以后台记录为准',
    applicationNoFallback: '审核中生成',
    pendingStatusText: '深度审核中',
    statusFallback: '待确认',
    expectedDoneFallback: '预计 1-3 个工作日',
    reapplySuffix: ' 后',
    reapplyNotifyFallback: '请关注后台通知',
    loadFailedText: '审核状态加载失败'
  },
  defaultRejectReasons: [
    '申请资料暂未达到当前角色审核要求',
    '部分证明材料或计划说明仍需补充完善'
  ],
  suggestionTemplates: [
    '多参与平台组局活动，积累带队经验',
    '完善个人资料，提升信用评分',
    '{improvePlanText}，详细描述你的服务优势',
    '获得同伴推荐背书可提升审核通过率'
  ],
  improvePlanTextByRole: {
    guide: '重新撰写领路计划书',
    expert: '补充服务说明'
  },
  reapplyDays: 7
}

const mockRoleApplicationPageConfig = {
  pageTitle: '选择你的身份',
  pageDesc: '玩家为默认身份。行家和领路人需提交申请，审核通过后开放对应能力。',
  texts: {
    loadingText: '加载中...',
    loadFailedText: '角色申请加载失败',
    conditionLabel: '条件达成',
    paymentLabel: '付费状态',
    pendingButtonText: '已进入审核',
    submittingText: '提交中...',
    submitButtonText: '提交申请',
    submitFailedText: '提交失败',
    requirementPrefix: '• ',
    defaultSuccessText: '申请已提交'
  }
}

const mockGuideApplyConfig = {
  applyRoleType: 'guide',
  applyRoleName: '领路人',
  requirements: [
    { title: '完成实名认证', text: '领路人必须实名', done: false },
    { title: '信用分 ≥ 80 分', text: '以信用记录为准', done: false }
  ],
  planTask: { title: '提交领路计划书', text: '描述你的带队风格、战绩、资源和规划', done: false, action: '去填写 ›' },
  perks: [
    { icon: '¥', text: '有权益的领路人引荐玩家组局可获得相应收入' },
    { icon: '★', text: '专属领路人标识与优先推荐位' },
    { icon: 'D', text: '数据看板：查看邀约数据与关系网络' }
  ],
  fields: [
    { key: 'city', label: '所在城市', type: 'input', required: true, placeholder: '请输入常驻城市', maxlength: 20, helper: '用于匹配同城玩家与组局推荐' },
    {
      key: 'audience',
      label: '可推荐人群',
      type: 'chips',
      required: true,
      options: [
        { name: '朋友', active: true },
        { name: '同事', active: true },
        { name: '同城玩家', active: true },
        { name: '社群成员', active: false }
      ],
      helper: '可多选，后续将用于关系网推荐'
    },
    { key: 'contact', label: '常用联系方式', type: 'input', required: true, placeholder: '请输入微信号或手机号', maxlength: 30 },
    { key: 'guidePlan', label: '领路计划书', type: 'textarea', required: true, placeholder: '请描述你的带队风格、战绩、资源和规划', maxlength: 300, helper: '不少于 50 字，说明你能帮助玩家完成组局的方式' }
  ],
  uploadField: {
    label: '资质证明',
    required: true,
    icon: '📎',
    title: '点击上传作品集及凭证',
    helper: '支持 JPG、PNG、PDF，最多 5 张',
    acceptTypes: ['JPG', 'PNG', 'PDF'],
    maxCount: 5
  },
  serviceCount: 0,
  serviceBlocks: [],
  validationRules: {
    guidePlan: { minLength: 50, maxLength: 300 },
    serviceName: { minLength: 2, maxLength: 20 },
    money: { integerMaxLength: 8, decimalMaxLength: 2 }
  },
  priceHint: '',
  primaryText: '提交领路人申请',
  helperText: '审核预计 1-3 个工作日'
}

const mockRoleBenefitConfig = {
  version: '2026-07-01',
  permissionPrompts: {
    expert: {
      roleType: 'expert',
      title: '我懂玩家需要什么！我申请成为行家',
      primary: '申请成为行家',
      secondary: '查看权益对比'
    },
    guide: {
      roleType: 'guide',
      title: '我愿意带领更多人一起玩！我申请成为领路人',
      primary: '申请成为领路人',
      secondary: '查看权益对比'
    }
  },
  roleComparison: {
    name: '权益对比页',
    mode: 'roleComparison',
    roleBadge: '权益',
    title: '角色权益对比',
    subtitle: '选择适合你的角色，开启不同玩法',
    roles: [
      { key: 'player', name: '玩家', level: '等级由后台配置', active: true },
      { key: 'guide', name: '领路人', level: '等级由后台配置', active: false },
      { key: 'expert', name: '行家', level: '等级由后台配置', active: false }
    ],
    benefits: [
      { name: '发起组局', player: '✓', guide: '—', expert: '✓' },
      { name: '加入组局', player: '✓', guide: '✓', expert: '✓' },
      { name: '发起组局', player: '✓', guide: '✓', expert: '✓' },
      { name: '数据看板', player: '—', guide: '✓', expert: '✓' },
      { name: '信用背书', player: '—', guide: '✓', expert: '✓' }
    ],
    primary: '立即申请角色'
  }
}

const mockReferralRecordsConfig = {
  pageTitle: '我的引荐记录',
  summary: {
    label: '本月引荐数据',
    background: 'linear-gradient(135deg, #ffb347 0%, #ff7b00 100%)',
    iconSrc: 'https://static.haowan.net.cn/miniprogram/pages/game/referral-record/assets/wallet.png',
    statTemplates: {
      success: '成功 {count}单',
      processing: '进行中 {count}单',
      review: '待评价 {count}单'
    }
  },
  tabs: [
    { key: 'processing', label: '进行中' },
    { key: 'completed', label: '已完成' },
    { key: 'canceled', label: '已取消' }
  ],
  texts: {
    emptyText: '暂无引荐记录',
    loadFailedText: '引荐记录加载失败',
    expertRoleText: '行家',
    playerRoleText: '玩家',
    selfLabel: '我',
    reviewedTagText: '已评价',
    pendingReviewTagText: '待评价',
    reviewedActionText: '已评价',
    reviewActionText: '评价双方',
    remindActionText: '提醒交付',
    chatActionText: '查看群聊',
    chatPrefill: '你好，我想查看本次引荐服务的群聊进度。',
    unavailableText: '该操作暂不可用',
    remindMessageTemplate: '请及时确认交付：{serviceTitle}',
    remindServiceFallback: '引荐服务',
    remindSuccessText: '已提醒交付',
    remindFailedText: '提醒交付失败',
    rewardPrefix: '¥'
  }
}

const mockUser = {
  id: '10001',
  nickname: '微信用户',
  avatarUrl: '',
  authStatus: 'pending',
  roles: ['player'],
  roleStatusMap: {
    player: 'approved'
  },
  defaultRole: 'player',
  default_role: 'player',
  needInvite: false,
  needRealname: true,
  creditScore: 100,
  todayCreditScore: 100
}

const mockCurrentUser = {
  id: mockUser.id,
  nickname: mockUser.nickname,
  avatarUrl: mockUser.avatarUrl,
  authStatus: mockUser.authStatus,
  realnameStatus: 'pending',
  roles: mockUser.roles,
  roleStatusMap: mockUser.roleStatusMap,
  defaultRole: mockUser.defaultRole,
  default_role: mockUser.default_role,
  member: {
    planCode: 'basic',
    planName: '基础会员',
    validUntil: '2026-12-31'
  },
  inviter: {
    nickname: '周领路人',
    city: '上海'
  },
  creditScore: 100,
  todayCreditScore: 100,
  points: 260,
  level: 2,
  experience: 680,
  todoCounts: {
    realname: 1,
    roleApplications: 0,
    gameApplications: 2,
    availableReviews: 1,
    messages: 3
  }
}

const mockHome = {
  user: mockCurrentUser,
  hero: {
    roleName: '玩家',
    dateLabel: '2026.03.30',
    subtitle: '开启你的今日副本',
    onlineText: '在线'
  },
  notices: [
    {
      id: 'notice-1',
      title: '完成实名后可发起和参与核心组局',
      type: 'realname'
    }
  ],
  quickActions: [
    { id: 'create', title: '发起组局', route: 'pages/game/create/index' },
    { id: 'hall', title: '局前大厅', route: 'pages/game/hall/index' },
    { id: 'role', title: '角色申请', route: 'pages/role/apply/index' },
    { id: 'map', title: '附近组局', route: 'pages/map/index' },
    { id: 'message', title: '消息', route: 'pages/message/index' },
    { id: 'profile', title: '我的', route: 'pages/profile/index' }
  ],
  nearbySection: {
    title: '附近正在发生',
    tabs: [
      { key: 'all', name: '全部' },
      { key: 'nearby', name: '附近' }
    ]
  },
  recommendedGames: [
    {
      id: '20001',
      title: 'AI赋能系统搭建交流局',
      cityName: '黄浦区',
      distanceText: '8.2km',
      memberText: '3/8人',
      timeText: '2026年5月1日 14:00--16:00',
      statusText: '任务局',
      priceText: '¥0/人',
      joinedText: '+3位玩家已入局',
      actionText: '加入',
      coverUrl: 'https://static.haowan.net.cn/miniprogram/assets/game-card/cover-sunset.png',
      scope: 'city',
      actions: ['share', 'follow', 'refer', 'greet'],
      route: 'pages/game/detail/index?id=20001',
      tags: ['AI', '系统搭建', '交流']
    },
    {
      id: '20002',
      title: '苏州河“记忆碎片”采集',
      cityName: '静安区',
      distanceText: '3.2km',
      memberText: '5/8人',
      timeText: '2026年5月1日 20:00--22:00',
      statusText: '探索局',
      priceText: '¥0/人',
      joinedText: '+5位玩家已入局',
      actionText: '加入',
      coverUrl: 'https://static.haowan.net.cn/miniprogram/components/game-card/assets/cover-city.png',
      scope: 'nearby',
      actions: ['share', 'follow', 'refer', 'greet'],
      route: 'pages/game/detail/index?id=20002',
      tags: ['城市故事', '探索', '同城']
    }
  ],
  playerSummary: {
    currentRole: 'player',
    roleType: 'player',
    displayName: 'Alex Chen',
    nickname: 'Alex Chen',
    roleLabel: '玩家 Lv.5',
    title: '活跃达人',
    level: 5,
    nextLevel: 6,
    experience: 580,
    nextLevelExperience: 1000,
    expToNextLevel: 420,
    xpText: '580/1000 XP',
    progressPercent: 58,
    nextLevelText: '距离下一等级还需 420 经验值',
    joinCount: 12,
    monthlyMvpCount: 3,
    participationRate: '98%',
    stats: [
      { label: '参与局数', value: '12' },
      { label: '本月MVP', value: '3' },
      { label: '参与率', value: '98%' }
    ]
  },
  rankingSection: {
    icon: '🏆',
    title: '本周玩霸榜',
    moreText: '查看全部榜单',
    defaultTab: 'player',
    tabs: [
      { key: 'player', name: '玩家' },
      { key: 'expert', name: '行家' },
      { key: 'guide', name: '领路人' }
    ]
  },
  rankingBoards: {
    player: {
      list: [
        {
          id: 'rank-player-01',
          rank: 1,
          nickname: '领域专家 PRO',
          avatarFallback: '👨🏾‍🎓',
          desc: '本周组局 12 · MVP 5次',
          xpText: '2,450 XP'
        },
        {
          id: 'rank-player-02',
          rank: 2,
          nickname: '社交达人',
          avatarFallback: '👩🏻‍🎤',
          desc: '本周组局 8 次',
          xpText: '1,890 XP'
        },
        {
          id: 'rank-player-03',
          rank: 3,
          nickname: '探险家',
          avatarFallback: '👨🏿‍🚀',
          desc: '本周组局 6 次',
          xpText: '1,560 XP'
        }
      ],
      myRank: {
        rank: 52,
        nickname: '我（Alex）',
        avatarFallback: '👩🏻‍💻',
        desc: '上周排名 65 ↑',
        xpText: '520 XP'
      }
    }
  },
  rankingList: [
    { rank: '01', name: '领域专家 PRO', avatarFallback: '👨🏾‍🎓', desc: '本周组局 12 · MVP 5次', xpText: '2,450 XP' },
    { rank: '02', name: '社交达人', avatarFallback: '👩🏻‍🎤', desc: '本周组局 8 次', xpText: '1,890 XP' },
    { rank: '03', name: '探险家', avatarFallback: '👨🏿‍🚀', desc: '本周组局 6 次', xpText: '1,560 XP' }
  ],
  achievementSection: {
    icon: '💎',
    title: '我的成就'
  },
  achievementList: [
    { id: 'hundred', code: 'hundred_king', title: '百场王者', icon: '🏆', statusText: '等级', unlocked: true },
    { id: 'pilot', code: 'pilot_king', title: '引航王者', icon: '🏆', statusText: '等级', unlocked: true },
    { id: 'earth', code: 'earth_roamer', title: '地球漫游者', icon: '🌍', statusText: '进度20%', progressPercent: 20, unlocked: true },
    { id: 'hidden', code: 'hidden_badge', title: '隐藏徽章', icon: '🔒', statusText: '未解锁', unlocked: false }
  ],
  friendSection: {
    icon: '🎲',
    title: '朋友在玩',
    count: 2,
    moreText: '查看全部'
  },
  friendGames: [
    {
      id: 'friend-1',
      title: '盲盒路线：3小时点亮天际线',
      cityName: '梧桐山',
      distanceText: '1.5km',
      memberText: '3/8人',
      timeText: '2026年5月1日 14:00--16:00',
      priceText: '¥29/人',
      statusText: '探索局',
      joinedText: '+3位玩家已入局',
      coverUrl: 'https://static.haowan.net.cn/miniprogram/assets/game-card/cover-sunset.png',
      actionText: '加入',
      actions: ['share', 'follow', 'refer', 'greet'],
      route: 'pages/game/detail/index?id=friend-1'
    }
  ],
  metaverseEntry: {
    title: '进入元宇宙',
    desc: '共创数字街区｜全球联机互动',
    tags: ['3D空间', 'NFT徽章'],
    avatars: [
      { avatarFallback: '👨🏾‍🎓' },
      { avatarFallback: '👩🏻‍🎤' },
      { avatarFallback: '👨🏿‍🚀' }
    ],
    joinedCount: 99,
    actionText: '进入元宇宙',
    route: 'pages/placeholder/metaverse/index'
  },
  earth: {
    onlineCount: 5,
    nearbyCount: 2,
    nodes: [
      { id: 'user-10001', type: 'self', label: '我', weight: 5 },
      { id: 'game-20001', type: 'game', label: 'AI交流局', longitude: 121.4737, latitude: 31.2304, weight: 3 },
      { id: 'game-20002', type: 'game', label: '城市采集局', longitude: 121.4837, latitude: 31.2204, weight: 5 },
      { id: 'user-20002', type: 'relation', label: '好友A', strength: 3 }
    ],
    heatPoints: [
      { cityCode: '310100', cityName: '上海', longitude: 121.4737, latitude: 31.2304, weight: 3 },
      { cityCode: '320500', cityName: '苏州', longitude: 120.5853, latitude: 31.2989, weight: 2 }
    ],
    edges: [
      { source: 'user-10001', target: 'user-20002', relationType: 'invite' }
    ]
  },
  visualization: {
    network: {
      nodes: [
        { id: 'user-10001', label: '我', type: 'self' },
        { id: 'user-20002', label: '好友A', type: 'relation' }
      ],
      edges: [
        { source: 'user-10001', target: 'user-20002', relationType: 'invite' }
      ]
    }
  },
  nearbySummary: {
    count: 12,
    nearbyGameCount: 12,
    checkedInCount: 8,
    cityName: '上海',
    accuracyText: '定位精度 300m 内'
  }
}

const mockRoleHomes = {
  expert: {
    user: mockCurrentUser,
    hero: {
      roleName: '行家',
      dateLabel: '2026.05.15',
      subtitle: '开启你的今日副本',
      onlineText: '在线'
    },
    roleDashboard: {
      roleName: '行家',
      roleEmoji: '🎯',
      deviceBadge: '💎',
      profileName: '摄影咖 小李',
      identity: '认证DM',
      levelText: '行家 ⭐️',
      scoreText: '650/1000 XP',
      nextLevelText: '距离下一等级还需 350 经验值',
      primaryTitle: '我的开局',
      primaryDesc: '管理进行中的局',
      sectionTitle: '即将带局',
      sectionMore: '查看全部 →',
      sectionCount: '2',
      stats: [
        { value: '156', label: '服务玩家' },
        { value: '99%', label: '复购率' },
        { value: '98%', label: '被选率' },
        { value: '3', label: '本月MVP' }
      ],
      roleTabs: [
        { label: '🎮 玩家', active: false },
        { label: '🎯 行家', active: true },
        { label: '🌐 领路人', active: false }
      ],
      quickActions: [
        { id: 'create-game', icon: '📍', title: '发起组局', desc: '创建新的一局', tone: 'cyan', route: 'pages/game/create/index' },
        { id: 'lobby', icon: '🎲', title: '我的开局', desc: '管理进行中的局', tone: 'blue', route: 'pages/game/hall/index' }
      ],
      onlineCard: {
        title: '地球online',
        desc: '管理我的组局足迹与城市打卡',
        ownedGameCount: 12,
        checkinCount: 8,
        tags: ['我的组局 12 个', '已打卡 8 处']
      },
      events: [
        {
          id: 'expert-game-1',
          time: '14:00',
          day: '今天',
          title: '《血染钟楼》8人局',
          meta: '星巴克(万科店) · 1.2km · 5/8人',
          dateText: '2026年5月1日 14:00--16:00',
          tags: ['专业场', '探索局'],
          status: '即将满员',
          statusTone: 'green',
          income: '¥200',
          joinedText: '+3位玩家已入局',
          participantAvatars: [
            'https://static.haowan.net.cn/miniprogram/pages/home/player/assets/ranking-avatar-01.png',
            'https://static.haowan.net.cn/miniprogram/pages/home/player/assets/ranking-avatar-02.png',
            'https://static.haowan.net.cn/miniprogram/pages/home/player/assets/ranking-avatar-03.png'
          ]
        },
        {
          id: 'expert-game-2',
          time: '19:30',
          day: '今天',
          title: '苏州河“记忆碎片”采集',
          meta: '黄浦区 · 8.2km · 3/8人',
          dateText: '2026年5月1日 14:00--16:00',
          tags: ['专业场', '任务局'],
          status: '新手友好',
          statusTone: 'lime',
          income: '¥240',
          joinedText: '+3位玩家已入局',
          participantAvatars: [
            'https://static.haowan.net.cn/miniprogram/pages/home/player/assets/ranking-avatar-01.png',
            'https://static.haowan.net.cn/miniprogram/pages/home/player/assets/ranking-avatar-02.png',
            'https://static.haowan.net.cn/miniprogram/pages/home/player/assets/ranking-avatar-03.png'
          ]
        }
      ],
      skills: [
        { id: 'dm-basic', icon: '🎲', title: 'DM入门', stateText: '已解锁', locked: false },
        { id: 'story', icon: '🎭', title: '沉浸演绎', stateText: '待解锁', locked: true },
        { id: 'control', icon: '🔒', title: '控场大师', stateText: '待解锁', locked: true }
      ],
      review: {
        title: '最新评价',
        count: 156,
        playerLevel: '萌新玩家',
        avatarText: '🎮',
        timeText: '2小时前',
        rating: 4,
        content: 'DM非常专业，带本节奏很好，气氛拉满！第一次玩血染就上瘾了，下次还找小王带局。💯'
      }
    },
    rankingSection: {
      ...mockHome.rankingSection,
      defaultTab: 'expert'
    },
    rankingBoards: {
      ...mockHome.rankingBoards,
      expert: {
        list: [
          { id: 'rank-expert-01', rank: 1, nickname: '领域专家 PRO', avatarFallback: '👨🏾‍🎓', desc: '本周服务玩家 90 位', xpText: '2,450 XP' },
          { id: 'rank-expert-02', rank: 2, nickname: '社交达人', avatarFallback: '👩🏻‍🎤', desc: '本周服务玩家 10 位', xpText: '1,890 XP' },
          { id: 'rank-expert-03', rank: 3, nickname: '探险家', avatarFallback: '👨🏿‍🚀', desc: '本周服务玩家 1 位', xpText: '1,560 XP' }
        ],
        myRank: {
          rank: 52,
          nickname: '我（Alex）',
          avatarFallback: '👩🏻‍💻',
          desc: '上周排名 65 ↑',
          xpText: '520 XP'
        }
      }
    },
    achievementSection: mockHome.achievementSection,
    achievementList: mockHome.achievementList,
    metaverseEntry: mockHome.metaverseEntry
  },
  guide: {
    user: mockCurrentUser,
    hero: {
      roleName: '领路人',
      dateLabel: '2026.05.15',
      subtitle: '开启你的今日副本',
      onlineText: '在线'
    },
    roleDashboard: {
      roleName: '领路人',
      roleEmoji: '🌐',
      profileName: '摄影咖 萧飒',
      identity: '百场辅助',
      levelText: '领路人 🐑',
      scoreText: '580/1000 XP',
      experience: 580,
      nextLevelExperience: 1000,
      progress: 58,
      nextLevelText: '距离下一等级还需 420 经验值',
      primaryTitle: '我的邀请',
      primaryDesc: '管理连接的玩家',
      sectionTitle: '附近正在发生',
      sectionDesc: '',
      sectionMore: '查看全部',
      stats: [
        { value: '99', label: '连接玩家' },
        { value: '99%', label: '玩家再玩率' },
        { value: '98%', label: '玩家完局率' },
        { value: '3', label: '本月MVP' }
      ],
      roleTabs: [
        { label: '🎮 玩家', active: false },
        { label: '🎯 行家', active: false },
        { label: '🌐 领路人', active: true }
      ],
      quickActions: [
        { id: 'lobby', icon: '📍', title: '局前大厅', desc: '准备加入一局', tone: 'pink', routeIcon: true, route: 'pages/game/hall/index' },
        { id: 'invite', icon: '📍', title: '我的邀约', desc: '管理连接的玩家', tone: 'cyan', routeIcon: true, route: 'pages/game/applications/index' }
      ],
      filterTabs: [
        { label: '全部', active: true },
        { label: '附近', active: false }
      ],
      events: [
        {
          id: 'guide-game-1',
          title: '苏州河“记忆碎片”采集',
          cityName: '静安区',
          distanceText: '3.2km',
          memberText: '5/8人',
          dateText: '2026年5月1日 20:00--22:00',
          statusText: '探索局',
          priceText: '¥0/人',
          actionText: '加入',
          joinedText: '+5位玩家已入局',
          coverUrl: 'https://static.haowan.net.cn/miniprogram/components/game-card/assets/cover-city.png',
          scope: 'nearby',
          actions: ['share', 'follow', 'refer', 'greet'],
          route: 'pages/game/detail/index?id=guide-game-1',
          tags: ['城市故事', '探索', '同城']
        },
        {
          id: 'guide-game-2',
          title: 'AI赋能系统搭建交流局',
          cityName: '黄浦区',
          distanceText: '8.2km',
          memberText: '3/8人',
          dateText: '2026年5月1日 14:00--16:00',
          statusText: '任务局',
          priceText: '¥0/人',
          actionText: '加入',
          joinedText: '+3位玩家已入局',
          coverUrl: 'https://static.haowan.net.cn/miniprogram/assets/game-card/cover-sunset.png',
          scope: 'city',
          actions: ['share', 'follow', 'refer', 'greet'],
          route: 'pages/game/detail/index?id=guide-game-2',
          tags: ['AI', '系统搭建', '交流']
        }
      ],
      recommendation: {
        title: '推荐行家',
        tabs: [
          { label: '行家', active: true },
          { label: '玩家', active: false }
        ],
        items: [
          { id: 'guide-reco-1', icon: '👑', name: '剧本杀小王', desc: '已完成 3 局', tag: '查看关系' },
          { id: 'guide-reco-2', icon: '🎓', name: '大学生DM', desc: '宁大节点', tag: '管理我的连接' },
          { id: 'guide-reco-3', icon: '🔬', name: '研究员阿伟', desc: '中科院', tag: '周活跃' }
        ]
      },
      network: {
        title: '我的关系网络',
        status: '实时连接中',
        hubTitle: '萧飒',
        hubDesc: '领路人',
        connectedCount: 156,
        summary: '● 已连接 156 位玩家',
        actionText: '查看全部',
        income: '本周新增连接 12 位',
        location: '📍 镇海区',
        items: [
          { id: 'script-master', icon: '👑', name: '剧本杀小王', desc: '已完成 3 局' },
          { id: 'student-dm', icon: '🎓', name: '大学生DM', desc: '宁大节点' },
          { id: 'mama-group', icon: '👶', name: '宝妈组局', desc: '周活跃' },
          { id: 'researcher', icon: '🔬', name: '研究员阿伟', desc: '中科院' },
          { id: 'all', icon: '+', name: '查看全部', desc: '156人', dashed: true }
        ],
        buttons: [
          { text: '管理我的连接', primary: true, route: 'pages/message/index' },
          { text: '查看邀请记录', route: 'pages/profile/service-center/invite/records/index' }
        ]
      }
    },
    rankingSection: {
      ...mockHome.rankingSection,
      defaultTab: 'guide'
    },
    rankingBoards: {
      ...mockHome.rankingBoards,
      guide: {
        list: [
          { id: 'rank-guide-01', rank: 1, nickname: '城市连接官', avatarFallback: '👨🏾‍🎓', desc: '本周连接玩家 156 位', xpText: '2,450 XP' },
          { id: 'rank-guide-02', rank: 2, nickname: '社群引路人', avatarFallback: '👩🏻‍🎤', desc: '本周成功引荐 32 次', xpText: '1,890 XP' },
          { id: 'rank-guide-03', rank: 3, nickname: '活动发现家', avatarFallback: '👨🏿‍🚀', desc: '本周活跃节点 18 个', xpText: '1,560 XP' }
        ],
        myRank: {
          rank: 52,
          nickname: '我（萧飒）',
          avatarFallback: '👩🏻‍💻',
          desc: '上周排名 65 ↑',
          xpText: '520 XP'
        }
      }
    },
    achievementSection: mockHome.achievementSection,
    achievementList: mockHome.achievementList,
    metaverseEntry: mockHome.metaverseEntry
  }
}

const mockProfileHome = {
  user: {
    nickname: '小明',
    memberLevel: '',
    memberStatus: 'none',
    roleLevel: 'V5 探险家',
    growthLevel: 'V5 探险家',
    role: '玩家',
    avatarText: '小',
    inviteCode: 'ZHW-2026'
  },
  stats: [
    { key: 'referrals', label: '引荐数', value: '128' },
    { key: 'successes', label: '成功数', value: '86' },
    { key: 'completedGames', label: '完成局数', value: '42' },
    { key: 'credit', label: '信用分', value: '98' }
  ],
  assets: [
    { key: 'availablePoints', label: '可用积分', value: '260' },
    { key: 'experience', label: '累计经验', value: '1,280' },
    { key: 'credit', label: '信用分', value: '98', tone: 'green' }
  ],
  vipBanner: {
    text: '升级会员，认证您的角色',
    actionText: '增购会员 >',
    route: '/pages/profile/member/index'
  },
  serviceSections: [
    {
      title: '服务中心',
      items: [
        { title: '我的局', iconSrc: '/pages/profile/assets/i66@3x.png', iconClass: 'purple-blue', badge: '2进行中', badgeClass: 'pink', route: '/pages/profile/service-center/my-games/index' },
        { title: '我的邀请', iconSrc: '/pages/profile/assets/i68@3x.png', iconClass: 'purple-blue', badge: '3个关系', badgeClass: 'orange', route: '/pages/profile/service-center/invite/overview/index' },
        { title: '评价管理', iconSrc: '/pages/profile/assets/i69@3x.png', iconClass: 'purple-blue', badge: '1待评价', badgeClass: 'pink', route: '/pages/profile/service-center/manage/review-manage/index' }
      ]
    },
    {
      title: '资产中心',
      items: [
        { title: '我的资产', iconSrc: '/pages/profile/assets/i70@3x.png', iconClass: 'orange', route: '/pages/profile/asset-center/manage/index' },
        { title: '我的押金', iconSrc: '/pages/profile/assets/i71@3x.png', iconClass: 'orange' },
        { title: '积分商城', iconSrc: '/pages/profile/assets/i72@3x.png', iconClass: 'orange', route: '/pages/profile/asset-center/mall/index' },
        { title: '我的积分', iconSrc: '/pages/profile/assets/i73@3x.png', iconClass: 'orange', route: '/pages/profile/asset-center/points/index' },
        { title: '开票中心', iconSrc: '/pages/profile/assets/i74@3x.png', iconClass: 'orange' },
        { title: '任务中心', iconSrc: '/pages/profile/assets/i73@3x.png', iconClass: 'orange', route: '/pages/profile/task-center/index' }
      ]
    }
  ],
  incomeSummary: {
    pendingAmountText: '¥0.00',
    settledAmountText: '¥0.00'
  }
}

const mockProfileAssets = {
  overview: {
    label: '积分资产',
    value: '260 积分',
    points: 260,
    updatedText: '实时同步积分、订单与评价记录'
  },
  assetStats: [
    { key: 'availablePoints', label: '可用积分', value: '260', tone: 'green' },
    { key: 'orderCount', label: '兑换订单', value: '2', tone: 'yellow' },
    { key: 'reviewTodo', label: '待评价', value: '1', tone: 'orange' }
  ],
  quickActions: [],
  menuItems: [
    { key: 'points', title: '积分明细', desc: '查看积分获取和使用记录', iconSrc: '/pages/profile/asset-center/manage/assets/fa/list-ul.svg', tone: 'blue', route: '/pages/profile/asset-center/points/index' },
    { key: 'mall', title: '积分商城', desc: '使用积分兑换权益', iconSrc: '/pages/profile/asset-center/manage/assets/fa/bag-shopping.svg', tone: 'green', route: '/pages/profile/asset-center/mall/index' },
    { key: 'orders', title: '我的订单', desc: '查看全部订单', iconSrc: '/pages/profile/asset-center/manage/assets/fa/bag-shopping.svg', tone: 'purple', route: '/pages/profile/asset-center/orders/index' }
  ],
  orderStatuses: [
    { key: 'pending', label: '待处理', count: 0, countText: '0', iconSrc: '/pages/profile/asset-center/manage/assets/fa/hourglass-half.svg', tone: 'blue', route: '/pages/profile/asset-center/orders/index?status=pending' },
    { key: 'processing', label: '进行中', count: 1, countText: '1', iconSrc: '/pages/profile/asset-center/manage/assets/fa/spinner.svg', tone: 'orange', route: '/pages/profile/asset-center/orders/index?status=pending' },
    { key: 'completed', label: '已完成', count: 1, countText: '1', iconSrc: '/pages/profile/asset-center/manage/assets/fa/check.svg', tone: 'green', route: '/pages/profile/asset-center/orders/index?status=fulfilled' },
    { key: 'refund', label: '已取消', count: 0, countText: '0', iconSrc: '/pages/profile/asset-center/manage/assets/fa/rotate-left.svg', tone: 'red', route: '/pages/profile/asset-center/orders/index?status=canceled' },
    { key: 'review', label: '待评价', count: 1, countText: '1', iconSrc: '/pages/profile/asset-center/manage/assets/fa/star.svg', tone: 'gray', route: '/pages/profile/service-center/manage/review-manage/index' }
  ],
  recentOrders: [
    { id: 'RDM-1', orderId: 1, title: '活动纪念徽章', status: '进行中', statusTone: 'blue', time: '2026-03-20 14:30:00', amount: '800积分', route: '/pages/profile/asset-center/orders/index' },
    { id: 'RDM-2', orderId: 2, title: '平台周边兑换', status: '已完成', statusTone: 'green', time: '2026-03-15 09:15:00', amount: '600积分', route: '/pages/profile/asset-center/orders/index' }
  ],
  balanceRecords: [],
  bankCards: { count: 0, summaryText: '一期不提供银行卡功能', items: [], canBind: false, needIdentity: false },
  faqLinks: [
    { key: 'pointsUse', label: '积分有什么用？', answer: '积分可用于积分商城兑换；具体商品以商城展示为准。' },
    { key: 'pointsRecord', label: '如何查看积分记录？', answer: '可在积分明细中查看积分获取和使用记录。' }
  ]
}

const mockSystemSkillConfig = {
  roleSummary: {
    roleName: '??',
    maxSkillCount: 3,
    monthlyLimit: 3,
    usedCount: 0,
    remainingCount: 3,
    configuredCount: 0
  },
  skillSlots: [
    { id: 'empty-1', title: '????', iconKey: 'plus', iconText: '+', tone: 'gray', active: false, empty: true },
    { id: 'empty-2', title: '????', iconKey: 'plus', iconText: '+', tone: 'gray', active: false, empty: true },
    { id: 'empty-3', title: '????', iconKey: 'plus', iconText: '+', tone: 'gray', active: false, empty: true, locked: true }
  ],
  tabs: [
    { key: 'visible', label: '????' },
    { key: 'hidden', label: '????' },
    { key: 'cases', label: '????' }
  ],
  sectionMap: {
    visible: {
      title: '???????',
      desc: '???????????'
    },
    hidden: {
      title: '????????',
      desc: '???????????????????????'
    },
    cases: {
      title: '??????',
      desc: '???????????????'
    }
  },
  skillGroups: {
    visible: [],
    hidden: [],
    cases: []
  },
  addableSkills: [
    { id: 'review', title: '?????', desc: '????????', iconText: '??', tone: 'orange' },
    { id: 'mood', title: '?????', desc: '?????????', iconText: '??', tone: 'pink' },
    { id: 'rules', title: '?????', desc: '????????', iconText: '??', tone: 'green' },
    { id: 'photo', title: '?????', desc: '????????', iconText: '??', tone: 'cyan' }
  ],
  unlockSuggestion: {
    title: '???????',
    desc: '?????????????????',
    actionText: '????'
  }
}
const mockPointsMall = {
  pointsAvailable: 2580,
  expireTip: '积分有效期12个月，请及时兑换',
  goods: [
    {
      id: 'mall-shirt',
      iconText: '👕',
      title: '平台限定T恤',
      cost: 500,
      stockLeft: 23
    },
    {
      id: 'mall-badge',
      iconText: '🏅',
      title: '真好玩徽章套装',
      cost: 300,
      stockLeft: 56
    },
    {
      id: 'mall-backpack',
      iconText: '🎒',
      title: '探险家背包',
      cost: 800,
      stockLeft: 12
    },
    {
      id: 'mall-camping',
      iconText: '⛺',
      title: '露营装备套装',
      cost: 1200,
      stockLeft: 8
    },
    {
      id: 'mall-card',
      iconText: '👑',
      title: '玩家桌游卡牌',
      cost: 200,
      stockLeft: 100
    },
    {
      id: 'mall-cup',
      iconText: '🥤',
      title: '定制水杯',
      cost: 350,
      stockLeft: 45
    }
  ]
}

const mockPointsPageConfig = {
  stats: [
    { key: 'total', label: '累计积分' },
    { key: 'redeemed', label: '已兑换' },
    { key: 'expired', label: '过期积分' }
  ],
  rules: [
    {
      text: '完成组局、提交评价、举报核实和平台活动等行为可产生积分，具体比例以后台配置为准',
      strong: '后台规则',
      suffix: ''
    },
    {
      text: '积分有效期按平台规则执行，到期后由后台任务处理',
      strong: '有效期规则',
      suffix: ''
    },
    {
      text: '积分仅可兑换',
      strong: '平台限定商品',
      suffix: '，不可提现'
    }
  ],
  filters: [
    { key: 'all', label: '全部', tone: 'all' },
    { key: 'income', label: '收入', tone: 'income' },
    { key: 'expense', label: '支出', tone: 'expense' }
  ],
  version: '2026-07-01'
}

const mockPointsOrders = {
  pageConfig: {
    emptyText: '\u6682\u65e0\u5151\u6362\u8ba2\u5355',
    logisticsEmptyText: '\u6682\u65e0\u7269\u6d41\u4fe1\u606f',
    detailEmptyText: '\u6682\u65e0\u8ba2\u5355\u8be6\u60c5',
    cancelConfirm: {
      title: '\u53d6\u6d88\u8ba2\u5355',
      content: '\u53d6\u6d88\u540e\u79ef\u5206\u5c06\u9000\u56de\u5230\u8d26\u6237\uff0c\u786e\u8ba4\u53d6\u6d88\u8fd9\u4e2a\u5151\u6362\u8ba2\u5355\u5417\uff1f',
      confirmText: '\u786e\u8ba4\u53d6\u6d88',
      cancelText: '\u518d\u60f3\u60f3',
      reason: '\u7528\u6237\u4e3b\u52a8\u53d6\u6d88'
    },
    actions: {
      detail: '\u67e5\u770b\u8be6\u60c5',
      cancel: '\u53d6\u6d88\u8ba2\u5355',
      logistics: '\u67e5\u770b\u7269\u6d41',
      again: '\u518d\u6b21\u5151\u6362'
    },
    version: '2026-07-01'
  },
  tabs: [
    { key: 'all', label: '全部' },
    { key: 'pending_ship', label: '待发货' },
    { key: 'shipping', label: '配送中' },
    { key: 'completed', label: '已完成' }
  ],
  orders: [
    {
      id: '20260615001',
      statusKey: 'shipping',
      statusText: '配送中',
      statusTone: 'blue',
      iconText: '👕',
      title: '平台限定T恤',
      pointsText: '500积分',
      exchangedAtText: '兑换时间: 2026-06-10 14:22',
      actions: [
        { key: 'logistics', label: '查看物流', type: 'ghost' }
      ]
    },
    {
      id: '20260528001',
      statusKey: 'completed',
      statusText: '已完成',
      statusTone: 'success',
      iconText: '🏅',
      title: '真好玩徽章套装',
      pointsText: '300积分',
      exchangedAtText: '兑换时间: 2026-05-20 09:15',
      actions: [
        { key: 'detail', label: '查看详情', type: 'ghost' },
        { key: 'again', label: '再次兑换', type: 'primary' }
      ]
    },
    {
      id: '20260614001',
      statusKey: 'pending_ship',
      statusText: '待发货',
      statusTone: 'orange',
      iconText: '🎒',
      title: '探险家背包',
      pointsText: '800积分',
      exchangedAtText: '兑换时间: 2026-06-14 16:30',
      actions: [
        { key: 'cancel', label: '取消订单', type: 'ghost' }
      ]
    }
  ]
}

const mockPointsOrderLogistics = {
  '20260615001': {
    orderId: '20260615001',
    courier: {
      name: '顺丰速运',
      trackingNo: 'SF1234567890',
      logoText: 'SF'
    },
    timeline: [
      {
        id: 'arrived-site',
        desc: '【深圳市】快件已到达 深圳南山营业点',
        time: '2026-06-15 08:30',
        active: true
      },
      {
        id: 'left-transfer',
        desc: '【深圳市】快件离开 深圳转运中心，已发往 南山营业点',
        time: '2026-06-15 06:15'
      },
      {
        id: 'arrived-transfer',
        desc: '【深圳市】快件已到达 深圳转运中心',
        time: '2026-06-14 23:40'
      },
      {
        id: 'shipped',
        desc: '【广州市】商家已发货，等待揽收',
        time: '2026-06-14 18:00'
      }
    ]
  }
}

const mockNewbieTasks = [
  {
    id: 'newbie-realname',
    type: 'realname',
    title: '完成实名认证',
    rewardText: '+50 经验值',
    completed: false,
    route: 'pages/login/index?ui=1&mode=realnameGuide',
    actionText: '去完成'
  },
  {
    id: 'newbie-profile',
    type: 'profile',
    title: '完善个人资料',
    rewardText: '+30 经验值',
    completed: false,
    route: 'pages/profile/index',
    actionText: '去完成'
  },
  {
    id: 'newbie-first-game',
    type: 'first_game',
    title: '发布第一个局',
    rewardText: '+100 经验值',
    completed: false,
    route: 'pages/game/create/index',
    actionText: '去完成'
  }
]

const mockRoleApplications = [
  {
    roleType: 'expert',
    title: '申请成为行家',
    status: 'available',
    statusText: '可申请',
    desc: '通过审核后可参与服务交付、反馈进度并获得评价沉淀。',
    requirements: ['完成实名认证', '提交服务能力说明', '等待后台审核']
  },
  {
    roleType: 'guide',
    title: '申请成为领路人',
    status: 'locked',
    statusText: '条件待完善',
    desc: '领路人需要满足条件达成和付费状态双门槛。',
    requirements: ['完成实名认证', '邀请关系和撮合记录达标', '付费状态确认'],
    qualification: {
      conditionMet: false,
      paymentMet: false,
      conditionText: '条件未达成',
      paymentText: '待付费确认'
    }
  }
]

const mockInvitePlayerConfig = {
  minPlayerCount: 1,
  maxPlayerCount: 1,
  budgetMaxAmount: 99999999,
  defaultBudget: '800',
  defaultTitle: '产品架构梳理咨询',
  defaultDetail: '需要资深产品经理帮忙梳理B端产品架构，预计咨询时长2小时，涉及模块划分和数据流转设计。',
  playerIntroTemplate: '我帮你邀请了行家，可以一起确认需求、预算和服务节奏。',
  expert: {
    id: 'expert-demo',
    userId: 'expert-demo',
    name: '行家',
    avatarText: 'EX',
    roleLabel: '行家',
    desc: '资深产品经理·10年经验',
    tags: ['产品咨询', '架构梳理']
  },
  activityTypes: [
    { key: 'product', name: '产品咨询' },
    { key: 'design', name: '设计服务' },
    { key: 'tech', name: '技术开发' }
  ],
  rewardRateConfig: {
    platformServiceRate: 10,
    systemGuideRewardRate: 10,
    inviteRewardRate: 40
  }
}

const mockInvitePlayers = []

const mockSystemRecommendations = {
  recommendationId: 'system-rec-20260624-001',
  title: '系统推荐适配局',
  desc: '暂无推荐结果，请完善资料或稍后重试',
  loadingText: '推荐数据加载中',
  emptyText: '暂无匹配行家',
  summaryTemplate: '已选择 {count} 位行家',
  summaryDesc: '还可以选择多位行家组成顾问团，或搭配玩家共同组局',
  cancelText: '取消',
  confirmText: '确认组局',
  minSelectToast: '请选择至少一位行家',
  confirmingText: '正在进入组局',
  defaultSelectedExpertIds: [],
  defaultCategory: 'all',
  categories: [
    { key: 'all', name: '全部' },
    { key: 'product', name: '产品架构' },
    { key: 'tech', name: '技术咨询' },
    { key: 'operation', name: '运营策略' }
  ],
  experts: []
}

const mockDeliveryPageConfig = {
  paid: {
    pageTitle: '确认服务完成',
    status: {
      theme: 'paid',
      title: '服务已完成!',
      desc: '双方确认后，资金将全额结算'
    },
    statePill: {
      theme: 'green',
      text: '待确认完成'
    },
    notice: {},
    confirmItems: [
      { id: 'completed', title: '服务已全部完成', desc: '约定的2小时咨询服务已完整交付', checked: false },
      { id: 'qualified', title: '服务质量达标', desc: '需求方对服务内容和质量无异议', checked: false },
      { id: 'communicated', title: '双方已沟通确认', desc: '已与需求方确认服务完成，对方同意结算', checked: false }
    ],
    confirmNote: '正常交付无需扣减任何费用，只需双方确认服务已完成，资金将按全额结算。如服务未完全达标，请与玩家沟通后再确认。',
    security: {
      title: '',
      desc: ''
    },
    submitHints: {
      ready: '确认后将通知玩家进行最终确认',
      pending: '需勾选上方确认项后方可提交'
    },
    submitToast: '服务完成确认已提交',
    submitLoadingText: '提交中',
    amountRowLabel: '合同金额'
  },
  free: {
    pageTitle: '确认服务完成',
    status: {
      theme: 'free',
      title: '服务已完成!',
      desc: '双方确认后，服务正式结束'
    },
    statePill: {
      theme: 'blue',
      text: '待确认完成'
    },
    notice: {
      iconText: '🎁',
      title: '免费局说明',
      parts: [
        { text: '本局为' },
        { text: '免费体验局', strong: true },
        { text: '不涉及资金结算。双方确认完成后，行家将获得' },
        { text: '信用积分+5和免费局贡献徽章', strong: true },
        { text: '，玩家' },
        { text: '优先推荐权益', strong: true }
      ]
    },
    confirmItems: [
      { id: 'completed', title: '服务已全部完成', desc: '约定的2小时咨询服务已完整交付', checked: true, locked: true },
      { id: 'qualified', title: '服务质量达标', desc: '需求方对服务内容和质量无异议', checked: true, locked: true },
      { id: 'communicated', title: '双方已沟通确认', desc: '已与需求方确认服务完成，对方同意归档', checked: false }
    ],
    confirmNote: '免费局无需扣除任何费用，只需双方确认服务已完成，系统将自动归档。如服务未完全达标，请与玩家沟通后再次确认。',
    security: {
      title: '服务保障',
      desc: '免费局同样享受平台服务保障，评价真实有效'
    },
    submitHints: {
      ready: '确认后将通知玩家进行最终确认',
      pending: '需勾选上方确认项后方可提交'
    },
    submitToast: '免费局服务完成确认已提交',
    submitLoadingText: '提交中',
    amountRowLabel: '服务类型'
  },
  quickActions: [
    { key: 'upload', title: '上传凭证', theme: 'blue', iconText: '📎' },
    { key: 'contact_player', title: '联系玩家', theme: 'blue', iconSrc: 'https://static.haowan.net.cn/miniprogram/pages/game/delivery/assets/i18@3x.png' },
    { key: 'contact_guide', title: '联系领路人', theme: 'orange', iconText: '👬' }
  ],
  version: '2026-07-01'
}

const mockMemberRadarConfig = {
  pages: {
    radar: {
      key: 'radar',
      title: '组局雷达',
      skin: 'dark',
      estimate: '预计可匹配7461位商界决策者',
      primaryAction: '开启适配人脉',
      tip: '信息填写越完整，人脉匹配越精准',
      linkText: '填写适配信息 >'
    },
    matching: {
      key: 'matching',
      title: '组局雷达',
      skin: 'dark',
      statusText: '人脉雷达正在寻找与您适配的企业家…'
    },
    info: {
      key: 'info',
      title: '适配信息',
      skin: 'light'
    },
    query: {
      key: 'query',
      title: '组局雷达',
      skin: 'dark result',
      foundPrefix: '为您找到',
      foundSuffix: '位适配您的优质行家信息'
    },
    result: {
      key: 'result',
      title: '组局雷达',
      skin: 'dark result',
      loadingText: '正在寻找与您适配的优质业务主.....',
      resultTitle: '本轮匹配组局已完成推荐',
      resultDescPrefix: '共推荐了',
      resultDescSuffix: '位优质行家',
      resultLink: '重新查看 >',
      actionText: '再次重新匹配'
    }
  },
  formRows: [
    { key: 'location', label: '地址定位', value: '', placeholder: '选择' },
    { key: 'industry', label: '所在行业', value: '', placeholder: '选择' },
    { key: 'revenueScale', label: '营收规模', value: '', placeholder: '选填' },
    { key: 'interestedGames', label: '感兴趣组局', value: '', placeholder: '选择' },
    { key: 'resources', label: '我的资源', value: '', placeholder: '前往个人主页填写' },
    { key: 'needs', label: '我的需求', value: '', placeholder: '前往个人主页填写' },
    { key: 'recentDemand', label: '近期诉求', value: '', placeholder: '自定义填写' }
  ],
  profile: {
    id: 'lu-yi',
    userId: 201,
    name: '陆毅',
    title: '总经理｜上海创世界科技有限公司',
    tag: '第一标签：上海TMT投资领军者，数字化内容服务',
    need: '我的需求：AI赋能与市场运营助力企业IP打造',
    resource: '我的资源：10年TMT投资经验',
    address: '上海市浦东新区沙新镇黄赵路310号',
    distance: '231 km',
    avatar: 'https://static.haowan.net.cn/miniprogram/pages/profile/member/assets/radar-avatar.png'
  },
  radarNodes: [
    { id: 'hu-fang', userId: 200, className: 'node-leader', name: '胡芳', title: '董事长、创始人｜千浪化研新材料（上海…', avatar: 'https://static.haowan.net.cn/miniprogram/pages/profile/member/assets/radar-avatar.png' },
    { id: 'lu-yi', userId: 201, className: 'node-maker', name: '陆毅', title: '总经理｜上海创世界科技有限公司', avatar: 'https://static.haowan.net.cn/miniprogram/pages/profile/member/assets/radar-avatar.png' },
    { id: 'chen-zong', userId: 202, className: 'node-owner', name: '陈总', title: '企业服务资源方', avatar: '', shortName: '陈' },
    { id: 'wang-zong', userId: 203, className: 'node-investor', name: '王总', title: '产业投资合伙人', avatar: '', shortName: '王' },
    { id: 'li-zong', userId: 204, className: 'node-expert', name: '李总', title: '品牌增长顾问', avatar: '', shortName: '李' },
    { id: 'zhao-zong', userId: 205, className: 'node-partner', name: '赵总', title: '渠道合作伙伴', avatar: '', shortName: '赵' },
    { id: 'sun-zong', userId: 206, className: 'node-small', name: '孙总', title: '本地服务主理人', avatar: '', shortName: '孙' }
  ],
  result: {
    total: 10
  },
  texts: {
    criteriaMatchLabel: '符合条件的企业家',
    allMatchLabel: '适配企业家',
    criteriaScanningText: '人脉雷达正在按您的适配信息寻找企业家…',
    allScanningText: '人脉雷达正在为您匹配全部适配企业家…',
    scanDoneTemplate: '已扫描到 {count} 位{label}',
    scanProgressTemplate: '正在扫描，已发现 {count} 位{label}',
    actionFailedText: '人脉雷达操作失败',
    entryMissingText: '请选择可用入口'
  },
  actionMessages: {
    save: '已保存匹配偏好',
    next: '已为你刷新下一位',
    follow: '已关注该成员',
    profile: '暂无成员主页',
    share: '请使用右上角分享'
  }
}

const mockReplayConfirmContext = {
  sourceGameId: 'game-replay-001',
  serviceOrderId: 'SO-20260613-001',
  inviter: {
    id: 'guide-wang',
    name: '领路人',
    roleType: 'guide',
    roleLabel: '领路人'
  },
  previousSession: {
    serviceType: '产品架构咨询',
    completedAtText: '2026-06-13 14:30',
    participantText: '3人（行家+玩家+领路人）'
  },
  quickActions: [
    {
      id: 'same-friends',
      theme: 'green',
      iconText: '👫',
      title: '同局好友再玩一局',
      desc: '立即邀请上一局成员',
      route: 'confirm',
      order: 10,
      visible: true
    },
    {
      id: 'smart-match',
      theme: 'blue',
      iconText: '🤖',
      title: '系统推荐适配组局',
      desc: '基于资料和关系数据返回适配候选',
      route: 'system_recommend',
      order: 20,
      visible: true
    },
    {
      id: 'create-new',
      theme: 'pink',
      iconType: 'plus',
      title: '玩家创建新局',
      desc: '自定义需求，开启全新组局',
      route: 'create',
      order: 30,
      visible: true
    }
  ],
  quickMessages: [
    '再来一局？',
    '上次合作很愉快，继续！',
    '有个新需求想聊聊',
    '有空再约一局'
  ],
  invitees: [
    {
      id: 'expert-demo',
      name: '行家',
      roleType: 'expert',
      roleLabel: '行家',
      desc: '产品架构咨询'
    },
    {
      id: 'player-demo',
      name: '玩家',
      roleType: 'player',
      roleLabel: '玩家',
      desc: '需求方'
    }
  ]
}

const mockReviewPageConfig = {
  navTitle: '服务评价',
  skipText: '跳过',
  statusTitle: '服务已完成！',
  statusDesc: '请对本次服务进行评价',
  satisfactionQuestion: '这一局好玩吗？',
  satisfactionOptions: [
    { id: 'great', emoji: '😀', title: '很好玩', desc: '五星体验' },
    { id: 'ok', emoji: '🙂', title: '还行', desc: '基本合格' },
    { id: 'bad', emoji: '😕', title: '不好玩', desc: '有待改进' }
  ],
  storyTitle: '发生了什么有趣的事？',
  aiTip: 'AI小助手提示：可以从收获、惊喜、合作感受等方面描述哦',
  storyPlaceholder: '我们碰撞出了新的思路，对方的经验帮了大忙！',
  storyMaxLength: 100,
  aiSummaryText: 'AI帮我总结',
  ratingHint: '点击星星评分',
  npsHeadTitle: '发起人专属',
  npsQuestion: '你会推荐“真好玩”给朋友吗？ (NPS)',
  npsLowLabel: '不可能',
  npsHighLabel: '极有可能',
  submitText: '提交评价',
  submitNote: '评价内容仅双方可见，请客观公正',
  skipToast: '已跳过评价',
  submitSuccessText: '评价已提交',
  noReviewTargetText: '暂无可评价对象',
  missingTargetText: '缺少评价对象，无法提交',
  missingScoreText: '请先为每个评价对象打分',
  submitFailedText: '提交评价失败',
  defaultSummary: '本次合作沟通顺畅，交付清晰，整体体验不错。',
  againIntentBySatisfaction: {
    great: 'yes',
    ok: 'maybe',
    bad: 'no'
  },
  roleConfigs: {
    expert: {
      id: 'expert',
      avatarText: 'ZH',
      avatarTheme: 'blue',
      title: '评价行家',
      desc: '本次服务已完成',
      ratingTitle: '服务质量',
      tagTitle: '行家标签（多选）',
      tags: ['专业能力强', '交付及时', '沟通顺畅', '超出预期', '性价比高', '推荐再合作'],
      placeholder: '分享你对本次服务的评价...'
    },
    player: {
      id: 'player',
      avatarText: 'WA',
      avatarTheme: 'pink',
      title: '评价玩家',
      desc: '需求已确认，开始反馈',
      ratingTitle: '合作满意度',
      tagTitle: '玩家标签（多选）',
      tags: ['需求明确', '配合度高', '付款及时', '沟通友好', '长期合作潜力'],
      placeholder: '写下你对需求方的评价...'
    },
    guide: {
      id: 'guide',
      avatarText: 'WA',
      avatarTheme: 'orange',
      title: '评价领路人',
      desc: '撮合已完成，协助交付',
      ratingTitle: '引荐满意度',
      tagTitle: '邀约标签（多选）',
      tags: ['匹配精准', '响应及时', '协助积极', '沟通高效', '值得信赖'],
      placeholder: '写下你对引荐人的服务评价...'
    }
  },
  version: '2026-07-01'
}

const mockGuideProgress = {
  activeCount: 2,
  activeParties: [
    {
      id: 'invite-progress-001',
      status: 'waiting_expert',
      statusText: '进行中',
      timeText: '剩余23小时',
      progressText: '等待行家确认',
      progressPercent: 49,
      noticeText: '行家尚未查看邀请，可发送提醒',
      players: [
        {
          id: 'player-lina',
          name: '李娜',
          avatarText: 'LN',
          avatarClass: 'pink',
          confirmStatus: 'confirmed',
          statusText: '已确认'
        }
      ],
      experts: [
        {
          id: 'expert-wangqiang',
          name: '王强',
          avatarText: 'WQ',
          avatarClass: 'blue',
          confirmStatus: 'pending',
          statusText: '待确认'
        }
      ]
    },
    {
      id: 'invite-progress-002',
      status: 'waiting_all',
      statusText: '待双方确认',
      timeText: '刚刚',
      progressText: '等待双方确认',
      progressPercent: 0,
      players: [
        {
          id: 'player-chenming',
          name: '陈明',
          avatarText: 'CM',
          avatarClass: 'blue',
          confirmStatus: 'pending',
          statusText: '待确认'
        }
      ],
      experts: [
        {
          id: 'expert-liuying',
          name: '刘颖',
          avatarText: 'LY',
          avatarClass: 'green',
          confirmStatus: 'pending',
          statusText: '待确认'
        }
      ]
    }
  ],
  completedParties: [
    {
      id: 'invite-complete-001',
      resultStatus: 'success',
      title: '组局成功',
      timeText: '昨天',
      summaryPrefix: '你引荐的',
      completedMemberText: '张伟 与 李娜',
      players: [
        { id: 'player-zhangwei', name: '张伟' }
      ],
      experts: [
        { id: 'expert-lina', name: '李娜' }
      ],
      summarySuffix: '已成功组局',
      rewardText: '+50积分',
      gameTitle: '产品经理交流会'
    },
    {
      id: 'invite-complete-002',
      resultStatus: 'canceled',
      title: '组局已取消',
      timeText: '3天前',
      players: [
        { id: 'player-wangfang', name: '王芳' }
      ],
      rejectName: '王芳',
      rejectRoleText: '玩家',
      rejectText: '婉拒了组局邀请',
      reasonText: '原因：时间冲突'
    }
  ]
}

const mockGuideCancelDetail = {
  id: 'invite-complete-002',
  invitationId: 'invite-complete-002',
  statusTitle: '组局已取消',
  statusDesc: '玩家取消了此次组局邀请',
  canceledBy: {
    id: 'player-wangfang',
    name: '王芳',
    roleType: 'player',
    roleLabel: '玩家',
    avatarText: 'WF',
    avatarClass: 'player'
  },
  reason: {
    title: '时间冲突',
    desc: '临时有事，无法按时参加'
  },
  message: '抱歉，最近项目比较忙，时间上有冲突，希望下次有机会再合作。',
  messageTimeText: '2小时前',
  timeline: [
    {
      key: 'invite',
      title: '发起邀请',
      desc: '你向双方发送了组局邀请',
      timeText: '03-21 10:23',
      state: 'active'
    },
    {
      key: 'cancel',
      title: '玩家取消',
      desc: '王芳因时间冲突取消本次组局',
      timeText: '03-21 16:45',
      state: 'error'
    },
    {
      key: 'canceled',
      title: '组局取消',
      desc: '因一方取消，组局自动取消',
      state: 'pending'
    }
  ]
}

const mockGameCancelConfig = {
  player: {
    reasonOptions: [
      { key: 'need_changed', text: '需求变更，不再需要服务' },
      { key: 'other_solution', text: '找到其他解决方案' },
      { key: 'service_unexpected', text: '业务主服务不符合预期' },
      { key: 'budget', text: '预算问题/资金紧张' }
    ],
    defaultReason: 'other_solution',
    agreementText: '我已阅读并同意上述赔付协议，理解主动取消需承担行家的时间成本损失，并同意按设置比例从托管资金中赔付行家。',
    agreementItems: [
      '我理解主动取消需承担行家的时间成本损失',
      '我同意按设置比例赔付行家，金额从托管资金扣除',
      '剩余金额将在3个工作日内原路退回',
      '此取消记录将影响信用分（-3分）'
    ]
  },
  expert: {
    reasonOptions: [
      { key: 'schedule_conflict', text: '个人时间冲突，无法交付' },
      { key: 'requirement_mismatch', text: '需求与描述不符，无法完成' },
      { key: 'emergency', text: '身体原因/突发状况' },
      { key: 'other', text: '其他原因' }
    ],
    defaultReason: 'schedule_conflict',
    agreementText: '我已阅读并同意《服务取消协议》，理解主动取消将对我的信用分产生影响（-5分），并同意按设置比例赔付玩家损失。'
  },
  version: '2026-07-01'
}

const mockMyGamesPageConfig = {
  pageTitle: '我的局',
  emptyText: '暂无相关局',
  detailMissing: '暂无组局详情',
  actionMissing: '暂无可执行操作',
  categoryTabs: [
    { key: 'joined', text: '我参与的' },
    { key: 'invited', text: '我受邀的' },
    { key: 'favorite', text: '我收藏的' }
  ],
  statusTabs: [
    { key: 'all', text: '全部' },
    { key: 'active', text: '进行中' },
    { key: 'complete', text: '已完成' },
    { key: 'overdue', text: '超时' },
    { key: 'canceled', text: '已取消' }
  ],
  version: '2026-07-01'
}

const mockGameManage = {
  pageConfig: mockMyGamesPageConfig,
  summary: {
    label: '本月服务收入',
    amount: 0,
    amountText: '¥0',
    activeCount: 0,
    pendingSettlementCount: 0,
    completedCount: 0,
    disputeCount: 0
  },
  orders: []
}

const mockPlayerGameManage = {
  pageConfig: mockMyGamesPageConfig,
  currentTime: '',
  summary: {
    label: '本月服务支出',
    amount: 0,
    amountText: '¥0',
    activeCount: 0,
    completedCount: 0,
    canceledCount: 0
  },
  orders: []
}
const mockGameProfitTemplates = {
  currentAccountType: 'player',
  depositRuleText: '连续打卡 7 天即完成。完成者拿回押金池金额，未完成者押金由完成者平分。',
  depositNoticeText: '支付金额：100元 = 服务费10元 + 押金池90元。服务费不退，押金池按完成情况结算。',
  templates: [
    {
      key: 'standard',
      name: '标准 1441',
      desc: '平台10% · 流量方40% · 交付方40% · 推荐上级10%',
      selectable: true,
      allowedAccountTypes: ['player', 'expert', 'guide', 'platform']
    },
    {
      key: 'aa',
      name: 'AA局',
      desc: '平台2.5% · 交付方90% · 流量方5% · 推荐上级2.5%',
      selectable: true,
      allowedAccountTypes: ['player', 'expert', 'guide', 'platform']
    },
    {
      key: 'deposit',
      name: '押金局',
      desc: '平台2.5% · 交付方0% · 流量方5% · 推荐上级2.5% + 押金池90%，完成返还，未完成瓜分',
      selectable: true,
      allowedAccountTypes: ['player', 'expert', 'guide', 'platform']
    },
    {
      key: 'crowdfunding',
      name: '众筹局',
      desc: '平台2.5% · 交付方90% · 流量方5% · 推荐上级2.5%',
      selectable: true,
      allowedAccountTypes: ['player', 'expert', 'guide', 'platform']
    },
    {
      key: 'publicBenefit',
      name: '公益局',
      desc: '平台0% · 交付方100% · 流量方0% · 推荐上级0%',
      selectable: false,
      disabledReason: '仅平台账户可发起',
      allowedAccountTypes: ['platform']
    }
  ]
}

const mockRelationNetworkHome = {
  onlineText: '在线',
  header: {
    titleIcon: '📍',
    title: '星巴克(镇海万科店)',
    statusText: '营业中',
    address: '宁波市镇海区庄市大道1088号万科广场1F'
  },
  tabs: [
    { key: 'network', text: '人脉网络' },
    { key: 'nearby', text: '附近玩家' }
  ],
  activeTab: 'network'
}

const mockMyCityConfig = {
  onlineText: '在线',
  pageTitle: '我的城市故事',
  profileName: '我的信息',
  routeTip: '收集8条更早行程，航线图更完整',
  switchMapText: '切换为火车',
  journeyTitle: '我的局迹',
  journeyDesc: '通过 12 个局，认识了 28 位朋友',
  participantLabel: '参与者：',
  endingTitle: '我是有底线的',
  endingDesc: '继续探索，创造更多故事',
  summaryStats: [
    { value: '28', label: '故事总数', tone: 'blue' },
    { value: '6', label: '覆盖城市', tone: 'violet' },
    { value: '52', label: '参与组局数', tone: 'pink' }
  ],
  mapLegends: [
    { label: '第一次', tone: 'pink' },
    { label: '夜游', tone: 'violet' },
    { label: '社交局', tone: 'blue' },
    { label: '最难忘', tone: 'gold' }
  ],
  mapStats: [
    { label: '里程', value: '40244', unit: '公里' },
    { label: '次数', value: '33', unit: '次' },
    { label: '国家/地区', value: '1', unit: '个' },
    { label: '城市', value: '10', unit: '个' }
  ],
  tagEmojis: {
    最难忘: '👑',
    桌游局: '🎲',
    微醺局: '🍷',
    脑暴局: '💡',
    篮球局: '🏀',
    第一次: '🌱',
    起点: '🌱',
    摄影局: '📷'
  },
  badgeEmojis: {
    创业伙伴: '🤝',
    深度密友: '💬',
    合伙人: '🤝',
    室友: '🏠',
    固定局友: '📌'
  },
  storyGroups: [
    {
      year: 2024,
      stories: [
        {
          id: 'countdown-night',
          tag: '最难忘',
          tagTone: 'gold',
          title: '跨年夜的倒计时',
          date: '12.31',
          sortDate: '2024-12-31',
          location: '',
          cover: 'https://static.haowan.net.cn/miniprogram/pages/map/my-city/assets/story-river-cover.png',
          coverLocation: '上海 · 外滩',
          desc: '和刚认识的摄影局朋友们一起在外滩等待新年钟声。江风吹得发抖，但倒数的那一刻，所有的陌生人都变成了朋友。',
          participants: ['a', 'b', 'c'],
          badgeTitle: '',
          badgeDesc: '',
          actionText: '',
          actionTone: ''
        },
        {
          id: 'script-rain',
          tag: '桌游局',
          subTag: '新手场',
          tagTone: 'violet',
          title: '暴雨中的剧本杀',
          date: '2024.09.20',
          sortDate: '2024-09-20',
          location: '杭州 · 西湖区 · 14:00-22:00',
          desc: '原定5人的局因为暴雨只来了3人，却因此有了最深入的交谈。认识了做AI的@阿杰，现在我们是创业合伙人。',
          participants: ['a', 'b', 'c', 'd'],
          badgeTitle: '创业伙伴',
          badgeDesc: '已共同发起 3 个项目',
          actionText: '查看项目',
          actionTone: 'violet'
        },
        {
          id: 'truth-night',
          tag: '微醺局',
          subTag: '深夜场',
          tagTone: 'orange',
          title: '周五晚上的坦白局',
          date: '2024.11.03',
          sortDate: '2024-11-03',
          location: '北京 · 三里屯 · 21:00',
          desc: '"你最后悔的事是什么？"那个问题让陌生人变成了知己。和@Lucy约定每月一次深度对话。',
          participants: ['a', 'b'],
          badgeTitle: '深度密友',
          badgeDesc: '',
          actionText: '约下次',
          actionTone: 'orange'
        },
        {
          id: 'business-canvas',
          tag: '脑暴局',
          subTag: '创始人专场',
          tagTone: 'gold',
          title: '凌晨的商业模式画布',
          date: '2024.12.15',
          sortDate: '2024-12-15',
          location: '深圳 · 科技园 · 通宵',
          desc: '从晚上8点到早上6点，8个人在黑板上画满了想法。这个局让我找到了技术合伙人@老K。',
          participants: ['a', 'b', 'c', 'd'],
          badgeTitle: '合伙人',
          badgeDesc: '公司估值 500w',
          actionText: '查看公司',
          actionTone: 'gold'
        },
        {
          id: 'court-weekly',
          tag: '篮球局',
          subTag: '每周固定',
          tagTone: 'cyan',
          title: '东华球场的汗水',
          date: '每周六',
          sortDate: '2024-06-15',
          location: '上海 · 东华大学 · 16:00',
          desc: '最纯粹的快乐。这里没有身份，只有队友。通过球局认识了现在的室友@阿强。',
          participants: ['a', 'b', 'c'],
          badgeTitle: '室友',
          badgeDesc: '合租 6 个月',
          actionText: '加入球局',
          actionTone: 'cyan'
        },
        {
          id: 'first-use',
          type: 'first',
          tag: '起点',
          subTag: '',
          tagTone: 'green',
          highlightTag: '第一次',
          title: '第一次使用真好玩',
          date: '03.12',
          sortDate: '2024-03-12',
          location: '广州 · 天河公园',
          desc: '抱着试试看的心态参加了第一次飞盘局，从此打开了城市探索的新世界。',
          participants: [],
          badgeTitle: '',
          badgeDesc: '',
          actionText: '',
          actionTone: ''
        }
      ]
    },
    {
      year: 2023,
      stories: [
        {
          id: 'sunrise-photo',
          tag: '摄影局',
          subTag: '第1次参与',
          tagTone: 'pink',
          title: '外滩 sunrise 拍摄',
          date: '2023.03.12',
          sortDate: '2023-03-12',
          location: '上海 · 外滩观景台 · 06:00',
          desc: '为了拍日出早上5点起床，认识了同样疯狂的@小林和@大为。后来我们组成了固定摄影小队，每周六早扫街。',
          participants: ['a', 'b', 'c'],
          badgeTitle: '固定局友',
          badgeDesc: '已持续组队 8 个月',
          actionText: '再组一局',
          actionTone: 'blue'
        }
      ]
    }
  ],
  version: '2026-06-30'
}

const mockMessageCenter = {
  pageTitle: '消息中心',
  onlineText: '在线',
  activeTab: 'all',
  quickActions: [
    { key: 'join', label: '组局加入', iconSrc: 'https://static.haowan.net.cn/miniprogram/pages/message/assets/i53@3x.png', tone: 'blue', unreadCount: 1 },
    { key: 'system', label: '系统通知', iconSrc: 'https://static.haowan.net.cn/miniprogram/pages/message/assets/i54@3x.png', tone: 'green', unreadCount: 0 },
    { key: 'achievement', label: '成就解锁', iconSrc: 'https://static.haowan.net.cn/miniprogram/pages/message/assets/i55@3x.png', tone: 'yellow', unreadCount: 0 },
    { key: 'warning', label: '预警通知', iconSrc: 'https://static.haowan.net.cn/miniprogram/pages/message/assets/i56@3x.png', tone: 'red', unreadCount: 0 },
    { key: 'friend', label: '好友', iconSrc: '/pages/message/assets/i57@3x.png', tone: 'cyan', unreadCount: 3 }
  ],
  tabs: [
    { key: 'all', label: '全部消息' },
    { key: 'unread', label: '未读 (3)', unreadCount: 3 },
    { key: 'trade', label: '交易通知' }
  ],
  actionTexts: {
    accept: '确认参加',
    reject: '婉拒',
    process: '立即处理',
    review: '立即评价',
    detail: '查看详情',
    game: '查看组局',
    contact: '联系发起人'
  },
  texts: {
    loadFailedText: '消息中心加载失败',
    entryMissingText: '暂无可打开的消息入口',
    openFailedText: '消息打开失败',
    actionMissingText: '操作信息不完整',
    actionSuccessText: '操作成功',
    actionHandledText: '已标记处理',
    actionFailedText: '消息操作失败',
    serviceMissingText: '缺少消息操作信息',
    serviceFailedText: '消息操作失败'
  },
  sections: [
    {
      key: 'system',
      title: '系统通知',
      items: [
        {
          id: 'platform-notice',
          routeKey: 'system',
          iconSrc: 'https://static.haowan.net.cn/miniprogram/pages/message/assets/i58@3x.png',
          tone: 'blue',
          title: '平台公告',
          timeText: '2小时前',
          desc: '关于组局功能升级的通知：新增“智能匹配”功能，可自动推荐合适的组局对象...'
        },
        {
          id: 'audit-result',
          iconSrc: 'https://static.haowan.net.cn/miniprogram/pages/message/assets/i59@3x.png',
          tone: 'purple',
          title: '活动审核结果',
          timeText: '昨天',
          desc: '你发布的活动“AI技术分享会”已通过审核，将于明天10:00开始展示或者前往组局中心手动发布',
          tagText: '审核通过',
          tagTone: 'success'
        }
      ]
    },
    {
      key: 'group',
      title: '组局动态',
      moreText: '查看全部',
      items: [
        {
          id: 'group-confirm',
          iconSrc: 'https://static.haowan.net.cn/miniprogram/pages/message/assets/i60@3x.png',
          tone: 'orange',
          unread: true,
          title: '组局确认通知',
          timeText: '10:23',
          desc: '张伟 发起组局邀请你参与“周末篮球局”，需要你确认是否参加',
          highlightText: '张伟',
          actions: [
            { key: 'accept', text: '确认参加', primary: true },
            { key: 'reject', text: '婉拒' }
          ]
        },
        {
          id: 'group-success',
          iconSrc: 'https://static.haowan.net.cn/miniprogram/pages/message/assets/i54@3x.png',
          tone: 'green',
          title: '组局已成局',
          timeText: '昨天',
          desc: '你引荐的 李娜 与 王强 已成功组局“产品经理交流会”',
          summaryText: '✓ 引荐成功',
          subText: '获得积分 +50'
        },
        {
          id: 'pay-success',
          iconSrc: '/pages/message/assets/i61@3x.png',
          tone: 'orangeLight',
          title: '支付成功通知',
          timeText: '昨天',
          desc: '你成功支付了“早起星人挑战”押金 ¥100.00，资金已进入押金池托管。'
        },
        {
          id: 'join-apply',
          avatarText: '小',
          title: '小红 申请加入你的局',
          timeText: '10:30',
          desc: '局：【武康路】复古胶片摄影局...',
          actions: [
            { key: 'decline', text: '拒绝' },
            { key: 'chat', text: '通过并私聊', primary: true, orange: true }
          ]
        }
      ]
    },
    {
      key: 'achievement',
      title: '新增成就',
      items: [
        {
          id: 'achievement-unlock',
          iconSrc: '/pages/message/assets/i62@3x.png',
          tone: 'yellow',
          title: '解锁新成就！',
          timeText: '3月30日',
          desc: '恭喜你解锁了“魔都探险家”成就，获得 200 积分奖励！'
        }
      ]
    },
    {
      key: 'warning',
      title: '预警提醒',
      items: [
        {
          id: 'delivery-warning',
          routeKey: 'warning',
          iconSrc: 'https://static.haowan.net.cn/miniprogram/pages/message/assets/i65@3x.png',
          tone: 'red',
          alert: true,
          title: '待交付订单提醒',
          timeText: '2小时前',
          descParts: [
            { text: '你有1个组局服务订单将于 ' },
            { text: '2小时后', danger: true },
            { text: ' 到期交付，请及时处理' }
          ],
          metaText: '订单号：GD2024032201',
          linkText: '立即处理'
        },
        {
          id: 'activity-soon',
          iconSrc: 'https://static.haowan.net.cn/miniprogram/pages/message/assets/i64@3x.png',
          tone: 'yellow',
          title: '活动即将开始',
          timeText: '30分钟后',
          desc: '你参与的组局“周末徒步”将于今天14:00开始，地点：奥林匹克森林公园南门',
          actions: [
            { key: 'route', text: '查看路线', primary: true },
            { key: 'contact', text: '联系发起人' }
          ]
        }
      ]
    },
    {
      key: 'friends',
      title: '好友消息',
      items: [
        {
          id: 'friend-user-a',
          routeKey: 'friend',
          avatarText: 'UA',
          online: true,
          title: '用户A',
          timeText: '12:30',
          desc: '好的，那我们就周六下午2点在咖啡店见，我带上项目资料...',
          unreadCount: 3
        }
      ]
    }
  ]
}

const mockMessageMyConfig = {
  pageTitle: '好友消息',
  onlineText: '在线',
  friend: {
    defaultInitials: 'IM',
    defaultName: '局内会话',
    defaultStatus: '在线',
    nameTemplate: '成员 {userId}'
  },
  quickActions: [
    { key: 'friend', label: '加好友' },
    { key: 'greet', label: '打招呼', messageText: '你好，我看到你的消息了。' },
    { key: 'card', label: '发名片', messageText: '这是我的名片，后续可以在局内继续沟通。' },
    { key: 'location', label: '发定位' }
  ],
  texts: {
    loadFailedText: '加载会话失败',
    actionMissingText: '操作信息不完整',
    sendFailedText: '发送失败',
    recordStartText: '开始录音',
    recordStopText: '当前支持文字、图片和文件消息',
    recordErrorText: '录音失败',
    fileEntryMissingText: '请从局内消息入口发送文件',
    justNowText: '刚刚'
  }
}

const mockTradeWarningDetail = {
  id: 'trade-warning-001',
  warningId: 'trade-warning-001',
  pageTitle: '交易预警',
  onlineText: '在线',
  warning: {
    title: '即将超时',
    prefixText: '该订单将于',
    highlightText: '1小时30分钟',
    suffixText: '后自动标记为逾期，请立即处理'
  },
  countdown: [
    { value: '01', label: '小时' },
    { value: '30', label: '分钟' },
    { value: '45', label: '秒' }
  ],
  order: {
    orderNo: 'GD2024032201',
    statusText: '待交付',
    customerAvatarText: 'CL',
    customerTitle: '客户需求',
    customerDesc: '寻找资深产品经理进行业务咨询',
    detailRows: [
      { label: '约定交付时间', value: '今天 16:00' },
      { label: '服务费用', value: '¥500', strong: true }
    ]
  },
  deliveryMethods: [
    {
      id: 'online',
      title: '线上确认',
      desc: '双方在线确认服务完成',
      active: true
    },
    {
      id: 'upload',
      title: '上传凭证',
      desc: '上传服务完成截图或文件',
      active: false
    }
  ],
  actions: {
    delayText: '申请延期',
    deliverText: '立即交付'
  },
  texts: {
    loadingText: '加载中...',
    loadFailedText: '获取交易预警失败',
    invalidActionText: '交易预警操作无效',
    actionFailedText: '交易预警处理失败',
    delaySuccessText: '延期申请已提交',
    delayStatusText: '已申请延期',
    deliverSuccessText: '已进入交付确认',
    countdownTitle: '剩余交付时间',
    orderNoLabel: '订单编号',
    deliveryTitle: '交付方式'
  }
}

const mockSystemNotificationDetail = {
  id: 'system-notification-001',
  messageId: 'system-notification-001',
  notificationId: 'system-notification-001',
  pageTitle: '系统通知',
  onlineText: '在线',
  article: {
    tagText: '重要更新',
    title: '组局功能全新升级：智能匹配系统上线',
    author: '官方运营团队',
    publishedAtText: '2026-03-20',
    readText: '阅读 1.2k',
    blocks: [
      {
        id: 'lead',
        type: 'paragraph',
        text: '亲爱的用户：',
        lead: true
      },
      {
        id: 'intro',
        type: 'paragraph',
        text: '为了提升组局效率和匹配精准度，我们于今日正式上新智能匹配功能，根据你的行业标签、兴趣爱好、地理位置等多维度信息，自动推荐最合适的组局对象。'
      },
      {
        id: 'update-content',
        type: 'updateBox',
        icon: '★',
        title: '主要更新内容',
        points: [
          'AI智能推荐：基于行为分析的个性化推荐',
          '匹配度评分：直观展示双方契合程度',
          '一键邀约：简化组局发起流程'
        ]
      },
      {
        id: 'message-center',
        type: 'paragraph',
        text: '同时，我们对消息触达中心进行了优化，新增消息分类和优先级标记，确保你不会错过任何重要组局信息。'
      },
      {
        id: 'cover',
        type: 'cover',
        imageUrl: 'https://static.haowan.net.cn/miniprogram/pages/message/system-detail/assets/system-update-cover.png',
        caption: '智能匹配界面示意图'
      },
      {
        id: 'closing',
        type: 'paragraph',
        text: '如有任何问题，欢迎联系客服团队。感谢你的支持与信任！'
      },
      {
        id: 'signature',
        type: 'signature',
        teamText: '产品团队',
        dateText: '2026年3月20日'
      }
    ]
  },
  feedback: {
    question: '这篇文章对你有帮助吗？',
    useful: {
      icon: '👍',
      label: '有用',
      count: 128,
      countText: '128'
    },
    useless: {
      icon: '👎',
      label: '没用',
      count: 10,
      countText: '10'
    }
  },
  texts: {
    loadFailedText: '获取系统通知失败',
    feedbackFailedText: '反馈提交失败',
    feedbackSuccessText: '已记录{label}反馈'
  }
}

const mockReportCenterConfig = {
  types: [
    { key: 'private-guide', label: '诱导私下交易', reportType: 'revenue_dispute', order: 10, visible: true },
    { key: 'private-done', label: '私下交易已完成', reportType: 'revenue_dispute', order: 20, visible: true },
    { key: 'harassment', label: '言语骚扰', reportType: 'user_complaint', order: 30, visible: true },
    { key: 'fake', label: '虚假信息', reportType: 'user_complaint', order: 40, visible: true },
    { key: 'cancel', label: '恶意取消', reportType: 'service_dispute', order: 50, visible: true },
    { key: 'other', label: '其他违规', reportType: 'other', order: 60, visible: true }
  ],
  defaultType: 'private-guide',
  maxEvidenceCount: 9,
  allowedUploadTypes: ['jpg', 'png', 'pdf'],
  tips: [
    '举报属实且能核实金额：罚款20% (50%奖励举报人)',
    '属实但无法核实：按后台处理规则发放奖励并记录信用变化',
    '不属实扣除举报人信用分2分，多次恶意举报封号'
  ],
  appealReasons: [
    { key: 'misjudge', label: '误判扣分', order: 10, visible: true },
    { key: 'system', label: '系统错误', order: 20, visible: true },
    { key: 'special', label: '特殊情况', order: 30, visible: true },
    { key: 'other', label: '其他', order: 40, visible: true }
  ],
  appealPlaceholder: '请详细说明申诉原因，包括但不限于事件经过、时间、涉及人员等信息...',
  appealUploadNote: '支持 JPG、PNG 格式，单张不超过 5MB，最多 4 张证明材料',
  appealFileMaxCount: 4,
  appealUploadFullText: '最多上传 4 张证明材料',
  appealUploadSelectedTemplate: '已选择 {selected}/{max} 张证明材料',
  appealReviewTitle: '处理时效',
  appealReviewRules: [
    '提交后24小时内初审',
    '复杂情况48小时内复核',
    '结果将通过站内消息通知'
  ],
  version: '2026-07-01'
}

module.exports = {
  validInvites,
  mockRoleStatusPageConfig,
  mockRoleApplicationPageConfig,
  mockGuideApplyConfig,
  mockRoleBenefitConfig,
  mockReferralRecordsConfig,
  mockUser,
  mockCurrentUser,
  mockHome,
  mockRoleHomes,
  mockProfileHome,
  mockProfileAssets,
  mockSystemSkillConfig,
  mockPointsMall,
  mockPointsPageConfig,
  mockPointsOrders,
  mockPointsOrderLogistics,
  mockNewbieTasks,
  mockRoleApplications,
  mockInvitePlayerConfig,
  mockInvitePlayers,
  mockSystemRecommendations,
  mockDeliveryPageConfig,
  mockMemberRadarConfig,
  mockReportCenterConfig,
  mockReviewPageConfig,
  mockReplayConfirmContext,
  mockGuideProgress,
  mockGuideCancelDetail,
  mockGameCancelConfig,
  mockMyGamesPageConfig,
  mockGameManage,
  mockPlayerGameManage,
  mockGameProfitTemplates,
  mockRelationNetworkHome,
  mockMyCityConfig,
  mockMessageCenter,
  mockMessageMyConfig,
  mockTradeWarningDetail,
  mockSystemNotificationDetail
}
