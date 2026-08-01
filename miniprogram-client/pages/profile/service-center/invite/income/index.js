const toast = require('../../../../../utils/toast')
const { navigateShellRoute } = require('../../../../../utils/shell-nav')

Page({
  onLoad() {
    toast.info('收益功能暂未开放')
    setTimeout(() => {
      const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []
      if (pages.length > 1 && typeof wx.navigateBack === 'function') {
        wx.navigateBack()
        return
      }
      navigateShellRoute('/pages/profile/service-center/invite/overview/index', {
        reuseExisting: false
      })
    }, 0)
  }
})
