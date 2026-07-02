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

function getExpertApplyConfig() {
  return request.get('/api/app/role-applications/expert/config')
}

function getGuideApplyConfig() {
  return request.get('/api/app/role-applications/guide/config')
}

function getRoleStatusPageConfig() {
  return request.get('/api/app/role-applications/status-config')
}

function getRoleBenefitConfig() {
  return request.get('/api/app/role-applications/benefit-config')
}

module.exports = {
  submitRoleApplication,
  getMyRoleApplications,
  getMyRoles,
  getExpertApplyConfig,
  getGuideApplyConfig,
  getRoleStatusPageConfig,
  getRoleBenefitConfig
}
