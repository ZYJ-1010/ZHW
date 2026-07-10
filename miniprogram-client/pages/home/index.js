const homeService = require('../../services/home')
const roleService = require('../../services/role')
const toast = require('../../utils/toast')
const { ROUTES } = require('../../config/routes')
const UI_ICONS = require('../../config/ui-icons')
const { navigateShellRoute } = require('../../utils/shell-nav')
const {
  readRoleApplyDraft,
  saveRoleApplyDraft,
  clearRoleApplyDraft
} = require('../../utils/role-apply-draft')
const { isRoleApplyResultViewed } = require('../../utils/role-apply-result-view')

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
  skillTags: '',
  experienceYearIndex: -1,
  experienceYears: '',
  intro: '',
  uploadFiles: [],
  services: []
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
    { name: '+自定义', active: false, custom: true }
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
  requirementsTitle: '申请条件',
  requirements: [
    { title: '玩家等级达到 Lv.20', text: '以后台资格规则为准', done: false },
    { title: '完成实名认证', text: '行家必须实名', done: false },
    { title: '完成企业认证', text: '以认证记录为准', done: false },
    { title: '发起过 5 次以上组局', text: '以后台组局记录为准', done: false },
    { title: '信用分 ≥ 90 分', text: '以信用记录为准', done: false },
    { title: '会员等级 ≥ 高级会员', text: '以会员状态为准', done: false }
  ],
  planTask: {
    title: '提交行家计划书',
    text: '需描述你的资源、能力和项目说明书',
    done: false,
    action: '去填写 ›'
  },
  perksTitle: '行家特权',
  perks: [
    { icon: UI_ICONS.panel.revenue, text: '有权益的行家可发起有偿局并可获得相应收入' },
    { icon: UI_ICONS.panel.featured, text: '专属行家标识与优先推荐位' },
    { icon: UI_ICONS.panel.data, text: '数据看板：查看服务数据与收益分析' }
  ],
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
  priceHint: '平台将收取 10% 服务费',
  primaryText: '下一步',
  helperText: '审核预计 1-3 个工作日'
}
const ROLE_APPLY_PREVIEW_META = {
  expert: {
    roleType: 'expert',
    navTitle: '申请行家',
    pageName: '申请行家',
    title: '申请成为行家',
    icon: UI_ICONS.role.expert,
    tagline: '我懂玩家需要什么！我申请成为行家',
    formPrimaryText: '提交行家申请',
    fallbackPrimaryText: DEFAULT_EXPERT_APPLY_CONFIG.primaryText
  },
  guide: {
    roleType: 'guide',
    navTitle: '申请领路人',
    pageName: '申请领路人',
    title: '申请成为领路人',
    icon: UI_ICONS.role.guide,
    tagline: '我愿意带领更多人一起玩！我申请成为领路人',
    formPrimaryText: '提交领路人申请',
    fallbackPrimaryText: DEFAULT_EXPERT_APPLY_CONFIG.primaryText
  }
}

function getPositiveInteger(value, fallback) {
  const number = Number(value)
  return Number.isInteger(number) && number > 0 ? number : fallback
}

