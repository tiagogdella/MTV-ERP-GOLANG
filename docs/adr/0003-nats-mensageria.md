# ADR-0003: NATS para mensageria assíncrona

## Status
Aceito

## Contexto
A comunicação entre `purchasing-service` e `inventory-service` (criar lote a partir de uma compra lançada) começa síncrona via gRPC (Fase 4 do TODO), mas isso acopla a disponibilidade dos dois serviços — se `inventory-service` estiver fora do ar, `purchasing-service` não consegue lançar compra nenhuma. O plano (Fase 6) é migrar essa comunicação para um evento assíncrono.

## Decisão
Usar NATS como barramento de eventos para a comunicação assíncrona entre serviços, começando com NATS Core (pub/sub simples) e migrando pra JetStream (persistência, replay) apenas se a necessidade de garantir entrega mesmo com consumidor fora do ar se confirmar na prática.

## Consequências
- Desacopla a disponibilidade de `purchasing-service` da de `inventory-service`.
- Exige lidar com entrega at-least-once (RNF2 em `docs/requisitos.md`) — consumidores precisam ser idempotentes (mesmo evento processado duas vezes não pode duplicar lote).
- NATS é operacionalmente mais simples que alternativas como Kafka, adequado ao porte do projeto (RNF7).
- Schema do evento (`CompraLançada`) precisa de estratégia de versionamento pra evoluir sem quebrar o consumidor (documentado quando a Fase 6 for implementada).
