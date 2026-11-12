import { z } from 'zod'
import { requiredText, positiveDecimal } from './rules'

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
