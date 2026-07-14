const profileService = require('../../../../../services/profile')
const { navigateShellRoute } = require('../../../../../utils/shell-nav')

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
      { icon: '🔗', label: '分享邀请码' },
      { icon: '▦', label: '二维码', iconClass: 'white' },
      { icon: '▧', label: '生成海报', iconClass: 'white' }
    ],
    tabs: ['数据概览', '关系网络', '邀约记录', '贡献排行', '收益明细'],
    trends: [],
    loadError: ''
  },

  onLoad() {
    this.loadOverview()
  },

  async loadOverview() {
    try {
      const result = await profileService.getInviteOverview()

      this.setData(result || {})
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
  }
})
