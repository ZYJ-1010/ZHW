const toast = require('../../../utils/toast')
const gameService = require('../../../services/game')

function formatSavedAt(value) {
  const normalized = typeof value === 'number' ? value : String(value || '').trim()
  const date = new Date(normalized || 0)
  if (!normalized || Number.isNaN(date.getTime())) {
    return '刚刚保存'
  }

  const pad = (number) => String(number).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function normalizeDraft(item = {}) {
  const payload = item.payload || {}
  const form = payload.form || {}
  const location = payload.locationInfo || {}
  const timeDraft = payload.timeDraft || {}
  const timeText = payload.gameTimeConfirmed && timeDraft.startDate && timeDraft.startTime
    ? `${timeDraft.startDate} ${timeDraft.startTime}`
    : '局时间待填写'

  return {
    id: String(item.id || ''),
    title: String(item.title || form.theme || '未命名组局'),
    savedAtText: formatSavedAt(item.updatedAt),
    typeText: String(payload.typeText || form.type || '局类型待选择'),
    timeText,
    locationText: String(location.name || location.address || '地点待填写')
  }
}

Page({
  data: {
    drafts: [],
    loading: true,
    loadingDraftId: '',
    deletingDraftId: ''
  },

  onShow() {
    this.loadDrafts()
  },

  async loadDrafts() {
    const requestSequence = Number(this.loadRequestSequence || 0) + 1
    this.loadRequestSequence = requestSequence
    this.setData({ loading: true })
    try {
      const data = await gameService.getGameDrafts()
      if (requestSequence !== this.loadRequestSequence) {
        return
      }
      const items = Array.isArray(data && data.items) ? data.items : []
      this.setData({ drafts: items.map(normalizeDraft), loading: false })
    } catch (error) {
      if (requestSequence !== this.loadRequestSequence) {
        return
      }
      this.setData({ drafts: [], loading: false })
      toast.info(error.message || '草稿箱读取失败，请稍后重试')
    }
  },

  async editDraft(event) {
    const draftId = String(event.currentTarget.dataset.id || '')
    if (!draftId || this.data.loadingDraftId || this.data.deletingDraftId) {
      return
    }
    const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []
    const previousPage = pages.length > 1 ? pages[pages.length - 2] : null
    if (!previousPage || previousPage.route !== 'pages/game/create/index'
      || typeof previousPage.restoreRequestedDraft !== 'function') {
      wx.navigateTo({ url: `/pages/game/create/index?draftId=${encodeURIComponent(draftId)}` })
      return
    }

    this.setData({ loadingDraftId: draftId })
    if (typeof wx.showLoading === 'function') {
      wx.showLoading({ title: '正在读取草稿', mask: true })
    }
    try {
      const restored = await previousPage.restoreRequestedDraft(draftId)
      if (restored) {
        wx.navigateBack({ delta: 1 })
      }
    } catch (error) {
      toast.info(error.message || '草稿读取失败，请稍后重试')
    } finally {
      if (typeof wx.hideLoading === 'function') {
        wx.hideLoading()
      }
      this.setData({ loadingDraftId: '' })
    }
  },

  deleteDraft(event) {
    const draftId = String(event.currentTarget.dataset.id || '')
    if (!draftId || this.data.loadingDraftId || this.data.deletingDraftId) {
      return
    }

    wx.showModal({
      title: '删除草稿',
      content: '删除后无法恢复，是否继续？',
      success: (result) => {
        if (!result.confirm) {
          return
        }
        this.setData({ deletingDraftId: draftId })
        gameService.deleteGameDraft(draftId)
          .then(() => this.loadDrafts())
          .catch((error) => {
            toast.info(error.message || '草稿删除失败，请稍后重试')
          })
          .finally(() => this.setData({ deletingDraftId: '' }))
      }
    })
  },

  onBack() {
    wx.navigateBack({ delta: 1 })
  }
})
