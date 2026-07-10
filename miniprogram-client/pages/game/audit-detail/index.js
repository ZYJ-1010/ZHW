const { ROUTES } = require('../../../config/routes')
const { navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')
const fileService = require('../../../services/file')
const gameService = require('../../../services/game')
const toast = require('../../../utils/toast')
const { getSurnameInitials } = require('../../../utils/avatar')

const DETAIL_SCROLL_TAP_STEP_RPX = 360
const DETAIL_SCROLL_HOLD_STEP_RPX = 72
const DETAIL_SCROLL_HOLD_INTERVAL_MS = 80
const DETAIL_SCROLL_HOLD_SUPPRESS_TAP_MS = 120
const IMAGE_EXTENSIONS = ['jpg', 'jpeg', 'png', 'gif', 'webp']
const PDF_EXTENSIONS = ['pdf']

const EMPTY_PLAYER = {
  requirementConfirmed: false,
  requirementStatusText: '',
  confirmed: false,
  statusText: '',
  name: '',
  avatarText: '',
  desc: '',
  tags: [],
  needText: '',
  expectedTime: '',
  remark: ''
}

const EMPTY_GUIDE = {
  iconSrc: '/pages/game/guide-chat/assets/icon-invite.png',
  online: false,
  avatarText: '',
  name: '',
  recommendation: ''
}

const EMPTY_GAME_INFO = {
  topic: '',
  time: '',
  location: '',
  activityType: '',
  serviceDuration: '',
  clientBudget: ''
}
const EMPTY_DETAIL_CONFIG = {
  pageTitle: '',
  referralText: '',
  statusTitles: {},
  countdownTexts: {},
  playerStatusTexts: {},
  texts: {},
  sessionItems: [],
  confirmRows: [],
  optionalActions: [],
  noticeBullets: []
}

function applyTemplate(template, values = {}) {
  let text = String(template || '')

  Object.keys(values).forEach((key) => {
    text = text.replace(new RegExp(`\\{${key}\\}`, 'g'), String(values[key]))
  })

  return text
}

function normalizeDetailConfig(source = {}) {
  const detail = source.auditPage && source.auditPage.detail || source.detail || source

  return {
    pageTitle: String(detail.pageTitle || ''),
    referralText: String(detail.referralText || ''),
    statusTitles: detail.statusTitles || {},
    countdownTexts: detail.countdownTexts || {},
    playerStatusTexts: detail.playerStatusTexts || {},
    texts: detail.texts || {},
    sessionItems: Array.isArray(detail.sessionItems) ? detail.sessionItems : [],
    confirmRows: Array.isArray(detail.confirmRows) ? detail.confirmRows : [],
    optionalActions: Array.isArray(detail.optionalActions) ? detail.optionalActions : [],
    noticeBullets: Array.isArray(detail.noticeBullets) ? detail.noticeBullets : []
  }
}

function normalizeStatus(status) {
  const value = String(status || '').toLowerCase()

  if (value === 'approved' || value === 'pass' || value === 'passed') {
    return 'approved'
  }

  if (value === 'rejected' || value === 'reject') {
    return 'rejected'
  }

  return 'pending'
}

function firstValue() {
  const values = Array.prototype.slice.call(arguments)

  for (let index = 0; index < values.length; index += 1) {
    if (values[index] !== undefined && values[index] !== null && values[index] !== '') {
      return values[index]
    }
  }

  return ''
}

function getFileSource(file = {}) {
  return file.name || file.fileName || file.tempFilePath || file.path || ''
}

function getFileName(file = {}) {
  const source = getFileSource(file)
  const parts = source.split(/[\\/]/)

  return parts[parts.length - 1] || ''
}

function getFileExtension(file = {}) {
  const name = getFileName(file)
  const matched = name.match(/\.([a-zA-Z0-9]+)(?:\?|#)?$/)

  return matched ? matched[1].toLowerCase() : ''
}

function normalizeUploadFile(file = {}) {
  return {
    name: getFileName(file),
    path: file.tempFilePath || file.path || '',
    size: file.size || 0,
    extension: getFileExtension(file)
  }
}

function isAllowedImageFile(file = {}) {
  const extension = getFileExtension(file)

  if (extension) {
    return IMAGE_EXTENSIONS.includes(extension)
  }

  return file.fileType === 'image' || file.type === 'image'
}

function isAllowedPdfFile(file = {}) {
  return PDF_EXTENSIONS.includes(getFileExtension(file))
}

function sessionValue(gameInfo, key) {
  if (key === 'topic') return gameInfo.topic
  if (key === 'time') return gameInfo.time
  if (key === 'location') return gameInfo.location
  return ''
}

function buildSessionInfo(gameInfo, config = EMPTY_DETAIL_CONFIG) {
  return (config.sessionItems || []).map((item) => ({
    label: item.label || '',
    value: sessionValue(gameInfo, item.key),
    iconText: item.iconText || '',
    iconSrc: item.iconSrc || '',
    iconClass: item.iconClass || '',
    actionText: item.actionText || ''
  }))
}

function confirmRowValue(gameInfo, settlement, key) {
  if (key === 'activityType') return gameInfo.activityType
  if (key === 'serviceDuration') return gameInfo.serviceDuration
  if (key === 'clientBudget') return gameInfo.clientBudget
  if (key === 'platformFee') return settlement.platformFee || ''
  if (key === 'guideReward') return settlement.guideReward || ''
  if (key === 'partnerReward') return settlement.partnerReward || ''
  if (key === 'expertIncome') return settlement.expertIncome || ''
  return ''
}

function buildConfirmRows(gameInfo, settlement = {}, config = EMPTY_DETAIL_CONFIG) {
  return (config.confirmRows || []).map((item, index) => {
    const value = confirmRowValue(gameInfo, settlement, item.key)

    return {
      label: item.label || '',
      value,
      highlight: item.key === 'clientBudget' && Boolean(value),
      divider: item.key === 'clientBudget' || item.key === 'partnerReward' || Boolean(item.divider),
      success: item.key === 'expertIncome' && Boolean(value),
      total: item.key === 'expertIncome' || Boolean(item.total),
      _index: index
    }
  })
}

function normalizeApplicationDetail(item = {}, detailConfig = EMPTY_DETAIL_CONFIG) {
  const rawId = firstValue(item.id, item.applicationId)
  const hasDetail = Boolean(rawId)
  const status = normalizeStatus(item.status || item.statusKey)
  const nickname = firstValue(item.nickname, item.userName, item.userNickname, item.user && item.user.nickname)
  const reason = String(firstValue(item.reason, item.remark, item.applyReason)).trim()
  const texts = detailConfig.texts || {}
  const playerStatusTexts = detailConfig.playerStatusTexts || {}
  const gameInfo = {
    ...EMPTY_GAME_INFO,
    topic: firstValue(item.gameTitle, item.title, item.game && item.game.title),
    time: firstValue(item.gameTimeText, item.timeText, item.game && item.game.timeText),
    location: firstValue(item.locationText, item.locationName, item.game && item.game.locationName),
    activityType: firstValue(item.activityType, item.gameTypeText, item.gameType),
    serviceDuration: firstValue(item.serviceDurationText, item.durationText),
    clientBudget: firstValue(item.clientBudgetText, item.budgetText, item.amountText)
  }
  const guideName = firstValue(item.guideName, item.referrerName, item.guide && item.guide.name)
  const guide = {
    ...EMPTY_GUIDE,
    online: Boolean(item.guideOnline || item.referrerOnline),
    avatarText: guideName ? getSurnameInitials(guideName, texts.guideAvatarFallback || '') : '',
    name: guideName,
    recommendation: firstValue(item.guideRecommendation, item.recommendation)
  }
  const player = {
    ...EMPTY_PLAYER,
    requirementConfirmed: status !== 'pending',
    requirementStatusText: status === 'pending' ? playerStatusTexts.pendingRequirement : playerStatusTexts.reviewedRequirement,
    confirmed: status !== 'rejected',
    statusText: playerStatusTexts[status] || '',
    name: nickname,
    avatarText: getSurnameInitials(nickname, texts.playerAvatarFallback || ''),
    desc: firstValue(item.userDesc, item.desc, item.profileText),
    tags: Array.isArray(item.tags) ? item.tags : [],
    needText: reason ? `${texts.needPrefix || ''}${reason}` : '',
    expectedTime: firstValue(item.expectedTimeText, item.expectedTime),
    remark: item.createdAt ? `${texts.remarkPrefix || ''}${item.createdAt}` : ''
  }
  const settlement = item.settlement || item.backendSettlement || {}

  return {
    hasDetail,
    auditId: rawId,
    gameId: item.gameId || '',
    status: {
      title: detailConfig.statusTitles[status] || '',
      quote: reason,
      guideName: guide.name,
      countdown: detailConfig.countdownTexts[status] || ''
    },
    relation: {
      title: texts.relationTitle || '',
      totalCount: Number(item.totalCount || item.memberCount || 0),
      confirmedCount: Number(item.confirmedCount || 0),
      expert: {
        avatarText: texts.expertAvatarText || '',
        name: texts.expertName || '',
        roleText: texts.expertRoleText || ''
      },
      guide,
      player: {
        avatarText: player.avatarText,
        name: player.name,
        confirmed: player.confirmed,
        statusText: player.statusText
      }
    },
    game: {
      player,
      info: gameInfo,
      guide
    },
    backendSettlement: settlement,
    sessionInfo: buildSessionInfo(gameInfo, detailConfig),
    confirmRows: buildConfirmRows(gameInfo, settlement, detailConfig)
  }
}

function buildEmptyDetail(detailConfig = EMPTY_DETAIL_CONFIG) {
  return Object.assign(normalizeApplicationDetail({}, detailConfig), { hasDetail: false })
}

Page({
  data: {
    auditId: '',
    hasDetail: false,
    actionLoading: false,
    uploadedFileIds: [],
    onlineText: '在线',
    detailScrollTop: 0,
    detailConfig: EMPTY_DETAIL_CONFIG,
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ],
    ...buildEmptyDetail(EMPTY_DETAIL_CONFIG),
    optionalActions: [],
    noticeBullets: []
  },

  onLoad(options = {}) {
    const auditId = options.auditId || options.id || ''

    this.setData({ auditId })
    this.loadDetailConfig()
      .then(() => this.loadApplicationDetail(auditId))
  },

  async loadDetailConfig() {
    try {
      const config = await gameService.getApplicationConfig()
      const detailConfig = normalizeDetailConfig(config)
      this.setData({
        detailConfig,
        optionalActions: detailConfig.optionalActions,
        noticeBullets: detailConfig.noticeBullets,
        ...buildEmptyDetail(detailConfig)
      })
    } catch (error) {
      toast.info(error.message || this.textOf('loadFailedText'))
    }
  },

  textOf(key, values = {}) {
    return applyTemplate(this.data.detailConfig && this.data.detailConfig.texts && this.data.detailConfig.texts[key], values)
  },

  async loadApplicationDetail(auditId = this.data.auditId) {
    if (!auditId) {
      return
    }

    try {
      const data = await gameService.getReceivedApplications({})
      const items = data.items || []
      const application = items.find((item) => String(item.id || item.applicationId || '') === String(auditId))

      if (!application) {
        toast.info(this.textOf('detailMissingText'))
        this.setData({
          ...buildEmptyDetail(this.data.detailConfig),
          auditId,
          hasDetail: false
        })
        return
      }

      this.setData(normalizeApplicationDetail(application, this.data.detailConfig))
    } catch (error) {
      toast.info(error.message || this.textOf('loadFailedText'))
    }
  },

  handleMapTap() {
    const params = []
    const gameId = this.data.gameId || ''
    const title = this.data.game && this.data.game.info ? this.data.game.info.topic : ''

    if (gameId) {
      params.push(`gameId=${encodeURIComponent(gameId)}`)
    }

    if (title) {
      params.push(`title=${encodeURIComponent(title)}`)
    }

    navigateShellRoute(`/${ROUTES.map}${params.length ? `?${params.join('&')}` : ''}`)
  },

  handleUploadImageTap() {
    if (wx.chooseMedia) {
      wx.chooseMedia({
        count: 1,
        mediaType: ['image'],
        sourceType: ['album', 'camera'],
        success: (result) => {
          const file = result.tempFiles && result.tempFiles[0]
          this.uploadAuditFile(file, 'image')
        },
        fail: (error) => {
          this.handleChooseFileFail(error)
        }
      })
      return
    }

    if (wx.chooseImage) {
      wx.chooseImage({
        count: 1,
        sourceType: ['album', 'camera'],
        success: (result) => {
          const file = (result.tempFiles && result.tempFiles[0]) || {
            tempFilePath: result.tempFilePaths && result.tempFilePaths[0],
            fileType: 'image'
          }
          this.uploadAuditFile(file, 'image')
        },
        fail: (error) => {
          this.handleChooseFileFail(error)
        }
      })
      return
    }

    toast.info(this.textOf('mediaUnsupportedText'))
  },

  handleUploadFileTap() {
    if (!wx.chooseMessageFile) {
      toast.info(this.textOf('fileUnsupportedText'))
      return
    }

    wx.chooseMessageFile({
      count: 1,
      type: 'file',
      extension: PDF_EXTENSIONS,
      success: (result) => {
        const file = result.tempFiles && result.tempFiles[0]
        this.uploadAuditFile(file, 'file')
      },
      fail: (error) => {
        this.handleChooseFileFail(error)
      }
    })
  },

  async uploadAuditFile(file, type) {
    if (!file) {
      return
    }

    if (type === 'image' && !isAllowedImageFile(file)) {
      toast.info(this.textOf('imageTypeErrorText'))
      return
    }

    if (type === 'file' && !isAllowedPdfFile(file)) {
      toast.info(this.textOf('fileTypeErrorText'))
      return
    }

    const uploadFile = normalizeUploadFile(file)
    if (!uploadFile.path) {
      toast.info(this.textOf('filePathInvalidText'))
      return
    }

    wx.showLoading({
      title: this.textOf('uploadingText'),
      mask: true
    })

    try {
      const fileIds = await fileService.uploadEvidenceImages([uploadFile.path], {
        bizType: 'game_application_audit',
        objectId: Number(this.data.gameId || 0) || 0
      })
      const nextFileIds = (this.data.uploadedFileIds || []).concat(fileIds || [])

      this.setData({
        uploadedFileIds: nextFileIds
      })
      toast.success(type === 'image' ? this.textOf('imageUploadedText') : this.textOf('fileUploadedText'))
    } catch (error) {
      toast.info(error.message || this.textOf('uploadFailedText'))
    } finally {
      wx.hideLoading()
    }
  },

  handleChooseFileFail(error = {}) {
    if (error.errMsg && error.errMsg.includes('cancel')) {
      return
    }

    toast.info(this.textOf('chooseFailedText'))
  },

  handleOptionalActionTap(event) {
    const key = event.currentTarget.dataset.key

    if (key === 'chat') {
      if (!this.data.hasDetail) {
        toast.info(this.textOf('detailRequiredActionText'))
        return
      }

      const params = [
        `gameId=${encodeURIComponent(this.data.gameId || '')}`,
        `prefill=${encodeURIComponent(this.textOf('chatPrefill'))}`
      ].join('&')
      this.navigateToRoute(`${ROUTES.imRoom}?${params}`)
      return
    }

    if (key === 'time') {
      if (!this.data.hasDetail) {
        toast.info(this.textOf('detailRequiredActionText'))
        return
      }

      const params = [
        `gameId=${encodeURIComponent(this.data.gameId || '')}`,
        `prefill=${encodeURIComponent(this.textOf('timePrefill'))}`
      ].join('&')
      this.navigateToRoute(`${ROUTES.imRoom}?${params}`)
      return
    }

    toast.info(this.textOf('unavailableActionText'))
  },

  handleDeclineTap() {
    this.reviewCurrentApplication(false)
  },

  handleApproveTap() {
    this.reviewCurrentApplication(true)
  },

  async reviewCurrentApplication(approve) {
    if (this.data.actionLoading || !this.data.auditId || !this.data.hasDetail) {
      if (!this.data.hasDetail) {
        toast.info(this.textOf('detailRequiredReviewText'))
      }
      return
    }

    this.setData({ actionLoading: true })
    wx.showLoading({
      title: approve ? this.textOf('approvingText') : this.textOf('rejectingText'),
      mask: true
    })

    try {
      const application = await gameService.reviewGameApplication(this.data.auditId, approve)
      this.setData(normalizeApplicationDetail(application, this.data.detailConfig))
      toast.success(approve ? this.textOf('approveSuccessText') : this.textOf('rejectSuccessText'))
    } catch (error) {
      toast.info(error.message || this.textOf('reviewFailedText'))
    } finally {
      wx.hideLoading()
      this.setData({ actionLoading: false })
    }
  },

  handleShellNavTap(event) {
    const key = event.detail && event.detail.key

    if (key === 'up' || key === 'down') {
      if (!this.suppressNextNavTap) {
        this.scrollDetail(key, DETAIL_SCROLL_TAP_STEP_RPX)
      }
      return
    }

    if (navigateShellKey(key, {
      currentRoute: ROUTES.gameAuditDetail
    })) {
      return
    }

    this.handleShellAction(key)
  },

  handleShellNavLongPress(event) {
    const key = event.detail && event.detail.key

    if (key !== 'up' && key !== 'down') {
      return
    }

    this.suppressNextNavTap = true
    this.stopDetailScrollHold(false)
    this.scrollDetail(key, DETAIL_SCROLL_HOLD_STEP_RPX)

    this.detailScrollHoldTimer = setInterval(() => {
      this.scrollDetail(key, DETAIL_SCROLL_HOLD_STEP_RPX)
    }, DETAIL_SCROLL_HOLD_INTERVAL_MS)
  },

  handleShellNavTouchEnd() {
    this.stopDetailScrollHold(true)
  },

  handleShellAction(key) {
    if (key === 'home') {
      this.scrollDetailToTop()
      return
    }

    if (key === 'search') {
      this.navigateToRoute(ROUTES.gameHall)
      return
    }

    if (key === 'comment' || key === 'message') {
      this.navigateToRoute(ROUTES.message)
      return
    }

    if (key === 'avatar' || key === 'mine') {
      this.navigateToRoute(ROUTES.profile)
      return
    }

    const routeMap = {
      metaverse: ROUTES.metaverse,
      map: ROUTES.map
    }

    this.navigateToRoute(routeMap[key])
  },

  navigateToRoute(route) {
    if (!route || route === ROUTES.gameAuditDetail) {
      return
    }

    navigateShellRoute(route)
  },

  handleDetailScroll(event) {
    const scrollTop = event.detail && event.detail.scrollTop

    if (typeof scrollTop === 'number') {
      this.detailScrollTopValue = scrollTop
    }
  },

  scrollDetail(direction, stepRpx = DETAIL_SCROLL_TAP_STEP_RPX) {
    const current = Number(this.detailScrollTopValue || this.data.detailScrollTop || 0)
    const distance = this.rpxToPx(stepRpx)
    const nextTop = direction === 'up'
      ? Math.max(0, current - distance)
      : current + distance

    this.detailScrollTopValue = nextTop
    this.setData({
      detailScrollTop: nextTop
    })
  },

  scrollDetailToTop() {
    this.detailScrollTopValue = 0
    this.setData({
      detailScrollTop: 0
    })
  },

  stopDetailScrollHold(resetTapSuppress) {
    if (this.detailScrollHoldTimer) {
      clearInterval(this.detailScrollHoldTimer)
      this.detailScrollHoldTimer = null
    }

    if (resetTapSuppress && this.suppressNextNavTap) {
      if (this.detailScrollSuppressTimer) {
        clearTimeout(this.detailScrollSuppressTimer)
      }

      this.detailScrollSuppressTimer = setTimeout(() => {
        this.suppressNextNavTap = false
        this.detailScrollSuppressTimer = null
      }, DETAIL_SCROLL_HOLD_SUPPRESS_TAP_MS)
    }
  },

  clearDetailScrollTimers() {
    this.stopDetailScrollHold(false)

    if (this.detailScrollSuppressTimer) {
      clearTimeout(this.detailScrollSuppressTimer)
      this.detailScrollSuppressTimer = null
    }

    this.suppressNextNavTap = false
  },

  rpxToPx(value) {
    if (!wx.getSystemInfoSync) {
      return value / 2
    }

    const system = wx.getSystemInfoSync()
    const windowWidth = system && system.windowWidth ? system.windowWidth : 375

    return Math.round((value * windowWidth) / 750)
  },

  onUnload() {
    this.clearDetailScrollTimers()
  }
})
