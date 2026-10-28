<script setup lang="ts">
const config = useRuntimeConfig()

const email = ref('')
const password = ref('')
const errorMessage = ref('')
const loading = ref(false)

async function submit() {
  errorMessage.value = ''
  loading.value = true

  try {
    const resp = await $fetch<{ token: string }>('/auth/login', {
      baseURL: config.public.apiBase,
      method: 'POST',
      body: { email: email.value, password: password.value },
    })
    console.log('token recebido:', resp.token.slice(0, 20) + '...')
  } catch (e) {
    const err = e as { data?: { error?: string } }
    errorMessage.value = err.data?.error ?? 'Não foi possível conectar ao servidor'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50">
    <form class="w-full max-w-sm bg-white p-8 rounded-xl shadow flex flex-col gap-4" @submit.prevent="submit">
      <div>
        <h1 class="text-2xl font-bold text-emerald-700">MTV ERP</h1>
        <p class="text-sm text-gray-600">Entre com seu e-mail e senha</p>
      </div>

      <InputText v-model="email" type="email" placeholder="E-mail" autocomplete="username" fluid />
      <Password v-model="password" placeholder="Senha" :feedback="false" toggle-mask autocomplete="current-password" fluid />

      <Message v-if="errorMessage" severity="error" size="small">{{ errorMessage }}</Message>
      <Button type="submit" label="Entrar" :loading="loading" />    
    </form>
  </div>
</template>
