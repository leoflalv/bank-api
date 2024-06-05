#!/bin/sh

set -e

echo "run db migrations"
/app/migrate -path /app/migrations -database postgresql://root:1dzjA7F0BA8QwXeTRXw2@bank.c3y0equmgn8w.eu-west-3.rds.amazonaws.com:5432/bank -verbose up

echo "start the app"
exec "$@"
