# CEP Weather

Sistema de microserviços em Go que recebe um CEP, identifica a cidade e retorna o clima atual (temperatura em Celsius, Fahrenheit e Kelvin). Implementa **OpenTelemetry (OTEL)** e **Zipkin** para observabilidade e tracing distribuído.

## 🏗️ Arquitetura de Microserviços

O sistema é composto por **dois serviços independentes** que se comunicam via HTTP, com instrumentação completa de tracing distribuído:

```
┌──────────────────┐                                                          
│     Cliente      │                                                          
│  (Postman/cURL)  │                                                          
└────────┬─────────┘                                                          
         │ POST /cep                                                          
         │ {"cep": "01001000"}                                                
         ▼                                                                    
┌─────────────────────────────────────────────────────────────────┐          
│  Serviço A - Input Service (Porta 8081)                         │          
│  ┌────────────────────────────────────────────────────────────┐ │          
│  │ • Valida formato do CEP (8 dígitos numéricos)              │ │          
│  │ • Retorna 422 se inválido                                  │ │          
│  │ • Encaminha para Serviço B via HTTP GET                    │ │          
│  │ • Propaga contexto OTEL via headers                        │ │          
│  └────────────────────────────────────────────────────────────┘ │          
└────────┬────────────────────────────────────────────────────────┘          
         │ GET /weather/01001000                                              
         │ (com traceparent header)                                           
         ▼                                                                    
┌─────────────────────────────────────────────────────────────────┐          
│  Serviço B - Weather Service (Porta 8080)                       │          
│  ┌────────────────────────────────────────────────────────────┐ │          
│  │ 1. Busca localização no ViaCEP                             │ │          
│  │    └─> Span: "viacep.Lookup" (mede latência)              │ │          
│  │                                                            │ │          
│  │ 2. Busca temperatura na WeatherAPI                        │ │          
│  │    └─> Span: "weatherapi.CurrentTemperatureC"            │ │          
│  │                                                            │ │          
│  │ 3. Converte temperaturas (C → F → K)                     │ │          
│  │                                                            │ │          
│  │ 4. Retorna JSON com cidade + temperaturas                 │ │          
│  └────────────────────────────────────────────────────────────┘ │          
└────────┬────────────────────────────────────────────────────────┘          
         │                                                                    
         ▼                                                                    
┌─────────────────────────────────────────────────────────────────┐          
│  Resposta Final                                                 │          
│  {                                                              │          
│    "city": "São Paulo",                                         │          
│    "temp_C": 28.5,                                              │          
│    "temp_F": 83.3,                                              │          
│    "temp_K": 301.5                                              │          
│  }                                                              │          
└─────────────────────────────────────────────────────────────────┘          
                                                                              
         │                                                                    
         └──────────────────────┐                                            
                                ▼                                            
                       ┌─────────────────┐                                   
                       │  Zipkin Server  │                                   
                       │  (Porta 9411)   │                                   
                       │                 │                                   
                       │  • UI Web       │                                   
                       │  • Query API    │                                   
                       │  • Visualização │                                   
                       └─────────────────┘                                   
                                                                              
    Serviço A e B enviam spans OTEL via HTTP para Zipkin                     
```

### 📊 Componentes do Sistema

#### 🔵 Serviço A - Input Service (Porta 8081)
- **Responsabilidade**: Validação de entrada e orquestração
- **Endpoint**: `POST /`
- **Validações**:
  - CEP deve ser string de exatamente 8 dígitos numéricos
  - Retorna `422` se formato inválido
- **Comportamento**:
  - Encaminha requisição válida para Serviço B via `GET /weather/{cep}`
  - Propaga contexto de tracing via header `traceparent` (W3C Trace Context)
  - Retorna resposta do Serviço B ao cliente
- **Observabilidade**: Cria span raiz para rastreamento end-to-end

#### 🟢 Serviço B - Weather Service (Porta 8080)
- **Responsabilidade**: Orquestração de APIs externas e lógica de negócio
- **Endpoint**: `GET /weather/{cep}`
- **Integrações**:
  1. **ViaCEP API**: Busca localização (cidade/estado) pelo CEP
  2. **WeatherAPI**: Busca temperatura atual da cidade
- **Processamento**:
  - Valida formato do CEP (8 dígitos)
  - Retorna `404` se CEP não encontrado no ViaCEP
  - Converte temperatura: Celsius → Fahrenheit → Kelvin
  - Combina dados de localização + clima em uma resposta unificada
