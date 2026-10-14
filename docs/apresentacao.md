# Cola da apresentação: api-gateway + catalog-service

**Quinta, 15/10/2026 · 15 minutos · sem slides · roda local no Mac** (o cluster não é acessível da UFSC)

```
curl ──HTTP/JSON──▶ api-gateway ──gRPC──▶ catalog-service ──SQL──▶ Postgres
                         └──HTTP──▶ ViaCEP (API externa)
```

Arquivos citados por **nome de função**, não por linha (as linhas mudam com os comentários).

---

## Antes de começar (10 min antes)

Abrir 4 terminais. ⚠️ Ainda não testado no Mac: validar no ensaio.

**Terminal 1, Postgres** (qualquer pasta):
```
docker start catalog-db-local
```
Na primeira vez, em vez de `start`, criar o container:
```
docker run -d --name catalog-db-local -e POSTGRES_USER=catalog -e POSTGRES_PASSWORD=catalog -e POSTGRES_DB=catalog -p 5432:5432 postgres:16
```
E rodar as migrations (na raiz do projeto, só na primeira vez):
```
migrate -path catalog-service/migrations -database "postgres://catalog:catalog@localhost:5432/catalog?sslmode=disable" up
```
Se a 5432 estiver ocupada (Postgres.app), use `-p 5433:5432` e troque a porta nas URLs.

**Terminal 2, catalog-service** (dentro de `catalog-service/`):
```
DATABASE_URL="postgres://catalog:catalog@localhost:5432/catalog?sslmode=disable" PORT=8081 GRPC_PORT=9090 go run ./cmd/server
```

**Terminal 3, api-gateway** (dentro de `api-gateway/`):
```
CATALOG_SERVICE_ADDR=localhost:9090 go run ./cmd/server
```

**Terminal 4, curl e testes** (dentro de `api-gateway/`). Aquecimento:
```
curl -i localhost:8080/api/v1/products
```

VS Code: fonte grande, árvore de `api-gateway/` aberta.

---

## 0. Abertura (1 min)

> "ERP de uma empresa de arroz em microsserviços Go. Vou mostrar a porta de entrada REST, o api-gateway, e o catalog-service, que é dono de produtos e fornecedores e tem banco próprio. O gateway fala gRPC com o catalog e HTTP com o ViaCEP. Ele não tem banco, de propósito: não é dono de nenhum dado."

## 1. Caminho feliz (2 min)

```
curl -i -X POST localhost:8080/api/v1/products -H "Content-Type: application/json" -d '{"name":"Arroz demo"}'
```
→ **201**. Copiar o `id`.
```
curl -i localhost:8080/api/v1/products/ID_AQUI
```
→ **200**
```
curl -i -X PATCH localhost:8080/api/v1/products/ID_AQUI -H "Content-Type: application/json" -d '{"active":false}'
```
→ **200**, com `"active":false`
```
curl -i localhost:8080/api/v1/addresses/01001-000
```
→ **200**, vindo do ViaCEP

> "201 porque criou. PATCH e não DELETE porque produto nunca é apagado: lote e compra antiga apontam pra ele."

## 2. Critério 1: organização (2 min)

- **1.1** Árvore de pastas: `cmd/` (entrada), `handlers/` (HTTP), `clients/viacep/` (API externa), `pb/` (client gRPC gerado), `models/` (DTOs), `json/` (respostas e erros), `config/`. Depois o `main.go`: middlewares, rotas em `/api/v1`, clients criados uma vez e injetados.
- **1.2** `config.go`: tudo vem do ambiente. Sem `CATALOG_SERVICE_ADDR`, nem sobe (fail-fast). No cluster vem do ConfigMap, aqui vem do terminal.
- **1.3** `go.mod`: só dependências usadas, `go mod tidy`.

## 3. Critério 2: persistência, no catalog (2 min)

