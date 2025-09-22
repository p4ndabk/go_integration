# Arquitetura do Worker - Go Integration

## Visão Geral

Este projeto implementa um worker assíncrono para processamento de emails utilizando Google Cloud Pub/Sub, seguindo as melhores práticas de arquitetura Go.

## Estrutura de Arquivos

```
go_integration/
├── cmd/queue/main.go              # Ponto de entrada - apenas inicialização
├── internal/
│   ├── config/config.go           # Configuração centralizada
│   ├── email/
│   │   ├── service.go            # Interface e implementação do serviço de email
│   │   ├── templates.go          # Templates HTML para emails
│   │   └── retry.go              # Utilitários de retry e resiliência
│   ├── handlers/
│   │   └── queue.go              # Handlers para processamento de mensagens
│   ├── models/
│   │   └── payloads.go           # Estruturas de dados
│   └── pubsub/
│       └── client.go             # Cliente Pub/Sub
├── .infra/                       # Infraestrutura Docker/Cloud Run
├── .env                          # Configurações ambiente
└── DESENVOLVIMENTO.md            # Guia de desenvolvimento
```

## Padrões Arquiteturais Implementados

### 1. Clean Architecture
- **Separação de responsabilidades**: handlers, serviços, e utilitários em pacotes distintos
- **Dependency Injection**: interfaces e implementações desacopladas
- **Single Responsibility**: cada arquivo/função tem uma única responsabilidade

### 2. Error Handling e Resilience
- **Retry Pattern**: função genérica `ExecuteWithRetry` para operações que podem falhar
- **Structured Logging**: logs padronizados com `slog`
- **Graceful Shutdown**: tratamento adequado de sinais do sistema

### 3. Template Pattern
- **Templates HTML centralizados**: welcome, verificação, e padrão
- **Configuração flexível**: fácil alteração de templates
- **Reusabilidade**: templates podem ser usados em diferentes contextos

## Componentes Principais

### main.go
- **Responsabilidade**: Apenas inicialização e orquestração
- **Características**:
  - Setup de logging estruturado
  - Inicialização de serviços
  - Configuração de graceful shutdown
  - Delegação de processamento para handlers

### internal/handlers/queue.go
- **EmailQueueHandler**: Struct centralizada para todos os handlers
- **Métodos especializados**:
  - `HandleEmailMessage`: emails regulares
  - `HandleWelcomeMessage`: emails de boas-vindas
  - `HandleVerificationMessage`: emails de verificação
  - `HandleUserMessage`: criação de usuários
- **Características**:
  - Retry automático usando `ExecuteWithRetry`
  - Logging estruturado consistente
  - Error handling padronizado

### internal/email/retry.go
- **ExecuteWithRetry**: Função genérica para retry com backoff (REMOVIDA - agora no handlers)
- **RetryConfig**: Configuração de tentativas e delays (REMOVIDA - agora no handlers)
- **IsWelcomeSubject**: Validação de tipos de email
- **Características**:
  - Função de retry movida para handlers internos
  - Validação de subject centralizada
  - Utilitários específicos do domínio email

### internal/email/templates.go
- **Templates HTML responsivos**:
  - `GetWelcomeEmailHTML`: Email de boas-vindas
  - `GetVerificationEmailHTML`: Email de verificação
  - `GetDefaultEmailHTML`: Template padrão
- **Características**:
  - Design moderno e responsivo
  - Fácil customização
  - Branding consistente

## Fluxos de Processamento

### 1. Email Regular
```
Pub/Sub → EmailQueueHandler.HandleEmailMessage → ExecuteWithRetry → ResendService → Template Padrão
```

### 2. Email de Boas-vindas
```
Pub/Sub → EmailQueueHandler.HandleUserMessage → HandleWelcomeMessage → ExecuteWithRetry → ResendService → Template Welcome
```

### 3. Email de Verificação
```
Pub/Sub → EmailQueueHandler.HandleVerificationMessage → ExecuteWithRetry → ResendService → Template Verification (Code/URL)
```

## Vantagens da Arquitetura

### 1. Manutenibilidade
- Código organizado e previsível
- Fácil localização de funcionalidades
- Testes unitários facilitados

### 2. Escalabilidade
- Handlers independentes
- Retry configurável por operação
- Logging estruturado para monitoramento

### 3. Flexibilidade
- Templates HTML facilmente customizáveis
- Novos tipos de email fáceis de adicionar
- Troca de providers de email sem impacto

### 4. Observabilidade
- Logs estruturados em JSON
- Métricas de retry e falhas
- Contexto completo em cada operação

## Próximos Passos

1. **Monitoramento**: Integrar métricas com Prometheus/Grafana
2. **Testes**: Implementar testes unitários e de integração
3. **Batch Processing**: Processar múltiplas mensagens em lote
4. **Rate Limiting**: Controlar taxa de envio de emails
5. **Dead Letter Queue**: Tratar mensagens que falharam todas as tentativas

## Deploy

A aplicação está preparada para deploy no Google Cloud Run com:
- Health check endpoint
- Graceful shutdown
- Variáveis de ambiente configuradas
- Docker multi-stage build otimizado
