const { ROUTES } = require('../../../config/routes')

const INVITE_SCROLL_TAP_STEP_RPX = 360
const INVITE_SCROLL_HOLD_STEP_RPX = 72
const INVITE_SCROLL_HOLD_INTERVAL_MS = 80
const INVITE_SCROLL_HOLD_SUPPRESS_TAP_MS = 120

const DEFAULT_BUDGET = '800'
const DEFAULT_BUDGET_MAX_AMOUNT = 99999999
const DEFAULT_REWARD_RATE_CONFIG = {
  platformServiceRate: 10,
  systemGuideRewardRate: 10,
  inviteRewardRate: 40
}
const WORK_IMAGE_MAX_COUNT = 3
const WORK_IMAGE_EXTENSIONS = ['jpg', 'jpeg', 'png', 'gif', 'webp']

function formatRateText(rate) {
  return `${rate}%`
}

function normalizeRate(value, fallback) {
  const number = Number(value)

  if (!Number.isFinite(number) || number < 0) {
    return fallback
  }

  const percent = number > 0 && number <= 1 ? number * 100 : number

  return Math.min(100, Math.round(percent * 100) / 100)
}

function normalizeRewardRateConfig(config = {}) {
  const normalized = {
    platformServiceRate: normalizeRate(config.platformServiceRate, DEFAULT_REWARD_RATE_CONFIG.platformServiceRate),
    systemGuideRewardRate: normalizeRate(config.systemGuideRewardRate, DEFAULT_REWARD_RATE_CONFIG.systemGuideRewardRate),
    inviteRewardRate: normalizeRate(config.inviteRewardRate, DEFAULT_REWARD_RATE_CONFIG.inviteRewardRate)
  }
  const totalRate = normalized.platformServiceRate + normalized.systemGuideRewardRate + normalized.inviteRewardRate

  return totalRate <= 100 ? normalized : { ...DEFAULT_REWARD_RATE_CONFIG }
}

function calculateReward(budget, rateConfig = DEFAULT_REWARD_RATE_CONFIG) {
  const amount = Number(budget) || 0
  const rates = normalizeRewardRateConfig(rateConfig)
  const serviceFee = Math.round((amount * rates.platformServiceRate) / 100)
  const systemReward = Math.round((amount * rates.systemGuideRewardRate) / 100)
  const inviteReward = Math.round((amount * rates.inviteRewardRate) / 100)

  return {
    serviceFee,
    systemReward,
    inviteReward,
    expertIncome: Math.max(0, amount - serviceFee - systemReward - inviteReward),
    platformServiceRateText: formatRateText(rates.platformServiceRate),
    systemGuideRewardRateText: formatRateText(rates.systemGuideRewardRate),
    inviteRewardRateText: formatRateText(rates.inviteRewardRate)
  }
}

