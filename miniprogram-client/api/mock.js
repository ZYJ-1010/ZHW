const {
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
} = require('./mock-data')

let mockCurrentRealnameStatus = 'pending'
let mockRoleStatusMap = Object.assign({}, mockCurrentUser.roleStatusMap)
let mockSubmittedRoleApplications = []
let mockMemberRadarState = {
  savedProfile: {},
  lastMatch: {},
  followed: []
}
let mockPointsMallState = JSON.parse(JSON.stringify(mockPointsMall))
let mockPointsOrdersState = JSON.parse(JSON.stringify(mockPointsOrders))
let mockPointsLogsState = [
  {
    id: 'points-log-report-001',
    userId: 1,
    changeValue: 20,
    beforePoints: Number(mockPointsMall.pointsAvailable || 0) - 20,
    afterPoints: Number(mockPointsMall.pointsAvailable || 0),
    bizType: 'report_reward',
    bizId: 1,
    reason: '举报核实奖励',
    createdAt: '2026-06-30T10:00:00+08:00'
  }
]
let mockSystemProfileInfoState = null
let mockSystemSkillConfigState = JSON.parse(JSON.stringify(mockSystemSkillConfig))
let mockProfileSettingsState = null
let mockSystemBlockSettingsState = null
let mockAgreementState = null
let mockFeedbackRecords = []
let mockServiceReviewsState = [
  {
    id: 'review-001',
    gameId: 1,
    reviewerUserId: 201,
    targetUserId: 1,
    user: '用户A',
    avatar: 'UA',
    avatarClass: 'green',
    time: '03-25',
    timeText: '03-25 14:30',
    rating: '★★★★★',
    score: 5,
    title: '产品架构咨询',
    content: '行家非常专业，沟通顺畅，响应及时，强烈推荐。',
    tags: ['专业能力强', '交付及时', '沟通顺畅'],
    reply: '感谢认可，期待下次合作。',
    statusType: 'replied',
    orderNo: 'ORD-20260325-001',
    amount: '¥800',
    detailUrl: '/pages/game/detail/index?id=1',
    reportUrl: '/pages/profile/system-management/report-center/index?gameId=1&targetUserId=201&targetName=用户A&reviewId=review-001'
  },
  {
    id: 'review-002',
    gameId: 1,
    reviewerUserId: 202,
    targetUserId: 1,
    user: '用户B',
    avatar: 'UB',
    avatarClass: 'green',
    time: '03-24',
    timeText: '03-24 18:20',
    rating: '★★★★★',
    score: 5,
    title: 'UI设计服务',
    content: '设计质量很高，修改响应也很快，整体体验很好。',
    tags: ['超出预期', '性价比高'],
    reply: '',
    statusType: 'pending',
    orderNo: 'ORD-20260324-002',
    amount: '¥600',
    detailUrl: '/pages/game/detail/index?id=1',
    reportUrl: '/pages/profile/system-management/report-center/index?gameId=1&targetUserId=202&targetName=用户B&reviewId=review-002'
  },
  {
    id: 'review-003',
    gameId: 1,
    reviewerUserId: 203,
    targetUserId: 1,
    user: '用户C',
    avatar: 'UC',
    avatarClass: 'indigo',
    time: '03-20',
    timeText: '03-20 10:05',
    rating: '★★★☆☆',
    score: 3,
    title: '技术架构咨询',
    content: '整体还可以，但是沟通效率有待提升。',
    tags: [],
    reply: '',
    statusType: 'neutral',
    orderNo: 'ORD-20260320-003',
    amount: '¥500',
    detailUrl: '/pages/game/detail/index?id=1',
    reportUrl: '/pages/profile/system-management/report-center/index?gameId=1&targetUserId=203&targetName=用户C&reviewId=review-003'
  }
]
let mockCreatedGames = []
let mockReceivedGameApplications = [
  {
    id: 1001,
    gameId: 1,
    userId: 201,
    userName: '极客少女小夏',
    roleKey: 'player',
    status: 'pending',
    reason: '我很想一起把这个局玩成。',
    createdAt: '2026-03-30 11:20'
  }
]
let mockSubmittedReviews = []
let mockSubmittedReports = []
let mockCreditLogs = [
  { id: 1, gameId: 1, reason: 'low_review', title: '低分评价', desc: '低分评价 · 局ID 1', score: '-5', tone: 'minus', createdAt: new Date().toISOString() }
]
let mockUploadedFiles = []
let mockSystemNotificationFeedback = {
  useful: 128,
  useless: 10
}
const MOCK_ROLE_APPLICATION_SUBMITTED_AT = '2026-06-08T10:30:00+08:00'
const MOCK_ROLE_APPLICATION_EXPECTED_REVIEW_AT = '2026-06-12T18:00:00+08:00'
const mockIMMessagesByGame = {
  1: [
    {
      id: 1,
      roomId: 1,
      gameId: 1,
      senderUserId: 1,
      messageType: 'text',
      content: '欢迎进入局内 IM，会话由 OpenIM 能力托底，业务权限由后端控制。',
      status: 'sent',
      createdAt: '2026-06-29T10:00:00+08:00'
    }
  ]
}

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
    authPageMode: inviteCode ? 'register' : 'login',
    boundWechat: !inviteCode,
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
    code: '000000',
    expiresIn: 60
  }))
}

