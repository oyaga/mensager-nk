# Mensager-NK (Chatwoot-Go)

Este projeto é uma plataforma de atendimento ao cliente inspirada no Chatwoot, reescrita com um backend de alta performance em **Go (Golang)** e um frontend moderno em **React**.

O objetivo é fornecer um sistema leve, rápido e fácil de implantar para gerenciar conversas de múltiplos canais (como WhatsApp via Evolution API).

---

## 🏗️ Arquitetura do Sistema

O sistema segue uma arquitetura monolítica modularizada, containerizada via Docker.

### Componentes Principais

1.  **Backend (API & WebSocket)**

    - **Linguagem**: Go (1.23+)
    - **Framework Web**: Gin
    - **ORM**: GORM (interagindo com PostgreSQL)
    - **Real-time**: WebSockets (Implementação customizada compatível com ActionCable)
    - **Função**: Gerencia autenticação, lógica de negócios, persistência de mensagens e comunicação em tempo real com o frontend.

2.  **Frontend (SPA)**

    - **Framework**: React (Vite)
    - **Linguagem**: TypeScript
    - **Gerenciamento de Estado**: Zustand (com persistência local)
    - **API Client**: Axios
    - **UI**: Tailwind CSS + Lucide Icons
    - **Função**: Interface do agente para responder mensagens, visualizar contatos e configurar o perfil.

3.  **Banco de Dados**

    - **PostgreSQL**: Armazena usuários, contas, contatos, conversas e mensagens.

4.  **Cache & Pub/Sub**

    - **Redis**: Utilizado para gerenciar sessões de WebSocket (pub/sub de eventos) e filas de processamento (se necessário no futuro).

5.  **Object Storage**
    - **MinIO**: Armazenamento compatível com S3 para uploads de arquivos (avatares, anexos de mensagens).

---

## 🚀 Como Rodar o Projeto

### Pré-requisitos

- Docker e Docker Compose

### Execução via Docker (Recomendado)

O projeto possui um arquivo `docker-compose.yml` que sobe toda a infraestrutura necessária (App + Postgres + Redis + MinIO).

```bash
# Na raiz do projeto
docker-compose up -d --build
```

O sistema estará disponível em: `http://localhost:8080`

### Credenciais Padrão (Ambiente Local)

- **Login**: admin@nakamura.com
- **Senha**: chatwoot123
- **Banco de Dados**: `postgres://chatwoot:chatwoot123@postgres:5432/chatwoot_go`

---

## 🛠️ Guia de Desenvolvimento

### Estrutura de Pastas

```
/
├── backend/                # Código fonte do servidor Go
│   ├── cmd/server/         # Ponto de entrada (main.go)
│   ├── internal/
│   │   ├── config/         # Carregamento de variáveis de ambiente
│   │   ├── handlers/       # Controladores HTTP (Auth, Chat, Webhook)
│   │   ├── models/         # Definições de Structs e Tabelas DB
│   │   ├── realtime/       # Lógica de WebSocket
│   │   └── services/       # Lógica de negócio (ex: MessageService)
│   └── go.mod              # Dependências Go
│
├── frontend/               # Código fonte da interface React
│   ├── src/
│   │   ├── components/     # Componentes Reutilizáveis (Modais, Paineis)
│   │   ├── pages/          # Páginas (Login, Chat, Configurações, Contatos)
│   │   ├── stores/         # Estados globais (AuthStore, ChatStore)
│   │   └── lib/            # Configuração do Axios e API
│
├── docs/                   # Documentação adicional
│   ├── WEBHOOKS.md         # Guia de integração de Webhooks
│   └── ...
│
├── docker-compose.yml      # Orquestração de containers
├── Dockerfile.local        # Build unificado (Backend + Frontend estático)
└── README.md               # Este arquivo
```

### Rodando Manualmente (Sem Docker para o App)

Se você quiser desenvolver e testar mudanças rapidamente sem rebuildar o Docker a cada vez:

1.  **Suba apenas a infraestrutura** (DB, Redis, MinIO):

    ```bash
    docker-compose up -d postgres redis minio createbuckets
    ```

2.  **Rode o Backend**:

    ```bash
    cd backend
    # Copie .env.example para .env e ajuste as credenciais se necessário
    go run cmd/server/main.go
    ```

3.  **Rode o Frontend**:
    ```bash
    cd frontend
    npm run dev
    ```

---

## 🔑 Funcionalidades Chave & Detalhes de Implementação

### 1. Webhooks de Entrada (Integração WhatsApp)

O sistema aceita mensagens de fontes externas (como Evolution API).

- **Rota**: `POST /webhooks/incoming/*pathParam`
- **Lógica**: A rota é "wildcard" para evitar conflitos. O sistema extrai o `account_id` da URL ou do token.
- **Autenticação**: Via Query Param `access_token` ou `account_id` na URL.
- Documentação detalhada em `docs/WEBHOOKS.md`.

### 2. Autenticação e Tokens

- **JWT**: Usado para sessões de login do usuário no frontend.
- **Access Token (API Key)**: Um token UUID estático gerado para cada usuário. Usado para autenticar webhooks externos e integrações API.
  - Visível em: `Configurações > Token de Acesso`.
  - Armazenado na tabela `users`, coluna `access_token`.

### 3. ID da Conta (Account ID)

Todo usuário pertence a uma "Conta" (Tenant).

- O **ID da Conta** é crucial para configurar a URL do Webhook.
- No Frontend, há um botão "Copiar ID da Conta" em Configurações que extrai esse ID diretamente do JWT ou do perfil carregado, com fallback visual.

### 4. Real-time (WebSockets)

- O frontend se conecta via WS em `/cable`.
- Eventos como `message.created`, `presence.update` são enviados pelo backend via Redis Pub/Sub ou diretamente pelo gerenciador de conexões em memória.

### 5. Upload de Arquivos

- Integrado com MinIO.
- Frontend faz upload para endpoint de API -> Backend salva no MinIO -> Retorna URL pública/assinada.

---

## 📝 Notas para Retomada (ToDo / Melhorias Futuras)

1.  **Fusão de Contatos**: A interface para "Mesclar Contatos" existe (`ContactDetailsPanel`), mas a lógica de backend ainda precisa ser refinada para unir históricos de conversas.
2.  **Canais Adicionais**: Atualmente focado em Webhooks genéricos/WhatsApp. Adicionar suporte nativo a E-mail ou Facebook.
3.  **Relatórios**: Implementar dashboard de métricas de atendimento.
4.  **Testes**: Aumentar cobertura de testes unitários no backend (`_test.go`).

---

## 🆘 Solução de Problemas Comuns

- **"Nenhuma conta associada encontrada"**:

  - Isso ocorre se o usuário foi criado manualmente no banco sem vínculo na tabela `account_users`.
  - _Correção_: O botão de copiar ID agora tenta extrair o ID do payload do JWT local como fallback. Se persistir, verifique a tabela `account_users`.

- **Erro de Conexão WebSocket**:

  - Verifique se o Redis está rodando. O WS depende do Redis para pub/sub.

- **Imagens não carregam**:
  - Verifique se o container MinIO está rodando e se a variável `MINIO_ENDPOINT` está acessível pelo navegador (cuidado com `localhost` vs `nome-do-container` dentro do Docker).

---

**Desenvolvido por Antigravity / Oyaga Tech**
