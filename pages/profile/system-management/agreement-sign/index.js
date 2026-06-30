const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

const ASSET_BASE = '/pages/profile/system-management/agreement-sign/assets'

const ROW_META = [
  { tone: 'green', icon: `${ASSET_BASE}/icon-agreement-doc.svg` },
  { tone: 'deep-green', icon: `${ASSET_BASE}/icon-agreement-lock.svg` },
  { tone: 'orange', icon: `${ASSET_BASE}/icon-agreement-box.svg` }
]

function pickFirstValue() {
  const values = Array.prototype.slice.call(arguments)

  for (let index = 0; index < values.length; index += 1) {
    if (values[index] !== undefined && values[index] !== null && values[index] !== '') {
      return values[index]
    }
  }

  return ''
}

function normalizeList(list) {
  return Array.isArray(list) ? list : []
}

function normalizeAgreements(data) {
  const source = data || {}
  const list = normalizeList(source.agreements || source.list || source.items)

  return list.map((item, index) => {
    const agreement = item || {}
    const meta = ROW_META[index % ROW_META.length]
    const key = pickFirstValue(agreement.key, agreement.agreementId, agreement.id)
    const status = String(agreement.status || '').toLowerCase()

    return {
      key,
      title: pickFirstValue(agreement.title, agreement.name),
      desc: pickFirstValue(agreement.desc, agreement.description, agreement.summary),
      signed: Boolean(agreement.signed || agreement.signedAt || status === 'signed'),
      tone: pickFirstValue(agreement.tone, agreement.iconTone, meta.tone),
      last: index === list.length - 1,
      icon: pickFirstValue(agreement.localIcon, meta.icon)
    }
  }).filter((item) => item.key)
}

Page({
  data: {
    agreements: [],
    icons: {
      check: `${ASSET_BASE}/icon-agreement-check.svg`,
      arrow: `${ASSET_BASE}/icon-agreement-arrow.svg`
    }
  },

  onShow() {
    this.loadAgreements()
  },

  async loadAgreements() {
    try {
      const data = await profileService.getSystemAgreements()

      this.setData({
        agreements: normalizeAgreements(data)
      })
    } catch (error) {
      this.setData({
        agreements: []
      })
      toast.info(error.message || '协议列表加载失败')
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
      `signed=${agreement.signed ? 1 : 0}`
    ].join('&')

    wx.navigateTo({
      url: `/pages/profile/system-management/agreement-detail/index?${query}`
    })
  }
})
