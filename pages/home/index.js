const homeService = require('../../services/home')
const roleService = require('../../services/role')
const toast = require('../../utils/toast')
const { ROUTES } = require('../../config/routes')
const UI_ICONS = require('../../config/ui-icons')

const APPLY_STAGE_TOP_RPX = 108
const APPLY_DEFAULT_NAV_TOP_RPX = 108
const APPLY_DEFAULT_NAV_HEIGHT_RPX = 64
const APPLY_DEFAULT_CONTENT_TOP_RPX = 181
const APPLY_NAV_BOTTOM_GAP_RPX = 13
const APPLY_FRAME_BOTTOM_PADDING_RPX = 10
const EXPERT_APPLY_SUCCESS_REDIRECT_DELAY_MS = 1200
const EXPERT_APPLY_YEAR_OPTIONS = Array.from({ length: 30 }, (_, index) => `${index + 1}年`).concat('30年以上')
const EXPERT_SERVICE_COUNT = 3
const DEFAULT_EXPERT_APPLY_FORM = {
  skillTags: '人像摄影、城市旅拍、活动记录',
  experienceYearIndex: 4,
  experienceYears: '5年',
  intro: '我长期从事城市旅拍、人像摄影和活动记录服务，熟悉前期沟通、现场引导、成片交付和用户体验管理。可以根据玩家的组局主题设计拍摄动线，提供稳定的审美表达、清晰的交付标准和可复用的服务流程。',
  uploadFiles: [
    {
      name: '城市旅拍作品集.pdf',
      path: 'mock://expert-portfolio.pdf',
      size: 102400,
      type: 'file'
    }
  ],
  services: [
    { name: '城市旅拍', price: '399', cost: '120' },
    { name: '活动跟拍', price: '599', cost: '180' },
    { name: '修图交付', price: '199', cost: '60' }
  ]
}
const DEFAULT_EXPERT_APPLY_CONFIG = {
  skillOptions: [
    { name: '摄影', active: true },
    { name: '户外', active: false },
    { name: '美食', active: false },
    { name: '文化', active: false },
    { name: '手工', active: false },
    { name: '运动', active: false },
    { name: '音乐', active: false },
    { name: '+ 自定义', active: false, custom: true }
  ],
  fields: [
    {
      type: 'chips',
      key: 'skillDomain',
      label: '选择技能领域',
      required: true
    },
    {
      type: 'input',
      key: 'skillTags',
      label: '技能标签',
      required: true,
      placeholder: '如：人像摄影、风光摄影、夜景拍摄',
      helper: '添加具体标签，让用户更容易找到你',
      maxlength: 30
    },
    {
      type: 'select',
      key: 'experienceYears',
      label: '从业年限',
      required: true,
      placeholder: '请选择从业年限'
    },
    {
      type: 'textarea',
      key: 'intro',
      label: '个人简介',
      required: true,
      placeholder: '介绍你的专业背景、服务风格、擅长领域...',
      helper: '不少于 50 字，突出你的专业优势',
      maxlength: 300
    }
  ],
  uploadField: {
    label: '资质证明',
    required: true,
    icon: UI_ICONS.panel.upload,
    title: '点击上传作品集及凭证',
    helper: '支持 JPG、PNG、PDF，最多 5 张',
    acceptTypes: ['JPG', 'PNG', 'PDF'],
    maxCount: 5
  },
  validationRules: {
    skillTags: {
      minLength: 2,
      maxLength: 30
    },
    intro: {
      minLength: 50,
      maxLength: 300
    },
    serviceName: {
      minLength: 2,
      maxLength: 20
    },
    customSkill: {
      minLength: 2,
      maxLength: 8
    },
    money: {
      integerMaxLength: 8,
      decimalMaxLength: 2
    }
  },
  yearOptions: EXPERT_APPLY_YEAR_OPTIONS,
  serviceCount: EXPERT_SERVICE_COUNT,
  priceHint: '平台将收取 10% 服务费'
}

function getPositiveInteger(value, fallback) {
  const number = Number(value)
  return Number.isInteger(number) && number > 0 ? number : fallback
}

function createExpertApplyServices(count = EXPERT_SERVICE_COUNT) {
  const serviceCount = getPositiveInteger(count, EXPERT_SERVICE_COUNT)

  return Array.from({ length: serviceCount }, (_, index) => {
    const demoService = DEFAULT_EXPERT_APPLY_FORM.services[index] || {}

    return {
      name: demoService.name || '',
      price: demoService.price || '',
      cost: demoService.cost || ''
    }
  })
}

function createExpertApplyServiceErrors(count = EXPERT_SERVICE_COUNT) {
  const serviceCount = getPositiveInteger(count, EXPERT_SERVICE_COUNT)

  return Array.from({ length: serviceCount }, () => ({
    name: '',
    price: '',
    cost: ''
  }))
}

function createExpertApplyForm(count = EXPERT_SERVICE_COUNT) {
  return {
    skillTags: DEFAULT_EXPERT_APPLY_FORM.skillTags,
    experienceYearIndex: DEFAULT_EXPERT_APPLY_FORM.experienceYearIndex,
    experienceYears: DEFAULT_EXPERT_APPLY_FORM.experienceYears,
    intro: DEFAULT_EXPERT_APPLY_FORM.intro,
    uploadFiles: DEFAULT_EXPERT_APPLY_FORM.uploadFiles.map((file) => ({ ...file })),
    services: createExpertApplyServices(count)
  }
}

