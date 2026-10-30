<script setup lang="ts">
const auth = useAuthStore()

async function logout() {
  auth.logout()
  await navigateTo('/login')
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-800">Bem-vindo</h1>
    <p class="mt-2 text-gray-600">Use o menu ao lado para navegar.</p>
  </div>
</template>
