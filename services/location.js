const locationApi = require('../api/modules/location')

async function getNearbyGames(params) {
  const result = await locationApi.getNearbyGames(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取附近局失败')
  }

  return result.data
}

module.exports = {
  getNearbyGames
}
