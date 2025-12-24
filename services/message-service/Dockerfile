# ============================================
# Multi-stage Dockerfile otimizado para tamanho mínimo
# Resultado final: ~15-20MB (com scratch) ou ~25MB (com alpine)
# ============================================

# Stage 1: Builder
FROM golang:1.23-alpine AS builder

# Instalar dependências de build
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copiar go.mod e go.sum primeiro (cache de dependências)
COPY go.mod go.sum ./

# Download de dependências (cacheable)
RUN go mod download && go mod verify

# Copiar código fonte
COPY . .

# Build com flags de otimização máxima
# -ldflags: remove informações de debug e path
# -trimpath: remove caminhos absolutos do binário
# CGO_ENABLED=0: binário estático sem dependências C
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.Version=1.0.0" \
    -trimpath \
    -o message-service \
    ./cmd/main.go

# Verificar tamanho do binário
RUN ls -lh message-service && file message-service

# Opcional: Comprimir ainda mais com UPX (descomente se quiser)
# RUN apk add --no-cache upx && upx --best --lzma message-service

# ============================================
# Stage 2: Runtime (opção 1 - scratch - menor)
# ============================================
FROM scratch AS runtime-scratch

# Copiar timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copiar certificados SSL
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copiar binário
COPY --from=builder /build/message-service /message-service

# Expor porta
EXPOSE 3001

# Executar aplicação
ENTRYPOINT ["/message-service"]

# ============================================
# Stage 3: Runtime (opção 2 - alpine - com shell para debug)
# ============================================
FROM alpine:3.19 AS runtime-alpine

# Adicionar usuário não-root para segurança
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Instalar apenas certificados SSL
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copiar binário
COPY --from=builder /build/message-service .

# Mudar ownership
RUN chown -R appuser:appuser /app

# Usar usuário não-root
USER appuser

# Expor porta
EXPOSE 3001

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/message-service", "--health"] || exit 1

# Executar aplicação
ENTRYPOINT ["/app/message-service"]

# ============================================
# Stage final: usar scratch por padrão (menor tamanho)
# Para usar alpine: docker build --target runtime-alpine -t message-service .
# ============================================
FROM runtime-scratch
