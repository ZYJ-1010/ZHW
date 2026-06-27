const request = require('../request')

function getProfileHome() {
  return request.get('/api/app/profile/home')
}

function getGrowth() {
  return request.get('/api/app/growth/me')
}

function getMemberStatus() {
  return request.get('/api/app/memberships/me')
}

function replyServiceReview(reviewId, data) {
  return request.post(`/api/app/profile/service-center/reviews/${reviewId}/reply`, data)
}

function getPointsMall() {
  return request.get('/api/app/profile/points/mall')
}

function exchangePointsMallGood(data) {
  return request.post('/api/app/profile/points/mall/exchange', data)
}

function getPointsOrders(data) {
  return request.get('/api/app/profile/points/orders', data)
}

function getPointsOrderLogistics(orderId) {
  return request.get(`/api/app/profile/points/orders/${orderId}/logistics`)
}

function saveSystemProfileInfo(data) {
  return request.put('/api/app/profile/system-management/profile-info', data)
}

function getSystemSkillConfig() {
  return request.get('/api/app/profile/system-management/skill-config')
}

function saveSystemSkillConfig(data) {
  return request.put('/api/app/profile/system-management/skill-config', data)
}

module.exports = {
  getProfileHome,
  getGrowth,
  getMemberStatus,
  replyServiceReview,
  getPointsMall,
  exchangePointsMallGood,
  getPointsOrders,
  getPointsOrderLogistics,
  saveSystemProfileInfo,
  getSystemSkillConfig,
  saveSystemSkillConfig
}
