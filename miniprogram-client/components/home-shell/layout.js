const viewportLayout = require('../../utils/viewport-layout')

const HOME_SHELL_FIXED_FRAME_DESIGN = Object.freeze({
  designWidthRpx: 750,
  minTopbarHeightRpx: 182,
  navigationBottomGapRpx: 13,
  toolbarHeightRpx: 58,
  toolbarCapsuleGapRpx: 14,
  pageTitleLineHeightRpx: 44,
  brandOnlineBottomOffsetRpx: 82,
  dockHeightRpx: 218,
  dockTopOverlapRpx: 58,
  dockBottomOverflowRpx: 2
})

function calculateHomeShellFixedFrameLayout(metrics, options = {}) {
  if (!metrics) {
    return null
  }

  const design = Object.assign({}, HOME_SHELL_FIXED_FRAME_DESIGN, options.design || {})
  const dockVisible = options.dockVisible !== false
  const contentBottomGapRpx = Math.max(0, Number(options.contentBottomGapRpx || 0))
  const contentTopRpx = viewportLayout.roundRpx(Math.max(
    design.minTopbarHeightRpx,
    metrics.navigationBottomRpx + design.navigationBottomGapRpx
  ))
  const topbarHeightRpx = viewportLayout.roundRpx(Math.max(
    design.minTopbarHeightRpx,
    contentTopRpx + 1
  ))
  const dockHeightRpx = dockVisible
    ? viewportLayout.roundRpx(design.dockHeightRpx + metrics.safeBottomRpx)
    : 0
  const dockTopRpx = dockVisible
    ? viewportLayout.roundRpx(Math.max(
      contentTopRpx,
      metrics.canvasHeightRpx - dockHeightRpx + design.dockBottomOverflowRpx
    ))
    : metrics.canvasHeightRpx
  // 内容不能延伸到固定导航之下；否则底部卡片会被导航遮住，且被遮住
  // 的区域会被导航层截获点击事件。
  const contentBottomRpx = dockVisible
    ? dockTopRpx
    : viewportLayout.roundRpx(Math.max(
      contentTopRpx,
      metrics.canvasHeightRpx - metrics.safeBottomRpx - contentBottomGapRpx
    ))
  const contentHeightRpx = viewportLayout.roundRpx(Math.max(0, contentBottomRpx - contentTopRpx))
  const toolbarTopRpx = viewportLayout.roundRpx(
    metrics.capsuleTopRpx + (metrics.capsuleHeightRpx - design.toolbarHeightRpx) / 2
  )
  const pageTitleTopRpx = viewportLayout.roundRpx(Math.max(
    0,
    toolbarTopRpx - metrics.statusBarHeightRpx
      + (design.toolbarHeightRpx - design.pageTitleLineHeightRpx) / 2
  ))

  return Object.assign({}, metrics, {
    contentTopRpx,
    topbarHeightRpx,
    dockTopRpx,
    dockHeightRpx,
    contentHeightRpx,
    toolbarTopRpx,
    toolbarRightRpx: viewportLayout.roundRpx(
      metrics.capsuleRightRpx + metrics.capsuleWidthRpx + design.toolbarCapsuleGapRpx
    ),
    pageTitleTopRpx,
    brandTopRpx: viewportLayout.roundRpx(
      metrics.capsuleBottomRpx
      - metrics.statusBarHeightRpx
      - design.brandOnlineBottomOffsetRpx
    )
  })
}

function getHomeShellFixedFrameLayout(options = {}) {
  const design = Object.assign({}, HOME_SHELL_FIXED_FRAME_DESIGN, options.design || {})
  const metrics = viewportLayout.getViewportMetrics({
    designWidthRpx: design.designWidthRpx
  })

  return calculateHomeShellFixedFrameLayout(metrics, Object.assign({}, options, { design }))
}

module.exports = {
  HOME_SHELL_FIXED_FRAME_DESIGN,
  calculateHomeShellFixedFrameLayout,
  getHomeShellFixedFrameLayout
}
