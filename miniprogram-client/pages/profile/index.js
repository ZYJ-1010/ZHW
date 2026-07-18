const toast = require('../../utils/toast')
const profileService = require('../../services/profile')
const { ROUTES } = require('../../config/routes')
const { navigateShellKey, navigateShellRoute } = require('../../utils/shell-nav')
const { AUTH_EXPIRED_MESSAGE, goLogin, isAuthExpiredError } = require('../../utils/auth-error')

Page({
  data: {
    loaded: false,
    loadError: '',
    loadErrorActionText: '',
    loadErrorAuthExpired: false,
    onlineText: '在线0人',
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
        loadError: '',
        loadErrorActionText: '',
        loadErrorAuthExpired: false
      }))
    } catch (error) {
      const authExpired = isAuthExpiredError(error)

      this.setData({
        loaded: true,
        loadError: authExpired ? AUTH_EXPIRED_MESSAGE : (error.message || '个人中心加载失败'),
        loadErrorActionText: authExpired ? '去登录' : '',
        loadErrorAuthExpired: authExpired,
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

  handleAvatarTap() {
    if (this.data.loadErrorAuthExpired) {
      goLogin(ROUTES.profile)
      return
    }

    navigateShellRoute(`/${ROUTES.profileSystemProfileInfo}`, {
      currentRoute: ROUTES.profile
    })
  },

  handleLoginTap() {
    goLogin(ROUTES.profile)
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

  if (data.onlineText) {
    patch.onlineText = data.onlineText
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
    patch.serviceSections = ensureTaskCenterEntry(data.serviceSections)
  } else if (Array.isArray(data.sections)) {
    patch.serviceSections = ensureTaskCenterEntry(data.sections)
  }

  if (data.vipBanner) {
    patch.vipBanner = data.vipBanner
  }

  return patch
}

function ensureTaskCenterEntry(sections) {
  const result = (Array.isArray(sections) ? sections : []).map((section) => ({
    ...section,
    items: Array.isArray(section.items) ? section.items.slice() : []
  }))
  const assetSection = result.find((section) => section && section.title === '资产中心')
  if (!assetSection || assetSection.items.some((item) => item && (item.key === 'taskCenter' || item.title === '任务中心'))) {
    return result
  }
  assetSection.items.push({
    key: 'taskCenter',
    title: '任务中心',
    iconSrc: '/pages/profile/assets/i73@3x.png',
    iconClass: 'orange',
    badge: '',
    badgeClass: '',
    route: '/pages/profile/task-center/index',
    enabled: true
  })
  return result
}
