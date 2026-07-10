const homeService = require('../../services/home')

const DEFAULT_TOPBAR_HEIGHT_RPX = 182
const BRAND_ONLINE_BOTTOM_OFFSET_RPX = 82

Page({
  data: {
    onlineText: '',
    entryTopbarStyle: '',
    entryStatusFillStyle: '',
    entryNavStyle: '',
    entryBrandStyle: ''
  },

  onLoad() {
    this.updateEntryTopbarLayout()
    this.loadOnlineText()
  },

  onShow() {
    this.updateEntryTopbarLayout()
  },

  onResize() {
    this.updateEntryTopbarLayout()
  },

  goGuestHome() {
    // Static walkthrough only. Wire this page into the app flow after the UI is confirmed.
  },

  async loadOnlineText() {
    try {
      const data = await homeService.getHome({})
      const hero = data && data.hero ? data.hero : {}
      const onlineText = hero.onlineText || data.onlineText || ''

      this.setData({
        onlineText
      })
    } catch (error) {
      this.setData({
        onlineText: ''
      })
    }
  },

  roundRpx(value) {
    return Math.round(value * 100) / 100
  },

  getWindowInfo() {
    if (wx.getWindowInfo) {
      return wx.getWindowInfo()
    }

    if (wx.getSystemInfoSync) {
      return wx.getSystemInfoSync()
    }

    return null
  },

  updateEntryTopbarLayout() {
    if (!wx.getMenuButtonBoundingClientRect) {
      return
    }

    const menuButton = wx.getMenuButtonBoundingClientRect()
    const windowInfo = this.getWindowInfo()

    if (!menuButton || !menuButton.width || !menuButton.height || !windowInfo || !windowInfo.windowWidth) {
      return
    }

    const ratio = 750 / windowInfo.windowWidth
    const statusBarHeight = Number(windowInfo.statusBarHeight) || Math.max(0, menuButton.top - 4)
    const capsuleTopGap = Math.max(0, menuButton.top - statusBarHeight)
    const navHeightPx = capsuleTopGap * 2 + menuButton.height
    const statusFillHeight = this.roundRpx(statusBarHeight * ratio)
    const navHeight = this.roundRpx(navHeightPx * ratio)
    const topbarHeight = Math.max(DEFAULT_TOPBAR_HEIGHT_RPX, this.roundRpx((statusBarHeight + navHeightPx) * ratio))
    const capsuleBottom = this.roundRpx((menuButton.top + menuButton.height) * ratio)
    const brandTop = this.roundRpx(capsuleBottom - statusFillHeight - BRAND_ONLINE_BOTTOM_OFFSET_RPX)

    this.setData({
      entryTopbarStyle: `height: ${topbarHeight}rpx;`,
      entryStatusFillStyle: `height: ${statusFillHeight}rpx;`,
      entryNavStyle: `height: ${navHeight}rpx;`,
      entryBrandStyle: `top: ${brandTop}rpx;`
    })
  }
})
