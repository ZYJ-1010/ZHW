const ASSET_BASE = '/pages/profile/system-management/agreement-sign/assets'
const STORAGE_KEY = 'profileAgreementSignedMap'

const DEFAULT_AGREEMENTS = [
  {
    key: 'user-service',
    title: '用户服务协议',
    desc: '平台服务条款与规则',
    signed: true,
    tone: 'green',
    last: false,
    icon: `${ASSET_BASE}/icon-agreement-doc.svg`
  },
  {
    key: 'privacy',
    title: '隐私政策',
    desc: '个人信息保护说明',
    signed: true,
    tone: 'deep-green',
    last: false,
    icon: `${ASSET_BASE}/icon-agreement-lock.svg`
  },
  {
    key: 'settlement',
    title: '入驻协议',
    desc: '服务与分润协议',
    signed: false,
    tone: 'orange',
    last: true,
    icon: `${ASSET_BASE}/icon-agreement-box.svg`
  }
]

function getSignedMap() {
  try {
    return wx.getStorageSync(STORAGE_KEY) || {}
  } catch (error) {
    return {}
  }
}

function getAgreements() {
  const signedMap = getSignedMap()

  return DEFAULT_AGREEMENTS.map((item) => Object.assign({}, item, {
    signed: Object.prototype.hasOwnProperty.call(signedMap, item.key) ? Boolean(signedMap[item.key]) : item.signed
  }))
}

Page({
  data: {
    agreements: getAgreements(),
    icons: {
      check: `${ASSET_BASE}/icon-agreement-check.svg`,
      arrow: `${ASSET_BASE}/icon-agreement-arrow.svg`
    }
  },

  onShow() {
    this.setData({
      agreements: getAgreements()
    })
  },

  handleAgreementTap(event) {
    const { key } = event.currentTarget.dataset
    const agreement = this.data.agreements.find((item) => item.key === key) || this.data.agreements[0]
    const query = [
      `agreement=${agreement.key}`,
      `title=${encodeURIComponent(agreement.title)}`,
      `signed=${agreement.signed ? 1 : 0}`
    ].join('&')

    wx.navigateTo({
      url: `/pages/profile/system-management/agreement-detail/index?${query}`
    })
  }
})
