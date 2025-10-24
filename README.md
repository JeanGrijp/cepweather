# CEP Weather

Serviço em Go que consulta o endereço de um CEP brasileiro usando ViaCEP e retorna a temperatura atual (Celsius, Fahrenheit e Kelvin) para a cidade correspondente utilizando a WeatherAPI. Está preparado para rodar localmente, via Docker e para ser implantado no Google Cloud Run.

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

```bash
export WEATHER_API_KEY=coloque_sua_chave_aqui
PORT=8080 go run ./cmd/api
```

Com o servidor no ar:

```bash
curl http://localhost:8080/weather/01001000
```

Resposta de exemplo (valores ilustrativos):

```json
{
  "temp_C": 28.5,
  "temp_F": 83.3,
  "temp_K": 301.5
}
```

## Testes

```bash
GOCACHE=$(pwd)/.cache go test ./...
```

## Docker

Build da imagem e execução:

```bash
docker build -t cepweather .
docker run --rm -p 8080:8080 -e WEATHER_API_KEY=coloque_sua_chave_aqui cepweather
```

### Docker Compose

```bash
export WEATHER_API_KEY=coloque_sua_chave_aqui
docker compose up --build
```

## Deploy no Google Cloud Run

1. Autentique-se no Google Cloud e selecione o projeto desejado:
   ```bash
   gcloud auth login
   gcloud config set project SEU_PROJETO
   ```
2. Construa e publique a imagem no Artifact Registry (ou Container Registry):
   ```bash
   gcloud builds submit --tag REGION-docker.pkg.dev/SEU_PROJETO/REPOSITORIO/cepweather .
   ```
3. Faça o deploy no Cloud Run:
   ```bash
   gcloud run deploy cepweather \
     --image REGION-docker.pkg.dev/SEU_PROJETO/REPOSITORIO/cepweather \
     --platform managed \
     --region REGION \
     --allow-unauthenticated \
     --set-env-vars WEATHER_API_KEY=coloque_sua_chave_aqui
   ```
4. Anote o `Service URL` retornado. Esse será o endpoint público da API.

### Teste pós-deploy

```bash
curl https://SEU_ENDPOINT/weather/01001000
```
