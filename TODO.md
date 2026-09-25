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
  *(ver `docs/contexto-c4.md` — inclui a relação com RabbitMQ/inventory e os contextos pós-MVP como caixa separada; era NATS, atualizado em 2026-08-27 — ver ADR-0009)*
- [ ] Listar os eventos de domínio principais que vão trafegar via RabbitMQ (ex: `MercadoriaRecebida`, `LoteCriado`, `EstoqueAtualizado`) — só a lista, sem payload ainda
  *(lista antiga ficou desatualizada pelas decisões de compra/venda — revisar contra `docs/requisitos.md` antes de fechar, ex: `MercadoriaRecebida` virou `CompraLançada`)*

### ADRs iniciais
- [x] Criar pasta `docs/adr/` e escrever ADR-0001: "Por que microsserviços + gRPC" (formato Michael Nygard: Title, Status, Context, Decision, Consequences)
- [x] ADR-0002: "Por que database-per-service"
- [x] ADR-0003: "Por que NATS para mensageria assíncrona"
  📚 Estudar: at-least-once vs exactly-once delivery — por que isso importa pro evento de recebimento de mercadoria
  *(superada por ADR-0009 em 2026-08-27 — projeto trocou pra RabbitMQ, motivo de aprendizado/currículo)*
- [x] ADR-0009 (novo): "RabbitMQ para mensageria assíncrona (substitui ADR-0003)"
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
- [x] Configurar `go.mod`, Makefile básico (build, test, run, lint) no template
  *(Makefile com os 4 alvos desde o início; `go.mod` ganhou grpc/gorm/otel como deps diretas em 2026-09-14, durante o resync do template)*
- [x] Adicionar setup de log/slog estruturado (JSON handler) como padrão do template
- [x] Adicionar setup de config via env vars (sem lib externa pesada — usar stdlib `os.Getenv` + struct de config validada na inicialização)
- [x] Adicionar healthcheck básico (endpoint/RPC de liveness e readiness)
- [x] Adicionar Dockerfile multi-stage padrão (build Go estático + imagem final mínima, ex: distroless ou alpine)
  *(imagem final ~18.5MB, distroless — vs quase 1GB da imagem golang usada só pra compilar)*
- [x] Adicionar setup base do OpenTelemetry (tracer provider + exporter) no template
  📚 Estudar: OpenTelemetry Go SDK — diferença entre TracerProvider, Exporter e Span; como instrumentar gRPC automaticamente com interceptors
  *(feito em 2026-09-14 — `internal/observability/tracing.go`: Resource + Exporter OTLP gRPC (`otlptracegrpc`, lê `OTEL_EXPORTER_OTLP_ENDPOINT`, precisa do esquema `http://` senão falha silencioso) + TracerProvider registrado global + propagator `TraceContext`. Retrofit no auth-service também, não só no template)*
- [x] Adicionar setup base do Prometheus (endpoint `/metrics` com client_golang) no template
- [x] Adicionar interceptors gRPC padrão (logging, recovery de panic, tracing) no template
  📚 Estudar: gRPC interceptors (unary e stream) — como compor múltiplos interceptors numa chain
  *(feito em 2026-09-14 — `internal/interceptors/recovery.go` (recupera panic, vira `codes.Internal`) e `logging.go` (método/duração/código por chamada), encadeados via `grpc.ChainUnaryInterceptor`. Tracing por request via `otelgrpc.NewServerHandler()` como `grpc.StatsHandler`, não como interceptor manual — é o padrão atual da lib)*

### Setup de infraestrutura do projeto

> **Decisão (2026-08-26):** confirmado (via SSH no servidor) que a empresa MTV **tem sim um cluster Kubernetes real** — `k3s` rodando no servidor (`tiagoserver`), com Prometheus, Grafana, node-exporter e cAdvisor já ativos via Docker no mesmo host. Primeira suposição (de que não existia) estava errada — bom ter conferido antes de montar tudo num cluster local à toa. Falta só resolver o acesso ao `kubectl` a partir da máquina de desenvolvimento (ver primeiro item abaixo).

- [x] Resolver acesso ao `kubectl` a partir da máquina de dev (copiar kubeconfig do k3s, ver se a API do k3s está acessível pela rede ou só via SSH/túnel)
  *(API do k3s acessível direto pela rede local, sem túnel — kubeconfig copiado do servidor pra `~/.kube/config`, fora do git)*
- [x] Criar namespace Kubernetes dedicado pro projeto no cluster
  *(namespace `mtv-erp`, definido em `deploy/base/namespace.yaml`; namespace `comprassularroz` já existente no cluster não foi tocado)*
- [x] Criar manifests base (Deployment, Service, ConfigMap, Secret) genéricos reutilizáveis por serviço
  *(em `deploy/service-template/` — Deployment/Service/ConfigMap aplicados e validados no cluster real; Secret fica como molde documentado, sem segredo real ainda)*
