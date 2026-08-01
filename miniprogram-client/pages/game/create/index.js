const toast = require('../../../utils/toast')
const gameService = require('../../../services/game')
const fileService = require('../../../services/file')
const mapService = require('../../../services/map')
const featureFlags = require('./feature-flags')
const { ROUTES } = require('../../../config/routes')
const { navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')

const CREATE_SCROLL_TAP_STEP_RPX = 360
const CREATE_SCROLL_HOLD_STEP_RPX = 72
const CREATE_SCROLL_HOLD_INTERVAL_MS = 80
const CREATE_SCROLL_HOLD_SUPPRESS_TAP_MS = 120
const COVER_ALLOWED_FORMATS = ['jpg', 'jpeg', 'png', 'webp']
const COVER_MAX_SIZE = 10 * 1024 * 1024
const COVER_CROP_SCALE = 5 / 3
const COVER_MIN_WIDTH = 750
const COVER_MIN_HEIGHT = 450
const DEFAULT_GAME_COVER = 'https://static.haowan.net.cn/miniprogram/components/game-card/assets/game-cover-default.png'
const DESCRIPTION_ALLOWED_IMAGE_FORMATS = ['jpg', 'jpeg', 'png']
const DESCRIPTION_ALLOWED_VIDEO_FORMATS = ['mp4', 'mov']
const DESCRIPTION_IMAGE_MAX_COUNT = 9
const DESCRIPTION_VIDEO_MAX_COUNT = 1
const DESCRIPTION_VIDEO_MAX_DURATION = 60
const THEME_MAX_LENGTH = 20
const INTRO_MAX_LENGTH = 200
const HIGHLIGHTS_MAX_LENGTH = 100
const DESCRIPTION_MAX_LENGTH = 1000
const NOTICE_MAX_LENGTH = 100
const AUDIENCE_MAX_LENGTH = 50
const EMPTY_GAME_TYPES = []
const MAX_SELECTED_TAGS = 3
const PARTICIPANT_ROLE_OPTIONS = [
  { key: 'player', name: '玩家' },
  { key: 'expert', name: '行家' },
  { key: 'guide', name: '领路人' }
]

function buildParticipantRoleOptions(roles) {
  const selected = new Set(Array.isArray(roles) ? roles : ['player'])
  return PARTICIPANT_ROLE_OPTIONS.map((item) => ({ ...item, active: selected.has(item.key) }))
}
const CREATE_PREVIEW_STORAGE_KEY = 'game_create_preview_v1'
const CREATE_PREVIEW_ACTION_KEY = 'game_create_preview_action_v1'
const CREATE_PREVIEW_ACTION_MAX_AGE_MS = 10 * 60 * 1000
const EMPTY_CREATE_FORM = {
  capacity: { min: 0, max: 0 },
  currentLocationText: '',
  participationModes: [],
  tags: [],
  completionRules: []
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

function isValidTimeDraft(draft = {}) {
  const startTimestamp = getDraftTimestamp(draft.startDate, draft.startTime)
  const endTimestamp = getDraftTimestamp(draft.endDate, draft.endTime)

  return Number.isFinite(startTimestamp)
    && Number.isFinite(endTimestamp)
    && endTimestamp > startTimestamp
}

function isValidSignupTimeDraft(signupDraft = {}, gameDraft = {}) {
  if (!isValidTimeDraft(signupDraft) || !isValidTimeDraft(gameDraft)) {
    return false
  }

  const signupEndTimestamp = getDraftTimestamp(signupDraft.endDate, signupDraft.endTime)
  const gameStartTimestamp = getDraftTimestamp(gameDraft.startDate, gameDraft.startTime)

  return signupEndTimestamp < gameStartTimestamp
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

function hasConfirmedGameLocation(location = {}) {
  const address = String(location.address || location.name || '').trim()
  const latitude = Number(location.latitude)
  const longitude = Number(location.longitude)

  return Boolean(address)
    && Number.isFinite(latitude)
    && Number.isFinite(longitude)
    && !(latitude === 0 && longitude === 0)
}

function normalizeGameLocation(location = {}) {
  const rawLatitude = location.latitude
  const rawLongitude = location.longitude
  const latitude = rawLatitude === '' || rawLatitude === null || rawLatitude === undefined ? NaN : Number(rawLatitude)
  const longitude = rawLongitude === '' || rawLongitude === null || rawLongitude === undefined ? NaN : Number(rawLongitude)

  return {
    name: String(location.name || '').trim(),
    address: String(location.address || location.name || '').trim(),
    cityCode: String(location.cityCode || '').trim(),
    cityName: String(location.cityName || '').trim(),
    latitude: Number.isFinite(latitude) ? latitude : '',
    longitude: Number.isFinite(longitude) ? longitude : ''
  }
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
    size: Number(file.size || 0) || 1,
    duration,
    durationText: duration ? formatMediaDuration(duration) : ''
  }
}

function normalizeDraftDescriptionMedia(items = []) {
  return (Array.isArray(items) ? items : []).map((source, index) => {
    const item = source && typeof source === 'object' ? source : {}
    const type = item.type === 'video' ? 'video' : 'image'
    const fileId = Number(item.fileId || 0)
    const duration = type === 'video' ? Math.max(0, Number(item.duration || 0)) : 0

    return {
      ...item,
      id: String(item.id || `draft-media-${fileId > 0 ? fileId : index}`),
      type,
      fileId: Number.isInteger(fileId) && fileId > 0 ? fileId : 0,
      url: String(item.url || '').trim(),
      tempFilePath: String(item.tempFilePath || item.path || '').trim(),
      thumbTempFilePath: String(item.thumbTempFilePath || '').trim(),
      size: Math.max(1, Number(item.size || 1)),
      duration,
      durationText: duration ? formatMediaDuration(duration) : ''
    }
  })
}

function serializeDraftDescriptionMedia(items = []) {
  return normalizeDraftDescriptionMedia(items).map((item) => {
    const serialized = {
      id: item.id,
      type: item.type,
      fileId: item.fileId,
      url: item.url,
      size: item.size,
      duration: item.duration,
      durationText: item.durationText
    }
    // 已上传媒体只保存服务器文件编号和当前下载地址，避免把仅在本机
    // 有效的 wxfile/http://tmp 路径写进服务器草稿。
    if (item.fileId <= 0) {
      serialized.tempFilePath = item.tempFilePath
      serialized.thumbTempFilePath = item.thumbTempFilePath
    }
    return serialized
  })
}

function descriptionMediaMimeType(item = {}) {
  const extension = getFileExtension(item.tempFilePath || item.path || '')
  if (item.type === 'video') {
    return extension === 'mov' ? 'video/quicktime' : 'video/mp4'
  }
  if (extension === 'png') return 'image/png'
  if (extension === 'webp') return 'image/webp'
  return 'image/jpeg'
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
        .filter((tag) => tag.key && tag.name)
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

  return {
    capacity: {
      min: Number(capacity.min || 0),
      max: Number(capacity.max || 0)
    },
    currentLocationText: String(source.currentLocationText || '').trim(),
    participationModes: normalizeOptions(source.participationModes),
    tags: normalizeOptions(source.tags),
    completionRules: normalizeOptions(source.completionRules)
  }
}

function getSelectedCategory(gameTypes = [], primaryKey = '') {
  const primary = gameTypes.find((item) => item.key === primaryKey) || gameTypes[0] || {}

  return { primary }
}

function findSelectedCategory(gameTypes = [], primaryKey = '') {
  const primary = gameTypes.find((item) => item.key === primaryKey) || null

  return { primary }
}

function createPreviewSessionId() {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
}

function createFlowErrorMessage(error, fallback) {
  const raw = String(error && (error.message || error.errMsg) || '').trim()
  if (/request:fail|network|timeout|timed out|socket/i.test(raw)) {
    return '网络连接异常，请检查网络后重试'
  }
  if (/uploadFile:fail/i.test(raw)) {
    return '图片或视频上传失败，请检查网络后重试'
  }
  return raw || fallback
}

function activeOptionKeys(items = []) {
  return items
    .filter((item) => item && item.active)
    .map((item) => item.key || item.name)
    .filter(Boolean)
}

function withScheduleFieldValue(fields = [], key, value) {
  return fields.map((item) => (
    item.key === key ? { ...item, value } : item
  ))
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
    createScrollIntoView: '',
    coverUploadEnabled: featureFlags.gameCoverUploadEnabled,
    defaultGameCover: DEFAULT_GAME_COVER,
    coverImage: '',
    coverFile: null,
    coverFileId: 0,
    draftId: '',
    draftCount: 0,
    draftSaving: false,
    previewPreparing: false,
    publishing: false,
    publishDisabled: true,
    themeMaxLength: THEME_MAX_LENGTH,
    themeLength: 0,
    introMaxLength: INTRO_MAX_LENGTH,
    introLength: 0,
    highlightsMaxLength: HIGHLIGHTS_MAX_LENGTH,
    highlightsLength: 0,
    descriptionMaxLength: DESCRIPTION_MAX_LENGTH,
    noticeMaxLength: NOTICE_MAX_LENGTH,
    noticeLength: 0,
    audienceMaxLength: AUDIENCE_MAX_LENGTH,
    audienceLength: 0,
    descriptionMedia: [],
    descriptionMediaPendingCount: 0,
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
      allowedRoles: ['player'],
      intro: '',
      highlights: '',
      description: '',
      notice: '',
      audience: ''
    },
    gameTypes: EMPTY_GAME_TYPES,
    scheduleFields: [
      { key: 'gameTime', label: '局时间', required: true, value: '', hint: '必填：请选择开始和结束日期+时间' },
      { key: 'signupTime', label: '报名时间', required: true, value: '', hint: '展开日期面板，选择报名开始和结束日期+时间' },
      { key: 'location', label: '组局地址', required: true, value: '', hint: '点击在地图上标记位置' }
    ],
    participationModes: [],
    participantRoleOptions: buildParticipantRoleOptions(['player']),
    tags: [],
    completionRules: []
  },

  onLoad(options = {}) {
    this.createSubmissionSessionId = createPreviewSessionId()
    this.draftSubmissionSessionId = createPreviewSessionId()
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
    this.initializeCreatePage(options.draftId)
  },

  async initializeCreatePage(draftId) {
    // 先固定分类、人数和选项配置，再恢复草稿。否则两个接口响应顺序
    // 不同时，分类默认值会覆盖草稿，或草稿标签会套用到错误的大类。
    await this.loadCategoryConfig()
    await this.restoreRequestedDraft(draftId)
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
      return false
    }

    const restoreSequence = Number(this.draftRestoreSequence || 0) + 1
    this.draftRestoreSequence = restoreSequence
    let record
    try {
      record = await gameService.getGameDraft(normalizedDraftId)
    } catch (error) {
      if (restoreSequence !== this.draftRestoreSequence) {
        return false
      }
      toast.info(createFlowErrorMessage(error, '草稿不存在或已删除'))
      return false
    }
    if (restoreSequence !== this.draftRestoreSequence) {
      return false
    }
    const payload = record && record.payload
    const draft = payload && typeof payload === 'object' && !Array.isArray(payload)
      ? { ...payload, id: record.id, title: record.title, savedAt: record.updatedAt }
      : null
    if (!draft || !draft.form || typeof draft.form !== 'object' || Array.isArray(draft.form)) {
      toast.info('草稿内容异常，无法继续编辑')
      return false
    }

    const timeDraft = draft.timeDraft || getInitialTimeDraft()
    const signupTimeDraft = draft.signupTimeDraft || getInitialSignupTimeDraft()
    const locationInfo = normalizeGameLocation(draft.locationInfo || {})
    const gameTimeConfirmed = draft.gameTimeConfirmed === true && isValidTimeDraft(timeDraft)
    const signupTimeConfirmed = draft.signupTimeConfirmed === true && isValidSignupTimeDraft(signupTimeDraft, timeDraft)
    const startTimestamp = getDraftTimestamp(timeDraft.startDate, timeDraft.startTime)
    const endTimestamp = getDraftTimestamp(timeDraft.endDate, timeDraft.endTime)
    const signupStartTimestamp = getDraftTimestamp(signupTimeDraft.startDate, signupTimeDraft.startTime)
    const signupEndTimestamp = getDraftTimestamp(signupTimeDraft.endDate, signupTimeDraft.endTime)
    const rawAllowedRoles = draft.form.allowedRoles
    const allowedRoles = Array.isArray(rawAllowedRoles)
      ? rawAllowedRoles.filter((role, index, items) => (
        PARTICIPANT_ROLE_OPTIONS.some((option) => option.key === role) && items.indexOf(role) === index
      ))
      : (Array.isArray(this.data.form.allowedRoles) ? this.data.form.allowedRoles : ['player'])
    const restoredType = String(draft.form.type || '').trim()
    const normalizedType = findSelectedCategory(this.data.gameTypes || EMPTY_GAME_TYPES, restoredType).primary
      ? restoredType
      : ''
    const normalizedForm = {
      ...this.data.form,
      ...draft.form,
      theme: String(draft.form.theme || '').slice(0, THEME_MAX_LENGTH),
      type: normalizedType,
      intro: String(draft.form.intro || '').slice(0, INTRO_MAX_LENGTH),
      highlights: String(draft.form.highlights || '').slice(0, HIGHLIGHTS_MAX_LENGTH),
      description: String(draft.form.description || '').slice(0, DESCRIPTION_MAX_LENGTH),
      notice: String(draft.form.notice || '').slice(0, NOTICE_MAX_LENGTH),
      audience: String(draft.form.audience || '').slice(0, AUDIENCE_MAX_LENGTH),
      allowedRoles
    }
    const refreshedFiles = await this.refreshDraftFileURLs(draft)
    if (restoreSequence !== this.draftRestoreSequence) {
      return false
    }

    this.gameTimeEditBackup = null
    this.signupTimeEditBackup = null
    this.activePreviewSessionId = ''
    this.descriptionMediaUpdateChain = Promise.resolve()
    this.pendingDraftSelections = { ...draft, form: normalizedForm }
    this.setData({
      draftId: draft.id,
      coverImage: refreshedFiles.coverImage,
      coverFileId: Number(draft.coverFileId || 0),
      coverFile: null,
      form: normalizedForm,
      participantRoleOptions: buildParticipantRoleOptions(allowedRoles),
      themeLength: normalizedForm.theme.length,
      introLength: normalizedForm.intro.length,
      highlightsLength: normalizedForm.highlights.length,
      noticeLength: normalizedForm.notice.length,
      audienceLength: normalizedForm.audience.length,
      timeDraft,
      signupTimeDraft,
      timePanelVisible: false,
      signupTimePanelVisible: false,
      gameTimeConfirmed,
      signupTimeConfirmed,
      locationInfo,
      descriptionMedia: refreshedFiles.descriptionMedia,
      gameTimeSummary: gameTimeConfirmed ? {
        startText: `${timeDraft.startDate} ${timeDraft.startTime}`,
        endText: `${timeDraft.endDate} ${timeDraft.endTime}`,
        durationText: getDurationText(startTimestamp, endTimestamp)
      } : { startText: '', endText: '', durationText: '' },
      signupTimeSummary: signupTimeConfirmed ? {
        startText: `${signupTimeDraft.startDate} ${signupTimeDraft.startTime}`,
        endText: `${signupTimeDraft.endDate} ${signupTimeDraft.endTime}`,
        durationText: getDurationText(signupStartTimestamp, signupEndTimestamp)
      } : { startText: '', endText: '', durationText: '' },
      scheduleFields: this.data.scheduleFields.map((item) => ({ ...item, value: '' }))
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
    if (refreshedFiles.refreshFailed) {
      toast.info('草稿中的部分图片或视频暂时无法加载，请检查后再发布')
    } else if (restoredType && !normalizedType) {
      toast.info('草稿中的局类型已调整，请重新选择局类型')
    }
    this.syncPublishState()
    return true
  },

  async refreshDraftFileURLs(draft = {}) {
    let refreshFailed = false
    let coverImage = String(draft.coverImage || '').trim()
    const coverFileId = Number(draft.coverFileId || 0)
    if (coverFileId > 0) {
      try {
        coverImage = String(await fileService.getDownloadURL(coverFileId) || '').trim()
      } catch (error) {
        refreshFailed = true
      }
    }

    const descriptionMedia = await Promise.all(normalizeDraftDescriptionMedia(draft.descriptionMedia).map(async (item) => {
      if (item.fileId <= 0) {
        return item
      }
      try {
        return { ...item, url: String(await fileService.getDownloadURL(item.fileId) || '').trim() }
      } catch (error) {
        refreshFailed = true
        return item
      }
    }))

    return { coverImage, descriptionMedia, refreshFailed }
  },

  applyPendingDraftSelections() {
    const draft = this.pendingDraftSelections
    if (!draft) {
      return
    }

    const activeTags = new Set((Array.isArray(draft.activeTags) ? draft.activeTags : []).slice(0, MAX_SELECTED_TAGS))
    const activeCompletionRules = new Set(Array.isArray(draft.activeCompletionRules) ? draft.activeCompletionRules : [])
    const selected = findSelectedCategory(this.data.gameTypes || EMPTY_GAME_TYPES, String(draft.form && draft.form.type || '').trim())
    const categoryTags = selected.primary && selected.primary.tags && selected.primary.tags.length
      ? selected.primary.tags
      : (this.data.createForm.tags || [])
    this.setData({
      tags: categoryTags.map((item) => ({ ...item, active: activeTags.has(item.key) })),
      completionRules: (this.data.completionRules || []).map((item) => ({ ...item, active: activeCompletionRules.has(item.key) }))
    }, () => this.syncPublishState())
    this.pendingDraftSelections = null
  },

  handlePreviewPublishAction() {
    try {
      const action = wx.getStorageSync(CREATE_PREVIEW_ACTION_KEY)
      wx.removeStorageSync(CREATE_PREVIEW_ACTION_KEY)
      const createdAt = Number(action && action.createdAt || 0)
      const age = Date.now() - createdAt
      const sessionId = String(action && action.sessionId || '')
      if (!action || action.type !== 'publish'
        || !sessionId
        || sessionId !== String(this.activePreviewSessionId || '')
        || !Number.isFinite(age)
        || age < 0
        || age > CREATE_PREVIEW_ACTION_MAX_AGE_MS) {
        return
      }
      this.activePreviewSessionId = ''
      setTimeout(() => this.publishGame(), 0)
    } catch (error) {
      // 预览操作标记读取失败时不影响创建页继续编辑。
    }
  },

  loadCategoryConfig() {
    return gameService.getCategoryConfig().then((data) => {
      const gameTypes = normalizeGameTypes(data)
      const createForm = normalizeCreateFormConfig(data)
      const currentType = String(this.data.form && this.data.form.type || '').trim()
      const defaultType = String(data && data.defaultPrimaryCategory || '').trim()
      const selected = getSelectedCategory(gameTypes, currentType || defaultType)
      const nextData = {
        gameTypes,
        createForm,
        capacityMin: createForm.capacity.min,
        capacityMax: createForm.capacity.max,
        participationModes: createForm.participationModes,
        tags: selected.primary && selected.primary.tags && selected.primary.tags.length ? selected.primary.tags : createForm.tags,
        completionRules: createForm.completionRules
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
      this.setData(nextData)
      this.applyPendingDraftSelections()
      this.syncPublishState()
    }).catch((error) => {
      this.setData({
        gameTypes: EMPTY_GAME_TYPES,
        createForm: EMPTY_CREATE_FORM,
        capacityMin: 0,
        capacityMax: 0,
        participationModes: [],
        tags: [],
        completionRules: []
      })
      this.syncPublishState()
      toast.info(createFlowErrorMessage(error, '组局类型配置加载失败'))
    })
  },

  chooseCover() {
    if (this.isCreateOperationBusy()) {
      toast.info('当前操作尚未完成，请稍候')
      return
    }
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
              coverFile: null,
              coverFileId: 0
            }, () => this.syncPublishState())
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
    if (this.isCreateOperationBusy()) {
      toast.info('当前操作尚未完成，请稍候')
      return
    }
    wx.chooseMedia({
      count: 1,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      sizeType: ['compressed'],
      success: (res) => {
        const file = res.tempFiles && res.tempFiles[0]
        const tempFilePath = file && file.tempFilePath

        if (tempFilePath) {
          this.cropAndSetCoverImage(file)
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

  cropAndSetCoverImage(file = {}) {
    const sourcePath = String(file.tempFilePath || '').trim()
    if (!sourcePath) {
      toast.info('图片读取失败，请重新选择')
      return
    }

    const continueWithPath = (tempFilePath) => {
      const croppedFile = { ...file, tempFilePath, size: 0 }
      if (typeof wx.getFileInfo !== 'function') {
        this.validateAndSetCoverImage(croppedFile)
        return
      }
      wx.getFileInfo({
        filePath: tempFilePath,
        success: (info = {}) => this.validateAndSetCoverImage({ ...croppedFile, size: Number(info.size || 0) }),
        fail: () => this.validateAndSetCoverImage(croppedFile)
      })
    }

    if (typeof wx.cropImage !== 'function') {
      toast.info('当前微信版本不支持封面裁剪，请升级后重试')
      return
    }

    wx.cropImage({
      src: sourcePath,
      cropScale: COVER_CROP_SCALE,
      success: (result = {}) => continueWithPath(String(result.tempFilePath || '').trim()),
      fail: (error = {}) => {
        if (String(error.errMsg || '').includes('cancel')) {
          return
        }
        toast.info('封面裁剪失败，请重新选择图片')
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

        const width = Number(imageInfo.width || 0)
        const height = Number(imageInfo.height || 0)
        const cropRatio = width / height
        if (width < COVER_MIN_WIDTH || height < COVER_MIN_HEIGHT || Math.abs(cropRatio - COVER_CROP_SCALE) > 0.02) {
          toast.info('封面需裁剪为 5:3 横图，且不小于 750 × 450')
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
          coverFile,
          coverFileId: 0
        }, () => this.syncPublishState())
      },
      fail: () => {
        toast.info('图片校验失败')
      }
    })
  },

  selectType(event) {
    const key = String(event.currentTarget.dataset.key || '').trim()
    const selected = findSelectedCategory(this.data.gameTypes || EMPTY_GAME_TYPES, key)
    if (!selected.primary) {
      toast.info('局类型已更新，请重新选择')
      return
    }
    const tags = selected.primary && selected.primary.tags && selected.primary.tags.length
      ? selected.primary.tags
      : (this.data.createForm.tags || [])

    this.setData({
      'form.type': selected.primary.key || key,
      tags: tags.map((item) => ({ ...item, active: false }))
    }, () => this.syncPublishState())
  },

  chooseSchedule(event) {
    const key = event.currentTarget.dataset.key

    if (key === 'gameTime') {
      if (this.data.timePanelVisible) {
        this.cancelGameTimePanel()
        return
      }
      const openPanel = () => this.openGameTimePanel()
      if (this.data.signupTimePanelVisible) {
        this.cancelSignupTimePanel(openPanel)
      } else {
        openPanel()
      }
      return
    }

    if (key === 'signupTime') {
      if (this.data.signupTimePanelVisible) {
        this.cancelSignupTimePanel()
        return
      }
      const openPanel = () => this.openSignupTimePanel()
      if (this.data.timePanelVisible) {
        this.cancelGameTimePanel(openPanel)
      } else {
        openPanel()
      }
      return
    }

    if (key === 'location') {
      const chooseLocation = () => this.chooseGameLocation()
      if (this.data.timePanelVisible) {
        this.cancelGameTimePanel(chooseLocation)
      } else if (this.data.signupTimePanelVisible) {
        this.cancelSignupTimePanel(chooseLocation)
      } else {
        chooseLocation()
      }
      return
    }

    toast.info('请选择有效的局属性')
  },

  openGameTimePanel() {
    const field = (this.data.scheduleFields || []).find((item) => item.key === 'gameTime') || {}
    this.gameTimeEditBackup = {
      draft: { ...(this.data.timeDraft || {}) },
      confirmed: Boolean(this.data.gameTimeConfirmed),
      summary: { ...(this.data.gameTimeSummary || {}) },
      value: String(field.value || '')
    }
    const resetExpiredDraft = isDraftBeforeNow(this.data.timeDraft)
    this.setData({
      timePanelVisible: true,
      signupTimePanelVisible: false,
      timeDraft: resetExpiredDraft ? getInitialTimeDraft() : this.data.timeDraft,
      gameTimeConfirmed: resetExpiredDraft ? false : this.data.gameTimeConfirmed,
      scheduleFields: resetExpiredDraft
        ? withScheduleFieldValue(this.data.scheduleFields, 'gameTime', '')
        : this.data.scheduleFields
    }, () => this.syncPublishState())
  },

  openSignupTimePanel() {
    const field = (this.data.scheduleFields || []).find((item) => item.key === 'signupTime') || {}
    this.signupTimeEditBackup = {
      draft: { ...(this.data.signupTimeDraft || {}) },
      confirmed: Boolean(this.data.signupTimeConfirmed),
      summary: { ...(this.data.signupTimeSummary || {}) },
      value: String(field.value || '')
    }
    const resetExpiredDraft = isDraftBeforeNow(this.data.signupTimeDraft)
    this.setData({
      signupTimePanelVisible: true,
      timePanelVisible: false,
      signupTimeDraft: resetExpiredDraft ? getInitialSignupTimeDraft() : this.data.signupTimeDraft,
      signupTimeConfirmed: resetExpiredDraft ? false : this.data.signupTimeConfirmed,
      scheduleFields: resetExpiredDraft
        ? withScheduleFieldValue(this.data.scheduleFields, 'signupTime', '')
        : this.data.scheduleFields
    }, () => this.syncPublishState())
  },

  chooseGameLocation() {
    const selectedLocation = normalizeGameLocation(this.data.locationInfo || {})
    const fallbackLocation = normalizeGameLocation(this.data.locationFallbackInfo || {})
    const hasSelectedCoordinate = Number.isFinite(Number(selectedLocation.latitude))
      && Number.isFinite(Number(selectedLocation.longitude))
      && !(Number(selectedLocation.latitude) === 0 && Number(selectedLocation.longitude) === 0)
    const locationInfo = hasSelectedCoordinate
      ? selectedLocation
      : fallbackLocation

    if (Number.isFinite(Number(locationInfo.latitude)) && Number.isFinite(Number(locationInfo.longitude))
      && !(Number(locationInfo.latitude) === 0 && Number(locationInfo.longitude) === 0)) {
      this.openNativeGameLocation(locationInfo)
      return
    }

    if (!wx.getLocation) {
      this.openNativeGameLocation()
      return
    }

    if (typeof wx.getSetting === 'function') {
      wx.getSetting({
        success: (setting = {}) => {
          const authSetting = setting.authSetting || {}
          if (authSetting['scope.userLocation'] === false) {
            this.showLocationDeniedGuide()
            return
          }
          this.requestCurrentLocationForGameMap()
        },
        fail: () => this.requestCurrentLocationForGameMap()
      })
      return
    }
    this.requestCurrentLocationForGameMap()
  },

  requestCurrentLocationForGameMap() {
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

  showLocationDeniedGuide() {
    wx.showModal({
      title: '开启定位权限',
      content: '开启后可在地图中自动定位当前位置；也可以不授权，直接手动搜索地点。',
      confirmText: '去设置',
      cancelText: '手动选点',
      success: (result = {}) => {
        if (!result.confirm || typeof wx.openSetting !== 'function') {
          this.openNativeGameLocation()
          return
        }
        wx.openSetting({
          success: (setting = {}) => {
            const enabled = Boolean(setting.authSetting && setting.authSetting['scope.userLocation'])
            if (enabled) {
              this.requestCurrentLocationForGameMap()
              return
            }
            this.openNativeGameLocation()
          },
          fail: () => this.openNativeGameLocation()
        })
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
        const nextLocationInfo = normalizeGameLocation({
          name: res.name || '',
          address: res.address || res.name || '',
          latitude: typeof res.latitude === 'number' ? res.latitude : '',
          longitude: typeof res.longitude === 'number' ? res.longitude : ''
        })
        if (!hasConfirmedGameLocation(nextLocationInfo)) {
          toast.info('未获取到有效地址，请重新选择位置')
          return
        }
        const text = String(nextLocationInfo.name || nextLocationInfo.address).trim()
        this.updateScheduleField('location', text)
        this.setData({ locationInfo: nextLocationInfo }, () => {
          this.syncPublishState()
          this.resolveGameLocationCity(nextLocationInfo)
        })
      },
      fail: (error = {}) => {
        if (error.errMsg && error.errMsg.indexOf('cancel') > -1) {
          return
        }
        toast.info('地图选点暂不可用，请稍后重试')
      }
    })
  },

  async resolveGameLocationCity(locationInfo = {}) {
    const latitude = Number(locationInfo.latitude)
    const longitude = Number(locationInfo.longitude)
    if (!Number.isFinite(latitude) || !Number.isFinite(longitude) || (latitude === 0 && longitude === 0)) {
      return
    }

    try {
      const result = await mapService.reverseGeocode({ latitude, longitude })
      const place = result && result.place ? result.place : result || {}
      const cityName = String(place.city || '').trim()
      const cityCode = String(place.cityCode || '').trim()
      if (!cityName && !cityCode) {
        return
      }
      const current = this.data.locationInfo || {}
      if (Number(current.latitude) !== latitude || Number(current.longitude) !== longitude) {
        return
      }
      this.setData({
        locationInfo: {
          ...current,
          cityName: cityName || current.cityName || '',
          cityCode: cityCode || current.cityCode || ''
        }
      })
    } catch (error) {
      // 原生地图已返回可发布的地点；城市归属获取失败时保留用户已选择的位置。
    }
  },

  onGameTimePickerChange(event) {
    const field = event.currentTarget.dataset.field
    const value = event.detail.value

    if (!field) {
      return
    }

    this.setData({
      [`timeDraft.${field}`]: value,
      gameTimeConfirmed: false,
      scheduleFields: withScheduleFieldValue(this.data.scheduleFields, 'gameTime', '')
    }, () => this.syncPublishState())
  },

  cancelGameTimePanel(callback) {
    const backup = this.gameTimeEditBackup
    this.gameTimeEditBackup = null
    const nextData = { timePanelVisible: false }
    if (backup) {
      nextData.timeDraft = backup.draft
      nextData.gameTimeConfirmed = backup.confirmed
      nextData.gameTimeSummary = backup.summary
      nextData.scheduleFields = withScheduleFieldValue(this.data.scheduleFields, 'gameTime', backup.value)
    }
    this.setData(nextData, () => {
      this.syncPublishState()
      if (typeof callback === 'function') {
        callback()
      }
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

    this.gameTimeEditBackup = null
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
    this.setData({ scheduleFields: withScheduleFieldValue(this.data.scheduleFields, key, value) })
  },

  onSignupTimePickerChange(event) {
    const field = event.currentTarget.dataset.field
    const value = event.detail.value

    if (!field) {
      return
    }

    this.setData({
      [`signupTimeDraft.${field}`]: value,
      signupTimeConfirmed: false,
      scheduleFields: withScheduleFieldValue(this.data.scheduleFields, 'signupTime', '')
    }, () => this.syncPublishState())
  },

  cancelSignupTimePanel(callback) {
    const backup = this.signupTimeEditBackup
    this.signupTimeEditBackup = null
    const nextData = { signupTimePanelVisible: false }
    if (backup) {
      nextData.signupTimeDraft = backup.draft
      nextData.signupTimeConfirmed = backup.confirmed
      nextData.signupTimeSummary = backup.summary
      nextData.scheduleFields = withScheduleFieldValue(this.data.scheduleFields, 'signupTime', backup.value)
    }
    this.setData(nextData, () => {
      this.syncPublishState()
      if (typeof callback === 'function') {
        callback()
      }
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

    this.signupTimeEditBackup = null
    this.updateScheduleField('signupTime', text)
    this.setData({
      signupTimePanelVisible: false,
      signupTimeConfirmed: true,
      signupTimeSummary: {
        startText: `${draft.startDate} ${draft.startTime}`,
        endText: `${draft.endDate} ${draft.endTime}`,
        durationText: getDurationText(signupStartTimestamp, signupEndTimestamp)
      }
    }, () => this.syncPublishState())
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

  toggleAllowedRole(event) {
    const role = String(event.currentTarget.dataset.key || '').trim()
    if (!PARTICIPANT_ROLE_OPTIONS.some((item) => item.key === role)) {
      return
    }
    const current = Array.isArray(this.data.form.allowedRoles) ? this.data.form.allowedRoles : []
    const selected = current.includes(role)
    if (selected && current.length === 1) {
      toast.info('至少选择一种可参与身份')
      return
    }
    const allowedRoles = selected
      ? current.filter((item) => item !== role)
      : [...current, role]
    this.setData({
      'form.allowedRoles': allowedRoles,
      participantRoleOptions: buildParticipantRoleOptions(allowedRoles)
    }, () => this.syncPublishState())
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
    }, () => this.syncPublishState())
  },

  onHighlightsInput(event) {
    const value = String(event.detail.value || '').slice(0, HIGHLIGHTS_MAX_LENGTH)

    this.setData({
      'form.highlights': value,
      highlightsLength: value.length
    }, () => this.syncPublishState())
  },

  onDescriptionInput(event) {
    const value = String(event.detail.value || '').slice(0, DESCRIPTION_MAX_LENGTH)
    this.setData({
      'form.description': value
    }, () => this.syncPublishState())
  },

  chooseDescriptionMedia(event) {
    if (this.isCreateOperationBusy()) {
      toast.info('当前操作尚未完成，请稍候')
      return
    }
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
    if (this.getDescriptionVideoCount() >= DESCRIPTION_VIDEO_MAX_COUNT) {
      toast.info('视频最多上传1个')
      return
    }
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

    let pending = files.length
    let hasInvalid = false
    this.setData({
      descriptionMediaPendingCount: Number(this.data.descriptionMediaPendingCount || 0) + pending
    })

    const finish = (item = null) => {
      pending -= 1

      if (pending > 0) {
        return
      }

      this.setData({
        descriptionMediaPendingCount: Math.max(0, Number(this.data.descriptionMediaPendingCount || 0) - files.length)
      })

      if (hasInvalid) {
        toast.info('部分媒体不符合格式或数量限制')
      }
    }

    const append = (item) => {
      if (!item) {
        hasInvalid = true
        finish()
        return
      }
      this.appendDescriptionMedia(item).then((appended) => {
        if (!appended) {
          hasInvalid = true
        }
        finish()
      }).catch(() => {
        hasInvalid = true
        finish()
      })
    }

    files.forEach((file) => {
      const mediaType = expectedType || file.fileType || file.type || ''

      if (mediaType === 'video') {
        this.validateDescriptionVideo(file, (item) => {
          append(item)
        })
        return
      }

      this.validateDescriptionImage(file, (item) => {
        append(item)
      })
    })
  },

  appendDescriptionMedia(item) {
    const previous = this.descriptionMediaUpdateChain || Promise.resolve()
    const next = previous.catch(() => undefined).then(() => new Promise((resolve) => {
      const currentMedia = Array.isArray(this.data.descriptionMedia) ? this.data.descriptionMedia : []
      if (item.type === 'image' && this.getMediaImageCount(currentMedia) >= DESCRIPTION_IMAGE_MAX_COUNT) {
        resolve(false)
        return
      }
      if (item.type === 'video' && this.getMediaVideoCount(currentMedia) >= DESCRIPTION_VIDEO_MAX_COUNT) {
        resolve(false)
        return
      }
      this.setData({ descriptionMedia: currentMedia.concat(item) }, () => resolve(true))
    }))

    this.descriptionMediaUpdateChain = next
    return next
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
    if (this.isCreateOperationBusy()) {
      toast.info('当前操作尚未完成，请稍候')
      return
    }
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

  getDescriptionVideoCount() {
    return this.getMediaVideoCount(this.data.descriptionMedia || [])
  },

  getMediaVideoCount(media = []) {
    return media.filter((item) => item.type === 'video').length
  },

  isCreateOperationBusy() {
    return Boolean(this.data.draftSaving || this.data.previewPreparing || this.data.publishing)
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

  onPaidGameTap() {
    toast.info('收费局暂未开放')
  },

  buildDraftSnapshot() {
    const form = this.data.form || {}
    const title = String(form.theme || '').trim()
    const selectedCategory = findSelectedCategory(this.data.gameTypes || EMPTY_GAME_TYPES, form.type)

    return {
      title: title || '未命名组局',
      savedAt: Date.now(),
      coverImage: String(this.data.coverImage || '').trim(),
      coverFileId: Number(this.data.coverFileId || 0),
      form: { ...form },
      timeDraft: { ...(this.data.timeDraft || {}) },
      signupTimeDraft: { ...(this.data.signupTimeDraft || {}) },
      gameTimeConfirmed: this.data.gameTimeConfirmed === true,
      signupTimeConfirmed: this.data.signupTimeConfirmed === true,
      locationInfo: { ...(this.data.locationInfo || {}) },
      descriptionMedia: serializeDraftDescriptionMedia(this.data.descriptionMedia),
      activeTags: activeOptionKeys(this.data.tags),
      activeCompletionRules: activeOptionKeys(this.data.completionRules),
      tagLabels: (this.data.tags || []).filter((item) => item.active).map((item) => item.name),
      completionRuleLabels: (this.data.completionRules || []).filter((item) => item.active).map((item) => item.name),
      typeText: String((selectedCategory.primary || {}).name || '').trim(),
      participationText: String(((this.data.participationModes || []).find((item) => item.key === form.participation) || {}).name || '').trim()
    }
  },

  async saveDraft() {
    if (this.isCreateOperationBusy()) {
      toast.info(this.data.draftSaving ? '草稿正在保存，请稍候' : '当前操作尚未完成，请稍候')
      return
    }
    if (this.data.descriptionMediaPendingCount > 0) {
      toast.info('图片或视频仍在处理中，请稍候再保存草稿')
      return
    }
    this.setData({ draftSaving: true })
    try {
      await this.persistCoverForDraft()
      const persistedMedia = await this.persistDescriptionMedia()
      await new Promise((resolve) => this.setData({ descriptionMedia: persistedMedia }, resolve))
      const snapshot = this.buildDraftSnapshot()
      const saved = await gameService.saveGameDraft({
        id: this.data.draftId,
        submissionSessionId: this.draftSubmissionSessionId,
        title: snapshot.title,
        payload: snapshot
      })
      this.setData({ draftId: saved.id })
      await this.loadDraftCount()
      toast.success('已保存到服务器草稿箱')
    } catch (error) {
      toast.info(createFlowErrorMessage(error, '草稿保存失败，请稍后重试'))
    } finally {
      this.setData({ draftSaving: false })
    }
  },

  async previewSubmit() {
    if (this.isCreateOperationBusy()) {
      toast.info(this.data.previewPreparing ? '预览正在生成，请稍候' : '当前操作尚未完成，请稍候')
      return
    }
    if (this.data.descriptionMediaPendingCount > 0) {
      toast.info('图片或视频仍在处理中，请稍候再预览')
      return
    }

    this.setData({ previewPreparing: true })
    try {
      // 预览必须使用已落库的媒体地址。临时图片路径在页面跳转后可能
      // 失效，导致同一组图片和视频只显示其中一部分。
      wx.showLoading({ title: '正在生成预览', mask: true })
      await this.persistCoverForDraft()
      const persistedMedia = await this.persistDescriptionMedia()
      await new Promise((resolve) => this.setData({ descriptionMedia: persistedMedia }, resolve))
      const preview = this.buildDraftSnapshot()
      const previewSessionId = createPreviewSessionId()
      preview.previewSessionId = previewSessionId
      this.activePreviewSessionId = previewSessionId
      wx.removeStorageSync(CREATE_PREVIEW_ACTION_KEY)
      wx.setStorageSync(CREATE_PREVIEW_STORAGE_KEY, preview)
      wx.hideLoading()
      navigateShellRoute(`/${ROUTES.gameCreatePreview}`, {
        currentRoute: ROUTES.gameCreate,
        reuseExisting: false
      })
    } catch (error) {
      wx.hideLoading()
      toast.info(createFlowErrorMessage(error, '预览生成失败，请稍后重试'))
    } finally {
      this.setData({ previewPreparing: false })
    }
  },

  openDraftBox() {
    if (this.isCreateOperationBusy() || this.data.descriptionMediaPendingCount > 0) {
      toast.info('当前操作尚未完成，请稍候')
      return
    }
    navigateShellRoute(`/${ROUTES.gameCreateDrafts}`, {
      currentRoute: ROUTES.gameCreate,
      reuseExisting: false
    })
  },

  async publishGame() {
    if (this.isCreateOperationBusy()) {
      toast.info(this.data.publishing ? '正在发布，请勿重复提交' : '当前操作尚未完成，请稍候')
      return
    }
    if (this.data.descriptionMediaPendingCount > 0) {
      toast.info('图片或视频仍在处理中，请稍候再发布')
      return
    }
    const missing = this.getPublishMissingFields()
    if (missing.length) {
      toast.info(this.getPublishBlockedMessage(missing))
      this.focusPublishIssue(missing[0])
      this.syncPublishState()
      return
    }

    const payload = this.buildCreateGamePayload()
    if (!payload) {
      toast.info('表单状态异常，请重新确认局类型、人数和局时间')
      return
    }

    this.setData({ publishing: true })
    wx.showLoading({ title: '发布中', mask: true })
    let usedDefaultCover = false
    try {
      if (!featureFlags.gameCoverUploadEnabled) {
        payload.coverFileId = 0
        payload.coverImage = DEFAULT_GAME_COVER
        usedDefaultCover = true
      } else {
        const persistedCover = await this.persistCoverForDraft()
        payload.coverFileId = Number(persistedCover.fileId || 0)
        payload.coverImage = payload.coverFileId > 0 ? '' : String(persistedCover.url || '').trim()
      }
      payload.descriptionMedia = await this.uploadDescriptionMedia()
      const game = await gameService.createGame(payload, this.createSubmissionSessionId)
      this.createSubmissionSessionId = createPreviewSessionId()
      this.draftSubmissionSessionId = createPreviewSessionId()
      const draftId = Number(this.data.draftId || 0)
      if (draftId > 0) {
        // 发布已成功，草稿删除失败不能把成功结果改成失败；下次进入
        // 草稿箱仍可由用户手动删除。
        try {
          await gameService.deleteGameDraft(draftId)
          this.setData({ draftId: '', draftCount: Math.max(0, Number(this.data.draftCount || 0) - 1) })
        } catch (cleanupError) {
          // 忽略清理错误，避免用户因误判失败而重复发布同一局。
        }
      }
      wx.hideLoading()
      try {
        wx.removeStorageSync(CREATE_PREVIEW_STORAGE_KEY)
        wx.removeStorageSync(CREATE_PREVIEW_ACTION_KEY)
      } catch (storageError) {
        // 发布结果已经由服务端确认，本地预览缓存清理失败不影响成功状态。
      }
      toast.success(usedDefaultCover ? '发布成功，暂用默认封面' : '已提交后台审核')
      setTimeout(() => {
        navigateShellRoute(`${ROUTES.gameDetail}?id=${game.id}`, {
          currentRoute: ROUTES.gameCreate
        })
      }, 500)
    } catch (error) {
      wx.hideLoading()
      const errorMessage = createFlowErrorMessage(error, '请检查填写内容后重试')
      this.focusPublishIssue(errorMessage)
      toast.info(/invalid file request/i.test(errorMessage)
        ? '封面上传失败：仅支持 JPG/JPEG、PNG、WEBP，且不能超过 10MB'
        : (`发布失败：${this.getFriendlyPublishError(errorMessage)}`))
    } finally {
      this.setData({ publishing: false })
    }
  },

  async persistCoverForDraft() {
    if (!featureFlags.gameCoverUploadEnabled) {
      return { fileId: 0, url: DEFAULT_GAME_COVER }
    }

    let fileId = Number(this.data.coverFileId || 0)
    const coverImage = String(this.data.coverImage || '').trim()
    if (!fileId) {
      const isWechatTempCover = /^https?:\/\/tmp\//i.test(coverImage)
      const isLocalCover = coverImage && (isWechatTempCover || (!/^https?:\/\//i.test(coverImage) && !/^mock:\/\//i.test(coverImage)))
      if (!isLocalCover) {
        return { fileId: 0, url: coverImage }
      }
      const coverFile = this.data.coverFile && this.data.coverFile.path === coverImage
        ? this.data.coverFile
        : coverImage
      try {
        fileId = await fileService.uploadSingleFile(coverFile, {
          bizType: 'game_cover',
          objectId: 0
        })
      } catch (error) {
        throw new Error(`封面上传失败：${createFlowErrorMessage(error, '请稍后重试')}`)
      }
      if (!fileId) {
        throw new Error('封面上传失败')
      }
    }

    // 下载地址可能带有效期。每次保存或打开预览时刷新，确保服务器
    // 草稿跨设备、跨会话恢复后仍能看到封面。
    let url
    try {
      url = await fileService.getDownloadURL(fileId)
    } catch (error) {
      throw new Error(`封面读取失败：${error.message || '请稍后重试'}`)
    }
    await new Promise((resolve) => this.setData({
      coverFileId: fileId,
      coverImage: url,
      coverFile: null
    }, resolve))
    return { fileId, url }
  },

  async uploadDescriptionMedia() {
    const persisted = await this.persistDescriptionMedia()
    return persisted.map((item) => ({
      fileId: Number(item.fileId || 0),
      type: item.type === 'video' ? 'video' : 'image',
      duration: item.type === 'video' ? Math.ceil(Number(item.duration || 0)) : 0
    }))
  },

  async persistDescriptionMedia() {
    const items = Array.isArray(this.data.descriptionMedia) ? this.data.descriptionMedia : []
    const persisted = []
    for (let index = 0; index < items.length; index += 1) {
      const item = items[index] || {}
      let fileId = Number(item.fileId || 0)
      if (!fileId) {
        const path = String(item.tempFilePath || item.path || '').trim()
        if (!path) {
          throw new Error('局详情媒体文件已失效，请重新选择')
        }
        try {
          fileId = await fileService.uploadSingleFile({
            path,
            fileName: `game-description-${Date.now()}-${index}.${item.type === 'video' ? (getFileExtension(path) || 'mp4') : (getFileExtension(path) || 'jpg')}`,
            mimeType: descriptionMediaMimeType(item),
            size: Number(item.size || 1) || 1
          }, {
            bizType: 'game_description',
            objectId: 0
          })
        } catch (error) {
          throw new Error(`局详情媒体上传失败：${createFlowErrorMessage(error, '请稍后重试')}`)
        }
      }
      if (!fileId) {
        throw new Error('局详情媒体上传失败')
      }
      // 草稿中的下载地址可能已过期，始终按 fileId 获取当前地址。
      let url
      try {
        url = await fileService.getDownloadURL(fileId)
      } catch (error) {
        throw new Error(`局详情媒体读取失败：${error.message || '请稍后重试'}`)
      }
      persisted.push({ ...item, fileId, url })
    }
    return persisted
  },

  focusPublishIssue(issue = '') {
    const text = String(issue || '')
    const fields = [
      [/主题|标题/, 'theme'],
      [/封面/, 'cover'],
      [/介绍/, 'introduction'],
      [/亮点/, 'highlights'],
      [/描述|详情媒体|图片|视频/, 'description'],
      [/类型|收费/, 'type'],
      [/人数/, 'capacity'],
      [/角色/, 'roles'],
      [/报名时间/, 'signupTime'],
      [/时间|日期/, 'gameTime'],
      [/完成规则|完成标准/, 'completionRules'],
      [/地址|地图|位置/, 'location']
    ]
    const matched = fields.find(([pattern]) => pattern.test(text))
    if (!matched) {
      return
    }
    const target = `create-field-${matched[1]}`
    this.setData({ createScrollIntoView: '', createScrollTop: 0 })
    setTimeout(() => this.setData({ createScrollIntoView: target }), 30)
  },

  getPublishBlockedMessage(missing = this.getPublishMissingFields()) {
    if (!missing.length) {
      return '信息已填写完整，请重新点击发布'
    }

    return `请先完善：${missing.slice(0, 3).join('、')}`
  },

  getFriendlyPublishError(message = '') {
    const raw = String(message || '').trim()
    const messageMap = [
      [/invalid primary category/i, '局类型无效，请重新选择'],
      [/invalid game cover/i, '封面文件无效，请重新上传'],
      [/invalid game input/i, '组局信息不完整或格式不正确，请检查后重试'],
      [/storage base url not configured/i, '文件存储服务暂不可用，请稍后重试'],
      [/daily limit reached/i, '今日发起组局次数已达上限'],
      [/realname required|强实名未完成/i, '请先完成实名认证后再发起组局'],
      [/signup.*time|报名时间/i, '请确认报名时间完整，且报名结束早于局开始时间'],
      [/description media|详情图片或视频/i, '局描述中的图片或视频无效，请重新上传']
    ]
    const matched = messageMap.find(([pattern]) => pattern.test(raw))
    return matched ? matched[1] : (raw || '请检查填写内容后重试')
  },

  getPublishMissingFields() {
    const form = this.data.form || {}
    const capacity = this.data.createForm.capacity || {}
    const missing = []
    const selectedCategory = findSelectedCategory(this.data.gameTypes || EMPTY_GAME_TYPES, String(form.type || '').trim())
    const gameTimeValid = this.data.gameTimeConfirmed && isValidTimeDraft(this.data.timeDraft)
    const signupTimeValid = this.data.signupTimeConfirmed && isValidSignupTimeDraft(this.data.signupTimeDraft, this.data.timeDraft)
    const allowedRoles = Array.isArray(form.allowedRoles) ? form.allowedRoles : []
    const descriptionMedia = Array.isArray(this.data.descriptionMedia) ? this.data.descriptionMedia : []

    if (!Array.isArray(this.data.gameTypes) || !this.data.gameTypes.length) {
      missing.push('局类型配置')
    }
    if (!String(form.theme || '').trim()) {
      missing.push('局主题')
    }
    if (!selectedCategory.primary) {
      missing.push('局类型')
    }
    if (featureFlags.gameCoverUploadEnabled && !String(this.data.coverImage || '').trim()) {
      missing.push('局封面')
    }
    if (!String(form.intro || '').trim()) {
      missing.push('局介绍')
    }
    if (!String(form.highlights || '').trim()) {
      missing.push('亮点')
    }
    if (!String(form.description || '').trim() || String(form.description || '').length > DESCRIPTION_MAX_LENGTH) {
      missing.push('局描述')
    }
    if (Number(form.capacity) < Number(capacity.min || 0) || Number(form.capacity) > Number(capacity.max || 0)) {
      missing.push('人数')
    }
    if (!allowedRoles.length || allowedRoles.some((role) => !PARTICIPANT_ROLE_OPTIONS.some((option) => option.key === role))) {
      missing.push('可参与角色')
    }
    if (!gameTimeValid) {
      missing.push('局时间')
    }
    if (!signupTimeValid) {
      missing.push('报名时间')
    }
    if (!hasConfirmedGameLocation(this.data.locationInfo)) {
      missing.push('组局地址')
    }
    if (!Array.isArray(this.data.completionRules) || !this.data.completionRules.some((item) => item.active)) {
      missing.push('完成规则')
    }
    if (this.getMediaImageCount(descriptionMedia) > DESCRIPTION_IMAGE_MAX_COUNT
      || this.getMediaVideoCount(descriptionMedia) > DESCRIPTION_VIDEO_MAX_COUNT) {
      missing.push('局详情媒体')
    }
    return missing
  },

  buildCreateGamePayload() {
    const form = this.data.form || {}
    const locationInfo = this.data.locationInfo || {}
    const title = String(form.theme || '').trim()
    const capacity = this.data.createForm.capacity || {}
    const minPlayers = Number(capacity.min || 0)
    const maxCapacity = Number(capacity.max || 0)
    const maxPlayers = clampCapacity(form.capacity, capacity)

    if (!title || !String(form.type || '').trim() || maxPlayers < minPlayers || maxPlayers > maxCapacity
      || !this.data.gameTimeConfirmed || !isValidTimeDraft(this.data.timeDraft)
      || !this.data.signupTimeConfirmed || !isValidSignupTimeDraft(this.data.signupTimeDraft, this.data.timeDraft)
      || !hasConfirmedGameLocation(locationInfo)) {
      return null
    }

    const gameTypes = this.data.gameTypes || EMPTY_GAME_TYPES
    if (!gameTypes.length) {
      toast.info('组局类型配置加载失败')
      return null
    }

    const selected = findSelectedCategory(gameTypes, form.type)
    const primary = selected.primary || {}
    const allowedRoles = Array.isArray(form.allowedRoles)
      ? form.allowedRoles.filter((role, index, items) => (
        PARTICIPANT_ROLE_OPTIONS.some((option) => option.key === role) && items.indexOf(role) === index
      ))
      : []
    if (!primary.key || !allowedRoles.length || !Array.isArray(this.data.completionRules)
      || !this.data.completionRules.some((item) => item.active)) {
      return null
    }

    return {
      title,
      gameType: 'free',
      coverImage: this.data.coverImage || '',
      introduction: String(form.intro || '').trim(),
      // 「局介绍」和「局描述」是两个独立字段。此前这里误取局介绍，
      // 会让用户填写的局描述在正式发布后丢失。
      description: String(form.description || '').trim(),
      highlights: String(form.highlights || '').trim(),
      notice: String(form.notice || '').trim(),
      audience: String(form.audience || '').trim(),
      participation: form.participation || '',
      allowedRoles,
      startAt: this.data.gameTimeConfirmed ? `${this.data.timeDraft.startDate} ${this.data.timeDraft.startTime}` : '',
      endAt: this.data.gameTimeConfirmed ? `${this.data.timeDraft.endDate} ${this.data.timeDraft.endTime}` : '',
      signupStartAt: this.data.signupTimeConfirmed ? `${this.data.signupTimeDraft.startDate} ${this.data.signupTimeDraft.startTime}` : '',
      signupEndAt: this.data.signupTimeConfirmed ? `${this.data.signupTimeDraft.endDate} ${this.data.signupTimeDraft.endTime}` : '',
      tags: activeOptionKeys(this.data.tags),
      completionRules: activeOptionKeys(this.data.completionRules),
      primaryCategory: primary.key || form.type || '',
      primaryCategoryText: primary.name || '',
      secondaryCategory: '',
      secondaryCategoryText: '',
      type: 'free',
      price: 0,
      profitTemplate: '',
      minPlayers,
      maxPlayers,
      cityCode: locationInfo.cityCode || '',
      cityName: locationInfo.cityName || locationInfo.name || '',
      address: locationInfo.address || locationInfo.name || '',
      longitude: Number(locationInfo.longitude),
      latitude: Number(locationInfo.latitude)
    }
  },

  syncPublishState() {
    const form = this.data.form || {}
    const hasGameTypes = Array.isArray(this.data.gameTypes) && this.data.gameTypes.length > 0
    const hasTheme = String(form.theme || '').trim().length > 0
    const hasType = Boolean(findSelectedCategory(this.data.gameTypes || EMPTY_GAME_TYPES, String(form.type || '').trim()).primary)
    const hasCover = !featureFlags.gameCoverUploadEnabled || String(this.data.coverImage || '').trim().length > 0
    const hasIntroduction = String(form.intro || '').trim().length > 0
    const hasHighlights = String(form.highlights || '').trim().length > 0
    const hasDescription = String(form.description || '').trim().length > 0 && String(form.description || '').length <= DESCRIPTION_MAX_LENGTH
    const capacity = this.data.createForm.capacity || {}
    const hasCapacity = Number(form.capacity) >= Number(capacity.min || 0) && Number(form.capacity) <= Number(capacity.max || 0)
    const hasAllowedRoles = Array.isArray(form.allowedRoles) && form.allowedRoles.length > 0
      && form.allowedRoles.every((role) => PARTICIPANT_ROLE_OPTIONS.some((option) => option.key === role))
    const hasGameTime = Boolean(this.data.gameTimeConfirmed) && isValidTimeDraft(this.data.timeDraft)
    const hasSignupTime = Boolean(this.data.signupTimeConfirmed) && isValidSignupTimeDraft(this.data.signupTimeDraft, this.data.timeDraft)
    const hasLocation = hasConfirmedGameLocation(this.data.locationInfo)
    const hasCompletionRule = Array.isArray(this.data.completionRules) && this.data.completionRules.some((item) => item.active)

    this.setData({
      publishDisabled: !(hasGameTypes && hasTheme && hasType && hasCover && hasIntroduction && hasHighlights && hasDescription && hasCapacity && hasAllowedRoles && hasGameTime && hasSignupTime && hasLocation && hasCompletionRule)
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
