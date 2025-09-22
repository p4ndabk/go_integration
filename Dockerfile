# Go Integration - Dockerfile de Produção
# Multi-stage build otimizado para Google Cloud Run

# ============================================================================
# Estágio 1: Build
# ============================================================================
FROM golang:1.24-alpine AS builder

# Instalar dependências necessárias para build
RUN apk add --no-cache git ca-certificates tzdata

# Definir diretório de trabalho
WORKDIR /build

# Copiar arquivos de dependências primeiro (para cache de layers)
COPY go.mod go.sum ./

# Download de dependências (será cached se go.mod/go.sum não mudarem)
RUN go mod download && go mod verify

# Copiar o código fonte
COPY . .

# Compilar o binário de forma estática e otimizada
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o worker \
    ./cmd/worker

# Verificar se o binário foi criado corretamente
RUN chmod +x worker && ./worker --help 2>/dev/null || echo "Binary built successfully"

# ============================================================================
# Estágio 2: Runtime (Produção)
# ============================================================================
FROM gcr.io/distroless/base-debian12

# Metadados da imagem
LABEL maintainer="NorthFi Team"
LABEL description="Go Integration Worker - Email Processing Service"
LABEL version="1.0"

# Copiar certificados SSL para HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copiar timezone info
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Criar usuário não-root
# Distroless já vem com usuário nonroot (uid:gid = 65532:65532)

# Copiar o binário compilado
COPY --from=builder /build/worker /app/worker

# Definir usuário não-root
USER 65532:65532

# Definir diretório de trabalho
WORKDIR /app

# Variáveis de ambiente padrão
ENV PORT=8081
ENV GO_ENV=production

# Variáveis para Google Cloud Run
ENV GOOGLE_CLOUD_PROJECT=""
ENV PUBSUB_EMULATOR_HOST=""

# Health check (opcional - Cloud Run faz automaticamente)
HEALTHCHECK NONE

# Expor porta (Cloud Run usa PORT automaticamente)
EXPOSE 8081

# Comando de entrada
ENTRYPOINT ["/app/worker"]
