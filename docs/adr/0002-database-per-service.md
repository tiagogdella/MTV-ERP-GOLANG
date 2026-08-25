# ADR-0002: Database-per-service

## Status
Aceito

## Contexto
Com múltiplos serviços independentes (ADR-0001), é preciso decidir se eles compartilham um banco de dados ou não. Banco compartilhado facilita joins e transações entre contexts, mas acopla os serviços no nível de esquema — mudar uma tabela de um serviço pode quebrar outro silenciosamente, e nenhum serviço é dono de verdade dos seus dados.

## Decisão
Cada serviço tem seu próprio banco PostgreSQL, sem acesso direto a bancos de outros serviços. Referências entre bounded contexts (ex: `purchasing-service` referenciando um fornecedor que "vive" em `catalog-service`) são feitas por ID lógico, nunca por foreign key física entre bancos.

## Consequências
- Sem transação ACID cruzando serviços — qualquer operação que precise de consistência entre dois contexts depende de chamada síncrona (gRPC) ou, mais adiante, de evento assíncrono + idempotência (ver ADR-0003).
- Cada serviço pode evoluir seu schema livremente sem quebrar outros serviços.
- Integridade referencial entre contexts (ex: "esse fornecedor_id realmente existe?") vira responsabilidade da aplicação, não do banco — validado na camada de serviço, não por constraint SQL.
- Custo operacional de manter múltiplos bancos pequenos, aceito dado o volume baixo de uso (RNF7).