function createExpertApplyServices(count = EXPERT_SERVICE_COUNT) {
  const serviceCount = getPositiveInteger(count, EXPERT_SERVICE_COUNT)

  return Array.from({ length: serviceCount }, () => ({
    name: '',
    price: '',
    cost: ''
  }))
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
      const rawName = trimText(item.name || item.label || item.text || item.value)
      const isCustom = Boolean(item.custom || item.type === 'custom' || rawName === '+ 自定义' || rawName === '+自定义')
      const name = isCustom ? '+自定义' : rawName

      if (!name) {
        return null
      }

      return {
        name,
        active: Boolean(item.active),
        custom: isCustom,
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
    nextOptions.push({ name: '+自定义', active: false, custom: true })
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

function normalizeExpertRequirements(requirements) {
  const sourceRequirements = Array.isArray(requirements) && requirements.length
    ? requirements
    : DEFAULT_EXPERT_APPLY_CONFIG.requirements

  return sourceRequirements
    .map((item) => {
      const title = trimText(item.title || item.name || item.label)
      const status = trimText(item.status || item.text || item.desc || item.description)

      if (!title && !status) {
        return null
      }

      return {
        title: title || status,
        status,
        checked: isExpertRequirementChecked(item, status)
      }
    })
    .filter(Boolean)
}

function isExpertRequirementChecked(item = {}, statusText = '') {
  const boolKeys = ['checked', 'done', 'completed', 'passed', 'met', 'satisfied']
  const matchedBoolKey = boolKeys.find((key) => typeof item[key] === 'boolean')

  if (matchedBoolKey) {
    return item[matchedBoolKey]
  }

  const stateText = trimText(item.state || item.result || item.statusText || statusText)

  if (/不满足|未满足|未通过|未完成|待|失败|failed|blocked|false/i.test(stateText)) {
    return false
  }

  if (/已满足|满足|已通过|通过|已完成|完成|达标|met|passed|done|completed|true/i.test(stateText)) {
    return true
  }

  return false
}

function normalizeExpertPlanTask(planTask) {
  const rawTask = planTask && typeof planTask === 'object'
    ? planTask
    : DEFAULT_EXPERT_APPLY_CONFIG.planTask

  return {
    title: rawTask.title || DEFAULT_EXPERT_APPLY_CONFIG.planTask.title,
    desc: rawTask.desc || rawTask.text || DEFAULT_EXPERT_APPLY_CONFIG.planTask.text,
    action: rawTask.action || DEFAULT_EXPERT_APPLY_CONFIG.planTask.action,
    checked: Boolean(rawTask.checked || rawTask.done || rawTask.completed)
  }
}

function normalizeExpertPerks(perks) {
  const sourcePerks = Array.isArray(perks) && perks.length
    ? perks
    : DEFAULT_EXPERT_APPLY_CONFIG.perks

  return sourcePerks
    .map((item) => {
      const text = trimText(item.text || item.title || item.desc || item.description)

      if (!text) {
        return null
      }

      return {
        icon: item.icon || item.iconText || UI_ICONS.panel.featured,
        text
      }
    })
    .filter(Boolean)
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

function extractResponseData(result) {
  if (
    result
    && typeof result === 'object'
    && Object.prototype.hasOwnProperty.call(result, 'data')
    && (
      Object.prototype.hasOwnProperty.call(result, 'code')
      || Object.prototype.hasOwnProperty.call(result, 'message')
      || Object.prototype.hasOwnProperty.call(result, 'requestId')
    )
  ) {
    return result.data || {}
  }

  return result || {}
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
    priceHint: rawConfig.priceHint || DEFAULT_EXPERT_APPLY_CONFIG.priceHint,
    requirementsTitle: rawConfig.requirementsTitle || DEFAULT_EXPERT_APPLY_CONFIG.requirementsTitle,
    requirements: normalizeExpertRequirements(rawConfig.requirements),
    planTask: normalizeExpertPlanTask(rawConfig.planTask),
    benefitsTitle: rawConfig.benefitsTitle || rawConfig.perksTitle || DEFAULT_EXPERT_APPLY_CONFIG.perksTitle,
    benefits: normalizeExpertPerks(rawConfig.benefits || rawConfig.perks),
    primaryText: rawConfig.primaryText || DEFAULT_EXPERT_APPLY_CONFIG.primaryText,
    helperText: rawConfig.helperText || DEFAULT_EXPERT_APPLY_CONFIG.helperText
  }
}

function applyExpertApplyConfigToPage(page, config) {
  if (!page || (page.mode !== 'expertApplyForm' && page.mode !== 'expertApplyOverview')) {
    return page
  }

  const normalizedConfig = normalizeExpertApplyConfig(config)

  if (page.mode === 'expertApplyOverview') {
    return Object.assign({}, page, {
      requirementsTitle: normalizedConfig.requirementsTitle,
      requirements: normalizedConfig.requirements,
      planTask: normalizedConfig.planTask,
      benefitsTitle: normalizedConfig.benefitsTitle,
      benefits: normalizedConfig.benefits,
      primary: normalizedConfig.primaryText,
      reviewHint: normalizedConfig.helperText
    })
  }

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

function normalizeExpertApplyDraftForm(draft, serviceCount) {
  const sourceForm = draft && draft.form && typeof draft.form === 'object' ? draft.form : {}
  const baseForm = createExpertApplyForm(serviceCount)

  return Object.assign({}, baseForm, sourceForm, {
    uploadFiles: Array.isArray(sourceForm.uploadFiles) ? sourceForm.uploadFiles : [],
    services: createExpertApplyServices(serviceCount).map((service, index) => (
      Object.assign({}, service, (sourceForm.services || [])[index] || {})
    ))
  })
}

function mergeExpertDraftSkillOptions(baseOptions, draftOptions) {
  if (!Array.isArray(draftOptions) || !draftOptions.length) {
    return baseOptions
  }

  const selectedDraftSkill = draftOptions.find((skill) => skill && skill.active && !skill.custom)

  if (!selectedDraftSkill || !selectedDraftSkill.name) {
    return baseOptions
  }

  let matched = false
  const nextOptions = (baseOptions || []).map((skill) => {
    const isSelected = !skill.custom && skill.name === selectedDraftSkill.name

    if (isSelected) {
      matched = true
    }

    return Object.assign({}, skill, {
      active: isSelected
    })
  })

  if (matched) {
    return nextOptions
  }

  const generatedSkill = {
    name: selectedDraftSkill.name,
    active: true,
    custom: false,
    generatedCustom: true
  }
  const customIndex = nextOptions.findIndex((skill) => skill.custom)

  if (customIndex >= 0) {
    nextOptions.splice(customIndex, 0, generatedSkill)
    return nextOptions
  }

  return nextOptions.concat(generatedSkill)
}

function applyExpertDraftToPage(page, draft) {
  if (!page || page.mode !== 'expertApplyForm' || !draft) {
    return page
  }

  return Object.assign({}, page, {
    skillOptions: mergeExpertDraftSkillOptions(page.skillOptions, draft.skillOptions)
  })
}

function getUnmetExpertRequirements(page = {}) {
  const requirements = Array.isArray(page.requirements) ? page.requirements : []

  return requirements.filter((requirement) => requirement && !requirement.checked)
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

function applyRoleBenefitConfigToPage(page, config = {}) {
  if (!page || page.mode !== 'roleComparison') {
    return page
  }

  const comparison = config.roleComparison || {}
  const roles = Array.isArray(comparison.roles) && comparison.roles.length
    ? comparison.roles
    : ROLE_COMPARISON_PREVIEW_PAGE.roles

  return {
    ...page,
    ...comparison,
    mode: 'roleComparison',
    roles: roles.map((role) => ({
      ...role,
      icon: role.icon || (UI_ICONS.role && UI_ICONS.role[role.key]) || ''
    })),
    benefits: Array.isArray(comparison.benefits) ? comparison.benefits : page.benefits,
    primary: comparison.primary || page.primary
  }
}

function normalizeComparisonRoleType(value) {
  const roleType = String(value || '').trim()

  if (roleType === 'guide' || roleType === 'leader' || roleType === '领路人') {
    return 'guide'
  }

  if (roleType === 'player' || roleType === '玩家') {
    return 'player'
  }

  return 'expert'
}

function normalizeApplyRoleType(value) {
  const roleType = normalizeComparisonRoleType(value)

  return roleType === 'guide' ? 'guide' : 'expert'
}

function normalizeRoleApplyStatus(value) {
  const status = String(value || '').trim()

  if (status === 'pending' || status === 'reviewing' || status === 'auditing' || status === 'pending_audit' || status === '待处理' || status === '审核中') {
    return 'pending'
  }

  if (status === 'approved' || status === 'active' || status === 'enabled' || status === 'passed' || status === 'success' || status === '已通过') {
    return 'approved'
  }

  if (status === 'rejected' || status === 'reject' || status === 'failed' || status === 'rejected_audit' || status === '已驳回' || status === '未通过') {
    return 'rejected'
  }

  return status
}

function extractRoleApplications(data = {}) {
  if (Array.isArray(data)) {
    return data
  }

  if (Array.isArray(data.applications)) {
    return data.applications
  }

  if (Array.isArray(data.items)) {
    return data.items
  }

  return []
}

function getExistingRoleApplication(roleInfo = {}, roleType = 'expert', status = '') {
  const normalizedRoleType = normalizeApplyRoleType(roleType)
  const applications = extractRoleApplications(roleInfo)
  const normalizedStatus = normalizeRoleApplyStatus(status)

  return applications.find((item) => {
    if (normalizeApplyRoleType(item && (item.roleCode || item.roleType || item.role_code || item.role)) !== normalizedRoleType) {
      return false
    }

    return normalizedStatus
      ? normalizeRoleApplyStatus(item.status || item.roleStatus || item.role_status) === normalizedStatus
      : true
  }) || null
}

function getExistingRoleApplyState(roleInfo = {}, roleType = 'expert') {
  const normalizedRoleType = normalizeApplyRoleType(roleType)
  const pendingApplication = getExistingRoleApplication(roleInfo, normalizedRoleType, 'pending')

  if (pendingApplication) {
    return {
      status: 'pending',
      viewed: false,
      application: pendingApplication
    }
  }

  const approvedApplication = getExistingRoleApplication(roleInfo, normalizedRoleType, 'approved')
  if (approvedApplication) {
    return {
      status: 'approved',
      viewed: isRoleApplyResultViewed({
        roleType: normalizedRoleType,
        status: 'approved',
        application: approvedApplication
      }),
      application: approvedApplication
    }
  }

  const rejectedApplication = getExistingRoleApplication(roleInfo, normalizedRoleType, 'rejected')
  if (rejectedApplication) {
    return {
      status: 'rejected',
      viewed: isRoleApplyResultViewed({
        roleType: normalizedRoleType,
        status: 'rejected',
        application: rejectedApplication
      }),
      application: rejectedApplication
    }
  }

  const roleStatusMap = roleInfo.roleStatusMap || roleInfo.role_status_map || {}
  const mappedStatus = normalizeRoleApplyStatus(roleStatusMap[normalizedRoleType])

  if (mappedStatus === 'approved' || mappedStatus === 'rejected') {
    const application = { id: `${normalizedRoleType}-${mappedStatus}` }

    return {
      status: mappedStatus,
      viewed: isRoleApplyResultViewed({
        roleType: normalizedRoleType,
        status: mappedStatus,
        application
      }),
      application
    }
  }

  const matchedApplication = getExistingRoleApplication(roleInfo, normalizedRoleType)

  return {
    status: matchedApplication
      ? normalizeRoleApplyStatus(matchedApplication.status || matchedApplication.roleStatus || matchedApplication.role_status)
      : mappedStatus,
    viewed: false,
    application: matchedApplication
  }
}

function getExistingRoleApplyStatus(roleInfo = {}, roleType = 'expert') {
  return getExistingRoleApplyState(roleInfo, roleType).status
}

function getRoleHomeRoute(roleType = 'expert') {
  const normalizedRoleType = normalizeApplyRoleType(roleType)

  return normalizedRoleType === 'guide' ? ROUTES.guideHome : ROUTES.expertHome
}

function getRoleApplyResultPageId(roleType = 'expert', status = 'approved') {
  const normalizedRoleType = normalizeApplyRoleType(roleType)
  const normalizedStatus = normalizeRoleApplyStatus(status)

  if (normalizedStatus === 'rejected') {
    return `${normalizedRoleType}Rejected`
  }

  if (normalizedStatus === 'approved') {
    return `${normalizedRoleType}Passed`
  }

  return ''
}

function getRoleApplyResultRoute(roleType = 'expert', status = 'approved') {
  const normalizedRoleType = normalizeApplyRoleType(roleType)
  const resultPageId = getRoleApplyResultPageId(normalizedRoleType, status)

  return resultPageId
    ? `${ROUTES.homeOther}?page=${resultPageId}&single=1&roleType=${normalizedRoleType}`
    : ''
}

function getApplyRoleName(roleType = 'expert') {
  const normalizedRoleType = normalizeApplyRoleType(roleType)
  const meta = ROLE_APPLY_PREVIEW_META[normalizedRoleType] || ROLE_APPLY_PREVIEW_META.expert

  return meta.roleName || (normalizedRoleType === 'guide' ? '领路人' : '行家')
}

function applyRoleComparisonApplyStatus(page, roleInfo, roleType = 'expert') {
  if (!page || page.mode !== 'roleComparison') {
    return page
  }

  const normalizedRoleType = normalizeApplyRoleType(roleType)
  const state = getExistingRoleApplyState(roleInfo, normalizedRoleType)
  const status = state.status
  const roleName = getApplyRoleName(normalizedRoleType)

  if (status === 'pending') {
    return Object.assign({}, page, {
      primaryDisabled: true,
      primaryDisabledReason: 'pending',
      primaryDisabledText: `申请${roleName}审核中`,
      primaryOverrideText: '',
      primaryResultStatus: ''
    })
  }

  if ((status === 'approved' || status === 'rejected') && !state.viewed) {
    return Object.assign({}, page, {
      primaryDisabled: false,
      primaryDisabledReason: '',
      primaryDisabledText: '',
      primaryOverrideText: `查看${roleName}审核结果`,
      primaryResultStatus: status
    })
  }

  if (status === 'approved' && state.viewed) {
    return Object.assign({}, page, {
      primaryDisabled: true,
      primaryDisabledReason: 'approved',
      primaryDisabledText: `已成为${roleName}`,
      primaryOverrideText: '',
      primaryResultStatus: ''
    })
  }

  return Object.assign({}, page, {
    primaryDisabled: false,
    primaryDisabledReason: '',
    primaryDisabledText: '',
    primaryOverrideText: '',
    primaryResultStatus: ''
  })
}

function getRoleApplyPreviewMeta(roleType) {
  return ROLE_APPLY_PREVIEW_META[normalizeApplyRoleType(roleType)] || ROLE_APPLY_PREVIEW_META.expert
}

function createRoleApplyPreviewPages(roleType = 'expert') {
  const meta = getRoleApplyPreviewMeta(roleType)
  const overviewPage = {
    name: `${meta.pageName}操作页`,
    mode: 'expertApplyOverview',
    applyRoleType: meta.roleType,
    roleBadge: '申请',
    navTitle: meta.navTitle,
    saveText: '保存',
    title: meta.title,
    icon: meta.icon,
    tagline: meta.tagline,
    reviewHint: DEFAULT_EXPERT_APPLY_CONFIG.helperText,
    primary: meta.fallbackPrimaryText,
    requirementsTitle: DEFAULT_EXPERT_APPLY_CONFIG.requirementsTitle,
    requirements: normalizeExpertRequirements(DEFAULT_EXPERT_APPLY_CONFIG.requirements),
    planTask: normalizeExpertPlanTask(DEFAULT_EXPERT_APPLY_CONFIG.planTask),
    benefitsTitle: DEFAULT_EXPERT_APPLY_CONFIG.perksTitle,
    benefits: normalizeExpertPerks(DEFAULT_EXPERT_APPLY_CONFIG.perks)
  }

  if (meta.roleType === 'guide') {
    return [overviewPage]
  }

  return [
    overviewPage,
    {
      name: `${meta.pageName}内页`,
      mode: 'expertApplyForm',
      applyRoleType: meta.roleType,
      roleBadge: '申请',
      navTitle: meta.navTitle,
      saveText: '保存',
      title: meta.title,
      icon: meta.icon,
      tagline: meta.tagline,
      reviewHint: '审核预计 1-3 个工作日',
      primary: meta.formPrimaryText,
      skillOptions: DEFAULT_EXPERT_APPLY_CONFIG.skillOptions,
      fields: DEFAULT_EXPERT_APPLY_CONFIG.fields,
      uploadField: DEFAULT_EXPERT_APPLY_CONFIG.uploadField,
      serviceBlocks: createExpertServiceBlocks(DEFAULT_EXPERT_APPLY_CONFIG.serviceCount),
      priceHint: DEFAULT_EXPERT_APPLY_CONFIG.priceHint || '平台将收取 10% 服务费'
    }
  ]
}

const HOME_CONVERTED_PAGES = [
  {
    name: '启动动画',
    mode: 'launchAnimation',
    brand: '眞好玩',
    onlineText: '在线',
    title: '真好玩',
    subtitle: 'Zhen Hao Wan',
    actionText: 'GO',
    featureDots: ['pink', 'purple', 'cyan']
  }
]

const EXPERT_APPLY_PREVIEW_PAGES = createRoleApplyPreviewPages('expert')

const EXPERT_APPLY_WALKTHROUGH_PAGES = [
  EXPERT_APPLY_PREVIEW_PAGES[0],
  EXPERT_APPLY_PREVIEW_PAGES[1]
]

const HOME_PREVIEW_PAGES = [
  { name: '玩家首页', mode: 'roleHome', roleType: 'player' },
  { name: '行家首页', mode: 'roleHome', roleType: 'expert' },
  { name: '领路人首页', mode: 'roleHome', roleType: 'guide' },
  ROLE_COMPARISON_PREVIEW_PAGE,
  EXPERT_APPLY_PREVIEW_PAGES[0],
  EXPERT_APPLY_PREVIEW_PAGES[1]
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
    roleApplyChecking: false,
    expertApplySubmitting: false,
    expertApplyNoticeVisible: false,
    expertApplyNoticeLines: [],
    roleComparisonReturnTo: '',
    roleComparisonRoleType: 'expert',
    roleBenefitConfig: {},
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
    friendMoreRoute: ROUTES.gameHall,
    metaverseEntry: {
      title: '',
      desc: '',
      actionText: '',
      route: ''
    },
    earth: {
      nodes: [],
      heatPoints: [],
      edges: []
    },
    visualization: {
      earth: {},
      network: {}
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

      if (this.redirectGuideSingleApply(options)) {
        return
      }

      this.setData({
        roleComparisonReturnTo: decodeURIComponent(options.returnTo || ''),
        roleComparisonRoleType: normalizeComparisonRoleType(options.role || options.roleType)
      })
      this.enterHomePreview(options.mode || '', options.single === '1', options.role || options.roleType)
      return
    }

    this.loadHome()
  },

  redirectGuideSingleApply(options = {}) {
    const mode = String(options.mode || '')
    const roleType = normalizeApplyRoleType(options.role || options.roleType)
    const isApplyMode = mode === 'expertApplyOverview'
      || mode === 'expertApplyForm'
      || mode === 'roleApplyOverview'
      || mode === 'roleApplyForm'

    if (!isApplyMode || roleType !== 'guide') {
      return false
    }

    const returnTo = decodeURIComponent(options.returnTo || ROUTES.playerHome)

    navigateShellRoute(`${ROUTES.homeOther}?page=guideApplyForm&single=1&roleType=guide&returnTo=${encodeURIComponent(returnTo)}`, {
      currentRoute: ROUTES.home
    })

    return true
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

    if (!route) {
      return false
    }

    navigateShellRoute(route)

    return true
  },

  enterHomePreview(mode, single = false, roleType = '') {
    const normalizedRoleType = String(roleType || '').trim()
    const isRoleApplyMode = mode === 'expertApplyOverview'
      || mode === 'expertApplyForm'
      || mode === 'roleApplyOverview'
      || mode === 'roleApplyForm'
    const requestedMode = mode === 'roleApplyOverview'
      ? 'expertApplyOverview'
      : mode === 'roleApplyForm'
        ? 'expertApplyForm'
        : mode
    const findPreviewIndex = (pages) => pages.findIndex((page) => {
      if (!page) {
        return false
      }

      if (requestedMode === 'roleHome' && normalizedRoleType) {
        return page.mode === 'roleHome' && page.roleType === normalizedRoleType
      }

      return page.name === requestedMode || page.mode === requestedMode
    })
    const previewPages = isRoleApplyMode
      ? createRoleApplyPreviewPages(normalizeApplyRoleType(normalizedRoleType))
      : (HOME_PREVIEW_GROUPS[requestedMode] || HOME_PREVIEW_PAGES)
    const index = findPreviewIndex(previewPages)
    const lookupIndex = index >= 0
      ? index
      : (isRoleApplyMode ? -1 : findPreviewIndex(HOME_PREVIEW_LOOKUP_PAGES))
    const currentHomePreview = index >= 0
      ? previewPages[index]
      : lookupIndex >= 0
        ? HOME_PREVIEW_LOOKUP_PAGES[lookupIndex]
        : (previewPages[0] || HOME_PREVIEW_PAGES[0])
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

    if (currentHomePreview && (currentHomePreview.mode === 'expertApplyForm' || currentHomePreview.mode === 'expertApplyOverview')) {
      this.loadExpertApplyConfig()
    }
    if (currentHomePreview && currentHomePreview.mode === 'roleComparison') {
      this.loadRoleBenefitConfig()
    }
  },

  async loadRoleBenefitConfig() {
    try {
      const config = extractResponseData(await roleService.getRoleBenefitConfig())
      this.applyRoleBenefitConfig(config)
    } catch (error) {
      this.applyRoleBenefitConfig({})
    }

    this.loadRoleComparisonApplyStatus()
  },

  applyRoleBenefitConfig(config = {}) {
    const currentHomePreview = applyRoleBenefitConfigToPage(this.data.currentHomePreview, config)
    const homePreviewPages = (this.data.homePreviewPages || []).map((page) => applyRoleBenefitConfigToPage(page, config))

    this.setData({
      roleBenefitConfig: config,
      currentHomePreview,
      homePreviewPages
    })
  },

  async loadRoleComparisonApplyStatus() {
    const currentHomePreview = this.data.currentHomePreview || {}

    if (currentHomePreview.mode !== 'roleComparison') {
      return
    }

    try {
      const roleInfo = await roleService.getMyRoles()

      this.applyRoleComparisonApplyStatus(roleInfo, this.data.roleComparisonRoleType)
    } catch (error) {
      this.applyRoleComparisonApplyStatus({}, this.data.roleComparisonRoleType)
    }
  },

  applyRoleComparisonApplyStatus(roleInfo, roleType = 'expert') {
    const currentHomePreview = applyRoleComparisonApplyStatus(
      this.data.currentHomePreview,
      roleInfo,
      roleType
    )
    const homePreviewPages = (this.data.homePreviewPages || []).map((page) => (
      applyRoleComparisonApplyStatus(page, roleInfo, roleType)
    ))

    this.setData({
      currentHomePreview,
      homePreviewPages
    })
  },

  async loadExpertApplyConfig() {
    const fallbackConfig = normalizeExpertApplyConfig(DEFAULT_EXPERT_APPLY_CONFIG)
    const roleType = normalizeApplyRoleType((this.data.currentHomePreview || {}).applyRoleType)
    const fetchApplyConfig = roleType === 'guide'
      ? roleService.getGuideApplyConfig
      : roleService.getExpertApplyConfig

    try {
      const [roleInfo, remoteConfigData] = await Promise.all([
        roleService.getMyRoles(),
        fetchApplyConfig()
      ])

      if (this.redirectExistingRoleApply(roleInfo, roleType)) {
        return
      }

      const remoteConfig = extractResponseData(remoteConfigData)
      const config = normalizeExpertApplyConfig(remoteConfig)
      this.applyExpertApplyConfig(config, roleType)
    } catch (error) {
      this.applyExpertApplyConfig(fallbackConfig, roleType)
      toast.info(error.message || '网络异常，请重试')
    }
  },

  applyExpertApplyConfig(config, roleType = 'expert') {
    const applyRoleType = normalizeApplyRoleType(roleType)
    const normalizedConfig = normalizeExpertApplyConfig(config)
    const savedDraft = applyRoleType === 'expert' ? readRoleApplyDraft('expert') : null
    const shouldResetForm = isExpertApplyFormEmpty(this.data.expertApplyForm)
    const draftEnabled = Boolean(savedDraft && shouldResetForm)
    const currentFormPreview = this.data.currentHomePreview && this.data.currentHomePreview.mode === 'expertApplyForm'
      ? this.data.currentHomePreview
      : (this.data.homePreviewPages || []).find((page) => page && page.mode === 'expertApplyForm')
    const preservedDraft = !draftEnabled && !shouldResetForm && currentFormPreview
      ? { skillOptions: currentFormPreview.skillOptions || [] }
      : null
    const previewDraft = draftEnabled ? savedDraft : preservedDraft
    const currentHomePreview = applyExpertDraftToPage(
      applyExpertApplyConfigToPage(this.data.currentHomePreview, normalizedConfig),
      previewDraft
    )
    const homePreviewPages = (this.data.homePreviewPages || []).map((page) => applyExpertDraftToPage(
      applyExpertApplyConfigToPage(page, normalizedConfig),
      previewDraft
    ))
    const nextExpertApplyForm = draftEnabled
      ? normalizeExpertApplyDraftForm(savedDraft, normalizedConfig.serviceCount)
      : shouldResetForm
        ? createExpertApplyForm(normalizedConfig.serviceCount)
        : this.data.expertApplyForm

    this.setData({
      currentHomePreview,
      homePreviewPages,
      expertApplyYearOptions: normalizedConfig.yearOptions,
      expertApplyForm: nextExpertApplyForm,
      expertApplyErrors: createExpertApplyErrors(),
      expertApplyServiceErrors: createExpertApplyServiceErrors(
        (nextExpertApplyForm.services || []).length
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

  canEnterExpertApplyForm(page = this.data.currentHomePreview || {}) {
    const unmetRequirements = getUnmetExpertRequirements(page)

    if (!unmetRequirements.length) {
      return true
    }

    const firstTitle = unmetRequirements[0].title || '申请条件'
    toast.info(`请先满足：${firstTitle}`)

    return false
  },

  showExpertApplyForm() {
    if (!this.canEnterExpertApplyForm()) {
      return
    }

    const currentHomePreview = this.data.currentHomePreview || {}
    const applyRoleType = normalizeApplyRoleType(currentHomePreview.applyRoleType)

    if (applyRoleType === 'guide') {
      navigateShellRoute(`${ROUTES.homeOther}?page=guideApplyForm&single=1&roleType=guide`, {
        currentRoute: ROUTES.home
      })
      return
    }

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
    this.showExpertApplyForm()
  },

  handleExpertApplyPrimaryTap() {
    const currentHomePreview = this.data.currentHomePreview || {}

    if (currentHomePreview.mode === 'expertApplyOverview') {
      this.showExpertApplyForm()
      return
    }

    this.handleExpertApplySubmit()
  },

  handleExpertApplySaveTap() {
    const currentHomePreview = this.data.currentHomePreview || {}
    const applyRoleType = normalizeApplyRoleType(currentHomePreview.applyRoleType)

    if (applyRoleType === 'guide') {
      const savedDraft = saveRoleApplyDraft('guide', readRoleApplyDraft('guide') || { form: {} })

      toast.info(savedDraft ? '已保存到本机草稿' : '草稿保存失败')
      return
    }

    const formPreview = currentHomePreview.mode === 'expertApplyForm'
      ? currentHomePreview
      : (this.data.homePreviewPages || []).find((page) => page && page.mode === 'expertApplyForm') || {}

    const savedDraft = saveRoleApplyDraft('expert', {
      form: this.data.expertApplyForm,
      skillOptions: formPreview.skillOptions || []
    })

    toast.info(savedDraft ? '已保存到本机草稿' : '草稿保存失败')
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
    nextOptions.push({ name: '+自定义', active: false, custom: true })

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
      const roleInfo = await roleService.getMyRoles()

      if (this.redirectExistingRoleApply(roleInfo, 'expert')) {
        return
      }

      const payload = buildExpertApplyPayload(this.data.currentHomePreview, this.data.expertApplyForm)
      await roleService.submitRoleApplication(payload)
      clearRoleApplyDraft('expert')

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
    const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []

    if (
      this.data.currentHomePreview
      && this.data.currentHomePreview.mode === 'expertApplyForm'
      && previousPage
    ) {
      this.setData({
        homePreviewIndex: currentIndex - 1,
        homePreviewNo: Math.max(1, currentIndex),
        currentHomePreview: previousPage,
        expertApplyCustomInputVisible: false,
        expertApplyCustomInput: '',
        expertApplyCustomError: '',
        expertApplyYearDropdownVisible: false
      })
      return
    }

    if (pages.length > 1 && typeof wx.navigateBack === 'function') {
      wx.navigateBack()
      return
    }

    if (this.data.roleComparisonReturnTo) {
      navigateShellRoute(this.data.roleComparisonReturnTo)
      return
    }

    if (typeof wx.reLaunch === 'function') {
      wx.reLaunch({
        url: `/${ROUTES.playerHome}`
      })
    }
  },

  redirectExistingRoleApply(roleInfo, roleType) {
    const normalizedRoleType = normalizeApplyRoleType(roleType)
    const state = getExistingRoleApplyState(roleInfo, normalizedRoleType)
    const status = state.status

    if (status === 'pending') {
      navigateShellRoute(`${ROUTES.homeOther}?page=pendingCards&single=1&roleType=${normalizedRoleType}`, {
        currentRoute: ROUTES.home
      })
      return true
    }

    if (status === 'approved' && state.viewed) {
      navigateShellRoute(getRoleHomeRoute(normalizedRoleType), {
        currentRoute: ROUTES.home
      })
      return true
    }

    if (status === 'rejected' && state.viewed) {
      return false
    }

    return this.navigateRoleApplyResult(normalizedRoleType, status)
  },

  navigateRoleApplyResult(roleType, status) {
    const route = getRoleApplyResultRoute(roleType, status)

    if (!route) {
      return false
    }

    navigateShellRoute(route, {
      currentRoute: ROUTES.home
    })

    return true
  },

  async handleRoleCompareApplyTap(event) {
    const roleType = normalizeComparisonRoleType(
      event && event.detail && event.detail.roleType
        ? event.detail.roleType
        : this.data.roleComparisonRoleType
    )
    const currentHomePreview = this.data.currentHomePreview || {}

    if (currentHomePreview.primaryDisabled) {
      return
    }

    if (this.data.roleApplyChecking) {
      return
    }

    this.setData({
      roleApplyChecking: true
    })

    try {
      const roleInfo = await roleService.getMyRoles()
      const state = getExistingRoleApplyState(roleInfo, roleType)
      const status = state.status

      if (status === 'pending') {
        this.applyRoleComparisonApplyStatus(roleInfo, roleType)
        return
      }

      if ((status === 'approved' || status === 'rejected') && !state.viewed) {
        this.applyRoleComparisonApplyStatus(roleInfo, roleType)
        this.navigateRoleApplyResult(roleType, status)
        return
      }

      if (status === 'approved' && state.viewed) {
        this.applyRoleComparisonApplyStatus(roleInfo, roleType)
        return
      }

      if (roleType === 'guide') {
        const returnTo = this.data.roleComparisonReturnTo || ROUTES.playerHome

        navigateShellRoute(`${ROUTES.homeOther}?page=guideApplyForm&single=1&roleType=guide&returnTo=${encodeURIComponent(returnTo)}`, {
          currentRoute: ROUTES.home
        })
        return
      }

      this.enterHomePreview('expertApplyOverview', this.data.homePreviewSingle, roleType)
    } catch (error) {
      toast.info(error.message || '角色申请状态加载失败，请重试')
    } finally {
      this.setData({
        roleApplyChecking: false
      })
    }
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
        friendMoreRoute: (home.friendSection && (home.friendSection.route || home.friendSection.moreRoute || home.friendSection.actionRoute)) || ROUTES.gameHall,
        metaverseEntry: home.metaverseEntry,
        nearbySummary: home.nearbySummary,
        earth: home.earth,
        visualization: home.visualization
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

    navigateShellRoute(route, {
      currentRoute: ROUTES.home
    })
  }
})
