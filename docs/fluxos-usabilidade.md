# Fluxos de usabilidade — ERP Arroz (MTV)

> Um fluxo por RF relevante do MVP, validado um de cada vez antes de seguir pro design (Claude Design). Baseado em `docs/requisitos.md`.

**Padrão de UX válido pra todos os fluxos abaixo (definido 2026-08-24):** cadastro de entidade auxiliar (fornecedor, cliente, unidade de comercialização, produto) é sempre **embutido** na tela que depende dele — nunca manda o operador sair, ir cadastrar em outro lugar, e voltar.

---

## Fluxo 1: Lançar compra (RF-PUR-1)

```mermaid
flowchart TD
    A[Operador abre tela Lançar Compra] --> B[Seleciona fornecedor]
    B --> C{Fornecedor já existe?}
    C -- Não --> C1[Cadastra fornecedor ali mesmo — RF-CAT-4]
    C -- Sim --> C2[Seleciona da lista]
    C1 --> N[Informa dados da nota fiscal de entrada: número, data, valor]
    C2 --> N
    N --> N1{Fornecedor + número de NF já lançados antes?}
    N1 -- Sim --> N2[Bloqueia: nota já lançada — RF-PUR-1/RN6]
    N1 -- Não --> D[Adiciona produto: seleciona produto + informa quantidade + unidade]
    D --> E[Sistema converte quantidade para kg automaticamente e mostra pro operador conferir]
    E --> F{Quer adicionar outro produto na mesma compra?}
    F -- Sim --> D
    F -- Não --> H[Confirma lançamento]
    H --> I{Validação: produtos ativos, quantidades > 0?}
    I -- Não --> I1[Mostra erro, mantém formulário preenchido]
    I -- Sim --> J[Sistema cria um lote por produto lançado, associa quantidade em kg]
    J --> K[Estoque atualizado — RF-INV-1]
    K --> L[Confirma compra lançada e mostra lotes criados]
```

**Validado (2026-08-24):** compra pode ter mais de um produto (um lote por produto); bloqueio de nota fiscal duplicada por fornecedor+número, refletido no `requisitos.md` (RF-PUR-1/RN1, RN3, RN6).

---

## Fluxo 2: Pedido de venda → Faturamento (RF-VEN-1, RF-VEN-2)

```mermaid
flowchart TD
    A[Vendedor abre tela Cadastrar Pedido de Venda] --> B[Seleciona cliente]
    B --> C{Cliente já existe?}
    C -- Não --> C1[Cadastra cliente ali mesmo — RF-CAT-5]
    C -- Sim --> C2[Seleciona da lista]
    C1 --> D[Adiciona produto: produto + quantidade + unidade + preço manual]
    C2 --> D
    D --> E{Quer adicionar outro produto no pedido?}
    E -- Sim --> D
    E -- Não --> F[Confirma pedido]
    F --> G[Pedido salvo, status em aberto — entra na projeção de caixa RF-VEN-4]

    G -.antes de faturar.-> H{O que acontece com o pedido em aberto?}
    H -- Alterar --> D
    H -- Cancelar --> H1[Pedido cancelado — some da projeção, fica no histórico RF-VEN-1/RN4]
    H -- Faturar --> I[Abre tela Faturar Pedido de Venda]

    S[Venda direta, sem pedido prévio] --> I
    I --> I2{Vincula a um pedido existente ou cria um novo na hora?}
    I2 -- Existente --> I3[Seleciona o pedido em aberto]
    I2 -- Novo --> I4[Cadastra cliente + itens ali mesmo, na tela de faturar]
    I3 --> J
    I4 --> J[Revisa/edita quantidade e preço de cada item — pode divergir do pedido original]
    J --> K[Confirma faturamento]
    K --> L{Estoque disponível pra quantidade faturada de cada lote?}
    L -- Não --> L1[Bloqueia: saldo negativo — RF-INV-2/RN2]
    L -- Sim --> M[Gera nota fiscal de venda]
    M --> N[Baixa estoque pela quantidade faturada — RF-INV-2]
    N --> O[Gera título a receber — RF-FIN-1]
    O --> P[Emite boleto junto, se a forma de pagamento exigir — RF-VEN-2/RN4]
    P --> Q[Pedido fecha por completo, mesmo com quantidade diferente da pedida — RF-VEN-2/RN5]
    Q --> R[Confirma venda faturada]
```

**Validado (2026-08-24):**
1. Faturamento nem sempre depende de pedido pré-existente — a tela de faturar permite vincular a um pedido em aberto **ou** criar um pedido novo ali mesmo (venda direta, sem precisar ir cadastrar em outro lugar antes). Refletido em `RF-VEN-2/RN1`.
2. Qualquer operador pode faturar, inclusive o vendedor que criou o pedido — sem restrição extra de permissão, empresa pequena. Refletido em `RF-VEN-2/RN6`.

---

## Fluxo 3: Cadastrar produto e unidade de comercialização (RF-CAT-1, RF-CAT-6, RF-CAT-2, RF-CAT-7)

