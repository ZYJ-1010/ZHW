const { ROUTES } = require('../../../../config/routes')
const toast = require('../../../../utils/toast')

const CONTENT_TOP_RPX = 160
const CONTENT_WIDTH_RPX = 750
const NAV_TITLE_HEIGHT_RPX = 50
const NAV_ICON_SIZE_RPX = 52
const FILTER_ICON_SIZE_RPX = 44
const DEFAULT_CAPSULE_BOTTOM_RPX = 142
const DEFAULT_FRAME_HEIGHT_RPX = 1620

const CATEGORY_TABS = [
  { key: 'joined', text: '我参与的' },
  { key: 'invited', text: '我受邀的' },
  { key: 'favorite', text: '我收藏的' }
]

const STATUS_TABS = [
  { key: 'all', text: '全部' },
  { key: 'active', text: '进行中(2)' },
  { key: 'complete', text: '已完成(5)' },
  { key: 'overdue', text: '超时(0)' },
  { key: 'canceled', text: '已取消(1)' }
]

const GAME_CARDS = [
  {
    id: 'ACT-20260320-001',
    category: 'joined',
    statusType: 'active',
    statusText: '进行中',
    ref: 'ACT-20260320-001',
    timeText: '3天前',
    avatar: 'ZH',
    title: '产品架构咨询',
    expertName: '张专家',
    guideName: '王引荐',
    amount: '¥800',
    fundStatus: '已托管',
    deliveryText: '预计交付：03-25 14:00',
    canExpand: true,
    actions: []
  },
  {
    id: 'ACT-20260315-002',
    category: 'joined',
    statusType: 'complete',
    statusText: '已完成',
    ref: 'ACT-20260315-002',
    timeText: '5天前',
    avatar: 'CH',
    title: 'UI设计服务',
    expertName: '陈设计师',
    guideName: '',
    amount: '¥600',
    fundStatus: '已完成',
    deliveryText: '',
    canExpand: false,
    actions: [
      { type: 'review', text: '评价', icon: '★' },
      { type: 'buy', text: '再次购买' }
    ]
  },
  {
    id: 'ACT-20260310-003',
    category: 'joined',
    statusType: 'canceled',
    statusText: '已取消',
    ref: 'ACT-20260310-003',
    timeText: '10天前',
    avatar: 'LI',
    title: '技术咨询服务',
    expertName: '刘工',
    guideName: '',
    reason: '时间冲突',
    amount: '¥500',
    fundStatus: '已退款',
    deliveryText: '',
    canExpand: false,
    actions: []
  },
  {
    id: 'ACT-20260308-004',
    category: 'invited',
    statusType: 'active',
    statusText: '待确认',
    ref: 'ACT-20260308-004',
    timeText: '12天前',
    avatar: 'WY',
    title: '品牌增长诊断',
    expertName: '吴顾问',
    guideName: '赵引荐',
    amount: '¥1200',
    fundStatus: '待托管',
    deliveryText: '等待你确认参与',
    canExpand: true,
    actions: []
  },
  {
    id: 'ACT-20260301-005',
    category: 'favorite',
    statusType: 'complete',
    statusText: '已收藏',
    ref: 'ACT-20260301-005',
    timeText: '20天前',
    avatar: 'ML',
    title: '商业模型梳理',
    expertName: '马老师',
    guideName: '',
    amount: '¥900',
    fundStatus: '可复购',
    deliveryText: '',
    canExpand: false,
    actions: [
      { type: 'buy', text: '再次购买' }
    ]
  }
]

function roundRpx(value) {
  return Math.round(value * 100) / 100
}

function getCapsuleBottomRpx() {
  try {
    if (typeof wx !== 'undefined' && wx.getSystemInfoSync && wx.getMenuButtonBoundingClientRect) {
      const menuButton = wx.getMenuButtonBoundingClientRect()
      const systemInfo = wx.getSystemInfoSync()

      if (menuButton && systemInfo && systemInfo.windowWidth) {
        return roundRpx((menuButton.top + menuButton.height) * 750 / systemInfo.windowWidth)
      }
    }
  } catch (error) {
    return DEFAULT_CAPSULE_BOTTOM_RPX
  }

  return DEFAULT_CAPSULE_BOTTOM_RPX
}

