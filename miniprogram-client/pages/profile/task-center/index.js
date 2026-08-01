const newbieService = require('../../../services/newbie')
const reviewService = require('../../../services/review')
const { ROUTES } = require('../../../config/routes')
const { navigateShellBack, navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')

const CATEGORY_TABS = [
  { key: 'newbie', label: '新手任务' },
  { key: 'daily', label: '每日任务' },
  { key: 'activity', label: '活动任务' }
]

function completedOf(item = {}) {
  return Boolean(item.completed || item.done || item.finished || item.status === 'completed' || item.status === 'done')
}

function rewardOf(item = {}) {
  if (item.rewardText) return String(item.rewardText)
  const rewardParts = []
  const rewardExperience = Number(item.rewardExperience)
  const rewardPoints = Number(item.rewardPoints)

  if (Number.isFinite(rewardExperience) && rewardExperience > 0) {
    rewardParts.push(`+${rewardExperience} 经验值`)
  }
  if (Number.isFinite(rewardPoints) && rewardPoints > 0) {
    rewardParts.push(`+${rewardPoints} 积分`)
  }
  if (rewardParts.length) return rewardParts.join(' · ')

  const value = item.reward || item.points || item.experience || item.xp
  if (value == null || value === '') return ''
  return /^[-+]?\d+(\.\d+)?$/.test(String(value)) ? `+${value} 经验值` : String(value)
}

function normalizeTask(item = {}, index = 0, highlightedCode = '') {
  const completed = completedOf(item)
  const isGuideTask = !completed && String(item.code || item.taskCode || item.id || '') === String(highlightedCode || '')
  return {
    ...item,
    id: item.id || item.code || `task-${index + 1}`,
    code: item.code || item.taskCode || item.id || '',
    title: item.title || item.name || '平台任务',
    rewardText: rewardOf(item),
    completed,
    guideTask: isGuideTask,
    statusText: item.statusText || (completed ? '已完成' : '去完成'),
    itemClass: `${completed ? 'is-completed' : ''} ${isGuideTask ? 'is-guide-task' : ''}`.trim(),
    statusClass: completed ? 'completed' : 'pending'
  }
}

function taskRoute(task = {}) {
  if (task.route) return task.route
  switch (task.type || task.code) {
    case 'realname':
    case 'complete_identity':
      return '/pages/login/realname/index'
    case 'role_apply':
    case 'apply_role':
      return `/${ROUTES.roleFlow}?mode=roleComparison&single=1`
    case 'first_game':
    case 'join_or_create_game':
      return `/${ROUTES.gameCreate}`
    case 'complete_game':
      return `/${ROUTES.gamePlayerManage}`
    case 'review':
    case 'submit_review':
      return `/${ROUTES.gameReview}`
    default:
      return ''
  }
}

function isReviewTask(task = {}) {
  const type = task.type || task.code
  return type === 'review' || type === 'submit_review'
}

Page({
  data: {
    pageTitle: '任务中心',
    onlineText: '在线',
    tabs: CATEGORY_TABS,
    activeCategory: 'newbie',
    tasks: [],
    categoryTasks: { newbie: [], daily: [], activity: [] },
    loading: true,
    error: '',
    completedCount: 0,
    totalCount: 0
  },

  onLoad() {
    this.loadTasks()
  },

  async loadTasks() {
    try {
      const data = await newbieService.getNewbieTasks()
      const sourceCategories = Array.isArray(data.categories) ? data.categories : []
      const guideTaskCode = String((data.guide || {}).currentTaskCode || '')
      const newbieItems = sourceCategories.length
        ? ((sourceCategories.find((item) => item.key === 'newbie') || {}).items || [])
        : (data.items || data.tasks || data.list || [])
      const categories = {}
      sourceCategories.forEach((category) => {
        categories[category.key] = (category.items || []).map((item, index) => normalizeTask(item, index, category.key === 'newbie' ? guideTaskCode : ''))
      })
      categories.newbie = (categories.newbie || newbieItems.map((item, index) => normalizeTask(item, index, guideTaskCode)))
      categories.daily = categories.daily || []
      categories.activity = categories.activity || []
      const tasks = categories.newbie
      this.setData({
        loading: false,
        error: '',
        tasks,
        categoryTasks: categories,
        completedCount: tasks.filter((item) => item.completed).length,
        totalCount: tasks.length
      })
    } catch (error) {
      this.setData({ loading: false, error: error.message || '任务加载失败' })
    }
  },

  onTabTap(event) {
    const key = event.currentTarget.dataset.key
    if (key) this.setData({ activeCategory: key, tasks: (this.data.categoryTasks || {})[key] || [] })
  },

  async onTaskTap(event) {
    const task = this.data.tasks.find((item) => String(item.id) === String(event.currentTarget.dataset.id))
    if (!task || task.completed) return

    if (isReviewTask(task)) {
      try {
        const result = await reviewService.getAvailableReviews()
        const items = Array.isArray(result && result.items) ? result.items : []
        const review = items.find((item) => item && item.gameId)
        if (!review) {
          wx.showToast({ title: '暂无待评价的局', icon: 'none' })
          return
        }
        navigateShellRoute(`/${ROUTES.gameReview}?gameId=${review.gameId}`, {
          currentRoute: ROUTES.profileTaskCenter
        })
      } catch (error) {
        wx.showToast({ title: '获取待评价局失败，请稍后重试', icon: 'none' })
      }
      return
    }

    const route = taskRoute(task)
    if (route) {
      navigateShellRoute(route, { currentRoute: ROUTES.profileTaskCenter })
      return
    }
    wx.showToast({ title: '任务暂未开放', icon: 'none' })
  },

  onBackTap() {
    if (!navigateShellBack()) navigateShellRoute(`/${ROUTES.profile}`)
  },

  handleShellNavTap(event) {
    navigateShellKey(event.detail && event.detail.key, { currentRoute: ROUTES.profileTaskCenter })
  }
})
