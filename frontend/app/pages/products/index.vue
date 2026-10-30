<script setup lang="ts">
type Product = {
  id: string
  name: string
  active: boolean
}

const api = useApi()
const products = ref<Product[]>([])
const loading = ref(true)
const errorMessage = ref('')

onMounted(async () => {
  try {
    products.value = await api<Product[]>('/products')
  } catch (e) {
    const err = e as { data?: { error?: string } }
    errorMessage.value = err.data?.error ?? 'Não foi possível carregar os produtos'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="p-8">
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-emerald-700">Produtos</h1>
      <NuxtLink to="/" class="text-emerald-700 underline">Voltar</NuxtLink>
    </div>

    <Message v-if="errorMessage" severity="error" class="mb-4">{{ errorMessage }}</Message>

    <DataTable :value="products" :loading="loading" striped-rows size="small">
      <Column field="name" header="Nome" sortable />
      <Column field="active" header="Ativo">
        <template #body="{ data }">
          <Tag :value="data.active ? 'Sim' : 'Não'" :severity="data.active ? 'success' : 'secondary'" />
        </template>
      </Column>
    </DataTable>
  </div>
</template>
