const roleService = require('../../services/role')
const { ROUTES } = require('../../config/routes')
const UI_ICONS = require('../../config/ui-icons')

const APPLY_STAGE_TOP_RPX = 108
const APPLY_DEFAULT_CONTENT_TOP_RPX = 181
const APPLY_NAV_BOTTOM_GAP_RPX = 13
const APPLY_FRAME_BOTTOM_PADDING_RPX = 10
const APPLY_DEFAULT_NAV_TOP_RPX = 108
const APPLY_DEFAULT_NAV_HEIGHT_RPX = 64
const HOME_ROUTE_MAP = {
  player: ROUTES.playerHome,
  expert: ROUTES.expertHome,
  guide: ROUTES.guideHome
}
const GUIDE_SERVICE_COUNT = 3
const GUIDE_SERVICE_NAME_MAX_LENGTH = 20
const GUIDE_UPLOAD_ACCEPT_TYPES = ['JPG', 'PNG', 'PDF']
const DEFAULT_GUIDE_AUDIENCE = ['朋友', '同事', '同城玩家']
const GUIDE_MONEY_RULE = {
  integerMaxLength: 8,
  decimalMaxLength: 2
}
const PENDING_TIMELINE_STEPS = [
  {
    title: '提交申请',
    time: {
      done: '2024.06.08 10:30',
      active: '进行中...',
      todo: '待开始'
    },
    text(meta) {
      return {
        done: meta.submittedText,
        active: `正在提交${meta.roleName}申请资料`,
        todo: `待提交${meta.roleName}申请资料`
      }
    }
  },
  {
    title: '资料初审',
    time: {
      done: '2024.06.08 11:15',
      active: '进行中...',
      todo: '待完成'
    },
    text() {
      return {
        done: '平台审核团队已完成初审',
        active: '平台审核团队已接收并开始初审',
        todo: '待平台完成资料初审'
      }
    }
  },
  {
    title: '深度审核',
    time: {
      done: '已完成',
      active: '进行中...',
      todo: '待完成'
    },
    text(meta) {
      return {
        done: '平台已完成深度审核',
        active: meta.reviewingText,
        todo: '待进入深度审核'
      }
    }
  },
  {
    title: '结果通知',
    time: {
      done: '已通知',
      active: '进行中...',
      todo: '待完成'
    },
    text() {
      return {
        done: '审核结果已通过消息通知你',
        active: '正在生成审核结果通知',
        todo: '审核结果将通过消息推送通知你'
      }
    }
  }
]
const PROGRESS_ROLE_META = {
  expert: {
    roleName: '行家',
    roleClass: 'role-expert',
    applyTitle: '行家申请',
    submittedText: '已成功提交行家申请资料',
    reviewingText: '正在评估你的专业能力、资质材料及服务说明',
    applicationNo: 'HJ20240608001'
  },
  guide: {
    roleName: '领路人',
    roleClass: 'role-guide',
    applyTitle: '领路人申请',
    submittedText: '已成功提交领路人申请资料',
    reviewingText: '正在评估你的组局记录、信用分及领路计划书',
    applicationNo: 'LR20240608001'
  }
}

const REJECTED_PAGE_META = {
  expert: {
    roleName: '行家',
    roleClass: 'role-expert',
    applicationNo: 'HJ20240608001',
    reasons: ['服务案例材料不足（需 ≥ 3 个，当前 1 个）', '专业能力说明不完整，需补充资质证明'],
    suggestions: [
      '补充更多可验证的服务案例和项目经历',
      '完善个人资料，突出专业能力与服务边界',
      '重新撰写行家申请说明，详细描述服务内容',
      '上传资质证明或过往成果可提升审核通过率'
    ],
    suggestionNote: '行家申请说明与资质证明不够完整，需补充具体服务案例和证明材料'
  },
  guide: {
    roleName: '领路人',
    roleClass: 'role-guide',
    applicationNo: 'LR20240608001',
    reasons: ['组局参与次数不足（需 ≥ 3 次，当前 2 次）', '信用分未达到要求（需 ≥ 80 分，当前 75 分）'],
    suggestions: [
      '多参与平台组局活动，积累带队经验',
      '完善个人资料，提升信用评分',
      '重新撰写领路计划书，详细描述你的服务优势',
      '获得其他领路人的推荐背书可加速审核'
    ],
    suggestionNote: '领路计划书描述过于简单，需补充具体战绩和规划说明'
  }
}

