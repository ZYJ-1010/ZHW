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

function getLayoutStyles() {
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

function normalizePercent(value) {
  const percent = Number(value)

  if (!Number.isFinite(percent)) {
    return 0
  }

  return Math.max(0, Math.min(100, percent))
}

function normalizeTask(task = {}) {
  return {
    title: task.title || task.statusText || task.status || '',
    desc: task.desc || task.description || task.content || ''
  }
}

function getMemberName(member = {}) {
  return member.displayName || member.name || member.nickname || ''
}

function normalizeMessage(message = {}) {
  return {
    name: message.name || message.senderName || message.nickname || '',
    content: message.content || message.text || message.message || ''
  }
}

function normalizeCollaboration(data = {}) {
  const progress = data.progress || {}
  const members = Array.isArray(data.members) ? data.members : []
  const tasks = Array.isArray(progress.tasks || data.tasks)
    ? (progress.tasks || data.tasks).map(normalizeTask)
    : []
  const messages = Array.isArray(data.messages || data.chatMessages)
    ? (data.messages || data.chatMessages).map(normalizeMessage)
    : []
  const membersText = data.membersText || members.map(getMemberName).filter(Boolean).join(' · ')

  return {
    progressPercent: normalizePercent(progress.percent || data.progressPercent),
    membersText,
    tasks,
    messages,
    memberManageRoute: data.memberManageRoute || '',
    endRoute: data.endRoute || ''
  }
}

Page({
  data: {
    layout: getLayoutStyles(),
    gameId: '',
    progressPercent: 0,
    membersText: '',
    tasks: [],
    messages: [],
    memberManageRoute: '',
    endRoute: '',
    loading: false,
    ending: false
  },

  onLoad(options = {}) {
    this.loadCollaboration(options)
  },

  onShow() {
    this.setData({
      layout: getLayoutStyles()
    })
  },

  async loadCollaboration(options = {}) {
    const gameId = options.gameId || options.id || this.data.gameId || ''

    this.setData({
      gameId,
      loading: true
    })

    try {
      const data = await gameService.getGameCollaboration({
        ...options,
        gameId
      })

      this.setData({
        ...normalizeCollaboration(data),
        gameId: data.gameId || gameId,
        loading: false
      })
    } catch (error) {
      this.setData({
        ...normalizeCollaboration({}),
        loading: false
      })
      wx.showToast({
        title: error.message || '协作信息加载失败',
        icon: 'none'
      })
    }
  },

  handleBack() {
    if (getCurrentPages().length > 1) {
      wx.navigateBack()
      return
    }

    wx.navigateTo({
      url: '/pages/game/hall/index'
    })
  },

  handleMemberManage() {
    if (this.data.memberManageRoute) {
      wx.navigateTo({
        url: this.data.memberManageRoute
      })
      return
    }

    wx.showToast({
      title: '成员管理待接入',
      icon: 'none'
    })
  },

  async handleEndSession() {
    if (this.data.endRoute) {
      wx.navigateTo({
        url: this.data.endRoute
      })
      return
    }

    if (this.data.ending) {
      return
    }

    this.setData({
      ending: true
    })

    try {
      await gameService.endGameCollaboration({
        gameId: this.data.gameId
      })
      wx.showToast({
        title: '已提交结束',
        icon: 'none'
      })
      this.loadCollaboration({
        gameId: this.data.gameId
      })
    } catch (error) {
      wx.showToast({
        title: error.message || '结束本局失败',
        icon: 'none'
      })
    } finally {
      this.setData({
        ending: false
      })
    }
  }
})
