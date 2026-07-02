const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

const ASSET_BASE = '/pages/profile/system-management/block-settings/assets'

Page({
  data: {
    icons: {
      lock: `${ASSET_BASE}/icon-lock.svg`,
      check: `${ASSET_BASE}/icon-check.svg`
    },
    blockSettings: null,
    mode: '',
    modes: [],
    activeRules: []
  },

  onLoad(options) {
    const nextMode = options && options.mode ? options.mode : ''
    this.loadProtectionMode(nextMode)
  },

  async loadProtectionMode(fallbackMode) {
    try {
      const data = await profileService.getSystemBlockSettings()
      const modes = Array.isArray(data.protectionModes) ? data.protectionModes : []
      const selectedMode = data.protectionMode || fallbackMode || (modes[0] && modes[0].key) || ''

      this.setData({
        blockSettings: data || null,
        modes
      })
      this.applyMode(selectedMode)
    } catch (error) {
      this.setData({ modes: [], activeRules: [], mode: '' })
      toast.info(error.message || '保护模式暂时不可用')
    }
  },

  applyMode(mode) {
    const active = this.data.modes.find((item) => item.key === mode) || this.data.modes[0]

    this.setData({
      mode: active ? active.key : '',
      activeRules: active && Array.isArray(active.rules) ? active.rules : []
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
      await profileService.saveSystemBlockSettings({
        ...(this.data.blockSettings || {}),
        protectionMode: this.data.mode
      })
      toast.success('保护模式已保存')
    } catch (error) {
      toast.info(error.message || '保存失败')
    }
  }
})
