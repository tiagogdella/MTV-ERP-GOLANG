import { z } from 'zod'

export function requiredText(message: string) {
  return z.string({ required_error: message, invalid_type_error: message }).trim().min(1, message)
}

export function positiveDecimal(field: string, maxDecimals: number) {
  const format = new RegExp(`^\\d+(\\.\\d{1,${maxDecimals}})?$`)
  return requiredText(`${field} é obrigatório`)
    .transform((value) => value.replace(',', '.'))
    .refine((value) => format.test(value), `${field} deve ser um número com até ${maxDecimals} casas decimais, sem ponto de milhar`)
    .refine((value) => Number(value) > 0, `${field} deve ser maior que zero`)
}
