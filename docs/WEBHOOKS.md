# Configuração de Webhooks de Entrada (Incoming Webhooks)

O Chatwoot-Go aceita webhooks de serviços externos (como Evolution API, WaAPI, etc.) para processar mensagens recebidas.

## Estrutura da URL

A URL para configuração do webhook suporta dois formatos:

1.  **Explícito (Recomendado - Padrão Original):**

    ```
    POST /api/v1/webhooks/incoming/:account_id/:instance
    ```

    Onde `:account_id` é o ID da conta (UUID ou ID Numérico se suportado) e `:instance` é o nome da instância.

2.  **Simplificado:**
    ```
    POST /api/v1/webhooks/incoming/:instance
    ```
    Neste caso, o sistema usará a primeira conta disponível do usuário autenticado.

## Autenticação (Token)

Para ativar o webhook, é necessário um **Token de Acesso**.
Você pode gerar um token na página de **Configurações > Token de Acesso**.

O token pode ser passado de duas formas:

1.  **Via Header (Recomendado):**

    - Header: `X-Api-Token` ou `api_access_token`
    - Valor: `SeuTokenAqui`

2.  **Via Query String:**
    - Adicione `?api_token=SeuTokenAqui` ao final da URL.

## Exemplo Completo

**URL:** `https://seu-chatwoot.com/api/v1/webhooks/incoming/1/WhatsappPrincipal?api_token=v1_xxxx`

**Componentes:**

1.  **URL Base:** `/api/v1/webhooks/incoming`
2.  **ID da Conta:** `1` (ou UUID da conta)
3.  **Identificador (+1):** `WhatsappPrincipal` (nome da Inbox)
4.  **Token:** `v1_xxxx` (Autenticação do Usuário)

## Lógica Interna

1.  O sistema valida o **Token** e identifica o **Usuário**.
2.  O sistema identifica a **Conta** principal do Usuário.
3.  O sistema procura uma **Inbox** com o nome `MinhaInstancia` nesta Conta.
4.  Se não existir, cria uma nova Inbox do tipo Whatsapp.
5.  A mensagem é processada e associada a um Contato e Conversa dentro desta Inbox.
