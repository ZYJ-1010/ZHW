const request = require('../request')

// 地图打卡、盲盒路线和挑战属于二期能力；一期不应向不存在的后端路由发起请求。
// 保留方法名是为了兼容已有页面，调用时返回明确的禁用错误，由页面统一展示空态/提示。
const PHASE_TWO_DISABLED_MESSAGE = '地图玩法将在后续版本开放'

function phaseTwoDisabled() {
  const error = new Error(PHASE_TWO_DISABLED_MESSAGE)
  error.code = 'FEATURE_DISABLED'
  return Promise.reject(error)
}

function getMyCity(params) {
  return phaseTwoDisabled()
}

function getIndexConfig(params) {
  return request.get('/api/app/map/index-config', params || {})
}

function reverseGeocode(params) {
  return request.get('/api/app/map/reverse-geocode', params || {})
}

function getPlayPage(params) {
  return phaseTwoDisabled()
}

function submitCheckin(data) {
  return phaseTwoDisabled()
}

function createBlindRoute(data) {
  return phaseTwoDisabled()
}

function completeBlindRoute(routeId, data) {
  return phaseTwoDisabled()
}

function createChallenge(data) {
  return phaseTwoDisabled()
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
