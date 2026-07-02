const locationApi = require('../api/modules/location')

async function unwrap(result, fallbackMessage) {
  if (result.code !== 0) {
    throw new Error(result.message || fallbackMessage)
  }

  return result.data
}

async function getNearbyGames(params) {
  return unwrap(await locationApi.getNearbyGames(params), '获取附近局失败')
}

async function searchMapPlaces(params) {
  return unwrap(await locationApi.searchMapPlaces(params), '地图地点搜索失败')
}

async function reverseGeocode(params) {
  return unwrap(await locationApi.reverseGeocode(params), '地图地点解析失败')
}

module.exports = {
  getNearbyGames,
  searchMapPlaces,
  reverseGeocode
}