function normalizeBudgetInput(value, maxAmount = DEFAULT_BUDGET_MAX_AMOUNT) {
  const safeMaxAmount = Number.isInteger(Number(maxAmount)) && Number(maxAmount) > 0
    ? Number(maxAmount)
    : DEFAULT_BUDGET_MAX_AMOUNT
  const maxInputLength = String(safeMaxAmount).length + 1
  const digits = String(value || '')
    .replace(/\D/g, '')
    .replace(/^0+(?=\d)/, '')
    .slice(0, maxInputLength)

  if (!digits) {
    return ''
  }

  return String(Math.min(Number(digits), safeMaxAmount))
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

function isAllowedWorkImageFile(file = {}) {
  const extension = getFileExtension(file)

  if (extension) {
    return WORK_IMAGE_EXTENSIONS.includes(extension)
  }

  return file.fileType === 'image' || file.type === 'image'
}

function createWorkImage(file = {}, index = 0) {
  const path = file.tempFilePath || file.path || ''
  const name = getFileName(file) || `作品图片${index + 1}.jpg`

  return {
    id: `work-image-${Date.now()}-${index}`,
    src: path,
    name,
    desc: ''
  }
}

function createWorkImageSlots(images = []) {
  return Array.from({ length: WORK_IMAGE_MAX_COUNT }, (_, index) => {
    const image = images[index]

    if (!image) {
      return {
        slotId: `work-slot-${index}`,
        empty: true
      }
    }

    return {
      ...image,
      slotId: `work-slot-${index}`,
      empty: false
    }
  })
}

Page({
  data: {
    onlineText: '3999人在线',
    inviteScrollTop: 0,
    selectedType: 'product',
    detailCount: 0,
    rewardRateConfig: DEFAULT_REWARD_RATE_CONFIG,
    reward: calculateReward(DEFAULT_BUDGET, DEFAULT_REWARD_RATE_CONFIG),
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ],
    expert: {
      avatarText: 'ZE',
      name: '张专家',
      roleName: '行家',
      desc: '资深产品经理·10年经验',
      tags: ['产品咨询', '架构梳理']
    },
    activityTypes: [
      { key: 'product', name: '产品咨询' },
      { key: 'design', name: '设计服务' },
      { key: 'tech', name: '技术开发' }
    ],
    form: {
      title: '产品架构梳理咨询',
      detail: '需要资深产品经理帮忙梳理B端产品架构，预计咨询时长2小时，涉及模块划分和数据流转设计。',
      budget: DEFAULT_BUDGET
    },
    budgetMaxAmount: DEFAULT_BUDGET_MAX_AMOUNT,
    workImageMaxCount: WORK_IMAGE_MAX_COUNT,
    workImages: [],
    workImageSlots: createWorkImageSlots(),
    activeWorkImageId: '',
    workDescEditor: {
      visible: false,
      imageId: '',
      value: '',
      count: 0
    }
  },

  onLoad() {
    this.setData({
      detailCount: String(this.data.form.detail || '').length
    })
  },

  onTypeTap(event) {
    const key = event.currentTarget.dataset.key

    if (!key) {
      return
    }

    this.setData({
      selectedType: key
    })
  },

  onTitleInput(event) {
    this.setData({
      'form.title': event.detail.value || ''
    })
  },

  onDetailInput(event) {
    const value = event.detail.value || ''

    this.setData({
      'form.detail': value,
      detailCount: value.length
    })
  },

  onBudgetInput(event) {
    const budget = normalizeBudgetInput(event.detail.value, this.data.budgetMaxAmount)

    this.setData({
      'form.budget': budget,
      reward: calculateReward(budget, this.data.rewardRateConfig)
    })

    return budget
  },

  applyInvitePricingConfig(config = {}) {
    const budgetMaxAmount = Number.isInteger(Number(config.budgetMaxAmount)) && Number(config.budgetMaxAmount) > 0
      ? Number(config.budgetMaxAmount)
      : DEFAULT_BUDGET_MAX_AMOUNT
    const rewardRateConfig = normalizeRewardRateConfig(config.rewardRateConfig || config.rewardRates || config)
    const budget = normalizeBudgetInput(this.data.form.budget, budgetMaxAmount)

    this.setData({
      budgetMaxAmount,
      rewardRateConfig,
      'form.budget': budget,
      reward: calculateReward(budget, rewardRateConfig)
    })
  },

  onReplaceExpert() {
    this.showInfo('更换行家待接入')
  },

  onAddWorkImage() {
    const currentImages = this.data.workImages || []
    this.closeWorkImageMenu()

    if (currentImages.length >= WORK_IMAGE_MAX_COUNT) {
      this.showInfo(`最多添加${WORK_IMAGE_MAX_COUNT}张图片`)
      return
    }

    this.chooseWorkImage((image) => {
      this.setWorkImages(currentImages.concat(image))
      this.showInfo('图片已添加')
    }, currentImages.length)
  },

  onWorkImageTap(event) {
    const id = event.currentTarget.dataset.id

    if (!id) {
      return
    }

    this.setData({
      activeWorkImageId: this.data.activeWorkImageId === id ? '' : id,
      workDescEditor: {
        visible: false,
        imageId: '',
        value: '',
        count: 0
      }
    })
  },

  onDeleteWorkImage(event) {
    const id = event.currentTarget.dataset.id

    if (!id) {
      return
    }

    this.deleteWorkImageById(id)
    this.showInfo('图片已删除')
  },

  onWorkImageDescAction(event) {
    const id = event.currentTarget.dataset.id
    const image = (this.data.workImages || []).find((item) => item.id === id)

    if (!image) {
      return
    }

    this.promptWorkImageDesc(image)
  },

  promptWorkImageDesc(image) {
    const value = String(image.desc || '').slice(0, 20)

    this.setData({
      activeWorkImageId: '',
      workDescEditor: {
        visible: true,
        imageId: image.id,
        value,
        count: value.length
      }
    })
  },

  onWorkDescInput(event) {
    const value = String(event.detail.value || '').slice(0, 20)

    this.setData({
      'workDescEditor.value': value,
      'workDescEditor.count': value.length
    })
  },

  onSaveWorkDesc() {
    const editor = this.data.workDescEditor || {}
    const desc = String(editor.value || '').trim().slice(0, 20)

    if (!editor.imageId) {
      return
    }

    if (!desc) {
      this.showInfo('请输入说明')
      return
    }

    this.setWorkImageDesc(editor.imageId, desc)
    this.showInfo('说明已保存')
  },

  closeWorkDescEditor() {
    this.setData({
      workDescEditor: {
        visible: false,
        imageId: '',
        value: '',
        count: 0
      }
    })
  },

  setWorkImageDesc(id, desc) {
    const workImages = (this.data.workImages || []).map((image) => image.id === id
      ? {
        ...image,
        desc
      }
      : image)

    this.setWorkImages(workImages)
  },

  deleteWorkImageById(id) {
    this.setWorkImages((this.data.workImages || []).filter((image) => image.id !== id))
  },

  replaceWorkImageById(id) {
    const workImages = this.data.workImages || []
    const targetIndex = workImages.findIndex((image) => image.id === id)

    if (targetIndex < 0) {
      return
    }

    this.chooseWorkImage((image) => {
      const nextImages = workImages.map((item, index) => index === targetIndex
        ? {
          ...image,
          id: item.id
        }
        : item)

      this.setWorkImages(nextImages)
      this.showInfo('图片已替换')
    }, targetIndex)
  },

  setWorkImages(workImages) {
    this.setData({
      workImages,
      workImageSlots: createWorkImageSlots(workImages),
      activeWorkImageId: '',
      workDescEditor: {
        visible: false,
        imageId: '',
        value: '',
        count: 0
      }
    })
  },

  closeWorkImageMenu() {
    if (!this.data.activeWorkImageId) {
      return
    }

    this.setData({
      activeWorkImageId: ''
    })
  },

  noop() {},

  chooseWorkImage(onSelect, index = 0) {
    if (wx.chooseMedia) {
      wx.chooseMedia({
        count: 1,
        mediaType: ['image'],
        sourceType: ['album', 'camera'],
        success: (result) => {
          const file = result.tempFiles && result.tempFiles[0]
          this.handleWorkImageSelected(file, onSelect, index)
        },
        fail: (error) => {
          this.handleChooseImageFail(error)
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
          this.handleWorkImageSelected(file, onSelect, index)
        },
        fail: (error) => {
          this.handleChooseImageFail(error)
        }
      })
      return
    }

    this.showInfo('当前微信版本不支持选择图片')
  },

  handleWorkImageSelected(file, onSelect, index) {
    if (!file) {
      return
    }

    if (!isAllowedWorkImageFile(file)) {
      this.showInfo('仅支持 JPG、PNG、GIF、WEBP 图片')
      return
    }

    if (typeof onSelect === 'function') {
      onSelect(createWorkImage(file, index))
    }
  },

  handleChooseImageFail(error = {}) {
    if (error.errMsg && error.errMsg.includes('cancel')) {
      return
    }

    this.showInfo('选择图片失败，请重试')
  },

  onNextTap() {
    this.showInfo('选择玩家待接入')
  },

  handleShellNavTap(event) {
    const key = event.detail && event.detail.key

    if (key === 'up' || key === 'down') {
      if (!this.suppressNextNavTap) {
        this.scrollInvite(key, INVITE_SCROLL_TAP_STEP_RPX)
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
    this.stopInviteScrollHold(false)
    this.scrollInvite(key, INVITE_SCROLL_HOLD_STEP_RPX)

    this.inviteScrollHoldTimer = setInterval(() => {
      this.scrollInvite(key, INVITE_SCROLL_HOLD_STEP_RPX)
    }, INVITE_SCROLL_HOLD_INTERVAL_MS)
  },

  handleShellNavTouchEnd() {
    this.stopInviteScrollHold(true)
  },

  handleShellAction(key) {
    if (key === 'home') {
      this.scrollInviteToTop()
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
    if (!route || route === ROUTES.gameInvite) {
      return
    }

    wx.navigateTo({
      url: `/${route}`
    })
  },

  handleInviteScroll(event) {
    const scrollTop = event.detail && event.detail.scrollTop

    if (typeof scrollTop === 'number') {
      this.inviteScrollTopValue = scrollTop
    }
  },

  scrollInvite(direction, stepRpx = INVITE_SCROLL_TAP_STEP_RPX) {
    const current = Number(this.inviteScrollTopValue || this.data.inviteScrollTop || 0)
    const distance = this.rpxToPx(stepRpx)
    const nextTop = direction === 'up'
      ? Math.max(0, current - distance)
      : current + distance

    this.inviteScrollTopValue = nextTop
    this.setData({
      inviteScrollTop: nextTop
    })
  },

  scrollInviteToTop() {
    this.inviteScrollTopValue = 0
    this.setData({
      inviteScrollTop: 0
    })
  },

  stopInviteScrollHold(resetTapSuppress) {
    if (this.inviteScrollHoldTimer) {
      clearInterval(this.inviteScrollHoldTimer)
      this.inviteScrollHoldTimer = null
    }

    if (resetTapSuppress && this.suppressNextNavTap) {
      if (this.inviteScrollSuppressTimer) {
        clearTimeout(this.inviteScrollSuppressTimer)
      }

      this.inviteScrollSuppressTimer = setTimeout(() => {
        this.suppressNextNavTap = false
        this.inviteScrollSuppressTimer = null
      }, INVITE_SCROLL_HOLD_SUPPRESS_TAP_MS)
    }
  },

  clearInviteScrollTimers() {
    this.stopInviteScrollHold(false)

    if (this.inviteScrollSuppressTimer) {
      clearTimeout(this.inviteScrollSuppressTimer)
      this.inviteScrollSuppressTimer = null
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
    this.clearInviteScrollTimers()
  }
})
