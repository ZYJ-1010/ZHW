const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

const ASSET_BASE = '/pages/profile/system-management/block-settings/assets'

Page({
  data: {
    icons: {
      avatar: `${ASSET_BASE}/icon-avatar.png`,
      search: `${ASSET_BASE}/icon-search.svg`
    },
    showAddSheet: false,
    whitelist: [],
    candidates: [],
    rules: []
  },

  onLoad(options) {
    this.loadWhitelist()

    if (options && options.sheet === 'add') {
      this.setData({
        showAddSheet: true
      })
    }
  },

  async loadWhitelist() {
    try {
      const data = await profileService.getSystemBlockWhitelist()

      this.setData({
        whitelist: this.normalizeList(data.whitelist || data.list || data.items),
        candidates: this.normalizeList(data.candidates || data.recommendations),
        rules: this.normalizeList(data.rules || data.ruleLines)
      })
    } catch (error) {
      this.setData({
        whitelist: [],
        candidates: [],
        rules: []
      })
      toast.info(error.message || '白名单加载失败')
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

  async handleRemoveTap(event) {
    const index = Number(event.currentTarget.dataset.index)
    const item = this.data.whitelist[index]
    const expertId = item && (item.expertId || item.id)

    if (!expertId) {
      toast.info('缺少白名单行家信息')
      return
    }

    try {
      await profileService.removeSystemBlockWhitelist({
        expertId
      })
      this.setData({
        whitelist: this.data.whitelist.filter((_, itemIndex) => itemIndex !== index)
      })
      toast.info('已从白名单移除')
    } catch (error) {
      toast.info(error.message || '移除白名单失败')
    }
  },

  async handleAddCandidateTap(event) {
    const index = Number(event.currentTarget.dataset.index)
    const candidate = this.data.candidates[index]

    if (!candidate) {
      return
    }

    const expertId = candidate.expertId || candidate.id

    if (!expertId) {
      toast.info('缺少候选行家信息')
      return
    }

    try {
      await profileService.addSystemBlockWhitelist({
        expertId
      })
      this.setData({
        showAddSheet: false
      })
      toast.success('已添加白名单行家')
      this.loadWhitelist()
    } catch (error) {
      toast.info(error.message || '添加白名单失败')
    }
  },

  handleDoneTap() {
    toast.success('白名单已保存')
  },

  normalizeList(list) {
    return Array.isArray(list) ? list : []
  }
})
