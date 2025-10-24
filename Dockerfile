# syntax=docker/dockerfile:1

FROM golang:1.22 AS builder
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/api

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /out/server ./server

ENV PORT=:8080
EXPOSE 8080

ENTRYPOINT ["/app/server"]
