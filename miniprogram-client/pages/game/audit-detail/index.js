const { ROUTES } = require('../../../config/routes')
const { navigateShellBack, navigateShellKey, navigateShellRoute } = require('../../../utils/shell-nav')
const fileService = require('../../../services/file')
const gameService = require('../../../services/game')
const toast = require('../../../utils/toast')

const DETAIL_SCROLL_TAP_STEP_RPX = 360
const DETAIL_SCROLL_HOLD_STEP_RPX = 72
const DETAIL_SCROLL_HOLD_INTERVAL_MS = 80
const DETAIL_SCROLL_HOLD_SUPPRESS_TAP_MS = 120
const IMAGE_EXTENSIONS = ['jpg', 'jpeg', 'png', 'gif', 'webp']
const PDF_EXTENSIONS = ['pdf']

const EMPTY_PLAYER = {
  requirementConfirmed: false,
  requirementStatusText: '',
  confirmed: false,
  statusText: '',
  name: '',
  avatarText: '',
  avatarSrc: '',
  avatarType: '',
  avatarClass: '',
  desc: '',
  roleKey: '',
  roleName: '',
  tags: [],
  badges: [],
  metrics: [],
  profileSections: [],
  needTitle: '',
  needText: '',
  expectedTime: '',
  remark: ''
}

const EMPTY_GUIDE = {
  iconSrc: '/pages/game/guide-chat/assets/icon-invite.png',
  online: false,
  avatarText: '',
  avatarSrc: '',
  avatarType: '',
  name: '',
  recommendation: ''
}

const EMPTY_GAME_INFO = {
  topic: '',
  time: '',
  location: '',
  activityType: '',
  serviceDuration: '',
  clientBudget: ''
}
const EMPTY_DETAIL_CONFIG = {
  pageTitle: '',
  referralText: '',
  statusTitles: {},
  countdownTexts: {},
  playerStatusTexts: {},
  roleNames: {},
  texts: {},
  sessionItems: [],
  confirmRows: [],
  optionalActions: [],
  noticeBullets: []
}

const OPTION_ICON_MAP = {
  time: '/pages/game/audit-detail/assets/option-time.svg',
  chat: '/pages/game/audit-detail/assets/option-chat.svg',
  message: '/pages/game/audit-detail/assets/option-chat.svg'
}

const PLAYER_INFO_ICON_MAP = {
  time: '/pages/game/audit-detail/assets/player-time.svg',
  chat: '/pages/game/audit-detail/assets/player-chat.svg'
}

function applyTemplate(template, values = {}) {
  let text = String(template || '')

  Object.keys(values).forEach((key) => {
    text = text.replace(new RegExp(`\\{${key}\\}`, 'g'), String(values[key]))
  })

  return text
}

function normalizeDetailConfig(source = {}) {
  const auditPage = source.auditPage || {}
  const detail = auditPage.detail || source.detail || source

  return {
    pageTitle: String(detail.pageTitle || ''),
    referralText: String(detail.referralText || ''),
    statusTitles: detail.statusTitles || {},
    countdownTexts: detail.countdownTexts || {},
    playerStatusTexts: detail.playerStatusTexts || {},
    roleNames: auditPage.roleNames || detail.roleNames || {},
    texts: detail.texts || {},
    sessionItems: Array.isArray(detail.sessionItems) ? detail.sessionItems : [],
    confirmRows: Array.isArray(detail.confirmRows) ? detail.confirmRows : [],
    optionalActions: normalizeOptionalActions(detail.optionalActions),
    noticeBullets: Array.isArray(detail.noticeBullets) ? detail.noticeBullets : []
  }
}

function normalizeOptionalActionKey(key) {
  return key === 'message' ? 'chat' : key
}

function normalizeOptionalActions(actions = []) {
  if (!Array.isArray(actions)) {
    return []
  }

  return actions.map((item = {}) => {
    const rawKey = item.key || ''
    const key = normalizeOptionalActionKey(rawKey)

    return {
      ...item,
      key,
      iconSrc: OPTION_ICON_MAP[rawKey] || OPTION_ICON_MAP[key] || item.iconSrc || ''
    }
  })
}

function playerInfoSectionIconSrc(section = {}) {
  const key = String(section.key || section.tone || '').toLowerCase()
  const title = String(section.title || section.label || '')

  if (key === 'time' || key === 'expectedtime' || title.includes('时间')) {
    return PLAYER_INFO_ICON_MAP.time
  }
  if (key === 'remark' || key === 'note' || key === 'chat' || title.includes('备注')) {
    return PLAYER_INFO_ICON_MAP.chat
  }
  return ''
}

function normalizePlayerInfoSections(sections = []) {
  if (!Array.isArray(sections)) {
    return []
  }

  return sections.map((section = {}) => {
    const iconSrc = section.iconSrc || playerInfoSectionIconSrc(section)

    return iconSrc ? { ...section, iconSrc } : { ...section }
  })
}

function normalizeAuditDetailRole(value) {
  const role = String(value || '').trim()

  if (role === 'expert' || role === '行家') {
    return 'expert'
  }
  if (role === 'guide' || role === 'leader' || role === '领路人' || role === 'main_guide' || role === '主行家') {
    return 'guide'
  }
  return 'player'
}

function normalizeStatus(status) {
  const value = String(status || '').toLowerCase()

  if (value === 'approved' || value === 'pass' || value === 'passed' || value === 'accepted' || value === 'confirmed') {
    return 'approved'
  }

  if (value === 'rejected' || value === 'reject' || value === 'declined' || value === 'refused') {
    return 'rejected'
  }

  return 'pending'
}

function asArray(value) {
  return Array.isArray(value) ? value : []
}

function firstValue() {
  const values = Array.prototype.slice.call(arguments)

  for (let index = 0; index < values.length; index += 1) {
    if (values[index] !== undefined && values[index] !== null && values[index] !== '') {
      return values[index]
    }
  }

  return ''
}

function formatCountdownSeconds(value) {
  let seconds = Number(value || 0)
  if (!Number.isFinite(seconds) || seconds < 0) {
    seconds = 0
  }
  seconds = Math.floor(seconds)
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const remainSeconds = seconds % 60
  return [hours, minutes, remainSeconds].map((item) => String(item).padStart(2, '0')).join(':')
}

