const profileService = require('../../../services/profile')

const authService = require('../../../services/auth')
const { navigateShellRoute } = require('../../../utils/shell-nav')

const ASSET_BASE = '/pages/profile/settings/assets'

function asset(name) {
  return `${ASSET_BASE}/${name}`
}

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

function normalizeSections(sections) {
  const source = Array.isArray(sections) ? sections : []

  return source.map((section) => ({
    title: section.title || '设置',
    rows: (Array.isArray(section.rows) ? section.rows : []).map((row) => {
      const iconKey = row.iconKey || row.id

      return Object.assign({}, row, {
        iconKey,
        icon: row.icon || ICON_MAP[iconKey] || ICON_MAP.about,
        enabled: Boolean(row.enabled)
      })
    })
  }))
}

Page({
  data: {
    sections: [],
    chevronIcon: asset('icon-chevron-right.svg'),
    previewMode: 'static',
    isSaving: false,
    loadError: ''
  },

  onLoad(options = {}) {
    this.setData({
      previewMode: options.mode || 'static'
    })
    this.loadSettings()
  },

  async loadSettings() {
    try {
      const settings = await profileService.getProfileSettings()
      this.setData({
        sections: normalizeSections(settings && settings.sections),
        loadError: ''
      })
    } catch (error) {
      this.setData({
        sections: [],
        loadError: error.message || '系统设置加载失败'
      })
      this.showToast(error.message || '系统设置加载失败')
    }
  },

  async handleSwitch(event) {
    if (this.data.isSaving) {
      return
    }

    const { sectionIndex, rowIndex } = event.currentTarget.dataset
    const key = `sections[${sectionIndex}].rows[${rowIndex}].enabled`
    const current = this.data.sections[sectionIndex].rows[rowIndex].enabled

    this.setData({
      [key]: !current,
      isSaving: true
    })

    try {
      const saved = await profileService.saveProfileSettings({
        sections: this.data.sections
      })
      this.setData({
        sections: normalizeSections(saved && saved.sections)
      })
    } catch (error) {
      this.setData({ [key]: current })
      this.showToast(error.message || '设置保存失败')
    } finally {
      this.setData({ isSaving: false })
    }
  },

  handleRowTap(event) {
    const { sectionIndex, rowIndex } = event.currentTarget.dataset
    const row = this.data.sections[sectionIndex] && this.data.sections[sectionIndex].rows[rowIndex]

    if (!row || row.switch) {
      return
    }

    if (row.disabledReason) {
      this.showToast(row.disabledReason)
      return
    }

    if (row.action === 'clear_cache') {
      this.clearCache()
      return
    }

    if (row.route) {
      navigateShellRoute(row.route)
      return
    }

    if (row.agreementKey) {
      navigateShellRoute(`/pages/profile/system-management/agreement-detail/index?agreement=${encodeURIComponent(row.agreementKey)}&title=${encodeURIComponent(row.label || '协议详情')}`)
      return
    }

    this.showToast('该设置项暂未开放')
  },

  clearCache() {
    try {
      wx.clearStorageSync()
      this.showToast('缓存已清理')
    } catch (error) {
      this.showToast('缓存清理失败')
    }
  },

  handleLogout() {
    wx.showModal({
      title: '退出登录',
      content: '退出后需要重新通过邀请入口登录。',
      confirmText: '退出',
      confirmColor: '#ff4d57',
      success: (res) => {
        if (!res.confirm) {
          return
        }

        authService.logout()
        wx.reLaunch({
          url: '/pages/login/index'
        })
      }
    })
  },

  showToast(title) {
    if (typeof wx === 'undefined') {
      return
    }

    wx.showToast({
      title,
      icon: 'none'
    })
  }
})
