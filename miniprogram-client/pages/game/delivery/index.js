const { getLegacyWhiteFrameLayoutStyles } = require('../../../utils/adaptive-shell-layout')
const gameService = require('../../../services/game')
const { ROUTES } = require('../../../config/routes')
const fileService = require('../../../services/file')
const { getSurnameInitials } = require('../../../utils/avatar')
const { navigateShellRoute } = require('../../../utils/shell-nav')
const { getActiveRole } = require('../../../utils/active-role')

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
const DEFAULT_DELIVERY_MODE = 'paid'

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
  const moreTop = Math.max(0, roundRpx(
    layout.capsuleTopRpx + (layout.capsuleHeightRpx - MORE_BUTTON_SIZE_RPX) / 2
  ))

  return Object.assign({}, layout, {
    moreStyle: `left: ${MORE_BUTTON_LEFT_RPX}rpx; top: ${moreTop}rpx; width: ${MORE_BUTTON_SIZE_RPX}rpx; height: ${MORE_BUTTON_SIZE_RPX}rpx;`
  })
}

const EMPTY_DELIVERY_PAGE_CONFIG = {
  paid: {
    pageTitle: '',
    status: {
      theme: 'paid',
      title: '',
      desc: ''
    },
    statePill: {
      theme: 'green',
      text: ''
    },
    activity: {},
    activityRows: [],
    notice: {},
    settlement: {},
    settlementRows: [],
    settlementNote: '',
    timeline: [],
    confirmItems: [],
    confirmNote: '',
    security: {
      title: '',
      desc: ''
    },
    submitHints: {
      ready: '',
      pending: ''
    },
    submitToast: '',
    submitLoadingText: '',
    amountRowLabel: ''
  },
  free: {
    pageTitle: '',
    status: {
      theme: 'free',
      title: '',
      desc: ''
    },
    statePill: {
      theme: 'blue',
      text: ''
    },
    activity: {},
    activityRows: [],
    notice: {},
    settlement: {},
    settlementRows: [],
    settlementNote: '',
    timeline: [],
    confirmItems: [],
    confirmNote: '',
    security: {
      title: '',
      desc: ''
    },
    submitHints: {
      ready: '',
      pending: ''
    },
    submitToast: '',
    submitLoadingText: '',
    amountRowLabel: ''
  },
  quickActions: []
}

function cloneList(list) {
  return Array.isArray(list) ? list.map((item) => ({ ...item })) : []
}

function cloneNotice(notice) {
  if (!notice || !notice.title) {
    return {}
  }

  return {
    ...notice,
    parts: cloneList(notice.parts || [])
  }
}

function normalizeDeliveryMode(mode) {
  return String(mode || '').toLowerCase() === 'free' ? 'free' : 'paid'
}

function resolveDeliveryMode(detail = {}, requestedMode = DEFAULT_DELIVERY_MODE) {
  if (detail.deliveryMode) {
    return normalizeDeliveryMode(detail.deliveryMode)
  }

  if (detail.fund && detail.fund.status === 'free_no_pay') {
    return 'free'
  }

  return normalizeDeliveryMode(requestedMode)
}

function normalizeViewerRole(role) {
  const value = String(role || '').toLowerCase()
  return value === 'player' ? 'member' : value
}

function mergeModeConfig(modeConfig = {}, fallback = {}) {
  return {
    ...fallback,
    ...modeConfig,
    status: { ...(fallback.status || {}), ...(modeConfig.status || {}) },
    statePill: { ...(fallback.statePill || {}), ...(modeConfig.statePill || {}) },
    notice: cloneNotice(modeConfig.notice || fallback.notice),
    confirmItems: cloneList(modeConfig.confirmItems || fallback.confirmItems),
    security: { ...(fallback.security || {}), ...(modeConfig.security || {}) },
    submitHints: { ...(fallback.submitHints || {}), ...(modeConfig.submitHints || {}) }
  }
}

