const profileService = require('../../../../../services/profile')
const { navigateShellRoute } = require('../../../../../utils/shell-nav')

const INVITE_TABS = [
  { key: 'overview', label: '数据概览', route: '' },
  { key: 'network', label: '关系网络', route: '/pages/profile/service-center/invite/network/index' },
  { key: 'records', label: '邀约记录', route: '/pages/profile/service-center/invite/records/index' },
  { key: 'ranking', label: '贡献排行', route: '/pages/profile/service-center/invite/ranking/index' },
  { key: 'income', label: '收益明细', route: '/pages/profile/service-center/invite/income/index' }
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
      { value: '0%', label: '转化率', trend: '', tone: 'up' },
      { value: '¥0', label: '分润收益', trend: '', tone: 'up' }
    ],
    actions: [
      { key: 'share_card', icon: '🔗', label: '分享邀请码' },
      { key: 'qrcode', icon: '▦', label: '二维码', iconClass: 'white' },
      { key: 'poster', icon: '▧', label: '生成海报', iconClass: 'white' }
    ],
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

      // 接口当前返回旧版字符串 tabs；页面点击路由需要稳定的 key 和 route，
      // 因此只使用本地的结构化标签定义，避免接口结果覆盖后标签消失。
      this.setData({
        ...(result || {}),
        tabs: INVITE_TABS
      })
    } catch (error) {
      console.warn('get invite overview failed', error)
      this.setData({
        loadError: error.message || '邀请概览加载失败'
      })
    }
  },

  handleActionTap(event) {
    const { key } = event.currentTarget.dataset
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
