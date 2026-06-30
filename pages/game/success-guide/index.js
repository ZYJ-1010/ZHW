const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')
const { getSurnameInitials } = require('../../../utils/avatar')

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
const FOLLOW_UP_ICON_MAP = {
  schedule: {
    iconSrc: './assets/follow-schedule.png',
    iconClass: 'schedule',
    theme: 'blue'
  },
  feedback: {
    iconSrc: './assets/follow-feedback.png',
    iconClass: 'feedback',
    theme: 'purple'
  },
  deal: {
    iconSrc: './assets/follow-deal.png',
    iconClass: 'deal',
    theme: 'orange'
  }
}

function pickFirstValue() {
  const values = Array.prototype.slice.call(arguments)

  for (let index = 0; index < values.length; index += 1) {
    if (values[index] !== undefined && values[index] !== null && values[index] !== '') {
      return values[index]
    }
  }

  return ''
}

function normalizeList(list) {
  return Array.isArray(list) ? list : []
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

function normalizeMember(member, roleType) {
  const source = member || {}
  const name = pickFirstValue(source.name, source.nickname, source.realname)

  return {
    id: pickFirstValue(source.id, source.userId),
    avatarClass: pickFirstValue(source.avatarClass, roleType === 'expert' ? 'blue' : 'pink'),
    avatarUrl: pickFirstValue(source.avatarUrl, source.avatar),
    name,
    avatarText: pickFirstValue(source.avatarText, source.initials, name ? getSurnameInitials(name, roleType === 'expert' ? 'EX' : 'PL') : ''),
    role: pickFirstValue(source.role, source.roleLabel, roleType === 'expert' ? '行家' : '玩家'),
    state: pickFirstValue(source.state, source.statusText),
    stateClass: pickFirstValue(source.stateClass, source.statusClass)
  }
}

function normalizeParty(data) {
  const source = data || {}
  const party = source.party || source
  const player = normalizeList(party.players || party.playerList)[0] || party.player || source.player || {}
  const expert = normalizeList(party.experts || party.expertList)[0] || party.expert || source.expert || {}

  return {
    confirmedText: pickFirstValue(party.confirmedText),
    cardClass: 'success-guide-party-card',
    cardStyle: 'width: 684rpx; height: 446rpx; margin: 32rpx auto 0; box-shadow: none;',
    titleClass: 'regular',
    player: normalizeMember(player, 'player'),
    expert: normalizeMember(expert, 'expert')
  }
}

function normalizeTimeline(data) {
  const source = data || {}

  return normalizeList(source.timeline || source.steps || source.progress)
    .map((item) => ({
      title: pickFirstValue(item.title, item.name),
      desc: pickFirstValue(item.desc, item.description),
      time: pickFirstValue(item.time, item.timeText, item.createdAt),
      active: Boolean(item.active || item.current)
    }))
    .filter((item) => item.title || item.desc || item.time)
}

function normalizeFollowUps(data) {
  const source = data || {}

  return normalizeList(source.followUps || source.actions || source.nextActions)
    .map((item) => {
      const key = pickFirstValue(item.key, item.id, item.action)
      const meta = FOLLOW_UP_ICON_MAP[key] || FOLLOW_UP_ICON_MAP.schedule

      return {
        key,
        title: pickFirstValue(item.title, item.name),
        desc: pickFirstValue(item.desc, item.description),
        reward: pickFirstValue(item.reward, item.rewardText),
        route: pickFirstValue(item.route, item.path),
        message: pickFirstValue(item.message, item.toastText),
        iconSrc: pickFirstValue(item.localIcon, meta.iconSrc),
        iconClass: pickFirstValue(item.iconClass, meta.iconClass),
        theme: pickFirstValue(item.theme, meta.theme)
      }
    })
    .filter((item) => item.key)
}

Page({
  data: {
    shellLayout: getWhiteShellLayoutStyles(),
    queryParams: {},
    timeline: [],
    party: normalizeParty({}),
    followUps: []
  },

  onLoad(options = {}) {
    this.setData({
      queryParams: options
    })
    this.loadSuccessGuide(options)
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

  async loadSuccessGuide(params = {}) {
    try {
      const data = await gameService.getGuideSuccess(params)

      this.setData({
        timeline: normalizeTimeline(data),
        party: normalizeParty(data),
        followUps: normalizeFollowUps(data)
      })
    } catch (error) {
      this.setData({
        timeline: [],
        party: normalizeParty({}),
        followUps: []
      })
      wx.showToast({
        title: error.message || '组局成功信息加载失败',
        icon: 'none'
      })
    }
  },

  onBackTap() {
    const pages = getCurrentPages()

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    wx.navigateTo({
      url: `/${ROUTES.gameGuideProgress}`
    })
  },

  onFollowTap(event) {
    const item = this.data.followUps.find((entry) => entry.key === event.currentTarget.dataset.key)

    if (item && item.route) {
      wx.navigateTo({
        url: item.route
      })
      return
    }

    wx.showToast({
      title: item && (item.message || item.title) || '后续跟进待接入',
      icon: 'none'
    })
  },

  onShareAppMessage() {
    return {
      title: '组局成功',
      path: `/${ROUTES.gameSuccessGuide}`
    }
  }
})
