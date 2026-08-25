# ADR-0004: Kg como unidade canônica de estoque

## Status
Aceito

## Contexto
Produtos são comercializados em várias unidades diferentes (fardo 30kg, saco 25kg, pacote 1kg, granel, etc — ver `docs/requisitos.md` RF-CAT-6). Se o estoque interno guardasse quantidade "em fardos" ou "em sacos" misturado, comparar ou somar saldo entre operações que usaram unidades diferentes vira ambíguo e propenso a erro.

## Decisão
O estoque interno é sempre representado em quilogramas (kg), independente da unidade de comercialização usada na operação de origem. A conversão para kg acontece só na borda — na entrada (compra) e na saída (venda) — nunca fica um saldo de estoque expresso em unidade de comercialização.

## Consequências
- Toda operação de entrada/saída de estoque precisa ter, obrigatoriamente, o fator de conversão da unidade usada disponível no momento (ver ADR-0005).
- Consultas agregadas de estoque (por produto, por lote) são sempre diretas, sem precisar normalizar unidades na hora da consulta.
- Unidades de comercialização passam a ser puramente metadata de exibição/negociação (RF-CAT-6), nunca a unidade de armazenamento real do saldo.
