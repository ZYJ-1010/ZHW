const profileService = require('../../services/profile')
const toast = require('../../utils/toast')

Page({
  data: {
    loading: true,
    user: {
      nickname: '',
      avatarText: '',
      realnameStatus: '',
      roleText: '',
      level: '',
      creditScore: '',
      points: ''
    },
    stats: [],
    menuItems: [],
    incomeSummary: {
      pendingAmountText: ''
    }
  },

  onLoad() {
    this.loadProfile()
  },

  async loadProfile() {
    try {
      const profile = await profileService.getProfileHome()
      const user = Object.assign({}, profile.user, {
        avatarText: String(profile.user.nickname || '微').slice(0, 1),
        roleText: (profile.user.roles || []).join(' / ')
      })

      this.setData({
        loading: false,
        user,
        stats: profile.stats,
        menuItems: profile.menuItems,
        incomeSummary: profile.incomeSummary
      })
    } catch (error) {
      this.setData({
        loading: false
      })
      toast.info(error.message || '个人中心加载失败')
    }
  },

  goMenu(event) {
    const route = event.currentTarget.dataset.route

    if (!route) {
      return
    }

    toast.developing()
  }
})
