<script setup lang="ts">
type Product = {
  id: string
  name: string
  active: boolean
}

const api = useApi()
const toast = useToast()
const confirm = useConfirm()

const products = ref<Product[]>([])
const loading = ref(true)
const errorMessage = ref('')

async function loadProducts() {
  loading.value = true
  errorMessage.value = ''
  try {
    products.value = await api<Product[]>('/products')
  } catch (e) {
    const err = e as { data?: { error?: string } }
    errorMessage.value = err.data?.error ?? 'Não foi possível carregar os produtos'
  } finally {
    loading.value = false
  }
}

onMounted(loadProducts)

// --- cadastro ---
const dialogVisible = ref(false)
const newName = ref('')
const saving = ref(false)
const formError = ref('')

function openDialog() {
  newName.value = ''
  formError.value = ''
  dialogVisible.value = true
}

async function createProduct() {
  saving.value = true
  formError.value = ''
  try {
    const created = await api<Product>('/products', {
      method: 'POST',
      body: { name: newName.value },
    })
    products.value.push(created)
    dialogVisible.value = false
    toast.add({ severity: 'success', summary: 'Produto cadastrado', detail: created.name, life: 3000 })
  } catch (e) {
    const err = e as { data?: { error?: string } }
    formError.value = err.data?.error ?? 'Não foi possível salvar'
  } finally {
    saving.value = false
  }
}

// --- desativação ---
function askDeactivate(product: Product) {
  confirm.require({
    header: 'Desativar produto',
    message: `Desativar "${product.name}"? Ele deixa de aparecer no catálogo, mas o histórico é mantido.`,
    acceptProps: { label: 'Desativar', severity: 'danger' },
    rejectProps: { label: 'Cancelar', severity: 'secondary', outlined: true },
    accept: () => deactivate(product),
  })
}

async function deactivate(product: Product) {
  try {
    await api(`/products/${product.id}`, {
      method: 'PATCH',
      body: { active: false },
    })
    products.value = products.value.filter((p) => p.id !== product.id)
    toast.add({ severity: 'info', summary: 'Produto desativado', detail: product.name, life: 3000 })
  } catch (e) {
    const err = e as { data?: { error?: string } }
    toast.add({ severity: 'error', summary: 'Erro', detail: err.data?.error ?? 'Não foi possível desativar', life: 5000 })
  }
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-gray-800">Produtos</h1>
      <Button label="Novo produto" icon="pi pi-plus" @click="openDialog" />
    </div>

    <Message v-if="errorMessage" severity="error" class="mb-4">{{ errorMessage }}</Message>

    <DataTable :value="products" :loading="loading" striped-rows size="small">
      <Column field="name" header="Nome" sortable />
      <Column header="Ações" class="w-32">
        <template #body="{ data }">
          <Button label="Desativar" severity="danger" text size="small" @click="askDeactivate(data)" />
        </template>
      </Column>
    </DataTable>

    <Dialog v-model:visible="dialogVisible" header="Novo produto" modal class="w-96">
      <form class="flex flex-col gap-3" @submit.prevent="createProduct">
        <InputText v-model="newName" placeholder="Nome do produto" autofocus fluid />
        <Message v-if="formError" severity="error" size="small">{{ formError }}</Message>
        <div class="flex justify-end gap-2">
          <Button label="Cancelar" severity="secondary" text @click="dialogVisible = false" />
          <Button type="submit" label="Salvar" :loading="saving" />
        </div>
      </form>
    </Dialog>
  </div>
</template>
