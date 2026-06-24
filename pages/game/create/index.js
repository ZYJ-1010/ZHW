const toast = require('../../../utils/toast')
const gameService = require('../../../services/game')
const { ROUTES } = require('../../../config/routes')

const CREATE_SCROLL_TAP_STEP_RPX = 360
const CREATE_SCROLL_HOLD_STEP_RPX = 72
const CREATE_SCROLL_HOLD_INTERVAL_MS = 80
const CREATE_SCROLL_HOLD_SUPPRESS_TAP_MS = 120
const COVER_ALLOWED_FORMATS = ['jpg', 'jpeg', 'png']
const DESCRIPTION_ALLOWED_IMAGE_FORMATS = ['jpg', 'jpeg', 'png']
const DESCRIPTION_ALLOWED_VIDEO_FORMATS = ['mp4', 'mov']
const DESCRIPTION_IMAGE_MAX_COUNT = 9
const DESCRIPTION_VIDEO_MAX_DURATION = 60
const THEME_MAX_LENGTH = 20
const INTRO_MAX_LENGTH = 200
const HIGHLIGHTS_MAX_LENGTH = 100
const NOTICE_MAX_LENGTH = 100
const AUDIENCE_MAX_LENGTH = 50
const CURRENT_LOCATION_TEXT = '当前位置'
const CAPACITY_MIN = 3
const CAPACITY_MAX = 10
const DEFAULT_DEPOSIT_RULE_TEXT = '连续打卡 7 天即完成。完成者拿回押金池金额，未完成者押金由完成者平分。'
const DEFAULT_DEPOSIT_NOTICE_TEXT = '支付金额：100元 = 服务费10元 + 押金池90元。服务费不退，押金池按完成情况结算。'
const DEFAULT_PROFIT_TEMPLATES = [
  { key: 'standard', name: '标准 1441', desc: '平台10% · 流量方40% · 交付方40% · 推荐上级10%' },
  { key: 'aa', name: 'AA局', desc: '平台2.5% · 交付方90% · 流量方5% · 推荐上级2.5%' },
  { key: 'deposit', name: '押金局', desc: '平台2.5% · 交付方0% · 流量方5% · 推荐上级2.5% + 押金池90%，完成返还，未完成瓜分' },
  { key: 'crowdfunding', name: '众筹局', desc: '平台2.5% · 交付方90% · 流量方5% · 推荐上级2.5%' },
  { key: 'publicBenefit', name: '公益局', desc: '平台0% · 交付方100% · 流量方0% · 推荐上级0%', disabled: true, disabledReason: '仅平台账户可发起' }
]

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
  const end = addHours(now, 2)

  return {
    startDate: formatDate(now),
    startTime: formatTime(now),
    endDate: formatDate(end),
    endTime: formatTime(end)
  }
}