function verifyPhoneCode(payload) {
  const phone = String(payload.phone || '').trim()
  const code = String(payload.code || '').trim()

  if (!/^1\d{10}$/.test(phone) || code !== '000000') {
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

  if (!/^1\d{10}$/.test(phone) || code !== '000000') {
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

  if (phone !== '13888888888' || code !== '000000') {
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
  const realname = String(payload.realName || payload.realname || '').trim()
  const idNumber = String(payload.idCard || payload.idNumber || '').trim().toUpperCase()

  if (!/^[\u4e00-\u9fa5A-Za-z·\s]{2,20}$/.test(realname) || !/^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dX]$/.test(idNumber)) {
    return wait(fail(40007, '实名认证失败，请重新核对后填写'))
  }

  setMockRealnameStatus('pending')

  return wait(ok({
    status: 'pending',
    realnameStatus: 'pending',
    verified: false
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

function createMockUploadToken(data = {}) {
  const fileName = String(data.fileName || '').trim()
  const bizType = String(data.bizType || 'report_attachment').trim()
  const mimeType = String(data.mimeType || 'image/jpeg').trim()
  const size = Number(data.size || 0)

  if (!fileName || !mimeType || size <= 0) {
    return wait(fail(42208, '文件参数错误'))
  }

  const fileId = mockUploadedFiles.length + 1
  const file = {
    fileId,
    uploaderUserId: 1,
    bizType,
    objectId: Number(data.objectId || 0),
    fileName,
    mimeType,
    size,
    storageKey: `${bizType}/${fileId}-${fileName}`,
    accessLevel: 'private',
    createdAt: new Date().toISOString()
  }
  const upload = {
    fileId,
    uploadUrl: `mock://upload/${file.storageKey}`,
    storageKey: file.storageKey,
    headers: {
      'x-zhw-file-id': String(fileId)
    },
    expiresAt: new Date(Date.now() + 15 * 60 * 1000).toISOString()
  }

  mockUploadedFiles.push(file)

  return wait(ok({
    upload,
    file
  }))
}

function buildMockDownloadURL(fileId) {
  const file = mockUploadedFiles.find((item) => Number(item.fileId) === Number(fileId))
  if (!file) {
    return wait(fail(40404, 'file not found'))
  }

  return wait(ok({
    fileId: file.fileId,
    downloadUrl: `mock://download/${file.storageKey}`,
    expiresAt: new Date(Date.now() + 15 * 60 * 1000).toISOString()
  }))
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
  const items = mockRoleApplications.map((item) => {
    const submitted = mockSubmittedRoleApplications.find((application) => application.roleType === item.roleType)

    if (!submitted) {
      return item
    }

    return Object.assign({}, item, submitted)
  })

  return {
    items,
    pageConfig: mockRoleApplicationPageConfig
  }
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
  const gameId = data.gameId || data.bizId || 1

  return Object.assign({}, detail, {
    id: warningId,
    warningId,
    gameId,
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
    notificationId,
    feedback: buildSystemNotificationFeedback()
  })
}

function submitMockTradeWarningAction(data = {}) {
  const action = String(data.action || '').trim()
  const warningId = data.warningId || data.id || mockTradeWarningDetail.warningId
  const orderId = data.orderId || mockTradeWarningDetail.order.orderNo
  const gameId = data.gameId || 1

  if (action !== 'delay' && action !== 'deliver') {
    return wait(fail(40001, mockTradeWarningDetail.texts.invalidActionText))
  }

  if (action === 'deliver') {
    return wait(ok({
      action,
      handled: true,
      message: mockTradeWarningDetail.texts.deliverSuccessText,
      target: {
        route: `pages/game/delivery/index?gameId=${encodeURIComponent(gameId)}&warningId=${encodeURIComponent(warningId)}&orderId=${encodeURIComponent(orderId)}`,
        routeKey: 'gameDelivery'
      }
    }))
  }

  return wait(ok({
    action,
    handled: true,
    message: mockTradeWarningDetail.texts.delaySuccessText,
    statusText: mockTradeWarningDetail.texts.delayStatusText,
    warningId,
    orderId
  }))
}

function buildSystemNotificationFeedback() {
  const sections = []
  if (item.caseDesc) {
    sections.push({ label: '服务亮点：', text: item.caseDesc })
  }
  if (item.sourceText || item.visibilityText) {
    sections.push({ label: '能力来源：', text: item.sourceText || item.visibilityText })
  }
  if (item.feedbackText) {
    sections.push({ label: '玩家反馈：', text: item.feedbackText })
  }

  return {
    question: mockSystemNotificationDetail.feedback.question,
    useful: Object.assign({}, mockSystemNotificationDetail.feedback.useful, {
      count: mockSystemNotificationFeedback.useful,
      countText: String(mockSystemNotificationFeedback.useful)
    }),
    useless: Object.assign({}, mockSystemNotificationDetail.feedback.useless, {
      count: mockSystemNotificationFeedback.useless,
      countText: String(mockSystemNotificationFeedback.useless)
    })
  }
}

function submitMockSystemNotificationFeedback(data = {}) {
  const value = String(data.value || data.feedback || '').trim()

  if (value !== 'useful' && value !== 'useless') {
    return wait(fail(40001, mockSystemNotificationDetail.texts.feedbackFailedText))
  }

  mockSystemNotificationFeedback[value] += 1

  return wait(ok({
    value,
    feedback: buildSystemNotificationFeedback()
  }))
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

function buildPointsSummary() {
  const pointsAvailable = Number(mockPointsMallState.pointsAvailable) || 0
  const earnedFromLogs = mockPointsLogsState.reduce((sum, item) => {
    const changeValue = Number(item.changeValue || 0)

    return changeValue > 0 ? sum + changeValue : sum
  }, 0)

  return {
    userId: 1,
    availablePoints: pointsAvailable,
    frozenPoints: 0,
    totalEarnedPoints: Math.max(pointsAvailable, earnedFromLogs),
    redeemedPoints: buildPointsLogs().reduce((sum, item) => {
      const changeValue = Number(item.changeValue || 0)

      return changeValue < 0 ? sum + Math.abs(changeValue) : sum
    }, 0),
    expiredPoints: 0,
    stats: mockPointsPageConfig.stats,
    rules: mockPointsPageConfig.rules,
    earnExample: mockPointsPageConfig.earnExample,
    roleExamples: mockPointsPageConfig.roleExamples,
    filters: mockPointsPageConfig.filters,
    noteText: mockPointsPageConfig.noteText,
    version: mockPointsPageConfig.version,
    updatedAt: new Date().toISOString()
  }
}

function buildPointsLogs() {
  const existingRedemptionBizIds = new Set(mockPointsLogsState
    .filter((item) => item.bizType === 'redemption_order')
    .map((item) => String(item.bizId || '')))
  const orderLogs = (mockPointsOrdersState.orders || []).map((order, index) => {
    const cost = Number(String(order.pointsCost || order.cost || order.pointsText || '').replace(/[^\d]/g, '')) || 0
    const bizId = order.id || order.orderId || 0

    return {
      id: `points-log-order-${bizId || index}`,
      userId: 1,
      changeValue: -cost,
      beforePoints: 0,
      afterPoints: 0,
      bizType: 'redemption_order',
      bizId,
      reason: '积分商城兑换',
      createdAt: order.createdAt || new Date(Date.now() - (index + 1) * 3600000).toISOString()
    }
  }).filter((item) => item.changeValue < 0 && !existingRedemptionBizIds.has(String(item.bizId || '')))

  return mockPointsLogsState.concat(orderLogs).sort((left, right) => {
    return new Date(right.createdAt).getTime() - new Date(left.createdAt).getTime()
  })
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
    emptyText: source.pageConfig && source.pageConfig.emptyText || ''
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
    ],
    emptyText: mockPointsOrdersState.pageConfig && mockPointsOrdersState.pageConfig.logisticsEmptyText || ''
  }
}

function findMockPointsOrder(orderId) {
  const id = String(orderId || '').trim()

  return (mockPointsOrdersState.orders || []).find((item) => String(item.id || item.orderId || item.orderNo || '') === id)
}

function buildPointsOrderDetail(orderId) {
  const order = findMockPointsOrder(orderId)

  if (!order) {
    return null
  }

  return {
    order: JSON.parse(JSON.stringify(order)),
    detailRows: [
      { label: '订单编号', value: order.orderNo || order.id },
      { label: '兑换商品', value: order.title || order.goodsName || '' },
      { label: '消耗积分', value: order.pointsText || order.points || '' },
      { label: '订单状态', value: order.statusText || order.status || '' },
      { label: '兑换时间', value: order.exchangedAtText || order.createdAtText || order.time || '' }
    ],
    emptyText: mockPointsOrdersState.pageConfig && mockPointsOrdersState.pageConfig.detailEmptyText || ''
  }
}

function mockPointsOrderActionLabel(key) {
  const actions = mockPointsOrdersState.pageConfig && mockPointsOrdersState.pageConfig.actions || {}
  return actions[key] || ''
}

function cancelMockPointsOrder(orderId) {
  const order = findMockPointsOrder(orderId)

  if (!order) {
    return wait(fail(40404, '兑换订单不存在'))
  }

  const statusKey = String(order.statusKey || order.status || '').trim()

  if (statusKey !== 'pending' && statusKey !== 'pending_ship') {
    return wait(fail(40903, '当前订单状态不允许取消'))
  }

  order.statusKey = 'canceled'
  order.status = 'canceled'
  order.statusText = '已取消'
  order.statusTone = 'muted'
  order.actions = [
    { key: 'detail', label: mockPointsOrderActionLabel('detail'), type: 'ghost' }
  ]

  const refundPoints = Number(String(order.pointsCost || order.cost || '').replace(/[^\d]/g, '')) ||
    Number(String(order.pointsText || '').replace(/[^\d]/g, '')) ||
    0
  const beforePoints = Number(mockPointsMallState.pointsAvailable || 0)
  mockPointsMallState.pointsAvailable = Number(mockPointsMallState.pointsAvailable || 0) + refundPoints
  mockPointsLogsState.unshift({
    id: `points-log-refund-${Date.now()}`,
    userId: 1,
    changeValue: refundPoints,
    beforePoints,
    afterPoints: mockPointsMallState.pointsAvailable,
    bizType: 'redemption_refund',
    bizId: order.id || order.orderId || 0,
    reason: '兑换订单退回',
    createdAt: new Date().toISOString()
  })

  return wait(ok({
    order,
    pointsSummary: {
      availablePoints: mockPointsMallState.pointsAvailable
    },
    message: '订单已取消，积分已退回'
  }))
}

function buildMockInviteOverview() {
  return {
    profile: {
      level: 'V5 探险家',
      name: '小明',
      desc: '邀请码: TEST2026 · 关系数 5'
    },
    metrics: [
      { value: '5', label: '总邀约数', trend: '▲', tone: 'up' },
      { value: '3', label: '成功转化', trend: '▲', tone: 'up' },
      { value: '60%', label: '转化率', trend: '▲', tone: 'up' },
      { value: '¥12.58', label: '分润收益', trend: '▲', tone: 'up' }
    ],
    actions: [
      { key: 'share_card', icon: '🔗', label: '分享邀请码', inviteCode: 'TEST2026' },
      { key: 'qrcode', icon: '▦', label: '二维码', iconClass: 'white', inviteCode: 'TEST2026' },
      { key: 'poster', icon: '▧', label: '生成海报', iconClass: 'white', inviteCode: 'TEST2026' }
    ],
    tabs: ['数据概览', '关系网络', '邀约记录', '贡献排行', '收益明细'],
    trends: [
      { label: '本周新增邀约', value: '+5', tone: 'cyan' },
      { label: '本周新增转化', value: '+3', tone: 'green' },
      { label: '本周分润', value: '¥12.58', tone: 'cyan' }
    ]
  }
}

function buildMockInviteMembers() {
  return [
    { id: 'member-001', avatar: '张', name: '张大山', desc: '邀约 8 · 转化 3 · 活跃 12天', direct: '+', team: '贡献 ¥42.60' },
    { id: 'member-002', avatar: '李', name: '李小红', desc: '邀约 6 · 转化 2 · 活跃 9天', direct: '+', team: '贡献 ¥31.80' },
    { id: 'member-003', avatar: '王', name: '王建国', desc: '邀约 4 · 转化 1 · 活跃 7天', direct: '+', team: '贡献 ¥25.40' }
  ]
}

function buildMockInviteNetwork() {
  return {
    summary: [
      { value: '5', label: '已服务\n位玩家' },
      { value: '¥12.58', label: '本周收益' }
    ],
    networkNodes: ['1', '2', '3', '4', '5'],
    avatars: ['张', '李', '王'],
    members: buildMockInviteMembers()
  }
}

function buildMockInviteRecords(data = {}) {
  const allRecords = [
    {
      role: 'referred',
      statusKey: 'progress',
      status: '进行中',
      statusClass: 'blue',
      id: 'REF-MOCK-001',
      time: '07-01 14:30',
      expertAvatar: '张',
      expert: '张大山',
      memberId: 'member-001',
      playerAvatar: '我',
      player: '我',
      title: '邀请关系服务',
      budget: '你的奖励：¥42.60',
      income: '',
      route: '/pages/profile/service-center/invite/member-detail/index?memberId=member-001'
    },
    {
      role: 'referred',
      statusKey: 'completed',
      status: '已完成',
      statusClass: 'green',
      id: 'REF-MOCK-002',
      time: '06-28 10:20',
      expertAvatar: '李',
      expert: '李小红',
      memberId: 'member-002',
      playerAvatar: '我',
      player: '我',
      title: '邀请关系服务',
      budget: '你的奖励：¥31.80',
      income: '已结算',
      route: '/pages/profile/service-center/invite/member-detail/index?memberId=member-002'
    }
  ]
  const activeRole = data.role || 'referred'
  const activeStatus = data.status || 'all'
  const records = allRecords.filter((record) => {
    const roleMatched = record.role === activeRole
    const statusMatched = activeStatus === 'all' || record.statusKey === activeStatus

    return roleMatched && statusMatched
  })

  return {
    activeRole,
    activeStatus,
    roleTabs: [
      { key: 'referred', label: '我引荐的', count: allRecords.length },
      { key: 'created', label: '我发起的', count: 0 }
    ],
    filters: [
      { key: 'all', label: '全部' },
      { key: 'progress', label: '进行中' },
      { key: 'completed', label: '已完成' },
      { key: 'timeout', label: '超时' },
      { key: 'cancelled', label: '已取消' }
    ],
    records,
    allRecords,
    warning: null
  }
}
function buildMockInviteRanking(data = {}) {
  return {
    activePeriodIndex: 0,
    activeType: data.type || 'inviteCount',
    periods: [
      { key: 'week', label: '本周' },
      { key: 'month', label: '本月' },
      { key: 'quarter', label: '本季' },
      { key: 'year', label: '本年' },
      { key: 'all', label: '全部' }
    ],
    rankTypes: [
      { key: 'inviteCount', label: '邀约数排行' },
      { key: 'profitContribution', label: '分润贡献排行' }
    ],
    members: buildMockInviteMembers().map((item, index) => ({
      id: item.id,
      rank: index + 1,
      avatar: item.avatar,
      name: item.name,
      level: '一级成员',
      activeDays: 10 - index,
      inviteCount: 8 - index,
      profitContribution: item.team.replace('贡献 ', '')
    }))
  }
}

function buildMockInviteIncome() {
  return {
    trendSeries: [
      { month: '1月', amount: 120 },
      { month: '2月', amount: 180 },
      { month: '3月', amount: 260 },
      { month: '4月', amount: 320 },
      { month: '5月', amount: 480 },
      { month: '6月', amount: 620 }
    ],
    metrics: [
      { label: '本月分润', value: '¥12.58', desc: '已结算收益' },
      { label: '累计分润', value: '¥42.60', desc: '含待结算收益' },
      { label: '活跃成员', value: '5', desc: '当前关系数' },
      { label: '产生分润局数', value: '2', desc: '累计流水' }
    ],
    flows: [
      { icon: '🎯', title: '组局分润 · member', time: '06-14 20:30', amount: '+¥8.00' },
      { icon: '🎯', title: '组局分润 · guide', time: '06-13 18:15', amount: '+¥4.58' }
    ]
  }
}

function buildMockInviteMemberDetail(memberId) {
  const member = buildMockInviteMembers().find((item) => item.id === memberId) || buildMockInviteMembers()[0]

  return {
    memberId: member.id,
    member: {
      avatar: member.avatar,
      name: member.name,
      level: '一级成员'
    },
    stats: [
      { value: '8', label: '总邀约' },
      { value: '3', label: '成功转化' },
      { value: '37%', label: '转化率' }
    ],
    income: [
      { label: '直接贡献收益', value: '¥24.00' },
      { label: '团队贡献收益', value: '¥18.60' },
      { label: '合计贡献', value: '¥42.60', highlight: true }
    ],
    activities: [
      { icon: '🎯', title: '邀请关系建立', time: '06-14 20:30', amount: '+' },
      { icon: '📈', title: '关系强度更新', time: '06-15 10:00', amount: '+' }
    ]
  }
}

function mockServiceReviewTemplates() {
  return ['感谢您的认可，', '期待下次合作', '有问题随时联系', '我们会继续努力', '感谢反馈，已改进']
}

function buildMockServiceReviewStats(reviews) {
  const total = reviews.length
  const pending = reviews.filter((item) => item.statusType !== 'replied').length
  const good = reviews.filter((item) => Number(item.score) >= 4).length

  return [
    { value: String(total), label: '近30天新增评价' },
    { value: total ? `${Math.round((total - pending) * 100 / total)}%` : '0%', label: '回复率' },
    { value: total ? `${Math.round(good * 100 / total)}%` : '0%', label: '好评率' },
    { value: '2小时', label: '平均响应' }
  ]
}

function buildMockServiceReviewScore(reviews) {
  const total = reviews.length
  const sum = reviews.reduce((value, item) => value + Number(item.score || 0), 0)
  const good = reviews.filter((item) => Number(item.score) >= 4).length
  const overall = total ? (sum / total).toFixed(1) : '0.0'

  return {
    overall,
    total: String(total),
    goodRate: total ? `${Math.round(good * 100 / total)}%` : '0%',
    stars: total ? '★'.repeat(Math.round(sum / total)) + '☆'.repeat(5 - Math.round(sum / total)) : '☆☆☆☆☆',
    breakdown: [5, 4, 3].map((score) => {
      const count = reviews.filter((item) => Number(item.score) === score).length
      const percent = total ? Math.round(count * 100 / total) : 0

      return { label: `${score}星`, percent: `${percent}%`, style: `width: ${percent}%;` }
    })
  }
}

function buildMockServiceReviews() {
  const reviews = mockServiceReviewsState.map((item) => Object.assign({}, item))

  return {
    score: buildMockServiceReviewScore(reviews),
    pendingCount: reviews.filter((item) => item.statusType !== 'replied').length,
    stats: buildMockServiceReviewStats(reviews),
    reviews,
    templates: mockServiceReviewTemplates()
  }
}

function buildMockServiceReviewDetail(reviewId) {
  const review = mockServiceReviewsState.find((item) => String(item.id) === String(reviewId))

  if (!review) {
    return null
  }

  const history = [
    {
      role: `${review.user}（玩家）`,
      avatar: review.avatar,
      time: review.timeText,
      content: review.content,
      side: 'user'
    }
  ]

  if (review.reply) {
    history.push({
      role: '我（行家）',
      avatar: '我',
      time: '已回复',
      content: review.reply,
      side: 'expert'
    })
  }

  return {
    review: Object.assign({}, review),
    templates: mockServiceReviewTemplates(),
    templateRows: [['感谢您的认可，', '期待下次合作'], ['有问题随时联系', '我们会继续努力'], ['感谢反馈，已改进']],
    history
  }
}

function submitMockServiceReviewReply(reviewId, data = {}) {
  const content = String(data.content || '').trim()
  const review = mockServiceReviewsState.find((item) => String(item.id) === String(reviewId))

  if (!review) {
    return wait(fail(40404, '评价不存在'))
  }

  if (!content) {
    return wait(fail(40002, '请输入回复内容'))
  }

  if (content.length > 200) {
    return wait(fail(40003, '回复内容不能超过200字'))
  }

  review.reply = content
  review.statusType = 'replied'

  return wait(ok({
    reviewId,
    reply: {
      content,
      repliedAt: '2026-06-30T10:00:00+08:00',
      timeText: '刚刚'
    },
    statusType: 'replied'
  }))
}

function exchangePointsMallGood(data = {}) {
  const goodId = String(data.goodId || data.productId || data.itemId || data.id || '').trim()

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
  mockPointsLogsState.unshift({
    id: `points-log-order-${orderId}`,
    userId: 1,
    changeValue: -cost,
    beforePoints: pointsAvailable,
    afterPoints: mockPointsMallState.pointsAvailable,
    bizType: 'redemption_order',
    bizId: orderId,
    reason: '积分商城兑换',
    createdAt: new Date().toISOString()
  })

  mockPointsOrdersState.orders = [
    {
      id: orderId,
      statusKey: 'pending_ship',
      statusText: '待发货',
      statusTone: 'orange',
      iconText: good.iconText || '🎁',
      title: good.title || '',
      pointsText: `${formatNumber(cost)}积分`,
      pointsCost: cost,
      exchangedAtText: '兑换时间: 刚刚',
      actions: [
        { key: 'detail', label: mockPointsOrderActionLabel('detail'), type: 'ghost' },
        { key: 'cancel', label: mockPointsOrderActionLabel('cancel'), type: 'ghost' }
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

function defaultSystemProfileInfo() {
  return {
    profile: {
      avatarText: 'ZW',
      name: '张伟',
      phone: '138****8888',
      contactVisibility: 'all',
      hobby: '未填写',
      company: '腾讯科技',
      jobTitle: '产品经理',
      businessCountText: '已设置3项',
      resources: '未填写',
      publicBusinessInfo: true
    },
    personalInfo: {
      avatarText: 'ZW',
      name: '张伟',
      phoneMasked: '138****8888',
      contactVisibility: 'all',
      hobby: '未填写'
    },
    enterpriseInfo: {
      company: '腾讯科技',
      jobTitle: '产品经理',
      businessCountText: '已设置3项',
      resources: '未填写',
      publicBusinessInfo: true
    },
    certifications: [
      { key: 'personal', status: '已认证', statusClass: 'verified', tone: 'green' },
      { key: 'enterprise', status: '未认证', statusClass: '', tone: 'blue' }
    ],
    visibilityOptions: [
      { key: 'all', label: '全部展示' },
      { key: 'member', label: '仅会员可见' },
      { key: 'hidden', label: '完全隐藏' }
    ]
  }
}

function buildSystemProfileInfo() {
  if (!mockSystemProfileInfoState) {
    return defaultSystemProfileInfo()
  }

  return JSON.parse(JSON.stringify(mockSystemProfileInfoState))
}

function saveSystemProfileInfo(data = {}) {
  const personalInfo = data.personalInfo || {}
  const enterpriseInfo = data.enterpriseInfo || {}
  const contactVisibility = String(personalInfo.contactVisibility || '').trim()
  const validVisibility = ['all', 'member', 'hidden'].indexOf(contactVisibility) !== -1

  if (!validVisibility) {
    return wait(fail(40001, '联系方式可见性设置错误'))
  }

  const base = defaultSystemProfileInfo()
  mockSystemProfileInfoState = {
    profile: Object.assign({}, base.profile, {
      avatarText: personalInfo.avatarText || base.profile.avatarText,
      name: personalInfo.name || base.profile.name,
      phone: personalInfo.phoneMasked || base.profile.phone,
      contactVisibility,
      hobby: personalInfo.hobby || base.profile.hobby,
      company: enterpriseInfo.company || base.profile.company,
      jobTitle: enterpriseInfo.jobTitle || base.profile.jobTitle,
      businessCountText: enterpriseInfo.businessCountText || base.profile.businessCountText,
      resources: enterpriseInfo.resources || base.profile.resources,
      publicBusinessInfo: typeof enterpriseInfo.publicBusinessInfo === 'boolean'
        ? enterpriseInfo.publicBusinessInfo
        : base.profile.publicBusinessInfo
    }),
    id: `profile-info-${Date.now()}`,
    personalInfo: Object.assign({}, base.personalInfo, personalInfo),
    enterpriseInfo: Object.assign({}, base.enterpriseInfo, enterpriseInfo),
    certifications: Array.isArray(data.certifications) ? data.certifications : base.certifications,
    visibilityOptions: base.visibilityOptions,
    savedAt: '2026-06-27T00:00:00+08:00'
  }

  return wait(ok(mockSystemProfileInfoState))
}

function submitMockEnterpriseCertification(data = {}) {
  if (!data.companyName || !data.unifiedSocialCreditCode || !data.legalPerson || !data.businessLicenseFileId || !data.publicAccountFileId) {
    return wait(fail(40001, '企业认证材料不完整'))
  }
  const base = defaultSystemProfileInfo()
  mockSystemProfileInfoState = Object.assign({}, buildSystemProfileInfo(), {
    certifications: (base.certifications || []).map((item) => item.key === 'enterprise'
      ? Object.assign({}, item, { status: '审核中', statusClass: 'pending', desc: '材料已提交，等待后台审核' })
      : item),
    enterpriseCertification: Object.assign({}, data, { status: 'pending' })
  })
  return wait(ok(mockSystemProfileInfoState.enterpriseCertification))
}

function buildSystemSkillConfig() {
  return JSON.parse(JSON.stringify(mockSystemSkillConfigState))
}

function buildMockServiceCaseDetail(caseId) {
  const config = buildSystemSkillConfig()
  const groups = config.skillGroups || {}
  const caseItems = Array.isArray(groups.cases) ? groups.cases : []
  const visibleItems = Array.isArray(groups.visible) ? groups.visible : []
  const item = caseItems.concat(visibleItems).find((entry) => String(entry.id || '') === String(caseId || '')) || caseItems[0] || visibleItems[0]

  if (!item) {
    return null
  }

  const match = String(item.casePlayers || '').match(/\d+/)
  const score = Number(item.rating || String(item.ratingText || '').replace('分', '')) || 5

  return {
    caseInfo: {
      id: item.id,
      title: item.caseTitle || item.title || '服务案例',
      date: item.caseDate || item.lockedAt || '',
      playersText: item.casePlayers || '待绑定',
      totalPlayers: match ? Number(match[0]) : 0,
      ratingText: item.ratingText || `${score.toFixed(1)}分`,
      iconText: item.iconText || '★',
      tone: item.tone || 'blue'
    },
    rating: {
      score: score.toFixed(1),
      tags: (Array.isArray(item.tags) ? item.tags : []).concat([item.linkedSkillTitle, item.badge, item.sourceText].filter(Boolean))
    },
    detailSections: Array.isArray(item.detailSections) ? item.detailSections : sections,
    players: Array.isArray(item.players)
      ? item.players.map((player, index) => Object.assign({}, player, { last: index === item.players.length - 1 }))
      : []
  }
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

function defaultProfileSettings() {
  return {
    sections: [
      {
        title: '账号安全',
        rows: [
          { id: 'wechatBind', label: '绑定微信', iconKey: 'wechatBind', value: '未绑定', arrow: true, action: 'bind_wechat' },
          { id: 'payPassword', label: '支付密码', iconKey: 'payPassword', value: '一期未开放', arrow: true, disabledReason: '一期未接真实支付，支付密码暂未开放' },
          { id: 'loginPassword', label: '登录密码', iconKey: 'loginPassword', value: '未设置', arrow: true, action: 'set_password' },
          { id: 'phone', label: '更换手机号', iconKey: 'phone', value: '138****8888', arrow: true, disabledReason: '请在我的资料中更新联系方式' },
          { id: 'facePay', label: '指纹/面容支付', iconKey: 'facePay', switch: true, enabled: true }
        ]
      },
      {
        title: '通知设置',
        rows: [
          { id: 'gamePush', label: '局消息推送', iconKey: 'gamePush', switch: true, enabled: true },
          { id: 'systemNotice', label: '系统通知', iconKey: 'systemNotice', switch: true, enabled: true },
          { id: 'subscribeNotice', label: '订阅消息', iconKey: 'subscribeNotice', switch: true, enabled: true },
          { id: 'emailNotice', label: '邮件通知', iconKey: 'emailNotice', switch: true, enabled: false },
          { id: 'quietHours', label: '消息免打扰', iconKey: 'quietHours', value: '22:00 - 08:00', arrow: true, disabledReason: '消息免打扰详细配置暂未开放' }
        ]
      },
      {
        title: '隐私设置',
        rows: [
          { id: 'showGames', label: '允许他人查看我的局', iconKey: 'showGames', switch: true, enabled: true },
          { id: 'showReviews', label: '允许他人查看我的评价', iconKey: 'showReviews', switch: true, enabled: true },
          { id: 'findByPhone', label: '允许通过手机号找到我', iconKey: 'findByPhone', switch: true, enabled: false },
          { id: 'personalized', label: '个性化推送', iconKey: 'personalized', switch: true, enabled: true },
          { id: 'privacySummary', label: '隐私政策摘要', iconKey: 'privacySummary', arrow: true, agreementKey: 'privacy' },
          { id: 'thirdPartyList', label: '第三方共享清单', iconKey: 'thirdPartyList', arrow: true, agreementKey: 'third_party_sharing' },
          { id: 'collectionList', label: '信息收集清单', iconKey: 'collectionList', arrow: true, agreementKey: 'personal_info_collection' }
        ]
      },
      {
        title: '通用设置',
        rows: [
          { id: 'clearCache', label: '清除缓存', iconKey: 'clearCache', value: '当前缓存 23.5MB', arrow: true, action: 'clear_cache' },
          { id: 'about', label: '关于我们', iconKey: 'about', value: '版本 v1.0.0', arrow: true, agreementKey: 'about' }
        ]
      }
    ],
    logoutText: '退出登录'
  }
}

function buildProfileSettings() {
  return JSON.parse(JSON.stringify(mockProfileSettingsState || defaultProfileSettings()))
}

function saveProfileSettings(data = {}) {
  const sections = Array.isArray(data.sections) ? data.sections : []

  if (!sections.length) {
    return wait(fail(40001, '系统设置不能为空'))
  }

  mockProfileSettingsState = Object.assign({}, defaultProfileSettings(), {
    sections,
    updatedAt: new Date().toISOString()
  })

  return wait(ok(buildProfileSettings()))
}

function defaultSystemBlockSettings() {
  const keywords = ['培训', '课程', '收费教学', '加微信', '私下']
  const whitelist = [{ id: `guide-${mockCurrentUser.id}`, name: mockCurrentUser.nickname || '微信用户', role: '我', reason: '默认不受屏蔽' }]

  return {
    enabled: true,
    protectionMode: 'hard',
    renewalDays: 30,
    summary: {
      protectedUserText: '0位用户',
      blockedExpertText: `${keywords.length}个关键词`,
      renewalDaysText: '30天'
    },
    stats: [
      { value: 0, label: '保护用户' },
      { value: keywords.length, label: '关键词' },
      { value: '100%', label: '生效率' }
    ],
    configRows: [
      { key: 'protection', title: '保护模式', desc: '当前：硬保护', badge: '已开启' },
      { key: 'scene', title: '分场景配置', desc: '推荐/附近/列表', badge: '4开' },
      { key: 'whitelist', title: '白名单', desc: `${whitelist.length}位不受保护`, badge: `${whitelist.length}/20` }
    ],
    manageRows: [
      { key: 'users', title: '用户屏蔽', desc: '双方互不可见', badge: '0人' },
      { key: 'keywords', title: '关键词屏蔽', desc: '自动过滤内容', badge: `${keywords.length}/20` }
    ],
    rules: [
      { prefix: '保护期默认', strong: '30天', suffix: '' },
      { prefix: '配置变更', strong: '5秒内', suffix: '对新请求生效' }
    ],
    protectionModes: [
      {
        key: 'hard',
        title: '硬保护',
        desc: '完全过滤，用户无感知',
        rules: [
          { prefix: '', strong: '完全过滤', suffix: '，同类行家内容不展示' },
          { prefix: '被保护用户 ', strong: '零曝光', suffix: '' },
          { prefix: '适合竞争激烈的同城市', strong: '', suffix: '' }
        ]
      },
      {
        key: 'soft',
        title: '软保护',
        desc: '排序降权×0.1，推至第5页后',
        rules: [
          { prefix: '排序权重 ', strong: '×0.1', suffix: '，大幅降权' },
          { prefix: '翻页至局列表第5页后', strong: '不受保护', suffix: '' },
          { prefix: '适合内容不足或过渡期', strong: '', suffix: '' }
        ]
      }
    ],
    renewalOptions: [
      { days: 30, label: '标准周期' },
      { days: 60, label: '双倍保护' },
      { days: 90, label: '季度保护' }
    ],
    renewalRules: [
      { prefix: '保护期最长 ', strong: '90天', suffix: '，到期需重新续期' },
      { prefix: '续期后立即生效，', strong: '5秒内', suffix: ' 覆盖所有新请求' },
      { prefix: '到期前 ', strong: '3天', suffix: ' 将发送提醒通知' }
    ],
    keywords,
    suggestions: ['微商', '直销', '刷单', '贷款', '兼职', '代理', '拉群', '推广'],
    keywordMaxCount: 20,
    keywordStats: [
      { value: keywords.length, label: '已设置' },
      { value: 20, label: '上限', color: 'gold' },
      { value: '模糊', label: '匹配模式', color: 'green' }
    ],
    keywordRules: [
      { prefix: '支持', strong: '模糊匹配', suffix: '，如“培训”匹配“培训机构”“培训课程”' },
      { prefix: '最多可设', strong: '20个', suffix: '关键词' },
      { prefix: '关键词屏蔽仅影响内容展示，', strong: '不影响用户间交互', suffix: '' },
      { prefix: '生效范围：局标题、描述、评论、私信内容', strong: '', suffix: '' }
    ],
    whitelist,
    whitelistCandidates: [],
    whitelistRules: [
      { prefix: '白名单行家', strong: '不受保护', suffix: '影响' },
      { prefix: '最多添加', strong: '20位', suffix: '' }
    ],
    blockedUsers: [],
    blockCandidates: [],
    blockRules: [
      { prefix: '可通过用户主页右上角菜单快速屏蔽', strong: '', suffix: '' },
      { prefix: '屏蔽后双方', strong: '互不可见', suffix: '，历史互动记录保留' },
      { prefix: '解除屏蔽后', strong: '24小时冷却期', suffix: '才能再次屏蔽' },
      { prefix: '屏蔽人数上限', strong: '100人', suffix: '' }
    ],
    scenes: [
      { key: 'recommend', title: '推荐场景', desc: '首页/发现页推荐', mode: 'hard', enabled: true },
      { key: 'nearby', title: '附近场景', desc: 'LBS地理位置推荐', mode: 'soft', enabled: true },
      { key: 'message', title: '私信场景', desc: '局内与私信内容', mode: 'hard', enabled: true },
      { key: 'list', title: '列表浏览', desc: '局列表/行家列表', mode: 'none', enabled: true }
    ],
    sceneRules: [
      { strong: '硬保护', suffix: ' = 完全过滤' },
      { strong: '软保护', suffix: ' = 降权至第5页后' },
      { strong: '不过滤', suffix: ' = 正常展示' }
    ]
  }
}

function buildSystemBlockSettings() {
  return JSON.parse(JSON.stringify(mockSystemBlockSettingsState || defaultSystemBlockSettings()))
}

function saveSystemBlockSettings(data = {}) {
  mockSystemBlockSettingsState = Object.assign({}, defaultSystemBlockSettings(), buildSystemBlockSettings(), data, {
    updatedAt: new Date().toISOString()
  })

  return wait(ok(buildSystemBlockSettings()))
}

function defaultAgreementSections() {
  return [
    { title: '一、协议范围', content: '本协议是您与本平台之间关于使用平台服务所订立的协议。请您仔细阅读本协议，如您不同意本协议的任何内容，请停止使用平台服务。' },
    { title: '二、账号注册', content: '您承诺以真实身份注册账号，并保证所提供的个人资料真实、准确、完整、合法有效。如有变动，应及时更新。' },
    { title: '三、服务内容', content: '平台向您提供组局管理、技能展示、社交互动等服务。您有权按照平台规则使用各项服务。' },
    { title: '四、用户行为规范', content: '您在使用平台服务时，应遵守法律法规，不得发布违法违规信息，不得侵犯他人合法权益。' },
    { title: '五、知识产权', content: '平台所有内容，包括但不限于文字、图片、音频、视频、软件等，均受知识产权法律保护。' },
    { title: '六、免责声明', content: '平台不对因不可抗力或第三方原因导致的服务中断承担责任。' },
    { title: '七、协议变更', content: '平台有权根据需要修改本协议，修改后的协议将在平台公示，公示期满即生效。' }
  ]
}

function defaultProfileAgreements() {
  return [
    { key: 'user-service', title: '用户服务协议', desc: '平台服务条款与规则', signed: true, signedVersion: '2026-07-01', version: '2026-07-01', tone: 'green', last: false, iconKey: 'doc', sections: defaultAgreementSections() },
    { key: 'privacy', title: '隐私政策', desc: '个人信息保护说明', signed: true, signedVersion: '2026-06-30', version: '2026-07-01', tone: 'deep-green', last: false, iconKey: 'lock', sections: defaultAgreementSections() },
    { key: 'settlement', title: '入驻协议', desc: '服务与分润协议', signed: false, version: '2026-07-01', tone: 'orange', last: false, iconKey: 'box', sections: defaultAgreementSections() },
    { key: 'third_party_sharing', title: '第三方共享清单', desc: '第三方服务与共享场景', signed: true, signedVersion: '2026-07-01', version: '2026-07-01', tone: 'deep-green', last: false, iconKey: 'doc', sections: defaultAgreementSections() },
    { key: 'personal_info_collection', title: '信息收集清单', desc: '平台收集和使用信息的说明', signed: true, signedVersion: '2026-07-01', version: '2026-07-01', tone: 'green', last: false, iconKey: 'lock', sections: defaultAgreementSections() },
    { key: 'about', title: '关于我们', desc: '平台介绍与服务说明', signed: true, signedVersion: '2026-07-01', version: '2026-07-01', tone: 'green', last: true, iconKey: 'doc', sections: defaultAgreementSections() }
  ]
}

function normalizeProfileAgreements(items = []) {
  return items.map((item) => {
    const version = item.version || '2026-07-01'
    const signedVersion = item.signedVersion || (item.signed ? version : '')
    const requiresResign = Boolean(item.signed) && signedVersion && signedVersion !== version

    return Object.assign({}, item, {
      version,
      signedVersion,
      signed: Boolean(item.signed) && !requiresResign,
      requiresResign,
      signActionText: item.signActionText || '同意并签署',
      signSuccessText: item.signSuccessText || '签署成功',
      signConfirm: item.signConfirm || {
        title: '确认签署协议？',
        desc: '签署后将视为同意协议全部条款，协议立即生效',
        cancelText: '取消',
        confirmText: '确认签署'
      }
    })
  })
}

function buildProfileAgreements() {
  return normalizeProfileAgreements(JSON.parse(JSON.stringify(mockAgreementState || defaultProfileAgreements())))
}

function getMockAgreement(key) {
  return buildProfileAgreements().find((item) => item.key === key)
}

function signMockAgreement(key) {
  const agreements = buildProfileAgreements()
  const agreement = agreements.find((item) => item.key === key)

  if (!agreement) {
    return wait(fail(40404, '协议不存在'))
  }

  agreement.signed = true
  agreement.signedAt = new Date().toISOString()
  agreement.signedVersion = agreement.version || '2026-07-01'
  agreement.requiresResign = false
  mockAgreementState = agreements

  return wait(ok(agreement))
}

function defaultFeedbackSuccessPage() {
  return {
    successTitle: '提交成功',
    successDesc: '感谢您的反馈，我们会认真阅读每一条建议',
    successDescSecond: '处理进度将通过消息通知您',
    backHomeText: '返回首页',
    viewRecordsText: '查看反馈记录',
    reward: {
      show: true,
      title: '获得经验值奖励',
      desc: '优质反馈被采纳后可获得更多经验值',
      value: '+20'
    },
    rating: {
      title: '您对我们的反馈体验满意吗？',
      desc: '0 = 非常不满意，10 = 非常满意',
      minLabel: '非常不满意',
      maxLabel: '非常满意',
      default: 7,
      scoreFrom: 0,
      scoreTo: 10
    },
    reasonsTitle: '哪些方面让您满意？（可多选）',
    reasons: [
      { value: '反馈流程简单' },
      { value: '响应速度快' },
      { value: '客服态度好' },
      { value: '问题解决彻底' },
      { value: '界面清晰易用' },
      { value: '有积分激励' }
    ],
    submitRatingText: '提交评价',
    ratingSavedText: '评价已记录'
  }
}

function buildSystemFeedbackHome() {
  return {
    activeType: 'feature',
    activeSession: 'general',
    feedbackTypes: [
      { key: 'feature', label: '功能建议', iconKey: 'pencil' },
      { key: 'problem', label: '问题反馈', iconKey: 'alert' },
      { key: 'experience', label: '体验优化', iconKey: 'star' },
      { key: 'game', label: '组局相关', iconKey: 'problem' },
      { key: 'expert', label: '行家相关', iconKey: 'problem' },
      { key: 'points', label: '积分/提现', iconKey: 'notify' },
      { key: 'other', label: '其他', iconKey: 'alert' }
    ],
    sessions: [
      { key: 'general', title: '通用反馈', meta: '不关联具体组局' },
      { key: 'latest_game', title: '最近组局', meta: '可在提交时补充说明' }
    ],
    quickTypes: [
      { key: 'problem', label: '遇到问题', iconKey: 'problem', tone: 'red' },
      { key: 'feature', label: '功能建议', iconKey: 'pencil', tone: 'cyan' },
      { key: 'experience', label: '体验优化', iconKey: 'star', tone: 'gold' },
      { key: 'other', label: '其他', iconKey: 'alert', tone: 'gray' }
    ],
    quickActions: [
      { key: 'feature', label: '功能建议', iconKey: 'pencil' },
      { key: 'problem', label: '问题反馈', iconKey: 'alert' },
      { key: 'screenshot', label: '截图反馈', iconKey: 'image' }
    ],
    limits: {
      contentMaxLength: 500,
      fileMaxCount: 9,
      uploadNote: '支持 JPG/PNG 图片和语音文件，单个附件不超过 5MB，最多 9 个附件',
      uploadFullText: '最多上传 9 个附件',
      uploadSelectedTemplate: '已选择 {selected}/{max} 个附件'
    },
    successPage: defaultFeedbackSuccessPage()
  }
}

function feedbackTypeLabel(typeKey) {
  const matched = buildSystemFeedbackHome().feedbackTypes.find((item) => item.key === typeKey)
  return matched ? matched.label : '功能建议'
}

function ensureMockFeedbackRecords() {
  return mockFeedbackRecords
}

function submitMockSystemFeedback(data = {}) {
  const typeKey = String(data.typeKey || '').trim()
  const content = String(data.content || '').trim()
  const contact = String(data.contact || '').trim()
  const fileIds = Array.isArray(data.fileIds) ? data.fileIds : []
  const finalContent = content || '附件反馈'

  if (!typeKey || (!content && !fileIds.length) || content.length > 500 || contact.length > 80 || fileIds.length > 9) {
    return wait(fail(42206, '反馈参数错误'))
  }

  const createdAt = new Date()
  const record = {
    id: `FB${Date.now()}`,
    type: feedbackTypeLabel(typeKey),
    typeKey,
    typeTone: typeKey === 'problem' ? 'red' : typeKey === 'experience' ? 'gold' : 'blue',
    status: '待处理',
    statusClass: 'pending',
    content: finalContent,
    contact,
    fileIds,
    quick: Boolean(data.quick),
    time: createdAt.toISOString().slice(0, 16).replace('T', ' '),
    createdAt: createdAt.toISOString(),
    images: [],
    messages: [
      { id: 'initial', role: 'me', avatar: '我', time: createdAt.toISOString().slice(11, 16), content: finalContent, fileIds }
    ]
  }

  mockFeedbackRecords = [record].concat(ensureMockFeedbackRecords())

  const successPage = defaultFeedbackSuccessPage()

  return wait(ok({
    record,
    successTitle: successPage.successTitle,
    successDesc: successPage.successDesc,
    successPage
  }))
}

function buildMockFeedbackRecords(data = {}) {
  const activeTab = String(data.tab || 'all')
  const allRecords = ensureMockFeedbackRecords()
  const records = allRecords.filter((record) => {
    if (activeTab === 'resolved') {
      return record.statusClass === 'resolved'
    }
    if (activeTab === 'processing') {
      return record.statusClass !== 'resolved'
    }
    return true
  })

  return {
    activeTab,
    tabs: [
      { key: 'all', label: '全部', count: allRecords.length },
      { key: 'processing', label: '处理中', count: allRecords.filter((item) => item.statusClass !== 'resolved').length },
      { key: 'resolved', label: '已解决', count: allRecords.filter((item) => item.statusClass === 'resolved').length }
    ],
    records,
    allRecords
  }
}

function mockFeedbackDetail(recordId) {
  const record = ensureMockFeedbackRecords().find((item) => item.id === recordId)

  if (!record) {
    return wait(fail(40405, '反馈不存在'))
  }

  return wait(ok({
    record,
    messages: record.messages || []
  }))
}

function appendMockFeedbackMessage(recordId, data = {}) {
  const record = ensureMockFeedbackRecords().find((item) => item.id === recordId)
  const content = String(data.content || '').trim()
  const fileIds = Array.isArray(data.fileIds) ? data.fileIds : []
  const finalContent = content || '附件补充'

  if (!record) {
    return wait(fail(40405, '反馈不存在'))
  }

  if (!content && !fileIds.length) {
    return wait(fail(42207, '补充说明不能为空'))
  }

  const message = {
    id: `MSG${Date.now()}`,
    role: 'me',
    avatar: '我',
    time: new Date().toISOString().slice(11, 16),
    content: finalContent,
    fileIds
  }

  record.messages = (record.messages || []).concat(message)
  record.replyText = '已补充说明'
  record.replyClass = 'processing'
  record.status = '处理中'
  record.statusClass = 'processing'

  return wait(ok({
    record,
    message,
    messages: record.messages
  }))
}

function toFiniteNumber(value, fallback) {
  const number = Number(value)

  return Number.isFinite(number) ? number : fallback
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
      route: 'pages/game/detail/index?id=map-nearby-001'
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
      route: 'pages/game/detail/index?id=map-nearby-002'
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
      route: 'pages/game/detail/index?id=map-nearby-003'
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

function buildMockMapSearch(data = {}) {
  const latitude = toFiniteNumber(data.latitude, 31.2304)
  const longitude = toFiniteNumber(data.longitude, 121.4737)
  const keyword = String(data.keyword || '').trim()

  if (!keyword) {
    return fail(42201, '请输入地点关键词')
  }

  return ok({
    provider: 'mock',
    total: 3,
    items: [
      {
        id: 'mock-place-001',
        title: `${keyword}·城市会客厅`,
        address: '人民广场附近',
        province: '上海市',
        city: '上海',
        cityCode: '310000',
        district: '黄浦区',
        longitude: longitude + 0.0012,
        latitude: latitude + 0.001
      },
      {
        id: 'mock-place-002',
        title: `${keyword}·共享空间`,
        address: '滨江步道 88 号',
        province: '上海市',
        city: '上海',
        cityCode: '310000',
        district: '浦东新区',
        longitude: longitude + 0.002,
        latitude: latitude - 0.0014
      },
      {
        id: 'mock-place-003',
        title: `${keyword}·开放广场`,
        address: '创意园区入口',
        province: '上海市',
        city: '上海',
        cityCode: '310000',
        district: '徐汇区',
        longitude: longitude - 0.0016,
        latitude: latitude + 0.0018
      }
    ]
  })
}

function buildMockMapIndexConfig() {
  return {
    onlineText: '在线',
    defaultLocation: {
      latitude: 31.2304,
      longitude: 121.4737
    },
    defaultRadiusMeters: 3000,
    radiusOptions: [1000, 3000, 5000],
    mapFilters: ['附近组局', '组局路线', '热力图', '好友分布', '解锁图鉴', 'AR'],
    texts: {
      searchPlaceholder: '搜局、搜人、搜地块...',
      loadingNearbyText: '正在获取附近数据...',
      locateToolText: '定',
      refreshToolText: '刷',
      dateRangeText: '2025.09.23 - 2026.03.30',
      detailActionText: '查看详情',
      joinActionText: '去组队',
      playerDetailActionText: '查看资料',
      currentInteractText: '当前可互动',
      offlineInteractText: '暂未在线，可查看轨迹',
      nearbySectionTitle: '附近玩法点',
      nearbyEmptyText: '当前位置附近暂无玩法点'
    },
    version: '2026-07-01'
  }
}

function buildMockMapPlayPages() {
  return {
    version: '2026-07-01',
    pages: {
      'blind-route': {
        title: '组局盲盒',
        description: '不知道去哪局？让命运决定你的下一次城市冒险',
        sectionTitle: '最近开启',
        selectToast: '已选择 {title}',
        cards: [
          { id: 'walk', title: '城市漫步盲盒', desc: '30 分钟内出发，随机匹配附近轻量局', tone: 'blue', icon: '/pages/map/blind-route/assets/i50.png', tags: ['轻松', '附近', '低门槛'] },
          { id: 'food', title: '深夜食堂盲盒', desc: '匹配同城饭搭子和夜宵路线', tone: 'orange', icon: '/pages/map/blind-route/assets/i52.png', tags: ['饭局', '夜间', '社交'] },
          { id: 'photo', title: '拍照路线盲盒', desc: '用一个主题串起三处城市机位', tone: 'purple', icon: '/pages/map/blind-route/assets/i54.png', tags: ['拍照', '路线', '打卡'] }
        ],
        recentRoutes: [
          { title: '徐汇夜风路线', timeText: '20 分钟前', statusText: '已成局' },
          { title: '周末咖啡搭子', timeText: '1 小时前', statusText: '招募中' }
        ]
      },
      'city-atlas': {
        title: '上海探索图鉴',
        progressLabel: '探索进度',
        lockedText: '未解锁',
        hiddenBadge: '隐藏',
        routeSectionTitle: '主题路线',
        filters: ['全部', '已解锁', '未解锁', '隐藏点'],
        unlockInfo: { unlockedCount: 1, totalCount: 36, progressPercent: 33 },
        unlockToast: '{name}已解锁',
        lockedToast: '{name}待解锁',
        unlockPoints: [
          { id: 'bund-night', name: '外滩夜景', statusType: 'unlocked', unlockText: '2024.01.15 解锁', footprintValue: '+50 足迹值', tone: 'blue', iconType: 'building', checked: true },
          { id: 'tianzifang', name: '田子坊', statusType: 'unlocked', unlockText: '2024.02.03 解锁', footprintValue: '+30 足迹值', tone: 'green', iconType: 'lantern', checked: true },
          { id: 'wukang-road', name: '武康路街角', statusType: 'locked', unlockText: '完成 2 次附近打卡后解锁', footprintValue: '+40 足迹值', tone: 'locked', iconType: 'lock', checked: false },
          { id: 'hidden-rooftop', name: '城市天台', statusType: 'hidden', unlockText: '隐藏点待发现', footprintValue: '+80 足迹值', tone: 'purple', iconType: 'hidden', checked: false }
        ],
        themeRoutes: [
          { id: 'couple-walk', name: '情侣漫步', meta: '6个地点 · 预计3小', progressText: '已解锁 2/6', tone: 'sunset', iconText: '💕' },
          { id: 'coffee-shop', name: '咖啡探店', meta: '8个地点 · 预计4小', progressText: '已解锁 0/8', tone: 'cyan', iconText: '☕' }
        ]
      },
      'footprint-heatmap': {
        title: '足迹热力图',
        heatTitle: '全国城市打卡热力',
        legendLabel: '城市打卡热度',
        friendTitle: '好友也在打卡',
        friendMoreText: '查看全部',
        hotTitle: '城市热点排行',
        rangeTabs: ['今日', '本周', '本月', '全部'],
        rangeStats: {
          今日: [{ value: '12', label: '打卡城市' }, { value: '1.8k', label: '玩家足迹' }, { value: '3', label: '热门城市' }],
          本周: [{ value: '38', label: '打卡城市' }, { value: '8.5k', label: '玩家足迹' }, { value: '9', label: '热门城市' }],
          本月: [{ value: '76', label: '打卡城市' }, { value: '26k', label: '玩家足迹' }, { value: '18', label: '热门城市' }],
          全部: [{ value: '126', label: '打卡城市' }, { value: '92k', label: '玩家足迹' }, { value: '31', label: '热门城市' }]
        },
        cityHeatPoints: [
          { id: 'beijing', city: '北京', level: 'mid', className: 'footprint-city-point beijing level-mid' },
          { id: 'shanghai', city: '上海', level: 'hot', className: 'footprint-city-point shanghai level-hot' },
          { id: 'chengdu', city: '成都', level: 'hot', className: 'footprint-city-point chengdu level-hot' },
          { id: 'guangzhou', city: '广州', level: 'mid', className: 'footprint-city-point guangzhou level-mid' },
          { id: 'shenzhen', city: '深圳', level: 'high', className: 'footprint-city-point shenzhen level-high' },
          { id: 'xian', city: '西安', level: 'low', className: 'footprint-city-point xian level-low' },
          { id: 'hangzhou', city: '杭州', level: 'high', className: 'footprint-city-point hangzhou level-high' }
        ],
        friendUpdates: [
          { id: 'alex', avatarText: 'AL', name: 'Alex', desc: '刚刚在成都宽窄巷子打卡', online: true },
          { id: 'sarah', avatarText: 'SA', name: 'Sarah', desc: '25分钟前在西安城墙打卡', online: false }
        ],
        hotCities: [
          { id: 'shanghai', rank: 1, city: '上海市中心', desc: '2456人在这里打卡', progress: 86, level: 'hot' },
          { id: 'chengdu', rank: 2, city: '成都市', desc: '1892人在这里打卡', progress: 72, level: 'warm' },
          { id: 'shenzhen', rank: 3, city: '深圳湾', desc: '1567人在这里打卡', progress: 58, level: 'active' }
        ]
      },
      'friend-city': {
        title: '好友',
        challengeTitle: '进行中的挑战',
        rankingTitle: '城　市　榜',
        nationalRankText: '查看全国榜',
        challengeToast: '{title}进行中',
        emptyChallengeText: '暂无挑战详情',
        rankingToast: '第{rank}名城市榜',
        emptyRankingText: '暂无城市榜详情',
        nationalToast: '已展示当前城市榜',
        duel: { selfName: '我', selfCount: '12区已点亮', rivalName: 'Sarah', rivalCount: '10区已点亮', vsText: 'VS', subtitle: '友谊赛', startButtonText: '发起挑战', recordButtonText: '查看记录' },
        challenges: [
          { id: 'jingan-first', title: '率先点亮静安区', timeLeft: '2天', statusText: '进行中', statusTone: 'pending', selfValue: 3, rivalValue: 2, total: 5, selfPercent: 60, rivalPercent: 40 },
          { id: 'landmark-speed', title: '10个地标速通', timeLeft: '5天', statusText: '领先中', statusTone: 'leading', selfValue: 7, rivalValue: 4, total: 10, selfPercent: 70, rivalPercent: 40 }
        ],
        rankings: [
          { rank: 1, name: 'Mike', desc: '已点亮 28 区', score: '2,450', tone: 'gold' },
          { rank: 2, name: 'Sarah', desc: '已点亮 24 区', score: '2,180', tone: 'silver' },
          { rank: 3, name: 'David', desc: '已点亮 22 区', score: '1,950', tone: 'bronze' },
          { rank: 4, name: '我', desc: '已点亮 12 区', score: '1,240', tone: 'normal' }
        ]
      },
      'real-checkin': {
        taskSectionTitle: '选择打卡任务',
        generateButtonText: '生成足迹碎片',
        texts: { unsupportedCamera: '当前基础库不支持拍照', photoSelected: '打卡照片已选择', storySaved: '打卡文字已记录', storyRequired: '请先填写打卡文字', taskSelected: '{title}已选中', taskRequired: '请选择打卡任务', fragmentPending: '足迹碎片待生成' },
        checkinDetail: { distanceText: '距离目标 15米', spotName: '外滩观景台', statusTitle: '地点已解锁', statusDesc: '完成打卡任务获得足迹值', storyTitle: '留下你的故事', storyPlaceholder: '用20个字记录此刻的心情...', storyMinLength: 20, storyMaxLength: 120, rewards: ['+20 足迹值', '+1 成就点'] },
        checkinTasks: [
          { id: 'photo', title: '拍摄地标合影', desc: '与标志性建筑合影', scoreText: '+10分' },
          { id: 'angle', title: '发现隐藏角度', desc: '拍摄独特的视角', scoreText: '+20分' }
        ]
      }
    }
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
    order: {
      id: Date.now(),
      orderNo: `FREE-${data.gameId || 'mock'}-${Date.now()}`,
      gameId: data.gameId || '',
      amountCent: 0,
      payStatus: 'free_no_pay',
      payChannel: 'none',
      needWechatPay: false
    },
    orderNo: `FREE-${data.gameId || 'mock'}`,
    payStatus: 'free_no_pay',
    amountCent: 0,
    needWechatPay: false,
    mode: 'free_no_pay_placeholder',
    paymentOrderId: `mock_game_payment_${Date.now()}`,
    gameId: data.gameId || '',
    scene: data.scene || 'deposit_game',
    payChannel: 'none',
    amount: 0,
    currency: data.currency || 'CNY',
    splits,
    mockPayment: true
  }))
}

function applyGame(gameId, data = {}) {
  const reason = String(data.reason || '').trim()
  const fileIds = Array.isArray(data.fileIds) ? data.fileIds.filter(Boolean) : []

  if (!gameId) {
    return wait(fail(40004, '缺少局信息'))
  }

  if (!reason) {
    return wait(fail(40005, '请填写申请理由'))
  }

  const application = {
    id: Date.now(),
    gameId,
    userId: 1,
    userName: '当前玩家',
    roleKey: 'player',
    status: 'pending',
    reason,
    fileIds,
    createdAt: new Date().toISOString()
  }
  mockReceivedGameApplications.unshift(application)

  return wait(ok(application))
}

function buildReceivedGameApplications(params = {}) {
  const status = String(params.status || '').trim()
  const items = status
    ? mockReceivedGameApplications.filter((item) => item.status === status)
    : mockReceivedGameApplications

  return {
    items
  }
}

function reviewMockGameApplication(applicationId, data = {}) {
  const idText = String(applicationId || '')
  const item = mockReceivedGameApplications.find((application) => String(application.id) === idText)

  if (!item) {
    return wait(fail(40404, '申请不存在'))
  }

  if (item.status !== 'pending') {
    return wait(fail(40901, '申请已处理'))
  }

  item.status = data.approve ? 'approved' : 'rejected'
  item.reviewedAt = new Date().toISOString()

  return wait(ok(item))
}

function buildGameCategoryConfig() {
  return {
    primaryCategories: [
      {
        key: 'social',
        name: '社交局',
        icon: 'category-social',
        visible: true,
        order: 10,
        children: [
          { key: 'meal', name: '饭局', visible: true, order: 10, selectable: true },
          { key: 'board_game', name: '桌游局', visible: true, order: 20, selectable: true },
          { key: 'friend', name: '交友局', visible: true, order: 30, selectable: true },
          { key: 'walk', name: '同城散步局', visible: true, order: 40, selectable: true }
        ]
      },
      {
        key: 'task',
        name: '任务局',
        icon: 'category-task',
        visible: true,
        order: 20,
        children: [
          { key: 'partner', name: '找合伙人', visible: true, order: 10, selectable: true },
          { key: 'project', name: '做项目', visible: true, order: 20, selectable: true },
          { key: 'brainstorm', name: '头脑风暴', visible: true, order: 30, selectable: true },
          { key: 'cowork', name: '组队共创', visible: true, order: 40, selectable: true }
        ]
      },
      {
        key: 'explore',
        name: '探索局',
        icon: 'category-income',
        visible: true,
        order: 30,
        children: [
          { key: 'city_explore', name: '城市探索', visible: true, order: 10, selectable: true },
          { key: 'route_blind_box', name: '路线盲盒', visible: true, order: 20, selectable: true },
          { key: 'checkin_challenge', name: '打卡挑战', visible: true, order: 30, selectable: true },
          { key: 'night_walk', name: '夜游/徒步/骑行', visible: true, order: 40, selectable: true }
        ]
      },
      {
        key: 'growth',
        name: '成长局',
        icon: 'category-growth',
        visible: true,
        order: 40,
        children: [
          { key: 'reading', name: '读书局', visible: true, order: 10, selectable: true },
          { key: 'fitness', name: '健身局', visible: true, order: 20, selectable: true },
          { key: 'checkin', name: '打卡局', visible: true, order: 30, selectable: true },
          { key: 'deposit_checkin', name: '押金局', visible: true, order: 40, selectable: true },
          { key: 'study', name: '学习共修局', visible: true, order: 50, selectable: true }
        ]
      }
    ],
    typeFilters: [
      { key: 'all', name: '类型', visible: true, order: 0, selectable: true },
      { key: 'free', name: '免费局', visible: true, order: 10, selectable: true },
      { key: 'standard', name: '标准局', visible: true, order: 20, selectable: true },
      { key: 'aa', name: 'AA局', visible: true, order: 30, selectable: true },
      { key: 'crowdfund', name: '众筹局', visible: true, order: 40, selectable: true },
      { key: 'deposit', name: '押金局', visible: true, order: 50, selectable: true },
      { key: 'public_welfare', name: '公益局', visible: true, order: 60, selectable: false }
    ],
    locationFilters: [
      { key: 'all', name: '全国', visible: true, order: 0, selectable: true },
      { key: 'nearby', name: '附近(50km)', visible: true, order: 10, selectable: true }
    ],
    sortOptions: [
      { key: 'comprehensive', name: '综合排序', sortKey: '', sortOrder: 'asc' },
      { key: 'latest', name: '最新发布', sortKey: 'time', sortOrder: 'desc' },
      { key: 'hot', name: '热度最高', sortKey: 'hot', sortOrder: 'desc' },
      { key: 'distance', name: '距离最近', sortKey: 'distance', sortOrder: 'asc' },
      { key: 'credit', name: '信用优先', sortKey: 'credit', sortOrder: 'desc' }
    ],
    eventActions: ['分享', '关注', '引荐', '打招呼'],
    defaultPrimaryCategory: 'task',
    defaultSecondaryCategory: 'project',
    defaultType: 'free',
    createForm: {
      capacity: { min: 5, max: 8 },
      currentLocationText: '当前位置',
      participationModes: [
        { key: 'online', name: '线上' },
        { key: 'offline', name: '线下' },
        { key: 'hybrid', name: '混合' }
      ],
      tags: [
        { key: 'product', name: '产品研发' },
        { key: 'startup', name: '创业' },
        { key: 'city_explore', name: '城市探索' },
        { key: 'cocreation', name: '共创' }
      ],
      completionRules: [
        { key: 'time', name: '时间截止' },
        { key: 'goal', name: '目标达成', active: true },
        { key: 'capacity', name: '人数满额' },
        { key: 'manual', name: '手动结束', active: true }
      ],
      feeTypes: [
        { key: 'free', name: '免费局' },
        { key: 'paid', name: '收费局' }
      ]
    },
    version: '2026-06-30'
  }
}

function buildGameApplicationConfig() {
  return {
    agreementTitle: '入局申请须知',
    agreementText: '申请入局前请确认本人已完成实名，了解局的主题、地点、时间和成员规则。',
    requireRealname: true,
    requireIntro: true,
    requireAgreement: true,
    allowDuplicateApply: false,
    uploadRequired: false,
    maxUploadCount: 3,
    allowedUploadTypes: ['jpg', 'png', 'pdf'],
    minIntroLength: 5,
    maxIntroLength: 200,
    maxMessageLength: 120,
    searchEnabled: false,
    recommendationHint: '一期优先展示审核通过且人数未满的局。',
    texts: {
      subtitle: '你的信息将展示给发起人',
      wechatTitle: '微信信息',
      nicknameLabel: '昵称',
      introLabel: '自我介绍',
      introPlaceholder: '介绍你的背景、能力和参与动机',
      portfolioLabel: '相关经历/作品',
      messageLabel: '申请留言',
      messagePlaceholder: '给发起人留一句话',
      agreementPrefix: '我已阅读并同意',
      cancelText: '取消',
      submitText: '提交申请',
      profileSyncedText: '资料已同步',
      profilePendingText: '资料待同步',
      profileNameFallback: '待同步',
      loadFailedText: '入局申请配置加载失败',
      mediaUnsupportedText: '当前微信版本不支持选择图片',
      fileUnsupportedText: '当前微信版本不支持选择文件',
      imageTypeErrorText: '仅支持 JPG、PNG、GIF、WEBP 图片',
      fileTypeErrorText: '仅支持 PDF 文件',
      imageSelectedText: '图片已选择',
      fileSelectedText: '文件已选择',
      chooseFailedText: '选择失败，请重试',
      introRequiredText: '请先填写自我介绍',
      introMinTemplate: '自我介绍不少于{min}字',
      agreementRequiredText: '请先勾选平台协议',
      uploadRequiredText: '请先上传相关经历/作品',
      gameMissingText: '缺少局信息',
      submittingText: '提交中',
      submitSuccessText: '申请已提交',
      submitFailedText: '提交失败，请重试',
      maxUploadTemplate: '最多上传{max}个文件',
      navUnavailableText: '当前页暂无左右切换'
    },
    auditPage: {
      pageTitle: '审核申请列表',
      filters: [
        { key: 'all', name: '全部' },
        { key: 'pending', name: '待审核' },
        { key: 'approved', name: '已通过' },
        { key: 'rejected', name: '已拒绝' }
      ],
      statusTexts: {
        pending: '待审核',
        approved: '已通过',
        rejected: '已拒绝'
      },
      roleNames: {
        expert: '行家',
        guide: '领路人',
        main_guide: '主行家',
        player: '玩家',
        member: '玩家'
      },
      texts: {
        userFallbackTemplate: '用户{userId}',
        avatarFallback: '玩',
        applyTimeLabel: '申请时间',
        approveText: '通过申请',
        rejectText: '拒绝',
        reviewedText: '已完成审核',
        detailText: '详情',
        selectAllText: '全选',
        batchRejectText: '批量拒绝',
        batchApproveText: '批量通过',
        loadFailedText: '申请列表加载失败',
        approvingText: '通过中',
        rejectingText: '拒绝中',
        approveSuccessText: '已通过申请',
        rejectSuccessText: '已拒绝申请',
        reviewFailedText: '审核失败',
        emptyPendingText: '暂无待审核申请',
        batchApprovingText: '批量通过中',
        batchRejectingText: '批量拒绝中',
        batchApproveSuccess: '已批量通过',
        batchRejectSuccess: '已批量拒绝',
        batchReviewFailedText: '批量审核失败'
      },
      detail: {
        pageTitle: '审核组局',
        referralText: '已撮合双方意向',
        statusTitles: { pending: '等待你审核', approved: '已通过申请', rejected: '已拒绝申请' },
        countdownTexts: { pending: '待处理', approved: '已处理', rejected: '已处理' },
        playerStatusTexts: {
          pendingRequirement: '待确认需求',
          reviewedRequirement: '已完成审核',
          pending: '等待审核',
          approved: '申请已通过',
          rejected: '申请已拒绝'
        },
        texts: {
          playerTitle: '玩家信息',
          portfolioTitle: '相关经历/作品',
          confirmTitle: '局信息确认',
          optionTitle: '可选操作',
          noticeTitle: '确认须知',
          relationTitle: '组局关系图',
          expertName: '我',
          expertRoleText: '审核方',
          expertAvatarText: '我',
          guideAvatarFallback: '领',
          playerAvatarFallback: '玩',
          needPrefix: '申请说明：',
          remarkPrefix: '申请时间：',
          detailMissingText: '申请详情不存在',
          loadFailedText: '申请详情加载失败',
          emptyTitle: '申请详情未加载',
          emptyText: '请确认审核入口携带的申请 ID 是否有效，或返回申请列表重新打开。',
          mediaUnsupportedText: '当前微信版本不支持选择图片',
          fileUnsupportedText: '当前微信版本不支持选择文件',
          imageTypeErrorText: '仅支持 JPG、PNG、GIF、WEBP 图片',
          fileTypeErrorText: '仅支持 PDF 文件',
          filePathInvalidText: '文件路径无效',
          uploadingText: '上传中',
          imageUploadedText: '图片已上传',
          fileUploadedText: '文件已上传',
          uploadFailedText: '上传失败，请重试',
          chooseFailedText: '选择失败，请重试',
          detailRequiredActionText: '申请详情加载后才可以沟通',
          detailRequiredReviewText: '申请详情加载后才可以审核',
          chatPrefill: '你好，我想进一步确认本次组局申请。',
          timePrefill: '我建议进一步确认本次组局的具体时间，请看是否方便。',
          unavailableActionText: '请选择可用操作',
          approvingText: '通过中',
          rejectingText: '拒绝中',
          approveSuccessText: '已确认通过',
          rejectSuccessText: '已拒绝申请',
          reviewFailedText: '审核失败',
          actionTip: '确认后将建立三方连接群并冻结资金',
          actionLoadingText: '处理中...',
          confirmText: '确认通过'
        },
        sessionItems: [
          { key: 'topic', label: '组局主题', iconText: 'H', iconClass: 'topic' },
          { key: 'time', label: '时间', iconSrc: 'https://static.haowan.net.cn/miniprogram/pages/game/detail/assets/icon-clock.png', iconClass: 'time' },
          { key: 'location', label: '地点', actionText: '地图位置', iconSrc: 'https://static.haowan.net.cn/miniprogram/pages/game/detail/assets/icon-location.png', iconClass: 'place' }
        ],
        confirmRows: [
          { key: 'activityType', label: '活动类型' },
          { key: 'serviceDuration', label: '服务时长' },
          { key: 'clientBudget', label: '客户预算' },
          { key: 'platformFee', label: '平台' },
          { key: 'guideReward', label: '领路人' },
          { key: 'partnerReward', label: '生态合伙人' },
          { key: 'expertIncome', label: '你的收益' }
        ],
        optionalActions: [
          { key: 'time', name: '提议具体时间', iconSrc: 'https://static.haowan.net.cn/miniprogram/pages/game/audit-detail/assets/option-time.png' },
          { key: 'chat', name: '与玩家沟通', iconSrc: 'https://static.haowan.net.cn/miniprogram/pages/game/audit-detail/assets/option-chat.png' }
        ],
        noticeBullets: [
          '确认后请准时参加，如需取消请提前通知',
          '双方确认后组局正式生效，领路人将获得积分奖励',
          '请保持专业态度，维护平台信誉'
        ]
      }
    },
    version: '2026-06-30'
  }
}

function buildGameConditionRuleConfig() {
  return {
    enabled: true,
    visibleInMiniProgram: true,
    adminOnlyCreate: true,
    ruleItems: [
      { key: 'realname_verified', name: '完成实名认证', description: '玩家必须完成实名后才能申请条件局', required: true, order: 10 },
      { key: 'credit_min_80', name: '信用分不低于 80', description: '用于测试条件局的信用门槛', required: true, order: 20 }
    ],
    defaultVisibility: 'approved_users',
    reviewRequired: true,
    paymentRequired: false,
    version: '2026-06-30'
  }
}

function createMockGame(data = {}) {
  const title = String(data.title || '').trim()
  const gameType = String(data.gameType || 'free').trim() || 'free'
  const minPlayers = Number(data.minPlayers || 5)
  const maxPlayers = Number(data.maxPlayers || 8)

  if (!title) {
    return wait(fail(40001, '请输入局标题'))
  }

  if (gameType !== 'free') {
    return wait(fail(42201, '小程序端当前只能创建免费局'))
  }

  if (minPlayers < 5 || maxPlayers > 8 || minPlayers > maxPlayers) {
    return wait(fail(42202, '每局人数必须为 5-8'))
  }

  const id = Date.now()
  const game = {
    id,
    creatorUserId: 1,
    title,
    gameType,
    primaryCategory: String(data.primaryCategory || '').trim(),
    primaryCategoryText: String(data.primaryCategoryText || '').trim(),
    secondaryCategory: String(data.secondaryCategory || '').trim(),
    secondaryCategoryText: String(data.secondaryCategoryText || '').trim(),
    type: String(data.type || 'free').trim() || 'free',
    gameSource: 'app',
    status: 'pending_audit',
    minPlayers,
    maxPlayers,
    currentPlayers: 1,
    cityCode: String(data.cityCode || '').trim(),
    cityName: String(data.cityName || '').trim(),
    address: String(data.address || '').trim(),
    longitude: Number(data.longitude || 0),
    latitude: Number(data.latitude || 0),
    createdAt: new Date().toISOString()
  }

  mockCreatedGames = [game].concat(mockCreatedGames)

  return wait(ok(game))
}

function mockGameIdFromDetailPath(url) {
  const match = String(url || '').match(/^\/api\/app\/games\/(\d+)$/)

  return match ? Number(match[1]) : 0
}

function buildMockGameDetail(gameId) {
  const game = mockCreatedGames.find((item) => Number(item.id) === Number(gameId)) || {
    id: gameId,
    creatorUserId: 1,
    title: '企业数字化转型及技术服务沙龙局',
    gameType: 'free',
    gameSource: 'app',
    status: 'recruiting',
    minPlayers: 5,
    maxPlayers: 8,
    currentPlayers: 5,
    cityCode: '310000',
    cityName: '上海',
    address: '人民广场附近',
    longitude: 121.4737,
    latitude: 31.2304,
    createdAt: '2026-06-29T10:00:00+08:00'
  }

  return {
    ...game,
    memberIds: [1, 2, 3, 4, 5].slice(0, game.currentPlayers || 1),
    myRelation: {
      role: game.creatorUserId === 1 ? 'creator' : 'player',
      isCreator: game.creatorUserId === 1,
      isMember: true,
      canApply: game.status === 'recruiting',
      canAudit: game.creatorUserId === 1,
      canStart: false,
      canEnterIM: ['in_progress', 'pending_confirm', 'pending_review', 'completed'].indexOf(game.status) !== -1,
      canConfirm: game.status === 'pending_confirm',
      canReview: game.status === 'pending_review'
    },
    progress: {
      feedbacks: [],
      milestones: [],
      checkins: []
    },
    im: {
      available: ['in_progress', 'pending_confirm', 'pending_review', 'completed'].indexOf(game.status) !== -1,
      roomId: game.id,
      status: 'active',
      engine: 'openim',
      openIMGroupId: `mock-openim-game-${game.id}`
    },
    review: {
      reviewable: game.status === 'pending_review',
      complete: false,
      todos: []
    }
  }
}

function mockGameIdFromPath(url, suffix) {
  const path = String(url || '')
  const prefix = '/api/app/games/'

  if (!path.startsWith(prefix) || !path.endsWith(suffix)) {
    return 0
  }

  const idText = path.slice(prefix.length, path.length - suffix.length)
  return /^\d+$/.test(idText) ? Number(idText) : 0
}

function buildMockChatRoom(gameId) {
  return {
    id: gameId || 1,
    gameId,
    status: 'active',
    memberIds: [1, 2, 3, 4, 5],
    engine: 'openim',
    openIMGroupId: `mock-openim-game-${gameId || 1}`,
    createdAt: '2026-06-29T10:00:00+08:00'
  }
}

function buildMockChatSession(gameId) {
  return {
    engine: 'openim',
    imUserId: 'mock-im-user-1',
    openIMGroupId: `mock-openim-game-${gameId || 1}`,
    openIMToken: 'mock-openim-token'
  }
}

function buildMockChatMessages(gameId) {
  return {
    items: mockIMMessagesByGame[gameId] || []
  }
}

function buildMockCollaboration(gameId) {
  const messages = mockIMMessagesByGame[gameId] || []
  return {
    gameId,
    title: '局内协作',
    status: 'in_progress',
    statusText: '进行中',
    dayText: '第 6 天 / 共 15 天',
    progress: {
      percent: 46,
      title: '进度 46%',
      tasks: [
        { key: 'group-success', title: '已完成', desc: '组局已成功，成员已进入协作', state: 'completed' },
        { key: 'service-active', title: '进行中', desc: '成员正在协作推进交付', state: 'active' },
        { key: 'service-done', title: '待完成', desc: '确认服务、评价双方并归档', state: 'pending' }
      ]
    },
    members: [
      { id: 1, userId: 1, name: '佩娜', roleText: '玩家', avatarText: 'PE' },
      { id: 2, userId: 2, name: '张伟', roleText: '行家', avatarText: 'ZH' },
      { id: 3, userId: 3, name: '李娜', roleText: '领路人', avatarText: 'LI' }
    ],
    membersText: '佩娜 (玩家) · 张伟 (行家) · 李娜 (领路人)',
    messages: messages.map((item) => ({
      id: item.id,
      messageId: item.id,
      senderId: item.senderUserId,
      senderName: item.senderUserId === 1 ? '佩娜' : '成员',
      content: item.content,
      createdAt: item.createdAt
    })),
    actions: {
      canManageMembers: true,
      canEndGame: true,
      completionMode: 'ordered_confirm',
      manageRoute: `pages/game/participants/index?gameId=${gameId}`,
      endConfirmRoute: `pages/game/collaboration/index?gameId=${gameId}`,
      reviewRoute: `pages/game/review/index?gameId=${gameId}`
    }
  }
}

function requestMockGameCompletion(gameId) {
  return wait(ok({
    game: { id: gameId, gameSource: 'app', status: 'pending_confirm' },
    directReview: false,
    nextRoute: `/pages/game/collaboration/index?gameId=${gameId}`
  }))
}

function sendMockChatMessage(gameId, data = {}) {
  const content = String(data.content || '').trim()
  const messageType = String(data.messageType || 'text').trim() || 'text'

  if (!content && messageType === 'text') {
    return wait(fail(40001, '请输入消息内容'))
  }

  const messages = mockIMMessagesByGame[gameId] || []
  const message = {
    id: messages.length + 1,
    roomId: gameId || 1,
    gameId,
    senderUserId: 1,
    messageType,
    content,
    fileId: Number(data.fileId || 0),
    status: 'sent',
    createdAt: new Date().toISOString()
  }

  mockIMMessagesByGame[gameId] = messages.concat(message)

  return wait(ok(message))
}

function confirmMockService(gameId, data = {}) {
  if (!gameId) {
    return wait(fail(40001, '缺少局ID'))
  }

  const fileIds = Array.isArray(data.fileIds) ? data.fileIds.filter(Boolean) : []

  return wait(ok({
    confirm: {
      id: gameId,
      gameId,
      status: 'pending',
      confirmedBy: [1],
      createdAt: new Date().toISOString()
    },
    items: [
      {
        id: 1,
        confirmId: gameId,
        gameId,
        userId: 1,
        note: String(data.note || '').trim(),
        fileIds,
        createdAt: new Date().toISOString()
      }
    ],
    game: {
      id: gameId,
      status: 'pending_confirm'
    }
  }))
}

function buildMockAvailableReviews() {
  const reward = { show: false }
  const candidates = [
    {
      id: 'mock-review-1',
      gameId: 1,
      targetUserId: 2,
      targetRole: 'member',
      targetName: '用户A',
      targetNickname: '用户A',
      deadlineAt: '2026-06-30T20:00:00+08:00'
    }
  ]
  return {
    reward,
    reviewPage: mockReviewPageConfig,
    items: candidates.filter((item) => (
      item.targetUserId !== Number(mockCurrentUser.id)
      && !mockSubmittedReviews.some((review) => (
        Number(review.reviewerUserId) === Number(mockCurrentUser.id)
        && Number(review.gameId) === Number(item.gameId)
        && Number(review.targetUserId) === Number(item.targetUserId)
      ))
    ))
  }
}

function submitMockReview(data = {}) {
  const gameId = Number(data.gameId || 0)
  const targetUserId = Number(data.targetUserId || 0)
  const score = Number(data.score || 0)

  if (!gameId || !targetUserId || targetUserId === Number(mockCurrentUser.id) || score < 1 || score > 5) {
    return wait(fail(42203, '评价参数错误'))
  }

  if (mockSubmittedReviews.some((item) => (
    Number(item.reviewerUserId) === Number(mockCurrentUser.id)
    && Number(item.gameId) === gameId
    && Number(item.targetUserId) === targetUserId
  ))) {
    return wait(fail(40932, '重复评价'))
  }

  const review = {
    id: mockSubmittedReviews.length + 1,
    gameId,
    reviewerUserId: Number(mockCurrentUser.id),
    targetUserId,
    targetRole: String(data.targetRole || 'member'),
    score,
    content: String(data.content || '').trim(),
    tags: Array.isArray(data.tags) ? data.tags : [],
    againIntent: String(data.againIntent || 'yes'),
    createdAt: new Date().toISOString()
  }

  mockSubmittedReviews.push(review)

  return wait(ok({
    review,
    profile: {
      userId: 1,
      creditScore: 100 + mockSubmittedReviews.length
    }
  }))
}

function submitMockReport(data = {}) {
  const gameId = Number(data.gameId || 0)
  const reportType = String(data.reportType || '').trim()
  const content = String(data.content || '').trim()
  const validTypes = ['service_dispute', 'im_message', 'review_dispute', 'revenue_dispute', 'user_complaint', 'other']

  if (!gameId || validTypes.indexOf(reportType) === -1 || !content) {
    return wait(fail(42204, '举报参数错误'))
  }

  const report = {
    id: mockSubmittedReports.length + 1,
    gameId,
    reporterUserId: 1,
    targetUserId: Number(data.targetUserId || 0),
    reportType,
    content,
    status: 'pending',
    chatMessageId: Number(data.chatMessageId || 0),
    fileId: Number(data.fileId || 0),
    reviewId: Number(data.reviewId || 0) || data.reviewId || '',
    revenueRecordId: Number(data.revenueRecordId || 0),
    revenueFrozen: reportType === 'revenue_dispute',
    revenueFreezeNote: reportType === 'revenue_dispute' ? 'report_created' : '',
    createdAt: new Date().toISOString()
  }

  mockSubmittedReports.push(report)

  return wait(ok(report))
}

function submitMockAppeal(reportId, data = {}) {
  const report = mockSubmittedReports.find((item) => item.id === reportId)
  const content = String(data.content || '').trim()
  const fileIds = Array.isArray(data.fileIds) ? data.fileIds.filter(Boolean) : []
  const firstFileId = Number(data.fileId || fileIds[0] || report && report.fileId || 0)

  if (!report) {
    return wait(fail(40404, '举报不存在'))
  }

  if (!content) {
    return wait(fail(42205, '申诉内容不能为空'))
  }

  report.status = 'appealed'
  report.handleResult = content
  report.handledAt = new Date().toISOString()
  report.fileId = firstFileId
  report.appealFileIds = fileIds.length ? fileIds : (firstFileId ? [firstFileId] : [])

  return wait(ok(report))
}

function submitMockCreditAppeal(data = {}) {
  const creditLogId = Number(data.creditLogId || 0)
  const content = String(data.content || '').trim()
  const log = mockCreditLogs.find((item) => item.id === creditLogId)
  const fileIds = Array.isArray(data.fileIds) ? data.fileIds.filter(Boolean) : []
  const firstFileId = Number(data.fileId || fileIds[0] || 0)

  if (!log) {
    return wait(fail(40404, '信用记录不存在'))
  }

  if (!content) {
    return wait(fail(42205, '申诉内容不能为空'))
  }

  const report = {
    id: mockSubmittedReports.length + 1,
    gameId: log.gameId || 0,
    reporterUserId: 1,
    targetUserId: 1,
    reportType: 'credit_appeal',
    content,
    status: 'appealed',
    creditLogId,
    fileId: firstFileId,
    appealFileIds: fileIds.length ? fileIds : (firstFileId ? [firstFileId] : []),
    handleResult: content,
    handledAt: new Date().toISOString(),
    createdAt: new Date().toISOString()
  }

  mockSubmittedReports.push(report)

  return wait(ok(report))
}

function withdrawMockAppeal(reportId) {
  const report = mockSubmittedReports.find((item) => item.id === reportId)

  if (!report) {
    return wait(fail(40404, '举报不存在'))
  }

  if (report.status !== 'appealed') {
    return wait(fail(42204, '当前申诉不可撤回'))
  }

  report.status = 'appeal_withdrawn'
  report.handledAt = new Date().toISOString()

  return wait(ok(report))
}

function buildMockCreditCenter() {
  const creditRecords = mockCreditLogs.map((item) => Object.assign({}, item, {
    id: `credit-${item.id}`,
    creditLogId: item.id
  }))
  const reportRecords = mockSubmittedReports
    .filter((item) => item.status === 'handled' || item.status === 'appealed' || item.status === 'appeal_withdrawn')
    .slice()
    .reverse()
    .map((item) => ({
      id: `report-${item.id}`,
      title: item.status === 'appealed' ? '申诉处理中' : (item.status === 'appeal_withdrawn' ? '申诉已撤回' : '举报处理'),
      desc: `${item.reportType || 'other'} · 局ID ${item.gameId || 0}`,
      score: item.status === 'appealed' || item.status === 'appeal_withdrawn' ? '+0' : '-2',
      tone: item.status === 'appealed' || item.status === 'appeal_withdrawn' ? 'plus' : 'minus',
      reportId: item.id,
      createdAt: item.handledAt || item.createdAt
    }))
  const records = creditRecords.concat(reportRecords)

  return {
    score: 95,
    scoreLabel: '信用分',
    todayScore: 100,
    level: '优秀',
    monthlyDelta: '+0',
    summary: [
      { label: '信用等级', value: '优秀' },
      { label: '奖励中心', value: '0条待查看' },
      { label: '惩罚中心', value: `${records.filter((item) => item.tone === 'minus').length}条记录` }
    ],
    records,
    appealEntry: {
      enabled: records.length > 0,
      route: records[0] && records[0].reportId
        ? `/pages/profile/system-management/credit-appeal/index?reportId=${records[0].reportId}`
        : (records[0] && records[0].creditLogId ? `/pages/profile/system-management/credit-appeal/index?creditLogId=${records[0].creditLogId}` : '/pages/profile/system-management/credit-appeal/index'),
      text: '信用申诉'
    },
    bottomNote: '信用分低于80分将限制部分功能，低于60分将暂停服务资格。'
  }
}

function submitMockMemberRadarAction(data = {}) {
  const action = String(data.action || '').trim()
  const allowedActions = ['start', 'save', 'next', 'follow', 'profile', 'rematch']

  if (!allowedActions.includes(action)) {
    return wait(fail(422, '雷达操作无效'))
  }

  const targetId = String(data.targetId || '').trim()
  const targetUserId = Number(data.targetUserId || 0) || 0

  if (action === 'save') {
    mockMemberRadarState.savedProfile = {
      formRows: Array.isArray(data.formRows) ? data.formRows : [],
      matchCriteria: data.matchCriteria || {}
    }
  }

  if (action === 'start' || action === 'rematch' || action === 'next') {
    mockMemberRadarState.lastMatch = {
      targetId,
      targetUserId,
      matchMode: data.matchMode || 'all',
      matchCriteria: data.matchCriteria || {}
    }
  }

  if (action === 'follow') {
    const key = targetId || String(targetUserId || '')
    if (key && !mockMemberRadarState.followed.some((item) => item.targetId === key)) {
      mockMemberRadarState.followed.push({ targetId: key, targetUserId })
    }
  }

  return wait(ok({
    action,
    status: 'ok',
    route: action === 'profile' && (targetUserId || targetId)
      ? `/pages/profile/service-center/invite/member-detail/index?id=${encodeURIComponent(targetUserId || targetId)}`
      : '',
    profile: mockMemberRadarState.savedProfile,
    last: mockMemberRadarState.lastMatch
  }))
}

function pagedItems(items = [], params = {}) {
  const page = Math.max(1, Number(params.page || 1) || 1)
  const pageSize = Math.min(100, Math.max(1, Number(params.pageSize || items.length || 20) || 20))
  const total = items.length
  const start = (page - 1) * pageSize
  const end = Math.min(total, start + pageSize)

  return {
    items: items.slice(start, end),
    page,
    pageSize,
    total,
    hasMore: end < total
  }
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

  if (method === 'POST' && url === '/api/app/files/upload-token') {
    return createMockUploadToken(options.data || {})
  }

  {
    const downloadMatched = String(url || '').match(/^\/api\/app\/files\/(\d+)\/download-url$/)
    if (method === 'GET' && downloadMatched) {
      return buildMockDownloadURL(Number(downloadMatched[1]))
    }
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

  if (method === 'GET' && url === '/api/app/game-applications/received') {
    return wait(ok(buildReceivedGameApplications(options.data || {})))
  }

  if (method === 'POST' && /^\/api\/app\/game-applications\/[^/]+\/audit$/.test(url)) {
    const applicationId = decodeURIComponent(url.split('/').slice(-2)[0] || '')

    return reviewMockGameApplication(applicationId, options.data || {})
  }

  if (method === 'GET' && url === '/api/app/home') {
    return wait(ok(buildHome(options.data || {})))
  }

  if (method === 'GET' && url === '/api/app/games/nearby') {
    return wait(ok(buildNearbyGames(options.data || {})))
  }

  if (method === 'GET' && url === '/api/app/map/search') {
    return wait(buildMockMapSearch(options.data || {}))
  }

  if (method === 'GET' && url === '/api/app/map/index-config') {
    return wait(ok(buildMockMapIndexConfig()))
  }

  if (method === 'GET' && url === '/api/app/map/play-pages') {
    const config = buildMockMapPlayPages()
    const pageKey = options.data && options.data.pageKey

    return wait(ok(pageKey ? config.pages[pageKey] : config))
  }

  if (method === 'POST' && url === '/api/app/map/checkins') {
    return wait(ok({
      id: Date.now(),
      pointId: options.data && options.data.pointId || '',
      challengeId: options.data && options.data.challengeId || '',
      gameId: options.data && options.data.gameId || '',
      status: 'pending_audit',
      statusText: '待审核',
      statusDesc: '打卡材料已提交，等待后台审核',
      createdAt: new Date().toISOString()
    }))
  }

  if (method === 'POST' && url === '/api/app/map/blind-routes') {
    return wait(ok({
      id: Date.now(),
      cardId: options.data && options.data.cardId || '',
      title: options.data && options.data.title || '盲盒路线',
      timeText: '刚刚开启',
      status: 'opened',
      statusText: '进行中',
      createdAt: new Date().toISOString()
    }))
  }

  if (method === 'POST' && /^\/api\/app\/map\/blind-routes\/[^/]+\/complete$/.test(url)) {
    const routeId = decodeURIComponent(url.split('/').slice(-2)[0] || '')

    return wait(ok({
      id: Number(routeId) || routeId,
      title: '盲盒路线',
      timeText: '刚刚完成',
      status: 'completed',
      statusText: '已完成',
      createdAt: new Date().toISOString()
    }))
  }

  if (method === 'POST' && url === '/api/app/map/challenges') {
    return wait(ok({
      id: Date.now(),
      title: options.data && options.data.title || '城市挑战',
      timeLeft: '24小时',
      statusText: '进行中',
      statusTone: 'active',
      selfValue: 0,
      rivalValue: 0,
      total: options.data && options.data.total || 3,
      selfPercent: 0,
      rivalPercent: 0,
      createdAt: new Date().toISOString()
    }))
  }

  if (method === 'GET' && url === '/api/app/map/my-city') {
    return wait(ok(mockMyCityConfig))
  }

  if (method === 'POST' && url === '/api/app/games') {
    return createMockGame(options.data || {})
  }

  {
    const gameId = mockGameIdFromPath(url, '/members')
    if (method === 'GET' && gameId) {
      return wait(ok({
        game: buildMockGameDetail(gameId).game,
        total: 5,
        items: [
          { userId: 1, role: 'creator', roleLabel: '发起人', isCreator: true, isCurrentUser: true, confirmed: false },
          { userId: 2, role: 'expert', roleLabel: '行家', confirmed: false },
          { userId: 3, role: 'guide', roleLabel: '领路人', confirmed: false },
          { userId: 4, role: 'player', roleLabel: '玩家', confirmed: true },
          { userId: 5, role: 'player', roleLabel: '玩家', confirmed: true }
        ]
      }))
    }
  }

  {
    const gameId = mockGameIdFromDetailPath(url)
    if (method === 'GET' && gameId) {
      return wait(ok(buildMockGameDetail(gameId)))
    }
  }

  {
    const gameId = mockGameIdFromPath(url, '/favorite')
    if (method === 'POST' && gameId) {
      return wait(ok({
        gameId,
        favorited: true,
        game: buildMockGameDetail(gameId).game
      }))
    }
  }

  if (method === 'GET' && url === '/api/app/relations/network-home') {
    return wait(ok(mockRelationNetworkHome))
  }

  if (method === 'GET' && url === '/api/app/messages/trade-warning') {
    return wait(ok(buildTradeWarningDetail(options.data || {})))
  }

  if (method === 'POST' && url === '/api/app/messages/trade-warning/actions') {
    return submitMockTradeWarningAction(options.data || {})
  }

  if (method === 'GET' && url === '/api/app/messages/center') {
    return wait(ok(buildMessageCenter(options.data || {})))
  }

  if (method === 'GET' && url === '/api/app/messages/my-config') {
    return wait(ok(JSON.parse(JSON.stringify(mockMessageMyConfig))))
  }

  if (method === 'GET' && url === '/api/app/messages/system-notification') {
    return wait(ok(buildSystemNotificationDetail(options.data || {})))
  }

  if (method === 'POST' && url === '/api/app/messages/system-notification/feedback') {
    return submitMockSystemNotificationFeedback(options.data || {})
  }

  if (method === 'GET' && url === '/api/app/profile/home') {
    return wait(ok(mockProfileHome))
  }

  if (method === 'GET' && url === '/api/app/profile/assets') {
    return wait(ok(mockProfileAssets))
  }

  if (method === 'GET' && url === '/api/app/profile/credit-center') {
    return wait(ok(buildMockCreditCenter()))
  }

  if (method === 'GET' && url === '/api/app/membership/radar-config') {
    return wait(ok(mockMemberRadarConfig))
  }

  if (method === 'POST' && url === '/api/app/membership/radar/actions') {
    return submitMockMemberRadarAction(options.data || {})
  }

  if (method === 'GET' && url === '/api/app/points/summary') {
    return wait(ok(buildPointsSummary()))
  }

  if (method === 'GET' && url === '/api/app/points/logs') {
    return wait(ok({
      items: buildPointsLogs()
    }))
  }

  if (method === 'GET' && (url === '/api/app/growth/my' || url === '/api/app/users/me/growth')) {
    return wait(ok({
      profile: {
        userId: mockCurrentUser.id,
        level: 5,
        experience: 350,
        creditScore: 100,
        todayCreditScore: 100,
        availablePoints: mockPointsMall.pointsAvailable,
        reviewCount: 8,
        achievements: [],
        updatedAt: '2026-06-30T10:00:00+08:00'
      },
      footprints: [
        { userId: mockCurrentUser.id, gameId: 1, action: 'completed_game', createdAt: '2026-06-30T10:00:00+08:00' },
        { userId: mockCurrentUser.id, gameId: 1, action: 'submitted_review', createdAt: '2026-06-30T10:05:00+08:00' }
      ],
      achievements: [],
      achievementConfig: {
        filters: [
          { key: 'all', label: '全部', order: 10 }
        ],
        locked: [],
        season: {
          title: '赛季',
          status: '进行中',
          remainTpl: '{footprintCount} 条记录'
        },
        onlineSuffix: '人成长中',
        levelTitlePrefix: 'Lv.',
        version: '2026-07-01'
      }
    }))
  }

  if (method === 'GET' && url === '/api/app/profile/system-management/profile-info') {
    return wait(ok(buildSystemProfileInfo()))
  }

  if (method === 'GET' && url === '/api/app/enterprise-certification') {
    const item = mockSystemProfileInfoState && mockSystemProfileInfoState.enterpriseCertification
    return wait(ok(item || { status: 'none', items: [] }))
  }

  if (method === 'POST' && url === '/api/app/enterprise-certification') {
    return submitMockEnterpriseCertification(options.data || {})
  }

  if (method === 'GET' && url === '/api/app/profile/settings') {
    return wait(ok(buildProfileSettings()))
  }

  if (method === 'PUT' && url === '/api/app/profile/settings') {
    return saveProfileSettings(options.data || {})
  }

  if (method === 'GET' && url === '/api/app/profile/system-management/block-settings') {
    return wait(ok(buildSystemBlockSettings()))
  }

  if (method === 'PUT' && url === '/api/app/profile/system-management/block-settings') {
    return saveSystemBlockSettings(options.data || {})
  }

  if (method === 'GET' && url === '/api/app/profile/agreements') {
    return wait(ok({
      items: buildProfileAgreements()
    }))
  }

  if (method === 'GET' && /^\/api\/app\/profile\/agreements\/[^/]+$/.test(url)) {
    const agreementKey = decodeURIComponent(url.split('/').pop() || '')
    const agreement = getMockAgreement(agreementKey)

    return agreement ? wait(ok(agreement)) : wait(fail(40404, '协议不存在'))
  }

  if (method === 'POST' && /^\/api\/app\/profile\/agreements\/[^/]+\/sign$/.test(url)) {
    const agreementKey = decodeURIComponent(url.split('/').slice(-2)[0] || '')

    return signMockAgreement(agreementKey)
  }

  if (method === 'GET' && url === '/api/app/profile/system-management/feedback') {
    return wait(ok(buildSystemFeedbackHome()))
  }

  if (method === 'POST' && url === '/api/app/profile/system-management/feedback') {
    return submitMockSystemFeedback(options.data || {})
  }

  if (method === 'GET' && /^\/api\/app\/profile\/system-management\/feedback-records\/[^/]+$/.test(url)) {
    const recordId = decodeURIComponent(url.split('/').pop() || '')

    return mockFeedbackDetail(recordId)
  }

  if (method === 'POST' && /^\/api\/app\/profile\/system-management\/feedback-records\/[^/]+\/messages$/.test(url)) {
    const recordId = decodeURIComponent(url.split('/').slice(-2)[0] || '')

    return appendMockFeedbackMessage(recordId, options.data || {})
  }

  if (method === 'GET' && url === '/api/app/profile/system-management/feedback-records') {
    return wait(ok(buildMockFeedbackRecords(options.data || {})))
  }

  if (method === 'GET' && url === '/api/app/profile/system-management/skill-config') {
    return wait(ok(buildSystemSkillConfig()))
  }

  if (method === 'GET' && /^\/api\/app\/profile\/system-management\/service-cases\/[^/]+$/.test(url)) {
    const caseId = decodeURIComponent(url.split('/').pop() || '')
    const detail = buildMockServiceCaseDetail(caseId)

    return detail ? wait(ok(detail)) : wait(fail(40404, '服务案例不存在'))
  }

  if (method === 'GET' && url === '/api/app/profile/service-center/invite/overview') {
    return wait(ok(buildMockInviteOverview()))
  }

  if (method === 'GET' && url === '/api/app/profile/service-center/invite/network') {
    return wait(ok(buildMockInviteNetwork()))
  }

  if (method === 'GET' && url === '/api/app/profile/service-center/invite/records') {
    return wait(ok(buildMockInviteRecords(options.data || {})))
  }

  if (method === 'GET' && url === '/api/app/profile/service-center/invite/ranking') {
    return wait(ok(buildMockInviteRanking(options.data || {})))
  }

  if (method === 'GET' && url === '/api/app/profile/service-center/invite/income') {
    return wait(ok(buildMockInviteIncome()))
  }

  if (method === 'GET' && url === '/api/app/profile/service-center/invite/member-detail') {
    const memberId = options.data && (options.data.memberId || options.data.id) || ''

    return wait(ok(buildMockInviteMemberDetail(memberId)))
  }

  if (method === 'GET' && (url === '/api/app/profile/points/mall' || url === '/api/app/redemption/items')) {
    return wait(ok(buildPointsMall()))
  }

  if (method === 'GET' && (url === '/api/app/profile/points/orders' || url === '/api/app/redemption/orders/my')) {
    return wait(ok(buildPointsOrders(options.data || {})))
  }

  if (method === 'GET' && /^\/api\/app\/redemption\/orders\/[^/]+$/.test(url)) {
    const orderId = decodeURIComponent(url.split('/').pop() || '')
    const detail = buildPointsOrderDetail(orderId)

    if (!detail) {
      return wait(fail(40404, '兑换订单不存在'))
    }

    return wait(ok(detail))
  }

  if (method === 'GET' && /^\/api\/app\/profile\/points\/orders\/[^/]+\/logistics$/.test(url)) {
    const orderId = decodeURIComponent(url.split('/').slice(-2)[0] || '')

    if (!orderId) {
      return wait(fail(40001, '缺少订单信息'))
    }

    return wait(ok(buildPointsOrderLogistics(orderId)))
  }

  if (method === 'POST' && /^\/api\/app\/redemption\/orders\/[^/]+\/cancel$/.test(url)) {
    const orderId = decodeURIComponent(url.split('/').slice(-2)[0] || '')

    return cancelMockPointsOrder(orderId)
  }

  if (method === 'POST' && (url === '/api/app/profile/points/mall/exchange' || url === '/api/app/redemption/orders')) {
    return exchangePointsMallGood(options.data || {})
  }

  if (method === 'GET' && url === '/api/app/profile/service-center/reviews') {
    return wait(ok(buildMockServiceReviews()))
  }

  if (method === 'GET' && /^\/api\/app\/profile\/service-center\/reviews\/[^/]+$/.test(url)) {
    const reviewId = decodeURIComponent(url.split('/').pop() || '')
    const detail = buildMockServiceReviewDetail(reviewId)

    return detail ? wait(ok(detail)) : wait(fail(40404, '评价不存在'))
  }

  if (method === 'POST' && /^\/api\/app\/profile\/service-center\/reviews\/[^/]+\/reply$/.test(url)) {
    const reviewId = decodeURIComponent(url.split('/').slice(-2)[0] || '')

    return submitMockServiceReviewReply(reviewId, options.data || {})
  }

  if (method === 'POST' && /^\/api\/app\/profile\/service-center\/reviews\/[^/]+\/like$/.test(url)) {
    const reviewId = decodeURIComponent(url.split('/').slice(-2)[0] || '')

    return wait(ok({
      reviewId,
      liked: true
    }))
  }

  if (method === 'POST' && /^\/api\/app\/profile\/service-center\/reviews\/[^/]+\/actions$/.test(url)) {
    const reviewId = decodeURIComponent(url.split('/').slice(-2)[0] || '')

    return wait(ok({
      reviewId,
      actions: [
        { key: 'detail', text: '查看关联局', route: '/pages/game/detail/index?id=1' },
        { key: 'reply', text: '回复评价', route: `/pages/profile/service-center/manage/review-reply/index?id=${encodeURIComponent(reviewId)}` },
        { key: 'intervention', text: '申请平台介入', route: '/pages/profile/system-management/report-center/index' }
      ]
    }))
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

  if (method === 'GET' && url === '/api/app/role-applications/status-config') {
    return wait(ok(mockRoleStatusPageConfig))
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
        { name: '+自定义', custom: true }
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

  if (method === 'GET' && url === '/api/app/role-applications/guide/config') {
    return wait(ok(mockGuideApplyConfig))
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

  if (method === 'POST' && url === '/api/app/game-invites/reminders') {
    return wait(ok({
      invitationId: options.data && options.data.invitationId || `mock_invite_${Date.now()}`,
      gameId: options.data && options.data.gameId || '',
      targetUserId: 2,
      notification: {
        id: Date.now(),
        notifyType: 'game_invitation_remind',
        bizType: 'game_invitation',
        bizId: options.data && options.data.invitationId || '',
        status: 'unread'
      }
    }))
  }

  if (method === 'GET' && url === '/api/app/game-invites/guide-progress') {
    return wait(ok(mockGuideProgress))
  }

  if (method === 'GET' && url === '/api/app/game-invites/referral-records') {
    return wait(ok({
      pageConfig: mockReferralRecordsConfig,
      summary: {
        label: mockReferralRecordsConfig.summary.label,
        amount: 0,
        amountText: '¥0',
        background: mockReferralRecordsConfig.summary.background,
        iconSrc: mockReferralRecordsConfig.summary.iconSrc,
        stats: [
          { key: 'success', text: '成功 0单' },
          { key: 'processing', text: '进行中 0单' },
          { key: 'review', text: '待评价 0单' }
        ]
      },
      tabs: mockReferralRecordsConfig.tabs.map((item) => ({ ...item, count: 0 })),
      records: []
    }))
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

  if (method === 'GET' && url === '/api/app/games/favorites/my') {
    return wait(ok({
      pageConfig: mockMyGamesPageConfig,
      items: (mockGameManage.orders || []).slice(0, 1).map((item, index) => ({
        id: `mock-favorite-${index + 1}`,
        gameId: item.gameId || index + 1,
        game: {
          id: item.gameId || index + 1,
          title: item.serviceTitle || item.title || item.serviceText || item.name || '收藏组局',
          status: item.status || item.statusType || 'recruiting',
          createdAt: item.startedAt || item.createdAt || '2026-03-20T14:30:00+08:00'
        },
        createdAt: item.createdAt || '2026-03-20T14:30:00+08:00'
      }))
    }))
  }

  if (method === 'GET' && url === '/api/app/games/category-config') {
    return wait(ok(buildGameCategoryConfig()))
  }

  if (method === 'GET' && url === '/api/app/games/application-config') {
    return wait(ok(buildGameApplicationConfig()))
  }

  if (method === 'GET' && url === '/api/app/games/condition-rule-config') {
    return wait(ok(buildGameConditionRuleConfig()))
  }

  if (method === 'GET' && url === '/api/app/role-applications/benefit-config') {
    return wait(ok(mockRoleBenefitConfig))
  }

  if (method === 'GET' && url === '/api/app/games/cancel-config') {
    return wait(ok(mockGameCancelConfig))
  }

  if (method === 'GET' && url === '/api/app/games/profit-templates') {
    return wait(ok(mockGameProfitTemplates))
  }

  if (method === 'GET' && url === '/api/app/reviews/available') {
    return wait(ok(buildMockAvailableReviews()))
  }

  if (method === 'GET' && url === '/api/app/reviews/complete-config') {
    return wait(ok({
      benefits: [],
      playOptions: [],
      version: '2026-07-01'
    }))
  }

  if (method === 'POST' && url === '/api/app/reviews') {
    return submitMockReview(options.data || {})
  }

  if (method === 'GET' && url === '/api/app/reports/config') {
    return wait(ok(mockReportCenterConfig))
  }

  if (method === 'GET' && url === '/api/app/reports/my') {
    return wait(ok(pagedItems(mockSubmittedReports.slice().reverse(), options.data || {})))
  }

  if (method === 'GET' && url === '/api/app/reports/appeals/my') {
    return wait(ok(pagedItems(
      mockSubmittedReports.filter((item) => item.status === 'appealed').slice().reverse(),
      options.data || {}
    )))
  }

  if (method === 'POST' && /^\/api\/app\/reports\/\d+\/appeal$/.test(url)) {
    const reportId = Number(url.split('/').slice(-2)[0])

    return submitMockAppeal(reportId, options.data || {})
  }

  if (method === 'POST' && url === '/api/app/profile/credit-appeals') {
    return submitMockCreditAppeal(options.data || {})
  }

  if (method === 'POST' && /^\/api\/app\/reports\/\d+\/appeal\/withdraw$/.test(url)) {
    const reportId = Number(url.split('/').slice(-3)[0])

    return withdrawMockAppeal(reportId)
  }

  if (method === 'GET' && /^\/api\/app\/reports\/\d+$/.test(url)) {
    const reportId = Number(url.split('/').pop())
    const report = mockSubmittedReports.find((item) => item.id === reportId)

    return wait(report ? ok(report) : fail(40404, '举报不存在'))
  }

  if (method === 'POST' && url === '/api/app/reports') {
    return submitMockReport(options.data || {})
  }

  {
    const gameId = mockGameIdFromPath(url, '/chat-room')
    if (method === 'GET' && gameId) {
      return wait(ok(buildMockChatRoom(gameId)))
    }
  }

  {
    const roomMatched = String(url || '').match(/^\/api\/app\/chat\/rooms\/(\d+)\/messages$/)
    if (roomMatched) {
      const roomId = Number(roomMatched[1])
      if (method === 'GET') {
        return wait(ok(buildMockChatMessages(roomId)))
      }
      if (method === 'POST') {
        return sendMockChatMessage(roomId, options.data || {})
      }
    }
  }

  {
    const gameId = mockGameIdFromPath(url, '/success-detail')
    if (method === 'GET' && gameId) {
      return wait(ok({
        viewer: { userId: 1, role: 'expert' },
        completion: {
          role: 'expert',
          canConfirm: true,
          hasConfirmed: false,
          allExpertsConfirmed: false,
          waitingText: '请确认本次服务已经完成'
        },
        group: {
          title: '产品架构咨询 - 三方群',
          roles: '行家、领路人、玩家',
          hint: '领路人将持续跟进活动进度，确保双方顺利对接',
          gameId
        },
        participants: [
          { id: 'expert', name: '行家', avatarText: 'EX', colorClass: 'blue', roleLabel: '行家' },
          { id: 'guide', name: '领路人', avatarText: 'GU', colorClass: 'orange', roleLabel: '领路人' },
          { id: 'player', name: '玩家', avatarText: 'PL', colorClass: 'pink', roleLabel: '玩家' }
        ],
        activityRows: [
          { label: '活动编号', value: `GAME-${gameId}` },
          { label: '创建时间', value: '2026-03-23 10:23' },
          { label: '组局时间', value: '2026-03-23 14:30' },
          { label: '当前阶段', value: '待交付服务', highlight: true }
        ],
        fund: {
          title: '资金已托管',
          desc: '服务完成后自动结算',
          amount: 80000,
          amountText: '¥800',
          status: 'escrowed',
          progress: 33,
          stepLabels: ['已托管', '服务中', '已完成']
        },
        nextSteps: [
          { index: 1, title: '联系玩家确认具体时间', desc: '建议24小时内完成', active: true, action: 'contact_player' },
          { index: 2, title: '按时交付服务', desc: '等待确认时间', action: 'deliver' },
          { index: 3, title: '确认完成并收款', desc: '等待服务完成', action: 'confirm' }
        ],
        deliveryPage: mockDeliveryPageConfig,
        deliveryProof: {
          maxCount: 4,
          emptyText: '可上传服务完成截图、交付材料截图等凭证，最多 4 张。',
          fullText: '最多上传 4 张凭证',
          selectedTemplate: '已选择 {selected}/{max} 张凭证',
          uploadActionText: '上传交付凭证',
          contactPlayerText: '联系玩家',
          contactGuideText: '联系领路人',
          cancelServiceText: '申请取消服务',
          unavailableTextTemplate: '{action}暂不可用',
          contactPlayerPrefill: '你好，麻烦确认一下服务完成情况。',
          contactGuidePrefill: '你好，辛苦帮忙同步一下服务完成状态。',
          timelinePrefill: '你好，服务已经完成，麻烦确认一下。',
          idleTimelineText: '当前节点无需额外操作',
          timelineActionText: '联系玩家',
          proofNoteTemplate: '交付凭证ID：{fileIds}'
        }
      }))
    }
  }

  {
    const gameId = mockGameIdFromPath(url, '/guide-success-detail')
    if (method === 'GET' && gameId) {
      return wait(ok({
        gameId,
        viewer: { userId: 1, role: 'guide' },
        timeline: [
          { title: '发起引荐', desc: '你向双方发送了组局邀请', time: '03-21 10:23' },
          { title: '玩家确认', desc: '李娜确认参加组局', time: '03-21 11:05' },
          { title: '行家确认', desc: '王强确认参加组局', time: '03-21 14:30' },
          { title: '组局成功！', desc: '双方已建立连接，进入交付阶段', time: '03-21 14:30', active: true }
        ],
        party: {
          confirmedText: '',
          cardClass: 'success-guide-party-card',
          cardStyle: 'width: 684rpx; height: 446rpx; margin: 32rpx auto 0; box-shadow: none;',
          titleClass: 'regular',
          player: { id: 'player', name: '李娜', avatarText: 'LI', role: '玩家', avatarClass: 'pink', state: '已确认', stateClass: 'confirmed' },
          expert: { id: 'expert', name: '王强', avatarText: 'WA', role: '行家', avatarClass: 'blue', state: '已确认', stateClass: 'confirmed' }
        },
        followUps: [
          { key: 'schedule', title: '查看组局日程', desc: '查看该局详情与交付进度', iconSrc: './assets/follow-schedule.png', iconClass: 'schedule', theme: 'blue', target: { type: 'game_detail', gameId } },
          { key: 'feedback', title: '询问双方反馈', desc: '进入三方群了解交流情况', iconSrc: './assets/follow-feedback.png', iconClass: 'feedback', theme: 'purple', reward: '+20积分', target: { type: 'im_room', gameId, prefill: '我来跟进一下本次组局双方反馈。' } },
          { key: 'deal', title: '促成交易', desc: '记录或跟进双方合作意向', iconSrc: './assets/follow-deal.png', iconClass: 'deal', theme: 'orange', reward: '+100积分', target: { type: 'referral_record', gameId } }
        ]
      }))
    }
    if (method === 'POST' && gameId) {
      const action = String(options.data && options.data.action || '').trim()
      const targetMap = {
        schedule: { type: 'game_detail', gameId },
        feedback: { type: 'im_room', gameId, prefill: '我来跟进一下本次组局双方反馈。' },
        deal: { type: 'referral_record', gameId }
      }

      if (!targetMap[action]) {
        return wait(fail(40001, '跟进动作错误'))
      }

      return wait(ok({
        gameId,
        action,
        status: 'recorded',
        target: targetMap[action]
      }))
    }
  }

  {
    const gameId = mockGameIdFromPath(url, '/collaboration')
    if (method === 'GET' && gameId) {
      return wait(ok(buildMockCollaboration(gameId)))
    }
  }

  {
    const gameId = mockGameIdFromPath(url, '/completion-request')
    if (method === 'POST' && gameId) {
      return requestMockGameCompletion(gameId)
    }
  }

  {
    const gameId = mockGameIdFromPath(url, '/chat-session')
    if (method === 'GET' && gameId) {
      return wait(ok(buildMockChatSession(gameId)))
    }
  }

  {
    const gameId = mockGameIdFromPath(url, '/chat/messages')
    if (method === 'GET' && gameId) {
      return wait(ok(buildMockChatMessages(gameId)))
    }
    if (method === 'POST' && gameId) {
      return sendMockChatMessage(gameId, options.data || {})
    }
  }

  {
    const gameId = mockGameIdFromPath(url, '/service-confirm')
    if (method === 'POST' && gameId) {
      return confirmMockService(gameId, options.data || {})
    }
  }

  {
    const gameId = mockGameIdFromPath(url, '/retrospectives')
    if (method === 'POST' && gameId) {
      return wait(ok(Object.assign({
        id: Date.now(),
        gameId,
        userId: 1,
        createdAt: new Date().toISOString()
      }, options.data || {})))
    }
  }

  {
    const gameId = mockGameIdFromPath(url, '/player-cancel')
    if (method === 'POST' && gameId) {
      return wait(ok({
        gameId,
        status: 'submitted',
        cancelRequest: Object.assign({
          gameId,
          userId: 1
        }, options.data || {})
      }))
    }
  }

  {
    const gameId = mockGameIdFromPath(url, '/expert-cancel')
    if (method === 'POST' && gameId) {
      return wait(ok({
        gameId,
        status: 'submitted',
        cancelRequest: Object.assign({
          gameId,
          userId: 1
        }, options.data || {})
      }))
    }
  }

  {
    const gameId = mockGameIdFromPath(url, '/applications')
    if (method === 'POST' && gameId) {
      return applyGame(gameId, options.data || {})
    }
  }

  if (method === 'POST' && (url === '/api/app/payment/precreate-placeholder' || url === '/api/app/game-payments/wechat')) {
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
  createGamePayment,
  applyGame
}
