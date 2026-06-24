const { ROUTES } = require('../../../config/routes')
const { getDefaultGameParticipants } = require('../shared/participants')

const DEFAULT_CONTENT_TOP_RPX = 160
const NAV_BOTTOM_GAP_RPX = 18
const NAV_TITLE_HEIGHT_RPX = 50
const NAV_BUTTON_SIZE_RPX = 44
const BOTTOM_ACTION_RPX = 148
const DEFAULT_CAPSULE_BOTTOM_RPX = 142
const DEFAULT_SHARE_RIGHT_RPX = 206
const SHARE_CAPSULE_GAP_RPX = 18

function roundRpx(value) {
  return Math.round(value * 100) / 100
}

function getMenuMetricsRpx() {
  try {
    if (typeof wx !== 'undefined' && wx.getMenuButtonBoundingClientRect && wx.getSystemInfoSync) {
      const menuButton = wx.getMenuButtonBoundingClientRect()
      const systemInfo = wx.getSystemInfoSync()

      if (menuButton && systemInfo && systemInfo.windowWidth) {
        const ratio = 750 / systemInfo.windowWidth
        const capsuleBottom = roundRpx((menuButton.top + menuButton.height) * ratio)
        const capsuleLeftGap = menuButton.left
          ? roundRpx((systemInfo.windowWidth - menuButton.left) * ratio)
          : DEFAULT_SHARE_RIGHT_RPX - SHARE_CAPSULE_GAP_RPX

        return {
          capsuleBottom,
          shareRight: roundRpx(capsuleLeftGap + SHARE_CAPSULE_GAP_RPX)
        }
      }
    }
  } catch (error) {
    return {
      capsuleBottom: DEFAULT_CAPSULE_BOTTOM_RPX,
      shareRight: DEFAULT_SHARE_RIGHT_RPX
    }
  }

  return {
    capsuleBottom: DEFAULT_CAPSULE_BOTTOM_RPX,
    shareRight: DEFAULT_SHARE_RIGHT_RPX
  }
}

function getWhiteDetailLayout() {
  const { capsuleBottom, shareRight } = getMenuMetricsRpx()
  const contentTop = Math.max(DEFAULT_CONTENT_TOP_RPX, roundRpx(capsuleBottom + NAV_BOTTOM_GAP_RPX))
  const titleTop = Math.max(0, roundRpx(capsuleBottom - NAV_TITLE_HEIGHT_RPX))
  const buttonTop = Math.max(0, roundRpx(capsuleBottom - NAV_BUTTON_SIZE_RPX))

  return {
    headerStyle: `height: ${contentTop}rpx;`,
    titleStyle: `top: ${titleTop}rpx; height: ${NAV_TITLE_HEIGHT_RPX}rpx; line-height: ${NAV_TITLE_HEIGHT_RPX}rpx;`,
    backStyle: `top: ${buttonTop}rpx; width: ${NAV_BUTTON_SIZE_RPX}rpx; height: ${NAV_BUTTON_SIZE_RPX}rpx;`,
    shareStyle: `top: ${buttonTop}rpx; right: ${shareRight}rpx; width: ${NAV_BUTTON_SIZE_RPX}rpx; height: ${NAV_BUTTON_SIZE_RPX}rpx;`,
    scrollStyle: `top: ${contentTop}rpx; height: calc(100vh - ${contentTop}rpx - ${BOTTOM_ACTION_RPX}rpx - env(safe-area-inset-bottom));`
  }
}

function isRealnameVerified(user) {
  if (!user) {
    return false
  }

  if (user.needRealname === true || user.realnameRequired === true) {
    return false
  }

  const status = user.realnameStatus || user.authStatus || user.certificationStatus

  return status === 'verified' ||
    status === 'approved' ||
    status === 'passed' ||
    status === 'success' ||
    status === true ||
    user.realnameVerified === true ||
    user.isRealnameVerified === true ||
    user.verified === true ||
    user.needRealname === false
}

