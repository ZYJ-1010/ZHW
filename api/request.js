const env = require('../config/env')
const mockApi = require('./mock')

function request(options) {
  return new Promise((resolve, reject) => {
    const token = wx.getStorageSync('enjoy_token')
    const method = options.method || 'GET'
    const data = options.data || {}
    const query = method === 'GET' && data ? buildQuery(data) : ''

    wx.request({
      url: `${env.baseUrl}${options.url}${query}`,
      method,
      data: method === 'GET' ? {} : data,
      header: Object.assign({
        'X-Client-Type': 'mp-wechat',
        'X-App-Version': '0.1.0'
      }, token ? {
        Authorization: `Bearer ${token}`
      } : {}, options.header || {}),
      success: (res) => resolve(res.data),
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

function get(url, data) {
  if (env.isMock) {
    return mockApi.handleRequest({
      url,
      method: 'GET',
      data
    })
  }

  return request({
    url,
    method: 'GET',
    data
  })
}

function post(url, data) {
  if (env.isMock) {
    return mockApi.handleRequest({
      url,
      method: 'POST',
      data
    })
  }

  return request({
    url,
    method: 'POST',
    data
  })
}

function put(url, data) {
  if (env.isMock) {
    return mockApi.handleRequest({
      url,
      method: 'PUT',
      data
    })
  }

  return request({
    url,
    method: 'PUT',
    data
  })
}

function loginWithWechat(payload) {
  if (env.isMock) {
    return mockApi.loginWithWechat(payload)
  }

  return request({
    url: '/api/app/auth/wechat-login',
    method: 'POST',
    data: payload
  })
}

function verifyInvite(code) {
  if (env.isMock) {
    return mockApi.verifyInvite(code)
  }

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
  verifyInvite
}
