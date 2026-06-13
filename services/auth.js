const api = require('../api/request')

function wxLogin() {
  return new Promise((resolve, reject) => {
    wx.login({
      success: (res) => resolve(res.code || ''),
      fail: reject
    })
  })
}

async function loginByWechat(options = {}) {
  const code = await wxLogin()
  const result = await api.loginWithWechat({
    code,
    inviteCode: options.inviteCode || '',
    encryptedData: options.encryptedData || '',
    iv: options.iv || ''
  })

  if (result.code !== 0) {
    throw new Error(result.message || '授权失败')
  }

  wx.setStorageSync('enjoy_token', result.data.token)
  wx.setStorageSync('enjoy_user', result.data.user)

  return result.data
}

async function sendPhoneCode(phone) {
  const result = await api.sendPhoneCode({
    phone
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

async function loginByPhone(options) {
  const result = await api.loginWithPhone({
    phone: options.phone || '',
    code: options.code || '',
    inviteCode: options.inviteCode || ''
  })

  if (result.code !== 0) {
    throw new Error(result.message || '手机号登录失败')
  }

  wx.setStorageSync('enjoy_token', result.data.token)
  wx.setStorageSync('enjoy_user', result.data.user)

  return result.data
}

async function loginByPassword(options) {
  const result = await api.loginWithPassword({
    phone: options.phone || '',
    password: options.password || ''
  })

  if (result.code !== 0) {
    throw new Error(result.message || '密码登录失败')
  }

  wx.setStorageSync('enjoy_token', result.data.token)
  wx.setStorageSync('enjoy_user', result.data.user)

  return result.data
}

async function resetPassword(options) {
  const result = await api.resetPassword({
    phone: options.phone || '',
    code: options.code || '',
    password: options.password || ''
  })

  if (result.code !== 0) {
    throw new Error(result.message || '密码找回失败')
  }

  return result.data
}

module.exports = {
  loginByWechat,
  sendPhoneCode,
  verifyPhoneCode,
  loginByPhone,
  loginByPassword,
  resetPassword
}
