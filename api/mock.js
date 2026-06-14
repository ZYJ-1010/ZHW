const {
  validInvites,
  mockUser,
  mockCurrentUser,
  mockHome,
  mockProfileHome,
  mockNewbieTasks,
  mockRoleApplications
} = require('./mock-data')

let mockCurrentRealnameStatus = 'pending'

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
    needRealname: !isRealnameVerified
  })
}

function buildCurrentUser() {
  const isRealnameVerified = mockCurrentRealnameStatus === 'verified' || getMockRealnameStorageVerified()

  return Object.assign({}, mockCurrentUser, {
    authStatus: isRealnameVerified ? 'verified' : 'pending',
    realnameStatus: isRealnameVerified ? 'verified' : 'pending',
    needRealname: !isRealnameVerified,
    todoCounts: Object.assign({}, mockCurrentUser.todoCounts, {
      realname: isRealnameVerified ? 0 : 1
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

  if (method === 'GET' && url === '/api/app/users/me') {
    return wait(ok(buildCurrentUser()))
  }

  if (method === 'POST' && url === '/api/app/users/me/realname-auth') {
    return wait(ok({
      url: '/pages/login/realname/index',
      provider: 'mock'
    }))
  }

  if (method === 'GET' && url === '/api/app/newbie-tasks') {
    return wait(ok(buildNewbieTaskSummary()))
  }

  if (method === 'GET' && url === '/api/app/home') {
    return wait(ok(mockHome))
  }

  if (method === 'GET' && url === '/api/app/profile/home') {
    return wait(ok(mockProfileHome))
  }

  if (method === 'GET' && url === '/api/app/role-applications/my') {
    return wait(ok(mockRoleApplications))
  }

  if (method === 'POST' && url === '/api/app/role-applications') {
    return wait(ok({
      applicationId: `mock-role-${Date.now()}`,
      roleType: options.data.roleType,
      status: 'pending_audit',
      statusText: '已进入审核'
    }))
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
  verifyInvite
}