- [x] Configurar PostgreSQL no k8s (um banco por serviço) — decidir: operator (ex: CloudNativePG) ou StatefulSet simples
  📚 Estudar: database-per-service na prática — isolamento de credenciais, backup por banco
  *(decisão: StatefulSet simples, sem operator — RNF7 não justifica a complexidade de failover/backup automático. Primeiro banco criado: `auth-db`, em `deploy/auth-db/`, credenciais via Secret criado imperativamente, nunca commitado)*
- [x] ~~Deployar NATS no cluster~~ — SUPERADO, ver item abaixo
  *(Deployment + Service ficaram em `deploy/nats/`, validado via /varz — mas a decisão de mensageria mudou pra RabbitMQ em 2026-08-27, ver ADR-0009. Manifests antigos ainda não removidos do cluster.)*
- [x] Deployar RabbitMQ no cluster (modo standalone, sem clustering — não precisa de HA no MVP) — substitui o item do NATS acima
  📚 Estudar: RabbitMQ — conceitos de exchange, queue, binding, virtual host
  *(Deployment + Service em `deploy/rabbitmq/`, imagem `rabbitmq:3.13-management` com UI web na porta 15672; NATS removido do cluster e da pasta `deploy/`)*
- [x] Validar que Prometheus/Grafana já rodando no servidor conseguem fazer scrape de um pod de teste
  *(feito em 2026-09-08 com o próprio auth-service. Prometheus roda como container Docker fora do k3s (rede `monitoring_default`, sem k8s SD), então: Service `NodePort` dedicado `auth-service-metrics` (`deploy/auth-service/service-metrics.yaml`, nodePort 30080) expõe o `/metrics` do pod; job estático `auth-service` no `prometheus.yml` do servidor apontando pra `172.18.0.1:30080` (gateway da bridge = host). Target UP, reload via `docker kill --signal=HUP prometheus`)*

### CI/CD básico
- [x] Escolher e configurar pipeline de CI (GitHub Actions ou similar) rodando lint + testes a cada push
  *(`.github/workflows/ci.yml`, escopo hoje só cobre `service-template` — quando o `auth-service` nascer, vira matrix strategy ou job duplicado)*
- [x] Adicionar build de imagem Docker automatizado no CI
  *(sem push pra registro ainda — só valida que o Dockerfile builda)*
- [x] Documentar (README) o fluxo de deploy manual pro k8s da empresa (sem CD automático ainda — over-engineering pro estágio atual)
  *(ver `docs/deploy.md` — inclui as pegadinhas descobertas na prática: `/tmp` isolado por SSH, imagePullPolicy, rollout restart com tag `:latest`)*

### Ferramentas de desenvolvimento
*(decisão revista em 2026-08-27: GORM no lugar de sqlc como camada de acesso a dados — recomendação do professor, banco continua Postgres; `golang-migrate` mantido para migrations versionadas em vez do `AutoMigrate` do GORM, pra não depender de inferência de schema em produção)*
- [x] Instalar e configurar `golang-migrate` no template (comando padrão pra criar/rodar migrations)
  *(feito em 2026-09-10 — `migrate` CLI instalado (`go install .../migrate/v4/cmd/migrate`), `service-template/migrations/README.md` documenta a convenção de nomes e os comandos `create`/`up`/`version`)*
- [x] Instalar e configurar `GORM` no template (conexão com Postgres via driver `gorm.io/driver/postgres`, structs de modelo por serviço)
  *(feito direto no `auth-service`, não no template vazio — mesmo raciocínio já usado pro `buf`. `internal/db/models.go` (struct User) + `internal/db/db.go` (Connect), `DATABASE_URL` no Config. Conexão testada de verdade contra o `auth-db` no cluster)*
  📚 Estudar: GORM — Active Record vs Data Mapper, `AutoMigrate` vs migrations versionadas, N+1 em preload de associações
- [x] Instalar `buf` e configurar `buf.gen.yaml` pra geração de código Go a partir de `.proto`
  📚 Estudar: Buf — lint de proto, breaking change detection, geração de código
  *(feito em 2026-09-10 — `service-template/proto/` com `buf.yaml` (lint `STANDARD`) + `buf.gen.yaml`, gera `example.v1` pra `internal/pb/`. Mesmo padrão aplicado no `auth-service`, migrado de `DEFAULT` pra `STANDARD` e de `auth` pra `auth.v1` no mesmo lote de trabalho)*

---

## 🔐 Fase 3 — Primeiro serviço (auth-service)

### Modelagem e proto
- [x] Definir entidades mínimas: `User`, `Role` (ex: admin, operador, financeiro — só o suficiente pro MVP)
  *(já decidido em `docs/modelo-dados.md`: sem tabela Role separada, é campo string em User — admin/operador, ver RF-AUTH-1/RN1)*
- [x] Escrever `auth.proto` com RPCs: `Login`, `ValidateToken`, `CreateUser`
  📚 Estudar: protobuf schema evolution — regras de compatibilidade (campos novos sempre opcionais, nunca reaproveitar número de campo removido)
