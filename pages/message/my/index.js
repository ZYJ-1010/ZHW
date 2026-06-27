const { ROUTES } = require('../../../config/routes')

const QUICK_ACTIONS = [
  { key: 'friend', label: '加好友' },
  { key: 'greet', label: '打招呼' },
  { key: 'card', label: '发名片' },
  { key: 'location', label: '发定位' }
]

const MESSAGES = [
  {
    id: 'm1',
    type: 'friend',
    text: '好的，那我们就周六下午2点在咖啡店见，我带上项目资料...',
    timeText: '12:30'
  }
]

Page({
  data: {
    pageTitle: '好友消息',
    onlineText: '3999人在线',
    navItems: [
      { name: '我的', key: 'mine' },
      { name: '元宇宙', key: 'metaverse' },
      { name: '地图', key: 'map' },
      { name: '消息', key: 'message' },
      { name: '首页', key: 'home' }
    ],
    friend: {
      initials: 'LM',
      name: '李明',
      status: '在线'
    },
    quickActions: QUICK_ACTIONS,
    messages: MESSAGES,
    chatScrollTop: 0
  },

  onQuickActionTap(event) {
    const action = QUICK_ACTIONS.find((item) => item.key === event.currentTarget.dataset.key)
    this.showInfo(action ? `${action.label}待接入` : '操作待接入')
  },

  onSendMessage(event) {
    const value = String(event.detail && event.detail.value || '').trim()

    if (!value) {
      return
    }

    this.setData({
      messages: this.data.messages.concat({
        id: `m-${Date.now()}`,
        type: 'self',
        text: value,
        timeText: '刚刚'
      }),
      chatScrollTop: 999999
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
    const { key } = event.detail || {}
    const routeMap = {
      home: ROUTES.playerHome || ROUTES.home,
      map: ROUTES.map,
      message: ROUTES.message,
      mine: ROUTES.profile,
      avatar: ROUTES.profile,
      metaverse: ROUTES.metaverse
    }
    const route = routeMap[key]

    if (!route || route === ROUTES.messageMy) {
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
