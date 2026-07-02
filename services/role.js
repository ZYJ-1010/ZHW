const roleApi = require('../api/modules/role')

function trimText(value) {
  return String(value || '').trim()
}

function normalizeRoleCode(payload = {}) {
  const roleText = trimText(payload.roleCode || payload.roleType || payload.role)

  if (roleText === '行家' || roleText === 'expert' || roleText === 'master') {
    return 'expert'
  }

  if (roleText === '领路人' || roleText === 'guide' || roleText === 'leader') {
    return 'guide'
  }

  if (roleText === '玩家' || roleText === 'player') {
    return 'player'
  }

  return roleText || 'player'
}

function normalizeProofFileIds(files = []) {
  if (!Array.isArray(files)) {
    return []
  }

  return files
    .map((file) => Number(file && (file.fileId || file.fileID || file.id)))
    .filter((fileId) => Number.isInteger(fileId) && fileId > 0)
}

function buildExpertApplication(payload = {}) {
  const services = Array.isArray(payload.services) ? payload.services : []
  const serviceLines = services
    .map((service, index) => {
      const name = trimText(service.name || service.serviceName)
      const price = trimText(service.price || service.hourlyPrice)
      const cost = trimText(service.cost)

      if (!name && !price && !cost) {
        return ''
      }

      return `服务${index + 1}：${name || '未命名'}，定价${price || '未填'}，成本${cost || '未填'}`
    })
    .filter(Boolean)

  const skillDomain = trimText(payload.skillDomain || (payload.skillDomains || [])[0])
  const skillTags = trimText(payload.skillTags)
  const intro = trimText(payload.intro || payload.personalIntro)

  return {
    roleCode: 'expert',
    reason: skillDomain || '申请成为行家',
    abilityDescription: [
      skillDomain ? `技能领域：${skillDomain}` : '',
      skillTags ? `技能标签：${skillTags}` : '',
      payload.experienceYears ? `从业年限：${payload.experienceYears}` : '',
      intro ? `个人简介：${intro}` : '',
      serviceLines.length ? `服务说明：${serviceLines.join('；')}` : ''
    ].filter(Boolean).join('\n'),
    proofFileIds: normalizeProofFileIds(payload.uploadFiles || payload.qualifications)
  }
}

function buildGuideApplication(payload = {}) {
  const form = payload.form || payload
  const services = Array.isArray(form.services) ? form.services : []
  const serviceLines = services
    .map((service, index) => {
      const name = trimText(service.name || service.serviceName)
      const price = trimText(service.price || service.hourlyPrice)
      const cost = trimText(service.cost)

      if (!name && !price && !cost) {
        return ''
      }

      return `业务${index + 1}：${name || '未命名'}，定价${price || '未填'}，成本${cost || '未填'}`
    })
    .filter(Boolean)

  const city = trimText(form.city)
  const contact = trimText(form.contact)
  const audience = Array.isArray(form.audience) ? form.audience.filter(Boolean).join('、') : trimText(form.audience)
  const guidePlan = trimText(form.guidePlan || form.reason || form.abilityDescription)

  return {
    roleCode: 'guide',
    reason: guidePlan || '申请成为领路人',
    abilityDescription: [
      city ? `所在城市：${city}` : '',
      audience ? `可推荐人群：${audience}` : '',
      contact ? `常用联系方式：${contact}` : '',
      guidePlan ? `领路计划书：${guidePlan}` : '',
      serviceLines.length ? `业务说明：${serviceLines.join('；')}` : ''
    ].filter(Boolean).join('\n'),
    proofFileIds: normalizeProofFileIds(form.uploadFiles)
  }
}

function normalizeRoleApplicationPayload(payload) {
  if (typeof payload === 'string') {
    return { roleCode: normalizeRoleCode({ roleType: payload }), reason: `申请成为${payload}` }
  }

  const data = payload || {}
  const roleCode = normalizeRoleCode(data)

  if (data.roleCode && data.reason) {
    return data
  }

  if (roleCode === 'expert') {
    return buildExpertApplication(data)
  }

  if (roleCode === 'guide') {
    return buildGuideApplication(data)
  }

  return Object.assign({}, data, {
    roleCode,
    reason: trimText(data.reason) || `申请成为${roleCode}`
  })
}

async function getMyRoles() {
  const result = await roleApi.getMyRoles()

  if (result.code !== 0) {
    throw new Error(result.message || '获取角色信息失败')
  }

  return result.data
}

async function getMyRoleApplications() {
  const result = await roleApi.getMyRoleApplications()

  if (result.code !== 0) {
    throw new Error(result.message || '获取角色申请信息失败')
  }

  if (Array.isArray(result.data)) {
    return { items: result.data, pageConfig: {} }
  }

  return Object.assign({ items: [], pageConfig: {} }, result.data || {})
}

async function submitRoleApplication(payload) {
  const data = normalizeRoleApplicationPayload(payload)
  const result = await roleApi.submitRoleApplication(data)

  if (result.code !== 0) {
    throw new Error(result.message || '提交角色申请失败')
  }

  return result.data
}

async function getExpertApplyConfig() {
  const result = await roleApi.getExpertApplyConfig()

  if (result.code !== 0) {
    throw new Error(result.message || '获取行家申请配置失败')
  }

  return result.data
}

async function getGuideApplyConfig() {
  const result = await roleApi.getGuideApplyConfig()

  if (result.code !== 0) {
    throw new Error(result.message || '获取领路人申请配置失败')
  }

  return result.data
}

async function getRoleStatusPageConfig() {
  const result = await roleApi.getRoleStatusPageConfig()

  if (result.code !== 0) {
    throw new Error(result.message || '获取角色审核页配置失败')
  }

  return result.data
}

async function getRoleBenefitConfig() {
  const result = await roleApi.getRoleBenefitConfig()

  if (result.code !== 0) {
    throw new Error(result.message || '获取角色权益配置失败')
  }

  return result.data
}

module.exports = {
  getMyRoles,
  getMyRoleApplications,
  submitRoleApplication,
  getExpertApplyConfig,
  getGuideApplyConfig,
  getRoleStatusPageConfig,
  getRoleBenefitConfig
}
