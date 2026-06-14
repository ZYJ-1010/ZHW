const ENV = {
  MOCK: 'mock',
  PROD: 'prod'
}

const APP_ENV = {
  DEVELOP: 'develop',
  TRIAL: 'trial',
  RELEASE: 'release'
}

const LOG_LEVEL = {
  DEBUG: 'debug',
  INFO: 'info',
  WARN: 'warn',
  ERROR: 'error',
  SILENT: 'silent'
}

const currentEnv = ENV.MOCK

function getMiniProgramEnvVersion() {
  if (typeof wx === 'undefined' || typeof wx.getAccountInfoSync !== 'function') {
    return APP_ENV.DEVELOP
  }

  try {
    const accountInfo = wx.getAccountInfoSync()
    const envVersion = accountInfo && accountInfo.miniProgram && accountInfo.miniProgram.envVersion

    return envVersion || APP_ENV.DEVELOP
  } catch (error) {
    return APP_ENV.DEVELOP
  }
}

const serverMap = {
  [ENV.MOCK]: '',
  [ENV.PROD]: 'https://api.example.com'
}

const currentMiniProgramEnv = getMiniProgramEnvVersion()
const logLevelMap = {
  [APP_ENV.DEVELOP]: LOG_LEVEL.DEBUG,
  [APP_ENV.TRIAL]: LOG_LEVEL.DEBUG,
  [APP_ENV.RELEASE]: LOG_LEVEL.WARN
}
const currentLogLevel = logLevelMap[currentMiniProgramEnv] || LOG_LEVEL.DEBUG

module.exports = {
  ENV,
  APP_ENV,
  LOG_LEVEL,
  currentEnv,
  currentMiniProgramEnv,
  currentLogLevel,
  isMock: currentEnv === ENV.MOCK,
  baseUrl: serverMap[currentEnv]
}
