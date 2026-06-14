const logger = require('./utils/logger')

const appLogger = logger.createLogger('app')

App({
  globalData: {
    userInfo: null
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
