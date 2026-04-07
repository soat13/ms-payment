#!/bin/bash
set -euo pipefail

REGION="us-east-1"

QUEUE_REPAIRORDER_FINISHED="repairorder-finished"
QUEUE_PAYMENT_STATUS="payment-status"
TOPIC_PAYMENT_STATUS="payment-status-topic"

DYNAMODB_TABLE="payments"
DYNAMODB_TABLE_TEST="payments_test"
DYNAMODB_GSI1="gsi1"

create_dynamodb_table_if_not_exists() {
  local table_name="$1"

  if awslocal dynamodb describe-table \
    --table-name "$table_name" \
    --region "$REGION" >/dev/null 2>&1; then
    echo "Table $table_name already exists, skipping..."
  else
    awslocal dynamodb create-table \
      --table-name "$table_name" \
      --attribute-definitions \
        AttributeName=pk,AttributeType=S \
        AttributeName=sk,AttributeType=S \
        AttributeName=gsi1pk,AttributeType=S \
        AttributeName=gsi1sk,AttributeType=S \
      --key-schema \
        AttributeName=pk,KeyType=HASH \
        AttributeName=sk,KeyType=RANGE \
      --global-secondary-indexes "[
        {
          \"IndexName\": \"$DYNAMODB_GSI1\",
          \"KeySchema\": [
            {\"AttributeName\": \"gsi1pk\", \"KeyType\": \"HASH\"},
            {\"AttributeName\": \"gsi1sk\", \"KeyType\": \"RANGE\"}
          ],
          \"Projection\": {
            \"ProjectionType\": \"ALL\"
          }
        }
      ]" \
      --billing-mode PAY_PER_REQUEST \
      --region "$REGION" >/dev/null

    echo "Table $table_name created with GSI $DYNAMODB_GSI1"
  fi
}

echo "Checking/creating SQS queues..."

if awslocal sqs get-queue-url \
  --queue-name "$QUEUE_REPAIRORDER_FINISHED" \
  --region "$REGION" >/dev/null 2>&1; then
  echo "Queue $QUEUE_REPAIRORDER_FINISHED already exists, skipping..."
else
  awslocal sqs create-queue \
    --queue-name "$QUEUE_REPAIRORDER_FINISHED" \
    --region "$REGION" >/dev/null
  echo "Queue $QUEUE_REPAIRORDER_FINISHED created"
fi

if awslocal sqs get-queue-url \
  --queue-name "$QUEUE_PAYMENT_STATUS" \
  --region "$REGION" >/dev/null 2>&1; then
  echo "Queue $QUEUE_PAYMENT_STATUS already exists, skipping..."
else
  awslocal sqs create-queue \
    --queue-name "$QUEUE_PAYMENT_STATUS" \
    --region "$REGION" >/dev/null
  echo "Queue $QUEUE_PAYMENT_STATUS created"
fi

echo "Checking/creating SNS topic..."

TOPIC_ARN=$(awslocal sns create-topic \
  --name "$TOPIC_PAYMENT_STATUS" \
  --region "$REGION" \
  --query 'TopicArn' \
  --output text)

echo "Topic ready: $TOPIC_ARN"

echo "Fetching queue URL and ARN..."

PAYMENT_STATUS_QUEUE_URL=$(awslocal sqs get-queue-url \
  --queue-name "$QUEUE_PAYMENT_STATUS" \
  --region "$REGION" \
  --query 'QueueUrl' \
  --output text)

PAYMENT_STATUS_QUEUE_ARN=$(awslocal sqs get-queue-attributes \
  --queue-url "$PAYMENT_STATUS_QUEUE_URL" \
  --attribute-names QueueArn \
  --region "$REGION" \
  --query 'Attributes.QueueArn' \
  --output text)

echo "Applying queue policy to allow SNS topic to publish..."

POLICY=$(cat <<EOF
{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"sqs:SendMessage","Resource":"$PAYMENT_STATUS_QUEUE_ARN","Condition":{"ArnEquals":{"aws:SourceArn":"$TOPIC_ARN"}}}]}
EOF
)

ESCAPED_POLICY=$(printf '%s' "$POLICY" | python3 -c 'import json,sys; print(json.dumps(sys.stdin.read()))')

awslocal sqs set-queue-attributes \
  --queue-url "$PAYMENT_STATUS_QUEUE_URL" \
  --attributes "{\"Policy\":$ESCAPED_POLICY}" \
  --region "$REGION"

echo "Checking SNS subscription..."

EXISTING_SUBSCRIPTION_ARN=$(awslocal sns list-subscriptions-by-topic \
  --topic-arn "$TOPIC_ARN" \
  --region "$REGION" \
  --query "Subscriptions[?Endpoint=='$PAYMENT_STATUS_QUEUE_ARN'].SubscriptionArn | [0]" \
  --output text)

if [ "$EXISTING_SUBSCRIPTION_ARN" = "None" ] || [ -z "$EXISTING_SUBSCRIPTION_ARN" ]; then
  awslocal sns subscribe \
    --topic-arn "$TOPIC_ARN" \
    --protocol sqs \
    --notification-endpoint "$PAYMENT_STATUS_QUEUE_ARN" \
    --region "$REGION" >/dev/null
  echo "Queue $QUEUE_PAYMENT_STATUS subscribed to topic $TOPIC_PAYMENT_STATUS"
else
  echo "Subscription already exists, skipping..."
fi

echo "Checking/creating DynamoDB tables..."

create_dynamodb_table_if_not_exists "$DYNAMODB_TABLE"
create_dynamodb_table_if_not_exists "$DYNAMODB_TABLE_TEST"

echo "Done."
echo "Resources ready:"
echo "- SQS queue: $QUEUE_REPAIRORDER_FINISHED"
echo "- SQS queue: $QUEUE_PAYMENT_STATUS"
echo "- SNS topic: $TOPIC_PAYMENT_STATUS"
echo "- DynamoDB table (dev): $DYNAMODB_TABLE"
echo "- DynamoDB table (test): $DYNAMODB_TABLE_TEST"
echo "- DynamoDB GSI: $DYNAMODB_GSI1"