- [x] Gerar código Go a partir do proto com buf
  *(`buf.yaml`/`buf.gen.yaml` em `auth-service/proto/`, gera pra `internal/authpb/`; `go build ./...` compilando limpo)*

### Implementação
- [x] Clonar o service-template pra `auth-service`
- [x] Criar migration inicial (tabela `users`, `roles`)
  *(só `users` — sem tabela `roles` separada, `role` é coluna texto, já decidido em `docs/modelo-dados.md`. Aplicada de verdade contra o `auth-db` no cluster via `golang-migrate`)*
- [x] Implementar acesso a dados com GORM (create user, find by email, etc.)
  *(`internal/db/users.go` — `UserRepository.Create`/`FindByEmail`. Item ficou órfão sem marcar de uma sessão anterior — já estava implementado e coberto pelo teste de integração)*
- [x] Implementar hash de senha (bcrypt via stdlib-adjacent lib, ex: `golang.org/x/crypto/bcrypt`)
  *(`internal/auth/password.go` — HashPassword/CheckPassword. Testado de ponta a ponta: hash real gravado no `auth-db`, confirmado via psql)*
- [x] Implementar geração e validação de JWT
  *(`internal/auth/jwt.go` — GenerateToken/ValidateToken, HS256. Assinatura simétrica é suficiente porque só o auth-service assina e valida — os outros serviços nunca verificam token sozinhos, sempre via RPC ValidateToken, RF-AUTH-3. Testado gerando e validando de volta, claims batendo)*
  📚 Estudar: JWT — claims padrão (exp, iat, sub), assinatura HS256 vs RS256, onde guardar a chave secreta
- [x] Implementar RPC `Login` (valida credenciais, retorna token)
- [x] Implementar RPC `ValidateToken` (usado pelos outros serviços via gRPC)
- [x] Implementar RPC `CreateUser` (admin cria novos usuários)
  *(`internal/grpcserver/server.go` — os três implementados e testados de ponta a ponta via grpcurl contra o servidor rodando de verdade: CreateUser → Login → ValidateToken, token e claims batendo. Servidor gRPC roda numa goroutine junto do HTTP de healthcheck/metrics, cfg.GRPCPort)*

### Qualidade e observabilidade
- [x] Escrever testes unitários das regras de negócio (hash, validação de token)
  *(`internal/auth/password_test.go` e `jwt_test.go` — cobre hash/verificação de senha e geração/validação de token, incluindo o caso de segredo errado)*
- [x] Escrever teste de integração do fluxo de login (contra banco real via testcontainers ou docker-compose)
  *(`internal/grpcserver/login_test.go` — sobe Postgres real via testcontainers, roda a migration de verdade, cria usuário e chama `server.Login` direto. Debugou um problema clássico de timing: Postgres reinicia sozinho na primeira subida, precisa esperar a 2ª ocorrência do log "ready to accept connections")*
  📚 Estudar: testcontainers-go — como subir Postgres descartável pra teste de integração
- [x] Validar que métricas Prometheus aparecem no Grafana pro auth-service
  *(feito em 2026-09-08 — datasource Prometheus já provisionado no Grafana (`http://prometheus:9090`); no Explore a query `promhttp_metric_handler_requests_total{job="auth-service"}` plota série do pod deployado. Depende do job de scrape configurado no item da Fase 2)*
- [x] Validar que traces do auth-service aparecem no backend de tracing configurado
  *(feito em 2026-09-14 — backend é Jaeger all-in-one, container Docker novo no servidor (`monitoring/docker-compose.yml`, junto do Prometheus/Grafana), UI em `192.168.1.44:16686`, OTLP gRPC na 4317. `OTEL_EXPORTER_OTLP_ENDPOINT` adicionado no ConfigMap do auth-service. Chamadas reais (`CreateUser`, `Login`) feitas contra o pod deployado via `grpcurl` — traces apareceram no Jaeger na hora, um span por RPC)*
- [x] Escrever README do serviço (o que faz, como rodar local, variáveis de ambiente)
  *(feito em 2026-09-08 — `auth-service/README.md`, em inglês: o que faz, os 3 RPCs, portas, tabela de env vars, passo a passo pra rodar local (Postgres + migrate + make run), make targets, testes, `buf generate`, ponteiro pro deploy)*

### Deploy
- [x] Escrever manifests k8s específicos do auth-service (a partir dos genéricos da Fase 2)
  *(feito em 2026-09-08 — `deploy/auth-service/` com configmap (ENVIRONMENT/PORT/GRPC_PORT), service (portas http 8080 + grpc 9090), deployment (envFrom config+secret, probes /healthz e /readyz, `imagePullPolicy: IfNotPresent`) e secret.yaml como molde comentado. Commit `e35ad2d`)*
- [x] Deployar auth-service no cluster e validar healthcheck respondendo
  *(feito em 2026-09-08 — primeiro deploy do auth-service no k3s. Secret `auth-service-secret` criado fora do git (`JWT_SECRET` + `DATABASE_URL` apontando pro Service `auth-db`). Pipeline manual do `docs/deploy.md`: build → save → scp → `k3s ctr images import` → `kubectl apply`. Pod 1/1 Running, `/healthz` e `/readyz` = 200)*
