const profileApi = require('../../api/modules/profile')
const profileService = require('../../services/profile')
const { navigateShellRoute } = require('../../utils/shell-nav')
const viewportLayout = require('../../utils/viewport-layout')

const DEFAULT_STATUS_HEIGHT_RPX = 88
const DEFAULT_NAV_HEIGHT_RPX = 88
const DEFAULT_FRAME_HEIGHT_RPX = 1624
const TITLE_HEIGHT_RPX = 44
const ACTION_SIZE_RPX = 42
const CAPSULE_ACTION_GAP_RPX = 24

const PROFILE_ROWS = [
  { key: 'location', label: '地址定位', value: '', placeholder: '选择' },
  { key: 'industry', label: '所在行业', value: '', placeholder: '选择' },
  { key: 'revenueScale', label: '营收规模', value: '', placeholder: '选填' },
  { key: 'interestedGames', label: '感兴趣组局', value: '', placeholder: '选择' },
  { key: 'resources', label: '我的资源', value: '', placeholder: '前往个人主页填写' },
  { key: 'needs', label: '我的需求', value: '', placeholder: '前往个人主页填写' },
  { key: 'recentDemand', label: '近期诉求', value: '', placeholder: '自定义填写' }
]

const MATCH_PROFILE = {
  name: '',
  title: '',
  tag: '',
  need: '',
  resource: '',
  address: '',
  distance: '',
  avatar: ''
}

const MATCHED_PEOPLE = []

const PAGE_CONFIGS = {
  radar: {
    key: 'radar',
    title: '',
    skin: 'dark',
    estimate: '',
    primaryAction: '',
    tip: '',
    linkText: ''
  },
  matching: {
    key: 'matching',
    title: '',
    skin: 'dark',
    statusText: ''
  },
  info: {
    key: 'info',
    title: '',
    skin: 'light',
    formRows: PROFILE_ROWS
  },
  query: {
    key: 'query',
    title: '',
    skin: 'dark result',
    foundPrefix: '',
    foundCount: '10',
    foundSuffix: '',
    profile: MATCH_PROFILE
  },
  result: {
    key: 'result',
    title: '',
    skin: 'dark result',
    loadingText: '',
    profile: MATCH_PROFILE,
    resultTitle: '',
    resultCount: '10',
    resultDescPrefix: '',
    resultDescSuffix: '',
    resultLink: '',
    actionText: ''
  }
}

function normalizeRadarConfig(config = {}) {
  const pages = config.pages && typeof config.pages === 'object' ? config.pages : {}
  const formRows = Array.isArray(config.formRows) && config.formRows.length ? config.formRows : PROFILE_ROWS
  const radarNodes = Array.isArray(config.radarNodes) ? config.radarNodes : MATCHED_PEOPLE
  const profile = config.profile && typeof config.profile === 'object'
    ? { ...MATCH_PROFILE, ...config.profile }
    : MATCH_PROFILE
  const result = config.result && typeof config.result === 'object' ? config.result : {}

  return {
    pages: {
      ...PAGE_CONFIGS,
      ...pages
    },
    formRows,
    radarNodes,
    profile,
    result: {
      total: Number(result.total || radarNodes.length || 0)
    }
  }
}

function buildPageData(pageKey, radarConfig) {
  const source = normalizeRadarConfig(radarConfig)
  const config = source.pages[pageKey] || source.pages.radar || PAGE_CONFIGS.radar

  return {
    ...config,
    formRows: config.formRows || source.formRows,
    radarNodes: source.radarNodes,
    profile: config.profile || source.profile,
    showShare: config.key !== 'info' && config.key !== 'matching',
    layout: buildHeaderLayout(),
    radarConfig: source
  }
}

function collectMatchCriteria(formRows = []) {
  return formRows.reduce((criteria, row) => {
    const value = typeof row.value === 'string' ? row.value.trim() : row.value

    if (value) {
      criteria[row.key || row.label] = value
    }

    return criteria
  }, {})
}

function formatRadarScanText(template, count, label) {
  return String(template || '')
    .replace('{count}', String(count))
    .replace('{label}', String(label || ''))
}

function radarText(config, key, fallback = '') {
  const texts = config && config.texts ? config.texts : {}
  return texts[key] || fallback
}

