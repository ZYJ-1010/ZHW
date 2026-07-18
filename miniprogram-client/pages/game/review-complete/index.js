const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')
const reviewService = require('../../../services/review')
const { navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')

const EMPTY_REWARD = { show: false, points: '', title: '', desc: '', iconText: '' }
const DEFAULT_PLAY_OPTIONS = []

function normalizeRewardIcon(iconText) {
  const normalized = String(iconText || '').trim()
  return !normalized || normalized === '礼' ? '🎁' : normalized
}

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
    pageScrollTop: 0,
    reviewCount: 0,
    reviewGameId: 0,
    rewardedPoints: 0,
    reviewSummaryText: '',
    reward: EMPTY_REWARD,
    benefits: [],
    playOptions: DEFAULT_PLAY_OPTIONS,
    selectedPlayIntent: '',
    loadingConfig: false
  },

  onLoad(options = {}) {
    const reviewCount = Number(options.count || 0) || 0
    const reviewGameId = Number(options.gameId || 0) || 0
    const rewardedPoints = Math.max(0, Number(options.rewardPoints || 0) || 0)

    this.setData({
      reviewCount,
      reviewGameId,
      rewardedPoints,
      reviewSummaryText: reviewCount
        ? `本次已提交 ${reviewCount} 条评价${reviewGameId ? ` · 局ID ${reviewGameId}` : ''}`
        : ''
    })
    this.loadCompleteConfig()
  },

  async loadCompleteConfig() {
    this.setData({ loadingConfig: true })

    try {
      const config = await reviewService.getCompleteConfig({})
      const benefits = Array.isArray(config.benefits) ? config.benefits : []
      const playOptions = Array.isArray(config.playOptions) ? config.playOptions : DEFAULT_PLAY_OPTIONS
      const rewardConfig = config.reward && typeof config.reward === 'object' ? config.reward : EMPTY_REWARD
      const rewardedPoints = this.data.rewardedPoints
      const reward = rewardedPoints > 0
        ? {
            ...rewardConfig,
            show: true,
            points: `+${rewardedPoints}`,
            iconText: normalizeRewardIcon(rewardConfig.iconText),
            title: rewardConfig.title || '积分已到账',
            desc: rewardConfig.desc || '评价奖励已存入积分账户'
          }
        : EMPTY_REWARD

      this.setData({
        reward,
        benefits,
        playOptions,
        selectedPlayIntent: playOptions[0] ? playOptions[0].id : '',
        loadingConfig: false
      })
    } catch (error) {
      this.setData({
        reward: EMPTY_REWARD,
        benefits: [],
        playOptions: DEFAULT_PLAY_OPTIONS,
        selectedPlayIntent: '',
        loadingConfig: false
      })
    }
  },

  handlePageScroll(event) {
    this.lastScrollTop = event.detail.scrollTop || 0
  },

  handleShellNavTap(event) {
    const { key } = event.detail

    if (key === 'up') {
      this.scrollBy(-180)
      return
    }

    if (key === 'down') {
      this.scrollBy(180)
      return
    }

    navigateShellKey(key, {
      currentRoute: ROUTES.gameReviewComplete,
      routeMap: { comment: ROUTES.message }
    })
  },

  handleShellNavLongPress(event) {
    const { key } = event.detail

    if (key === 'up') {
      this.scrollBy(-520)
      return
    }

    if (key === 'down') {
      this.scrollBy(520)
    }
  },

  handleShellNavTouchEnd() {},

  scrollBy(distance) {
    const nextTop = Math.max(0, Math.round((this.lastScrollTop || 0) + distance))

    this.setData({
      pageScrollTop: nextTop
    })
  },

  async onPlayIntentTap(event) {
    const id = event.currentTarget.dataset.id

    if (!id) {
      return
    }

    this.setData({
      selectedPlayIntent: id
    })

    await this.submitPlayIntent(id)
    this.navigateByIntent(id)
  },

  async submitPlayIntent(id) {
    if (!this.data.reviewGameId) {
      return
    }

    try {
      await gameService.createRetrospective(this.data.reviewGameId, {
        content: `评价完成页选择：${this.intentTitle(id)}`,
        againIntent: this.intentValue(id)
      })
    } catch (error) {
      this.showInfo(error && error.message ? error.message : '再玩意愿记录失败')
    }
  },

  navigateByIntent(id) {
    const option = this.data.playOptions.find((item) => item.id === id) || {}

    if (option.route === 'play_again' || id === 'again') {
      this.replaceRoute(`/${ROUTES.gamePlayAgain}${this.data.reviewGameId ? `?gameId=${this.data.reviewGameId}` : ''}`)
      return
    }

    navigateShellRoute(ROUTES.gameHall)
  },

  intentValue(id) {
    const option = this.data.playOptions.find((item) => item.id === id)
    return option && option.intent ? option.intent : 'yes'
  },

  intentTitle(id) {
    const option = this.data.playOptions.find((item) => item.id === id)
    return option ? option.title : '再玩一局'
  },

  replaceRoute(route) {
    wx.redirectTo({
      url: route,
      fail: () => wx.reLaunch({ url: route })
    })
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
