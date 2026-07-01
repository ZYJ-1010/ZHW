const request = require('../request')

function submitRoleApplication(data) {
  return request.post('/api/app/role-applications', data)
}

function getMyRoleApplications() {
  return request.get('/api/app/role-applications/my')
}

function getMyRoles() {
  return request.get('/api/app/roles/my')
}

function getExpertApplyConfig(params) {
  return request.get('/api/app/role-applications/expert/config', params || {})
}

function getExpertApplyPrecheck(params) {
  return request.get('/api/app/role-applications/expert/precheck', params || {})
}

module.exports = {
  submitRoleApplication,
  getMyRoleApplications,
  getMyRoles,
  getExpertApplyConfig,
  getExpertApplyPrecheck
}
