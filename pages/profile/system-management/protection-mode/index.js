const toast = require('../../../../utils/toast')

const ASSET_BASE = '/pages/profile/system-management/block-settings/assets'

Page({
  data: {
    icons: {
      lock: `${ASSET_BASE}/icon-lock.svg`,
      check: `${ASSET_BASE}/icon-check.svg`
    },
    mode: 'hard',
    modes: [
      {
        key: 'hard',
        title: '硬保护',
        desc: '完全过滤，用户无感知',
        rules: [
          { prefix: '', strong: '完全过滤', suffix: '，同类行家内容不展示' },
          { prefix: '被保护用户 ', strong: '零曝光', suffix: '' },
          { prefix: '适合竞争激烈的同城市', strong: '', suffix: '' }
        ]
      },
      {
        key: 'soft',
        title: '软保护',
        desc: '排序降权×0.1，推至第5页后',
        rules: [
          { prefix: '排序权重 ', strong: '×0.1', suffix: '，大幅降权' },
          { prefix: '翻页至局列表第5页后', strong: '不受保护', suffix: '' },
          { prefix: '适合内容不足或过渡期', strong: '', suffix: '' }
        ]
      }
    ],
    activeRules: []
  },

  onLoad(options) {
    const nextMode = options && options.mode === 'soft' ? 'soft' : 'hard'
    this.applyMode(nextMode)
  },

  applyMode(mode) {
    const active = this.data.modes.find((item) => item.key === mode) || this.data.modes[0]

    this.setData({
      mode: active.key,
      activeRules: active.rules
    })
  },

  handleModeTap(event) {
    const { mode } = event.currentTarget.dataset
    this.applyMode(mode)
  },

  handleConfirmTap() {
    toast.success(this.data.mode === 'hard' ? '已选择硬保护' : '已选择软保护')
  }
})
