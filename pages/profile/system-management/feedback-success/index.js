const toast = require('../../../../utils/toast')
const { navigateShellRoute } = require('../../../../utils/shell-nav')

const ASSET_BASE = '/pages/profile/system-management/feedback/assets'
const SUCCESS_PAGE_STORAGE_KEY = 'enjoy_feedback_success_page'

function buildScores(config = {}) {
  const start = Number(config.scoreFrom || 0)
  const end = Number(config.scoreTo || 10)
  const scores = []

  for (let value = start; value <= end; value += 1) {
    scores.push(value)
  }

  return scores
}

function normalizeReasons(reasons = []) {
  if (!Array.isArray(reasons)) {
    return []
  }

  return reasons
    .map((item) => ({
      value: item.value || item.label || item.text || '',
      selected: Boolean(item.selected)
    }))
    .filter((item) => item.value)
}

function reportSuccessConfig() {
  return {
    successTitle: '举报已提交',
    successDesc: '平台已收到您的举报材料，将尽快核实处理',
    successDescSecond: '处理进度会同步到处理记录',
    backHomeText: '返回举报中心',
    viewRecordsText: '查看处理记录'
  }
}

Page({
  data: {
    source: 'feedback',
    reportId: '',
    successTitle: '',
    successDesc: '',
    successDescSecond: '',
    backHomeText: '',
    viewRecordsText: '',
    icons: {
      check: `${ASSET_BASE}/icon-check-white.svg`,
      star: `${ASSET_BASE}/icon-star-outline.svg`
    },
    reward: {},
    rating: {},
    score: 0,
    scores: [],
    reasonsTitle: '',
    reasons: [],
    submitRatingText: '',
    ratingSavedText: ''
  },

  onLoad(options = {}) {
    const source = options.source === 'report' ? 'report' : 'feedback'
    const successPage = source === 'report' ? reportSuccessConfig() : this.getSuccessPageConfig()
    const rating = successPage.rating || {}

    this.setData({
      source,
      reportId: options.reportId || '',
      successTitle: successPage.successTitle || '',
      successDesc: successPage.successDesc || '',
      successDescSecond: successPage.successDescSecond || '',
      backHomeText: successPage.backHomeText || '',
      viewRecordsText: successPage.viewRecordsText || '',
      reward: successPage.reward || {},
      rating,
      score: Number(rating.default || 0),
      scores: buildScores(rating),
      reasonsTitle: successPage.reasonsTitle || '',
      reasons: normalizeReasons(successPage.reasons),
      submitRatingText: successPage.submitRatingText || '',
      ratingSavedText: successPage.ratingSavedText || ''
    })
  },

  getSuccessPageConfig() {
    if (typeof wx === 'undefined' || !wx.getStorageSync) {
      return {}
    }

    const config = wx.getStorageSync(SUCCESS_PAGE_STORAGE_KEY) || {}

    if (wx.removeStorageSync) {
      wx.removeStorageSync(SUCCESS_PAGE_STORAGE_KEY)
    }

    return config
  },

  handleScoreTap(event) {
    const { score } = event.currentTarget.dataset

    this.setData({
      score: Number(score)
    })
  },

  handleReasonTap(event) {
    const reasonIndex = Number(event.currentTarget.dataset.index)

    if (Number.isNaN(reasonIndex)) {
      return
    }

    const selected = !this.data.reasons[reasonIndex].selected

    this.setData({
      [`reasons[${reasonIndex}].selected`]: selected
    })
  },

  handleSubmitRating() {
    toast.success(this.data.ratingSavedText || '')
  },

  handleBackHome() {
    navigateShellRoute(this.data.source === 'report'
      ? '/pages/profile/system-management/report-center/index'
      : '/pages/profile/system-management/feedback/index')
  },

  handleViewRecords() {
    if (this.data.source === 'report') {
      navigateShellRoute(this.data.reportId
        ? `/pages/profile/system-management/report-detail/index?reportId=${this.data.reportId}`
        : '/pages/profile/system-management/report-records/index')
      return
    }

    navigateShellRoute('/pages/profile/system-management/feedback-records/index')
  }
})
