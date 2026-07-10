const request = require('../request')

function getIncomeSummary() {
  return request.get('/api/app/income/summary')
}

function getIncomeLogs(params) {
  return request.get('/api/app/income/logs', params)
}

module.exports = {
  getIncomeSummary,
  getIncomeLogs
}