function createExpertApplyErrors() {
  return {
    skillDomain: '',
    skillTags: '',
    experienceYears: '',
    intro: '',
    uploadFiles: ''
  }
}

function trimText(value) {
  return String(value || '').trim()
}

function hasAnyError(errorMap) {
  return Object.keys(errorMap).some((key) => Boolean(errorMap[key]))
}

function getMoneyRule(rules) {
  const moneyRule = (rules && rules.money) || DEFAULT_EXPERT_APPLY_CONFIG.validationRules.money

  return {
    integerMaxLength: getPositiveInteger(moneyRule.integerMaxLength, 8),
    decimalMaxLength: getPositiveInteger(moneyRule.decimalMaxLength, 2)
  }
}

function isValidMoney(value, rules) {
  const text = trimText(value)
  const moneyRule = getMoneyRule(rules)
  const pattern = new RegExp(`^(0|[1-9]\\d{0,${moneyRule.integerMaxLength - 1}})(\\.\\d{1,${moneyRule.decimalMaxLength}})?$`)

  return pattern.test(text) && Number(text) > 0
}

function normalizeUploadFile(file, index) {
  const path = file.path || file.tempFilePath || ''
  const fallbackName = path ? path.split(/[\\/]/).pop() : `文件${index + 1}`

  return {
    name: file.name || fallbackName,
    path,
    size: file.size || 0,
    type: file.type || ''
  }
}

function getUploadFileExtension(file) {
  const name = trimText(file.name)
  const path = trimText(file.path)
  const source = name || path
  const dotIndex = source.lastIndexOf('.')

  return dotIndex >= 0 ? source.slice(dotIndex + 1).toUpperCase() : ''
}

function isAllowedUploadFile(file, acceptTypes) {
  const accepted = (acceptTypes || []).map((type) => String(type).replace(/^\./, '').toUpperCase())

  if (!accepted.length) {
    return true
  }

  return accepted.includes(getUploadFileExtension(file))
}

function cloneObject(value) {
  return JSON.parse(JSON.stringify(value))
}

function normalizeSkillOptions(options) {
  const sourceOptions = Array.isArray(options) && options.length
    ? options
    : DEFAULT_EXPERT_APPLY_CONFIG.skillOptions
  const normalized = sourceOptions
    .map((item) => {
      const name = trimText(item.name || item.label || item.text || item.value)

      if (!name) {
        return null
      }

      return {
        name,
        active: Boolean(item.active),
        custom: Boolean(item.custom || item.type === 'custom' || name === '+ 自定义'),
        generatedCustom: Boolean(item.generatedCustom)
      }
    })
    .filter(Boolean)
  const hasActive = normalized.some((item) => item.active && !item.custom)
  const nextOptions = normalized.map((item, index) => ({
    ...item,
    active: hasActive ? item.active : index === 0 && !item.custom
  }))

  if (!nextOptions.some((item) => item.custom)) {
    nextOptions.push({ name: '+ 自定义', active: false, custom: true })
  }

  return nextOptions
}

function normalizeUploadField(uploadField) {
  const rawUpload = uploadField || {}
  const fallbackUpload = DEFAULT_EXPERT_APPLY_CONFIG.uploadField
  const acceptTypes = Array.isArray(rawUpload.acceptTypes) && rawUpload.acceptTypes.length
    ? rawUpload.acceptTypes
    : Array.isArray(rawUpload.types) && rawUpload.types.length
      ? rawUpload.types
      : fallbackUpload.acceptTypes
  const maxCount = getPositiveInteger(rawUpload.maxCount || rawUpload.maxFiles || rawUpload.count, fallbackUpload.maxCount)
  const helper = rawUpload.helper || `支持 ${acceptTypes.join('、')}，最多 ${maxCount} 张`

  return Object.assign({}, fallbackUpload, rawUpload, {
    acceptTypes,
    maxCount,
    helper
  })
}

function createExpertServiceBlocks(count = EXPERT_SERVICE_COUNT) {
  const serviceCount = getPositiveInteger(count, EXPERT_SERVICE_COUNT)

  return Array.from({ length: serviceCount }, (_, index) => ({
    id: `service-${index + 1}`,
    title: '业务：',
    fields: [
      { key: 'price', label: '服务定价（元/小时）', required: true },
      { key: 'cost', label: '服务成本', required: false }
    ]
  }))
}

function normalizeValidationRules(rules) {
  return Object.assign(
    {},
    cloneObject(DEFAULT_EXPERT_APPLY_CONFIG.validationRules),
    rules || {}
  )
}

function normalizeExpertApplyConfig(config) {
  const rawConfig = config && typeof config === 'object' ? config : {}
  const serviceCount = getPositiveInteger(rawConfig.serviceCount, DEFAULT_EXPERT_APPLY_CONFIG.serviceCount)

  return {
    skillOptions: normalizeSkillOptions(rawConfig.skillOptions || rawConfig.skillDomains),
    fields: Array.isArray(rawConfig.fields) && rawConfig.fields.length
      ? rawConfig.fields
      : cloneObject(DEFAULT_EXPERT_APPLY_CONFIG.fields),
    uploadField: normalizeUploadField(rawConfig.uploadField || rawConfig.upload),
    validationRules: normalizeValidationRules(rawConfig.validationRules),
    yearOptions: Array.isArray(rawConfig.yearOptions) && rawConfig.yearOptions.length
      ? rawConfig.yearOptions
      : DEFAULT_EXPERT_APPLY_CONFIG.yearOptions,
    serviceCount,
    serviceBlocks: Array.isArray(rawConfig.serviceBlocks) && rawConfig.serviceBlocks.length
      ? rawConfig.serviceBlocks
      : createExpertServiceBlocks(serviceCount),
    priceHint: rawConfig.priceHint || '平台将收取 10% 服务费'
  }
}

