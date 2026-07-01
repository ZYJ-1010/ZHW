const profileService = require('../../services/profile')

const ASSET_BASE = '/pages/profile/member/assets'
const DEFAULT_STATUS_HEIGHT_RPX = 88
const DEFAULT_NAV_HEIGHT_RPX = 88
const DEFAULT_FRAME_HEIGHT_RPX = 1624
const TITLE_HEIGHT_RPX = 44
const ACTION_SIZE_RPX = 42
const CAPSULE_ACTION_GAP_RPX = 24
const DEFAULT_RESULT_TOTAL = 0

const PROFILE_ROWS = []

const MATCH_PROFILE = {
  id: '',
  name: '',
  title: '',
  tag: '',
  need: '',
  resource: '',
  address: '',
  distance: '',
  avatar: '',
  profileRoute: ''
}

const MATCHED_PEOPLE = []

const PAGE_CONFIGS = {
  radar: {
    key: 'radar',
    title: '组局雷达',
    skin: 'dark',
    estimate: '',
    primaryAction: '',
    tip: '',
    linkText: ''
  },
  matching: {
    key: 'matching',
    title: '组局雷达',
    skin: 'dark',
    statusText: ''
  },
  info: {
    key: 'info',
    title: '适配信息',
    skin: 'light',
    formRows: PROFILE_ROWS
  },
  query: {
    key: 'query',
    title: '组局雷达',
    skin: 'dark result',
    foundPrefix: '',
    foundCount: '',
    foundSuffix: '',
    profile: MATCH_PROFILE
  },
  result: {
    key: 'result',
    title: '组局雷达',
    skin: 'dark result',
    loadingText: '',
    profile: MATCH_PROFILE,
    resultTitle: '',
    resultCount: '',
    resultDescPrefix: '',
    resultDescSuffix: '',
    resultLink: '',
    actionText: ''
  }
}

