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
  achievementList: [
    { id: 'hundred', title: '百场王者', statusText: '等级' },
    { id: 'guide', title: '引航王者', statusText: '等级' },
    { id: 'earth', title: '地球漫游者', statusText: '进度20%' }
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
    title: '共创数字街区｜全球联机互动',
    desc: '3D空间 · NFT徽章',
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

module.exports = {
  validInvites,
  mockUser,
  mockCurrentUser,
  mockHome,
  mockProfileHome,
  mockNewbieTasks,
  mockRoleApplications
}