- [x] Testar login end-to-end via `grpcurl` contra o serviço deployado
  *(feito em 2026-09-08 — servidor sem reflection, então `grpcurl -import-path auth-service/proto -proto auth.proto`. Fluxo CreateUser → Login → ValidateToken contra o pod: token gerado e validado, `valid: true` com `userId`/`role` batendo. Usuário de teste `e2e@mtv.com.br` ficou no `auth-db` do cluster)*

---

## 📦 Fase 4 — Serviços seguintes (clonando o template)

### catalog-service
- [x] Clonar template pra `catalog-service`
  *(feito em 2026-09-15 — `cp -r service-template catalog-service` + módulo renomeado (`mtv-erp/catalog-service`) em todos os `.go` via `sed`, proto de exemplo removido, `go build`/`vet` limpos)*
- [x] Modelar entidade `Product` (tipo/marca de arroz — ex: tipo 1, tipo 2, parboilizado, integral)
  *(`internal/db/models.go`/`products.go` — `id`, `name`, `active` (soft delete, RF-CAT-2). `ProductRepository.Create`/`ListActive`/`Deactivate`, sem `Delete` de verdade)*
- [x] Modelar entidade `UnitOfMeasure` com os valores definidos na Fase 1 (fardo 30kg, fardo 10kg, pacote 5kg, pacote 1kg, saco 25kg, saco 50kg, saco 60kg, granel) e campo `peso_base_kg`
  *(`ConversionFactorKg` como `decimal.Decimal` (shopspring/decimal, não float64 — evita erro de arredondamento), coluna `NUMERIC(12,4)`. Sem método de update no repositório — trava depois de criado é garantido pela ausência do caminho de código, não por constraint de banco, conforme RF-CAT-6/RN2)*
- [x] Implementar mecanismo de conversão: função/método que recebe quantidade + unidade e retorna quantidade em kg (e o inverso, kg → unidade de venda)
  *(`internal/db/conversion.go` — `UnitOfMeasure.ToKg`/`FromKg`, usando `.Mul`/`.Div` do decimal. 3 testes unitários: caso normal, fração, e ida-e-volta `FromKg(ToKg(x)) == x` com valor fracionário — prova que a precisão do decimal se mantém)*
- [x] Escrever proto `catalog.proto` (RPCs: CRUD de Product, CRUD/list de UnitOfMeasure, RPC de conversão)
  *(feito em 2026-09-16 — `proto/catalog/v1/catalog.proto`, pacote `catalog.v1` desde o início (sem precisar migrar depois, como o auth). `CreateProduct`/`ListProducts`/`DeactivateProduct`, `CreateUnitOfMeasure`/`ListUnitsOfMeasure`, `ConvertToKg`/`ConvertFromKg` espelhando os métodos Go. Fatores decimais trafegam como `string` no proto — protobuf não tem tipo decimal nativo, evita perder a precisão do `shopspring/decimal` na rede. `buf lint` limpo, gerado em `internal/pb/catalog/v1/`)*
- [x] Migrations + modelos GORM (products, units_of_measure)
  *(`migrations/000001_create_products_table.*` e `000002_create_units_of_measure_table.*`, aplicadas e revertidas contra Postgres real — `up`/`down` testados)*
- [x] Implementar RPCs e testes unitários (especial atenção nos testes de conversão de unidade — casos de borda tipo fração de kg)
  *(feito em 2026-09-16 — `internal/grpcserver/server.go`: os 7 RPCs (`CreateProduct`, `ListProducts`, `DeactivateProduct`, `CreateUnitOfMeasure`, `ListUnitsOfMeasure`, `ConvertToKg`, `ConvertFromKg`), registrados no `main.go`. Achado no caminho: `UnitOfMeasure` precisou de `TableName() string` — o GORM adivinha nome de tabela pluralizando a struct (`unit_of_measures`), que não batia com a migration (`units_of_measure`); `Product` só funcionou por coincidência. Teste de integração (`catalog_test.go`, testcontainers) cobre o fluxo completo create→convert→list→deactivate→list. Validado também na mão via `grpcurl` com reflection. Primeira adoção do `testify` (`assert`/`require`) no projeto — só em testes novos daqui pra frente, os antigos (auth-service, `conversion_test.go`) ficam com `testing` puro, decisão deliberada de não misturar refactor de estilo com trabalho novo)*
- [x] Seed de dados inicial (as unidades padrão já listadas, alguns produtos de exemplo)
  *(feito em 2026-09-16 — `migrations/000003_seed_catalog_data.{up,down}.sql`: as 8 unidades da Fase 1 (fardo 30/10kg, pacote 5/1kg, saco 25/50/60kg, granel com fator 1 — vende direto em kg) + 5 produtos de exemplo mais realistas que o mínimo (incluindo subprodutos: resíduo e farelo de arroz). `down` remove só as linhas por nome, não dá DROP — cada migration desfaz só o que ela própria fez. Testado up+down em sequência com as 3 migrations juntas)*
