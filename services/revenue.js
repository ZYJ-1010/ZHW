const revenueApi = require('../api/modules/revenue')

async function getIncomeSummary() {
  const result = await revenueApi.getIncomeSummary()

  if (result.code !== 0) {
    throw new Error(result.message || '获取收益信息失败')
  }

  return result.data
}

module.exports = {
  getIncomeSummary
}
