const homeService = require('../../services/home')
const inviteService = require('../../services/invite')
const authService = require('../../services/auth')
const { ROUTES } = require('../../config/routes')
const { getAuthToken } = require('../../utils/auth-session')
const entryLayout = require('./layout')

const DEFAULT_ONLINE_COUNT = '0'
function formatOnlineText(value) {
  const match = String(value || '').match(/\d[\d,]*/)
  const count = match ? match[0].replace(/,/g, '') : DEFAULT_ONLINE_COUNT

  return `在线${count}人`
}

function safeDecode(value) {
  const text = String(value || '').trim()
  try {
    return decodeURIComponent(text)
  } catch (error) {
    return text
  }
}

function decodeInviteScene(scene) {
  const decoded = safeDecode(scene)
  if (!decoded) {
    return {}
  }
  if (!decoded.includes('=') && !decoded.includes('&')) {
    return { inviteCode: decoded }
  }
  return decoded.split('&').reduce((result, item) => {
    const index = item.indexOf('=')
    const key = safeDecode(index >= 0 ? item.slice(0, index) : item)
    const value = safeDecode(index >= 0 ? item.slice(index + 1) : '')
    if (key) {
      result[key] = value
    }
    return result
  }, {})
}

function normalizeEntryType(value) {
  const entryType = String(value || '').trim().toLowerCase()
  const map = {
    p: 'poster',
    poster: 'poster',
    card: 'poster',
    q: 'qrcode',
    qr: 'qrcode',
    qrcode: 'qrcode',
    code: 'qrcode',
    l: 'link',
    link: 'link'
  }
  return map[entryType] || entryType
}

function resolveInviteParams(options = {}) {
  const scene = decodeInviteScene(options.scene)
  return {
    inviteCode: inviteService.normalizeInviteCode(
      options.inviteCode || options.code || scene.inviteCode || scene.code || scene.i
    ),
    entryType: normalizeEntryType(
      options.entryType || options.type || scene.entryType || scene.type || scene.t || scene.entry
    )
  }
}

