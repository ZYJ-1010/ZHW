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
    referralText: {
      type: String,
      value: '邀请你参与组局'
    }
  }
})
