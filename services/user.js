const userApi = require('../api/modules/user')

async function getCurrentUser() {
  const result = await userApi.getCurrentUser()

  if (result.code !== 0) {
    throw new Error(result.message || '获取用户信息失败')
  }

  return result.data
}

async function startRealnameAuth() {
  const result = await userApi.startRealnameAuth()

  if (result.code !== 0) {
    throw new Error(result.message || '实名认证页面打开失败')
  }

  return result.data
}

async function submitRealnameAuth(data) {
  const result = await userApi.submitRealnameAuth(data)

  if (result.code !== 0) {
    throw new Error(result.message || '实名认证失败')
  }

  return result.data
}

module.exports = {
  getCurrentUser,
  startRealnameAuth,
  submitRealnameAuth
}
