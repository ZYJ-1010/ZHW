const { ROUTES } = require('../../../config/routes')
const messageService = require('../../../services/message')

const DEFAULT_SYSTEM_NOTIFICATION_DETAIL = {
  pageTitle: '系统通知',
  onlineText: '',
  article: {},
  feedback: {
    question: '',
    useful: {},
    useless: {}
  }
}

function normalizeSystemNotificationDetail(data = {}) {
  const source = Object.assign({}, DEFAULT_SYSTEM_NOTIFICATION_DETAIL, data)
  const article = Object.assign({}, DEFAULT_SYSTEM_NOTIFICATION_DETAIL.article, data.article || {})
  const feedback = normalizeFeedback(data.feedback)
  const blocks = normalizeArticleBlocks(article.blocks)

  return {
    pageTitle: source.pageTitle || DEFAULT_SYSTEM_NOTIFICATION_DETAIL.pageTitle,
    onlineText: source.onlineText || DEFAULT_SYSTEM_NOTIFICATION_DETAIL.onlineText,
    article: Object.assign({}, article, {
      blocks
    }),
    feedback,
    messageId: source.messageId || source.notificationId || source.id || '',
    hasArticle: Boolean(article.title || blocks.length),
    hasFeedback: Boolean(feedback.question),
    loading: false,
    errorText: ''
  }
}

function normalizeArticleBlocks(blocks) {
  return Array.isArray(blocks) ? blocks : []
}

function normalizeFeedback(feedback = {}) {
  const useful = Object.assign({}, DEFAULT_SYSTEM_NOTIFICATION_DETAIL.feedback.useful, feedback.useful || {})
  const useless = Object.assign({}, DEFAULT_SYSTEM_NOTIFICATION_DETAIL.feedback.useless, feedback.useless || {})

  return {
    question: feedback.question || DEFAULT_SYSTEM_NOTIFICATION_DETAIL.feedback.question,
    useful: Object.assign({}, useful, {
      countText: useful.countText || (useful.count != null ? String(useful.count) : '')
    }),
    useless: Object.assign({}, useless, {
      countText: useless.countText || (useless.count != null ? String(useless.count) : '')
    })
  }
}

Page({
  data: {
    pageTitle: '系统通知',
    onlineText: '3999人在线',
    navItems: [
      { name: '我的', key: 'mine' },
      { name: '元宇宙', key: 'metaverse' },
      { name: '地图', key: 'map' },
      { name: '消息', key: 'message' },
      { name: '首页', key: 'home' }
    ],
    article: DEFAULT_SYSTEM_NOTIFICATION_DETAIL.article,
    feedback: DEFAULT_SYSTEM_NOTIFICATION_DETAIL.feedback,
    messageId: '',
    hasArticle: false,
    hasFeedback: false,
    loading: false,
    errorText: ''
  },

  onLoad(options = {}) {
    this.loadSystemNotificationDetail(options)
  },

  async loadSystemNotificationDetail(options = {}) {
    this.setData({
      loading: true,
      errorText: ''
    })

    try {
      const detail = await messageService.getSystemNotificationDetail({
        messageId: options.messageId || options.notificationId || options.id || ''
      })

      this.setData(normalizeSystemNotificationDetail(detail))
    } catch (error) {
      const errorText = error && error.message ? error.message : '获取系统通知失败'

      this.setData(Object.assign({}, normalizeSystemNotificationDetail(), {
        errorText
      }))
      this.showInfo(errorText)
    }
  },

  onFeedbackTap(event) {
    if (!this.data.hasFeedback) {
      return
    }

    const { value } = event.currentTarget.dataset
    const feedback = value === 'useful' ? this.data.feedback.useful : this.data.feedback.useless
    this.showInfo(`已记录${feedback.label}反馈`)
  },

  handleShellNavTap(event) {
    const { key } = event.detail || {}

    if (key === 'map') {
      wx.showToast({
        title: '地图功能开发中',
        icon: 'none'
      })
      return
    }
    const routeMap = {
      home: ROUTES.playerHome || ROUTES.home,
      map: '',
      message: ROUTES.message,
      mine: ROUTES.profile,
      avatar: ROUTES.profile,
      metaverse: ROUTES.metaverse
    }
    const route = routeMap[key]

    if (!route || route === ROUTES.messageSystemDetail) {
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
