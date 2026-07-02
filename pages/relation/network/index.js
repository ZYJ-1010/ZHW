const { ROUTES } = require('../../../config/routes')
const relationService = require('../../../services/relation')
const { navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')

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
  onlineText: '在线',
  header: {
    title: '星巴克(镇海万科店)',
    titleIcon: '📍',
    statusText: '营业中',
    address: '宁波市镇海区庄市大道1088号万科广场1F'
  },
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
    activeTab: source.activeTab || source.defaultTab || FALLBACK_NETWORK_HOME.activeTab
  }
}

Page({
  data: {
    onlineText: '在线',
    navItems: NAV_ITEMS,
    relationTabs: RELATION_TABS,
    activeTab: 'network',
    header: FALLBACK_NETWORK_HOME.header,
    loading: false
  },

  onLoad(options) {
    this.loadNetworkHome(options || {})
  },

  async loadNetworkHome(params = {}) {
    this.setData({
      loading: true
    })

    try {
      const data = await relationService.getNetworkHome(params)
      const normalized = normalizeNetworkHome(data)

      this.setData({
        onlineText: normalized.onlineText,
        header: normalized.header,
        relationTabs: normalized.relationTabs,
        activeTab: normalized.activeTab,
        loading: false
      })
    } catch (error) {
      const normalized = normalizeNetworkHome(FALLBACK_NETWORK_HOME)

      this.setData({
        onlineText: normalized.onlineText,
        header: normalized.header,
        relationTabs: normalized.relationTabs,
        activeTab: normalized.activeTab,
        loading: false
      })
    }
  },

  handleBackTap() {
    wx.navigateBack({
      delta: 1,
      fail: () => {
        navigateShellRoute(ROUTES.home)
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
    navigateShellKey(key, {
      currentRoute: ROUTES.relationNetwork
    })
  }
})
