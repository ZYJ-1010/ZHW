const GAME_ICON_SRC = {
  successCheck: '/pages/game/assets/icons/success-check-green.svg',
  trophy: '/pages/game/assets/icons/trophy-white.svg',
  guideFollowSchedule: '/pages/game/assets/icons/schedule-check-blue.svg',
  guideFollowFeedback: '/pages/game/assets/icons/feedback-chat-purple.svg',
  guideFollowDeal: '/pages/game/assets/icons/deal-handshake-yellow.svg',
  share: '/pages/game/assets/icons/share-white.svg'
}

const GUIDE_FOLLOW_ICON_SRC = {
  schedule: GAME_ICON_SRC.guideFollowSchedule,
  feedback: GAME_ICON_SRC.guideFollowFeedback,
  deal: GAME_ICON_SRC.guideFollowDeal
}

function resolveGuideFollowIconSrc(key, fallback = '') {
  return GUIDE_FOLLOW_ICON_SRC[key] || fallback
}

module.exports = {
  GAME_ICON_SRC,
  GUIDE_FOLLOW_ICON_SRC,
  resolveGuideFollowIconSrc
}
