<script setup lang="ts">
const auth = useAuthStore()

const links = [
  { to: '/', label: 'Início' },
  { to: '/products', label: 'Produtos' },
]

async function logout() {
  auth.logout()
  await navigateTo('/login')
}
</script>

<template>
  <div class="min-h-screen flex bg-gray-50">
    <aside class="w-56 bg-white border-r border-gray-200 flex flex-col">
      <div class="p-4 border-b border-gray-200">
        <span class="text-lg font-bold text-emerald-700">MTV ERP</span>
      </div>

      <nav class="flex-1 p-2 flex flex-col gap-1">
        <NuxtLink
          v-for="link in links"
          :key="link.to"
          :to="link.to"
          class="px-3 py-2 rounded-md text-gray-700 hover:bg-gray-100"
          active-class="bg-emerald-50 text-emerald-700 font-medium"
        >
          {{ link.label }}
        </NuxtLink>
      </nav>

      <div class="p-4 border-t border-gray-200 flex items-center justify-between">
        <span class="text-sm text-gray-600">{{ auth.role }}</span>
        <Button label="Sair" severity="secondary" text size="small" @click="logout" />
      </div>
    </aside>

    <main class="flex-1 p-8">
      <slot />
    </main>
  </div>
</template>
