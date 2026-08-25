# ADR-0006: Rastreabilidade por lote é premissa obrigatória do modelo de dados

## Status
Aceito

## Contexto
O negócio da distribuidora exige rastrear a origem de cada quantidade de produto em estoque — de qual fornecedor veio, de qual safra, quando foi recebida — tanto por exigência prática (qualidade, disputa com fornecedor) quanto por ser um diferencial de controle que o sistema deve garantir de forma estrutural, não opcional.

## Decisão
Toda movimentação de estoque (entrada por compra, saída por venda, ajuste manual, troca) referencia obrigatoriamente um lote. Essa regra é validada como invariante de domínio na camada de serviço do `inventory-service` — não é só uma constraint `NOT NULL` de banco, é rejeitada explicitamente como erro de negócio (`FailedPrecondition` ou equivalente) se alguém tentar movimentar estoque sem lote associado.

## Consequências
- Nenhum fluxo do sistema pode criar ou mover estoque "solto" — compra, venda, ajuste e troca (RF-PUR-1, RF-VEN-2, RF-INV-4, RF-FIN-7 em `docs/requisitos.md`) todos precisam modelar a criação ou referência de um lote.
- Habilita a consulta de rastreabilidade completa por lote (RF-INV-5) — origem e todos os destinos de cada lote, de ponta a ponta.
- Testes de domínio precisam cobrir explicitamente o caso de erro "movimentação sem lote" (ver Fase 4 do TODO).
