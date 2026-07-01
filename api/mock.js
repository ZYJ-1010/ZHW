const {
  validInvites,
  mockUser,
  mockCurrentUser,
  mockHome,
  mockRoleHomes,
  mockProfileHome,
  mockSystemSkillConfig,
  mockPointsMall,
  mockPointsOrders,
  mockPointsOrderLogistics,
  mockNewbieTasks,
  mockRoleApplications,
  mockInvitePlayerConfig,
  mockInvitePlayers,
  mockSystemRecommendations,
  mockReplayConfirmContext,
  mockGuideProgress,
  mockGuideCancelDetail,
  mockGameManage,
  mockPlayerGameManage,
  mockGameProfitTemplates,
  mockRelationNetworkHome,
  mockMessageCenter,
  mockTradeWarningDetail,
  mockSystemNotificationDetail
} = require('./mock-data')

let mockCurrentRealnameStatus = 'pending'
let mockRoleStatusMap = Object.assign({}, mockCurrentUser.roleStatusMap)
let mockSubmittedRoleApplications = []
let mockPointsMallState = JSON.parse(JSON.stringify(mockPointsMall))
let mockPointsOrdersState = JSON.parse(JSON.stringify(mockPointsOrders))
let mockSystemProfileInfoState = null
let mockSystemSkillConfigState = JSON.parse(JSON.stringify(mockSystemSkillConfig))
const MOCK_ROLE_APPLICATION_SUBMITTED_AT = '2026-06-08T10:30:00+08:00'
const MOCK_ROLE_APPLICATION_EXPECTED_REVIEW_AT = '2026-06-12T18:00:00+08:00'

function ok(data) {
  return {
    code: 0,
    message: 'ok',
    data,
    requestId: `mock_${Date.now()}`
  }
}

function fail(code, message, data) {
  return {
    code,
    message,
    data: data || null,
    requestId: `mock_${Date.now()}`
  }
}

function wait(result) {
  return new Promise((resolve) => {
    setTimeout(() => resolve(result), 300)
  })
}

function loginWithWechat(payload) {
  const inviteCode = String(payload.inviteCode || '').trim().toUpperCase()
  const invite = inviteCode ? validInvites[inviteCode] : null

  if (inviteCode && !invite) {
    return wait(fail(40001, '邀请码无效'))
  }

  setMockRealnameStatus(inviteCode ? 'pending' : 'verified')

  return wait(ok({
    token: 'mock-token-enjoy-login',
    user: buildLoginUser(!inviteCode),
    isNewUser: Boolean(inviteCode),
    inviteRelation: buildInviteRelation(invite),
    code: payload.code || 'mock-wx-login-code'
  }))
}

function sendPhoneCode(payload) {
  const phone = String(payload.phone || '').trim()

  if (!/^1\d{10}$/.test(phone)) {
    return wait(fail(40002, '请输入正确手机号'))
  }

  return wait(ok({
    phone,
    code: '123456',
    expiresIn: 60
  }))
}

function verifyPhoneCode(payload) {
  const phone = String(payload.phone || '').trim()
  const code = String(payload.code || '').trim()

  if (!/^1\d{10}$/.test(phone) || code !== '123456') {
    return wait(fail(40005, '手机号或验证码错误'))
  }

  return wait(ok({
    phone,
    verified: true
  }))
}

function loginWithPhone(payload) {
  const phone = String(payload.phone || '').trim()
  const code = String(payload.code || '').trim()
  const inviteCode = String(payload.inviteCode || '').trim().toUpperCase()
  const invite = inviteCode ? validInvites[inviteCode] : null
  const isRegisteredUser = phone === '13888888888' || phone === '13900000000'
  const isRealnameVerified = phone === '13888888888'

  if (!/^1\d{10}$/.test(phone) || code !== '123456') {
    return wait(fail(40003, '手机号或验证码错误'))
  }

  if (!isRegisteredUser && !invite) {
    return wait(fail(40001, '邀请码无效'))
  }

  setMockRealnameStatus(isRealnameVerified ? 'verified' : 'pending')

  return wait(ok({
    token: 'mock-token-enjoy-phone-login',
    user: buildLoginUser(isRealnameVerified),
    loginType: 'phone',
    isNewUser: !isRegisteredUser,
    inviteRelation: buildInviteRelation(isRegisteredUser ? null : invite)
  }))
}

function loginWithPassword(payload) {
  const phone = String(payload.phone || '').trim()
  const password = String(payload.password || '')

  if (phone !== '13888888888' || password !== 'Test123456') {
    return wait(fail(40004, '账号或密码错误'))
  }

  setMockRealnameStatus('verified')

  return wait(ok({
    token: 'mock-token-enjoy-password-login',
    user: buildLoginUser(true),
    loginType: 'password'
  }))
}

