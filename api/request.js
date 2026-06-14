const env = require('../config/env')
const logger = require('../utils/logger')
const mockApi = require('./mock')

const requestLogger = logger.createLogger('request')
const APP_VERSION = '0.1.0'

function request(options) {
  const method = String(options.method || 'GET').toUpperCase()
  const data = options.data || {}
  const url = options.url || ''
  const startedAt = Date.now()
  const token = typeof wx !== 'undefined' && typeof wx.getStorageSync === 'function'
    ? wx.getStorageSync('enjoy_token')
    : ''
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

  if (env.isMock) {
    return mockApi.handleRequest({
      url,
      method,
      data
    }).then((result) => ({
      statusCode: 200,
      header: {},
      data: result
    }))
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
    url: '/api/app/auth/phone-code',
    method: 'POST',
    data: payload
  })
}

function verifyPhoneCode(payload) {
  return request({
    url: '/api/app/auth/phone-code/verify',
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
  return request({
    url: '/api/app/auth/password-login',
    method: 'POST',
    data: payload
  })
}

function resetPassword(payload) {
  return request({
    url: '/api/app/auth/password/reset',
    method: 'POST',
    data: payload
  })
}

function verifyInvite(code) {
  return request({
    url: '/api/app/invites/verify',
    method: 'POST',
    data: { code }
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
  verifyInvite
}
