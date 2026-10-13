# Roteiro de apresentação: primeiro microsserviço (api-gateway + catalog-service)

Data: **quinta, 15/10/2026**. Formato: **sem slides**. Abrir o software, mostrar o código e testar ao vivo com `curl`.

Esse documento é a "cola" da apresentação. Cada bloco diz **o que abrir**, **o que falar** (2 ou 3 frases, não precisa decorar, só entender) e **o que rodar**. A ordem foi pensada pra contar uma história: primeiro funciona, depois como está organizado, depois o que acontece quando dá errado.

Tempo estimado: **12 a 15 minutos**.

---

## 0. Antes de começar (10 minutos antes)

Deixe tudo aberto e testado. A demo não pode depender de "deixa eu abrir aqui".

- [ ] **Terminal 1, o túnel.** Fica aberto o tempo todo:
  ```
  kubectl port-forward -n mtv-erp svc/api-gateway 8080:8080
  ```
- [ ] **Terminal 2, os `curl`.** É onde você digita durante a demo.
- [ ] **Terminal 3, dentro de `api-gateway/`.** É pra rodar os testes.
- [ ] Conferir se os pods estão de pé:
  ```
  kubectl get pods -n mtv-erp
  ```
  `api-gateway` e `catalog-service` devem estar `1/1 Running`.
- [ ] Rodar **um** `curl` de aquecimento, pra confirmar que o túnel funciona:
  ```
  curl -i localhost:8080/api/v1/products
  ```
- [ ] VS Code aberto na raiz do projeto, com estas abas já abertas, nesta ordem:
  1. `api-gateway/cmd/server/main.go`
  2. `api-gateway/internal/config/config.go`
  3. `api-gateway/internal/handlers/handlers_products.go`
  4. `api-gateway/internal/json/grpcerror.go`
  5. `api-gateway/internal/clients/viacep/viacep.go`
  6. `api-gateway/internal/handlers/handlers_addresses.go`
  7. `catalog-service/internal/db/products.go`
  8. `catalog-service/internal/grpcserver/server.go`
  9. `api-gateway/internal/handlers/handlers_products_test.go`
- [ ] Aumentar a fonte do terminal e do VS Code (`Ctrl +`). Quem está assistindo precisa ler.

> ⚠️ **Pegadinha:** se precisar rebuildar alguma coisa, confira no prompt se está na pasta certa antes do `docker build` (aconteceu em 13/10: o build feito de dentro de `catalog-service/` empacotou o serviço errado).

---

## 1. Abertura: o que é e como se encaixa (1 min)

**Abrir:** nada ainda, só falar. Se quiser, mostrar a árvore de pastas na lateral do VS Code.

**Falar:**
> "É um ERP pra uma empresa de arroz, em microsserviços Go. O que eu vou mostrar é a **porta de entrada REST**, o `api-gateway`, junto com o serviço que ele consulta, o `catalog-service`, que cuida de produtos e fornecedores e tem banco próprio. O gateway recebe HTTP/JSON e conversa com o catalog por gRPC. Ele também consulta uma API externa, o ViaCEP. Tudo roda num cluster Kubernetes de verdade, no servidor da empresa."

Desenho mental pra explicar (pode desenhar no ar ou no quadro, se tiver):
```
curl ──HTTP/JSON──▶ api-gateway ──gRPC──▶ catalog-service ──SQL──▶ catalog-db (Postgres)
                         │
                         └──HTTP/JSON──▶ ViaCEP (API externa)
```

**Se perguntarem "por que o gateway não tem banco?":**
> "Porque ele não é dono de nenhum dado. Ele só traduz REST pra gRPC. Quem é dono do produto é o catalog, e cada serviço tem o próprio banco (database-per-service). Se o gateway tivesse banco, ia duplicar dado que pertence a outro serviço."

---

## 2. Funcionando: caminho feliz (2 min)

**Rodar** (terminal 2):

