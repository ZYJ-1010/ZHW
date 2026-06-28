const toast = require('../../../../utils/toast')

const ASSET_BASE = '/pages/profile/system-management/feedback/assets'

Page({
  data: {
    icons: {
      check: `${ASSET_BASE}/icon-check-white.svg`,
      star: `${ASSET_BASE}/icon-star-outline.svg`
    },
    score: 7,
    scores: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
    reasons: [
      { value: '反馈流程简单', selected: false },
      { value: '响应速度快', selected: false },
      { value: '客服态度好', selected: false },
      { value: '问题解决彻底', selected: false },
      { value: '界面清晰易用', selected: false },
      { value: '有积分激励', selected: false }
    ]
  },

  handleScoreTap(event) {
    const { score } = event.currentTarget.dataset

    this.setData({
      score: Number(score)
    })
  },

  handleReasonTap(event) {
    const reasonIndex = Number(event.currentTarget.dataset.index)

    if (Number.isNaN(reasonIndex)) {
      return
    }

    const selected = !this.data.reasons[reasonIndex].selected

    this.setData({
      [`reasons[${reasonIndex}].selected`]: selected
    })
  },

  handleSubmitRating() {
    toast.success('评价已记录')
  },

  handleBackHome() {
    wx.redirectTo({
      url: '/pages/profile/system-management/feedback/index'
    })
  },

  handleViewRecords() {
    wx.redirectTo({
      url: '/pages/profile/system-management/feedback-records/index'
    })
  }
})
