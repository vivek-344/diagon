#!/bin/bash
set -e

psql \
  --username "$POSTGRES_USER" \
  --dbname "$POSTGRES_DB" \
  --command "REVOKE CREATE ON DATABASE \"$POSTGRES_DB\" FROM PUBLIC;"