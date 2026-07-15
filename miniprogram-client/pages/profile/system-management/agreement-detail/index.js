const profileService = require('../../../../services/profile')
const { navigateShellRoute } = require('../../../../utils/shell-nav')

const ASSET_BASE = '/pages/profile/system-management/agreement-detail/assets'

const PUBLIC_AGREEMENTS = {
  'user-service': {
    title: '用户服务协议',
    sections: [
      {
        title: '一、服务范围',
        content: '真好玩为受邀用户和已注册用户提供组局、报名、消息通知、资料管理、评价成长等服务。用户进入和使用本小程序，应遵守平台规则及相关法律法规。'
      },
      {
        title: '二、账号与登录鉴权',
        content: '本小程序属于特定服务人群使用。已注册账号可通过手机号验证码、微信登录或密码登录完成身份鉴权；新用户需持有效邀请码完成注册。用户应妥善保管账号信息，不得转让、出借或冒用他人账号。'
      },
      {
        title: '三、用户行为规范',
        content: '用户不得发布违法违规、侵权、虚假、骚扰、欺诈、低俗或其他影响平台秩序的内容。平台可根据运营规则对违规内容、账号、组局或交易行为进行处理。'
      },
      {
        title: '四、服务变更与风险提示',
        content: '平台可能根据业务、安全、审核或合规要求调整功能范围、服务规则或访问权限。用户在参与组局、报名、沟通和交易前，应自行确认活动信息并承担相应风险。'
      }
    ]
  },
  privacy: {
    title: '隐私政策',
    sections: [
      {
        title: '一、信息收集',
        content: '为实现登录鉴权、邀请注册、实名认证、组局报名、消息通知、头像资料、位置选择和文件上传等功能，平台会在必要范围内收集手机号、微信身份标识、昵称头像、实名信息、位置信息及用户主动提交的内容。'
      },
      {
        title: '二、信息使用',
        content: '平台仅将收集的信息用于账号识别、身份审核、服务履约、安全风控、内容审核、消息通知、用户资料展示及法律法规要求的场景，不会超出实现功能所必要的范围使用。'
      },
      {
        title: '三、信息保护',
        content: '平台会采取合理的技术和管理措施保护用户信息安全。涉及实名认证、联系方式、头像审核等敏感信息时，将按业务权限和审核流程控制访问与展示。'
      },
      {
        title: '四、用户权利',
        content: '用户可在小程序内查看、修改个人资料，按平台流程重新提交实名认证或头像审核。如需了解、撤回授权或处理个人信息相关问题，可通过平台客服或后台留存渠道联系处理。'
      }
    ]
  }
}

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
    title: title || fallback.title,
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
