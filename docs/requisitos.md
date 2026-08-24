# Requisitos — ERP Arroz (Migra Alimentos)

> Rascunho gerado a partir do event storming inicial, refinado com as decisões de negócio abaixo. Itens ainda com `⚠️ REVISAR` seguem em aberto — o resto já reflete decisão tomada.

Convenção: `RF-<CONTEXTO>-N` agrupado por bounded context, cada um com suas `RN` (regras de negócio). RNF ao final, sem agrupamento por contexto (são transversais).

---

## Decisão: escopo de Vendas/Financeiro/Fiscal

**Fica como Pós-MVP** — não entra nos bounded contexts iniciais (`auth, catalog, inventory, purchasing, gateway`), mas continua documentado aqui (seção `RF-VEN` / `RF-FIN` abaixo) em vez de removido, pra não perder o raciocínio já feito quando chegar a hora de implementar.

Ainda em aberto: `Fornecedor` e `Cliente` ficam em `catalog` (junto com Produto, como dado de referência/cadastro) — assumi isso abaixo por não termos indicação contrária, mas revisita se fizer mais sentido em `purchasing`/`sales` quando esses contextos existirem de fato.

---

## RF-AUTH — Autenticação e usuários

**RF-AUTH-1: Cadastrar usuário**
- RN1: Papéis (roles) do sistema: `admin` e `operador`
- RN2: Somente `admin` pode cadastrar usuário do sistema
- RN3: "Vendedor" não é um papel (role) novo — todo mundo no sistema é `operador` (exceto `admin`); vendedor é só uma função de negócio exercida por um `operador`, sem distinção de permissão no sistema

**RF-AUTH-2: Autenticar usuário (login)**
- RN1: Login inválido (credenciais erradas) não deve informar se foi o e-mail ou a senha que errou (evita enumeração de usuários)
- RN2: Autenticação bem-sucedida gera um token (JWT) com tempo de expiração

**RF-AUTH-3: Validar token de sessão**
- RN1: Token expirado ou inválido bloqueia a chamada no gateway antes de chegar ao serviço de destino

---

## RF-CAT — Catálogo (produto, fornecedor, cliente)

**RF-CAT-1: Cadastrar produto**
- RN1: Produto deve ter ao menos uma unidade de comercialização associada, cada uma com fator de conversão para kg (ver ADR-0004/0005 do TODO)
- RN2: Cadastro de produto é liberado — não restrito a `admin`

**RF-CAT-2: Inativar produto**
- RN1: Produto nunca é excluído, apenas inativado (soft delete)
- RN2: Produto inativo não aparece na consulta de catálogo (RF-CAT-3)

**RF-CAT-3: Consultar catálogo de produtos**
- RN1: Somente produtos ativos aparecem na consulta

**RF-CAT-4: Cadastrar fornecedor**
- RN1: Cadastro liberado — não restrito a `admin`
- RN2: Pessoa jurídica: CNPJ e Inscrição Estadual obrigatórios, além de endereço
- RN3: Pessoa física: CPF obrigatório, além de endereço
- RN4: E-mail aceito em ambos os casos, mas não obrigatório

**RF-CAT-5: Cadastrar cliente**
- RN1: Cadastro liberado — não restrito a `admin`
- RN2: Mesmas regras de documento de RF-CAT-4 (CNPJ+IE para PJ, CPF para PF, endereço obrigatório, e-mail opcional)

**RF-CAT-6: Cadastrar unidade de comercialização** 🆕
*(faltava um RF próprio pra isso — o TODO já trata como modelagem separada: fardo 30kg, saco 25kg, granel etc.)*
- RN1: Toda unidade de comercialização tem um fator de conversão para kg
- RN2: Fator de conversão é travado — nunca é editado depois de criado. Se precisar corrigir ou mudar o peso de uma unidade, cria-se uma unidade nova (ex: "saco 25kg" errado vira "saco 25,4kg" como cadastro à parte); a unidade antiga continua existindo só pra explicar o que já foi lançado com ela, preservando a rastreabilidade do histórico (RF-INV-5)