function normalizeDeliveryPageConfig(config = {}) {
  return {
    paid: mergeModeConfig(config.paid, EMPTY_DELIVERY_PAGE_CONFIG.paid),
    free: mergeModeConfig(config.free, EMPTY_DELIVERY_PAGE_CONFIG.free),
    quickActions: cloneList(config.quickActions || EMPTY_DELIVERY_PAGE_CONFIG.quickActions)
  }
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

function emptyActivity() {
  return {
    avatar: '',
    expertName: '',
    serviceName: '',
    orderNo: ''
  }
}

function emptyDeliveryBusinessState() {
  return {
    activity: emptyActivity(),
    activityRows: [],
    settlement: {},
    settlementRows: [],
    hasSettlement: false,
    settlementNote: '',
    timeline: [],
    canSubmit: false,
    submitBlockText: ''
  }
}

function createDeliveryState(mode, pageConfig = EMPTY_DELIVERY_PAGE_CONFIG) {
  const deliveryMode = normalizeDeliveryMode(mode)
  const deliveryPage = normalizeDeliveryPageConfig(pageConfig)
  const config = deliveryPage[deliveryMode]
  const confirmItems = cloneList(config.confirmItems)

  return {
    deliveryMode,
    roleType: 'member',
    pageTitle: config.pageTitle,
    status: { ...config.status },
    statePill: { ...config.statePill },
    activity: emptyActivity(),
    activityRows: [],
    notice: cloneNotice(config.notice),
    settlement: {},
    settlementRows: [],
    hasSettlement: false,
    settlementNote: '',
    timeline: [],
    confirmItems,
    allConfirmed: confirmItems.length > 0 && confirmItems.every((item) => item.checked),
    confirmNote: config.confirmNote,
    security: { ...config.security },
    submitHints: { ...config.submitHints },
    submitToast: config.submitToast,
    proofImages: [],
    proofFileIds: [],
    showProofCard: deliveryMode !== 'free',
    canSubmit: false,
    hasConfirmed: false,
    submitBlockText: '',
    deliveryProof: {},
    quickActions: cloneList(deliveryPage.quickActions)
  }
}

function normalizeDeliveryDetail(detail = {}, mode = DEFAULT_DELIVERY_MODE) {
  const deliveryPage = normalizeDeliveryPageConfig(detail.deliveryPage)
  const completion = detail.completion || {}
  const viewer = detail.viewer || {}
  const group = detail.group || {}
  const fund = detail.fund || {}
  const deliveryMode = resolveDeliveryMode(detail, mode)
  const pageState = createDeliveryState(deliveryMode, deliveryPage)
  const deliveryProof = normalizeDeliveryProof(detail.deliveryProof)
  const participants = Array.isArray(detail.participants) ? detail.participants : []
  const roleType = normalizeViewerRole(completion.role || viewer.role || 'member')
  const hasConfirmed = Boolean(completion.hasConfirmed)
  const player = participants.find((item) => item.role === 'member' || item.role === 'player' || item.roleLabel === '玩家') || participants[0] || {}
  const expert = participants.find((item) => item.role === 'expert' || item.roleLabel === '行家') || participants[0] || {}
  const counterpart = roleType === 'member' ? expert : player
  const activityRows = Array.isArray(detail.activityRows) ? detail.activityRows : []
  const amountText = firstValue(fund.amountText, fund.amount ? `¥${Number(fund.amount).toLocaleString('zh-CN')}` : '')
  const normalizedRows = amountText
    ? [{ label: pageState.amountRowLabel, value: amountText, strong: true }].concat(activityRows)
    : activityRows
  const timeline = Array.isArray(detail.nextSteps) ? detail.nextSteps.map((item, index) => {
    const isContactAction = String(item.action || '').indexOf('contact_') === 0

    return {
      key: item.key || item.action || `step-${index}`,
      action: isContactAction ? (deliveryProof.primaryContactKey || item.action) : (item.action || ''),
      title: item.title || '',
      desc: item.desc || '',
      timeText: item.timeText || '',
      state: item.state || (item.active ? 'active' : index === 0 ? 'done' : 'pending'),
      lineState: item.lineState || '',
      hasLine: index < detail.nextSteps.length - 1,
      actionText: isContactAction ? (deliveryProof.primaryContactText || deliveryProof.timelineActionText) : '',
      actionPlacement: isContactAction ? 'head' : ''
    }
  }) : []

  return {
    ...pageState,
    confirmItems: pageState.confirmItems,
    allConfirmed: pageState.allConfirmed,
    activity: {
      avatar: counterpart.avatarText || counterpart.avatar || getSurnameInitials(counterpart.name || '', roleType === 'member' ? '行' : '玩'),
      expertName: counterpart.displayName || counterpart.name || '',
      roleTag: counterpart.roleLabel || (roleType === 'member' ? '行家' : '玩家'),
      serviceName: group.title || '',
      orderNo: group.gameId ? `GAME-${group.gameId}` : ''
    },
    activityRows: normalizedRows,
    settlement: amountText ? { actualAmount: amountText } : {},
    settlementRows: [],
    hasSettlement: false,
    settlementNote: '',
    timeline,
    roleType,
    canSubmit: completion.canConfirm === true,
    hasConfirmed,
    submitBlockText: completion.waitingText || '',
    deliveryProof,
    security: fund.status === 'free_no_pay'
      ? pageState.security
      : {
          title: fund.title || '',
          desc: fund.desc || ''
        }
  }
}

function normalizeDeliveryProof(proof = {}) {
  const maxCount = Math.max(0, Number(proof.maxCount || 0))

  return {
    maxCount,
    emptyText: proof.emptyText || '',
    fullText: proof.fullText || '',
    selectedTemplate: proof.selectedTemplate || '',
    uploadActionText: proof.uploadActionText || '',
    contactPlayerText: proof.contactPlayerText || '',
    contactExpertText: proof.contactExpertText || '',
    contactGuideText: proof.contactGuideText || '',
    cancelServiceText: proof.cancelServiceText || '',
    unavailableTextTemplate: proof.unavailableTextTemplate || '',
    contactPlayerPrefill: proof.contactPlayerPrefill || '',
    contactExpertPrefill: proof.contactExpertPrefill || '',
    contactGuidePrefill: proof.contactGuidePrefill || '',
    primaryContactKey: proof.primaryContactKey || '',
    primaryContactText: proof.primaryContactText || '',
    primaryContactPrefill: proof.primaryContactPrefill || '',
    timelinePrefill: proof.timelinePrefill || '',
    idleTimelineText: proof.idleTimelineText || '',
    timelineActionText: proof.timelineActionText || '',
    proofNoteTemplate: proof.proofNoteTemplate || ''
  }
}

function quickActionIconSrc(item = {}) {
  if (item.key === 'contact_player') {
    return '/pages/game/delivery/assets/contact-person.svg'
  }

  return item.iconSrc || ''
}

function deliveryQuickActions(baseActions = [], proof = {}) {
  const titleMap = {
    upload: proof.uploadActionText,
    contact_player: proof.contactPlayerText,
    contact_guide: proof.contactGuideText
  }

  return baseActions.map((item) => {
    const usePrimaryContact = item.key === 'contact_player' && proof.primaryContactKey
    return {
      ...item,
      key: usePrimaryContact ? proof.primaryContactKey : item.key,
      title: usePrimaryContact ? proof.primaryContactText : (titleMap[item.key] || item.title),
      iconSrc: quickActionIconSrc(item)
    }
  })
}

function filterQuickActions(actions = [], mode = DEFAULT_DELIVERY_MODE) {
  if (normalizeDeliveryMode(mode) !== 'free') {
    return actions
  }

  return actions.filter((item) => item.key !== 'upload')
}

function buildReviewRoute(gameId, roleType) {
  const params = [
    `gameId=${encodeURIComponent(gameId || 0)}`,
    `role=${encodeURIComponent(roleType || 'member')}`
  ].join('&')

  return `/${ROUTES.gameReview}?${params}`
}

Page({
  data: {
    shellLayout: getWhiteShellLayoutStyles(),
    ...createDeliveryState(DEFAULT_DELIVERY_MODE)
  },

  onLoad(options = {}) {
    this.applyDeliveryMode(options.mode)
    this.gameId = Number(options.gameId || options.id || 0) || 0
    this.confirmNote = String(options.note || options.confirmNote || '').trim()
    this.hasShown = false
    this.logServiceConfirm('page_enter', {
      autoConfirm: false,
      routeMode: options.mode || '',
      source: options.source || ''
    })
    this.loadDeliveryDetail()
  },

  async loadDeliveryDetail() {
    const requestedGameId = this.gameId

    if (!this.gameId) {
      this.setData({
        ...createDeliveryState(this.data.deliveryMode),
        ...emptyDeliveryBusinessState()
      })
      return
    }

    this.setData({
      ...createDeliveryState(this.data.deliveryMode),
      ...emptyDeliveryBusinessState()
    })

    try {
      const detail = await gameService.getGameSuccessDetail(requestedGameId, {
        view: 'service-confirm'
      })
      if (requestedGameId !== this.gameId) {
        return
      }
      const nextState = normalizeDeliveryDetail(detail, this.data.deliveryMode)
      nextState.quickActions = filterQuickActions(
        deliveryQuickActions(nextState.quickActions, nextState.deliveryProof),
        nextState.deliveryMode
      )
      this.setData(nextState)
      this.logServiceConfirm('detail_loaded', {
        viewerUserId: detail.viewer && detail.viewer.userId,
        viewerRole: nextState.roleType,
        canSubmit: nextState.canSubmit,
        hasConfirmed: nextState.hasConfirmed,
        confirmItemCount: nextState.confirmItems.length
      })
    } catch (error) {
      if (requestedGameId !== this.gameId) {
        return
      }
      this.setData({
        ...createDeliveryState(this.data.deliveryMode),
        ...emptyDeliveryBusinessState()
      })
      this.logServiceConfirm('entry_rejected', {
        errorCode: error && error.code,
        errorMessage: error && error.message
      })
      this.showInfo(error.message || '服务确认数据加载失败')

      if (String(error && error.message || '').indexOf('仅组局绑定') >= 0) {
        setTimeout(() => {
          navigateShellRoute(`${ROUTES.gameDetail}?gameId=${encodeURIComponent(requestedGameId)}`, {
            reuseExisting: false
          })
        }, 300)
      }
    }
  },

  onShow() {
    this.updateShellLayout()
    if (this.hasShown) {
      this.loadDeliveryDetail()
    }
    this.hasShown = true
  },

  onResize() {
    this.updateShellLayout()
  },

  updateShellLayout() {
    this.setData({
      shellLayout: getWhiteShellLayoutStyles()
    })
  },

  applyDeliveryMode(mode) {
    this.setData(createDeliveryState(mode || DEFAULT_DELIVERY_MODE))
  },

  onBackTap() {
    const pages = getCurrentPages()

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    navigateShellRoute(ROUTES.gamePlayerManage || 'pages/game/player-manage/index')
  },

  onMoreTap() {
    const actions = [
      this.data.showProofCard ? { key: 'upload', title: this.data.deliveryProof.uploadActionText } : null,
      { key: this.data.deliveryProof.primaryContactKey, title: this.data.deliveryProof.primaryContactText },
      { key: 'cancel_service', title: this.data.deliveryProof.cancelServiceText }
    ].filter((item) => item && item.title)

    if (!actions.length) {
      return
    }

    wx.showActionSheet({
      itemList: actions.map((item) => item.title),
      success: (result) => {
        const action = actions[result.tapIndex]

        if (!action) {
          return
        }

        if (action.key === 'upload') {
          this.chooseProofImages()
          return
        }

        if (String(action.key || '').indexOf('contact_') === 0) {
          this.openIM(this.data.deliveryProof.primaryContactPrefill)
          return
        }

        if (action.key === 'cancel_service') {
          navigateShellRoute(`/${ROUTES.gameExpertCancel || 'pages/game/expert-cancel/index'}?gameId=${encodeURIComponent(this.gameId || 0)}`)
        }
      }
    })
  },

  onTimelineActionTap(event) {
    const item = event.detail && event.detail.item
    const actionKey = item && (item.action || item.key)

    if (String(actionKey || '').indexOf('contact_') === 0 || (item && item.key === 'waiting-player')) {
      const prefillMap = {
        contact_player: this.data.deliveryProof.contactPlayerPrefill,
        contact_expert: this.data.deliveryProof.contactExpertPrefill,
        contact_guide: this.data.deliveryProof.contactGuidePrefill
      }
      this.openIM(this.data.deliveryProof.primaryContactPrefill || prefillMap[actionKey] || this.data.deliveryProof.timelinePrefill)
      return
    }

    this.showInfo(this.data.deliveryProof.idleTimelineText)
  },

  onQuickActionTap(event) {
    const { key, title } = event.currentTarget.dataset

    if (key === 'upload') {
      this.chooseProofImages()
      return
    }

    if (key === 'contact_player' || key === 'contact_expert' || key === 'contact_guide') {
      const prefillMap = {
        contact_player: this.data.deliveryProof.contactPlayerPrefill,
        contact_expert: this.data.deliveryProof.contactExpertPrefill,
        contact_guide: this.data.deliveryProof.contactGuidePrefill
      }
      this.openIM(prefillMap[key])
      return
    }

    this.showInfo((this.data.deliveryProof.unavailableTextTemplate || '').replace('{action}', title || ''))
  },

  chooseProofImages() {
    const maxCount = Number(this.data.deliveryProof.maxCount || 0)
    const remain = Math.max(0, maxCount - this.data.proofImages.length)

    if (!remain) {
      this.showInfo(this.data.deliveryProof.fullText || '')
      return
    }

    const onSuccess = (result = {}) => {
      const files = Array.isArray(result.tempFiles) ? result.tempFiles : []
      const proofImages = this.data.proofImages.concat(files
        .map((item) => item.tempFilePath || item.path)
        .filter(Boolean)
        .slice(0, remain)
        .map((path) => ({ path }))).slice(0, this.data.proofImages.length + remain)

      this.setData({ proofImages })
      this.showInfo(this.proofSelectedText(proofImages.length))
    }

    if (wx.chooseMedia) {
      wx.chooseMedia({
        count: remain,
        mediaType: ['image'],
        sourceType: ['album', 'camera'],
        success: onSuccess
      })
      return
    }

    wx.chooseImage({
      count: remain,
      sourceType: ['album', 'camera'],
      success: onSuccess
    })
  },

  proofSelectedText(selectedCount) {
    const maxCount = Number(this.data.deliveryProof.maxCount || 0)
    const template = this.data.deliveryProof.selectedTemplate || ''

    if (template) {
      return template.replace('{selected}', selectedCount).replace('{max}', maxCount)
    }

    return ''
  },

  removeProofImage(event) {
    const index = Number(event.currentTarget.dataset.index)

    if (Number.isNaN(index)) {
      return
    }

    this.setData({
      proofImages: this.data.proofImages.filter((_, itemIndex) => itemIndex !== index)
    })
  },

  openIM(prefill) {
    navigateShellRoute(`/${ROUTES.imRoom}?gameId=${encodeURIComponent(this.gameId || 0)}&prefill=${encodeURIComponent(prefill)}`)
  },

  toggleConfirm(event) {
    const { id } = event.currentTarget.dataset
    const confirmItems = this.data.confirmItems.map((item) => {
      if (item.id !== id || item.locked) {
        return item
      }

      return {
        ...item,
        checked: !item.checked
      }
    })
    const allConfirmed = confirmItems.every((item) => item.checked)

    this.setData({
      confirmItems,
      allConfirmed
    })
    this.logServiceConfirm('selection_changed', {
      changedItemKey: id,
      allConfirmed
    })
  },

  async onSubmitTap() {
    this.logServiceConfirm('submit_attempt')
    if (!this.data.canSubmit) {
      this.logServiceConfirm('submit_blocked', {
        reason: 'not_allowed',
        blockText: this.data.submitBlockText || ''
      })
      this.showInfo(this.data.submitBlockText || '当前暂不能确认完成')
      return
    }

    if (!this.data.allConfirmed) {
      this.logServiceConfirm('submit_blocked', {
        reason: 'confirm_items_incomplete'
      })
      this.showInfo(this.data.submitHints.pending || '')
      return
    }

    if (!this.gameId) {
      this.logServiceConfirm('submit_blocked', { reason: 'missing_game_id' })
      this.showInfo('缺少局信息，无法确认服务')
      return
    }

    wx.showLoading({ title: this.data.submitLoadingText || '', mask: true })

    try {
      const fileIds = await fileService.uploadEvidenceImages(
        this.data.proofImages.map((item) => item.path).filter(Boolean),
        {
          bizType: 'delivery_proof',
          objectId: this.gameId
        }
      )
      const note = this.confirmNote || this.data.confirmNote || ''

      const proofNote = fileIds.length
        ? (this.data.deliveryProof.proofNoteTemplate || '').replace('{fileIds}', fileIds.join(','))
        : ''

      const confirmItemKeys = this.data.confirmItems.filter((item) => item.checked).map((item) => item.id)
      this.logServiceConfirm('request_start', {
        confirmItemKeys,
        fileCount: fileIds.length,
        noteProvided: Boolean(note)
      })
      const result = await gameService.confirmService(this.gameId, {
        note: [note, proofNote].filter(Boolean).join('\n'),
        fileIds,
        confirmItemKeys,
        confirmSource: 'delivery_page',
        pageGameId: this.gameId
      })

      this.logServiceConfirm('request_success', {
        confirmId: result && result.confirm && result.confirm.id,
        confirmStatus: result && result.confirm && result.confirm.status,
        gameStatus: result && result.game && result.game.status,
        confirmedCount: result && Array.isArray(result.items) ? result.items.length : 0
      })

      await this.loadDeliveryDetail()
      wx.hideLoading()
      this.showInfo(this.data.submitToast || '')
      const gameStatus = result && result.game && result.game.status
      const isExpert = this.data.roleType === 'expert'
      if (!isExpert && (gameStatus === 'pending_review' || gameStatus === 'completed')) {
        const reviewRoute = buildReviewRoute(this.gameId, this.data.roleType)
        this.logServiceConfirm('navigate_to_review', { reviewRoute, gameStatus })
        setTimeout(() => {
          navigateShellRoute(reviewRoute, { reuseExisting: false })
        }, 300)
      } else {
        const homeRoute = `/${this.data.roleType === 'expert' ? (ROUTES.expertHome || ROUTES.home) : (ROUTES.playerHome || ROUTES.home)}`
        this.logServiceConfirm('review_deferred', {
          gameStatus,
          reason: isExpert ? 'expert_returns_home' : 'waiting_for_other_confirmation',
          homeRoute
        })
        setTimeout(() => {
          navigateShellRoute(homeRoute, { reuseExisting: false })
        }, 300)
      }
      return
    } catch (error) {
      wx.hideLoading()
      this.logServiceConfirm('request_failed', {
        errorCode: error && error.code,
        errorMessage: error && error.message
      })
      this.showInfo(error.message || '')
    }
  },
  logServiceConfirm(event, detail = {}) {
    const confirmItems = Array.isArray(this.data.confirmItems) ? this.data.confirmItems : []
    const selectedItemKeys = confirmItems.filter((item) => item.checked).map((item) => item.id)
    console.info('[service-confirm]', {
      event,
      gameId: this.gameId || 0,
      viewerRole: this.data.roleType || '',
      canSubmit: Boolean(this.data.canSubmit),
      hasConfirmed: Boolean(this.data.hasConfirmed),
      selectedItemKeys,
      selectedItemCount: selectedItemKeys.length,
      timestamp: new Date().toISOString(),
      ...detail
    })
  },
  showInfo(title) {
    if (!title) {
      return
    }

    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
