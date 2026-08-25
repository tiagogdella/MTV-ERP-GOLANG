# ADR-0005: Unidade de comercialização e conversão para kg

## Status
Aceito

## Contexto
Toda unidade de comercialização (RF-CAT-6) tem um fator de conversão para kg. Ao longo do tempo pode ser necessário corrigir esse fator (ex: descobrir que "saco 25kg" na prática pesa 25,4kg) ou criar variações. Se o fator pudesse ser editado livremente numa unidade já usada em movimentações históricas, toda rastreabilidade de kg calculado no passado (ADR-0006) ficaria inconsistente com o fator atual.

## Decisão
O fator de conversão de uma unidade de comercialização é imutável depois de criado. Qualquer correção ou mudança de peso cria uma unidade de comercialização nova — a antiga permanece cadastrada (nunca é editada ou excluída) só para explicar o histórico que já existe.

## Consequências
- Toda movimentação de estoque antiga continua explicável exatamente como foi calculada no momento em que aconteceu, mesmo anos depois.
- O cadastro de unidades cresce ao longo do tempo com variações (ex: "saco 25kg", "saco 25kg (corrigido)") — aceito como custo necessário pela rastreabilidade.
- Reforça, no nível de unidade de medida, o mesmo princípio de "nunca editar o que já virou fato" que aparece em outras partes do domínio (ver ADR-0008).
