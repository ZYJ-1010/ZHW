const api = require('../api/request')

function wxLogin() {
  return new Promise((resolve, reject) => {
    wx.login({
      success: (res) => resolve(res.code || ''),
      fail: reject
    })
  })
}

async function loginByWechat(options) {
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

module.exports = {
  loginByWechat
}
