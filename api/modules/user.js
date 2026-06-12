const request = require('../request')

function getCurrentUser() {
  return request.get('/api/app/users/me')
}

function updateProfile(data) {
  return request.put('/api/app/users/me/profile', data)
}

module.exports = {
  getCurrentUser,
  updateProfile
}
