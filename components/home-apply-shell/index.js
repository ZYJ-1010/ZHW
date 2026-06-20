const APPLY_STAGE_TOP_RPX = 108
const APPLY_DEFAULT_CONTENT_TOP_RPX = 181
const APPLY_NAV_BOTTOM_GAP_RPX = 13
const APPLY_FRAME_BOTTOM_PADDING_RPX = 10
const APPLY_DEFAULT_NAV_TOP_RPX = 108
const APPLY_DEFAULT_NAV_HEIGHT_RPX = 64

function roundRpx(value) {
  return Math.round(value * 100) / 100
}

function getApplyShellLayoutStyles() {
  let contentTop = APPLY_DEFAULT_CONTENT_TOP_RPX
  let navTop = APPLY_DEFAULT_NAV_TOP_RPX
  let navHeight = APPLY_DEFAULT_NAV_HEIGHT_RPX

  try {
    if (typeof wx !== 'undefined' && wx.getSystemInfoSync && wx.getMenuButtonBoundingClientRect) {
      const menuButton = wx.getMenuButtonBoundingClientRect()
      const systemInfo = wx.getSystemInfoSync()

      if (menuButton && menuButton.height && systemInfo && systemInfo.windowWidth) {
        const ratio = 750 / systemInfo.windowWidth
        navTop = roundRpx(menuButton.top * ratio)
        navHeight = roundRpx(menuButton.height * ratio)
        const capsuleBottom = (menuButton.top + menuButton.height) * ratio
        contentTop = Math.max(
          APPLY_DEFAULT_CONTENT_TOP_RPX,
          roundRpx(capsuleBottom + APPLY_NAV_BOTTOM_GAP_RPX)
        )
      }
    }
  } catch (error) {
    contentTop = APPLY_DEFAULT_CONTENT_TOP_RPX
  }

  const frameTopPadding = Math.max(
    APPLY_FRAME_BOTTOM_PADDING_RPX,
    roundRpx(contentTop - APPLY_STAGE_TOP_RPX)
  )

  return {
    frameStyle: `padding-top: ${frameTopPadding}rpx; padding-bottom: ${APPLY_FRAME_BOTTOM_PADDING_RPX}rpx;`,
    topBgStyle: `top: -${contentTop}rpx; height: ${contentTop}rpx;`,
    navStyle: `top: ${roundRpx(navTop - APPLY_STAGE_TOP_RPX)}rpx; height: ${navHeight}rpx;`,
    phoneStyle: `min-height: calc(100vh - ${contentTop}rpx);`,
    switchStyle: `height: ${frameTopPadding}rpx;`
  }
}

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

    handlePreviewSwitch(event) {
      const direction = Number(event.currentTarget.dataset.direction) || 1

      this.triggerEvent('previewswitch', { direction })
    }
  }
})
