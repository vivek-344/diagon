#!/bin/bash
set -e

psql \
  --username "$POSTGRES_USER" \
  --dbname "$POSTGRES_DB" \
  --command "CREATE USER auth_user WITH PASSWORD '$AUTH_DB_PASSWORD';"

psql \
  --username "$POSTGRES_USER" \
  --dbname "$POSTGRES_DB" \
  --command "CREATE SCHEMA auth AUTHORIZATION auth_user;"

psql \
  --username "$POSTGRES_USER" \
  --dbname "$POSTGRES_DB" \
  --command "REVOKE ALL ON SCHEMA auth FROM PUBLIC;"