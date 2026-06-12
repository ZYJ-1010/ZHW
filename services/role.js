const roleApi = require('../api/modules/role')

async function getMyRoles() {
  const result = await roleApi.getMyRoles()

  if (result.code !== 0) {
    throw new Error(result.message || '获取角色信息失败')
  }

  return result.data
}

async function getMyRoleApplications() {
  const result = await roleApi.getMyRoleApplications()

  if (result.code !== 0) {
    throw new Error(result.message || '获取角色申请信息失败')
  }

  return result.data
}

async function submitRoleApplication(roleType) {
  const result = await roleApi.submitRoleApplication({
    roleType
  })

  if (result.code !== 0) {
    throw new Error(result.message || '提交角色申请失败')
  }

  return result.data
}

module.exports = {
  getMyRoles,
  getMyRoleApplications,
  submitRoleApplication
}
