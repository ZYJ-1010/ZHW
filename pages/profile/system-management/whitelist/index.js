const toast = require('../../../../utils/toast')

const ASSET_BASE = '/pages/profile/system-management/block-settings/assets'

Page({
  data: {
    icons: {
      avatar: `${ASSET_BASE}/icon-avatar.png`,
      search: `${ASSET_BASE}/icon-search.svg`
    },
    showAddSheet: false,
    whitelist: [
      { name: '合作行家 · 互推协议' },
      { name: '联盟行家 · 城市联盟' },
      { name: '导师行家 · 带教关系' }
    ],
    candidates: [
      { name: '张三 · 桌游行家', meta: '北京 · 128场局 · 4.9分' },
      { name: '李四 · 剧本杀行家', meta: '上海 · 256场局 · 4.8分' },
      { name: '王五 · 密室逃脱', meta: '广州 · 89场局 · 4.7分' }
    ],
    rules: [
      { prefix: '白名单行', strong: '不受保护', suffix: '影响' },
      { prefix: '最多 添加', strong: '20位', suffix: '' }
    ]
  },

  onLoad(options) {
    if (options && options.sheet === 'add') {
      this.setData({
        showAddSheet: true
      })
    }
  },

  handleOpenAddSheet() {
    this.setData({
      showAddSheet: true
    })
  },

  handleCloseSheet() {
    this.setData({
      showAddSheet: false
    })
  },

  handleRemoveTap(event) {
    const index = Number(event.currentTarget.dataset.index)
    const whitelist = this.data.whitelist.filter((_, itemIndex) => itemIndex !== index)

    this.setData({ whitelist })
    toast.info('已从白名单移除')
  },

  handleAddCandidateTap(event) {
    const index = Number(event.currentTarget.dataset.index)
    const candidate = this.data.candidates[index]

    if (!candidate) {
      return
    }

    this.setData({
      whitelist: this.data.whitelist.concat({ name: candidate.name }),
      showAddSheet: false
    })
    toast.success('已添加白名单行家')
  },

  handleDoneTap() {
    toast.success('白名单已保存')
  }
})
