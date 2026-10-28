export default defineNuxtRouteMiddleware((to) => {
  const auth = useAuthStore()

  if (!auth.isLoggedIn && to.path !== '/login') {
    return navigateTo('/login')
  }
  if (auth.isLoggedIn && to.path === '/login') {
    return navigateTo('/')
  }
})
