const profileService = require('../../../../../services/profile')

Page({
  data: {
    member: {
      avatar: '',
      name: '',
      level: ''
    },
    stats: [],
    income: [],
    activities: []
  },

  onLoad(options) {
    this.setData({
      memberId: options.id || ''
    })
    this.loadMemberDetail()
  },

  async loadMemberDetail() {
    if (!this.data.memberId) {
      return
    }

    try {
      const data = await profileService.getInviteMemberDetail({
        memberId: this.data.memberId
      })

      this.setData({
        member: data.member || data.profile || this.data.member,
        stats: Array.isArray(data.stats) ? data.stats : [],
        income: Array.isArray(data.income || data.incomeItems) ? (data.income || data.incomeItems) : [],
        activities: Array.isArray(data.activities || data.records) ? (data.activities || data.records) : []
      })
    } catch (error) {
      wx.showToast({
        title: error.message || '成员详情加载失败',
        icon: 'none'
      })
    }
  }
})
