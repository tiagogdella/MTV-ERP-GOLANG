# ADR-0003: NATS para mensageria assíncrona

## Status
Aceito, revisitável (ver nota sobre RabbitMQ abaixo)

## Contexto
A comunicação entre `purchasing-service` e `inventory-service` (criar lote a partir de uma compra lançada) começa síncrona via gRPC (Fase 4 do TODO), mas isso acopla a disponibilidade dos dois serviços — se `inventory-service` estiver fora do ar, `purchasing-service` não consegue lançar compra nenhuma. O plano (Fase 6) é migrar essa comunicação para um evento assíncrono.

## Decisão
Usar NATS como barramento de eventos para a comunicação assíncrona entre serviços, começando com NATS Core (pub/sub simples) e migrando pra JetStream (persistência, replay) apenas se a necessidade de garantir entrega mesmo com consumidor fora do ar se confirmar na prática.

## Consequências
- Desacopla a disponibilidade de `purchasing-service` da de `inventory-service`.
- Exige lidar com entrega at-least-once (RNF2 em `docs/requisitos.md`) — consumidores precisam ser idempotentes (mesmo evento processado duas vezes não pode duplicar lote).
- NATS é operacionalmente mais simples que alternativas como Kafka, adequado ao porte do projeto (RNF7).
- Schema do evento (`CompraLançada`) precisa de estratégia de versionamento pra evoluir sem quebrar o consumidor (documentado quando a Fase 6 for implementada).

## Alternativa considerada: RabbitMQ
RabbitMQ é a opção mais estabelecida do mercado (mais tutorial, mais recorrente em vaga de emprego, roteamento mais expressivo via exchanges). Não foi escolhido agora porque NATS encaixa melhor na stack Go já adotada (ADR-0001) e no porte do projeto (RNF7) — mas a escala e a expressividade de roteamento do RabbitMQ podem justificar reabrir essa decisão mais adiante, se o volume de eventos ou a complexidade de roteamento crescerem além do que NATS resolve confortavelmente. Discutido em 2026-08-25, decisão de manter NATS por ora, revisitar quando/se a necessidade aparecer.
