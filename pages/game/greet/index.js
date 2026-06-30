const { ROUTES } = require('../../../config/routes')
const { getSurnameInitials } = require('../../../utils/avatar')

const DEFAULT_CONTACT = {
  name: '王引荐',
  realName: '',
  nickname: '王引荐',
  avatarText: 'WA'
}

function decodeQueryText(value = '') {
  try {
    return decodeURIComponent(value)
  } catch (error) {
    return value
  }
}

function getAvatarText(name = '') {
  return getSurnameInitials(name, DEFAULT_CONTACT.avatarText)
}

function getDisplayName(contact) {
  return contact.realName || contact.name || contact.nickname || DEFAULT_CONTACT.name
}

Page({
  data: {
    onlineText: '3999人在线',
    pageTitle: DEFAULT_CONTACT.name,
    contact: DEFAULT_CONTACT,
    groupInfo: {
      headerTitle: '组局信息',
      title: '产品架构梳理咨询',
      guideLabel: '领路人',
      guideName: DEFAULT_CONTACT.name,
      expertLabel: '行家',
      miniProgramText: '小程序 · 真好玩',
      expert: {
        name: '张专家',
        avatarText: getSurnameInitials('张专家', 'ZH'),
        desc: '资深产品经理 · 10年经验',
        intro: '擅长产品架构设计、MVP规划，10年大厂经验，服务过50+企业客户...',
        tags: ['产品咨询', '架构梳理']
      },
      stats: [
        {
          label: '4.9分',
          icon: '/pages/game/greet/assets/rating.png'
        },
        {
          label: '已认证',
          icon: '/pages/game/greet/assets/verified.png'
        },
        {
          label: '¥800',
          icon: '/pages/game/greet/assets/price.png'
        }
      ],
      confirmText: '查看详情并确认'
    },
    messageText: '李明，这位张专家是我认识的产品大牛，正好符合你之前说的产品架构咨询需求，我帮你们牵个线！',
    sentMessages: [],
    canvasMinHeight: 1280
  },

  onLoad(options = {}) {
    const realName = decodeQueryText(options.realName || '')
    const name = decodeQueryText(options.playerName || options.name || '')
    const nickname = decodeQueryText(options.nickname || '')
    const contactName = realName || name || nickname

    if (!contactName) {
      return
    }

    const contact = {
      ...this.data.contact,
      realName,
      name: contactName,
      nickname,
      avatarText: getAvatarText(contactName)
    }

    this.setData({
      contact,
      pageTitle: getDisplayName(contact),
      'groupInfo.guideName': getDisplayName(contact)
    })

    if (options.message) {
      this.setData({
        messageText: decodeQueryText(options.message)
      })
    }
  },

  onConfirmTap() {
    wx.showToast({
      title: '已确认组局信息',
      icon: 'none'
    })
  },

  onSendMessage(event) {
    const value = String(event.detail && event.detail.value || '').trim()

    if (!value) {
      return
    }

    const sentMessages = this.data.sentMessages.concat({
      id: `message-${Date.now()}`,
      text: value
    })

    this.setData({
      sentMessages,
      canvasMinHeight: 1280 + sentMessages.length * 120
    })
  },

  onRecordStart() {
    this.showInfo('开始录音')
  },

  onRecordStop() {
    this.showInfo('录音发送功能待接入')
  },

  onRecordError() {
    this.showInfo('录音失败')
  },

  onChooseImage() {
    this.showInfo('图片发送功能待接入')
  },

  onChooseFile() {
    this.showInfo('文件发送功能待接入')
  },

  handleShellNavTap(event) {
    const key = event.detail && event.detail.key

    if (key === 'map') {
      wx.showToast({
        title: '地图功能开发中',
        icon: 'none'
      })
      return
    }

    if (key === 'avatar' || key === 'mine') {
      this.navigateToRoute(ROUTES.profile)
      return
    }

    if (key === 'comment' || key === 'message') {
      this.navigateToRoute(ROUTES.message)
      return
    }

    const routeMap = {
      home: ROUTES.home,
      metaverse: ROUTES.metaverse,
      map: ''
    }

    this.navigateToRoute(routeMap[key])
  },

  navigateToRoute(route) {
    if (!route || route === ROUTES.gameGreet) {
      return
    }

    wx.navigateTo({
      url: `/${route}`
    })
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
