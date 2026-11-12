import { z } from 'zod'
import { requiredText, positiveDecimal } from './rules'

export const unitSchema = z.object({
  name: requiredText('Informe o nome da unidade'),
  conversionFactorKg: positiveDecimal('Fator de conversão', 4)
    .refine((value) => Number(value) < 100_000_000, 'Fator de conversão deve ser menor que 100.000.000'),
})

export type UnitInput = z.infer<typeof unitSchema>
