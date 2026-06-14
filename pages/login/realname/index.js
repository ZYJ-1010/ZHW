const { ROUTES } = require('../../../config/routes')

Page({
  data: {
    isCompleting: false
  },

  completeRealname() {
    if (this.data.isCompleting) {
      return
    }

    this.setData({
      isCompleting: true
    })
    wx.setStorageSync('enjoy_mock_realname_verified', '1')
    wx.redirectTo({
      url: `/${ROUTES.login}?ui=1&mode=newbieTasks`
    })
  }
})
