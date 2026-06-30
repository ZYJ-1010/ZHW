const { ROUTES } = require('../../../../config/routes')
const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

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
  { key: 'active', text: '进行中' },
  { key: 'complete', text: '已完成' },
  { key: 'overdue', text: '超时' },
  { key: 'canceled', text: '已取消' }
]

const GAME_CARDS = []

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

function buildDisplayState(categoryKey, statusKey, cards = GAME_CARDS, statusTabs = STATUS_TABS) {
  return {
    categoryTabs: buildTabs(CATEGORY_TABS, categoryKey),
    statusTabs: buildTabs(statusTabs, statusKey),
    displayCards: getDisplayCards(cards, categoryKey, statusKey)
  }
}

const INITIAL_CATEGORY_KEY = 'joined'
const INITIAL_STATUS_KEY = 'all'

Page({
  data: {
    shellLayout: getShellLayoutStyles(),
    activeCategoryKey: INITIAL_CATEGORY_KEY,
    activeStatusKey: INITIAL_STATUS_KEY,
    cards: GAME_CARDS,
    ...buildDisplayState(INITIAL_CATEGORY_KEY, INITIAL_STATUS_KEY)
  },

  onLoad() {
    this.loadGames()
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

  async loadGames() {
    try {
      const data = await profileService.getProfileGames({
        category: this.data.activeCategoryKey,
        status: this.data.activeStatusKey
      })
      const cards = this.normalizeCards(data.list || data.records || data.items || data.games)
      const statusTabs = Array.isArray(data.statusTabs) && data.statusTabs.length ? data.statusTabs : STATUS_TABS

      this.setData({
        cards,
        ...buildDisplayState(this.data.activeCategoryKey, this.data.activeStatusKey, cards, statusTabs)
      })
    } catch (error) {
      toast.info(error.message || '我的局加载失败')
    }
  },

  normalizeCards(cards) {
    return Array.isArray(cards) ? cards : []
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
      ...buildDisplayState(key, this.data.activeStatusKey, this.data.cards)
    })
    this.loadGames()
  },

  onStatusTabTap(event) {
    const key = event.currentTarget.dataset.key

    this.setData({
      activeStatusKey: key,
      ...buildDisplayState(this.data.activeCategoryKey, key, this.data.cards)
    })
    this.loadGames()
  },

  onActionTap(event) {
    const action = event.currentTarget.dataset.action || '操作'

    toast.info(`${action}功能待接入`)
  }
})