```mermaid
flowchart TD
    A[Operador abre tela Cadastrar Produto] --> B[Informa nome/tipo do produto]
    B --> C[Associa unidade de comercialização]
    C --> D{Unidade já existe?}
    D -- Sim --> E[Seleciona da lista]
    D -- Não --> F[Cadastra nova unidade ali mesmo: nome + fator de conversão pra kg]
    F --> E
    E --> G{Quer associar outra unidade ao produto?}
    G -- Sim --> C
    G -- Não --> H[Confirma produto]
    H --> I[Produto salvo, ativo — aparece no catálogo RF-CAT-3]

    I -.depois.-> J{O que fazer com o produto?}
    J -- Editar --> K[Altera nome/unidades associadas — fator de conversão de unidade já existente continua travado RF-CAT-6/RN2]
    J -- Inativar --> L[Produto inativado — some do catálogo, mas fica no histórico RF-CAT-2]
```

**Validado (2026-08-24):** cadastro auxiliar sempre embutido na tela que depende dele — padronizado em todos os fluxos (ver nota no topo do arquivo). Fluxo 1 (fornecedor) e Fluxo 2 (cliente) já ajustados pro mesmo padrão.

---

## Fluxo 4: Consultar estoque, rastreabilidade e ajuste manual (RF-INV-3, RF-INV-4, RF-INV-5)

```mermaid
flowchart TD
    A[Operador abre tela Consultar Estoque] --> B{Buscar por produto ou por lote?}
    B -- Produto --> C[Seleciona produto]
    C --> D[Mostra total em kg do produto + lista de lotes com saldo cada]
    B -- Lote --> E[Busca pelo código do lote]
    E --> D2[Mostra saldo em kg daquele lote]

    D --> F[Seleciona um lote da lista]
    F --> G{O que fazer com esse lote?}
    D2 --> G

    G -- Ver rastreabilidade --> H[Mostra origem: fornecedor, compra, data — RF-INV-5]
    H --> I[Mostra todos os destinos: vendas, trocas, ajustes associados a esse lote]

    G -- Ajustar estoque --> J[Informa quantidade do ajuste, positiva ou negativa]
    J --> K[Informa justificativa — obrigatória, RF-INV-4/RN1]
    K --> L[Confirma ajuste]
    L --> M[Sistema registra movimentação vinculada ao lote, atualiza saldo]
    M --> N[Registra autor + timestamp do ajuste — RNF9]
```

**Notas:**
- Ajuste manual pode somar ou subtrair estoque (perda/quebra reduz, contagem física a mais soma) — já implícito em RF-INV-4, não precisou de decisão nova.
- Ajuste liberado pra qualquer operador, mesmo padrão de liberdade já definido pra faturamento (RF-VEN-2/RN6) — assumi por consistência, me avisa se quiser restringir só pra esse caso.

---

## Fluxo 5: Login e sessão (RF-AUTH-2, RF-AUTH-3)

```mermaid
flowchart TD
    A[Usuário abre tela de Login] --> B[Informa e-mail e senha]
    B --> C[Confirma]
    C --> D{Credenciais válidas?}
    D -- Não --> D1[Erro genérico: e-mail ou senha inválidos, sem dizer qual — RF-AUTH-2/RN1]
    D1 --> B
    D -- Sim --> E[Sistema gera token JWT com expiração — RF-AUTH-2/RN2]
    E --> F[Redireciona pra tela inicial]

    F -.durante o uso.-> G[Toda chamada ao gateway valida o token]
    G --> H{Token válido e não expirado?}
    H -- Sim --> I[Chamada segue normalmente — RF-AUTH-3]
    H -- Não --> J[Bloqueia chamada, desloga, volta pra tela de Login]
```

**Nota:** não desenhei fluxo de "esqueci minha senha" — não existe RF pra isso hoje (RF-AUTH só tem cadastrar/autenticar/validar). Pra 4-5 usuários, provavelmente o `admin` resolve resetando a senha diretamente por fora do sistema, mas se você quiser isso como funcionalidade formal, é um RF novo em `requisitos.md`. Deixo como está, só avisando — não acho que trava o MVP.

Com esse, os **5 fluxos do MVP estão fechados**: compra, pedido de venda + faturamento, catálogo, estoque/rastreabilidade, e login.

---

## Revisão final (2026-08-24)

Conferido cada RF do MVP (não pós-MVP) contra os 5 fluxos acima:

| RF | Coberto por |
|---|---|
| RF-AUTH-1 Cadastrar usuário | **Não desenhado** — mesmo padrão de formulário simples já repetido 3x (fornecedor, cliente, produto); não achei que valia um 6º diagrama igual |
| RF-AUTH-2/3 Login e sessão | Fluxo 5 |
| RF-CAT-1/2/6/7 Produto, unidade, inativar, editar | Fluxo 3 |
| RF-CAT-3 Consultar catálogo | **Não desenhado** — é só listagem/busca simples, sem decisão de fluxo relevante |
| RF-CAT-4/5 Fornecedor, cliente | Embutido nos Fluxos 1 e 2 (RNF10) |
| RF-PUR-1 Lançar compra | Fluxo 1 |
| RF-INV-1/2 Atualizar/baixar estoque | Consequência automática dos Fluxos 1 e 2, sem tela própria |
| RF-INV-3/4/5 Consultar, ajustar, rastrear | Fluxo 4 |

Os dois "não desenhados" são deliberados, não esquecidos: telas de formulário/listagem simples que não têm decisão de UX nova pra validar — desenhar deixaria de ser útil e viraria repetição. Se qualquer um dos dois esconder alguma regra que só aparece na hora de fazer a tela, dá pra voltar aqui.
