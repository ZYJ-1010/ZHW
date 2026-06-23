const {
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
  mockGuideProgress,
  mockGuideCancelDetail,
  mockGameManage,
  mockPlayerGameManage
} = require('./mock-data')

let mockCurrentRealnameStatus = 'pending'
let mockRoleStatusMap = Object.assign({}, mockCurrentUser.roleStatusMap)
let mockSubmittedRoleApplications = []
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

  if (method === 'GET' && url === '/api/app/profile/home') {
    return wait(ok(mockProfileHome))
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
