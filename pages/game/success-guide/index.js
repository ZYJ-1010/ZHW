const gameService = require('../../../services/game')
const { ROUTES } = require('../../../config/routes')
const { getSurnameInitials } = require('../../../utils/avatar')
const { navigateShellRoute } = require('../../../utils/shell-nav')

const CONTENT_LEFT_RPX = 2
const CONTENT_TOP_RPX = 160
const CONTENT_WIDTH_RPX = 750
const DESIGN_FRAME_HEIGHT_PT = 810
const DESIGN_BOTTOM_HEIGHT_PT = 78
const NAV_TITLE_HEIGHT_RPX = 50
const BACK_BUTTON_SIZE_RPX = 40
const DEFAULT_CAPSULE_BOTTOM_RPX = 142
const DEFAULT_FRAME_HEIGHT_RPX = DESIGN_FRAME_HEIGHT_PT * 2
const CTA_BAR_HEIGHT_RPX = 158

const DEFAULT_SUCCESS_GUIDE_DETAIL = {
  timeline: [],
  party: {
    confirmedText: '',
    cardClass: 'success-guide-party-card',
    cardStyle: 'width: 684rpx; height: 446rpx; margin: 32rpx auto 0; box-shadow: none;',
    titleClass: 'regular',
    player: {
      avatarClass: 'pink',
      name: '',
      avatarText: '',
      role: '玩家',
      state: '',
      stateClass: ''
    },
    expert: {
      avatarClass: 'blue',
      name: '',
      avatarText: '',
      role: '行家',
      state: '',
      stateClass: ''
    }
  },
  followUps: [],
  reward: {
    show: false,
    value: '',
    label: ''
  }
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

function getWhiteShellLayoutStyles() {
  const capsuleBottom = getCapsuleBottomRpx()
  const frameHeight = getFrameHeightRpx()
  const bottomHeight = roundRpx(frameHeight * DESIGN_BOTTOM_HEIGHT_PT / DESIGN_FRAME_HEIGHT_PT)
  const bottomTop = Math.max(CONTENT_TOP_RPX, roundRpx(frameHeight - bottomHeight))
  const contentHeight = Math.max(0, roundRpx(bottomTop - CONTENT_TOP_RPX))
  const titleTop = Math.max(0, roundRpx(capsuleBottom - NAV_TITLE_HEIGHT_RPX))
  const backTop = Math.max(0, roundRpx(capsuleBottom - BACK_BUTTON_SIZE_RPX))
  const ctaTop = Math.max(CONTENT_TOP_RPX, roundRpx(frameHeight - CTA_BAR_HEIGHT_RPX))

  return {
    frameStyle: `height: ${frameHeight}rpx; min-height: ${frameHeight}rpx;`,
    contentStyle: [
      `left: ${CONTENT_LEFT_RPX}rpx`,
      `top: ${CONTENT_TOP_RPX}rpx`,
      `width: ${CONTENT_WIDTH_RPX}rpx`,
      `height: ${contentHeight}rpx`
    ].join('; '),
    bottomStyle: `top: ${bottomTop}rpx; height: ${bottomHeight}rpx;`,
    ctaStyle: `top: ${ctaTop}rpx; height: ${CTA_BAR_HEIGHT_RPX}rpx;`,
    titleStyle: `top: ${titleTop}rpx; height: ${NAV_TITLE_HEIGHT_RPX}rpx; line-height: ${NAV_TITLE_HEIGHT_RPX}rpx;`,
    backStyle: `top: ${backTop}rpx; width: ${BACK_BUTTON_SIZE_RPX}rpx; height: ${BACK_BUTTON_SIZE_RPX}rpx;`
  }
}

function normalizePartyMember(member = {}, fallback = {}) {
  const name = member.name || fallback.name || ''

  return Object.assign({}, fallback, member, {
    avatarText: member.avatarText || member.avatar || getSurnameInitials(name, fallback.avatarText || 'ME')
  })
}

function normalizeSuccessGuideDetail(data = {}) {
  const source = data && typeof data === 'object' ? data : {}
  const party = Object.assign({}, DEFAULT_SUCCESS_GUIDE_DETAIL.party, source.party || {})

  return {
    viewer: source.viewer || {},
    timeline: Array.isArray(source.timeline) && source.timeline.length ? source.timeline : [],
    party: Object.assign({}, party, {
      player: normalizePartyMember(party.player, DEFAULT_SUCCESS_GUIDE_DETAIL.party.player),
      expert: normalizePartyMember(party.expert, DEFAULT_SUCCESS_GUIDE_DETAIL.party.expert)
    }),
    followUps: Array.isArray(source.followUps) && source.followUps.length ? source.followUps : [],
    reward: Object.assign({}, DEFAULT_SUCCESS_GUIDE_DETAIL.reward, source.reward || {})
  }
}

Page({
  data: {
    shellLayout: getWhiteShellLayoutStyles(),
    gameId: '',
    loading: false,
    errorText: '',
    ...normalizeSuccessGuideDetail(DEFAULT_SUCCESS_GUIDE_DETAIL)
  },

  onLoad(options = {}) {
    const gameId = options.gameId || options.id || ''

    this.setData({ gameId })
    if (gameId) {
      this.loadSuccessGuideDetail(gameId)
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
      shellLayout: getWhiteShellLayoutStyles()
    })
  },

  async loadSuccessGuideDetail(gameId) {
    this.setData({
      loading: true,
      errorText: ''
    })

    try {
      const data = await gameService.getGameGuideSuccessDetail(gameId)

      this.setData({
        loading: false,
        errorText: '',
        ...normalizeSuccessGuideDetail(data)
      })
    } catch (error) {
      this.setData({
        loading: false,
        errorText: error.message || '领路人成功详情加载失败，请稍后重试'
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

  async onFollowTap(event) {
    const item = this.data.followUps.find((entry) => entry.key === event.currentTarget.dataset.key)

    if (!item) {
      wx.showToast({ title: '未找到跟进项', icon: 'none' })
      return
    }

    const gameId = this.data.gameId || item.target && item.target.gameId || ''
    let target = item.target || {}

    if (gameId) {
      try {
        const result = await gameService.createGuideFollowUp(gameId, { action: item.key })
        target = result.target || target
      } catch (error) {
        wx.showToast({ title: error.message || '跟进记录失败，继续打开页面', icon: 'none' })
      }
    }

    this.navigateFollowTarget(item.key, target)
  },

  navigateFollowTarget(action, target = {}) {
    const gameId = this.data.gameId || target.gameId || ''

    if (target.type === 'im_room' || action === 'feedback') {
      navigateShellRoute(`/${ROUTES.imRoom}?gameId=${encodeURIComponent(gameId)}&role=guide&prefill=${encodeURIComponent(target.prefill || '我来跟进一下本次组局双方反馈。')}`)
      return
    }

    if (target.type === 'referral_record' || action === 'deal') {
      navigateShellRoute(`/${ROUTES.gameReferralRecord}${gameId ? `?gameId=${encodeURIComponent(gameId)}` : ''}`)
      return
    }

    navigateShellRoute(`/${ROUTES.gameDetail}${gameId ? `?id=${encodeURIComponent(gameId)}` : ''}`)
  },

  onShareAppMessage() {
    return {
      title: '组局成功',
      path: `/${ROUTES.gameSuccessGuide}${this.data.gameId ? `?gameId=${encodeURIComponent(this.data.gameId)}` : ''}`
    }
  }
})
