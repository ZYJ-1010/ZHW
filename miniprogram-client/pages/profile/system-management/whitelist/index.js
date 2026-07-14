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
    blockSettings: null,
    whitelist: [],
    candidates: [],
    rules: []
  },

  onLoad(options) {
    if (options && options.sheet === 'add') {
      this.setData({
        showAddSheet: true
      })
    }
    this.loadWhitelist()
  },

  async loadWhitelist() {
    try {
      const data = await profileService.getSystemBlockSettings()
      this.setData({
        blockSettings: data || null,
        whitelist: Array.isArray(data.whitelist) ? data.whitelist : [],
        candidates: Array.isArray(data.whitelistCandidates) ? data.whitelistCandidates : [],
        rules: Array.isArray(data.whitelistRules) ? data.whitelistRules : []
      })
    } catch (error) {
      this.setData({ whitelist: [], candidates: [], rules: [] })
      toast.info(error.message || '白名单暂时不可用')
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
    profileService.saveSystemBlockSettings({ ...(this.data.blockSettings || {}), whitelist }).then(() => {
      toast.info('已从白名单移除')
    }).catch((error) => {
      toast.info(error.message || '保存失败')
    })
  },

  handleAddCandidateTap(event) {
    const index = Number(event.currentTarget.dataset.index)
    const candidate = this.data.candidates[index]

    if (!candidate) {
      return
    }

    const whitelist = this.data.whitelist.concat({
      id: candidate.id,
      name: candidate.name,
      role: candidate.role,
      reason: candidate.reason
    })

    this.setData({
      whitelist,
      showAddSheet: false
    })
    profileService.saveSystemBlockSettings({
      ...(this.data.blockSettings || {}),
      whitelist
    }).then(() => {
      toast.success('已添加白名单')
    }).catch((error) => {
      toast.info(error.message || '保存失败')
    })
  },

  async handleDoneTap() {
    try {
      await profileService.saveSystemBlockSettings({
        ...(this.data.blockSettings || {}),
        whitelist: this.data.whitelist
      })
      toast.success('白名单已保存')
    } catch (error) {
      toast.info(error.message || '保存失败')
    }
  }
})
