const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

const ASSET_BASE = '/pages/profile/system-management/block-settings/assets'

Page({
  data: {
    icons: {
      search: `${ASSET_BASE}/icon-search.svg`,
      lightbulb: `${ASSET_BASE}/icon-lightbulb.svg`,
      plus: `${ASSET_BASE}/icon-plus.png`,
      avatar: `${ASSET_BASE}/icon-avatar.png`,
      close: `${ASSET_BASE}/icon-close.svg`,
      check: `${ASSET_BASE}/icon-check.svg`
    },
    showAddSheet: false,
    selectedCount: 0,
    hasSelection: false,
    blockSettings: null,
    users: [],
    candidates: [],
    rules: []
  },

  onLoad(options) {
    if (options && options.sheet === 'add') {
      this.setData({
        showAddSheet: true
      })
    }
    this.loadBlockedUsers()
  },

  async loadBlockedUsers() {
    try {
      const data = await profileService.getSystemBlockSettings()
      const candidates = Array.isArray(data.blockCandidates) ? data.blockCandidates : []

      this.setData({
        blockSettings: data || null,
        users: Array.isArray(data.blockedUsers) ? data.blockedUsers : [],
        candidates: candidates.map((item) => ({ ...item, checked: false })),
        rules: Array.isArray(data.blockRules) ? data.blockRules : []
      })
    } catch (error) {
      this.setData({ users: [], candidates: [], rules: [] })
      toast.info(error.message || '屏蔽用户暂时不可用')
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

  handleUnblockTap(event) {
    const index = Number(event.currentTarget.dataset.index)

    const users = this.data.users.filter((_, itemIndex) => itemIndex !== index)

    this.setData({ users })
    profileService.saveSystemBlockSettings({ ...(this.data.blockSettings || {}), blockedUsers: users }).then(() => {
      toast.info('已解除屏蔽')
    }).catch((error) => {
      toast.info(error.message || '保存失败')
    })
  },

  handleCandidateToggle(event) {
    const index = Number(event.currentTarget.dataset.index)
    const candidates = this.data.candidates.map((item, itemIndex) => (
      itemIndex === index ? { ...item, checked: !item.checked } : item
    ))
    const selectedCount = candidates.filter((item) => item.checked).length

    this.setData({
      candidates,
      selectedCount,
      hasSelection: selectedCount > 0
    })
  },

  handleConfirmBlock() {
    const selected = this.data.candidates.filter((item) => item.checked)

    if (!selected.length) {
      toast.info('请选择要屏蔽的用户')
      return
    }

    const users = this.data.users.concat(selected.map((item) => ({
      id: item.id,
      name: item.name,
      role: item.role,
      reason: item.reason,
      date: item.date || ''
    })))

    this.setData({
      users,
      showAddSheet: false,
      selectedCount: 0,
      hasSelection: false,
      candidates: this.data.candidates.map((item) => ({ ...item, checked: false }))
    })
    profileService.saveSystemBlockSettings({ ...(this.data.blockSettings || {}), blockedUsers: users }).then(() => {
      toast.success('已添加屏蔽用户')
    }).catch((error) => {
      toast.info(error.message || '保存失败')
    })
  }
})
