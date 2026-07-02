const profileService = require('../../../../../services/profile')
const { navigateShellRoute } = require('../../../../../utils/shell-nav')

Page({
  data: {
    summary: [
      { value: '0', label: '已服务\n位玩家' },
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

      this.setData(result || {})
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
