const mapApi = require('../api/modules/map')

async function getIndexConfig(params) {
  const result = await mapApi.getIndexConfig(params || {})

  if (result.code !== 0) {
    throw new Error(result.message || '获取地图配置失败')
  }

  return result.data
}

async function reverseGeocode(params) {
  const result = await mapApi.reverseGeocode(params || {})

  if (result.code !== 0) {
    throw new Error(result.message || '获取地点城市信息失败')
  }

  return result.data
}

async function getMyCity(params) {
  const result = await mapApi.getMyCity(params || {})

  if (result.code !== 0) {
    throw new Error(result.message || '获取城市故事失败')
  }

  return result.data
}

async function getPlayPage(pageKey) {
  const result = await mapApi.getPlayPage({ pageKey })

  if (result.code !== 0) {
    throw new Error(result.message || '获取地图玩法失败')
  }

  return result.data
}

async function submitCheckin(data) {
  const result = await mapApi.submitCheckin(data || {})

  if (result.code !== 0) {
    throw new Error(result.message || '提交打卡失败')
  }

  return result.data
}

async function createBlindRoute(data) {
  const result = await mapApi.createBlindRoute(data || {})

  if (result.code !== 0) {
    throw new Error(result.message || '开启路线失败')
  }

  return result.data
}

async function completeBlindRoute(routeId, data) {
  if (!routeId) {
    throw new Error('缺少路线信息')
  }

  const result = await mapApi.completeBlindRoute(routeId, data || {})

  if (result.code !== 0) {
    throw new Error(result.message || '完成路线失败')
  }

  return result.data
}

async function createChallenge(data) {
  const result = await mapApi.createChallenge(data || {})

  if (result.code !== 0) {
    throw new Error(result.message || '发起挑战失败')
  }

  return result.data
}

module.exports = {
  completeBlindRoute,
  createBlindRoute,
  createChallenge,
  getIndexConfig,
  getMyCity,
  getPlayPage,
  reverseGeocode,
  submitCheckin
}
