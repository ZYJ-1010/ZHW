const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')
const inviteService = require('../../../services/invite')
const { navigateShellRoute } = require('../../../utils/shell-nav')

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

  const timeText = event.time || event.timeText || event.startTimeText || ''

  return /\d{4}年/.test(String(timeText)) ? parseChineseEventEndTime(timeText) : null
}

function getGameEndedState(event = {}) {
  const endTimestamp = getEventEndTimestamp(event)

  return typeof endTimestamp === 'number' ? endTimestamp <= Date.now() : false
}

function gameStatusText(status = '') {
  const map = {
    pending_audit: '待后台审核',
    recruiting: '招募中',
    full: '已满员',
    in_progress: '进行中',
    pending_confirm: '待确认完成',
    pending_review: '待评价',
    completed: '已完成'
  }

  return map[status] || status || '招募中'
}

function gameTypeText(type = '') {
  const map = {
    free: '免费局',
    standard: '标准局',
    public_welfare: '公益局',
    aa: 'AA局',
    crowdfund: '众筹局',
    deposit: '押金局',
    condition: '条件局'
  }

  return map[type] || type || '免费局'
}

function formatCreatedAt(value) {
  if (!value) {
    return ''
  }

  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return ''
  }

  const month = date.getMonth() + 1
  const day = date.getDate()
  const hour = String(date.getHours()).padStart(2, '0')
  const minute = String(date.getMinutes()).padStart(2, '0')

  return `${month}月${day}日 ${hour}:${minute}`
}

function makeParticipant(userId, index, game = {}) {
  const isCreator = Number(userId) === Number(game.creatorUserId)
  const isMainGuide = Number(userId) === Number(game.mainGuideUserId)
  const role = isCreator ? '发起人' : isMainGuide ? '主行家' : '成员'
  const roleClass = isCreator || isMainGuide ? 'guide' : 'player'

  return {
    id: userId || `member-${index + 1}`,
    userId,
    name: userId ? `成员${userId}` : `成员${index + 1}`,
    avatarSrc: '/pages/home/player/assets/ranking-avatar-01.png',
    avatarText: String(userId || index + 1).slice(-2),
    role,
    roleClass,
    position: role,
    topic: '参与本次组局',
    primaryTag: isCreator ? '组局发起人' : isMainGuide ? '主行家' : '组局成员',
    tags: [],
    location: game.cityName || '同城组局',
    distance: game.distanceLabel || ''
  }
}

function normalizeParticipants(data = {}, game = {}) {
  const ids = Array.isArray(data.memberIds) ? data.memberIds : []

  if (ids.length) {
    return ids.slice(0, Number(game.maxPlayers || 8)).map((id, index) => makeParticipant(id, index, game))
  }

  if (game.creatorUserId) {
    return [makeParticipant(game.creatorUserId, 0, game)]
  }

  return []
}

function buildOrganizer(game = {}) {
  const userId = game.creatorUserId || ''

  return {
    name: userId ? `发起人 ${userId}` : '组局发起人',
    avatarSrc: '/pages/home/player/assets/ranking-avatar-01.png',
    avatarText: userId ? String(userId).slice(-2) : '发',
    role: game.mainGuideUserId ? `主行家 ${game.mainGuideUserId}` : '组局发起人',
    summary: `人数 ${Number(game.currentPlayers || 0)}/${Number(game.maxPlayers || 8)}`,
    rating: '--'
  }
}

function compactList(items) {
  return items.map((item) => String(item || '').trim()).filter(Boolean)
}

