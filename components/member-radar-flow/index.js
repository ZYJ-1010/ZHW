const ASSET_BASE = '/pages/profile/member/assets'
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
  name: '陆毅',
  title: '总经理｜上海创世界科技有限公司',
  tag: '第一标签：上海TMT投资领军者，数字化内容服务',
  need: '我的需求：AI赋能与市场运营助力企业IP打造',
  resource: '我的资源：10年TMT投资经验',
  address: '上海市浦东新区沙新镇黄赵路310号',
  distance: '231 km',
  avatar: `${ASSET_BASE}/radar-avatar.png`
}

const QUERY_PROFILE = {
  name: '胡芳',
  title: '董事长、创始人｜千浪化研新材料（上海…',
  tag: '第一标签：手机漆，汽车漆深耕者',
  industry: '化学原料和化学制品制造业',
  supply: '化工涂料的生产销售,专业的塑胶工业漆及手机漆生产者',
  need: '期待与更多需要油漆涂料的岛亲链接交流',
  city: '上海',
  avatar: `${ASSET_BASE}/radar-avatar.png`
}

const MATCHED_PEOPLE = [
  {
    id: 'hu-fang',
    className: 'node-leader',
    name: QUERY_PROFILE.name,
    title: QUERY_PROFILE.title,
    avatar: QUERY_PROFILE.avatar
  },
  {
    id: 'lu-yi',
    className: 'node-maker',
    name: MATCH_PROFILE.name,
    title: MATCH_PROFILE.title,
    avatar: MATCH_PROFILE.avatar
  },
  {
    id: 'chen-zong',
    className: 'node-owner',
    name: '陈总',
    title: '企业服务资源方',
    avatar: '',
    shortName: '陈'
  },
  {
    id: 'wang-zong',
    className: 'node-investor',
    name: '王总',
    title: '产业投资合伙人',
    avatar: '',
    shortName: '王'
  },
  {
    id: 'li-zong',
    className: 'node-expert',
    name: '李总',
    title: '品牌增长顾问',
    avatar: '',
    shortName: '李'
  },
  {
    id: 'zhao-zong',
    className: 'node-partner',
    name: '赵总',
    title: '渠道合作伙伴',
    avatar: '',
    shortName: '赵'
  },
  {
    id: 'sun-zong',
    className: 'node-small',
    name: '孙总',
    title: '本地服务主理人',
    avatar: '',
    shortName: '孙'
  }
]

const PAGE_CONFIGS = {
  radar: {
    key: 'radar',
    title: '组局雷达',
    skin: 'dark',
    estimate: '预计可匹配7461位商界决策者',
    primaryAction: '开启适配人脉',
    tip: '信息填写越完整，人脉匹配越精准',
    linkText: '填写适配信息 >'
  },
  matching: {
    key: 'matching',
    title: '组局雷达',
    skin: 'dark',
    statusText: '人脉雷达正在寻找与您适配的企业家…'
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
    foundPrefix: '为您找到',
    foundCount: '10',
    foundSuffix: '位适配您的优质行家信息',
    profile: MATCH_PROFILE
  },
  result: {
    key: 'result',
    title: '组局雷达',
    skin: 'dark result',
    loadingText: '正在寻找与您适配的优质业务主.....',
    profile: MATCH_PROFILE,
    resultTitle: '本轮匹配组局已完成推荐',
    resultCount: '10',
    resultDescPrefix: '共推荐了',
    resultDescSuffix: '位优质行家',
    resultLink: '重新查看 >',
    actionText: '再次重新匹配'
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
    // Fall back to the static rows until the backend profile-form API is connected.
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
    // Fall back to the static rows until the backend match API is connected.
  }

  return buildMatchRequest(getSavedProfileRows())
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
    // Keep the fallback layout below when running outside a mini-program runtime.
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
      this.setData(pageKey === 'matching' ? buildMatchingData() : buildPageData(pageKey))

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
      this.setData(pageKey === 'matching' ? buildMatchingData() : buildPageData(pageKey))

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

    handleAction(event) {
      const { action } = event.currentTarget.dataset
      const messageMap = {
        save: '保存接口待接入',
        next: '下一位待接入',
        follow: '关注接口待接入',
        rematch: '重新匹配待接入',
        profile: '个人主页待接入',
        share: '分享功能待接入'
      }

      if (action === 'start') {
        const request = buildMatchRequest(getSavedProfileRows())

        try {
          if (typeof wx !== 'undefined' && wx.setStorageSync) {
            wx.setStorageSync('memberRadarMatchRequest', request)
          }
        } catch (error) {
          // Ignore local storage failures for the static prototype state.
        }

        wx.navigateTo({
          url: '/pages/profile/member/match/index'
        })
        return
      }

      if (action === 'save') {
        try {
          if (typeof wx !== 'undefined' && wx.setStorageSync) {
            wx.setStorageSync('memberRadarProfileForm', this.data.formRows || PROFILE_ROWS)
          }
        } catch (error) {
          // Ignore local storage failures for the static prototype state.
        }

        wx.showToast({
          title: messageMap[action],
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

      wx.showToast({
        title: messageMap[action] || '功能待接入',
        icon: 'none'
      })
    }
  }
})