- [x] Deploy no k8s + validação de métricas/traces (deve ser mais rápido que o auth-service, já que o template está validado)
  *(feito em 2026-09-16 — `deploy/catalog-db/` e `deploy/catalog-service/` copiados de `auth-db`/`auth-service` via `cp`+`sed` (nomes trocados em lote). Imagem no ghcr.io (camadas reaproveitadas do `auth-service`, mesma base). Achado no caminho: senha gerada com `openssl rand -base64` continha `+`, que quebra dentro de uma DATABASE_URL — trocado pra `-hex` (só `0-9a-f`, sempre seguro em URL) e banco recriado. Pod rodando, 8 unidades + 5 produtos do seed confirmados via grpcurl contra o pod real, conversão testada (3×25=75kg). Prometheus com job novo (`catalog-service` → NodePort `30081`) confirmado `UP`; traces das chamadas reais aparecendo no Jaeger. Deploy inteiro (da imagem pronta até validado) bem mais rápido que o do auth-service, como esperado)*

*(reaberto em 2026-09-24 — `CATALOG_SUPPLIER` já estava desenhado em `docs/modelo-dados.md` desde a Fase 1, mas nunca virou tarefa aqui. Achado ao planejar o purchasing-service: RF-PUR-1/RN2 exige "compra deve referenciar um fornecedor cadastrado", e não existe cadastro de fornecedor em lugar nenhum do código. Bloqueia o purchasing-service, então entra antes dele.)*
- [x] Modelar entidade `Supplier` (fornecedor — `id`, `nome`, `documento`, `endereco`, `active`), campos simples conforme `docs/modelo-dados.md` (sem separar CNPJ/CPF/IE por tipo de pessoa, mais simples que a leitura literal de RF-CAT-4)
  *(feito em 2026-09-24 — `internal/db/models.go`/`suppliers.go`, mesmo molde do `Product`: `SupplierRepository.Create`/`ListActive`/`Deactivate`, sem `Delete` de verdade. Não precisou de `TableName()` — `Supplier` pluraliza certo pra `suppliers` sozinho)*
- [x] Migration + modelo GORM (`suppliers`)
  *(feito em 2026-09-24 — `migrations/000004_create_suppliers_table.{up,down}.sql`, mesmo estilo do `products`)*
- [x] Adicionar RPCs no `catalog.proto`: `CreateSupplier`, `ListSuppliers`, `DeactivateSupplier` (mesmo padrão de `Product` — soft delete, sem `Delete` de verdade)
  *(feito em 2026-09-24 — `buf lint` limpo, gerado em `internal/pb/catalog/v1/`)*
- [x] Implementar RPCs + testes (mesmo padrão de integração com testcontainers)
  *(feito em 2026-09-24 — os 3 RPCs em `internal/grpcserver/server.go`, `NewServer` agora recebe `supplierRepo` como terceiro argumento (`main.go` atualizado). Fluxo create→list→deactivate→list adicionado no `TestCatalogServiceIntegration` existente. Achado no caminho: o assert antigo do `Product` (`require.Len(listResp.Products, 1)`) estava quebrado desde que a migration de seed (000003) foi criada — banco de teste nunca foi realmente vazio, tinha 5 produtos de seed + o criado no teste = 6, não 1. Corrigido pra checar por ID específico (`containsProductID`) em vez de contagem total, resistente a seed data. Bug pré-existente, não relacionado ao Supplier, só nunca tinha sido notado porque o teste não rodava de novo desde 2026-09-16)*
- [x] Deploy: nova migration precisa rodar contra o banco já em produção (`catalog-db`), depois rebuild + rollout da imagem
  *(feito em 2026-09-24 — migration `000004` aplicada contra `catalog-db` via port-forward, imagem rebuildada/pushada, `rollout restart` (sem `apply` na pasta, pra não repetir a cilada do secret do inventory). `CreateSupplier` confirmado via `grpcurl` contra o pod real, `active: true`)*

*(reaberto de novo em 2026-09-24 — ainda no mesmo dia: pro `purchasing-service` validar `supplier_id`/`product_id` antes de lançar uma compra, precisa de `GetSupplier`/`GetProduct` por ID, que não existiam — só tinha `List`. Sem FK física entre bancos (ADR-0002), essa validação é responsabilidade da aplicação, não do banco.)*
- [x] Adicionar `FindByID` em `ProductRepository`/`SupplierRepository` + RPCs `GetProduct`/`GetSupplier` no `catalog.proto`, implementados com `codes.NotFound` (não erro genérico — adiantando o padrão que a Fase 5 já exige)
  *(feito em 2026-09-24 — testado no `TestCatalogServiceIntegration`: caso feliz (acha e confere os dados) + caso `NotFound` (UUID aleatório) pros dois. `buf lint` limpo)*
