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
    { id: 'role', title: '申请身份', route: 'pages/role/apply/index' },
    { id: 'map', title: '附近组局', route: 'pages/map/index' },
    { id: 'message', title: '消息', route: 'pages/message/index' },
    { id: 'profile', title: '我的', route: 'pages/profile/index' }
  ],
  recommendedGames: [
    {
      id: '20001',
      title: '周末咖啡创业交流局',
      cityName: '上海',
      distanceText: '1.2km',
      memberText: '5-8人',
      statusText: '招募中',
      tags: ['创业', '咖啡', '同城']
    },
    {
      id: '20002',
      title: '城市夜景拍照路线',
      cityName: '上海',
      distanceText: '420m',
      memberText: '2-4人',
      statusText: '可预约',
      tags: ['路线', '摄影', '打卡']
    }
  ],
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
