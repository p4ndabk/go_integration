# Go Pub/Sub Integration

Sistema de pub/sub em Go usando Google Cloud Pub/Sub com arquitetura modular seguindo padrões da comunidade Go.

## 📁 Estrutura do Projeto

```
go_integration/
├── cmd/
│   ├── api/             # API HTTP para envio de emails
│   │   └── main.go
│   └── queue/           # Worker para processar mensagens
│       └── main.go
├── internal/            # Código privado da aplicação
│   ├── config/         # Configuração centralizada
│   │   └── config.go
│   ├── email/          # Serviço de email
│   │   └── service.go
│   ├── handlers/       # Handlers HTTP
│   │   └── email.go
│   ├── models/         # Modelos de dados
│   │   ├── email.go
│   │   └── errors.go
│   └── pubsub/         # Cliente Pub/Sub
│       └── client.go
├── docker-compose.yml   # Orquestração local
├── Dockerfile          # Imagem da aplicação
├── .env               # Variáveis de ambiente
├── .dockerignore      # Arquivos ignorados no Docker
├── go.mod             # Dependências Go
└── README.md          # Documentação
```

## 🏗️ Arquitetura

### Padrão Internal/Pkg
- **`internal/`**: Código privado que não pode ser importado por outros projetos
- **`cmd/`**: Pontos de entrada da aplicação (main.go)
- Separação clara de responsabilidades entre camadas

### Componentes

#### 1. **internal/pubsub** - Cliente Pub/Sub
```go
// Wrapper para Google Cloud Pub/Sub
type Client struct {
    client    *pubsub.Client
    projectID string
}

// Funcionalidades:
- EnsureTopic()      // Cria tópico se não existir
- EnsureSubscription() // Cria subscription se não existir
- Receive()          // Recebe e processa mensagens
```

#### 2. **internal/models** - Modelos de Dados
```go
type EmailPayload struct {
    To      string `json:"to"`
    Subject string `json:"subject"`
    Body    string `json:"body"`
}
```

#### 3. **internal/email** - Serviço de Email
```go
// Publica mensagens no Pub/Sub
func PublishEmail(ctx context.Context, topic *pubsub.Topic, payload *models.EmailPayload) error

// Processa mensagens recebidas
func ProcessMessage(ctx context.Context, msg *pubsub.Message, handler func(context.Context, *models.EmailPayload) error) error
```

#### 4. **internal/handlers** - Handlers HTTP
```go
func SendEmail(emailService *email.Service) http.HandlerFunc
```

#### 5. **internal/config** - Configuração
```go
type Config struct {
    ProjectID string
    TopicID   string
    SubID     string
    Port      string
}
```

## 🚀 Como Usar

### 1. Configuração

Crie o arquivo `.env`:
```bash
PUBSUB_PROJECT_ID=go-integration-project
PUBSUB_TOPIC_ID=send-email
PUBSUB_SUB_ID=send-email-sub
PORT=8080
```

### 2. Executar com Docker (Recomendado)

```bash
# Inicia emulador + aplicação
docker-compose up --build

# Logs separados
docker-compose logs -f api
docker-compose logs -f queue
```

### 3. Executar Localmente

```bash
# Terminal 1 - Emulador
docker run --rm -p 8085:8085 gcr.io/google.com/cloudsdktool/cloud-sdk:latest \
  gcloud beta emulators pubsub start --host-port=0.0.0.0:8085

# Terminal 2 - API
export PUBSUB_EMULATOR_HOST=localhost:8085
go run cmd/api/main.go

# Terminal 3 - Queue Worker
export PUBSUB_EMULATOR_HOST=localhost:8085
go run cmd/queue/main.go
```

### 4. Testando a API

```bash
# Enviar email
curl -X POST http://localhost:8080/send-email \
  -H "Content-Type: application/json" \
  -d '{
    "to": "usuario@exemplo.com",
    "subject": "Teste Pub/Sub",
    "body": "Mensagem de teste do sistema pub/sub"
  }'

# Resposta
{"message":"Email enviado com sucesso"}
```

## 🔄 Fluxo de Funcionamento

### Envio de Email
1. **API** recebe POST `/send-email`
2. **Handler** valida payload
3. **Service** publica no tópico Pub/Sub
4. Retorna sucesso para cliente

### Processamento
1. **Queue Worker** escuta subscription
2. Recebe mensagens do tópico
3. Deserializa `EmailPayload`
4. Processa email (simula envio)
5. Confirma processamento (ACK)

```
Cliente → API → Pub/Sub Topic → Subscription → Queue Worker → Email Service
```

## 📊 Monitoramento

### Logs da API
```
2024/01/15 10:30:15 Server starting on :8080
2024/01/15 10:30:20 Email sent successfully to: usuario@exemplo.com
```

### Logs do Worker
```
2024/01/15 10:30:15 Starting to receive messages from subscription: send-email-sub
📧 Email enviado para: usuario@exemplo.com
   Assunto: Teste Pub/Sub
   Mensagem: Mensagem de teste do sistema pub/sub
```

## 🛠️ Desenvolvimento

### Comandos Úteis

```bash
# Instalar dependências
go mod tidy

# Executar testes
go test ./...

# Build
go build -o bin/api cmd/api/main.go
go build -o bin/queue cmd/queue/main.go

# Linting
golangci-lint run

# Verificar módulos
go mod verify
```

### Estrutura de Testes
```bash
# Testes unitários
go test ./internal/...

# Testes de integração
go test ./cmd/...

# Coverage
go test -cover ./...
```

## 🐳 Docker

### Build Manual
```bash
docker build -t go-pubsub-api .
docker build -t go-pubsub-queue .
```

### Variáveis de Ambiente
- `PUBSUB_PROJECT_ID`: ID do projeto GCP
- `PUBSUB_TOPIC_ID`: Nome do tópico
- `PUBSUB_SUB_ID`: Nome da subscription  
- `PORT`: Porta da API (padrão: 8080)
- `PUBSUB_EMULATOR_HOST`: Host do emulador (desenvolvimento)

## 🎯 Vantagens desta Arquitetura

### Modularidade
- Cada pacote tem responsabilidade única
- Fácil teste e manutenção
- Reutilização de código

### Escalabilidade  
- API e workers independentes
- Processamento assíncrono
- Horizontal scaling

### Observabilidade
- Logs estruturados
- Métricas por componente
- Rastreamento de erros

### Padrões Go
- Arquitetura internal/pkg
- Context propagation
- Error handling idiomático
- Graceful shutdown

## 📝 Próximos Passos

1. **Testes**: Adicionar testes unitários e integração
2. **Métricas**: Implementar Prometheus metrics
3. **Tracing**: Adicionar OpenTelemetry
4. **CI/CD**: Pipeline GitHub Actions
5. **Deploy**: Configuração Kubernetes
6. **Monitoring**: Health checks e alertas

## 🤝 Contribuição

1. Fork o projeto
2. Crie feature branch (`git checkout -b feature/amazing-feature`)
3. Commit (`git commit -m 'Add amazing feature'`)
4. Push (`git push origin feature/amazing-feature`)  
5. Abra Pull Request
