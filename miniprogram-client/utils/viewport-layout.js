const DEFAULT_DESIGN_WIDTH_RPX = 750

function roundRpx(value) {
  return Math.round(value * 100) / 100
}

function clamp(value, min, max) {
  return Math.min(max, Math.max(min, value))
}

function getWindowInfo() {
  if (typeof wx === 'undefined') {
    return null
  }

  if (typeof wx.getWindowInfo === 'function') {
    return wx.getWindowInfo()
  }

  if (typeof wx.getSystemInfoSync === 'function') {
    return wx.getSystemInfoSync()
  }

  return null
}

function calculateViewportMetrics(options = {}) {
  const windowInfo = options.windowInfo || null
  const menuButton = options.menuButton || null
  const designWidthRpx = Number(options.designWidthRpx) || DEFAULT_DESIGN_WIDTH_RPX
  const windowWidth = Number(windowInfo && windowInfo.windowWidth)
  const windowHeight = Number(windowInfo && (windowInfo.windowHeight || windowInfo.screenHeight))

  if (!windowWidth || !windowHeight || !menuButton || !menuButton.width || !menuButton.height) {
    return null
  }

  const ratio = designWidthRpx / windowWidth
  const statusBarHeightPx = Number(windowInfo.statusBarHeight) || Math.max(0, menuButton.top - 4)
  const capsuleTopGapPx = Math.max(0, menuButton.top - statusBarHeightPx)
  const navigationHeightPx = capsuleTopGapPx * 2 + menuButton.height
  const screenHeight = Number(windowInfo.screenHeight || windowHeight)
  const safeAreaBottom = Number(windowInfo.safeArea && windowInfo.safeArea.bottom)

  return {
    ratio,
    designWidthRpx,
    windowWidth,
    windowHeight,
    canvasHeightRpx: roundRpx(windowHeight * ratio),
    statusBarHeightRpx: roundRpx(statusBarHeightPx * ratio),
    navigationHeightRpx: roundRpx(navigationHeightPx * ratio),
    navigationBottomRpx: roundRpx((statusBarHeightPx + navigationHeightPx) * ratio),
    capsuleTopRpx: roundRpx(menuButton.top * ratio),
    capsuleBottomRpx: roundRpx((menuButton.top + menuButton.height) * ratio),
    capsuleLeftRpx: roundRpx(menuButton.left * ratio),
    capsuleRightRpx: roundRpx(Math.max(0, windowWidth - menuButton.left - menuButton.width) * ratio),
    capsuleWidthRpx: roundRpx(menuButton.width * ratio),
    capsuleHeightRpx: roundRpx(menuButton.height * ratio),
    safeBottomRpx: safeAreaBottom > 0
      ? roundRpx(Math.max(0, screenHeight - safeAreaBottom) * ratio)
      : 0
  }
}

function getViewportMetrics(options = {}) {
  if (typeof wx === 'undefined' || typeof wx.getMenuButtonBoundingClientRect !== 'function') {
    return null
  }

  return calculateViewportMetrics({
    windowInfo: getWindowInfo(),
    menuButton: wx.getMenuButtonBoundingClientRect(),
    designWidthRpx: options.designWidthRpx
  })
}

module.exports = {
  DEFAULT_DESIGN_WIDTH_RPX,
  roundRpx,
  clamp,
  calculateViewportMetrics,
  getViewportMetrics
}
