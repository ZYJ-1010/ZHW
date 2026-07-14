const homeApi = require('../api/modules/home')
const { createAuthExpiredError, isAuthExpiredResult } = require('../utils/auth-error')

async function getHome(params) {
  const result = await homeApi.getHome(params || {})

  if (isAuthExpiredResult(result)) {
    throw createAuthExpiredError(result.message)
  }

  if (result.code !== 0) {
    throw new Error(result.message || '获取首页信息失败')
  }

  return result.data
}

module.exports = {
  getHome
}
