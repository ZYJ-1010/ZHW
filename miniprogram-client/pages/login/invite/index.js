const inviteService = require('../../../services/invite')
const { ROUTES } = require('../../../config/routes')

function decodeScene(scene) {
  const rawScene = String(scene || '').trim()
  let decoded = rawScene

  try {
    decoded = decodeURIComponent(rawScene)
  } catch (error) {
    decoded = rawScene
  }

  if (!decoded) {
    return {}
  }

  if (!decoded.includes('=') && !decoded.includes('&')) {
    return { inviteCode: decoded }
  }

  return decoded.split('&').reduce((result, item) => {
    const index = item.indexOf('=')
    const key = index >= 0 ? item.slice(0, index) : item
    const value = index >= 0 ? item.slice(index + 1) : ''

    if (key) {
      result[key] = value
    }

    return result
  }, {})
}

function normalizeSceneEntryType(entryType) {
  const value = String(entryType || '').trim().toLowerCase()
  const map = {
    p: 'poster',
    q: 'qrcode',
    l: 'link',
    qr: 'qrcode',
    code: 'qrcode',
    poster: 'poster',
    qrcode: 'qrcode',
    link: 'link',
    card: 'poster',
    '小程序卡片': 'poster',
    '海报': 'poster',
    '二维码': 'qrcode',
    '链接': 'link'
  }
  return map[value] || value
}

function resolveInviteParams(options = {}) {
  const sceneParams = decodeScene(options.scene)
  const inviteCode = inviteService.normalizeInviteCode(
    options.inviteCode || options.code || sceneParams.inviteCode || sceneParams.code || sceneParams.i
  )
  const entryType = normalizeSceneEntryType(
    options.entryType || options.type || sceneParams.entryType || sceneParams.type ||
    sceneParams.t || sceneParams.entry || sceneParams['邀请入口'] || ''
  )

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
