const { getApplyShellLayoutStyles } = require('../../utils/adaptive-shell-layout')

Component({
  options: {
    multipleSlots: true,
    addGlobalClass: true,
    styleIsolation: 'shared'
  },

  properties: {
    navTitle: {
      type: String,
      value: '标题'
    },
    shellClass: {
      type: String,
      value: ''
    },
    showBack: {
      type: Boolean,
      value: false
    },
    rightText: {
      type: String,
      value: ''
    },
    previewSwitchEnabled: {
      type: Boolean,
      value: false
    }
  },

  data: {
    applyShellLayout: getApplyShellLayoutStyles()
  },

  lifetimes: {
    attached() {
      this.updateApplyShellLayout()
    },
    ready() {
      this.updateApplyShellLayout()
    }
  },

  pageLifetimes: {
    show() {
      this.updateApplyShellLayout()
    },
    resize() {
      this.updateApplyShellLayout()
    }
  },

  methods: {
    updateApplyShellLayout() {
      this.setData({
        applyShellLayout: getApplyShellLayoutStyles()
      })
    },

    handleBackTap() {
      this.triggerEvent('backtap')
    },

    handleRightTap() {
      this.triggerEvent('righttap')
    },

    handlePreviewSwitch(event) {
      const direction = Number(event.currentTarget.dataset.direction) || 1

      this.triggerEvent('previewswitch', { direction })
    }
  }
})
