const { getLegacyWhiteFrameLayoutStyles } = require('../../../utils/adaptive-shell-layout')
const gameService = require('../../../services/game')
const { ROUTES } = require('../../../config/routes')
const { getSurnameInitials } = require('../../../utils/avatar')
const { navigateShellRoute } = require('../../../utils/shell-nav')
const { GAME_ICON_SRC } = require('../shared/icons')

const CONTENT_LEFT_RPX = 2
const CONTENT_TOP_RPX = 160
const CONTENT_WIDTH_RPX = 750
const DESIGN_FRAME_HEIGHT_PT = 810
const DESIGN_BOTTOM_HEIGHT_PT = 78
const NAV_TITLE_HEIGHT_RPX = 50
const BACK_BUTTON_SIZE_RPX = 40
const MORE_BUTTON_SIZE_RPX = 44
const MORE_BUTTON_LEFT_RPX = 658
const DEFAULT_CAPSULE_BOTTOM_RPX = 142
const DEFAULT_FRAME_HEIGHT_RPX = DESIGN_FRAME_HEIGHT_PT * 2
const LOCAL_LOAD_FAILED_TEXT = '组局成功详情加载失败，请稍后重试'

const DEFAULT_SUCCESS_DETAIL = {
  pageTexts: {
    navTitle: '组局成功',
    successHeading: '组局成功!',
    subtitleFallback: '成功详情以接口返回为准',
    groupSectionTitle: '局内成员',
    onlineText: '局内',
    chatButtonText: '进入局内群聊',
    activityTitle: '本局信息',
    fundTitle: '',
    fundEmptyText: '',
    stepsTitle: '下一步',
    stepsEmptyText: '暂无下一步动作',
    manageButtonText: '进入局管理',
    loadFailedText: LOCAL_LOAD_FAILED_TEXT,
    chatPrefill: '你好，想和大家确认一下本局安排。'
  },
  gameId: '',
  manageRoute: '',
  group: {
    title: '',
    roles: '',
    hint: ''
  },
  participants: [],
  activityRows: [],
  fund: {
    show: false,
    title: '',
    desc: '',
    amountText: '',
    progress: 0,
    stepLabels: []
  },
  nextSteps: [],
  iconSources: GAME_ICON_SRC
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

function getBlankShellLayoutStyles() {
  const layout = getLegacyWhiteFrameLayoutStyles({
    contentLeftRpx: CONTENT_LEFT_RPX,
    contentWidthRpx: CONTENT_WIDTH_RPX,
    bottomHeightRatio: DESIGN_BOTTOM_HEIGHT_PT / DESIGN_FRAME_HEIGHT_PT,
    titleHeightRpx: NAV_TITLE_HEIGHT_RPX,
    backSizeRpx: BACK_BUTTON_SIZE_RPX
  })
  const moreTop = Math.max(0, roundRpx(
    layout.capsuleTopRpx + (layout.capsuleHeightRpx - MORE_BUTTON_SIZE_RPX) / 2
  ))

  return Object.assign({}, layout, {
    moreStyle: `left: ${MORE_BUTTON_LEFT_RPX}rpx; top: ${moreTop}rpx; width: ${MORE_BUTTON_SIZE_RPX}rpx; height: ${MORE_BUTTON_SIZE_RPX}rpx;`
  })
}

function normalizeSuccessDetail(data = {}) {
  const source = data && typeof data === 'object' ? data : {}
  const fund = Object.assign({}, DEFAULT_SUCCESS_DETAIL.fund, source.fund || {})
  const showFund = typeof fund.show === 'boolean'
    ? fund.show
    : Boolean(fund.title || fund.desc || fund.amountText || (Array.isArray(fund.stepLabels) && fund.stepLabels.length))
  const participants = Array.isArray(source.participants) && source.participants.length
    ? source.participants.map((item, index) => ({
      id: item.id || item.userId || `participant-${index}`,
      avatar: item.avatar || item.avatarText || item.initials || getSurnameInitials(item.name || '', item.roleLabel || '成员'),
      avatarUrl: item.avatarUrl || item.avatarURL || '',
      avatarType: item.avatarType || '',
      colorClass: item.colorClass || item.avatarClass || 'blue',
      roleLabel: item.roleLabel || item.roleText || item.role || '',
      name: item.name || item.nickname || ''
    }))
    : []
  const avatarCounts = participants.reduce((counts, item) => {
    if (item.avatar) {
      counts[item.avatar] = (counts[item.avatar] || 0) + 1
    }
    return counts
  }, {})
  participants.forEach((item) => {
    if (item.avatar && avatarCounts[item.avatar] > 1) {
      item.avatar = String(item.id || '').slice(-2) || item.avatar
    }
  })

  return {
    pageTexts: Object.assign({}, DEFAULT_SUCCESS_DETAIL.pageTexts, source.pageTexts || {}),
    viewer: source.viewer || {},
    gameId: source.gameId || (source.group && source.group.gameId) || '',
    manageRoute: source.manageRoute || (source.group && source.group.manageRoute) || '',
    group: Object.assign({}, DEFAULT_SUCCESS_DETAIL.group, source.group || {}),
    participants,
    participantCount: Number(source.participantCount || 0) || participants.length,
    activityRows: Array.isArray(source.activityRows) && source.activityRows.length ? source.activityRows : [],
    fund: Object.assign({}, fund, {
      show: showFund,
      progressStyle: `width: ${Math.max(0, Math.min(100, Number(fund.progress || 0)))}%;`,
      stepLabels: Array.isArray(fund.stepLabels) && fund.stepLabels.length ? fund.stepLabels : []
    }),
    nextSteps: Array.isArray(source.nextSteps) && source.nextSteps.length ? source.nextSteps : [],
    iconSources: GAME_ICON_SRC
  }
}

Page({
  data: {
    shellLayout: getBlankShellLayoutStyles(),
    gameId: '',
    loading: false,
    errorText: '',
    iconSources: GAME_ICON_SRC,
    ...normalizeSuccessDetail(DEFAULT_SUCCESS_DETAIL)
  },

  onLoad(options = {}) {
    const gameId = options.gameId || options.id || ''

    this.setData({ gameId })
    if (gameId) {
      this.loadSuccessDetail(gameId)
    }
  },

  onShow() {
    this.updateShellLayout()
  },

  onResize() {
    this.updateShellLayout()
  },

  updateShellLayout() {
    this.setData({
      shellLayout: getBlankShellLayoutStyles()
    })
  },

  async loadSuccessDetail(gameId) {
    this.setData({
      loading: true,
      errorText: ''
    })

    try {
      const data = await gameService.getGameSuccessDetail(gameId, { role: 'expert' })

      this.setData({
        loading: false,
        errorText: '',
        ...normalizeSuccessDetail(data)
      })
    } catch (error) {
      this.setData({
        loading: false,
        errorText: error.message || this.data.pageTexts.loadFailedText || LOCAL_LOAD_FAILED_TEXT
      })
    }
  },

  onBackTap() {
    const pages = getCurrentPages()

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    navigateShellRoute(ROUTES.gameGuideProgress)
  },

  onMoreTap() {
    this.onManageTap()
  },

  onEnterChatTap() {
    const gameId = this.data.gameId || this.data.group.gameId || ''
    const roomId = this.data.group.roomId || ''

    if (!roomId) {
      this.showInfo('本局尚未开始，暂不可进入局内群聊')
      return
    }
    const query = [
      `gameId=${encodeURIComponent(gameId)}`,
      `role=${encodeURIComponent(this.data.viewer.role || 'expert')}`
    ]

    if (roomId) {
      query.push(`roomId=${encodeURIComponent(roomId)}`)
    }
    if (this.data.pageTexts.chatPrefill) {
      query.push(`prefill=${encodeURIComponent(this.data.pageTexts.chatPrefill)}`)
    }

    navigateShellRoute(`/${ROUTES.imRoom}?${query.join('&')}`)
  },

  onStepTap(event) {
    const action = event.currentTarget.dataset.action

    if (action === 'contact_player') {
      this.onEnterChatTap()
      return
    }

    this.onManageTap()
  },

  onManageTap() {
    if (this.data.manageRoute) {
      navigateShellRoute(this.data.manageRoute)
      return
    }

    const gameId = this.data.gameId || this.data.group.gameId || ''
    navigateShellRoute(`/pages/profile/service-center/my-games/index?category=created${gameId ? `&gameId=${encodeURIComponent(gameId)}` : ''}`)
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
