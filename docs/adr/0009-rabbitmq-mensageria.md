# ADR-0009: RabbitMQ para mensageria assíncrona (substitui ADR-0003)

## Status
Aceito — substitui ADR-0003

## Contexto
ADR-0003 escolheu NATS pelos motivos documentados lá (leveza operacional, encaixe com a stack Go, porte do projeto — RNF7). A decisão foi revisitada em 2026-08-27: RabbitMQ é a ferramenta de mensageria mais estabelecida do mercado, mais recorrente em vaga de emprego, e o valor de aprendizado/currículo de trabalhar com ela pesa mais, pro dono do projeto, do que a economia de complexidade operacional que o NATS trazia. Essa troca já tinha sido antecipada como possibilidade na própria ADR-0003 ("Alternativa considerada: RabbitMQ").

## Decisão
Usar RabbitMQ (via AMQP) como barramento de eventos para a comunicação assíncrona entre serviços, no lugar de NATS. NATS já havia sido deployado no cluster (Fase 2 do TODO) e será substituído por RabbitMQ na infraestrutura.

## Consequências
- Ganho de experiência com a ferramenta de mensageria mais recorrente no mercado — motivo explicitamente de aprendizado/currículo, não de necessidade técnica do projeto (é importante registrar isso, porque justifica abrir mão da simplicidade operacional do NATS conscientemente).
- Mais complexidade operacional que o NATS: broker Erlang, conceitos de exchange/queue/binding a configurar antes de publicar/consumir qualquer coisa.
- Modelo de entrega muda de "subject" simples (NATS) pra roteamento via exchange (direct/topic/fanout) — o desenho do evento `CompraLançada` (Fase 6 do TODO) precisa considerar esse modelo na hora de implementar.
- O estudo de "NATS Core vs JetStream" da Fase 6 original vira "filas duráveis + publisher confirms do RabbitMQ" — resolve o mesmo problema (garantir que a mensagem não se perde) por um caminho diferente.
- Manifests de deploy do NATS (`deploy/nats/`) saem de uso, substituídos por manifests equivalentes de RabbitMQ.
