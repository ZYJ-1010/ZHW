const roleService = require('../../../services/role')
const toast = require('../../../utils/toast')
const { ROUTES } = require('../../../config/routes')
const UI_ICONS = require('../../../config/ui-icons')
const { navigateShellRoute } = require('../../../utils/shell-nav')

const EMPTY_STATUS_CONFIG = {
  roleAliases: {},
  statusMap: {},
  roleMeta: {},
  pendingTimeline: [],
  approvedActions: [],
  texts: {},
  defaultRejectReasons: [],
  suggestionTemplates: [],
  improvePlanTextByRole: {},
  reapplyDays: 7
}

function normalizeStatusConfig(config = {}) {
  return {
    roleAliases: config.roleAliases || {},
    statusMap: config.statusMap || {},
    roleMeta: config.roleMeta || {},
    pendingTimeline: Array.isArray(config.pendingTimeline) ? config.pendingTimeline : [],
    approvedActions: Array.isArray(config.approvedActions) ? config.approvedActions : [],
    texts: config.texts || {},
    defaultRejectReasons: Array.isArray(config.defaultRejectReasons) ? config.defaultRejectReasons : [],
    suggestionTemplates: Array.isArray(config.suggestionTemplates) ? config.suggestionTemplates : [],
    improvePlanTextByRole: config.improvePlanTextByRole || {},
    reapplyDays: Number(config.reapplyDays || 7)
  }
}

function textOf(config, key) {
  const texts = config && config.texts ? config.texts : {}
  return texts[key] || ''
}

function applyTemplate(template, values = {}) {
  return String(template || '').replace(/\{(\w+)\}/g, (_, key) => values[key] == null ? '' : values[key])
}

function normalizeRoleType(roleType, config = EMPTY_STATUS_CONFIG) {
  return config.roleAliases[roleType] || 'guide'
}