function invitationRemainingSeconds(item = {}, detailDisplay = {}) {
  const rawSeconds = firstValue(
    detailDisplay.remainingSeconds,
    detailDisplay.remainingSecond,
    detailDisplay.countdownSeconds,
    item.remainingSeconds,
    item.remainingSecond,
    item.countdownSeconds
  )
  const seconds = Number(rawSeconds)
  if (Number.isFinite(seconds)) {
    return Math.max(0, Math.floor(seconds))
  }

  const expireText = firstValue(detailDisplay.expiresAt, detailDisplay.timeoutAt, item.expiresAt, item.timeoutAt)
  const expireTime = expireText ? new Date(expireText).getTime() : 0
  if (!expireTime || Number.isNaN(expireTime)) {
    return 0
  }

  return Math.max(0, Math.ceil((expireTime - Date.now()) / 1000))
}

function invitationCountdownText(item = {}, statusDisplay = {}, detailDisplay = {}) {
  const backendText = firstValue(
    statusDisplay.countdown,
    detailDisplay.countdownText,
    detailDisplay.remainingText,
    item.countdownText,
    item.remainingText
  )
  if (backendText) {
    return backendText
  }

  if (String(item.status || '').toLowerCase() !== 'pending') {
    return ''
  }

  const rawSeconds = firstValue(
    detailDisplay.remainingSeconds,
    detailDisplay.remainingSecond,
    detailDisplay.countdownSeconds,
    item.remainingSeconds,
    item.remainingSecond,
    item.countdownSeconds
  )
  const seconds = rawSeconds !== '' ? invitationRemainingSeconds(item, detailDisplay) : 0
  return seconds > 0 ? formatCountdownSeconds(seconds) : ''
}

function getFileSource(file = {}) {
  return file.name || file.fileName || file.tempFilePath || file.path || ''
}

function getFileName(file = {}) {
  const source = getFileSource(file)
  const parts = source.split(/[\\/]/)

  return parts[parts.length - 1] || ''
}

