import { defineStore } from 'pinia'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('decoy-club.token') ?? '',
    user: (() => {
      const raw = localStorage.getItem('decoy-club.user')
      return raw ? (JSON.parse(raw) as Record<string, unknown>) : null
    })(),
  }),
  actions: {
    setSession(token: string, user: Record<string, unknown>) {
      this.token = token
      this.user = user
      localStorage.setItem('decoy-club.token', token)
      localStorage.setItem('decoy-club.user', JSON.stringify(user))
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem('decoy-club.token')
      localStorage.removeItem('decoy-club.user')
    },
  },
})
