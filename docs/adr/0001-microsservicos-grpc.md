# ADR-0001: Microsserviços com gRPC

## Status
Aceito

## Contexto
O projeto é um exercício de aprendizado com objetivo de ganhar experiência real com arquitetura de microsserviços em Go. O domínio (ERP para distribuidora de arroz) já tem fronteiras de negócio naturais — autenticação, catálogo de produtos, compras, estoque — que mapeiam bem para bounded contexts independentes (ver `docs/contexto-c4.md`).

A comunicação entre esses serviços é majoritariamente interna (serviço chamando serviço), não exposta a clientes externos diretamente.

## Decisão
Cada bounded context vira um serviço Go independente, com contratos definidos via Protocol Buffers e comunicação síncrona via gRPC entre eles. REST só existe na borda, no `api-gateway` (BFF), que traduz requisições REST externas em chamadas gRPC internas — nenhum serviço de domínio expõe REST diretamente.

## Consequências
- Ganho de experiência real com o padrão mais usado em sistemas distribuídos em Go, incluindo contratos tipados (protobuf) e evolução de schema.
- Contratos gRPC exigem disciplina de versionamento (buf, breaking-change detection) — custo extra comparado a REST solto.
- Mais complexidade operacional (deploy, observabilidade, service discovery) do que um monólito — aceito como parte do objetivo de aprendizado, mitigado por manter o número de usuários/carga baixo (RNF7 em `docs/requisitos.md`).
- Precisa de um `api-gateway` como Backend For Frontend, sem lógica de negócio própria (só tradução REST → gRPC e autenticação).
