const env = require('../config/env')
const logger = require('../utils/logger')
const { getAuthToken } = require('../utils/auth-session')
const {
  AUTH_EXPIRED_MESSAGE,
  isAuthExpiredResult,
  markAuthExpired
} = require('../utils/auth-error')

const requestLogger = logger.createLogger('request')
const APP_VERSION = '0.1.0'

function request(options) {
  const method = String(options.method || 'GET').toUpperCase()
  const data = options.data || {}
  const url = options.url || ''
  const startedAt = Date.now()
  const token = getAuthToken()
  const header = Object.assign({
    'X-Client-Type': 'mp-wechat',
    'X-App-Version': APP_VERSION
  }, token ? {
    Authorization: `Bearer ${token}`
  } : {}, options.header || {})

  requestLogger.debug('request start', {
    method,
    url,
    data
  })

  return executeRequest({
    method,
    url,
    data,
    header
  }).then((response) => {
    const body = response ? response.data : undefined
    const statusCode = response && typeof response.statusCode === 'number' ? response.statusCode : 0
    const requestId = extractRequestId(body, response)
    const code = getBusinessCode(body)
    const message = getBusinessMessage(body)
    const logPayload = {
      method,
      url,
      duration: Date.now() - startedAt,
      statusCode,
      requestId,
      code,
      message
    }

    if (statusCode >= 500) {
      requestLogger.error('request http error', logPayload)
    } else if (statusCode >= 400) {
      requestLogger.warn('request http error', logPayload)
    } else if (typeof code === 'number' && code !== 0) {
      requestLogger.warn('request business error', logPayload)
    } else {
      requestLogger.info('request success', logPayload)
    }

    if (isAuthExpiredResult(Object.assign({}, body && typeof body === 'object' ? body : {}, { statusCode }))) {
      markAuthExpired()

      if (body && typeof body === 'object') {
        return Object.assign({}, body, {
          authExpired: true,
          code: typeof code === 'number' ? code : 40102,
          message: message || AUTH_EXPIRED_MESSAGE
        })
      }

      return {
        code: 40102,
        message: AUTH_EXPIRED_MESSAGE,
        data: null,
        authExpired: true
      }
    }

    return body
  }).catch((error) => {
    requestLogger.error('request network error', {
      method,
      url,
      duration: Date.now() - startedAt,
      error
    })

    throw error
  })
}

function executeRequest(options) {
  const method = options.method
  const url = options.url
  const data = options.data || {}
  const header = options.header || {}
  const query = method === 'GET' ? buildQuery(data) : ''
  const baseUrlError = getBaseUrlError()

  if (env.isMock) {
    return Promise.reject(new Error('当前包已关闭前端 mock，请启动本机后端服务'))
  }

  if (baseUrlError) {
    return Promise.reject(new Error(baseUrlError))
  }

  return new Promise((resolve, reject) => {
    wx.request({
      url: `${env.baseUrl}${url}${query}`,
      method,
      data: method === 'GET' ? {} : data,
      header,
      success: resolve,
      fail: reject
    })
  })
}

function getBaseUrlError() {
  if (!env.baseUrl) {
    return '接口地址未配置'
  }
  if (env.currentEnv === env.ENV.PROD && !env.isProdBaseUrlReady) {
    return '生产接口地址未配置，请在 miniprogram-client/config/env.js 配置 HTTPS API 域名'
  }
  return ''
}

function buildQuery(data) {
  const pairs = Object.keys(data)
    .filter((key) => data[key] !== undefined && data[key] !== null && data[key] !== '')
    .map((key) => `${encodeURIComponent(key)}=${encodeURIComponent(data[key])}`)

  return pairs.length ? `?${pairs.join('&')}` : ''
}

function extractRequestId(body, response) {
  if (body && typeof body === 'object' && body.requestId) {
    return body.requestId
  }

  const header = response && (response.header || response.headers) ? (response.header || response.headers) : {}

  return header['x-request-id'] || header['X-Request-Id'] || header.requestId || ''
}

function getBusinessCode(body) {
  return body && typeof body.code === 'number' ? body.code : null
}

function getBusinessMessage(body) {
  return body && typeof body.message === 'string' ? body.message : ''
}

function get(url, data) {
  return request({
    url,
    method: 'GET',
    data
  })
}

function post(url, data) {
  return request({
    url,
    method: 'POST',
    data
  })
}

function put(url, data) {
  return request({
    url,
    method: 'PUT',
    data
  })
}

function loginWithWechat(payload) {
  return request({
    url: '/api/app/auth/wechat-login',
    method: 'POST',
    data: payload
  })
}

function sendPhoneCode(payload) {
  return request({
    url: '/api/app/sms/send-code',
    method: 'POST',
    data: payload
  })
}

function verifyPhoneCode(payload) {
  return request({
    url: '/api/app/sms/verify-code',
    method: 'POST',
    data: payload
  })
}

function loginWithPhone(payload) {
  return request({
    url: '/api/app/auth/phone-login',
    method: 'POST',
    data: payload
  })
}

function loginWithPassword(payload) {
  return Promise.resolve({
    code: 410,
    message: '\u5f53\u524d\u5c0f\u7a0b\u5e8f\u4ec5\u652f\u6301\u901a\u8fc7\u9080\u8bf7\u5165\u53e3\u5fae\u4fe1\u767b\u5f55',
    data: null
  })
}

function resetPassword(payload) {
  return Promise.resolve({
    code: 410,
    message: '\u5f53\u524d\u5c0f\u7a0b\u5e8f\u4ec5\u652f\u6301\u901a\u8fc7\u9080\u8bf7\u5165\u53e3\u5fae\u4fe1\u767b\u5f55\uff0c\u65e0\u9700\u627e\u56de\u5bc6\u7801',
    data: null
  })
}

function issueTokenAfterIdentity(payload) {
  return request({
    url: '/api/app/auth/issue-token-after-identity',
    method: 'POST',
    data: payload || {}
  })
}

function verifyInvite(code, entryType) {
  const data = { inviteCode: code }
  if (entryType) {
    data.entryType = entryType
  }
  return request({
    url: '/api/app/invites/precheck',
    method: 'POST',
    data
  })
}

module.exports = {
  get,
  post,
  put,
  loginWithWechat,
  sendPhoneCode,
  verifyPhoneCode,
  loginWithPhone,
  loginWithPassword,
  resetPassword,
  issueTokenAfterIdentity,
  verifyInvite
}
