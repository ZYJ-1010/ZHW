const ENV = {
  MOCK: 'mock',
  LOCAL: 'local',
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

// 本地调试需要强制切换时填 ENV.MOCK / ENV.LOCAL / ENV.PROD；上线前保持空字符串。
const ENV_OVERRIDE = ''
// 当前测试阶段开发版仍默认直连线上 API；真正本地联调时改为 ENV.LOCAL。
const DEVELOP_DEFAULT_ENV = ENV.PROD

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

function isValidEnv(value) {
  return value === ENV.MOCK || value === ENV.LOCAL || value === ENV.PROD
}

function resolveCurrentEnv(envVersion) {
  if (isValidEnv(ENV_OVERRIDE)) {
    return ENV_OVERRIDE
  }
  if (envVersion === APP_ENV.TRIAL || envVersion === APP_ENV.RELEASE) {
    return ENV.PROD
  }
  return DEVELOP_DEFAULT_ENV
}

const serverMap = {
  [ENV.MOCK]: '',
  [ENV.LOCAL]: 'http://127.0.0.1:8080',
  [ENV.PROD]: 'https://api.haowan.net.cn'
}

const socketServerMap = {
  [ENV.MOCK]: '',
  [ENV.LOCAL]: 'ws://127.0.0.1:8080',
  [ENV.PROD]: 'wss://im.haowan.net.cn'
}

const currentMiniProgramEnv = getMiniProgramEnvVersion()
const currentEnv = resolveCurrentEnv(currentMiniProgramEnv)
const logLevelMap = {
  [APP_ENV.DEVELOP]: LOG_LEVEL.DEBUG,
  [APP_ENV.TRIAL]: LOG_LEVEL.DEBUG,
  [APP_ENV.RELEASE]: LOG_LEVEL.WARN
}
const currentLogLevel = logLevelMap[currentMiniProgramEnv] || LOG_LEVEL.DEBUG
const baseUrl = serverMap[currentEnv]
const socketBaseUrl = socketServerMap[currentEnv]

function isPlaceholderBaseUrl(value) {
  return !value || value.indexOf('api.example.com') >= 0
}

const isProdBaseUrlReady = currentEnv !== ENV.PROD || (
  /^https:\/\//.test(baseUrl || '') && !isPlaceholderBaseUrl(baseUrl)
)

module.exports = {
  ENV,
  APP_ENV,
  LOG_LEVEL,
  currentEnv,
  currentMiniProgramEnv,
  currentLogLevel,
  isMock: currentEnv === ENV.MOCK,
  baseUrl,
  socketBaseUrl,
  isProdBaseUrlReady
}
