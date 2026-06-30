const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')
const FA_BASE = '/pages/profile/asset-center/manage/assets/fa'

Page({
  data: {
    overview: {
      label: '总资产（元）',
      value: ''
    },
    assetStats: [],
    quickActions: [
      { key: 'withdraw', label: '提现', tone: 'green', iconSrc: `${FA_BASE}/download.svg` },
      { key: 'recharge', label: '充值', tone: 'blue', iconSrc: `${FA_BASE}/plus.svg` }
    ],
    menuItems: [
      {
        key: 'balance',
        title: '余额明细',
        desc: '收入支出记录',
        iconSrc: `${FA_BASE}/list-ul.svg`,
        tone: 'blue'
      },
      {
        key: 'bankCards',
        title: '银行卡',
        desc: '管理收款账户',
        value: '',
        iconSrc: `${FA_BASE}/credit-card.svg`,
        tone: 'green'
      },
      {
        key: 'orders',
        title: '我的订单',
        desc: '查看全部订单',
        iconSrc: `${FA_BASE}/bag-shopping.svg`,
        tone: 'purple',
        route: '/pages/profile/asset-center/orders/index'
      }
    ],
    orderStatuses: [
      { key: 'pendingPay', label: '待付款', iconSrc: `${FA_BASE}/hourglass-half.svg`, tone: 'blue' },
      { key: 'processing', label: '进行中', iconSrc: `${FA_BASE}/spinner.svg`, tone: 'orange' },
      { key: 'completed', label: '已完成', iconSrc: `${FA_BASE}/check.svg`, tone: 'green' },
      { key: 'refund', label: '退款/售后', iconSrc: `${FA_BASE}/rotate-left.svg`, tone: 'red' },
      { key: 'review', label: '待评价', iconSrc: `${FA_BASE}/star.svg`, tone: 'gray' }
    ],
    recentOrders: [],
    faqLinks: [
      { key: 'withdrawArrival', label: '提现多久到账？' },
      { key: 'bindBankCard', label: '如何绑定银行卡？' }
    ]
  },

  onLoad() {
    this.loadAssets()
  },

  async loadAssets() {
    try {
      const data = await profileService.getProfileAssets()

      this.setData({
        overview: this.normalizeOverview(data.overview || data.summary || {}),
        assetStats: this.normalizeList(data.assetStats || data.stats || data.summaryItems),
        menuItems: this.mergeMenuItems(data.menuItems || data.menus),
        orderStatuses: this.normalizeOrderStatuses(data.orderStatuses || data.statuses),
        recentOrders: this.normalizeList(data.recentOrders || data.orders),
        faqLinks: this.normalizeList(data.faqLinks || data.faqs)
      })
    } catch (error) {
      toast.info(error.message || '资产信息加载失败')
    }
  },

  normalizeOverview(source = {}) {
    return {
      label: source.label || source.title || '总资产（元）',
      value: source.value || source.amountText || source.totalAssetText || ''
    }
  },

  normalizeList(list) {
    return Array.isArray(list) ? list : []
  },

  mergeMenuItems(list) {
    if (!Array.isArray(list) || list.length === 0) {
      return this.data.menuItems
    }

    return this.data.menuItems.map((item) => {
      const remote = list.find((entry) => entry.key === item.key || entry.title === item.title) || {}

      return Object.assign({}, item, remote)
    })
  },

  normalizeOrderStatuses(list) {
    if (!Array.isArray(list) || list.length === 0) {
      return this.data.orderStatuses
    }

    return list
  },

  handleMenuTap(event) {
    const { route } = event.currentTarget.dataset

    if (!route) {
      toast.developing()
      return
    }

    wx.navigateTo({ url: route })
  },

  handleDeveloping() {
    toast.developing()
  }
})
