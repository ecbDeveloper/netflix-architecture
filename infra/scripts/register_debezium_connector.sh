#!/bin/sh

echo "Waiting for Kafka Connect to be ready at ${KAFKA_CONNECT_HOST:-localhost}:${KAFKA_CONNECT_PORT:-8083}..."
until curl -s -o /dev/null -w "%{http_code}" "http://${KAFKA_CONNECT_HOST:-localhost}:${KAFKA_CONNECT_PORT:-8083}/connectors" | grep -E "^(200|409)$" > /dev/null; do
  echo "Kafka Connect is not ready yet. Retrying in 3 seconds..."
  sleep 3
done
echo "Kafka Connect is ready! Registering Debezium connector..."

curl -i -X POST -H "Accept:application/json" -H "Content-Type:application/json" \
  "http://${KAFKA_CONNECT_HOST:-localhost}:${KAFKA_CONNECT_PORT:-8083}/connectors/" \
  -d @- <<EOF
{
  "name": "history-connector",
  "config": {
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "tasks.max": "1",
    "database.hostname": "${HISTORY_DB_HOST:-localhost}",
    "database.port": "${HISTORY_DB_PORT}",
    "database.user": "${HISTORY_DB_USER}",
    "database.password": "${HISTORY_DB_PASS}",
    "database.dbname": "${HISTORY_DB_NAME}",
    "topic.prefix": "netflix_history",
    "plugin.name": "pgoutput"
  }
}
EOF

echo -e "\nDebezium connector registered!"
