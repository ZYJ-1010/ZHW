const roleService = require('../../../services/role')
const toast = require('../../../utils/toast')
const { ROUTES } = require('../../../config/routes')
const UI_ICONS = require('../../../config/ui-icons')

const ROLE_TYPE_MAP = {
  player: 'player',
  expert: 'expert',
  master: 'expert',
  guide: 'guide',
  leader: 'guide',
  玩家: 'player',
  行家: 'expert',
  领路人: 'guide'
}

const ROLE_META = {
  expert: {
    roleName: '行家',
    applyTitle: '行家申请',
    certNo: 'ZHW-00126-2026',
    successAccent: 'cyan',
    approvedCopy: '现在可以开始创建新局、交付服务',
    primaryText: '开启行家之旅',
    rewards: [
      { icon: UI_ICONS.panel.giftLimit, name: '每月添加行家30位', tag: '限时', tone: 'yellow' },
      { icon: UI_ICONS.panel.traffic, name: '首页推荐 7 天', tag: '流量', tone: 'blue' },
      { icon: UI_ICONS.panel.reward, name: '赠送300经验值', tag: '奖励', tone: 'green' }
    ]
  },
  guide: {
    roleName: '领路人',
    applyTitle: '领路人申请',
    certNo: 'ZHW-00115-2026',
    successAccent: 'orange',
    approvedCopy: '现在可以开始邀约玩家进入组局',
    primaryText: '开启领路人之旅',
    rewards: [
      { icon: UI_ICONS.panel.giftLimit, name: '每月添加行家15位', tag: '限时', tone: 'yellow' },
      { icon: UI_ICONS.panel.traffic, name: '首页推荐 7 天', tag: '流量', tone: 'blue' },
      { icon: UI_ICONS.panel.reward, name: '赠送100经验值', tag: '奖励', tone: 'green' }
    ]
  }
}

const APPROVED_ACTIONS = [
  { icon: UI_ICONS.action.network, text: '关系网开启', route: ROUTES.relationNetwork },
  { icon: UI_ICONS.action.invite, text: '邀请玩家', route: ROUTES.gameInvite },
  { icon: UI_ICONS.action.profile, text: '完善资料', route: ROUTES.profileSystemProfileInfo }
]

const STATUS_MAP = {
  active: 'approved',
  enabled: 'approved',
  passed: 'approved',
  success: 'approved',
  waiting: 'pending',
  reviewing: 'pending',
  auditing: 'pending',
  pending_audit: 'pending',
  rejected_audit: 'rejected',
  reject: 'rejected',
  disabled: 'disabled',
  available: 'none',
  locked: 'none',
  unavailable: 'none'
}

function normalizeRoleType(roleType) {
  return ROLE_TYPE_MAP[roleType] || 'guide'
}

function normalizeStatus(status) {
  if (status == null || status === '') {
    return 'pending'
  }

  const value = String(status).trim()

  return STATUS_MAP[value] || value
}

function pickFirstValue(...values) {
  return values.find((value) => value != null && value !== '')
}

function formatDateTime(value) {
  if (!value || typeof value !== 'string') {
    return ''
  }

  return value
    .replace(/-/g, '.')
    .replace('T', ' ')
    .replace(/\+.*$/, '')
    .replace(/Z$/, '')
    .replace(/(\d{2}:\d{2}):\d{2}(?:\.\d+)?$/, '$1')
    .trim()
}

function addDays(dateText, days) {
  if (!dateText) {
    return ''
  }

  const normalized = dateText.replace(/\./g, '-').replace(' ', 'T')
  const date = new Date(normalized)

  if (Number.isNaN(date.getTime())) {
    return ''
  }

  date.setDate(date.getDate() + days)

  const year = date.getFullYear()
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  const day = `${date.getDate()}`.padStart(2, '0')

  return `${year}.${month}.${day}`
}

function createFallbackApplication(roleType, status) {
  return {
    roleType,
    status
  }
}

