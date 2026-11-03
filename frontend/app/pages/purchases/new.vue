<script setup lang="ts">
import { useForm, useFieldArray } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import { purchaseSchema } from '~/schemas/purchase'

type Supplier = { id: string; name: string; document: string; active: boolean }

const api = useApi()
const toast = useToast()

type Product = { id: string; name: string; active: boolean }
type Unit = { id: string; name: string; conversionFactorKg: string }

const products = ref<Product[]>([])
const activeProducts = computed(() => products.value.filter((p) => p.active))
const units = ref<Unit[]>([])

const suppliers = ref<Supplier[]>([])
const activeSuppliers = computed(() => suppliers.value.filter((s) => s.active))

onMounted(async () => {
  try {
    const [supplierList, productList, unitList] = await Promise.all([
      api<Supplier[]>('/suppliers'),
      api<Product[]>('/products'),
      api<Unit[]>('/units'),
    ])
    suppliers.value = supplierList
    products.value = productList
    units.value = unitList
  } catch {
    toast.add({ severity: 'error', summary: 'Erro', detail: 'Não foi possível carregar fornecedores, produtos ou unidades', life: 5000 })
  }
})

function emptyItem() {
  return { productId: '', unitId: '', harvest: '', quantity: '' }
}

const { defineField, errors, handleSubmit, isSubmitting, resetForm, submitCount } = useForm({
  validationSchema: toTypedSchema(purchaseSchema),
  initialValues: { items: [emptyItem()] },
})

function itemError(idx: number, field: string): string | undefined {
  if (submitCount.value === 0) return undefined
  return (errors.value as Record<string, string | undefined>)[`items[${idx}].${field}`]
}

const { fields: itemFields, push: addItem, remove: removeItem } = useFieldArray<ReturnType<typeof emptyItem>>('items')

const [supplierId] = defineField('supplierId')
const [invoiceNumber] = defineField('invoiceNumber')
const [invoiceDate] = defineField('invoiceDate')
const [invoiceValue] = defineField('invoiceValue')

const onSubmit = handleSubmit(async (values) => {
  try {
    await api('/purchases', { method: 'POST', body: values })
    toast.add({ severity: 'success', summary: 'Compra lançada', detail: `Nota ${values.invoiceNumber} registrada`, life: 4000 })
    resetForm()
  } catch (e) {
    const err = e as { data?: { error?: string } }
    toast.add({ severity: 'error', summary: 'Erro', detail: err.data?.error ?? 'Não foi possível lançar a compra', life: 6000 })
  }
})
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-800 mb-4">Nova compra</h1>

    <form class="flex flex-col gap-6 max-w-4xl" @submit.prevent="onSubmit">
      <section class="grid grid-cols-1 md:grid-cols-3 gap-4 p-4 bg-white rounded-lg border border-gray-200">
        <h2 class="md:col-span-3 font-semibold text-gray-800">Nota fiscal</h2>

        <div class="flex flex-col gap-1 md:col-span-3">
          <label for="supplier" class="text-sm font-medium text-gray-700">Fornecedor</label>
          <Select
            v-model="supplierId"
            input-id="supplier"
            :options="activeSuppliers"
            option-label="name"
            option-value="id"
            placeholder="Escolha o fornecedor"
            filter
            :invalid="!!errors.supplierId"
          />
          <small class="text-red-600">{{ errors.supplierId }}</small>
        </div>

        <div class="flex flex-col gap-1">
          <label for="invoiceNumber" class="text-sm font-medium text-gray-700">Número da nota</label>
          <InputText id="invoiceNumber" v-model="invoiceNumber" placeholder="Ex.: 000123" :invalid="!!errors.invoiceNumber" />
          <small class="text-red-600">{{ errors.invoiceNumber }}</small>
        </div>

        <div class="flex flex-col gap-1">
          <label for="invoiceDate" class="text-sm font-medium text-gray-700">Data da nota</label>
          <InputText id="invoiceDate" v-model="invoiceDate" type="date" :invalid="!!errors.invoiceDate" />
          <small class="text-red-600">{{ errors.invoiceDate }}</small>
        </div>

        <div class="flex flex-col gap-1">
          <label for="invoiceValue" class="text-sm font-medium text-gray-700">Valor da nota (R$)</label>
          <InputText id="invoiceValue" v-model="invoiceValue" inputmode="decimal" placeholder="Ex.: 1500,50" :invalid="!!errors.invoiceValue" />
          <small class="text-red-600">{{ errors.invoiceValue }}</small>
        </div>
      </section>

      <section class="flex flex-col gap-3 p-4 bg-white rounded-lg border border-gray-200">
        <div class="flex items-center justify-between">
          <h2 class="font-semibold text-gray-800">Itens da nota</h2>
          <Button label="Adicionar item" icon="pi pi-plus" size="small" severity="secondary" @click="addItem(emptyItem())" />
        </div>

        <div class="hidden md:grid md:grid-cols-[2fr_1fr_1fr_1fr_2.5rem] gap-2 text-sm font-medium text-gray-700">
          <span>Produto</span>
          <span>Unidade</span>
          <span>Safra</span>
          <span>Quantidade</span>
          <span />
        </div>

        <div
          v-for="(item, idx) in itemFields"
          :key="item.key"
          class="grid grid-cols-1 md:grid-cols-[2fr_1fr_1fr_1fr_2.5rem] gap-2 items-start"
        >
          <div class="flex flex-col gap-1">
            <Select v-model="item.value.productId" :options="activeProducts" option-label="name" option-value="id" placeholder="Escolha o produto" filter :invalid="!!itemError(idx, 'productId')" />
            <small class="text-red-600">{{ itemError(idx, 'productId') }}</small>
          </div>
          <div class="flex flex-col gap-1">
            <Select v-model="item.value.unitId" :options="units" option-label="name" option-value="id" placeholder="Unidade" :invalid="!!itemError(idx, 'unitId')" />
            <small class="text-red-600">{{ itemError(idx, 'unitId') }}</small>
          </div>
          <div class="flex flex-col gap-1">
            <InputText v-model="item.value.harvest" placeholder="Ex.: 11/2026" :invalid="!!itemError(idx, 'harvest')" />
            <small class="text-red-600">{{ itemError(idx, 'harvest') }}</small>
          </div>
          <div class="flex flex-col gap-1">
            <InputText v-model="item.value.quantity" inputmode="decimal" placeholder="Ex.: 10" :invalid="!!itemError(idx, 'quantity')" />
            <small class="text-red-600">{{ itemError(idx, 'quantity') }}</small>
          </div>
          <Button icon="pi pi-trash" severity="danger" text aria-label="Remover item" :disabled="itemFields.length === 1" @click="removeItem(idx)" />
        </div>

        <small v-if="errors.items" class="text-red-600">{{ errors.items }}</small>
      </section>

      <div class="flex justify-end">
        <Button type="submit" label="Lançar compra" icon="pi pi-check" :loading="isSubmitting" />
      </div>
    </form>
  </div>
</template>