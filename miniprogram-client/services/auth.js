const api = require('../api/request')
const logger = require('../utils/logger')

const authLogger = logger.createLogger('auth')

function wxLogin() {
  return new Promise((resolve, reject) => {
    wx.login({
      success: (res) => {
        const code = res.code || ''

        if (!code) {
          authLogger.error('wx.login missing code', {
            result: res
          })
          reject(new Error('wx.login returned empty code'))
          return
        }

        resolve(code)
      },
      fail: (error) => {
        authLogger.error('wx.login failed', {
          error
        })
        reject(error)
      }
    })
  })
}

async function loginByWechat(options = {}) {
  const code = await wxLogin()
  const result = await api.loginWithWechat({
    code,
    inviteCode: options.inviteCode || '',
    entryType: options.entryType || '',
    encryptedData: options.encryptedData || '',
    iv: options.iv || ''
  })

  if (result.code !== 0) {
    throw new Error(result.message || '授权失败')
  }

  const token = result.data.token || ''
  const preAuthToken = result.data.preAuthToken || ''
  wx.setStorageSync('enjoy_token', token || preAuthToken)
  if (token) {
    wx.removeStorageSync('enjoy_pre_auth_token')
  } else if (preAuthToken) {
    wx.setStorageSync('enjoy_pre_auth_token', preAuthToken)
  }
  wx.setStorageSync('enjoy_user', result.data.user)

  return result.data
}

async function sendPhoneCode(options) {
  const payload = options && typeof options === 'object' ? options : {
    phone: options
  }
  const result = await api.sendPhoneCode({
    phone: payload.phone || '',
    scene: payload.scene || 'login'
  })

  if (result.code !== 0) {
    throw new Error(result.message || '验证码发送失败')
  }

  return result.data
}

async function verifyPhoneCode(options) {
  const result = await api.verifyPhoneCode({
    phone: options.phone || '',
    code: options.code || ''
  })

  if (result.code !== 0) {
    throw new Error(result.message || '手机号或验证码错误')
  }

  return result.data
}

async function loginByPhone(options = {}) {
  if (options.inviteCode) {
    return loginByWechat({
      inviteCode: options.inviteCode,
      entryType: options.entryType || ''
    })
  }

  throw new Error('\u5c0f\u7a0b\u5e8f\u9700\u8981\u9080\u8bf7\u624d\u53ef\u4ee5\u8fdb\u5165')
}

async function loginByPassword(options = {}) {
  if (options.inviteCode) {
    return loginByWechat({
      inviteCode: options.inviteCode,
      entryType: options.entryType || ''
    })
  }

  throw new Error('\u5c0f\u7a0b\u5e8f\u9700\u8981\u9080\u8bf7\u624d\u53ef\u4ee5\u8fdb\u5165')
}

async function resetPassword() {
  throw new Error('\u5f53\u524d\u5c0f\u7a0b\u5e8f\u4ec5\u652f\u6301\u901a\u8fc7\u9080\u8bf7\u5165\u53e3\u5fae\u4fe1\u767b\u5f55\uff0c\u65e0\u9700\u627e\u56de\u5bc6\u7801')
}

async function issueTokenAfterIdentity() {
  const result = await api.issueTokenAfterIdentity({})

  if (result.code !== 0) {
    throw new Error(result.message || '身份认证后登录态换发失败')
  }

  wx.setStorageSync('enjoy_token', result.data.token || '')
  wx.removeStorageSync('enjoy_pre_auth_token')
  if (result.data.user) {
    wx.setStorageSync('enjoy_user', result.data.user)
  }

  return result.data
}

function logout() {
  wx.removeStorageSync('enjoy_token')
  wx.removeStorageSync('enjoy_pre_auth_token')
  wx.removeStorageSync('enjoy_user')
  wx.removeStorageSync('enjoy_invite_context')
}

module.exports = {
  loginByWechat,
  sendPhoneCode,
  verifyPhoneCode,
  loginByPhone,
  loginByPassword,
  resetPassword,
  issueTokenAfterIdentity,
  logout
}