```
curl -i -X POST localhost:8080/api/v1/products -H "Content-Type: application/json" -d '{"name":"Arroz demo"}'
```
→ **201 Created**. Copie o `id`.

```
curl -i localhost:8080/api/v1/products/ID_AQUI
```
→ **200**.

```
curl -i -X PATCH localhost:8080/api/v1/products/ID_AQUI -H "Content-Type: application/json" -d '{"active":false}'
```
→ **200**, com `"active":false`.

```
curl -i localhost:8080/api/v1/addresses/01001-000
```
→ **200**, com o endereço da Praça da Sé, vindo do ViaCEP.

**Falar:**
> "Criação devolve 201, não 200, porque criou um recurso. Desativar é PATCH, e não DELETE, porque produto nunca é apagado (é regra de negócio: tem lote e compra antiga apontando pra ele). É só uma atualização parcial do campo `active`."

---

## 3. Critério 1: arquitetura e organização

### 1.1 Estrutura de pastas

**Abrir:** a árvore de `api-gateway/` na lateral.

**Falar**, apontando cada pasta:
| Pasta | Papel |
|---|---|
| `cmd/server/` | ponto de entrada: monta tudo e sobe o servidor |
| `internal/handlers/` | rotas e regras de entrada (validação 400/422) |
| `internal/clients/viacep/` | client da API externa |
| `internal/pb/` | client gRPC do catalog, gerado do `.proto` |
| `internal/models/` | DTOs: o formato do JSON da API |
| `internal/json/` | helpers de resposta e de tradução de erro |
| `internal/config/` | leitura das variáveis de ambiente |

> "O acesso a dados não está aqui. Ele está no catalog, em `internal/db/`, que é o repository. Cada serviço só tem as camadas de que precisa."

**Abrir:** [`main.go`](../api-gateway/cmd/server/main.go), linhas 53 a 80.

> "Aqui fica a montagem: os middlewares globais (log, recuperação de panic, request ID, **timeout de 10 segundos**), a versão `/api/v1` e as rotas, agrupadas por recurso. Os clients são criados uma vez só e passados pros handlers."

### 1.2 Configuração por variável de ambiente

**Abrir:** [`config.go`](../api-gateway/internal/config/config.go) e depois `deploy/api-gateway/configmap.yaml`.

> "Porta, endereço do catalog e URL do ViaCEP vêm de variável de ambiente. No Kubernetes, elas vêm desse ConfigMap. Se uma variável obrigatória faltar, o serviço **nem sobe**: falha na largada, com mensagem clara, em vez de quebrar no meio de uma requisição."

### 1.3 go.mod

**Abrir:** `api-gateway/go.mod`.

> "O módulo é `mtv-erp/api-gateway`. As dependências diretas são só o que é usado: chi (router), grpc/protobuf (catalog), uuid (validar ID), prometheus/otel (observabilidade) e testify (testes). Mantido limpo com `go mod tidy`."

---

## 4. Critério 3: comunicação REST

### 3.2 e 3.3: rotas, status e JSON

**Abrir:** [`handlers_products.go`](../api-gateway/internal/handlers/handlers_products.go), na função `HandleCreateProduct`.

**Falar**, seguindo o código de cima pra baixo:
> "Todo handler segue a mesma ordem: (1) lê e decodifica o JSON: se estiver quebrado, **400**; (2) valida a regra de negócio: nome vazio, **422**; (3) só então chama o catalog. Entrada ruim nunca gasta chamada de rede."

**Rodar:**
```
curl -i -X POST localhost:8080/api/v1/products -H "Content-Type: application/json" -d '{"name":'
```
→ **400**, JSON inválido.

```
curl -i -X POST localhost:8080/api/v1/products -H "Content-Type: application/json" -d '{"name":"   "}'
```
→ **422**, nome obrigatório.

```
curl -i localhost:8080/api/v1/products/abc
```
→ **400**, ID não é UUID.

```
curl -i localhost:8080/api/v1/products/00000000-0000-0000-0000-000000000000
```
→ **404**, não existe.

