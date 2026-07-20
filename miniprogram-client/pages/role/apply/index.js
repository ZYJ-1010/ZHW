const roleService = require('../../../services/role')
const toast = require('../../../utils/toast')

function normalizeRoleType(value) {
  const roleType = String(value || '').trim().toLowerCase()

  if (roleType === 'expert' || roleType === 'master' || roleType === '行家') {
    return 'expert'
  }

  if (roleType === 'guide' || roleType === 'leader' || roleType === '领路人') {
    return 'guide'
  }

  return ''
}

function normalizePageConfig(config = {}) {
  return {
    pageTitle: config.pageTitle || '',
    pageDesc: config.pageDesc || '',
    texts: config.texts || {}
  }
}

function textOf(config, key) {
  const texts = config && config.texts ? config.texts : {}
  return texts[key] || ''
}

Page({
  data: {
    loading: true,
    submittingRole: '',
    pageConfig: normalizePageConfig(),
    texts: {},
    applications: []
  },

  onLoad(options = {}) {
    // 旧路由仍可能被收藏、任务或历史消息命中。统一进入新版申请流，
    // 保证所有行家/领路人申请都先展示条件页，再进入资料填写页。
    const roleType = normalizeRoleType(options.roleType || options.role)
    const target = roleType
      ? `/pages/role/flow/index?mode=${roleType === 'expert' ? 'expertApplyOverview' : 'roleApplyOverview'}&single=1&roleType=${roleType}`
      : '/pages/role/flow/index?mode=roleComparison&single=1'

    wx.redirectTo({
      url: target,
      fail: () => {
        wx.reLaunch({ url: target })
      }
    })
  },

  async loadApplications() {
    try {
      const data = await roleService.getMyRoleApplications()
      const pageConfig = normalizePageConfig(data.pageConfig)
      this.setData({
        loading: false,
        pageConfig,
        texts: pageConfig.texts,
        applications: Array.isArray(data.items) ? data.items : []
      })
    } catch (error) {
      this.setData({
        loading: false
      })
      toast.info(error.message || textOf(this.data.pageConfig, 'loadFailedText'))
    }
  },

  async submitApplication(event) {
    const roleType = event.currentTarget.dataset.role

    if (!roleType || this.data.submittingRole) {
      return
    }

    this.setData({
      submittingRole: roleType
    })

    try {
      const result = await roleService.submitRoleApplication(roleType)
      const applications = this.data.applications.map((item) => {
        if (item.roleType !== roleType) {
          return item
        }

        return Object.assign({}, item, {
          status: result.status,
          statusText: result.statusText
        })
      })

      this.setData({
        applications
      })
      toast.success(result.statusText || textOf(this.data.pageConfig, 'defaultSuccessText'))
    } catch (error) {
      toast.info(error.message || textOf(this.data.pageConfig, 'submitFailedText'))
    } finally {
      this.setData({
        submittingRole: ''
      })
    }
  }
})
