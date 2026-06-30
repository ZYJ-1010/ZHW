const profileService = require('../../../../../services/profile')

Page({
  data: {
    activePeriodIndex: 0,
    activeType: 'inviteCount',
    periods: [
      { key: 'week', label: '本周' },
      { key: 'month', label: '本月' },
      { key: 'quarter', label: '本季' },
      { key: 'year', label: '本年' },
      { key: 'all', label: '全部' }
    ],
    rankTypes: [
      { key: 'inviteCount', label: '邀约数排行' },
      { key: 'profitContribution', label: '分润贡献排行' }
    ],
    members: []
  },

  onLoad() {
    this.loadRanking()
  },

  async loadRanking() {
    try {
      const period = this.data.periods[this.data.activePeriodIndex] || {}
      const data = await profileService.getInviteRanking({
        period: period.key,
        type: this.data.activeType
      })

      this.setData({
        members: Array.isArray(data.members || data.list || data.items) ? (data.members || data.list || data.items) : []
      })
    } catch (error) {
      wx.showToast({
        title: error.message || '邀请排行加载失败',
        icon: 'none'
      })
    }
  },

  handlePeriodTap(event) {
    const { index } = event.currentTarget.dataset

    if (index === this.data.activePeriodIndex) {
      return
    }

    this.setData({
      activePeriodIndex: index
    })
    this.loadRanking()
  },

  handleTypeTap(event) {
    const { type } = event.currentTarget.dataset

    if (!type || type === this.data.activeType) {
      return
    }

    this.setData({
      activeType: type
    })
    this.loadRanking()
  }
})
