const { ROUTES } = require('../../../../config/routes')
const toast = require('../../../../utils/toast')
const gameService = require('../../../../services/game')
const { navigateShellRoute } = require('../../../../utils/shell-nav')
const { getProfileWhiteShellLayoutStyles } = require('../../../../utils/adaptive-shell-layout')

const CONTENT_TOP_RPX = 160
const CONTENT_WIDTH_RPX = 750
const NAV_TITLE_HEIGHT_RPX = 50
const NAV_ICON_SIZE_RPX = 52
const FILTER_ICON_SIZE_RPX = 44
const DEFAULT_CAPSULE_BOTTOM_RPX = 142
const DEFAULT_FRAME_HEIGHT_RPX = 1620

const EMPTY_GAME_CARDS = []
const EMPTY_PAGE_CONFIG = {
  pageTitle: '',
  emptyText: '',
  detailMissing: '',
  actionMissing: '',
  categoryTabs: [],
  statusTabs: []
}

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
  const layout = getProfileWhiteShellLayoutStyles({
    contentMinTopRpx: CONTENT_TOP_RPX,
    titleHeightRpx: NAV_TITLE_HEIGHT_RPX,
    backSizeRpx: NAV_ICON_SIZE_RPX
  })
  const filterTop = roundRpx(
    layout.capsuleTopRpx + (layout.capsuleHeightRpx - FILTER_ICON_SIZE_RPX) / 2
  )

  return Object.assign({}, layout, {
    contentStyle: `${layout.contentStyle} left: 0; width: ${CONTENT_WIDTH_RPX}rpx;`,
    filterStyle: `top: ${filterTop}rpx; width: ${FILTER_ICON_SIZE_RPX}rpx; height: ${FILTER_ICON_SIZE_RPX}rpx;`
  })
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

const INITIAL_CATEGORY_KEY = 'joined'
const INITIAL_STATUS_KEY = 'all'

Page({
  data: {
    shellLayout: getShellLayoutStyles(),
    loaded: false,
    pageConfig: EMPTY_PAGE_CONFIG,
    activeCategoryKey: INITIAL_CATEGORY_KEY,
    activeStatusKey: INITIAL_STATUS_KEY,
    ...buildDisplayState(INITIAL_CATEGORY_KEY, INITIAL_STATUS_KEY)
  },

  onLoad() {
    this.loadMyGames()
  },

  onShow() {
    this.updateShellLayout()
    if (this.data.loaded) {
      this.loadMyGames()
    }
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

    navigateShellRoute(ROUTES.profile)
  },

  onFilterTap() {
    const nextKey = this.data.activeStatusKey === 'all' ? 'active' : 'all'

    this.setData({
      activeStatusKey: nextKey,
      ...buildDisplayState(this.data.activeCategoryKey, nextKey, this.data.cards || EMPTY_GAME_CARDS, this.data.pageConfig)
    })
  },

  onCategoryTabTap(event) {
    const key = event.currentTarget.dataset.key

    this.setData({
      activeCategoryKey: key,
      ...buildDisplayState(key, this.data.activeStatusKey, this.data.cards || EMPTY_GAME_CARDS, this.data.pageConfig)
    })
  },

  onStatusTabTap(event) {
    const key = event.currentTarget.dataset.key

    this.setData({
      activeStatusKey: key,
      ...buildDisplayState(this.data.activeCategoryKey, key, this.data.cards || EMPTY_GAME_CARDS, this.data.pageConfig)
    })
  },

  onCardTap(event) {
    const card = findCardById(this.data.cards, event.currentTarget.dataset.id)
    const route = buildGameDetailRoute(card && card.gameId)

    if (route) {
      navigateShellRoute(route)
      return
    }

    toast.info(this.data.pageConfig.detailMissing || '')
  },

  onActionTap(event) {
    const { id, type, route } = event.currentTarget.dataset
    const card = findCardById(this.data.cards, id)
    const targetRoute = route || buildActionRoute(type, card)

    if (targetRoute) {
      navigateShellRoute(targetRoute)
      return
    }

    toast.info(this.data.pageConfig.actionMissing || '')
  },

  async loadMyGames() {
    try {
      const [playerData, favoriteData] = await Promise.all([
        gameService.getPlayerGameManage(),
        gameService.getMyFavoriteGames()
      ])
      const cards = normalizeMyGameCards(playerData, null, favoriteData)
      const pageConfig = normalizeMyGamesPageConfig(playerData.pageConfig || favoriteData.pageConfig)

      this.setData({
        loaded: true,
        pageConfig,
        cards,
        ...buildDisplayState(this.data.activeCategoryKey, this.data.activeStatusKey, cards, pageConfig)
      })
    } catch (error) {
      console.warn('[profile-my-games] load failed', error)
      this.setData({
        loaded: true,
        cards: EMPTY_GAME_CARDS,
        ...buildDisplayState(this.data.activeCategoryKey, this.data.activeStatusKey, EMPTY_GAME_CARDS, this.data.pageConfig)
      })
    }
  }
})