function resetPassword(payload) {
  const phone = String(payload.phone || '').trim()
  const code = String(payload.code || '').trim()
  const password = String(payload.password || '')

  if (phone !== '13888888888' || code !== '123456') {
    return wait(fail(40005, '手机号或验证码错误'))
  }

  if (!/^(?=.*[A-Za-z])(?=.*\d)[A-Za-z\d]{8,20}$/.test(password)) {
    return wait(fail(40006, '仅支持字母和数字，长度8-20位'))
  }

  return wait(ok({
    phone,
    reset: true
  }))
}

function submitRealnameAuth(payload) {
  const realname = String(payload.realname || '').trim()
  const idNumber = String(payload.idNumber || '').trim().toUpperCase()

  if (!/^[\u4e00-\u9fa5A-Za-z·\s]{2,20}$/.test(realname) || !/(^\d{15}$)|(^\d{17}[\dX]$)/.test(idNumber)) {
    return wait(fail(40007, '实名认证失败，请重新核对后填写'))
  }

  setMockRealnameStatus('verified')

  return wait(ok({
    realnameStatus: 'verified',
    verified: true
  }))
}

function verifyInvite(code) {
  const invite = validInvites[code]

  if (!invite) {
    return wait(fail(40001, '邀请码无效'))
  }

  return wait(ok(invite))
}

function buildInviteRelation(invite) {
  if (!invite) {
    return null
  }

  return {
    inviteCode: invite.code,
    inviterUserId: invite.ownerUserId,
    inviterName: invite.inviterName,
    bound: true
  }
}

function buildLoginUser(isRealnameVerified) {
  return Object.assign({}, mockUser, {
    authStatus: isRealnameVerified ? 'verified' : 'pending',
    roleStatusMap: mockRoleStatusMap,
    defaultRole: mockUser.defaultRole || 'player',
    default_role: mockUser.default_role || mockUser.defaultRole || 'player',
    needRealname: !isRealnameVerified
  })
}

function buildCurrentUser() {
  const isRealnameVerified = mockCurrentRealnameStatus === 'verified' || getMockRealnameStorageVerified()

  return Object.assign({}, mockCurrentUser, {
    authStatus: isRealnameVerified ? 'verified' : 'pending',
    realnameStatus: isRealnameVerified ? 'verified' : 'pending',
    roleStatusMap: mockRoleStatusMap,
    needRealname: !isRealnameVerified,
    todoCounts: Object.assign({}, mockCurrentUser.todoCounts, {
      realname: isRealnameVerified ? 0 : 1,
      roleApplications: mockSubmittedRoleApplications.filter((item) => item.status === 'pending').length
    })
  })
}

function buildNewbieTaskSummary() {
  const isRealnameVerified = mockCurrentRealnameStatus === 'verified' || getMockRealnameStorageVerified()
  const tasks = mockNewbieTasks.map((task) => {
    const completed = task.type === 'realname' ? isRealnameVerified : Boolean(task.completed)

    return Object.assign({}, task, {
      completed,
      status: completed ? 'completed' : 'pending',
      statusText: completed ? '已完成' : task.actionText || '去完成',
      actionText: task.actionText || '去完成'
    })
  })
  const completedCount = tasks.filter((task) => task.completed).length
  const totalCount = tasks.length

  return {
    completedCount,
    totalCount,
    progressPercent: totalCount ? Math.round((completedCount / totalCount) * 100) : 0,
    tasks
  }
}

function buildHome(data) {
  const roleType = String(data && data.roleType || 'player').trim()

  return Object.assign({}, mockRoleHomes[roleType] || mockHome, {
    user: buildCurrentUser(),
    roleApplications: mockSubmittedRoleApplications
  })
}

function buildRoleApplications() {
  return mockRoleApplications.map((item) => {
    const submitted = mockSubmittedRoleApplications.find((application) => application.roleType === item.roleType)

    if (!submitted) {
      return item
    }

    return Object.assign({}, item, submitted)
  })
}

function submitMockRoleApplication(data = {}) {
  const roleType = data.roleType || 'expert'
  const application = {
    applicationId: `mock-role-${Date.now()}`,
    roleType,
    status: 'pending',
    statusText: '已进入审核',
    submittedAt: MOCK_ROLE_APPLICATION_SUBMITTED_AT,
    expectedReviewAt: MOCK_ROLE_APPLICATION_EXPECTED_REVIEW_AT,
    submitted: data
  }

  mockRoleStatusMap = Object.assign({}, mockRoleStatusMap, {
    [roleType]: 'pending'
  })
  mockSubmittedRoleApplications = mockSubmittedRoleApplications
    .filter((item) => item.roleType !== roleType)
    .concat(application)

  return application
}

