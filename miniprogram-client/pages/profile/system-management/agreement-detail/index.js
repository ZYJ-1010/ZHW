const profileService = require('../../../../services/profile')
const { navigateShellRoute } = require('../../../../utils/shell-nav')
const PUBLIC_AGREEMENTS = require('./public-agreements')

const ASSET_BASE = '/pages/profile/system-management/agreement-detail/assets'

function getTitle(options = {}) {
  if (options.title) {
    return decodeURIComponent(options.title)
  }

  return ''
}

function getPublicAgreement(agreementKey, title) {
  const fallback = PUBLIC_AGREEMENTS[agreementKey]

  if (!fallback) {
    return null
  }

  return {
    title: fallback.title || title,
    sections: fallback.sections,
    signed: true
  }
}

function normalizeSignConfirm(confirm = {}) {
  return {
    title: confirm.title || '',
    desc: confirm.desc || '',
    cancelText: confirm.cancelText || '',
    confirmText: confirm.confirmText || ''
  }
}

Page({
  data: {
    agreementKey: 'user-service',
    title: '',
    sections: [],
    canSign: false,
    signActionText: '',
    signSuccessText: '',
    signConfirm: normalizeSignConfirm(),
    confirmVisible: false,
    isSigning: false,
    dialogIcon: `${ASSET_BASE}/icon-agreement-file.svg`
  },

  onLoad(options = {}) {
    const agreementKey = options.agreement || 'user-service'
    const hasSignedOption = options.signed === '0' || options.signed === '1'
    const fallbackSigned = hasSignedOption ? options.signed === '1' : false

    this.setData({
      agreementKey,
      title: getTitle(options),
      canSign: !fallbackSigned,
      confirmVisible: options.confirm === '1'
    })
    this.loadAgreementDetail(agreementKey)
  },

  async loadAgreementDetail(agreementKey) {
    try {
      const detail = await profileService.getProfileAgreementDetail(agreementKey)
      const sections = Array.isArray(detail && detail.sections) ? detail.sections : []

      this.setData({
        title: detail.title || this.data.title,
        sections,
        canSign: !Boolean(detail.signed),
        signActionText: detail.signActionText || this.data.signActionText,
        signSuccessText: detail.signSuccessText || this.data.signSuccessText,
        signConfirm: normalizeSignConfirm(detail.signConfirm || {})
      })
    } catch (error) {
      const fallback = getPublicAgreement(agreementKey, this.data.title)
      if (fallback) {
        this.setData({
          title: fallback.title,
          sections: fallback.sections,
          canSign: false,
          signActionText: '',
          signSuccessText: '',
          signConfirm: normalizeSignConfirm()
        })
        return
      }

      this.showToast(error.message || '协议详情加载失败')
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
    if (this.data.isSigning) {
      return
    }

    this.setData({ isSigning: true })

    try {
      await profileService.signProfileAgreement(this.data.agreementKey)
      this.setData({
        canSign: false,
        confirmVisible: false
      })
      wx.showToast({
        title: this.data.signSuccessText,
        icon: 'success'
      })
      setTimeout(() => {
        const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []

        if (pages.length > 1) {
          wx.navigateBack()
          return
        }

        navigateShellRoute('/pages/profile/system-management/agreement-sign/index')
      }, 500)
    } catch (error) {
      this.showToast(error.message || '签署失败，请稍后再试')
    } finally {
      this.setData({ isSigning: false })
    }
  },

  showToast(title) {
    if (typeof wx === 'undefined') {
      return
    }

    wx.showToast({
      title,
      icon: 'none'
    })
  },

  noop() {}
})
