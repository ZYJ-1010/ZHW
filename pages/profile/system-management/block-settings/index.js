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

Page({
  data: {
    icons: {
      chevron: `${ASSET_BASE}/icon-chevron-right.svg`
    },
    routes: PAGE_ROUTES,
    enabled: true,
    debugMessage: '',
    summary: {
      protectedUserText: '128位用户',
      blockedExpertText: '12位',
      renewalDaysText: '18天'
    },
    stats: [
      { value: '128', label: '保护用户' },
      { value: '12', label: '已屏蔽行家' },
      { value: '92%', label: '过滤率' }
    ],
    configRows: [
      {
        key: 'protection',
        title: '保护模式',
        desc: '当前：硬保护（完全过滤）',
        badge: '硬保护',
        iconText: '🎯',
        tone: '',
        route: PAGE_ROUTES.protection
      },
      {
        key: 'scene',
        title: '分场景配置',
        desc: '推荐/搜索/附近/列表',
        badge: '4开',
        iconText: '⚙️',
        tone: 'green',
        route: PAGE_ROUTES.scene
      },
      {
        key: 'whitelist',
        title: '白名单',
        desc: '3位行家不受保护',
        badge: '3/20',
        iconText: '📋',
        tone: 'gold',
        route: PAGE_ROUTES.whitelist
      }
    ],
    manageRows: [
      {
        key: 'users',
        title: '用户屏蔽',
        desc: '已屏蔽3位用户，双方互不可见',
        badge: '3人',
        badgeClass: 'red',
        iconText: '🚫',
        tone: 'red',
        route: PAGE_ROUTES.users
      },
      {
        key: 'keywords',
        title: '关键词屏蔽',
        desc: '已设置5个关键词，自动过滤内容',
        badge: '5/20',
        iconText: '🔤',
        tone: '',
        route: PAGE_ROUTES.keywords
      }
    ],
    rules: [
      { prefix: '保护期默认', strong: '30天', suffix: '，到期需重新配置' },
      { prefix: '主动搜索/行家主页/已收藏', strong: '均受保护', suffix: '' },
      { prefix: '新行', strong: '无豁免', suffix: '，一视同仁' },
      { prefix: '配置变更', strong: '5秒内', suffix: '对所有新请求生效' }
    ]
  },

  handleToggle() {
    this.setData({
      enabled: !this.data.enabled
    })
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
