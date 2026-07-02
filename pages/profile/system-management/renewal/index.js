const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

const ASSET_BASE = '/pages/profile/system-management/block-settings/assets'

Page({
  data: {
    icons: {
      clock: `${ASSET_BASE}/icon-clock.svg`
    },
    blockSettings: null,
    selectedDays: 0,
    remainingDaysText: '',
    options: [],
    rules: []
  },

  onLoad() {
    this.loadRenewalConfig()
  },

  async loadRenewalConfig() {
    try {
      const data = await profileService.getSystemBlockSettings()
      const options = Array.isArray(data.renewalOptions) ? data.renewalOptions : []
      const renewalDays = Number(data.renewalDays || (options[0] && options[0].days) || 0)

      this.setData({
        blockSettings: data || null,
        selectedDays: renewalDays,
        remainingDaysText: renewalDays ? `当前剩余 ${renewalDays} 天` : '',
        options,
        rules: Array.isArray(data.renewalRules) ? data.renewalRules : []
      })
    } catch (error) {
      this.setData({ options: [], rules: [], remainingDaysText: '' })
      toast.info(error.message || '续期配置暂时不可用')
    }
  },

  handleOptionTap(event) {
    this.setData({
      selectedDays: Number(event.currentTarget.dataset.days)
    })
  },

  async handleConfirmTap() {
    if (!this.data.selectedDays) {
      toast.info('请选择续期时长')
      return
    }
    try {
      await profileService.saveSystemBlockSettings({
        ...(this.data.blockSettings || {}),
        renewalDays: this.data.selectedDays
      })
      toast.success(`已确认续期 ${this.data.selectedDays} 天`)
    } catch (error) {
      toast.info(error.message || '保存失败')
    }
  }
})
