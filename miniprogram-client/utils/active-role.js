const { ROUTES } = require('../config/routes')

const ACTIVE_ROLE_STORAGE_KEY = 'enjoy_active_role'
const ROLE_HOME_ROUTES = {
  player: ROUTES.playerHome || ROUTES.home,
  expert: ROUTES.expertHome || ROUTES.playerHome || ROUTES.home,
  guide: ROUTES.guideHome || ROUTES.playerHome || ROUTES.home
}

function normalizeActiveRole(value) {
  const role = String(value || '').trim().toLowerCase()

  if (role === 'guide' || role === 'leader' || value === '领路人') {
    return 'guide'
  }

  if (role === 'expert' || role === 'master' || value === '行家') {
    return 'expert'
  }

  return 'player'
}

function getActiveRole() {
  if (typeof wx === 'undefined' || typeof wx.getStorageSync !== 'function') {
    return 'player'
  }

  return normalizeActiveRole(wx.getStorageSync(ACTIVE_ROLE_STORAGE_KEY))
}

function setActiveRole(roleType) {
  const role = normalizeActiveRole(roleType)

  if (typeof wx !== 'undefined' && typeof wx.setStorageSync === 'function') {
    wx.setStorageSync(ACTIVE_ROLE_STORAGE_KEY, role)
  }

  return role
}

function activeRoleHomeRoute(roleType = getActiveRole()) {
  return ROLE_HOME_ROUTES[normalizeActiveRole(roleType)] || ROLE_HOME_ROUTES.player
}

function isBackendRoleActive(roleInfo = {}, roleType) {
  const role = normalizeActiveRole(roleType)

  if (role === 'player') {
    return true
  }

  const roles = Array.isArray(roleInfo.roles) ? roleInfo.roles.map(normalizeActiveRole) : []
  const statusMap = roleInfo.roleStatusMap && typeof roleInfo.roleStatusMap === 'object'
    ? roleInfo.roleStatusMap
    : {}
  const status = String(statusMap[role] || '').trim().toLowerCase()

  return roles.includes(role) && (status === 'active' || status === 'approved')
}

function syncCachedUserRoles(roleInfo = {}) {
  if (typeof wx === 'undefined' || typeof wx.getStorageSync !== 'function' || typeof wx.setStorageSync !== 'function') {
    return
  }

  const cachedUser = wx.getStorageSync('enjoy_user')

  if (!cachedUser || typeof cachedUser !== 'object') {
    return
  }

  wx.setStorageSync('enjoy_user', Object.assign({}, cachedUser, {
    roles: Array.isArray(roleInfo.roles) ? roleInfo.roles.slice() : cachedUser.roles,
    roleStatusMap: roleInfo.roleStatusMap && typeof roleInfo.roleStatusMap === 'object'
      ? Object.assign({}, roleInfo.roleStatusMap)
      : cachedUser.roleStatusMap
  }))
}

module.exports = {
  ACTIVE_ROLE_STORAGE_KEY,
  activeRoleHomeRoute,
  getActiveRole,
  isBackendRoleActive,
  normalizeActiveRole,
  setActiveRole,
  syncCachedUserRoles
}
