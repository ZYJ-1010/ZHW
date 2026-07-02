const profileService = require('../../../../services/profile')
const { navigateShellRoute } = require('../../../../utils/shell-nav')

const ASSET_BASE = '/pages/profile/system-management/agreement-sign/assets'

const ICON_MAP = {
  doc: `${ASSET_BASE}/icon-agreement-doc.svg`,
  lock: `${ASSET_BASE}/icon-agreement-lock.svg`,
  box: `${ASSET_BASE}/icon-agreement-box.svg`
}

function normalizeAgreement(item = {}, index = 0, total = 0) {
  return Object.assign({}, item, {
    key: item.key || `agreement-${index}`,
    title: item.title || '平台协议',
    desc: item.desc || '协议说明',
    signed: Boolean(item.signed) && !Boolean(item.requiresResign),
    requiresResign: Boolean(item.requiresResign),
    version: item.version || '',
    tone: item.tone || 'green',
    last: typeof item.last === 'boolean' ? item.last : index === total - 1,
    icon: item.icon || ICON_MAP[item.iconKey] || ICON_MAP.doc
  })
}

Page({
  data: {
    agreements: [],
    icons: {
      check: `${ASSET_BASE}/icon-agreement-check.svg`,
      arrow: `${ASSET_BASE}/icon-agreement-arrow.svg`
    }
  },

  onLoad() {
    this.loadAgreements()
  },

  onShow() {
    this.loadAgreements()
  },

  async loadAgreements() {
    try {
      const result = await profileService.getProfileAgreements()
      const items = Array.isArray(result && result.items) ? result.items : []

      this.setData({
        agreements: items.map((item, index) => normalizeAgreement(item, index, items.length))
      })
    } catch (error) {
      this.showToast(error.message || '协议列表加载失败')
    }
  },

  handleAgreementTap(event) {
    const { key } = event.currentTarget.dataset
    const agreement = this.data.agreements.find((item) => item.key === key)

    if (!agreement) {
      return
    }

    const query = [
      `agreement=${encodeURIComponent(agreement.key)}`,
      `title=${encodeURIComponent(agreement.title)}`,
      `signed=${agreement.signed ? 1 : 0}`,
      `requiresResign=${agreement.requiresResign ? 1 : 0}`
    ].join('&')

    navigateShellRoute(`/pages/profile/system-management/agreement-detail/index?${query}`)
  },

  showToast(title) {
    if (typeof wx === 'undefined') {
      return
    }

    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
