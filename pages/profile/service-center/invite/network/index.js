const profileService = require('../../../../../services/profile')

Page({
  data: {
    summary: [],
    networkNodes: [],
    avatars: [],
    members: []
  },

  onLoad() {
    this.loadNetwork()
  },

  async loadNetwork() {
    try {
      const data = await profileService.getInviteNetwork()

      this.setData({
        summary: Array.isArray(data.summary || data.stats) ? (data.summary || data.stats) : [],
        networkNodes: Array.isArray(data.networkNodes || data.nodes) ? (data.networkNodes || data.nodes) : [],
        avatars: Array.isArray(data.avatars) ? data.avatars : [],
        members: Array.isArray(data.members || data.list) ? (data.members || data.list) : []
      })
    } catch (error) {
      wx.showToast({
        title: error.message || '邀请网络加载失败',
        icon: 'none'
      })
    }
  },

  handleMemberTap(event) {
    const { id } = event.currentTarget.dataset

    wx.navigateTo({
      url: `/pages/profile/service-center/invite/member-detail/index?id=${id}`
    })
  }
})
