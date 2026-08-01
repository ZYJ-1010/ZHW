const { ROUTES } = require('../../../config/routes')
const gameApi = require('../../../api/modules/game')
const { navigateShellRoute } = require('../../../utils/shell-nav')

const REQUIRED_OPENING_MODES = [
  { id: 'same-friends', theme: 'green', iconText: '👥', title: '同局开局', desc: '邀请上一局成员再次组局', route: 'confirm', order: 10 },
  { id: 'smart-match', theme: 'blue', iconText: '🤖', title: '系统自配', desc: '根据你的偏好推荐适配组局', route: 'system_recommend', order: 20 },
  { id: 'create-new', theme: 'pink', iconType: 'plus', title: '玩家创建', desc: '自行创建一场全新的局', route: 'create', order: 30 }
]
const EMPTY_RECOMMEND_OPTIONS = []

Page({
  data: {
    onlineText: '在线',
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ],
    panelTitle: '太棒了！你想怎么开启下一局？',
    panelSubtitle: '后台会根据当前局和配置返回可用方式',
    recommendOptions: EMPTY_RECOMMEND_OPTIONS,
    selectedOption: '',
    sourceGameId: '',
    loading: false
  },

  onLoad(options = {}) {
    this.setData({
      sourceGameId: options.sourceGameId || options.gameId || ''
    })
    this.loadReplayContext()
  },

  loadReplayContext() {
    this.setData({ loading: true })
    gameApi.getReplayConfirmContext({
      sourceGameId: this.data.sourceGameId
    }).then((res) => {
      const data = res && res.data ? res.data : {}
      const options = this.normalizeRecommendOptions(data.quickActions)

      this.setData({
        panelTitle: data.title || data.panelTitle || this.data.panelTitle,
        panelSubtitle: data.desc || data.panelSubtitle || this.data.panelSubtitle,
        recommendOptions: options,
        selectedOption: '',
        loading: false
      })
    }).catch(() => {
      this.setData({
        recommendOptions: this.normalizeRecommendOptions(),
        selectedOption: '',
        loading: false
      })
    })
  },

  normalizeRecommendOptions(items) {
    const configured = Array.isArray(items) ? items : []
    const byRoute = new Map()

    configured.forEach((item) => {
      const route = String(item && item.route || '').trim()
      const fallback = REQUIRED_OPENING_MODES.find((mode) => mode.route === route)
      if (!fallback || byRoute.has(route)) return
      byRoute.set(route, Object.assign({}, fallback, item, { route, id: item.id || fallback.id }))
    })

    return REQUIRED_OPENING_MODES.map((fallback) => byRoute.get(fallback.route) || Object.assign({}, fallback))
  },

  onOptionTap(event) {
    const id = event.currentTarget.dataset.id
    const option = this.data.recommendOptions.find((item) => item.id === id)

    if (!option) {
      return
    }

    this.setData({
      selectedOption: id
    })
  },

  handleShellNavTap(event) {
    this.showRequiredChoiceNotice()
  },

  onConfirmTap() {
    const id = this.data.selectedOption
    if (!id) {
      this.showRequiredChoiceNotice()
      return
    }
    this.navigateByOption(id)
  },

  showRequiredChoiceNotice() {
    wx.showToast({
      title: '请选择一种开局方式后继续',
      icon: 'none'
    })
  },

  navigateByOption(id) {
    const option = this.data.recommendOptions.find((item) => item.id === id) || {}
    const route = option.route || id

    if (route === 'confirm' || id === 'same-friends') {
      this.navigateToRoute(ROUTES.gameConfirm, {
        source: 'playAgain',
        sourceGameId: this.data.sourceGameId
      })
      return
    }

    if (route === 'system_recommend' || id === 'smart-match') {
      this.navigateToRoute(ROUTES.gameSystemRecommend, {
        source: 'playAgain',
        sourceGameId: this.data.sourceGameId
      })
      return
    }

    if (route === 'create' || id === 'create-new') {
      this.navigateToRoute(ROUTES.gameCreate, {
        source: 'playAgain',
        sourceGameId: this.data.sourceGameId
      })
    }
  },

  navigateToRoute(route, params = {}) {
    const url = this.buildUrl(route, params)

    // 再玩一局是结束后的单向流程，替换当前页，避免返回上一页重复选择。
    wx.redirectTo({
      url,
      fail: () => wx.reLaunch({ url })
    })
  },

  buildUrl(route, params = {}) {
    const query = Object.keys(params)
      .filter((key) => params[key])
      .map((key) => `${key}=${encodeURIComponent(params[key])}`)
      .join('&')

    return `/${route}${query ? `?${query}` : ''}`
  }
})