**Falar, a diferença entre os códigos** (é pergunta provável):
> "**400**: a requisição está mal formada (JSON quebrado, ID que nem é UUID). **422**: está bem formada, mas viola regra de negócio (nome vazio). **404**: está tudo certo, só que o recurso não existe. **500**: falhou do nosso lado."

**Abrir:** `api-gateway/internal/models/models.go`.

> "O JSON tem formato próprio, os DTOs. Eu não serializo direto a struct do protobuf, então se o contrato interno mudar, a API pública não muda junto. E lista vazia sai `[]`, nunca `null`."

### 3.1: API externa (ViaCEP)

**Abrir:** [`viacep.go`](../api-gateway/internal/clients/viacep/viacep.go), na função `Lookup` (linha 46).

**Falar**, apontando cada linha:
> "Linha 52: `context.WithTimeout`. A chamada tem prazo próprio de 5 segundos, e se o ViaCEP travar, a gente desiste. O `defer cancel()` libera o timer."
>
> "Linha 68: `defer resp.Body.Close()`. A resposta é uma conexão aberta, e se não fechar, vaza conexão. O `defer` garante que fecha, qualquer que seja o `return`."
>
> "Linhas 70 a 75: confiro o `StatusCode` antes de assumir sucesso."
>
> "Linha 81, **o detalhe mais interessante:** o ViaCEP responde **200 mesmo quando o CEP não existe**, com `{"erro": "true"}` no corpo. Então conferir só o status não basta. Eu detecto pelo campo `cep` vazio."

**Rodar:**
```
curl -i localhost:8080/api/v1/addresses/123
```
→ **400**, formato inválido. Esse nem chega a chamar o ViaCEP.

```
curl -i localhost:8080/api/v1/addresses/99999999
```
→ **404**, o caso do "200 com erro".

---

## 5. Critério 4: erro de negócio × erro de infra

**Abrir:** [`viacep.go`](../api-gateway/internal/clients/viacep/viacep.go), linhas 15 a 20 (os erros sentinela), e depois [`handlers_addresses.go`](../api-gateway/internal/handlers/handlers_addresses.go), na função `writeAddressError` (linha 40).

**Falar:**
> "O client devolve 4 erros com nome. Dois são de **negócio**, CEP inválido e CEP não encontrado: a culpa é do dado, então é 400 e 404 e não precisa logar. Dois são de **infra**, timeout e ViaCEP fora do ar: a culpa é de quem a gente chamou, então é **504 e 502**, com log, e a mensagem interna não vaza pro cliente. O handler separa com `errors.Is`."

**Abrir:** [`grpcerror.go`](../api-gateway/internal/json/grpcerror.go).

> "Pro catalog é a mesma ideia, com os códigos do gRPC: `NotFound` vira 404 e `InvalidArgument` vira 400, que são negócio. `Unavailable` (catalog fora do ar) vira **503** e `DeadlineExceeded` vira **504**, que são infra. Qualquer outro vira 500 com log."

### Demo opcional: derrubar o catalog ao vivo

> ⚠️ Ensaie antes. É a parte mais impressionante, mas é a única que mexe no cluster.

```
kubectl scale deployment/catalog-service -n mtv-erp --replicas=0
```
```
curl -i localhost:8080/api/v1/products
```
→ **503**, serviço temporariamente indisponível. O gateway continua de pé e diz **honestamente** que a dependência caiu.

**Restaurar (não esquecer):**
```
kubectl scale deployment/catalog-service -n mtv-erp --replicas=1
```
```
kubectl rollout status deployment/catalog-service -n mtv-erp
```

Se preferir não arriscar, mostre esse mesmo cenário no teste (`catalog fora do ar` → 503, na seção 8).

---

## 6. Critério 2: persistência própria (no catalog-service)

**Falar:**
> "Persistência fica no catalog, que é o dono do dado."

### 2.1 Banco próprio

**Abrir:** a pasta `deploy/catalog-db/` (o StatefulSet do Postgres) e `catalog-service/migrations/`.

