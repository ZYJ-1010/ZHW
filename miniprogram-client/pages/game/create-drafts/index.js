const toast = require('../../../utils/toast')
const gameService = require('../../../services/game')

function formatSavedAt(value) {
  const date = new Date(Number(value) || 0)
  if (Number.isNaN(date.getTime()) || !Number(value)) {
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
    drafts: []
  },

  onShow() {
    this.loadDrafts()
  },

  async loadDrafts() {
    try {
      const data = await gameService.getGameDrafts()
      const items = Array.isArray(data && data.items) ? data.items : []
      this.setData({ drafts: items.map(normalizeDraft) })
    } catch (error) {
      this.setData({ drafts: [] })
      toast.info(error.message || '草稿箱读取失败，请稍后重试')
    }
  },

  editDraft(event) {
    const draftId = String(event.currentTarget.dataset.id || '')
    if (!draftId) {
      return
    }
    wx.navigateTo({ url: `/pages/game/create/index?draftId=${encodeURIComponent(draftId)}` })
  },

  deleteDraft(event) {
    const draftId = String(event.currentTarget.dataset.id || '')
    if (!draftId) {
      return
    }

    wx.showModal({
      title: '删除草稿',
      content: '删除后无法恢复，是否继续？',
      success: (result) => {
        if (!result.confirm) {
          return
        }
        gameService.deleteGameDraft(draftId).then(() => this.loadDrafts()).catch((error) => {
          toast.info(error.message || '草稿删除失败，请稍后重试')
        })
      }
    })
  },

  onBack() {
    wx.navigateBack({ delta: 1 })
  }
})
