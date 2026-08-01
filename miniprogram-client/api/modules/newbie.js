const request = require('../request')

function getNewbieTasks() {
  return request.get('/api/app/newbie-tasks')
}

function completeTask(code) {
  return request.post(`/api/app/newbie-tasks/${encodeURIComponent(code)}/complete`, {})
}

function recordProfileGuideReminder() {
  return request.post('/api/app/newbie-tasks/guide-profile-reminder', {})
}

module.exports = {
  getNewbieTasks,
  completeTask,
  recordProfileGuideReminder
}