function buildInvitePlayers(data = {}, recentOnly = false) {
  const keyword = String(data.keyword || data.search || '').trim().toLowerCase()
  const source = recentOnly ? mockInvitePlayers.slice(0, 3) : mockInvitePlayers
  const list = keyword
    ? source.filter((item) => {
      const text = `${item.name || ''} ${item.desc || ''} ${item.meta || ''}`.toLowerCase()

      return text.includes(keyword)
    })
    : source

  return {
    list,
    total: list.length
  }
}

function buildSystemRecommendations(data = {}) {
  const category = String(data.category || 'all').trim()
  const source = mockSystemRecommendations || {}
  const experts = Array.isArray(source.experts) ? source.experts : []
  const list = category && category !== 'all'
    ? experts.filter((item) => item.category === category)
    : experts

  return Object.assign({}, source, {
    activeCategory: category || 'all',
    experts: list
  })
}

function buildTradeWarningDetail(data = {}) {
  const detail = JSON.parse(JSON.stringify(mockTradeWarningDetail))
  const warningId = data.warningId || data.id || detail.warningId
  const orderId = data.orderId || detail.order.orderNo

  return Object.assign({}, detail, {
    id: warningId,
    warningId,
    order: Object.assign({}, detail.order, {
      orderNo: orderId
    })
  })
}

function buildSystemNotificationDetail(data = {}) {
  const detail = JSON.parse(JSON.stringify(mockSystemNotificationDetail))
  const notificationId = data.notificationId || data.messageId || data.id || detail.notificationId

  return Object.assign({}, detail, {
    id: notificationId,
    messageId: notificationId,
    notificationId
  })
}

function buildMessageCenter(data = {}) {
  const center = JSON.parse(JSON.stringify(mockMessageCenter))
  const activeTab = data.tab || data.activeTab || center.activeTab || 'all'

  return Object.assign({}, center, {
    activeTab
  })
}

function formatNumber(value) {
  const number = Number(value) || 0

  return number.toLocaleString('en-US')
}

function buildPointsMall() {
  const pointsAvailable = Number(mockPointsMallState.pointsAvailable) || 0
  const goods = (mockPointsMallState.goods || []).map((item) => Object.assign({}, item, {
    points: `${formatNumber(item.cost)}积分`,
    pointsText: `${formatNumber(item.cost)}积分`,
    stock: `库存: 剩余${Number(item.stockLeft) || 0}件`,
    stockText: `库存: 剩余${Number(item.stockLeft) || 0}件`,
    remaining: formatNumber(Math.max(pointsAvailable - (Number(item.cost) || 0), 0))
  }))

  return {
    pointsAvailable,
    points: formatNumber(pointsAvailable),
    pointsText: formatNumber(pointsAvailable),
    expireTip: mockPointsMallState.expireTip,
    goods
  }
}

function buildPointsOrders(data = {}) {
  const status = String(data.status || data.statusKey || data.orderStatus || '').trim()
  const source = JSON.parse(JSON.stringify(mockPointsOrdersState))
  const orders = status
    ? source.orders.filter((item) => item.statusKey === status)
    : source.orders

  return Object.assign({}, source, {
    activeStatus: status || 'all',
    orders,
    emptyText: status ? '暂无该状态订单' : '暂无订单'
  })
}

function buildPointsOrderLogistics(orderId) {
  const detail = mockPointsOrderLogistics[orderId]

  if (detail) {
    return JSON.parse(JSON.stringify(detail))
  }

  return {
    orderId,
    courier: {
      name: '顺丰速运',
      trackingNo: 'SF1234567890',
      logoText: 'SF'
    },
    timeline: [
      {
        id: 'created',
        desc: '商家已创建物流单，等待揽收',
        time: '刚刚',
        active: true
      }
    ]
  }
}

function exchangePointsMallGood(data = {}) {
  const goodId = String(data.goodId || data.productId || data.id || '').trim()

  if (!goodId) {
    return wait(fail(40001, '请选择兑换商品', buildPointsMall()))
  }

  const good = (mockPointsMallState.goods || []).find((item) => item.id === goodId)

  if (!good) {
    return wait(fail(40404, '商品不存在或已下架', buildPointsMall()))
  }

  const stockLeft = Number(good.stockLeft) || 0
  const cost = Number(good.cost) || 0
  const pointsAvailable = Number(mockPointsMallState.pointsAvailable) || 0

  if (stockLeft <= 0) {
    return wait(fail(40902, '库存不足，请稍后再试', buildPointsMall()))
  }

  if (pointsAvailable < cost) {
    return wait(fail(40901, '积分不足，无法兑换', buildPointsMall()))
  }

  good.stockLeft = stockLeft - 1
  mockPointsMallState.pointsAvailable = pointsAvailable - cost
  const orderId = `POINTS-${Date.now()}`

  mockPointsOrdersState.orders = [
    {
      id: orderId,
      statusKey: 'pending_ship',
      statusText: '待发货',
      statusTone: 'orange',
      iconText: good.iconText || '🎁',
      title: good.title || '兑换商品',
      pointsText: `${formatNumber(cost)}积分`,
      exchangedAtText: '兑换时间: 刚刚',
      actions: [
        { key: 'cancel', label: '取消订单', type: 'ghost' }
      ]
    }
  ].concat(mockPointsOrdersState.orders || [])

  return wait(ok({
    exchangeSuccess: true,
    message: '兑换成功，订单已进入待发货',
    orderId,
    orderStatus: 'pending_ship',
    mall: buildPointsMall()
  }))
}