function getSavedProfileRows(defaultRows = PROFILE_ROWS) {
  try {
    if (typeof wx !== 'undefined' && wx.getStorageSync) {
      const rows = wx.getStorageSync('memberRadarProfileForm')

      if (Array.isArray(rows)) {
        return rows
      }
    }
  } catch (error) {
    // Keep unsaved form edits local; the display schema still comes from radar-config.
  }

  return defaultRows
}

function buildMatchRequest(formRows = []) {
  const criteria = collectMatchCriteria(formRows)
  const hasCriteria = Object.keys(criteria).length > 0

  return {
    matchCriteria: criteria,
    matchMode: hasCriteria ? 'criteria' : 'all'
  }
}

function primaryProfile(config) {
  const source = normalizeRadarConfig(config)
  const profile = source.profile || {}
  const node = (source.radarNodes || []).find((item) => item && item.id === profile.id) || (source.radarNodes || [])[0] || {}

  return {
    targetId: String(profile.id || node.id || ''),
    targetUserId: Number(profile.userId || node.userId || 0) || 0
  }
}

function getSavedMatchRequest(defaultRows) {
  try {
    if (typeof wx !== 'undefined' && wx.getStorageSync) {
      const request = wx.getStorageSync('memberRadarMatchRequest')

      if (request && typeof request === 'object' && request.matchMode) {
        return request
      }
    }
  } catch (error) {
    // Keep the last match request local so a page switch does not erase it.
  }

  return buildMatchRequest(getSavedProfileRows(defaultRows))
}

function getSavedResultSummary(radarConfig) {
  const source = normalizeRadarConfig(radarConfig)

  try {
    if (typeof wx !== 'undefined' && wx.getStorageSync) {
      const summary = wx.getStorageSync('memberRadarResultSummary')

      if (summary && typeof summary === 'object') {
        return {
          total: Number(summary.total) || source.result.total,
          profile: summary.profile || source.profile
        }
      }
    }
  } catch (error) {
    // Keep the latest scan summary local until the next radar-config/action response.
  }

  return {
    total: source.result.total,
    profile: source.profile
  }
}

function saveResultSummary(summary) {
  try {
    if (typeof wx !== 'undefined' && wx.setStorageSync) {
      wx.setStorageSync('memberRadarResultSummary', summary)
    }
  } catch (error) {
    // Ignore local storage failures.
  }
}

function buildQueryData(radarConfig) {
  const summary = getSavedResultSummary(radarConfig)

  return {
    ...buildPageData('query', radarConfig),
    foundCount: String(summary.total),
    profile: summary.profile
  }
}

function buildResultData(radarConfig) {
  const summary = getSavedResultSummary(radarConfig)

  return {
    ...buildPageData('result', radarConfig),
    resultCount: String(summary.total),
    profile: summary.profile
  }
}

function buildDataForPage(pageKey, radarConfig) {
  if (pageKey === 'matching') {
    return buildMatchingData(radarConfig)
  }

  if (pageKey === 'query') {
    return buildQueryData(radarConfig)
  }

  if (pageKey === 'result') {
    return buildResultData(radarConfig)
  }

  return buildPageData(pageKey, radarConfig)
}

function buildMatchingData(radarConfig) {
  const source = normalizeRadarConfig(radarConfig)
  const request = getSavedMatchRequest(source.formRows)
  const texts = source.texts || {}
  const matchResultLabel = request.matchMode === 'criteria'
    ? (texts.criteriaMatchLabel || '')
    : (texts.allMatchLabel || '')

  return {
    ...buildPageData('matching', source),
    ...request,
    scannedNodes: [],
    scanFoundCount: 0,
    scanFoundTotal: source.radarNodes.length,
    matchResultLabel,
    statusText: request.matchMode === 'criteria'
      ? (texts.criteriaScanningText || '')
      : (texts.allScanningText || '')
  }
}

function roundRpx(value) {
  return Math.round(value * 100) / 100
}

