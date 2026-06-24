const { ROUTES } = require('../../../config/routes')

const APPLY_SCROLL_TAP_STEP_RPX = 360
const APPLY_SCROLL_HOLD_STEP_RPX = 72
const APPLY_SCROLL_HOLD_INTERVAL_MS = 80
const APPLY_SCROLL_HOLD_SUPPRESS_TAP_MS = 120
const IMAGE_EXTENSIONS = ['jpg', 'jpeg', 'png', 'gif', 'webp']
const PDF_EXTENSIONS = ['pdf']

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

Page({
  data: {
    gameId: '',
    onlineText: '3999人在线',
    applyScrollTop: 0,
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ],
    wechatInfo: {
      avatarFallback: '张',
      avatarUrl: '',
      nickname: '张伟',
      syncText: '头像已同步'
    },
    form: {
      intro: '我有3年产品经验，做过增长和内容工具，对游戏化社交产品很感兴趣。',
      message: '我很想一起把这个局玩成。',
      imageFiles: [],
      attachmentFiles: [],
      agreed: true
    }
  },

  onLoad(options = {}) {
    this.setData({
      gameId: options.gameId || options.id || ''
    })
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

    this.showInfo('当前微信版本不支持选择图片')
  },

  onUploadFile() {
    if (!wx.chooseMessageFile) {
      this.showInfo('当前微信版本不支持选择文件')
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

    if (!isAllowedImageFile(file)) {
      this.showInfo('仅支持 JPG、PNG、GIF、WEBP 图片')
      return
    }

    this.setData({
      'form.imageFiles': [normalizeUploadFile(file)]
    })
    this.showInfo('图片已选择')
  },

  handlePdfFileSelected(file) {
    if (!file) {
      return
    }

    if (!isAllowedPdfFile(file)) {
      this.showInfo('仅支持 PDF 文件')
      return
    }

    this.setData({
      'form.attachmentFiles': [normalizeUploadFile(file)]
    })
    this.showInfo('文件已选择')
  },

  handleChooseFileFail(error = {}) {
    if (error.errMsg && error.errMsg.includes('cancel')) {
      return
    }

    this.showInfo('选择失败，请重试')
  },

  onAgreementTap() {
    this.showInfo('平台协议页面待接入')
  },

  onCancel() {
    const pages = getCurrentPages()
    const previousPage = pages[pages.length - 2]

    if (previousPage && previousPage.route === ROUTES.gameHall) {
      wx.navigateBack()
      return
    }

    wx.redirectTo({
      url: `/${ROUTES.gameHall}`
    })
  },

  onSubmit() {
    if (!String(this.data.form.intro || '').trim()) {
      this.showInfo('请先填写自我介绍')
      return
    }

    if (!this.data.form.agreed) {
      this.showInfo('请先勾选平台协议')
      return
    }

    this.showInfo('功能开发中')
  },

  handleShellNavTap(event) {
    const key = event.detail && event.detail.key

    if (key === 'up' || key === 'down') {
      if (!this.suppressNextNavTap) {
        this.scrollApply(key, APPLY_SCROLL_TAP_STEP_RPX)
      }
      return
    }

    if (key === 'left' || key === 'right') {
      this.showInfo('功能正在开发中')
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
      this.showInfo('搜索功能开发中')
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

    wx.navigateTo({
      url: `/${route}`
    })
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