function getFileExtension(file = {}) {
  const name = getFileName(file)
  const matched = name.match(/\.([a-zA-Z0-9]+)(?:\?|#)?$/)

  return matched ? matched[1].toLowerCase() : ''
}

function normalizeUploadFile(file = {}) {
  return {
    name: getFileName(file),
    path: file.tempFilePath || file.path || '',
    size: file.size || 0,
    extension: getFileExtension(file)
  }
}

function isAllowedImageFile(file = {}) {
  const extension = getFileExtension(file)

  if (extension) {
    return IMAGE_EXTENSIONS.includes(extension)
  }

  return file.fileType === 'image' || file.type === 'image'
}

function isAllowedPdfFile(file = {}) {
  return PDF_EXTENSIONS.includes(getFileExtension(file))
}

function sessionValue(gameInfo, key) {
  if (key === 'topic') return gameInfo.topic
  if (key === 'time') return gameInfo.time
  if (key === 'location') return gameInfo.location
  return ''
}

function buildSessionInfo(gameInfo, config = EMPTY_DETAIL_CONFIG) {
  return (config.sessionItems || []).map((item) => ({
    label: item.label || '',
    value: sessionValue(gameInfo, item.key),
    iconText: item.iconText || '',
    iconSrc: item.iconSrc || '',
    iconClass: item.iconClass || '',
    actionText: item.actionText || ''
  }))
}

function confirmRowValue(gameInfo, settlement, key) {
  if (key === 'activityType') return gameInfo.activityType
  if (key === 'serviceDuration') return gameInfo.serviceDuration
  if (key === 'clientBudget') return gameInfo.clientBudget
  if (key === 'platformFee') return settlement.platformFee || ''
  if (key === 'guideReward') return settlement.guideReward || ''
  if (key === 'partnerReward') return settlement.partnerReward || ''
  if (key === 'expertIncome') return settlement.expertIncome || ''
  return ''
}

function buildConfirmRows(gameInfo, settlement = {}, config = EMPTY_DETAIL_CONFIG) {
  return (config.confirmRows || []).map((item, index) => {
    const value = confirmRowValue(gameInfo, settlement, item.key)

    return {
      label: item.label || '',
      value,
      highlight: item.key === 'clientBudget' && Boolean(value),
      divider: item.key === 'clientBudget' || item.key === 'partnerReward' || Boolean(item.divider),
      success: item.key === 'expertIncome' && Boolean(value),
      total: item.key === 'expertIncome' || Boolean(item.total),
      _index: index
    }
  }).filter((row) => row.label && row.value)
}

function normalizeDisplayRows(rows = []) {
  if (!Array.isArray(rows)) {
    return []
  }

  return rows.map((item, index) => ({
    key: item.key || item.label || `row-${index}`,
    label: item.label || '',
    value: item.value || '',
    iconText: item.iconText || '',
    iconSrc: item.iconSrc || '',
    iconClass: item.iconClass || '',
    actionText: item.actionText || '',
    highlight: Boolean(item.highlight),
    divider: Boolean(item.divider),
    success: Boolean(item.success),
    total: Boolean(item.total),
    _index: index
  })).filter((row) => row.label && (row.value || row.actionText))
}

function attachInvitationInfoIcons(rows = []) {
  return rows.map((row) => {
    if (row.iconSrc || row.iconText) {
      return row
    }
    if (row.label === '主题') {
      return { ...row, iconText: 'H' }
    }
    if (row.label === '时间') {
      return { ...row, iconSrc: '/pages/game/guide-chat/assets/icon-time.svg' }
    }
    if (row.label === '地点') {
      return { ...row, iconSrc: '/pages/game/guide-chat/assets/icon-location.svg' }
    }
    if (row.label === '费用') {
      return { ...row, iconSrc: '/pages/game/audit-detail/assets/icon-confirm-green.svg' }
    }
    return row
  })
}

function invitationBudgetDetailText(game = {}, gameFee = '') {
  const amountText = firstValue(
    game.clientBudget,
    game.clientBudgetText,
    game.budgetAmountText,
    game.budgetText,
    game.amountText
  )
  if (amountText) {
    return amountText
  }
  return ''
}

function buildInvitationDetailRows(backendDisplay = {}, game = {}, gameFee = '') {
  const backendRows = normalizeDisplayRows(Array.isArray(backendDisplay.confirmRows) ? backendDisplay.confirmRows : [])
  if (backendRows.length) {
    return backendRows
  }

  return normalizeDisplayRows([
    {
      label: '服务类型',
      value: firstValue(game.activityType, game.categoryText, game.typeText, game.gameTypeText)
    },
    {
      label: '咨询时长',
      value: firstValue(game.serviceDuration, game.serviceDurationText, game.durationText)
    },
    {
      label: '预算金额',
      value: invitationBudgetDetailText(game, gameFee),
      highlight: true
    },
    {
      label: '预计时间',
      value: firstValue(game.expectedTime, game.expectedTimeText, game.estimateTimeText, game.deliveryTimeText)
    }
  ])
}

function splitProfileTags(text = '') {
  const tags = String(text || '').split(/[、,，/；;|\s]+/).map((item) => item.trim()).filter(Boolean)
  return tags.filter((item, index) => tags.indexOf(item) === index)
}

function sectionText(sections = [], title) {
  const match = sections.find((item) => item && item.title === title && item.text)
  return match ? match.text : ''
}

function sectionsText(sections = [], titles = []) {
  for (let index = 0; index < titles.length; index += 1) {
    const text = sectionText(sections, titles[index])
    if (text) {
      return text
    }
  }
  return ''
}

function normalizeInvitationProfileDisplay(profile = {}) {
  if (profile.roleKey !== 'expert') {
    return profile
  }
  const sections = Array.isArray(profile.profileSections) ? profile.profileSections : []
  const professionalText = firstValue(sectionsText(sections, ['专业领域', '服务标签']), profile.tags && profile.tags.join('、'))
  const serviceText = firstValue(sectionsText(sections, ['服务介绍', '服务说明', '个人简介']), profile.needText)
  const profileSections = []
  if (professionalText) {
    profileSections.push({
      title: '专业领域',
      text: professionalText,
      tags: splitProfileTags(professionalText),
      tone: 'domain'
    })
  }
  if (serviceText) {
    profileSections.push({
      title: '服务介绍',
      text: serviceText,
      tone: 'service'
    })
  }
  return {
    ...profile,
    hideRoleBadge: false,
    hideTags: true,
    compactCard: true,
    compactProfile: true,
    needTitle: '',
    needText: '',
    profileSections
  }
}

function findInvitationProgressItem(data = {}, id) {
  const activeList = asArray(data.activeParties || data.activeList || data.ongoingList || data.processingList)
  const completedList = asArray(data.completedParties || data.completedList || data.recentCompleted || data.historyList)
  const list = activeList.concat(completedList)

  if (!id) {
    return list[0]
  }

  return list.find((item) => String(item.id || item.invitationId || '') === String(id)) || null
}

function invitationAuditTarget(item = {}) {
  const targetUserId = String(firstValue(item.targetUserId, item.targetUserID)).trim()
  const players = asArray(item.players).length ? asArray(item.players) : (item.player ? [item.player] : [])
  const experts = asArray(item.experts).length ? asArray(item.experts) : (item.expert ? [item.expert] : [])
  let targetRole = normalizeAuditDetailRole(firstValue(item.targetRole, item.roleKey, item.role))

  if (targetUserId) {
    const matchedExpert = experts.find((member) => String(firstValue(member.userId, member.id)).trim() === targetUserId)
    const matchedPlayer = players.find((member) => String(firstValue(member.userId, member.id)).trim() === targetUserId)

    if (matchedExpert) {
      targetRole = 'expert'
      return {
        target: matchedExpert,
        targetRole,
        targetRoleName: matchedExpert.roleLabel || '行家',
        players,
        experts
      }
    }

    if (matchedPlayer) {
      targetRole = 'player'
      return {
        target: matchedPlayer,
        targetRole,
        targetRoleName: matchedPlayer.roleLabel || '玩家',
        players,
        experts
      }
    }
  }

  const target = targetRole === 'expert' ? (experts[0] || {}) : (players[0] || {})

  return {
    target,
    targetRole,
    targetRoleName: target.roleLabel || (targetRole === 'expert' ? '行家' : '玩家'),
    players,
    experts
  }
}

function normalizeInvitationAuditDetail(item = {}, detailConfig = EMPTY_DETAIL_CONFIG) {
  const game = item.game || item.gameInfo || {}
  const inviter = item.inviter || item.guide || {}
  const targetInfo = invitationAuditTarget(item)
  const target = targetInfo.target
  const isPlayerInvitation = targetInfo.targetRole === 'player'
  const player = targetInfo.players[0] || item.player || {}
  const expert = targetInfo.experts[0] || item.expert || {}
  const displayRole = targetInfo.targetRole === 'player' ? 'expert' : 'player'
  const displayPerson = displayRole === 'expert' ? expert : player
  const displayRoleName = displayPerson.roleLabel || (displayRole === 'expert' ? '行家' : '玩家')
  const backendDisplay = item.detailDisplay && typeof item.detailDisplay === 'object' ? item.detailDisplay : {}
  const backendPlayerDisplay = backendDisplay.player && typeof backendDisplay.player === 'object' ? backendDisplay.player : {}
  const backendRelationDisplay = backendDisplay.relation && typeof backendDisplay.relation === 'object' ? backendDisplay.relation : {}
  const backendStatusDisplay = backendDisplay.status && typeof backendDisplay.status === 'object' ? backendDisplay.status : {}
  const backendGuideDisplay = backendDisplay.guide && typeof backendDisplay.guide === 'object' ? backendDisplay.guide : {}
  const status = String(item.status || '').toLowerCase()
  const statusTitle = status === 'accepted'
    ? '已确认参加'
    : (status === 'rejected' ? '已婉拒' : '等待确认')
  const countdownText = invitationCountdownText(item, backendStatusDisplay, backendDisplay)
  const gameTitle = firstValue(game.topic, game.title, item.gameTitle)
  const gameTime = firstValue(game.timeText, game.time, item.timeText)
  const gameLocation = firstValue(game.locationText, game.location)
  const gameFee = firstValue(game.feeText, game.fee, item.feeText, item.priceText)
  const baseInfoRows = normalizeDisplayRows(item.infoRows && item.infoRows.length ? item.infoRows : [
    { label: '主题', value: gameTitle },
    { label: '时间', value: gameTime },
    { label: '地点', value: gameLocation }
  ])
  const infoRows = attachInvitationInfoIcons(gameFee && !baseInfoRows.some((row) => row.label === '费用')
    ? baseInfoRows.concat([{ label: '费用', value: gameFee }])
    : baseInfoRows)
  const visibleInfoRows = infoRows
    .filter((row) => row.label !== '费用')
    .map((row) => (row.label === '地点'
      ? { ...row, actionText: row.actionText || '地图位置' }
      : row))
  const detailRows = buildInvitationDetailRows(backendDisplay, game, gameFee)
  const hasBackendPlayerDisplay = Object.keys(backendPlayerDisplay).length > 0
  const displayProfile = normalizeInvitationProfileDisplay(hasBackendPlayerDisplay ? {
    ...EMPTY_PLAYER,
    ...backendPlayerDisplay,
    tags: asArray(backendPlayerDisplay.tags),
    badges: asArray(backendPlayerDisplay.badges),
    metrics: asArray(backendPlayerDisplay.metrics),
    profileSections: asArray(backendPlayerDisplay.profileSections)
  } : {
    requirementConfirmed: status !== 'pending',
    requirementStatusText: statusTitle,
    confirmed: status === 'accepted',
    statusText: statusTitle,
    name: displayPerson.name,
    avatarText: displayPerson.avatarText,
    avatarSrc: displayPerson.avatarSrc || displayPerson.avatarUrl,
    avatarType: displayPerson.avatarType,
    avatarClass: displayPerson.avatarClass,
    desc: displayRoleName,
    roleKey: displayRole,
    roleName: displayRoleName,
    tags: [displayRoleName],
    badges: [],
    metrics: [],
    profileSections: [],
    needTitle: '',
    needText: firstValue(item.reason, item.chatMessageText),
    expectedTime: '',
    remark: firstValue(item.timeText, item.createdAt) ? `邀请时间：${firstValue(item.timeText, item.createdAt)}` : ''
  })
  if (hasBackendPlayerDisplay && !displayProfile.requirementStatusText) {
    const profileStatus = String(backendPlayerDisplay.status || '').toLowerCase()
    const profileStatusText = firstValue(
      backendPlayerDisplay.statusText,
      backendPlayerDisplay.state
    )
    const profileConfirmed = profileStatus === 'approved' ||
      profileStatus === 'accepted' ||
      backendPlayerDisplay.stateClass === 'confirmed' ||
      backendPlayerDisplay.confirmed === true
    displayProfile.requirementConfirmed = profileConfirmed
    displayProfile.requirementStatusText = profileStatusText || (profileConfirmed ? '已确认' : '')
  }
  displayProfile.profileSections = normalizePlayerInfoSections(displayProfile.profileSections)
  const relationDisplay = Object.keys(backendRelationDisplay).length ? { ...backendRelationDisplay } : {
    variant: 'parties',
    title: '成局双方',
    confirmedText: firstValue(item.confirmedText, item.confirmedCountText, game.confirmedText),
    expert: {
      avatarText: expert.avatarText,
      avatarSrc: expert.avatarSrc || expert.avatarUrl,
      avatarType: expert.avatarType,
      avatarClass: expert.avatarClass || 'blue',
      name: expert.name,
      role: expert.roleLabel,
      roleText: expert.roleLabel,
      state: expert.statusText || expert.state,
      stateClass: expert.stateClass
    },
    player: {
      avatarText: player.avatarText,
      avatarSrc: player.avatarSrc || player.avatarUrl,
      avatarType: player.avatarType,
      avatarClass: player.avatarClass || 'pink',
      name: player.name,
      role: player.roleLabel,
      roleText: player.roleLabel,
      state: player.statusText || player.state,
      stateClass: player.stateClass
    }
  }
  relationDisplay.cardClass = `${relationDisplay.cardClass || ''} audit-invitation-party-card`.trim()
  relationDisplay.titleClass = relationDisplay.titleClass || 'regular'

  return normalizeApplicationDetail({
    id: item.invitationId || item.id,
    invitationId: item.invitationId || item.id,
    gameId: item.gameId || game.id || '',
    status,
    role: displayRole,
    roleKey: displayRole,
    nickname: displayPerson.name,
    avatarText: displayPerson.avatarText,
    avatarSrc: displayPerson.avatarSrc || displayPerson.avatarUrl,
    avatarType: displayPerson.avatarType,
    avatarClass: displayPerson.avatarClass,
    reason: firstValue(item.reason, item.chatMessageText),
    createdAtText: firstValue(item.timeText, item.createdAt),
    tags: [displayRoleName],
    guideName: inviter.name,
    guideAvatarText: inviter.avatarText,
    guideAvatarSrc: inviter.avatarSrc || inviter.avatarUrl,
    guideAvatarType: inviter.avatarType,
    detailDisplay: {
      pageTitle: firstValue(backendDisplay.pageTitle, isPlayerInvitation ? '组局详情' : '审核组局'),
      referralText: firstValue(backendDisplay.referralText, '邀请你参与组局'),
      applicantTitle: firstValue(backendDisplay.applicantTitle, `${displayRoleName}信息`),
      confirmTitle: firstValue(backendDisplay.confirmTitle, '组局信息'),
      optionTitle: firstValue(backendDisplay.optionTitle, '可选操作'),
      noticeTitle: firstValue(backendDisplay.noticeTitle, '确认须知'),
      actionTip: firstValue(backendDisplay.actionTip, '确认后将建立三方连接群并冻结资金'),
      confirmText: firstValue(backendDisplay.confirmText, item.chatAcceptButtonText, '确认通过'),
      declineText: firstValue(backendDisplay.declineText, item.chatDeclineButtonText, '婉拒'),
      canConfirm: backendDisplay.canConfirm !== false,
      confirmDisabledReason: firstValue(backendDisplay.confirmDisabledReason, backendDisplay.disabledReason),
      showRelation: displayFlag(backendDisplay, 'showRelation', true),
      showPortfolio: displayFlag(backendDisplay, 'showPortfolio', !isPlayerInvitation),
      showSession: displayFlag(backendDisplay, 'showSession', true),
      showConfirm: displayFlag(backendDisplay, 'showConfirm', true),
      showRecommend: displayFlag(backendDisplay, 'showRecommend', !isPlayerInvitation),
      showOptions: displayFlag(backendDisplay, 'showOptions', true),
      showNotice: displayFlag(backendDisplay, 'showNotice', true),
      showActionBar: displayFlag(backendDisplay, 'showActionBar', status === 'pending'),
      reviewReadonlyText: firstValue(backendDisplay.reviewReadonlyText, statusTitle),
      status: {
        title: firstValue(backendStatusDisplay.title, statusTitle),
        quote: firstValue(backendStatusDisplay.quote, item.reason, item.chatMessageText),
        hideQuote: backendStatusDisplay.hideQuote === true,
        guideName: firstValue(backendStatusDisplay.guideName, inviter.name),
        playerName: firstValue(backendStatusDisplay.playerName, player.name),
        countdown: countdownText
      },
      countdownText,
      guide: {
        name: firstValue(backendGuideDisplay.name, inviter.name),
        recommendation: firstValue(backendGuideDisplay.recommendation, item.guideRecommendation, item.recommendation)
      },
      relation: relationDisplay,
      player: displayProfile,
      gameInfo: {
        topic: gameTitle,
        time: gameTime,
        location: gameLocation,
        activityType: firstValue(game.activityType, game.categoryText),
        serviceDuration: '',
        clientBudget: '',
        ...(backendDisplay.gameInfo && typeof backendDisplay.gameInfo === 'object' ? backendDisplay.gameInfo : {})
      },
      sessionInfo: Array.isArray(backendDisplay.sessionInfo) ? backendDisplay.sessionInfo : visibleInfoRows,
      confirmRows: detailRows,
      optionalActions: normalizeOptionalActions(Array.isArray(backendDisplay.optionalActions) ? backendDisplay.optionalActions : []),
      noticeBullets: Array.isArray(backendDisplay.noticeBullets) ? backendDisplay.noticeBullets : [
        '确认后请准时参加，如需取消请提前通知',
        '双方确认后组局正式生效，领路人将获得积分奖励',
        '请保持专业态度，维护平台信誉'
      ]
    }
  }, detailConfig)
}

function displayFlag(display, key, fallback) {
  if (display && Object.prototype.hasOwnProperty.call(display, key)) {
    return display[key] === true
  }

  return fallback
}

function normalizeApplicationDetail(item = {}, detailConfig = EMPTY_DETAIL_CONFIG) {
  const detailDisplay = item.detailDisplay && typeof item.detailDisplay === 'object' ? item.detailDisplay : {}
  const hasBackendDisplay = Object.keys(detailDisplay).length > 0
  const rawId = firstValue(item.id, item.applicationId)
  const hasDetail = Boolean(rawId)
  const status = normalizeStatus(item.status || item.statusKey)
  const nickname = firstValue(item.nickname, item.userName, item.userNickname, item.user && item.user.nickname)
  const reason = String(firstValue(item.reason, item.remark, item.applyReason)).trim()
  const rejectReason = String(firstValue(item.rejectReason, item.reject_reason)).trim()
  const texts = Object.assign({}, detailConfig.texts || {}, detailDisplay.texts || {})
  const playerStatusTexts = detailConfig.playerStatusTexts || {}
  const roleKey = normalizeAuditDetailRole(firstValue(item.roleKey, item.role, item.roleType))
  const roleName = detailConfig.roleNames[roleKey] || (roleKey === 'expert' ? '行家' : (roleKey === 'guide' ? '领路人' : '玩家'))
  const tags = Array.isArray(item.tags) ? item.tags.slice() : []
  if (roleName && !tags.includes(roleName)) {
    tags.unshift(roleName)
  }
  const gameInfo = {
    ...EMPTY_GAME_INFO,
    topic: firstValue(item.gameTitle, item.title, item.game && item.game.title),
    time: firstValue(item.gameTimeText, item.timeText, item.game && item.game.timeText),
    location: firstValue(item.locationText, item.locationName, item.game && item.game.locationName),
    activityType: firstValue(item.activityType, item.gameTypeText, item.gameType),
    serviceDuration: firstValue(item.serviceDurationText, item.durationText),
    clientBudget: firstValue(item.clientBudgetText, item.budgetText, item.amountText)
  }
  if (detailDisplay.gameInfo && typeof detailDisplay.gameInfo === 'object') {
    Object.assign(gameInfo, detailDisplay.gameInfo)
  }
  const guideName = firstValue(item.guideName, item.referrerName, item.guide && item.guide.name)
  const guide = {
    ...EMPTY_GUIDE,
    online: Boolean(item.guideOnline || item.referrerOnline),
    avatarText: firstValue(item.guideAvatarText, item.referrerAvatarText, item.guide && item.guide.avatarText, texts.guideAvatarFallback),
    avatarSrc: firstValue(item.guideAvatarSrc, item.referrerAvatarSrc, item.guide && (item.guide.avatarSrc || item.guide.avatarUrl)),
    avatarType: firstValue(item.guideAvatarType, item.referrerAvatarType, item.guide && item.guide.avatarType),
    name: guideName,
    recommendation: firstValue(item.guideRecommendation, item.recommendation)
  }
  const player = {
    ...EMPTY_PLAYER,
    requirementConfirmed: status !== 'pending',
    requirementStatusText: status === 'pending' ? playerStatusTexts.pendingRequirement : playerStatusTexts.reviewedRequirement,
    confirmed: status !== 'rejected',
    statusText: playerStatusTexts[status] || '',
    name: nickname,
    avatarText: firstValue(item.avatarText, item.avatar, item.initials, texts.playerAvatarFallback),
    avatarSrc: firstValue(item.avatarSrc, item.avatarUrl),
    avatarType: firstValue(item.avatarType),
    desc: firstValue(item.userDesc, item.desc, item.profileText, roleName),
    roleKey,
    roleName,
    tags,
    needText: status === 'rejected' && rejectReason
      ? `驳回原因：${rejectReason}`
      : (reason ? `${texts.needPrefix || ''}${reason}` : ''),
    expectedTime: firstValue(item.expectedTimeText, item.expectedTime),
    remark: firstValue(item.createdAtText, item.applyTime, item.createdAt) ? `${texts.remarkPrefix || ''}${firstValue(item.createdAtText, item.applyTime, item.createdAt)}` : ''
  }
  if (detailDisplay.player && typeof detailDisplay.player === 'object') {
    Object.assign(player, detailDisplay.player)
  }
  player.profileSections = normalizePlayerInfoSections(player.profileSections)
  const settlement = item.settlement || item.backendSettlement || {}
  const sessionInfo = hasBackendDisplay && Array.isArray(detailDisplay.sessionInfo)
    ? normalizeDisplayRows(detailDisplay.sessionInfo)
    : buildSessionInfo(gameInfo, detailConfig)
  const confirmRows = hasBackendDisplay && Array.isArray(detailDisplay.confirmRows)
    ? normalizeDisplayRows(detailDisplay.confirmRows)
    : buildConfirmRows(gameInfo, settlement, detailConfig)
  const optionalActions = normalizeOptionalActions(hasBackendDisplay && Array.isArray(detailDisplay.optionalActions)
    ? detailDisplay.optionalActions
    : detailConfig.optionalActions)
  const noticeBullets = hasBackendDisplay && Array.isArray(detailDisplay.noticeBullets)
    ? detailDisplay.noticeBullets
    : detailConfig.noticeBullets
  const statusDisplay = detailDisplay.status && typeof detailDisplay.status === 'object' ? detailDisplay.status : {}
  const relationDisplay = detailDisplay.relation && typeof detailDisplay.relation === 'object' ? detailDisplay.relation : null
  const guideDisplay = detailDisplay.guide && typeof detailDisplay.guide === 'object' ? detailDisplay.guide : {}
  Object.assign(guide, guideDisplay)

  return {
    hasDetail,
    auditId: rawId,
    gameId: item.gameId || '',
    detailPageTitle: firstValue(detailDisplay.pageTitle, detailConfig.pageTitle),
    referralText: firstValue(detailDisplay.referralText, detailConfig.referralText),
    applicantTitle: firstValue(detailDisplay.applicantTitle, texts.playerTitle),
    confirmTitle: firstValue(detailDisplay.confirmTitle, texts.confirmTitle),
    optionTitle: firstValue(detailDisplay.optionTitle, texts.optionTitle),
    noticeTitle: firstValue(detailDisplay.noticeTitle, texts.noticeTitle),
    actionTip: firstValue(detailDisplay.actionTip, texts.actionTip),
    confirmText: firstValue(detailDisplay.confirmText, texts.confirmText),
    declineText: firstValue(detailDisplay.declineText, texts.declineText, '婉拒'),
    canConfirm: detailDisplay.canConfirm !== false,
    confirmDisabledReason: firstValue(detailDisplay.confirmDisabledReason, detailDisplay.disabledReason),
    showRelation: displayFlag(detailDisplay, 'showRelation', true),
    showPortfolio: displayFlag(detailDisplay, 'showPortfolio', !hasBackendDisplay),
    showSession: displayFlag(detailDisplay, 'showSession', true),
    showConfirm: displayFlag(detailDisplay, 'showConfirm', true),
    showRecommend: displayFlag(detailDisplay, 'showRecommend', !hasBackendDisplay),
    showOptions: displayFlag(detailDisplay, 'showOptions', true),
    showNotice: displayFlag(detailDisplay, 'showNotice', true),
    showActionBar: displayFlag(detailDisplay, 'showActionBar', status === 'pending'),
    reviewReadonlyText: firstValue(detailDisplay.reviewReadonlyText, statusDisplay.title, detailConfig.statusTitles[status]),
    optionalActions,
    noticeBullets,
    status: {
      title: firstValue(statusDisplay.title, detailConfig.statusTitles[status]),
      quote: firstValue(statusDisplay.quote, reason),
      guideName: firstValue(statusDisplay.guideName, guide.name),
      countdown: firstValue(statusDisplay.countdown, detailConfig.countdownTexts[status])
    },
    relation: relationDisplay || {
      variant: 'relation',
      title: '组局关系图',
      totalCount: Number(item.totalCount || item.memberCount || 0),
      confirmedCount: Number(item.confirmedCount || 0),
      expert: {
        avatarText: texts.expertAvatarText || '',
        name: texts.expertName || '',
        roleText: texts.expertRoleText || ''
      },
      guide,
      player: {
        avatarText: player.avatarText,
        name: player.name,
        confirmed: player.confirmed,
        statusText: player.statusText
      }
    },
    game: {
      player,
      info: gameInfo,
      guide
    },
    backendSettlement: settlement,
    sessionInfo,
    confirmRows
  }
}

function buildEmptyDetail(detailConfig = EMPTY_DETAIL_CONFIG) {
  return Object.assign(normalizeApplicationDetail({}, detailConfig), { hasDetail: false })
}

Page({
  data: {
    auditId: '',
    invitationId: '',
    detailMode: 'application',
    hasDetail: false,
    actionLoading: false,
    rejectDialogVisible: false,
    rejectReasonDraft: '',
    uploadedFileIds: [],
    onlineText: '在线',
    countdownText: '',
    detailScrollTop: 0,
    detailConfig: EMPTY_DETAIL_CONFIG,
    navItems: [
      { name: '我的', active: false },
      { name: '元宇宙', active: false },
      { name: '地图', active: false },
      { name: '消息', active: false },
      { name: '首页', active: true }
    ],
    ...buildEmptyDetail(EMPTY_DETAIL_CONFIG),
    optionalActions: [],
    noticeBullets: []
  },

  onLoad(options = {}) {
    const invitationId = options.invitationId || options.inviteId || ''
    const auditId = options.auditId || options.applicationId || options.id || ''
    const detailMode = invitationId ? 'invitation' : 'application'

    // 处理开发环境 token
    const devToken = options.devToken
    if (devToken && typeof wx !== 'undefined' && typeof wx.setStorageSync === 'function') {
      const env = require('../../../config/env')
      if (env.currentMiniProgramEnv !== env.APP_ENV.RELEASE) {
        require('../../../utils/auth-session').setAuthToken(devToken)
      }
    }

    this.setData({ auditId, invitationId, detailMode })
    this.loadDetailConfig()
      .then(() => {
        if (detailMode === 'invitation') {
          return this.loadInvitationDetail(invitationId)
        }

        return this.loadApplicationDetail(auditId)
      })
  },

  async loadDetailConfig() {
    try {
      const config = await gameService.getApplicationConfig()
      const detailConfig = normalizeDetailConfig(config)
      this.setData({
        detailConfig,
        optionalActions: detailConfig.optionalActions,
        noticeBullets: detailConfig.noticeBullets,
        ...buildEmptyDetail(detailConfig)
      })
    } catch (error) {
      toast.info(error.message || this.textOf('loadFailedText'))
    }
  },

  textOf(key, values = {}) {
    return applyTemplate(this.data.detailConfig && this.data.detailConfig.texts && this.data.detailConfig.texts[key], values)
  },

  async loadApplicationDetail(auditId = this.data.auditId) {
    if (!auditId) {
      return
    }

    this.clearInvitationCountdownTimer()

    try {
      const data = await gameService.getReceivedApplications({})
      const items = data.items || data.applications || []
      const application = items.find((item) => String(item.id || item.applicationId || '') === String(auditId))

      if (!application) {
        toast.info(this.textOf('detailMissingText'))
        this.setData({
          ...buildEmptyDetail(this.data.detailConfig),
          auditId,
          hasDetail: false
        })
        return
      }

      this.setData(normalizeApplicationDetail(application, this.data.detailConfig))
    } catch (error) {
      toast.info(error.message || this.textOf('loadFailedText'))
    }
  },

  async loadInvitationDetail(invitationId = this.data.invitationId) {
    if (!invitationId) {
      return
    }

    this.clearInvitationCountdownTimer()

    try {
      const data = await gameService.getGuideProgress({ invitationId })
      const invitation = findInvitationProgressItem(data, invitationId)

      if (!invitation) {
        toast.info(this.textOf('detailMissingText'))
        this.setData({
          ...buildEmptyDetail(this.data.detailConfig),
          auditId: invitationId,
          invitationId,
          detailMode: 'invitation',
          hasDetail: false
        })
        return
      }

      const normalized = normalizeInvitationAuditDetail(invitation, this.data.detailConfig)

      this.setData({
        ...normalized,
        auditId: invitation.id || invitation.invitationId || invitationId,
        invitationId: invitation.invitationId || invitation.id || invitationId,
        detailMode: 'invitation'
      })
      this.startInvitationCountdown(invitation)
    } catch (error) {
      toast.info(error.message || this.textOf('loadFailedText'))
    }
  },

  startInvitationCountdown(invitation = {}) {
    this.clearInvitationCountdownTimer()

    if (String(invitation.status || '').toLowerCase() !== 'pending') {
      return
    }

    this.invitationCountdownSeconds = invitationRemainingSeconds(invitation, invitation.detailDisplay || {})
    if (this.invitationCountdownSeconds <= 0) {
      const staticCountdownText = invitationCountdownText(invitation, this.data.status || {}, invitation.detailDisplay || {})
      if (staticCountdownText) {
        this.setData({
          countdownText: staticCountdownText,
          'status.countdown': staticCountdownText
        })
      }
      return
    }

    this.updateInvitationCountdown()

    this.invitationCountdownTimer = setInterval(() => {
      this.invitationCountdownSeconds = Math.max(0, Number(this.invitationCountdownSeconds || 0) - 1)
      this.updateInvitationCountdown()

      if (this.invitationCountdownSeconds <= 0) {
        this.clearInvitationCountdownTimer()
      }
    }, 1000)
  },

  updateInvitationCountdown() {
    if (this.data.detailMode !== 'invitation' || !this.data.status) {
      return
    }

    this.setData({
      countdownText: formatCountdownSeconds(this.invitationCountdownSeconds),
      'status.countdown': formatCountdownSeconds(this.invitationCountdownSeconds)
    })
  },

  clearInvitationCountdownTimer() {
    if (this.invitationCountdownTimer) {
      clearInterval(this.invitationCountdownTimer)
      this.invitationCountdownTimer = null
    }
  },

  handleMapTap() {
    const params = []
    const gameId = this.data.gameId || ''
    const title = this.data.game && this.data.game.info ? this.data.game.info.topic : ''

    if (gameId) {
      params.push(`gameId=${encodeURIComponent(gameId)}`)
    }

    if (title) {
      params.push(`title=${encodeURIComponent(title)}`)
    }

    navigateShellRoute(`/${ROUTES.map}${params.length ? `?${params.join('&')}` : ''}`)
  },

  handleUploadImageTap() {
    if (wx.chooseMedia) {
      wx.chooseMedia({
        count: 1,
        mediaType: ['image'],
        sourceType: ['album', 'camera'],
        success: (result) => {
          const file = result.tempFiles && result.tempFiles[0]
          this.uploadAuditFile(file, 'image')
        },
        fail: (error) => {
          this.handleChooseFileFail(error)
        }
      })
      return
    }

    if (wx.chooseImage) {
      wx.chooseImage({
        count: 1,
        sourceType: ['album', 'camera'],
        success: (result) => {
          const file = (result.tempFiles && result.tempFiles[0]) || {
            tempFilePath: result.tempFilePaths && result.tempFilePaths[0],
            fileType: 'image'
          }
          this.uploadAuditFile(file, 'image')
        },
        fail: (error) => {
          this.handleChooseFileFail(error)
        }
      })
      return
    }

    toast.info(this.textOf('mediaUnsupportedText'))
  },

  handleUploadFileTap() {
    if (!wx.chooseMessageFile) {
      toast.info(this.textOf('fileUnsupportedText'))
      return
    }

    wx.chooseMessageFile({
      count: 1,
      type: 'file',
      extension: PDF_EXTENSIONS,
      success: (result) => {
        const file = result.tempFiles && result.tempFiles[0]
        this.uploadAuditFile(file, 'file')
      },
      fail: (error) => {
        this.handleChooseFileFail(error)
      }
    })
  },

  async uploadAuditFile(file, type) {
    if (!file) {
      return
    }

    if (type === 'image' && !isAllowedImageFile(file)) {
      toast.info(this.textOf('imageTypeErrorText'))
      return
    }

    if (type === 'file' && !isAllowedPdfFile(file)) {
      toast.info(this.textOf('fileTypeErrorText'))
      return
    }

    const uploadFile = normalizeUploadFile(file)
    if (!uploadFile.path) {
      toast.info(this.textOf('filePathInvalidText'))
      return
    }

    wx.showLoading({
      title: this.textOf('uploadingText'),
      mask: true
    })

    try {
      const fileIds = await fileService.uploadEvidenceImages([uploadFile.path], {
        bizType: 'game_application_audit',
        objectId: Number(this.data.gameId || 0) || 0
      })
      const nextFileIds = (this.data.uploadedFileIds || []).concat(fileIds || [])

      this.setData({
        uploadedFileIds: nextFileIds
      })
      toast.success(type === 'image' ? this.textOf('imageUploadedText') : this.textOf('fileUploadedText'))
    } catch (error) {
      toast.info(error.message || this.textOf('uploadFailedText'))
    } finally {
      wx.hideLoading()
    }
  },

  handleChooseFileFail(error = {}) {
    if (error.errMsg && error.errMsg.includes('cancel')) {
      return
    }

    toast.info(this.textOf('chooseFailedText'))
  },

  handleOptionalActionTap(event) {
    const key = event.currentTarget.dataset.key

    if (key === 'detail' || key === 'game' || key === 'gameDetail') {
      if (!this.data.hasDetail || !this.data.gameId) {
        toast.info(this.textOf('detailRequiredActionText'))
        return
      }

      this.navigateToRoute(`${ROUTES.gameDetail}?id=${encodeURIComponent(this.data.gameId)}`)
      return
    }

    if (key === 'chat') {
      if (!this.data.hasDetail) {
        toast.info(this.textOf('detailRequiredActionText'))
        return
      }

      const params = [
        `gameId=${encodeURIComponent(this.data.gameId || '')}`,
        `prefill=${encodeURIComponent(this.textOf('chatPrefill'))}`
      ].join('&')
      this.navigateToRoute(`${ROUTES.imRoom}?${params}`)
      return
    }

    if (key === 'time') {
      if (!this.data.hasDetail) {
        toast.info(this.textOf('detailRequiredActionText'))
        return
      }

      const params = [
        `gameId=${encodeURIComponent(this.data.gameId || '')}`,
        `prefill=${encodeURIComponent(this.textOf('timePrefill'))}`
      ].join('&')
      this.navigateToRoute(`${ROUTES.imRoom}?${params}`)
      return
    }

    toast.info(this.textOf('unavailableActionText'))
  },

  handleDeclineTap() {
    if (this.data.detailMode === 'invitation') {
      this.reviewCurrentApplication(false)
      return
    }
    this.setData({ rejectDialogVisible: true, rejectReasonDraft: '' })
  },

  handleRejectReasonInput(event) {
    this.setData({ rejectReasonDraft: event.detail.value || '' })
  },

  closeRejectDialog() {
    this.setData({ rejectDialogVisible: false, rejectReasonDraft: '' })
  },

  noop() {},

  handleRejectReasonConfirm() {
    const reason = String(this.data.rejectReasonDraft || '').trim()
    if (!reason) {
      toast.info('请填写驳回理由')
      return
    }
    this.closeRejectDialog()
    this.reviewCurrentApplication(false, reason)
  },

  handleApproveTap() {
    this.reviewCurrentApplication(true)
  },

  handleConfirmDisabledTap(event) {
    const reason = event.detail && event.detail.reason
    toast.info(reason || this.data.confirmDisabledReason || '当前暂不能确认')
  },

  async reviewCurrentApplication(approve, rejectReason = '') {
    if (this.data.actionLoading || !this.data.auditId || !this.data.hasDetail) {
      if (!this.data.hasDetail) {
        toast.info(this.textOf('detailRequiredReviewText'))
      }
      return
    }
    if (approve && this.data.canConfirm === false) {
      toast.info(this.data.confirmDisabledReason || '当前暂不能确认')
      return
    }

    this.setData({ actionLoading: true })
    wx.showLoading({
      title: approve ? this.textOf('approvingText') : this.textOf('rejectingText'),
      mask: true
    })

    try {
      if (this.data.detailMode === 'invitation') {
        const invitationId = this.data.invitationId || this.data.auditId
        await gameService.respondGameInvitation(invitationId, approve ? 'accept' : 'reject')
        await this.loadInvitationDetail(invitationId)
        toast.success(approve ? '已确认参加' : '已婉拒')
        if (approve && this.data.gameId) {
          navigateShellRoute(`${ROUTES.gameSuccessExpert}?gameId=${encodeURIComponent(this.data.gameId)}`, {
            reuseExisting: false
          })
          return
        }
      } else {
        const application = await gameService.reviewGameApplication(this.data.auditId, approve, approve ? '' : rejectReason)
        this.setData(normalizeApplicationDetail(application, this.data.detailConfig))
        toast.success(approve ? this.textOf('approveSuccessText') : this.textOf('rejectSuccessText'))
      }
    } catch (error) {
      toast.info(error.message || this.textOf('reviewFailedText'))
    } finally {
      wx.hideLoading()
      this.setData({ actionLoading: false })
    }
  },

  handleShellNavTap(event) {
    const key = event.detail && event.detail.key

    if (key === 'left') {
      if (navigateShellBack()) {
        return
      }

      navigateShellRoute(ROUTES.message || ROUTES.home)
      return
    }

    if (key === 'up' || key === 'down') {
      if (!this.suppressNextNavTap) {
        this.scrollDetail(key, DETAIL_SCROLL_TAP_STEP_RPX)
      }
      return
    }

    if (navigateShellKey(key, {
      currentRoute: ROUTES.gameAuditDetail
    })) {
      return
    }

    this.handleShellAction(key)
  },

  handleShellNavLongPress(event) {
    const key = event.detail && event.detail.key

    if (key !== 'up' && key !== 'down') {
      return
    }

    this.suppressNextNavTap = true
    this.stopDetailScrollHold(false)
    this.scrollDetail(key, DETAIL_SCROLL_HOLD_STEP_RPX)

    this.detailScrollHoldTimer = setInterval(() => {
      this.scrollDetail(key, DETAIL_SCROLL_HOLD_STEP_RPX)
    }, DETAIL_SCROLL_HOLD_INTERVAL_MS)
  },

  handleShellNavTouchEnd() {
    this.stopDetailScrollHold(true)
  },

  handleShellAction(key) {
    if (key === 'home') {
      this.scrollDetailToTop()
      return
    }

    if (key === 'search') {
      this.navigateToRoute(ROUTES.gameHall)
      return
    }

    if (key === 'comment' || key === 'message') {
      this.navigateToRoute(ROUTES.message)
      return
    }

    if (key === 'avatar' || key === 'mine') {
      this.navigateToRoute(ROUTES.profile)
      return
    }

    const routeMap = {
      metaverse: ROUTES.metaverse,
      map: ROUTES.map
    }

    this.navigateToRoute(routeMap[key])
  },

  navigateToRoute(route) {
    if (!route || route === ROUTES.gameAuditDetail) {
      return
    }

    navigateShellRoute(route)
  },

  handleDetailScroll(event) {
    const scrollTop = event.detail && event.detail.scrollTop

    if (typeof scrollTop === 'number') {
      this.detailScrollTopValue = scrollTop
    }
  },

  scrollDetail(direction, stepRpx = DETAIL_SCROLL_TAP_STEP_RPX) {
    const current = Number(this.detailScrollTopValue || this.data.detailScrollTop || 0)
    const distance = this.rpxToPx(stepRpx)
    const nextTop = direction === 'up'
      ? Math.max(0, current - distance)
      : current + distance

    this.detailScrollTopValue = nextTop
    this.setData({
      detailScrollTop: nextTop
    })
  },

  scrollDetailToTop() {
    this.detailScrollTopValue = 0
    this.setData({
      detailScrollTop: 0
    })
  },

  stopDetailScrollHold(resetTapSuppress) {
    if (this.detailScrollHoldTimer) {
      clearInterval(this.detailScrollHoldTimer)
      this.detailScrollHoldTimer = null
    }

    if (resetTapSuppress && this.suppressNextNavTap) {
      if (this.detailScrollSuppressTimer) {
        clearTimeout(this.detailScrollSuppressTimer)
      }

      this.detailScrollSuppressTimer = setTimeout(() => {
        this.suppressNextNavTap = false
        this.detailScrollSuppressTimer = null
      }, DETAIL_SCROLL_HOLD_SUPPRESS_TAP_MS)
    }
  },

  clearDetailScrollTimers() {
    this.stopDetailScrollHold(false)

    if (this.detailScrollSuppressTimer) {
      clearTimeout(this.detailScrollSuppressTimer)
      this.detailScrollSuppressTimer = null
    }

    this.suppressNextNavTap = false
  },

  rpxToPx(value) {
    if (!wx.getSystemInfoSync) {
      return value / 2
    }

    const system = wx.getSystemInfoSync()
    const windowWidth = system && system.windowWidth ? system.windowWidth : 375

    return Math.round((value * windowWidth) / 750)
  },

  onUnload() {
    this.clearInvitationCountdownTimer()
    this.clearDetailScrollTimers()
  }
})