function getFrameHeightRpx() {
  try {
    if (typeof wx !== 'undefined' && wx.getSystemInfoSync) {
      const systemInfo = wx.getSystemInfoSync()

      if (systemInfo && systemInfo.windowWidth && systemInfo.windowHeight) {
        return roundRpx(systemInfo.windowHeight * 750 / systemInfo.windowWidth)
      }
    }
  } catch (error) {
    return DEFAULT_FRAME_HEIGHT_RPX
  }

  return DEFAULT_FRAME_HEIGHT_RPX
}

function getShellLayoutStyles() {
  const capsuleBottom = getCapsuleBottomRpx()
  const frameHeight = getFrameHeightRpx()
  const contentHeight = Math.max(0, roundRpx(frameHeight - CONTENT_TOP_RPX))
  const titleTop = Math.max(0, roundRpx(capsuleBottom - NAV_TITLE_HEIGHT_RPX))
  const backTop = Math.max(0, roundRpx(capsuleBottom - NAV_ICON_SIZE_RPX))
  const filterTop = Math.max(0, roundRpx(capsuleBottom - FILTER_ICON_SIZE_RPX - 2))

  return {
    frameStyle: `height: ${frameHeight}rpx; min-height: ${frameHeight}rpx;`,
    contentStyle: [
      'left: 0',
      `top: ${CONTENT_TOP_RPX}rpx`,
      `width: ${CONTENT_WIDTH_RPX}rpx`,
      `height: ${contentHeight}rpx`
    ].join('; '),
    titleStyle: `top: ${titleTop}rpx; height: ${NAV_TITLE_HEIGHT_RPX}rpx; line-height: ${NAV_TITLE_HEIGHT_RPX}rpx;`,
    backStyle: `top: ${backTop}rpx; width: ${NAV_ICON_SIZE_RPX}rpx; height: ${NAV_ICON_SIZE_RPX}rpx;`,
    filterStyle: `top: ${filterTop}rpx; width: ${FILTER_ICON_SIZE_RPX}rpx; height: ${FILTER_ICON_SIZE_RPX}rpx;`
  }
}

function buildTabs(tabs, activeKey) {
  return tabs.map((item) => ({
    ...item,
    active: item.key === activeKey
  }))
}

function getDisplayCards(cards, categoryKey, statusKey) {
  return cards.filter((item) => (
    item.category === categoryKey && (statusKey === 'all' || item.statusType === statusKey)
  ))
}

function buildDisplayState(categoryKey, statusKey) {
  return {
    categoryTabs: buildTabs(CATEGORY_TABS, categoryKey),
    statusTabs: buildTabs(STATUS_TABS, statusKey),
    displayCards: getDisplayCards(GAME_CARDS, categoryKey, statusKey)
  }
}

const INITIAL_CATEGORY_KEY = 'joined'
const INITIAL_STATUS_KEY = 'all'

Page({
  data: {
    shellLayout: getShellLayoutStyles(),
    activeCategoryKey: INITIAL_CATEGORY_KEY,
    activeStatusKey: INITIAL_STATUS_KEY,
    ...buildDisplayState(INITIAL_CATEGORY_KEY, INITIAL_STATUS_KEY)
  },

  onShow() {
    this.updateShellLayout()
  },

  onResize() {
    this.updateShellLayout()
  },

  updateShellLayout() {
    this.setData({
      shellLayout: getShellLayoutStyles()
    })
  },

  onBackTap() {
    const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    wx.redirectTo({
      url: `/${ROUTES.profile}`
    })
  },

  onFilterTap() {
    toast.info('筛选功能待接入')
  },

  onCategoryTabTap(event) {
    const key = event.currentTarget.dataset.key

    this.setData({
      activeCategoryKey: key,
      ...buildDisplayState(key, this.data.activeStatusKey)
    })
  },

  onStatusTabTap(event) {
    const key = event.currentTarget.dataset.key

    this.setData({
      activeStatusKey: key,
      ...buildDisplayState(this.data.activeCategoryKey, key)
    })
  },

  onActionTap(event) {
    const action = event.currentTarget.dataset.action || '操作'

    toast.info(`${action}功能待接入`)
  }
})
