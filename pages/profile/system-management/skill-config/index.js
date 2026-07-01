const profileService = require('../../../../services/profile')
const toast = require('../../../../utils/toast')

const ASSET_BASE = '/pages/profile/system-management/skill-config/assets'
const DEFAULT_QUOTA_RESET_DATE = ''
const MIN_CASE_DESC_LENGTH = 100
const MAX_CASE_DESC_LENGTH = 300

const ADDABLE_SKILLS = []

function buildStars(activeCount) {
  return Array.from({ length: 5 }, (_, index) => ({
    key: `star-${index}`,
    active: index < activeCount
  }))
}

function buildIcons() {
  return {
    skill: `${ASSET_BASE}/skill-config.png`,
    credit: `${ASSET_BASE}/credit.png`,
    caseFile: `${ASSET_BASE}/icon-case-file.svg`,
    calendar: `${ASSET_BASE}/icon-calendar.svg`,
    users: `${ASSET_BASE}/icon-users.svg`,
    feedback: `${ASSET_BASE}/feedback.png`,
    layers: `${ASSET_BASE}/icon-status-layers.svg`,
    bolt: `${ASSET_BASE}/icon-status-bolt.svg`,
    plus: `${ASSET_BASE}/icon-plus.svg`,
    star: `${ASSET_BASE}/icon-star.svg`,
    warning: `${ASSET_BASE}/icon-warning.png`,
    hourglass: `${ASSET_BASE}/icon-hourglass.png`,
    lock: `${ASSET_BASE}/icon-lock.svg`,
    check: `${ASSET_BASE}/icon-check-plain.svg`,
    chevronRight: `${ASSET_BASE}/icon-chevron-right.svg`
  }
}

function getIconSrc(item, icons) {
  if (item.iconSrc) {
    return item.iconSrc
  }

  return icons[item.iconKey] || icons.skill
}

function shouldUseImageIcon(item) {
  const key = item.iconKey || item.id

  return key === 'caseFile'
}

function getIconText(item) {
  if (shouldUseImageIcon(item)) {
    return ''
  }

  if (item.iconText) {
    return item.iconText
  }

  const key = item.iconKey || item.id
  const title = item.title || ''

  if (key === 'atmosphere' || title.indexOf('氛围') !== -1) {
    return '🎭'
  }

  if (key === 'detail' || title.indexOf('细节') !== -1) {
    return '🔍'
  }

  if (key === 'plus' || key === 'add' || title.indexOf('添加') !== -1) {
    return '+'
  }

  return ''
}

function normalizeSkillItem(item, icons) {
  const rating = Number(item.rating)

  return {
    ...item,
    iconSrc: getIconSrc(item, icons),
    iconAsImage: shouldUseImageIcon(item),
    iconText: getIconText(item),
    ratingStars: Array.isArray(item.ratingStars)
      ? item.ratingStars
      : buildStars(Number.isFinite(rating) ? Math.max(0, Math.min(5, Math.round(rating))) : 5)
  }
}

function clone(value) {
  return JSON.parse(JSON.stringify(value || null))
}

function getConfiguredCount(skillSlots = []) {
  return skillSlots.filter((item) => item && !item.empty).length
}

function buildEmptySlot(sourceId) {
  return {
    id: `empty-${sourceId || Date.now()}`,
    title: '添加技能',
    iconKey: 'plus',
    iconText: '+',
    tone: 'gray',
    active: false,
    empty: true
  }
}

function buildSkillFromCandidate(candidate, sourceType = 'manual') {
  const isAi = sourceType === 'ai'

  return normalizeSkillItem({
    id: candidate.id,
    title: candidate.title,
    iconText: candidate.iconText,
    tone: candidate.tone,
    badge: isAi ? 'AI解锁' : '手动解锁',
    badgeTone: isAi ? 'purple' : 'info',
    visibilityText: '显性技能 · 玩家可见',
    sourceText: isAi ? '系统综合评估自动解锁' : '行家手动配置',
    lockedAt: '刚刚',
    caseTitle: '定制服务案例',
    caseBadge: '待完善',
    caseDesc: candidate.caseDesc || candidate.desc,
    caseDate: '待补充',
    casePlayers: '待绑定',
    ratingText: '待评分',
    ratingStars: buildStars(0)
  }, buildIcons())
}

