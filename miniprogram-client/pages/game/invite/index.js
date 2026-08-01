const { ROUTES } = require('../../../config/routes')
const { navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')
const gameService = require('../../../services/game')
const { getSurnameInitials } = require('../../../utils/avatar')

const INVITE_SCROLL_TAP_STEP_RPX = 360
const INVITE_SCROLL_HOLD_STEP_RPX = 72
const INVITE_SCROLL_HOLD_INTERVAL_MS = 80
const INVITE_SCROLL_HOLD_SUPPRESS_TAP_MS = 120

const WORK_IMAGE_MAX_COUNT = 3
const WORK_IMAGE_EXTENSIONS = ['jpg', 'jpeg', 'png', 'gif', 'webp']
const EMPTY_INVITE_PLAYER_RULE = {
  minPlayerCount: 1,
  maxPlayerCount: 1
}
const EMPTY_ACTIVITY_TYPES = []
const EMPTY_EXPERT = {
  id: '',
  userId: '',
  name: '',
  avatarText: '',
  avatarSrc: '',
  roleName: '',
  desc: '',
  tags: []
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
  const minPlayerCount = Math.max(1, parsePositiveInteger(config.minPlayerCount, EMPTY_INVITE_PLAYER_RULE.minPlayerCount))
  const maxPlayerCount = Math.max(minPlayerCount, parsePositiveInteger(config.maxPlayerCount, EMPTY_INVITE_PLAYER_RULE.maxPlayerCount))

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
  const name = String(player.name || player.nickname || '').trim()

  return getSurnameInitials(name, player.avatarText || 'WA')
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
    ,
    inGame: Boolean(player.inGame),
    canSelect: player.canSelect !== false && !player.inGame
  }
}

function normalizeInvitePlayers(result) {
  const list = Array.isArray(result) ? result : (result && result.list) || []

  return list.map(normalizeInvitePlayer)
}

function normalizeActivityTypes(config = {}) {
  const list = Array.isArray(config.activityTypes) ? config.activityTypes : []

  return list
    .filter((item) => item && item.key && item.name)
    .map((item) => ({
      key: String(item.key),
      name: String(item.name)
    }))
}

function normalizeInviteExpert(expert = {}) {
  const name = String(expert.name || expert.nickname || '').trim()
  const id = expert.userId || expert.id || ''
  const tags = Array.isArray(expert.tags)
    ? expert.tags.filter(Boolean).slice(0, 4)
    : []

  return {
    id,
    userId: id,
    name,
    avatarText: getSurnameInitials(name, expert.avatarText || 'EX'),
    avatarSrc: expert.avatarSrc || expert.avatarUrl || '',
    avatarType: expert.avatarType || '',
    roleName: expert.roleLabel || expert.roleName || '行家',
    desc: expert.desc || expert.description || expert.title || '',
    tags
  }
}