function normalizeGameDetailPayload(data = {}, fallbackEvent = {}) {
  const game = data.game || data
  const memberIds = Array.isArray(data.memberIds) ? data.memberIds : []
  const currentPlayers = Number(game.currentPlayers || memberIds.length || 0)
  const minPlayers = Number(game.minPlayers || 5)
  const maxPlayers = Number(game.maxPlayers || 8)
  const cityName = String(game.cityName || '').trim()
  const address = String(game.address || '').trim()
  const title = String(game.title || fallbackEvent.title || '').trim()
  const statusText = gameStatusText(game.status)
  const gameType = gameTypeText(game.gameType)
  const categoryText = String(game.primaryCategoryText || game.secondaryCategoryText || '').trim()
  const createdAtText = formatCreatedAt(game.createdAt)

  return {
    event: Object.assign({}, fallbackEvent, {
      title,
      location: address || cityName || fallbackEvent.location,
      category: categoryText || gameType,
      fee: game.gameType === 'free' ? '免费局' : gameType,
      time: statusText
    }),
    stats: [
      { iconText: '状态', text: statusText, action: 'status' },
      { iconText: '评价', text: data.review && data.review.complete ? '评价已完成' : '待评价', action: 'reviews' },
      { iconSrc: '/pages/game/assets/icons/icon-participants.svg', text: `${currentPlayers}/${maxPlayers}人已报名` }
    ],
    tags: compactList([gameType, statusText, categoryText, cityName]).map((name, index) => ({
      name: `#${name}`,
      tone: ['blue', 'green', 'purple'][index % 3]
    })),
    organizer: buildOrganizer(game),
    introduction: title ? `${title}。${cityName || address ? `地点：${address || cityName}。` : ''}` : '暂无组局介绍',
    highlights: compactList([
      categoryText ? `分类：${categoryText}` : '',
      `人数规则：最少${minPlayers}人，最多${maxPlayers}人`,
      game.mainGuideUserId ? `主行家：${game.mainGuideUserId}` : '',
      statusText ? `当前状态：${statusText}` : ''
    ]),
    schedule: createdAtText ? [
      { title: '组局发布', time: createdAtText, desc: '后台审核通过后进入报名和组局流程。' }
    ] : [],
    detailImages: [],
    noticeLead: '请按平台规则参与组局。',
    noticeBullets: [
      `人数限制：${minPlayers}-${maxPlayers}人，未满${minPlayers}人不能开始，满${maxPlayers}人后不可继续报名。`,
      '领路人和行家需要完成实名认证后参与对应身份流程。',
      '请以平台内报名、审核、确认和评价流程为准。'
    ],
    audience: categoryText ? `适合关注${categoryText}的用户参与。` : '适合符合本局条件的用户参与。',
    participants: normalizeParticipants(data, game),
    isGameEnded: ['pending_review', 'completed'].indexOf(game.status) !== -1,
    endedActionText: statusText,
    endedNoticeText: game.status === 'pending_audit' ? '新建组局正在后台审核，审核通过后对外展示' : '服务状态已更新，请按流程继续处理'
  }
}

