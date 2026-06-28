const toast = require('../../../../utils/toast')

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
    users: [
      { name: '张三', role: '玩家', reason: '言语骚扰', date: '2026-06-10' },
      { name: '李四', role: '行家', reason: '诱导私下交易', date: '2026-06-08' },
      { name: '王五', role: '玩家', reason: '恶意取消', date: '2026-06-05' }
    ],
    candidates: [
      { name: '赵六', meta: '玩家 · 最近互动：3天前', checked: false },
      { name: '钱七', meta: '玩家 · 最近互动：1周前', checked: false },
      { name: '孙八', meta: '行家 · 最近互动：2周前', checked: false }
    ],
    rules: [
      { prefix: '可通过用户主页右上角「⋮」菜单快速屏蔽', strong: '', suffix: '' },
      { prefix: '屏蔽后双方', strong: '互不可见', suffix: '，历史互动记录保留' },
      { prefix: '解除屏蔽后', strong: '24小时冷却期', suffix: '才能再次屏蔽' },
      { prefix: '屏蔽人数上限', strong: '100人', suffix: '' }
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

  handleUnblockTap(event) {
    const index = Number(event.currentTarget.dataset.index)

    this.setData({
      users: this.data.users.filter((_, itemIndex) => itemIndex !== index)
    })
    toast.info('已解除屏蔽')
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
      name: item.name,
      role: item.meta.indexOf('行家') > -1 ? '行家' : '玩家',
      reason: '手动添加',
      date: '2026-06-28'
    })))

    this.setData({
      users,
      showAddSheet: false,
      selectedCount: 0,
      hasSelection: false,
      candidates: this.data.candidates.map((item) => ({ ...item, checked: false }))
    })
    toast.success('已添加屏蔽用户')
  }
})
