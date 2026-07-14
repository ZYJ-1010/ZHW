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
	if (normalized === 'member') {
		return 'player'
	}
	if (normalized === 'main_guide') {
		return 'guide'
	}
  if (['expert', 'player', 'guide'].includes(normalized)) {
    return normalized
  }
  return ROLE_ALIASES[normalized] || 'expert'
}

function toPositiveInt(value) {
  const number = Number(value)
  return Number.isInteger(number) && number > 0 ? number : 0
}

function decodeOption(value) {
  if (value === undefined || value === null) {
    return ''
  }

  try {
    return decodeURIComponent(String(value))
  } catch (error) {
    return String(value)
  }
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
    avatarText: fallback.avatarText || '',
    avatarType: fallback.avatarType || '',
    avatarTheme: fallback.avatarTheme || base.avatarTheme || '',
	title: fallback.title || `评价：${targetLabel}`,
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
	return (Array.isArray(todos) ? todos : []).map((todo, index) => createSectionBase(normalizeRole(todo.targetRole || role), index, {
    id: todo.id,
    gameId: todo.gameId,
    targetUserId: todo.targetUserId,
    targetRole: todo.targetRole,
    targetName: todo.targetName,
    targetNickname: todo.targetNickname,
    avatarText: todo.avatarText,
    avatarType: todo.avatarType,
    avatarTheme: todo.avatarTheme,
    deadlineAt: todo.deadlineAt,
	title: todo.targetName ? `评价：${todo.targetName}` : '',
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
	showOrganizerNps: false,
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
      targetName: decodeOption(options.targetName || options.name).trim(),
      targetNickname: decodeOption(options.targetNickname).trim(),
      title: decodeOption(options.title).trim()
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
		const result = await reviewService.getAvailableReviews({ gameId: this.options.gameId || undefined })
      const reviewPage = result.reviewPage || {}
      const roleConfigs = reviewPage.roleConfigs || {}
      let sections = []

      this.applyReviewPageConfig(reviewPage, result.reward)

      const todos = Array.isArray(result.items) ? result.items : []
      const gameTodos = this.options.gameId
        ? todos.filter((todo) => toPositiveInt(todo.gameId) === this.options.gameId)
        : todos
		// Always render every server-authorized target for this game. Route
		// parameters identify the entry point but must not hide other todos.
		sections = buildSectionsFromTodos(role, gameTodos, roleConfigs)

      this.setData({
        loading: false,
        evaluationSections: sections,
		pendingCount: sections.length,
		showOrganizerNps: Boolean(result.viewer && result.viewer.showOrganizerNps)
      })
    } catch (error) {
      this.setData({
        loading: false,
        evaluationSections: [],
        pendingCount: 0
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
		const title = this.data.reviewPage.skipToast || ''
		if (title) {
			this.showInfo(title)
		}
		setTimeout(() => this.onBackTap(), title ? 300 : 0)
	},

  async onSubmitTap() {
    const sections = this.data.evaluationSections || []
    if (!sections.length) {
      this.showInfo(this.data.reviewPage.noReviewTargetText || '')
      return
    }

	const reviews = sections
		.map((section) => ({
			sectionId: section.id,
			gameId: section.gameId || this.options.gameId,
        targetUserId: section.targetUserId || this.options.targetUserId,
        targetRole: section.targetRole || 'member',
        score: Number(section.score) || 0,
        content: String(section.comment || this.data.storyText || '').trim(),
        tags: (section.tags || []).filter((tag) => tag.selected).map((tag) => tag.label),
			againIntent: resolveAgainIntent(this.data.selectedSatisfaction, this.data.reviewPage.againIntentBySatisfaction || {}),
			npsScore: this.data.showOrganizerNps ? Number(this.data.npsScore) : undefined
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
	let submittedCount = 0
	let rewardedPoints = 0

	try {
		for (const review of reviews) {
			const { sectionId, ...payload } = review
			const result = await reviewService.submitReview(payload)
			const rewardPoints = Number(result && result.reward && result.reward.points)
			const logPoints = Number(result && result.pointsLog && result.pointsLog.changeValue)
			rewardedPoints += Number.isFinite(rewardPoints)
				? rewardPoints
				: (Number.isFinite(logPoints) ? logPoints : 0)
			submittedCount += 1
			const remainingSections = this.data.evaluationSections.filter((section) => section.id !== sectionId)
			this.setData({
				evaluationSections: remainingSections,
				pendingCount: remainingSections.length
			})
		}

      wx.showToast({
        title: this.data.reviewPage.submitSuccessText || '',
        icon: 'success'
      })

      setTimeout(() => {
        const completeRoute = `/${ROUTES.gameReviewComplete}?count=${submittedCount}&rewardPoints=${Math.max(0, rewardedPoints)}${this.options.gameId ? `&gameId=${this.options.gameId}` : ''}`
        wx.redirectTo({
          url: completeRoute,
          fail: () => navigateShellRoute(completeRoute)
        })
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
