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

module.exports = {
  getProfileHome,
  getGrowth,
  getMemberStatus
}
