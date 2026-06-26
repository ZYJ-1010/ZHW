const CONTENT_TOP_RPX = 160
const NAV_TITLE_HEIGHT_RPX = 50
const BACK_BUTTON_SIZE_RPX = 40
const DEFAULT_CAPSULE_BOTTOM_RPX = 142
const DEFAULT_FRAME_HEIGHT_RPX = 1620

function roundRpx(value) {
  return Math.round(value * 100) / 100
}

function getMenuCapsuleBottomRpx() {
  try {
    if (typeof wx !== 'undefined' && wx.getSystemInfoSync && wx.getMenuButtonBoundingClientRect) {
      const menuButton = wx.getMenuButtonBoundingClientRect()
      const systemInfo = wx.getSystemInfoSync()

      if (menuButton && systemInfo && systemInfo.windowWidth) {
        return roundRpx((menuButton.top + menuButton.height) * 750 / systemInfo.windowWidth)
      }
    }
  } catch (error) {
    return DEFAULT_CAPSULE_BOTTOM_RPX
  }

  return DEFAULT_CAPSULE_BOTTOM_RPX
}

function getLayoutStyles() {
  const capsuleBottom = getMenuCapsuleBottomRpx()
  const titleTop = Math.max(0, roundRpx(capsuleBottom - NAV_TITLE_HEIGHT_RPX))
  const backTop = Math.max(0, roundRpx(capsuleBottom - BACK_BUTTON_SIZE_RPX))
  let frameHeight = DEFAULT_FRAME_HEIGHT_RPX

  try {
    if (typeof wx !== 'undefined' && wx.getSystemInfoSync) {
      const systemInfo = wx.getSystemInfoSync()

      if (systemInfo && systemInfo.windowWidth && systemInfo.windowHeight) {
        frameHeight = roundRpx(systemInfo.windowHeight * 750 / systemInfo.windowWidth)
      }
    }
  } catch (error) {
    frameHeight = DEFAULT_FRAME_HEIGHT_RPX
  }

  return {
    frameStyle: `height: ${frameHeight}rpx; min-height: ${frameHeight}rpx;`,
    contentStyle: `top: ${CONTENT_TOP_RPX}rpx; height: calc(100% - ${CONTENT_TOP_RPX}rpx);`,
    titleStyle: `top: ${titleTop}rpx; height: ${NAV_TITLE_HEIGHT_RPX}rpx; line-height: ${NAV_TITLE_HEIGHT_RPX}rpx;`,
    rightStyle: `top: ${titleTop}rpx; height: ${NAV_TITLE_HEIGHT_RPX}rpx; line-height: ${NAV_TITLE_HEIGHT_RPX}rpx;`,
    backStyle: `top: ${backTop}rpx; width: ${BACK_BUTTON_SIZE_RPX}rpx; height: ${BACK_BUTTON_SIZE_RPX}rpx;`
  }
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
    rightText: {
      type: String,
      value: ''
    },
    rightColor: {
      type: String,
      value: '#101010'
    }
  },

  data: {
    layout: getLayoutStyles()
  },

  lifetimes: {
    attached() {
      this.setData({
        layout: getLayoutStyles()
      })
    }
  },

  methods: {
    handleBack() {
      const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []

      if (pages.length > 1) {
        wx.navigateBack()
        return
      }

      wx.redirectTo({
        url: '/pages/profile/index'
      })
    }
  }
})
