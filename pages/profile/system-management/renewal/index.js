const toast = require('../../../../utils/toast')
const profileService = require('../../../../services/profile')

const ASSET_BASE = '/pages/profile/system-management/block-settings/assets'

Page({
  data: {
    icons: {
      clock: `${ASSET_BASE}/icon-clock.svg`
    },
    selectedDays: '',
    options: [],
    rules: []
  },

  onLoad() {
    this.loadRenewalOptions()
  },

  async loadRenewalOptions() {
    try {
      const data = await profileService.getSystemBlockRenewalOptions()
      const options = this.normalizeList(data.options || data.list || data.items)
      const selected = data.selectedDays || data.defaultDays || options[0] && options[0].days || ''

      this.setData({
        selectedDays: selected,
        options,
        rules: this.normalizeList(data.rules || data.ruleLines)
      })
    } catch (error) {
      this.setData({
        selectedDays: '',
        options: [],
        rules: []
      })
      toast.info(error.message || '续期选项加载失败')
    }
  },

  handleOptionTap(event) {
    this.setData({
      selectedDays: Number(event.currentTarget.dataset.days)
    })
  },

  async handleConfirmTap() {
    if (!this.data.selectedDays) {
      toast.info('请选择续期天数')
      return
    }

    try {
      await profileService.renewSystemBlockSettings({
        days: this.data.selectedDays
      })
      toast.success('续期已提交')
    } catch (error) {
      toast.info(error.message || '续期保护期失败')
    }
  },

  normalizeList(list) {
    return Array.isArray(list) ? list : []
  }
})
