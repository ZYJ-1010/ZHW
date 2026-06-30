const { ROUTES } = require('../../../config/routes')
const relationService = require('../../../services/relation')

const NAV_ITEMS = [
  { name: '我的', key: 'mine' },
  { name: '元宇宙', key: 'metaverse' },
  { name: '地图', key: 'map' },
  { name: '消息', key: 'message' },
  { name: '首页', key: 'home' }
]

const RELATION_TABS = [
  { key: 'network', text: '人脉网络' },
  { key: 'nearby', text: '附近玩家' }
]

const FALLBACK_NETWORK_HOME = {
  onlineText: '',
  header: {},
  tabs: RELATION_TABS,
  activeTab: 'network'
}

function normalizeNetworkHome(data) {
  const source = data && typeof data === 'object' ? data : {}
  const fallbackHeader = FALLBACK_NETWORK_HOME.header
  const header = source.header || source.venue || {}
  const tabs = Array.isArray(source.tabs) && source.tabs.length
    ? source.tabs
    : FALLBACK_NETWORK_HOME.tabs

  return {
    onlineText: source.onlineText || FALLBACK_NETWORK_HOME.onlineText,
    header: {
      title: header.title || header.name || fallbackHeader.title,
      titleIcon: header.titleIcon || header.icon || fallbackHeader.titleIcon,
      statusText: header.statusText || header.businessStatusText || fallbackHeader.statusText,
      address: header.address || header.addressText || fallbackHeader.address
    },
    relationTabs: tabs.map((item) => ({
      key: item.key || item.id || 'network',
      text: item.text || item.name || item.title || ''
    })).filter((item) => item.key && item.text),
    activeTab: source.activeTab || source.defaultTab || FALLBACK_NETWORK_HOME.activeTab,
    hasHeader: Boolean(header.title || header.name || header.address || header.addressText),
    errorText: ''
  }
}

Page({
  data: {
    onlineText: '',
    navItems: NAV_ITEMS,
    relationTabs: RELATION_TABS,
    activeTab: 'network',
    header: FALLBACK_NETWORK_HOME.header,
    hasHeader: false,
    loading: false,
    errorText: ''
  },

  onLoad(options) {
    this.loadNetworkHome(options || {})
  },

  async loadNetworkHome(params = {}) {
    this.setData({
      loading: true,
      errorText: ''
    })

    try {
      const data = await relationService.getNetworkHome(params)
      const normalized = normalizeNetworkHome(data)

      this.setData({
        onlineText: normalized.onlineText,
        header: normalized.header,
        relationTabs: normalized.relationTabs,
        activeTab: normalized.activeTab,
        hasHeader: normalized.hasHeader,
        errorText: '',
        loading: false
      })
    } catch (error) {
      const normalized = normalizeNetworkHome()

      this.setData({
        onlineText: normalized.onlineText,
        header: normalized.header,
        relationTabs: normalized.relationTabs,
        activeTab: normalized.activeTab,
        hasHeader: false,
        errorText: error && error.message ? error.message : '关系网络加载失败',
        loading: false
      })
    }
  },

  handleBackTap() {
    wx.navigateBack({
      delta: 1,
      fail: () => {
        wx.navigateTo({
          url: `/${ROUTES.home}`
        })
      }
    })
  },

  handleRefreshTap() {
    this.loadNetworkHome({
      refresh: 1
    })
  },

  handleTabTap(event) {
    const { key } = event.currentTarget.dataset

    if (!key || key === this.data.activeTab) {
      return
    }

    this.setData({
      activeTab: key
    })
  },

  handleShellNavTap(event) {
    const { key } = event.detail || {}

    if (key === 'map') {
      wx.showToast({
        title: '地图功能开发中',
        icon: 'none'
      })
      return
    }
    const routeMap = {
      home: ROUTES.playerHome,
      map: '',
      message: ROUTES.message,
      mine: ROUTES.profile,
      metaverse: ROUTES.metaverse
    }
    const route = routeMap[key]

    if (!route) {
      return
    }

    wx.navigateTo({
      url: `/${route}`
    })
  }
})
