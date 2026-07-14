const request = require('../request')

function getMyCity(params) {
  return request.get('/api/app/map/my-city', params || {})
}

function getIndexConfig(params) {
  return request.get('/api/app/map/index-config', params || {})
}

function getPlayPage(params) {
  return request.get('/api/app/map/play-pages', params || {})
}

function submitCheckin(data) {
  return request.post('/api/app/map/checkins', data || {})
}

function createBlindRoute(data) {
  return request.post('/api/app/map/blind-routes', data || {})
}

function completeBlindRoute(routeId, data) {
  return request.post(`/api/app/map/blind-routes/${encodeURIComponent(routeId)}/complete`, data || {})
}

function createChallenge(data) {
  return request.post('/api/app/map/challenges', data || {})
}

module.exports = {
  completeBlindRoute,
  createBlindRoute,
  createChallenge,
  getIndexConfig,
  getMyCity,
  getPlayPage,
  submitCheckin
}
