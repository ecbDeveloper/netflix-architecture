#!/bin/sh

# The script is executed by the init container.
# It expects the environment variables below to be injected.

echo "Waiting for Kafka Connect to be ready at ${KAFKA_CONNECT_HOST:-localhost}:${KAFKA_CONNECT_PORT:-8083}..."
sleep 15

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
