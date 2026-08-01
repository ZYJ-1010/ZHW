const profileService = require('../../../../../services/profile')
const { navigateShellRoute } = require('../../../../../utils/shell-nav')
const { toUserMessage } = require('../../../../../utils/user-message')

const INVITE_TABS = [
  { key: 'overview', label: '数据概览', route: '' },
  { key: 'network', label: '关系网络', route: '/pages/profile/service-center/invite/network/index' },
  { key: 'records', label: '邀约记录', route: '/pages/profile/service-center/invite/records/index' }
]

const INVITE_ACTIONS = [
  { key: 'share_card', icon: '🔗', label: '分享邀请码' },
  { key: 'qrcode', icon: '▦', label: '二维码', iconClass: 'white' },
  { key: 'poster', icon: '▧', label: '生成海报', iconClass: 'white' },
  { key: 'manage_codes', icon: '⚙', label: '邀请码管理', iconClass: 'white' }
]

Page({
  data: {
    profile: {
      level: '',
      name: '',
      desc: ''
    },
    metrics: [
      { value: '0', label: '总邀约数', trend: '', tone: 'up' },
      { value: '0', label: '成功转化', trend: '', tone: 'up' },
      { value: '0%', label: '转化率', trend: '', tone: 'up' }
    ],
    actions: [],
    tabs: INVITE_TABS,
    trends: [],
    loadError: ''
  },

  onLoad() {
    this.loadOverview()
  },

  async loadOverview() {
    try {
      const result = await profileService.getInviteOverview()

      // 标签需要稳定的 key 和 route；快捷操作则由服务端按已分配的邀请码
      // 返回，避免未分配邀请码的角色看到二维码或分享入口。
      this.setData({
        ...(result || {}),
        tabs: INVITE_TABS,
        actions: Array.isArray(result && result.actions) ? result.actions : INVITE_ACTIONS.filter((item) => item.key === 'manage_codes')
      })
    } catch (error) {
      console.warn('get invite overview failed', error)
      this.setData({
        loadError: toUserMessage(error && error.message, '邀请概览加载失败')
      })
    }
  },

  handleActionTap(event) {
    const { key } = event.currentTarget.dataset
    if (key === 'manage_codes') {
      navigateShellRoute('/pages/profile/service-center/invite/code-management/index', {
        currentRoute: '/pages/profile/service-center/invite/overview/index'
      })
      return
    }
    const entryTypeMap = {
      share_card: 'link',
      qrcode: 'qrcode',
      poster: 'poster'
    }
    const entryType = entryTypeMap[key] || 'link'
    const query = [
      `entryType=${encodeURIComponent(entryType)}`,
      `mode=${encodeURIComponent(key || 'share_card')}`
    ].filter(Boolean).join('&')

    navigateShellRoute(`/pages/profile/service-center/invite/qrcode-share/index${query ? `?${query}` : ''}`)
  },

  handleTabTap(event) {
    const { key } = event.currentTarget.dataset
    const target = INVITE_TABS.find((item) => item.key === key)

    if (!target || !target.route) {
      return
    }

    navigateShellRoute(target.route, {
      currentRoute: '/pages/profile/service-center/invite/overview/index'
    })
  }
})
