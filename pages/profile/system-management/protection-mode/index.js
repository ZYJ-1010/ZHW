const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

const ASSET_BASE = '/pages/profile/system-management/block-settings/assets'

Page({
  data: {
    icons: {
      lock: `${ASSET_BASE}/icon-lock.svg`,
      check: `${ASSET_BASE}/icon-check.svg`
    },
    mode: '',
    modes: [],
    activeRules: []
  },

  onLoad(options) {
    this.loadProtectionMode(options && options.mode)
  },

  async loadProtectionMode(optionMode) {
    try {
      const data = await profileService.getSystemBlockSettings()
      const protection = data.protectionMode || data.protection || {}
      const modes = Array.isArray(protection.modes || data.modes) ? (protection.modes || data.modes) : []
      const mode = optionMode || protection.mode || protection.currentMode || data.mode || ''

      this.setData({
        modes
      })
      this.applyMode(mode)
    } catch (error) {
      this.setData({
        mode: '',
        modes: [],
        activeRules: []
      })
      toast.info(error.message || '保护模式加载失败')
    }
  },

  applyMode(mode) {
    const active = this.data.modes.find((item) => item.key === mode) || this.data.modes[0] || {}

    this.setData({
      mode: active.key || '',
      activeRules: Array.isArray(active.rules) ? active.rules : []
    })
  },

  handleModeTap(event) {
    const { mode } = event.currentTarget.dataset
    this.applyMode(mode)
  },

  async handleConfirmTap() {
    if (!this.data.mode) {
      toast.info('请选择保护模式')
      return
    }

    try {
      await profileService.saveSystemProtectionMode({
        mode: this.data.mode
      })
      toast.success('保护模式已保存')
    } catch (error) {
      toast.info(error.message || '保护模式保存失败')
    }
  }
})
