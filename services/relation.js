const relationApi = require('../api/modules/relation')

async function getNetworkHome(params) {
  const result = await relationApi.getNetworkHome(params || {})

  if (result.code !== 0) {
    throw new Error(result.message || '获取关系网首页失败')
  }

  return result.data
}

module.exports = {
  getNetworkHome
}
