const STORAGE_PREFIX = 'enjoy_role_apply_draft'
const USER_STORAGE_KEY = 'enjoy_user'

function normalizeRoleType(roleType) {
  const text = String(roleType || '').trim()

  if (text === 'guide' || text === 'leader' || text === '领路人') {
    return 'guide'
  }

  if (text === 'expert' || text === 'master' || text === '行家') {
    return 'expert'
  }

  return text || 'unknown'
}

function cloneData(value) {
  try {
    return JSON.parse(JSON.stringify(value || {}))
  } catch (error) {
    return {}
  }
}

function getCurrentUserKey() {
  try {
    if (typeof wx === 'undefined' || typeof wx.getStorageSync !== 'function') {
      return 'anonymous'
    }

    const user = wx.getStorageSync(USER_STORAGE_KEY) || {}
    const userId = user.userId || user.user_id || user.id || user.openId || user.openid || user.phone || ''

    return userId ? String(userId) : 'anonymous'
  } catch (error) {
    return 'anonymous'
  }
}

function getDraftStorageKey(roleType) {
  return `${STORAGE_PREFIX}_${getCurrentUserKey()}_${normalizeRoleType(roleType)}`
}

function readRoleApplyDraft(roleType) {
  try {
    if (typeof wx === 'undefined' || typeof wx.getStorageSync !== 'function') {
      return null
    }

    const draft = wx.getStorageSync(getDraftStorageKey(roleType))

    return draft && typeof draft === 'object' ? draft : null
  } catch (error) {
    return null
  }
}

function saveRoleApplyDraft(roleType, draft) {
  if (typeof wx === 'undefined' || typeof wx.setStorageSync !== 'function') {
    return null
  }

  const payload = Object.assign({}, cloneData(draft), {
    roleType: normalizeRoleType(roleType),
    updatedAt: new Date().toISOString()
  })

  try {
    wx.setStorageSync(getDraftStorageKey(roleType), payload)
  } catch (error) {
    return null
  }

  return payload
}

function clearRoleApplyDraft(roleType) {
  try {
    if (typeof wx !== 'undefined' && typeof wx.removeStorageSync === 'function') {
      wx.removeStorageSync(getDraftStorageKey(roleType))
    }
  } catch (error) {
    // Ignore local cache cleanup failures; submitting the application has already succeeded.
  }
}

module.exports = {
  readRoleApplyDraft,
  saveRoleApplyDraft,
  clearRoleApplyDraft
}