- **Observabilidade**: 
  - Span `viacep.Lookup` com atributos: cep, city, state
  - Span `weatherapi.CurrentTemperatureC` com atributos: city, state, temp_c

#### 🟡 Zipkin - Distributed Tracing (Porta 9411)
- **Responsabilidade**: Coleta, armazenamento e visualização de traces
- **Interface Web**: `http://localhost:9411`
- **Funcionalidades**:
  - Visualização de traces end-to-end
  - Análise de latência por serviço/operação
  - Detecção de gargalos e erros
  - Query API para busca de traces

## 🌐 Ambientes de Execução

### 🐳 Ambiente Local (Sistema Completo)

Execute todo o sistema localmente com Docker Compose:

```bash
make docker-watch
```

**Serviços disponíveis:**
- **Serviço A (Input)**: `http://localhost:8081` - Ponto de entrada principal
- **Serviço B (Weather)**: `http://localhost:8080` - API de clima (pode ser acessado diretamente)
- **Zipkin UI**: `http://localhost:9411` - Interface de tracing distribuído

### ☁️ API em Produção (Cloud Run)

> **Nota**: Atualmente apenas o **Serviço B** está em produção. O Serviço A roda apenas localmente.

**Base URL:** `https://cepweather-763272253855.us-central1.run.app`

#### Endpoints Disponíveis (Serviço B)

##### 1. Consultar Temperatura por CEP
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
  "city": "São Lourenço da Mata",
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

O sistema implementa **OpenTelemetry (OTEL)** para instrumentação automática e **Zipkin** para visualização de traces distribuídos, permitindo rastrear requisições através de múltiplos serviços.

### 🎯 Por que Tracing Distribuído?

Em arquiteturas de microserviços, uma única requisição do usuário pode atravessar múltiplos serviços. O tracing distribuído permite:

- 🔍 **Visibilidade end-to-end**: Ver o caminho completo de uma requisição
- ⏱️ **Análise de latência**: Identificar quais serviços/operações são lentos
- 🐛 **Debug facilitado**: Rastrear erros através de múltiplos serviços
- 📊 **Métricas de performance**: Medir SLA e identificar gargalos

### 📡 O que é Instrumentado?

#### Serviço A (Input Service)
- **Span HTTP**: Toda requisição POST cria um span raiz
- **Propagação de contexto**: Injeta `traceparent` header ao chamar Serviço B

#### Serviço B (Weather Service)
- **Span HTTP**: Recebe contexto do Serviço A via header
- **Span `viacep.Lookup`**: Mede tempo de consulta ao ViaCEP
  - Atributos: `cep`, `city`, `state`
- **Span `weatherapi.CurrentTemperatureC`**: Mede tempo de consulta à WeatherAPI
  - Atributos: `city`, `state`, `temp_c`

### 🖥️ Como Visualizar Traces no Zipkin

1. **Inicie o sistema completo**:
```bash
make docker-watch
```

2. **Gere traces fazendo requisições**:
```bash
# Requisição de sucesso
curl -X POST http://localhost:8081 \
  -H "Content-Type: application/json" \
  -d '{"cep": "01001000"}'

# CEP não encontrado
curl -X POST http://localhost:8081 \
  -H "Content-Type: application/json" \
  -d '{"cep": "99999999"}'

# CEP inválido
curl -X POST http://localhost:8081 \
  -H "Content-Type: application/json" \
  -d '{"cep": "123"}'
```

3. **Acesse a UI do Zipkin**: `http://localhost:9411`

4. **Explore os traces**:
   - Clique em **"Run Query"** para buscar traces recentes
   - Selecione um trace na lista para ver detalhes
   - Visualize a **linha do tempo** (timeline) de cada span
   - Clique em cada span para ver **atributos** e **tags**

### 📊 Exemplo de Trace Completo

Quando você faz uma requisição bem-sucedida, o Zipkin mostra:

