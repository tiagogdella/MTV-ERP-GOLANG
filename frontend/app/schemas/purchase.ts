import { z } from 'zod'

function requiredText(message: string) {
  return z.string({ required_error: message, invalid_type_error: message }).trim().min(1, message)
}

function positiveDecimal(field: string, maxDecimals: number) {
  const format = new RegExp(`^\\d+(\\.\\d{1,${maxDecimals}})?$`)
  return requiredText(`${field} é obrigatório`)
    .transform((value) => value.replace(',', '.'))
    .refine((value) => format.test(value), `${field} deve ser um número com até ${maxDecimals} casas decimais, sem ponto de milhar`)
    .refine((value) => Number(value) > 0, `${field} deve ser maior que zero`)
}

const purchaseItemSchema = z.object({
  productId: requiredText('Escolha o produto'),
  unitId: requiredText('Escolha a unidade'),
  harvest: requiredText('Informe a safra')
    .refine((value) => /^(0[1-9]|1[0-2])\/\d{4}$/.test(value), 'A safra deve estar no formato MM/AAAA (ex.: 11/2026)'),  
  quantity: positiveDecimal('Quantidade', 4),
})

export const purchaseSchema = z.object({
  supplierId: requiredText('Escolha o fornecedor'),
  invoiceNumber: requiredText('Informe o número da nota fiscal')
    .refine((value) => /^\d{1,9}$/.test(value), 'O número da nota deve ter só dígitos (até 9)'),
  invoiceDate: z.string({ required_error: 'Informe a data da nota' }).regex(/^\d{4}-\d{2}-\d{2}$/, 'Informe a data da nota'),
  invoiceValue: positiveDecimal('Valor da nota', 2),
  items: z.array(purchaseItemSchema).min(1, 'Adicione pelo menos um item'),
})

export type PurchaseInput = z.infer<typeof purchaseSchema>
