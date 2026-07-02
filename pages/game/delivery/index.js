const gameService = require('../../../services/game')
const { ROUTES } = require('../../../config/routes')
const fileService = require('../../../services/file')
const { getSurnameInitials } = require('../../../utils/avatar')
const { navigateShellRoute } = require('../../../utils/shell-nav')

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
  const capsuleBottom = getCapsuleBottomRpx()
  const frameHeight = getFrameHeightRpx()
  const bottomHeight = roundRpx(frameHeight * DESIGN_BOTTOM_HEIGHT_PT / DESIGN_FRAME_HEIGHT_PT)
  const bottomTop = Math.max(CONTENT_TOP_RPX, roundRpx(frameHeight - bottomHeight))
  const contentHeight = Math.max(0, roundRpx(bottomTop - CONTENT_TOP_RPX))
  const titleTop = Math.max(0, roundRpx(capsuleBottom - NAV_TITLE_HEIGHT_RPX))
  const backTop = Math.max(0, roundRpx(capsuleBottom - BACK_BUTTON_SIZE_RPX))
  const moreTop = Math.max(0, roundRpx(capsuleBottom - MORE_BUTTON_SIZE_RPX - 3))

  return {
    frameStyle: `height: ${frameHeight}rpx; min-height: ${frameHeight}rpx;`,
    contentStyle: [
      `left: ${CONTENT_LEFT_RPX}rpx`,
      `top: ${CONTENT_TOP_RPX}rpx`,
      `width: ${CONTENT_WIDTH_RPX}rpx`,
      `height: ${contentHeight}rpx`
    ].join('; '),
    bottomStyle: `top: ${bottomTop}rpx; height: ${bottomHeight}rpx;`,
    titleStyle: `top: ${titleTop}rpx; height: ${NAV_TITLE_HEIGHT_RPX}rpx; line-height: ${NAV_TITLE_HEIGHT_RPX}rpx;`,
    backStyle: `top: ${backTop}rpx; width: ${BACK_BUTTON_SIZE_RPX}rpx; height: ${BACK_BUTTON_SIZE_RPX}rpx;`,
    moreStyle: `left: ${MORE_BUTTON_LEFT_RPX}rpx; top: ${moreTop}rpx; width: ${MORE_BUTTON_SIZE_RPX}rpx; height: ${MORE_BUTTON_SIZE_RPX}rpx;`
  }
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
    timeline: []
  }
}

function createDeliveryState(mode, pageConfig = EMPTY_DELIVERY_PAGE_CONFIG) {
  const deliveryMode = normalizeDeliveryMode(mode)
  const deliveryPage = normalizeDeliveryPageConfig(pageConfig)
  const config = deliveryPage[deliveryMode]
  const confirmItems = cloneList(config.confirmItems)

  return {
    deliveryMode,
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
    deliveryProof: {},
    quickActions: cloneList(deliveryPage.quickActions)
  }
}

