export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const isLoggedIn = computed(() => token.value !== null)

  const role = computed(() => {
    if (!token.value) return null
    try {
      const payload = token.value.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')
      return JSON.parse(atob(payload)).role as string
    } catch {
      return null
    }
  })

  function setToken(newToken: string) {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  function logout() {
    token.value = null
    localStorage.removeItem('token')
  }

  return { token, isLoggedIn, role, setToken, logout }
})