```
┌─────────────────────────────────────────────────────────────────────┐
│ service-a: POST / (200ms total)                                     │
│ ├─ http.method: POST                                                │
│ ├─ http.url: /                                                      │
│ └─ http.status_code: 200                                            │
│                                                                     │
│   └─> service-b: GET /weather/01001000 (190ms)                     │
│       ├─ http.method: GET                                           │
│       ├─ http.url: /weather/01001000                                │
│       └─ http.status_code: 200                                      │
│                                                                     │
│         ├─> viacep.Lookup (45ms)                                    │
│         │   ├─ cep: 01001000                                        │
│         │   ├─ city: São Paulo                                      │
│         │   └─ state: SP                                            │
│         │                                                           │
│         └─> weatherapi.CurrentTemperatureC (140ms)                  │
│             ├─ city: São Paulo                                      │
│             ├─ state: SP                                            │
│             └─ temp_c: 28.5                                         │
└─────────────────────────────────────────────────────────────────────┘

Timeline:
|----service-a (200ms)-------------------------------------------|
  |----service-b (190ms)-------------------------------------|
    |--viacep (45ms)--|
                        |----weatherapi (140ms)----------|
```

### 🏷️ Atributos Capturados nos Spans

| Span | Atributos | Exemplo |
|------|-----------|---------|
| `POST /` | `http.method`, `http.url`, `http.status_code` | `POST`, `/`, `200` |
| `GET /weather/{cep}` | `http.method`, `http.url`, `http.status_code` | `GET`, `/weather/01001000`, `200` |
| `viacep.Lookup` | `cep`, `city`, `state` | `01001000`, `São Paulo`, `SP` |
| `weatherapi.CurrentTemperatureC` | `city`, `state`, `temp_c` | `São Paulo`, `SP`, `28.5` |

### 🔧 Propagação de Contexto (W3C Trace Context)

O sistema usa o padrão **W3C Trace Context** para propagar o contexto de tracing entre serviços:

```http
# Serviço A envia para Serviço B:
GET /weather/01001000 HTTP/1.1
Host: service-b:8080
traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
```

**Formato do traceparent**:
```
00-<trace-id>-<parent-span-id>-<trace-flags>
│  │          │                 │
│  │          │                 └─ Flags (01 = sampled)
│  │          └─ ID do span pai (16 bytes hex)
│  └─ ID do trace (32 bytes hex)
└─ Versão (sempre 00)
```

Isso garante que todos os spans sejam correlacionados ao mesmo trace, mesmo atravessando múltiplos serviços.

### 🚀 Recursos Avançados do Zipkin

- **Filtros de busca**: Busque traces por service name, span name, duration, etc.
- **Dependências**: Visualize o grafo de dependências entre serviços
- **Comparação de traces**: Compare traces lentos vs rápidos
- **JSON export**: Exporte traces para análise offline

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

## 📝 Testando o Sistema

### 🔵 Fluxo Completo: Serviço A → Serviço B (Recomendado)

O Serviço A é o **ponto de entrada principal** do sistema e implementa a validação de CEP antes de encaminhar para o Serviço B.

#### ✅ 1. CEP Válido (Sucesso)
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

**O que acontece internamente:**
1. Serviço A valida formato do CEP ✅
2. Serviço A faz `GET http://service-b:8080/weather/01001000`
3. Serviço B busca localização no ViaCEP
4. Serviço B busca temperatura na WeatherAPI
5. Serviço B retorna JSON com city + temperaturas
6. Serviço A repassa resposta ao cliente

#### ❌ 2. CEP com Formato Inválido (422)
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

**Por que 422?** 
- CEP tem menos de 8 dígitos
- Validação falha no **Serviço A**
- Requisição não chega ao Serviço B

#### ❌ 3. CEP Válido mas Não Encontrado (404)
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

**O que acontece:**
1. Serviço A valida formato ✅ (8 dígitos)
2. Serviço A encaminha para Serviço B
3. Serviço B consulta ViaCEP
4. ViaCEP retorna `{"erro": "true"}` (CEP não existe)
5. Serviço B retorna 404
6. Serviço A repassa o 404 ao cliente

### 🟢 Acesso Direto ao Serviço B (Bypass)

Você também pode testar o Serviço B diretamente, sem passar pelo Serviço A:

```bash
# Via GET direto no Serviço B
curl http://localhost:8080/weather/01001000
```

**Diferença:**
- ✅ **Via Serviço A (POST)**: Validação de CEP + Tracing distribuído completo
- ⚠️ **Direto no Serviço B (GET)**: Sem validação prévia, apenas tracing do Serviço B

### 🧪 Collection Postman Completa

