const profileService = require('../../../../services/profile')
const toast = require('../../../../utils/toast')

const ASSET_BASE = '/pages/profile/system-management/profile-info/assets'

Page({
  data: {
    icons: {
      user: `${ASSET_BASE}/icon-user.svg`,
      building: `${ASSET_BASE}/icon-building.svg`,
      shield: `${ASSET_BASE}/icon-shield.svg`,
      idCard: `${ASSET_BASE}/icon-id-card.svg`,
      enterprise: `${ASSET_BASE}/icon-enterprise.svg`,
      lock: `${ASSET_BASE}/icon-lock.svg`,
      chevron: `${ASSET_BASE}/icon-chevron-right.svg`
    },
    profile: {
      avatarText: 'ZW',
      name: '张伟',
      phone: '138****8888',
      contactVisibility: 'all',
      hobby: '未填写',
      company: '腾讯科技',
      jobTitle: '产品经理',
      businessCountText: '已设置3项',
      resources: '未填写',
      publicBusinessInfo: true
    },
    isSaving: false,
    visibilityOptions: [
      { key: 'all', label: '全部展示' },
      { key: 'member', label: '仅会员可见' },
      { key: 'hidden', label: '完全隐藏' }
    ],
    personalRows: [
      { key: 'avatar', label: '头像', type: 'avatar' },
      { key: 'name', label: '姓名', value: '张伟' },
      { key: 'contact', label: '联系方式', type: 'contact' },
      { key: 'hobby', label: '兴趣爱好', value: '未填写', muted: true }
    ],
    enterpriseRows: [
      { key: 'company', label: '公司名称', value: '腾讯科技' },
      { key: 'jobTitle', label: '职务', value: '产品经理' },
      { key: 'business', label: '主营业务', value: '已设置3项', muted: true },
      { key: 'resources', label: '可提供资源', value: '未填写', muted: true }
    ],
    certifications: [
      {
        key: 'personal',
        title: '个人身份认证',
        desc: '身份证+人脸识别',
        status: '已认证',
        statusClass: 'verified',
        iconSrc: `${ASSET_BASE}/icon-id-card.svg`,
        tone: 'green'
      },
      {
        key: 'enterprise',
        title: '企业认证',
        desc: '营业执照+对公账户',
        status: '未认证',
        statusClass: '',
        iconSrc: `${ASSET_BASE}/icon-enterprise.svg`,
        tone: 'blue'
      }
    ]
  },

  handleVisibilityTap(event) {
    const { key } = event.currentTarget.dataset

    if (!key || key === this.data.profile.contactVisibility) {
      return
    }

    this.setData({
      'profile.contactVisibility': key
    })
  },

  handleToggleBusinessPublic() {
    this.setData({
      'profile.publicBusinessInfo': !this.data.profile.publicBusinessInfo
    })
  },

  handleEditableRowTap() {
    toast.developing('资料编辑功能待接入后台后完善')
  },

  handleCertificationTap() {
    toast.developing('认证流程待接入正式认证接口后完善')
  },

  buildSavePayload() {
    const profile = this.data.profile

    return {
      personalInfo: {
        avatarText: profile.avatarText,
        name: profile.name,
        phoneMasked: profile.phone,
        contactVisibility: profile.contactVisibility,
        hobby: profile.hobby
      },
      enterpriseInfo: {
        company: profile.company,
        jobTitle: profile.jobTitle,
        businessCountText: profile.businessCountText,
        resources: profile.resources,
        publicBusinessInfo: profile.publicBusinessInfo
      },
      certifications: this.data.certifications.map((item) => ({
        key: item.key,
        status: item.status,
        statusClass: item.statusClass
      }))
    }
  },

  async handleSaveTap() {
    if (this.data.isSaving) {
      return
    }

    this.setData({
      isSaving: true
    })

    try {
      await profileService.saveSystemProfileInfo(this.buildSavePayload())
      toast.success('资料已保存')
    } catch (error) {
      toast.info(error.message || '资料保存失败，请稍后再试')
    } finally {
      this.setData({
        isSaving: false
      })
    }
  }
})
