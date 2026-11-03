const kgFormatter = new Intl.NumberFormat('pt-BR', { maximumFractionDigits: 4 })

export function formatKg(value: string): string {
  return `${kgFormatter.format(Number(value))} kg`
}

export function formatDate(value: string): string {
  const [year, month, day] = value.slice(0, 10).split('-')
  return `${day}/${month}/${year}`
}
