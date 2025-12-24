# 🔍 Diagnóstico: Contatos Não Persistem no Banco

## ✅ Correções Aplicadas

### 1. Handlers Update e Delete Implementados
- [handlers.go:425-491](backend/internal/handlers/handlers.go#L425-L491)
- Update agora aceita campos parciais (name, email, phone_number)
- Delete implementa soft delete (usa DeletedAt do GORM)
- Ambos validam permissões (account_id)

---

## 🧪 Passo a Passo para Diagnosticar

### Etapa 1: Verificar se Backend Está Rodando

```bash
# No terminal do backend
cd backend
go run cmd/server/main.go
```

**Esperado:**
```
✅ Database connected successfully
✅ Minio storage initialized
🚀 Server starting on port 8080
```

---

### Etapa 2: Testar Criação Direto na API

Abra o Postman/Insomnia ou use curl:

```bash
# 1. Fazer login para obter token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "seu-email@example.com",
    "password": "sua-senha"
  }'
```

**Resposta esperada:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": { ... }
}
```

```bash
# 2. Criar contato usando o token
curl -X POST http://localhost:8080/api/v1/contacts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer SEU_TOKEN_AQUI" \
  -d '{
    "name": "Teste API",
    "email": "teste@api.com",
    "phone_number": "+5511999999999"
  }'
```

**Respostas possíveis:**

✅ **Sucesso (201):**
```json
{
  "id": "uuid-aqui",
  "name": "Teste API",
  "email": "teste@api.com",
  "phone_number": "+5511999999999",
  "account_id": "uuid-da-conta",
  "created_at": "2025-12-23T..."
}
```

❌ **Erro 400 (Bad Request):**
```json
{
  "error": "mensagem de erro de validação"
}
```
**Causa:** Dados inválidos ou faltando campos

❌ **Erro 401 (Unauthorized):**
```json
{
  "error": "Unauthorized"
}
```
**Causa:** Token inválido ou expirado

❌ **Erro 500 (Internal Server Error):**
```json
{
  "error": "Failed to create contact: erro do banco"
}
```
**Causa:** Problema no banco de dados

---

### Etapa 3: Verificar Logs do Backend

Olhe o terminal do backend após tentar criar contato:

**Logs esperados (sucesso):**
```
ContactHandler.List: accountID=uuid-aqui
Fetching contacts page: 1
>>> DEBUG CREATE CONTACT ERROR: (não deve aparecer)
```

**Logs de erro (problema):**
```
>>> DEBUG CREATE CONTACT ERROR: pq: relation "contacts" does not exist
```
**Solução:** Rodar migrations

```
>>> DEBUG CREATE CONTACT ERROR: pq: null value in column "account_id"
```
**Solução:** Middleware de autenticação não está passando account_id

---

### Etapa 4: Verificar Diretamente no Banco

**Se usar PostgreSQL local:**
```bash
psql -U chatwoot -d chatwoot_go

-- Verificar tabela existe
\dt contacts

-- Ver contatos
SELECT id, name, email, phone_number, account_id, created_at
FROM contacts
ORDER BY created_at DESC
LIMIT 10;
```

**Se usar Supabase:**
1. Abra o painel do Supabase
2. Vá em "Table Editor"
3. Selecione tabela "contacts"
4. Verifique se há registros

---

### Etapa 5: Testar Frontend → Backend

Abra DevTools do navegador (F12) → aba Network:

1. Criar contato no frontend
2. Procurar requisição POST `/api/v1/contacts`
3. Verificar:
   - **Request Headers:** Tem `Authorization: Bearer ...`?
   - **Request Payload:** Dados estão corretos?
   - **Response Status:** 201, 400, 401 ou 500?
   - **Response Body:** Qual erro retornou?

---

## 🔧 Problemas Comuns e Soluções

### Problema 1: "Unauthorized" ou "token missing"

**Sintoma:** Frontend retorna erro 401

**Causa:** Token não está sendo enviado ou é inválido

**Solução:**
```typescript
// Verificar em frontend/src/stores/contactStore.ts:84
const token = useAuthStore.getState().token
console.log('Token:', token) // Deve aparecer no console