function applyExpertApplyConfigToPage(page, config) {
  if (!page || page.mode !== 'expertApplyForm') {
    return page
  }

  const normalizedConfig = normalizeExpertApplyConfig(config)

  return Object.assign({}, page, {
    skillOptions: normalizedConfig.skillOptions,
    fields: normalizedConfig.fields,
    uploadField: normalizedConfig.uploadField,
    serviceBlocks: normalizedConfig.serviceBlocks,
    priceHint: normalizedConfig.priceHint
  })
}

function isExpertApplyFormEmpty(form) {
  if (!form) {
    return true
  }

  const services = form.services || []

  return !trimText(form.skillTags)
    && !form.experienceYears
    && !trimText(form.intro)
    && (!form.uploadFiles || form.uploadFiles.length === 0)
    && services.every((service) => !trimText(service.name) && !trimText(service.price) && !trimText(service.cost))
}

function buildExpertApplyPayload(page, form) {
  const selectedSkill = ((page || {}).skillOptions || []).find((skill) => skill.active)
  const services = (form.services || []).map((service) => ({
    name: trimText(service.name),
    serviceName: trimText(service.name),
    price: trimText(service.price),
    hourlyPrice: trimText(service.price),
    cost: trimText(service.cost)
  }))
  const uploadFiles = (form.uploadFiles || []).map((file) => ({
    name: file.name,
    fileName: file.name,
    path: file.path,
    tempFilePath: file.path,
    size: file.size,
    type: file.type,
    extension: getUploadFileExtension(file)
  }))

  return {
    roleType: 'expert',
    skillDomain: selectedSkill ? selectedSkill.name : '',
    skillDomains: selectedSkill ? [selectedSkill.name] : [],
    skillTags: trimText(form.skillTags),
    experienceYears: form.experienceYears,
    intro: trimText(form.intro),
    personalIntro: trimText(form.intro),
    uploadFiles,
    qualifications: uploadFiles,
    services
  }
}

function roundRpx(value) {
  return Math.round(value * 100) / 100
}

function getApplyShellLayoutStyles() {
  let navTop = APPLY_DEFAULT_NAV_TOP_RPX
  let navHeight = APPLY_DEFAULT_NAV_HEIGHT_RPX
  let contentTop = APPLY_DEFAULT_CONTENT_TOP_RPX

  try {
    if (typeof wx !== 'undefined' && wx.getSystemInfoSync && wx.getMenuButtonBoundingClientRect) {
      const menuButton = wx.getMenuButtonBoundingClientRect()
      const systemInfo = wx.getSystemInfoSync()

      if (menuButton && menuButton.height && systemInfo && systemInfo.windowWidth) {
        const ratio = 750 / systemInfo.windowWidth
        navTop = roundRpx(menuButton.top * ratio)
        navHeight = roundRpx(menuButton.height * ratio)
        const capsuleBottom = (menuButton.top + menuButton.height) * ratio
        contentTop = Math.max(
          APPLY_DEFAULT_CONTENT_TOP_RPX,
          roundRpx(capsuleBottom + APPLY_NAV_BOTTOM_GAP_RPX)
        )
      }
    }
  } catch (error) {
    navTop = APPLY_DEFAULT_NAV_TOP_RPX
    navHeight = APPLY_DEFAULT_NAV_HEIGHT_RPX
    contentTop = APPLY_DEFAULT_CONTENT_TOP_RPX
  }

  const frameTopPadding = Math.max(
    APPLY_FRAME_BOTTOM_PADDING_RPX,
    roundRpx(contentTop - APPLY_STAGE_TOP_RPX)
  )
  return {
    frameStyle: `padding-top: ${frameTopPadding}rpx; padding-bottom: ${APPLY_FRAME_BOTTOM_PADDING_RPX}rpx;`,
    topBgStyle: `top: -${contentTop}rpx; height: ${contentTop}rpx;`,
    navStyle: `top: ${roundRpx(navTop - APPLY_STAGE_TOP_RPX)}rpx; height: ${navHeight}rpx;`,
    phoneStyle: `height: calc(100vh - ${roundRpx(contentTop + APPLY_FRAME_BOTTOM_PADDING_RPX)}rpx); min-height: 0;`
  }
}

const ROLE_COMPARISON_PREVIEW_PAGE = {
  name: '权益对比页',
  mode: 'roleComparison',
  roleBadge: '权益',
  title: '角色权益对比',
  subtitle: '选择适合你的角色，开启不同玩法',
  roles: [
    { key: 'player', name: '玩家', level: 'Lv.1+', icon: UI_ICONS.role.player, active: true },
    { key: 'guide', name: '领路人', level: 'Lv.5+', icon: UI_ICONS.role.guide, active: false },
    { key: 'expert', name: '行家', level: 'Lv.20+', icon: UI_ICONS.role.expert, active: false }
  ],
  benefits: [
    { name: '发起组局', player: '✓', leader: '—', expert: '✓' },
    { name: '加入组局', player: '✓', leader: '✓', expert: '✓' },
    { name: '创建路线', player: '✓', leader: '—', expert: '✓' },
    { name: '分润收益', player: '—', leader: '基础会员40%', expert: '高级会员40%' },
    { name: '服务交易', player: '—', leader: '—', expert: '✓' },
    { name: '数据看板', player: '—', leader: '✓', expert: '✓' },
    { name: '信用背书', player: '—', leader: '✓', expert: '✓' }
  ],
  primary: '立即申请角色'
}

