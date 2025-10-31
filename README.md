# CEP Weather

Sistema distribuído em Go que recebe um CEP, identifica a cidade e retorna o clima atual (temperatura em Celsius, Fahrenheit e Kelvin) juntamente com a cidade. Implementa **OpenTelemetry (OTEL)** e **Zipkin** para tracing distribuído.

## 🏗️ Arquitetura

O sistema é composto por dois serviços que se comunicam via HTTP:

```
┌─────────────┐      POST      ┌─────────────┐     GET     ┌──────────┐
│  Serviço A  │  ─────────────> │  Serviço B  │ ──────────> │  ViaCEP  │
│   (Input)   │  {"cep":"..."}  │  (Weather)  │             └──────────┘
└─────────────┘                 └─────────────┘
      │                               │                      ┌────────────┐
      │                               └─────────────────────>│ WeatherAPI │
      │                                                      └────────────┘
      │                          ┌─────────┐
      └─────────────────────────>│ Zipkin  │<───────────────┘
                 OTEL Traces      └─────────┘    OTEL Traces
```

### Serviço A - Input Service (Porta 8081)
- Recebe requisições POST com `{"cep": "12345678"}`
- Valida se o CEP tem 8 dígitos e é uma string
- Encaminha para o Serviço B via HTTP
- Retorna 422 se o CEP for inválido

### Serviço B - Weather Service (Porta 8080)
- Recebe CEP do Serviço A (ou diretamente via GET)
- Consulta o ViaCEP para obter a localização
- Consulta a WeatherAPI para obter a temperatura
- Retorna: `{"city": "São Paulo", "temp_C": 28.5, "temp_F": 83.3, "temp_K": 301.5}`

### Zipkin - Distributed Tracing (Porta 9411)
- Coleta e visualiza traces distribuídos
- Interface web em `http://localhost:9411`
- Permite rastrear requisições end-to-end

## 🌐 API em Produção

A API está disponível publicamente no Google Cloud Run:

**Base URL:** `https://cepweather-763272253855.us-central1.run.app`

### Endpoints Disponíveis

#### 1. Consultar Temperatura por CEP
```http
GET /weather/{cep}
```

**Exemplo de requisição com CEP válido:**
```bash
curl https://cepweather-763272253855.us-central1.run.app/weather/54735220
```

**Resposta de sucesso (200 OK):**
```json
{
  "temp_C": 28.5,
  "temp_F": 83.3,
  "temp_K": 301.5
}
```

**Respostas de erro:**

| Status | Mensagem | Descrição |
|--------|----------|-----------|
| `422` | `{"message":"invalid zipcode"}` | CEP com formato inválido (tamanho incorreto, caracteres especiais, etc.) |
| `404` | `{"message":"can not find zipcode"}` | CEP não encontrado na base de dados do ViaCEP |
| `500` | `{"message":"internal server error"}` | Erro inesperado no servidor ou nas APIs externas |

**Exemplos de erros:**

```bash
# CEP não encontrado
curl https://cepweather-763272253855.us-central1.run.app/weather/53424543
# Resposta: 404 {"message":"can not find zipcode"}

# CEP com formato inválido (muito longo)
curl https://cepweather-763272253855.us-central1.run.app/weather/012345678
# Resposta: 422 {"message":"invalid zipcode"}
```

#### 2. Health Check
```http
GET /healthz
```

**Exemplo de requisição:**
```bash
curl https://cepweather-763272253855.us-central1.run.app/healthz
```

**Resposta:**
```
ok
```

### CEPs para Teste

| CEP | Cidade | Estado | Status Esperado |
|-----|--------|--------|-----------------|
| `01001000` | São Paulo | SP | ✅ 200 OK |
| `20040020` | Rio de Janeiro | RJ | ✅ 200 OK |
| `30140071` | Belo Horizonte | MG | ✅ 200 OK |
| `80010000` | Curitiba | PR | ✅ 200 OK |
| `54735220` | São Lourenço da Mata | PE | ✅ 200 OK |
| `53424543` | CEP não encontrado | - | ❌ 404 Not Found |
| `00000000` | CEP inválido | - | ❌ 404 Not Found |
| `123` | Formato inválido | - | ❌ 422 Invalid |

## Requisitos

