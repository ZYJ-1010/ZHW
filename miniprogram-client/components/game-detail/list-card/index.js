Component({
  options: {
    addGlobalClass: true,
    styleIsolation: 'shared'
  },

  properties: {
    title: {
      type: String,
      value: ''
    },
    rows: {
      type: Array,
      value: []
    },
    cardClass: {
      type: String,
      value: ''
    },
    listClass: {
      type: String,
      value: ''
    }
  }
})
