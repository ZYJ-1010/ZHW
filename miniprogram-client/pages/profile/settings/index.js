const profileService = require('../../../services/profile')

const authService = require('../../../services/auth')
const { navigateShellRoute } = require('../../../utils/shell-nav')
const { toUserMessage } = require('../../../utils/user-message')

const ASSET_BASE = '/pages/profile/settings/assets'

function asset(name) {
  return `${ASSET_BASE}/${name}`
}

const ICON_MAP = {
  payPassword: asset('icon-pay-lock.png'),
  loginPassword: asset('icon-login-lock.png'),
  wechatBind: asset('icon-users.png'),
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
        loadError: toUserMessage(error && error.message, '系统设置加载失败')
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

    if (row.action === 'bind_wechat') {
      this.bindWechat()
      return
    }

    if (row.action === 'set_password') {
      this.setLoginPassword()
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

  async bindWechat() {
    try {
      await authService.bindWechatAccount()
      this.showToast('微信已绑定')
      this.loadSettings()
    } catch (error) {
      this.showToast(error.message || '微信绑定失败')
    }
  },

  setLoginPassword() {
    wx.showModal({
      title: '设置登录密码',
      editable: true,
      placeholderText: '8-64 位字母和数字',
      success: async (res) => {
        if (!res.confirm) return
        const password = String(res.content || '')
        if (!/^[A-Za-z0-9]{8,64}$/.test(password)) {
          this.showToast('密码须为8-64位字母和数字')
          return
        }
        try {
          await authService.setLoginPassword(password)
          this.showToast('登录密码已设置')
          this.loadSettings()
        } catch (error) {
          this.showToast(error.message || '设置密码失败')
        }
      }
    })
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

  handleDeleteAccount() {
    wx.showModal({
      title: '确认注销账号',
      content: '注销后将解除微信和手机号绑定，当前邀请码不可继续使用。后续注册需重新获取邀请码。',
      confirmText: '确认注销',
      confirmColor: '#ff4d57',
      success: async (res) => {
        if (!res.confirm) {
          return
        }
        try {
          await authService.deleteAccount()
          wx.showModal({
            title: '账号已注销',
            content: '请重新获取邀请码后再注册。',
            showCancel: false,
            success: () => wx.reLaunch({ url: '/pages/entry/index' })
          })
        } catch (error) {
          this.showToast(error.message || '账号注销失败，请稍后重试')
        }
      }
    })
  },

  showToast(title) {
    if (typeof wx === 'undefined') {
      return
    }

    wx.showToast({
      title: toUserMessage(title),
      icon: 'none'
    })
  }
})
