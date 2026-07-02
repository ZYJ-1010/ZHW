const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')
const { navigateShellRoute } = require('../../../../utils/shell-nav')

Page({
  data: {
    loaded: false,
    loadError: '',
    overview: {},
    assetStats: [],
    quickActions: [],
    menuItems: [],
    orderStatuses: [],
    recentOrders: [],
    balanceRecords: [],
    bankCards: null,
    faqLinks: []
  },

  onLoad() {
    this.loadAssets()
  },

  onShow() {
    if (this.data.loaded) {
      this.loadAssets()
    }
  },

  async loadAssets() {
    try {
      const data = await profileService.getProfileAssets()

      this.setData(Object.assign({}, normalizeAssetHome(data), {
        loadError: ''
      }))
    } catch (error) {
      this.setData({
        loaded: true,
        loadError: error.message || '资产数据加载失败',
        quickActions: [],
        menuItems: [],
        orderStatuses: [],
        recentOrders: [],
        balanceRecords: [],
        faqLinks: []
      })
      console.warn('[profile-assets] load failed', error)
    }
  },

  handleMenuTap(event) {
    const { route, key } = event.currentTarget.dataset

    if (!route) {
      if (key === 'balance') {
        this.showBalanceRecords()
        return
      }
      if (key === 'bankCards') {
        this.showBankCards()
        return
      }
      toast.info('该资产入口暂未开放')
      return
    }

    navigateShellRoute(route)
  },

  handleActionTap(event) {
    const { enabled, reason, key } = event.currentTarget.dataset

    if (enabled === false || enabled === 'false') {
      toast.info(reason || '该功能暂未开放')
      return
    }

    if (key === 'withdraw') {
      toast.info(reason || '提现需后台财务审核后处理')
      return
    }

    toast.info(reason || '一期未开放真实支付充值')
  },

  handleStatusTap(event) {
    const { route, key } = event.currentTarget.dataset

    const targetRoute = route || (key ? `/pages/profile/asset-center/orders/index?status=${encodeURIComponent(key)}` : '/pages/profile/asset-center/orders/index')

    navigateShellRoute(targetRoute)
  },

  handleOrderTap(event) {
    const { route, orderId, id } = event.currentTarget.dataset
    const targetId = orderId || id
    const targetRoute = route || (targetId ? `/pages/profile/asset-center/orders/index?orderId=${encodeURIComponent(targetId)}` : '/pages/profile/asset-center/orders/index')

    navigateShellRoute(targetRoute)
  },

  handleFAQTap(event) {
    const key = event.currentTarget.dataset.key
    const item = this.data.faqLinks.find((faq) => faq.key === key)

    wx.showModal({
      title: item && item.label || '常见问题',
      content: item && item.answer || '暂无说明',
      showCancel: false
    })
  },

  showBalanceRecords() {
    const records = this.data.balanceRecords || []
    const menu = this.data.menuItems.find((item) => item.key === 'balance') || {}

    if (!records.length) {
      toast.info('暂无余额流水')
      return
    }

    wx.showModal({
      title: menu.title || '余额明细',
      content: records.map((item) => `${item.title} ${item.amount} ${item.status || ''}`).join('\n'),
      showCancel: false
    })
  },

  showBankCards() {
    const bankCards = this.data.bankCards || {}
    const items = Array.isArray(bankCards.items) ? bankCards.items : []
    const menu = this.data.menuItems.find((item) => item.key === 'bankCards') || {}

    wx.showModal({
      title: menu.title || '银行卡',
      content: items.length
        ? items.map((item) => `${item.bankName || '银行卡'} ${item.cardNo || ''}`).join('\n')
        : (bankCards.summaryText || '未绑定'),
      showCancel: false
    })
  }
})

function normalizeAssetHome(data = {}) {
  const patch = { loaded: true }

  ;[
    'overview',
    'assetStats',
    'quickActions',
    'menuItems',
    'orderStatuses',
    'recentOrders',
    'faqLinks',
    'balanceRecords',
    'bankCards'
  ].forEach((key) => {
    if (data[key]) {
      patch[key] = data[key]
    }
  })

  return patch
}
