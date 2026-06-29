const APPLY_STAGE_TOP_RPX = 108
const APPLY_DEFAULT_CONTENT_TOP_RPX = 181
const APPLY_NAV_BOTTOM_GAP_RPX = 13
const APPLY_FRAME_BOTTOM_PADDING_RPX = 10
const APPLY_DEFAULT_NAV_TOP_RPX = 108
const APPLY_DEFAULT_NAV_HEIGHT_RPX = 64
const APPLY_NAV_LEFT_RPX = 23
const APPLY_NAV_WIDTH_RPX = 704
const APPLY_DEFAULT_RIGHT_ACTION_RIGHT_RPX = 170
const APPLY_RIGHT_ACTION_CAPSULE_GAP_RPX = 16

function roundRpx(value) {
  return Math.round(value * 100) / 100
}

function getApplyShellLayoutStyles() {
  let contentTop = APPLY_DEFAULT_CONTENT_TOP_RPX
  let navTop = APPLY_DEFAULT_NAV_TOP_RPX
  let navHeight = APPLY_DEFAULT_NAV_HEIGHT_RPX
  let rightActionRight = APPLY_DEFAULT_RIGHT_ACTION_RIGHT_RPX

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

        if (typeof menuButton.left === 'number') {
          const capsuleLeft = menuButton.left * ratio
          rightActionRight = Math.max(
            96,
            roundRpx(APPLY_NAV_LEFT_RPX + APPLY_NAV_WIDTH_RPX - capsuleLeft + APPLY_RIGHT_ACTION_CAPSULE_GAP_RPX)
          )
        }
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
    phoneStyle: `height: calc(100vh - ${roundRpx(contentTop + APPLY_FRAME_BOTTOM_PADDING_RPX)}rpx); min-height: 0;`,
    switchStyle: `height: ${frameTopPadding}rpx;`,
    rightActionStyle: `right: ${rightActionRight}rpx; height: ${navHeight}rpx;`
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
