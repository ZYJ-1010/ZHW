const { ROUTES } = require('../../../config/routes')
const gameService = require('../../../services/game')

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
const PLAYER_INTRO_DEFAULT = '李明，这位张专家是我认识的产品大牛，正好符合你之前说的产品架构咨询需求，我帮你们牵个线！'
const DEFAULT_INVITE_PLAYER_RULE = {
  minPlayerCount: 1,
  maxPlayerCount: 1
}

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

function normalizeInvitePlayerRule(config = {}) {
  const minPlayerCount = Math.max(1, parsePositiveInteger(config.minPlayerCount, DEFAULT_INVITE_PLAYER_RULE.minPlayerCount))
  const maxPlayerCount = Math.max(minPlayerCount, parsePositiveInteger(config.maxPlayerCount, DEFAULT_INVITE_PLAYER_RULE.maxPlayerCount))

  return {
    minPlayerCount,
    maxPlayerCount
  }
}

function parsePositiveInteger(value, fallback) {
  const number = Number(value)

  return Number.isInteger(number) && number > 0 ? number : fallback
}

function getAvatarText(player = {}) {
  if (player.avatarText) {
    return player.avatarText
  }

  const name = String(player.name || player.nickname || '').trim()

  if (!name) {
    return '玩'
  }

  return /^[A-Za-z]/.test(name) ? name.slice(0, 2).toUpperCase() : name.slice(0, 1)
}

function normalizeInvitePlayer(player = {}, index = 0) {
  const id = String(player.id || player.userId || player.playerId || `invite-player-${index}`)

  return {
    id,
    avatarText: getAvatarText(player),
    avatarClass: player.avatarClass || ['pink', 'teal', 'purple', 'blue', 'orange'][index % 5],
    name: player.name || player.nickname || '玩家',
    tag: player.tag || player.tagText || '',
    desc: player.desc || player.description || player.title || '',
    meta: player.meta || player.metaText || player.extraText || ''
  }
}

function normalizeInvitePlayers(result) {
  const list = Array.isArray(result) ? result : (result && result.list) || []

  return list.map(normalizeInvitePlayer)
}

