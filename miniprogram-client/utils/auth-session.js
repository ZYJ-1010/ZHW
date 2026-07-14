const TOKEN_STORAGE_KEY = 'enjoy_token'

let runtimeToken = ''
let initialized = false

function normalizeToken(value) {
  return String(value || '').replace(/^Bearer\s+/i, '').trim()
}

function canUseStorage() {
  return typeof wx !== 'undefined'
    && typeof wx.getStorageSync === 'function'
    && typeof wx.setStorageSync === 'function'
}

function initializeAuthToken(preferredToken) {
  const preferred = normalizeToken(preferredToken)

  if (preferred) {
    runtimeToken = preferred
    initialized = true
    if (canUseStorage()) {
      wx.setStorageSync(TOKEN_STORAGE_KEY, preferred)
    }
    return runtimeToken
  }

  if (!initialized) {
    runtimeToken = canUseStorage()
      ? normalizeToken(wx.getStorageSync(TOKEN_STORAGE_KEY))
      : ''
    initialized = true
  }

  return runtimeToken
}

function getAuthToken() {
  return initialized ? runtimeToken : initializeAuthToken()
}

function setAuthToken(token) {
  runtimeToken = normalizeToken(token)
  initialized = true

  if (canUseStorage()) {
    if (runtimeToken) {
      wx.setStorageSync(TOKEN_STORAGE_KEY, runtimeToken)
    } else if (typeof wx.removeStorageSync === 'function') {
      wx.removeStorageSync(TOKEN_STORAGE_KEY)
    }
  }

  return runtimeToken
}

function clearAuthToken() {
  return setAuthToken('')
}

module.exports = {
  initializeAuthToken,
  getAuthToken,
  setAuthToken,
  clearAuthToken
}