function buildDisplayState(categoryKey, statusKey, cards = EMPTY_GAME_CARDS, pageConfig = EMPTY_PAGE_CONFIG) {
  return {
    categoryTabs: buildTabs(buildCategoryTabs(cards, pageConfig), categoryKey),
    statusTabs: buildTabs(buildStatusTabs(cards, categoryKey, pageConfig), statusKey),
    displayCards: getDisplayCards(cards, categoryKey, statusKey)
  }
}

function buildCategoryTabs(cards, pageConfig = EMPTY_PAGE_CONFIG) {
  return pageConfig.categoryTabs.map((tab) => {
    const count = cards.filter((item) => item.category === tab.key).length

    return {
      ...tab,
      text: count > 0 ? `${tab.text}(${count})` : tab.text
    }
  })
}

function buildStatusTabs(cards, categoryKey, pageConfig = EMPTY_PAGE_CONFIG) {
  return pageConfig.statusTabs.map((tab) => {
    if (tab.key === 'all') {
      return tab
    }

    const count = cards.filter((item) => item.category === categoryKey && item.statusType === tab.key).length
    const baseText = tab.text.replace(/\(.+\)/, '')

    return {
      ...tab,
      text: `${baseText}(${count})`
    }
  })
}

function normalizeMyGamesPageConfig(config = {}) {
  return {
    ...EMPTY_PAGE_CONFIG,
    ...config,
    categoryTabs: Array.isArray(config.categoryTabs) ? config.categoryTabs : EMPTY_PAGE_CONFIG.categoryTabs,
    statusTabs: Array.isArray(config.statusTabs) ? config.statusTabs : EMPTY_PAGE_CONFIG.statusTabs
  }
}

function normalizeMyGameCards(playerData, managedData, favoriteData) {
  const playerOrders = normalizeOrders(playerData).map((item, index) => normalizeGameCard(item, 'joined', index))
  const joinedGameIds = new Set(playerOrders.map((item) => String(item.gameId || '')).filter(Boolean))
  const managedOrders = normalizeOrders(managedData)
    .map((item, index) => normalizeGameCard(item, 'created', index))
    .filter((item) => !item.gameId || !joinedGameIds.has(String(item.gameId)))
  const favoriteOrders = normalizeOrders(favoriteData).map((item, index) => normalizeGameCard(item, 'favorite', index))

  return playerOrders.concat(managedOrders, favoriteOrders)
}

function normalizeOrders(data) {
  if (!data) {
    return []
  }

  if (Array.isArray(data.orders)) {
    return data.orders
  }

  if (Array.isArray(data.items)) {
    return data.items
  }

  return []
}

