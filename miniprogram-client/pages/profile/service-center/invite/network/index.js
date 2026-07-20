const profileService = require('../../../../../services/profile')
const { navigateShellRoute } = require('../../../../../utils/shell-nav')

Page({
  data: {
    summary: [
      { value: '0', label: '已服务' },
      { value: '¥0', label: '本周收益' }
    ],
    networkNodes: [],
    avatars: [],
    members: [],
    loadError: ''
  },

  onLoad() {
    this.loadNetwork()
  },

  async loadNetwork() {
    try {
      const result = await profileService.getInviteNetwork()

      const payload = result || {}
      this.setData({
        ...payload,
        // 关系网概览只展示可读的头像数量，剩余人数由 +N 表达，避免挤压成员说明。
        avatars: Array.isArray(payload.avatars) ? payload.avatars.slice(0, 4) : []
      })
    } catch (error) {
      console.warn('get invite network failed', error)
      this.setData({
        loadError: error.message || '关系网络加载失败'
      })
    }
  },

  handleMemberTap(event) {
    const { id } = event.currentTarget.dataset

    navigateShellRoute(`/pages/profile/service-center/invite/member-detail/index?id=${id}`)
  },

  handleManageTap() {
    navigateShellRoute('/pages/profile/service-center/invite/records/index')
  },

  handleIncomeTap() {
    navigateShellRoute('/pages/profile/service-center/invite/income/index')
  }
})