function normalizeApplication(rawApplication = {}, roleType, status) {
  const normalizedRole = normalizeRoleType(
    rawApplication.roleType ||
    rawApplication.role_type ||
    rawApplication.type ||
    rawApplication.key ||
    rawApplication.name ||
    roleType
  )
  const normalizedStatus = normalizeStatus(
    rawApplication.status ||
    rawApplication.applyStatus ||
    rawApplication.applicationStatus ||
    rawApplication.roleStatus ||
    rawApplication.role_status ||
    status
  )
  const submittedAt = formatDateTime(pickFirstValue(
    rawApplication.submittedAt,
    rawApplication.createdAt,
    rawApplication.applyTime,
    rawApplication.created_at
  ))
  const reviewedAt = formatDateTime(pickFirstValue(
    rawApplication.reviewedAt,
    rawApplication.reviewTime,
    rawApplication.reviewed_at
  ))
  const expectedReviewAt = formatDateTime(pickFirstValue(
    rawApplication.expectedReviewAt,
    rawApplication.expectedReviewedAt,
    rawApplication.estimatedReviewAt,
    rawApplication.estimatedReviewedAt,
    rawApplication.expectedReviewTime
  ))

  return {
    roleType: normalizedRole,
    status: normalizedStatus,
    statusText: rawApplication.statusText || rawApplication.statusLabel || rawApplication.applyStatusText || '',
    submittedAt,
    reviewedAt,
    expectedReviewAt,
    applicationId: pickFirstValue(rawApplication.applicationId, rawApplication.id, rawApplication.application_id, ''),
    rejectReason: pickFirstValue(rawApplication.rejectReason, rawApplication.reject_reason, ''),
    rejectReasons: Array.isArray(rawApplication.rejectReasons) ? rawApplication.rejectReasons : [],
    canReapplyAt: formatDateTime(pickFirstValue(rawApplication.canReapplyAt, rawApplication.reapplyAt, ''))
  }
}

function findApplication(applications, roleType, status) {
  const normalizedRole = normalizeRoleType(roleType)
  const normalizedStatus = normalizeStatus(status)
  const normalizedList = Array.isArray(applications)
    ? applications.map((item) => normalizeApplication(item, normalizedRole, normalizedStatus))
    : []
  const sameRole = normalizedList.filter((item) => item.roleType === normalizedRole)

  if (!sameRole.length) {
    return normalizeApplication(createFallbackApplication(normalizedRole, normalizedStatus), normalizedRole, normalizedStatus)
  }

  const matchedApplication = sameRole.find((item) => item.status === normalizedStatus)

  if (matchedApplication) {
    return matchedApplication
  }

  if (normalizedStatus && normalizedStatus !== 'none') {
    return Object.assign({}, sameRole[0], {
      status: normalizedStatus
    })
  }

  return sameRole[0]
}

function buildPendingTimeline(application, roleName) {
  return [
    {
      title: '提交申请',
      desc: `已成功提交${roleName}申请资料`,
      time: application.submittedAt || '已提交',
      state: 'done'
    },
    {
      title: '资料初审',
      desc: '平台审核团队已接收并开始初审',
      time: application.submittedAt ? '已接收' : '待系统同步',
      state: 'done'
    },
    {
      title: '深度审核',
      desc: roleName === '领路人'
        ? '正在评估你的组局记录、信用分及领路计划书'
        : '正在评估你的专业能力、资质材料及服务说明',
      time: '进行中...',
      state: 'active'
    },
    {
      title: '结果通知',
      desc: '审核结果将通过消息推送通知你',
      time: '待完成',
      state: 'pending'
    }
  ]
}

function buildDetails(application, meta) {
  return [
    { label: '申请角色', value: meta.roleName, highlight: true },
    { label: '申请时间', value: application.submittedAt || '以后台记录为准' },
    { label: '申请编号', value: application.applicationId || '审核中生成' },
    { label: '当前状态', value: application.status === 'pending' ? '深度审核中' : application.statusText || '待确认', highlight: true },
    { label: '预计完成', value: application.expectedReviewAt || '预计 1-3 个工作日' }
  ]
}

function buildRejectReasons(application) {
  if (application.rejectReasons.length) {
    return application.rejectReasons
  }

  if (application.rejectReason) {
    return [application.rejectReason]
  }

  return [
    '申请资料暂未达到当前角色审核要求',
    '部分证明材料或计划说明仍需补充完善'
  ]
}

