const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')
const { navigateShellRoute } = require('../../../utils/shell-nav')

const WHITE_CONTENT_LEFT_RPX = 2
const WHITE_CONTENT_TOP_RPX = 160
const WHITE_CONTENT_WIDTH_RPX = 750
const WHITE_DESIGN_FRAME_HEIGHT_PT = 810
const WHITE_DESIGN_BOTTOM_HEIGHT_PT = 78
const WHITE_NAV_TITLE_HEIGHT_RPX = 50
const WHITE_BACK_BUTTON_SIZE_RPX = 40
const WHITE_DEFAULT_CAPSULE_BOTTOM_RPX = 142
const WHITE_DEFAULT_FRAME_HEIGHT_RPX = WHITE_DESIGN_FRAME_HEIGHT_PT * 2

function roundRpx(value) {
  return Math.round(value * 100) / 100
}

function getMenuCapsuleBottomRpx() {
  try {
    if (typeof wx !== 'undefined' && wx.getSystemInfoSync && wx.getMenuButtonBoundingClientRect) {
      const menuButton = wx.getMenuButtonBoundingClientRect()
      const systemInfo = wx.getSystemInfoSync()

      if (menuButton && systemInfo && systemInfo.windowWidth) {
        return roundRpx((menuButton.top + menuButton.height) * 750 / systemInfo.windowWidth)
      }
    }
  } catch (error) {
    return WHITE_DEFAULT_CAPSULE_BOTTOM_RPX
  }

  return WHITE_DEFAULT_CAPSULE_BOTTOM_RPX
}

function getWhiteShellLayoutStyles() {
  const capsuleBottom = getMenuCapsuleBottomRpx()
  const titleTop = Math.max(0, roundRpx(capsuleBottom - WHITE_NAV_TITLE_HEIGHT_RPX))
  const backTop = Math.max(0, roundRpx(capsuleBottom - WHITE_BACK_BUTTON_SIZE_RPX))
  let frameHeight = WHITE_DEFAULT_FRAME_HEIGHT_RPX

  try {
    if (typeof wx !== 'undefined' && wx.getSystemInfoSync) {
      const systemInfo = wx.getSystemInfoSync()

      if (systemInfo && systemInfo.windowWidth && systemInfo.windowHeight) {
        frameHeight = roundRpx(systemInfo.windowHeight * 750 / systemInfo.windowWidth)
      }
    }
  } catch (error) {
    frameHeight = WHITE_DEFAULT_FRAME_HEIGHT_RPX
  }

  const bottomHeight = roundRpx(frameHeight * WHITE_DESIGN_BOTTOM_HEIGHT_PT / WHITE_DESIGN_FRAME_HEIGHT_PT)
  const bottomTop = Math.max(WHITE_CONTENT_TOP_RPX, roundRpx(frameHeight - bottomHeight))
  const contentHeight = Math.max(0, roundRpx(bottomTop - WHITE_CONTENT_TOP_RPX))

  return {
    frameStyle: `height: ${frameHeight}rpx; min-height: ${frameHeight}rpx;`,
    contentStyle: [
      `left: ${WHITE_CONTENT_LEFT_RPX}rpx`,
      `top: ${WHITE_CONTENT_TOP_RPX}rpx`,
      `width: ${WHITE_CONTENT_WIDTH_RPX}rpx`,
      `height: ${contentHeight}rpx`
    ].join('; '),
    bottomStyle: `top: ${bottomTop}rpx; height: ${bottomHeight}rpx;`,
    titleStyle: `top: ${titleTop}rpx; height: ${WHITE_NAV_TITLE_HEIGHT_RPX}rpx; line-height: ${WHITE_NAV_TITLE_HEIGHT_RPX}rpx;`,
    backStyle: `top: ${backTop}rpx; width: ${WHITE_BACK_BUTTON_SIZE_RPX}rpx; height: ${WHITE_BACK_BUTTON_SIZE_RPX}rpx;`
  }
}

function normalizeMember(item = {}, index = 0) {
  const userId = item.userId || item.id || item.memberId || ''
  const role = item.roleLabel || item.role || (item.isCreator ? '发起人' : '玩家')
  const roleKey = String(item.role || '').toLowerCase()
  const roleClass = roleKey.indexOf('expert') !== -1
    ? 'expert'
    : roleKey.indexOf('guide') !== -1 || item.isCreator
      ? 'guide'
      : 'player'
  const avatarText = item.avatarText || item.avatar || String(userId || index + 1).slice(-2)

  return {
    id: userId || `member-${index + 1}`,
    userId,
    name: item.name || item.nickname || `成员${index + 1}`,
    avatarSrc: item.avatarSrc || item.avatarUrl || 'https://static.haowan.net.cn/miniprogram/pages/home/player/assets/ranking-avatar-01.png',
    avatarText,
    role,
    roleClass,
    position: item.position || item.roleText || (item.isCreator ? '组局发起人' : '局内成员'),
    topic: item.topic || (item.confirmed ? '已确认服务完成' : '参与本次组局'),
    primaryTag: item.primaryTag || (item.isCurrentUser ? '当前用户' : '组局成员'),
    tags: Array.isArray(item.tags) ? item.tags : [],
    location: item.location || item.cityName || '同城组局',
    distance: item.distance || item.distanceText || ''
  }
}

Page({
  data: {
    gameId: '',
    whiteShellLayout: getWhiteShellLayoutStyles(),
    participants: []
  },

  onLoad(options = {}) {
    const gameId = options.gameId || options.id || ''

    this.setData({
      gameId,
      whiteShellLayout: getWhiteShellLayoutStyles(),
      participants: []
    })
    this.loadParticipants(gameId)
  },

  async loadParticipants(gameId) {
    if (!gameId) {
      return
    }

    try {
      const data = await gameService.getGameMembers(gameId)
      const items = Array.isArray(data.items) ? data.items : []

      this.setData({
        participants: items.map(normalizeMember)
      })
    } catch (error) {
      wx.showToast({
        title: error.message || '参与者加载失败',
        icon: 'none'
      })
    }
  },

  onResize() {
    this.setData({
      whiteShellLayout: getWhiteShellLayoutStyles()
    })
  },

  handleBack() {
    const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    navigateShellRoute(`/${ROUTES.gameDetail}${this.data.gameId ? `?id=${encodeURIComponent(this.data.gameId)}` : ''}`)
  },

  onParticipantTap(event) {
    const participant = event.detail && event.detail.participant
    const memberId = participant && (participant.userId || participant.id)

    if (memberId) {
      navigateShellRoute(`/pages/profile/service-center/invite/member-detail/index?id=${encodeURIComponent(memberId)}`)
      return
    }

    navigateShellRoute(ROUTES.profile)
  },

  onShareAppMessage() {
    return {
      title: '全部参与者',
      path: `/${ROUTES.gameParticipants}${this.data.gameId ? `?gameId=${encodeURIComponent(this.data.gameId)}` : ''}`
    }
  }
})
