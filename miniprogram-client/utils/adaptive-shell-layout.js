const viewportLayout = require('./viewport-layout')

const DESIGN_WIDTH_RPX = 750

function normalizeOptions(options = {}) {
  return Object.assign({
    contentMinTopRpx: 160,
    navigationBottomGapRpx: 13,
    titleHeightRpx: 50,
    backSizeRpx: 40
  }, options)
}

function getMetrics() {
  return viewportLayout.getViewportMetrics({ designWidthRpx: DESIGN_WIDTH_RPX })
}

function getFallbackMetrics() {
  return {
    canvasHeightRpx: 1620,
    safeBottomRpx: 0,
    navigationBottomRpx: 147,
    capsuleTopRpx: 108,
    capsuleLeftRpx: 584,
    capsuleHeightRpx: 64
  }
}

function resolveMetrics() {
  return getMetrics() || getFallbackMetrics()
}

function getProfileWhiteShellLayoutStyles(options = {}) {
  const design = normalizeOptions(options)
  const metrics = resolveMetrics()
  const contentTopRpx = viewportLayout.roundRpx(Math.max(
    design.contentMinTopRpx,
    metrics.navigationBottomRpx + design.navigationBottomGapRpx
  ))
  const contentHeightRpx = viewportLayout.roundRpx(Math.max(
    0,
    metrics.canvasHeightRpx - contentTopRpx - metrics.safeBottomRpx
  ))
  const titleTopRpx = viewportLayout.roundRpx(
    metrics.capsuleTopRpx + (metrics.capsuleHeightRpx - design.titleHeightRpx) / 2
  )
  const backTopRpx = viewportLayout.roundRpx(
    metrics.capsuleTopRpx + (metrics.capsuleHeightRpx - design.backSizeRpx) / 2
  )
  const rightLeftRpx = viewportLayout.roundRpx(Math.max(0, metrics.capsuleLeftRpx - 76))

  return {
    frameStyle: `height: ${metrics.canvasHeightRpx}rpx; min-height: ${metrics.canvasHeightRpx}rpx;`,
    navStyle: `height: ${contentTopRpx}rpx;`,
    contentStyle: `top: ${contentTopRpx}rpx; height: ${contentHeightRpx}rpx;`,
    titleStyle: `top: ${titleTopRpx}rpx; height: ${design.titleHeightRpx}rpx; line-height: ${design.titleHeightRpx}rpx;`,
    rightStyle: `top: ${titleTopRpx}rpx; left: ${rightLeftRpx}rpx; height: ${design.titleHeightRpx}rpx; line-height: ${design.titleHeightRpx}rpx;`,
    backStyle: `top: ${backTopRpx}rpx; width: ${design.backSizeRpx}rpx; height: ${design.backSizeRpx}rpx;`,
    safeBottomRpx: metrics.safeBottomRpx,
    canvasHeightRpx: metrics.canvasHeightRpx,
    contentTopRpx,
    capsuleTopRpx: metrics.capsuleTopRpx,
    capsuleHeightRpx: metrics.capsuleHeightRpx
  }
}

