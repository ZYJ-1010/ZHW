const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')

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

Page({
  data: {
    gameId: '',
    whiteShellLayout: getWhiteShellLayoutStyles(),
    loading: false,
    loadErrorText: '',
    participants: []
  },

  onLoad(options = {}) {
    const gameId = options.gameId || options.id || ''

    this.setData({
      gameId,
      whiteShellLayout: getWhiteShellLayoutStyles(),
      participants: []
    })

    if (gameId) {
      this.loadParticipants(gameId)
    }
  },

  async loadParticipants(gameId) {
    this.setData({
      loading: true,
      loadErrorText: ''
    })

    try {
      const data = await gameService.getGameMembers(gameId, {
        page: 1,
        pageSize: 100
      })
      const source = Array.isArray(data) ? data : (data.list || data.records || data.items || data.members || [])

      this.setData({
        loading: false,
        participants: source.map(this.normalizeParticipant)
      })
    } catch (error) {
      this.setData({
        loading: false,
        loadErrorText: error.message || '参与者加载失败',
        participants: []
      })
      wx.showToast({
        title: error.message || '参与者加载失败',
        icon: 'none'
      })
    }
  },

  normalizeParticipant(member = {}) {
    const user = member.user || member.profile || member
    const name = user.name || user.nickname || member.name || member.nickname || ''

    return {
      id: member.id || member.userId || user.id || name,
      name,
      avatarSrc: user.avatarSrc || user.avatarUrl || member.avatarSrc || member.avatarUrl || '',
      avatarText: user.avatarText || member.avatarText || name.slice(0, 1),
      role: member.roleText || member.role || user.roleText || '',
      roleClass: member.roleClass || member.role || '',
      position: user.position || user.title || member.position || member.title || '',
      topic: member.topic || member.summary || user.summary || '',
      primaryTag: member.primaryTag || member.tagText || '',
      tags: Array.isArray(member.tags || user.tags) ? (member.tags || user.tags) : [],
      location: member.location || member.address || user.location || '',
      distance: member.distanceText || member.distance || ''
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

    wx.redirectTo({
      url: `/${ROUTES.gameDetail}${this.data.gameId ? `?id=${encodeURIComponent(this.data.gameId)}` : ''}`
    })
  },

  onParticipantTap(event) {
    const participant = event.detail && event.detail.participant
    const name = (participant && participant.name) || '参与者'

    wx.showToast({
      title: `${name}资料待接入`,
      icon: 'none'
    })
  },

  onShareAppMessage() {
    return {
      title: '全部参与者',
      path: `/${ROUTES.gameParticipants}${this.data.gameId ? `?gameId=${encodeURIComponent(this.data.gameId)}` : ''}`
    }
  }
})
