const CONTENT_TOP_RPX = 160
const NAV_TITLE_HEIGHT_RPX = 50
const BACK_BUTTON_SIZE_RPX = 40
const DEFAULT_CAPSULE_BOTTOM_RPX = 142
const DEFAULT_CAPSULE_LEFT_RPX = 584
const DEFAULT_FRAME_HEIGHT_RPX = 1620

const { navigateShellRoute } = require('../../utils/shell-nav')

let cachedLayout = null

function roundRpx(value) {
  return Math.round(value * 100) / 100
}

function getWindowMetrics() {
  try {
    if (typeof wx !== 'undefined' && wx.getWindowInfo) {
      return wx.getWindowInfo()
    }
  } catch (error) {
    return null
  }

  try {
    if (typeof wx !== 'undefined' && wx.getSystemInfoSync) {
      return wx.getSystemInfoSync()
    }
  } catch (error) {
    return null
  }

  return null
}

function getMenuCapsuleLayoutRpx() {
  try {
    if (typeof wx !== 'undefined' && wx.getMenuButtonBoundingClientRect) {
      const menuButton = wx.getMenuButtonBoundingClientRect()
      const windowInfo = getWindowMetrics()

      if (menuButton && windowInfo && windowInfo.windowWidth) {
        return {
          bottom: roundRpx((menuButton.top + menuButton.height) * 750 / windowInfo.windowWidth),
          left: roundRpx(menuButton.left * 750 / windowInfo.windowWidth)
        }
      }
    }
  } catch (error) {
    return {
      bottom: DEFAULT_CAPSULE_BOTTOM_RPX,
      left: DEFAULT_CAPSULE_LEFT_RPX
    }
  }

  return {
    bottom: DEFAULT_CAPSULE_BOTTOM_RPX,
    left: DEFAULT_CAPSULE_LEFT_RPX
  }
}

function getLayoutStyles() {
  if (cachedLayout) {
    return cachedLayout
  }

  const capsuleLayout = getMenuCapsuleLayoutRpx()
  const titleTop = Math.max(0, roundRpx(capsuleLayout.bottom - NAV_TITLE_HEIGHT_RPX))
  const backTop = Math.max(0, roundRpx(capsuleLayout.bottom - BACK_BUTTON_SIZE_RPX))
  const rightLeft = Math.max(0, roundRpx(capsuleLayout.left - 76))
  let frameHeight = DEFAULT_FRAME_HEIGHT_RPX

  const windowInfo = getWindowMetrics()

  if (windowInfo && windowInfo.windowWidth && windowInfo.windowHeight) {
    frameHeight = roundRpx(windowInfo.windowHeight * 750 / windowInfo.windowWidth)
  }

  cachedLayout = {
    frameStyle: `height: ${frameHeight}rpx; min-height: ${frameHeight}rpx;`,
    contentStyle: `top: ${CONTENT_TOP_RPX}rpx; height: calc(100% - ${CONTENT_TOP_RPX}rpx);`,
    titleStyle: `top: ${titleTop}rpx; height: ${NAV_TITLE_HEIGHT_RPX}rpx; line-height: ${NAV_TITLE_HEIGHT_RPX}rpx;`,
    rightStyle: `top: ${titleTop}rpx; left: ${rightLeft}rpx; height: ${NAV_TITLE_HEIGHT_RPX}rpx; line-height: ${NAV_TITLE_HEIGHT_RPX}rpx;`,
    backStyle: `top: ${backTop}rpx; width: ${BACK_BUTTON_SIZE_RPX}rpx; height: ${BACK_BUTTON_SIZE_RPX}rpx;`
  }

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

  pageLifetimes: {
    resize() {
      cachedLayout = null
      this.setData({
        layout: getLayoutStyles()
      })
    }
  },

  methods: {
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