const HOME_CONVERTED_PAGES = [
  {
    name: '启动动画',
    mode: 'launchAnimation',
    brand: '眞好玩',
    onlineText: '2999+人在线',
    title: '真好玩',
    subtitle: 'Zhen Hao Wan',
    actionText: 'GO',
    featureDots: ['pink', 'purple', 'cyan']
  }
]

const EXPERT_APPLY_PREVIEW_PAGES = [
  {
    name: '申请行家内页',
    mode: 'expertApplyForm',
    roleBadge: '申请',
    navTitle: '申请行家',
    saveText: '保存',
    title: '申请成为行家',
    icon: UI_ICONS.role.expert,
    tagline: '我懂玩家需要什么！我申请成为行家',
    reviewHint: '审核预计 1-3 个工作日',
    primary: '提交行家申请',
    skillOptions: DEFAULT_EXPERT_APPLY_CONFIG.skillOptions,
    fields: DEFAULT_EXPERT_APPLY_CONFIG.fields,
    uploadField: DEFAULT_EXPERT_APPLY_CONFIG.uploadField,
    serviceBlocks: createExpertServiceBlocks(DEFAULT_EXPERT_APPLY_CONFIG.serviceCount),
    priceHint: DEFAULT_EXPERT_APPLY_CONFIG.priceHint || '平台将收取 10% 服务费'
  },
  {
    name: '申请行家操作页',
    mode: 'expertApplyOverview',
    roleBadge: '申请',
    navTitle: '申请行家',
    saveText: '保存',
    title: '申请成为行家',
    icon: UI_ICONS.role.expert,
    tagline: '我懂玩家需要什么！我申请成为行家',
    reviewHint: '审核预计 1-3 个工作日',
    primary: '下一步',
    requirementsTitle: '申请条件',
    requirements: [
      { title: '玩家等级达到 Lv.20', status: '当前等级: Lv.21 / 已满足', checked: true },
      { title: '完成实名认证', status: '认证状态: 已通过 / 已满足', checked: true },
      { title: '完成企业认证', status: '认证状态: 已通过 / 已满足', checked: true },
      { title: '发起过 5次以上组局', status: '当前: 5 次 / 已满足', checked: true },
      { title: '信用分 ≥ 90 分', status: '当前: 92 分 / 已满足', checked: true },
      { title: '会员等级≥ 高级会员', status: '当前: 高级会员 / 已满足', checked: true }
    ],
    planTask: {
      title: '提交行家计划书',
      desc: '需描述你的资源、能力和项目说明书',
      action: '去填写 ›',
      checked: false
    },
    benefitsTitle: '行家特权',
    benefits: [
      { icon: UI_ICONS.panel.revenue, text: '有权益的行家可发起有偿局并可获得相应收入' },
      { icon: UI_ICONS.panel.featured, text: '专属行家标识与优先推荐位' },
      { icon: UI_ICONS.panel.data, text: '数据看板：查看服务数据与收益分析' }
    ]
  }
]

const EXPERT_APPLY_WALKTHROUGH_PAGES = [
  EXPERT_APPLY_PREVIEW_PAGES[1],
  EXPERT_APPLY_PREVIEW_PAGES[0]
]

const HOME_PREVIEW_PAGES = [
  { name: '玩家首页', mode: 'roleHome', roleType: 'player' },
  { name: '行家首页', mode: 'roleHome', roleType: 'expert' },
  { name: '领路人首页', mode: 'roleHome', roleType: 'guide' },
  ROLE_COMPARISON_PREVIEW_PAGE,
  EXPERT_APPLY_PREVIEW_PAGES[1],
  EXPERT_APPLY_PREVIEW_PAGES[0]
]

const HOME_PREVIEW_LOOKUP_PAGES = [
  ...HOME_PREVIEW_PAGES,
  ...HOME_CONVERTED_PAGES,
  ...EXPERT_APPLY_WALKTHROUGH_PAGES,
  ROLE_COMPARISON_PREVIEW_PAGE
]

const HOME_PREVIEW_GROUPS = {
  homeAll: HOME_PREVIEW_PAGES,
  expertApplyWalkthrough: EXPERT_APPLY_WALKTHROUGH_PAGES
}