- [x] Redeploy do catalog-service com o `GetProduct`/`GetSupplier` novos (mesmo fluxo: rebuild + push + rollout restart, sem migration dessa vez — não mudou schema)
  *(feito em 2026-09-24 — `GetSupplier` confirmado via `grpcurl` contra o pod real: `NotFound` pra ID inválido, dados certos pro ID real)*

### inventory-service
- [x] Clonar template pra `inventory-service`
  *(feito em 2026-09-17 — mesmo processo do catalog-service: `cp -r` + módulo renomeado via `sed` em todos os `.go`, proto de exemplo removido, build/vet/test limpos)*
- [x] Modelar entidade `Lot` (código, safra, fornecedor de origem, data de recebimento, quantidade em kg, produto associado — produto referenciado por ID lógico do catalog-service)
  *(feito em 2026-09-17 — `internal/db/models.go`: `ProductID`/`PurchaseItemID` são referência lógica (outro serviço/banco, sem FK física, ADR-0002) — fornecedor não é campo direto, chega-se nele via `PurchaseItemID` → a compra, evita duplicar dado. `ReceivedAt` com `gorm:"type:date"` — só a data, sem hora, como o modelo pede)*
- [x] Modelar entidade `StockMovement` (tipo: entrada/saída, quantidade em kg, lote associado, timestamp, origem do movimento)
  *(`LotID` é FK **física** de verdade (`REFERENCES lots(id)` na migration) — diferente do `Lot`, porque `StockMovement` e `Lot` moram no mesmo banco. Testado: insert com `lot_id` inventado é rejeitado pelo Postgres, prova que a FK está ativa)*
- [x] Implementar validação de domínio explícita: **rejeitar qualquer movimentação de estoque sem lote associado** (regra de negócio, não só constraint de banco)
  *(feito em 2026-09-17 — `RegisterMovement` chama `lotRepo.FindByID` **antes** de gravar; se não achar, devolve `codes.NotFound` com mensagem clara, em vez de deixar a FK do Postgres estourar um erro genérico. Validado na mão via grpcurl)*
  📚 Estudar: onde colocar validação de invariante de domínio em Go — validação na camada de serviço vs constraint NOT NULL no banco (fazer as duas, mas a de domínio é a que dá erro de negócio claro)
- [x] Escrever proto `inventory.proto` (RPCs: CreateLot, RegisterMovement, GetStockByProduct, GetLotDetails)
  *(feito em 2026-09-17 — `proto/inventory/v1/inventory.proto`, pacote `inventory.v1` desde o início. `GetLotDetails` devolve o lote + todas as movimentações + saldo calculado (cobre RF-INV-5 de uma vez); `GetStockByProduct` só o total agregado em kg. Valores decimais como `string`, mesmo padrão do catalog. `buf lint` limpo, gerado em `internal/pb/inventory/v1/`)*
- [x] Migrations + modelos GORM (lots, stock_movements)
  *(migrations/000001_create_lots_table e 000002_create_stock_movements_table, aplicadas e revertidas contra Postgres real, nessa ordem por causa da FK. `LotRepository` (Create/FindByID) e `StockMovementRepository` (Create/ListByLot))*
- [x] Implementar RPCs com a validação de lote obrigatório
  *(feito em 2026-09-17 — os 4 RPCs (`CreateLot`, `RegisterMovement`, `GetStockByProduct`, `GetLotDetails`) em `internal/grpcserver/server.go`, registrados no `main.go`. Decisão de design: `quantity_kg` da movimentação é **assinado** (positivo entrada, negativo saída/ajuste-pra-baixo) — saldo é soma direta, `type` fica só como metadado descritivo, não determina o sinal. Testado de ponta a ponta via grpcurl contra Postgres real: CreateLot → 2 movimentações → saldo 700 (1000-300) certo em `GetLotDetails` e `GetStockByProduct`)*
- [x] Testes unitários da regra "sem lote não existe estoque" (caso de erro esperado)
  *(feito em 2026-09-23 — `internal/grpcserver/inventory_test.go`, `TestRegisterMovement_SemLoteFalha`, mesmo padrão do catalog-service: testcontainers + Postgres real + migrations aplicadas de verdade. Confirma que `RegisterMovement` com `lot_id` inexistente devolve `codes.NotFound`, não erro genérico)*
- [x] Testes de consulta de saldo de estoque agregado por produto (soma de lotes)
  *(feito em 2026-09-23 — `TestGetStockByProduct_SomaVariosLotes`: 2 lotes do mesmo produto, movimentações em cada um, confirma que `GetStockByProduct` soma através dos lotes (1000 - 300 + 500 = 1200), não só dentro de um lote)*
- [x] Deploy no k8s + validação de métricas/traces
  *(feito em 2026-09-23 — `deploy/inventory-db/` e `deploy/inventory-service/` copiados de `catalog-db`/`catalog-service` via `cp`+`sed`, NodePort de métricas `30082`. Achado no caminho: `kubectl apply -f deploy/inventory-service/` aplica a pasta inteira, inclusive o `secret.yaml` (que é só molde de exemplo com senha `SENHA`) — isso sobrescreveu o Secret real criado na mão e quebrou a conexão com o banco (`password authentication failed`). Corrigido recriando o secret e documentado o aviso em `docs/deploy.md` (aplicar arquivos específicos, nunca a pasta inteira, quando o secret real já existe). Pod `Running` 1/1, `/healthz` OK, job `inventory-service` adicionado no `prometheus.yml` do servidor (`172.18.0.1:30082`) e confirmado `UP` em `/targets`)*