function saveSystemProfileInfo(data = {}) {
  const personalInfo = data.personalInfo || {}
  const enterpriseInfo = data.enterpriseInfo || {}
  const contactVisibility = String(personalInfo.contactVisibility || '').trim()
  const validVisibility = ['all', 'member', 'hidden'].indexOf(contactVisibility) !== -1

  if (!validVisibility) {
    return wait(fail(40001, '联系方式可见性设置错误'))
  }

  mockSystemProfileInfoState = {
    id: `profile-info-${Date.now()}`,
    personalInfo,
    enterpriseInfo,
    certifications: Array.isArray(data.certifications) ? data.certifications : [],
    savedAt: '2026-06-27T00:00:00+08:00'
  }

  return wait(ok({
    saved: true,
    profileInfo: mockSystemProfileInfoState,
    message: '资料保存成功'
  }))
}

function buildSystemSkillConfig() {
  return JSON.parse(JSON.stringify(mockSystemSkillConfigState))
}

function saveSystemSkillConfig(data = {}) {
  const roleSummary = data.roleSummary || {}
  const skillSlots = Array.isArray(data.skillSlots) ? data.skillSlots : []
  const skillGroups = data.skillGroups && typeof data.skillGroups === 'object' ? data.skillGroups : {}

  if (!skillSlots.length) {
    return wait(fail(40001, '至少需要保留一个技能槽位'))
  }

  mockSystemSkillConfigState = Object.assign({}, mockSystemSkillConfigState, {
    roleSummary,
    activeTab: data.activeTab || 'visible',
    skillSlots,
    skillGroups,
    unlockSuggestion: data.unlockSuggestion || mockSystemSkillConfigState.unlockSuggestion,
    savedAt: '2026-06-27T00:00:00+08:00'
  })

  return wait(ok({
    saved: true,
    skillConfig: buildSystemSkillConfig(),
    message: '技能配置保存成功'
  }))
}

function toFiniteNumber(value, fallback) {
  const number = Number(value)

  return Number.isFinite(number) ? number : fallback
}

function getMockGameMembers() {
  return [
    {
      id: 'member-20001-01',
      userId: 'user-luyi',
      name: '陆毅',
      avatarUrl: '/pages/home/player/assets/ranking-avatar-01.png',
      roleText: '玩家',
      roleClass: 'player',
      title: '总经理 | 上海创世界科技有限公司',
      summary: 'AI赋能与市场运营助力企业IP打造',
      primaryTag: '第一标签：上海TMT投资领军者',
      tags: ['数字化内容服务'],
      location: '上海市浦东新区沙新镇黄赵路310号',
      distanceText: '2.1 km'
    },
    {
      id: 'member-20001-02',
      userId: 'user-linyi',
      name: '林一',
      avatarUrl: '/pages/home/player/assets/ranking-avatar-02.png',
      roleText: '行家',
      roleClass: 'expert',
      title: '品牌创始人 | 杭州欣悦服装工作',
      summary: '企业家服务平台',
      primaryTag: '第一标签：女性高品质服装领先者',
      tags: ['品牌增长', '企业服务'],
      location: '杭州市上城区',
      distanceText: '2.1 km'
    }
  ]
}

function getMockGameList() {
  return (mockHome.recommendedGames || []).map((item) => Object.assign({}, item, {
    gameType: item.gameType || item.type || 'social',
    gameTypeText: item.gameTypeText || item.statusText,
    addressName: item.addressName || item.cityName,
    approvedMemberCount: item.approvedMemberCount || 3,
    maxParticipants: item.maxParticipants || 8,
    coverFileUrl: item.coverUrl,
    themeTags: item.themeTags || item.tags
  }))
}

function buildGameList(data = {}) {
  const keyword = String(data.keyword || '').trim()
  let list = getMockGameList()

  if (keyword) {
    list = list.filter((item) => String(item.title || '').indexOf(keyword) !== -1)
  }

  return {
    page: Number(data.page) || 1,
    pageSize: Number(data.pageSize) || 20,
    total: list.length,
    list
  }
}

