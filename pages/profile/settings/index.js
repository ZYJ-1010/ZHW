const toast = require('../../../utils/toast')
const profileService = require('../../../services/profile')

const ASSET_BASE = '/pages/profile/settings/assets'

function asset(name) {
  return `${ASSET_BASE}/${name}`
}

const DEFAULT_ICON = asset('icon-document.png')
const ICON_MAP = {
  payPassword: asset('icon-pay-lock.png'),
  loginPassword: asset('icon-login-lock.png'),
  phone: asset('icon-mobile.png'),
  facePay: asset('icon-pay-shield.png'),
  gamePush: asset('icon-bell.png'),
  systemNotice: asset('icon-calendar.png'),
  subscribeNotice: asset('icon-mail.png'),
  emailNotice: asset('icon-mail-off.png'),
  quietHours: asset('icon-clock.png'),
  showGames: asset('icon-eye.png'),
  showReviews: asset('icon-comment.png'),
  findByPhone: asset('icon-phone.png'),
  personalized: asset('icon-star.png'),
  privacySummary: asset('icon-document.png'),
  thirdPartyList: asset('icon-users.png'),
  collectionList: asset('icon-document-list.png'),
  clearCache: asset('icon-trash.png'),
  about: asset('icon-info.png')
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

function normalizeIcon(row, rowId) {
  const icon = pickFirstValue(row.localIcon, row.iconLocal, row.icon)

  if (typeof icon === 'string' && icon.indexOf('/pages/') === 0) {
    return icon
  }

  return ICON_MAP[rowId] || DEFAULT_ICON
}

function normalizeRow(row) {
  const source = row || {}
  const id = pickFirstValue(source.id, source.key, source.settingKey)
  const label = pickFirstValue(source.label, source.title, source.name)
  const isSwitch = source.switch === true || source.type === 'switch' || typeof source.enabled === 'boolean'

  return {
    id,
    label,
    icon: normalizeIcon(source, id),
    value: pickFirstValue(source.valueText, source.displayValue, source.value),
    arrow: source.arrow !== undefined ? Boolean(source.arrow) : Boolean(source.route || source.url || source.action),
    switch: isSwitch,
    enabled: normalizeBoolean(source.enabled !== undefined ? source.enabled : source.value),
    disabled: Boolean(source.disabled),
    route: pickFirstValue(source.route, source.path),
    url: pickFirstValue(source.url, source.webUrl),
    action: pickFirstValue(source.action, source.actionKey),
    message: pickFirstValue(source.message, source.toastText)
  }
}

function normalizeSections(data) {
  const source = data || {}
  const sections = normalizeList(source.sections || source.groups || source.items)

  return sections.map((section) => {
    const group = section || {}
    const rows = normalizeList(group.rows || group.items || group.children)
      .map(normalizeRow)
      .filter((row) => row.id || row.label)

    return {
      title: pickFirstValue(group.title, group.name, group.label),
      rows
    }
  }).filter((section) => section.title || section.rows.length)
}

Page({
  data: {
    sections: [],
    chevronIcon: asset('icon-chevron-right.svg'),
    previewMode: 'static'
  },

  onLoad(options = {}) {
    this.setData({
      previewMode: options.mode || 'static'
    })
    this.loadSettings()
  },

  async loadSettings() {
    try {
      const data = await profileService.getProfileSettings()

      this.setData({
        sections: normalizeSections(data)
      })
    } catch (error) {
      this.setData({
        sections: []
      })
      toast.info(error.message || '系统设置加载失败')
    }
  },

  async handleSwitch(event) {
    const { sectionIndex, rowIndex } = event.currentTarget.dataset
    const row = this.getRow(sectionIndex, rowIndex)

    if (!row || row.disabled) {
      return
    }

    const key = `sections[${sectionIndex}].rows[${rowIndex}].enabled`
    const nextValue = !row.enabled

    this.setData({
      [key]: nextValue
    })

    try {
      const data = await profileService.saveProfileSettings({
        key: row.id,
        value: nextValue,
        enabled: nextValue
      })

      if (data && (data.sections || data.groups || data.items)) {
        this.setData({
          sections: normalizeSections(data)
        })
      }
    } catch (error) {
      this.setData({
        [key]: row.enabled
      })
      toast.info(error.message || '设置保存失败')
    }
  },

  handleRowTap(event) {
    const { label } = event.currentTarget.dataset
    const row = this.findRowByLabel(label)

    if (!row || typeof wx === 'undefined') {
      return
    }

    if (row.disabled) {
      toast.info(row.message || '当前不可操作')
      return
    }

    if (row.id === 'clearCache' || row.action === 'clearCache') {
      this.handleClearCache()
      return
    }

    if (row.route) {
      wx.navigateTo({
        url: row.route
      })
      return
    }

    toast.info(row.message || '待接入')
  },

  async handleClearCache() {
    try {
      await profileService.clearProfileSettingsCache()
      toast.info('缓存已清除')
      this.loadSettings()
    } catch (error) {
      toast.info(error.message || '清除缓存失败')
    }
  },

  async handleLogout() {
    if (typeof wx === 'undefined') {
      return
    }

    try {
      await profileService.logoutProfile()
      wx.removeStorageSync('enjoy_token')
      wx.removeStorageSync('enjoy_user')
      wx.showToast({
        title: '已退出登录',
        icon: 'success'
      })
    } catch (error) {
      toast.info(error.message || '退出登录失败')
    }
  },

  getRow(sectionIndex, rowIndex) {
    const section = this.data.sections[sectionIndex]

    return section && section.rows ? section.rows[rowIndex] : null
  },

  findRowByLabel(label) {
    const sections = this.data.sections || []

    for (let sectionIndex = 0; sectionIndex < sections.length; sectionIndex += 1) {
      const rows = sections[sectionIndex].rows || []

      for (let rowIndex = 0; rowIndex < rows.length; rowIndex += 1) {
        if (rows[rowIndex].label === label) {
          return rows[rowIndex]
        }
      }
    }

    return null
  }
})
