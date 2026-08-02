const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')
const inviteService = require('../../../services/invite')
const userService = require('../../../services/user')
const { navigateShellRoute } = require('../../../utils/shell-nav')
const { toUserMessage } = require('../../../utils/user-message')

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

function gameStatusText(status = '') {
  const map = {
    pending_audit: '待后台审核',
    rejected: '审核未通过',
    recruiting: '招募中',
    full: '已满员',
    in_progress: '进行中',
    pending_confirm: '已结束',
    pending_review: '待评价',
    completed: '已完成',
    canceled: '已取消',
    cancelled: '已取消',
    disputed: '争议中',
    settling: '结算中',
    closed: '已关闭'
  }

  return map[status] || status || '招募中'
}

function gameTypeText(type = '') {
  return '免费局'
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

function normalizeParticipant(item = {}, index, game = {}) {
	const userId = item.userId || item.id || ''
	const roleKey = String(item.role || '').trim().toLowerCase()
	const roleClass = roleKey === 'expert'
		? 'expert'
		: (roleKey === 'guide' || roleKey === 'main_guide' || roleKey === 'guide_escort' ? 'guide' : (roleKey === 'member' || roleKey === 'player' ? 'player' : 'unknown'))
	const role = item.roleLabel || '后台未返回'
	const name = item.displayName || item.name || item.nickname || '后台未返回'

	return {
		id: userId || `member-${index + 1}`,
		userId,
		name,
		avatarSrc: item.avatarUrl || item.avatarSrc || '',
		avatarText: item.avatarText || '未',
		role,
		roleClass,
		position: item.position || '后台未返回',
		topic: item.topic || '后台未返回',
		primaryTag: item.primaryTag || '后台未返回',
		blueBadge: item.expertBlueBadge || { enabled: false },
		tags: Array.isArray(item.tags) ? item.tags : [],
		location: item.location || '后台未返回',
		distance: item.distance || ''
	}
}

function normalizeParticipants(data = {}, game = {}) {
	const members = Array.isArray(data.members)
		? data.members
		: (Array.isArray(data.memberIds) ? data.memberIds.map((userId) => ({ userId })) : [])
	return members
		.slice(0, Number(game.maxPlayers || 8))
		.map((item, index) => normalizeParticipant(item, index, game))
}

function buildOrganizer(game = {}, display = {}) {
	const rating = String(display.rating || '').trim()
  const ratingCount = Number(display.ratingCount || 0)
  const ratingValue = Number(rating)

	return {
		name: display.name || '后台未返回',
		avatarSrc: display.avatarUrl || display.avatarSrc || '',
		avatarText: display.avatarText || '未',
		role: display.roleLabel || '后台未返回',
		blueBadge: display.expertBlueBadge || { enabled: false },
    summary: `人数 ${Number(game.currentPlayers || 0)}/${Number(game.maxPlayers || 8)}`,
    rating,
    ratingCount,
    ratingVisible: ratingCount > 0 && Number.isFinite(ratingValue) && ratingValue > 0
  }
}

function normalizePrimaryAction(detailDisplay = {}, game = {}, statusText = '', relation = {}) {
  const source = detailDisplay.primaryAction || {}
  const relationReason = String(relation.applyDisabledReason || relation.ApplyDisabledReason || '').trim()
  const applicationStatus = String(relation.applicationStatus || relation.ApplicationStatus || '').trim().toLowerCase()
  const applicationId = Number(relation.applicationId || relation.ApplicationID || 0)
  const isCreator = relation.isCreator === true || relation.IsCreator === true
  const isMember = relation.isMember === true || relation.IsMember === true

  if (applicationStatus === 'pending' && applicationId > 0) {
    return { text: '申请中 · 撤回', disabled: false, action: 'withdraw_application', applicationId, route: '', confirmText: '确定撤回本次入局申请？' }
  }

  if (!isCreator && isMember && (game.status === 'recruiting' || game.status === 'full')) {
    return { text: '退出本局', disabled: false, action: 'exit_game', route: '', confirmText: '退出后将释放本局席位，并按后台信用规则扣分，确定退出？' }
  }

  if (source.text) {
    const sourceText = String(source.text)
    const shouldUseRelationReason = source.disabled === true && relationReason &&
      ['等待开局', '等待报名', '状态处理中', '招募中'].indexOf(sourceText) !== -1
    return {
      text: shouldUseRelationReason ? relationReason : sourceText,
      disabled: source.disabled === true,
      action: String(source.action || 'none'),
      route: String(source.route || ''),
      confirmText: String(source.confirmText || ''),
      disabledReason: String(source.disabledReason || (shouldUseRelationReason ? relationReason : '') || '')
    }
  }

  if (game.status === 'pending_audit') {
    return { text: '后台审核中', disabled: true, action: 'none', route: '', confirmText: '' }
  }

  if (game.status === 'rejected') {
    return { text: '审核未通过', disabled: true, action: 'none', route: '', confirmText: '' }
  }

  return {
    text: statusText || '状态处理中',
    disabled: true,
    action: 'none',
    route: '',
    confirmText: ''
  }
}

function buildReviewStat(data = {}, game = {}, relation = {}, primaryAction = {}) {
  const review = data.review || {}
  const status = String(game.status || '').toLowerCase()
  const canOpen = status === 'pending_review' && (
    review.reviewable === true ||
    relation.canReview === true ||
    (primaryAction.action === 'review' && primaryAction.disabled !== true) ||
    (Array.isArray(review.todos) && review.todos.length > 0)
  )
  const complete = review.complete === true || review.completed === true

  return {
    key: 'reviews',
    label: '评价',
    value: complete ? '已完成' : (canOpen ? '待评价' : '未开启'),
    action: 'reviews',
    disabled: !canOpen,
    disabledToast: status !== 'pending_review'
      ? '本局尚未进入评价阶段'
      : (!relation.isMember ? '加入并完成本局后才能评价' : '暂无可评价内容')
  }
}

function compactList(items) {
  return items.map((item) => String(item || '').trim()).filter(Boolean)
}

function normalizeShareComponent(source = {}) {
  return {
    enabled: source.enabled !== false,
    variant: source.variant === 'icon_button' ? 'icon_button' : 'channel_sheet',
    label: source.label || '分享',
    enableInternal: source.enableInternal !== false,
    enableWechat: source.enableWechat !== false
  }
}

function buildBottomTools(relation = {}, game = {}, shareComponent = {}) {
  if (['pending_review', 'completed', 'canceled', 'cancelled'].indexOf(game.status) !== -1) {
    return []
  }

  const role = String(relation.role || '').toLowerCase()
  const isCreator = Boolean(relation.isCreator)
  const isMember = Boolean(relation.isMember)
  const share = normalizeShareComponent(shareComponent)
  const tools = share.enabled ? [
    { key: 'share', text: share.label, iconSrc: '/pages/game/detail/assets/i45@3x.png' }
  ] : []

  if (isMember && ['in_progress', 'pending_confirm'].indexOf(game.status) !== -1) {
    tools.push({ key: 'checkin', text: '签到', iconSrc: '/pages/game/detail/assets/i47@3x.png' })
  }

  if (relation.canEnterIM) {
    tools.push({ key: 'chat', text: '聊天', iconSrc: '/pages/game/detail/assets/i48@3x.png' })
  } else if (!isCreator && game.creatorUserId) {
    tools.push({ key: 'greet', text: '打招呼', iconSrc: '/pages/game/detail/assets/i48@3x.png' })
  }

  return tools
}

function normalizeGameDetailPayload(data = {}, fallbackEvent = {}) {
  const game = data.game || data
  const detailDisplay = data.detailDisplay || {}
  const relation = data.myRelation || {}
  const memberIds = Array.isArray(data.memberIds) ? data.memberIds : []
  const currentPlayers = Number(game.currentPlayers || memberIds.length || 0)
  const minPlayers = Number(game.minPlayers || 5)
  const maxPlayers = Number(game.maxPlayers || 8)
  const cityName = String(game.cityName || '').trim()
  const address = String(game.address || '').trim()
  const title = String(game.title || fallbackEvent.title || '').trim()
  const statusText = String(detailDisplay.statusText || gameStatusText(game.status))
  const gameType = gameTypeText(game.gameType)
  const categoryText = String(game.primaryCategoryText || game.secondaryCategoryText || '').trim()
	const roleLabels = { player: '玩家', expert: '行家', guide: '领路人' }
	const allowedRoles = Array.isArray(game.allowedRoles) && game.allowedRoles.length ? game.allowedRoles : ['player']
	const allowedRoleLabels = allowedRoles.map((role) => roleLabels[role]).filter(Boolean)
  const createdAtText = formatCreatedAt(game.createdAt)
  const shareComponent = normalizeShareComponent(data.shareComponent || {})
  const bottomTools = buildBottomTools(relation, game, shareComponent)
  const primaryAction = normalizePrimaryAction(detailDisplay, game, statusText, relation)
  const descriptionMedia = Array.isArray(game.descriptionMedia) ? game.descriptionMedia : []
  const detailMedia = descriptionMedia.map((item, index) => ({
    id: item.id || item.fileId || `media-${index}`,
    type: item.type === 'video' ? 'video' : 'image',
    src: String(item.url || item.downloadUrl || item.src || '').trim()
  })).filter((item) => item.src)

  return {
    event: Object.assign({}, fallbackEvent, {
      coverSrc: game.coverImage || game.coverUrl || game.coverSrc || fallbackEvent.coverSrc,
      title,
      location: address || cityName || fallbackEvent.location,
      category: categoryText || gameType,
      fee: game.gameType === 'free' ? '免费局' : gameType,
      time: statusText
    }),
    stats: [
      { key: 'status', label: '状态', value: statusText, action: 'status' },
      buildReviewStat(data, game, relation, primaryAction),
      { key: 'participants', iconSrc: 'https://static.haowan.net.cn/miniprogram/pages/game/assets/icons/icon-participants.svg', value: `${currentPlayers}/${maxPlayers}人已报名` }
    ],
    tags: compactList([gameType, statusText, categoryText, cityName]).map((name, index) => ({
      name: `#${name}`,
      tone: ['blue', 'green', 'purple'][index % 3]
    })),
    allowedRoleLabels,
    organizer: buildOrganizer(game, detailDisplay.organizer || {}),
    introduction: String(game.introduction || '').trim() || (title ? `${title}。${cityName || address ? `地点：${address || cityName}。` : ''}` : '暂无组局介绍'),
    highlights: compactList([
      categoryText ? `分类：${categoryText}` : '',
      `人数规则：最少${minPlayers}人，最多${maxPlayers}人`,
      game.mainGuideUserId ? `主领路人：${game.mainGuideUserId}` : '',
      statusText ? `当前状态：${statusText}` : ''
    ]),
    detailDescription: String(game.description || '').trim(),
    schedule: createdAtText ? [
      { title: '组局发布', time: createdAtText, desc: '后台审核通过后进入报名和组局流程。' }
    ] : [],
    detailImages: detailMedia.filter((item) => item.type === 'image').map((item) => item.src),
    detailMedia,
    noticeLead: '请按平台规则参与组局。',
    noticeBullets: [
      `人数限制：${minPlayers}-${maxPlayers}人，未满${minPlayers}人不能开始，满${maxPlayers}人后不可继续报名。`,
      '领路人和行家需要完成实名认证后参与对应身份流程。',
      '请以平台内报名、审核、确认和评价流程为准。'
    ],
    audience: categoryText ? `适合关注${categoryText}的用户参与。` : '适合符合本局条件的用户参与。',
    participants: normalizeParticipants(data, game),
    primaryAction,
    myRelation: relation,
    game: {
      id: game.id || game.gameId || '',
      title,
      creatorUserId: game.creatorUserId || game.creatorID || 0,
      auditRejectReason: game.auditRejectReason || data.auditRejectReason || ''
    },
    interested: data.isFavorited === true || data.favorited === true || relation.isFavorited === true,
    shareComponent,
    bottomTools,
    showBottomTools: bottomTools.length > 0
  }
}

Page({
  data: {
    gameId: '',
    entryIntent: '',
    interested: false,
    authPromptVisible: false,
    showShareWindow: false,
    shareComponent: normalizeShareComponent({ enabled: false, enableInternal: false, enableWechat: false }),
    shareEntry: null,
    detailScrollTop: 0,
    navLayout: getWhiteDetailLayout(),
    event: {
      coverSrc: 'https://static.haowan.net.cn/miniprogram/assets/game/hall/hall-featured-city.jpg',
      title: '',
      time: '',
      location: '',
      category: '',
      categoryIcon: 'https://static.haowan.net.cn/miniprogram/pages/game/assets/icons/icon-social-handshake.svg',
      fee: ''
    },
    primaryAction: {
      text: '',
      disabled: true,
      action: 'none',
      route: '',
      confirmText: ''
    },
    game: {},
    myRelation: {},
    bottomTools: [],
    showBottomTools: true,
    stats: [],
    tags: [],
    allowedRoleLabels: [],
    organizer: {
      name: '',
      avatarSrc: '',
      avatarText: '',
      role: '',
      summary: '',
      rating: '',
      ratingCount: 0,
      ratingVisible: false
    },
    introduction: '暂无组局介绍',
    highlights: [],
    schedule: [],
    detailDescription: '',
    detailImages: [],
    detailMedia: [],
    noticeLead: '请按平台规则参与组局。',
    noticeBullets: [],
    audience: '',
    participants: []
  },

  onLoad(options = {}) {
    this.saveInviteEntryContext(options)
    this._skipNextShowRefresh = true

    this.setData({
      gameId: options.gameId || options.id || '',
      entryIntent: options.intent || '',
      navLayout: getWhiteDetailLayout()
    })

    this.loadGameDetail()
    this.syncNativeShareMenu(this.data.shareComponent)
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
      const normalized = normalizeGameDetailPayload(detail, this.data.event)
      const nativeShareEnabled = normalized.shareComponent.enabled && normalized.shareComponent.enableWechat
      this.setData(Object.assign({}, normalized, nativeShareEnabled ? {} : {
        shareEntry: null,
        showShareWindow: false
      }))
      this.syncNativeShareMenu(normalized.shareComponent)
      if (nativeShareEnabled) {
        this.ensureShareEntry()
      }
      this.showEntryIntentHint()
    } catch (error) {
      this.showInfo(error.message || '局详情加载失败')
    }
  },

  async ensureShareEntry() {
    const shareComponent = this.data.shareComponent || {}
    if (!this.data.gameId || this.data.shareEntry || shareComponent.enabled === false || shareComponent.enableWechat === false) {
      return this.data.shareEntry
    }

    if (!this._shareEntryPromise) {
      const requestedGameId = String(this.data.gameId)
      this._shareEntryPromise = (async () => {
        try {
          const shareEntry = await gameService.createInviteEntry({
            entryType: 'link',
            gameId: Number(this.data.gameId),
            title: this.data.event.title || '真好玩组局邀请'
          })

          const currentShareComponent = this.data.shareComponent || {}
          if (String(this.data.gameId) !== requestedGameId || currentShareComponent.enabled === false || currentShareComponent.enableWechat === false) {
            return null
          }

          this.setData({ shareEntry })
          return shareEntry
        } catch (error) {
          return null
        } finally {
          this._shareEntryPromise = null
        }
      })()
    }

    return this._shareEntryPromise
  },

  syncNativeShareMenu(shareComponent = {}) {
    if (typeof wx === 'undefined') {
      return
    }

    const nativeShareEnabled = shareComponent.enabled !== false && shareComponent.enableWechat !== false
    if (nativeShareEnabled && typeof wx.showShareMenu === 'function') {
      wx.showShareMenu({
        withShareTicket: true,
        menus: ['shareAppMessage', 'shareTimeline']
      })
      return
    }

    if (!nativeShareEnabled && typeof wx.hideShareMenu === 'function') {
      wx.hideShareMenu({
        menus: ['shareAppMessage', 'shareTimeline']
      })
    }
  },

  showEntryIntentHint() {
    if (this.data.entryIntent === 'join') {
      if (this.joinIntentHintShown) {
        return
      }
      this.joinIntentHintShown = true
      this.showInfo('已进入组队详情，可点击底部按钮提交报名')
      return
    }

    if (this.entryIntentHandled) {
      return
    }

    const tools = this.data.bottomTools || []
    if (this.data.entryIntent === 'greet') {
      const tool = tools.find((item) => item.key === 'greet' || item.key === 'chat')
      if (tool) {
        this.entryIntentHandled = true
        this.onToolTap({ currentTarget: { dataset: { action: tool.key } } })
      }
      return
    }

    if (this.data.entryIntent === 'refer') {
      const tool = tools.find((item) => item.key === 'refer' || item.key === 'invite')
      if (tool) {
        this.entryIntentHandled = true
        this.onToolTap({ currentTarget: { dataset: { action: tool.key } } })
      }
    }
  },

  async onShow() {
    this.updateDetailLayout()

    if (this._skipNextShowRefresh) {
      this._skipNextShowRefresh = false
    } else if (this.data.gameId) {
      await this.loadGameDetail()
    }

    const user = await this.getLatestEnrollUser()

    if (isRealnameVerified(user) && this.data.authPromptVisible) {
      this.setData({ authPromptVisible: false })
    }
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
      const reviewStat = (this.data.stats || []).find((item) => item && item.key === 'reviews') || {}

      if (reviewStat.disabled) {
        this.showInfo(reviewStat.disabledToast || '本局尚未进入评价阶段')
        return
      }

      navigateShellRoute(`${ROUTES.gameReview}${this.data.gameId ? `?gameId=${encodeURIComponent(this.data.gameId)}` : ''}`, {
        currentRoute: ROUTES.gameDetail
      })
      return
    }

    if (action === 'status') {
      this.showInfo(this.data.event.time)
      return
    }

    this.onViewAllParticipants()
  },

  async onToolTap(event) {
    const action = event.currentTarget.dataset.action
    const gameIdQuery = this.data.gameId ? `?gameId=${encodeURIComponent(this.data.gameId)}` : ''

    if (action === 'share') {
      this.onOpenShare()
      return
    }

    if (action === 'greet') {
      this.navigateToCreatorPrivateChat()
      return
    }

    if (action === 'checkin') {
      if (!this.data.myRelation || !this.data.myRelation.isMember) {
        this.showInfo('仅局内成员可签到')
        return
      }

      navigateShellRoute(`${ROUTES.mapRealCheckin || 'pages/map/real-checkin/index'}${gameIdQuery}`, {
        currentRoute: ROUTES.gameDetail
      })
      return
    }

    if (action === 'chat') {
      if (!this.data.myRelation || !this.data.myRelation.canEnterIM) {
        this.showInfo('当前身份不可进入聊天')
        return
      }

      navigateShellRoute(`${ROUTES.imRoom}${gameIdQuery}`, {
        currentRoute: ROUTES.gameDetail
      })
      return
    }

    this.showInfo('暂无可执行操作')
  },

  navigateToCreatorPrivateChat() {
    const creatorUserId = Number(this.data.game && this.data.game.creatorUserId)
    const relation = this.data.myRelation || {}

    if (!Number.isInteger(creatorUserId) || creatorUserId <= 0) {
      this.showInfo('缺少组局者信息')
      return
    }

    if (relation.canEnterIM !== true) {
      this.showInfo(relation.isMember === true ? '局还未开' : '仅局内玩家可用，请先报名')
      return
    }

    const query = [
      'mode=private',
      `targetUserId=${encodeURIComponent(creatorUserId)}`,
      this.data.gameId ? `sourceGameId=${encodeURIComponent(this.data.gameId)}` : '',
      this.data.event && this.data.event.title ? `gameTitle=${encodeURIComponent(this.data.event.title)}` : '',
      this.data.organizer && this.data.organizer.name ? `targetName=${encodeURIComponent(this.data.organizer.name)}` : ''
    ].filter(Boolean).join('&')

    navigateShellRoute(`${ROUTES.messageMy}?${query}`, {
      currentRoute: ROUTES.gameDetail
    })
  },

  noop() {},

  onOpenShare() {
    if (!this.data.shareComponent || this.data.shareComponent.enabled === false) {
      return
    }
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

  async onPrimaryAction() {
    const primaryAction = this.data.primaryAction || {}

    if (primaryAction.disabled || primaryAction.action === 'none') {
      this.showInfo(primaryAction.text || this.data.event.time)
      return
    }

    if (primaryAction.action === 'start') {
      this.confirmStartGame(primaryAction)
      return
    }

    if (primaryAction.action === 'withdraw_application') {
      this.confirmWithdrawApplication(primaryAction)
      return
    }

    if (primaryAction.action === 'exit_game') {
      this.confirmExitGame(primaryAction)
      return
    }

    if (primaryAction.action !== 'apply') {
      if (primaryAction.route) {
        navigateShellRoute(primaryAction.route, {
          currentRoute: ROUTES.gameDetail
        })
      }
      return
    }

    this.navigateToApply(primaryAction.route)
  },

  confirmStartGame(primaryAction = {}) {
    wx.showModal({
      title: '开始组局',
      content: primaryAction.confirmText || '确认开始本局？',
      success: async (result) => {
        if (!result.confirm) {
          return
        }

        wx.showLoading({ title: '开始中', mask: true })
        try {
          await gameService.startGame(this.data.gameId)
          wx.hideLoading()
          navigateShellRoute(`/${ROUTES.imRoom}?gameId=${encodeURIComponent(this.data.gameId)}&role=creator`, {
            currentRoute: ROUTES.gameDetail
          })
        } catch (error) {
          wx.hideLoading()
          this.showInfo(error.message || '开始组局失败')
        }
      }
    })
  },

  confirmWithdrawApplication(primaryAction = {}) {
    const applicationId = Number(primaryAction.applicationId || this.data.myRelation.applicationId || 0)
    if (!applicationId) {
      this.showInfo('未找到可撤回的申请')
      return
    }
    wx.showModal({
      title: '撤回申请',
      content: primaryAction.confirmText || '确定撤回本次入局申请？',
      success: async (result) => {
        if (!result.confirm) return
        wx.showLoading({ title: '撤回中', mask: true })
        try {
          await gameService.cancelGameApplication(applicationId)
          wx.hideLoading()
          this.showInfo('申请已撤回')
          await this.loadGameDetail()
        } catch (error) {
          wx.hideLoading()
          this.showInfo(error.message || '撤回申请失败')
        }
      }
    })
  },

  confirmExitGame(primaryAction = {}) {
    wx.showModal({
      title: '退出本局',
      content: primaryAction.confirmText || '退出后将释放本局席位，并按后台信用规则扣分，确定退出？',
      success: async (result) => {
        if (!result.confirm) return
        wx.showLoading({ title: '退出中', mask: true })
        try {
          await gameService.exitGame(this.data.gameId)
          wx.hideLoading()
          this.showInfo('已退出本局')
          await this.loadGameDetail()
        } catch (error) {
          wx.hideLoading()
          this.showInfo(error.message || '退出本局失败')
        }
      }
    })
  },

  getCachedEnrollUser() {
    return wx.getStorageSync('enjoy_user') || null
  },

  async getLatestEnrollUser() {
    try {
      const user = await userService.getCurrentUser()

      if (user) {
        wx.setStorageSync('enjoy_user', user)
        return user
      }
    } catch (error) {
      // Keep the cached user as a temporary fallback when the profile request fails.
    }

    return this.getCachedEnrollUser()
  },

  showAuthPrompt() {
    this.setData({
      authPromptVisible: true
    })
  },

  navigateToApply(route = '') {
    const query = this.data.gameId ? `?gameId=${encodeURIComponent(this.data.gameId)}` : ''

    navigateShellRoute(route || `${ROUTES.gameApply}${query}`, {
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

  onParticipantTap(event) {
    const participant = event.detail && event.detail.participant
    const memberId = participant && (participant.userId || participant.id)

    if (memberId) {
      navigateShellRoute(`/pages/profile/service-center/invite/member-detail/index?id=${encodeURIComponent(memberId)}`, {
        currentRoute: ROUTES.gameDetail
      })
      return
    }

    this.onViewAllParticipants()
  },

  onOrganizerBlueBadgeTap() {
    const badge = this.data.organizer && this.data.organizer.blueBadge
    wx.showModal({
      title: (badge && badge.label) || '蓝标认证',
      content: [badge && badge.ruleText, badge && badge.contactText].filter(Boolean).join('\n'),
      showCancel: false,
      confirmText: '我知道了'
    })
  },

  handleDetailScroll(event) {
    const scrollTop = event.detail && event.detail.scrollTop

    if (typeof scrollTop === 'number') {
      this.detailScrollTopValue = scrollTop
    }
  },

  showInfo(title) {
    wx.showToast({
      title: toUserMessage(title),
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