> "O catalog tem o próprio Postgres, o `catalog-db`, que nenhum outro serviço acessa. Cada serviço do projeto tem o seu (auth-db, inventory-db, purchasing-db). O schema é versionado com migrations do `golang-migrate`. Não uso o `AutoMigrate` do GORM, pra não depender de inferência de schema em produção."

### 2.2 Repository isolado

**Abrir:** [`products.go`](../catalog-service/internal/db/products.go).

> "Essa é a única camada que fala com o banco. O handler gRPC nunca escreve SQL, ele só chama `Create`, `FindByID`, `ListActive` e `Deactivate`."

### 2.3 Não encontrado ≠ falha de banco

**Abrir:** [`products.go`](../catalog-service/internal/db/products.go), na função `Deactivate` (linha 34).

> "Pegadinha que eu achei: um `UPDATE` com um ID que não existe **não dá erro** no Postgres, só atualiza 0 linhas. Então eu olho o `RowsAffected`: se for zero, devolvo `ErrRecordNotFound`. Se o `Error` vier preenchido, é falha de banco de verdade. São dois caminhos diferentes."

**Abrir:** [`server.go`](../catalog-service/internal/grpcserver/server.go), na função `GetProduct` (linha 219).

> "No handler gRPC, só o `ErrRecordNotFound` vira `NotFound`. Qualquer outro erro vira `Internal`. Antes, tudo virava 'não encontrado', e se o banco caísse, o usuário ouviria 'produto não existe', o que é mentira. E o `validateID` no topo barra ID mal formado antes de ir ao banco."

---

## 7. Critério 5: qualidade do código

### 5.1 Tratamento de erro idiomático

> "Todo erro é tratado com `if err != nil`, e quando eu repasso, embrulho com `%w` pra não perder a causa. Os `_` que aparecem no código descartam **valores**, nunca erros: por exemplo, a resposta vazia do `DeactivateProduct`, ou o UUID convertido que eu só uso pra validar o formato."

### 5.2 Interfaces

**Abrir:** [`handlers_addresses.go`](../api-gateway/internal/handlers/handlers_addresses.go), linhas 16 a 18 (a interface `AddressLookup`).

> "O handler não depende do client concreto do ViaCEP. Ele depende de uma interface com um método, `Lookup`. Em produção, o client real encaixa nela. No teste, um fake. A interface fica no pacote de quem **usa**, que é o costume do Go. O client do catalog também é uma interface, gerada pelo gRPC."

### 5.3 Nomes consistentes

> "Código e campos JSON em inglês. Mensagens pro usuário em português. Os arquivos seguem o padrão `handlers_<recurso>.go` e o teste do lado, `handlers_<recurso>_test.go`."

---

## 8. Critério 6: testes

**Rodar** (terminal 3, dentro de `api-gateway/`):
```
go test ./internal/handlers/ -v
```

São 22 subtestes, todos `PASS`. Os `ERROR` no meio da saída **são esperados**: são os logs dos casos de infra, provando que erro de infra é logado.

**Abrir:** [`handlers_products_test.go`](../api-gateway/internal/handlers/handlers_products_test.go).

**Falar:**
> "Testes em tabela: cada linha é um caso, com entrada e resultado esperado. O catalog é substituído por um **fake**, então não precisa de rede nem de banco. O fake embute a interface e sobrescreve só os métodos usados."
>
> "O mais importante é o `wantCalled`: nos casos 400 e 422, o teste **prova** que o catalog nem foi chamado. A validação barra antes."
>
> "No teste do endereço, o fake simula o ViaCEP lento e fora do ar, coisa que eu não consigo forçar no ViaCEP real, e confere 504 e 502."

**Se perguntarem do catalog:**
> "O catalog tem teste de integração com **testcontainers**: sobe um Postgres de verdade num container, roda as migrations e testa o fluxo completo, incluindo os casos de NotFound e ID inválido."

---

## 9. Fechamento (30s)

