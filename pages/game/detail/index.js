const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')

const DEFAULT_CONTENT_TOP_RPX = 160
const NAV_BOTTOM_GAP_RPX = 18
const NAV_TITLE_HEIGHT_RPX = 50
const NAV_BUTTON_SIZE_RPX = 44
const BOTTOM_ACTION_RPX = 148
const DEFAULT_CAPSULE_BOTTOM_RPX = 142
const DEFAULT_SHARE_RIGHT_RPX = 206
const SHARE_CAPSULE_GAP_RPX = 18
const DEFAULT_CATEGORY_ICON = '/pages/game/assets/icons/icon-social-handshake.svg'
const PARTICIPANTS_ICON = '/pages/game/assets/icons/icon-participants.svg'

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

function parseChineseEventEndTime(text = '') {
  const matched = String(text).match(/(?:(\d{4})年)?(\d{1,2})月(\d{1,2})日(?:[^\d]*(\d{1,2}):(\d{2}))?(?:\s*(?:-|--|—|~|～|至)\s*(?:(\d{1,2})月(\d{1,2})日\s*)?(\d{1,2}):(\d{2}))?/)

  if (!matched) {
    return null
  }

  if (!matched[1]) {
    return null
  }

  const year = Number(matched[1])
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

  const timeText = event.time || event.timeText || event.startTimeText || ''

  return /\d{4}年/.test(String(timeText)) ? parseChineseEventEndTime(timeText) : null
}

function getBackendTimestamp(value) {
  if (!value) {
    return null
  }

  const timestamp = Date.parse(value)

  return Number.isNaN(timestamp) ? null : timestamp
}

function getGameEndedState(event = {}, backendTime) {
  const endTimestamp = getEventEndTimestamp(event)
  const backendTimestamp = getBackendTimestamp(pickFirstValue(
    event.serverTime,
    event.currentTime,
    event.now,
    backendTime
  ))

  return typeof endTimestamp === 'number' && typeof backendTimestamp === 'number'
    ? endTimestamp <= backendTimestamp
    : false
}

function pickFirstValue(...values) {
  return values.find((value) => value !== undefined && value !== null && value !== '')
}

function asArray(value) {
  return Array.isArray(value) ? value : []
}

function formatDateTimeText(value) {
  if (!value) {
    return ''
  }

  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return String(value)
  }

  const month = date.getMonth() + 1
  const day = date.getDate()
  const hour = String(date.getHours()).padStart(2, '0')
  const minute = String(date.getMinutes()).padStart(2, '0')

  return `${date.getFullYear()}年${month}月${day}日 ${hour}:${minute}`
}

function formatTimeRange(item = {}) {
  const text = pickFirstValue(item.time, item.timeText, item.startTimeText)

  if (text) {
    return text
  }

  const start = formatDateTimeText(item.startAt || item.startTime)
  const end = formatDateTimeText(item.endAt || item.endTime)

  return end ? `${start}-${end}` : start
}

function formatFeeText(item = {}) {
  const text = pickFirstValue(item.fee, item.feeText, item.priceText, item.costText)

  if (text) {
    return text
  }

  const amount = Number(pickFirstValue(item.feeAmount, item.priceAmount, item.costAmount))

  return Number.isFinite(amount) ? `场地费${amount}/位` : ''
}

function normalizeParticipant(member = {}) {
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
    tags: asArray(member.tags || user.tags),
    location: member.location || member.address || user.location || '',
    distance: member.distanceText || member.distance || ''
  }
}

function normalizeTag(tag, index) {
  if (typeof tag === 'string') {
    return {
      name: tag.charAt(0) === '#' ? tag : `#${tag}`,
      tone: ['blue', 'green', 'purple'][index % 3]
    }
  }

  return {
    name: tag.name || tag.label || '',
    tone: tag.tone || ['blue', 'green', 'purple'][index % 3]
  }
}

function normalizeOrganizer(source = {}) {
  const creator = source.creator || source.organizer || source.owner || {}
  const name = creator.name || creator.nickname || creator.realname || ''

  return {
    name,
    avatarSrc: creator.avatarSrc || creator.avatarUrl || '',
    avatarText: creator.avatarText || name.slice(0, 1),
    role: creator.role || creator.title || creator.company || '',
    summary: creator.summary || creator.statText || '',
    rating: String(creator.rating || creator.score || '')
  }
}

