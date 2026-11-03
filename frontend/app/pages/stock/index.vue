<script setup lang="ts">
type Product = { id: string; name: string; active: boolean }
type Lot = {
  id: string
  productId: string
  purchaseItemId: string
  harvest: string
  quantityKg: string
  receivedAt: string
}
type LotStock = Lot & { balanceKg: string }
type ProductStock = { productId: string; totalKg: string; lots: LotStock[] }
type StockMovement = { id: string; type: string; quantityKg: string; occurredAt: string; origin: string }
type LotDetails = Lot & { balanceKg: string; movements: StockMovement[] }

const api = useApi()
const toast = useToast()

const products = ref<Product[]>([])
const selectedProductId = ref<string | null>(null)
const stock = ref<ProductStock | null>(null)
const loadingStock = ref(false)

onMounted(async () => {
  try {
    products.value = await api<Product[]>('/products')
  } catch {
    toast.add({ severity: 'error', summary: 'Erro', detail: 'Não foi possível carregar os produtos', life: 5000 })
  }
})

watch(selectedProductId, async (productId) => {
  stock.value = null
  if (!productId) return

  loadingStock.value = true
  try {
    stock.value = await api<ProductStock>(`/products/${productId}/stock`)
  } catch (e) {
    const err = e as { data?: { error?: string } }
    toast.add({ severity: 'error', summary: 'Erro', detail: err.data?.error ?? 'Não foi possível carregar o estoque', life: 5000 })
  } finally {
    loadingStock.value = false
  }
})

const drawerVisible = ref(false)
const lotDetails = ref<LotDetails | null>(null)

async function openLot(event: { data: LotStock }) {
  lotDetails.value = null
  drawerVisible.value = true
  try {
    lotDetails.value = await api<LotDetails>(`/lots/${event.data.id}`)
  } catch (e) {
    const err = e as { data?: { error?: string } }
    toast.add({ severity: 'error', summary: 'Erro', detail: err.data?.error ?? 'Não foi possível carregar o lote', life: 5000 })
  }
}

function movementSeverity(type: string): string {
  if (type === 'entrada') return 'success'
  if (type === 'saida') return 'danger'
  return 'warn'
}
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-800 mb-4">Estoque</h1>

    <Select
      v-model="selectedProductId"
      :options="products"
      option-label="name"
      option-value="id"
      placeholder="Escolha um produto"
      filter
      show-clear
      class="w-full md:w-96 mb-6"
    />

    <div v-if="stock" class="mb-4 p-4 bg-white rounded-lg border border-gray-200 inline-block">
      <span class="text-sm text-gray-600">Total em estoque</span>
      <p class="text-2xl font-bold text-emerald-700">{{ formatKg(stock.totalKg) }}</p>
    </div>

    <DataTable
      v-if="selectedProductId"
      :value="stock?.lots ?? []"
      :loading="loadingStock"
      striped-rows
      size="small"
      :row-class="() => 'cursor-pointer'"
      @row-click="openLot"
    >
      <template #empty>Nenhum lote para este produto.</template>
      <Column field="harvest" header="Safra" sortable />
      <Column field="receivedAt" header="Recebido em" sortable>
        <template #body="{ data }">{{ formatDate(data.receivedAt) }}</template>
      </Column>
      <Column header="Entrada">
        <template #body="{ data }">{{ formatKg(data.quantityKg) }}</template>
      </Column>
      <Column header="Saldo">
        <template #body="{ data }"><strong>{{ formatKg(data.balanceKg) }}</strong></template>
      </Column>
    </DataTable>

    <Drawer v-model:visible="drawerVisible" header="Rastreabilidade do lote" position="right" class="w-full md:w-[32rem]">
      <div v-if="lotDetails" class="flex flex-col gap-4">
        <div class="grid grid-cols-2 gap-2 text-sm">
          <span class="text-gray-500">Safra</span><span>{{ lotDetails.harvest }}</span>
          <span class="text-gray-500">Recebido em</span><span>{{ formatDate(lotDetails.receivedAt) }}</span>
          <span class="text-gray-500">Entrada</span><span>{{ formatKg(lotDetails.quantityKg) }}</span>
          <span class="text-gray-500">Saldo</span><strong>{{ formatKg(lotDetails.balanceKg) }}</strong>
        </div>

        <h3 class="font-semibold text-gray-800">Movimentações</h3>
        <DataTable :value="lotDetails.movements" size="small">
          <Column header="Data">
            <template #body="{ data }">{{ formatDate(data.occurredAt) }}</template>
          </Column>
          <Column header="Tipo">
            <template #body="{ data }"><Tag :value="data.type" :severity="movementSeverity(data.type)" /></template>
          </Column>
          <Column header="Quantidade">
            <template #body="{ data }">{{ formatKg(data.quantityKg) }}</template>
          </Column>
          <Column field="origin" header="Origem" />
        </DataTable>
      </div>
      <ProgressSpinner v-else />
    </Drawer>
  </div>
</template>