function getInitialSignupTimeDraft() {
  const now = new Date()
  const end = addHours(now, 24)

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

function clampCapacity(value) {
  const numberValue = Number(value)

  if (Number.isNaN(numberValue)) {
    return CAPACITY_MIN
  }

  return Math.min(CAPACITY_MAX, Math.max(CAPACITY_MIN, Math.round(numberValue)))
}

function getFileExtension(filePath = '') {
  const normalized = String(filePath || '').split('?')[0].split('#')[0]
  const matched = normalized.match(/\.([a-zA-Z0-9]+)$/)

  return matched ? matched[1].toLowerCase() : ''
}

function isAllowedCoverFormat(file = {}, imageInfo = {}) {
  return isAllowedMediaFormat(file, imageInfo, COVER_ALLOWED_FORMATS)
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

  return templates.length ? templates : DEFAULT_PROFIT_TEMPLATES
}

function getSelectableProfitTemplateKey(templates = [], currentKey = '') {
  const current = templates.find((item) => item.key === currentKey)

  if (current && !current.disabled) {
    return current.key
  }

  const fallback = templates.find((item) => !item.disabled)

  return fallback ? fallback.key : ''
}

Page({
  data: {
    onlineText: '3999人在线',
    createScrollTop: 0,
    coverImage: '',
    publishDisabled: true,
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
    capacityMin: CAPACITY_MIN,
    capacityMax: CAPACITY_MAX,
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
      theme: '我想组一局，找到产品经理一起梳理真好玩 MVP',
      type: '',
      capacity: 5,
      participation: '',
      intro: '',
      highlights: '',
      description: '',
      notice: '',
      audience: '',
      feeType: 'paid',
      price: '',
      profitTemplate: 'deposit'
    },
    gameTypes: [
      { key: 'task', name: '任务局' },
      { key: 'social', name: '社交局' },
      { key: 'explore', name: '探索局' },
      { key: 'growth', name: '成长局' }
    ],
    scheduleFields: [
      { key: 'gameTime', label: '局时间', required: false, value: '', hint: '展开日期面板，选择开始和结束日期+时间' },
      { key: 'signupTime', label: '报名时间', required: true, value: '', hint: '展开日期面板，选择报名开始和结束日期+时间' },
      { key: 'location', label: '组局地址', required: false, value: '', hint: '点击在地图上标记位置' }
    ],
    participationModes: [
      { key: 'online', name: '线上' },
      { key: 'offline', name: '线下' },
      { key: 'hybrid', name: '混合' }
    ],
    tags: [
      { name: '产品研发', active: false },
      { name: '创业', active: false },
      { name: '城市探索', active: false },
      { name: '共创', active: false }
    ],
    completionRules: [
      { key: 'time', name: '时间截止', active: false },
      { key: 'goal', name: '目标达成', active: true },
      { key: 'capacity', name: '人数满额', active: false },
      { key: 'manual', name: '手动结束', active: true }
    ],
    feeTypes: [
      { key: 'free', name: '免费局' },
      { key: 'paid', name: '收费局' }
    ],
    profitTemplates: DEFAULT_PROFIT_TEMPLATES,
    depositRuleText: DEFAULT_DEPOSIT_RULE_TEXT,
    depositNoticeText: DEFAULT_DEPOSIT_NOTICE_TEXT
  },

  onLoad() {
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
    this.initGameLocation()
    this.loadProfitTemplates()
  },

  loadProfitTemplates(feeType) {
    gameService.getProfitTemplates({
      feeType: feeType || (this.data.form && this.data.form.feeType)
    }).then((data) => {
      const profitTemplates = normalizeProfitTemplates(data)
      const profitTemplate = getSelectableProfitTemplateKey(
        profitTemplates,
        this.data.form && this.data.form.profitTemplate
      )
      const depositRuleText = getProfitConfigText(
        data,
        ['depositRuleText', 'ruleText', 'rule', 'depositRule'],
        DEFAULT_DEPOSIT_RULE_TEXT
      )
      const depositNoticeText = getProfitConfigText(
        data,
        ['depositNoticeText', 'noticeText', 'tipText', 'notice', 'depositNotice'],
        DEFAULT_DEPOSIT_NOTICE_TEXT
      )

      this.setData({
        profitTemplates,
        'form.profitTemplate': profitTemplate,
        depositRuleText,
        depositNoticeText
      })
      this.syncPublishState()
    }).catch(() => {
      const profitTemplates = DEFAULT_PROFIT_TEMPLATES
      const profitTemplate = getSelectableProfitTemplateKey(
        profitTemplates,
        this.data.form && this.data.form.profitTemplate
      )

      this.setData({
        profitTemplates,
        'form.profitTemplate': profitTemplate,
        depositRuleText: DEFAULT_DEPOSIT_RULE_TEXT,
        depositNoticeText: DEFAULT_DEPOSIT_NOTICE_TEXT
      })
      this.syncPublishState()
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
              coverImage: ''
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
    })
    this.syncPublishState()
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
          toast.info('仅支持 JPG、PNG 等图片格式')
          return
        }

        this.setData({
          coverImage: file.tempFilePath
        })
      },
      fail: () => {
        toast.info('图片校验失败')
      }
    })
  },

  selectType(event) {
    this.setData({
      'form.type': event.currentTarget.dataset.key
    })
    this.syncPublishState()
  },

  chooseSchedule(event) {
    const key = event.currentTarget.dataset.key

    if (key === 'gameTime') {
      this.setData({
        timePanelVisible: !this.data.timePanelVisible,
        signupTimePanelVisible: false
      })
      return
    }

    if (key === 'signupTime') {
      this.setData({
        signupTimePanelVisible: !this.data.signupTimePanelVisible,
        timePanelVisible: false
      })
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

    toast.info('选择器待接入')
  },

  chooseGameLocation() {
    const locationOptions = {}
    const locationInfo = this.data.locationInfo || {}

    if (typeof locationInfo.latitude === 'number' && typeof locationInfo.longitude === 'number') {
      locationOptions.latitude = locationInfo.latitude
      locationOptions.longitude = locationInfo.longitude
    }

    wx.chooseLocation({
      ...locationOptions,
      success: (res = {}) => {
        const name = res.name || ''
        const address = res.address || ''
        const text = name || address || '已选择位置'

        this.updateScheduleField('location', text)
        this.setData({
          locationInfo: {
            name,
            address,
            latitude: typeof res.latitude === 'number' ? res.latitude : '',
            longitude: typeof res.longitude === 'number' ? res.longitude : ''
          }
        })
      },
      fail: (error = {}) => {
        if (error.errMsg && error.errMsg.indexOf('cancel') > -1) {
          return
        }

        toast.info('地图选点失败')
      }
    })
  },

  initGameLocation() {
    if (!wx.getLocation) {
      return
    }

    wx.getLocation({
      type: 'gcj02',
      success: (res = {}) => {
        if (typeof res.latitude !== 'number' || typeof res.longitude !== 'number') {
          return
        }

        this.updateScheduleField('location', CURRENT_LOCATION_TEXT)
        this.setData({
          locationInfo: {
            name: CURRENT_LOCATION_TEXT,
            address: '',
            latitude: res.latitude,
            longitude: res.longitude
          }
        })
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
    const nowTimestamp = getCurrentMinuteTimestamp()

    if (!startTimestamp || !endTimestamp || Number.isNaN(startTimestamp) || Number.isNaN(endTimestamp)) {
      toast.info('请选择完整局时间')
      return
    }

    if (startTimestamp < nowTimestamp) {
      toast.info('局开始时间不能早于当前时间')
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
    })
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
    const nowTimestamp = getCurrentMinuteTimestamp()

    if (!signupStartTimestamp || !signupEndTimestamp || Number.isNaN(signupStartTimestamp) || Number.isNaN(signupEndTimestamp)) {
      toast.info('请选择完整报名时间')
      return
    }

    if (signupStartTimestamp < nowTimestamp) {
      toast.info('报名开始时间不能早于当前时间')
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
      'form.capacity': clampCapacity(event.detail.value)
    })
    this.syncPublishState()
  },

  onCapacityInput(event) {
    this.setData({
      'form.capacity': clampCapacity(event.detail.value)
    })
    this.syncPublishState()
  },

  selectParticipation(event) {
    this.setData({
      'form.participation': event.currentTarget.dataset.key
    })
  },

  toggleTag(event) {
    const index = Number(event.currentTarget.dataset.index)
    const tags = this.data.tags.map((item, itemIndex) => ({
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

    this.setData({ completionRules })
    this.syncPublishState()
  },

  selectFeeType(event) {
    const feeType = event.currentTarget.dataset.key
    const nextData = {
      'form.feeType': feeType
    }

    if (feeType === 'free') {
      nextData['form.price'] = ''
    }

    this.setData({
      ...nextData
    })
    this.syncPublishState()
    this.loadProfitTemplates(feeType)
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
    })
    this.syncPublishState()
  },

  saveDraft() {
    toast.info('草稿保存待接入')
  },

  previewSubmit() {
    toast.info('预览提交待接入')
  },

  publishGame() {
    if (this.data.publishDisabled) {
      toast.info('请先完善局属性')
      return
    }

    toast.info('发布接口待接入')
  },

  syncPublishState() {
    const form = this.data.form || {}
    const hasTheme = String(form.theme || '').trim().length > 0
    const hasType = String(form.type || '').trim().length > 0
    const hasCapacity = Number(form.capacity) >= CAPACITY_MIN && Number(form.capacity) <= CAPACITY_MAX
    const hasCompletionRule = Array.isArray(this.data.completionRules) && this.data.completionRules.some((item) => item.active)
    const hasFeeType = Array.isArray(this.data.feeTypes) && this.data.feeTypes.some((item) => item.key === form.feeType)
    const hasProfitTemplate = Array.isArray(this.data.profitTemplates)
      && this.data.profitTemplates.some((item) => item.key === form.profitTemplate && !item.disabled)

    this.setData({
      publishDisabled: !(hasTheme && hasType && hasCapacity && hasCompletionRule && hasFeeType && hasProfitTemplate)
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

    if (key === 'left' || key === 'right') {
      toast.info('功能正在开发中')
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
      this.scrollCreateToTop()
      return
    }

    if (key === 'search') {
      toast.info('搜索功能开发中')
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

    wx.navigateTo({
      url: `/${route}`
    })
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
