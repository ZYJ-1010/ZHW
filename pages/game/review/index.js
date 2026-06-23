const SATISFACTION_OPTIONS = [
  { id: 'great', emoji: '😄', title: '真好玩', desc: '五星体验' },
  { id: 'ok', emoji: '😐', title: '还行', desc: '合格' },
  { id: 'bad', emoji: '☹️', title: '不好玩', desc: '有待改进' }
]

const TARGET_EVALUATION_MAP = {
  expert: {
    id: 'expert',
    avatarText: 'ZH',
    avatarTheme: 'blue',
    title: '评价行家：张专家',
    desc: '产品架构咨询 · 已完成',
    ratingTitle: '服务质量',
    tagTitle: '行家标签（多选）',
    tags: ['专业能力强', '交付及时', '沟通顺畅', '超出预期', '性价比高', '推荐再合作'],
    placeholder: '分享你的服务体验...'
  },
  player: {
    id: 'player',
    avatarText: 'WA',
    avatarTheme: 'pink',
    title: '评价玩家：王总',
    desc: '需求确认 · 配合度',
    ratingTitle: '合作满意度',
    tagTitle: '玩家标签（多选）',
    tags: ['需求明确', '配合度高', '付款及时', '沟通友好', '长期合作潜力'],
    placeholder: '写下对需求方的评价...'
  },
  guide: {
    id: 'guide',
    avatarText: 'WA',
    avatarTheme: 'orange',
    title: '评价领路人：王引荐',
    desc: '撮合匹配度 · 协助交付',
    ratingTitle: '引荐满意度',
    tagTitle: '邀请标签（多选）',
    tags: ['匹配精准', '响应及时', '协助积极', '沟通高效', '值得信赖'],
    placeholder: '评价引荐人的服务质量...'
  }
}

const ROLE_TARGETS = {
  expert: ['player', 'guide'],
  player: ['expert', 'guide'],
  guide: ['expert', 'player']
}

const ROLE_ALIASES = {
  master: 'expert',
  specialist: 'expert',
  leader: 'guide',
  referrer: 'guide'
}

function normalizeRole(role) {
  const normalized = String(role || '').trim()

  return ROLE_TARGETS[normalized]
    ? normalized
    : ROLE_ALIASES[normalized] || 'expert'
}

function cloneEvaluationSection(targetType) {
  const section = TARGET_EVALUATION_MAP[targetType]

  return {
    ...section,
    score: 0,
    tags: section.tags.map((label) => ({ label, selected: false })),
    comment: ''
  }
}

function getEvaluationSectionsByRole(role) {
  return ROLE_TARGETS[role].map(cloneEvaluationSection)
}

Page({
  data: {
    satisfactionOptions: SATISFACTION_OPTIONS,
    selectedSatisfaction: 'great',
    storyText: '',
    storyLength: 0,
    storyMaxLength: 100,
    stars: [1, 2, 3, 4, 5],
    viewerRole: 'expert',
    evaluationSections: getEvaluationSectionsByRole('expert'),
    npsScores: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
    npsScore: 7
  },

  onLoad(options = {}) {
    const viewerRole = normalizeRole(options.role)

    this.setData({
      viewerRole,
      evaluationSections: getEvaluationSectionsByRole(viewerRole)
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
    if (this.data.storyText.trim()) {
      this.showInfo('AI总结功能开发中')
      return
    }

    const summary = '本次合作沟通顺畅，需求清晰，过程里有不少新启发。'

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

    this.showInfo('返回入口待接入')
  },

  onSkipTap() {
    this.showInfo('已跳过评价')
  },

  onSubmitTap() {
    this.showInfo('提交评价待接入')
  },

  showInfo(title) {
    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