function isEndedStatus(value) {
  const text = String(value == null ? '' : value).trim().toLowerCase()

  return /^(1|true|yes|y|ended|end|finished|finish|closed|expired|completed|complete|done)$/.test(text) ||
    /已结束|结束|报名结束|已完成|完成|已关闭|关闭|已过期|过期/.test(text)
}

function isActiveStatus(value) {
  const text = String(value == null ? '' : value).trim().toLowerCase()

  return /^(0|false|no|n|active|open|opening|available|pending|processing|ongoing|upcoming|unstarted)$/.test(text) ||
    /报名中|可报名|进行中|未开始|待开始|即将开始|开放/.test(text)
}

function hasStatusSignal(value) {
  return value !== undefined && value !== null && String(value).trim() !== ''
}

function parseChineseEventEndTime(text = '') {
  const matched = String(text).match(/(?:(\d{4})年)?(\d{1,2})月(\d{1,2})日(?:[^\d]*(\d{1,2}):(\d{2}))?(?:\s*(?:-|--|—|~|～|至)\s*(?:(\d{1,2})月(\d{1,2})日\s*)?(\d{1,2}):(\d{2}))?/)

  if (!matched) {
    return null
  }

  const now = new Date()
  const year = Number(matched[1] || now.getFullYear())
  const startMonth = Number(matched[2])
  const startDay = Number(matched[3])
  const startHour = Number(matched[4] || 23)
  const startMinute = Number(matched[5] || 59)
  const endMonth = Number(matched[6] || startMonth)
  const endDay = Number(matched[7] || startDay)
  const endHour = Number(matched[8] || matched[4] || 23)
  const endMinute = Number(matched[9] || matched[5] || 59)
  const timestamp = new Date(year, endMonth - 1, endDay, endHour, endMinute).getTime()

  if (Number.isNaN(timestamp)) {
    return null
  }

  const startTimestamp = new Date(year, startMonth - 1, startDay, startHour, startMinute).getTime()

  return timestamp < startTimestamp ? timestamp + 24 * 60 * 60 * 1000 : timestamp
}

function getEventEndTimestamp(event = {}) {
  const directTime = event.endAt || event.endedAt || event.endTime || event.registrationEndAt || event.registrationEndTime

  if (directTime) {
    const timestamp = Date.parse(directTime)

    if (!Number.isNaN(timestamp)) {
      return timestamp
    }
  }

  return parseChineseEventEndTime(event.time || event.timeText || event.startTimeText || '')
}

function getGameEndedState(options = {}, event = {}) {
  const statusSignals = [
    options.ended,
    options.isEnded,
    options.status,
    options.state,
    options.registrationStatus,
    event.ended,
    event.isEnded,
    event.status,
    event.state,
    event.registrationStatus,
    event.registrationState
  ]

  for (let index = 0; index < statusSignals.length; index += 1) {
    const signal = statusSignals[index]

    if (!hasStatusSignal(signal)) {
      continue
    }

    if (isEndedStatus(signal)) {
      return true
    }

    if (isActiveStatus(signal)) {
      return false
    }
  }

  const endTimestamp = getEventEndTimestamp(event)

  return typeof endTimestamp === 'number' ? endTimestamp <= Date.now() : false
}

