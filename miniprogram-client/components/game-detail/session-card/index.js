Component({
  options: {
    addGlobalClass: true,
    styleIsolation: 'shared'
  },

  properties: {
    items: {
      type: Array,
      value: []
    },
    titleWeight: {
      type: String,
      value: ''
    },
    cardClass: {
      type: String,
      value: ''
    },
    showTitleIcon: {
      type: Boolean,
      value: true
    },
    showItemIcon: {
      type: Boolean,
      value: true
    },
    showMapButton: {
      type: Boolean,
      value: true
    }
  },

  methods: {
    handleMapTap(event) {
      this.triggerEvent('maptap', {
        item: event.currentTarget.dataset.item
      })
    }
  }
})