function normalizeGameCard(raw = {}, category, index) {
  const source = raw.game || raw
  const statusType = normalizeStatusType(raw.statusType || raw.status || raw.statusKey || source.status)
  const name = raw.expertName || raw.playerName || raw.name || (raw.expert && raw.expert.name) || (raw.player && raw.player.name) || '用户'
  const guideName = raw.guideName || (raw.guide && raw.guide.name) || ''
  const amountText = raw.amountText || raw.serviceAmountText || raw.fundAmountText || formatAmount(raw.amount || raw.fundAmount || raw.amountCent)
  const title = raw.serviceTitle || raw.title || raw.gameTitle || source.title || '未命名组局'
  const gameId = raw.gameId || source.gameId || source.id || ''
  const canReview = Boolean(raw.canReview || raw.canReviewBoth)

  return {
    id: raw.id || raw.serviceOrderId || raw.ref || (gameId ? `${category}-${gameId}` : `my-game-${category}-${index}`),
    gameId,
    category,
    statusType,
    statusText: raw.statusText || statusTextByType(statusType),
    ref: raw.ref || raw.serviceOrderId || raw.orderNo || raw.id || (gameId ? `GAME-${gameId}` : ''),
    timeText: formatTimeText(raw.startedAt || raw.createdAt || raw.timeText || source.createdAt),
    avatar: raw.avatar || raw.avatarText || getAvatarText(name),
    title,
    expertName: raw.expertName || (category === 'joined' ? name : ''),
    guideName,
    amount: amountText,
    fundStatus: raw.fundStatus || raw.resultText || statusTextByType(statusType),
    deliveryText: raw.deliveryText || buildDeliveryText(raw),
    canExpand: statusType === 'active',
    reason: raw.reason || raw.cancelReason || '',
    actions: buildCardActions(raw, gameId, canReview)
  }
}

function normalizeStatusType(status) {
  switch (String(status || '').toLowerCase()) {
    case 'complete':
    case 'completed':
    case 'fulfilled':
    case 'pending_confirm':
    case 'pending_review':
      return 'complete'
    case 'canceled':
    case 'cancelled':
      return 'canceled'
    case 'overdue':
      return 'overdue'
    default:
      return 'active'
  }
}

function statusTextByType(statusType) {
  switch (statusType) {
    case 'complete':
      return '已完成'
    case 'canceled':
      return '已取消'
    case 'overdue':
      return '超时'
    default:
      return '进行中'
  }
}

function buildCardActions(raw, gameId, canReview) {
  const actions = []

  actions.push({
    type: 'detail',
    text: '详情',
    route: buildGameDetailRoute(gameId)
  })

  if (canReview) {
    actions.push({
      type: 'review',
      text: raw.reviewActionText || '评价',
      icon: '★',
      route: gameId ? `/pages/game/review/index?gameId=${encodeURIComponent(gameId)}` : ''
    })
  }

  if (raw.statusType === 'complete' || raw.status === 'completed') {
    actions.push({
      type: 'buy',
      text: '再次购买',
      route: gameId ? `/pages/game/detail/index?gameId=${encodeURIComponent(gameId)}` : '/pages/game/hall/index'
    })
  }

  return actions
}

function findCardById(cards = [], id = '') {
  return cards.find((item) => String(item.id) === String(id)) || null
}

function buildGameDetailRoute(gameId) {
  return gameId ? `/${ROUTES.gameDetail}?id=${encodeURIComponent(gameId)}` : ''
}

function buildActionRoute(type, card) {
  if (!card) {
    return ''
  }

  if (type === 'review' && card.gameId) {
    return `/${ROUTES.gameReview}?gameId=${encodeURIComponent(card.gameId)}`
  }

  if ((type === 'buy' || type === 'detail') && card.gameId) {
    return buildGameDetailRoute(card.gameId)
  }

  if (type === 'contact' && card.gameId) {
    return `/${ROUTES.imRoom}?gameId=${encodeURIComponent(card.gameId)}&prefill=${encodeURIComponent('你好，我想确认一下本次组局进度。')}`
  }

  return ''
}

function buildDeliveryText(raw) {
  if (raw.expectedDeliveryAt) {
    return `预计交付：${formatDateText(raw.expectedDeliveryAt)}`
  }

  return raw.noticeText || ''
}

function formatAmount(value) {
  if (typeof value === 'number') {
    if (value > 10000) {
      return `¥${Math.round(value / 100)}`
    }
    return `¥${value}`
  }

  return value || ''
}

function formatTimeText(value) {
  if (!value || typeof value !== 'string') {
    return value || ''
  }

  return formatDateText(value)
}

function formatDateText(value) {
  const match = String(value).match(/(\d{4})-(\d{2})-(\d{2})(?:T|\s)(\d{2}):(\d{2})/)

  if (!match) {
    return value
  }

  return `${match[2]}-${match[3]} ${match[4]}:${match[5]}`
}

function getAvatarText(name) {
  const text = String(name || '').trim()
  if (!text) {
    return '用'
  }

  return Array.from(text).slice(0, 2).join('')
}