Page({
  data: {
    isHomePreview: false,
    homePreviewIndex: 0,
    homePreviewNo: 1,
    homePreviewTotal: HOME_PREVIEW_PAGES.length,
    homePreviewPages: HOME_PREVIEW_PAGES,
    currentHomePreview: HOME_PREVIEW_PAGES[0],
    homePreviewSingle: false,
    previewWindowWidth: 375,
    applyShellLayout: getApplyShellLayoutStyles(),
    expertApplyYearOptions: EXPERT_APPLY_YEAR_OPTIONS,
    expertApplyForm: createExpertApplyForm(),
    expertApplyErrors: createExpertApplyErrors(),
    expertApplyServiceErrors: createExpertApplyServiceErrors(),
    expertApplyValidationRules: DEFAULT_EXPERT_APPLY_CONFIG.validationRules,
    expertApplyCustomMaxLength: DEFAULT_EXPERT_APPLY_CONFIG.validationRules.customSkill.maxLength,
    expertApplyServiceNameMaxLength: DEFAULT_EXPERT_APPLY_CONFIG.validationRules.serviceName.maxLength,
    uiIcons: UI_ICONS,
    expertApplyCustomInputVisible: false,
    expertApplyCustomInput: '',
    expertApplyCustomError: '',
    expertApplyYearDropdownVisible: false,
    expertApplySubmitting: false,
    expertApplyNoticeVisible: false,
    expertApplyNoticeLines: [],
    roleComparisonReturnTo: '',
    roleComparisonShell: {
      onlineText: '3999人在线',
      navItems: [
        { name: '我的', active: false },
        { name: '元宇宙', active: false },
        { name: '地图', active: false },
        { name: '消息', active: false },
        { name: '首页', active: true }
      ]
    },
    loading: true,
    user: {
      nickname: '',
      todayCreditScore: '',
      level: '',
      member: {
        planName: ''
      },
      needRealname: true
    },
    hero: {
      roleName: '',
      dateLabel: '',
      subtitle: '',
      onlineText: ''
    },
    notices: [],
    quickActions: [],
    recommendedGames: [],
    playerSummary: {
      displayName: '',
      roleLabel: '',
      title: '',
      xpText: '',
      progressPercent: 0,
      nextLevelText: '',
      stats: []
    },
    rankingList: [],
    achievementList: [],
    friendGames: [],
    metaverseEntry: {
      title: '',
      desc: '',
      actionText: '',
      route: ''
    },
    nearbySummary: {
      cityName: '',
      count: 0
    }
  },

  onLoad(options = {}) {
    if (options.ui === '1') {
      if (this.redirectRoleHomePreview(options)) {
        return
      }

      this.setData({
        roleComparisonReturnTo: decodeURIComponent(options.returnTo || '')
      })
      this.enterHomePreview(options.mode || '', options.single === '1', options.role || options.roleType)
      return
    }

    this.loadHome()
  },

  redirectRoleHomePreview(options = {}) {
    if (options.mode !== 'roleHome') {
      return false
    }

    const roleType = String(options.role || options.roleType || '').trim()
    const roleHomeRoutes = {
      player: ROUTES.playerHome,
      expert: ROUTES.expertHome,
      guide: ROUTES.guideHome
    }
    const route = roleHomeRoutes[roleType]

    if (!route || typeof wx.redirectTo !== 'function') {
      return false
    }

    wx.redirectTo({
      url: `/${route}`
    })

    return true
  },

  enterHomePreview(mode, single = false, roleType = '') {
    const normalizedRoleType = String(roleType || '').trim()
    const findPreviewIndex = (pages) => pages.findIndex((page) => {
      if (!page) {
        return false
      }

      if (mode === 'roleHome' && normalizedRoleType) {
        return page.mode === 'roleHome' && page.roleType === normalizedRoleType
      }

      return page.name === mode || page.mode === mode
    })
    const previewPages = HOME_PREVIEW_GROUPS[mode] || HOME_PREVIEW_PAGES
    const index = findPreviewIndex(previewPages)
    const lookupIndex = index >= 0
      ? index
      : findPreviewIndex(HOME_PREVIEW_LOOKUP_PAGES)
    const currentHomePreview = lookupIndex >= 0
      ? HOME_PREVIEW_LOOKUP_PAGES[lookupIndex]
      : previewPages[0]
    const previewWindowWidth = wx.getSystemInfoSync ? wx.getSystemInfoSync().windowWidth : 375

    this.setData({
      isHomePreview: true,
      loading: false,
      applyShellLayout: getApplyShellLayoutStyles(),
      previewWindowWidth,
      homePreviewSingle: single,
      homePreviewPages: previewPages,
      homePreviewIndex: index >= 0 ? index : 0,
      homePreviewNo: index >= 0 ? index + 1 : 1,
      homePreviewTotal: previewPages.length,
      currentHomePreview
    })

    if (currentHomePreview && currentHomePreview.mode === 'expertApplyForm') {
      this.loadExpertApplyConfig()
    }
  },

  async loadExpertApplyConfig() {
    const fallbackConfig = normalizeExpertApplyConfig(DEFAULT_EXPERT_APPLY_CONFIG)

    try {
      const remoteConfig = await roleService.getExpertApplyConfig()
      const config = normalizeExpertApplyConfig(remoteConfig)
      this.applyExpertApplyConfig(config)
    } catch (error) {
      this.applyExpertApplyConfig(fallbackConfig)
    }
  },

  applyExpertApplyConfig(config) {
    const normalizedConfig = normalizeExpertApplyConfig(config)
    const currentHomePreview = applyExpertApplyConfigToPage(this.data.currentHomePreview, normalizedConfig)
    const homePreviewPages = (this.data.homePreviewPages || []).map((page) => applyExpertApplyConfigToPage(page, normalizedConfig))
    const shouldResetForm = isExpertApplyFormEmpty(this.data.expertApplyForm)

    this.setData({
      currentHomePreview,
      homePreviewPages,
      expertApplyYearOptions: normalizedConfig.yearOptions,
      expertApplyForm: shouldResetForm ? createExpertApplyForm(normalizedConfig.serviceCount) : this.data.expertApplyForm,
      expertApplyErrors: createExpertApplyErrors(),
      expertApplyServiceErrors: createExpertApplyServiceErrors(
        shouldResetForm ? normalizedConfig.serviceCount : (this.data.expertApplyForm.services || []).length
      ),
      expertApplyValidationRules: normalizedConfig.validationRules,
      expertApplyCustomMaxLength: getPositiveInteger(normalizedConfig.validationRules.customSkill.maxLength, 8),
      expertApplyServiceNameMaxLength: getPositiveInteger(normalizedConfig.validationRules.serviceName.maxLength, 20),
      expertApplyCustomInputVisible: false,
      expertApplyCustomInput: '',
      expertApplyCustomError: '',
      expertApplyYearDropdownVisible: false
    })
  },

  handleHomePreviewTap(event) {
    if (!this.data.isHomePreview) {
      return
    }

    if (this.data.homePreviewSingle) {
      return
    }

    const dataset = (event.currentTarget && event.currentTarget.dataset) || {}
    const detail = event.detail || {}
    const detailDirection = Number(detail.direction)
    const datasetDirection = Number(dataset.direction)
    const touch = event.changedTouches && event.changedTouches[0]
    const x = touch ? touch.clientX : detail.x
    const direction = detailDirection || datasetDirection || (x < this.data.previewWindowWidth / 2 ? -1 : 1)
    const previewPages = this.data.homePreviewPages || HOME_PREVIEW_PAGES
    const total = previewPages.length
    const currentIndex = this.data.homePreviewIndex
    const nextIndex = currentIndex + direction

    if (nextIndex < 0 || nextIndex >= total) {
      return
    }

    this.setData({
      homePreviewIndex: nextIndex,
      homePreviewNo: nextIndex + 1,
      currentHomePreview: previewPages[nextIndex]
    })
  },

  stopExpertApplyTap() {},

  showExpertApplyForm() {
    const previewPages = this.data.homePreviewPages || []
    const formIndex = previewPages.findIndex((page) => page && page.mode === 'expertApplyForm')

    if (formIndex < 0) {
      return
    }

    this.setData({
      homePreviewIndex: formIndex,
      homePreviewNo: formIndex + 1,
      currentHomePreview: previewPages[formIndex],
      expertApplyCustomInputVisible: false,
      expertApplyCustomInput: '',
      expertApplyCustomError: '',
      expertApplyYearDropdownVisible: false
    })

    this.loadExpertApplyConfig()
  },

  handleExpertApplyPlanTap() {
  },

  handleExpertApplyPrimaryTap() {
    const currentHomePreview = this.data.currentHomePreview || {}

    if (currentHomePreview.mode === 'expertApplyOverview') {
      this.showExpertApplyForm()
      return
    }

    this.handleExpertApplySubmit()
  },

  onExpertApplySkillTap(event) {
    const index = Number(event.currentTarget.dataset.index)
    const skillOptions = (((this.data.currentHomePreview || {}).skillOptions) || []).map((skill) => ({ ...skill }))
    const selectedSkill = skillOptions[index]

    if (!selectedSkill) {
      return
    }

    if (selectedSkill.custom) {
      this.setData({
        expertApplyCustomInputVisible: true,
        expertApplyCustomError: '',
        expertApplyYearDropdownVisible: false
      })
      return
    }

    const nextOptions = skillOptions.map((skill, skillIndex) => ({
      ...skill,
      active: skillIndex === index
    }))

    this.setData({
      'currentHomePreview.skillOptions': nextOptions,
      'expertApplyErrors.skillDomain': ''
    })
  },

  onExpertApplyCustomInput(event) {
    this.setData({
      expertApplyCustomInput: event.detail.value,
      expertApplyCustomError: ''
    })
  },

  confirmExpertApplyCustomSkill() {
    const value = trimText(this.data.expertApplyCustomInput)
    const rules = this.data.expertApplyValidationRules || DEFAULT_EXPERT_APPLY_CONFIG.validationRules
    const customRule = rules.customSkill || DEFAULT_EXPERT_APPLY_CONFIG.validationRules.customSkill
    const minLength = getPositiveInteger(customRule.minLength, 2)
    const maxLength = getPositiveInteger(customRule.maxLength, 8)

    if (!value) {
      this.setData({
        expertApplyCustomError: '请输入技能领域'
      })
      return
    }

    if (value.length < minLength || value.length > maxLength) {
      this.setData({
        expertApplyCustomError: `技能领域需为 ${minLength}-${maxLength} 个字`
      })
      return
    }

    const skillOptions = (((this.data.currentHomePreview || {}).skillOptions) || []).map((skill) => ({ ...skill }))
    const nextOptions = skillOptions
      .filter((skill) => !skill.custom && !skill.generatedCustom)
      .map((skill) => ({ ...skill, active: false }))

    nextOptions.push({
      name: value,
      active: true,
      generatedCustom: true
    })
    nextOptions.push({ name: '+ 自定义', active: false, custom: true })

    this.setData({
      'currentHomePreview.skillOptions': nextOptions,
      'expertApplyErrors.skillDomain': '',
      expertApplyCustomInputVisible: false,
      expertApplyCustomInput: '',
      expertApplyCustomError: ''
    })
  },

  onExpertApplyInput(event) {
    const field = event.currentTarget.dataset.field
    const value = event.detail.value

    if (!field) {
      return
    }

    this.setData({
      [`expertApplyForm.${field}`]: value,
      [`expertApplyErrors.${field}`]: ''
    })
  },

  toggleExpertApplyYearDropdown() {
    this.setData({
      expertApplyYearDropdownVisible: !this.data.expertApplyYearDropdownVisible,
      expertApplyCustomInputVisible: false,
      expertApplyCustomError: ''
    })
  },

  onExpertApplyYearOptionTap(event) {
    const index = Number(event.currentTarget.dataset.index)
    const option = this.data.expertApplyYearOptions[index] || ''

    if (!option) {
      return
    }

    this.setData({
      'expertApplyForm.experienceYearIndex': index,
      'expertApplyForm.experienceYears': option,
      'expertApplyErrors.experienceYears': '',
      expertApplyYearDropdownVisible: false
    })
  },

  onExpertApplyServiceNameInput(event) {
    const index = Number(event.currentTarget.dataset.index)
    const value = event.detail.value

    this.setData({
      [`expertApplyForm.services[${index}].name`]: value,
      [`expertApplyServiceErrors[${index}].name`]: ''
    })
  },

  onExpertApplyMoneyInput(event) {
    const index = Number(event.currentTarget.dataset.index)
    const field = event.currentTarget.dataset.field
    const moneyRule = getMoneyRule(this.data.expertApplyValidationRules)
    const rawValue = String(event.detail.value || '').replace(/[^\d.]/g, '')
    const firstDotIndex = rawValue.indexOf('.')
    const normalizedValue = firstDotIndex >= 0
      ? rawValue.slice(0, firstDotIndex + 1) + rawValue.slice(firstDotIndex + 1).replace(/\./g, '')
      : rawValue
    const value = normalizedValue.replace(
      new RegExp(`^(\\d{0,${moneyRule.integerMaxLength}})(\\.\\d{0,${moneyRule.decimalMaxLength}})?.*$`),
      '$1$2'
    )

    if (!field) {
      return
    }

    this.setData({
      [`expertApplyForm.services[${index}].${field}`]: value,
      [`expertApplyServiceErrors[${index}].${field}`]: ''
    })
  },

  onExpertApplyUploadTap() {
    const uploadField = (this.data.currentHomePreview || {}).uploadField || DEFAULT_EXPERT_APPLY_CONFIG.uploadField
    const maxCount = getPositiveInteger(uploadField.maxCount, DEFAULT_EXPERT_APPLY_CONFIG.uploadField.maxCount)
    const acceptTypes = Array.isArray(uploadField.acceptTypes) ? uploadField.acceptTypes : DEFAULT_EXPERT_APPLY_CONFIG.uploadField.acceptTypes
    const acceptText = acceptTypes.join('、')
    const currentFiles = this.data.expertApplyForm.uploadFiles || []
    const remainingCount = maxCount - currentFiles.length

    if (remainingCount <= 0) {
      toast.info(`最多上传 ${maxCount} 个文件`)
      return
    }

    if (typeof wx === 'undefined' || typeof wx.chooseMessageFile !== 'function') {
      toast.info('当前环境不支持文件选择')
      return
    }

    wx.chooseMessageFile({
      count: remainingCount,
      type: 'file',
      success: (result) => {
        const rawFiles = (result.tempFiles || []).slice(0, remainingCount).map(normalizeUploadFile)
        const selectedFiles = rawFiles.filter((file) => isAllowedUploadFile(file, acceptTypes))
        const rejectedFiles = rawFiles.filter((file) => !isAllowedUploadFile(file, acceptTypes))
        const files = currentFiles.concat(selectedFiles).slice(0, maxCount)

        this.setData({
          'expertApplyForm.uploadFiles': files,
          'expertApplyErrors.uploadFiles': rejectedFiles.length
            ? `仅支持 ${acceptText} 格式：${rejectedFiles.map((file) => file.name).join('、')}`
            : ''
        })
      },
      fail: (error) => {
        if (error && String(error.errMsg || '').includes('cancel')) {
          return
        }

        toast.info('文件选择失败，请重试')
      }
    })
  },

  onExpertApplyRemoveFileTap(event) {
    const index = Number(event.currentTarget.dataset.index)
    const files = (this.data.expertApplyForm.uploadFiles || []).filter((file, fileIndex) => fileIndex !== index)

    this.setData({
      'expertApplyForm.uploadFiles': files
    })
  },

  validateExpertApplyForm() {
    const form = this.data.expertApplyForm || createExpertApplyForm()
    const errors = createExpertApplyErrors()
    const serviceErrors = createExpertApplyServiceErrors((form.services || []).length)
    const rules = this.data.expertApplyValidationRules || DEFAULT_EXPERT_APPLY_CONFIG.validationRules
    const skillTagsRule = rules.skillTags || DEFAULT_EXPERT_APPLY_CONFIG.validationRules.skillTags
    const introRule = rules.intro || DEFAULT_EXPERT_APPLY_CONFIG.validationRules.intro
    const serviceNameRule = rules.serviceName || DEFAULT_EXPERT_APPLY_CONFIG.validationRules.serviceName
    const moneyRule = getMoneyRule(rules)
    const selectedSkill = ((this.data.currentHomePreview || {}).skillOptions || []).find((skill) => skill.active)
    const skillTagsMinLength = getPositiveInteger(skillTagsRule.minLength, 2)
    const skillTagsMaxLength = getPositiveInteger(skillTagsRule.maxLength, 30)
    const introMinLength = getPositiveInteger(introRule.minLength, 50)
    const introMaxLength = getPositiveInteger(introRule.maxLength, 300)
    const serviceNameMinLength = getPositiveInteger(serviceNameRule.minLength, 2)
    const serviceNameMaxLength = getPositiveInteger(serviceNameRule.maxLength, 20)

    if (!selectedSkill) {
      errors.skillDomain = '请选择技能领域'
    }

    const skillTags = trimText(form.skillTags)
    if (!skillTags) {
      errors.skillTags = '请填写技能标签'
    } else if (skillTags.length < skillTagsMinLength || skillTags.length > skillTagsMaxLength) {
      errors.skillTags = `技能标签需为 ${skillTagsMinLength}-${skillTagsMaxLength} 个字`
    }

    if (!form.experienceYears) {
      errors.experienceYears = '请选择从业年限'
    }

    const intro = trimText(form.intro)
    if (!intro) {
      errors.intro = '请填写个人简介'
    } else if (intro.length < introMinLength || intro.length > introMaxLength) {
      errors.intro = `个人简介需为 ${introMinLength}-${introMaxLength} 个字`
    }

    if (!form.uploadFiles || form.uploadFiles.length === 0) {
      errors.uploadFiles = '请上传作品集及凭证'
    }

    const services = form.services || []
    services.forEach((service, index) => {
      const name = trimText(service.name)
      const price = trimText(service.price)
      const cost = trimText(service.cost)

      if (!name) {
        serviceErrors[index].name = '请填写业务名称'
      } else if (name.length < serviceNameMinLength || name.length > serviceNameMaxLength) {
        serviceErrors[index].name = `业务名称需为 ${serviceNameMinLength}-${serviceNameMaxLength} 个字`
      }

      if (!price) {
        serviceErrors[index].price = '请填写服务定价'
      } else if (!isValidMoney(price, rules)) {
        serviceErrors[index].price = `金额需为最多 ${moneyRule.integerMaxLength} 位整数和 ${moneyRule.decimalMaxLength} 位小数`
      }

      if (cost && !isValidMoney(cost, rules)) {
        serviceErrors[index].cost = `金额需为最多 ${moneyRule.integerMaxLength} 位整数和 ${moneyRule.decimalMaxLength} 位小数`
      }
    })

    this.setData({
      expertApplyErrors: errors,
      expertApplyServiceErrors: serviceErrors
    })

    const hasServiceError = serviceErrors.some((serviceError) => hasAnyError(serviceError))

    if (hasAnyError(errors) || hasServiceError) {
      return false
    }

    return true
  },

  async handleExpertApplySubmit() {
    if (this.data.expertApplySubmitting) {
      return
    }

    if (!this.validateExpertApplyForm()) {
      toast.info('请完善必填项并检查格式')
      return
    }

    this.setData({
      expertApplySubmitting: true
    })

    try {
      const payload = buildExpertApplyPayload(this.data.currentHomePreview, this.data.expertApplyForm)
      await roleService.submitRoleApplication(payload)

      toast.info('申请已提交，\r\n请耐心等待审核')
      setTimeout(() => {
        if (typeof wx.reLaunch === 'function') {
          wx.reLaunch({
            url: `/${ROUTES.playerHome}`
          })
        }
      }, EXPERT_APPLY_SUCCESS_REDIRECT_DELAY_MS)
    } catch (error) {
      toast.info(error.message || '提交失败')
    } finally {
      this.setData({
        expertApplySubmitting: false
      })
    }
  },

  handleExpertApplyBackTap() {
    const previewPages = this.data.homePreviewPages || []
    const currentIndex = this.data.homePreviewIndex
    const previousPage = previewPages[currentIndex - 1]

    if (this.data.currentHomePreview && this.data.currentHomePreview.mode === 'expertApplyForm' && previousPage) {
      this.setData({
        homePreviewIndex: currentIndex - 1,
        homePreviewNo: currentIndex,
        currentHomePreview: previousPage,
        expertApplyCustomInputVisible: false,
        expertApplyCustomInput: '',
        expertApplyCustomError: '',
        expertApplyYearDropdownVisible: false
      })
      return
    }

    const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []

    if (pages.length > 1 && typeof wx.navigateBack === 'function') {
      wx.navigateBack()
      return
    }

    if (typeof wx.reLaunch === 'function') {
      wx.reLaunch({
        url: `/${ROUTES.playerHome}`
      })
    }
  },

  handleRoleCompareBackTap() {
    const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []

    if (pages.length > 1 && typeof wx.navigateBack === 'function') {
      wx.navigateBack()
      return
    }

    if (this.data.roleComparisonReturnTo && typeof wx.redirectTo === 'function') {
      wx.redirectTo({
        url: this.data.roleComparisonReturnTo.startsWith('/')
          ? this.data.roleComparisonReturnTo
          : `/${this.data.roleComparisonReturnTo}`
      })
      return
    }

    if (typeof wx.reLaunch === 'function') {
      wx.reLaunch({
        url: `/${ROUTES.playerHome}`
      })
    }
  },

  handleRoleCompareApplyTap() {
    this.enterHomePreview('expertApplyOverview', this.data.homePreviewSingle, 'expert')
  },

  async loadHome() {
    try {
      const home = await homeService.getHome()
      this.setData({
        loading: false,
        user: home.user,
        hero: home.hero,
        notices: home.notices,
        quickActions: home.quickActions,
        recommendedGames: home.recommendedGames,
        playerSummary: home.playerSummary,
        rankingList: home.rankingList,
        achievementList: home.achievementList,
        friendGames: home.friendGames,
        metaverseEntry: home.metaverseEntry,
        nearbySummary: home.nearbySummary
      })
    } catch (error) {
      this.setData({
        loading: false
      })
      toast.info(error.message || '首页加载失败')
    }
  },

  goAction(event) {
    const route = event.currentTarget.dataset.route

    if (!route) {
      return
    }

    toast.developing()
  }
})