function buildHeaderLayout() {
  const metrics = viewportLayout.getViewportMetrics({ designWidthRpx: 750 })

  if (metrics) {
    const contentHeightRpx = roundRpx(Math.max(
      0,
      metrics.canvasHeightRpx - metrics.safeBottomRpx
    ))
    const controlTopRpx = roundRpx(
      metrics.capsuleTopRpx + (metrics.capsuleHeightRpx - ACTION_SIZE_RPX) / 2
    )
    const titleTopRpx = roundRpx(
      metrics.capsuleTopRpx + (metrics.capsuleHeightRpx - TITLE_HEIGHT_RPX) / 2
    )
    const actionRightRpx = roundRpx(
      metrics.capsuleRightRpx + metrics.capsuleWidthRpx + CAPSULE_ACTION_GAP_RPX
    )

    return {
      frameStyle: `height: ${metrics.canvasHeightRpx}rpx; min-height: ${metrics.canvasHeightRpx}rpx;`,
      contentStyle: `height: ${contentHeightRpx}rpx;`,
      headerStyle: `height: ${metrics.navigationBottomRpx}rpx;`,
      statusStyle: `height: ${metrics.statusBarHeightRpx}rpx;`,
      navStyle: `top: ${metrics.statusBarHeightRpx}rpx; height: ${metrics.navigationHeightRpx}rpx;`,
      titleStyle: `top: ${titleTopRpx}rpx; height: ${TITLE_HEIGHT_RPX}rpx; line-height: ${TITLE_HEIGHT_RPX}rpx;`,
      backStyle: `top: ${controlTopRpx}rpx; width: ${ACTION_SIZE_RPX}rpx; height: ${ACTION_SIZE_RPX}rpx;`,
      actionStyle: `top: ${controlTopRpx}rpx; right: ${actionRightRpx}rpx; width: ${ACTION_SIZE_RPX}rpx; height: ${ACTION_SIZE_RPX}rpx;`,
      infoStyle: `padding-top: ${metrics.navigationBottomRpx}rpx;`,
      safeBottomRpx: metrics.safeBottomRpx
    }
  }

  const headerHeight = DEFAULT_STATUS_HEIGHT_RPX + DEFAULT_NAV_HEIGHT_RPX

  return {
    frameStyle: `height: ${DEFAULT_FRAME_HEIGHT_RPX}rpx; min-height: ${DEFAULT_FRAME_HEIGHT_RPX}rpx;`,
    contentStyle: `height: ${DEFAULT_FRAME_HEIGHT_RPX}rpx;`,
    headerStyle: `height: ${headerHeight}rpx;`,
    statusStyle: `height: ${DEFAULT_STATUS_HEIGHT_RPX}rpx;`,
    navStyle: `top: ${DEFAULT_STATUS_HEIGHT_RPX}rpx; height: ${DEFAULT_NAV_HEIGHT_RPX}rpx;`,
    titleStyle: 'top: 106rpx; height: 44rpx; line-height: 44rpx;',
    backStyle: `top: 108rpx; width: ${ACTION_SIZE_RPX}rpx; height: ${ACTION_SIZE_RPX}rpx;`,
    actionStyle: `top: 108rpx; right: 52rpx; width: ${ACTION_SIZE_RPX}rpx; height: ${ACTION_SIZE_RPX}rpx;`,
    infoStyle: `padding-top: ${headerHeight}rpx;`
  }
}

