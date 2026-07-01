const profileApi = require('../api/modules/profile')

async function getProfileHome() {
  const result = await profileApi.getProfileHome()

  if (result.code !== 0) {
    throw new Error(result.message || '获取个人中心失败')
  }

  return result.data
}

async function getMemberCenterConfig(params = {}) {
  const result = await profileApi.getMemberCenterConfig(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取会员中心配置失败')
  }

  return result.data
}

async function getMemberRadarOverview(params = {}) {
  const result = await profileApi.getMemberRadarOverview(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取人脉雷达概览失败')
  }

  return result.data
}

async function getMemberRadarProfile(params = {}) {
  const result = await profileApi.getMemberRadarProfile(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取适配信息失败')
  }

  return result.data
}

async function saveMemberRadarProfile(payload = {}) {
  const result = await profileApi.saveMemberRadarProfile(payload)

  if (result.code !== 0) {
    throw new Error(result.message || '保存适配信息失败')
  }

  return result.data
}

async function startMemberRadarMatch(payload = {}) {
  const result = await profileApi.startMemberRadarMatch(payload)

  if (result.code !== 0) {
    throw new Error(result.message || '发起适配失败')
  }

  return result.data
}

async function getMemberRadarMatch(params = {}) {
  const matchId = String(params.matchId || params.id || '').trim()

  if (!matchId) {
    throw new Error('缺少适配任务信息')
  }

  const result = await profileApi.getMemberRadarMatch(encodeURIComponent(matchId))

  if (result.code !== 0) {
    throw new Error(result.message || '获取适配任务失败')
  }

  return result.data
}

async function getMemberRadarMatchResults(params = {}) {
  const matchId = String(params.matchId || params.id || '').trim()

  if (!matchId) {
    throw new Error('缺少适配任务信息')
  }

  const result = await profileApi.getMemberRadarMatchResults(encodeURIComponent(matchId), params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取适配结果失败')
  }

  return result.data
}

