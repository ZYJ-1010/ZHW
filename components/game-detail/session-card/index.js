Component({
  options: {
    addGlobalClass: true,
    styleIsolation: 'shared'
  },

  properties: {
    items: {
      type: Array,
      value: []
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