function consumeMonthlyQuota(roleSummary = {}, configuredCount) {
  const usedCount = Number(roleSummary.usedCount) || 0
  const remainingCount = Number(roleSummary.remainingCount) || 0
  const monthlyLimit = Number(roleSummary.monthlyLimit) || 0

  return {
    ...roleSummary,
    usedCount: monthlyLimit ? Math.min(monthlyLimit, usedCount + 1) : usedCount + 1,
    remainingCount: Math.max(0, remainingCount - 1),
    configuredCount
  }
}

function findSkillInGroups(skillGroups = {}, skillId, preferredGroup) {
  const keys = preferredGroup
    ? [preferredGroup].concat(Object.keys(skillGroups).filter((key) => key !== preferredGroup))
    : Object.keys(skillGroups)

  for (let keyIndex = 0; keyIndex < keys.length; keyIndex += 1) {
    const groupKey = keys[keyIndex]
    const list = Array.isArray(skillGroups[groupKey]) ? skillGroups[groupKey] : []
    const itemIndex = list.findIndex((item) => item.id === skillId)

    if (itemIndex !== -1) {
      return {
        groupKey,
        itemIndex,
        skill: list[itemIndex]
      }
    }
  }

  return null
}

function normalizeSkillConfig(config = {}) {
  const fallback = getDefaultSkillConfig()
  const icons = buildIcons()
  const rawSkillGroups = config.skillGroups && typeof config.skillGroups === 'object'
    ? config.skillGroups
    : fallback.skillGroups
  const skillGroups = Object.keys(fallback.sectionMap).reduce((result, key) => {
    const list = Array.isArray(rawSkillGroups[key]) ? rawSkillGroups[key] : fallback.skillGroups[key]

    result[key] = list.map((item) => normalizeSkillItem(item, icons))
    return result
  }, {})

  return {
    ...fallback,
    ...config,
    icons,
    roleSummary: {
      ...fallback.roleSummary,
      ...(config.roleSummary || {})
    },
    skillSlots: (Array.isArray(config.skillSlots) ? config.skillSlots : fallback.skillSlots)
      .map((item) => ({
        ...item,
        iconSrc: getIconSrc(item, icons),
        iconAsImage: shouldUseImageIcon(item),
        iconText: getIconText(item)
      })),
    tabs: Array.isArray(config.tabs) && config.tabs.length ? config.tabs : fallback.tabs,
    sectionMap: {
      ...fallback.sectionMap,
      ...(config.sectionMap || {})
    },
    skillGroups,
    unlockSuggestion: {
      ...fallback.unlockSuggestion,
      ...(config.unlockSuggestion || {})
    }
  }
}

function getDefaultSkillConfig() {
  const icons = buildIcons()

  return {
    icons,
    roleSummary: {
      roleName: '',
      maxSkillCount: '',
      monthlyLimit: '',
      usedCount: '',
      remainingCount: '',
      configuredCount: ''
    },
    skillSlots: [],
    tabs: [
      { key: 'visible', label: '显性技能' },
      { key: 'hidden', label: '隐形技能' },
      { key: 'cases', label: '服务案例' }
    ],
    sectionMap: {
      visible: {
        title: '已配置显性技能',
        desc: '玩家可见，用于建立信任'
      },
      hidden: {
        title: '系统识别隐形技能',
        desc: '由履约、评价与复盘内容沉淀，暂不直接展示给玩家'
      },
      cases: {
        title: '服务案例沉淀',
        desc: '用于支撑技能标签和后续智能推荐'
      }
    },
    skillGroups: {
      visible: [],
      hidden: [],
      cases: []
    },
    unlockSuggestion: {
      title: '',
      desc: '',
      actionText: ''
    }
  }
}

