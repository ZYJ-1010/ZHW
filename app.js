const logger = require('./utils/logger')
const env = require('./config/env')

const appLogger = logger.createLogger('app')

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

App({
  globalData: {
    userInfo: null
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

    if (devToken && env.currentMiniProgramEnv !== env.APP_ENV.RELEASE) {
      wx.setStorageSync('enjoy_token', devToken)
    }
  },

  onError(error) {
    appLogger.error('app error', {
      error
    })
  },

  onUnhandledRejection(event) {
    const reason = event && (event.reason || event.errMsg || event.message || event)

    appLogger.error('unhandled rejection', {
      reason
    })
  }
})
