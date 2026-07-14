const { ROUTES } = require('../config/routes')
const { clearAuthToken } = require('./auth-session')
const { navigateShellRoute } = require('./shell-nav')

const AUTH_EXPIRED_MESSAGE = '登录已失效'
const AUTH_EXPIRED_CODES = [401, 40101, 40102]

function normalizeCode(value) {
  const code = Number(value)

  return Number.isFinite(code) ? code : null
}

function isAuthExpiredCode(value) {
  const code = normalizeCode(value)

  return code !== null && AUTH_EXPIRED_CODES.includes(code)
}

function isAuthExpiredMessage(message = '') {
  const value = String(message || '').trim().toLowerCase()

  return value === '未登录'
    || value.includes('登录已失效')
    || value.includes('登录已过期')
    || value.includes('请先登录')
    || value.includes('token expired')
    || value.includes('invalid token')
    || value.includes('unauthorized')
}

function isAuthExpiredResult(result = {}) {
  if (!result || typeof result !== 'object') {
    return false
  }

  return result.authExpired === true
    || isAuthExpiredCode(result.code)
    || isAuthExpiredCode(result.status)
    || isAuthExpiredCode(result.statusCode)
    || isAuthExpiredMessage(result.message)
}

function createAuthExpiredError(message = AUTH_EXPIRED_MESSAGE) {
  const error = new Error(message || AUTH_EXPIRED_MESSAGE)

  error.isAuthExpired = true
  error.code = 40102

  return error
}

function isAuthExpiredError(error) {
  if (!error) {
    return false
  }

  return error.isAuthExpired === true
    || isAuthExpiredCode(error.code)
    || isAuthExpiredCode(error.status)
    || isAuthExpiredCode(error.statusCode)
    || isAuthExpiredMessage(error.message)
}

function markAuthExpired() {
  clearAuthToken()

  if (typeof wx === 'undefined' || !wx) {
    return
  }

  ;['enjoy_user', 'enjoy_pre_auth_token'].forEach((key) => {
    if (typeof wx.removeStorageSync === 'function') {
      wx.removeStorageSync(key)
    }
  })
}

function goLogin(currentRoute = '') {
  markAuthExpired()

  navigateShellRoute(`${ROUTES.login}?reason=expired`, {
    currentRoute,
    reuseExisting: false
  })
}

module.exports = {
  AUTH_EXPIRED_MESSAGE,
  createAuthExpiredError,
  goLogin,
  isAuthExpiredError,
  isAuthExpiredResult,
  markAuthExpired
}