function buildPageData(pageKey) {
  const config = PAGE_CONFIGS[pageKey] || PAGE_CONFIGS.radar

  return {
    ...config,
    radarNodes: MATCHED_PEOPLE,
    showShare: config.key !== 'info' && config.key !== 'matching',
    layout: buildHeaderLayout()
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

function getSavedProfileRows() {
  try {
    if (typeof wx !== 'undefined' && wx.getStorageSync) {
      const rows = wx.getStorageSync('memberRadarProfileForm')

      if (Array.isArray(rows)) {
        return rows
      }
    }
  } catch (error) {
  }

  return PROFILE_ROWS
}

function buildMatchRequest(formRows = []) {
  const criteria = collectMatchCriteria(formRows)
  const hasCriteria = Object.keys(criteria).length > 0

  return {
    matchCriteria: criteria,
    matchMode: hasCriteria ? 'criteria' : 'all'
  }
}

function getSavedMatchRequest() {
  try {
    if (typeof wx !== 'undefined' && wx.getStorageSync) {
      const request = wx.getStorageSync('memberRadarMatchRequest')

      if (request && typeof request === 'object' && request.matchMode) {
        return request
      }
    }
  } catch (error) {
  }

  return buildMatchRequest(getSavedProfileRows())
}

function getSavedResultSummary() {
  try {
    if (typeof wx !== 'undefined' && wx.getStorageSync) {
      const summary = wx.getStorageSync('memberRadarResultSummary')

      if (summary && typeof summary === 'object') {
        return {
          total: Number(summary.total) || DEFAULT_RESULT_TOTAL,
          profile: summary.profile || MATCH_PROFILE
        }
      }
    }
  } catch (error) {
  }

  return {
    total: DEFAULT_RESULT_TOTAL,
    profile: MATCH_PROFILE
  }
}

function saveResultSummary(summary) {
  try {
    if (typeof wx !== 'undefined' && wx.setStorageSync) {
      wx.setStorageSync('memberRadarResultSummary', summary)
    }
  } catch (error) {
  }
}

function getStoredMatchId() {
  try {
    if (typeof wx !== 'undefined' && wx.getStorageSync) {
      return wx.getStorageSync('memberRadarMatchId') || ''
    }
  } catch (error) {
  }

  return ''
}

function saveMatchId(matchId) {
  try {
    if (typeof wx !== 'undefined' && wx.setStorageSync) {
      wx.setStorageSync('memberRadarMatchId', matchId || '')
    }
  } catch (error) {
  }
}

function normalizeRadarProfile(item = {}) {
  return {
    id: item.id || item.resultId || item.expertId || '',
    name: item.name || item.nickname || '',
    title: item.title || item.position || '',
    tag: item.tag || item.firstTag || '',
    need: item.need || item.needText || '',
    resource: item.resource || item.supply || item.resourceText || '',
    address: item.address || item.location || '',
    distance: item.distance || item.distanceText || '',
    avatar: item.avatar || item.avatarUrl || '',
    profileRoute: item.profileRoute || item.route || '',
    followed: Boolean(item.followed)
  }
}

function normalizeRadarNode(item = {}, index = 0) {
  return {
    id: item.id || item.resultId || item.expertId || `radar-node-${index}`,
    className: item.className || item.nodeClass || '',
    name: item.name || item.nickname || '',
    title: item.title || '',
    avatar: item.avatar || item.avatarUrl || '',
    shortName: item.shortName || item.avatarText || ''
  }
}

function getMatchResultsList(data = {}) {
  return Array.isArray(data.results || data.list || data.items)
    ? (data.results || data.list || data.items)
    : []
}

function buildQueryData() {
  const summary = getSavedResultSummary()

  return {
    ...buildPageData('query'),
    foundCount: String(summary.total),
    profile: summary.profile
  }
}

function buildResultData() {
  const summary = getSavedResultSummary()

  return {
    ...buildPageData('result'),
    resultCount: String(summary.total),
    profile: summary.profile
  }
}

function buildDataForPage(pageKey) {
  if (pageKey === 'matching') {
    return buildMatchingData()
  }

  if (pageKey === 'query') {
    return buildQueryData()
  }

  if (pageKey === 'result') {
    return buildResultData()
  }

  return buildPageData(pageKey)
}

function buildMatchingData() {
  const request = getSavedMatchRequest()
  const matchResultLabel = request.matchMode === 'criteria' ? '符合条件的企业家' : '适配企业家'

  return {
    ...buildPageData('matching'),
    ...request,
    scannedNodes: [],
    scanFoundCount: 0,
    scanFoundTotal: MATCHED_PEOPLE.length,
    matchResultLabel,
    statusText: request.matchMode === 'criteria'
      ? '人脉雷达正在按您的适配信息寻找企业家…'
      : '人脉雷达正在为您匹配全部适配企业家…'
  }
}

function roundRpx(value) {
  return Math.round(value * 100) / 100
}

function getWindowInfo() {
  if (typeof wx === 'undefined') {
    return null
  }

  if (wx.getWindowInfo) {
    return wx.getWindowInfo()
  }

  if (wx.getSystemInfoSync) {
    return wx.getSystemInfoSync()
  }

  return null
}

function buildHeaderLayout() {
  try {
    if (typeof wx !== 'undefined' && wx.getMenuButtonBoundingClientRect) {
      const menuButton = wx.getMenuButtonBoundingClientRect()
      const windowInfo = getWindowInfo()

      if (menuButton && menuButton.width && menuButton.height && windowInfo && windowInfo.windowWidth) {
        const ratio = 750 / windowInfo.windowWidth
        const windowHeight = Number(windowInfo.windowHeight || windowInfo.screenHeight || 0)
        const frameHeight = windowHeight > 0 ? roundRpx(windowHeight * ratio) : DEFAULT_FRAME_HEIGHT_RPX
        const statusHeight = roundRpx((windowInfo.statusBarHeight || Math.max(0, menuButton.top - 4)) * ratio)
        const capsuleTopGap = Math.max(0, menuButton.top - (windowInfo.statusBarHeight || 0))
        const navHeight = roundRpx((capsuleTopGap * 2 + menuButton.height) * ratio)
        const navTop = statusHeight
        const headerHeight = roundRpx(statusHeight + navHeight)
        const controlTop = roundRpx(navTop + (navHeight - ACTION_SIZE_RPX) / 2)
        const titleTop = roundRpx(navTop + (navHeight - TITLE_HEIGHT_RPX) / 2)
        const actionRight = roundRpx((windowInfo.windowWidth - menuButton.left) * ratio + CAPSULE_ACTION_GAP_RPX)

        return {
          frameStyle: `height: ${frameHeight}rpx; min-height: ${frameHeight}rpx;`,
          contentStyle: `height: ${frameHeight}rpx;`,
          headerStyle: `height: ${headerHeight}rpx;`,
          statusStyle: `height: ${statusHeight}rpx;`,
          navStyle: `top: ${navTop}rpx; height: ${navHeight}rpx;`,
          titleStyle: `top: ${titleTop}rpx; height: ${TITLE_HEIGHT_RPX}rpx; line-height: ${TITLE_HEIGHT_RPX}rpx;`,
          backStyle: `top: ${controlTop}rpx; width: ${ACTION_SIZE_RPX}rpx; height: ${ACTION_SIZE_RPX}rpx;`,
          actionStyle: `top: ${controlTop}rpx; right: ${actionRight}rpx; width: ${ACTION_SIZE_RPX}rpx; height: ${ACTION_SIZE_RPX}rpx;`,
          infoStyle: `padding-top: ${headerHeight}rpx;`
        }
      }
    }
  } catch (error) {
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
      this.setData(buildDataForPage(pageKey))
      this.loadRemotePageData(pageKey)

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
      this.setData(buildDataForPage(pageKey))
      this.loadRemotePageData(pageKey)

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
    async loadRemotePageData(pageKey) {
      try {
        if (pageKey === 'radar') {
          const data = await profileService.getMemberRadarOverview()
          const overview = data.overview || data

          this.setData({
            estimate: overview.estimateText || overview.estimate || '',
            primaryAction: overview.primaryActionText || overview.primaryAction || '',
            tip: overview.tip || overview.tipText || '',
            linkText: overview.linkText || '',
            radarNodes: (Array.isArray(data.nodes || data.radarNodes) ? (data.nodes || data.radarNodes) : []).map(normalizeRadarNode)
          })
          return
        }

        if (pageKey === 'info') {
          const data = await profileService.getMemberRadarProfile()
          const profileForm = data.profileForm || data

          this.setData({
            formRows: Array.isArray(profileForm.fields || profileForm.rows)
              ? (profileForm.fields || profileForm.rows)
              : []
          })
          return
        }

        if (pageKey === 'query' || pageKey === 'result') {
          const matchId = getStoredMatchId()

          if (!matchId) {
            return
          }

          const data = await profileService.getMemberRadarMatchResults({ matchId })
          const results = getMatchResultsList(data)
          const profile = normalizeRadarProfile(data.current || data.currentResult || results[0] || {})
          const total = Number(data.total || results.length) || 0

          saveResultSummary({
            matchId,
            total,
            profile
          })

          this.setData({
            foundCount: String(total || ''),
            resultCount: String(total || ''),
            profile,
            radarNodes: results.map(normalizeRadarNode)
          })
        }
      } catch (error) {
        wx.showToast({
          title: error.message || '人脉雷达数据加载失败',
          icon: 'none'
        })
      }
    },

    stopScan() {
      if (this.scanTimer) {
        clearInterval(this.scanTimer)
        this.scanTimer = null
      }
    },

    startScan() {
      const nodes = this.data.radarNodes || MATCHED_PEOPLE
      const label = this.data.matchResultLabel || '符合条件的企业家'

      this.stopScan()
      this.setData({
        scannedNodes: [],
        scanFoundCount: 0,
        statusText: this.data.matchMode === 'criteria'
          ? '人脉雷达正在按您的适配信息寻找企业家…'
          : '人脉雷达正在为您匹配全部适配企业家…'
      })

      if (!nodes.length) {
        saveResultSummary({
          matchId: getStoredMatchId(),
          total: DEFAULT_RESULT_TOTAL,
          profile: MATCH_PROFILE
        })
        return
      }

      let index = 0

      this.scanTimer = setInterval(() => {
        index += 1

        const scannedNodes = nodes.slice(0, Math.min(index, nodes.length))
        const done = scannedNodes.length >= nodes.length

        this.setData({
          scannedNodes,
          scanFoundCount: scannedNodes.length,
          statusText: done
            ? `已扫描到 ${scannedNodes.length} 位${label}`
            : `正在扫描，已发现 ${scannedNodes.length} 位${label}`
        })

        if (done) {
          saveResultSummary(getSavedResultSummary())
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

      wx.redirectTo({
        url: '/pages/profile/member/index'
      })
    },

    async handleAction(event) {
      const { action } = event.currentTarget.dataset
      const messageMap = {
        save: '保存接口待接入',
        next: '下一位待接入',
        follow: '关注接口待接入',
        profile: '个人主页待接入',
        share: '分享功能待接入'
      }

      if (action === 'start') {
        const request = buildMatchRequest(getSavedProfileRows())

        try {
          const result = await profileService.startMemberRadarMatch({
            criteria: request.matchCriteria,
            matchMode: request.matchMode
          })
          const matchId = result.matchId || result.id || ''

          saveMatchId(matchId)
          saveResultSummary({
            matchId,
            total: Number(result.total) || 0,
            profile: normalizeRadarProfile(result.current || result.profile || {})
          })

          if (typeof wx !== 'undefined' && wx.setStorageSync) {
            wx.setStorageSync('memberRadarMatchRequest', request)
          }
        } catch (error) {
          wx.showToast({
            title: error.message || '发起适配失败',
            icon: 'none'
          })
          return
        }

        wx.navigateTo({
          url: '/pages/profile/member/match/index'
        })
        return
      }

      if (action === 'save') {
        try {
          await profileService.saveMemberRadarProfile({
            fields: this.data.formRows || []
          })

          if (typeof wx !== 'undefined' && wx.setStorageSync) {
            wx.setStorageSync('memberRadarProfileForm', this.data.formRows || PROFILE_ROWS)
          }
        } catch (error) {
          wx.showToast({
            title: error.message || '保存适配信息失败',
            icon: 'none'
          })
          return
        }

        wx.showToast({
          title: '已保存',
          icon: 'none'
        })
        return
      }

      if (action === 'info') {
        wx.navigateTo({
          url: '/pages/profile/member/match-info/index'
        })
        return
      }

      if (action === 'review') {
        wx.redirectTo({
          url: '/pages/profile/member/match-query/index'
        })
        return
      }

      if (action === 'rematch') {
        const request = buildMatchRequest(getSavedProfileRows())

        saveMatchId('')
        saveResultSummary({
          total: 0,
          profile: MATCH_PROFILE
        })

        try {
          if (typeof wx !== 'undefined' && wx.setStorageSync) {
            wx.setStorageSync('memberRadarMatchRequest', request)
          }
        } catch (error) {
        }

        wx.redirectTo({
          url: '/pages/profile/member/match/index'
        })
        return
      }

      if (action === 'follow') {
        const resultId = this.data.profile && this.data.profile.id

        if (!resultId) {
          wx.showToast({
            title: '缺少适配对象信息',
            icon: 'none'
          })
          return
        }

        try {
          await profileService.followMemberRadarResult({ resultId })
          this.setData({
            'profile.followed': true
          })
          wx.showToast({
            title: '已关注',
            icon: 'none'
          })
        } catch (error) {
          wx.showToast({
            title: error.message || '关注失败',
            icon: 'none'
          })
        }
        return
      }

      if (action === 'profile' && this.data.profile && this.data.profile.profileRoute) {
        wx.navigateTo({
          url: this.data.profile.profileRoute.startsWith('/')
            ? this.data.profile.profileRoute
            : `/${this.data.profile.profileRoute}`
        })
        return
      }

      wx.showToast({
        title: messageMap[action] || '功能待接入',
        icon: 'none'
      })
    }
  }
})
