# CEP Weather

Serviço em Go que consulta o endereço de um CEP brasileiro usando ViaCEP e retorna a temperatura atual (Celsius, Fahrenheit e Kelvin) para a cidade correspondente utilizando a WeatherAPI. Está preparado para rodar localmente, via Docker e para ser implantado no Google Cloud Run.

## 🌐 API em Produção

A API está disponível publicamente no Google Cloud Run:

**Base URL:** `https://cepweather-763272253855.us-central1.run.app`

### Endpoints Disponíveis

#### 1. Consultar Temperatura por CEP
```http
GET /weather/{cep}
```

**Exemplo de requisição:**
```bash
curl https://cepweather-763272253855.us-central1.run.app/weather/54735220
```

**Exemplo de resposta (200 OK):**
```json
{
  "temp_C": 33.7,
  "temp_F": 92.7,
  "temp_K": 306.7
}
```

**Possíveis respostas de erro:**
- `422 Unprocessable Entity` - CEP inválido ou com formato incorreto
- `404 Not Found` - CEP não encontrado na base de dados
- `500 Internal Server Error` - Erro ao processar a requisição

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

| CEP | Cidade | Estado |
|-----|--------|--------|
| `01001000` | São Paulo | SP |
| `20040020` | Rio de Janeiro | RJ |
| `30140071` | Belo Horizonte | MG |
| `80010000` | Curitiba | PR |
| `54735220` | São Lourenço da Mata | PE |

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
| `PORT`                  | Não         | `8080`                               | Porta exposta pelo servidor HTTP.         |

## Execução local

### Usando arquivo .env

1. Crie um arquivo `.env` na raiz do projeto:
```bash
WEATHER_API_KEY=sua_chave_aqui
```

2. Execute com Make:
```bash
make docker-watch
```

### Usando variáveis de ambiente diretamente

```bash
export WEATHER_API_KEY=coloque_sua_chave_aqui
PORT=8080 go run ./cmd/api
```

### Testando localmente

Com o servidor no ar:

```bash
curl http://localhost:8080/weather/01001000
```

Resposta de exemplo:

```json
{
  "temp_C": 28.5,
  "temp_F": 83.3,
  "temp_K": 301.5
}
```

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
| `make run` | Executa a aplicação localmente com Go |
| `make test` | Executa os testes unitários |
| `make build` | Compila o binário da aplicação |
| `make docker-build` | Cria a imagem Docker |
| `make docker-run` | Executa o container Docker |
| `make docker-watch` | Build, sobe container e mostra logs (usa .env) |
| `make compose` | Executa com Docker Compose |
| `make clean` | Remove arquivos compilados e cache |

## 📝 Testando com Postman

### Importar Collection

Você pode testar a API usando o Postman com as seguintes requisições:

**1. Health Check**
- Method: `GET`
- URL: `https://cepweather-763272253855.us-central1.run.app/healthz`
- Headers: Nenhum necessário

**2. Buscar Temperatura por CEP**
- Method: `GET`
- URL: `https://cepweather-763272253855.us-central1.run.app/weather/54735220`
- Headers: Nenhum necessário

### Exemplos de Testes

**Teste com CEP válido:**
```
GET https://cepweather-763272253855.us-central1.run.app/weather/01001000
Resposta esperada: 200 OK
```

**Teste com CEP inválido:**
```
GET https://cepweather-763272253855.us-central1.run.app/weather/00000000
Resposta esperada: 422 Unprocessable Entity
```

**Teste com CEP não encontrado:**
```
GET https://cepweather-763272253855.us-central1.run.app/weather/99999999
Resposta esperada: 404 Not Found
```
