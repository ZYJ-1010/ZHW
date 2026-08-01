const api = require('../api/request')
const logger = require('../utils/logger')
const { setAuthToken, clearAuthToken } = require('../utils/auth-session')

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
  setAuthToken(token || preAuthToken)
  if (token) {
    wx.removeStorageSync('enjoy_pre_auth_token')
  } else if (preAuthToken) {
    wx.setStorageSync('enjoy_pre_auth_token', preAuthToken)
  }
  wx.setStorageSync('enjoy_user', result.data.user)

  return result.data
}

async function precheckWechatEntry() {
  const code = await wxLogin()
  const result = await api.precheckWechatEntry({ code })

  if (result.code !== 0) {
    throw new Error(result.message || '微信账号检测失败')
  }

  return result.data || {}
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
  const result = await api.loginWithPhone({
    phone: options.phone || '',
    code: options.code || '',
    inviteCode: options.inviteCode || '',
    entryType: options.entryType || ''
  })

  if (result.code !== 0) {
    throw new Error(result.message || '手机号登录/注册失败')
  }

  const token = result.data.token || ''
  if (!token) {
    throw new Error('手机号登录/注册未返回登录凭证')
  }
  setAuthToken(token)
  wx.removeStorageSync('enjoy_pre_auth_token')
  wx.setStorageSync('enjoy_user', result.data.user)

  return result.data
}

async function loginByPassword(options = {}) {
  const result = await api.loginWithPassword({ phone: options.phone || '', password: options.password || '' })
  if (result.code !== 0) throw new Error(result.message || '手机号或密码错误')
  const token = result.data && result.data.token || ''
  if (!token) {
    throw new Error('账号密码登录未返回登录凭证')
  }
  setAuthToken(token)
  wx.removeStorageSync('enjoy_pre_auth_token')
  wx.setStorageSync('enjoy_user', result.data.user)
  return result.data
}

async function bindWechatAccount() {
  const code = await wxLogin()
  const result = await api.bindWechatAccount({ code })
  if (result.code !== 0) throw new Error(result.message || '微信绑定失败')
  wx.setStorageSync('enjoy_user', result.data)
  return result.data
}

async function setLoginPassword(password) {
  const result = await api.setLoginPassword({ password: password || '' })
  if (result.code !== 0) throw new Error(result.message || '设置密码失败')
  wx.setStorageSync('enjoy_user', result.data)
  return result.data
}

async function resetPassword(options = {}) {
  const result = await api.resetPassword({
    phone: options.phone || '',
    code: options.code || '',
    password: options.password || ''
  })
  if (result.code !== 0) {
    throw new Error(result.message || '密码重置失败')
  }
  return result.data || {}
}

async function issueTokenAfterIdentity() {
  const result = await api.issueTokenAfterIdentity({})

  if (result.code !== 0) {
    throw new Error(result.message || '身份认证后登录态换发失败')
  }

  setAuthToken(result.data.token || '')
  wx.removeStorageSync('enjoy_pre_auth_token')
  if (result.data.user) {
    wx.setStorageSync('enjoy_user', result.data.user)
  }

  return result.data
}

async function deleteAccount() {
  const result = await api.deleteAccount({ confirm: true })

  if (result.code !== 0) {
    throw new Error(result.message || '账号注销失败')
  }

  logout()
  return result.data || {}
}

function logout() {
  clearAuthToken()
  wx.removeStorageSync('enjoy_pre_auth_token')
  wx.removeStorageSync('enjoy_user')
  wx.removeStorageSync('enjoy_invite_context')
  wx.removeStorageSync('enjoy_active_role')
}

module.exports = {
  precheckWechatEntry,
  loginByWechat,
  sendPhoneCode,
  verifyPhoneCode,
  loginByPhone,
  loginByPassword,
  bindWechatAccount,
  setLoginPassword,
  resetPassword,
  issueTokenAfterIdentity,
  deleteAccount,
  logout
}
