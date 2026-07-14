const { getLegacyWhiteFrameLayoutStyles } = require('../../../utils/adaptive-shell-layout')
const gameService = require('../../../services/game')
const { ROUTES } = require('../../../config/routes')
const { getSurnameInitials } = require('../../../utils/avatar')
const { navigateShellRoute } = require('../../../utils/shell-nav')
const { GAME_ICON_SRC, resolveGuideFollowIconSrc } = require('../shared/icons')

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
const LOCAL_LOAD_FAILED_TEXT = '领路人成功详情加载失败，请稍后重试'
const LOCAL_FOLLOW_MISSING_TEXT = '未找到跟进项'
const LOCAL_FOLLOW_FAILED_TEXT = '跟进记录失败，继续打开页面'
const LOCAL_SHARE_TITLE = '组局成功'

const DEFAULT_SUCCESS_GUIDE_DETAIL = {
  pageTexts: {
    navTitle: '组局成功',
    heroTitle: '恭喜！组局成功',
    heroDesc: '你成功促成了这次连接',
    timelineTitle: '成局历程',
    followTitle: '后续跟进',
    followEmptyText: '暂无后续跟进项',
    followMissingText: LOCAL_FOLLOW_MISSING_TEXT,
    loadFailedText: LOCAL_LOAD_FAILED_TEXT,
    followFailedText: LOCAL_FOLLOW_FAILED_TEXT,
    defaultImPrefill: '我来跟进一下本次组局双方反馈。',
    shareTitle: LOCAL_SHARE_TITLE,
    shareButtonText: '分享成局喜悦'
  },
  gameId: '',
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
  },
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

function getWhiteShellLayoutStyles() {
  const layout = getLegacyWhiteFrameLayoutStyles({
    contentLeftRpx: CONTENT_LEFT_RPX,
    contentWidthRpx: CONTENT_WIDTH_RPX,
    bottomHeightRatio: DESIGN_BOTTOM_HEIGHT_PT / DESIGN_FRAME_HEIGHT_PT,
    titleHeightRpx: NAV_TITLE_HEIGHT_RPX,
    backSizeRpx: BACK_BUTTON_SIZE_RPX
  })
  const ctaTop = Math.max(
    layout.contentTopRpx,
    roundRpx(layout.canvasHeightRpx - layout.safeBottomRpx - CTA_BAR_HEIGHT_RPX)
  )

  return Object.assign({}, layout, {
    ctaStyle: `top: ${ctaTop}rpx; height: ${CTA_BAR_HEIGHT_RPX}rpx;`
  })
}

function normalizePartyMember(member = {}, fallback = {}) {
  const name = member.name || fallback.name || ''

  return Object.assign({}, fallback, member, {
    avatarText: member.avatarText || member.avatar || member.initials || getSurnameInitials(name, fallback.avatarText || 'ME'),
    avatarType: member.avatarType || fallback.avatarType || ''
  })
}

function normalizeSuccessGuideDetail(data = {}) {
  const source = data && typeof data === 'object' ? data : {}
  const party = Object.assign({}, DEFAULT_SUCCESS_GUIDE_DETAIL.party, source.party || {})
  const followUps = Array.isArray(source.followUps) && source.followUps.length
    ? source.followUps.map((item) => Object.assign({}, item, {
      iconSrc: resolveGuideFollowIconSrc(item.key, item.iconSrc)
    }))
    : []

  return {
    pageTexts: Object.assign({}, DEFAULT_SUCCESS_GUIDE_DETAIL.pageTexts, source.pageTexts || {}),
    viewer: source.viewer || {},
    gameId: source.gameId || '',
    timeline: Array.isArray(source.timeline) && source.timeline.length ? source.timeline : [],
    party: Object.assign({}, party, {
      player: normalizePartyMember(party.player, DEFAULT_SUCCESS_GUIDE_DETAIL.party.player),
      expert: normalizePartyMember(party.expert, DEFAULT_SUCCESS_GUIDE_DETAIL.party.expert)
    }),
    followUps,
    reward: Object.assign({}, DEFAULT_SUCCESS_GUIDE_DETAIL.reward, source.reward || {}),
    iconSources: GAME_ICON_SRC
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

  async onFollowTap(event) {
    const item = this.data.followUps.find((entry) => entry.key === event.currentTarget.dataset.key)

    if (!item) {
      wx.showToast({ title: this.data.pageTexts.followMissingText || LOCAL_FOLLOW_MISSING_TEXT, icon: 'none' })
      return
    }

    const gameId = this.data.gameId || item.target && item.target.gameId || ''
    let target = item.target || {}

    if (gameId) {
      try {
        const result = await gameService.createGuideFollowUp(gameId, { action: item.key })
        target = result.target || target
      } catch (error) {
        wx.showToast({ title: error.message || this.data.pageTexts.followFailedText || LOCAL_FOLLOW_FAILED_TEXT, icon: 'none' })
      }
    }

    this.navigateFollowTarget(item.key, target)
  },

  navigateFollowTarget(action, target = {}) {
    const gameId = this.data.gameId || target.gameId || ''

    if (target.type === 'im_room' || action === 'feedback') {
      const query = [
        `gameId=${encodeURIComponent(gameId)}`,
        'role=guide'
      ]

      if (target.roomId) {
        query.push(`roomId=${encodeURIComponent(target.roomId)}`)
      }
      if (target.prefill || this.data.pageTexts.defaultImPrefill) {
        query.push(`prefill=${encodeURIComponent(target.prefill || this.data.pageTexts.defaultImPrefill)}`)
      }

      navigateShellRoute(`/${ROUTES.imRoom}?${query.join('&')}`)
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
      title: this.data.pageTexts.shareTitle || LOCAL_SHARE_TITLE,
      path: `/${ROUTES.gameSuccessGuide}${this.data.gameId ? `?gameId=${encodeURIComponent(this.data.gameId)}` : ''}`
    }
  }
})
