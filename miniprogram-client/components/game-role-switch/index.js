const userService = require('../../services/user')
const { navigateShellRoute } = require('../../utils/shell-nav')

const ROLE_ITEMS = [
  { key: 'player', label: '玩家', route: '/pages/game/player-manage/index' },
  { key: 'guide', label: '领路人', route: '/pages/game/referral-record/index' },
  { key: 'expert', label: '行家', route: '/pages/game/manage/index' }
]

Component({
  properties: {
    currentRole: { type: String, value: 'player' }
  },
  data: {
    roles: [ROLE_ITEMS[0]]
  },
  lifetimes: {
    attached() { this.loadRoles() }
  },
  methods: {
    async loadRoles() {
      try {
        const user = await userService.getCurrentUser()
        const roles = Array.isArray(user.roles) ? user.roles : []
        const statusMap = user.roleStatusMap || {}
        this.setData({
          roles: ROLE_ITEMS.filter((item) => item.key === 'player' || roles.includes(item.key) || statusMap[item.key] === 'active' || statusMap[item.key] === 'approved')
        })
      } catch (error) {
        this.setData({ roles: [ROLE_ITEMS[0]] })
      }
    },
    onRoleTap(event) {
      const role = this.data.roles.find((item) => item.key === event.currentTarget.dataset.key)
      if (!role || role.key === this.properties.currentRole) return
      navigateShellRoute(role.route)
    }
  }
})
