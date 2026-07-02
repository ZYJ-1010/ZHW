const request = require('../request')

function getCurrentUser() {
  return request.get('/api/app/users/me')
}

function updateProfile(data) {
  return request.put('/api/app/users/me/profile', data)
}

function startRealnameAuth() {
  return request.post('/api/app/identity/faceid/detect-auth', {})
}

function submitRealnameAuth(data) {
  return request.post('/api/app/identity/phone/verify', data)
}

function bindPhone(data) {
  return request.post('/api/app/identity/phone/bind', data)
}

function getIdentityStatus() {
  return request.get('/api/app/identity/status')
}

module.exports = {
  getCurrentUser,
  updateProfile,
  startRealnameAuth,
  submitRealnameAuth,
  bindPhone,
  getIdentityStatus
}