### purchasing-service
*(atualizado em 2026-08-24 — sem "pedido de compra" como entidade separada, ver RF-PUR-1 em `docs/requisitos.md`: compra é lançada num único passo)*
*(corrigido em 2026-09-24 — a descrição abaixo estava simplificada demais: RF-PUR-1/RN1 é "fornecedor + **um ou mais** produtos", não um produto só. `docs/modelo-dados.md` já desenhava isso como duas entidades desde a Fase 1 — `Purchase` (cabeçalho) 1:N `PurchaseItem` (linha por produto) — e o `inventory-service.Lot.PurchaseItemID` já pressupõe isso. Fornecedor agora existe de verdade: `Supplier` no catalog-service, feito hoje)*
- [x] Clonar template pra `purchasing-service`
  *(feito em 2026-09-24 — `cp -r` + módulo renomeado via `sed` em todos os `.go`, proto/pb de exemplo removidos, build/vet limpos. Achado no caminho: o `sed` do módulo (`mtv-erp/service-template` → `mtv-erp/purchasing-service`) não pega strings soltas sem o prefixo `mtv-erp/` — sobraram duas em `main.go` (`InitTracer(ctx, "service-template")` e o `slog.Info` de startup) que não são erro de compilação, só ficariam com o nome errado no Jaeger/logs. Corrigido na mão. Vale conferir isso da próxima vez que clonar o template também)*
- [ ] Modelar entidade `Purchase` (cabeçalho da compra — `id`, `supplier_id` (ref lógica pro `Supplier` do catalog-service), `invoice_number`, `invoice_date`, `invoice_value`)
- [ ] Modelar entidade `PurchaseItem` (linha da compra — `id`, `purchase_id` (FK física, mesmo banco), `product_id`/`unit_id` (ref lógica pro catalog-service), `quantity`), FK física `purchase_id → purchases(id)`
- [ ] Constraint de unicidade `(supplier_id, invoice_number)` — RF-PUR-1/RN6, bloqueia lançar a mesma nota fiscal duas vezes
- [ ] Escrever proto `purchasing.proto` (RPC: `CreatePurchase`, recebendo fornecedor + nota fiscal + lista de itens de uma vez)
- [ ] Migrations + modelos GORM (purchases, purchase_items)
- [ ] Implementar lógica de lançamento: `CreatePurchase` grava `Purchase` + `PurchaseItem`s, converte cada item pra kg (RN4 — chama `ConvertToKg` do catalog-service) e chama (via gRPC síncrono, por enquanto) o inventory-service `CreateLot` **uma vez por item**, antes de confirmar a compra
- [ ] Testes unitários e de integração do fluxo compra lançada → lote(s) criado(s) no inventory (cobrir caso de múltiplos itens numa mesma compra)
- [ ] Deploy no k8s + validação de métricas/traces

---

## 🌐 Fase 5 — API Gateway + Frontend

### api-gateway (BFF REST)
*(planejamento alinhado em 2026-09-23 com a checklist da disciplina — Aulas 1 e 2, APIs REST em Go. A arquitetura em camadas `handler→service→repository` da aula não se aplica 1:1 aqui: pelo ADR-0001 o gateway não tem lógica de negócio nem banco próprio, então vira `handler REST → client gRPC` direto, sem camada de service/repository — divergência intencional, não omissão.)*
- [ ] Clonar template pra `api-gateway`, mas reescrever o essencial: trocar `grpc.NewServer` (servidor principal do template) por um servidor `net/http` com `go-chi/chi/v5` como router — recomendação da disciplina, 100% compatível com `http.Handler` (ao contrário de gin/echo, que têm API própria; gorilla/mux tem manutenção reduzida). `internal/grpcserver/` e `internal/db/` do template não se aplicam ao gateway
- [ ] Adicionar `github.com/go-chi/chi/v5` ao `go.mod`; montar `chi.NewRouter()` na raiz com middlewares globais via `r.Use`: `middleware.Logger`, `middleware.Recoverer`, `middleware.RequestID`, `middleware.Timeout`, `middleware.AllowContentType`
- [ ] Versionar a API com `r.Mount("/api/v1", subRouter)` — faltava no planejamento original, a disciplina exige prefixo de versão
- [ ] Definir rotas REST do gateway mapeando pros RPCs dos 4 serviços (auth, catalog, inventory, purchasing), agrupadas por recurso com `r.Route("/produtos", ...)` + sub-grupo `/{id}`; parâmetros de rota via `chi.URLParam(r, "id")`, query params via `r.URL.Query().Get()`
  📚 Estudar: tradução verbo-RPC → substantivo-REST — RPCs como `CreateProduct`/`ListProducts`/`DeactivateProduct` não viram rota 1:1 por nome; ex: `DeactivateProduct` vira `PATCH /produtos/{id}`, não uma rota própria. URIs sempre substantivo no plural, sem verbo (`GET /produtos`, nunca `/getProdutos`)
