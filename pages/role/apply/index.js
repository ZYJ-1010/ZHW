const roleService = require('../../../services/role')
const toast = require('../../../utils/toast')

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

  onLoad() {
    this.loadApplications()
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
