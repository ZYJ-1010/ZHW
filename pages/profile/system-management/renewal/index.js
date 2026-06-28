const toast = require('../../../../utils/toast')

const ASSET_BASE = '/pages/profile/system-management/block-settings/assets'

Page({
  data: {
    icons: {
      clock: `${ASSET_BASE}/icon-clock.svg`
    },
    selectedDays: 30,
    options: [
      { days: 30, label: '标准周期' },
      { days: 60, label: '双倍保护' },
      { days: 90, label: '季度保护' }
    ],
    rules: [
      { prefix: '保护期最长 ', strong: '90天', suffix: '，到期需重新续期' },
      { prefix: '续期后立即生效，', strong: '5秒内', suffix: ' 覆盖所有新请求' },
      { prefix: '到期前 ', strong: '3天', suffix: ' 将发送提醒通知' }
    ]
  },

  handleOptionTap(event) {
    this.setData({
      selectedDays: Number(event.currentTarget.dataset.days)
    })
  },

  handleConfirmTap() {
    toast.success(`已确认续期 ${this.data.selectedDays} 天`)
  }
})