function normalizeGameDetail(data = {}) {
  const serverTime = pickFirstValue(data.serverTime, data.currentTime, data.now, data.responseTime)
  const memberCount = pickFirstValue(data.approvedMemberCount, data.memberCount, data.joinedCount)
  const maxParticipants = data.maxParticipants || data.maxMemberCount
  const membersText = memberCount != null && maxParticipants ? `${memberCount}/${maxParticipants}人已报名` : ''
  const commentsText = data.commentCount != null ? `${data.commentCount}条评价` : ''
  const viewsText = data.viewCount != null ? `${data.viewCount}次浏览` : ''
  const members = asArray(data.members || data.participants).map(normalizeParticipant)

  return {
    event: {
      coverSrc: data.coverSrc || data.coverUrl || data.coverFileUrl || '',
      title: data.title || '',
      time: formatTimeRange(data),
      location: data.addressName || data.address || data.locationName || '',
      category: data.gameTypeText || data.categoryText || data.typeText || '',
      categoryIcon: data.categoryIcon || DEFAULT_CATEGORY_ICON,
      fee: formatFeeText(data),
      endAt: data.endAt || data.endTime,
      registrationEndAt: data.registrationEndAt || data.applyEndAt,
      serverTime
    },
    serverTime,
    stats: [
      viewsText ? { iconText: '👁️', text: viewsText, action: 'views' } : null,
      commentsText ? { iconText: '💬', text: commentsText, action: 'reviews' } : null,
      membersText ? { iconSrc: PARTICIPANTS_ICON, text: membersText } : null
    ].filter(Boolean),
    tags: asArray(data.themeTags || data.tags).map(normalizeTag).filter((item) => item.name),
    organizer: normalizeOrganizer(data),
    introduction: data.introduction || data.description || data.summary || '',
    highlights: asArray(data.highlights),
    schedule: asArray(data.schedule || data.agenda),
    detailImages: asArray(data.detailImages || data.images || data.imageUrls),
    noticeLead: data.noticeLead || '',
    noticeBullets: asArray(data.noticeBullets || data.notices),
    audience: data.audience || data.targetAudience || '',
    participants: members
  }
}

Page({
  data: {
    gameId: '',
    interested: false,
    authPromptVisible: false,
    showShareWindow: false,
    detailScrollTop: 0,
    navLayout: getWhiteDetailLayout(),
    loading: false,
    loadErrorText: '',
    event: {
      coverSrc: '',
      title: '',
      time: '',
      location: '',
      category: '',
      categoryIcon: DEFAULT_CATEGORY_ICON,
      fee: ''
    },
    isGameEnded: false,
    endedActionText: '报名结束',
    endedNoticeText: '新建组局将经过平台审核，审核通过后才能正式发布',
    stats: [],
    tags: [],
    organizer: {
      name: '',
      avatarSrc: '',
      avatarText: '',
      role: '',
      summary: '',
      rating: ''
    },
    introduction: '',
    highlights: [],
    schedule: [],
    detailImages: [],
    noticeLead: '',
    noticeBullets: [],
    audience: '',
    participants: []
  },

  onLoad(options = {}) {
    const gameId = options.gameId || options.id || ''

    this.setData({
      gameId,
      isGameEnded: false,
      navLayout: getWhiteDetailLayout()
    })

    if (gameId) {
      this.loadGameDetail(gameId)
    }

    if (wx.showShareMenu) {
      wx.showShareMenu({
        withShareTicket: true,
        menus: ['shareAppMessage', 'shareTimeline']
      })
    }
  },

  async loadGameDetail(gameId) {
    this.setData({
      loading: true,
      loadErrorText: ''
    })

    try {
      const detail = normalizeGameDetail(await gameService.getGameDetail(gameId))

      this.setData(Object.assign({
        loading: false,
        loadErrorText: ''
      }, detail, {
        isGameEnded: getGameEndedState(detail.event, detail.serverTime)
      }))
    } catch (error) {
      this.setData({
        loading: false,
        loadErrorText: error.message || '局详情加载失败'
      })
      this.showInfo(error.message || '局详情加载失败')
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