function buildGameDetail(gameId) {
  const matched = getMockGameList().find((item) => String(item.id) === String(gameId)) || getMockGameList()[0] || {}

  return Object.assign({}, matched, {
    id: gameId || matched.id,
    serverTime: mockPlayerGameManage.currentTime,
    gameTypeText: matched.gameTypeText || matched.statusText,
    addressName: matched.addressName || matched.cityName,
    feeText: matched.priceText,
    viewCount: 1234,
    commentCount: 3,
    approvedMemberCount: 5,
    maxParticipants: 8,
    creator: {
      id: 'creator-001',
      name: '陆毅',
      avatarUrl: '/pages/home/player/assets/ranking-avatar-01.png',
      title: '总经理 | 上海创世界科技有限公司',
      summary: '已组局 88次 · 推荐20人',
      rating: '4.8'
    },
    introduction: '本场活动围绕主题交流、资源对接和现场协作展开，具体内容由后台活动配置返回。',
    highlights: ['主题分享', '自由交流', '资源对接'],
    schedule: [
      { title: '签到入场', time: '08:30-08:50', desc: '完成签到并熟悉现场。' },
      { title: '主题交流', time: '08:50-10:20', desc: '围绕活动主题进行分享和讨论。' }
    ],
    noticeLead: '请按活动要求准时到场，并遵守现场秩序。',
    noticeBullets: ['报名成功后请提前确认行程。', '如需取消，请提前联系发起人。'],
    audience: '活动报名通过用户',
    members: getMockGameMembers()
  })
}

function buildNearbyGames(data = {}) {
  const centerLatitude = toFiniteNumber(data.latitude || data.lat, 31.2304)
  const centerLongitude = toFiniteNumber(data.longitude || data.lng, 121.4737)
  const radiusMeters = toFiniteNumber(data.radiusMeters || data.radius, 3000)
  const nearbyGames = [
    {
      id: 'map-nearby-001',
      title: '鱼尾狮夜景打卡点',
      cityName: '海尚广场',
      latitude: centerLatitude + 0.0022,
      longitude: centerLongitude + 0.0016,
      distanceText: '420m',
      memberText: '3/6人',
      timeText: '今晚 20:00',
      statusText: '探索局',
      priceText: '¥0/人',
      route: 'pages/game/detail/index'
    },
    {
      id: 'map-nearby-002',
      title: '3点路线盲盒: 港湾微风版',
      cityName: '滨江步道',
      latitude: centerLatitude - 0.0018,
      longitude: centerLongitude + 0.0026,
      distanceText: '860m',
      memberText: '2/4人',
      timeText: '明天 15:30',
      statusText: '路线局',
      priceText: '¥29/人',
      route: 'pages/game/detail/index'
    },
    {
      id: 'map-nearby-003',
      title: '苏州河记忆碎片采集',
      cityName: '桥下空间',
      latitude: centerLatitude + 0.001,
      longitude: centerLongitude - 0.0028,
      distanceText: '1.2km',
      memberText: '5/8人',
      timeText: '周六 19:00',
      statusText: '任务局',
      priceText: '¥0/人',
      route: 'pages/game/detail/index'
    }
  ]
  const onlinePlayers = [
    {
      id: 'map-online-player-001',
      latitude: centerLatitude + 0.0018,
      longitude: centerLongitude + 0.0008,
      statusText: '在线玩家'
    },
    {
      id: 'map-online-player-002',
      latitude: centerLatitude - 0.0012,
      longitude: centerLongitude + 0.0019,
      statusText: '在线玩家'
    },
    {
      id: 'map-online-player-003',
      latitude: centerLatitude + 0.0028,
      longitude: centerLongitude - 0.0017,
      statusText: '在线玩家'
    }
  ]
  const offlinePlayers = [
    {
      id: 'map-offline-player-001',
      latitude: centerLatitude - 0.0022,
      longitude: centerLongitude - 0.0021,
      statusText: '离线玩家'
    },
    {
      id: 'map-offline-player-002',
      latitude: centerLatitude + 0.0006,
      longitude: centerLongitude + 0.0032,
      statusText: '离线玩家'
    }
  ]

  return {
    center: {
      latitude: centerLatitude,
      longitude: centerLongitude
    },
    radiusMeters,
    nearestDistanceText: '157m',
    onlinePlayerCount: 23,
    offlinePlayerCount: 8,
    total: nearbyGames.length,
    list: nearbyGames,
    onlinePlayers,
    offlinePlayers
  }
}

function respondGameInvitation(data = {}) {
  const action = String(data.action || '').trim()

  if (action !== 'accept' && action !== 'reject') {
    return wait(fail(40001, '邀约处理动作错误'))
  }

  return wait(ok({
    invitationStatus: action === 'accept' ? 'accepted' : 'rejected',
    applicationStatus: action === 'accept' ? 'pending' : null
  }))
}

function createGamePayment(data = {}) {
  const amount = Number(data.amount || 0)
  const splits = Array.isArray(data.splits) ? data.splits : []

  if (!data.agreementChecked) {
    return wait(fail(40002, '请先同意押金局规则'))
  }

  if (!amount || amount <= 0) {
    return wait(fail(40003, '支付金额错误'))
  }

  return wait(ok({
    paymentOrderId: `mock_game_payment_${Date.now()}`,
    gameId: data.gameId || '',
    scene: data.scene || 'deposit_game',
    payChannel: data.payChannel || 'wechat',
    amount,
    currency: data.currency || 'CNY',
    splits,
    mockPayment: true
  }))
}