Page({
  data: {
    gameId: '',
    interested: false,
    authPromptVisible: false,
    showShareWindow: false,
    detailScrollTop: 0,
    navLayout: getWhiteDetailLayout(),
    event: {
      coverSrc: '/pages/game/hall/assets/hall-featured-city.jpg',
      title: '企业数字化转型及技术服务沙龙局',
      time: '2月12日 08:30-11:30',
      location: '上海市浦东新区沙新镇黄赵路310号',
      category: '社交局',
      categoryIcon: '/pages/game/detail/assets/i26@3x.png',
      fee: '场地费AA/位'
    },
    isGameEnded: false,
    endedActionText: '报名结束',
    endedNoticeText: '新建组局将经过平台审核，审核通过后才能正式发布',
    stats: [
      { iconText: '👁️', text: '1,234次浏览', action: 'views' },
      { iconText: '💬', text: '3条评价', action: 'reviews' },
      { iconSrc: '/pages/game/detail/assets/i44@3x.png', text: '5/8人已报名' }
    ],
    tags: [
      { name: '#产品研发', tone: 'blue' },
      { name: '#创业', tone: 'green' },
      { name: '#社交', tone: 'purple' }
    ],
    organizer: {
      name: '陆毅',
      avatarSrc: '/pages/home/player/assets/ranking-avatar-01.png',
      avatarText: '陆',
      role: '总经理 | 上海创世界科技有限公司',
      summary: '已组局 88次 · 推荐20人',
      rating: '4.8'
    },
    introduction: '本场沙龙围绕企业数字化转型及技术服务话题，邀请多位成功创业者分享经验。活动包含主题分享、自由交流、资源对接三个环节，帮助参与者拓展人脉、获取资源。',
    highlights: [
      '实战大咖亲授，拒绝空泛理论',
      '精准资源对接，高效链接人脉',
      '全流程干货输出，内容覆盖全面',
      '轻量高效参会，时间成本可控'
    ],
    schedule: [
      {
        title: '签到入场',
        time: '08:30-08:50',
        desc: '参会人员现场签到，领取活动资料与伴手礼，自由熟悉场地，初步交流破冰'
      },
      {
        title: '主题分享环节',
        time: '08:50-10:20',
        desc: '多位成功创业者依次登台，围绕企业数字化转型实战经验、技术服务选型技巧、行业转型趋势、低成本高效转型方案等核心主题展开分享。预留简短提问时间，现场答疑解惑'
      },
      {
        title: '中场休息+自由交流',
        time: '10:20-10:40',
        desc: '短暂休整，参会者自由沟通，互换名片，初步对接需求'
      },
      {
        title: '资源对接+深度交流',
        time: '10:40-11:25',
        desc: '定向资源配对环节，主办方引导供需双方精准对接，针对性洽谈合作，针对共性问题展开集体讨论，搭建长期交流合作平台'
      },
      {
        title: '组局总结+合影留念',
        time: '11:25-11:30',
        desc: '主办方总结组局核心内容，公布后续社群交流渠道，全体参会人员合影留念，活动圆满结束'
      }
    ],
    detailImages: [
      '/pages/game/hall/assets/hall-card-desk.jpg',
      '/components/game-card/assets/cover-city.png'
    ],
    noticeLead: '为保障活动秩序与参会体验，敬请所有参会人员提前知悉以下事项，遵守活动规则：',
    noticeBullets: [
      '签到要求：请务必携带个人名片参会，便于现场人脉拓展与资源对接；需在08:50前完成签到入场，迟到超过30分钟将无法进入会场，敬请准时。',
      '参会对象限制：本次沙龙仅限企业负责人、核心管理层、技术负责人及创业团队成员参与，谢绝无关人员、非商务推广人员入场；仅限报名成功且收到确认通知的人员参与，不接受临时空降参会。',
      '行为规范：活动期间禁止随意打断分享、大声喧哗，保持会场安静；禁止发放无关小广告、恶意推销产品，违规者将被劝离会场；禁止录制嘉宾完整分享内容、私自传播活动内部资料，尊重知识产权与嘉宾隐私。',
      '资料与物品：活动资料、饮品由主办方统一提供，请勿自带零食饮料入内；个人贵重物品请自行妥善保管，主办方不负责财物保管。',
      '防疫与安全：参会期间请自觉维护会场卫生，遵守场地安全管理规定；如遇特殊情况，请及时联系现场工作人员协助处理。',
      '报名与取消：报名成功后如需取消参会，请至少提前1天告知主办方，方便释放名额给其他有需求的人员；无故缺席将影响后续参与各级活动报名资格。'
    ],
    audience: '企业负责人、运营管理者、技术负责人、创业团队核心成员、数字化服务相关从业者',
    participants: getDefaultGameParticipants()
  },

  onLoad(options = {}) {
    this.setData({
      gameId: options.gameId || options.id || '',
      isGameEnded: getGameEndedState(options, this.data.event),
      navLayout: getWhiteDetailLayout()
    })

    if (wx.showShareMenu) {
      wx.showShareMenu({
        withShareTicket: true,
        menus: ['shareAppMessage', 'shareTimeline']
      })
    }
  },

  onShow() {
    this.updateDetailLayout()
  },

  onResize() {
    this.updateDetailLayout()
  },

  updateDetailLayout() {
    this.setData({
      navLayout: getWhiteDetailLayout()
    })
  },

  onBack() {
    const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    wx.redirectTo({
      url: `/${ROUTES.gameHall}`
    })
  },

  toggleInterest() {
    this.showPendingFeature()
  },

  onMapTap() {
    this.showPendingFeature()
  },

  onStatTap() {
    this.showPendingFeature()
  },

  onToolTap() {
    this.showPendingFeature()
  },

  noop() {},

  onOpenShare() {
    this.setData({
      showShareWindow: true
    })
  },

  onCloseShare() {
    this.setData({
      showShareWindow: false
    })
  },

  onNativeShareTap() {
    this.onCloseShare()
  },

  onShareTimelineTap() {
    this.showInfo('请通过右上角菜单分享到朋友圈')
  },

  onShareDirect() {
    this.onCloseShare()
    wx.navigateTo({
      url: `/${ROUTES.message}?from=gameShare${this.data.gameId ? `&gameId=${encodeURIComponent(this.data.gameId)}` : ''}`
    })
  },

  onPreventTouch() {},

  onPreventBubble() {},

  onEnroll() {
    if (this.data.isGameEnded) {
      this.showInfo(this.data.endedActionText)
      return
    }

    const cachedUser = this.getCachedEnrollUser()

    if (isRealnameVerified(cachedUser)) {
      this.navigateToApply()
      return
    }

    this.showAuthPrompt()
  },

  getCachedEnrollUser() {
    return wx.getStorageSync('enjoy_user') || null
  },

  showAuthPrompt() {
    this.setData({
      authPromptVisible: true
    })
  },

  navigateToApply() {
    const query = this.data.gameId ? `?gameId=${encodeURIComponent(this.data.gameId)}` : ''

    wx.navigateTo({
      url: `/${ROUTES.gameApply}${query}`
    })
  },

  closeAuthPrompt() {
    this.setData({
      authPromptVisible: false
    })
  },

  goRealnameAuth() {
    this.setData({
      authPromptVisible: false
    })

    wx.navigateTo({
      url: '/pages/login/realname/index'
    })
  },

  onViewAllParticipants() {
    const query = this.data.gameId ? `?gameId=${encodeURIComponent(this.data.gameId)}` : ''

    wx.navigateTo({
      url: `/${ROUTES.gameParticipants}${query}`
    })
  },

  onParticipantTap() {
    this.showPendingFeature()
  },

  handleDetailScroll(event) {
    const scrollTop = event.detail && event.detail.scrollTop

    if (typeof scrollTop === 'number') {
      this.detailScrollTopValue = scrollTop
    }
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  },

  showPendingFeature() {
    this.showInfo('功能待开发')
  },

  onShareAppMessage() {
    return {
      title: this.data.event.title,
      path: `/${ROUTES.gameDetail}${this.data.gameId ? `?id=${this.data.gameId}` : ''}`,
      imageUrl: this.data.event.coverSrc
    }
  },

  onShareTimeline() {
    return {
      title: this.data.event.title,
      query: this.data.gameId ? `id=${this.data.gameId}` : '',
      imageUrl: this.data.event.coverSrc
    }
  }
})
