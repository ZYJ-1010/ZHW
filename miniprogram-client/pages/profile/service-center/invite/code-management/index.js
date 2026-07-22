const profileService = require('../../../../../services/profile')

function statusText(status) {
  return ({ active: '可使用', used: '已使用', disabled: '已禁用', voided: '已作废', expired: '已过期', exhausted: '已用尽', pending: '待审核', approved: '已通过', rejected: '已驳回' })[status] || status || '-'
}

function entryText(entryType) {
  return ({ poster: '小程序卡片', qrcode: '二维码', link: '链接' })[entryType] || '-'
}

Page({
  data: { items: [], requests: [], summary: {}, maxRequestCount: 200, loading: true, quantity: '', reason: '' },

  onLoad() { this.loadData() },

  async loadData() {
    try {
      const data = await profileService.getInviteCodes()
      this.setData({
        items: (data.items || []).map((item) => ({ ...item, statusText: statusText(item.useStatus === 'used' ? 'used' : (item.displayStatus || item.status)), entryText: entryText(item.entryType), expiresText: item.expiresAt ? String(item.expiresAt).slice(0, 10) : '长期有效', usedUserText: item.boundWechatNickname || (item.boundWechatUserId ? `用户 ${item.boundWechatUserId}` : '未使用') })),
        requests: (data.requests || []).map((item) => ({ ...item, statusText: statusText(item.status) })),
        summary: data.summary || {},
        maxRequestCount: Number((data.config || {}).maxRequestCount || 200),
        loading: false
      })
    } catch (error) {
      this.setData({ loading: false })
      wx.showToast({ title: error.message || '邀请码加载失败', icon: 'none' })
    }
  },

  onQuantityInput(event) { this.setData({ quantity: event.detail.value }) },
  onReasonInput(event) { this.setData({ reason: event.detail.value }) },

  async submitQuotaRequest() {
    try {
      const quantity = Number(this.data.quantity || 0)
      if (quantity > this.data.maxRequestCount) throw new Error(`申请数量不能超过 ${this.data.maxRequestCount} 个`)
      await profileService.createInviteQuotaRequest({ quantity, reason: this.data.reason })
      this.setData({ quantity: '', reason: '' })
      wx.showToast({ title: '申请已提交', icon: 'success' })
      this.loadData()
    } catch (error) {
      wx.showToast({ title: error.message || '提交失败', icon: 'none' })
    }
  }
})
