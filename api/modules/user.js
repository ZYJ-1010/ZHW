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

function submitRealnameAuth(data) {
  return request.post('/api/app/users/me/realname-auth/submit', data)
}

module.exports = {
  getCurrentUser,
  updateProfile,
  startRealnameAuth,
  submitRealnameAuth
}