**RF-CAT-7: Editar cadastro de produto, fornecedor ou cliente** 🆕
*(só havia cadastro/inativação, faltava edição)*
- RN1: Edição não altera documentos (compras, vendas, notas) já emitidos anteriormente com os dados antigos

---

## RF-PUR — Compras (purchasing)

> **Decisão (2026-08-24):** sem conceito de "pedido de compra" como entidade separada. Como a nota fiscal sempre acompanha a mercadoria na chegada e não há necessidade de rastrear encomenda pendente, a compra é lançada num único passo — já é fato consumado (fornecedor + mercadoria + nota fiscal + lote, tudo de uma vez). Isso também simplifica o event storming original: os eventos `PedidoDeCompraCriado`, `MercadoriaRecebida` e `NotaFiscalDeEntradaLançada` viram um só (`CompraLançada` ou equivalente). 💡 candidato a ADR, junto com a de baixa de estoque na emissão da nota de venda (RF-INV-2/RN1) — as duas seguem o mesmo princípio: só existe o que já é fato, nada fica "pendente" no sistema.

**RF-PUR-1: Lançar compra**
- RN1: Lançamento é feito num único passo: fornecedor + um ou mais produtos (cada um com quantidade/unidade) + nota fiscal de entrada, tudo registrado de uma vez
- RN2: Compra deve referenciar um fornecedor cadastrado
- RN3: Todo lançamento gera um lote por produto (código, safra, fornecedor de origem, data) — premissa central do domínio (ver TODO, "não existe estoque sem lote")
- RN4: Quantidade lançada é convertida para kg conforme o fator de conversão da unidade usada, por produto
- RN5: Se a mesma compra chegar em mais de uma entrega (ex: dois caminhões), cada entrega é lançada separadamente — cada lançamento gera seu(s) próprio(s) lote(s), sem vínculo de "pedido" em comum
- RN6: Combinação fornecedor + número da nota fiscal não pode se repetir — bloqueia lançar a mesma nota fiscal duas vezes por engano

---

## RF-INV — Estoque (inventory)

**RF-INV-1: Atualizar estoque a partir de um recebimento**
- RN1: Toda movimentação de estoque referencia obrigatoriamente um lote
- RN2: Estoque interno é sempre representado em kg, independente da unidade usada na operação de origem

**RF-INV-2: Baixar estoque a partir de uma venda**
- RN1: Baixa de estoque ocorre na emissão da nota fiscal de venda (decisão de simplificação — não existe etapa separada de expedição no MVP) — 💡 candidato a ADR, já que afeta diretamente RF-FIN-6 (cancelamento de nota já com estoque baixado)
- RN2: Venda é bloqueada se resultar em saldo negativo de um lote (regra rígida, não é só alerta)

**RF-INV-3: Consultar estoque por lote**
- RN1: Consulta deve mostrar quantidade disponível em kg por lote e por produto

**RF-INV-4: Ajustar estoque manualmente** 🆕
*(perda, quebra, contagem de inventário físico divergindo do sistema — sem isso o estoque trava sempre que a realidade não bate com o registrado)*
- RN1: Todo ajuste manual exige justificativa registrada
- RN2: Ajuste referencia o lote afetado, como qualquer outra movimentação (nunca mexe em estoque "sem lote")

**RF-INV-5: Consultar rastreabilidade completa de um lote** 🆕
*(é a premissa central do domínio segundo o próprio TODO — "rastreabilidade por lote" merecia um RF explícito, não só uma regra dentro de outro)*
- RN1: Deve mostrar origem do lote (fornecedor, compra de origem, data de recebimento) e todos os destinos (vendas, trocas, ajustes) associados a ele

---

## Pós-MVP (documentado agora, implementação futura)

### RF-VEN — Vendas