Page({
  data: {
    onlineText: '3999人在线',
    inviteStep: 1,
    inviteScrollTop: 0,
    selectedType: 'product',
    detailCount: 0,
    playerSearchKeyword: '',
    selectedPlayerId: '',
    selectedPlayerIds: [],
    selectedPlayerCount: 0,
    minPlayerCount: DEFAULT_INVITE_PLAYER_RULE.minPlayerCount,
    maxPlayerCount: DEFAULT_INVITE_PLAYER_RULE.maxPlayerCount,
    playerRuleText: '至少 1 位，最多 1 位',
    playerCountText: '已添加 0/1 位玩家',
    invitePlayerLoading: false,
    playerPickerVisible: false,
    playerPickerLoading: false,
    allPlayerSearchKeyword: '',
    allPlayers: [],
    playerIntroMessage: PLAYER_INTRO_DEFAULT,
    playerIntroCount: PLAYER_INTRO_DEFAULT.length,
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
    players: [],
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

  onLoad(options = {}) {
    const inviteStep = this.resolveInviteStep(options)

    this.setData({
      inviteStep,
      detailCount: String(this.data.form.detail || '').length,
      playerIntroCount: String(this.data.playerIntroMessage || '').length
    })

    if (inviteStep === 2) {
      this.loadInvitePlayerStep()
    }
  },

  resolveInviteStep(options = {}) {
    const step = Number(options.step)

    if (step === 2 || options.mode === 'invite2') {
      return 2
    }

    return 1
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

  onPlayerSearchInput(event) {
    const keyword = event.detail.value || ''

    this.setData({
      playerSearchKeyword: keyword
    })
    this.loadInviteRecentPlayers({ keyword })
  },

  onPlayerTap(event) {
    const id = event.currentTarget.dataset.id

    if (!id) {
      return
    }

    this.selectInvitePlayer(id)
  },

  onAddPlayer() {
    this.setData({
      playerPickerVisible: true
    })
    this.loadAllInvitePlayers()
  },

  onClosePlayerPicker() {
    this.setData({
      playerPickerVisible: false,
      allPlayerSearchKeyword: ''
    })
  },

  onAllPlayerSearchInput(event) {
    const keyword = event.detail.value || ''

    this.setData({
      allPlayerSearchKeyword: keyword
    })
    this.loadAllInvitePlayers({ keyword })
  },

  onAllPlayerTap(event) {
    const id = event.currentTarget.dataset.id

    if (!id) {
      return
    }

    this.selectInvitePlayer(id)
    this.ensurePlayerVisible(id)
    this.onClosePlayerPicker()
  },

  onPlayerIntroInput(event) {
    const value = event.detail.value || ''

    this.setData({
      playerIntroMessage: value,
      playerIntroCount: value.length
    })
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
    this.setInviteStep(2)
  },

  onPrevTap() {
    this.setInviteStep(1)
  },

  onConfirmInviteTap() {
    if (this.data.selectedPlayerCount < this.data.minPlayerCount) {
      this.showInfo(`请至少添加${this.data.minPlayerCount}位玩家`)
      return
    }

    this.showInfo('邀请已发起')
  },

  setInviteStep(inviteStep) {
    this.inviteScrollTopValue = 0
    this.setData({
      inviteStep,
      inviteScrollTop: 0
    })

    if (inviteStep === 2) {
      this.loadInvitePlayerStep()
    }
  },

  async loadInvitePlayerStep() {
    if (this.invitePlayerStepLoaded) {
      return
    }

    this.invitePlayerStepLoaded = true
    this.setData({
      invitePlayerLoading: true
    })

    try {
      const [config, playersResult] = await Promise.all([
        gameService.getInvitePlayerConfig(),
        gameService.getInviteRecentPlayers()
      ])
      const rule = normalizeInvitePlayerRule(config)
      const players = normalizeInvitePlayers(playersResult)
      const selectedPlayerIds = this.normalizeSelectedPlayerIds(this.data.selectedPlayerIds, rule, players)

      this.setData({
        players,
        invitePlayerLoading: false,
        ...this.buildPlayerRuleData(rule, selectedPlayerIds)
      })
    } catch (error) {
      this.invitePlayerStepLoaded = false
      this.setData({
        invitePlayerLoading: false,
        ...this.buildPlayerRuleData(DEFAULT_INVITE_PLAYER_RULE, this.data.selectedPlayerIds)
      })
      this.showInfo(error.message || '玩家规则加载失败')
    }
  },

  async loadInviteRecentPlayers(params = {}) {
    try {
      const result = await gameService.getInviteRecentPlayers(params)
      const players = normalizeInvitePlayers(result)

      this.setData({
        players
      })
    } catch (error) {
      this.showInfo(error.message || '最近联系加载失败')
    }
  },

  async loadAllInvitePlayers(params = {}) {
    this.setData({
      playerPickerLoading: true
    })

    try {
      const result = await gameService.getInvitePlayers(params)

      this.setData({
        allPlayers: normalizeInvitePlayers(result),
        playerPickerLoading: false
      })
    } catch (error) {
      this.setData({
        playerPickerLoading: false
      })
      this.showInfo(error.message || '玩家列表加载失败')
    }
  },

  selectInvitePlayer(id) {
    const rule = {
      minPlayerCount: this.data.minPlayerCount,
      maxPlayerCount: this.data.maxPlayerCount
    }
    const selectedPlayerIds = this.normalizeSelectedPlayerIds([id], rule)

    this.setData(this.buildPlayerRuleData(rule, selectedPlayerIds))
  },

  ensurePlayerVisible(id) {
    const existsInRecent = (this.data.players || []).some((player) => player.id === id)

    if (existsInRecent) {
      return
    }

    const selectedPlayer = (this.data.allPlayers || []).find((player) => player.id === id)

    if (!selectedPlayer) {
      return
    }

    this.setData({
      players: [selectedPlayer].concat(this.data.players || [])
    })
  },

  normalizeSelectedPlayerIds(selectedPlayerIds = [], rule = DEFAULT_INVITE_PLAYER_RULE, players = []) {
    const maxCount = Math.max(1, Number(rule.maxPlayerCount) || 1)
    let ids = Array.from(new Set((selectedPlayerIds || []).filter(Boolean))).slice(0, maxCount)

    if (!ids.length && players.length) {
      ids = [players[0].id]
    }

    return ids
  },

  buildPlayerRuleData(rule = DEFAULT_INVITE_PLAYER_RULE, selectedPlayerIds = []) {
    const normalizedRule = normalizeInvitePlayerRule(rule)
    const selectedIds = selectedPlayerIds.slice(0, normalizedRule.maxPlayerCount)
    const selectedPlayerCount = selectedIds.length
    const selectedPlayerId = selectedIds[0] || ''

    return {
      minPlayerCount: normalizedRule.minPlayerCount,
      maxPlayerCount: normalizedRule.maxPlayerCount,
      selectedPlayerId,
      selectedPlayerIds: selectedIds,
      selectedPlayerCount,
      playerRuleText: `至少 ${normalizedRule.minPlayerCount} 位，最多 ${normalizedRule.maxPlayerCount} 位`,
      playerCountText: `已添加 ${selectedPlayerCount}/${normalizedRule.maxPlayerCount} 位玩家`
    }
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