function normalizeDeliveryDetail(detail = {}, mode = DEFAULT_DELIVERY_MODE) {
  const deliveryPage = normalizeDeliveryPageConfig(detail.deliveryPage)
  const pageState = createDeliveryState(mode, deliveryPage)
  const deliveryProof = normalizeDeliveryProof(detail.deliveryProof)
  const group = detail.group || {}
  const fund = detail.fund || {}
  const participants = Array.isArray(detail.participants) ? detail.participants : []
  const player = participants.find((item) => item.role === 'member' || item.roleLabel === '玩家') || participants[0] || {}
  const activityRows = Array.isArray(detail.activityRows) ? detail.activityRows : []
  const amountText = firstValue(fund.amountText, fund.amount ? `¥${Number(fund.amount).toLocaleString('zh-CN')}` : '')
  const normalizedRows = amountText
    ? [{ label: pageState.amountRowLabel, value: amountText, strong: true }].concat(activityRows)
    : activityRows
  const timeline = Array.isArray(detail.nextSteps) ? detail.nextSteps.map((item, index) => ({
    key: item.action || item.key || `step-${index}`,
    title: item.title || '',
    desc: item.desc || '',
    timeText: item.timeText || '',
    state: item.active ? 'active' : index === 0 ? 'done' : 'pending',
    hasLine: index < detail.nextSteps.length - 1,
    actionText: item.action === 'contact_player' ? deliveryProof.timelineActionText : '',
    actionPlacement: item.action === 'contact_player' ? 'head' : ''
  })) : []

  return {
    ...pageState,
    activity: {
      avatar: player.avatarText || player.avatar || getSurnameInitials(player.name || '', '玩'),
      expertName: player.name || '',
      serviceName: group.title || '',
      orderNo: group.gameId ? `GAME-${group.gameId}` : ''
    },
    activityRows: normalizedRows,
    settlement: amountText ? { actualAmount: amountText } : {},
    settlementRows: [],
    hasSettlement: false,
    settlementNote: '',
    timeline,
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
    contactGuideText: proof.contactGuideText || '',
    cancelServiceText: proof.cancelServiceText || '',
    unavailableTextTemplate: proof.unavailableTextTemplate || '',
    contactPlayerPrefill: proof.contactPlayerPrefill || '',
    contactGuidePrefill: proof.contactGuidePrefill || '',
    timelinePrefill: proof.timelinePrefill || '',
    idleTimelineText: proof.idleTimelineText || '',
    timelineActionText: proof.timelineActionText || '',
    proofNoteTemplate: proof.proofNoteTemplate || ''
  }
}

function deliveryQuickActions(baseActions = [], proof = {}) {
  const titleMap = {
    upload: proof.uploadActionText,
    contact_player: proof.contactPlayerText,
    contact_guide: proof.contactGuideText
  }

  return baseActions.map((item) => ({
    ...item,
    title: titleMap[item.key] || item.title
  }))
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
    this.loadDeliveryDetail()
  },

  async loadDeliveryDetail() {
    if (!this.gameId) {
      this.setData(emptyDeliveryBusinessState())
      return
    }

    try {
      const detail = await gameService.getGameSuccessDetail(this.gameId, { role: 'expert' })
      const nextState = normalizeDeliveryDetail(detail, this.data.deliveryMode)
      nextState.quickActions = deliveryQuickActions(nextState.quickActions, nextState.deliveryProof)
      this.setData(nextState)
    } catch (error) {
      this.setData(emptyDeliveryBusinessState())
      this.showInfo(error.message || '服务确认数据加载失败')
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
      { key: 'upload', title: this.data.deliveryProof.uploadActionText },
      { key: 'contact_player', title: this.data.deliveryProof.contactPlayerText },
      { key: 'cancel_service', title: this.data.deliveryProof.cancelServiceText }
    ].filter((item) => item.title)

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

        if (action.key === 'contact_player') {
          this.openIM(this.data.deliveryProof.contactPlayerPrefill)
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

    if (item && item.key === 'waiting-player') {
      this.openIM(this.data.deliveryProof.timelinePrefill)
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

    if (key === 'contact_player' || key === 'contact_guide') {
      this.openIM(key === 'contact_player' ? this.data.deliveryProof.contactPlayerPrefill : this.data.deliveryProof.contactGuidePrefill)
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
  },

  async onSubmitTap() {
    if (!this.data.allConfirmed) {
      this.showInfo(this.data.submitHints.pending || '')
      return
    }

    if (!this.gameId) {
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

      await gameService.confirmService(this.gameId, {
        note: [note, proofNote].filter(Boolean).join('\n'),
        fileIds
      })

      wx.hideLoading()
      this.showInfo(this.data.submitToast || '')
      this.setData({
        proofFileIds: fileIds,
        statePill: {
          ...this.data.statePill,
          text: this.data.statePill.text
        }
      })
    } catch (error) {
      wx.hideLoading()
      this.showInfo(error.message || '')
    }
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

