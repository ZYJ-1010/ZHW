const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')
const { getSurnameInitials } = require('../../../utils/avatar')

const DEFAULT_CATEGORIES = [
  { key: 'all', name: '全部' },
  { key: 'product', name: '产品架构' },
  { key: 'tech', name: '技术咨询' },
  { key: 'operation', name: '运营策略' }
]

function filterExperts(experts, category) {
  if (category === 'all') {
    return experts
  }

  return experts.filter((item) => item.category === category)
}

function getSelectedExperts(experts) {
  return experts.filter((item) => item.selected)
}

function getMatchTone(match) {
  const value = Number(match || 0)

  if (value >= 90) {
    return 'high'
  }

  if (value >= 80) {
    return 'medium'
  }

  return 'low'
}

function normalizeExpert(expert = {}, selectedIds = []) {
  const name = expert.name || expert.nickname || '行家'
  const id = String(expert.id || expert.userId || expert.expertId || name)

  return {
    id,
    name,
    role: expert.role || expert.desc || expert.title || '',
    avatarText: getSurnameInitials(name, expert.avatarText || expert.initials || 'EX'),
    avatarClass: expert.avatarClass || 'purple',
    rating: expert.rating || expert.score || '5.0',
    stars: expert.stars || '★★★★★',
    reviewCount: Number(expert.reviewCount || expert.reviews || 0),
    match: Number(expert.match || expert.matchPercent || 0),
    matchTone: expert.matchTone || getMatchTone(expert.match || expert.matchPercent),
    price: Number(expert.price || expert.pricePerHour || 0),
    category: expert.category || 'product',
    tags: Array.isArray(expert.tags) ? expert.tags : [],
    selected: Boolean(expert.selected || selectedIds.includes(id))
  }
}

function normalizeRecommendationContext(context = {}) {
  const categories = Array.isArray(context.categories) && context.categories.length
    ? context.categories
    : DEFAULT_CATEGORIES
  const selectedIds = Array.isArray(context.defaultSelectedExpertIds)
    ? context.defaultSelectedExpertIds.map(String)
    : []
  const experts = (Array.isArray(context.experts) ? context.experts : [])
    .map((item) => normalizeExpert(item, selectedIds))

  return {
    recommendationId: context.recommendationId || '',
    title: context.title || '系统推荐适配局',
    desc: context.desc || '基于你的偏好，已找到高匹配度行家',
    categories,
    experts
  }
}

Page({
  data: {
    recommendationId: '',
    recommendationTitle: '系统推荐适配局',
    recommendationDesc: '基于你的偏好，已找到高匹配度行家',
    sourceGameId: '',
    serviceOrderId: '',
    categories: DEFAULT_CATEGORIES,
    activeCategory: 'all',
    experts: [],
    visibleExperts: [],
    selectedExperts: [],
    loading: false
  },

  onLoad(options = {}) {
    this.setData({
      sourceGameId: options.sourceGameId || options.gameId || '',
      serviceOrderId: options.serviceOrderId || ''
    })
    this.loadRecommendations(options)
  },

  async loadRecommendations(options = {}) {
    this.setData({
      loading: true
    })

    try {
      const context = await gameService.getSystemRecommendations({
        sourceGameId: options.sourceGameId || options.gameId || '',
        serviceOrderId: options.serviceOrderId || '',
        recommendationId: options.recommendationId || '',
        category: this.data.activeCategory
      })

      this.applyRecommendationContext(context)
    } catch (error) {
      wx.showToast({
        title: error.message || '推荐数据加载失败',
        icon: 'none'
      })
      this.setData({
        loading: false
      })
    }
  },

  applyRecommendationContext(context) {
    const normalized = normalizeRecommendationContext(context)
    const activeCategory = this.data.activeCategory || 'all'

    this.setData({
      recommendationId: normalized.recommendationId,
      recommendationTitle: normalized.title,
      recommendationDesc: normalized.desc,
      categories: normalized.categories,
      experts: normalized.experts,
      visibleExperts: filterExperts(normalized.experts, activeCategory),
      selectedExperts: getSelectedExperts(normalized.experts),
      loading: false
    })
  },

  onCategoryTap(event) {
    const key = event.currentTarget.dataset.key || 'all'
    const visibleExperts = filterExperts(this.data.experts, key)

    this.setData({
      activeCategory: key,
      visibleExperts
    })
  },

  onExpertTap(event) {
    const id = event.currentTarget.dataset.id

    if (!id) {
      return
    }

    const experts = this.data.experts.map((item) => item.id === id
      ? {
        ...item,
        selected: !item.selected
      }
      : item)

    this.setData({
      experts,
      visibleExperts: filterExperts(experts, this.data.activeCategory),
      selectedExperts: getSelectedExperts(experts)
    })
  },

  onBackTap() {
    this.navigateBackOrHall()
  },

  onCancelTap() {
    this.navigateBackOrHall()
  },

  onConfirmTap() {
    if (!this.data.selectedExperts.length) {
      wx.showToast({
        title: '请选择至少一位行家',
        icon: 'none'
      })
      return
    }

    wx.showToast({
      title: '正在进入组局',
      icon: 'none',
      duration: 800
    })

    wx.navigateTo({
      url: this.buildCreateGameUrl(),
      fail: () => {
        wx.redirectTo({
          url: this.buildCreateGameUrl()
        })
      }
    })
  },

  buildCreateGameUrl() {
    const selectedExpertIds = this.data.selectedExperts.map((item) => item.id).join(',')
    const query = [
      ['source', 'systemRecommend'],
      ['recommendationId', this.data.recommendationId],
      ['selectedExpertIds', selectedExpertIds],
      ['sourceGameId', this.data.sourceGameId],
      ['serviceOrderId', this.data.serviceOrderId]
    ]
      .filter((item) => item[1])
      .map((item) => `${item[0]}=${encodeURIComponent(item[1])}`)
      .join('&')

    return `/${ROUTES.gameCreate}${query ? `?${query}` : ''}`
  },

  navigateBackOrHall() {
    const pages = getCurrentPages()

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    wx.redirectTo({
      url: `/${ROUTES.gameHall}`
    })
  }
})