function getPositiveInteger(value, fallback) {
  const number = Number(value)

  return Number.isInteger(number) && number > 0 ? number : fallback
}

function createGuideServiceBlocks(count = GUIDE_SERVICE_COUNT) {
  const serviceCount = getPositiveInteger(count, GUIDE_SERVICE_COUNT)

  return Array.from({ length: serviceCount }, (_, index) => ({
    id: `guide-service-${index + 1}`,
    title: '业务'
  }))
}

function createGuideApplyServices(count = GUIDE_SERVICE_COUNT) {
  const serviceCount = getPositiveInteger(count, GUIDE_SERVICE_COUNT)

  return Array.from({ length: serviceCount }, () => ({
    name: '',
    price: '',
    cost: ''
  }))
}

function createGuideApplyForm() {
  return {
    audience: DEFAULT_GUIDE_AUDIENCE.slice(),
    uploadFiles: [],
    services: createGuideApplyServices()
  }
}

function normalizeUploadFile(file, index) {
  const path = file.path || file.tempFilePath || ''
  const name = file.name || path.split('/').pop() || `file-${index + 1}`

  return {
    name,
    path,
    size: file.size || 0,
    type: file.type || 'file'
  }
}

function isAllowedUploadFile(file, acceptTypes = GUIDE_UPLOAD_ACCEPT_TYPES) {
  const name = String(file.name || '').toUpperCase()

  return acceptTypes.some((type) => name.endsWith(`.${type}`))
}

function normalizeMoneyInput(value) {
  const rawValue = String(value || '').replace(/[^\d.]/g, '')
  const firstDotIndex = rawValue.indexOf('.')
  const normalizedValue = firstDotIndex >= 0
    ? rawValue.slice(0, firstDotIndex + 1) + rawValue.slice(firstDotIndex + 1).replace(/\./g, '')
    : rawValue

  return normalizedValue.replace(
    new RegExp(`^(\\d{0,${GUIDE_MONEY_RULE.integerMaxLength}})(\\.\\d{0,${GUIDE_MONEY_RULE.decimalMaxLength}})?.*$`),
    '$1$2'
  )
}

function normalizeProgressRoleType(roleType) {
  const text = String(roleType || '').trim()

  if (text === 'expert' || text === 'master' || text === '行家') {
    return 'expert'
  }

  if (text === 'guide' || text === 'leader' || text === '领路人') {
    return 'guide'
  }

  return 'guide'
}

function getProgressRoleMeta(roleType) {
  return PROGRESS_ROLE_META[normalizeProgressRoleType(roleType)] || PROGRESS_ROLE_META.guide
}

function createPendingTimeline(meta, activeStep = 'deepReview') {
  const stepKeys = ['submitted', 'preReview', 'deepReview', 'result']
  const activeStepIndex = Math.max(stepKeys.indexOf(activeStep), 0)

  return PENDING_TIMELINE_STEPS.map((step, index) => {
    const state = index < activeStepIndex
      ? 'done'
      : (index === activeStepIndex ? 'active' : 'todo')
    const textMap = step.text(meta)

    return {
      title: step.title,
      text: textMap[state],
      time: step.time[state],
      state
    }
  })
}

function createPendingDetails(meta, withStatus = false) {
  const details = [
    { label: '申请角色', value: meta.roleName, cyan: true },
    { label: '申请时间', value: '2024.06.08 10:30' },
    { label: '申请编号', value: meta.applicationNo }
  ]

  if (withStatus) {
    details.push(
      { label: '当前状态', value: '深度审核中', cyan: true },
      { label: '预计完成', value: '2024.06.12 18:00' }
    )
  }

  return details
}

