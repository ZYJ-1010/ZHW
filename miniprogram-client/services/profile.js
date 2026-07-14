const profileApi = require('../api/modules/profile')
const { createAuthExpiredError, isAuthExpiredResult } = require('../utils/auth-error')

async function getProfileHome() {
  const result = await profileApi.getProfileHome()

  if (isAuthExpiredResult(result)) {
    throw createAuthExpiredError(result.message)
  }

  if (result.code !== 0) {
    throw new Error(result.message || '获取个人中心失败')
  }

  return result.data
}

async function getServiceReviews(params = {}) {
  const result = await profileApi.getServiceReviews(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取评价列表失败')
  }

  return result.data
}

async function getServiceReviewDetail(reviewId) {
  const id = String(reviewId || '').trim()

  if (!id) {
    throw new Error('缺少评价信息，无法查看详情')
  }

  const result = await profileApi.getServiceReviewDetail(encodeURIComponent(id))

  if (result.code !== 0) {
    throw new Error(result.message || '获取评价详情失败')
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

async function likeServiceReview(reviewId) {
  const id = String(reviewId || '').trim()

  if (!id) {
    throw new Error('缺少评价信息，无法点赞')
  }

  const result = await profileApi.likeServiceReview(encodeURIComponent(id))

  if (result.code !== 0) {
    throw new Error(result.message || '评价点赞失败')
  }

  return result.data
}

async function getServiceReviewActions(reviewId) {
  const id = String(reviewId || '').trim()

  if (!id) {
    throw new Error('缺少评价信息，无法查看操作')
  }

  const result = await profileApi.getServiceReviewActions(encodeURIComponent(id))

  if (result.code !== 0) {
    throw new Error(result.message || '获取评价操作失败')
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

async function getPointsSummary() {
  const result = await profileApi.getPointsSummary()

  if (result.code !== 0) {
    throw new Error(result.message || '获取积分概览失败')
  }

  return result.data
}

async function getPointsLogs(params = {}) {
  const result = await profileApi.getPointsLogs(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取积分明细失败')
  }

  return result.data
}

async function getGrowth() {
  const result = await profileApi.getGrowth()

  if (result.code !== 0) {
    throw new Error(result.message || '获取成长数据失败')
  }

  return result.data
}

async function submitMemberRadarAction(payload = {}) {
  const action = String(payload.action || '').trim()

  if (!action) {
    throw new Error('请选择可用的人脉雷达操作')
  }

  const formRows = Array.isArray(payload.formRows) ? payload.formRows : []

  if (formRows.length > 20) {
    throw new Error('适配信息过多，请精简后再保存')
  }

  const result = await profileApi.submitMemberRadarAction({
    action,
    targetId: String(payload.targetId || '').trim(),
    targetUserId: Number(payload.targetUserId || 0) || 0,
    matchMode: payload.matchMode || 'all',
    matchCriteria: payload.matchCriteria || {},
    formRows
  })

  if (result.code !== 0) {
    throw new Error(result.message || '人脉雷达操作失败')
  }

  return result.data
}

async function exchangePointsMallGood(payload = {}) {
  const goodId = String(payload.goodId || payload.productId || '').trim()

  if (!goodId) {
    throw new Error('请选择兑换商品')
  }

  const result = await profileApi.exchangePointsMallGood({
    itemId: Number(goodId) || goodId
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

async function getProfileAssets() {
  const result = await profileApi.getProfileAssets()

  if (result.code !== 0) {
    throw new Error(result.message || '获取资产中心失败')
  }

  return result.data
}

async function getCreditCenter() {
  const result = await profileApi.getCreditCenter()

  if (result.code !== 0) {
    throw new Error(result.message || '获取信用中心失败')
  }

  return result.data
}

async function getPointsOrderDetail(params = {}) {
  const orderId = String(params.orderId || params.id || '').trim()

  if (!orderId) {
    throw new Error('缺少订单信息，无法查看详情')
  }

  const result = await profileApi.getPointsOrderDetail(encodeURIComponent(orderId))

  if (result.code !== 0) {
    throw new Error(result.message || '获取订单详情失败')
  }

  return result.data
}

async function cancelPointsOrder(params = {}) {
  const orderId = String(params.orderId || params.id || '').trim()

  if (!orderId) {
    throw new Error('缺少订单信息，无法取消订单')
  }

  const result = await profileApi.cancelPointsOrder(encodeURIComponent(orderId), {
    reason: params.reason || '用户主动取消'
  })

  if (result.code !== 0) {
    throw new Error(result.message || '取消订单失败')
  }

  return result.data
}

async function getInviteOverview() {
  const result = await profileApi.getInviteOverview()

  if (result.code !== 0) {
    throw new Error(result.message || '获取邀请概览失败')
  }

  return result.data
}

async function getInviteNetwork() {
  const result = await profileApi.getInviteNetwork()

  if (result.code !== 0) {
    throw new Error(result.message || '获取关系网络失败')
  }

  return result.data
}

async function getInviteRecords(params = {}) {
  const result = await profileApi.getInviteRecords(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取邀约记录失败')
  }

  return result.data
}

async function getInviteRanking(params = {}) {
  const result = await profileApi.getInviteRanking(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取贡献排行失败')
  }

  return result.data
}

async function getInviteIncome(params = {}) {
  const result = await profileApi.getInviteIncome(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取收益明细失败')
  }

  return result.data
}

async function getInviteMemberDetail(params = {}) {
  const memberId = String(params.memberId || params.id || '').trim()

  if (!memberId) {
    throw new Error('缺少成员信息')
  }

  const result = await profileApi.getInviteMemberDetail(encodeURIComponent(memberId))

  if (result.code !== 0) {
    throw new Error(result.message || '获取成员详情失败')
  }

  return result.data
}

async function getSystemProfileInfo() {
  const result = await profileApi.getSystemProfileInfo()

  if (result.code !== 0) {
    throw new Error(result.message || '获取资料失败')
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

async function getSystemServiceCaseDetail(caseId) {
  const id = String(caseId || '').trim()

  if (!id) {
    throw new Error('缺少案例信息')
  }

  const result = await profileApi.getSystemServiceCaseDetail(encodeURIComponent(id))

  if (result.code !== 0) {
    throw new Error(result.message || '获取案例详情失败')
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

async function getSystemFeedbackHome() {
  const result = await profileApi.getSystemFeedbackHome()

  if (result.code !== 0) {
    throw new Error(result.message || '获取反馈配置失败')
  }

  return result.data
}

async function submitSystemFeedback(payload = {}) {
  const typeKey = String(payload.typeKey || '').trim()
  const sessionKey = String(payload.sessionKey || '').trim()
  const content = String(payload.content || '').trim()
  const contact = String(payload.contact || '').trim()
  const fileIds = Array.isArray(payload.fileIds) ? payload.fileIds : []

  if (!typeKey) {
    throw new Error('请选择反馈类型')
  }

  if (!content && !fileIds.length) {
    throw new Error('请输入反馈内容或添加附件')
  }

  if (content.length > 500) {
    throw new Error('反馈内容不能超过500字')
  }

  if (contact.length > 80) {
    throw new Error('联系方式不能超过80字')
  }

  if (fileIds.length > 9) {
    throw new Error('最多上传9个反馈附件')
  }

  const result = await profileApi.submitSystemFeedback({
    typeKey,
    sessionKey,
    content,
    contact,
    fileIds,
    quick: Boolean(payload.quick)
  })

  if (result.code !== 0) {
    throw new Error(result.message || '反馈提交失败')
  }

  return result.data
}

async function getSystemFeedbackRecords(params = {}) {
  const result = await profileApi.getSystemFeedbackRecords(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取反馈记录失败')
  }

  return result.data
}

async function getSystemFeedbackDetail(recordId) {
  const id = String(recordId || '').trim()

  if (!id) {
    throw new Error('缺少反馈记录')
  }

  const result = await profileApi.getSystemFeedbackDetail(encodeURIComponent(id))

  if (result.code !== 0) {
    throw new Error(result.message || '获取反馈详情失败')
  }

  return result.data
}

async function appendSystemFeedbackMessage(recordId, payload = {}) {
  const id = String(recordId || '').trim()
  const content = String(payload.content || '').trim()
  const fileIds = Array.isArray(payload.fileIds) ? payload.fileIds : []

  if (!id) {
    throw new Error('缺少反馈记录')
  }

  if (!content && !fileIds.length) {
    throw new Error('请输入补充说明或添加附件')
  }

  if (content.length > 500) {
    throw new Error('补充说明不能超过500字')
  }

  const result = await profileApi.appendSystemFeedbackMessage(encodeURIComponent(id), {
    content,
    fileIds
  })

  if (result.code !== 0) {
    throw new Error(result.message || '补充说明提交失败')
  }

  return result.data
}

async function getSystemBlockSettings() {
  const result = await profileApi.getSystemBlockSettings()

  if (result.code !== 0) {
    throw new Error(result.message || '获取屏蔽设置失败')
  }

  return result.data
}

async function saveSystemBlockSettings(payload = {}) {
  const result = await profileApi.saveSystemBlockSettings(payload)

  if (result.code !== 0) {
    throw new Error(result.message || '屏蔽设置保存失败')
  }

  return result.data
}

async function getProfileSettings() {
  const result = await profileApi.getProfileSettings()

  if (result.code !== 0) {
    throw new Error(result.message || '获取系统设置失败')
  }

  return result.data
}

async function saveProfileSettings(payload = {}) {
  const result = await profileApi.saveProfileSettings(payload)

  if (result.code !== 0) {
    throw new Error(result.message || '保存系统设置失败')
  }

  return result.data
}

async function getProfileAgreements() {
  const result = await profileApi.getProfileAgreements()

  if (result.code !== 0) {
    throw new Error(result.message || '获取协议列表失败')
  }

  return result.data
}

async function getProfileAgreementDetail(agreementKey) {
  const key = String(agreementKey || '').trim()

  if (!key) {
    throw new Error('缺少协议信息')
  }

  const result = await profileApi.getProfileAgreementDetail(encodeURIComponent(key))

  if (result.code !== 0) {
    throw new Error(result.message || '获取协议详情失败')
  }

  return result.data
}

async function signProfileAgreement(agreementKey) {
  const key = String(agreementKey || '').trim()

  if (!key) {
    throw new Error('缺少协议信息')
  }

  const result = await profileApi.signProfileAgreement(encodeURIComponent(key), {})

  if (result.code !== 0) {
    throw new Error(result.message || '签署协议失败')
  }

  return result.data
}

module.exports = {
  getProfileHome,
  getProfileAssets,
  getCreditCenter,
  getServiceReviews,
  getServiceReviewDetail,
  replyServiceReview,
  likeServiceReview,
  getServiceReviewActions,
  getPointsSummary,
  getPointsLogs,
  getGrowth,
  getPointsMall,
  exchangePointsMallGood,
  getPointsOrders,
  getPointsOrderLogistics,
  getPointsOrderDetail,
  cancelPointsOrder,
  getInviteOverview,
  getInviteNetwork,
  getInviteRecords,
  getInviteRanking,
  getInviteIncome,
  getInviteMemberDetail,
  getSystemProfileInfo,
  saveSystemProfileInfo,
  getSystemSkillConfig,
  getSystemServiceCaseDetail,
  saveSystemSkillConfig,
  getSystemFeedbackHome,
  submitSystemFeedback,
  getSystemFeedbackRecords,
  getSystemFeedbackDetail,
  appendSystemFeedbackMessage,
  getSystemBlockSettings,
  saveSystemBlockSettings,
  getProfileSettings,
  saveProfileSettings,
  getProfileAgreements,
  getProfileAgreementDetail,
  signProfileAgreement
}