function normalizeStatus(status, config = EMPTY_STATUS_CONFIG) {
  if (status == null || status === '') {
    return 'pending'
  }

  const value = String(status).trim()

  return config.statusMap[value] || value
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

function normalizeApplication(rawApplication = {}, roleType, status, config = EMPTY_STATUS_CONFIG) {
  const normalizedRole = normalizeRoleType(
    rawApplication.roleType ||
    rawApplication.role_type ||
    rawApplication.type ||
    rawApplication.key ||
    rawApplication.name ||
    roleType,
    config
  )
  const normalizedStatus = normalizeStatus(
    rawApplication.status ||
    rawApplication.applyStatus ||
    rawApplication.applicationStatus ||
    rawApplication.roleStatus ||
    rawApplication.role_status ||
    status,
    config
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
    ,
    approvedCopy: pickFirstValue(rawApplication.approvedCopy, rawApplication.successCopy, ''),
    primaryText: pickFirstValue(rawApplication.primaryText, rawApplication.primaryActionText, ''),
    certNo: pickFirstValue(rawApplication.certNo, rawApplication.certNumber, rawApplication.certificateNo, ''),
    rewards: Array.isArray(rawApplication.rewards) ? rawApplication.rewards : []
  }
}

function findApplication(applications, roleType, status, config = EMPTY_STATUS_CONFIG) {
  const normalizedRole = normalizeRoleType(roleType, config)
  const normalizedStatus = normalizeStatus(status, config)
  const normalizedList = Array.isArray(applications)
    ? applications.map((item) => normalizeApplication(item, normalizedRole, normalizedStatus, config))
    : []
  const sameRole = normalizedList.filter((item) => item.roleType === normalizedRole)

  if (!sameRole.length) {
    return normalizeApplication(createFallbackApplication(normalizedRole, normalizedStatus), normalizedRole, normalizedStatus, config)
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

function buildPendingTimeline(application, roleName, config) {
  return config.pendingTimeline.map((item) => {
    const descByRole = item.descByRole || {}

    return {
      title: item.title || '',
      desc: applyTemplate(descByRole[application.roleType] || item.descTemplate || item.desc || '', { roleName }),
      time: item.timeField === 'submittedAt'
        ? (application.submittedAt || item.fallbackTime || '')
        : (item.timeWhenSubmitted ? (application.submittedAt ? item.timeWhenSubmitted : item.fallbackTime || '') : item.time || ''),
      state: item.state || ''
    }
  })
}

function buildDetails(application, meta, config) {
  return [
    { label: textOf(config, 'fieldRoleLabel'), value: meta.roleName, highlight: true },
    { label: textOf(config, 'fieldApplyTimeLabel'), value: application.submittedAt || textOf(config, 'backendRecordFallback') },
    { label: textOf(config, 'fieldApplicationNoLabel'), value: application.applicationId || textOf(config, 'applicationNoFallback') },
    { label: textOf(config, 'fieldCurrentStatusLabel'), value: application.status === 'pending' ? textOf(config, 'pendingStatusText') : application.statusText || textOf(config, 'statusFallback'), highlight: true },
    { label: textOf(config, 'fieldExpectedLabel'), value: application.expectedReviewAt || textOf(config, 'expectedDoneFallback') }
  ]
}

function buildRejectReasons(application, config) {
  if (application.rejectReasons.length) {
    return application.rejectReasons
  }

  if (application.rejectReason) {
    return [application.rejectReason]
  }

  return config.defaultRejectReasons
}

function buildSuggestions(application, meta, config) {
  const improvePlanText = config.improvePlanTextByRole[application.roleType] || ''

  return config.suggestionTemplates.map((item) => applyTemplate(item, {
    roleName: meta.roleName,
    improvePlanText
  }))
}

function buildApprovedActions(config) {
  return config.approvedActions.map((item) => ({
    icon: UI_ICONS.action[item.iconKey] || '',
    text: item.text || '',
    route: ROUTES[item.routeKey] || ''
  }))
}

function buildPageState(application, rawConfig = EMPTY_STATUS_CONFIG) {
  const config = normalizeStatusConfig(rawConfig)
  const meta = config.roleMeta[application.roleType] || config.roleMeta.guide || {}
  const status = normalizeStatus(application.status, config)
  const isApproved = status === 'approved'
  const isRejected = status === 'rejected'
  const isPending = !isApproved && !isRejected
  const reviewedAt = application.reviewedAt || ''
  const canReapplyAt = application.canReapplyAt || addDays(reviewedAt || application.submittedAt, config.reapplyDays)

  return {
    loading: false,
    loadError: '',
    roleType: application.roleType,
    status,
    uiIcons: UI_ICONS,
    pageConfig: config,
    texts: config.texts,
    meta,
    isPending,
    isApproved,
    isRejected,
    pageTitle: isPending ? textOf(config, 'pendingPageTitle') : textOf(config, 'resultPageTitle'),
    statusTitle: isPending ? textOf(config, 'pendingTitle') : (isApproved ? textOf(config, 'approvedTitle') : textOf(config, 'rejectedTitle')),
    statusSubTitle: isPending
      ? applyTemplate(textOf(config, 'pendingSubtitleTemplate'), { roleName: meta.roleName })
      : (isApproved ? applyTemplate(textOf(config, 'approvedSubtitleTemplate'), { roleName: meta.roleName }) : textOf(config, 'rejectedSubtitle')),
    statusDesc: isPending
      ? textOf(config, 'pendingDesc')
      : (isApproved ? (application.approvedCopy || meta.approvedCopy || '') : textOf(config, 'rejectedDesc')),
    expectedText: application.expectedReviewAt
      ? applyTemplate(textOf(config, 'expectedTemplate'), { expectedReviewAt: application.expectedReviewAt })
      : textOf(config, 'expectedFallback'),
    timeline: buildPendingTimeline(application, meta.roleName, config),
    details: buildDetails(application, meta, config),
    rejectReasons: buildRejectReasons(application, config),
    suggestions: buildSuggestions(application, meta, config),
    history: [
      { label: textOf(config, 'fieldRoleLabel'), value: meta.roleName, highlight: true },
      { label: textOf(config, 'fieldApplyTimeLabel'), value: application.submittedAt || textOf(config, 'backendRecordFallback') },
      { label: textOf(config, 'fieldRejectTimeLabel'), value: reviewedAt || textOf(config, 'backendRecordFallback') },
      { label: textOf(config, 'fieldReapplyLabel'), value: canReapplyAt ? `${canReapplyAt}${textOf(config, 'reapplySuffix')}` : textOf(config, 'reapplyNotifyFallback'), highlight: true }
    ],
    certNo: application.certNo || application.applicationId || '',
    certTime: reviewedAt || textOf(config, 'backendRecordFallback'),
    rewards: application.rewards,
    hasRewards: application.rewards.length > 0,
    primaryText: application.primaryText || meta.primaryText,
    nextActions: buildApprovedActions(config)
  }
}

Page({
  data: buildPageState(normalizeApplication(createFallbackApplication('guide', 'pending'), 'guide', 'pending', EMPTY_STATUS_CONFIG), EMPTY_STATUS_CONFIG),

  onLoad(options = {}) {
    const roleType = options.roleType || 'guide'
    const status = options.status || 'pending'

    this.setData({
      loading: true,
      loadError: '',
      roleType,
      status
    })
    this.loadRoleStatus(roleType, status)
  },

  async loadRoleStatus(roleType, status) {
    try {
      const results = await Promise.all([
        roleService.getRoleStatusPageConfig(),
        roleService.getMyRoleApplications()
      ])
      const config = normalizeStatusConfig(results[0])
      const applications = Array.isArray(results[1])
        ? results[1]
        : (results[1] && Array.isArray(results[1].items) ? results[1].items : [])
      const application = findApplication(applications, roleType, status, config)

      this.setData(buildPageState(application, config))
    } catch (error) {
      const config = normalizeStatusConfig(this.data.pageConfig)
      this.setData({
        loading: false,
        loadError: error.message || textOf(config, 'loadFailedText'),
        roleType: normalizeRoleType(roleType, config),
        status: normalizeStatus(status, config),
        pageConfig: config,
        texts: config.texts
      })
      toast.info(error.message || textOf(config, 'loadFailedText'))
    }
  },

  handleRetryTap() {
    this.setData({ loading: true, loadError: '' })
    this.loadRoleStatus(this.data.roleType, this.data.status)
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
      toast.info(textOf(this.data.pageConfig, 'routeMissingText'))
      return
    }

    navigateShellRoute(route)
  },

  handleBenefitsTap() {
    navigateShellRoute(`${ROUTES.home}?ui=1&mode=roleComparison&single=1&returnTo=${encodeURIComponent(ROUTES.roleStatus)}`)
  },

  handleHelpTap() {
    navigateShellRoute('/pages/profile/system-management/feedback/index?sheet=quick')
  },

  handleImproveTap() {
    navigateShellRoute(ROUTES.profileSystemProfileInfo)
  }
})