function setMockRealnameStatus(status) {
  mockCurrentRealnameStatus = status

  if (typeof wx === 'undefined') {
    return
  }

  if (status === 'verified') {
    wx.setStorageSync('enjoy_mock_realname_verified', '1')
    return
  }

  wx.removeStorageSync('enjoy_mock_realname_verified')
}

function getMockRealnameStorageVerified() {
  if (typeof wx === 'undefined') {
    return false
  }

  return wx.getStorageSync('enjoy_mock_realname_verified') === '1'
}

function handleRequest(options) {
  const method = options.method || 'GET'
  const url = options.url

  if (method === 'POST' && url === '/api/app/auth/wechat-login') {
    return loginWithWechat(options.data || {})
  }

  if (method === 'POST' && url === '/api/app/auth/phone-code') {
    return sendPhoneCode(options.data || {})
  }

  if (method === 'POST' && url === '/api/app/auth/phone-code/verify') {
    return verifyPhoneCode(options.data || {})
  }

  if (method === 'POST' && url === '/api/app/auth/phone-login') {
    return loginWithPhone(options.data || {})
  }

  if (method === 'POST' && url === '/api/app/auth/password-login') {
    return loginWithPassword(options.data || {})
  }

  if (method === 'POST' && url === '/api/app/auth/password/reset') {
    return resetPassword(options.data || {})
  }

  if (method === 'POST' && url === '/api/app/invites/verify') {
    return verifyInvite(String(options.data && options.data.code || '').trim().toUpperCase())
  }

  if (method === 'GET' && url === '/api/app/users/me') {
    return wait(ok(buildCurrentUser()))
  }

  if (method === 'POST' && url === '/api/app/users/me/realname-auth') {
    return wait(ok({
      url: '/pages/login/realname/index',
      provider: 'mock'
    }))
  }

  if (method === 'POST' && url === '/api/app/users/me/realname-auth/submit') {
    return submitRealnameAuth(options.data || {})
  }

  if (method === 'GET' && url === '/api/app/newbie-tasks') {
    return wait(ok(buildNewbieTaskSummary()))
  }

  if (method === 'GET' && url === '/api/app/home') {
    return wait(ok(buildHome(options.data || {})))
  }

  if (method === 'GET' && url === '/api/app/games') {
    return wait(ok(buildGameList(options.data || {})))
  }

  if (method === 'GET' && /^\/api\/app\/games\/[^/]+\/members$/.test(url)) {
    const gameId = url.split('/')[4]

    return wait(ok({
      page: Number(options.data && options.data.page) || 1,
      pageSize: Number(options.data && options.data.pageSize) || 100,
      total: getMockGameMembers().length,
      list: getMockGameMembers().map((item) => Object.assign({}, item, {
        gameId
      }))
    }))
  }

  if (method === 'GET' && url === '/api/app/games/nearby') {
    return wait(ok(buildNearbyGames(options.data || {})))
  }

  if (method === 'GET' && /^\/api\/app\/games\/[^/]+$/.test(url)) {
    return wait(ok(buildGameDetail(url.split('/')[4])))
  }

  if (method === 'GET' && url === '/api/app/relations/network-home') {
    return wait(ok(mockRelationNetworkHome))
  }

  if (method === 'GET' && url === '/api/app/messages/trade-warning') {
    return wait(ok(buildTradeWarningDetail(options.data || {})))
  }

  if (method === 'GET' && url === '/api/app/messages/center') {
    return wait(ok(buildMessageCenter(options.data || {})))
  }

  if (method === 'GET' && url === '/api/app/messages/system-notification') {
    return wait(ok(buildSystemNotificationDetail(options.data || {})))
  }

  if (method === 'GET' && /^\/api\/im\/rooms\/[^/]+\/messages$/.test(url)) {
    return wait(ok({
      page: Number(options.data && options.data.page) || 1,
      pageSize: Number(options.data && options.data.pageSize) || 50,
      total: 0,
      list: []
    }))
  }

  if (method === 'POST' && /^\/api\/im\/rooms\/[^/]+\/messages$/.test(url)) {
    return wait(ok({
      messageId: `mock-message-${Date.now()}`,
      type: options.data && options.data.type || 'text',
      content: options.data && options.data.content || ''
    }))
  }

  if (method === 'GET' && url === '/api/app/profile/home') {
    return wait(ok(mockProfileHome))
  }

  if (method === 'GET' && url === '/api/app/profile/assets') {
    return wait(ok({
      overview: {},
      assetStats: [],
      menuItems: [],
      orderStatuses: [],
      recentOrders: [],
      faqLinks: []
    }))
  }

  if (method === 'GET' && url === '/api/app/profile/points') {
    return wait(ok({
      summary: {},
      rules: [],
      earnExample: {},
      roleExamples: [],
      records: []
    }))
  }

  if (method === 'GET' && url === '/api/app/profile/games') {
    return wait(ok({
      list: [],
      statusTabs: []
    }))
  }

  if (method === 'GET' && url === '/api/app/profile/service-center/reviews') {
    return wait(ok({
      score: {},
      stats: [],
      reviews: [],
      pendingCount: 0
    }))
  }

  if (method === 'GET' && /^\/api\/app\/profile\/service-center\/reviews\/[^/]+$/.test(url)) {
    return wait(ok({
      review: {},
      templates: [],
      history: []
    }))
  }

  if (method === 'GET' && url === '/api/app/profile/invite/overview') {
    return wait(ok({
      user: {},
      stats: [],
      cards: [],
      actions: []
    }))
  }

  if (method === 'GET' && url === '/api/app/profile/invite-records') {
    return wait(ok({
      list: [],
      tabs: []
    }))
  }

  if (method === 'GET' && url === '/api/app/profile/invite/income') {
    return wait(ok({
      stats: [],
      records: []
    }))
  }

  if (method === 'GET' && url === '/api/app/profile/invite/network') {
    return wait(ok({
      stats: [],
      members: []
    }))
  }

  if (method === 'GET' && url === '/api/app/profile/invite/ranking') {
    return wait(ok({
      list: [],
      tabs: []
    }))
  }

  if (method === 'GET' && /^\/api\/app\/profile\/invite\/members\/[^/]+$/.test(url)) {
    return wait(ok({
      member: {},
      stats: [],
      records: []
    }))
  }

  if (method === 'GET' && url === '/api/app/profile/system-management/skill-config') {
    return wait(ok(buildSystemSkillConfig()))
  }

  if (method === 'GET' && url === '/api/app/profile/points/mall') {
    return wait(ok(buildPointsMall()))
  }

  if (method === 'GET' && url === '/api/app/profile/points/orders') {
    return wait(ok(buildPointsOrders(options.data || {})))
  }

  if (method === 'GET' && /^\/api\/app\/profile\/points\/orders\/[^/]+\/logistics$/.test(url)) {
    const orderId = decodeURIComponent(url.split('/').slice(-2)[0] || '')

    if (!orderId) {
      return wait(fail(40001, '缺少订单信息'))
    }

    return wait(ok(buildPointsOrderLogistics(orderId)))
  }

  if (method === 'POST' && url === '/api/app/profile/points/mall/exchange') {
    return exchangePointsMallGood(options.data || {})
  }

  if (method === 'PUT' && url === '/api/app/profile/system-management/profile-info') {
    return saveSystemProfileInfo(options.data || {})
  }

  if (method === 'PUT' && url === '/api/app/profile/system-management/skill-config') {
    return saveSystemSkillConfig(options.data || {})
  }

  if (method === 'POST' && /^\/api\/app\/profile\/service-center\/reviews\/[^/]+\/reply$/.test(url)) {
    const reviewId = decodeURIComponent(url.split('/').slice(-2)[0] || '')
    const content = String(options.data && options.data.content || '').trim()

    if (!reviewId) {
      return wait(fail(40001, '缺少评价信息'))
    }

    if (!content) {
      return wait(fail(40002, '请输入回复内容'))
    }

    if (content.length > 200) {
      return wait(fail(40003, '回复内容不能超过200字'))
    }

    return wait(ok({
      reviewId,
      content,
      replied: true,
      repliedAt: '2026-06-27T00:00:00+08:00'
    }))
  }

  if (method === 'GET' && url === '/api/app/role-applications/my') {
    return wait(ok(buildRoleApplications()))
  }

  if (method === 'GET' && url === '/api/app/roles/my') {
    return wait(ok(buildCurrentUser()))
  }

  if (method === 'GET' && url === '/api/app/role-applications/expert/config') {
    return wait(ok({
      skillOptions: [
        { name: '摄影', active: true },
        { name: '户外', active: false },
        { name: '美食', active: false },
        { name: '文化', active: false },
        { name: '手工', active: false },
        { name: '运动', active: false },
        { name: '音乐', active: false },
        { name: '+ 自定义', custom: true }
      ],
      fields: [
        { type: 'chips', key: 'skillDomain', label: '选择技能领域', required: true },
        {
          type: 'input',
          key: 'skillTags',
          label: '技能标签',
          required: true,
          placeholder: '如：人像摄影、风光摄影、夜景拍摄',
          helper: '添加具体标签，让用户更容易找到你',
          maxlength: 30
        },
        { type: 'select', key: 'experienceYears', label: '从业年限', required: true, placeholder: '请选择从业年限' },
        {
          type: 'textarea',
          key: 'intro',
          label: '个人简介',
          required: true,
          placeholder: '介绍你的专业背景、服务风格、擅长领域...',
          helper: '不少于 50 字，突出你的专业优势',
          maxlength: 300
        }
      ],
      uploadField: {
        label: '资质证明',
        required: true,
        icon: '📎',
        title: '点击上传作品集及凭证',
        acceptTypes: ['JPG', 'PNG', 'PDF'],
        maxCount: 5
      },
      validationRules: {
        skillTags: { minLength: 2, maxLength: 30 },
        intro: { minLength: 50, maxLength: 300 },
        serviceName: { minLength: 2, maxLength: 20 },
        customSkill: { minLength: 2, maxLength: 8 },
        money: { integerMaxLength: 8, decimalMaxLength: 2 }
      },
      yearOptions: Array.from({ length: 30 }, (_, index) => `${index + 1}年`).concat('30年以上'),
      serviceCount: 3,
      priceHint: '平台将收取 10% 服务费'
    }))
  }

  if (method === 'POST' && url === '/api/app/role-applications') {
    return wait(ok(submitMockRoleApplication(options.data || {})))
  }

  if (method === 'GET' && url === '/api/app/game-invites/player-config') {
    return wait(ok(mockInvitePlayerConfig))
  }

  if (method === 'GET' && url === '/api/app/game-invites/recent-players') {
    return wait(ok(buildInvitePlayers(options.data || {}, true)))
  }

  if (method === 'GET' && url === '/api/app/game-invites/players') {
    return wait(ok(buildInvitePlayers(options.data || {}, false)))
  }

  if (method === 'GET' && url === '/api/app/game-invites/replay-context') {
    return wait(ok(Object.assign({}, mockReplayConfirmContext, {
      sourceGameId: options.data && (options.data.sourceGameId || options.data.gameId) || mockReplayConfirmContext.sourceGameId,
      serviceOrderId: options.data && options.data.serviceOrderId || mockReplayConfirmContext.serviceOrderId
    })))
  }

  if (method === 'GET' && url === '/api/app/game-invites/system-recommendations') {
    return wait(ok(buildSystemRecommendations(options.data || {})))
  }

  if (method === 'POST' && url === '/api/app/game-invites/replay') {
    return wait(ok({
      replayInvitationId: `mock_replay_invite_${Date.now()}`,
      status: 'pending',
      statusText: '等待双方确认',
      sourceGameId: options.data && options.data.sourceGameId || '',
      serviceOrderId: options.data && options.data.serviceOrderId || '',
      invitees: options.data && options.data.invitees || []
    }))
  }

  if (method === 'GET' && url === '/api/app/game-invites/guide-progress') {
    return wait(ok(mockGuideProgress))
  }

  if (method === 'GET' && url === '/api/app/game-invites/guide-cancel-detail') {
    const data = Object.assign({}, mockGuideCancelDetail)

    if (String(options.data && (options.data.cancelRole || options.data.role) || '').toLowerCase() === 'expert') {
      data.statusDesc = '行家取消了此次组局邀请'
      data.canceledBy = {
        id: 'expert-wangqiang',
        name: '王强',
        roleType: 'expert',
        roleLabel: '行家',
        avatarText: 'WQ',
        avatarClass: 'expert'
      }
      data.reason = {
        title: '档期冲突',
        desc: '行家临时档期调整，无法按时参加'
      }
      data.message = '抱歉，临时档期有冲突，本次无法继续参加。辛苦帮忙协调，下次有机会再合作。'
      data.timeline = [
        {
          key: 'invite',
          title: '发起邀请',
          desc: '你向双方发送了组局邀请',
          timeText: '03-21 10:23',
          state: 'active'
        },
        {
          key: 'cancel',
          title: '行家取消',
          desc: '王强因档期冲突取消本次组局',
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

    return wait(ok(data))
  }

  if (method === 'GET' && url === '/api/app/games/my/manage') {
    return wait(ok(mockGameManage))
  }

  if (method === 'GET' && url === '/api/app/games/player/manage') {
    return wait(ok(mockPlayerGameManage))
  }

  if (method === 'GET' && url === '/api/app/games/profit-templates') {
    return wait(ok(mockGameProfitTemplates))
  }

  if (method === 'POST' && url === '/api/app/game-payments/wechat') {
    return createGamePayment(options.data || {})
  }

  if (method === 'POST' && /^\/api\/app\/game-invitations\/[^/]+\/respond$/.test(url)) {
    return respondGameInvitation(options.data || {})
  }

  return wait(fail(40401, `mock 未配置接口：${method} ${url}`))
}

module.exports = {
  handleRequest,
  loginWithWechat,
  sendPhoneCode,
  verifyPhoneCode,
  loginWithPhone,
  loginWithPassword,
  resetPassword,
  submitRealnameAuth,
  verifyInvite,
  createGamePayment
}
