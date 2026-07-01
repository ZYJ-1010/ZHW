const gameService = require('../../../services/game')

function normalizeOption(option = {}) {
  return {
    id: option.id || option.optionId || option.key || '',
    theme: option.theme || '',
    iconText: option.iconText || option.emoji || '',
    iconType: option.iconType || '',
    title: option.title || option.name || '',
    desc: option.desc || option.description || '',
    route: option.route || option.path || '',
    message: option.message || option.toastText || ''
  }
}

function normalizePlayAgainData(data = {}) {
  const recommendOptions = Array.isArray(data.recommendOptions || data.options)
    ? (data.recommendOptions || data.options).map(normalizeOption).filter((item) => item.id)
    : []

  return {
    onlineText: data.onlineText || '3999人在线',
    recommendOptions,
    selectedOption: data.selectedOption || data.defaultOptionId || recommendOptions[0] && recommendOptions[0].id || ''
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
    recommendOptions: [],
    selectedOption: '',
    loading: false,
    submitting: false,
    queryParams: {}
  },

  onLoad(options = {}) {
    this.setData({
      queryParams: options
    })
    this.loadPlayAgainOptions(options)
  },

  async loadPlayAgainOptions(options = {}) {
    this.setData({
      loading: true
    })

    try {
      const data = await gameService.getPlayAgainOptions(options)

      this.setData({
        ...normalizePlayAgainData(data),
        loading: false
      })
    } catch (error) {
      this.setData({
        ...normalizePlayAgainData({}),
        loading: false
      })
      this.showInfo(error.message || '再玩一局推荐加载失败')
    }
  },

  async onOptionTap(event) {
    const id = event.currentTarget.dataset.id
    const option = this.data.recommendOptions.find((item) => item.id === id)

    if (!option || this.data.submitting) {
      return
    }

    this.setData({
      selectedOption: id,
      submitting: true
    })

    try {
      const result = await gameService.selectPlayAgainOption({
        ...this.data.queryParams,
        optionId: id
      })
      const route = result && (result.route || result.path) || option.route

      if (route) {
        wx.navigateTo({
          url: route
        })
        return
      }

      this.showInfo(result && (result.message || result.toastText) || option.message || '已提交')
    } catch (error) {
      this.showInfo(error.message || '推荐方式提交失败')
    } finally {
      this.setData({
        submitting: false
      })
    }
  },

  onCloseTap() {
    const pages = getCurrentPages()

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    this.showInfo('已关闭')
  },

  handleShellNavTap() {
    this.showInfo('功能正在开发中')
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
