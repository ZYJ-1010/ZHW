const toast = require('../../utils/toast')
const profileService = require('../../services/profile')
const { ROUTES } = require('../../config/routes')
const { navigateShellKey, navigateShellRoute } = require('../../utils/shell-nav')

Page({
  data: {
    loaded: false,
    loadError: '',
    user: {
      nickname: '未登录',
      memberLevel: '',
      roleLevel: '',
      role: '',
      avatarText: '我'
    },
    stats: [
      { label: '引荐数', value: '0' },
      { label: '成功数', value: '0' },
      { label: '成交总额', value: '¥0.00' },
      { label: '信用度', value: '0' }
    ],
    assets: [
      { label: '总成交额', value: '¥0.00', tone: '' },
      { label: '可提现', value: '¥0.00', tone: 'green' },
      { label: '待结算', value: '¥0.00', tone: 'orange' }
    ],
    vipBanner: {
      text: '升级会员，认证您的角色',
      actionText: '增购会员 >',
      route: '/pages/profile/member/index'
    },
    serviceSections: []
  },

  onLoad() {
    this.loadProfileHome()
  },

  onShow() {
    if (this.data.loaded) {
      this.loadProfileHome()
    }
  },

  async loadProfileHome() {
    try {
      const data = await profileService.getProfileHome()

      this.setData(Object.assign({}, normalizeProfileHome(data), {
        loadError: ''
      }))
    } catch (error) {
      this.setData({
        loaded: true,
        loadError: error.message || '个人中心加载失败',
        serviceSections: []
      })
      console.warn('[profile] load home failed', error)
    }
  },

  handleMenuTap(event) {
    const { route, enabled, title, reason } = event.currentTarget.dataset

    if (enabled === false || enabled === 'false' || !route) {
      toast.info(reason || `${title || '该入口'}暂未开放`)
      return
    }

    navigateShellRoute(route, {
      currentRoute: ROUTES.profile
    })
  },

  handleAssetAllTap() {
    navigateShellRoute('/pages/profile/asset-center/manage/index', {
      currentRoute: ROUTES.profile
    })
  },

  handleUpgradeTap() {
    navigateShellRoute(this.data.vipBanner.route || '/pages/profile/member/index', {
      currentRoute: ROUTES.profile
    })
  },

  handleShellNavTap(event) {
    const { key } = event.detail || {}

    navigateShellKey(key, {
      currentRoute: ROUTES.profile
    })
  }
})

function normalizeProfileHome(data = {}) {
  const patch = { loaded: true }

  if (data.user) {
    patch.user = data.user
  }

  if (Array.isArray(data.stats)) {
    patch.stats = data.stats
  }

  if (Array.isArray(data.assets)) {
    patch.assets = data.assets
  } else if (data.assets && Array.isArray(data.assets.summary)) {
    patch.assets = data.assets.summary
  }

  if (Array.isArray(data.serviceSections)) {
    patch.serviceSections = data.serviceSections
  } else if (Array.isArray(data.sections)) {
    patch.serviceSections = data.sections
  }

  if (data.vipBanner) {
    patch.vipBanner = data.vipBanner
  }

  return patch
}