Page({
  data: {
    onlineText: formatOnlineText(''),
    entryTopbarStyle: '',
    entryStatusFillStyle: '',
    entryNavStyle: '',
    entryBrandStyle: '',
    entryContentStyle: '',
    entryHeroStyle: '',
    entryTitleStyle: '',
    entrySubtitleStyle: '',
    entryAccentStyle: '',
    inviteCode: '',
    entryType: '',
    isInviteVerifying: false,
    isInviteVerified: false,
    isInviteNavigating: false,
    isWechatChecking: false
  },

  onLoad(options = {}) {
    const invite = resolveInviteParams(options)
    this.setData(invite)
    this.updateEntryTopbarLayout()
    this.loadOnlineText()
    if (!invite.inviteCode && this.restoreExistingSession()) {
      return
    }
    this.checkWechatEntryAndContinue()
  },

  onShow() {
    this.updateEntryTopbarLayout()
  },

  onUnload() {
    // No fixed startup timer: navigation is completed by the account precheck.
  },

  onResize() {
    this.updateEntryTopbarLayout()
  },

  goGuestHome() {
    if (this.data.isInviteVerifying || this.data.isWechatChecking) {
      wx.showToast({ title: '账号检测中', icon: 'none' })
      return
    }
    if (this.restoreExistingSession()) {
      return
    }
    if (!this.data.inviteCode) {
      this.continueToLoginWithoutInvite()
      return
    }
    if (this.data.isInviteVerified) {
      this.continueToPhoneRegister(inviteService.getInviteContext())
      return
    }
    this.verifyInviteAndContinue()
  },

  async checkWechatEntryAndContinue() {
    if (this.data.isWechatChecking || this.data.isInviteNavigating) {
      return
    }
    this.setData({ isWechatChecking: true })
    try {
      const entry = await authService.precheckWechatEntry()
      if (entry.boundWechat) {
        await this.loginBoundWechatAndContinue()
        return
      }
      if (this.data.inviteCode) {
        await this.verifyInviteAndContinue()
        return
      }
      this.continueToLoginWithoutInvite()
    } catch (error) {
      // Do not hold the startup page on a network failure. The login page keeps
      // both manual phone login and a user-triggered WeChat retry available.
      if (this.data.inviteCode) {
        await this.verifyInviteAndContinue()
      } else {
        this.continueToLoginWithoutInvite()
      }
    } finally {
      this.setData({ isWechatChecking: false })
    }
  },

  async loginBoundWechatAndContinue() {
    if (this.data.isInviteNavigating) {
      return
    }
    this.setData({ isInviteNavigating: true })
    try {
      // The precheck only determines whether this WeChat account is bound.
      // Obtain a fresh one-time code for the actual session login, then enter
      // the app directly instead of showing the login screen again.
      const loginData = await authService.loginByWechat()
      if (!loginData || !loginData.token) {
        throw new Error('微信登录凭证无效')
      }
      const url = `/${ROUTES.playerHome}`
      wx.reLaunch({
        url,
        fail: () => wx.redirectTo({ url })
      })
    } catch (error) {
      // Keep the existing login page as the recovery path when WeChat's
      // temporary credential cannot be exchanged, such as a network failure.
      this.setData({ isInviteNavigating: false })
      this.continueToRegisteredLogin()
    }
  },

  async verifyInviteAndContinue() {
    if (this.data.isInviteVerifying || this.data.isInviteNavigating) {
      return
    }
    this.setData({ isInviteVerifying: true })
    try {
      const result = await inviteService.verifyInviteCode(this.data.inviteCode, this.data.entryType)
      if (!result || result.status !== 'valid' || !result.invite) {
        this.setData({ isInviteVerified: false })
        if (result && result.expired) {
          this.handleExpiredInvite()
          return
        }
        wx.showToast({ title: result && result.message ? result.message : '邀请码无效', icon: 'none' })
        return
      }
      const inviteContext = inviteService.saveInviteContext(result.invite)
      if (!inviteContext) {
        throw new Error('邀请码返回格式异常')
      }
      this.setData({
        inviteCode: inviteContext.code,
        entryType: inviteContext.entryType || this.data.entryType,
        isInviteVerified: true
      })
      this.continueToPhoneRegister(inviteContext)
    } catch (error) {
      this.setData({ isInviteVerified: false })
      wx.showToast({ title: error.message || '邀请码校验失败', icon: 'none' })
    } finally {
      this.setData({ isInviteVerifying: false })
    }
  },

  handleExpiredInvite() {
    inviteService.clearInviteContext()
    this.setData({ inviteCode: '', entryType: '', isInviteVerified: false })
    wx.showModal({
      title: '邀请码已失效',
      content: '当前邀请码已失效，请重新获取邀请码后再次进入。',
      showCancel: false,
      confirmText: '重新获取',
      success: () => {
        const url = `/${ROUTES.login}?flow=register`
        wx.redirectTo({ url, fail: () => wx.reLaunch({ url }) })
      }
    })
  },

  continueToPhoneRegister(inviteContext) {
    if (!inviteContext || !inviteContext.code || this.data.isInviteNavigating) {
      return
    }
    this.setData({ isInviteNavigating: true })
    const query = [
      `inviteCode=${encodeURIComponent(inviteContext.code)}`,
      inviteContext.entryType ? `entryType=${encodeURIComponent(inviteContext.entryType)}` : ''
    ].filter(Boolean).join('&')
    const url = `/${ROUTES.login}?flow=register${query ? `&${query}` : ''}`
    wx.redirectTo({
      url,
      fail: () => {
        wx.reLaunch({ url })
      }
    })
  },

  continueToLoginWithoutInvite() {
    if (this.data.isInviteNavigating) {
      return
    }
    this.setData({ isInviteNavigating: true })
    const url = `/${ROUTES.login}?flow=register`
    wx.redirectTo({
      url,
      fail: () => {
        wx.reLaunch({ url })
      }
    })
  },

  continueToRegisteredLogin() {
    if (this.data.isInviteNavigating) {
      return
    }
    this.setData({ isInviteNavigating: true })
    const url = `/${ROUTES.login}?flow=existing`
    wx.redirectTo({
      url,
      fail: () => {
        wx.reLaunch({ url })
      }
    })
  },

  async loadOnlineText() {
    try {
      const data = await homeService.getHome({})
      const hero = data && data.hero ? data.hero : {}
      const onlineText = hero.onlineText || data.onlineText || ''

      this.setData({
        onlineText: formatOnlineText(onlineText)
      })
    } catch (error) {
      this.setData({
        onlineText: formatOnlineText('')
      })
    }
  },

  restoreExistingSession() {
    if (!getAuthToken()) {
      return false
    }

    const url = `/${ROUTES.playerHome}`
    wx.reLaunch({
      url,
      fail: () => {
        wx.redirectTo({ url })
      }
    })
    return true
  },

  updateEntryTopbarLayout() {
    const layout = entryLayout.getEntryBrandSingleScreenLayout()

    if (!layout) {
      return
    }

    this.setData({
      entryTopbarStyle: `height: ${layout.topbarHeightRpx}rpx;`,
      entryStatusFillStyle: `height: ${layout.statusBarHeightRpx}rpx;`,
      entryNavStyle: `height: ${layout.navigationHeightRpx}rpx;`,
      entryBrandStyle: `top: ${layout.brandTopRpx}rpx;`,
      entryContentStyle: `height: ${layout.contentHeightRpx}rpx; min-height: 0;`,
      entryHeroStyle: `width: ${layout.heroSizeRpx}rpx; height: ${layout.heroSizeRpx}rpx; margin-top: ${layout.heroTopRpx}rpx; --entry-bounce-large: ${layout.bounceLargeRpx}rpx; --entry-bounce-small: ${layout.bounceSmallRpx}rpx; --entry-go-font-size: ${layout.goFontSizeRpx}rpx;`,
      entryTitleStyle: `width: ${layout.titleWidthRpx}rpx; height: ${layout.titleHeightRpx}rpx; margin-top: ${layout.titleTopGapRpx}rpx;`,
      entrySubtitleStyle: `margin-top: ${layout.subtitleTopGapRpx}rpx; font-size: ${layout.subtitleFontSizeRpx}rpx; line-height: ${layout.subtitleLineHeightRpx}rpx; letter-spacing: ${layout.subtitleLetterSpacingRpx}rpx;`,
      entryAccentStyle: `top: ${layout.accentTopRpx}rpx;`
    })
  }
})
