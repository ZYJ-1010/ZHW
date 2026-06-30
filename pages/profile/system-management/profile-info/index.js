const profileService = require('../../../../services/profile')
const toast = require('../../../../utils/toast')

const ASSET_BASE = '/pages/profile/system-management/profile-info/assets'

const EMPTY_PROFILE = {
  avatarText: '',
  name: '',
  phone: '',
  contactVisibility: '',
  hobby: '',
  company: '',
  jobTitle: '',
  businessCountText: '',
  resources: '',
  publicBusinessInfo: false
}

const PERSONAL_ROW_META = [
  { key: 'avatar', label: '头像', type: 'avatar' },
  { key: 'name', label: '姓名', field: 'name' },
  { key: 'contact', label: '联系方式', type: 'contact' },
  { key: 'hobby', label: '兴趣爱好', field: 'hobby', muted: true }
]

const ENTERPRISE_ROW_META = [
  { key: 'company', label: '公司名称', field: 'company' },
  { key: 'jobTitle', label: '职务', field: 'jobTitle' },
  { key: 'business', label: '主营业务', field: 'businessCountText', muted: true },
  { key: 'resources', label: '可提供资源', field: 'resources', muted: true }
]

const CERT_META = {
  personal: {
    iconSrc: `${ASSET_BASE}/icon-id-card.svg`,
    tone: 'green'
  },
  enterprise: {
    iconSrc: `${ASSET_BASE}/icon-enterprise.svg`,
    tone: 'blue'
  }
}

function pickFirstValue() {
  const values = Array.prototype.slice.call(arguments)

  for (let index = 0; index < values.length; index += 1) {
    if (values[index] !== undefined && values[index] !== null && values[index] !== '') {
      return values[index]
    }
  }

  return ''
}

function normalizeList(list) {
  return Array.isArray(list) ? list : []
}

function normalizeBoolean(value) {
  if (typeof value === 'boolean') {
    return value
  }

  if (typeof value === 'number') {
    return value === 1
  }

  return String(value || '').toLowerCase() === 'true'
}

function normalizeProfile(data) {
  const source = data || {}
  const personal = source.personalInfo || source.personal || source.profile || {}
  const enterprise = source.enterpriseInfo || source.enterprise || {}

  return Object.assign({}, EMPTY_PROFILE, {
    avatarText: pickFirstValue(personal.avatarText, source.avatarText),
    name: pickFirstValue(personal.name, source.name),
    phone: pickFirstValue(personal.phoneMasked, personal.phone, source.phoneMasked, source.phone),
    contactVisibility: pickFirstValue(personal.contactVisibility, source.contactVisibility),
    hobby: pickFirstValue(personal.hobby, source.hobby),
    company: pickFirstValue(enterprise.company, source.company),
    jobTitle: pickFirstValue(enterprise.jobTitle, source.jobTitle),
    businessCountText: pickFirstValue(enterprise.businessCountText, enterprise.businessText, source.businessCountText),
    resources: pickFirstValue(enterprise.resources, source.resources),
    publicBusinessInfo: normalizeBoolean(
      enterprise.publicBusinessInfo !== undefined ? enterprise.publicBusinessInfo : source.publicBusinessInfo
    )
  })
}

function normalizeRows(rows, metaRows, profile) {
  const remoteRows = normalizeList(rows)
  const sourceRows = remoteRows.length ? remoteRows : metaRows

  return sourceRows.map((item) => {
    const row = item || {}
    const key = pickFirstValue(row.key, row.id)
    const meta = metaRows.find((metaItem) => metaItem.key === key) || {}
    const field = pickFirstValue(row.field, meta.field)

    return {
      key,
      label: pickFirstValue(row.label, row.title, meta.label),
      type: pickFirstValue(row.type, meta.type),
      value: pickFirstValue(row.value, row.valueText, field ? profile[field] : ''),
      muted: Boolean(row.muted !== undefined ? row.muted : meta.muted)
    }
  }).filter((item) => item.key)
}

function normalizeVisibilityOptions(data) {
  const source = data || {}

  return normalizeList(source.visibilityOptions || source.contactVisibilityOptions)
    .map((item) => ({
      key: pickFirstValue(item.key, item.value, item.id),
      label: pickFirstValue(item.label, item.title, item.name)
    }))
    .filter((item) => item.key && item.label)
}

function normalizeCertifications(data) {
  const certifications = normalizeList(data && data.certifications)

  return certifications.map((item) => {
    const cert = item || {}
    const key = pickFirstValue(cert.key, cert.id, cert.type)
    const meta = CERT_META[key] || {}

    return {
      key,
      title: pickFirstValue(cert.title, cert.name),
      desc: pickFirstValue(cert.desc, cert.description),
      status: pickFirstValue(cert.statusText, cert.status),
      statusClass: pickFirstValue(cert.statusClass),
      iconSrc: pickFirstValue(cert.localIcon, meta.iconSrc),
      tone: pickFirstValue(cert.tone, meta.tone)
    }
  }).filter((item) => item.key)
}

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
    profile: Object.assign({}, EMPTY_PROFILE),
    isSaving: false,
    visibilityOptions: [],
    personalRows: normalizeRows([], PERSONAL_ROW_META, EMPTY_PROFILE),
    enterpriseRows: normalizeRows([], ENTERPRISE_ROW_META, EMPTY_PROFILE),
    certifications: []
  },

  onLoad() {
    this.loadProfileInfo()
  },

  async loadProfileInfo() {
    try {
      const data = await profileService.getSystemProfileInfo()
      const profile = normalizeProfile(data)

      this.setData({
        profile,
        visibilityOptions: normalizeVisibilityOptions(data),
        personalRows: normalizeRows(data && data.personalRows, PERSONAL_ROW_META, profile),
        enterpriseRows: normalizeRows(data && data.enterpriseRows, ENTERPRISE_ROW_META, profile),
        certifications: normalizeCertifications(data)
      })
    } catch (error) {
      this.setData({
        profile: Object.assign({}, EMPTY_PROFILE),
        visibilityOptions: [],
        personalRows: normalizeRows([], PERSONAL_ROW_META, EMPTY_PROFILE),
        enterpriseRows: normalizeRows([], ENTERPRISE_ROW_META, EMPTY_PROFILE),
        certifications: []
      })
      toast.info(error.message || '资料设置加载失败')
    }
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
