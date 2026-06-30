const request = require('../request')

function getProfileHome() {
  return request.get('/api/app/profile/home')
}

function getGrowth() {
  return request.get('/api/app/growth/me')
}

function getMemberStatus() {
  return request.get('/api/app/memberships/me')
}

function replyServiceReview(reviewId, data) {
  return request.post(`/api/app/profile/service-center/reviews/${reviewId}/reply`, data)
}

function getProfileAssets() {
  return request.get('/api/app/profile/assets')
}

function getProfilePoints(data) {
  return request.get('/api/app/profile/points', data)
}

function getProfileGames(data) {
  return request.get('/api/app/profile/games', data)
}

function getServiceReviews(data) {
  return request.get('/api/app/profile/service-center/reviews', data)
}

function getServiceReviewDetail(reviewId) {
  return request.get(`/api/app/profile/service-center/reviews/${reviewId}`)
}

function getInviteOverview() {
  return request.get('/api/app/profile/invite/overview')
}

function getInviteRecords(data) {
  return request.get('/api/app/profile/invite-records', data)
}

function getInviteIncome(data) {
  return request.get('/api/app/profile/invite/income', data)
}

function getInviteNetwork(data) {
  return request.get('/api/app/profile/invite/network', data)
}

function getInviteRanking(data) {
  return request.get('/api/app/profile/invite/ranking', data)
}

function getInviteMemberDetail(memberId) {
  return request.get(`/api/app/profile/invite/members/${memberId}`)
}

function getProfileCredit() {
  return request.get('/api/app/profile/credit')
}

function getProfileAchievements(data) {
  return request.get('/api/app/profile/footprint/achievements', data)
}

function getPointsMall() {
  return request.get('/api/app/profile/points/mall')
}

function exchangePointsMallGood(data) {
  return request.post('/api/app/profile/points/mall/exchange', data)
}

function getPointsOrders(data) {
  return request.get('/api/app/profile/points/orders', data)
}

function getPointsOrderLogistics(orderId) {
  return request.get(`/api/app/profile/points/orders/${orderId}/logistics`)
}

function saveSystemProfileInfo(data) {
  return request.put('/api/app/profile/system-management/profile-info', data)
}

function getSystemProfileInfo() {
  return request.get('/api/app/profile/system-management/profile-info')
}

function getSystemSkillConfig() {
  return request.get('/api/app/profile/system-management/skill-config')
}

function saveSystemSkillConfig(data) {
  return request.put('/api/app/profile/system-management/skill-config', data)
}

function getSystemSkillCaseDetail(caseId) {
  return request.get(`/api/app/profile/system-management/skill-config/cases/${caseId}`)
}

function getSystemBlockSettings() {
  return request.get('/api/app/profile/system-management/block-settings')
}

function saveSystemBlockStatus(data) {
  return request.put('/api/app/profile/system-management/block-settings/status', data)
}

function saveSystemProtectionMode(data) {
  return request.put('/api/app/profile/system-management/block-settings/protection-mode', data)
}

function saveSystemBlockScenes(data) {
  return request.put('/api/app/profile/system-management/block-settings/scenes', data)
}

function getSystemBlockWhitelist(data) {
  return request.get('/api/app/profile/system-management/block-settings/whitelist', data)
}

function addSystemBlockWhitelist(data) {
  return request.post('/api/app/profile/system-management/block-settings/whitelist', data)
}

function removeSystemBlockWhitelist(expertId) {
  return request.delete(`/api/app/profile/system-management/block-settings/whitelist/${expertId}`)
}

function getSystemBlockedUsers(data) {
  return request.get('/api/app/profile/system-management/block-settings/users', data)
}

function addSystemBlockedUsers(data) {
  return request.post('/api/app/profile/system-management/block-settings/users', data)
}

function removeSystemBlockedUser(userId) {
  return request.delete(`/api/app/profile/system-management/block-settings/users/${userId}`)
}

function getSystemBlockRenewalOptions() {
  return request.get('/api/app/profile/system-management/block-settings/renewal-options')
}

function renewSystemBlockSettings(data) {
  return request.post('/api/app/profile/system-management/block-settings/renewal', data)
}

function getSystemBlockKeywords(data) {
  return request.get('/api/app/profile/system-management/block-settings/keywords', data)
}

function addSystemBlockKeyword(data) {
  return request.post('/api/app/profile/system-management/block-settings/keywords', data)
}

function removeSystemBlockKeyword(keywordId) {
  return request.delete(`/api/app/profile/system-management/block-settings/keywords/${keywordId}`)
}

function getSystemFeedbackRecords(data) {
  return request.get('/api/app/profile/system-management/feedback/records', data)
}

function getSystemFeedbackDetail(feedbackId) {
  return request.get(`/api/app/profile/system-management/feedback/records/${feedbackId}`)
}

function getSystemFeedbackOptions() {
  return request.get('/api/app/profile/system-management/feedback/options')
}

function getSystemFeedbackGames(data) {
  return request.get('/api/app/profile/system-management/feedback/games', data)
}

function submitSystemFeedback(data) {
  return request.post('/api/app/profile/system-management/feedback', data)
}

function getCreditAppealOptions(data) {
  return request.get('/api/app/profile/credit/appeal/options', data)
}

function submitCreditAppeal(data) {
  return request.post('/api/app/profile/credit/appeals', data)
}

function getSystemReportRecords(data) {
  return request.get('/api/app/profile/system-management/reports/records', data)
}

function getSystemReportOptions() {
  return request.get('/api/app/profile/system-management/reports/options')
}

function submitSystemReport(data) {
  return request.post('/api/app/profile/system-management/reports', data)
}

function getSystemReportDetail(reportId) {
  return request.get(`/api/app/profile/system-management/reports/${reportId}`)
}

function getSystemReportRecordDetail(recordId) {
  return request.get(`/api/app/profile/system-management/reports/records/${recordId}`)
}

function getSystemReportAppeals(data) {
  return request.get('/api/app/profile/system-management/reports/appeals', data)
}

function getSystemReportAppealDetail(appealId) {
  return request.get(`/api/app/profile/system-management/reports/appeals/${appealId}`)
}

function withdrawSystemReportAppeal(appealId) {
  return request.post(`/api/app/profile/system-management/reports/appeals/${appealId}/withdraw`)
}

function getSystemAgreements(data) {
  return request.get('/api/app/profile/system-management/agreements', data)
}

function getSystemAgreementDetail(agreementId) {
  return request.get(`/api/app/profile/system-management/agreements/${agreementId}`)
}

function signSystemAgreement(agreementId, data) {
  return request.post(`/api/app/profile/system-management/agreements/${agreementId}/sign`, data)
}

function getProfileSettings() {
  return request.get('/api/app/profile/settings')
}

function saveProfileSettings(data) {
  return request.put('/api/app/profile/settings', data)
}

function clearProfileSettingsCache() {
  return request.post('/api/app/profile/settings/cache/clear')
}

function logoutProfile() {
  return request.post('/api/app/auth/logout')
}

module.exports = {
  getProfileHome,
  getGrowth,
  getMemberStatus,
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
