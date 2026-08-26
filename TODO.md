# TODO — ERP Arroz (MTV) — Microsserviços Go+gRPC

## 📖 Como usar este TODO

Este é um **documento vivo**. Conforme eu completo um item, marco `- [x]`. Não preciso seguir a ordem à risca dentro de uma fase, mas **a ordem entre fases importa** (ex: não dá pra fazer inventory antes de catalog, nem E2E antes dos serviços existirem).

Convenções:
- `- [ ]` tarefa pendente / `- [x]` tarefa concluída
- Itens marcados com **📚 Estudar:** trazem o conceito-chave a pesquisar antes de implementar — não precisa dominar o assunto, só entender o suficiente pra tomar a decisão de design.
- Cada item foi pensado pra caber numa sessão de 2-4h. Se um item crescer muito na prática, é sinal de que ele deveria virar dois — quebre e documente aqui.
- Sem prazo. O objetivo é aprender bem feito e ter algo apresentável no fim do ano, não correr.
- Seção **Decisões em aberto (hotspots)** no fim é pra perguntas de negócio — não trave o código por causa delas, documente a suposição assumida e sigualiante.

---

## 🗺️ Fase 1 — Planejamento e modelagem

### Event storming e bounded contexts
- [x] Fazer event storming do fluxo completo (compra → recebimento → estoque → venda → expedição → financeiro → fiscal) num quadro (Miro/FigJam/papel mesmo)
  *(feito em formato diferente do previsto — RF/RN/RNF em vez de quadro de event storming clássico, ver `docs/requisitos.md`; cobre o mesmo objetivo de entender o fluxo antes de modelar contextos)*
- [x] Definir os bounded contexts do MVP (auth, catalog, inventory, purchasing, gateway) e desenhar um diagrama de contexto simples (C4 Context)
  *(ver `docs/contexto-c4.md` — inclui a relação com NATS/inventory e os contextos pós-MVP como caixa separada)*
- [ ] Listar os eventos de domínio principais que vão trafegar via NATS (ex: `MercadoriaRecebida`, `LoteCriado`, `EstoqueAtualizado`) — só a lista, sem payload ainda
  *(lista antiga ficou desatualizada pelas decisões de compra/venda — revisar contra `docs/requisitos.md` antes de fechar, ex: `MercadoriaRecebida` virou `CompraLançada`)*

### ADRs iniciais
- [x] Criar pasta `docs/adr/` e escrever ADR-0001: "Por que microsserviços + gRPC" (formato Michael Nygard: Title, Status, Context, Decision, Consequences)
- [x] ADR-0002: "Por que database-per-service"
- [x] ADR-0003: "Por que NATS para mensageria assíncrona"
  📚 Estudar: at-least-once vs exactly-once delivery — por que isso importa pro evento de recebimento de mercadoria
- [x] ADR-0004: "Kg como unidade canônica de estoque — conversões acontecem na borda"
- [x] ADR-0008 (novo, não previsto originalmente): "Operações modelam fato consumado, não estado pendente, por padrão" — princípio descoberto ao decidir os fluxos de compra/venda, ver `docs/adr/0008-fato-consumado-sem-estado-pendente.md`

### Modelagem de unidades de comercialização
- [x] Desenhar (papel/diagrama) o modelo de dados de `UnitOfMeasure`: fardo 30kg, fardo 10kg, pacote 5kg, pacote 1kg, saco 25kg, saco 50kg, saco 60kg, granel — cada unidade com peso-base em kg
  *(ver `docs/modelo-dados.md` — entidade `CATALOG_UNIT`)*
- [x] Definir a regra de conversão: toda unidade tem um `fator_conversao_kg`; estoque interno sempre em kg; conversão só acontece na entrada (compra) e saída (venda/expedição)
  *(definido em `docs/requisitos.md` — RF-CAT-6, RNF5; inclui a regra extra de fator travado após uso, corrige criando unidade nova)*
- [x] Documentar essa modelagem num ADR-0005: "Unidade de comercialização e conversão para kg"

### Modelagem de lote (rastreabilidade)
- [x] Desenhar o modelo de dados de `Lot`: código do lote, safra, fornecedor de origem, data de recebimento, quantidade em kg, produto associado
  *(ver `docs/modelo-dados.md` — entidade `INVENTORY_LOT`)*