Page({
  data: {
    onlineText: '在线',
    inviteStep: 1,
    confirmInviteText: '确认发起',
    inviteScrollTop: 0,
    submittingInvite: false,
    selectedType: '',
    detailCount: 0,
    playerSearchKeyword: '',
    selectedPlayerId: '',
    selectedPlayerIds: [],
    selectedPlayerCount: 0,
    minPlayerCount: EMPTY_INVITE_PLAYER_RULE.minPlayerCount,
    maxPlayerCount: EMPTY_INVITE_PLAYER_RULE.maxPlayerCount,
    playerRuleText: '至少 1 位，最多 1 位',
    playerCountText: '已添加 0/1 位玩家',
    invitePlayerLoading: false,
    playerPickerVisible: false,
    playerPickerLoading: false,
    allPlayerSearchKeyword: '',
    allPlayers: [],
    playerIntroMessage: '',
    playerIntroCount: 0,
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ],
    expert: EMPTY_EXPERT,
    expertCandidates: [],
    expertCandidateIndex: 0,
    sourceGame: null,
    sourceGameRows: [],
    activityTypes: EMPTY_ACTIVITY_TYPES,
    players: [],
    form: {
      title: '',
      detail: ''
    },
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

  async onLoad(options = {}) {
    const inviteStep = this.resolveInviteStep(options)
    const gameId = options.gameId || options.id || options.sourceGameId || ''

    this.setData({
      gameId,
      sourceGameId: gameId,
      inviteStep,
      detailCount: String(this.data.form.detail || '').length,
      playerIntroCount: String(this.data.playerIntroMessage || '').length
    })

    try {
      const permission = await gameService.getInvitePermission({ gameId })

      if (!permission.allowed) {
        this.showInfo(permission.reason || '仅局创建者或主领路人可发起引荐')
        setTimeout(() => wx.navigateBack({ delta: 1 }), 500)
        return
      }
    } catch (error) {
      this.showInfo(error.message || '引荐权限校验失败')
      setTimeout(() => wx.navigateBack({ delta: 1 }), 500)
      return
    }

    if (inviteStep === 2) {
      this.loadInvitePlayerStep()
    } else {
      this.loadInviteConfig()
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

  onPlayerSearchInput(event) {
    const keyword = event.detail.value || ''

    this.setData({
      playerSearchKeyword: keyword
    })
    this.loadInviteRecentPlayers({ keyword })
  },

  onPlayerTap(event) {
    const id = event.currentTarget.dataset.id
	const disabled = event.currentTarget.dataset.disabled

    if (!id || disabled) {
	  if (disabled) this.showInfo('该玩家已在局内，无需重复邀请')
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
	const disabled = event.currentTarget.dataset.disabled

    if (!id || disabled) {
	  if (disabled) this.showInfo('该玩家已在局内，无需重复邀请')
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

  async loadInviteConfig() {
    try {
      const config = await gameService.getInvitePlayerConfig({ sourceGameId: this.data.sourceGameId })

      this.applyInviteConfig(config)
    } catch (error) {
      this.showInfo(error.message || '邀请配置加载失败')
    }
  },

  applyInviteConfig(config = {}) {
    const activityTypes = normalizeActivityTypes(config)
    const experts = Array.isArray(config.experts) ? config.experts.map(normalizeInviteExpert) : []
    const selectedType = activityTypes.length ? activityTypes[0].key : ''
    const defaultTitle = String(config.defaultTitle || '').trim()
    const defaultDetail = String(config.defaultDetail || '').trim()
    const intro = String(config.playerIntroTemplate || '').trim()
    const sourceGameRows = (Array.isArray(config.sourceGameRows) ? config.sourceGameRows : [])
      .filter((item) => !/费用|预算|价格|分润|奖励/.test(String(item && item.label || '')))

    this.setData({
      activityTypes,
      selectedType,
      expert: experts[0] || normalizeInviteExpert(config.expert || {}),
      expertCandidates: experts,
      expertCandidateIndex: 0,
      sourceGame: config.sourceGame || null,
      sourceGameRows,
      playerIntroMessage: intro,
      playerIntroCount: intro.length,
      'form.title': defaultTitle,
      'form.detail': defaultDetail,
      detailCount: defaultDetail.length
    })
  },

  async onReplaceExpert() {
    let experts = this.data.expertCandidates || []

    if (!experts.length) {
      try {
        const result = await gameService.getSystemRecommendations({
          category: this.data.selectedType || 'all'
        })
        experts = Array.isArray(result.experts) ? result.experts.map(normalizeInviteExpert) : []
      } catch (error) {
        this.showInfo(error.message || '行家列表加载失败')
        return
      }
    }

    if (!experts.length) {
      this.showInfo('暂无可更换行家')
      return
    }

    const nextIndex = (Number(this.data.expertCandidateIndex || 0) + 1) % experts.length

    this.setData({
      expert: experts[nextIndex],
      expertCandidates: experts,
      expertCandidateIndex: nextIndex
    })
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

  async onConfirmInviteTap() {
    if (this.data.selectedPlayerCount < this.data.minPlayerCount) {
      this.showInfo(`请至少添加${this.data.minPlayerCount}位玩家`)
      return
    }

    if (!this.data.sourceGameId) {
      this.showInfo('缺少组局信息')
      return
    }

    const expertUserId = Number(this.data.expert && (this.data.expert.userId || this.data.expert.id))

    if (!Number.isInteger(expertUserId) || expertUserId <= 0) {
      this.showInfo('请选择行家')
      return
    }

    if (this.data.submittingInvite) {
      return
    }

    this.setData({ submittingInvite: true })
    wx.showLoading({
      title: '发起邀请中',
      mask: true
    })

    try {
      const playerUserIds = (this.data.selectedPlayerIds || [])
        .map((id) => Number(id))
        .filter((id) => Number.isInteger(id) && id > 0)
      const payload = {
        sourceGameId: this.data.sourceGameId,
        expertUserId,
        playerUserIds,
        expertUserIds: [expertUserId],
        invitees: playerUserIds.map((id) => ({ id, roleType: 'player' })),
        message: this.data.playerIntroMessage || this.data.form.detail || this.data.form.title || '',
        serviceType: this.data.form.title || '',
        serviceDuration: this.data.sourceGame && (this.data.sourceGame.serviceDuration || this.data.sourceGame.durationText) || '',
        demandDetail: this.data.form.detail || '',
        budgetAmountCent: 0,
        expectedTime: this.data.sourceGame && (this.data.sourceGame.expectedTime || this.data.sourceGame.expectedTimeText) || '',
        overrides: {
          title: this.data.form.title || '',
          description: this.data.form.detail || '',
          price: 0,
          startAt: this.data.sourceGame && this.data.sourceGame.startAt || '',
          endAt: this.data.sourceGame && this.data.sourceGame.endAt || ''
        }
      }
      const result = await gameService.createCurrentGameInvitation(payload)

      this.showInfo('邀请已发起')
      setTimeout(() => {
        const invitationId = result.invitationId || result.replayInvitationId || result.primaryInvitationId || ''
        const query = [
          invitationId ? `invitationId=${encodeURIComponent(invitationId)}` : '',
          result.gameId ? `gameId=${encodeURIComponent(result.gameId)}` : ''
        ].filter(Boolean).join('&')

        navigateShellRoute(`/${ROUTES.gameGuideProgressDetail}${query ? `?${query}` : ''}`)
      }, 500)
    } catch (error) {
      this.showInfo(error.message || '邀请发起失败')
    } finally {
      wx.hideLoading()
      this.setData({ submittingInvite: false })
    }
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
      const expertUserId = Number(this.data.expert && (this.data.expert.userId || this.data.expert.id)) || 0
      const [config, playersResult] = await Promise.all([
        gameService.getInvitePlayerConfig({ sourceGameId: this.data.sourceGameId }),
        gameService.getInviteRecentPlayers({ sourceGameId: this.data.sourceGameId, excludeUserId: expertUserId })
      ])
      const rule = normalizeInvitePlayerRule(config)
      const players = normalizeInvitePlayers(playersResult).filter((player) => Number(player.id) !== expertUserId)
      const selectedPlayerIds = this.normalizeSelectedPlayerIds(this.data.selectedPlayerIds, rule, players)

      this.applyInviteConfig(config)
      this.setData({
        players,
        invitePlayerLoading: false,
        ...this.buildPlayerRuleData(rule, selectedPlayerIds)
      })
    } catch (error) {
      this.invitePlayerStepLoaded = false
      this.setData({
        invitePlayerLoading: false,
        ...this.buildPlayerRuleData(EMPTY_INVITE_PLAYER_RULE, this.data.selectedPlayerIds)
      })
      this.showInfo(error.message || '玩家规则加载失败')
    }
  },

  async loadInviteRecentPlayers(params = {}) {
    try {
      const expertUserId = Number(this.data.expert && (this.data.expert.userId || this.data.expert.id)) || 0
      const result = await gameService.getInviteRecentPlayers({
        ...params,
        sourceGameId: params.sourceGameId || this.data.sourceGameId,
        excludeUserId: expertUserId
      })
      const players = normalizeInvitePlayers(result).filter((player) => Number(player.id) !== expertUserId)

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
      const expertUserId = Number(this.data.expert && (this.data.expert.userId || this.data.expert.id)) || 0
      const result = await gameService.getInvitePlayers({
        ...params,
        sourceGameId: params.sourceGameId || this.data.sourceGameId,
        excludeUserId: expertUserId
      })

      this.setData({
        allPlayers: normalizeInvitePlayers(result).filter((player) => Number(player.id) !== expertUserId),
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

  normalizeSelectedPlayerIds(selectedPlayerIds = [], rule = EMPTY_INVITE_PLAYER_RULE, players = []) {
    const maxCount = Math.max(1, Number(rule.maxPlayerCount) || 1)
    let ids = Array.from(new Set((selectedPlayerIds || []).filter(Boolean))).slice(0, maxCount)

    if (!ids.length && players.length) {
      ids = [players[0].id]
    }

    return ids
  },

  buildPlayerRuleData(rule = EMPTY_INVITE_PLAYER_RULE, selectedPlayerIds = []) {
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

    if (navigateShellKey(key, {
      currentRoute: ROUTES.gameInvite
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
    if (!route || route === ROUTES.gameInvite) {
      return
    }

    navigateShellRoute(route)
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
