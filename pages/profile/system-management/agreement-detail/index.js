const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

const ASSET_BASE = '/pages/profile/system-management/agreement-detail/assets'

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

function normalizeSections(data) {
  const source = data || {}
  const sections = normalizeList(source.sections || source.contentSections || source.clauses)

  if (sections.length) {
    return sections.map((item) => ({
      title: pickFirstValue(item.title, item.heading),
      content: pickFirstValue(item.content, item.body, item.text)
    })).filter((item) => item.title || item.content)
  }

  const content = pickFirstValue(source.content, source.body, source.text)

  if (content) {
    return [{
      title: '',
      content
    }]
  }

  return []
}

function isSigned(data) {
  const source = data || {}
  const status = String(source.status || '').toLowerCase()

  return Boolean(source.signed || source.signedAt || status === 'signed')
}

Page({
  data: {
    agreementKey: '',
    title: '',
    sections: [],
    canSign: false,
    confirmVisible: false,
    dialogIcon: `${ASSET_BASE}/icon-agreement-file.svg`
  },

  onLoad(options = {}) {
    const agreementKey = options.agreement || options.agreementId || options.id || ''

    this.setData({
      agreementKey,
      title: options.title ? decodeURIComponent(options.title) : '',
      canSign: false,
      confirmVisible: false
    })

    this.loadAgreementDetail({
      agreementKey,
      showConfirm: options.confirm === '1'
    })
  },

  async loadAgreementDetail(options = {}) {
    const agreementKey = options.agreementKey || this.data.agreementKey

    if (!agreementKey) {
      toast.info('缺少协议信息')
      return
    }

    try {
      const detail = await profileService.getSystemAgreementDetail({
        agreementId: agreementKey
      })
      const source = detail || {}
      const signed = isSigned(source)
      const canSign = source.canSign !== false && !signed

      this.setData({
        title: pickFirstValue(source.title, source.name, this.data.title),
        sections: normalizeSections(source),
        canSign,
        confirmVisible: options.showConfirm && canSign
      })
    } catch (error) {
      this.setData({
        sections: [],
        canSign: false,
        confirmVisible: false
      })
      toast.info(error.message || '协议详情加载失败')
    }
  },

  handleSignTap() {
    this.setData({
      confirmVisible: true
    })
  },

  handleCancelTap() {
    this.setData({
      confirmVisible: false
    })
  },

  async handleConfirmTap() {
    try {
      await profileService.signSystemAgreement({
        agreementId: this.data.agreementKey
      })

      this.setData({
        canSign: false,
        confirmVisible: false
      })

      wx.showToast({
        title: '签署成功',
        icon: 'success'
      })

      setTimeout(() => {
        const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []

        if (pages.length > 1) {
          wx.navigateBack()
          return
        }

        wx.redirectTo({
          url: '/pages/profile/system-management/agreement-sign/index'
        })
      }, 500)
    } catch (error) {
      toast.info(error.message || '协议签署失败')
    }
  },

  noop() {}
})
