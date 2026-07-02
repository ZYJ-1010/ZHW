const reviewService = require('../../../services/review')
const { ROUTES } = require('../../../config/routes')
const { navigateShellRoute } = require('../../../utils/shell-nav')

const ROLE_ALIASES = {
  master: 'expert',
  specialist: 'expert',
  leader: 'guide',
  referrer: 'guide'
}

function normalizeRole(role) {
  const normalized = String(role || '').trim()
  if (['expert', 'player', 'guide'].includes(normalized)) {
    return normalized
  }
  return ROLE_ALIASES[normalized] || 'expert'
}

function toPositiveInt(value) {
  const number = Number(value)
  return Number.isInteger(number) && number > 0 ? number : 0
}

function formatDeadline(deadlineAt) {
  if (!deadlineAt) {
    return ''
  }

  const text = String(deadlineAt)
  return text.replace('T', ' ').replace(/:\d{2}(?:\.\d+)?Z?$/, '')
}

function createTagItems(tags) {
  return (Array.isArray(tags) ? tags : []).map((label) => ({
    label,
    selected: false
  }))
}

function createSectionBase(role, index, fallback = {}, roleConfigs = {}) {
  const base = roleConfigs[role] || roleConfigs.expert || {}
  const targetUserId = toPositiveInt(fallback.targetUserId)
  const gameId = toPositiveInt(fallback.gameId)
  const targetLabel = fallback.targetName || fallback.targetNickname || (targetUserId ? `成员 ${targetUserId}` : `待评价对象 ${index + 1}`)

  return {
    id: fallback.id || `review-${gameId || 'draft'}-${targetUserId || index + 1}`,
    gameId,
    targetUserId,
    targetRole: fallback.targetRole || 'member',
    avatarText: fallback.avatarText || String(targetUserId || index + 1).slice(-2).toUpperCase(),
    avatarTheme: fallback.avatarTheme || base.avatarTheme || '',
    title: fallback.title || `评价 ${targetLabel}`,
    desc: fallback.desc || [
      gameId ? `局 ${gameId}` : '',
      fallback.deadlineAt ? `截止 ${formatDeadline(fallback.deadlineAt)}` : ''
    ].filter(Boolean).join(' · ') || '待评价',
    ratingTitle: fallback.ratingTitle || base.ratingTitle || '',
    tagTitle: fallback.tagTitle || base.tagTitle || '',
    tags: createTagItems(fallback.tags || base.tags),
    placeholder: fallback.placeholder || base.placeholder || '',
    score: Number(fallback.score) || 0,
    comment: String(fallback.comment || '').trim()
  }
}

function buildSectionsFromTodos(role, todos, roleConfigs) {
  return (Array.isArray(todos) ? todos : []).map((todo, index) => createSectionBase(role, index, {
    id: todo.id,
    gameId: todo.gameId,
    targetUserId: todo.targetUserId,
    targetRole: todo.targetRole,
    deadlineAt: todo.deadlineAt,
    title: todo.targetName ? `评价 ${todo.targetName}` : '',
    desc: todo.targetRole ? `局 ${todo.gameId} · ${todo.targetRole}` : `局 ${todo.gameId}`
  }, roleConfigs))
}

function resolveAgainIntent(satisfactionId, intentMap = {}) {
  return intentMap[satisfactionId] || 'yes'
}

function buildReviewSummary(text = '', defaultSummary = '') {
  const content = String(text || '').trim().replace(/\s+/g, ' ')

  if (!content) {
    return defaultSummary
  }

  if (content.length <= 60) {
    return content
  }

  return `${content.slice(0, 56)}...`
}

function normalizeReward(reward = {}) {
  return {
    show: Boolean(reward.show),
    title: String(reward.title || ''),
    desc: String(reward.desc || reward.description || '')
  }
}