Importe essa collection no Postman para testar todos os cenários:

#### 📋 Requests para Ambiente Local

##### 1. [Serviço A] POST CEP Válido
- **Method:** `POST`
- **URL:** `http://localhost:8081`
- **Headers:** 
  ```
  Content-Type: application/json
  ```
- **Body (raw JSON):**
  ```json
  {
    "cep": "01001000"
  }
  ```
- **Resposta esperada:** `200 OK` com city + temperaturas

##### 2. [Serviço A] POST CEP Inválido (Formato)
- **Method:** `POST`
- **URL:** `http://localhost:8081`
- **Headers:** 
  ```
  Content-Type: application/json
  ```
- **Body (raw JSON):**
  ```json
  {
    "cep": "123"
  }
  ```
- **Resposta esperada:** `422 Unprocessable Entity`

##### 3. [Serviço A] POST CEP Não Encontrado
- **Method:** `POST`
- **URL:** `http://localhost:8081`
- **Headers:** 
  ```
  Content-Type: application/json
  ```
- **Body (raw JSON):**
  ```json
  {
    "cep": "99999999"
  }
  ```
- **Resposta esperada:** `404 Not Found`

##### 4. [Serviço B] GET Direto (Bypass do A)
- **Method:** `GET`
- **URL:** `http://localhost:8080/weather/54735220`
- **Headers:** Nenhum
- **Resposta esperada:** `200 OK` com city + temperaturas

##### 5. [Serviço A] Health Check
- **Method:** `GET`
- **URL:** `http://localhost:8081/healthz`
- **Headers:** Nenhum
- **Resposta esperada:** `200 OK` com body `ok`

##### 6. [Serviço B] Health Check
- **Method:** `GET`
- **URL:** `http://localhost:8080/healthz`
- **Headers:** Nenhum
- **Resposta esperada:** `200 OK` com body `ok`

#### ☁️ Requests para Produção (Cloud Run - Apenas Serviço B)

> **Nota**: O Serviço A não está em produção, apenas o Serviço B.

##### 1. [Produção] GET CEP Válido
- **Method:** `GET`
- **URL:** `https://cepweather-763272253855.us-central1.run.app/weather/54735220`
- **Headers:** Nenhum
- **Resposta esperada:** `200 OK`

##### 2. [Produção] GET CEP Não Encontrado
- **Method:** `GET`
- **URL:** `https://cepweather-763272253855.us-central1.run.app/weather/99999999`
- **Headers:** Nenhum
- **Resposta esperada:** `404 Not Found`

##### 3. [Produção] Health Check
- **Method:** `GET`
- **URL:** `https://cepweather-763272253855.us-central1.run.app/healthz`
- **Headers:** Nenhum
- **Resposta esperada:** `200 OK` com body `ok`

### 📊 Matriz de Testes Recomendada

| # | Cenário | Endpoint | Body/Param | Status | Response |
|---|---------|----------|------------|--------|----------|
| 1 | ✅ CEP válido (São Paulo) | `POST http://localhost:8081` | `{"cep":"01001000"}` | 200 | City + Temps |
| 2 | ✅ CEP válido (Rio) | `POST http://localhost:8081` | `{"cep":"20040020"}` | 200 | City + Temps |
| 3 | ✅ CEP válido (Recife) | `POST http://localhost:8081` | `{"cep":"50010000"}` | 200 | City + Temps |
| 4 | ❌ CEP curto (3 dígitos) | `POST http://localhost:8081` | `{"cep":"123"}` | 422 | Invalid zipcode |
| 5 | ❌ CEP longo (9 dígitos) | `POST http://localhost:8081` | `{"cep":"012345678"}` | 422 | Invalid zipcode |
| 6 | ❌ CEP não numérico | `POST http://localhost:8081` | `{"cep":"abcd1234"}` | 422 | Invalid zipcode |
| 7 | ❌ CEP não encontrado | `POST http://localhost:8081` | `{"cep":"99999999"}` | 404 | Cannot find zipcode |
| 8 | ✅ Direto no Serviço B | `GET http://localhost:8080/weather/01001000` | - | 200 | City + Temps |
| 9 | ✅ Health Check A | `GET http://localhost:8081/healthz` | - | 200 | ok |
| 10 | ✅ Health Check B | `GET http://localhost:8080/healthz` | - | 200 | ok |

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
