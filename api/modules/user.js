const request = require('../request')

function getCurrentUser() {
  return request.get('/api/app/users/me')
}

function updateProfile(data) {
  return request.put('/api/app/users/me/profile', data)
}

function startRealnameAuth() {
  return request.post('/api/app/users/me/realname-auth', {})
}

module.exports = {
  getCurrentUser,
  updateProfile,
  startRealnameAuth
}
