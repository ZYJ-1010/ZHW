const DEFAULT_GAME_PARTICIPANTS = []

function getDefaultGameParticipants() {
  return DEFAULT_GAME_PARTICIPANTS.map((participant) => ({
    ...participant,
    tags: Array.isArray(participant.tags) ? participant.tags.slice() : []
  }))
}

module.exports = {
  getDefaultGameParticipants
}
