const profileService = require('../../../services/profile')
const toast = require('../../../utils/toast')
const { navigateShellRoute } = require('../../../utils/shell-nav')

const PROFILE_INFO_ROUTE = '/pages/profile/system-management/profile-info/index'
const MAP_ROUTE = '/pages/map/index'

function buildFields(profile = {}) {
  const personal = profile.personalInfo || {}
  const enterprise = profile.enterpriseInfo || {}

  return [
    { key: 'needs', label: '我的需求', value: personal.needs || '前往个人主页填写 >', route: PROFILE_INFO_ROUTE },
    { key: 'resources', label: '我的资源', value: enterprise.resources || '前往个人主页填写 >', route: PROFILE_INFO_ROUTE },
    { key: 'interestGames', label: '感兴趣组局', value: personal.interestGames || '选择 >', route: '/pages/game/hall/index' },
    { key: 'businessScale', label: '营收规模', value: enterprise.businessCountText || '选填 >', route: PROFILE_INFO_ROUTE },
    { key: 'industry', label: '所在行业', value: enterprise.industry || '选择 >', route: PROFILE_INFO_ROUTE },
    { key: 'address', label: '地址定位', value: personal.address || '选择 >', route: MAP_ROUTE }
  ]
}

Page({
  data: {
    fields: buildFields(),
    profileInfo: {},
    isSaving: false
  },

  onLoad() {
    this.loadProfileInfo()
  },

  async loadProfileInfo() {
    try {
      const profileInfo = await profileService.getSystemProfileInfo()

      this.setData({
        profileInfo,
        fields: buildFields(profileInfo)
      })
    } catch (error) {
      toast.info(error.message || '获取资料失败')
    }
  },

  handleFieldTap(event) {
    const index = Number(event.currentTarget.dataset.index)
    const field = this.data.fields[index]

    if (!field || !field.route) {
      toast.info('暂无可用入口')
      return
    }

    navigateShellRoute(field.route)
  },

  async handleSaveTap() {
    if (this.data.isSaving) {
      return
    }

    this.setData({ isSaving: true })
    try {
      const profileInfo = this.data.profileInfo || {}
      const saved = await profileService.saveSystemProfileInfo({
        ...profileInfo,
        matchInfo: {
          fields: this.data.fields.map((item) => ({
            key: item.key,
            label: item.label,
            value: item.value
          }))
        }
      })

      this.setData({
        profileInfo: saved,
        fields: buildFields(saved),
        isSaving: false
      })
      toast.info('已保存')
    } catch (error) {
      this.setData({ isSaving: false })
      toast.info(error.message || '保存失败')
    }
  }
})
