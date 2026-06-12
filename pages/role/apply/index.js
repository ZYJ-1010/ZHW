const roleService = require('../../../services/role')
const toast = require('../../../utils/toast')

Page({
  data: {
    loading: true,
    submittingRole: '',
    applications: []
  },

  onLoad() {
    this.loadApplications()
  },

  async loadApplications() {
    try {
      const applications = await roleService.getMyRoleApplications()
      this.setData({
        loading: false,
        applications
      })
    } catch (error) {
      this.setData({
        loading: false
      })
      toast.info(error.message || '角色申请加载失败')
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
      toast.success(result.statusText)
    } catch (error) {
      toast.info(error.message || '提交失败')
    } finally {
      this.setData({
        submittingRole: ''
      })
    }
  }
})
