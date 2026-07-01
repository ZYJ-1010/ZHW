const gameService = require('../../../services/game')

function normalizeBenefit(item = {}) {
  return {
    iconText: item.iconText || item.icon || '',
    theme: item.theme || '',
    title: item.title || item.name || '',
    desc: item.desc || item.description || ''
  }
}

function normalizePlayOption(item = {}) {
  return {
    id: item.id || item.key || '',
    theme: item.theme || '',
    title: item.title || item.name || '',
    desc: item.desc || item.description || ''
  }
}

function normalizeCompleteConfig(data = {}) {
  const success = data.success || data.successSection || {}
  const reward = data.rewardCard || data.reward || {}
  const playOptions = Array.isArray(data.playOptions || data.options)
    ? (data.playOptions || data.options).map(normalizePlayOption).filter((item) => item.id)
    : []

  return {
    onlineText: data.onlineText || '3999人在线',
    success: {
      title: success.title || data.successTitle || '',
      desc: success.desc || success.description || data.successDesc || ''
    },
    rewardCard: {
      iconText: reward.iconText || reward.icon || '',
      value: reward.value || reward.points || data.rewardValue || data.rewardText || '',
      title: reward.title || reward.label || data.rewardTitle || '',
      desc: reward.desc || reward.description || data.rewardDesc || ''
    },
    benefits: Array.isArray(data.benefits) ? data.benefits.map(normalizeBenefit).filter((item) => item.title || item.desc) : [],
    playOptions,
    selectedPlayIntent: data.selectedPlayIntent || data.defaultPlayIntent || playOptions[0] && playOptions[0].id || ''
  }
}

Page({
  data: {
    onlineText: '3999人在线',
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ],
    pageScrollTop: 0,
    success: {
      title: '',
      desc: ''
    },
    rewardCard: {
      iconText: '',
      value: '',
      title: '',
      desc: ''
    },
    benefits: [],
    playOptions: [],
    selectedPlayIntent: '',
    queryParams: {},
    loading: false,
    submitting: false
  },

  onLoad(options = {}) {
    this.setData({
      queryParams: options
    })
    this.loadCompleteConfig(options)
  },

  async loadCompleteConfig(options = {}) {
    this.setData({
      loading: true
    })

    try {
      const data = await gameService.getReviewCompleteConfig(options)

      this.setData({
        ...normalizeCompleteConfig(data),
        loading: false
      })
    } catch (error) {
      this.setData({
        ...normalizeCompleteConfig({}),
        loading: false
      })
      this.showInfo(error.message || '评价完成信息加载失败')
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

    this.showInfo('功能正在开发中')
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

    if (!id || this.data.submitting) {
      return
    }

    this.setData({
      selectedPlayIntent: id,
      submitting: true
    })

    try {
      await gameService.selectReviewCompleteIntent({
        ...this.data.queryParams,
        intentId: id
      })
    } catch (error) {
      this.showInfo(error.message || '后续意向提交失败')
    } finally {
      this.setData({
        submitting: false
      })
    }
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
