const toast = require('../../utils/toast')
const profileService = require('../../services/profile')
const { ROUTES } = require('../../config/routes')
const { navigateShellKey, navigateShellRoute } = require('../../utils/shell-nav')
const { AUTH_EXPIRED_MESSAGE, goLogin, isAuthExpiredError } = require('../../utils/auth-error')
const { getActiveRole } = require('../../utils/active-role')
const { toUserMessage } = require('../../utils/user-message')

function createEmptyProfileHomeData() {
  return {
    onlineText: '在线',
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
      { label: '完成局数', value: '0' },
      { label: '信用度', value: '0' }
    ],
    assets: [
      { label: '可用积分', value: '0', tone: 'green' },
      { label: '累计经验', value: '0', tone: 'orange' },
      { label: '信用分', value: '0', tone: '' }
    ],
    vipBanner: null,
    serviceSections: []
  }
}

Page({
  data: {
    loaded: false,
    loadError: '',
    loadErrorActionText: '',
    loadErrorAuthExpired: false,
    ...createEmptyProfileHomeData()
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
      const data = await profileService.getProfileHome({ roleType: getActiveRole() })

      this.setData(Object.assign({}, normalizeProfileHome(data), {
        loadError: '',
        loadErrorActionText: '',
        loadErrorAuthExpired: false
      }))
    } catch (error) {
      const authExpired = isAuthExpiredError(error)

      this.setData(Object.assign({}, createEmptyProfileHomeData(), {
        loaded: true,
        loadError: authExpired ? AUTH_EXPIRED_MESSAGE : toUserMessage(error && error.message, '个人中心加载失败'),
        loadErrorActionText: authExpired ? '去登录' : '',
        loadErrorAuthExpired: authExpired,
        serviceSections: []
      }))
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
    navigateShellRoute(`/${ROUTES.profileAssetPoints}`, {
      currentRoute: ROUTES.profile
    })
  },

  handleUpgradeTap() {
    if (!this.data.vipBanner || this.data.vipBanner.visible !== true || !this.data.vipBanner.route) {
      return
    }
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
  const patch = Object.assign({ loaded: true }, createEmptyProfileHomeData())

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
    patch.serviceSections = removeTaskCenterEntry(data.serviceSections)
  } else if (Array.isArray(data.sections)) {
    patch.serviceSections = removeTaskCenterEntry(data.sections)
  }

  patch.vipBanner = data.vipBanner && data.vipBanner.visible === true ? data.vipBanner : null

  return patch
}

function removeTaskCenterEntry(sections) {
  return (Array.isArray(sections) ? sections : []).map((section) => ({
    ...section,
    items: (Array.isArray(section.items) ? section.items : []).filter((item) => (
      item && item.key !== 'taskCenter' && item.title !== '任务中心'
    ))
  }))
}
