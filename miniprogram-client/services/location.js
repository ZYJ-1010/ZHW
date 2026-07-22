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

async function saveCurrentLocation(payload) {
  return unwrap(await locationApi.saveCurrentLocation(payload), '保存当前位置失败')
}

async function saveManualLocation(payload) {
  return unwrap(await locationApi.saveManualLocation(payload), '保存手动位置失败')
}

async function getLocationFallback() {
  return unwrap(await locationApi.getLocationFallback(), '获取同城推荐失败')
}

async function searchMapPlaces(params) {
  return unwrap(await locationApi.searchMapPlaces(params), '地图地点搜索失败')
}

async function reverseGeocode(params) {
  return unwrap(await locationApi.reverseGeocode(params), '地图地点解析失败')
}

module.exports = {
  getNearbyGames,
  saveCurrentLocation,
  saveManualLocation,
  getLocationFallback,
  searchMapPlaces,
  reverseGeocode
}
