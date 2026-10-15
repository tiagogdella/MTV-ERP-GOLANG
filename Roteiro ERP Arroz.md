[Abertura](#abertura)[1 Arquitetura](#s1)[2 Persistência](#s2)[3 REST externa](#s3)[4 Erros](#s4)[5 Qualidade Go](#s5)[6 Testes](#s6)

# Roteiro — ERP de arroz em microsserviços Go

api-gateway + catalog-service · fatia vertical

## 00Abertura

O projeto é um ERP para uma empresa de beneficiamento de arroz, em microsserviços Go. O sistema completo tem 5 serviços (gateway, autenticação, catálogo, estoque e compras), cada um com o próprio banco.\
\
Hoje eu apresento a primeira fatia vertical funcionando: o api-gateway, que é a porta de entrada REST, e o catalog-service, dono dos produtos e fornecedores.\
\
O fluxo é: o cliente chama o gateway em HTTP/JSON, o gateway fala com o catalog por gRPC, e o catalog grava no Postgres dele. O gateway também consulta uma API externa, o ViaCEP. Ele não tem banco, de propósito: não é dono de nenhum dado.\
\
Normalmente roda num cluster Kubernetes no servidor da empresa. Aqui está rodando local, com as mesmas configurações vindas de variável de ambiente.

```
curl ──HTTP/JSON──▶ api-gateway ──gRPC──▶ catalog-service ──SQL──▶ Postgres
                         └──HTTP──▶ ViaCEP
```

200teste

`curl -i localhost:8080/api/v1/products`

## 01Arquitetura e organização do serviço

### 1.1Estrutura de pastas coerente

Separação entre handlers/rotas, lógica de negócio, acesso a dados, clients externos.

```
api-gateway/
├── cmd/server/main.go          → ponto de entrada: monta e liga tudo
└── internal/
    ├── handlers/               → camada de entrada HTTP: lê JSON, valida (400/422), chama o client
    ├── clients/viacep/         → client da API externa (ViaCEP)
    ├── pb/catalog/v1/          → client gRPC do catalog, GERADO a partir do .proto
    ├── models/                 → DTOs: o formato do JSON que sai pra fora
    ├── json/                   → helpers de resposta + tradução erro gRPC → status HTTP
    ├── config/                 → lê as variáveis de ambiente
    ├── health/                 → /healthz e /readyz (o Kubernetes usa)
    └── observability/          → tracing (OpenTelemetry: coleta dados)
```

Mostrar

main.go

Monta as peças e liga o servidor.

main.go 25–26

Logs

### 1.2Configuração via variáveis de ambiente

Porta, URL de banco, URL de serviços externos.

Variáveis de ambiente

main.go 39–42

**config.Load()** — carrega as variáveis de ambiente.\
Lógica: config.Load() lê as variáveis e, se falta uma obrigatória, devolve erro → os.Exit(1). **(FAILFAST)** melhor quebrar agora do que no meio da app.\
Sai: o cfg, com porta, endereço do catalog e URL do ViaCEP.

Clientes

main.go 46–53

**grpc.NewClient** prepara a conexão gRPC com o catalog, mas não conecta agora, só na primeira chamada.\
**NewCatalogServiceClient** embrulha a conexão no client gerado pelo .proto, que é uma interface (5.2).\
A diferença entre as duas variáveis é importante:

- **catalogConn**: representa a conexão gRPC.
- **catalogClient**: fornece os métodos para realizar as chamadas ao serviço.

**viacep.NewClient(..., 5\*time.Second)** cria o client da API externa com timeout de 5s.

Middlewares

main.go 56–61

- **Log**
- **Recoverer** — um panic vira 500 (Internal Server Error) em vez de derrubar o processo
- **RequestID** — serve pra identificar cada req com um ID
- **Timeout(10s)** — limite para processamento de reqs. O ViaCEP tem 5s próprios, e vale o que vencer primeiro.
- **AllowContentType** — aceita só JSON

Rotas

main.go 70–89

Mostrar as rotas apenas.

Fluxo de camadas — onde o main.go entra

```
curl → [middlewares + router: main.go] → handler → client gRPC → (rede) → handler gRPC → repository → Postgres
```

config.go

Mostrar principais.

deploy/api-gateway/configmap.yaml

### 1.3go.mod limpo

Nome de módulo e dependências fazem sentido para o que é usado.

Módulo mtv-erp/api-gateway, Go 1.25. As dependências diretas são só as usadas: chi (router), grpc/protobuf (catalog), uuid (validar ID), prometheus/otel (observabilidade) e testify (testes). Mantido limpo com go mod tidy.

## 02Persistência própria

### 2.1Banco/schema próprio

Não compartilha tabelas com outro microsserviço, mesmo que seja o mesmo SGBD.

O catalog tem Postgres próprio, que nenhum outro serviço acessa. O schema é versionado em migrations/.

### 2.2Camada de acesso a dados isolada

repository/store.

Mostrar

Abre o catalog e mostra o products.

### 2.3"Não encontrado" ≠ erro de conexão/infra

Mostrar

catalog products.go 35–45

catalog server.go

404Prova o RowsAffected ao vivo

`curl -i -X PATCH localhost:8080/api/v1/products/00000000-0000-0000-0000-000000000000 -H "Content-Type: application/json" -d '{"active":false}'`

## 03Comunicação REST externa

### 3.1API externa

Usa context com timeout, fecha o Body da resposta, confere StatusCode antes de assumir sucesso.

Mostrar · clients/viacep.go

viacep.go 53

Timeout no context.

viacep.go 70

Fechamento do body da função.

viacep.go 73

Não assume sucesso: 400 do ViaCEP → CEP inválido, e qualquer coisa diferente de 200 → indisponível.

viacep.go 85

Conferir status code: o ViaCEP responde 200 mesmo com CEP inexistente, com {"erro": "true"}. Então conferir só o status não basta: se o cep vier vazio no corpo → ErrNotFound.

200CEP válido

`curl -i localhost:8080/api/v1/addresses/88920-000`

404CEP inexistente

`curl -i localhost:8080/api/v1/addresses/99999999`

### 3.2 / 3.3Rotas RESTful e JSON bem definido

3.2 — Rotas RESTful coerentes (verbos e status HTTP corretos — 200/201/400/404/500 usados com critério).\
3.3 — JSON de entrada e saída bem definido.

Mostrar

handlers_products.go

Testes

201Created — copia o id

`curl -i -X POST localhost:8080/api/v1/products -H "Content-Type: application/json" -d '{"name":"Arroz demo"}'`

200Com active false — troque ID_AQUI pelo id

`curl -i -X PATCH localhost:8080/api/v1/products/ID_AQUI -H "Content-Type: application/json" -d '{"active":false}'`

400JSON quebrado

`curl -i -X POST localhost:8080/api/v1/products -H "Content-Type: application/json" -d '{"name":'`

422Nome vazio

`curl -i -X POST localhost:8080/api/v1/products -H "Content-Type: application/json" -d '{"name":" "}'`

400Não é id

`curl -i localhost:8080/api/v1/products/abc`

## 04Tratamento de erros

### 4.1Erro de negócio ≠ erro de infraestrutura

Ex: CEP inválido, aluno não encontrado vs. serviço externo fora do ar, timeout.

Mostrar

handlers_address.go 40

writeAddressError

503Encerrar o catalog (Ctrl + C) e testar sem. Apontar o ERROR no terminal do gateway.

`curl -i localhost:8080/api/v1/products`

## 05Qualidade do código Go

### 5.1Tratamento de erro idiomático

if err != nil, sem ignorar erros silenciosamente com \_.

Todo erro é tratado com if err != nil. Quando repasso, embrulho com %w para não perder a causa, e é por isso que o errors.Is funciona. Os \_ do código descartam valores, nunca erros.

### 5.2Uso de interfaces onde faz sentido

handlers_address.go 16

**AddressLookup** e **Lookup()** — bom para poder testar; adiciona uma camada que diz "precisa ter lookup".

### 5.3Nomes e organização consistentes

Em português ou inglês, mas consistentes.

Código e JSON em inglês, mensagens ao usuário em português. Arquivos no padrão handlers\_\<recurso>.go, com o teste ao lado.

## 06Testes

### 6.1Testes de unidade para lógica de negócio e/ou handlers

22 subtestes em tabela. O fake é um dublê do catalog com a mesma interface, então eu simulo catalog fora do ar sem derrubar nada. O wantCalled prova que 400 e 422 nem chamam o catalog.

runRodar os testes

`make test`