- **2.1** > "O catalog tem Postgres próprio, schema versionado com migrations (pasta `migrations/`)."
- **2.2** `catalog-service/internal/db/products.go`: > "Única camada que fala com o banco."
- **2.3** Função `Deactivate`: > "UPDATE num ID inexistente não dá erro, só afeta 0 linhas. Eu olho o `RowsAffected`."
  Depois `server.go`, função `GetProduct`: > "Só `ErrRecordNotFound` vira NotFound. Banco caído vira Internal, não 'produto não existe'."

## 4. Critério 3: REST e API externa (3 min)

- **3.2/3.3** `handlers_products.go`, função `HandleCreateProduct`: > "Sempre na ordem: JSON quebrado é 400, regra violada é 422, e só então chama o catalog."
  ```
  curl -i -X POST localhost:8080/api/v1/products -H "Content-Type: application/json" -d '{"name":"   "}'
  ```
  → **422**
  ```
  curl -i localhost:8080/api/v1/products/abc
  ```
  → **400**
- **3.1** `viacep.go`, função `Lookup`: > "Timeout de 5s no context, `defer Body.Close()`, confiro o StatusCode. E o ViaCEP responde 200 mesmo com CEP inexistente, então eu checo o corpo."
  ```
  curl -i localhost:8080/api/v1/addresses/99999999
  ```
  → **404**

## 5. Critério 4: erro de negócio × infra (2 min)

`grpcerror.go`: > "NotFound vira 404 e InvalidArgument vira 400: culpa do dado. Unavailable vira 503 e DeadlineExceeded vira 504: culpa da infra, com log. O ViaCEP segue a mesma ideia (`writeAddressError`: 400/404 × 502/504)."

**Demo:** `Ctrl+C` no **terminal 2** (catalog) e depois:
```
curl -i localhost:8080/api/v1/products
```
→ **503**. O gateway continua de pé.
Religar o catalog (seta pra cima + Enter no terminal 2).

## 6. Critério 5: qualidade (1 min)

`handlers_addresses.go`, interface `AddressLookup`: > "O handler depende de interface, não do client concreto. No teste entra um fake."
> "Erro sempre com `if err != nil`, embrulhado com `%w`. Os `_` descartam valores, nunca erros. Código em inglês, mensagens em português."

## 7. Critério 6: testes (1 min)

```
go test ./internal/handlers/ -v
```
> "22 subtestes em tabela, com fakes. O `wantCalled` prova que 400 e 422 nem chamam o catalog. Os ERROR no log são esperados: são os casos de infra."

## 8. Fechamento (30s)

> "REST com validação em camadas, gRPC pro catalog, HTTP pro ViaCEP, erro de usuário separado de erro de infra, persistência isolada no dono do dado. Próximos passos: JWT, estoque e compras."

---

## Perguntas prováveis

- **400 × 422?** 400: mal formado (JSON quebrado, ID não UUID). 422: bem formado, mas viola regra (nome vazio).
- **PATCH × PUT?** PUT substitui tudo, PATCH muda parte. Desativar só muda `active`.
- **Por que `*bool` no PATCH?** Pra distinguir "não mandou" (`nil`, 422) de `false`. Com `bool`, um `{}` desativaria sem querer.
- **Por que gRPC dentro e REST fora?** REST é universal pra quem consome. Entre serviços, o gRPC dá contrato tipado (`.proto`).
- **Por que DTO e não a struct do proto?** Mudar o proto não quebra a API pública.
- **502/504 × 500?** 500 é "eu falhei". 502/504 é "quem eu chamei falhou ou demorou".
- **Por que o gateway sobe sem o catalog?** O `grpc.NewClient` só conecta na primeira chamada.
- **Por que o gateway não tem banco?** Não é dono de dado. Cada serviço tem o seu (database-per-service).
- **Senha vai no ConfigMap?** Não, vai em Secret.

## Plano B

1. Catalog ou Docker não sobe: rodar só o gateway (terminal 3). `/addresses` funciona e `/products` dá 503, e isso já mostra o critério 4.
2. Sem internet: os testes rodam offline e cobrem quase todos os critérios.
