<script setup lang="ts">
const auth = useAuthStore()

async function logout() {
  auth.logout()
  await navigateTo('/login')
}
</script>

<template>
  <div class="p-8">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-emerald-700">MTV ERP</h1>
      <Button label="Sair" severity="secondary" text @click="logout" />
    </div>
    <p class="mt-2 text-gray-600">Você está logado.</p>
  </div>
</template>
