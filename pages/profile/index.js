const toast = require('../../utils/toast')
const profileService = require('../../services/profile')

const DEFAULT_SERVICE_SECTIONS = [
  {
    key: 'service',
    title: '服务中心',
    items: [
      { key: 'myGames', title: '我的局', iconSrc: '/pages/profile/assets/i66@3x.png', iconClass: 'purple-blue', route: '/pages/profile/service-center/my-games/index' },
      { key: 'invite', title: '我的邀请', iconSrc: '/pages/profile/assets/i68@3x.png', iconClass: 'purple-blue', route: '/pages/profile/service-center/invite/overview/index' },
      { key: 'reviews', title: '评价管理', iconSrc: '/pages/profile/assets/i69@3x.png', iconClass: 'purple-blue', route: '/pages/profile/service-center/manage/review-manage/index' }
    ]
  },
  {
    key: 'assets',
    title: '资产中心',
    items: [
      { key: 'assets', title: '我的资产', iconSrc: '/pages/profile/assets/i70@3x.png', iconClass: 'orange', route: '/pages/profile/asset-center/manage/index' },
      { key: 'deposit', title: '我的押金', iconSrc: '/pages/profile/assets/i71@3x.png', iconClass: 'orange' },
      { key: 'mall', title: '积分商城', iconSrc: '/pages/profile/assets/i72@3x.png', iconClass: 'orange', route: '/pages/profile/asset-center/mall/index' },
      { key: 'points', title: '我的积分', iconSrc: '/pages/profile/assets/i73@3x.png', iconClass: 'orange', route: '/pages/profile/asset-center/points/index' },
      { key: 'invoice', title: '开票中心', iconSrc: '/pages/profile/assets/i74@3x.png', iconClass: 'orange' }
    ]
  },
  {
    key: 'footprint',
    title: '足迹中心',
    items: [
      { key: 'footprint', title: '我的足迹', iconSrc: '/pages/profile/assets/i75@3x.png', iconClass: 'teal' },
      { key: 'cityStories', title: '我的城市故事', iconSrc: '/pages/profile/assets/i76@3x.png', iconClass: 'teal' },
      { key: 'achievements', title: '我的成就墙', iconSrc: '/pages/profile/assets/i77@3x.png', iconClass: 'teal', route: '/pages/profile/footprint/achievements/index' }
    ]
  },
  {
    key: 'system',
    title: '系统管理',
    items: [
      { key: 'profileInfo', title: '我的资料', iconSrc: '/pages/profile/assets/i78@3x.png', iconClass: 'blue-purple', route: '/pages/profile/system-management/profile-info/index' },
      { key: 'skills', title: '技能配置', iconSrc: '/pages/profile/assets/i79@3x.png', iconClass: 'blue-purple', route: '/pages/profile/system-management/skill-config/index' },
      { key: 'blockSettings', title: '屏蔽设置', iconSrc: '/pages/profile/assets/i80@3x.png', iconClass: 'blue-purple', route: '/pages/profile/system-management/block-settings/index' },
      { key: 'credit', title: '信用中心', iconSrc: '/pages/profile/assets/i81@3x.png', iconClass: 'blue-purple', route: '/pages/profile/credit-center/index' },
      { key: 'reports', title: '举报中心', iconSrc: '/pages/profile/assets/i82@3x.png', iconClass: 'blue-purple', route: '/pages/profile/system-management/report-center/index' },
      { key: 'agreements', title: '签署协议', iconSrc: '/pages/profile/assets/i83@3x.png', iconClass: 'blue-purple', route: '/pages/profile/system-management/agreement-sign/index' },
      { key: 'feedback', title: '建议反馈', iconSrc: '/pages/profile/assets/i84@3x.png', iconClass: 'blue-purple', route: '/pages/profile/system-management/feedback/index' },
      { key: 'settings', title: '系统设置', iconSrc: '/pages/profile/assets/i85@3x.png', iconClass: 'blue-purple', route: '/pages/profile/settings/index' }
    ]
  }
]

function createEmptyUser() {
  return {
    nickname: '',
    memberLevel: '',
    roleLevel: '',
    role: '',
    avatarText: '',
    avatarUrl: ''
  }
}

function cloneDefaultSections() {
  return DEFAULT_SERVICE_SECTIONS.map((section) => ({
    key: section.key,
    title: section.title,
    items: section.items.map((item) => Object.assign({}, item, {
      badge: '',
      badgeClass: ''
    }))
  }))
}

function normalizeRoute(route) {
  if (!route) {
    return ''
  }

  return route.indexOf('/') === 0 ? route : `/${route}`
}

function findDefaultSection(source = {}) {
  return DEFAULT_SERVICE_SECTIONS.find((section) => {
    return section.key === source.key || section.title === source.title
  }) || {}
}

function findDefaultEntry(section = {}, source = {}) {
  return (section.items || []).find((item) => {
    return item.key === source.key || item.title === source.title
  }) || {}
}

