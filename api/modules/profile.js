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


module.exports = {
  getProfileHome,
  getGrowth,
  getMemberStatus,
  getPointsMall,
  exchangePointsMallGood,
  getPointsOrders,
  getPointsOrderLogistics
}
