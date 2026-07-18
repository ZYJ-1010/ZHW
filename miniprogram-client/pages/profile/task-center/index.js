const newbieService = require('../../../services/newbie')
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
  const value = item.reward || item.points || item.experience || item.xp
  if (value == null || value === '') return ''
  return /^[-+]?\d+(\.\d+)?$/.test(String(value)) ? `+${value} 经验值` : String(value)
}

function normalizeTask(item = {}, index = 0) {
  const completed = completedOf(item)
  return {
    ...item,
    id: item.id || item.code || `task-${index + 1}`,
    code: item.code || item.taskCode || item.id || '',
    title: item.title || item.name || '平台任务',
    rewardText: rewardOf(item),
    completed,
    statusText: item.statusText || (completed ? '已完成' : '去完成'),
    itemClass: completed ? 'is-completed' : '',
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

Page({
  data: {
    pageTitle: '任务中心',
    onlineText: '在线',
    tabs: CATEGORY_TABS,
    activeCategory: 'newbie',
    tasks: [],
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
      const newbieItems = sourceCategories.length
        ? ((sourceCategories.find((item) => item.key === 'newbie') || {}).items || [])
        : (data.items || data.tasks || data.list || [])
      const tasks = newbieItems.map(normalizeTask)
      this.setData({
        loading: false,
        error: '',
        tasks,
        completedCount: tasks.filter((item) => item.completed).length,
        totalCount: tasks.length
      })
    } catch (error) {
      this.setData({ loading: false, error: error.message || '任务加载失败' })
    }
  },

  onTabTap(event) {
    const key = event.currentTarget.dataset.key
    if (key) this.setData({ activeCategory: key })
  },

  onTaskTap(event) {
    const task = this.data.tasks.find((item) => String(item.id) === String(event.currentTarget.dataset.id))
    if (!task || task.completed) return
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
