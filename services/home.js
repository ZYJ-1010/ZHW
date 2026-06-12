const homeApi = require('../api/modules/home')

async function getHome() {
  const result = await homeApi.getHome()

  if (result.code !== 0) {
    throw new Error(result.message || '获取首页信息失败')
  }

  return result.data
}

module.exports = {
  getHome
}
