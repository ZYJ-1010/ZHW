Component({
  options: {
    addGlobalClass: true,
    styleIsolation: 'shared'
  },

  properties: {
    status: {
      type: Object,
      value: {}
    },
    variant: {
      type: String,
      value: ''
    },
    countdownLabel: {
      type: String,
      value: '响应倒计时'
    },
    countdownText: {
      type: String,
      value: ''
    },
    referralText: {
      type: String,
      value: '邀请你参与组局'
    }
  }
})
