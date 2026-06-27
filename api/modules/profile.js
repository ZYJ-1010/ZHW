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

module.exports = {
  getProfileHome,
  getGrowth,
  getMemberStatus,
  getPointsMall,
  exchangePointsMallGood
}
