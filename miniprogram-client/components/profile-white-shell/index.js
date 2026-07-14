const { navigateShellRoute } = require('../../utils/shell-nav')
const { getProfileWhiteShellLayoutStyles } = require('../../utils/adaptive-shell-layout')

let cachedLayout = null

function getLayoutStyles() {
  if (cachedLayout) {
    return cachedLayout
  }

  cachedLayout = getProfileWhiteShellLayoutStyles()

  return cachedLayout
}

Component({
  options: {
    multipleSlots: true
  },

  properties: {
    title: {
      type: String,
      value: ''
    },
    background: {
      type: String,
      value: '#f8fafd'
    },
    navBackground: {
      type: String,
      value: '#ffffff'
    },
    titleColor: {
      type: String,
      value: '#101010'
    },
    titleWeight: {
      type: String,
      value: '700'
    },
    titleSize: {
      type: Number,
      value: 36
    },
    backBackground: {
      type: String,
      value: '#ffffff'
    },
    backBorderColor: {
      type: String,
      value: 'rgba(51, 51, 51, 0.28)'
    },
    backIconColor: {
      type: String,
      value: 'rgba(51, 51, 51, 0.71)'
    },
    showBack: {
      type: Boolean,
      value: true
    },
    backUrl: {
      type: String,
      value: ''
    },
    rightText: {
      type: String,
      value: ''
    },
    rightType: {
      type: String,
      value: ''
    },
    rightColor: {
      type: String,
      value: '#101010'
    },
    rightWidth: {
      type: Number,
      value: 60
    }
  },

  data: {
    layout: getLayoutStyles()
  },

  lifetimes: {
    attached() {
      this.updateLayout()
    },
    ready() {
      this.updateLayout()
    }
  },

  pageLifetimes: {
    show() {
      this.updateLayout()
    },
    resize() {
      this.updateLayout()
    }
  },

  methods: {
    updateLayout() {
      cachedLayout = null
      this.setData({
        layout: getLayoutStyles()
      })
    },

    handleBack() {
      if (this.properties.backUrl) {
        navigateShellRoute(this.properties.backUrl)
        return
      }

      const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []

      if (pages.length > 1) {
        wx.navigateBack()
        return
      }

      navigateShellRoute('/pages/profile/index')
    },

    handleRightTap() {
      this.triggerEvent('righttap')
    }
  }
})