- [ ] Implementar tradução REST → gRPC (handlers HTTP chamando clients gRPC internos direto, sem camada de service/repository intermediária — ver nota acima)
  📚 Estudar: padrão BFF (Backend For Frontend) — por que o gateway não deveria ter lógica de negócio própria
- [ ] **Pré-requisito, adiantado da Fase 7 (2026-09-23):** revisar auth/catalog/inventory/purchasing e garantir que cada erro de negócio devolve `status.Error(codes.X, ...)` com o código certo (`NotFound`, `InvalidArgument`, `AlreadyExists`, `FailedPrecondition`), não só `return nil, err` cru — sem isso o gateway não tem o que traduzir e todo erro de negócio vira 500 genérico no REST
- [ ] Implementar tradução de status gRPC → HTTP com helper padronizado `writeError(w, status, msg)` retornando `{"error": "..."}` (nunca `http.Error` texto puro): `NotFound`→404, `InvalidArgument`→400, `FailedPrecondition`/`AlreadyExists`→409/422, erro inesperado→500 sempre logado
- [ ] Definir DTOs REST próprios (structs com `json:"campo"` + `omitempty`) em vez de serializar direto as structs geradas do protobuf; listas vazias retornam `[]` (`make([]T, 0, n)`), nunca `null`
- [ ] Implementar middleware de autenticação no gateway (valida JWT via chamada ao auth-service), no formato `func(http.Handler) http.Handler`, registrado só nas rotas que exigem login (não no router raiz)
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
*(evento renomeado em 2026-08-24 de `MercadoriaRecebida` pra `CompraLançada`, refletindo o lançamento em passo único — ver RF-PUR-1. Atualizado em 2026-08-27: mensageria trocada de NATS pra RabbitMQ, ver ADR-0009 — motivo de aprendizado/currículo, não técnico.)*
- [ ] Definir schema do evento `CompraLançada` (payload: lote, produto, quantidade em kg, timestamp)
  📚 Estudar: versionamento de eventos em mensageria — como evoluir o schema de um evento sem quebrar consumidores antigos
- [ ] Trocar a chamada síncrona gRPC de lançamento de compra (Fase 4) por publish do evento `CompraLançada` via RabbitMQ no purchasing-service
  📚 Estudar: exchanges e filas no RabbitMQ (direct/topic/fanout) — qual tipo de exchange faz sentido pra esse evento
- [ ] Implementar subscriber do evento no inventory-service (cria o lote a partir do evento recebido)
- [ ] Implementar idempotência no consumidor (mesmo evento processado duas vezes não duplica lote)
  📚 Estudar: idempotência em consumidores de mensageria — chave de deduplicação, at-least-once delivery do RabbitMQ
- [ ] Testar cenário de falha: inventory-service fora do ar durante publish — validar que a mensagem não se perde (fila durável + publisher confirms)
  📚 Estudar: filas duráveis e publisher confirms do RabbitMQ — equivalente ao que JetStream resolveria no NATS
- [ ] Documentar o fluxo assíncrono resultante num ADR-0007: "Lançamento de compra via evento assíncrono (RabbitMQ)"

### Observabilidade da mensageria
- [ ] Propagar trace context através do evento RabbitMQ (correlacionar span do publish com o do consumo)
- [ ] Adicionar métrica de mensagens processadas/falhas no inventory-service

---

## ✅ Fase 7 — Testes end-to-end e polimento

- [ ] Escrever teste E2E do fluxo completo: login → cadastro de produto → criação de pedido de compra → recebimento → evento assíncrono → lote criado → consulta de estoque
- [ ] Revisar mensagens de erro consistentes em todos os serviços (texto, não código — o código gRPC apropriado já foi adiantado pra Fase 5, é pré-requisito do gateway)
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
- [GORM](https://gorm.io/docs/) — ORM em Go, mapeamento de structs e queries
- [RabbitMQ Docs](https://www.rabbitmq.com/docs) — mensageria (exchanges, filas, publisher confirms)
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/) — instrumentação e tracing distribuído
- [Focus NFe — Documentação da API](https://focusnfe.com.br/doc/) — emissão de NFe via API terceira
- [Nuxt 3](https://nuxt.com/docs) — framework Vue
- [PrimeVue](https://primevue.org/) — biblioteca de componentes
- [Naive UI](https://www.naiveui.com/) — biblioteca de componentes (alternativa)
- [VitePress](https://vitepress.dev/) — portal de documentação
- [Buf Docs](https://buf.build/docs/) — lint, breaking change detection e geração de docs de proto
- [golang-migrate](https://github.com/golang-migrate/migrate) — migrations de banco
- [ADR - Documenting Architecture Decisions (Michael Nygard)](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions) — formato original de ADR