Page({
  data: {
    gameId: '',
    entryIntent: '',
    interested: false,
    authPromptVisible: false,
    showShareWindow: false,
    shareEntry: null,
    detailScrollTop: 0,
    navLayout: getWhiteDetailLayout(),
    event: {
      coverSrc: '/pages/game/hall/assets/hall-featured-city.jpg',
      title: '',
      time: '',
      location: '',
      category: '',
      categoryIcon: '/pages/game/assets/icons/icon-social-handshake.svg',
      fee: ''
    },
    isGameEnded: false,
    endedActionText: '????',
    endedNoticeText: '',
    stats: [],
    tags: [],
    organizer: {
      name: '',
      avatarSrc: '/pages/home/player/assets/ranking-avatar-01.png',
      avatarText: '',
      role: '',
      summary: '',
      rating: '--'
    },
    introduction: '??????',
    highlights: [],
    schedule: [],
    detailImages: [],
    noticeLead: '???????????',
    noticeBullets: [],
    audience: '',
    participants: []
  },

  onLoad(options = {}) {
    this.saveInviteEntryContext(options)

    this.setData({
      gameId: options.gameId || options.id || '',
      entryIntent: options.intent || '',
      isGameEnded: getGameEndedState(this.data.event),
      navLayout: getWhiteDetailLayout()
    })

    this.loadGameDetail()

    if (wx.showShareMenu) {
      wx.showShareMenu({
        withShareTicket: true,
        menus: ['shareAppMessage', 'shareTimeline']
      })
    }
  },

  saveInviteEntryContext(options = {}) {
    const inviteCode = String(options.inviteCode || options.code || '').trim().toUpperCase()

    if (!inviteCode) {
      return
    }

    inviteService.saveInviteContext({
      code: inviteCode,
      entryType: String(options.entryType || '').trim(),
      gameId: options.gameId || options.id || '',
      source: 'game_detail_share'
    })
  },

  async loadGameDetail() {
    if (!this.data.gameId) {
      return
    }

    try {
      const detail = await gameService.getGameDetail(this.data.gameId)
      this.setData(normalizeGameDetailPayload(detail, this.data.event))
      this.ensureShareEntry()
      this.showEntryIntentHint()
    } catch (error) {
      this.showInfo(error.message || '局详情加载失败')
    }
  },

  async ensureShareEntry() {
    if (!this.data.gameId || this.data.shareEntry) {
      return this.data.shareEntry
    }

    try {
      const shareEntry = await gameService.createInviteEntry({
        entryType: 'link',
        gameId: Number(this.data.gameId),
        title: this.data.event.title || '真好玩组局邀请'
      })

      this.setData({ shareEntry })
      return shareEntry
    } catch (error) {
      return null
    }
  },

  showEntryIntentHint() {
    if (this.data.entryIntent !== 'join' || this.joinIntentHintShown) {
      return
    }

    this.joinIntentHintShown = true
    this.showInfo('已进入组队详情，可点击底部按钮提交报名')
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

    navigateShellRoute(ROUTES.gameHall, {
      currentRoute: ROUTES.gameDetail
    })
  },

  async toggleInterest() {
    if (!this.data.gameId) {
      this.setData({ interested: true })
      this.showInfo('已标记感兴趣')
      return
    }

    try {
      await gameService.favoriteGame(this.data.gameId)
      this.setData({ interested: true })
      this.showInfo('已加入感兴趣')
    } catch (error) {
      this.showInfo(error.message || '收藏失败')
    }
  },

  onMapTap() {
    const query = this.data.gameId ? `?gameId=${encodeURIComponent(this.data.gameId)}&mode=route` : ''

    navigateShellRoute(`${ROUTES.map}${query}`, {
      currentRoute: ROUTES.gameDetail
    })
  },

  onStatTap(event) {
    const action = event.currentTarget.dataset.action

    if (action === 'reviews') {
      navigateShellRoute(`${ROUTES.gameReview}${this.data.gameId ? `?gameId=${encodeURIComponent(this.data.gameId)}` : ''}`, {
        currentRoute: ROUTES.gameDetail
      })
      return
    }

    if (action === 'status') {
      this.showInfo(this.data.endedActionText || this.data.event.time)
      return
    }

    this.onViewAllParticipants()
  },

  onToolTap(event) {
    const action = event.currentTarget.dataset.action
    const gameIdQuery = this.data.gameId ? `?gameId=${encodeURIComponent(this.data.gameId)}` : ''

    if (action === '引荐') {
      navigateShellRoute(`${ROUTES.gameInvite}${gameIdQuery}`, {
        currentRoute: ROUTES.gameDetail
      })
      return
    }

    if (action === '签到') {
      navigateShellRoute(`${ROUTES.mapRealCheckin || 'pages/map/real-checkin/index'}${gameIdQuery}`, {
        currentRoute: ROUTES.gameDetail
      })
      return
    }

    if (action === '打招呼') {
      navigateShellRoute(`${ROUTES.gameGreet}${gameIdQuery}`, {
        currentRoute: ROUTES.gameDetail
      })
      return
    }

    this.showInfo('暂无可执行操作')
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
    navigateShellRoute(`${ROUTES.message}?from=gameShare${this.data.gameId ? `&gameId=${encodeURIComponent(this.data.gameId)}` : ''}`, {
      currentRoute: ROUTES.gameDetail
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

    navigateShellRoute(`${ROUTES.gameApply}${query}`, {
      currentRoute: ROUTES.gameDetail
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

    navigateShellRoute('/pages/login/realname/index', {
      currentRoute: ROUTES.gameDetail
    })
  },

  onViewAllParticipants() {
    const query = this.data.gameId ? `?gameId=${encodeURIComponent(this.data.gameId)}` : ''

    navigateShellRoute(`${ROUTES.gameParticipants}${query}`, {
      currentRoute: ROUTES.gameDetail
    })
  },

  onParticipantTap() {
    this.onViewAllParticipants()
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
    this.showInfo('暂无可执行操作')
  },

  onShareAppMessage() {
    const entry = this.data.shareEntry || {}
    const inviteCode = entry.inviteCode || ''
    const entryType = entry.entryType || 'link'
    const fallbackPath = `/${ROUTES.gameDetail}${this.data.gameId ? `?id=${encodeURIComponent(this.data.gameId)}` : ''}`
    const path = entry.path
      ? entry.path.replace(/^\/+/, '/')
      : `${fallbackPath}${inviteCode ? `&inviteCode=${encodeURIComponent(inviteCode)}&entryType=${encodeURIComponent(entryType)}` : ''}`

    return {
      title: entry.title || this.data.event.title,
      path,
      imageUrl: this.data.event.coverSrc
    }
  },

  onShareTimeline() {
    const entry = this.data.shareEntry || {}
    const inviteCode = entry.inviteCode || ''
    const entryType = entry.entryType || 'link'
    const query = [
      this.data.gameId ? `id=${encodeURIComponent(this.data.gameId)}` : '',
      inviteCode ? `inviteCode=${encodeURIComponent(inviteCode)}` : '',
      entryType ? `entryType=${encodeURIComponent(entryType)}` : ''
    ].filter(Boolean).join('&')

    return {
      title: entry.title || this.data.event.title,
      query,
      imageUrl: this.data.event.coverSrc
    }
  }
})
