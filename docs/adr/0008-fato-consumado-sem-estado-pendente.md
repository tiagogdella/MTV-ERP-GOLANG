# ADR-0008: Operações modelam fato consumado, não estado pendente, por padrão

## Status
Aceito

## Contexto
Durante a definição dos fluxos de compra e venda (`docs/requisitos.md`, `docs/fluxos-usabilidade.md`), surgiu repetidamente a mesma pergunta de design: uma operação deveria existir como um estado "pendente" no sistema (ex: um pedido aguardando confirmação) ou deveria ser lançada diretamente como um fato já consumado?

A resposta não foi a mesma nos dois casos, e a diferença revelou o critério certo:
- **Compra** (RF-PUR-1): não existe "pedido de compra" pendente — a nota fiscal sempre acompanha a mercadoria, então não há necessidade real de negócio de rastrear uma encomenda antes dela chegar. Lançar num passo só é suficiente.
- **Venda** (RF-VEN-1/RF-VEN-2): existe "pedido de venda" como etapa própria, porque vendedores de fato registram o que o cliente encomendou antes do faturamento acontecer (uso real: abandonar planilhas, gerar projeção de caixa) — e pedidos são alterados/cancelados com frequência antes de virar venda.

O critério que decide não é "essa operação parece com outra que já resolvemos", é **se o estado intermediário existe de verdade na operação, independente do sistema**.

## Decisão
Por padrão, toda operação nova do sistema é modelada como fato consumado, lançada num único passo. Um estado pendente/intermediário só é introduzido quando há uma necessidade de negócio comprovada para ele existir — nunca por semelhança superficial com outro fluxo, nem por "pode ser útil algum dia".

## Consequências
- Reduz a quantidade de máquinas de estado e telas de "confirmar depois" que o sistema precisa suportar — a maior parte das operações (compra, cadastermos, ajuste de estoque) são ações diretas.
- Toda vez que uma nova funcionalidade for modelada, a pergunta de design correta é "isso existe fora do sistema, com ou sem ele?", não "outros ERPs fazem isso em duas etapas?".
- Onde o estado pendente é mesmo necessário (pedido de venda), ele precisa ser bem suportado — edição, cancelamento e projeção de caixa (RF-VEN-1, RF-VEN-4) não podem ser tratados como afterthought.
- Esse princípio também apareceu, num nível mais local, na decisão de fator de conversão de unidade imutável (ADR-0005): a resposta pra "isso deveria poder mudar depois?" segue o mesmo raciocínio.