- [x] Documentar a regra de negócio explícita: **não existe estoque sem lote associado** — toda movimentação de estoque referencia obrigatoriamente um lote
  *(documentado em `docs/requisitos.md` — RF-PUR-1/RN3, RF-INV-1/RN1, RF-INV-4/RN2, reforçado em RF-INV-5)*
- [x] Escrever ADR-0006: "Rastreabilidade por lote é premissa obrigatória do modelo de dados" — justificar com o requisito de negócio (safra, origem, rastreio)

### Modelagem de dados alto nível
- [x] Desenhar diagrama entidade-relacionamento simplificado dos 4 serviços de domínio (catalog, inventory, purchasing + auth) mostrando como os IDs cruzam entre bounded contexts (sem FK cross-database — anotar que a referência é lógica, não física)
  📚 Estudar: como referenciar entidades entre bounded contexts sem FK física (ID como referência fraca + validação assíncrona/eventual)
  *(ver `docs/modelo-dados.md` — mesmo diagrama cobre esse item e os dois de UnitOfMeasure/Lot acima, são zooms do mesmo modelo)*

### Fluxos de usabilidade
*(adicionado em 2026-08-24 — validar o fluxo humano de cada RF antes de desenhar tela ou escrever código, usando `docs/requisitos.md` como base; concluído em 2026-08-24, ver `docs/fluxos-usabilidade.md`)*
- [x] Desenhar (Mermaid `flowchart`) o fluxo de uso de "Lançar compra" (RF-PUR-1)
- [x] Desenhar o fluxo de "Cadastrar pedido de venda" (RF-VEN-1) e "Faturar pedido de venda" (RF-VEN-2) — incluir o caso de quantidade faturada divergente do pedido
- [x] Desenhar o fluxo de cadastro de produto e unidade de comercialização (RF-CAT-1, RF-CAT-6)
- [x] Desenhar o fluxo de consulta de estoque e rastreabilidade de lote (RF-INV-3, RF-INV-5)
- [x] Desenhar o fluxo de login e sessão (RF-AUTH-2, RF-AUTH-3) — adicionado durante a validação, tela mínima da Fase 5 que não tinha entrado na lista original
- [x] Revisar os fluxos desenhados contra `docs/requisitos.md` — todo RF relevante ao MVP tem fluxo correspondente ou foi marcado como dispensável por repetir padrão já coberto (ver seção "Revisão final" em `docs/fluxos-usabilidade.md`)

### Design de interface (Claude Design)
*(adicionado em 2026-08-24 — só começa depois dos fluxos de usabilidade acima, pra não gerar tela genérica desgarrada do fluxo real; concluído em 2026-08-25)*
- [x] Escrever prompt para o Claude Design cobrindo: papéis (admin/operador), telas do MVP, o fluxo de cada uma (baseado nos fluxos acima) e o contexto de uso (ferramenta interna, poucos usuários simultâneos, prioridade em densidade de informação e velocidade de digitação sobre estética)
  *(feito direto como canvas, sem prompt intermediário — ver nota abaixo)*
- [x] Gerar o design e revisar contra os fluxos desenhados
  *(protótipo clicável com as 6 telas do MVP, publicado como Artifact — login, lançar compra, pedido de venda, faturar venda, cadastrar produto, consultar estoque)*
- [x] Ajustar `docs/requisitos.md` ou os fluxos se o design revelar alguma lacuna de regra de negócio
  *(nenhuma lacuna nova — desenhar as telas confirmou que os fluxos já validados cobriam o necessário)*

---

## 🏗️ Fase 2 — Fundação técnica

### Template reutilizável de microsserviço
- [x] Criar repositório/diretório `service-template` com estrutura de pastas padrão Go (cmd/, internal/, pkg/, migrations/, proto/)
  📚 Estudar: Standard Go Project Layout — o que faz sentido adotar e o que é exagero pro nosso caso
- [ ] Configurar `go.mod`, Makefile básico (build, test, run, lint) no template
  *(go.mod feito; falta o Makefile)*
- [x] Adicionar setup de log/slog estruturado (JSON handler) como padrão do template
- [x] Adicionar setup de config via env vars (sem lib externa pesada — usar stdlib `os.Getenv` + struct de config validada na inicialização)
- [x] Adicionar healthcheck básico (endpoint/RPC de liveness e readiness)
- [x] Adicionar Dockerfile multi-stage padrão (build Go estático + imagem final mínima, ex: distroless ou alpine)
  *(imagem final ~18.5MB, distroless — vs quase 1GB da imagem golang usada só pra compilar)*
