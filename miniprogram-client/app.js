const logger = require('./utils/logger')
const env = require('./config/env')
const { setActiveRole } = require('./utils/active-role')
const { initializeAuthToken, getAuthToken, setAuthToken } = require('./utils/auth-session')
const sharedModules = [
  require('./config/page-map'),
  require('./services/chat-media'),
  require('./services/game'),
  require('./services/location'),
  require('./services/map'),
  require('./services/message'),
  require('./services/relation'),
  require('./services/review')
]

const appLogger = logger.createLogger('app')

function formatGlobalError(error) {
  if (error == null) {
    return {
      message: ''
    }
  }

  if (typeof error === 'string') {
    return {
      message: error
    }
  }

  return {
    name: String(error.name || 'Error'),
    message: String(error.message || error.errMsg || error),
    stack: String(error.stack || ''),
    errMsg: error.errMsg ? String(error.errMsg) : ''
  }
}

function normalizeDevToken(value) {
  const token = String(value || '').trim()

  if (!token || token === 'HOME_TEST_TOKEN') {
    return ''
  }

  try {
    return decodeURIComponent(token).replace(/^Bearer\s+/i, '').trim()
  } catch (error) {
    return token.replace(/^Bearer\s+/i, '').trim()
  }
}

function isDevToolsRuntime() {
  try {
    if (typeof wx !== 'undefined' && typeof wx.getDeviceInfo === 'function') {
      return wx.getDeviceInfo().platform === 'devtools'
    }

    if (typeof wx !== 'undefined' && typeof wx.getSystemInfoSync === 'function') {
      return wx.getSystemInfoSync().platform === 'devtools'
    }
  } catch (error) {
    return false
  }

  return false
}

function canApplyDevToken() {
  return isDevToolsRuntime() || env.currentMiniProgramEnv !== env.APP_ENV.RELEASE
}

App({
  globalData: {
    userInfo: null,
    sharedModules
  },

  onLaunch(options = {}) {
    this.applyDevToken(options)
  },

  onShow(options = {}) {
    this.applyDevToken(options)
  },

  applyDevToken(options = {}) {
    const query = options.query || {}
    const devToken = normalizeDevToken(query.devToken)

    if (devToken && canApplyDevToken()) {
      const currentToken = getAuthToken()
      const roleType = String(query.roleType || '').trim()

      if (currentToken && currentToken !== devToken) {
        wx.removeStorageSync('enjoy_user')
      }

      setAuthToken(devToken)
      setActiveRole(roleType || (currentToken !== devToken ? 'player' : wx.getStorageSync('enjoy_active_role')))
      return
    }

    initializeAuthToken()
  },

  onError(error) {
    appLogger.error('app error', formatGlobalError(error))
  },

  onUnhandledRejection(event) {
    const reason = event && (event.reason || event.errMsg || event.message || event)

    appLogger.error('unhandled rejection', formatGlobalError(reason))
  }
})
