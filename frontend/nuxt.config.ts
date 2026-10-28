import tailwindcss from '@tailwindcss/vite'
import Aura from '@primeuix/themes/aura'

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  runtimeConfig: {
    public: {
      apiBase: 'http://api.192.168.1.44.nip.io/api/v1',
    },
  },
  css: ['~/assets/css/main.css'],
  vite: {
    plugins: [tailwindcss()],
  },
  modules: ['@primevue/nuxt-module'],
  primevue: {
    options: {
      theme: {
        preset: Aura,
      },
    },
  },
})