function createPendingSimplePage(roleType = 'guide') {
  const meta = getProgressRoleMeta(roleType)

  return {
    id: 'pendingSimple',
    title: '审核状态',
    variant: `pending simple ${meta.roleClass}`,
    roleType: normalizeProgressRoleType(roleType),
    toolbar: true,
    toolbarSave: false,
    statusIconText: UI_ICONS.status.pending,
    statusTitle: '审核中',
    statusSubtitle: `${meta.applyTitle}正在审核`,
    description: ['平台正在评估你的申请资料，请耐心等待'],
    estimateTitle: '预计完成时间',
    estimateText: '预计 2024.06.12 18:00 前完成审核，届时将通过站内消息和短信通知你审核结果。',
    timeline: createPendingTimeline(meta),
    detailsTitle: '申请详情',
    detailsIconText: UI_ICONS.panel.record,
    details: createPendingDetails(meta),
    primaryText: '审核中，请耐心等待'
  }
}

function createPendingCardsPage(roleType = 'guide') {
  const meta = getProgressRoleMeta(roleType)

  return {
    id: 'pendingCards',
    title: '审核状态',
    variant: `pending cards ${meta.roleClass}`,
    roleType: normalizeProgressRoleType(roleType),
    toolbar: true,
    toolbarSave: false,
    statusIconText: UI_ICONS.status.pending,
    statusTitle: '审核中',
    statusSubtitle: `${meta.applyTitle}正在审核`,
    description: ['平台正在评估你的申请资料'],
    progressTextLeft: '已提交',
    progressTextRight: '预计 2024.06.12 完成',
    timeline: createPendingTimeline(meta),
    detailsTitle: '申请详情',
    detailsIconText: UI_ICONS.panel.record,
    details: createPendingDetails(meta, true),
    helperText: '审核期间你可以继续使用玩家身份',
    footerButtons: [
      { text: '返回玩家首页' },
      { text: `查看${meta.roleName}权益对比`, green: true }
    ]
  }
}

function createRejectedPage(roleType = 'guide') {
  const normalizedRoleType = normalizeProgressRoleType(roleType)
  const meta = REJECTED_PAGE_META[normalizedRoleType] || REJECTED_PAGE_META.guide

  return {
    id: `${normalizedRoleType}Rejected`,
    title: '审核结果',
    variant: `rejected ${meta.roleClass}`,
    rejected: true,
    roleType: normalizedRoleType,
    statusIconText: UI_ICONS.status.rejected,
    statusTitle: '审核未通过',
    description: ['感谢你的申请，但本次审核未通过', '查看原因并完善后可再次申请'],
    reasons: meta.reasons,
    suggestions: meta.suggestions,
    suggestionNote: meta.suggestionNote,
    retryText: '完善资料后可在 7 天后重新提交申请。建议根据驳回原因逐项改进，提高通过率。',
    detailsTitle: '申请记录',
    details: [
      { label: '申请角色', value: meta.roleName, accent: true },
      { label: '申请时间', value: '2024.06.08 10:30' },
      { label: '驳回时间', value: '2024.06.10 16:45' },
      { label: '可重新申请', value: '2024.06.17 后', cyan: true }
    ],
    footerButtons: [
      { text: '查看帮助', ghost: true },
      { text: '完善资料' }
    ]
  }
}

function applyProgressRoleToPage(page, roleType) {
  if (!page || page.id !== 'pendingSimple' && page.id !== 'pendingCards') {
    return page
  }

  const progressRoleType = roleType
    ? normalizeProgressRoleType(roleType)
    : normalizeProgressRoleType(page.roleType)

  return page.id === 'pendingSimple'
    ? createPendingSimplePage(progressRoleType)
    : createPendingCardsPage(progressRoleType)
}

function roundRpx(value) {
  return Math.round(value * 100) / 100
}