function buildPageState(application) {
  const meta = ROLE_META[application.roleType] || ROLE_META.guide
  const status = normalizeStatus(application.status)
  const isApproved = status === 'approved'
  const isRejected = status === 'rejected'
  const isPending = !isApproved && !isRejected
  const reviewedAt = application.reviewedAt || ''
  const canReapplyAt = application.canReapplyAt || addDays(reviewedAt || application.submittedAt, 7)

  return {
    loading: false,
    roleType: application.roleType,
    status,
    uiIcons: UI_ICONS,
    meta,
    isPending,
    isApproved,
    isRejected,
    pageTitle: isPending ? '审核进度' : '审核结果',
    statusTitle: isPending ? '审核中' : (isApproved ? '恭喜审核通过！' : '审核未通过'),
    statusSubTitle: isPending
      ? `${meta.roleName}申请正在审核`
      : (isApproved ? `你已成为「${meta.roleName}」` : '查看原因并完善后可再次申请'),
    statusDesc: isPending
      ? '平台正在评估你的申请资料，请耐心等待'
      : (isApproved ? meta.approvedCopy : '感谢你的申请，但本次审核未通过'),
    expectedText: application.expectedReviewAt
      ? `预计 ${application.expectedReviewAt} 前完成审核，届时将通过站内消息通知你审核结果。`
      : '审核预计 1-3 个工作日，结果将通过站内消息通知你。',
    timeline: buildPendingTimeline(application, meta.roleName),
    details: buildDetails(application, meta),
    rejectReasons: buildRejectReasons(application),
    suggestions: [
      '多参与平台组局活动，积累带队经验',
      '完善个人资料，提升信用评分',
      `${meta.roleName === '领路人' ? '重新撰写领路计划书' : '补充服务说明'}，详细描述你的服务优势`,
      '获得同伴推荐背书可提升审核通过率'
    ],
    history: [
      { label: '申请角色', value: meta.roleName, highlight: true },
      { label: '申请时间', value: application.submittedAt || '以后台记录为准' },
      { label: '驳回时间', value: reviewedAt || '以后台记录为准' },
      { label: '可重新申请', value: canReapplyAt ? `${canReapplyAt} 后` : '请关注后台通知', highlight: true }
    ],
    certNo: application.applicationId || meta.certNo,
    certTime: reviewedAt || '以后台记录为准',
    rewards: meta.rewards,
    nextActions: APPROVED_ACTIONS
  }
}

Page({
  data: buildPageState(normalizeApplication(createFallbackApplication('guide', 'pending'))),

  onLoad(options = {}) {
    const roleType = normalizeRoleType(options.roleType || 'guide')
    const status = normalizeStatus(options.status || 'pending')

    this.setData({
      loading: true,
      roleType,
      status
    })
    this.loadRoleStatus(roleType, status)
  },

  async loadRoleStatus(roleType, status) {
    try {
      const applications = await roleService.getMyRoleApplications()
      const application = findApplication(applications, roleType, status)

      this.setData(buildPageState(application))
    } catch (error) {
      const fallback = normalizeApplication(createFallbackApplication(roleType, status), roleType, status)

      this.setData(buildPageState(fallback))
      toast.info(error.message || '审核状态加载失败，已展示本地状态')
    }
  },

  handleBackTap() {
    const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []

    if (pages.length > 1 && typeof wx.navigateBack === 'function') {
      wx.navigateBack()
      return
    }

    wx.reLaunch({
      url: `/${ROUTES.playerHome}`
    })
  },

  handleBackHomeTap() {
    const roleHomeRoutes = {
      expert: ROUTES.expertHome,
      guide: ROUTES.guideHome,
      player: ROUTES.playerHome
    }
    const route = this.data.isApproved
      ? (roleHomeRoutes[this.data.roleType] || ROUTES.playerHome)
      : ROUTES.playerHome

    wx.reLaunch({
      url: `/${route}`
    })
  },

  handleApprovedActionTap(event) {
    const route = event.currentTarget.dataset.route

    if (!route) {
      toast.info('功能开发中')
      return
    }

    wx.navigateTo({
      url: route.indexOf('/') === 0 ? route : `/${route}`
    })
  },

  handleBenefitsTap() {
    wx.navigateTo({
      url: `/${ROUTES.home}?ui=1&mode=roleComparison&single=1&returnTo=${encodeURIComponent(ROUTES.roleStatus)}`
    })
  },

  handleHelpTap() {
    toast.info('审核帮助正在完善中')
  },

  handleImproveTap() {
    wx.navigateTo({
      url: `/${ROUTES.profile}`
    })
  }
})
