const inviteService = require('../../../services/invite')
const { ROUTES } = require('../../../config/routes')

function decodeScene(scene) {
  const decoded = decodeURIComponent(String(scene || '')).trim()

  if (!decoded) {
    return {}
  }

  return decoded.split('&').reduce((result, item) => {
    const pair = item.split('=')
    const key = pair[0]

    if (key) {
      result[key] = pair[1] || ''
    }

    return result
  }, {})
}

function resolveInviteParams(options = {}) {
  const sceneParams = decodeScene(options.scene)
  const inviteCode = inviteService.normalizeInviteCode(
    options.inviteCode || options.code || sceneParams.inviteCode || sceneParams.code || options.scene
  )
  const entryType = String(options.entryType || sceneParams.entryType || '').trim()

  return {
    inviteCode,
    entryType
  }
}

Page({
  onLoad(options = {}) {
    const { inviteCode, entryType } = resolveInviteParams(options)
    const query = [
      inviteCode ? `inviteCode=${encodeURIComponent(inviteCode)}` : '',
      entryType ? `entryType=${encodeURIComponent(entryType)}` : ''
    ].filter(Boolean).join('&')
    const url = `/${ROUTES.login}${query ? `?${query}` : ''}`

    wx.redirectTo({
      url,
      fail: () => {
        wx.reLaunch({ url })
      }
    })
  }
})
