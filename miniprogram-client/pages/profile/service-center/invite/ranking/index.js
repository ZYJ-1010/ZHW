const profileService = require('../../../../../services/profile')
const { toUserMessage } = require('../../../../../utils/user-message')

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
    rankTypes: [{ key: 'inviteCount', label: '邀约数排行' }],
    members: [],
    loadError: ''
  },

  onLoad() {
    this.loadRanking()
  },

  async loadRanking() {
    try {
      const period = this.data.periods[this.data.activePeriodIndex] || {}
      const result = await profileService.getInviteRanking({
        period: period.key,
        type: this.data.activeType
      })

      this.setData(result || {})
    } catch (error) {
      console.warn('get invite ranking failed', error)
      this.setData({
        loadError: toUserMessage(error && error.message, '贡献排行加载失败')
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
    }, () => this.loadRanking())
  },

  handleTypeTap(event) {
    const { type } = event.currentTarget.dataset

    if (!type || type === this.data.activeType) {
      return
    }

    this.setData({
      activeType: type
    }, () => this.loadRanking())
  }
})
