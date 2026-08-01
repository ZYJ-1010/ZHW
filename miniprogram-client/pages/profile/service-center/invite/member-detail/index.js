const profileService = require('../../../../../services/profile')
const { toUserMessage } = require('../../../../../utils/user-message')

Page({
  data: {
    member: {
      avatar: '',
      name: '',
      level: ''
    },
    stats: [],
    activities: [],
    loadError: ''
  },

  onLoad(options) {
    this.setData({
      memberId: options.memberId || options.id || ''
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
        loadError: toUserMessage(error && error.message, '成员详情加载失败')
      })
    }
  }
})
