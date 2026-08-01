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
        faqLinks: []
      })
      console.warn('[profile-assets] load failed', error)
    }
  },

  handleMenuTap(event) {
    const { route, key } = event.currentTarget.dataset

    if (!route) {
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

    toast.info(reason || '该功能暂未开放')
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
    'faqLinks'
  ].forEach((key) => {
    if (data[key]) {
      patch[key] = data[key]
    }
  })

  return patch
}
