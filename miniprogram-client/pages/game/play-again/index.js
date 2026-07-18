const { ROUTES } = require('../../../config/routes')
const gameApi = require('../../../api/modules/game')
const { navigateShellRoute } = require('../../../utils/shell-nav')

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
        selectedOption: options[0] ? options[0].id : '',
        loading: false
      })
    }).catch(() => {
      this.setData({
        recommendOptions: EMPTY_RECOMMEND_OPTIONS,
        selectedOption: '',
        loading: false
      })
    })
  },

  normalizeRecommendOptions(items) {
    const options = Array.isArray(items)
      ? items.filter((item) => item && item.id && item.title)
      : []

    return options
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
    this.navigateByOption(id)
  },

  onCloseTap() {
    this.navigateBackOrHall()
  },

  handleShellNavTap(event) {
    const route = this.getShellRoute(event.detail && event.detail.key)

    if (!route) {
      this.navigateBackOrHall()
      return
    }

    navigateShellRoute(route, {
      currentRoute: ROUTES.gamePlayAgain,
      onSameRoute: () => this.navigateBackOrHall()
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
  },

  getShellRoute(key) {
    const routes = {
      home: ROUTES.playerHome || ROUTES.home,
      mine: ROUTES.profile,
      profile: ROUTES.profile,
      map: ROUTES.map,
      message: ROUTES.message,
      metaverse: ROUTES.metaverse
    }

    return routes[key]
  },

  navigateBackOrHall() {
    const url = `/${ROUTES.gameHall}`
    wx.reLaunch({ url })
  }
})
