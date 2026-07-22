const toast = require('../../../utils/toast')
const gameService = require('../../../services/game')
const fileService = require('../../../services/file')
const featureFlags = require('./feature-flags')
const { ROUTES } = require('../../../config/routes')
const { navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')

const CREATE_SCROLL_TAP_STEP_RPX = 360
const CREATE_SCROLL_HOLD_STEP_RPX = 72
const CREATE_SCROLL_HOLD_INTERVAL_MS = 80
const CREATE_SCROLL_HOLD_SUPPRESS_TAP_MS = 120
const COVER_ALLOWED_FORMATS = ['jpg', 'jpeg', 'png', 'webp']
const COVER_MAX_SIZE = 10 * 1024 * 1024
const DEFAULT_GAME_COVER = 'https://static.haowan.net.cn/miniprogram/components/game-card/assets/game-cover-default.png'
const DESCRIPTION_ALLOWED_IMAGE_FORMATS = ['jpg', 'jpeg', 'png']
const DESCRIPTION_ALLOWED_VIDEO_FORMATS = ['mp4', 'mov']
const DESCRIPTION_IMAGE_MAX_COUNT = 9
const DESCRIPTION_VIDEO_MAX_DURATION = 60
const THEME_MAX_LENGTH = 20
const INTRO_MAX_LENGTH = 200
const HIGHLIGHTS_MAX_LENGTH = 100
const NOTICE_MAX_LENGTH = 100
const AUDIENCE_MAX_LENGTH = 50
const EMPTY_DEPOSIT_RULE_TEXT = ''
const EMPTY_DEPOSIT_NOTICE_TEXT = ''
const EMPTY_GAME_TYPES = []
const MAX_SELECTED_TAGS = 3
// 一期只保留收费局入口的视觉与交互提示；实际创建始终为免费局。
// 不依赖后台临时配置，避免配置遗漏时收费局入口消失。
const PHASE_ONE_FEE_TYPES = [
  { key: 'free', name: '免费局', active: false },
  { key: 'paid', name: '收费局', active: false }
]
const EMPTY_PROFIT_TEMPLATES = []
const CREATE_PREVIEW_STORAGE_KEY = 'game_create_preview_v1'
const CREATE_PREVIEW_ACTION_KEY = 'game_create_preview_action_v1'
const EMPTY_CONDITION_RULE_CONFIG = {
  enabled: false,
  visibleInMiniProgram: false,
  adminOnlyCreate: true,
  ruleItems: [],
  defaultVisibility: '',
  reviewRequired: false,
  paymentRequired: false
}
const EMPTY_CREATE_FORM = {
  capacity: { min: 0, max: 0 },
  currentLocationText: '',
  participationModes: [],
  tags: [],
  completionRules: [],
  feeTypes: []
}

function padNumber(value) {
  return String(value).padStart(2, '0')
}

function formatDate(date) {
  return [
    date.getFullYear(),
    padNumber(date.getMonth() + 1),
    padNumber(date.getDate())
  ].join('-')
}

function formatTime(date) {
  return `${padNumber(date.getHours())}:${padNumber(date.getMinutes())}`
}

function addHours(date, hours) {
  const next = new Date(date.getTime())

  next.setHours(next.getHours() + hours)

  return next
}

function getInitialTimeDraft() {
  const now = new Date()
  const start = addHours(now, 1)
  const end = addHours(start, 2)

  return {
    startDate: formatDate(start),
    startTime: formatTime(start),
    endDate: formatDate(end),
    endTime: formatTime(end)
  }
}

function getInitialSignupTimeDraft() {
  const now = new Date()
  const gameStart = addHours(now, 1)
  const end = new Date(gameStart.getTime() - 10 * 60 * 1000)

  return {
    startDate: formatDate(now),
    startTime: formatTime(now),
    endDate: formatDate(end),
    endTime: formatTime(end)
  }
}

function getTimeDraftText(draft = {}) {
  if (!draft.startDate || !draft.startTime || !draft.endDate || !draft.endTime) {
    return ''
  }

  return `${draft.startDate} ${draft.startTime} - ${draft.endDate} ${draft.endTime}`
}

function getDraftTimestamp(dateText, timeText) {
  return Date.parse(`${dateText}T${timeText}:00+08:00`)
}

function getCurrentMinuteTimestamp() {
  const now = new Date()

  now.setSeconds(0, 0)

  return now.getTime()
}

function isDraftBeforeNow(draft = {}) {
  const timestamp = getDraftTimestamp(draft.startDate, draft.startTime)

  return !timestamp || Number.isNaN(timestamp) || timestamp < getCurrentMinuteTimestamp()
}

function getDurationText(startTimestamp, endTimestamp) {
  const diffMs = endTimestamp - startTimestamp

  if (diffMs <= 0) {
    return ''
  }

  const totalHours = Math.ceil(diffMs / (60 * 60 * 1000))

  if (totalHours < 24) {
    return `${totalHours}小时`
  }

  const days = Math.floor(totalHours / 24)
  const hours = totalHours % 24

  return hours > 0 ? `${days}天${hours}小时` : `${days}天`
}

function clampCapacity(value, capacity = EMPTY_CREATE_FORM.capacity) {
  const numberValue = Number(value)
  const min = Number(capacity.min || 0)
  const max = Number(capacity.max || min)

  if (!min || !max) {
    return 0
  }

  if (Number.isNaN(numberValue)) {
    return min
  }

  return Math.min(max, Math.max(min, Math.round(numberValue)))
}

function getFileExtension(filePath = '') {
  const normalized = String(filePath || '').split('?')[0].split('#')[0]
  const matched = normalized.match(/\.([a-zA-Z0-9]+)$/)

  return matched ? matched[1].toLowerCase() : ''
}

function isAllowedCoverFormat(file = {}, imageInfo = {}) {
  return isAllowedMediaFormat(file, imageInfo, COVER_ALLOWED_FORMATS)
}

function normalizeCoverUploadFile(file = {}, imageInfo = {}) {
  const path = String(file.tempFilePath || '').trim()
  const rawFormat = String(imageInfo.type || file.type || getFileExtension(path) || 'jpg').toLowerCase()
  const format = rawFormat === 'jpeg' ? 'jpg' : rawFormat
  const mimeType = format === 'png'
    ? 'image/png'
    : (format === 'webp' ? 'image/webp' : 'image/jpeg')
  const pathName = path.split('?')[0].split('/').filter(Boolean).pop() || ''
  const fileName = COVER_ALLOWED_FORMATS.includes(getFileExtension(pathName))
    ? pathName
    : `game-cover-${Date.now()}.${format}`

  return {
    path,
    fileName,
    mimeType,
    size: Number(file.size || 0) || 0
  }
}

function isAllowedMediaFormat(file = {}, mediaInfo = {}, allowedFormats = []) {
  const format = String(mediaInfo.type || file.type || '').toLowerCase()
  const extension = getFileExtension(file.tempFilePath || '')

  return allowedFormats.includes(format) || allowedFormats.includes(extension)
}

function formatMediaDuration(duration) {
  const totalSeconds = Math.max(0, Math.ceil(Number(duration) || 0))
  const minutes = Math.floor(totalSeconds / 60)
  const seconds = totalSeconds % 60

  return `${minutes}:${padNumber(seconds)}`
}

function createDescriptionMediaItem(file = {}, type, mediaInfo = {}) {
  const duration = Number(mediaInfo.duration || file.duration || 0)
  const tempFilePath = file.tempFilePath || ''

  return {
    id: `${Date.now()}-${Math.floor(Math.random() * 100000)}`,
    type,
    tempFilePath,
    thumbTempFilePath: file.thumbTempFilePath || tempFilePath,
    duration,
    durationText: duration ? formatMediaDuration(duration) : ''
  }
}

function normalizeProfitTemplate(item = {}, currentAccountType = '') {
  const key = String(item.key || item.code || item.id || '').trim()
  const isPublicBenefit = key === 'publicBenefit'
  const disabled = item.selectable === false
    || item.enabled === false
    || item.disabled === true
    || (isPublicBenefit && currentAccountType !== 'platform')

  if (!key) {
    return null
  }

  return {
    key,
    name: String(item.name || item.title || key).trim(),
    desc: String(item.desc || item.description || item.summary || item.splitText || '').trim(),
    disabled,
    disabledReason: String(item.disabledReason || item.reason || (isPublicBenefit && disabled ? '仅平台账户可发起' : '')).trim()
  }
}

function getProfitConfigText(data, keys = [], fallback = '') {
  const sources = []
  const isObject = data && typeof data === 'object' && !Array.isArray(data)

  if (isObject) {
    sources.push(data)
    ;['deposit', 'depositRule', 'depositConfig', 'ruleConfig'].forEach((key) => {
      if (data[key] && typeof data[key] === 'object') {
        sources.push(data[key])
      }
    })
  }

  const templates = Array.isArray(data)
    ? data
    : Array.isArray(data && data.templates)
      ? data.templates
      : []
  const depositTemplate = templates.find((item) => (
    item && ['deposit', 'depositGame'].includes(String(item.key || item.code || item.id || ''))
  ))

  if (depositTemplate) {
    sources.push(depositTemplate)
  }

  for (let sourceIndex = 0; sourceIndex < sources.length; sourceIndex += 1) {
    const source = sources[sourceIndex]

    for (let keyIndex = 0; keyIndex < keys.length; keyIndex += 1) {
      const value = String(source[keys[keyIndex]] || '').trim()

      if (value) {
        return value
      }
    }
  }

  return fallback
}

function normalizeProfitTemplates(data) {
  const source = Array.isArray(data)
    ? data
    : Array.isArray(data && data.templates)
      ? data.templates
      : []
  const currentAccountType = String(data && data.currentAccountType || '').trim()
  const templates = source
    .map((item) => normalizeProfitTemplate(item, currentAccountType))
    .filter(Boolean)

  return templates
}

function normalizeGameTypes(data) {
  const source = Array.isArray(data && data.primaryCategories) ? data.primaryCategories : []
  const list = source
    .filter((item) => item && item.visible !== false && item.key)
    .sort((left, right) => Number(left.order || 0) - Number(right.order || 0))
    .map((item) => ({
      key: String(item.key || '').trim(),
      name: String(item.name || item.key || '').trim(),
      tags: (Array.isArray(item.tags) ? item.tags : [])
        .map((tag) => ({
          key: String(tag.key || '').trim(),
          name: String(tag.name || tag.label || tag.key || '').trim(),
          active: Boolean(tag.active)
        }))
        .filter((tag) => tag.key && tag.name),
      children: (Array.isArray(item.children) ? item.children : [])
        .filter((child) => child && child.visible !== false && child.key)
        .sort((left, right) => Number(left.order || 0) - Number(right.order || 0))
        .map((child) => ({
          key: String(child.key || '').trim(),
          name: String(child.name || child.key || '').trim()
        }))
        .filter((child) => child.key && child.name)
    }))
    .filter((item) => item.key && item.name)

  return list
}

function normalizeCreateFormConfig(data = {}) {
  const source = data.createForm || {}
  const capacity = source.capacity || {}
  const normalizeOptions = (items) => (Array.isArray(items) ? items : [])
    .map((item) => ({
      key: String(item.key || '').trim(),
      name: String(item.name || item.label || item.key || '').trim(),
      active: Boolean(item.active)
    }))
    .filter((item) => item.key && item.name)

  const configuredFeeTypes = normalizeOptions(source.feeTypes)
  const feeTypeNames = new Map(configuredFeeTypes.map((item) => [item.key, item.name]))

  return {
    capacity: {
      min: Number(capacity.min || 0),
      max: Number(capacity.max || 0)
    },
    currentLocationText: String(source.currentLocationText || '').trim(),
    participationModes: normalizeOptions(source.participationModes),
    tags: normalizeOptions(source.tags),
    completionRules: normalizeOptions(source.completionRules),
    // 收费局一期不可创建，但入口需要稳定保留；只接受这两个一级选项，
    // 避免后台的分润/押金等二期配置直接进入小程序表单。
    feeTypes: PHASE_ONE_FEE_TYPES.map((item) => ({
      ...item,
      name: feeTypeNames.get(item.key) || item.name
    }))
  }
}

function getSelectedCategory(gameTypes = [], primaryKey = '', secondaryKey = '') {
  const primary = gameTypes.find((item) => item.key === primaryKey) || gameTypes[0] || {}
  const secondary = (Array.isArray(primary.children) ? primary.children : []).find((item) => item.key === secondaryKey)
    || (Array.isArray(primary.children) && primary.children.length ? primary.children[0] : {})

  return { primary, secondary }
}
function getSelectableProfitTemplateKey(templates = [], currentKey = '') {
  const current = templates.find((item) => item.key === currentKey)

  if (current && !current.disabled) {
    return current.key
  }

  const fallback = templates.find((item) => !item.disabled)

  return fallback ? fallback.key : ''
}

function activeOptionKeys(items = []) {
  return items
    .filter((item) => item && item.active)
    .map((item) => item.key || item.name)
    .filter(Boolean)
}

function normalizeConditionRuleConfig(data = {}) {
  const ruleItems = Array.isArray(data.ruleItems)
    ? data.ruleItems
      .filter((item) => item && item.key && item.name)
      .sort((left, right) => Number(left.order || 0) - Number(right.order || 0))
      .map((item) => ({
        key: String(item.key || '').trim(),
        name: String(item.name || '').trim(),
        description: String(item.description || '').trim(),
        required: item.required !== false
      }))
    : []

  return {
    enabled: Boolean(data.enabled),
    visibleInMiniProgram: Boolean(data.visibleInMiniProgram),
    adminOnlyCreate: data.adminOnlyCreate !== false,
    ruleItems,
    defaultVisibility: String(data.defaultVisibility || '').trim(),
    reviewRequired: Boolean(data.reviewRequired),
    paymentRequired: Boolean(data.paymentRequired)
  }
}

Page({
  data: {
    onlineText: '在线',
    createScrollTop: 0,
    coverUploadEnabled: featureFlags.gameCoverUploadEnabled,
    defaultGameCover: DEFAULT_GAME_COVER,
    coverImage: '',
    coverFile: null,
    draftId: '',
    draftCount: 0,
    publishDisabled: true,
    showProfitTemplate: false,
    themeMaxLength: THEME_MAX_LENGTH,
    themeLength: 0,
    introMaxLength: INTRO_MAX_LENGTH,
    introLength: 0,
    highlightsMaxLength: HIGHLIGHTS_MAX_LENGTH,
    highlightsLength: 0,
    noticeMaxLength: NOTICE_MAX_LENGTH,
    noticeLength: 0,
    audienceMaxLength: AUDIENCE_MAX_LENGTH,
    audienceLength: 0,
    descriptionMedia: [],
    createForm: EMPTY_CREATE_FORM,
    capacityMin: 0,
    capacityMax: 0,
    todayDate: formatDate(new Date()),
    timePanelVisible: false,
    timeDraft: getInitialTimeDraft(),
    gameTimeConfirmed: false,
    gameTimeSummary: {
      startText: '',
      endText: '',
      durationText: ''
    },
    signupTimePanelVisible: false,
    signupTimeDraft: getInitialSignupTimeDraft(),
    signupTimeConfirmed: false,
    signupTimeSummary: {
      startText: '',
      endText: '',
      durationText: ''
    },
    locationInfo: {
      name: '',
      address: '',
      cityCode: '',
      cityName: '',
      latitude: '',
      longitude: ''
    },
    locationFallbackInfo: {
      name: '',
      address: '',
      cityCode: '',
      cityName: '',
      latitude: '',
      longitude: ''
    },
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ],
    form: {
      theme: '',
      type: '',
      capacity: 0,
      participation: '',
      intro: '',
      highlights: '',
      description: '',
      notice: '',
      audience: '',
      feeType: '',
      price: '',
      profitTemplate: ''
    },
    gameTypes: EMPTY_GAME_TYPES,
    secondaryCategories: [],
    scheduleFields: [
      { key: 'gameTime', label: '局时间', required: true, value: '', hint: '必填：请选择开始和结束日期+时间' },
      { key: 'signupTime', label: '报名时间', required: true, value: '', hint: '展开日期面板，选择报名开始和结束日期+时间' },
      { key: 'location', label: '组局地址', required: false, value: '', hint: '点击在地图上标记位置' }
    ],
    participationModes: [],
    tags: [],
    completionRules: [],
    feeTypes: [],
    profitTemplates: EMPTY_PROFIT_TEMPLATES,
    conditionRuleConfig: EMPTY_CONDITION_RULE_CONFIG,
    conditionRuleItems: [],
    conditionRuleNotice: '',
    depositRuleText: EMPTY_DEPOSIT_RULE_TEXT,
    depositNoticeText: EMPTY_DEPOSIT_NOTICE_TEXT
  },

  onLoad(options = {}) {
    this.restoreRequestedDraft(options.draftId)
    const theme = String((this.data.form && this.data.form.theme) || '')
    const intro = String((this.data.form && this.data.form.intro) || '')
    const highlights = String((this.data.form && this.data.form.highlights) || '')
    const notice = String((this.data.form && this.data.form.notice) || '')
    const audience = String((this.data.form && this.data.form.audience) || '')

    this.setData({
      'form.theme': theme.slice(0, THEME_MAX_LENGTH),
      themeLength: Math.min(theme.length, THEME_MAX_LENGTH),
      'form.intro': intro.slice(0, INTRO_MAX_LENGTH),
      introLength: Math.min(intro.length, INTRO_MAX_LENGTH),
      'form.highlights': highlights.slice(0, HIGHLIGHTS_MAX_LENGTH),
      highlightsLength: Math.min(highlights.length, HIGHLIGHTS_MAX_LENGTH),
      'form.notice': notice.slice(0, NOTICE_MAX_LENGTH),
      noticeLength: Math.min(notice.length, NOTICE_MAX_LENGTH),
      'form.audience': audience.slice(0, AUDIENCE_MAX_LENGTH),
      audienceLength: Math.min(audience.length, AUDIENCE_MAX_LENGTH)
    })
    this.syncPublishState()
    this.loadCategoryConfig()
    this.loadConditionRuleConfig()
    this.loadProfitTemplates()
  },

  onShow() {
    this.loadDraftCount()
    this.handlePreviewPublishAction()
  },

  async loadDraftCount() {
    try {
      const data = await gameService.getGameDrafts()
      const items = Array.isArray(data && data.items) ? data.items : []
      this.setData({ draftCount: items.length })
    } catch (error) {
      // 草稿箱入口保持可用，具体错误在用户点击保存或打开草稿箱时提示。
    }
  },

  async restoreRequestedDraft(draftId) {
    const normalizedDraftId = String(draftId || '').trim()
    if (!normalizedDraftId) {
      return
    }

    let record
    try {
      record = await gameService.getGameDraft(normalizedDraftId)
    } catch (error) {
      toast.info(error.message || '草稿不存在或已删除')
      return
    }
    const payload = record && record.payload
    const draft = payload && typeof payload === 'object'
      ? { ...payload, id: record.id, title: record.title, savedAt: record.updatedAt }
      : null
    if (!draft || !draft.form) {
      toast.info('草稿内容异常，无法继续编辑')
      return
    }

    const timeDraft = draft.timeDraft || getInitialTimeDraft()
    const signupTimeDraft = draft.signupTimeDraft || getInitialSignupTimeDraft()
    const locationInfo = draft.locationInfo || {}
    const gameTimeConfirmed = draft.gameTimeConfirmed === true
    const signupTimeConfirmed = draft.signupTimeConfirmed === true
    const startTimestamp = getDraftTimestamp(timeDraft.startDate, timeDraft.startTime)
    const endTimestamp = getDraftTimestamp(timeDraft.endDate, timeDraft.endTime)
    const signupStartTimestamp = getDraftTimestamp(signupTimeDraft.startDate, signupTimeDraft.startTime)
    const signupEndTimestamp = getDraftTimestamp(signupTimeDraft.endDate, signupTimeDraft.endTime)

    this.pendingDraftSelections = draft
    this.setData({
      draftId: draft.id,
      coverImage: String(draft.coverImage || '').trim(),
      // 历史草稿可能来自收费局尚未关闭前；一期恢复草稿时统一回退为免费局，
      // 不展示局费、分润或押金配置，也不会把收费数据提交给后端。
      form: {
        ...this.data.form,
        ...draft.form,
        feeType: 'free',
        price: '',
        profitTemplate: ''
      },
      themeLength: Math.min(String(draft.form.theme || '').length, THEME_MAX_LENGTH),
      introLength: Math.min(String(draft.form.intro || '').length, INTRO_MAX_LENGTH),
      highlightsLength: Math.min(String(draft.form.highlights || '').length, HIGHLIGHTS_MAX_LENGTH),
      noticeLength: Math.min(String(draft.form.notice || '').length, NOTICE_MAX_LENGTH),
      audienceLength: Math.min(String(draft.form.audience || '').length, AUDIENCE_MAX_LENGTH),
      timeDraft,
      signupTimeDraft,
      gameTimeConfirmed,
      signupTimeConfirmed,
      locationInfo,
      descriptionMedia: Array.isArray(draft.descriptionMedia) ? draft.descriptionMedia : [],
      gameTimeSummary: gameTimeConfirmed ? {
        startText: `${timeDraft.startDate} ${timeDraft.startTime}`,
        endText: `${timeDraft.endDate} ${timeDraft.endTime}`,
        durationText: getDurationText(startTimestamp, endTimestamp)
      } : this.data.gameTimeSummary,
      signupTimeSummary: signupTimeConfirmed ? {
        startText: `${signupTimeDraft.startDate} ${signupTimeDraft.startTime}`,
        endText: `${signupTimeDraft.endDate} ${signupTimeDraft.endTime}`,
        durationText: getDurationText(signupStartTimestamp, signupEndTimestamp)
      } : this.data.signupTimeSummary
    })

    if (gameTimeConfirmed) {
      this.updateScheduleField('gameTime', getTimeDraftText(timeDraft))
    }
    if (signupTimeConfirmed) {
      this.updateScheduleField('signupTime', getTimeDraftText(signupTimeDraft))
    }
    this.updateScheduleField('location', String(locationInfo.name || locationInfo.address || '').trim())
    if (Array.isArray(this.data.gameTypes) && this.data.gameTypes.length) {
      this.applyPendingDraftSelections()
    }
  },

  applyPendingDraftSelections() {
    const draft = this.pendingDraftSelections
    if (!draft) {
      return
    }

    const activeTags = new Set((Array.isArray(draft.activeTags) ? draft.activeTags : []).slice(0, MAX_SELECTED_TAGS))
    const activeCompletionRules = new Set(Array.isArray(draft.activeCompletionRules) ? draft.activeCompletionRules : [])
    this.setData({
      tags: (this.data.tags || []).map((item) => ({ ...item, active: activeTags.has(item.key) })),
      completionRules: (this.data.completionRules || []).map((item) => ({ ...item, active: activeCompletionRules.has(item.key) }))
    }, () => this.syncPublishState())
    this.pendingDraftSelections = null
  },

  handlePreviewPublishAction() {
    try {
      const action = wx.getStorageSync(CREATE_PREVIEW_ACTION_KEY)
      if (!action || action.type !== 'publish') {
        return
      }
      wx.removeStorageSync(CREATE_PREVIEW_ACTION_KEY)
      setTimeout(() => this.publishGame(), 0)
    } catch (error) {
      // 预览操作标记读取失败时不影响创建页继续编辑。
    }
  },

  loadCategoryConfig() {
    gameService.getCategoryConfig().then((data) => {
      const gameTypes = normalizeGameTypes(data)
      const createForm = normalizeCreateFormConfig(data)
      const currentType = String(this.data.form && this.data.form.type || '').trim()
      const defaultType = String(data && data.defaultPrimaryCategory || '').trim()
      const currentSecondaryCategory = String(this.data.form && this.data.form.secondaryCategory || '').trim()
      const selected = getSelectedCategory(gameTypes, currentType || defaultType, currentSecondaryCategory)
      const nextData = {
        gameTypes,
        createForm,
        capacityMin: createForm.capacity.min,
        capacityMax: createForm.capacity.max,
        participationModes: createForm.participationModes,
        tags: selected.primary && selected.primary.tags && selected.primary.tags.length ? selected.primary.tags : createForm.tags,
        secondaryCategories: selected.primary && selected.primary.children ? selected.primary.children : [],
        completionRules: createForm.completionRules,
        feeTypes: createForm.feeTypes
      }

      if (!currentType && selected.primary && selected.primary.key) {
        nextData['form.type'] = selected.primary.key
      }
      if (!Number(this.data.form && this.data.form.capacity) && createForm.capacity.min) {
        nextData['form.capacity'] = createForm.capacity.min
      }
      if (!this.data.form.participation && createForm.participationModes[0]) {
        nextData['form.participation'] = createForm.participationModes[0].key
      }
      if (this.data.form.feeType !== 'free' && createForm.feeTypes[0]) {
        const freeFeeType = createForm.feeTypes.find((item) => item && item.key === 'free')
        nextData['form.feeType'] = (freeFeeType || createForm.feeTypes[0]).key
        nextData['form.price'] = ''
        nextData['form.profitTemplate'] = ''
      }

      if (selected.secondary && selected.secondary.key) {
        nextData['form.secondaryCategory'] = selected.secondary.key
      }

      this.setData(nextData)
      this.applyPendingDraftSelections()
      if (nextData['form.feeType']) {
        this.loadProfitTemplates(nextData['form.feeType'])
      }
      this.syncPublishState()
    }).catch((error) => {
      this.setData({
        gameTypes: EMPTY_GAME_TYPES,
        createForm: EMPTY_CREATE_FORM,
        capacityMin: 0,
        capacityMax: 0,
        participationModes: [],
        tags: [],
        secondaryCategories: [],
        completionRules: [],
        feeTypes: []
      })
      this.syncPublishState()
      toast.info(error.message || '组局类型配置加载失败')
    })
  },

  loadConditionRuleConfig() {
    gameService.getConditionRuleConfig().then((data) => {
      const config = normalizeConditionRuleConfig(data)

      this.setData({
        conditionRuleConfig: config,
        conditionRuleItems: config.enabled && config.visibleInMiniProgram ? config.ruleItems : [],
        conditionRuleNotice: this.formatConditionRuleNotice(config)
      })
    }).catch(() => {
      this.setData({
        conditionRuleConfig: EMPTY_CONDITION_RULE_CONFIG,
        conditionRuleItems: [],
        conditionRuleNotice: ''
      })
    })
  },

  formatConditionRuleNotice(config = EMPTY_CONDITION_RULE_CONFIG) {
    if (!config.enabled || !config.visibleInMiniProgram) {
      return ''
    }

    const parts = []
    if (config.adminOnlyCreate) {
      parts.push('条件局由后台开局')
    }
    if (config.reviewRequired) {
      parts.push('需平台审核')
    }
    if (config.paymentRequired) {
      parts.push('需满足付费条件')
    }
    if (config.defaultVisibility === 'invite_only') {
      parts.push('默认仅邀请可见')
    }

    return parts.join(' · ')
  },

  loadProfitTemplates(feeType) {
    const normalizedFeeType = feeType || (this.data.form && this.data.form.feeType)
    const shouldShowProfitTemplate = this.shouldRequireProfitTemplate(normalizedFeeType)

    if (!shouldShowProfitTemplate) {
      this.setData({
        profitTemplates: EMPTY_PROFIT_TEMPLATES,
        'form.profitTemplate': '',
        showProfitTemplate: false,
        depositRuleText: EMPTY_DEPOSIT_RULE_TEXT,
        depositNoticeText: EMPTY_DEPOSIT_NOTICE_TEXT
      })
      this.syncPublishState()
      return
    }

    gameService.getProfitTemplates({
      feeType: normalizedFeeType
    }).then((data) => {
      const profitTemplates = normalizeProfitTemplates(data)
      const profitTemplate = getSelectableProfitTemplateKey(
        profitTemplates,
        this.data.form && this.data.form.profitTemplate
      )
      const depositRuleText = getProfitConfigText(
        data,
        ['depositRuleText', 'ruleText', 'rule', 'depositRule'],
        EMPTY_DEPOSIT_RULE_TEXT
      )
      const depositNoticeText = getProfitConfigText(
        data,
        ['depositNoticeText', 'noticeText', 'tipText', 'notice', 'depositNotice'],
        EMPTY_DEPOSIT_NOTICE_TEXT
      )

      this.setData({
        profitTemplates,
        'form.profitTemplate': profitTemplate,
        showProfitTemplate: shouldShowProfitTemplate,
        depositRuleText,
        depositNoticeText
      })
      this.syncPublishState()
    }).catch(() => {
      this.setData({
        profitTemplates: EMPTY_PROFIT_TEMPLATES,
        'form.profitTemplate': '',
        showProfitTemplate: false,
        depositRuleText: EMPTY_DEPOSIT_RULE_TEXT,
        depositNoticeText: EMPTY_DEPOSIT_NOTICE_TEXT
      })
      this.syncPublishState()
      toast.info('收益模板配置加载失败')
    })
  },

  chooseCover() {
    if (this.data.coverImage) {
      wx.showActionSheet({
        itemList: ['替换图片', '删除图片'],
        success: (res) => {
          if (res.tapIndex === 0) {
            this.selectCoverImage()
            return
          }

          if (res.tapIndex === 1) {
            this.setData({
              coverImage: '',
              coverFile: null
            })
          }
        }
      })
      return
    }

    this.selectCoverImage()
  },

  onThemeInput(event) {
    const value = String(event.detail.value || '').slice(0, THEME_MAX_LENGTH)

    this.setData({
      'form.theme': value,
      themeLength: value.length
    }, () => this.syncPublishState())
  },

  selectCoverImage() {
    wx.chooseMedia({
      count: 1,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      sizeType: ['compressed'],
      success: (res) => {
        const file = res.tempFiles && res.tempFiles[0]
        const tempFilePath = file && file.tempFilePath

        if (tempFilePath) {
          this.validateAndSetCoverImage(file)
        }
      },
      fail: (error = {}) => {
        if (error.errMsg && error.errMsg.indexOf('cancel') > -1) {
          return
        }

        toast.info('图片选择失败')
      }
    })
  },

  validateAndSetCoverImage(file = {}) {
    wx.getImageInfo({
      src: file.tempFilePath,
      success: (imageInfo) => {
        if (!isAllowedCoverFormat(file, imageInfo)) {
          toast.info('封面仅支持 JPG/JPEG、PNG、WEBP')
          return
        }

        const coverFile = normalizeCoverUploadFile(file, imageInfo)
        if (!coverFile.size) {
          toast.info('无法读取封面大小，请重新选择图片')
          return
        }
        if (coverFile.size > COVER_MAX_SIZE) {
          toast.info('封面不能超过 10MB')
          return
        }

        this.setData({
          coverImage: file.tempFilePath,
          coverFile
        })
      },
      fail: () => {
        toast.info('图片校验失败')
      }
    })
  },

  selectType(event) {
    const key = String(event.currentTarget.dataset.key || '').trim()
    const selected = getSelectedCategory(this.data.gameTypes || EMPTY_GAME_TYPES, key)
    const tags = selected.primary && selected.primary.tags && selected.primary.tags.length
      ? selected.primary.tags
      : (this.data.createForm.tags || [])

    this.setData({
      'form.type': selected.primary.key || key,
      'form.secondaryCategory': selected.secondary.key || '',
      secondaryCategories: selected.primary.children || [],
      tags: tags.map((item) => ({ ...item, active: false }))
    }, () => this.syncPublishState())
  },

  selectSecondaryCategory(event) {
    const key = String(event.currentTarget.dataset.key || '').trim()
    const selected = getSelectedCategory(this.data.gameTypes || EMPTY_GAME_TYPES, this.data.form.type, key)
    if (!selected.secondary || !selected.secondary.key) {
      return
    }
    this.setData({
      'form.secondaryCategory': selected.secondary.key
    })
  },

  chooseSchedule(event) {
    const key = event.currentTarget.dataset.key

    if (key === 'gameTime') {
      const nextData = {
        timePanelVisible: !this.data.timePanelVisible,
        signupTimePanelVisible: false
      }

      if (!this.data.timePanelVisible && isDraftBeforeNow(this.data.timeDraft)) {
        nextData.timeDraft = getInitialTimeDraft()
      }

      this.setData(nextData)
      return
    }

    if (key === 'signupTime') {
      const nextData = {
        signupTimePanelVisible: !this.data.signupTimePanelVisible,
        timePanelVisible: false
      }

      if (!this.data.signupTimePanelVisible && isDraftBeforeNow(this.data.signupTimeDraft)) {
        nextData.signupTimeDraft = getInitialSignupTimeDraft()
      }

      this.setData(nextData)
      return
    }

    if (key === 'location') {
      this.setData({
        timePanelVisible: false,
        signupTimePanelVisible: false
      })
      this.chooseGameLocation()
      return
    }

    toast.info('请选择有效的局属性')
  },

  chooseGameLocation() {
    const selectedLocation = this.data.locationInfo || {}
    const fallbackLocation = this.data.locationFallbackInfo || {}
    const locationInfo = (typeof selectedLocation.latitude === 'number' && typeof selectedLocation.longitude === 'number')
      ? selectedLocation
      : fallbackLocation

    if (typeof locationInfo.latitude === 'number' && typeof locationInfo.longitude === 'number') {
      this.openNativeGameLocation(locationInfo)
      return
    }

    if (!wx.getLocation) {
      this.openNativeGameLocation()
      return
    }

    // Request location only after the user chooses to select an address.
    // A refusal still falls back to the native map, where manual search works.
    wx.getLocation({
      type: 'gcj02',
      success: (res = {}) => {
        const currentLocation = {
          latitude: typeof res.latitude === 'number' ? res.latitude : '',
          longitude: typeof res.longitude === 'number' ? res.longitude : ''
        }
        this.setData({ locationFallbackInfo: currentLocation })
        this.openNativeGameLocation(currentLocation)
      },
      fail: () => this.openNativeGameLocation()
    })
  },

  openNativeGameLocation(locationInfo = {}) {
    const locationOptions = {}

    if (typeof locationInfo.latitude === 'number' && typeof locationInfo.longitude === 'number') {
      locationOptions.latitude = locationInfo.latitude
      locationOptions.longitude = locationInfo.longitude
    }

    wx.chooseLocation({
      ...locationOptions,
      success: (res = {}) => {
        const nextLocationInfo = {
          name: res.name || '',
          address: res.address || '',
          latitude: typeof res.latitude === 'number' ? res.latitude : '',
          longitude: typeof res.longitude === 'number' ? res.longitude : ''
        }
        const text = String(nextLocationInfo.name || nextLocationInfo.address || '已选择位置').trim()
        this.updateScheduleField('location', text)
        this.setData({ locationInfo: nextLocationInfo })
      },
      fail: (error = {}) => {
        if (error.errMsg && error.errMsg.indexOf('cancel') > -1) {
          return
        }
        toast.info('地图选点暂不可用，请稍后重试')
      }
    })
  },

  onGameTimePickerChange(event) {
    const field = event.currentTarget.dataset.field
    const value = event.detail.value

    if (!field) {
      return
    }

    this.setData({
      [`timeDraft.${field}`]: value
    })
  },

  cancelGameTimePanel() {
    this.setData({
      timePanelVisible: false
    })
  },

  confirmGameTimePanel() {
    const draft = this.data.timeDraft || {}
    const startTimestamp = getDraftTimestamp(draft.startDate, draft.startTime)
    const endTimestamp = getDraftTimestamp(draft.endDate, draft.endTime)
    if (!startTimestamp || !endTimestamp || Number.isNaN(startTimestamp) || Number.isNaN(endTimestamp)) {
      toast.info('请选择完整局时间')
      return
    }

    if (endTimestamp <= startTimestamp) {
      toast.info('结束时间需晚于开始时间')
      return
    }

    if (this.data.signupTimeConfirmed) {
      const signupDraft = this.data.signupTimeDraft || {}
      const signupEndTimestamp = getDraftTimestamp(signupDraft.endDate, signupDraft.endTime)

      if (!Number.isNaN(signupEndTimestamp) && signupEndTimestamp >= startTimestamp) {
        toast.info('报名结束时间需早于局开始时间')
        return
      }
    }

    this.updateScheduleField('gameTime', getTimeDraftText(draft))
    this.setData({
      timePanelVisible: false,
      gameTimeConfirmed: true,
      gameTimeSummary: {
        startText: `${draft.startDate} ${draft.startTime}`,
        endText: `${draft.endDate} ${draft.endTime}`,
        durationText: getDurationText(startTimestamp, endTimestamp)
      }
    }, () => this.syncPublishState())
  },

  updateScheduleField(key, value) {
    const scheduleFields = this.data.scheduleFields.map((item) => (
      item.key === key ? { ...item, value } : item
    ))

    this.setData({ scheduleFields })
  },

  onSignupTimePickerChange(event) {
    const field = event.currentTarget.dataset.field
    const value = event.detail.value

    if (!field) {
      return
    }

    this.setData({
      [`signupTimeDraft.${field}`]: value
    })
  },

  cancelSignupTimePanel() {
    this.setData({
      signupTimePanelVisible: false
    })
  },

  confirmSignupTimePanel() {
    const draft = this.data.signupTimeDraft || {}
    const signupStartTimestamp = getDraftTimestamp(draft.startDate, draft.startTime)
    const signupEndTimestamp = getDraftTimestamp(draft.endDate, draft.endTime)

    if (!signupStartTimestamp || !signupEndTimestamp || Number.isNaN(signupStartTimestamp) || Number.isNaN(signupEndTimestamp)) {
      toast.info('请选择完整报名时间')
      return
    }

    if (signupEndTimestamp <= signupStartTimestamp) {
      toast.info('报名结束时间需晚于报名开始时间')
      return
    }

    if (this.data.gameTimeConfirmed) {
      const gameStartTimestamp = getDraftTimestamp(this.data.timeDraft.startDate, this.data.timeDraft.startTime)

      if (!Number.isNaN(gameStartTimestamp) && signupEndTimestamp >= gameStartTimestamp) {
        toast.info('报名结束时间需早于局开始时间')
        return
      }
    }

    const text = getTimeDraftText(draft)

    this.updateScheduleField('signupTime', text)
    this.setData({
      signupTimePanelVisible: false,
      signupTimeConfirmed: true,
      signupTimeSummary: {
        startText: `${draft.startDate} ${draft.startTime}`,
        endText: `${draft.endDate} ${draft.endTime}`,
        durationText: getDurationText(signupStartTimestamp, signupEndTimestamp)
      }
    })
  },

  onCapacityChange(event) {
    this.setData({
      'form.capacity': clampCapacity(event.detail.value, this.data.createForm.capacity)
    }, () => this.syncPublishState())
  },

  onCapacityInput(event) {
    this.setData({
      'form.capacity': clampCapacity(event.detail.value, this.data.createForm.capacity)
    }, () => this.syncPublishState())
  },

  selectParticipation(event) {
    this.setData({
      'form.participation': event.currentTarget.dataset.key
    })
  },

  toggleTag(event) {
    const index = Number(event.currentTarget.dataset.index)
    const current = this.data.tags || []
    const target = current[index]
    if (!target) {
      return
    }
    if (!target.active && current.filter((item) => item && item.active).length >= MAX_SELECTED_TAGS) {
      toast.info(`标签最多选择${MAX_SELECTED_TAGS}个`)
      return
    }
    const tags = current.map((item, itemIndex) => ({
      ...item,
      active: itemIndex === index ? !item.active : item.active
    }))

    this.setData({ tags })
  },

  onIntroInput(event) {
    const value = String(event.detail.value || '').slice(0, INTRO_MAX_LENGTH)

    this.setData({
      'form.intro': value,
      introLength: value.length
    })
  },

  onHighlightsInput(event) {
    const value = String(event.detail.value || '').slice(0, HIGHLIGHTS_MAX_LENGTH)

    this.setData({
      'form.highlights': value,
      highlightsLength: value.length
    })
  },

  onDescriptionInput(event) {
    this.setData({
      'form.description': event.detail.value
    })
  },

  chooseDescriptionMedia(event) {
    const type = event.currentTarget.dataset.type

    if (type === 'image') {
      this.chooseDescriptionImages()
      return
    }

    if (type === 'video') {
      this.chooseDescriptionVideo()
    }
  },

  chooseDescriptionImages() {
    const remainImageCount = DESCRIPTION_IMAGE_MAX_COUNT - this.getDescriptionImageCount()

    if (remainImageCount <= 0) {
      toast.info('图片最多上传9张')
      return
    }

    wx.chooseMedia({
      count: remainImageCount,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      sizeType: ['compressed'],
      success: (res = {}) => {
        this.validateAndAppendDescriptionMedia(res.tempFiles || [], 'image')
      },
      fail: (error = {}) => {
        if (error.errMsg && error.errMsg.indexOf('cancel') > -1) {
          return
        }

        toast.info('图片选择失败')
      }
    })
  },

  chooseDescriptionVideo() {
    wx.chooseMedia({
      count: 1,
      mediaType: ['video'],
      sourceType: ['album', 'camera'],
      maxDuration: DESCRIPTION_VIDEO_MAX_DURATION,
      success: (res = {}) => {
        this.validateAndAppendDescriptionMedia(res.tempFiles || [], 'video')
      },
      fail: (error = {}) => {
        if (error.errMsg && error.errMsg.indexOf('cancel') > -1) {
          return
        }

        toast.info('视频选择失败')
      }
    })
  },

  validateAndAppendDescriptionMedia(files = [], expectedType = '') {
    if (!files.length) {
      return
    }

    const currentMedia = this.data.descriptionMedia || []
    let nextMedia = currentMedia.slice()
    let pending = files.length
    let hasInvalid = false

    const finish = () => {
      pending -= 1

      if (pending > 0) {
        return
      }

      this.setData({
        descriptionMedia: nextMedia
      })

      if (hasInvalid) {
        toast.info('部分媒体不符合格式或数量限制')
      }
    }

    files.forEach((file) => {
      const mediaType = expectedType || file.fileType || file.type || ''

      if (mediaType === 'video') {
        this.validateDescriptionVideo(file, (item) => {
          if (item) {
            nextMedia = nextMedia.concat(item)
          } else {
            hasInvalid = true
          }

          finish()
        })
        return
      }

      if (this.getMediaImageCount(nextMedia) >= DESCRIPTION_IMAGE_MAX_COUNT) {
        hasInvalid = true
        finish()
        return
      }

      this.validateDescriptionImage(file, (item) => {
        if (item) {
          nextMedia = nextMedia.concat(item)
        } else {
          hasInvalid = true
        }

        finish()
      })
    })
  },

  validateDescriptionImage(file = {}, callback) {
    wx.getImageInfo({
      src: file.tempFilePath,
      success: (imageInfo) => {
        if (!isAllowedMediaFormat(file, imageInfo, DESCRIPTION_ALLOWED_IMAGE_FORMATS)) {
          callback(null)
          return
        }

        callback(createDescriptionMediaItem(file, 'image', imageInfo))
      },
      fail: () => {
        callback(null)
      }
    })
  },

  validateDescriptionVideo(file = {}, callback) {
    const handleVideoInfo = (videoInfo = {}) => {
      const duration = Number(videoInfo.duration || file.duration || 0)

      if (!isAllowedMediaFormat(file, videoInfo, DESCRIPTION_ALLOWED_VIDEO_FORMATS)) {
        callback(null)
        return
      }

      if (duration > DESCRIPTION_VIDEO_MAX_DURATION) {
        callback(null)
        return
      }

      callback(createDescriptionMediaItem(file, 'video', {
        ...videoInfo,
        duration
      }))
    }

    if (!wx.getVideoInfo) {
      handleVideoInfo(file)
      return
    }

    wx.getVideoInfo({
      src: file.tempFilePath,
      success: handleVideoInfo,
      fail: () => {
        handleVideoInfo(file)
      }
    })
  },

  removeDescriptionMedia(event) {
    const id = event.currentTarget.dataset.id
    const descriptionMedia = (this.data.descriptionMedia || []).filter((item) => item.id !== id)

    this.setData({ descriptionMedia })
  },

  getDescriptionImageCount() {
    return this.getMediaImageCount(this.data.descriptionMedia || [])
  },

  getMediaImageCount(media = []) {
    return media.filter((item) => item.type === 'image').length
  },

  onNoticeInput(event) {
    const value = String(event.detail.value || '').slice(0, NOTICE_MAX_LENGTH)

    this.setData({
      'form.notice': value,
      noticeLength: value.length
    })
  },

  onAudienceInput(event) {
    const value = String(event.detail.value || '').slice(0, AUDIENCE_MAX_LENGTH)

    this.setData({
      'form.audience': value,
      audienceLength: value.length
    })
  },

  toggleCompletionRule(event) {
    const index = Number(event.currentTarget.dataset.index)
    const completionRules = this.data.completionRules.map((item, itemIndex) => ({
      ...item,
      active: itemIndex === index ? !item.active : item.active
    }))

    this.setData({ completionRules }, () => this.syncPublishState())
  },

  selectFeeType(event) {
    const feeType = event.currentTarget.dataset.key
    if (feeType && feeType !== 'free') {
      toast.info('收费局暂未开放')
      return
    }
    const nextData = {
      'form.feeType': feeType
    }

    if (feeType === 'free') {
      nextData['form.price'] = ''
    }

    if (feeType === 'free') {
      nextData['form.profitTemplate'] = ''
    }

    this.setData(nextData, () => {
      this.syncPublishState()
      this.loadProfitTemplates(feeType)
    })
  },

  onPriceInput(event) {
    if (this.data.form && this.data.form.feeType !== 'paid') {
      return
    }

    this.setData({
      'form.price': event.detail.value
    })
  },

  selectProfitTemplate(event) {
    const key = event.currentTarget.dataset.key
    const template = (this.data.profitTemplates || []).find((item) => item.key === key)

    if (!template || template.disabled) {
      toast.info(template && template.disabledReason ? template.disabledReason : '当前模板不可选择')
      return
    }

    this.setData({
      'form.profitTemplate': key
    }, () => this.syncPublishState())
  },

  buildDraftSnapshot() {
    const form = this.data.form || {}
    const title = String(form.theme || '').trim()

    return {
      title: title || '未命名组局',
      savedAt: Date.now(),
      coverImage: String(this.data.coverImage || '').trim(),
      form: { ...form },
      timeDraft: { ...(this.data.timeDraft || {}) },
      signupTimeDraft: { ...(this.data.signupTimeDraft || {}) },
      gameTimeConfirmed: this.data.gameTimeConfirmed === true,
      signupTimeConfirmed: this.data.signupTimeConfirmed === true,
      locationInfo: { ...(this.data.locationInfo || {}) },
      descriptionMedia: Array.isArray(this.data.descriptionMedia) ? this.data.descriptionMedia.slice() : [],
      activeTags: activeOptionKeys(this.data.tags),
      activeCompletionRules: activeOptionKeys(this.data.completionRules),
      tagLabels: (this.data.tags || []).filter((item) => item.active).map((item) => item.name),
      completionRuleLabels: (this.data.completionRules || []).filter((item) => item.active).map((item) => item.name),
      typeText: String((getSelectedCategory(this.data.gameTypes || EMPTY_GAME_TYPES, form.type, form.secondaryCategory).primary || {}).name || '').trim(),
      participationText: String(((this.data.participationModes || []).find((item) => item.key === form.participation) || {}).name || '').trim(),
      feeTypeText: String(((this.data.feeTypes || []).find((item) => item.key === form.feeType) || {}).name || '').trim()
    }
  },

  async saveDraft() {
    const snapshot = this.buildDraftSnapshot()
    try {
      const saved = await gameService.saveGameDraft({
        id: this.data.draftId,
        title: snapshot.title,
        payload: snapshot
      })
      this.setData({ draftId: saved.id })
      await this.loadDraftCount()
      toast.success('已保存到服务器草稿箱')
    } catch (error) {
      toast.info(error.message || '草稿保存失败，请稍后重试')
      return
    }
  },

  previewSubmit() {
    const preview = this.buildDraftSnapshot()

    try {
      wx.setStorageSync(CREATE_PREVIEW_STORAGE_KEY, preview)
      navigateShellRoute(`/${ROUTES.gameCreatePreview}`, {
        currentRoute: ROUTES.gameCreate,
        reuseExisting: false
      })
    } catch (error) {
      toast.info('预览打开失败，请稍后重试')
    }
  },

  openDraftBox() {
    navigateShellRoute(`/${ROUTES.gameCreateDrafts}`, {
      currentRoute: ROUTES.gameCreate,
      reuseExisting: false
    })
  },

  async publishGame() {
    const missing = this.getPublishMissingFields()
    if (missing.length) {
      toast.info(this.getPublishBlockedMessage(missing))
      this.syncPublishState()
      return
    }

    const payload = this.buildCreateGamePayload()
    if (!payload) {
      toast.info('表单状态异常，请重新确认局类型、人数和局时间')
      return
    }

    wx.showLoading({ title: '发布中', mask: true })
    let usedDefaultCover = false
    try {
      const coverImage = String(this.data.coverImage || '').trim()
      const isWechatTempCover = /^https?:\/\/tmp\//i.test(coverImage)
      if (!featureFlags.gameCoverUploadEnabled) {
        payload.coverFileId = 0
        payload.coverImage = DEFAULT_GAME_COVER
        usedDefaultCover = true
      } else if (coverImage && (isWechatTempCover || (!/^https?:\/\//i.test(coverImage) && !/^mock:\/\//i.test(coverImage)))) {
        const coverFile = this.data.coverFile && this.data.coverFile.path === coverImage
          ? this.data.coverFile
          : coverImage
        payload.coverFileId = await fileService.uploadSingleFile(coverFile, {
          bizType: 'game_cover',
          objectId: 0
        })
        payload.coverImage = ''
        if (!payload.coverFileId) {
          throw new Error('封面上传失败')
        }
      }
      const game = await gameService.createGame(payload)
      wx.hideLoading()
      toast.success(usedDefaultCover ? '发布成功，暂用默认封面' : '已提交后台审核')
      setTimeout(() => {
        navigateShellRoute(`${ROUTES.gameDetail}?id=${game.id}`, {
          currentRoute: ROUTES.gameCreate
        })
      }, 500)
    } catch (error) {
      wx.hideLoading()
      const errorMessage = String(error && error.message || '')
      toast.info(/invalid file request/i.test(errorMessage)
        ? '封面上传失败：仅支持 JPG/JPEG、PNG、WEBP，且不能超过 10MB'
        : (errorMessage || '发布失败'))
    }
  },

  getPublishBlockedMessage(missing = this.getPublishMissingFields()) {
    if (!missing.length) {
      return '信息已填写完整，请重新点击发布'
    }

    return `请先完善：${missing.slice(0, 3).join('、')}`
  },

  getPublishMissingFields() {
    const form = this.data.form || {}
    const capacity = this.data.createForm.capacity || {}
    const missing = []

    if (!Array.isArray(this.data.gameTypes) || !this.data.gameTypes.length) {
      missing.push('局类型配置')
    }
    if (!String(form.theme || '').trim()) {
      missing.push('局主题')
    }
    if (!String(form.type || '').trim()) {
      missing.push('局类型')
    }
    if (Number(form.capacity) < Number(capacity.min || 0) || Number(form.capacity) > Number(capacity.max || 0)) {
      missing.push('人数')
    }
    if (!this.data.gameTimeConfirmed) {
      missing.push('局时间')
    }
    if (!Array.isArray(this.data.completionRules) || !this.data.completionRules.some((item) => item.active)) {
      missing.push('完成规则')
    }
    if (!Array.isArray(this.data.feeTypes) || !this.data.feeTypes.some((item) => item.key === form.feeType)) {
      missing.push('收费类型')
    }
    if (this.shouldRequireProfitTemplate(form.feeType) && !this.hasSelectableProfitTemplate(form.profitTemplate)) {
      missing.push('分润模板')
    }

    return missing
  },

  shouldRequireProfitTemplate(feeType) {
    return Boolean(feeType) && feeType !== 'free'
  },

  hasSelectableProfitTemplate(profitTemplate) {
    return Array.isArray(this.data.profitTemplates)
      && this.data.profitTemplates.some((item) => item.key === profitTemplate && !item.disabled)
  },

  buildCreateGamePayload() {
    const form = this.data.form || {}
    const locationInfo = this.data.locationInfo || {}
    const title = String(form.theme || '').trim()
    const capacity = this.data.createForm.capacity || {}
    const minPlayers = Number(capacity.min || 0)
    const maxCapacity = Number(capacity.max || 0)
    const maxPlayers = clampCapacity(form.capacity, capacity)

    if (!title || !String(form.type || '').trim() || maxPlayers < minPlayers || maxPlayers > maxCapacity || !this.data.gameTimeConfirmed) {
      return null
    }

    const gameTypes = this.data.gameTypes || EMPTY_GAME_TYPES
    if (!gameTypes.length) {
      toast.info('组局类型配置加载失败')
      return null
    }

    const selected = getSelectedCategory(gameTypes, form.type, form.secondaryCategory)
    const primary = selected.primary || {}
    const secondary = selected.secondary || {}

    return {
      title,
      gameType: 'free',
      coverImage: this.data.coverImage || '',
      description: String(form.intro || '').trim(),
      highlights: String(form.highlights || '').trim(),
      notice: String(form.notice || '').trim(),
      audience: String(form.audience || '').trim(),
      participation: form.participation || '',
      startAt: this.data.gameTimeConfirmed ? `${this.data.timeDraft.startDate} ${this.data.timeDraft.startTime}` : '',
      endAt: this.data.gameTimeConfirmed ? `${this.data.timeDraft.endDate} ${this.data.timeDraft.endTime}` : '',
      signupStartAt: this.data.signupTimeConfirmed ? `${this.data.signupTimeDraft.startDate} ${this.data.signupTimeDraft.startTime}` : '',
      signupEndAt: this.data.signupTimeConfirmed ? `${this.data.signupTimeDraft.endDate} ${this.data.signupTimeDraft.endTime}` : '',
      tags: activeOptionKeys(this.data.tags),
      completionRules: activeOptionKeys(this.data.completionRules),
      primaryCategory: primary.key || form.type || '',
      primaryCategoryText: primary.name || '',
      secondaryCategory: form.secondaryCategory || secondary.key || '',
      secondaryCategoryText: secondary.name || '',
      type: 'free',
      price: 0,
      profitTemplate: '',
      minPlayers,
      maxPlayers,
      cityCode: locationInfo.cityCode || '',
      cityName: locationInfo.cityName || locationInfo.name || '',
      address: locationInfo.address || '',
      longitude: typeof locationInfo.longitude === 'number' ? locationInfo.longitude : 0,
      latitude: typeof locationInfo.latitude === 'number' ? locationInfo.latitude : 0
    }
  },

  syncPublishState() {
    const form = this.data.form || {}
    const hasGameTypes = Array.isArray(this.data.gameTypes) && this.data.gameTypes.length > 0
    const hasTheme = String(form.theme || '').trim().length > 0
    const hasType = String(form.type || '').trim().length > 0
    const capacity = this.data.createForm.capacity || {}
    const hasCapacity = Number(form.capacity) >= Number(capacity.min || 0) && Number(form.capacity) <= Number(capacity.max || 0)
    const hasGameTime = Boolean(this.data.gameTimeConfirmed)
    const hasCompletionRule = Array.isArray(this.data.completionRules) && this.data.completionRules.some((item) => item.active)
    const hasFeeType = Array.isArray(this.data.feeTypes) && this.data.feeTypes.some((item) => item.key === form.feeType)
    const hasProfitTemplate = !this.shouldRequireProfitTemplate(form.feeType)
      || this.hasSelectableProfitTemplate(form.profitTemplate)

    this.setData({
      publishDisabled: !(hasGameTypes && hasTheme && hasType && hasCapacity && hasGameTime && hasCompletionRule && hasFeeType && hasProfitTemplate)
    })
  },

  handleShellNavTap(event) {
    const key = event.detail && event.detail.key

    if (key === 'up' || key === 'down') {
      if (!this.suppressNextNavTap) {
        this.scrollCreate(key, CREATE_SCROLL_TAP_STEP_RPX)
      }
      return
    }

    if (navigateShellKey(key, {
      currentRoute: ROUTES.gameCreate
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
    this.stopCreateScrollHold(false)
    this.scrollCreate(key, CREATE_SCROLL_HOLD_STEP_RPX)

    this.createScrollHoldTimer = setInterval(() => {
      this.scrollCreate(key, CREATE_SCROLL_HOLD_STEP_RPX)
    }, CREATE_SCROLL_HOLD_INTERVAL_MS)
  },

  handleShellNavTouchEnd() {
    this.stopCreateScrollHold(true)
  },

  handleShellAction(key) {
    if (key === 'home') {
      this.navigateToRoute(ROUTES.playerHome || ROUTES.home)
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
    if (!route || route === ROUTES.gameCreate) {
      return
    }

    navigateShellRoute(route)
  },

  handleCreateScroll(event) {
    const scrollTop = event.detail && event.detail.scrollTop

    if (typeof scrollTop === 'number') {
      this.createScrollTopValue = scrollTop
    }
  },

  scrollCreate(direction, stepRpx = CREATE_SCROLL_TAP_STEP_RPX) {
    const current = Number(this.createScrollTopValue || this.data.createScrollTop || 0)
    const distance = this.rpxToPx(stepRpx)
    const nextTop = direction === 'up'
      ? Math.max(0, current - distance)
      : current + distance

    this.createScrollTopValue = nextTop
    this.setData({
      createScrollTop: nextTop
    })
  },

  scrollCreateToTop() {
    this.createScrollTopValue = 0
    this.setData({
      createScrollTop: 0
    })
  },

  stopCreateScrollHold(resetTapSuppress) {
    if (this.createScrollHoldTimer) {
      clearInterval(this.createScrollHoldTimer)
      this.createScrollHoldTimer = null
    }

    if (resetTapSuppress && this.suppressNextNavTap) {
      if (this.createScrollSuppressTimer) {
        clearTimeout(this.createScrollSuppressTimer)
      }

      this.createScrollSuppressTimer = setTimeout(() => {
        this.suppressNextNavTap = false
        this.createScrollSuppressTimer = null
      }, CREATE_SCROLL_HOLD_SUPPRESS_TAP_MS)
    }
  },

  clearCreateScrollTimers() {
    this.stopCreateScrollHold(false)

    if (this.createScrollSuppressTimer) {
      clearTimeout(this.createScrollSuppressTimer)
      this.createScrollSuppressTimer = null
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
    this.clearCreateScrollTimers()
  }
})
