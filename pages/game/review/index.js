const SATISFACTION_OPTIONS = [
  { id: 'great', emoji: '😄', title: '真好玩', desc: '五星体验' },
  { id: 'ok', emoji: '😐', title: '还行', desc: '合格' },
  { id: 'bad', emoji: '☹️', title: '不好玩', desc: '有待改进' }
]

const EVALUATION_SECTIONS = [
  {
    id: 'player',
    avatarText: 'WA',
    avatarTheme: 'pink',
    title: '评价玩家：王总',
    desc: '需求确认 · 配合度',
    ratingTitle: '合作满意度',
    score: 0,
    tagTitle: '玩家标签（多选）',
    tags: [
      { label: '需求明确', selected: false },
      { label: '配合度高', selected: false },
      { label: '付款及时', selected: false },
      { label: '沟通友好', selected: false },
      { label: '长期合作潜力', selected: false }
    ],
    comment: '',
    placeholder: '写下对需求方的评价...'
  },
  {
    id: 'guide',
    avatarText: 'WA',
    avatarTheme: 'orange',
    title: '评价领路人：王引荐',
    desc: '撮合匹配度 · 协助交付',
    ratingTitle: '引荐满意度',
    score: 0,
    tagTitle: '邀请标签（多选）',
    tags: [
      { label: '匹配精准', selected: false },
      { label: '响应及时', selected: false },
      { label: '协助积极', selected: false },
      { label: '沟通高效', selected: false },
      { label: '值得信赖', selected: false }
    ],
    comment: '',
    placeholder: '评价引荐人的服务质量...'
  }
]

function cloneEvaluationSections() {
  return EVALUATION_SECTIONS.map((section) => ({
    ...section,
    tags: section.tags.map((tag) => ({ ...tag }))
  }))
}

Page({
  data: {
    satisfactionOptions: SATISFACTION_OPTIONS,
    selectedSatisfaction: 'great',
    storyText: '',
    storyLength: 0,
    storyMaxLength: 100,
    stars: [1, 2, 3, 4, 5],
    evaluationSections: cloneEvaluationSections(),
    npsScores: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
    npsScore: 7
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
