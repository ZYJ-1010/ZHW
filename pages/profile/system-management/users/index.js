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
    users: [],
    candidates: [],
    rules: []
  },

  onLoad(options) {
    this.loadBlockedUsers()

    if (options && options.sheet === 'add') {
      this.setData({
        showAddSheet: true
      })
    }
  },

  async loadBlockedUsers() {
    try {
      const data = await profileService.getSystemBlockedUsers()

      this.setData({
        users: this.normalizeList(data.users || data.list || data.items),
        candidates: this.normalizeList(data.candidates || data.recommendations)
          .map((item) => ({ ...item, checked: false })),
        rules: this.normalizeList(data.rules || data.ruleLines)
      })
      this.updateSelectionState()
    } catch (error) {
      this.setData({
        users: [],
        candidates: [],
        rules: []
      })
      toast.info(error.message || '屏蔽用户加载失败')
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

  async handleUnblockTap(event) {
    const index = Number(event.currentTarget.dataset.index)
    const user = this.data.users[index]
    const userId = user && (user.userId || user.id)

    if (!userId) {
      toast.info('缺少屏蔽用户信息')
      return
    }

    try {
      await profileService.removeSystemBlockedUser({
        userId
      })
      this.setData({
        users: this.data.users.filter((_, itemIndex) => itemIndex !== index)
      })
      toast.info('已解除屏蔽')
    } catch (error) {
      toast.info(error.message || '解除屏蔽失败')
    }
  },

  handleCandidateToggle(event) {
    const index = Number(event.currentTarget.dataset.index)
    const candidates = this.data.candidates.map((item, itemIndex) => (
      itemIndex === index ? { ...item, checked: !item.checked } : item
    ))

    this.setData({
      candidates
    }, () => this.updateSelectionState())
  },

  async handleConfirmBlock() {
    const selected = this.data.candidates.filter((item) => item.checked)

    if (!selected.length) {
      toast.info('请选择要屏蔽的用户')
      return
    }

    const userIds = selected
      .map((item) => item.userId || item.id)
      .filter(Boolean)

    if (!userIds.length) {
      toast.info('缺少可屏蔽用户信息')
      return
    }

    try {
      await profileService.addSystemBlockedUsers({
        userIds
      })
      this.setData({
        showAddSheet: false,
        selectedCount: 0,
        hasSelection: false,
        candidates: this.data.candidates.map((item) => ({ ...item, checked: false }))
      })
      toast.success('已添加屏蔽用户')
      this.loadBlockedUsers()
    } catch (error) {
      toast.info(error.message || '添加屏蔽用户失败')
    }
  },

  updateSelectionState() {
    const selectedCount = this.data.candidates.filter((item) => item.checked).length

    this.setData({
      selectedCount,
      hasSelection: selectedCount > 0
    })
  },

  normalizeList(list) {
    return Array.isArray(list) ? list : []
  }
})
