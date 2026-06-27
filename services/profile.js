const profileApi = require('../api/modules/profile')

async function getProfileHome() {
  const result = await profileApi.getProfileHome()

  if (result.code !== 0) {
    throw new Error(result.message || '获取个人中心失败')
  }

  return result.data
}

async function getPointsMall() {
  const result = await profileApi.getPointsMall()

  if (result.code !== 0) {
    throw new Error(result.message || '获取积分商城失败')
  }

  return result.data
}

async function exchangePointsMallGood(payload = {}) {
  const goodId = String(payload.goodId || payload.productId || '').trim()

  if (!goodId) {
    throw new Error('请选择兑换商品')
  }

  const result = await profileApi.exchangePointsMallGood({
    goodId
  })

  return {
    success: result.code === 0,
    code: result.code,
    message: result.code === 0
      ? (result.data && result.data.message || result.message || '兑换成功')
      : (result.message || '兑换失败'),
    data: result.data
  }
}
async function getPointsOrders(params = {}) {
  const result = await profileApi.getPointsOrders(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取订单列表失败')
  }

  return result.data
}

async function getPointsOrderLogistics(params = {}) {
  const orderId = String(params.orderId || params.id || '').trim()

  if (!orderId) {
    throw new Error('缺少订单信息，无法查看物流')
  }

  const result = await profileApi.getPointsOrderLogistics(encodeURIComponent(orderId))

  if (result.code !== 0) {
    throw new Error(result.message || '获取物流详情失败')
  }

  return result.data
}


module.exports = {
  getProfileHome,
  getPointsMall,
  exchangePointsMallGood,
  getPointsOrders,
  getPointsOrderLogistics
}
