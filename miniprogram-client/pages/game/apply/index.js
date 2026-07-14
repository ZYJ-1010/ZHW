const { ROUTES } = require('../../../config/routes')
const { navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')
const gameService = require('../../../services/game')
const fileService = require('../../../services/file')
const profileService = require('../../../services/profile')
const { getActiveRole } = require('../../../utils/active-role')

const APPLY_SCROLL_TAP_STEP_RPX = 360
const APPLY_SCROLL_HOLD_STEP_RPX = 72
const APPLY_SCROLL_HOLD_INTERVAL_MS = 80
const APPLY_SCROLL_HOLD_SUPPRESS_TAP_MS = 120
const IMAGE_EXTENSIONS = ['jpg', 'jpeg', 'png', 'gif', 'webp']
const PDF_EXTENSIONS = ['pdf']
const EMPTY_APPLICATION_CONFIG = {
  agreementTitle: '',
  agreementText: '',
  requireIntro: false,
  requireAgreement: false,
  uploadRequired: false,
  maxUploadCount: 0,
  allowedUploadTypes: [],
  minIntroLength: 0,
  maxIntroLength: 0,
  maxMessageLength: 0,
  searchEnabled: false,
  texts: {}
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

Page({
  data: {
    gameId: '',
    onlineText: '在线',
    applyScrollTop: 0,
    submitting: false,
    eligibilityLoading: true,
    eligibilityReady: false,
    eligibilityError: '',
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ],
    wechatInfo: {
      avatarFallback: '我',
      avatarUrl: '',
      nickname: '',
      syncText: ''
    },
    applicationConfig: EMPTY_APPLICATION_CONFIG,
    form: {
      intro: '',
      message: '',
      imageFiles: [],
      attachmentFiles: [],
      agreed: false
    }
  },

  onLoad(options = {}) {
    const gameId = options.gameId || options.id || ''
    this.setData({ gameId, eligibilityLoading: true, eligibilityReady: false, eligibilityError: '' })
    this.loadApplicationConfig()
    this.loadWechatInfo()
    this.loadApplicationEligibility(gameId)
  },

  async loadApplicationEligibility(gameId) {
    if (!gameId) {
      this.setData({ eligibilityLoading: false, eligibilityError: '后端未返回申请入局所需的局信息' })
      return
    }
    try {
      const detail = await gameService.getGameDetail(gameId)
      const relation = detail && detail.myRelation
      if (!relation || !Object.prototype.hasOwnProperty.call(relation, 'canApply')) {
        throw new Error('后端未返回入局申请资格')
      }
      if (relation.canApply !== true) {
        const action = detail.detailDisplay && detail.detailDisplay.primaryAction || {}
        throw new Error(action.disabledReason || action.text || '当前账号不具备本局申请资格')
      }
      this.setData({ eligibilityLoading: false, eligibilityReady: true, eligibilityError: '' })
    } catch (error) {
      this.setData({ eligibilityLoading: false, eligibilityReady: false, eligibilityError: error.message || '后端返回申请资格失败' })
    }
  },

  async loadApplicationConfig() {
    try {
      const config = await gameService.getApplicationConfig()
      this.setData({
        applicationConfig: Object.assign({}, EMPTY_APPLICATION_CONFIG, config || {}, {
          texts: Object.assign({}, EMPTY_APPLICATION_CONFIG.texts, config && config.texts || {})
        })
      })
    } catch (error) {
      this.showInfo(error.message || this.textOf('loadFailedText'))
    }
  },

  async loadWechatInfo() {
    try {
      const data = await profileService.getSystemProfileInfo()
      const profile = data && data.profile || {}
      const personalInfo = data && data.personalInfo || {}
      const nickname = personalInfo.name || profile.name || ''
      const avatarFallback = personalInfo.avatarText || profile.avatarText || (nickname ? nickname.slice(0, 1) : '我')
      this.setData({
        wechatInfo: {
          avatarFallback,
          avatarUrl: personalInfo.avatarUrl || profile.avatarUrl || '',
          nickname: nickname || this.textOf('profileNameFallback'),
          syncText: nickname ? this.textOf('profileSyncedText') : this.textOf('profilePendingText')
        }
      })
    } catch (error) {
      this.setData({
        wechatInfo: {
          avatarFallback: '我',
          avatarUrl: '',
          nickname: this.textOf('profileNameFallback'),
          syncText: this.textOf('profilePendingText')
        }
      })
    }
  },

  onIntroInput(event) {
    this.setData({
      'form.intro': event.detail.value || ''
    })
  },

  onMessageInput(event) {
    this.setData({
      'form.message': event.detail.value || ''
    })
  },

  onAgreementChange(event) {
    const values = event.detail && event.detail.value

    this.setData({
      'form.agreed': Array.isArray(values) && values.includes('agreed')
    })
  },

  textOf(key, values = {}) {
    const texts = this.data.applicationConfig && this.data.applicationConfig.texts || {}
    let text = String(texts[key] || '')

    Object.keys(values).forEach((name) => {
      text = text.replace(new RegExp(`\\{${name}\\}`, 'g'), String(values[name]))
    })

    return text
  },

  onUploadImage() {
    if (wx.chooseMedia) {
      wx.chooseMedia({
        count: 1,
        mediaType: ['image'],
        sourceType: ['album', 'camera'],
        success: (result) => {
          const file = result.tempFiles && result.tempFiles[0]
          this.handleImageFileSelected(file)
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
          this.handleImageFileSelected(file)
        },
        fail: (error) => {
          this.handleChooseFileFail(error)
        }
      })
      return
    }

    this.showInfo(this.textOf('mediaUnsupportedText'))
  },

  onUploadFile() {
    if (!wx.chooseMessageFile) {
      this.showInfo(this.textOf('fileUnsupportedText'))
      return
    }

    wx.chooseMessageFile({
      count: 1,
      type: 'file',
      extension: PDF_EXTENSIONS,
      success: (result) => {
        const file = result.tempFiles && result.tempFiles[0]
        this.handlePdfFileSelected(file)
      },
      fail: (error) => {
        this.handleChooseFileFail(error)
      }
    })
  },

  handleImageFileSelected(file) {
    if (!file) {
      return
    }

    if (!this.isAllowedApplicationFile(file, IMAGE_EXTENSIONS)) {
      this.showInfo(this.textOf('imageTypeErrorText'))
      return
    }

    if (!this.canAppendUploadFile(1)) {
      return
    }

    this.setData({
      'form.imageFiles': [normalizeUploadFile(file)]
    })
    this.showInfo(this.textOf('imageSelectedText'))
  },

  handlePdfFileSelected(file) {
    if (!file) {
      return
    }

    if (!this.isAllowedApplicationFile(file, PDF_EXTENSIONS)) {
      this.showInfo(this.textOf('fileTypeErrorText'))
      return
    }

    if (!this.canAppendUploadFile(1)) {
      return
    }

    this.setData({
      'form.attachmentFiles': [normalizeUploadFile(file)]
    })
    this.showInfo(this.textOf('fileSelectedText'))
  },

  handleChooseFileFail(error = {}) {
    if (error.errMsg && error.errMsg.includes('cancel')) {
      return
    }

    this.showInfo(this.textOf('chooseFailedText'))
  },

  onAgreementTap() {
    const agreementTitle = this.data.applicationConfig.agreementTitle || ''

    navigateShellRoute(`/pages/profile/system-management/agreement-detail/index?agreement=user-service&title=${encodeURIComponent(agreementTitle)}&signed=0`, {
      currentRoute: ROUTES.gameApply
    })
  },

  onCancel() {
    const pages = getCurrentPages()
    const previousPage = pages[pages.length - 2]

    if (previousPage && previousPage.route === ROUTES.gameHall) {
      wx.navigateBack()
      return
    }

    navigateShellRoute(ROUTES.gameHall, {
      currentRoute: ROUTES.gameApply
    })
  },

  async onSubmit() {
    if (!this.data.eligibilityReady) {
      this.showInfo(this.data.eligibilityError || '后端尚未确认申请资格')
      return
    }
    const intro = String(this.data.form.intro || '').trim()
    const config = this.data.applicationConfig || EMPTY_APPLICATION_CONFIG

    if (config.requireIntro && !intro) {
      this.showInfo(this.textOf('introRequiredText'))
      return
    }

    if (config.minIntroLength > 0 && intro && intro.length < config.minIntroLength) {
      this.showInfo(this.textOf('introMinTemplate', { min: config.minIntroLength }))
      return
    }

    if (config.requireAgreement && !this.data.form.agreed) {
      this.showInfo(this.textOf('agreementRequiredText'))
      return
    }

    if (config.uploadRequired && !this.selectedUploadFileCount()) {
      this.showInfo(this.textOf('uploadRequiredText'))
      return
    }

    if (!this.data.gameId) {
      this.showInfo(this.textOf('gameMissingText'))
      return
    }

    if (this.data.submitting) {
      return
    }

    this.setData({ submitting: true })
    wx.showLoading({
      title: this.textOf('submittingText'),
      mask: true
    })

    try {
      const fileIds = await this.uploadApplicationFiles()
      await gameService.applyGame(this.data.gameId, {
        reason: this.buildApplyReason(),
        fileIds,
        roleType: getActiveRole()
      })
      this.showInfo(this.textOf('submitSuccessText'))
      setTimeout(() => {
        navigateShellRoute(`${ROUTES.gameDetail}?gameId=${encodeURIComponent(this.data.gameId)}`, {
          currentRoute: ROUTES.gameApply
        })
      }, 500)
    } catch (error) {
      this.showInfo(error.message || this.textOf('submitFailedText'))
    } finally {
      wx.hideLoading()
      this.setData({ submitting: false })
    }
  },

  buildApplyReason() {
    const intro = String(this.data.form.intro || '').trim()
    const message = String(this.data.form.message || '').trim()

    return [intro, message].filter(Boolean).join('\n')
  },

  async uploadApplicationFiles() {
    const files = []
      .concat(this.data.form.imageFiles || [])
      .concat(this.data.form.attachmentFiles || [])
      .filter((file) => file && file.path)

    if (!files.length) {
      return []
    }

    return fileService.uploadEvidenceImages(files.map((file) => file.path), {
      bizType: 'game_application',
      objectId: Number(this.data.gameId || 0) || 0
    })
  },

  selectedUploadFileCount() {
    return (this.data.form.imageFiles || []).length + (this.data.form.attachmentFiles || []).length
  },

  canAppendUploadFile(count) {
    const maxUploadCount = Number(this.data.applicationConfig.maxUploadCount || 0)

    if (maxUploadCount > 0 && this.selectedUploadFileCount() + count > maxUploadCount) {
      this.showInfo(this.textOf('maxUploadTemplate', { max: maxUploadCount }))
      return false
    }

    return true
  },

  isAllowedApplicationFile(file, fallbackExtensions) {
    const extension = getFileExtension(file)
    const allowed = (this.data.applicationConfig.allowedUploadTypes || fallbackExtensions || [])
      .map((item) => String(item || '').toLowerCase())

    if (!extension) {
      return fallbackExtensions.some((item) => item === 'jpg' || item === 'png') && (file.fileType === 'image' || file.type === 'image')
    }

    return allowed.includes(extension) || fallbackExtensions.includes(extension)
  },

  handleShellNavTap(event) {
    const key = event.detail && event.detail.key

    if (key === 'up' || key === 'down') {
      if (!this.suppressNextNavTap) {
        this.scrollApply(key, APPLY_SCROLL_TAP_STEP_RPX)
      }
      return
    }

    if (navigateShellKey(key, {
      currentRoute: ROUTES.gameApply
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
    this.stopApplyScrollHold(false)
    this.scrollApply(key, APPLY_SCROLL_HOLD_STEP_RPX)

    this.applyScrollHoldTimer = setInterval(() => {
      this.scrollApply(key, APPLY_SCROLL_HOLD_STEP_RPX)
    }, APPLY_SCROLL_HOLD_INTERVAL_MS)
  },

  handleShellNavTouchEnd() {
    this.stopApplyScrollHold(true)
  },

  handleShellAction(key) {
    if (key === 'home') {
      this.scrollApplyToTop()
      return
    }

    if (key === 'search') {
      navigateShellRoute(ROUTES.gameHall, {
        currentRoute: ROUTES.gameApply
      })
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
    if (!route) {
      return
    }

    navigateShellRoute(route)
  },

  handleApplyScroll(event) {
    const scrollTop = event.detail && event.detail.scrollTop

    if (typeof scrollTop === 'number') {
      this.applyScrollTopValue = scrollTop
    }
  },

  scrollApply(direction, stepRpx = APPLY_SCROLL_TAP_STEP_RPX) {
    const current = Number(this.applyScrollTopValue || this.data.applyScrollTop || 0)
    const distance = this.rpxToPx(stepRpx)
    const nextTop = direction === 'up'
      ? Math.max(0, current - distance)
      : current + distance

    this.applyScrollTopValue = nextTop
    this.setData({
      applyScrollTop: nextTop
    })
  },

  scrollApplyToTop() {
    this.applyScrollTopValue = 0
    this.setData({
      applyScrollTop: 0
    })
  },

  stopApplyScrollHold(resetTapSuppress) {
    if (this.applyScrollHoldTimer) {
      clearInterval(this.applyScrollHoldTimer)
      this.applyScrollHoldTimer = null
    }

    if (resetTapSuppress && this.suppressNextNavTap) {
      if (this.applyScrollSuppressTimer) {
        clearTimeout(this.applyScrollSuppressTimer)
      }

      this.applyScrollSuppressTimer = setTimeout(() => {
        this.suppressNextNavTap = false
        this.applyScrollSuppressTimer = null
      }, APPLY_SCROLL_HOLD_SUPPRESS_TAP_MS)
    }
  },

  clearApplyScrollTimers() {
    this.stopApplyScrollHold(false)

    if (this.applyScrollSuppressTimer) {
      clearTimeout(this.applyScrollSuppressTimer)
      this.applyScrollSuppressTimer = null
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

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  },

  onUnload() {
    this.clearApplyScrollTimers()
  }
})
