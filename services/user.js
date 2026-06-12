const userApi = require('../api/modules/user')

async function getCurrentUser() {
  const result = await userApi.getCurrentUser()

  if (result.code !== 0) {
    throw new Error(result.message || '获取用户信息失败')
  }

  return result.data
}

module.exports = {
  getCurrentUser
}