- [ ] Adicionar setup base do OpenTelemetry (tracer provider + exporter) no template
  📚 Estudar: OpenTelemetry Go SDK — diferença entre TracerProvider, Exporter e Span; como instrumentar gRPC automaticamente com interceptors
- [x] Adicionar setup base do Prometheus (endpoint `/metrics` com client_golang) no template
- [ ] Adicionar interceptors gRPC padrão (logging, recovery de panic, tracing) no template
  📚 Estudar: gRPC interceptors (unary e stream) — como compor múltiplos interceptors numa chain

### Setup de infraestrutura do projeto

> **Decisão (2026-08-26):** confirmado (via SSH no servidor) que a empresa MTV **tem sim um cluster Kubernetes real** — `k3s` rodando no servidor (`tiagoserver`), com Prometheus, Grafana, node-exporter e cAdvisor já ativos via Docker no mesmo host. Primeira suposição (de que não existia) estava errada — bom ter conferido antes de montar tudo num cluster local à toa. Falta só resolver o acesso ao `kubectl` a partir da máquina de desenvolvimento (ver primeiro item abaixo).

- [x] Resolver acesso ao `kubectl` a partir da máquina de dev (copiar kubeconfig do k3s, ver se a API do k3s está acessível pela rede ou só via SSH/túnel)
  *(API do k3s acessível direto pela rede local, sem túnel — kubeconfig copiado do servidor pra `~/.kube/config`, fora do git)*
- [ ] Criar namespace Kubernetes dedicado pro projeto no cluster
- [ ] Criar manifests base (Deployment, Service, ConfigMap, Secret) genéricos reutilizáveis por serviço
- [ ] Configurar PostgreSQL no k8s (um banco por serviço) — decidir: operator (ex: CloudNativePG) ou StatefulSet simples
  📚 Estudar: database-per-service na prática — isolamento de credenciais, backup por banco
- [ ] Deployar NATS no cluster (modo standalone, sem clustering por enquanto — não precisa de HA no MVP)
- [ ] Validar que Prometheus/Grafana já rodando no servidor conseguem fazer scrape de um pod de teste

### CI/CD básico
- [ ] Escolher e configurar pipeline de CI (GitHub Actions ou similar) rodando lint + testes a cada push
- [ ] Adicionar build de imagem Docker automatizado no CI
- [ ] Documentar (README) o fluxo de deploy manual pro k8s da empresa (sem CD automático ainda — over-engineering pro estágio atual)

### Ferramentas de desenvolvimento
- [ ] Instalar e configurar `golang-migrate` no template (comando padrão pra criar/rodar migrations)
- [ ] Instalar e configurar `sqlc` no template (config `sqlc.yaml`, geração de código a partir de queries SQL)
  📚 Estudar: sqlc — como ele gera código type-safe a partir de SQL puro, diferença pra um ORM tradicional
- [ ] Instalar `buf` e configurar `buf.gen.yaml` pra geração de código Go a partir de `.proto`
  📚 Estudar: Buf — lint de proto, breaking change detection, geração de código

---

## 🔐 Fase 3 — Primeiro serviço (auth-service)

### Modelagem e proto
- [ ] Definir entidades mínimas: `User`, `Role` (ex: admin, operador, financeiro — só o suficiente pro MVP)
- [ ] Escrever `auth.proto` com RPCs: `Login`, `ValidateToken`, `CreateUser`
  📚 Estudar: protobuf schema evolution — regras de compatibilidade (campos novos sempre opcionais, nunca reaproveitar número de campo removido)
- [ ] Gerar código Go a partir do proto com buf

### Implementação
- [ ] Clonar o service-template pra `auth-service`
- [ ] Criar migration inicial (tabela `users`, `roles`)
- [ ] Implementar queries sqlc (create user, find by email, etc.)
- [ ] Implementar hash de senha (bcrypt via stdlib-adjacent lib, ex: `golang.org/x/crypto/bcrypt`)
- [ ] Implementar geração e validação de JWT
  📚 Estudar: JWT — claims padrão (exp, iat, sub), assinatura HS256 vs RS256, onde guardar a chave secreta
- [ ] Implementar RPC `Login` (valida credenciais, retorna token)
- [ ] Implementar RPC `ValidateToken` (usado pelos outros serviços via gRPC)
- [ ] Implementar RPC `CreateUser` (admin cria novos usuários)

