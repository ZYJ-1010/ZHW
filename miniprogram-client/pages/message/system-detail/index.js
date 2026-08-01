const { ROUTES } = require('../../../config/routes')
const messageService = require('../../../services/message')
const { navigateShellKey } = require('../../../utils/shell-nav')
const { toUserMessage } = require('../../../utils/user-message')

const EMPTY_SYSTEM_NOTIFICATION_DETAIL = {
  pageTitle: '',
  onlineText: '',
  article: {
    tagText: '',
    title: '',
    author: '',
    publishedAtText: '',
    readText: '',
    blocks: []
  },
  feedback: {
    question: '',
    useful: { icon: '', label: '', countText: '0' },
    useless: { icon: '', label: '', countText: '0' }
  },
  texts: {}
}

function applyTemplate(template, values = {}) {
  return String(template || '').replace(/\{(\w+)\}/g, (_, key) => values[key] == null ? '' : values[key])
}

function textOf(config, key) {
  const texts = config && config.texts ? config.texts : {}
  return texts[key] || ''
}

function normalizeSystemNotificationDetail(data = {}) {
  const source = Object.assign({}, EMPTY_SYSTEM_NOTIFICATION_DETAIL, data)
  const article = Object.assign({}, EMPTY_SYSTEM_NOTIFICATION_DETAIL.article, data.article || {})
  const feedback = normalizeFeedback(data.feedback)

  return {
    pageTitle: source.pageTitle || '',
    onlineText: source.onlineText || '',
    article: Object.assign({}, article, {
      blocks: normalizeArticleBlocks(article.blocks)
    }),
    feedback,
    texts: source.texts || {},
    messageId: source.messageId || source.notificationId || source.id || '',
    loading: false,
    errorText: ''
  }
}

function normalizeArticleBlocks(blocks) {
  return Array.isArray(blocks) ? blocks : []
}

function normalizeFeedback(feedback = {}) {
  const useful = Object.assign({}, EMPTY_SYSTEM_NOTIFICATION_DETAIL.feedback.useful, feedback.useful || {})
  const useless = Object.assign({}, EMPTY_SYSTEM_NOTIFICATION_DETAIL.feedback.useless, feedback.useless || {})

  return {
    question: feedback.question || EMPTY_SYSTEM_NOTIFICATION_DETAIL.feedback.question,
    useful: Object.assign({}, useful, {
      countText: useful.countText || String(useful.count || 0)
    }),
    useless: Object.assign({}, useless, {
      countText: useless.countText || String(useless.count || 0)
    })
  }
}

Page({
  data: {
    pageTitle: '',
    onlineText: '',
    navItems: [
      { name: '我的', key: 'mine' },
      { name: '元宇宙', key: 'metaverse' },
      { name: '地图', key: 'map' },
      { name: '消息', key: 'message' },
      { name: '首页', key: 'home' }
    ],
    article: EMPTY_SYSTEM_NOTIFICATION_DETAIL.article,
    feedback: EMPTY_SYSTEM_NOTIFICATION_DETAIL.feedback,
    texts: EMPTY_SYSTEM_NOTIFICATION_DETAIL.texts,
    messageId: '',
    loading: false,
    feedbackSubmitting: false,
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
      const errorText = toUserMessage(error && error.message, textOf(this.data, 'loadFailedText') || '系统通知加载失败')
      this.setData(Object.assign({}, normalizeSystemNotificationDetail(EMPTY_SYSTEM_NOTIFICATION_DETAIL), {
        errorText
      }))
      this.showInfo(errorText)
    }
  },

  async onFeedbackTap(event) {
    if (this.data.feedbackSubmitting) {
      return
    }

    const { value } = event.currentTarget.dataset
    const feedback = value === 'useful' ? this.data.feedback.useful : this.data.feedback.useless

    if (value !== 'useful' && value !== 'useless') {
      return
    }

    this.setData({ feedbackSubmitting: true })

    try {
      const data = await messageService.submitSystemNotificationFeedback({
        messageId: this.data.messageId,
        value
      })

      if (data.feedback) {
        this.setData({ feedback: normalizeFeedback(data.feedback) })
      }

      this.showInfo(applyTemplate(textOf(this.data, 'feedbackSuccessText'), { label: feedback.label }))
    } catch (error) {
      this.showInfo(error.message || textOf(this.data, 'feedbackFailedText'))
    } finally {
      this.setData({ feedbackSubmitting: false })
    }
  },

  handleShellNavTap(event) {
    const { key } = event.detail || {}

    navigateShellKey(key, {
      currentRoute: ROUTES.messageSystemDetail
    })
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
