const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')
const { navigateShellRoute } = require('../../../../utils/shell-nav')

const ASSET_BASE = '/pages/profile/system-management/block-settings/assets'
const PAGE_ROUTES = {
  protection: '/pages/profile/system-management/protection-mode/index',
  scene: '/pages/profile/system-management/scene-config/index',
  whitelist: '/pages/profile/system-management/whitelist/index',
  renewal: '/pages/profile/system-management/renewal/index',
  keywords: '/pages/profile/system-management/keywords/index',
  users: '/pages/profile/system-management/users/index'
}

const PAGE_NAMES = {
  protection: '保护模式',
  scene: '分场景配置',
  whitelist: '白名单',
  renewal: '续期保护',
  keywords: '关键词屏蔽',
  users: '用户屏蔽'
}

Page({
  data: {
    icons: {
      chevron: `${ASSET_BASE}/icon-chevron-right.svg`
    },
    routes: PAGE_ROUTES,
    enabled: true,
    debugMessage: '',
    summary: {
      protectedUserText: '',
      blockedExpertText: '',
      renewalDaysText: ''
    },
    stats: [],
    configRows: [],
    manageRows: [],
    rules: []
  },

  onLoad() {
    this.loadBlockSettings()
  },

  async loadBlockSettings() {
    try {
      const data = await profileService.getSystemBlockSettings()
      this.setData({
        enabled: data.enabled !== false,
        summary: data.summary || this.data.summary,
        stats: Array.isArray(data.stats) ? data.stats : [],
        configRows: Array.isArray(data.configRows) ? data.configRows.map(withRoute) : [],
        manageRows: Array.isArray(data.manageRows) ? data.manageRows.map(withRoute) : [],
        rules: Array.isArray(data.rules) ? data.rules : []
      })
    } catch (error) {
      this.setData({
        stats: [],
        configRows: [],
        manageRows: [],
        rules: []
      })
      toast.info(error.message || '屏蔽设置暂时不可用')
    }
  },

  async handleToggle() {
    const enabled = !this.data.enabled
    this.setData({ enabled })

    try {
      await profileService.saveSystemBlockSettings({ enabled })
      toast.success(enabled ? '已开启保护' : '已关闭保护')
    } catch (error) {
      this.setData({ enabled: !enabled })
      toast.info(error.message || '保存失败')
    }
  },

  handleNavigateTap(event) {
    const { key } = event.currentTarget.dataset
    const url = PAGE_ROUTES[key]
    const name = PAGE_NAMES[key] || '页面'

    if (!url) {
      this.setData({
        debugMessage: `未找到页面路由：${key || '空'}`
      })
      return
    }

    if (this.data.debugMessage) {
      this.setData({ debugMessage: '' })
    }

    if (!navigateShellRoute(url)) {
      this.setData({
        debugMessage: `${name}暂时无法打开`
      })
    }
  }
})

function withRoute(item) {
  return {
    ...item,
    route: PAGE_ROUTES[item.key] || ''
  }
}
