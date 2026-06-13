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
      route: 'pages/game/detail/index',
      tags: ['城市故事', '探索', '同城']
    }
  ],
  playerSummary: {
    displayName: 'Alex Chen',
    roleLabel: '玩家 Lv.5',
    title: '探险家',
    xpText: '580/1000 XP',
    progressPercent: 58,
    nextLevelText: '距离下一等级还需 420 经验值',
    stats: [
      { label: '参与局数', value: '12' },
      { label: '本月MVP', value: '3' },
      { label: '参与率', value: '98%' }
    ]
  },
  rankingList: [
    { rank: '01', name: '领域专家 PRO', desc: '本周组局 12 · MVP 5次', xpText: '2,450 XP' },
    { rank: '02', name: '社交达人', desc: '本周组局 8 次', xpText: '1,890 XP' },
    { rank: '03', name: '探险家', desc: '本周组局 6 次', xpText: '1,560 XP' }
  ],
  achievementList: [
    { id: 'hundred', title: '百场王者', statusText: '等级' },
    { id: 'guide', title: '引航王者', statusText: '等级' },
    { id: 'earth', title: '地球漫游者', statusText: '进度20%' }
  ],
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
  mockRoleApplications
}