### Qualidade e observabilidade
- [ ] Escrever testes unitários das regras de negócio (hash, validação de token)
- [ ] Escrever teste de integração do fluxo de login (contra banco real via testcontainers ou docker-compose)
  📚 Estudar: testcontainers-go — como subir Postgres descartável pra teste de integração
- [ ] Validar que métricas Prometheus aparecem no Grafana pro auth-service
- [ ] Validar que traces do auth-service aparecem no backend de tracing configurado
- [ ] Escrever README do serviço (o que faz, como rodar local, variáveis de ambiente)

### Deploy
- [ ] Escrever manifests k8s específicos do auth-service (a partir dos genéricos da Fase 2)
- [ ] Deployar auth-service no cluster e validar healthcheck respondendo
- [ ] Testar login end-to-end via `grpcurl` contra o serviço deployado

---

## 📦 Fase 4 — Serviços seguintes (clonando o template)

### catalog-service
- [ ] Clonar template pra `catalog-service`
- [ ] Modelar entidade `Product` (tipo/marca de arroz — ex: tipo 1, tipo 2, parboilizado, integral)
- [ ] Modelar entidade `UnitOfMeasure` com os valores definidos na Fase 1 (fardo 30kg, fardo 10kg, pacote 5kg, pacote 1kg, saco 25kg, saco 50kg, saco 60kg, granel) e campo `peso_base_kg`
- [ ] Implementar mecanismo de conversão: função/método que recebe quantidade + unidade e retorna quantidade em kg (e o inverso, kg → unidade de venda)
- [ ] Escrever proto `catalog.proto` (RPCs: CRUD de Product, CRUD/list de UnitOfMeasure, RPC de conversão)
- [ ] Migrations + queries sqlc (products, units_of_measure)
- [ ] Implementar RPCs e testes unitários (especial atenção nos testes de conversão de unidade — casos de borda tipo fração de kg)
- [ ] Seed de dados inicial (as unidades padrão já listadas, alguns produtos de exemplo)
- [ ] Deploy no k8s + validação de métricas/traces (deve ser mais rápido que o auth-service, já que o template está validado)

### inventory-service
- [ ] Clonar template pra `inventory-service`
- [ ] Modelar entidade `Lot` (código, safra, fornecedor de origem, data de recebimento, quantidade em kg, produto associado — produto referenciado por ID lógico do catalog-service)
- [ ] Modelar entidade `StockMovement` (tipo: entrada/saída, quantidade em kg, lote associado, timestamp, origem do movimento)
- [ ] Implementar validação de domínio explícita: **rejeitar qualquer movimentação de estoque sem lote associado** (regra de negócio, não só constraint de banco)
  📚 Estudar: onde colocar validação de invariante de domínio em Go — validação na camada de serviço vs constraint NOT NULL no banco (fazer as duas, mas a de domínio é a que dá erro de negócio claro)
- [ ] Escrever proto `inventory.proto` (RPCs: CreateLot, RegisterMovement, GetStockByProduct, GetLotDetails)
- [ ] Migrations + queries sqlc (lots, stock_movements)
- [ ] Implementar RPCs com a validação de lote obrigatório
- [ ] Testes unitários da regra "sem lote não existe estoque" (caso de erro esperado)
- [ ] Testes de consulta de saldo de estoque agregado por produto (soma de lotes)
- [ ] Deploy no k8s + validação de métricas/traces

### purchasing-service
*(atualizado em 2026-08-24 — sem "pedido de compra" como entidade separada, ver RF-PUR-1 em `docs/requisitos.md`: compra é lançada num único passo)*
- [ ] Clonar template pra `purchasing-service`
- [ ] Modelar entidade `Purchase` (compra — fornecedor, produto/quantidade/unidade, dados da nota fiscal de entrada, tudo lançado de uma vez)
- [ ] Escrever proto `purchasing.proto` (RPC: CreatePurchase)
- [ ] Migrations + queries sqlc (purchases)
- [ ] Implementar lógica de lançamento: ao registrar uma `Purchase`, chamar (via gRPC síncrono, por enquanto) o inventory-service pra **criar o lote correspondente** antes de confirmar a compra
- [ ] Testes unitários e de integração do fluxo compra lançada → lote criado no inventory
- [ ] Deploy no k8s + validação de métricas/traces

---

