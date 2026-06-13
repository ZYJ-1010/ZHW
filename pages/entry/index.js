const { ROUTES } = require('../../config/routes')

Page({
  data: {
    jumped: false,
    mode: 'guest',
    isUiPreview: false
  },

  onLoad(options) {
    const isUiPreview = options.ui === '1'

    this.setData({
      mode: options.mode === 'login' ? 'login' : 'guest',
      isUiPreview
    })

    if (isUiPreview) {
      return
    }

    this.timer = setTimeout(() => {
      this.goNext()
    }, 1000)
  },

  onUnload() {
    if (this.timer) {
      clearTimeout(this.timer)
    }
  },

  goNext() {
    if (this.data.isUiPreview) {
      return
    }

    if (this.data.jumped) {
      return
    }

    this.setData({
      jumped: true
    })

    wx.redirectTo({
      url: `/${this.data.mode === 'login' ? ROUTES.home : ROUTES.guestHome}`
    })
  },

  goGuestHome() {
    this.goNext()
  }
})