- Go 1.22 ou superior (para execução local sem Docker)
- Docker e Docker Compose (para a execução containerizada)
- Conta na [WeatherAPI](https://www.weatherapi.com/) e chave de acesso (`WEATHER_API_KEY`)
- Conta Google Cloud com o SDK `gcloud` configurado (para deploy)

## Variáveis de ambiente

| Nome                    | Obrigatório | Default                              | Descrição                                |
|-------------------------|-------------|--------------------------------------|-------------------------------------------|
| `WEATHER_API_KEY`       | Sim         | —                                    | Chave da WeatherAPI.                      |
| `VIACEP_BASE_URL`       | Não         | `https://viacep.com.br/ws`           | Endpoint do serviço ViaCEP.               |
| `WEATHER_API_BASE_URL`  | Não         | `https://api.weatherapi.com/v1`      | Endpoint da WeatherAPI.                   |
| `SERVICE_B_URL`         | Não         | `http://localhost:8080`              | URL do Serviço B (usado pelo Serviço A). |
| `ZIPKIN_URL`            | Não         | `http://zipkin:9411/api/v2/spans`    | URL do exportador Zipkin.                |
| `PORT`                  | Não         | `8080` (B) / `8081` (A)              | Porta exposta pelos servidores HTTP.      |

## 🚀 Execução local

### Opção 1: Sistema Completo com Docker Compose (Recomendado)

Esta é a forma mais simples de rodar todo o sistema com tracing distribuído:

1. Crie um arquivo `.env` na raiz do projeto:
```bash
WEATHER_API_KEY=sua_chave_aqui
```

2. Execute o sistema completo:
```bash
make docker-watch
```

Isso irá iniciar:
- **Serviço A (Input)** em `http://localhost:8081`
- **Serviço B (Weather)** em `http://localhost:8080`
- **Zipkin UI** em `http://localhost:9411`

3. Teste o sistema completo (via Serviço A):
```bash
curl -X POST http://localhost:8081 \
  -H "Content-Type: application/json" \
  -d '{"cep": "01001000"}'
```

Resposta esperada:
```json
{
  "city": "São Paulo",
  "temp_C": 28.5,
  "temp_F": 83.3,
  "temp_K": 301.5
}
```

4. Visualize os traces no Zipkin:
   - Abra `http://localhost:9411` no navegador
   - Clique em "Run Query" para ver os traces
   - Explore o trace distribuído Service-A → Service-B → ViaCEP/WeatherAPI

### Opção 2: Executar serviços individuais (sem Docker)

#### Serviço B (Weather API):
```bash
export WEATHER_API_KEY=sua_chave_aqui
make run
```

#### Serviço A (Input Service):
```bash
make run-input-service
```

Teste direto no Serviço B:
```bash
curl http://localhost:8080/weather/01001000
```

Teste via Serviço A:
```bash
curl -X POST http://localhost:8081 \
  -H "Content-Type: application/json" \
  -d '{"cep": "01001000"}'
```

## 🔍 Observabilidade e Tracing Distribuído

O sistema implementa **OpenTelemetry (OTEL)** para instrumentação e **Zipkin** para visualização de traces distribuídos.

### O que é rastreado?

O sistema cria spans para medir o tempo de:

1. **Serviço A → Serviço B**: Propagação de contexto via HTTP headers
2. **Busca de CEP no ViaCEP**: Span `viacep.Lookup`
3. **Busca de temperatura na WeatherAPI**: Span `weatherapi.CurrentTemperatureC`

### Como visualizar os traces no Zipkin?

1. Acesse a interface do Zipkin: `http://localhost:9411`

2. Faça uma requisição para gerar traces:
```bash
curl -X POST http://localhost:8081 \
  -H "Content-Type: application/json" \
  -d '{"cep": "01001000"}'
```

3. No Zipkin UI:
   - Clique em **"Run Query"** para buscar traces
   - Selecione um trace para ver detalhes
   - Visualize a linha do tempo completa: Service-A → Service-B → APIs externas
   - Veja os atributos de cada span (CEP, cidade, temperatura, etc.)

### Exemplo de Trace

```
service-a: handle-cep (150ms)
  └─> service-b: GET /weather/01001000 (140ms)
      ├─> viacep.Lookup (50ms)
      │   └─ Attributes: cep=01001000, city=São Paulo, state=SP
      └─> weatherapi.CurrentTemperatureC (85ms)
          └─ Attributes: city=São Paulo, state=SP, temp_c=28.5
```

### Atributos Capturados nos Spans

| Span | Atributos |
|------|-----------|
| `viacep.Lookup` | `cep`, `city`, `state` |
| `weatherapi.CurrentTemperatureC` | `city`, `state`, `temp_c` |

## Testes

### Executar todos os testes

```bash
make test
```

Ou diretamente com Go:

```bash
GOCACHE=$(pwd)/.cache go test ./...
```

### Cobertura de testes

```bash
go test -cover ./...
```

## Docker

### Build da imagem

```bash
make docker-build
```

Ou diretamente:

```bash
docker build -t cepweather .
```

### Executar container

```bash
make docker-run WEATHER_API_KEY=sua_chave_aqui
```

Ou diretamente:

```bash
docker run --rm -p 8080:8080 -e WEATHER_API_KEY=sua_chave_aqui cepweather
```

### Docker Compose

Com arquivo `.env` configurado:

```bash
make docker-watch
```

Ou manualmente:

```bash
export WEATHER_API_KEY=coloque_sua_chave_aqui
docker compose up --build
```

Para rodar em background:

```bash
docker compose up -d
```

Para parar:

```bash
docker compose down
```

## Deploy no Google Cloud Run

### Pré-requisitos

1. Instalar o Google Cloud SDK:
   ```bash
   brew install google-cloud-sdk
   ```

2. Autenticar-se no Google Cloud:
   ```bash
   gcloud auth login
   ```

3. Configurar o projeto:
   ```bash
   gcloud config set project SEU_PROJETO_ID
   ```

### Deploy Simplificado (Recomendado)

Deploy direto do código-fonte:

```bash
gcloud run deploy cepweather \
  --source . \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars WEATHER_API_KEY=sua_chave_aqui
```

### Deploy via Artifact Registry (Alternativo)

1. Criar repositório no Artifact Registry (apenas uma vez):
   ```bash
   gcloud artifacts repositories create cepweather \
     --repository-format=docker \
     --location=us-central1
   ```

2. Fazer build e push da imagem:
   ```bash
   gcloud builds submit --tag us-central1-docker.pkg.dev/SEU_PROJETO/cepweather/app
   ```

3. Fazer deploy:
   ```bash
   gcloud run deploy cepweather \
     --image us-central1-docker.pkg.dev/SEU_PROJETO/cepweather/app \
     --platform managed \
     --region us-central1 \
     --allow-unauthenticated \
     --set-env-vars WEATHER_API_KEY=sua_chave_aqui
   ```

### Gerenciar o serviço

Ver logs:
```bash
gcloud run services logs read cepweather --region us-central1
```

Atualizar variáveis de ambiente:
```bash
gcloud run services update cepweather \
  --region us-central1 \
  --set-env-vars WEATHER_API_KEY=nova_chave
```

Deletar o serviço:
```bash
gcloud run services delete cepweather --region us-central1
```

### Teste pós-deploy

Após o deploy, a URL do serviço será exibida. Teste com:

```bash
curl https://SEU_ENDPOINT/weather/01001000
```

## 🛠️ Comandos Make Disponíveis

| Comando | Descrição |
|---------|-----------|
| `make run` | Executa o Serviço B (Weather API) localmente |
| `make run-input-service` | Executa o Serviço A (Input Service) localmente |
| `make test` | Executa os testes unitários |
| `make build` | Compila os binários de ambos os serviços |
| `make docker-build` | Cria as imagens Docker de ambos os serviços |
| `make docker-run` | Executa o container do Serviço B |
| `make docker-watch` | **Sistema completo** com Docker Compose + Zipkin |
| `make compose` | Executa com Docker Compose |
| `make clean` | Remove arquivos compilados e cache |

## 📝 Testando com Postman ou cURL

### Testando o Sistema Completo (Serviço A)

O Serviço A é o ponto de entrada principal e valida o CEP antes de encaminhar para o Serviço B.

#### 1. Enviar CEP via POST (Método correto)
```bash
curl -X POST http://localhost:8081 \
  -H "Content-Type: application/json" \
  -d '{"cep": "01001000"}'
```

**Resposta esperada (200 OK):**
```json
{
  "city": "São Paulo",
  "temp_C": 28.5,
  "temp_F": 83.3,
  "temp_K": 301.5
}
```

#### 2. Teste com CEP inválido (formato incorreto)
```bash
curl -X POST http://localhost:8081 \
  -H "Content-Type: application/json" \
  -d '{"cep": "123"}'
```

**Resposta esperada (422 Unprocessable Entity):**
```json
{
  "message": "invalid zipcode"
}
```

#### 3. Teste com CEP não encontrado
```bash
curl -X POST http://localhost:8081 \
  -H "Content-Type: application/json" \
  -d '{"cep": "99999999"}'
```

**Resposta esperada (404 Not Found):**
```json
{
  "message": "can not find zipcode"
}
```

### Testando Diretamente no Serviço B (Bypass do Serviço A)

Você também pode testar o Serviço B diretamente via GET:

```bash
curl http://localhost:8080/weather/01001000
```

### Collection Postman

**Collection para importar no Postman:**

#### 1. Serviço A - POST CEP Válido
- **Method:** `POST`
- **URL:** `http://localhost:8081`
- **Headers:** `Content-Type: application/json`
- **Body (raw JSON):**
```json
{
  "cep": "01001000"
}
```

#### 2. Serviço A - POST CEP Inválido
- **Method:** `POST`
- **URL:** `http://localhost:8081`
- **Headers:** `Content-Type: application/json`
- **Body (raw JSON):**
```json
{
  "cep": "123"
}
```

#### 3. Serviço B - GET Direto
- **Method:** `GET`
- **URL:** `http://localhost:8080/weather/54735220`
- **Headers:** Nenhum necessário

#### 4. Health Check - Serviço A
- **Method:** `GET`
- **URL:** `http://localhost:8081/healthz`
- **Headers:** Nenhum necessário

#### 5. Health Check - Serviço B
- **Method:** `GET`
- **URL:** `http://localhost:8080/healthz`
- **Headers:** Nenhum necessário

### Testando com Postman - API Produção (Serviço B apenas)

Você pode testar a API usando o Postman com as seguintes requisições:

#### 1. Health Check
- **Method:** `GET`
- **URL:** `https://cepweather-763272253855.us-central1.run.app/healthz`
- **Headers:** Nenhum necessário
- **Resposta esperada:** `200 OK` com corpo `ok`

#### 2. Consultar Temperatura - CEP Válido
- **Method:** `GET`
- **URL:** `https://cepweather-763272253855.us-central1.run.app/weather/54735220`
- **Headers:** Nenhum necessário
- **Resposta esperada:** `200 OK`
```json
{
  "temp_C": 28.5,
  "temp_F": 83.3,
  "temp_K": 301.5
}
```

#### 3. Consultar Temperatura - CEP Não Encontrado
- **Method:** `GET`
- **URL:** `https://cepweather-763272253855.us-central1.run.app/weather/53424543`
- **Headers:** Nenhum necessário
- **Resposta esperada:** `404 Not Found`
```json
{
  "message": "can not find zipcode"
}
```

#### 4. Consultar Temperatura - CEP Inválido
- **Method:** `GET`
- **URL:** `https://cepweather-763272253855.us-central1.run.app/weather/123456789`
- **Headers:** Nenhum necessário
- **Resposta esperada:** `422 Unprocessable Entity`
```json
{
  "message": "invalid zipcode"
}
```

### Casos de Teste Recomendados

| Caso de Teste | URL | Status Esperado | Descrição |
|---------------|-----|-----------------|-----------|
| ✅ CEP válido | `/weather/01001000` | 200 | Retorna temperaturas |
| ❌ CEP não encontrado | `/weather/99999999` | 404 | CEP não existe |
| ❌ CEP não encontrado | `/weather/53424543` | 404 | CEP inexistente |
| ❌ Formato inválido | `/weather/123` | 422 | Menos de 8 dígitos |
| ❌ Formato inválido | `/weather/012345678` | 422 | Mais de 8 dígitos |
| ❌ Rota vazia | `/weather/` | 404 | Sem CEP |
| ✅ Health check | `/healthz` | 200 | Servidor funcionando |

## 🐛 Tratamento de Erros

A API trata corretamente os seguintes cenários de erro:

### 1. CEP com formato inválido (422)
- CEP com menos ou mais de 8 dígitos
- CEP com letras ou caracteres especiais
- Retorna: `{"message":"invalid zipcode"}`

### 2. CEP não encontrado (404)
- CEP com formato válido mas não existe na base do ViaCEP
- Retorna: `{"message":"can not find zipcode"}`

### 3. Erros de APIs externas (500)
- Timeout na comunicação com ViaCEP ou WeatherAPI
- Erro de parsing de resposta
- Retorna: `{"message":"internal server error"}`

### 4. Rotas não encontradas (404)
- Acesso a endpoints inexistentes
- Retorna resposta padrão do servidor

## 🔧 Melhorias Implementadas

### Correção de Bug - ViaCEP Response
O ViaCEP retorna o campo `"erro"` como string `"true"` em vez de boolean quando um CEP não é encontrado. A aplicação foi corrigida para tratar ambos os casos:

```go
// Trata tanto "erro": true quanto "erro": "true"
hasError := false
if payload.Erro != nil {
    switch v := payload.Erro.(type) {
    case bool:
        hasError = v
    case string:
        hasError = v == "true"
    }
}
```

Isso evita erros 500 quando CEPs inválidos são consultados e retorna corretamente 404 com a mensagem apropriada.
