const ASSET_BASE = '/pages/profile/system-management/agreement-detail/assets'
const STORAGE_KEY = 'profileAgreementSignedMap'

const AGREEMENT_TITLES = {
  'user-service': '用户服务协议',
  privacy: '隐私政策',
  settlement: '入驻协议'
}

const SECTIONS = [
  {
    title: '一、协议范围',
    content: '本协议是您与本平台之间关于使用平台服务所订立的协议。 请您仔细阅读本协议，如您不同意本协议的任何内容，请停止使用平台服务。'
  },
  {
    title: '二、账号注册',
    content: '您承诺以真实身份注册账号，并保证所提供的个人资料真实、准确、完整、合法有效。如有变动，应及时更新。'
  },
  {
    title: '三、服务内容',
    content: '平台向您提供组局管理、技能展示、社交互动等服务。您有权按照平台规则使用各项服务。'
  },
  {
    title: '四、用户行为规范',
    content: '您在使用平台服务时，应遵守法律法规，不得发布违法违规信息，不得侵犯他人合法权益。'
  },
  {
    title: '五、知识产权',
    content: '平台所有内容，包括但不限于文字、图片、音频、视频、软件等，均受知识产权法律保护。'
  },
  {
    title: '六、免责声明',
    content: '平台不对因不可抗力或第三方原因导致的服务中断承担责任。'
  },
  {
    title: '七、协议变更',
    content: '平台有权根据需要修改本协议，修改后的协议将在平台公示，公示期满即生效。'
  }
]

function getTitle(options = {}) {
  if (options.title) {
    return decodeURIComponent(options.title)
  }

  return AGREEMENT_TITLES[options.agreement] || AGREEMENT_TITLES['user-service']
}

function getSignedMap() {
  try {
    return wx.getStorageSync(STORAGE_KEY) || {}
  } catch (error) {
    return {}
  }
}

function getSignedState(agreementKey, fallbackSigned) {
  const signedMap = getSignedMap()

  if (Object.prototype.hasOwnProperty.call(signedMap, agreementKey)) {
    return Boolean(signedMap[agreementKey])
  }

  return fallbackSigned
}

function saveSignedState(agreementKey) {
  try {
    const signedMap = getSignedMap()
    signedMap[agreementKey] = true
    wx.setStorageSync(STORAGE_KEY, signedMap)
  } catch (error) {
  }
}

Page({
  data: {
    agreementKey: 'user-service',
    title: AGREEMENT_TITLES['user-service'],
    sections: SECTIONS,
    canSign: false,
    confirmVisible: false,
    dialogIcon: `${ASSET_BASE}/icon-agreement-file.svg`
  },

  onLoad(options = {}) {
    const agreementKey = options.agreement || 'user-service'
    const hasSignedOption = options.signed === '0' || options.signed === '1'
    const fallbackSigned = hasSignedOption ? options.signed === '1' : false
    const signed = getSignedState(agreementKey, fallbackSigned)

    this.setData({
      agreementKey,
      title: getTitle(options),
      canSign: !signed,
      confirmVisible: options.confirm === '1'
    })
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

  handleConfirmTap() {
    saveSignedState(this.data.agreementKey)

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
  },

  noop() {}
})