Page({
  data: {
    navTitle: '',
    skipText: '',
    statusTitle: '',
    statusDesc: '',
    satisfactionQuestion: '',
    satisfactionOptions: [],
    selectedSatisfaction: '',
    storyTitle: '',
    aiTip: '',
    storyPlaceholder: '',
    aiSummaryText: '',
    ratingHint: '',
    npsHeadTitle: '',
    npsQuestion: '',
    npsLowLabel: '',
    npsHighLabel: '',
    submitText: '',
    submitNote: '',
    reviewPage: {},
    storyText: '',
    storyLength: 0,
    storyMaxLength: 100,
    stars: [1, 2, 3, 4, 5],
    viewerRole: 'expert',
    evaluationSections: [],
    npsScores: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
    npsScore: 7,
    loading: false,
    submitting: false,
    pendingCount: 0,
    reward: {
      show: false,
      title: '',
      desc: ''
    }
  },

  onLoad(options = {}) {
    this.options = {
      role: normalizeRole(options.role),
      gameId: toPositiveInt(options.gameId),
      targetUserId: toPositiveInt(options.targetUserId),
      targetRole: 'member',
      targetName: String(options.targetName || options.name || '').trim(),
      targetNickname: String(options.targetNickname || '').trim(),
      title: String(options.title || '').trim()
    }

    this.loadReviewContext()
  },

  async loadReviewContext() {
    const role = this.options.role || 'expert'

    this.setData({
      loading: true,
      viewerRole: role
    })

    try {
      const result = await reviewService.getAvailableReviews({})
      const reviewPage = result.reviewPage || {}
      const roleConfigs = reviewPage.roleConfigs || {}
      let sections = []

      this.applyReviewPageConfig(reviewPage, result.reward)

      if (this.options.targetUserId && this.options.gameId) {
        sections = [createSectionBase(role, 0, this.options, roleConfigs)]
      } else {
        const todos = Array.isArray(result.items) ? result.items : []
        sections = buildSectionsFromTodos(role, todos, roleConfigs)
      }

      if (!sections.length) {
        sections = [createSectionBase(role, 0, this.options, roleConfigs)]
      }

      this.setData({
        loading: false,
        evaluationSections: sections,
        pendingCount: sections.length
      })
    } catch (error) {
      const fallbackSections = [createSectionBase(role, 0, this.options, this.data.reviewPage.roleConfigs || {})]

      this.setData({
        loading: false,
        evaluationSections: fallbackSections,
        pendingCount: fallbackSections.length
      })

      if (error && error.message) {
        this.showInfo(error.message)
      }
    }
  },

  applyReviewPageConfig(config = {}, reward) {
    const options = Array.isArray(config.satisfactionOptions) ? config.satisfactionOptions : []

    this.setData({
      navTitle: config.navTitle || '',
      skipText: config.skipText || '',
      statusTitle: config.statusTitle || '',
      statusDesc: config.statusDesc || '',
      satisfactionQuestion: config.satisfactionQuestion || '',
      satisfactionOptions: options,
      selectedSatisfaction: options[0] ? options[0].id : this.data.selectedSatisfaction,
      storyTitle: config.storyTitle || '',
      aiTip: config.aiTip || '',
      storyPlaceholder: config.storyPlaceholder || '',
      storyMaxLength: Number(config.storyMaxLength) || this.data.storyMaxLength,
      aiSummaryText: config.aiSummaryText || '',
      ratingHint: config.ratingHint || '',
      npsHeadTitle: config.npsHeadTitle || '',
      npsQuestion: config.npsQuestion || '',
      npsLowLabel: config.npsLowLabel || '',
      npsHighLabel: config.npsHighLabel || '',
      submitText: config.submitText || '',
      submitNote: config.submitNote || '',
      reviewPage: config,
      reward: normalizeReward(reward)
    })
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
    const summary = buildReviewSummary(this.data.storyText, this.data.reviewPage.defaultSummary || '')
    this.setData({
      storyText: summary,
      storyLength: summary.length
    })
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

    navigateShellRoute(ROUTES.gameManage || 'pages/game/manage/index')
  },

  onSkipTap() {
    this.showInfo(this.data.reviewPage.skipToast || '')
  },

  async onSubmitTap() {
    const sections = this.data.evaluationSections || []
    if (!sections.length) {
      this.showInfo(this.data.reviewPage.noReviewTargetText || '')
      return
    }

    const reviews = sections
      .map((section) => ({
        gameId: section.gameId || this.options.gameId,
        targetUserId: section.targetUserId || this.options.targetUserId,
        targetRole: section.targetRole || 'member',
        score: Number(section.score) || 0,
        content: String(section.comment || this.data.storyText || '').trim(),
        tags: (section.tags || []).filter((tag) => tag.selected).map((tag) => tag.label),
        againIntent: resolveAgainIntent(this.data.selectedSatisfaction, this.data.reviewPage.againIntentBySatisfaction || {})
      }))
      .filter((item) => item.gameId && item.targetUserId)

    if (!reviews.length) {
      this.showInfo(this.data.reviewPage.missingTargetText || '')
      return
    }

    const invalid = reviews.find((item) => item.score < 1 || item.score > 5)
    if (invalid) {
      this.showInfo(this.data.reviewPage.missingScoreText || '')
      return
    }

    this.setData({ submitting: true })

    try {
      for (const review of reviews) {
        await reviewService.submitReview(review)
      }

      wx.showToast({
        title: this.data.reviewPage.submitSuccessText || '',
        icon: 'success'
      })

      setTimeout(() => {
        navigateShellRoute(`${ROUTES.gameReviewComplete}?count=${reviews.length}${this.options.gameId ? `&gameId=${this.options.gameId}` : ''}`)
      }, 300)
    } catch (error) {
      this.showInfo(error.message || this.data.reviewPage.submitFailedText || '')
    } finally {
      this.setData({ submitting: false })
    }
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
