const toast = require('../../../../utils/toast')
const FA_BASE = '/pages/profile/asset-center/manage/assets/fa'

Page({
  data: {
    overview: {
      label: '总资产（元）',
      value: '¥16,580.00'
    },
    assetStats: [
      { key: 'totalDealAmount', label: '总成交额', value: '¥12,580', tone: '' },
      { key: 'withdrawable', label: '可提现', value: '¥3,200', tone: 'green' },
      { key: 'pendingSettlement', label: '待结算', value: '¥800', tone: 'yellow' }
    ],
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
        value: '已绑定2张',
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
    recentOrders: [
      {
        id: 'ORD-20260320-001',
        title: '产品架构咨询',
        status: '进行中',
        statusTone: 'blue',
        time: '2026-03-20 14:30',
        amount: '¥800.00'
      },
      {
        id: 'ORD-20260315-002',
        title: 'UI设计服务',
        status: '已完成',
        statusTone: 'green',
        time: '2026-03-15 09:15',
        amount: '¥600.00'
      }
    ],
    faqLinks: [
      { key: 'withdrawArrival', label: '提现多久到账？' },
      { key: 'bindBankCard', label: '如何绑定银行卡？' }
    ]
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
