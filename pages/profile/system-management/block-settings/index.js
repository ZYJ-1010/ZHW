const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

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
  renewal: '续期保护期',
  keywords: '关键词屏蔽',
  users: '用户屏蔽'
}

const DEBUG_PREFIX = '[block-settings]'
const CONFIG_ROW_META = [
  { key: 'protection', title: '保护模式', iconText: '🎯', tone: '', route: PAGE_ROUTES.protection },
  { key: 'scene', title: '分场景配置', iconText: '⚙️', tone: 'green', route: PAGE_ROUTES.scene },
  { key: 'whitelist', title: '白名单', iconText: '📋', tone: 'gold', route: PAGE_ROUTES.whitelist }
]
const MANAGE_ROW_META = [
  { key: 'users', title: '用户屏蔽', iconText: '🚫', tone: 'red', route: PAGE_ROUTES.users },
  { key: 'keywords', title: '关键词屏蔽', iconText: '🔤', tone: '', route: PAGE_ROUTES.keywords }
]

function normalizeList(list) {
  return Array.isArray(list) ? list : []
}

function mergeRows(metaRows, remoteRows) {
  const remoteMap = normalizeList(remoteRows).reduce((result, item) => {
    result[item.key] = item
    return result
  }, {})

  return metaRows.map((item) => ({
    ...item,
    ...(remoteMap[item.key] || {})
  }))
}

Page({
  data: {
    icons: {
      chevron: `${ASSET_BASE}/icon-chevron-right.svg`
    },
    routes: PAGE_ROUTES,
    enabled: false,
    debugMessage: '',
    summary: {},
    stats: [],
    configRows: CONFIG_ROW_META,
    manageRows: MANAGE_ROW_META,
    rules: []
  },

  onLoad() {
    this.loadBlockSettings()
  },

  async loadBlockSettings() {
    try {
      const data = await profileService.getSystemBlockSettings()

      this.setData({
        enabled: !!data.enabled,
        summary: data.summary || {},
        stats: normalizeList(data.stats),
        configRows: mergeRows(CONFIG_ROW_META, data.configRows),
        manageRows: mergeRows(MANAGE_ROW_META, data.manageRows),
        rules: normalizeList(data.rules || data.ruleLines)
      })
    } catch (error) {
      this.setData({
        enabled: false,
        summary: {},
        stats: [],
        configRows: CONFIG_ROW_META,
        manageRows: MANAGE_ROW_META,
        rules: []
      })
      toast.info(error.message || '屏蔽设置加载失败')
    }
  },

  async handleToggle() {
    const enabled = !this.data.enabled

    this.setData({
      enabled
    })

    try {
      await profileService.saveSystemBlockStatus({
        enabled
      })
    } catch (error) {
      this.setData({
        enabled: !enabled
      })
      toast.info(error.message || '屏蔽设置保存失败')
    }
  },

  handleNavigateTap(event) {
    const { key } = event.currentTarget.dataset
    const url = PAGE_ROUTES[key]
    const name = PAGE_NAMES[key] || '页面'
    const page = this

    if (!url) {
      this.setData({
        debugMessage: `未找到页面路由：${key || '空'}`
      })
      return
    }

    wx.navigateTo({
      url,
      success() {
        if (page.data.debugMessage) {
          page.setData({ debugMessage: '' })
        }
      },
      fail(error) {
        const errMsg = error && error.errMsg ? error.errMsg : ''

        if (errMsg.indexOf('timeout') > -1) {
          console.warn(`${DEBUG_PREFIX} navigate timeout`, {
            key,
            name,
            url
          })
          return
        }

        console.error(`${DEBUG_PREFIX} navigate failed`, {
          key,
          name,
          url,
          errMsg
        })

        page.setData({
          debugMessage: `打开失败：${name}；${errMsg}`
        })
      }
    })
  }
})
