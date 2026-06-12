const {
  validInvites,
  mockUser,
  mockCurrentUser,
  mockHome,
  mockProfileHome,
  mockRoleApplications
} = require('./mock-data')

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

  if (inviteCode && !validInvites[inviteCode]) {
    return wait(fail(40001, '邀请码无效'))
  }

  return wait(ok({
    token: 'mock-token-enjoy-login',
    user: mockUser,
    isNewUser: Boolean(inviteCode),
    code: payload.code || 'mock-wx-login-code'
  }))
}

function verifyInvite(code) {
  const invite = validInvites[code]

  if (!invite) {
    return wait(fail(40001, '邀请码无效'))
  }

  return wait(ok(invite))
}

function handleRequest(options) {
  const method = options.method || 'GET'
  const url = options.url

  if (method === 'GET' && url === '/api/app/users/me') {
    return wait(ok(mockCurrentUser))
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
  verifyInvite
}
