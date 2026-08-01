const newbieApi = require('../api/modules/newbie')

async function getNewbieTasks() {
  const result = await newbieApi.getNewbieTasks()

  if (result.code !== 0) {
    throw new Error(result.message || '获取新手任务失败')
  }

  return result.data
}

async function completeTask(code) {
  const result = await newbieApi.completeTask(code)
  if (result.code !== 0) throw new Error(result.message || '任务完成记录失败')
  return result.data
}

async function recordProfileGuideReminder() {
  const result = await newbieApi.recordProfileGuideReminder()
  if (result.code !== 0) throw new Error(result.message || '记录资料提醒失败')
  return result.data || {}
}

module.exports = {
  getNewbieTasks,
  completeTask,
  recordProfileGuideReminder
}
