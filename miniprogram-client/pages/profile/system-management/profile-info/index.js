const profileService = require('../../../../services/profile')
const fileService = require('../../../../services/file')
const toast = require('../../../../utils/toast')
const { navigateShellRoute } = require('../../../../utils/shell-nav')

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
      avatarText: '',
      avatarUrl: '',
      pendingAvatarUrl: '',
      avatarAuditStatus: '',
      avatarAuditText: '',
      name: '',
      phone: '',
      contactVisibility: 'all',
      hobby: '',
      company: '',
      jobTitle: '',
      businessCountText: '',
      resources: '',
      publicBusinessInfo: false
    },
    avatarFileId: 0,
    pendingAvatarFileId: 0,
    isSaving: false,
    isAvatarSubmitting: false,
    isEnterpriseSubmitting: false,
    loadError: '',
    visibilityOptions: [],
    personalRows: [
      { key: 'avatar', label: '头像', type: 'avatar' },
      { key: 'name', label: '姓名', value: '' },
      { key: 'contact', label: '联系方式', type: 'contact' },
      { key: 'hobby', label: '兴趣爱好', value: '', muted: true }
    ],
    enterpriseRows: [
      { key: 'company', label: '公司名称', value: '' },
      { key: 'jobTitle', label: '职务', value: '' },
      { key: 'business', label: '主营业务', value: '', muted: true },
      { key: 'resources', label: '可提供资源', value: '', muted: true }
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

  onLoad() {
    this.loadSystemProfileInfo()
  },

  async loadSystemProfileInfo() {
    try {
      const remoteProfile = await profileService.getSystemProfileInfo()
      this.applySystemProfileInfo(remoteProfile || {})
    } catch (error) {
      this.setData({
        loadError: error.message || '资料加载失败'
      })
    }
  },

  applySystemProfileInfo(remote = {}) {
    const current = this.data.profile || {}
    const profile = remote.profile || {}
    const personalInfo = remote.personalInfo || {}
    const enterpriseInfo = remote.enterpriseInfo || {}
    const nextProfile = {
      ...current,
      ...profile,
      avatarText: personalInfo.avatarText || profile.avatarText || current.avatarText,
      avatarUrl: personalInfo.avatarUrl || profile.avatarUrl || current.avatarUrl,
      pendingAvatarUrl: personalInfo.pendingAvatarUrl || profile.pendingAvatarUrl || '',
      avatarAuditStatus: personalInfo.avatarAuditStatus || profile.avatarAuditStatus || '',
      avatarAuditText: personalInfo.avatarAuditText || profile.avatarAuditText || '',
      name: personalInfo.name || profile.name || current.name,
      phone: personalInfo.phoneMasked || profile.phone || current.phone,
      contactVisibility: personalInfo.contactVisibility || profile.contactVisibility || current.contactVisibility,
      hobby: personalInfo.hobby || profile.hobby || current.hobby,
      company: enterpriseInfo.company || profile.company || current.company,
      jobTitle: enterpriseInfo.jobTitle || profile.jobTitle || current.jobTitle,
      businessCountText: enterpriseInfo.businessCountText || profile.businessCountText || current.businessCountText,
      resources: enterpriseInfo.resources || profile.resources || current.resources,
      publicBusinessInfo: typeof enterpriseInfo.publicBusinessInfo === 'boolean'
        ? enterpriseInfo.publicBusinessInfo
        : (typeof profile.publicBusinessInfo === 'boolean' ? profile.publicBusinessInfo : current.publicBusinessInfo)
    }
    const nextData = {
      profile: nextProfile,
      avatarFileId: Number(personalInfo.avatarFileId || profile.avatarFileId || this.data.avatarFileId || 0) || 0,
      pendingAvatarFileId: Number(personalInfo.pendingAvatarFileId || profile.pendingAvatarFileId || 0) || 0,
      personalRows: this.patchPersonalRows(nextProfile),
      enterpriseRows: this.patchEnterpriseRows(nextProfile),
      loadError: ''
    }

    if (Array.isArray(remote.certifications) && remote.certifications.length) {
      nextData.certifications = this.mergeCertifications(remote.certifications)
    }

    if (Array.isArray(remote.visibilityOptions)) {
      nextData.visibilityOptions = remote.visibilityOptions.filter((item) => item && item.key && item.label)
    }

    this.setData(nextData)
  },

  patchPersonalRows(profile) {
    const valueMap = {
      name: profile.name,
      hobby: profile.hobby
    }

    return this.data.personalRows.map((row) => ({
      ...row,
      value: Object.prototype.hasOwnProperty.call(valueMap, row.key) ? valueMap[row.key] : row.value
    }))
  },

  patchEnterpriseRows(profile) {
    const valueMap = {
      company: profile.company,
      jobTitle: profile.jobTitle,
      business: profile.businessCountText,
      resources: profile.resources
    }

    return this.data.enterpriseRows.map((row) => ({
      ...row,
      value: Object.prototype.hasOwnProperty.call(valueMap, row.key) ? valueMap[row.key] : row.value
    }))
  },

  mergeCertifications(remoteItems = []) {
    return this.data.certifications.map((item) => ({
      ...item,
      ...(remoteItems.find((remoteItem) => remoteItem.key === item.key) || {})
    }))
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

  handleEditableRowTap(event = {}) {
    return this.editProfileRow(event)
  },

  handleCertificationTap(event = {}) {
    return this.openCertification(event)
  },

  editProfileRow(event = {}) {
    const { key, section } = (event.currentTarget && event.currentTarget.dataset) || {}
    const fieldMap = {
      name: { path: 'profile.name', prop: 'name', title: '姓名', current: this.data.profile.name },
      hobby: { path: 'profile.hobby', prop: 'hobby', title: '兴趣爱好', current: this.data.profile.hobby },
      company: { path: 'profile.company', prop: 'company', title: '公司名称', current: this.data.profile.company },
      jobTitle: { path: 'profile.jobTitle', prop: 'jobTitle', title: '职务', current: this.data.profile.jobTitle },
      business: { path: 'profile.businessCountText', prop: 'businessCountText', title: '主营业务', current: this.data.profile.businessCountText },
      resources: { path: 'profile.resources', prop: 'resources', title: '可提供资源', current: this.data.profile.resources }
    }
    const field = fieldMap[key]

    if (key === 'contact') {
      toast.info('联系方式来自实名认证/手机号绑定，请在认证流程中更新')
      return
    }

    if (key === 'avatar') {
      this.chooseAvatar()
      return
    }

    if (!field || !wx.showModal) {
      return
    }

    wx.showModal({
      title: `编辑${field.title}`,
      editable: true,
      placeholderText: `请输入${field.title}`,
      content: field.current || '',
      success: (res) => {
        if (!res.confirm) {
          return
        }

        const value = String(res.content || '').trim()

        if (!value) {
          toast.info(`${field.title}不能为空`)
          return
        }

        const nextProfile = {
          ...this.data.profile,
          [field.prop]: value
        }
        const nextData = {
          [field.path]: value
        }

        if (section === 'personal') {
          nextData.personalRows = this.patchPersonalRows(nextProfile)
        }

        if (section === 'enterprise') {
          nextData.enterpriseRows = this.patchEnterpriseRows(nextProfile)
        }

        this.setData(nextData)
      }
    })
  },

  chooseAvatar() {
    const onSuccess = async (result = {}) => {
      if (this.data.isAvatarSubmitting) {
        return
      }

      const file = Array.isArray(result.tempFiles) ? result.tempFiles[0] : null
      const path = (file && (file.tempFilePath || file.path)) || (result.tempFilePaths || [])[0]

      if (!path) {
        return
      }

      this.setData({
        isAvatarSubmitting: true
      })

      try {
        const fileId = await fileService.uploadSingleFile(path, {
          bizType: 'avatar',
          objectId: 0
        })
        const avatarText = this.data.profile.name ? this.data.profile.name.slice(0, 1) : '我'
        this.setData({
          pendingAvatarFileId: fileId,
          'profile.avatarText': avatarText,
          'profile.pendingAvatarUrl': path,
          'profile.avatarAuditStatus': 'pending',
          'profile.avatarAuditText': '待审核'
        })
        const savedProfile = await profileService.saveSystemProfileInfo(this.buildSavePayload({
          avatarFileId: fileId,
          avatarText
        }))
        this.applySystemProfileInfo(savedProfile || {})
        toast.success('头像已提交后台审核')
      } catch (error) {
        toast.info(error.message || '头像提交审核失败')
        this.loadSystemProfileInfo()
      } finally {
        this.setData({
          isAvatarSubmitting: false
        })
      }
    }

    if (wx.chooseMedia) {
      wx.chooseMedia({
        count: 1,
        mediaType: ['image'],
        sourceType: ['album', 'camera'],
        sizeType: ['compressed'],
        success: onSuccess
      })
      return
    }

    wx.chooseImage({
      count: 1,
      sourceType: ['album', 'camera'],
      sizeType: ['compressed'],
      success: onSuccess
    })
  },

  handleAvatarImageError() {
    this.setData({
      'profile.avatarUrl': '',
      'profile.pendingAvatarUrl': ''
    })
  },

  async openCertification(event = {}) {
    const { key } = (event.currentTarget && event.currentTarget.dataset) || {}

    if (key === 'personal') {
      navigateShellRoute('/pages/login/realname/index')
      return
    }

    if (this.data.isEnterpriseSubmitting) {
      return
    }
    const certification = (this.data.certifications || []).find((item) => item.key === 'enterprise') || {}
    if (certification.status === '审核中') {
      toast.info('企业认证材料正在后台审核')
      return
    }
    if (certification.status === '已认证') {
      toast.info('企业认证已通过')
      return
    }
    const ask = (title, content = '') => new Promise((resolve) => {
      wx.showModal({
        title,
        editable: true,
        content,
        placeholderText: `请输入${title}`,
        success: (result) => resolve(result.confirm ? String(result.content || '').trim() : '')
      })
    })
    const companyName = await ask('公司名称')
    if (!companyName) return
    const unifiedSocialCreditCode = await ask('统一社会信用代码')
    if (!unifiedSocialCreditCode) return
    const legalPerson = await ask('法定代表人')
    if (!legalPerson) return
    if (typeof wx.chooseMessageFile !== 'function') {
      toast.info('当前环境不支持选择企业认证材料')
      return
    }
    const chooseFile = (title) => new Promise((resolve) => {
      wx.showModal({
        title,
        content: '请选择文件后继续',
        showCancel: true,
        success: (modal) => {
          if (!modal.confirm) {
            resolve('')
            return
          }
          wx.chooseMessageFile({
            count: 1,
            type: 'file',
            success: (result) => {
              const file = Array.isArray(result.tempFiles) ? result.tempFiles[0] : null
              resolve(file && (file.path || file.tempFilePath) || '')
            },
            fail: () => resolve('')
          })
        },
        fail: () => resolve('')
      })
    })
    const businessLicensePath = await chooseFile('选择营业执照')
    if (!businessLicensePath) return
    const publicAccountPath = await chooseFile('选择对公账户证明')
    if (!publicAccountPath) return
    this.setData({ isEnterpriseSubmitting: true })
    try {
      const fileIds = []
      for (const path of [businessLicensePath, publicAccountPath]) {
        fileIds.push(await fileService.uploadSingleFile(path, {
          bizType: 'enterprise_material',
          objectId: 0
        }))
      }
      await profileService.submitEnterpriseCertification({
        companyName,
        unifiedSocialCreditCode,
        legalPerson,
        businessLicenseFileId: fileIds[0],
        publicAccountFileId: fileIds[1]
      })
      toast.success('企业认证已提交，等待后台审核')
      await this.loadSystemProfileInfo()
    } catch (error) {
      toast.info(error.message || '企业认证提交失败，请稍后重试')
    } finally {
      this.setData({ isEnterpriseSubmitting: false })
    }
  },

  buildSavePayload(options = {}) {
    const profile = this.data.profile
    const avatarFileId = Number(options.avatarFileId || this.data.pendingAvatarFileId || 0) || 0
    const avatarText = options.avatarText || profile.avatarText

    return {
      personalInfo: {
        avatarText,
        avatarFileId,
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
      const savedProfile = await profileService.saveSystemProfileInfo(this.buildSavePayload())
      this.applySystemProfileInfo(savedProfile || {})
      const savedPersonal = (savedProfile && savedProfile.personalInfo) || {}
      toast.success(savedPersonal.avatarAuditStatus === 'pending' ? '资料已保存，头像待审核' : '资料已保存')
    } catch (error) {
      toast.info(error.message || '资料保存失败，请稍后再试')
    } finally {
      this.setData({
        isSaving: false
      })
    }
  }
})
