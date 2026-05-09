# Payment Service

## Sumário

- [Contexto](#contexto)
- [Stack Tecnológica](#stack-tecnológica)
- [Responsabilidade](#responsabilidade)
- [Integração](#integração)
- [Arquitetura](#arquitetura)
- [Eventos](#eventos)
- [API](#api)
- [Execução](#execução)

---

## Contexto

Este serviço é responsável pelo gerenciamento do fluxo de pagamentos da oficina.

Ele recebe solicitações de criação de pagamento, cria registros de pagamento, gera links de pagamento via provedor externo e processa atualizações de status recebidas por webhook.

Características principais:

- Serviço orientado a eventos
- Integração com SNS/SQS
- Integração com Mercado Pago
- Persistência em DynamoDB
- API HTTP com Fiber
- Implementado seguindo Arquitetura Hexagonal

---

## Stack Tecnológica

- Linguagem: Go
- Framework HTTP: Fiber
- Banco: DynamoDB
- Mensageria: AWS SNS/SQS
- Provedor de pagamento: Mercado Pago Checkout Pro
- Infra local: Docker Compose + LocalStack
- Arquitetura: Hexagonal Architecture

---

## Responsabilidade

O serviço de payment é responsável exclusivamente por:

1. Criar pagamentos pendentes
2. Persistir pagamentos
3. Gerar links de pagamento
4. Publicar eventos de mudança de status
5. Processar webhooks do Mercado Pago
6. Atualizar o status do pagamento

Ele não é responsável por regras de negócio da oficina, pedido ou comunicação com o cliente.

---

## Integração

O serviço se integra com:

- SNS/SQS para consumo e publicação de eventos
- Mercado Pago para criação e consulta de pagamentos
- DynamoDB para persistência
- API HTTP para recebimento de webhooks

---

## Fluxo Simplificado

```text
Order Service
│
│ payment requested event
▼
SNS/SQS
│
▼
Payment Service
│
├─ Cria pagamento pendente
├─ Persiste no DynamoDB
├─ Gera link no Mercado Pago
├─ Atualiza pagamento com o link
▼
Publica evento de pagamento atualizado
```

## Fluxo de Webhook

```text
Mercado Pago
│
│ POST /webhooks/mercado-pago
▼
Payment Service
│
├─ Valida payload recebido
├─ Consulta pagamento no Mercado Pago
├─ Mapeia status externo para status interno
├─ Atualiza pagamento
▼
Publica evento de status alterado
```

## Eventos
### Consumidos

**Payment Requested**

Evento utilizado para iniciar a criação de um pagamento.

```json
{
  "id": "019db1c5-e000-7874-a7d1-909d361bd7c9",
  "amount": 10000,
  "description": "Pagamento do pedido"
}
```

**Create Payment Link**

Evento utilizado para gerar o link de pagamento de um pagamento já criado.
```json
{
  "id": "019db1c5-e000-7874-a7d1-909d361bd7c9",
  "status": "PENDING"
}
```

### Publicados
**Payment Status Changed**

Evento publicado quando o status do pagamento é alterado.

```json
{
  "id": "019db1c5-e000-7874-a7d1-909d361bd7c9",
  "external_id": "019db1c5-e000-7874-a7d1-909d361bd7c9",
  "status": "SUCCEEDED",
  "payment_url": "https://mercadopago.com/checkout/v1/..."
}
```

## API
**POST** /webhooks/mercado-pago

Endpoint responsável por receber notificações do Mercado Pago.

Request aceito
```json
{
  "resource": "https://api.mercadolibre.com/merchant_orders/40165307899",
  "topic": "merchant_order"
}
```
Payloads diferentes podem ser ignorados ou rejeitados, conforme regra do handler.

**Response - 200 OK**
```json
{
  "status": "ok"
}
```

## Arquitetura
Este serviço segue os princípios de Arquitetura Hexagonal.

Resumo das camadas:

- **Domain**
  - Entidade Payment
  - Status de pagamento
  - Regras de transição de status
  - Validações de domínio
- **Application**
  - Caso de uso CreatePayment
  - Caso de uso CreateLink
  - Caso de uso ProcessPaymentStatus
  - Ports de repositório, publisher e provider
- **Infrastructure**
  - DynamoDB Repository
  - Mercado Pago Provider
  - SNS/SQS Publishers e Consumers
  - Fiber HTTP Handler
  - Bootstrap da aplicação

## Execução
Variáveis de Ambiente Necessárias
```env
AWS_ENDPOINT=http://localstack:4566
AWS_REGION=us-east-1

PAYMENT_TABLE_NAME=payments

PAYMENT_REQUEST_QUEUE=payment-request
PAYMENT_STATUS_QUEUE=payment-status

PAYMENT_STATUS_CHANGED_TOPIC=payment-status-changed

MERCADO_PAGO_ACCESS_TOKEN=
MERCADO_PAGO_WEBHOOK_URL=
```

### Executando Localmente
**Requisitos**
- Go instalado
- Docker
- Docker Compose
- LocalStack

**Subindo o ambiente**
```bash
make up
```
--- 

## Last Sonar Overview:
<img width="1668" height="825" alt="image" src="https://github.com/user-attachments/assets/cb9119f2-0d6d-4b42-a62a-28b21c822264" />