function hasLockedSkillSlot(skillSlots = []) {
  return (skillSlots || []).some((item) => item && item.empty && item.locked)
}

function getUnlockedEmptySlotCount(skillSlots = []) {
  return (skillSlots || []).filter((item) => item && item.empty && !item.locked).length
}

function unlockFirstLockedSkillSlot(skillSlots = []) {
  let unlocked = false

  return (skillSlots || []).map((item) => {
    if (unlocked || !item || !item.empty || !item.locked) {
      return item
    }

    unlocked = true
    return {
      ...item,
      iconKey: 'plus',
      iconText: '+',
      locked: false
    }
  })
}

function buildPageData(config = getDefaultSkillConfig(), activeTab = 'visible') {
  const normalizedConfig = normalizeSkillConfig(config)
  const nextActiveTab = activeTab || 'visible'
  const displayedSkills = normalizedConfig.skillGroups[nextActiveTab] || []
  const hiddenSkills = normalizedConfig.skillGroups.hidden || []
  const isCasesTab = nextActiveTab === 'cases'
  const isHiddenTab = nextActiveTab === 'hidden'
  const hasLockedSlot = hasLockedSkillSlot(normalizedConfig.skillSlots)

  return {
    ...normalizedConfig,
    activeTab: nextActiveTab,
    sectionInfo: normalizedConfig.sectionMap[nextActiveTab] || normalizedConfig.sectionMap.visible,
    displayedSkills,
    hasDisplayedSkills: displayedSkills.length > 0,
    hasHiddenSkills: hiddenSkills.length > 0,
    isCasesTab,
    isHiddenTab,
    hasLockedSlot,
    showUnlockSuggestion: hasLockedSlot && nextActiveTab === 'visible'
  }
}

