const profileService = require('../../../../../services/profile')

Page({
  data: {
    profile: {
      level: '',
      name: '',
      desc: ''
    },
    metrics: [],
    actions: [
      { icon: '🔗', label: '分享邀请码' },
      { icon: '▦', label: '二维码', iconClass: 'white' },
      { icon: '▧', label: '生成海报', iconClass: 'white' }
    ],
    tabs: ['数据概览', '关系网络', '邀约记录', '贡献排行', '收益明细'],
    trends: []
  },

  onLoad() {
    this.loadOverview()
  },

  async loadOverview() {
    try {
      const data = await profileService.getInviteOverview()

      this.setData({
        profile: data.profile || data.user || this.data.profile,
        metrics: Array.isArray(data.metrics || data.stats) ? (data.metrics || data.stats) : [],
        trends: Array.isArray(data.trends || data.trendItems) ? (data.trends || data.trendItems) : []
      })
    } catch (error) {
      wx.showToast({
        title: error.message || '邀请概览加载失败',
        icon: 'none'
      })
    }
  }
})
