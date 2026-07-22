const locationService = require('../services/location')

let denialGuideShown = false

function callWx(name, options = {}) {
  return new Promise((resolve, reject) => {
    if (typeof wx === 'undefined' || typeof wx[name] !== 'function') {
      reject(new Error('当前环境不支持位置功能'))
      return
    }
    wx[name]({ ...options, success: resolve, fail: reject })
  })
}

async function getPreciseLocation() {
  const result = await callWx('getLocation', { type: 'gcj02' })
  const location = {
    latitude: Number(result.latitude),
    longitude: Number(result.longitude),
    accuracyMeter: Number(result.accuracy || result.horizontalAccuracy || 0)
  }
  if (!Number.isFinite(location.latitude) || !Number.isFinite(location.longitude)) {
    throw new Error('定位结果异常')
  }
  await locationService.saveCurrentLocation(location).catch(() => {})
  return location
}

async function chooseManualLocation() {
  const place = await callWx('chooseLocation')
  const location = {
    latitude: Number(place.latitude),
    longitude: Number(place.longitude),
    address: String(place.address || place.name || '').trim(),
    cityName: '',
    cityCode: '',
    accuracyMeter: 0
  }
  if (!Number.isFinite(location.latitude) || !Number.isFinite(location.longitude)) {
    throw new Error('手动位置无效')
  }
  try {
    const data = await locationService.reverseGeocode({ latitude: location.latitude, longitude: location.longitude })
    const placeInfo = data.place || data || {}
    location.cityName = String(placeInfo.city || '').trim()
    location.cityCode = String(placeInfo.cityCode || '').trim()
    location.address = location.address || String(placeInfo.address || placeInfo.title || '').trim()
  } catch (error) {
    // 手动选点本身仍可作为附近查询坐标，地址解析失败不阻断流程。
  }
  await locationService.saveManualLocation(location).catch(() => {})
  return location
}

async function getFallbackLocation() {
  return locationService.getLocationFallback()
}

async function showDeniedGuide({ onManual, onFallback }) {
  const choices = ['前往设置开启定位', '手动选择位置', '使用同城推荐']
  if (denialGuideShown) {
    if (typeof onFallback === 'function') return onFallback()
    return null
  }
  denialGuideShown = true
  const result = await callWx('showActionSheet', { itemList: choices }).catch(() => null)
  if (!result) return null
  if (result.tapIndex === 0) {
    await callWx('openSetting').catch(() => {})
    denialGuideShown = false
    return null
  }
  if (result.tapIndex === 1 && typeof onManual === 'function') return onManual()
  if (result.tapIndex === 2 && typeof onFallback === 'function') return onFallback()
  return null
}

module.exports = {
  chooseManualLocation,
  getFallbackLocation,
  getPreciseLocation,
  showDeniedGuide
}
