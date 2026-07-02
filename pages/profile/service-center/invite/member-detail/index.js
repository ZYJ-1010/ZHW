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
    activities: [],
    loadError: ''
  },

  onLoad(options) {
    this.setData({
      memberId: options.id || ''
    }, () => this.loadMemberDetail())
  },

  async loadMemberDetail() {
    try {
      const result = await profileService.getInviteMemberDetail({
        memberId: this.data.memberId
      })

      this.setData(result || {})
    } catch (error) {
      console.warn('get invite member detail failed', error)
      this.setData({
        loadError: error.message || '成员详情加载失败'
      })
    }
  }
})