> **Atualização de decisão (2026-08-24):** voltamos atrás na fusão "pedido + faturamento" — só para venda, não para compra (confirmado: compra segue em RF-PUR-1, lançamento único). Motivo: aqui existe uma necessidade de negócio real por trás do pedido — vendedores cadastrando encomenda por cliente pra abandonar as planilhas, projeção de contas a receber a partir do que foi encomendado, pedidos que mudam ou são cancelados com frequência antes de virar venda de fato. Isso não invalida o princípio usado em RF-PUR — só confirma que o crivo certo é "esse estado intermediário existe na operação real, independente do sistema?", não "parece com pedido de compra". Aqui a resposta é sim; em compra, não.

**RF-VEN-1: Cadastrar pedido de venda**
- RN1: Pedido é cadastrado por um vendedor, associado a um cliente (comprador)
- RN2: Pedido pode ser alterado enquanto não faturado
- RN3: Pedido pode ser cancelado enquanto não faturado — depois de faturado (vendido), não é mais possível cancelar o pedido; a única forma de desfazer nesse ponto é cancelar a nota fiscal (RF-FIN-6, com suas próprias restrições: só até 24h e sem baixa vinculada)
- RN4: "Excluir pedido antigo" é cancelamento (RN3), não exclusão de fato — consistente com o resto do sistema, que não apaga nada, só inativa/cancela (ver RF-CAT-2). Pedido cancelado fica no histórico, mas para de contar na projeção de caixa (RF-VEN-4/RN2)
- RN5: Preço é definido manualmente a cada pedido, não é um valor fixo do produto — varia conforme a negociação de cada venda (não existe cadastro de preço no catálogo — ver RF-VEN-3, removido)

**RF-VEN-2: Faturar pedido de venda**
- RN1: Faturamento parte de um pedido de venda — existente (criado antes pelo vendedor) ou criado ali mesmo, na própria tela de faturar, pra venda direta/imediata que nunca passou por cadastro em etapa separada
- RN2: Quantidade e preço faturados podem divergir dos registrados no pedido original — o pedido é estimativa, não compromisso rígido
- RN3: Disponibilidade de estoque é verificada sobre a quantidade efetivamente faturada (não a do pedido) — venda continua bloqueada se resultar em saldo negativo (RF-INV-2, regra mantida como rígida)
- RN4: Boleto (quando a forma de pagamento exigir) é emitido no mesmo momento do faturamento — nunca como etapa separada (ver RF-FIN-5, removido)
- RN5: Faturar um pedido o fecha por completo, mesmo que a quantidade faturada seja diferente da quantidade original do pedido — não existe estado de "pedido parcialmente atendido" aguardando complemento. Ex: pedido de 30.000kg de parboilizado a granel, o caminhão carrega 29.000kg, operador fatura os 29.000kg e o pedido fecha ali — os 1.000kg que faltaram não ficam pendentes em lugar nenhum
- RN6: Qualquer operador pode faturar um pedido, inclusive o próprio vendedor que o criou — sem restrição adicional de permissão (empresa pequena, liberdade total entre operadores)

**RF-VEN-3: ~~Definir preço de venda de um produto~~** — REMOVIDO
- Não existe preço fixo/cadastrado por produto: preço é sempre manual, decidido pedido a pedido (RF-VEN-1/RN5)

**RF-VEN-4: Consultar projeção de caixa a partir de pedidos de venda em aberto** 🆕
*(era o motivo de negócio por trás de trazer o pedido de volta — sem um RF próprio pra isso, o dado fica sem uso)*
- RN1: Projeção soma o valor dos pedidos de venda ainda não faturados, por cliente e agregado (total geral)
- RN2: Pedidos cancelados não entram na projeção

### RF-FIN — Financeiro e fiscal

**RF-FIN-1: Gerar título a receber**
- RN1: Título a receber é gerado a partir da emissão de uma nota de venda

**RF-FIN-2: Registrar pagamento**
- RN1: Pagamento deve ser vinculado a um título em aberto
- RN2: Pagamento parcial é permitido

