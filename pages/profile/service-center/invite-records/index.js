const profileService = require('../../../../services/profile')

Page({
  data: {
    roleTabs: [
      { label: '我引荐的', active: true },
      { label: '我发起的', active: false }
    ],
    filters: [
      { label: '全部', active: true },
      { label: '进行中', active: false },
      { label: '已完成', active: false },
      { label: '超时', active: false },
      { label: '已取消', active: false }
    ],
    records: []
  },

  onLoad() {
    this.loadRecords()
  },

  async loadRecords() {
    try {
      const data = await profileService.getInviteRecords()

      this.setData({
        records: Array.isArray(data.records || data.list || data.items) ? (data.records || data.list || data.items) : []
      })
    } catch (error) {
      wx.showToast({
        title: error.message || '邀请记录加载失败',
        icon: 'none'
      })
    }
  }
})
