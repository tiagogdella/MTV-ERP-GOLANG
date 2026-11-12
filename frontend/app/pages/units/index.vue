<script setup lang="ts">
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import { unitSchema } from '~/schemas/unit'

type Unit = { id: string; name: string; conversionFactorKg: string }

const api = useApi()
const toast = useToast()

const units = ref<Unit[]>([])
const loading = ref(true)
const errorMessage = ref('')

async function loadUnits() {
  loading.value = true
  errorMessage.value = ''
  try {
    units.value = await api<Unit[]>('/units')
  } catch (e) {
    const err = e as { data?: { error?: string } }
    errorMessage.value = err.data?.error ?? 'Não foi possível carregar as unidades'
  } finally {
    loading.value = false
  }
}

onMounted(loadUnits)

// --- cadastro ---
const dialogVisible = ref(false)
const formError = ref('')

const { defineField, errors, handleSubmit, isSubmitting, resetForm } = useForm({
  validationSchema: toTypedSchema(unitSchema),
})

const [name] = defineField('name')
const [conversionFactorKg] = defineField('conversionFactorKg')

function openDialog() {
  resetForm()
  formError.value = ''
  dialogVisible.value = true
}

const createUnit = handleSubmit(async (values) => {
  formError.value = ''
  try {
    const created = await api<Unit>('/units', { method: 'POST', body: values })
    units.value.push(created)
    dialogVisible.value = false
    toast.add({ severity: 'success', summary: 'Unidade cadastrada', detail: created.name, life: 3000 })
  } catch (e) {
    const err = e as { data?: { error?: string } }
    formError.value = err.data?.error ?? 'Não foi possível salvar'
  }
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold text-gray-800">Unidades de medida</h1>
      <Button label="Nova unidade" icon="pi pi-plus" @click="openDialog" />
    </div>

    <Message v-if="errorMessage" severity="error" class="mb-4">{{ errorMessage }}</Message>

    <DataTable :value="units" :loading="loading" striped-rows size="small">
      <template #empty>Nenhuma unidade cadastrada.</template>
      <Column field="name" header="Nome" sortable />
      <Column header="Equivale a">
        <template #body="{ data }">1 {{ data.name }} = <strong>{{ formatKg(data.conversionFactorKg) }}</strong></template>
      </Column>
    </DataTable>

    <Dialog v-model:visible="dialogVisible" header="Nova unidade" modal class="w-96">
      <form class="flex flex-col gap-3" @submit.prevent="createUnit">
        <div class="flex flex-col gap-1">
          <label for="unitName" class="text-sm font-medium text-gray-700">Nome</label>
          <InputText id="unitName" v-model="name" placeholder="Ex.: Saco 50 kg" autofocus fluid :invalid="!!errors.name" />
          <small class="text-red-600">{{ errors.name }}</small>
        </div>
        <div class="flex flex-col gap-1">
          <label for="unitFactor" class="text-sm font-medium text-gray-700">Quantos kg tem 1 unidade?</label>
          <InputText id="unitFactor" v-model="conversionFactorKg" inputmode="decimal" placeholder="Ex.: 50" fluid :invalid="!!errors.conversionFactorKg" />
          <small class="text-red-600">{{ errors.conversionFactorKg }}</small>
        </div>
        <Message v-if="formError" severity="error" size="small">{{ formError }}</Message>
        <div class="flex justify-end gap-2">
          <Button label="Cancelar" severity="secondary" text @click="dialogVisible = false" />
          <Button type="submit" label="Salvar" :loading="isSubmitting" />
        </div>
      </form>
    </Dialog>
  </div>
</template>