**RF-FIN-3: Baixar título em aberto**
- RN1: Título é baixado quando a soma dos pagamentos registrados cobre o valor total do título

**RF-FIN-4: Atualizar fluxo de caixa**
- RN1: Toda baixa de título (receber ou pagar) reflete no fluxo de caixa

**RF-FIN-5: ~~Emitir boleto para nota fiscal~~** — REMOVIDO
- Mesclado em RF-VEN-2/RN4: boleto nasce junto do faturamento, não é mais etapa financeira separada

**RF-FIN-6: Cancelar nota fiscal**
- RN1: Cancelamento só é permitido se dentro de 24h da emissão
- RN2: Cancelamento só é permitido se não houver baixa (pagamento) vinculada ao título

**RF-FIN-7: Registrar troca de mercadoria**
- RN1: Troca gera duas movimentações de estoque: entrada (devolução da mercadoria trocada) e saída (nova mercadoria entregue)

**RF-FIN-8: Alterar forma de pagamento de um título**
- RN1: Financeiro pode alterar a forma de pagamento de um título em aberto (ex: era boleto, cliente pagou via PIX)

**RF-FIN-9: Gerar título a pagar** 🆕
*(espelho do lado compra — o fluxo financeiro só cobria venda; sem isso não existe contas a pagar)*
- RN1: Título a pagar é gerado a partir do lançamento da compra (RF-PUR-1)

**RF-FIN-10: Registrar pagamento a fornecedor / baixar título a pagar** 🆕
- RN1: Mesmas regras de RF-FIN-2/RF-FIN-3 (pagamento parcial permitido, baixa quando soma cobre o total)

---

## RNF — Requisitos não funcionais

- RNF1: RPCs internos devem responder em até ~500ms sob carga normal (p95) — número frouxo de propósito, dado o volume baixo de uso (ver RNF7); ajuste se algum fluxo específico precisar de garantia mais rígida
- RNF2: Comunicação assíncrona entre serviços via NATS deve tolerar entrega at-least-once; consumidores devem ser idempotentes
- RNF3: Cada serviço possui seu próprio banco de dados (database-per-service); não há FK física entre bancos de serviços diferentes
- RNF4: Toda chamada que passa pelo gateway exige token JWT válido
- RNF5: Estoque interno sempre em kg — conversão de unidade só ocorre na borda (entrada/saída)
- RNF6: Todo serviço expõe métricas (Prometheus) e tracing distribuído (OpenTelemetry)
- RNF7: Sistema é de uso interno, no máximo 3 usuários simultâneos — não há requisito de escala horizontal no MVP (vale documentar essa suposição, ela simplifica bastante decisões de infra mais adiante)
- RNF8: Senha de usuário nunca é armazenada em texto puro (hash) 🆕
- RNF9: Operações sensíveis (cancelamento de NF, alteração de forma de pagamento, ajuste manual de estoque) registram autor e timestamp de quem executou 🆕
- RNF10: Cadastro de entidade auxiliar (fornecedor, cliente, unidade de comercialização, produto) é sempre embutido na tela que depende dele — nunca exige sair da tela atual pra cadastrar em outro lugar e voltar (ver `docs/fluxos-usabilidade.md`)

---

## Próximos passos sugeridos
1. Documento fechado — não sobrou nenhum `⚠️ REVISAR` em aberto.
2. Vale registrar como ADR o princípio que guiou boa parte das decisões de hoje: "operação só vira estado pendente/intermediário quando esse estado existe de verdade na operação, independente do sistema" — explica por que compra é lançamento único mas venda tem pedido, e por que a baixa de estoque acontece na emissão da nota (RF-INV-2/RN1), não numa etapa de expedição.
3. Revisitar a suposição de que Fornecedor/Cliente pertencem a `catalog` quando desenhar o diagrama de contexto — só assumi por falta de indicação contrária.
4. Com isso fechado, volta pro TODO: "Definir os bounded contexts do MVP" e "diagrama de contexto simples (C4 Context)" — esse doc já dá material suficiente pra isso.
