const STORAGE_PREFIX = 'enjoy_role_apply_result_viewed'

function normalizeRoleType(value) {
  const text = String(value || '').trim()

  if (text === 'guide' || text === 'leader' || text === '领路人') {
    return 'guide'
  }

  if (text === 'expert' || text === 'master' || text === 'pro' || text === '行家') {
    return 'expert'
  }

  return text
}

function normalizeStatus(value) {
  const text = String(value || '').trim()

  if (text === 'approved' || text === 'passed' || text === 'pass' || text === 'active' || text === 'enabled' || text === '已通过') {
    return 'approved'
  }

  if (text === 'rejected' || text === 'reject' || text === 'failed' || text === '未通过' || text === '已驳回') {
    return 'rejected'
  }

  return text
}

function getApplicationIdentity(application = {}) {
  return application.id ||
    application.applicationId ||
    application.application_id ||
    application.applicationNo ||
    application.application_no ||
    application.updatedAt ||
    application.updated_at ||
    application.createdAt ||
    application.created_at ||
    ''
}

function getRoleApplyResultViewKey(options = {}) {
  const roleType = normalizeRoleType(options.roleType || options.roleCode || options.role)
  const status = normalizeStatus(options.status || options.roleStatus)
  const identity = getApplicationIdentity(options.application || options)

  if (!roleType || !status || !identity) {
    return ''
  }

  return `${STORAGE_PREFIX}:${roleType}:${status}:${identity}`
}

function isRoleApplyResultViewed(options = {}) {
  const key = getRoleApplyResultViewKey(options)

  if (!key || typeof wx === 'undefined' || typeof wx.getStorageSync !== 'function') {
    return false
  }

  try {
    return Boolean(wx.getStorageSync(key))
  } catch (error) {
    return false
  }
}

function markRoleApplyResultViewed(options = {}) {
  const key = getRoleApplyResultViewKey(options)

  if (!key || typeof wx === 'undefined' || typeof wx.setStorageSync !== 'function') {
    return false
  }

  try {
    wx.setStorageSync(key, true)
    return true
  } catch (error) {
    return false
  }
}

module.exports = {
  isRoleApplyResultViewed,
  markRoleApplyResultViewed
}