> "Resumindo: o gateway expõe REST com validação em camadas, consulta o catalog por gRPC e uma API externa por HTTP, e separa o que é erro do usuário do que é erro de infra. Cada um tem o código HTTP certo. A persistência fica no catalog, com banco próprio e repository isolado. Próximos passos: autenticação JWT no gateway, rotas de estoque e compras, e documentação OpenAPI."

---

## Perguntas prováveis (e respostas curtas)

**Por que PATCH e não PUT?**
> PUT substitui o recurso inteiro. PATCH atualiza parte dele. Desativar só muda o `active`.

**Por que não DELETE?**
> Produto nunca é apagado. Lote e compra antiga referenciam ele, e a rastreabilidade precisa do histórico.

**Por que não o soft delete do GORM (`DeletedAt`)?**
> Ele **esconde** o registro de todas as consultas. Produto inativo precisa continuar visível no histórico, só não deve aparecer na listagem do catálogo. O `active` deixa o filtro explícito, só no `ListActive`.

**Por que gRPC entre os serviços e REST pra fora?**
> REST/JSON é universal pra quem consome de fora (frontend, curl). Entre serviços, o gRPC dá contrato tipado (o `.proto`), código gerado e é mais eficiente.

**O que é um middleware?**
> Uma função que embrulha o handler e roda antes e depois dele em toda requisição. Exemplos: log, recuperação de panic, timeout. No chi, tem o formato `func(http.Handler) http.Handler`.

**Por que DTO separado da struct do protobuf?**
> Desacopla o contrato público (JSON) do interno (proto). Mudar o proto não quebra quem consome a API.

**Por que 502/504 e não 500 quando o ViaCEP cai?**
> O 500 diz "eu falhei". O 502/504 diz "um serviço que eu chamei falhou ou demorou". É mais honesto e ajuda a achar o problema.

**Por que o timeout está no context?**
> O context carrega o prazo por todas as camadas. O middleware dá 10s pra requisição inteira, e o client do ViaCEP põe 5s só na chamada externa. Vale o que vencer primeiro.

**O que acontece se o ID não for UUID?**
> O gateway barra com 400 antes de chamar o catalog (`ParseId`). E o catalog também confere (`validateID` → `InvalidArgument`), pra proteger quem chama ele direto, sem passar pelo gateway.

---

## Plano B (se a rede ou o cluster falharem no dia)

1. **Os testes rodam offline**, e já mostram quase todos os critérios:
   ```
   go test ./internal/handlers/ -v
   ```
2. **Rodar o gateway local.** Antes, feche o port-forward do terminal 1, senão a porta 8080 fica ocupada. A rota do ViaCEP funciona sem o catalog (testado em 13/10). O `grpc.NewClient` não conecta na subida, só na primeira chamada. Dentro de `api-gateway/`:
   ```
   CATALOG_SERVICE_ADDR=localhost:9090 go run ./cmd/server
   ```
   Em outro terminal:
   ```
   curl -i localhost:8080/api/v1/addresses/01001-000
   ```
   E `curl -i localhost:8080/api/v1/products` mostra o **503** ao vivo, porque não tem catalog rodando local. Isso vira a demo de erro de infra de graça.
3. Se nada funcionar: explicar o código seguindo esse roteiro e mostrar os testes.

---

## Plano de estudo (quarta)

1. **Ler esse roteiro inteiro uma vez**, com o código aberto do lado, conferindo cada linha citada.
2. **Ensaiar a demo do começo ao fim, em voz alta, cronometrando.** Fazer pelo menos 2 vezes. Na primeira, vai travar em algum ponto, e é pra isso que serve.
3. **Responder em voz alta as "perguntas prováveis"** sem olhar a resposta.
4. Revisar o fluxo das camadas (cliente → router → handler → client gRPC → handler gRPC → repository → banco), porque é a base de quase toda pergunta de "por quê".
5. Ensaiar a demo de derrubar o catalog **e restaurar**. Se der qualquer problema no ensaio, cortar ela e usar o teste.