## 🌐 Fase 5 — API Gateway + Frontend

### api-gateway (BFF REST)
- [ ] Clonar template pra `api-gateway` (adaptado: expõe REST, não gRPC, pro mundo externo)
- [ ] Definir rotas REST do gateway mapeando pros RPCs dos 4 serviços (auth, catalog, inventory, purchasing)
- [ ] Implementar tradução REST → gRPC (handlers HTTP chamando clients gRPC internos)
  📚 Estudar: padrão BFF (Backend For Frontend) — por que o gateway não deveria ter lógica de negócio própria
- [ ] Implementar middleware de autenticação no gateway (valida JWT via chamada ao auth-service)
- [ ] Gerar documentação OpenAPI/Swagger das rotas do gateway
- [ ] Deploy do gateway no k8s + Ingress configurado

### Fundação do frontend
- [ ] Decidir entre PrimeVue e Naive UI (avaliar componentes de tabela/formulário disponíveis, tamanho do bundle)
- [ ] Criar projeto Nuxt 3 (`frontend/`)
- [ ] Configurar Tailwind CSS no projeto
- [ ] Configurar Pinia (store de auth com token JWT)
- [ ] Configurar TanStack Query pra Vue (client de dados apontando pro api-gateway)
- [ ] Configurar VeeValidate + Zod (schema de validação de formulário compartilhado)
  📚 Estudar: VeeValidate + Zod — como conectar schema Zod ao form do VeeValidate pra validação tipada

### Telas mínimas do MVP
- [ ] Tela de login (consome auth-service via gateway)
- [ ] Tela de listagem/cadastro de produtos e unidades de medida (catalog)
- [ ] Tela de consulta de estoque por produto/lote, incluindo rastreabilidade (inventory)
- [ ] Tela de lançamento de compra — passo único (purchasing, ver RF-PUR-1)
- [ ] Deploy do frontend (build estático ou SSR no k8s, conforme decisão de infra)

---

## 🔄 Fase 6 — Mensageria e sagas

### Primeira comunicação assíncrona
*(evento renomeado em 2026-08-24 de `MercadoriaRecebida` pra `CompraLançada`, refletindo o lançamento em passo único — ver RF-PUR-1)*
- [ ] Definir schema do evento `CompraLançada` (payload: lote, produto, quantidade em kg, timestamp)
  📚 Estudar: versionamento de eventos em mensageria — como evoluir o schema de um evento sem quebrar consumidores antigos
- [ ] Trocar a chamada síncrona gRPC de lançamento de compra (Fase 4) por publish do evento `CompraLançada` via NATS no purchasing-service
- [ ] Implementar subscriber do evento no inventory-service (cria o lote a partir do evento recebido)
- [ ] Implementar idempotência no consumidor (mesmo evento processado duas vezes não duplica lote)
  📚 Estudar: idempotência em consumidores de mensageria — chave de deduplicação, at-least-once delivery do NATS
- [ ] Testar cenário de falha: inventory-service fora do ar durante publish — validar que a mensagem não se perde (JetStream com persistência, se aplicável)
  📚 Estudar: NATS Core vs JetStream — quando vale a pena usar JetStream (persistência, replay) em vez do Core pub/sub
- [ ] Documentar o fluxo assíncrono resultante num ADR-0007: "Lançamento de compra via evento assíncrono (NATS)"

### Observabilidade da mensageria
- [ ] Propagar trace context através do evento NATS (correlacionar span do publish com o do consumo)
- [ ] Adicionar métrica de mensagens processadas/falhas no inventory-service

---

## ✅ Fase 7 — Testes end-to-end e polimento

- [ ] Escrever teste E2E do fluxo completo: login → cadastro de produto → criação de pedido de compra → recebimento → evento assíncrono → lote criado → consulta de estoque
- [ ] Revisar tratamento de erros em todos os serviços (mensagens de erro consistentes, códigos gRPC apropriados)
  📚 Estudar: gRPC status codes — quando usar `InvalidArgument` vs `FailedPrecondition` vs `NotFound`
- [ ] Revisar logs estruturados de todos os serviços (padronizar campos: `service`, `trace_id`, `level`)
- [ ] Criar dashboard Grafana consolidado do MVP (requests/s, latência, taxa de erro por serviço)
- [ ] Revisar e ajustar timeouts/retries nas chamadas gRPC entre serviços
- [ ] Rodar teste de carga leve (ex: `k6` ou `hey`) no gateway pra validar que não quebra sob uso simultâneo básico
- [ ] Revisão de segurança básica (secrets não hardcoded, HTTPS no ingress, rate limit simples no gateway)

