const profileApi = require('../api/modules/profile')

async function getProfileHome() {
  const result = await profileApi.getProfileHome()

  if (result.code !== 0) {
    throw new Error(result.message || '获取个人中心失败')
  }

  return result.data
}

module.exports = {
  getProfileHome
}
