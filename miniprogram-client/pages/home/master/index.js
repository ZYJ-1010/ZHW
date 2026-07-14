const APPLY_STAGE_TOP_RPX = 108
const APPLY_DEFAULT_CONTENT_TOP_RPX = 181
const APPLY_NAV_BOTTOM_GAP_RPX = 13
const APPLY_FRAME_BOTTOM_PADDING_RPX = 10
const APPLY_DEFAULT_NAV_TOP_RPX = 108
const APPLY_DEFAULT_NAV_HEIGHT_RPX = 64
const WHITE_CONTENT_LEFT_RPX = 2
const WHITE_CONTENT_TOP_RPX = 160
const WHITE_CONTENT_WIDTH_RPX = 750
const WHITE_DESIGN_FRAME_HEIGHT_PT = 810
const WHITE_DESIGN_BOTTOM_HEIGHT_PT = 78
const WHITE_NAV_TITLE_HEIGHT_RPX = 50
const WHITE_BACK_BUTTON_SIZE_RPX = 40
const WHITE_DEFAULT_CAPSULE_BOTTOM_RPX = 142
const WHITE_DEFAULT_FRAME_HEIGHT_RPX = WHITE_DESIGN_FRAME_HEIGHT_PT * 2
const { navigateShellKey } = require('../../../utils/shell-nav')
const {
  getApplyShellLayoutStyles: getAdaptiveApplyShellLayoutStyles,
  getLegacyWhiteFrameLayoutStyles
} = require('../../../utils/adaptive-shell-layout')

function roundRpx(value) {
  return Math.round(value * 100) / 100
}

function getApplyShellLayoutStyles() {
  return getAdaptiveApplyShellLayoutStyles()
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
    return WHITE_DEFAULT_CAPSULE_BOTTOM_RPX
  }

  return WHITE_DEFAULT_CAPSULE_BOTTOM_RPX
}

function getWhiteShellLayoutStyles() {
  return getLegacyWhiteFrameLayoutStyles()
}

Page({
  data: {
    masterMode: 'default',
    brand: '真好玩',
    onlineText: '在线',
    dockVisible: true,
    contentScrollY: false,
    shellClass: '',
    shellVariant: '',
    toolbarActionsVisible: true,
    applyMaster: {
      navTitle: '标题'
    },
    whiteMaster: {
      navTitle: '标题'
    },
    applyShellLayout: getApplyShellLayoutStyles(),
    whiteShellLayout: getWhiteShellLayoutStyles(),
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ]
  },

  onLoad(options = {}) {
    if (options.mode === 'apply') {
      this.setData({
        masterMode: 'apply',
        dockVisible: false,
        contentScrollY: false,
        shellClass: '',
        shellVariant: '',
        applyShellLayout: getApplyShellLayoutStyles()
      })
      return
    }

    if (options.mode === 'whiteBackground') {
      this.setData({
        masterMode: 'whiteBackground',
        dockVisible: false,
        contentScrollY: false,
        shellClass: '',
        shellVariant: '',
        toolbarActionsVisible: false,
        whiteShellLayout: getWhiteShellLayoutStyles()
      })
      return
    }

    if (options.mode === 'contentOnly') {
      this.setData({
        masterMode: 'contentOnly',
        dockVisible: false,
        contentScrollY: true,
        shellClass: 'home-shell--content-only',
        shellVariant: '',
        toolbarActionsVisible: true
      })
      return
    }

    if (options.mode === 'joinApply') {
      this.setData({
        masterMode: 'joinApply',
        dockVisible: true,
        contentScrollY: false,
        shellClass: '',
        shellVariant: 'joinApply',
        toolbarActionsVisible: false
      })
      return
    }

    if (options.mode === 'topNoBrand') {
      this.setData({
        masterMode: 'topNoBrand',
        dockVisible: true,
        contentScrollY: false,
        shellClass: '',
        shellVariant: 'topNoBrand',
        toolbarActionsVisible: false
      })
    }
  },

  onShow() {
    this.refreshPreviewLayout()
  },

  onResize() {
    this.refreshPreviewLayout()
  },

  refreshPreviewLayout() {
    this.setData({
      applyShellLayout: getApplyShellLayoutStyles(),
      whiteShellLayout: getWhiteShellLayoutStyles()
    })
  },

  handleShellNavTap(event) {
    const { key } = event.detail

    if (!key) {
      return
    }

    navigateShellKey(key, {
      currentRoute: 'pages/home/master/index'
    })
  }
})
