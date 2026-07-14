const viewportLayout = require('../../utils/viewport-layout')

const ENTRY_BRAND_SINGLE_SCREEN_DESIGN = Object.freeze({
  designWidthRpx: 750,
  minTopbarHeightRpx: 182,
  navigationBottomGapRpx: 24,
  brandOnlineBottomOffsetRpx: 82,
  designContentHeightRpx: 1152,
  minLayoutScale: 0.78,
  maxLayoutScale: 1,
  heroSizeRpx: 400,
  heroTopRatio: 0.13,
  heroTopMinRpx: 96,
  heroTopMaxRpx: 184,
  titleWidthRpx: 400,
  titleHeightRpx: 260,
  titleTopGapRpx: 20,
  subtitleTopGapRpx: -50,
  subtitleFontSizeRpx: 20,
  subtitleLineHeightRpx: 24,
  subtitleLetterSpacingRpx: 12,
  accentTopRatio: 0.82,
  accentContentGapRpx: 80,
  accentSafeBottomGapRpx: 80,
  bounceLargeRpx: -72,
  bounceSmallRpx: -35,
  goFontSizeRpx: 28
})

function calculateEntryBrandSingleScreenLayout(metrics, designOverrides = {}) {
  if (!metrics) {
    return null
  }

  const design = Object.assign({}, ENTRY_BRAND_SINGLE_SCREEN_DESIGN, designOverrides)
  const topbarHeightRpx = viewportLayout.roundRpx(Math.max(
    design.minTopbarHeightRpx,
    metrics.navigationBottomRpx + design.navigationBottomGapRpx
  ))
  const contentHeightRpx = Math.max(
    0,
    viewportLayout.roundRpx(metrics.canvasHeightRpx - topbarHeightRpx)
  )
  const usableContentHeightRpx = Math.max(0, contentHeightRpx - metrics.safeBottomRpx)
  const scale = viewportLayout.clamp(
    usableContentHeightRpx / design.designContentHeightRpx,
    design.minLayoutScale,
    design.maxLayoutScale
  )
  const heroSizeRpx = viewportLayout.roundRpx(design.heroSizeRpx * scale)
  const heroTopRpx = viewportLayout.roundRpx(viewportLayout.clamp(
    usableContentHeightRpx * design.heroTopRatio,
    design.heroTopMinRpx * scale,
    design.heroTopMaxRpx * scale
  ))
  const titleWidthRpx = viewportLayout.roundRpx(design.titleWidthRpx * scale)
  const titleHeightRpx = viewportLayout.roundRpx(design.titleHeightRpx * scale)
  const titleTopGapRpx = viewportLayout.roundRpx(design.titleTopGapRpx * scale)
  const subtitleTopGapRpx = viewportLayout.roundRpx(design.subtitleTopGapRpx * scale)
  const subtitleFontSizeRpx = viewportLayout.roundRpx(design.subtitleFontSizeRpx * scale)
  const subtitleLineHeightRpx = viewportLayout.roundRpx(design.subtitleLineHeightRpx * scale)
  const subtitleLetterSpacingRpx = viewportLayout.roundRpx(design.subtitleLetterSpacingRpx * scale)
  const contentBlockBottomRpx = heroTopRpx
    + heroSizeRpx
    + titleTopGapRpx
    + titleHeightRpx
    + subtitleTopGapRpx
    + subtitleLineHeightRpx
  const accentTopRpx = viewportLayout.roundRpx(Math.min(
    Math.max(0, usableContentHeightRpx - design.accentSafeBottomGapRpx * scale),
    Math.max(
      contentBlockBottomRpx + design.accentContentGapRpx * scale,
      usableContentHeightRpx * design.accentTopRatio
    )
  ))

  return Object.assign({}, metrics, {
    scale,
    topbarHeightRpx,
    contentHeightRpx,
    usableContentHeightRpx,
    brandTopRpx: viewportLayout.roundRpx(
      metrics.capsuleBottomRpx
      - metrics.statusBarHeightRpx
      - design.brandOnlineBottomOffsetRpx
    ),
    heroSizeRpx,
    heroTopRpx,
    titleWidthRpx,
    titleHeightRpx,
    titleTopGapRpx,
    subtitleTopGapRpx,
    subtitleFontSizeRpx,
    subtitleLineHeightRpx,
    subtitleLetterSpacingRpx,
    accentTopRpx,
    bounceLargeRpx: viewportLayout.roundRpx(design.bounceLargeRpx * scale),
    bounceSmallRpx: viewportLayout.roundRpx(design.bounceSmallRpx * scale),
    goFontSizeRpx: viewportLayout.roundRpx(design.goFontSizeRpx * scale)
  })
}

function getEntryBrandSingleScreenLayout(designOverrides = {}) {
  const design = Object.assign({}, ENTRY_BRAND_SINGLE_SCREEN_DESIGN, designOverrides)
  const metrics = viewportLayout.getViewportMetrics({
    designWidthRpx: design.designWidthRpx
  })

  return calculateEntryBrandSingleScreenLayout(metrics, design)
}

module.exports = {
  ENTRY_BRAND_SINGLE_SCREEN_DESIGN,
  calculateEntryBrandSingleScreenLayout,
  getEntryBrandSingleScreenLayout
}
