const profileApi = require('../api/modules/profile')

async function getProfileHome() {
  const result = await profileApi.getProfileHome()

  if (result.code !== 0) {
    throw new Error(result.message || '获取个人中心失败')
  }

  return result.data
}

async function replyServiceReview(payload = {}) {
  const reviewId = String(payload.reviewId || '').trim()
  const content = String(payload.content || '').trim()

  if (!reviewId) {
    throw new Error('缺少评价信息，无法提交回复')
  }

  if (!content) {
    throw new Error('请输入回复内容')
  }

  if (content.length > 200) {
    throw new Error('回复内容不能超过200字')
  }

  const result = await profileApi.replyServiceReview(reviewId, {
    content
  })

  if (result.code !== 0) {
    throw new Error(result.message || '回复评价提交失败')
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

async function saveSystemProfileInfo(payload = {}) {
  const result = await profileApi.saveSystemProfileInfo(payload)

  if (result.code !== 0) {
    throw new Error(result.message || '资料保存失败')
  }

  return result.data
}

async function getSystemSkillConfig() {
  const result = await profileApi.getSystemSkillConfig()

  if (result.code !== 0) {
    throw new Error(result.message || '获取技能配置失败')
  }

  return result.data
}

async function saveSystemSkillConfig(payload = {}) {
  const result = await profileApi.saveSystemSkillConfig(payload)

  if (result.code !== 0) {
    throw new Error(result.message || '技能配置保存失败')
  }

  return result.data
}

module.exports = {
  getProfileHome,
  replyServiceReview,
  getPointsMall,
  exchangePointsMallGood,
  getPointsOrders,
  getPointsOrderLogistics,
  saveSystemProfileInfo,
  getSystemSkillConfig,
  saveSystemSkillConfig
}