function getApplyShellLayoutStyles(options = {}) {
  const design = Object.assign({
    stageTopRpx: 108,
    contentMinTopRpx: 181,
    navigationBottomGapRpx: 13,
    frameBottomPaddingRpx: 10,
    topBackgroundOverlapRpx: 2,
    navLeftRpx: 23,
    navWidthRpx: 704,
    rightActionCapsuleGapRpx: 16,
    rightActionMinRightRpx: 96
  }, options)
  const metrics = resolveMetrics()
  const contentTopRpx = viewportLayout.roundRpx(Math.max(
    design.contentMinTopRpx,
    metrics.navigationBottomRpx + design.navigationBottomGapRpx
  ))
  const frameTopPaddingRpx = viewportLayout.roundRpx(Math.max(
    design.frameBottomPaddingRpx,
    contentTopRpx - design.stageTopRpx
  ))
  const phoneHeightRpx = viewportLayout.roundRpx(Math.max(
    0,
    metrics.canvasHeightRpx - contentTopRpx - design.frameBottomPaddingRpx
  ))
  const navTopRpx = viewportLayout.roundRpx(metrics.capsuleTopRpx - design.stageTopRpx)
  const rightActionRightRpx = viewportLayout.roundRpx(Math.max(
    design.rightActionMinRightRpx,
    design.navLeftRpx + design.navWidthRpx - metrics.capsuleLeftRpx + design.rightActionCapsuleGapRpx
  ))
  const topBackgroundTopRpx = viewportLayout.roundRpx(
    contentTopRpx + design.topBackgroundOverlapRpx
  )
  const topBackgroundHeightRpx = viewportLayout.roundRpx(
    contentTopRpx + design.topBackgroundOverlapRpx * 2
  )

  return {
    pageStyle: `height: ${metrics.canvasHeightRpx}rpx; min-height: ${metrics.canvasHeightRpx}rpx;`,
    frameStyle: `padding-top: ${frameTopPaddingRpx}rpx; padding-bottom: ${design.frameBottomPaddingRpx}rpx;`,
    topBgStyle: `top: -${topBackgroundTopRpx}rpx; height: ${topBackgroundHeightRpx}rpx;`,
    navStyle: `top: ${navTopRpx}rpx; height: ${metrics.capsuleHeightRpx}rpx;`,
    phoneStyle: `height: ${phoneHeightRpx}rpx; min-height: 0; padding-bottom: ${metrics.safeBottomRpx}rpx; box-sizing: border-box;`,
    switchStyle: `height: ${frameTopPaddingRpx}rpx;`,
    rightActionStyle: `right: ${rightActionRightRpx}rpx; height: ${metrics.capsuleHeightRpx}rpx;`,
    safeBottomRpx: metrics.safeBottomRpx,
    canvasHeightRpx: metrics.canvasHeightRpx,
    contentTopRpx,
    capsuleTopRpx: metrics.capsuleTopRpx,
    capsuleHeightRpx: metrics.capsuleHeightRpx
  }
}

function getLegacyWhiteFrameLayoutStyles(options = {}) {
  const design = Object.assign({
    contentLeftRpx: 2,
    contentWidthRpx: 750,
    contentMinTopRpx: 160,
    navigationBottomGapRpx: 13,
    bottomHeightRatio: 78 / 810,
    titleHeightRpx: 50,
    backSizeRpx: 40
  }, options)
  const metrics = resolveMetrics()
  const contentTopRpx = viewportLayout.roundRpx(Math.max(
    design.contentMinTopRpx,
    metrics.navigationBottomRpx + design.navigationBottomGapRpx
  ))
  const visualBottomHeightRpx = viewportLayout.roundRpx(
    metrics.canvasHeightRpx * design.bottomHeightRatio
  )
  const bottomHeightRpx = viewportLayout.roundRpx(
    visualBottomHeightRpx + metrics.safeBottomRpx
  )
  const bottomTopRpx = viewportLayout.roundRpx(Math.max(
    contentTopRpx,
    metrics.canvasHeightRpx - bottomHeightRpx
  ))
  const contentHeightRpx = viewportLayout.roundRpx(Math.max(0, bottomTopRpx - contentTopRpx))
  const titleTopRpx = viewportLayout.roundRpx(
    metrics.capsuleTopRpx + (metrics.capsuleHeightRpx - design.titleHeightRpx) / 2
  )
  const backTopRpx = viewportLayout.roundRpx(
    metrics.capsuleTopRpx + (metrics.capsuleHeightRpx - design.backSizeRpx) / 2
  )

  return {
    frameStyle: `height: ${metrics.canvasHeightRpx}rpx; min-height: ${metrics.canvasHeightRpx}rpx;`,
    contentStyle: [
      `left: ${design.contentLeftRpx}rpx`,
      `top: ${contentTopRpx}rpx`,
      `width: ${design.contentWidthRpx}rpx`,
      `height: ${contentHeightRpx}rpx`
    ].join('; '),
    bottomStyle: `top: ${bottomTopRpx}rpx; height: ${bottomHeightRpx}rpx;`,
    titleStyle: `top: ${titleTopRpx}rpx; height: ${design.titleHeightRpx}rpx; line-height: ${design.titleHeightRpx}rpx;`,
    backStyle: `top: ${backTopRpx}rpx; width: ${design.backSizeRpx}rpx; height: ${design.backSizeRpx}rpx;`,
    safeBottomRpx: metrics.safeBottomRpx,
    canvasHeightRpx: metrics.canvasHeightRpx,
    contentTopRpx,
    bottomTopRpx,
    bottomHeightRpx,
    capsuleTopRpx: metrics.capsuleTopRpx,
    capsuleHeightRpx: metrics.capsuleHeightRpx
  }
}

module.exports = {
  getProfileWhiteShellLayoutStyles,
  getApplyShellLayoutStyles,
  getLegacyWhiteFrameLayoutStyles
}