async function followMemberRadarResult(params = {}) {
  const resultId = String(params.resultId || params.id || '').trim()

  if (!resultId) {
    throw new Error('缺少适配对象信息')
  }

  const result = await profileApi.followMemberRadarResult(encodeURIComponent(resultId), params)

  if (result.code !== 0) {
    throw new Error(result.message || '关注失败')
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

async function getProfileAssets() {
  const result = await profileApi.getProfileAssets()

  if (result.code !== 0) {
    throw new Error(result.message || '获取资产信息失败')
  }

  return result.data
}

async function getProfilePoints(params = {}) {
  const result = await profileApi.getProfilePoints(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取积分信息失败')
  }

  return result.data
}

async function getProfileGames(params = {}) {
  const result = await profileApi.getProfileGames(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取我的局失败')
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

async function getServiceReviewDetail(params = {}) {
  const reviewId = String(params.reviewId || params.id || '').trim()

  if (!reviewId) {
    throw new Error('缺少评价信息')
  }

  const result = await profileApi.getServiceReviewDetail(encodeURIComponent(reviewId))

  if (result.code !== 0) {
    throw new Error(result.message || '获取评价详情失败')
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

async function getInviteRecords(params = {}) {
  const result = await profileApi.getInviteRecords(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取邀请记录失败')
  }

  return result.data
}

async function getInviteIncome(params = {}) {
  const result = await profileApi.getInviteIncome(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取邀请收益失败')
  }

  return result.data
}

async function getInviteNetwork(params = {}) {
  const result = await profileApi.getInviteNetwork(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取邀请网络失败')
  }

  return result.data
}

async function getInviteRanking(params = {}) {
  const result = await profileApi.getInviteRanking(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取邀请排行失败')
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

async function getProfileCredit() {
  const result = await profileApi.getProfileCredit()

  if (result.code !== 0) {
    throw new Error(result.message || '获取信用中心失败')
  }

  return result.data
}

async function getProfileAchievements(params = {}) {
  const result = await profileApi.getProfileAchievements(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取成就墙失败')
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

async function getSystemProfileInfo() {
  const result = await profileApi.getSystemProfileInfo()

  if (result.code !== 0) {
    throw new Error(result.message || '获取资料设置失败')
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

async function getSystemSkillCaseDetail(params = {}) {
  const caseId = String(params.caseId || params.id || '').trim()

  if (!caseId) {
    throw new Error('缺少案例信息')
  }

  const result = await profileApi.getSystemSkillCaseDetail(encodeURIComponent(caseId))

  if (result.code !== 0) {
    throw new Error(result.message || '获取案例详情失败')
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

async function saveSystemBlockStatus(payload = {}) {
  const result = await profileApi.saveSystemBlockStatus({
    enabled: !!payload.enabled
  })

  if (result.code !== 0) {
    throw new Error(result.message || '屏蔽设置保存失败')
  }

  return result.data
}

async function saveSystemProtectionMode(payload = {}) {
  const mode = String(payload.mode || '').trim()

  if (!mode) {
    throw new Error('请选择保护模式')
  }

  const result = await profileApi.saveSystemProtectionMode({
    mode
  })

  if (result.code !== 0) {
    throw new Error(result.message || '保护模式保存失败')
  }

  return result.data
}

async function saveSystemBlockScenes(payload = {}) {
  const scenes = Array.isArray(payload.scenes) ? payload.scenes : []

  if (!scenes.length) {
    throw new Error('缺少分场景配置')
  }

  const result = await profileApi.saveSystemBlockScenes({
    scenes
  })

  if (result.code !== 0) {
    throw new Error(result.message || '分场景配置保存失败')
  }

  return result.data
}

async function getSystemBlockWhitelist(params = {}) {
  const result = await profileApi.getSystemBlockWhitelist(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取白名单失败')
  }

  return result.data
}

async function addSystemBlockWhitelist(payload = {}) {
  const expertId = String(payload.expertId || payload.id || '').trim()

  if (!expertId) {
    throw new Error('请选择要添加的行家')
  }

  const result = await profileApi.addSystemBlockWhitelist({
    expertId
  })

  if (result.code !== 0) {
    throw new Error(result.message || '添加白名单失败')
  }

  return result.data
}

async function removeSystemBlockWhitelist(payload = {}) {
  const expertId = String(payload.expertId || payload.id || '').trim()

  if (!expertId) {
    throw new Error('缺少白名单行家信息')
  }

  const result = await profileApi.removeSystemBlockWhitelist(encodeURIComponent(expertId))

  if (result.code !== 0) {
    throw new Error(result.message || '移除白名单失败')
  }

  return result.data
}

async function getSystemBlockedUsers(params = {}) {
  const result = await profileApi.getSystemBlockedUsers(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取屏蔽用户失败')
  }

  return result.data
}

async function addSystemBlockedUsers(payload = {}) {
  const userIds = Array.isArray(payload.userIds) ? payload.userIds.filter(Boolean) : []

  if (!userIds.length) {
    throw new Error('请选择要屏蔽的用户')
  }

  const result = await profileApi.addSystemBlockedUsers({
    userIds
  })

  if (result.code !== 0) {
    throw new Error(result.message || '添加屏蔽用户失败')
  }

  return result.data
}

async function removeSystemBlockedUser(payload = {}) {
  const userId = String(payload.userId || payload.id || '').trim()

  if (!userId) {
    throw new Error('缺少屏蔽用户信息')
  }

  const result = await profileApi.removeSystemBlockedUser(encodeURIComponent(userId))

  if (result.code !== 0) {
    throw new Error(result.message || '解除屏蔽失败')
  }

  return result.data
}

async function getSystemBlockRenewalOptions() {
  const result = await profileApi.getSystemBlockRenewalOptions()

  if (result.code !== 0) {
    throw new Error(result.message || '获取续期选项失败')
  }

  return result.data
}

async function renewSystemBlockSettings(payload = {}) {
  const days = Number(payload.days)

  if (!Number.isFinite(days) || days <= 0) {
    throw new Error('请选择续期天数')
  }

  const result = await profileApi.renewSystemBlockSettings({
    days
  })

  if (result.code !== 0) {
    throw new Error(result.message || '续期保护期失败')
  }

  return result.data
}

async function getSystemBlockKeywords(params = {}) {
  const result = await profileApi.getSystemBlockKeywords(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取关键词屏蔽失败')
  }

  return result.data
}

async function addSystemBlockKeyword(payload = {}) {
  const keyword = String(payload.keyword || payload.text || '').trim()

  if (!keyword) {
    throw new Error('请输入关键词')
  }

  const result = await profileApi.addSystemBlockKeyword({
    keyword
  })

  if (result.code !== 0) {
    throw new Error(result.message || '添加关键词失败')
  }

  return result.data
}

async function removeSystemBlockKeyword(payload = {}) {
  const keywordId = String(payload.keywordId || payload.id || '').trim()

  if (!keywordId) {
    throw new Error('缺少关键词信息')
  }

  const result = await profileApi.removeSystemBlockKeyword(encodeURIComponent(keywordId))

  if (result.code !== 0) {
    throw new Error(result.message || '删除关键词失败')
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

async function getSystemFeedbackDetail(params = {}) {
  const feedbackId = String(params.feedbackId || params.id || '').trim()

  if (!feedbackId) {
    throw new Error('缺少反馈信息')
  }

  const result = await profileApi.getSystemFeedbackDetail(encodeURIComponent(feedbackId))

  if (result.code !== 0) {
    throw new Error(result.message || '获取反馈详情失败')
  }

  return result.data
}

async function getSystemFeedbackOptions() {
  const result = await profileApi.getSystemFeedbackOptions()

  if (result.code !== 0) {
    throw new Error(result.message || '获取反馈配置失败')
  }

  return result.data
}

async function getSystemFeedbackGames(params = {}) {
  const result = await profileApi.getSystemFeedbackGames(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取可反馈组局失败')
  }

  return result.data
}

async function submitSystemFeedback(payload = {}) {
  const type = String(payload.type || payload.feedbackType || '').trim()
  const content = String(payload.content || '').trim()

  if (!type) {
    throw new Error('请选择反馈类型')
  }

  if (!content) {
    throw new Error('请填写反馈内容')
  }

  const result = await profileApi.submitSystemFeedback(payload)

  if (result.code !== 0) {
    throw new Error(result.message || '反馈提交失败')
  }

  return result.data
}

async function getCreditAppealOptions(params = {}) {
  const result = await profileApi.getCreditAppealOptions(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取信用申诉配置失败')
  }

  return result.data
}

async function submitCreditAppeal(payload = {}) {
  const reason = String(payload.reason || payload.reasonKey || '').trim()
  const content = String(payload.content || '').trim()

  if (!reason) {
    throw new Error('请选择申诉原因')
  }

  if (!content) {
    throw new Error('请填写详细说明')
  }

  const result = await profileApi.submitCreditAppeal(payload)

  if (result.code !== 0) {
    throw new Error(result.message || '信用申诉提交失败')
  }

  return result.data
}

async function getSystemReportRecords(params = {}) {
  const result = await profileApi.getSystemReportRecords(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取举报记录失败')
  }

  return result.data
}

async function getSystemReportOptions() {
  const result = await profileApi.getSystemReportOptions()

  if (result.code !== 0) {
    throw new Error(result.message || '获取举报配置失败')
  }

  return result.data
}

async function submitSystemReport(payload = {}) {
  const type = String(payload.type || payload.reportType || '').trim()
  const reportedUser = String(payload.reportedUser || payload.reportedUserId || '').trim()
  const reason = String(payload.reason || '').trim()

  if (!type) {
    throw new Error('请选择举报类型')
  }

  if (!reportedUser) {
    throw new Error('请输入被举报人')
  }

  if (!reason) {
    throw new Error('请填写举报原因')
  }

  const result = await profileApi.submitSystemReport(payload)

  if (result.code !== 0) {
    throw new Error(result.message || '举报提交失败')
  }

  return result.data
}

async function getSystemReportDetail(params = {}) {
  const reportId = String(params.reportId || params.id || '').trim()

  if (!reportId) {
    throw new Error('缺少举报信息')
  }

  const result = await profileApi.getSystemReportDetail(encodeURIComponent(reportId))

  if (result.code !== 0) {
    throw new Error(result.message || '获取举报详情失败')
  }

  return result.data
}

async function getSystemReportRecordDetail(params = {}) {
  const recordId = String(params.recordId || params.id || '').trim()

  if (!recordId) {
    throw new Error('缺少处理记录信息')
  }

  const result = await profileApi.getSystemReportRecordDetail(encodeURIComponent(recordId))

  if (result.code !== 0) {
    throw new Error(result.message || '获取处理详情失败')
  }

  return result.data
}

async function getSystemReportAppeals(params = {}) {
  const result = await profileApi.getSystemReportAppeals(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取申诉列表失败')
  }

  return result.data
}

async function getSystemReportAppealDetail(params = {}) {
  const appealId = String(params.appealId || params.id || '').trim()

  if (!appealId) {
    throw new Error('缺少申诉信息')
  }

  const result = await profileApi.getSystemReportAppealDetail(encodeURIComponent(appealId))

  if (result.code !== 0) {
    throw new Error(result.message || '获取申诉详情失败')
  }

  return result.data
}

async function withdrawSystemReportAppeal(params = {}) {
  const appealId = String(params.appealId || params.id || '').trim()

  if (!appealId) {
    throw new Error('缺少申诉信息')
  }

  const result = await profileApi.withdrawSystemReportAppeal(encodeURIComponent(appealId))

  if (result.code !== 0) {
    throw new Error(result.message || '撤回申诉失败')
  }

  return result.data
}

async function getSystemAgreements(params = {}) {
  const result = await profileApi.getSystemAgreements(params)

  if (result.code !== 0) {
    throw new Error(result.message || '获取协议列表失败')
  }

  return result.data
}

async function getSystemAgreementDetail(params = {}) {
  const agreementId = String(params.agreementId || params.agreement || params.id || '').trim()

  if (!agreementId) {
    throw new Error('缺少协议信息')
  }

  const result = await profileApi.getSystemAgreementDetail(encodeURIComponent(agreementId))

  if (result.code !== 0) {
    throw new Error(result.message || '获取协议详情失败')
  }

  return result.data
}

async function signSystemAgreement(params = {}) {
  const agreementId = String(params.agreementId || params.agreement || params.id || '').trim()

  if (!agreementId) {
    throw new Error('缺少协议信息')
  }

  const result = await profileApi.signSystemAgreement(encodeURIComponent(agreementId), params)

  if (result.code !== 0) {
    throw new Error(result.message || '协议签署失败')
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
  const key = String(payload.key || payload.settingKey || payload.id || '').trim()

  if (!key) {
    throw new Error('缺少设置信息')
  }

  const result = await profileApi.saveProfileSettings(payload)

  if (result.code !== 0) {
    throw new Error(result.message || '设置保存失败')
  }

  return result.data
}

async function clearProfileSettingsCache() {
  const result = await profileApi.clearProfileSettingsCache()

  if (result.code !== 0) {
    throw new Error(result.message || '清除缓存失败')
  }

  return result.data
}

async function logoutProfile() {
  const result = await profileApi.logoutProfile()

  if (result.code !== 0) {
    throw new Error(result.message || '退出登录失败')
  }

  return result.data
}

module.exports = {
  getProfileHome,
  getMemberCenterConfig,
  getMemberRadarOverview,
  getMemberRadarProfile,
  saveMemberRadarProfile,
  startMemberRadarMatch,
  getMemberRadarMatch,
  getMemberRadarMatchResults,
  followMemberRadarResult,
  replyServiceReview,
  getProfileAssets,
  getProfilePoints,
  getProfileGames,
  getServiceReviews,
  getServiceReviewDetail,
  getInviteOverview,
  getInviteRecords,
  getInviteIncome,
  getInviteNetwork,
  getInviteRanking,
  getInviteMemberDetail,
  getProfileCredit,
  getProfileAchievements,
  getPointsMall,
  exchangePointsMallGood,
  getPointsOrders,
  getPointsOrderLogistics,
  saveSystemProfileInfo,
  getSystemProfileInfo,
  getSystemSkillConfig,
  saveSystemSkillConfig,
  getSystemSkillCaseDetail,
  getSystemBlockSettings,
  saveSystemBlockStatus,
  saveSystemProtectionMode,
  saveSystemBlockScenes,
  getSystemBlockWhitelist,
  addSystemBlockWhitelist,
  removeSystemBlockWhitelist,
  getSystemBlockedUsers,
  addSystemBlockedUsers,
  removeSystemBlockedUser,
  getSystemBlockRenewalOptions,
  renewSystemBlockSettings,
  getSystemBlockKeywords,
  addSystemBlockKeyword,
  removeSystemBlockKeyword,
  getSystemFeedbackRecords,
  getSystemFeedbackDetail,
  getSystemFeedbackOptions,
  getSystemFeedbackGames,
  submitSystemFeedback,
  getCreditAppealOptions,
  submitCreditAppeal,
  getSystemReportRecords,
  getSystemReportOptions,
  submitSystemReport,
  getSystemReportDetail,
  getSystemReportRecordDetail,
  getSystemReportAppeals,
  getSystemReportAppealDetail,
  withdrawSystemReportAppeal,
  getSystemAgreements,
  getSystemAgreementDetail,
  signSystemAgreement,
  getProfileSettings,
  saveProfileSettings,
  clearProfileSettingsCache,
  logoutProfile
}
