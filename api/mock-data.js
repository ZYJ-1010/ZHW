const validInvites = {
  ENJOY2026: {
    id: 'invite-code-001',
    code: 'ENJOY2026',
    ownerUserId: 'user-guide-001',
    inviterName: '周领路人',
    city: '上海',
    status: 'active'
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
    onlineText: '3999人在线'
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
      coverUrl: '/components/game-card/assets/cover-sunset.png',
      scope: 'city',
      actions: ['share', 'follow', 'refer', 'greet'],
      route: 'pages/game/detail/index',
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
      coverUrl: '/components/game-card/assets/cover-city.png',
      scope: 'nearby',
      actions: ['share', 'follow', 'refer', 'greet'],
      route: 'pages/game/detail/index',
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
          avatarUrl: '/pages/home/player/assets/ranking-avatar-01.png',
          avatarFallback: 'PRO',
          desc: '本周组局 12 · MVP 5次',
          xpText: '2,450 XP'
        },
        {
          id: 'rank-player-02',
          rank: 2,
          nickname: '社交达人',
          avatarUrl: '/pages/home/player/assets/ranking-avatar-02.png',
          avatarFallback: '星',
          desc: '本周组局 8 次',
          xpText: '1,890 XP'
        },
        {
          id: 'rank-player-03',
          rank: 3,
          nickname: '探险家',
          avatarUrl: '/pages/home/player/assets/ranking-avatar-03.png',
          avatarFallback: '探',
          desc: '本周组局 6 次',
          xpText: '1,560 XP'
        }
      ],
      myRank: {
        rank: 52,
        nickname: '我（Alex）',
        avatarUrl: '/pages/home/player/assets/ranking-avatar-me.png',
        avatarFallback: 'A',
        desc: '上周排名 65 ↑',
        xpText: '520 XP'
      }
    }
  },
  rankingList: [
    { rank: '01', name: '领域专家 PRO', avatarUrl: '/pages/home/player/assets/ranking-avatar-01.png', desc: '本周组局 12 · MVP 5次', xpText: '2,450 XP' },
    { rank: '02', name: '社交达人', avatarUrl: '/pages/home/player/assets/ranking-avatar-02.png', desc: '本周组局 8 次', xpText: '1,890 XP' },
    { rank: '03', name: '探险家', avatarUrl: '/pages/home/player/assets/ranking-avatar-03.png', desc: '本周组局 6 次', xpText: '1,560 XP' }
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
      coverUrl: '/components/game-card/assets/cover-sunset.png',
      actionText: '加入',
      actions: ['share', 'follow', 'refer', 'greet'],
      route: 'pages/game/detail/index'
    }
  ],
  metaverseEntry: {
    title: '进入元宇宙',
    desc: '共创数字街区｜全球联机互动',
    tags: ['3D空间', 'NFT徽章'],
    avatars: [
      { avatarUrl: '/pages/home/player/assets/ranking-avatar-01.png', avatarFallback: 'A' },
      { avatarUrl: '/pages/home/player/assets/ranking-avatar-02.png', avatarFallback: 'L' },
      { avatarUrl: '/pages/home/player/assets/ranking-avatar-03.png', avatarFallback: 'M' }
    ],
    joinedCount: 99,
    actionText: '进入元宇宙',
    route: 'pages/placeholder/metaverse/index'
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
      onlineText: '3999人在线'
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
          title: '《血染钟楼》10人局',
          meta: '星巴克(万科店) · 1.2km · 3/10人',
          dateText: '2026年5月1日 14:00--16:00',
          tags: ['专业场', '探索局'],
          status: '即将满员',
          statusTone: 'green',
          income: '¥200',
          joinedText: '+3位玩家已入局',
          participantAvatars: [
            '/pages/home/player/assets/ranking-avatar-01.png',
            '/pages/home/player/assets/ranking-avatar-02.png',
            '/pages/home/player/assets/ranking-avatar-03.png'
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
            '/pages/home/player/assets/ranking-avatar-01.png',
            '/pages/home/player/assets/ranking-avatar-02.png',
            '/pages/home/player/assets/ranking-avatar-03.png'
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
          { id: 'rank-expert-01', rank: 1, nickname: '领域专家 PRO', avatarUrl: '/pages/home/player/assets/ranking-avatar-01.png', avatarFallback: 'PRO', desc: '本周服务玩家 90 位', xpText: '2,450 XP' },
          { id: 'rank-expert-02', rank: 2, nickname: '社交达人', avatarUrl: '/pages/home/player/assets/ranking-avatar-02.png', avatarFallback: '星', desc: '本周服务玩家 10 位', xpText: '1,890 XP' },
          { id: 'rank-expert-03', rank: 3, nickname: '探险家', avatarUrl: '/pages/home/player/assets/ranking-avatar-03.png', avatarFallback: '探', desc: '本周服务玩家 1 位', xpText: '1,560 XP' }
        ],
        myRank: {
          rank: 52,
          nickname: '我（Alex）',
          avatarUrl: '/pages/home/player/assets/ranking-avatar-me.png',
          avatarFallback: 'A',
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
      onlineText: '3999人在线'
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
          coverUrl: '/components/game-card/assets/cover-city.png',
          scope: 'nearby',
          actions: ['share', 'follow', 'refer', 'greet'],
          route: 'pages/game/detail/index',
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
          coverUrl: '/components/game-card/assets/cover-sunset.png',
          scope: 'city',
          actions: ['share', 'follow', 'refer', 'greet'],
          route: 'pages/game/detail/index',
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
          { id: 'guide-reco-1', icon: '👑', name: '剧本杀小王', desc: '贡献¥320', tag: '查看分润' },
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
        income: '本周收益 ¥1,240',
        location: '📍 镇海区',
        items: [
          { id: 'script-master', icon: '👑', name: '剧本杀小王', desc: '贡献¥320' },
          { id: 'student-dm', icon: '🎓', name: '大学生DM', desc: '宁大节点' },
          { id: 'mama-group', icon: '👶', name: '宝妈组局', desc: '周活跃' },
          { id: 'researcher', icon: '🔬', name: '研究员阿伟', desc: '中科院' },
          { id: 'all', icon: '+', name: '查看全部', desc: '156人', dashed: true }
        ],
        buttons: [
          { text: '管理我的连接', primary: true, route: 'pages/message/index' },
          { text: '查看分润', route: 'pages/profile/index' }
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
          { id: 'rank-guide-01', rank: 1, nickname: '城市连接官', avatarUrl: '/pages/home/player/assets/ranking-avatar-01.png', avatarFallback: '城', desc: '本周连接玩家 156 位', xpText: '2,450 XP' },
          { id: 'rank-guide-02', rank: 2, nickname: '社群引路人', avatarUrl: '/pages/home/player/assets/ranking-avatar-02.png', avatarFallback: '社', desc: '本周成功引荐 32 次', xpText: '1,890 XP' },
          { id: 'rank-guide-03', rank: 3, nickname: '活动发现家', avatarUrl: '/pages/home/player/assets/ranking-avatar-03.png', avatarFallback: '活', desc: '本周活跃节点 18 个', xpText: '1,560 XP' }
        ],
        myRank: {
          rank: 52,
          nickname: '我（萧飒）',
          avatarUrl: '/pages/home/player/assets/ranking-avatar-me.png',
          avatarFallback: '萧',
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
  user: mockCurrentUser,
  stats: [
    { label: '我的局', value: 3 },
    { label: '我的申请', value: 2 },
    { label: '可评价', value: 1 },
    { label: '消息', value: 3 }
  ],
  menuItems: [
    { id: 'member', title: '会员中心', desc: '基础会员 · 分润资格待完善', route: 'pages/profile/member/index' },
    { id: 'achievements', title: '成长与成就', desc: 'Lv.2 · 信用 100 · 积分 260', route: 'pages/profile/achievements/index' },
    { id: 'games', title: '我的局', desc: '发起和参与的局', route: 'pages/game/hall/index' },
    { id: 'reviews', title: '可评价的局', desc: '1 个局待评价', route: 'pages/game/review/index' },
    { id: 'income', title: '收益信息', desc: '待结算收益和明细', route: 'pages/profile/index' }
  ],
  incomeSummary: {
    pendingAmountText: '¥0.00',
    settledAmountText: '¥0.00'
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
  maxPlayerCount: 1
}

const mockInvitePlayers = [
  {
    id: 'liming',
    avatarText: 'LM',
    avatarClass: 'pink',
    name: '李明',
    tag: '需求匹配',
    desc: '某互联网公司 · 产品总监',
    meta: '预算: ¥500-1000 | 时间: 本周'
  },
  {
    id: 'wanghua',
    avatarText: 'WH',
    avatarClass: 'teal',
    name: '王华',
    tag: '',
    desc: '寻找UI设计合作',
    meta: '预算: ¥2000+ | 长期合作'
  },
  {
    id: 'chenzhe',
    avatarText: 'CZ',
    avatarClass: 'purple',
    name: '陈哲',
    tag: '',
    desc: '需要技术顾问',
    meta: '预算: 面议 | 长期需求'
  },
  {
    id: 'zhaomin',
    avatarText: 'ZM',
    avatarClass: 'blue',
    name: '赵敏',
    tag: '常合作',
    desc: '品牌运营 · 社群增长',
    meta: '预算: ¥1000-2000 | 下周可约'
  },
  {
    id: 'sunyan',
    avatarText: 'SY',
    avatarClass: 'orange',
    name: '孙岩',
    tag: '',
    desc: '独立开发者 · 技术顾问',
    meta: '预算: 面议 | 晚间方便'
  }
]

const mockSystemRecommendations = {
  recommendationId: 'system-rec-20260624-001',
  title: '系统推荐适配局',
  desc: '基于你的偏好，已找到5个高匹配度行家',
  defaultSelectedExpertIds: ['li-senior'],
  categories: [
    { key: 'all', name: '全部' },
    { key: 'product', name: '产品架构' },
    { key: 'tech', name: '技术咨询' },
    { key: 'operation', name: '运营策略' }
  ],
  experts: [
    {
      id: 'li-senior',
      name: '李资深',
      role: '前阿里P8 · 产品架构专家',
      avatarText: 'LI',
      avatarClass: 'purple',
      rating: '5.0',
      stars: '★★★★★',
      reviewCount: 128,
      match: 98,
      price: 800,
      category: 'product',
      tags: ['产品架构', '技术方案', '团队管理', '响应及时'],
      selected: true
    },
    {
      id: 'chen-consultant',
      name: '陈顾问',
      role: '腾讯T3 · 技术架构师',
      avatarText: 'CH',
      avatarClass: 'teal',
      rating: '4.8',
      stars: '★★★★☆',
      reviewCount: 86,
      match: 95,
      price: 600,
      category: 'tech',
      tags: ['系统架构', '微服务', '云原生', '专业深度']
    },
    {
      id: 'zhao-growth',
      name: '赵顾问',
      role: '字节跳动 · 增长专家',
      avatarText: 'ZH',
      avatarClass: 'indigo',
      rating: '4.9',
      stars: '★★★★★',
      reviewCount: 64,
      match: 88,
      price: 700,
      category: 'operation',
      tags: ['用户增长', '数据分析', 'A/B测试']
    },
    {
      id: 'meng-designer',
      name: '孟设计师',
      role: '独立设计顾问 · UI/UX',
      avatarText: 'ME',
      avatarClass: 'cyan',
      rating: '4.7',
      stars: '★★★★☆',
      reviewCount: 52,
      match: 85,
      price: 500,
      category: 'product',
      tags: ['UI设计', '交互设计', '设计系统']
    },
    {
      id: 'sun-operation',
      name: '孙运营',
      role: '美团 · 运营策略专家',
      avatarText: 'SU',
      avatarClass: 'orange',
      rating: '5.0',
      stars: '★★★★★',
      reviewCount: 93,
      match: 72,
      price: 550,
      category: 'operation',
      tags: ['运营策略', '社群运营', '活动策划']
    }
  ]
}

const mockReplayConfirmContext = {
  sourceGameId: 'game-replay-001',
  serviceOrderId: 'SO-20260613-001',
  inviter: {
    id: 'guide-wang',
    name: '王引荐',
    roleType: 'guide',
    roleLabel: '领路人'
  },
  previousSession: {
    serviceType: '产品架构咨询',
    completedAtText: '2026-06-13 14:30',
    participantText: '3人（行家+玩家+领路人）'
  },
  invitees: [
    {
      id: 'expert-zhang',
      name: '张专家',
      roleType: 'expert',
      roleLabel: '行家',
      desc: '产品架构咨询'
    },
    {
      id: 'player-wang',
      name: '王总',
      roleType: 'player',
      roleLabel: '玩家',
      desc: '需求方'
    }
  ]
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

const mockGameManage = {
  summary: {
    label: '本月服务收入',
    amount: 5280,
    amountText: '¥5,280',
    activeCount: 2,
    pendingSettlementCount: 1,
    completedCount: 8,
    disputeCount: 0
  },
  orders: [
    {
      id: 'business-active-001',
      statusType: 'active',
      statusText: '服务进行中',
      ref: 'REF-20260320-001',
      avatarText: 'LI',
      avatarClass: 'pink',
      name: '李明',
      roleTag: '玩家',
      serviceText: '产品架构咨询 · ¥800',
      guideName: '王引荐',
      timeline: [
        { id: 'group-success', title: '组局成功', time: '03-20 14:30', state: 'done' },
        { id: 'service-active', title: '服务进行中', time: '预计交付：03-25', state: 'current' },
        { id: 'waiting-confirm', title: '等待确认完成', state: 'future' }
      ],
      primaryActionText: '提前结束交付',
      secondaryActionText: '取消并赔付',
      playerActionText: '联系玩家',
      guideActionText: '联系领路人'
    },
    {
      id: 'business-early-001',
      statusType: 'early',
      statusText: '已提前交付',
      ref: 'REF-20260318-004',
      avatarText: 'ZH',
      avatarClass: 'purple',
      name: '赵经理',
      serviceText: '技术咨询 · ¥600',
      guideText: '提前2天完成',
      guideTone: 'success',
      settlementRows: [
        { label: '实际服务时长', value: '1.5小时 (原定2小时)' },
        { label: '实际收入', value: '¥450 (按比例结算)', highlight: true }
      ],
      reviewStatus: 'pending',
      reviewActionText: '评价双方'
    },
    {
      id: 'business-complete-001',
      statusType: 'complete',
      statusText: '已完成',
      ref: 'REF-20260312-006',
      avatarText: 'WA',
      avatarClass: 'green',
      name: '王同学',
      serviceText: '品牌定位咨询 · ¥1,200',
      guideName: '陈引荐',
      completeSummary: '服务已完成',
      amountText: '¥1,200',
      completedAtText: '完成时间：03-15 18:30',
      actualDurationText: '2小时',
      actualIncomeText: '¥1,200',
      resultText: '双方已确认，收入已进入结算',
      reviewStatus: 'reviewed',
      reviewedActionText: '已评价'
    }
  ]
}

const mockPlayerGameManage = {
  currentTime: '2026-03-19T14:00:00+08:00',
  summary: {
    label: '本月服务支出',
    amount: 3200,
    amountText: '¥3,200',
    activeCount: 1,
    completedCount: 4,
    canceledCount: 1
  },
  orders: [
    {
      id: 'player-manage-active-001',
      statusType: 'active',
      statusText: '服务进行中',
      ref: 'REF-20260320-001',
      serviceOrderId: 'SO-20260320-001',
      fundAmount: 800,
      expert: {
        id: 'expert-zhang',
        name: '张专家',
        avatarText: 'ZH'
      },
      guide: {
        id: 'guide-wang',
        name: '王引荐'
      },
      serviceTitle: '产品架构咨询',
      startedAt: '2026-03-10T14:00:00+08:00',
      expectedDeliveryAt: '2026-03-25T14:00:00+08:00',
      noticeText: '取消需赔付一定比例金额给行家',
      canContactExpert: true,
      canCancel: true
    },
    {
      id: 'player-manage-complete-001',
      statusType: 'complete',
      statusText: '已完成',
      ref: 'REF-20260318-002',
      fundAmount: 1200,
      expert: {
        id: 'expert-wang',
        name: '王导师',
        avatarText: 'WM'
      },
      guide: {
        id: 'guide-chen',
        name: '陈引荐'
      },
      serviceTitle: '品牌定位咨询',
      completeSummary: '服务已完成',
      completedAtText: '完成时间：2026-03-19 18:30',
      resultText: '已完成验收，可查看服务记录',
      reviewStatus: 'pending',
      canReview: true,
      reviewActionText: '评价双方'
    },
    {
      id: 'player-manage-canceled-001',
      statusType: 'canceled',
      statusText: '已取消（已赔付）',
      ref: 'REF-20260310-003',
      compensationAmountText: '¥120',
      expert: {
        id: 'expert-liu',
        name: '刘设计师',
        avatarText: 'LI'
      },
      serviceTitle: 'UI设计服务',
      reasonSummary: '我主动取消 · 赔付15%',
      reasonText: '取消原因：需求变更，不再需要服务'
    }
  ]
}

module.exports = {
  validInvites,
  mockUser,
  mockCurrentUser,
  mockHome,
  mockRoleHomes,
  mockProfileHome,
  mockNewbieTasks,
  mockRoleApplications,
  mockInvitePlayerConfig,
  mockInvitePlayers,
  mockSystemRecommendations,
  mockReplayConfirmContext,
  mockGuideProgress,
  mockGuideCancelDetail,
  mockGameManage,
  mockPlayerGameManage
}