Page({
  data: {
    ...buildPageData(),
    isSaving: false,
    quotaResetDate: DEFAULT_QUOTA_RESET_DATE,
    caseDescMinLength: MIN_CASE_DESC_LENGTH,
    caseDescMaxLength: MAX_CASE_DESC_LENGTH,
    responseDialog: {
      visible: false
    },
    editPanel: {
      visible: false
    },
    addPanel: {
      visible: false
    }
  },

  onLoad() {
    this.loadSkillConfig()
  },

  async loadSkillConfig() {
    try {
      const remoteConfig = await profileService.getSystemSkillConfig()
      const config = remoteConfig && typeof remoteConfig === 'object'
        ? remoteConfig
        : getDefaultSkillConfig()

      this.setData(buildPageData(config, config.activeTab || this.data.activeTab))
    } catch (error) {
      this.setData(buildPageData(getDefaultSkillConfig(), this.data.activeTab))
    }
  },

  handleTabTap(event) {
    const { key } = event.currentTarget.dataset

    if (!key || key === this.data.activeTab) {
      return
    }

    this.setData(buildPageData({
      icons: this.data.icons,
      roleSummary: this.data.roleSummary,
      skillSlots: this.data.skillSlots,
      tabs: this.data.tabs,
      sectionMap: this.data.sectionMap,
      skillGroups: this.data.skillGroups,
      unlockSuggestion: this.data.unlockSuggestion
    }, key))
  },

  handleSlotTap(event) {
    const slotId = event.currentTarget.dataset.id
    const slot = (this.data.skillSlots || []).find((item) => item.id === slotId)

    if (!slot) {
      return
    }

    if (slot.empty) {
      if (slot.locked) {
        toast.info('请先解锁第三个技能')
        return
      }

      this.openAddPanel()
      return
    }

    toast.info(`${slot.title} 已在下方展示`)
  },

  handleCaseSummaryTap(event) {
    const { id } = event.currentTarget.dataset

    if (!id) {
      toast.info('未找到案例详情')
      return
    }

    const caseItem = (this.data.displayedSkills || []).find((item) => item.id === id) || {}
    const params = [
      `caseId=${encodeURIComponent(id)}`,
      caseItem.iconText ? `iconText=${encodeURIComponent(caseItem.iconText)}` : '',
      caseItem.tone ? `tone=${encodeURIComponent(caseItem.tone)}` : ''
    ].filter(Boolean).join('&')

    wx.navigateTo({
      url: `/pages/profile/system-management/service-case-detail/index?${params}`
    })
  },

  handleSkillEditTap(event) {
    const { id } = event.currentTarget.dataset

    if (!this.hasRemainingQuota()) {
      this.showQuotaDialog()
      return
    }

    const target = findSkillInGroups(this.data.skillGroups, id, this.data.activeTab)

    if (!target) {
      toast.info('未找到要编辑的技能')
      return
    }

    const caseDesc = String(target.skill.caseDesc || '').slice(0, MAX_CASE_DESC_LENGTH)

    this.setData({
      editPanel: {
        visible: true,
        skillId: id,
        groupKey: target.groupKey,
        skillName: target.skill.title,
        caseDesc,
        caseDescCount: caseDesc.length,
        tagType: target.skill.badge === 'AI解锁' ? 'ai' : 'manual'
      }
    })
  },

  handleSkillRemoveTap(event) {
    const { id } = event.currentTarget.dataset

    if (!this.hasRemainingQuota()) {
      this.showQuotaDialog()
      return
    }

    const target = findSkillInGroups(this.data.skillGroups, id, this.data.activeTab)

    if (!target) {
      toast.info('未找到要移除的技能')
      return
    }

    this.showRemoveDialog(target.skill)
  },

  handleUnlockTap() {
    if (!hasLockedSkillSlot(this.data.skillSlots)) {
      toast.info('第三个技能已解锁')
      return
    }

    this.showUnlockSlotDialog()
  },

  hasRemainingQuota() {
    return Number(this.data.roleSummary && this.data.roleSummary.remainingCount) > 0
  },

  getCurrentConfig() {
    return {
      roleSummary: clone(this.data.roleSummary),
      skillSlots: clone(this.data.skillSlots),
      tabs: clone(this.data.tabs),
      sectionMap: clone(this.data.sectionMap),
      skillGroups: clone(this.data.skillGroups),
      unlockSuggestion: clone(this.data.unlockSuggestion)
    }
  },

  applyConfigState(config, activeTab, extraData = {}) {
    this.setData({
      ...buildPageData(config, activeTab || this.data.activeTab),
      ...extraData
    })
  },

  showQuotaDialog() {
    const roleSummary = this.data.roleSummary || {}
    const usedCount = Number(roleSummary.usedCount) || 0
    const monthlyLimit = Number(roleSummary.monthlyLimit) || 0
    const resetDate = roleSummary.resetDate || this.data.quotaResetDate
    const descLines = [
      `本月修改次数已用完（${usedCount}/${monthlyLimit}）`
    ]

    if (resetDate) {
      descLines.push(`下次重置时间为 ${resetDate}`)
    }

    this.setData({
      responseDialog: {
        visible: true,
        type: 'quota',
        iconSrc: this.data.icons.hourglass,
        iconTone: 'warning',
        title: '修改次数不足',
        descLines,
        confirmText: '我知道了',
        onlyConfirm: true
      }
    })
  },

  showRemoveDialog(skill) {
    const remainingCount = Number(this.data.roleSummary && this.data.roleSummary.remainingCount) || 0

    this.setData({
      responseDialog: {
        visible: true,
        type: 'remove',
        skillId: skill.id,
        iconSrc: this.data.icons.warning,
        iconTone: 'danger',
        title: '确认移除技能',
        descLines: [
          `确定要移除「${skill.title}」吗？移除后将释放该技能槽位，本月剩余修改次数将减少 1 次`
        ],
        hint: `（剩余 ${Math.max(0, remainingCount - 1)} 次）`,
        cancelText: '再想想',
        confirmText: '确认移除',
        confirmTone: 'danger'
      }
    })
  },

  showUnlockDialog(skill) {
    this.setData({
      responseDialog: {
        visible: true,
        type: 'unlock',
        skillId: skill.id,
        iconSrc: this.data.icons.lock,
        iconTone: 'lock',
        title: '解锁新技能',
        descLines: [
          `系统检测到您具备“${skill.title}”潜力，解锁后玩家可见该技能标签，有助于提升匹配率`
        ],
        cancelText: '暂不解锁',
        confirmText: '立即解锁'
      }
    })
  },

  showUnlockSlotDialog() {
    this.setData({
      responseDialog: {
        visible: true,
        type: 'unlock-slot',
        iconSrc: this.data.icons.lock,
        iconTone: 'lock',
        title: '解锁第三个技能',
        descLines: [
          '解锁后上方第三个技能槽位可添加新的显性技能'
        ],
        cancelText: '暂不解锁',
        confirmText: '立即解锁'
      }
    })
  },

  closeResponseDialog() {
    this.setData({
      responseDialog: {
        visible: false
      }
    })
  },

  handleDialogCancel() {
    this.closeResponseDialog()
  },

  handleDialogConfirm() {
    const dialog = this.data.responseDialog || {}

    if (dialog.type === 'quota') {
      this.closeResponseDialog()
      return
    }

    if (dialog.type === 'remove') {
      this.confirmRemoveSkill(dialog.skillId)
      return
    }

    if (dialog.type === 'unlock') {
      this.confirmUnlockSkill(dialog.skillId)
      return
    }

    if (dialog.type === 'unlock-slot') {
      this.confirmUnlockSlot()
    }
  },

  confirmRemoveSkill(skillId) {
    const config = this.getCurrentConfig()
    const target = findSkillInGroups(config.skillGroups, skillId, this.data.activeTab)

    if (!target) {
      this.closeResponseDialog()
      toast.info('未找到要移除的技能')
      return
    }

    config.skillGroups[target.groupKey] = config.skillGroups[target.groupKey]
      .filter((item) => item.id !== skillId)

    config.skillSlots = (config.skillSlots || []).map((item) => {
      if (item.id !== skillId) {
        return item
      }

      return buildEmptySlot(skillId)
    })

    config.roleSummary = consumeMonthlyQuota(
      config.roleSummary,
      getConfiguredCount(config.skillSlots)
    )

    this.applyConfigState(config, this.data.activeTab, {
      responseDialog: { visible: false }
    })
    toast.success('技能已移除')
  },

  confirmUnlockSkill(skillId) {
    if (!this.hasRemainingQuota()) {
      this.showQuotaDialog()
      return
    }

    const config = this.getCurrentConfig()
    const target = findSkillInGroups(config.skillGroups, skillId, 'hidden')
    const visibleList = Array.isArray(config.skillGroups.visible) ? config.skillGroups.visible : []

    if (!target) {
      this.closeResponseDialog()
      toast.info('暂无可解锁技能')
      return
    }

    if (visibleList.some((item) => item.id === skillId)) {
      this.closeResponseDialog()
      toast.info('该技能已解锁')
      return
    }

    if (this.getRemainingSlotCount(config.skillSlots, config.roleSummary) <= 0) {
      this.closeResponseDialog()
      toast.info('技能槽位已满')
      return
    }

    const unlockedSkill = buildSkillFromCandidate({
      id: target.skill.id,
      title: target.skill.title,
      desc: target.skill.caseDesc,
      caseDesc: target.skill.caseDesc,
      iconText: target.skill.iconText || '🧠',
      tone: target.skill.tone || 'orange'
    }, 'ai')

    config.skillGroups.visible = visibleList.concat(unlockedSkill)
    config.skillGroups.hidden = (config.skillGroups.hidden || [])
      .filter((item) => item.id !== skillId)
    config.skillSlots = this.fillFirstEmptySlot(config.skillSlots, unlockedSkill)
    config.unlockSuggestion = {
      title: '暂无待解锁技能',
      desc: '系统会根据后续服务表现继续识别潜力标签',
      actionText: '已解锁'
    }
    config.roleSummary = consumeMonthlyQuota(
      config.roleSummary,
      getConfiguredCount(config.skillSlots)
    )

    this.applyConfigState(config, 'visible', {
      responseDialog: { visible: false }
    })
    toast.success('技能已解锁')
  },

  confirmUnlockSlot() {
    const config = this.getCurrentConfig()

    if (!hasLockedSkillSlot(config.skillSlots)) {
      this.closeResponseDialog()
      toast.info('第三个技能已解锁')
      return
    }

    config.skillSlots = unlockFirstLockedSkillSlot(config.skillSlots)
    config.unlockSuggestion = {
      title: '第三个技能已解锁',
      desc: '点击上方添加技能槽位即可配置新的显性技能',
      actionText: '已解锁'
    }

    this.applyConfigState(config, 'visible', {
      responseDialog: { visible: false }
    })
    toast.success('第三个技能已解锁')
  },

  handleEditCaseInput(event) {
    const caseDesc = String(event.detail.value || '').slice(0, MAX_CASE_DESC_LENGTH)

    this.setData({
      'editPanel.caseDesc': caseDesc,
      'editPanel.caseDescCount': caseDesc.length
    })
  },

  handleCloseEditPanel() {
    this.setData({
      editPanel: {
        visible: false
      }
    })
  },

  handleSaveEditPanel() {
    const panel = this.data.editPanel || {}

    if (!this.hasRemainingQuota()) {
      this.handleCloseEditPanel()
      this.showQuotaDialog()
      return
    }

    const config = this.getCurrentConfig()
    const target = findSkillInGroups(config.skillGroups, panel.skillId, panel.groupKey)

    if (!target) {
      this.handleCloseEditPanel()
      toast.info('未找到要保存的技能')
      return
    }

    config.skillGroups[target.groupKey] = config.skillGroups[target.groupKey].map((item) => {
      if (item.id !== panel.skillId) {
        return item
      }

      return {
        ...item,
        caseDesc: panel.caseDesc || item.caseDesc
      }
    })
    config.roleSummary = consumeMonthlyQuota(
      config.roleSummary,
      getConfiguredCount(config.skillSlots)
    )

    this.applyConfigState(config, this.data.activeTab, {
      editPanel: {
        visible: false
      }
    })
    toast.success('修改已保存')
  },

  getRemainingSlotCount(skillSlots = this.data.skillSlots, roleSummary = this.data.roleSummary) {
    const maxSkillCount = Number(roleSummary && roleSummary.maxSkillCount) || 0
    const configuredCount = getConfiguredCount(skillSlots || [])
    const remainingByLimit = Math.max(0, maxSkillCount - configuredCount)
    const unlockedEmptyCount = getUnlockedEmptySlotCount(skillSlots || [])

    if (unlockedEmptyCount > 0) {
      return Math.min(remainingByLimit, unlockedEmptyCount)
    }

    if (hasLockedSkillSlot(skillSlots || [])) {
      return 0
    }

    return remainingByLimit
  },

  getAddableCandidates() {
    const visibleIds = new Set(((this.data.skillGroups && this.data.skillGroups.visible) || [])
      .map((item) => item.id))

    return ADDABLE_SKILLS.filter((item) => !visibleIds.has(item.id))
  },

  openAddPanel() {
    const remainingSlots = this.getRemainingSlotCount()

    if (remainingSlots <= 0 && hasLockedSkillSlot(this.data.skillSlots)) {
      toast.info('请先解锁第三个技能')
      return
    }

    if (remainingSlots <= 0) {
      toast.info('技能槽位已满')
      return
    }

    if (!this.hasRemainingQuota()) {
      this.showQuotaDialog()
      return
    }

    const candidates = this.getAddableCandidates()

    if (!candidates.length) {
      toast.info('暂无可添加技能')
      return
    }

    this.setData({
      addPanel: {
        visible: true,
        remainingSlots,
        selectedId: candidates[0].id,
        candidates
      }
    })
  },

  handleAddCandidateTap(event) {
    const { id } = event.currentTarget.dataset

    if (!id) {
      return
    }

    this.setData({
      'addPanel.selectedId': id
    })
  },

  handleCloseAddPanel() {
    this.setData({
      addPanel: {
        visible: false
      }
    })
  },

  handleConfirmAddSkill() {
    const panel = this.data.addPanel || {}
    const candidate = (panel.candidates || []).find((item) => item.id === panel.selectedId)

    if (!candidate) {
      toast.info('请选择要添加的技能')
      return
    }

    if (!this.hasRemainingQuota()) {
      this.handleCloseAddPanel()
      this.showQuotaDialog()
      return
    }

    const config = this.getCurrentConfig()

    if (getUnlockedEmptySlotCount(config.skillSlots) <= 0 && hasLockedSkillSlot(config.skillSlots)) {
      this.handleCloseAddPanel()
      toast.info('请先解锁第三个技能')
      return
    }

    if (this.getRemainingSlotCount(config.skillSlots, config.roleSummary) <= 0) {
      this.handleCloseAddPanel()
      toast.info('技能槽位已满')
      return
    }

    const nextSkill = buildSkillFromCandidate(candidate, 'manual')

    config.skillGroups.visible = (config.skillGroups.visible || []).concat(nextSkill)
    config.skillGroups.hidden = (config.skillGroups.hidden || [])
      .filter((item) => item.id !== candidate.id)
    config.skillSlots = this.fillFirstEmptySlot(config.skillSlots, nextSkill)
    if (!config.skillGroups.hidden.length) {
      config.unlockSuggestion = {
        title: '暂无待解锁技能',
        desc: '系统会根据后续服务表现继续识别潜力标签',
        actionText: '已解锁'
      }
    }
    config.roleSummary = consumeMonthlyQuota(
      config.roleSummary,
      getConfiguredCount(config.skillSlots)
    )

    this.applyConfigState(config, 'visible', {
      addPanel: {
        visible: false
      }
    })
    toast.success('技能已添加')
  },

  fillFirstEmptySlot(skillSlots = [], skill) {
    const nextSlots = (skillSlots || []).map((item) => ({ ...item }))
    const emptyIndex = nextSlots.findIndex((item) => item.empty && !item.locked)
    const nextSlot = {
      id: skill.id,
      title: skill.title,
      iconText: skill.iconText,
      tone: skill.tone,
      active: false,
      empty: false
    }

    if (emptyIndex === -1) {
      if (nextSlots.some((item) => item.empty && item.locked)) {
        return nextSlots
      }

      return nextSlots.concat(nextSlot)
    }

    nextSlots[emptyIndex] = nextSlot
    return nextSlots
  },

  handlePanelTap() {},

  handlePreventTouch() {
    return false
  },

  buildSavePayload() {
    return {
      roleSummary: this.data.roleSummary,
      activeTab: this.data.activeTab,
      skillSlots: this.data.skillSlots,
      skillGroups: this.data.skillGroups,
      unlockSuggestion: this.data.unlockSuggestion
    }
  },

  async handleSaveTap() {
    if (this.data.isSaving) {
      return
    }

    this.setData({
      isSaving: true
    })

    try {
      await profileService.saveSystemSkillConfig(this.buildSavePayload())
      toast.success('技能配置已保存')
    } catch (error) {
      toast.info(error.message || '技能配置保存失败，请稍后再试')
    } finally {
      this.setData({
        isSaving: false
      })
    }
  }
})
