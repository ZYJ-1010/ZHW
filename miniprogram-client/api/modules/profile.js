const request = require('../request')

function getProfileHome() {
  return request.get('/api/app/profile/home')
}

function getProfileAssets() {
  return request.get('/api/app/profile/assets')
}

function getCreditCenter() {
  return request.get('/api/app/profile/credit-center')
}

function getPointsSummary() {
  return request.get('/api/app/points/summary')
}

function getPointsLogs(data) {
  return request.get('/api/app/points/logs', data)
}

function getGrowth() {
  return request.get('/api/app/growth/my')
}

function getMemberStatus() {
  return request.get('/api/app/membership/my')
}

function getMemberPlans() {
  return request.get('/api/app/membership/plans')
}

function getMemberRadarConfig() {
  return request.get('/api/app/membership/radar-config')
}

function submitMemberRadarAction(data) {
  return request.post('/api/app/membership/radar/actions', data)
}

function getServiceReviews(data) {
  return request.get('/api/app/profile/service-center/reviews', data)
}

function getServiceReviewDetail(reviewId) {
  return request.get(`/api/app/profile/service-center/reviews/${reviewId}`)
}

function replyServiceReview(reviewId, data) {
  return request.post(`/api/app/profile/service-center/reviews/${reviewId}/reply`, data)
}

function likeServiceReview(reviewId) {
  return request.post(`/api/app/profile/service-center/reviews/${reviewId}/like`, {})
}

function getServiceReviewActions(reviewId) {
  return request.post(`/api/app/profile/service-center/reviews/${reviewId}/actions`, {})
}

function getPointsMall() {
  return request.get('/api/app/redemption/items')
}

function exchangePointsMallGood(data) {
  return request.post('/api/app/redemption/orders', data)
}

function getPointsOrders(data) {
  return request.get('/api/app/redemption/orders/my', data)
}

function getPointsOrderLogistics(orderId) {
  return request.get(`/api/app/profile/points/orders/${orderId}/logistics`)
}

function getPointsOrderDetail(orderId) {
  return request.get(`/api/app/redemption/orders/${orderId}`)
}

function cancelPointsOrder(orderId, data) {
  return request.post(`/api/app/redemption/orders/${orderId}/cancel`, data || {})
}

function getInviteOverview() {
  return request.get('/api/app/profile/service-center/invite/overview')
}

function getInviteNetwork() {
  return request.get('/api/app/profile/service-center/invite/network')
}

function getInviteRecords(data) {
  return request.get('/api/app/profile/service-center/invite/records', data)
}

function getInviteRanking(data) {
  return request.get('/api/app/profile/service-center/invite/ranking', data)
}

function getInviteIncome(data) {
  return request.get('/api/app/profile/service-center/invite/income', data)
}

function getInviteMemberDetail(memberId) {
  return request.get('/api/app/profile/service-center/invite/member-detail', { memberId })
}

function getSystemProfileInfo() {
  return request.get('/api/app/profile/system-management/profile-info')
}

function saveSystemProfileInfo(data) {
  return request.put('/api/app/profile/system-management/profile-info', data)
}

function restartRealname(data) {
  return request.post('/api/app/identity/realname/restart', data)
}

function getSystemSkillConfig() {
  return request.get('/api/app/profile/system-management/skill-config')
}

function getSystemServiceCaseDetail(caseId) {
  return request.get(`/api/app/profile/system-management/service-cases/${caseId}`)
}

function saveSystemSkillConfig(data) {
  return request.put('/api/app/profile/system-management/skill-config', data)
}

function getSystemFeedbackHome() {
  return request.get('/api/app/profile/system-management/feedback')
}

function submitSystemFeedback(data) {
  return request.post('/api/app/profile/system-management/feedback', data)
}

function getSystemFeedbackRecords(data) {
  return request.get('/api/app/profile/system-management/feedback-records', data)
}

function getSystemFeedbackDetail(recordId) {
  return request.get(`/api/app/profile/system-management/feedback-records/${recordId}`)
}

function appendSystemFeedbackMessage(recordId, data) {
  return request.post(`/api/app/profile/system-management/feedback-records/${recordId}/messages`, data)
}

function getSystemBlockSettings() {
  return request.get('/api/app/profile/system-management/block-settings')
}

function saveSystemBlockSettings(data) {
  return request.put('/api/app/profile/system-management/block-settings', data)
}

function getProfileSettings() {
  return request.get('/api/app/profile/settings')
}

function saveProfileSettings(data) {
  return request.put('/api/app/profile/settings', data)
}

function getProfileAgreements() {
  return request.get('/api/app/profile/agreements')
}

function getProfileAgreementDetail(agreementKey) {
  return request.get(`/api/app/profile/agreements/${agreementKey}`)
}

function signProfileAgreement(agreementKey, data) {
  return request.post(`/api/app/profile/agreements/${agreementKey}/sign`, data || {})
}

module.exports = {
  getProfileHome,
  getProfileAssets,
  getCreditCenter,
  getPointsSummary,
  getPointsLogs,
  getGrowth,
  getMemberStatus,
  getMemberPlans,
  getMemberRadarConfig,
  submitMemberRadarAction,
  getServiceReviews,
  getServiceReviewDetail,
  replyServiceReview,
  likeServiceReview,
  getServiceReviewActions,
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
  restartRealname,
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