Page({
  data: {
    onlineText: '',
    user: createEmptyUser(),
    stats: [],
    assetTitle: '我的资产',
    assetMoreText: '查看全部',
    assets: [],
    vipBanner: null,
    serviceSections: cloneDefaultSections(),
    loading: false,
    errorText: ''
  },

  onLoad() {
    this.loadProfileHome()
  },

  async loadProfileHome() {
    this.setData({
      loading: true,
      errorText: ''
    })

    try {
      const home = await profileService.getProfileHome()
      const normalized = this.normalizeProfileHome(home || {})

      this.setData(Object.assign({}, normalized, {
        loading: false,
        errorText: ''
      }))
    } catch (error) {
      const errorText = error && error.message ? error.message : '个人中心加载失败'

      this.setData({
        loading: false,
        errorText,
        user: createEmptyUser(),
        stats: [],
        assets: [],
        vipBanner: null
      })
      toast.info(errorText)
    }
  },

  normalizeProfileHome(source = {}) {
    const assetSource = source.assets && !Array.isArray(source.assets) ? source.assets : {}

    return {
      onlineText: source.onlineText || '',
      user: this.normalizeUser(source.user || source.profile || source.currentUser || {}),
      stats: this.normalizeStats(source.stats || source.statItems || source.summaryStats),
      assetTitle: assetSource.title || source.assetTitle || this.data.assetTitle,
      assetMoreText: assetSource.moreText || source.assetMoreText || this.data.assetMoreText,
      assets: this.normalizeAssets(source),
      vipBanner: this.normalizeVipBanner(assetSource.vipBanner || source.vipBanner),
      serviceSections: this.normalizeServiceSections(source)
    }
  },

  normalizeUser(user = {}) {
    const nickname = user.nickname || user.displayName || user.name || ''
    const member = user.member || {}

    return {
      nickname,
      memberLevel: user.memberLevel || user.memberLevelText || user.memberName || member.planName || '',
      roleLevel: user.roleLevel || user.growthLevel || user.levelText || user.roleLevelText || '',
      role: user.role || user.roleName || user.roleText || '',
      avatarText: user.avatarText || (nickname ? nickname.trim().slice(0, 1) : ''),
      avatarUrl: user.avatarUrl || user.avatar || ''
    }
  },

  normalizeStats(stats) {
    if (!Array.isArray(stats)) {
      return []
    }

    return stats
      .map((item) => ({
        key: item.key || item.id || item.label,
        label: item.label || item.title || item.name || '',
        value: item.value != null ? String(item.value) : ''
      }))
      .filter((item) => item.label || item.value)
  },

  normalizeAssets(source = {}) {
    const assetSource = source.assets
    const summary = Array.isArray(assetSource)
      ? assetSource
      : (assetSource && (assetSource.summary || assetSource.items || assetSource.list)) || source.assetSummary

    if (!Array.isArray(summary)) {
      return []
    }

    return summary
      .map((item) => ({
        key: item.key || item.id || item.label,
        label: item.label || item.title || item.name || '',
        value: item.value != null ? String(item.value) : '',
        tone: item.tone || item.color || ''
      }))
      .filter((item) => item.label || item.value)
  },

  normalizeVipBanner(vipBanner) {
    if (!vipBanner || typeof vipBanner !== 'object') {
      return null
    }

    const text = vipBanner.text || vipBanner.desc || vipBanner.content || ''
    const actionText = vipBanner.actionText || vipBanner.buttonText || ''

    if (!text && !actionText) {
      return null
    }

    return {
      chip: vipBanner.chip || vipBanner.tag || 'VIP',
      text,
      actionText,
      route: normalizeRoute(vipBanner.route || vipBanner.url || '/pages/profile/member/index')
    }
  },

  normalizeServiceSections(source = {}) {
    const sections = source.sections || source.serviceSections || source.menuSections

    if (!Array.isArray(sections)) {
      return cloneDefaultSections()
    }

    return sections
      .filter((section) => section && section.visible !== false)
      .map((section) => {
        const defaultSection = findDefaultSection(section)
        const items = Array.isArray(section.items) ? section.items : []

        return {
          key: section.key || defaultSection.key || section.title,
          title: section.title || section.name || defaultSection.title || '',
          items: items
            .filter((item) => item && item.visible !== false && item.enabled !== false)
            .map((item) => this.normalizeServiceEntry(item, defaultSection))
        }
      })
      .filter((section) => section.title || section.items.length)
  },

  normalizeServiceEntry(item = {}, defaultSection = {}) {
    const fallback = findDefaultEntry(defaultSection, item)

    return {
      key: item.key || fallback.key || item.title,
      title: item.title || item.name || fallback.title || '',
      iconSrc: item.iconSrc || item.iconUrl || fallback.iconSrc || '',
      iconClass: item.iconClass || item.iconTone || fallback.iconClass || '',
      route: normalizeRoute(item.route || fallback.route || ''),
      badge: item.badgeText || item.badge || '',
      badgeClass: item.badgeClass || item.badgeTone || ''
    }
  },

  handleMenuTap(event) {
    const { route } = event.currentTarget.dataset

    if (!route) {
      toast.developing()
      return
    }

    wx.navigateTo({ url: route })
  },

  handleAssetAllTap() {
    wx.navigateTo({
      url: '/pages/profile/asset-center/manage/index'
    })
  },

  handleUpgradeTap() {
    const route = this.data.vipBanner && this.data.vipBanner.route

    wx.navigateTo({
      url: route || '/pages/profile/member/index'
    })
  }
})
