const gameService = require('../../../services/game')

function normalizeRole(role) {
  const normalized = String(role || '').trim()
  const roleAliases = {
    master: 'expert',
    specialist: 'expert',
    leader: 'guide',
    referrer: 'guide'
  }

  return roleAliases[normalized] || normalized
}

function normalizeSatisfactionOption(option = {}) {
  return {
    id: option.id || option.key || '',
    emoji: option.emoji || '',
    title: option.title || option.name || '',
    desc: option.desc || option.description || ''
  }
}

function normalizeTag(tag) {
  if (typeof tag === 'string') {
    return {
      label: tag,
      selected: false
    }
  }

  return {
    label: tag.label || tag.name || '',
    selected: Boolean(tag.selected)
  }
}

function normalizeEvaluationSection(section = {}) {
  return {
    id: section.id || section.targetType || section.key || '',
    avatarText: section.avatarText || '',
    avatarTheme: section.avatarTheme || section.avatarClass || '',
    title: section.title || '',
    desc: section.desc || section.description || '',
    ratingTitle: section.ratingTitle || '',
    tagTitle: section.tagTitle || '',
    tags: Array.isArray(section.tags) ? section.tags.map(normalizeTag).filter((item) => item.label) : [],
    placeholder: section.placeholder || '',
    score: Number(section.score || 0),
    comment: section.comment || ''
  }
}

function normalizeReviewConfig(data = {}, viewerRole = '') {
  const satisfactionOptions = Array.isArray(data.satisfactionOptions)
    ? data.satisfactionOptions.map(normalizeSatisfactionOption).filter((item) => item.id)
    : []
  const evaluationSections = Array.isArray(data.evaluationSections || data.sections)
    ? (data.evaluationSections || data.sections).map(normalizeEvaluationSection).filter((item) => item.id)
    : []

  return {
    satisfactionOptions,
    selectedSatisfaction: data.selectedSatisfaction || data.defaultSatisfaction || satisfactionOptions[0] && satisfactionOptions[0].id || '',
    storyText: data.storyText || '',
    storyLength: String(data.storyText || '').length,
    storyMaxLength: Number(data.storyMaxLength || 100),
    stars: Array.isArray(data.stars) ? data.stars : [1, 2, 3, 4, 5],
    viewerRole: data.viewerRole || viewerRole,
    evaluationSections,
    npsScores: Array.isArray(data.npsScores) ? data.npsScores : [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
    npsScore: Number(data.npsScore || data.defaultNpsScore || 0),
    rewardText: data.rewardText || '',
    queryContext: data.queryContext || {}
  }
}

Page({
  data: {
    satisfactionOptions: [],
    selectedSatisfaction: '',
    storyText: '',
    storyLength: 0,
    storyMaxLength: 100,
    stars: [1, 2, 3, 4, 5],
    viewerRole: '',
    evaluationSections: [],
    npsScores: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
    npsScore: 0,
    rewardText: '',
    queryParams: {},
    queryContext: {},
    loading: false,
    submitting: false
  },

  onLoad(options = {}) {
    const viewerRole = normalizeRole(options.role)

    this.setData({
      queryParams: options,
      viewerRole
    })
    this.loadReviewConfig({
      ...options,
      role: viewerRole
    })
  },

  async loadReviewConfig(params = {}) {
    this.setData({
      loading: true
    })

    try {
      const data = await gameService.getGameReviewConfig(params)

      this.setData({
        ...normalizeReviewConfig(data, this.data.viewerRole),
        loading: false
      })
    } catch (error) {
      this.setData({
        ...normalizeReviewConfig({}, this.data.viewerRole),
        loading: false
      })
      this.showInfo(error.message || '评价配置加载失败')
    }
  },

  onSatisfactionTap(event) {
    const id = event.currentTarget.dataset.id

    if (!id) {
      return
    }

    this.setData({
      selectedSatisfaction: id
    })
  },

  onStoryInput(event) {
    const value = event.detail.value || ''

    this.setData({
      storyText: value,
      storyLength: value.length
    })
  },

  onAiSummaryTap() {
    this.showInfo('AI总结功能待接入')
  },

  onStarTap(event) {
    const sectionId = event.currentTarget.dataset.sectionId
    const score = Number(event.currentTarget.dataset.score) || 0

    this.updateSection(sectionId, (section) => ({
      ...section,
      score
    }))
  },

  onTagTap(event) {
    const sectionId = event.currentTarget.dataset.sectionId
    const tagIndex = Number(event.currentTarget.dataset.tagIndex)

    if (Number.isNaN(tagIndex)) {
      return
    }

    this.updateSection(sectionId, (section) => ({
      ...section,
      tags: section.tags.map((tag, index) => index === tagIndex
        ? { ...tag, selected: !tag.selected }
        : tag)
    }))
  },

  onEvaluationCommentInput(event) {
    const sectionId = event.currentTarget.dataset.sectionId
    const value = event.detail.value || ''

    this.updateSection(sectionId, (section) => ({
      ...section,
      comment: value
    }))
  },

  updateSection(sectionId, updater) {
    if (!sectionId || typeof updater !== 'function') {
      return
    }

    this.setData({
      evaluationSections: this.data.evaluationSections.map((section) => (
        section.id === sectionId ? updater(section) : section
      ))
    })
  },

  onNpsTap(event) {
    const score = Number(event.currentTarget.dataset.score)

    if (Number.isNaN(score)) {
      return
    }

    this.setData({
      npsScore: score
    })
  },

  onBackTap() {
    const pages = getCurrentPages()

    if (pages.length > 1) {
      wx.navigateBack()
      return
    }

    this.showInfo('返回入口待接入')
  },

  onSkipTap() {
    this.showInfo('已跳过评价')
  },

  async onSubmitTap() {
    if (this.data.submitting) {
      return
    }

    this.setData({
      submitting: true
    })

    try {
      await gameService.submitGameReview({
        ...this.data.queryParams,
        ...this.data.queryContext,
        role: this.data.viewerRole,
        satisfaction: this.data.selectedSatisfaction,
        storyText: this.data.storyText,
        npsScore: this.data.npsScore,
        evaluations: this.data.evaluationSections.map((section) => ({
          id: section.id,
          score: section.score,
          tags: section.tags.filter((tag) => tag.selected).map((tag) => tag.label),
          comment: section.comment
        }))
      })
      this.showInfo('评价已提交')
    } catch (error) {
      this.showInfo(error.message || '提交评价失败')
    } finally {
      this.setData({
        submitting: false
      })
    }
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