Component({
  properties: {
    pageKey: {
      type: String,
      value: 'radar'
    }
  },

  data: buildPageData('radar'),

  observers: {
    pageKey(pageKey) {
      this.setData(buildDataForPage(pageKey, this.data.radarConfig))

      if (pageKey === 'matching') {
        this.startScan()
      } else {
        this.stopScan()
      }
    }
  },

  lifetimes: {
    attached() {
      const pageKey = this.properties.pageKey
      this.setData(buildDataForPage(pageKey, this.data.radarConfig))
      this.loadRadarConfig()

      if (pageKey === 'matching') {
        this.startScan()
      }
    },
    detached() {
      this.stopScan()
    }
  },

  pageLifetimes: {
    show() {
      this.setData({
        layout: buildHeaderLayout()
      })

      if (this.properties.pageKey === 'matching') {
        this.startScan()
      }
    },
    hide() {
      this.stopScan()
    },
    resize() {
      this.setData({
        layout: buildHeaderLayout()
      })
    }
  },

  methods: {
    async loadRadarConfig() {
      try {
        const result = await profileApi.getMemberRadarConfig()
        const config = normalizeRadarConfig(result.data || result)

        this.setData(buildDataForPage(this.properties.pageKey, config))

        if (this.properties.pageKey === 'matching') {
          this.startScan()
        }
      } catch (error) {
        this.setData(buildDataForPage(this.properties.pageKey, this.data.radarConfig))
      }
    },

    stopScan() {
      if (this.scanTimer) {
        clearInterval(this.scanTimer)
        this.scanTimer = null
      }
    },

    startScan() {
      const nodes = this.data.radarNodes || []
      const texts = this.data.radarConfig && this.data.radarConfig.texts ? this.data.radarConfig.texts : {}
      const label = this.data.matchResultLabel || texts.criteriaMatchLabel || ''

      this.stopScan()
      this.setData({
        scannedNodes: [],
        scanFoundCount: 0,
        statusText: this.data.matchMode === 'criteria'
          ? (texts.criteriaScanningText || '')
          : (texts.allScanningText || '')
      })

      let index = 0

      this.scanTimer = setInterval(() => {
        index += 1

        const scannedNodes = nodes.slice(0, Math.min(index, nodes.length))
        const done = scannedNodes.length >= nodes.length

        this.setData({
          scannedNodes,
          scanFoundCount: scannedNodes.length,
          statusText: done
            ? formatRadarScanText(texts.scanDoneTemplate, scannedNodes.length, label)
            : formatRadarScanText(texts.scanProgressTemplate, scannedNodes.length, label)
        })

        if (done) {
          saveResultSummary({
            total: this.data.radarConfig && this.data.radarConfig.result
              ? this.data.radarConfig.result.total
              : nodes.length,
            profile: this.data.profile || MATCH_PROFILE
          })
          this.stopScan()
        }
      }, 560)
    },

    handleBack() {
      const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []

      if (pages.length > 1) {
        wx.navigateBack()
        return
      }

      navigateShellRoute('/pages/profile/member/index')
    },

    async submitRadarAction(action, extra = {}) {
      try {
        return await profileService.submitMemberRadarAction({
          action,
          formRows: this.data.formRows || PROFILE_ROWS,
          ...buildMatchRequest(this.data.formRows || PROFILE_ROWS),
          ...primaryProfile(this.data.radarConfig),
          ...extra
        })
      } catch (error) {
        wx.showToast({
          title: error.message || radarText(this.data.radarConfig, 'actionFailedText'),
          icon: 'none'
        })
        return null
      }
    },

    async handleAction(event) {
      const { action } = event.currentTarget.dataset
      const messageMap = this.data.radarConfig && this.data.radarConfig.actionMessages
        ? this.data.radarConfig.actionMessages
        : {}

      if (action === 'start') {
        const request = buildMatchRequest(this.data.formRows || PROFILE_ROWS)
        const result = await this.submitRadarAction('start', request)

        if (!result) {
          return
        }

        try {
          if (typeof wx !== 'undefined' && wx.setStorageSync) {
            wx.setStorageSync('memberRadarMatchRequest', request)
          }
        } catch (error) {
          // Ignore local storage failures.
        }

        navigateShellRoute('/pages/profile/member/match/index')
        return
      }

      if (action === 'save') {
        const result = await this.submitRadarAction('save')

        if (!result) {
          return
        }

        try {
          if (typeof wx !== 'undefined' && wx.setStorageSync) {
            wx.setStorageSync('memberRadarProfileForm', this.data.formRows || PROFILE_ROWS)
          }
        } catch (error) {
          // Ignore local storage failures.
        }

        wx.showToast({
          title: messageMap[action],
          icon: 'none'
        })
        return
      }

      if (action === 'info') {
        navigateShellRoute('/pages/profile/member/match-info/index')
        return
      }

      if (action === 'review') {
        navigateShellRoute('/pages/profile/member/match-query/index')
        return
      }

      if (action === 'rematch') {
        const request = buildMatchRequest(this.data.formRows || PROFILE_ROWS)
        const result = await this.submitRadarAction('rematch', request)

        if (!result) {
          return
        }

        try {
          if (typeof wx !== 'undefined' && wx.setStorageSync) {
            wx.setStorageSync('memberRadarMatchRequest', request)
          }
        } catch (error) {
          // Ignore local storage failures.
        }

        navigateShellRoute('/pages/profile/member/match/index')
        return
      }

      if (action === 'next' || action === 'follow') {
        const result = await this.submitRadarAction(action)

        if (!result) {
          return
        }

        wx.showToast({
          title: messageMap[action],
          icon: 'none'
        })
        return
      }

      if (action === 'profile') {
        const result = await this.submitRadarAction('profile')
        const route = result && result.route

        if (route) {
          navigateShellRoute(route)
          return
        }

        wx.showToast({
          title: messageMap[action],
          icon: 'none'
        })
        return
      }

      wx.showToast({
        title: messageMap[action] || radarText(this.data.radarConfig, 'entryMissingText'),
        icon: 'none'
      })
    }
  }
})