// Se token for null, fazer login novamente
```

---

### Problema 2: "relation contacts does not exist"

**Sintoma:** Erro no backend ao criar contato

**Causa:** Migrations não foram executadas

**Solução:**
```bash
cd backend
go run cmd/server/main.go
# Migrations rodam automaticamente ao iniciar
```

Ou execute migrations manualmente:
```sql
-- Ver database/migrations.go para schema completo
CREATE TABLE IF NOT EXISTS contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL,
    name VARCHAR(255),
    email VARCHAR(255),
    phone_number VARCHAR(50),
    avatar VARCHAR(500),
    identifier VARCHAR(255),
    custom_attributes JSONB,
    additional_attributes JSONB,
    last_activity_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);
```

---

### Problema 3: Contato Criado mas Não Aparece na Lista

**Sintoma:** POST retorna 201, mas GET retorna lista vazia

**Causa 1:** account_id diferente
```bash
# Verificar no banco
SELECT id, name, account_id FROM contacts;

# Comparar com account_id do usuário logado
SELECT id, email FROM users WHERE email = 'seu-email@example.com';
SELECT account_id FROM account_users WHERE user_id = 'user-id-acima';
```

**Causa 2:** Soft delete ativo
```bash
# Verificar deleted_at
SELECT id, name, deleted_at FROM contacts;

# Se deleted_at não for NULL, o contato está "deletado"
UPDATE contacts SET deleted_at = NULL WHERE id = 'uuid-do-contato';
```

---

### Problema 4: Frontend Mostra Contato, Backend Não Tem

**Sintoma:** Lista aparece no frontend mas banco está vazio

**Causa:** Dados apenas no cache local do Zustand

**Solução:**
```typescript
// Limpar cache do navegador
localStorage.clear()

// Ou forçar refresh no código
fetchContacts(1, true) // force=true
```

---

## 📋 Checklist Completo

Marque cada item conforme testar:

- [ ] Backend está rodando sem erros
- [ ] Migrations executadas (tabela contacts existe)
- [ ] Login funciona e retorna token
- [ ] Token é válido (não expirou)
- [ ] POST /api/v1/contacts retorna 201
- [ ] Contato aparece no banco de dados
- [ ] GET /api/v1/contacts retorna o contato
- [ ] Frontend envia Authorization header
- [ ] account_id está correto
- [ ] deleted_at é NULL

---

## 🐛 Se Nada Funcionar

### Opção 1: Reset Completo

```bash
# 1. Parar backend
# 2. Limpar banco (cuidado!)
psql -U chatwoot -d chatwoot_go -c "DROP TABLE IF EXISTS contacts CASCADE;"

# 3. Reiniciar backend (recria tabelas)
cd backend
go run cmd/server/main.go

# 4. Limpar cache frontend
# DevTools → Application → Storage → Clear site data

# 5. Fazer login novamente
# 6. Tentar criar contato
```

### Opção 2: Logs Detalhados

Adicione mais logs temporários:

```go
// Em backend/internal/handlers/handlers.go:394
func (h *ContactHandler) Create(c *gin.Context) {
	accountIDStr := c.GetString("account_id")
	log.Printf("DEBUG: Creating contact, account_id=%s", accountIDStr)

	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		log.Printf("DEBUG: Failed to parse account_id: %v", err)
	}

	var input models.Contact
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("DEBUG: Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("DEBUG: Input data: %+v", input)
	input.AccountID = accountID

	if err := h.db.Create(&input).Error; err != nil {
		log.Printf(">>> DEBUG CREATE CONTACT ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create contact: " + err.Error()})
		return
	}

	log.Printf("DEBUG: Contact created successfully: %s", input.ID)
	c.JSON(http.StatusCreated, input)
}
```

---

## 📞 Próximos Passos

1. Execute os testes acima na ordem
2. Anote qual etapa falha
3. Verifique a seção "Problemas Comuns" correspondente
4. Se precisar, compartilhe:
   - Logs do backend
   - Erro específico do console do navegador
   - Status HTTP retornado

---

**Última Atualização:** 2025-12-23