---

## 📚 Fase 8 — Documentação final e preparação da apresentação

- [ ] Montar portal VitePress (`docs/`) com estrutura de navegação (Visão Geral, ADRs, Serviços, Domínio)
- [ ] Gerar e publicar docs dos `.proto` via Buf no portal
- [ ] Publicar Swagger/OpenAPI do gateway no portal
- [ ] Desenhar diagramas C4 finais (Context, Container, Component dos 4 serviços de domínio) em Mermaid
- [ ] Escrever página de domínio explicando lote, unidades de comercialização e conversão pra kg (pública, didática)
- [ ] Revisar todos os ADRs (status atualizado: aceito/superado)
- [ ] Preparar roteiro de demo (login → cadastro → compra → recebimento → estoque atualizado em tempo real via evento)
- [ ] Preparar slides/material de apresentação acadêmica (arquitetura, decisões, aprendizados, o que ficou de fora)
- [ ] Gravar vídeo de backup da demo (caso algo falhe ao vivo no dia da apresentação)

---

## ❓ Decisões em aberto (hotspots)

Perguntas de negócio que preciso validar com a empresa antes (ou durante) de modelar certas partes com mais profundidade. Não bloqueiam o MVP — assumo a suposição mais simples e documento, mas preciso confirmar antes de ir pra produção real.

- [ ] Contratos de fornecimento de longo prazo ou só compra pontual por pedido?
- [ ] Múltiplos depósitos/armazéns ou depósito único no início?
- [x] Existe alçada de aprovação para pedidos de compra acima de um valor X? → **Não.** Resolvido em `docs/requisitos.md` (RF-PUR-1): compra lançada num único passo, sem aprovação.
- [ ] Existe limite de crédito por cliente (relevante pro futuro sales-service)?
- [ ] Frete: CIF ou FOB — quem contrata a transportadora?
- [ ] Quais formas de pagamento são aceitas (boleto, transferência, prazo)?
- [ ] Existe controle de qualidade no recebimento (umidade, impureza, quebra de grão)? Isso afeta o modelo de `Lot`?
- [x] Política de precificação: preço é por unidade de venda (fardo/saco/pacote) ou sempre por kg? O preço pode variar por lote (safra) ou é fixo por produto? → **Variável.** Resolvido em `docs/requisitos.md` (RF-VEN-1/RN5): preço é sempre manual, decidido pedido a pedido, não é fixo por produto.

---

## 🔮 Backlog pós-MVP

Serviços que ficam pra depois da apresentação, sem detalhamento de tarefas ainda — só o escopo alto nível:

- **partners-service** — cadastro de fornecedores e clientes (hoje simplificado como referência solta em purchasing/sales)
- **sales-service** — pedidos de venda, expedição, integração com inventory pra baixa de estoque por lote
- **financial-service** — contas a pagar/receber, conciliação, vinculado a purchasing e sales
- **fiscal-service** — emissão de NFe via Focus NFe (ou similar), vinculado a sales e financial
- **notification-service** — notificações (email/whatsapp) de eventos de negócio (pedido aprovado, NFe emitida, etc.)

---

## 🔗 Referências úteis

- [gRPC-Go](https://grpc.io/docs/languages/go/) — documentação oficial gRPC para Go
- [sqlc](https://docs.sqlc.dev/) — geração de código SQL type-safe
- [NATS Docs](https://docs.nats.io/) — mensageria (Core e JetStream)
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/) — instrumentação e tracing distribuído
- [Focus NFe — Documentação da API](https://focusnfe.com.br/doc/) — emissão de NFe via API terceira
- [Nuxt 3](https://nuxt.com/docs) — framework Vue
- [PrimeVue](https://primevue.org/) — biblioteca de componentes
- [Naive UI](https://www.naiveui.com/) — biblioteca de componentes (alternativa)
- [VitePress](https://vitepress.dev/) — portal de documentação
- [Buf Docs](https://buf.build/docs/) — lint, breaking change detection e geração de docs de proto
- [golang-migrate](https://github.com/golang-migrate/migrate) — migrations de banco
- [ADR - Documenting Architecture Decisions (Michael Nygard)](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions) — formato original de ADR