function getApplyShellLayoutStyles() {
  let contentTop = APPLY_DEFAULT_CONTENT_TOP_RPX
  let navTop = APPLY_DEFAULT_NAV_TOP_RPX
  let navHeight = APPLY_DEFAULT_NAV_HEIGHT_RPX

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

const HOME_OTHER_PAGES = [
  createRejectedPage('guide'),
  createRejectedPage('expert'),
  createPendingCardsPage('expert'),
  {
    id: 'guideApply',
    title: '申请领路人',
    variant: 'apply',
    toolbar: true,
    toolbarSave: true,
    statusIconText: UI_ICONS.status.apply,
    statusTitle: '申请成为领路人',
    statusSubtitle: '我愿意带领更多人一起玩！我申请成为领路人',
    requirements: [
      { title: '玩家等级达到 Lv.5', text: '当前等级: Lv.6 ✓ 已满足', done: true },
      { title: '完成实名认证', text: '认证状态: 已通过 ✓ 已满足', done: true },
      { title: '完成企业认证', text: '认证状态: 已通过 ✓ 已满足', done: true },
      { title: '参与过 3 次以上组局', text: '当前: 5 次 ✓ 已满足', done: true },
      { title: '已成功邀请≥ 1人完成组局', text: '当前: 2 次 ✓ 已满足', done: true },
      { title: '信用分 ≥ 80 分', text: '当前: 82 分 ✓ 已满足', done: true },
      { title: '会员等级 ≥ 基础会员', text: '当前: 基础会员 ✓ 已满足', done: true }
    ],
    planTask: { title: '提交领路计划书', text: '描述你的带队风格、战绩、资源和规划', done: false, action: '去填写 ›' },
    perks: [
      { icon: UI_ICONS.panel.revenue, text: '有权益的领路人引荐玩家组局可获得相应收入' },
      { icon: UI_ICONS.panel.featured, text: '专属领路人标识与优先推荐位' },
      { icon: UI_ICONS.panel.data, text: '数据看板：查看邀约数据与关系网络' }
    ],
    primaryText: '提交申请',
    helperText: '审核预计 1-3 个工作日'
  },
  {
    id: 'guideApplyForm',
    title: '申请领路人',
    variant: 'apply',
    applyRoleType: 'guide',
    applyRoleName: '领路人',
    toolbar: true,
    toolbarSave: true,
    statusIconText: UI_ICONS.status.apply,
    statusTitle: '申请成为领路人',
    statusSubtitle: '我愿意带领更多人一起玩！我申请成为领路人',
    formFields: [
      {
        key: 'city',
        label: '所在城市',
        type: 'input',
        required: true,
        placeholder: '请输入常驻城市',
        maxlength: 20,
        helper: '用于匹配同城玩家与组局推荐'
      },
      {
        key: 'audience',
        label: '可推荐人群',
        type: 'chips',
        required: true,
        options: [
          { name: '朋友', active: true },
          { name: '同事', active: true },
          { name: '同城玩家', active: true },
          { name: '社群成员', active: false }
        ],
        helper: '可多选，后续将用于关系网推荐'
      },
      {
        key: 'contact',
        label: '常用联系方式',
        type: 'input',
        required: true,
        placeholder: '请输入微信号或手机号',
        maxlength: 30
      },
      {
        key: 'guidePlan',
        label: '领路计划书',
        type: 'textarea',
        required: true,
        placeholder: '请描述你的带队风格、战绩、资源和规划',
        maxlength: 300,
        helper: '不少于 50 字，说明你能帮助玩家完成组局的方式'
      }
    ],
    uploadField: {
      label: '资质证明',
      required: true,
      icon: UI_ICONS.panel.upload,
      title: '点击上传作品集及凭证',
      helper: '支持 JPG、PNG、PDF，最多 5 张',
      acceptTypes: GUIDE_UPLOAD_ACCEPT_TYPES,
      maxCount: 5
    },
    serviceBlocks: createGuideServiceBlocks(),
    priceHint: '平台将收取 10% 服务费',
    primaryText: '提交领路人申请',
    helperText: '审核预计 1-3 个工作日'
  },
  {
    id: 'expertPassed',
    title: '审核结果',
    variant: 'passed role-expert',
    roleTheme: 'expert',
    statusIconText: UI_ICONS.status.approved,
    statusTitle: '恭喜审核通过！',
    statusSubtitle: '你已成为「行家」',
    targetRole: 'expert',
    description: ['你的申请已通过平台审核', '现在可以开始创建新局、交付服务'],
    certNo: 'ZHW-00126-2026',
    certTime: '2026.06.10 14:30',
    giftTitle: '新手礼包',
    gifts: [
      { icon: UI_ICONS.panel.giftLimit, text: '每月添加行家30位', tag: '限时', type: 'limit' },
      { icon: UI_ICONS.panel.traffic, text: '首页推荐 7 天', tag: '流量', type: 'traffic' },
      { icon: UI_ICONS.panel.reward, text: '赠送300经验值', tag: '奖励', type: 'reward' }
    ],
    actions: [
      { iconText: UI_ICONS.action.network, text: '关系网开启', route: ROUTES.relationNetwork },
      { iconText: UI_ICONS.action.invite, text: '邀请玩家', route: ROUTES.gameInvite },
      { iconText: UI_ICONS.action.profile, text: '完善资料', route: ROUTES.profileSystemProfileInfo }
    ],
    primaryText: '开启行家之旅'
  },
  {
    id: 'guidePassed',
    title: '审核结果',
    variant: 'passed role-guide',
    roleTheme: 'guide',
    statusIconText: UI_ICONS.status.approved,
    statusTitle: '恭喜审核通过！',
    statusSubtitle: '你已成为「领路人」',
    targetRole: 'guide',
    description: ['你的申请已通过平台审核', '现在可以开始邀约玩家进入组局'],
    certNo: 'ZHW-00115-2026',
    certTime: '2026.06.10 14:30',
    giftTitle: '新手礼包',
    gifts: [
      { icon: UI_ICONS.panel.giftLimit, text: '每月添加行家15位', tag: '限时', type: 'limit' },
      { icon: UI_ICONS.panel.traffic, text: '首页推荐 7 天', tag: '流量', type: 'traffic' },
      { icon: UI_ICONS.panel.reward, text: '赠送100经验值', tag: '奖励', type: 'reward' }
    ],
    actions: [
      { iconText: UI_ICONS.action.network, text: '关系网开启', route: ROUTES.relationNetwork },
      { iconText: UI_ICONS.action.invite, text: '邀请玩家', route: ROUTES.gameInvite },
      { iconText: UI_ICONS.action.profile, text: '完善资料', route: ROUTES.profileSystemProfileInfo }
    ],
    primaryText: '开启领路人之旅'
  },
  createPendingCardsPage('guide'),
  createPendingSimplePage('guide'),
]

Page({
  data: {
    pages: HOME_OTHER_PAGES,
    currentIndex: 0,
    currentPage: applyProgressRoleToPage(HOME_OTHER_PAGES[0], 'expert'),
    pageNo: 1,
    pageTotal: HOME_OTHER_PAGES.length,
    previewSingle: false,
    progressRoleType: '',
    previewWindowWidth: 375,
    primaryNavigating: false,
    guideApplyForm: createGuideApplyForm(),
    guideServiceNameMaxLength: GUIDE_SERVICE_NAME_MAX_LENGTH,
    applyShellLayout: getApplyShellLayoutStyles(),
    uiIcons: UI_ICONS
  },

  onLoad(options = {}) {
    const previewWindowWidth = wx.getSystemInfoSync ? wx.getSystemInfoSync().windowWidth : 375
    const optionRoleType = options.roleType || options.applyRoleType || ''
    const progressRoleType = optionRoleType ? normalizeProgressRoleType(optionRoleType) : ''
    const requestedPageId = options.page || options.id || ''
    const previewSingle = requestedPageId ? options.single !== '0' : options.single === '1'
    const requestedPageIndex = requestedPageId
      ? HOME_OTHER_PAGES.findIndex((page) => page.id === requestedPageId)
      : -1
    const optionIndex = Number(options.index || 0)
    const currentIndex = requestedPageIndex >= 0
      ? requestedPageIndex
      : (optionIndex >= 0 && optionIndex < HOME_OTHER_PAGES.length ? optionIndex : 0)
    const currentPage = applyProgressRoleToPage(HOME_OTHER_PAGES[currentIndex], progressRoleType)

    this.setData({
      previewWindowWidth,
      progressRoleType,
      previewSingle,
      currentIndex,
      currentPage,
      pageNo: previewSingle ? 1 : currentIndex + 1,
      pageTotal: previewSingle ? 1 : HOME_OTHER_PAGES.length,
      applyShellLayout: getApplyShellLayoutStyles()
    })
  },

  handlePreviewTap(event) {
    if (this.data.previewSingle) {
      return
    }

    const datasetDirection = Number(event.currentTarget.dataset.direction)
    const touch = event.changedTouches && event.changedTouches[0]
    const x = touch ? touch.clientX : event.detail.x
    const direction = datasetDirection || (x < this.data.previewWindowWidth / 2 ? -1 : 1)
    const nextIndex = this.data.currentIndex + direction

    if (nextIndex < 0 || nextIndex >= HOME_OTHER_PAGES.length) {
      return
    }

    this.setData({
      currentIndex: nextIndex,
      currentPage: applyProgressRoleToPage(HOME_OTHER_PAGES[nextIndex], this.data.progressRoleType),
      pageNo: nextIndex + 1
    })
  },

  handleUnavailableTap() {
    wx.showToast({
      title: '功能开发中',
      icon: 'none'
    })
  },

  handlePassedActionTap(event) {
    const route = event.currentTarget.dataset.route

    if (!route) {
      this.handleUnavailableTap()
      return
    }

    wx.navigateTo({
      url: route.indexOf('/') === 0 ? route : `/${route}`
    })
  },

  handleGuideApplyBackTap() {
    if (this.data.currentPage && this.data.currentPage.id === 'guideApplyForm') {
      if (this.data.previewSingle) {
        const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []

        if (pages.length > 1 && typeof wx.navigateBack === 'function') {
          wx.navigateBack()
          return
        }

        wx.redirectTo({
          url: `/${ROUTES.homeOther}?page=guideApply`
        })
        return
      }

      const previousIndex = HOME_OTHER_PAGES.findIndex((page) => page.id === 'guideApply')

      if (previousIndex >= 0) {
        this.setData({
          currentIndex: previousIndex,
          currentPage: applyProgressRoleToPage(HOME_OTHER_PAGES[previousIndex], this.data.progressRoleType),
          pageNo: this.data.previewSingle ? 1 : previousIndex + 1
        })
        return
      }
    }

    if (this.data.currentPage && (this.data.currentPage.id === 'pendingSimple' || this.data.currentPage.id === 'pendingCards')) {
      wx.reLaunch({
        url: `/${ROUTES.playerHome}`
      })
      return
    }

    const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []

    if (pages.length > 1 && typeof wx.navigateBack === 'function') {
      wx.navigateBack()
      return
    }

    wx.reLaunch({
      url: `/${ROUTES.playerHome}`
    })
  },

  handleGuideApplySaveTap() {
    this.handleUnavailableTap()
  },

  handleGuidePlanTap() {
    if (this.data.previewSingle) {
      wx.navigateTo({
        url: `/${ROUTES.homeOther}?page=guideApplyForm`
      })
      return
    }

    const nextIndex = HOME_OTHER_PAGES.findIndex((page) => page.id === 'guideApplyForm')

    if (nextIndex < 0) {
      this.handleUnavailableTap()
      return
    }

    this.setData({
      currentIndex: nextIndex,
      currentPage: applyProgressRoleToPage(HOME_OTHER_PAGES[nextIndex], this.data.progressRoleType),
      pageNo: this.data.previewSingle ? 1 : nextIndex + 1
    })
  },

  onGuideApplyServiceNameInput(event) {
    const index = Number(event.currentTarget.dataset.index)

    if (!Number.isInteger(index) || index < 0) {
      return
    }

    this.setData({
      [`guideApplyForm.services[${index}].name`]: event.detail.value || ''
    })
  },

  onGuideApplyMoneyInput(event) {
    const index = Number(event.currentTarget.dataset.index)
    const field = event.currentTarget.dataset.field

    if (!Number.isInteger(index) || index < 0 || !field) {
      return
    }

    this.setData({
      [`guideApplyForm.services[${index}].${field}`]: normalizeMoneyInput(event.detail.value)
    })
  },

  onGuideApplyChipTap(event) {
    const fieldIndex = Number(event.currentTarget.dataset.fieldIndex)
    const optionIndex = Number(event.currentTarget.dataset.optionIndex)
    const formFields = (this.data.currentPage && this.data.currentPage.formFields) || []
    const field = formFields[fieldIndex]

    if (!field || field.type !== 'chips' || !Array.isArray(field.options)) {
      return
    }

    const option = field.options[optionIndex]

    if (!option) {
      return
    }

    const nextOptions = field.options.map((item, index) => (
      index === optionIndex
        ? Object.assign({}, item, { active: !item.active })
        : item
    ))
    const selectedValues = nextOptions
      .filter((item) => item.active)
      .map((item) => item.name)

    this.setData({
      [`currentPage.formFields[${fieldIndex}].options`]: nextOptions,
      [`guideApplyForm.${field.key}`]: selectedValues
    })
  },

  onGuideApplyUploadTap() {
    const uploadField = (this.data.currentPage || {}).uploadField || {}
    const maxCount = getPositiveInteger(uploadField.maxCount, 5)
    const acceptTypes = Array.isArray(uploadField.acceptTypes) ? uploadField.acceptTypes : GUIDE_UPLOAD_ACCEPT_TYPES
    const acceptText = acceptTypes.join('、')
    const currentFiles = this.data.guideApplyForm.uploadFiles || []
    const remainingCount = maxCount - currentFiles.length

    if (remainingCount <= 0) {
      wx.showToast({
        title: `最多上传 ${maxCount} 个文件`,
        icon: 'none'
      })
      return
    }

    if (typeof wx === 'undefined' || typeof wx.chooseMessageFile !== 'function') {
      wx.showToast({
        title: '当前环境不支持文件选择',
        icon: 'none'
      })
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
          'guideApplyForm.uploadFiles': files
        })

        if (rejectedFiles.length) {
          wx.showToast({
            title: `仅支持 ${acceptText} 格式`,
            icon: 'none'
          })
        }
      },
      fail: (error) => {
        if (error && String(error.errMsg || '').includes('cancel')) {
          return
        }

        wx.showToast({
          title: '文件选择失败，请重试',
          icon: 'none'
        })
      }
    })
  },

  onGuideApplyRemoveFileTap(event) {
    const index = Number(event.currentTarget.dataset.index)
    const files = (this.data.guideApplyForm.uploadFiles || []).filter((file, fileIndex) => fileIndex !== index)

    this.setData({
      'guideApplyForm.uploadFiles': files
    })
  },

  showGuideApplySubmitSuccess(roleName = '领路人') {
    const goPlayerHome = () => {
      wx.reLaunch({
        url: `/${ROUTES.playerHome}`
      })
    }

    wx.showToast({
      title: `${roleName}申请已提交，请耐心等待审核结果`,
      icon: 'none'
    })

    setTimeout(goPlayerHome, 1200)
  },

  async handleGuideApplySubmit() {
    if (this.data.primaryNavigating) {
      return
    }

    const currentPage = this.data.currentPage || {}
    const applyRoleType = currentPage.applyRoleType || 'guide'
    const applyRoleName = currentPage.applyRoleName || '领路人'

    this.setData({
      primaryNavigating: true
    })

    try {
      await roleService.submitRoleApplication({
        roleType: applyRoleType,
        source: 'home-other',
        form: this.data.guideApplyForm
      })
      this.showGuideApplySubmitSuccess(applyRoleName)
    } catch (error) {
      this.setData({
        primaryNavigating: false
      })
      wx.showToast({
        title: '提交失败，请重试',
        icon: 'none'
      })
    }
  },

  normalizeRoleType(roleType) {
    const text = String(roleType || '').trim()

    if (text === '行家' || text === 'expert') {
      return 'expert'
    }

    if (text === '领路人' || text === 'leader' || text === 'guide') {
      return 'guide'
    }

    if (text === '玩家' || text === 'player') {
      return 'player'
    }

    return ''
  },

  normalizeRoleStatus(status) {
    const text = String(status || '').trim()

    if (text === 'approved' || text === 'pass' || text === 'passed' || text === 'active' || text === 'enabled' || text === '已通过') {
      return 'approved'
    }

    if (text === 'pending' || text === 'reviewing' || text === '审核中') {
      return 'pending'
    }

    if (text === 'rejected' || text === 'reject' || text === 'failed' || text === '未通过' || text === '已驳回') {
      return 'rejected'
    }

    return text
  },

  isApprovedRole(roleInfo, roleType) {
    return this.getRoleStatus(roleInfo, roleType) === 'approved'
  },

  getRoleStatus(roleInfo, roleType) {
    const normalizedRoleType = this.normalizeRoleType(roleType)
    const roleStatusMap = (roleInfo && (roleInfo.roleStatusMap || roleInfo.role_status_map)) || {}
    const mappedStatus = this.normalizeRoleStatus(roleStatusMap[normalizedRoleType])

    if (mappedStatus === 'approved') {
      return 'approved'
    }

    if (mappedStatus) {
      return mappedStatus
    }

    const roles = (roleInfo && (roleInfo.roles || roleInfo.roleList || roleInfo.role_list)) || []
    const matchedRole = Array.isArray(roles)
      ? roles.find((item) => {
        if (typeof item === 'string') {
          return this.normalizeRoleType(item) === normalizedRoleType
        }

        return this.normalizeRoleType(item.roleType || item.role_type || item.key || item.name || item.type) === normalizedRoleType
      })
      : null

    if (typeof matchedRole === 'string') {
      return 'approved'
    }

    if (matchedRole) {
      const itemStatus = this.normalizeRoleStatus(item.status || item.roleStatus || item.role_status)

      return itemStatus || 'approved'
    }

    return ''
  },

  resolveHomeRole(roleInfo, fallbackRole) {
    const preferredRole = this.normalizeRoleType(
      roleInfo && (roleInfo.defaultRole || roleInfo.default_role || roleInfo.currentRole || roleInfo.current_role)
    )

    const targetRole = this.normalizeRoleType(fallbackRole)

    if (targetRole && this.getRoleStatus(roleInfo, targetRole) !== 'pending' && this.getRoleStatus(roleInfo, targetRole) !== 'rejected') {
      return targetRole
    }

    if (preferredRole && this.isApprovedRole(roleInfo, preferredRole)) {
      return preferredRole
    }

    return targetRole || preferredRole || 'player'
  },

  goRoleHome(roleType) {
    const route = HOME_ROUTE_MAP[this.normalizeRoleType(roleType)] || ROUTES.playerHome

    wx.reLaunch({
      url: `/${route}`
    })
  },

  async handlePrimaryTap() {
    const currentPage = this.data.currentPage || {}

    if (currentPage.id === 'guideApplyForm') {
      this.handleGuideApplySubmit()
      return
    }

    if (!currentPage.targetRole) {
      this.handleUnavailableTap()
      return
    }

    if (this.data.primaryNavigating) {
      return
    }

    this.setData({
      primaryNavigating: true
    })

    try {
      const roleInfo = await roleService.getMyRoles()
      const homeRole = this.resolveHomeRole(roleInfo, currentPage.targetRole)
      this.goRoleHome(homeRole)
    } catch (error) {
      this.goRoleHome(currentPage.targetRole)
    } finally {
      this.setData({
        primaryNavigating: false
      })
    }
  }
})
