#!/usr/bin/env bash
set -euo pipefail

cd -- "$(dirname -- "${BASH_SOURCE[0]}")"

command -v go >/dev/null || {
    echo "ERROR: Go no está instalado."
    exit 1
}

command -v make >/dev/null || {
    echo "ERROR: make no está instalado."
    exit 1
}

docker compose version >/dev/null
docker info >/dev/null

cleanup() {
    echo "==> Eliminando contenedores y volúmenes del proyecto..."
    docker compose down -v --remove-orphans
}

trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

echo "==> Generando código con sqlc y templ..."
make generate

echo "==> Compilando la aplicación dentro de Docker..."
docker compose build app

echo "==> Eliminando el entorno anterior..."
docker compose down -v --remove-orphans

echo "==> Levantando PostgreSQL..."
docker compose up -d db

echo "==> Esperando a que PostgreSQL esté disponible..."
ready=0

for ((i = 1; i <= 60; i++)); do
    if docker compose exec -T db \
        pg_isready -h localhost -U postgres -d tpe_pw_test \
        >/dev/null 2>&1; then
        ready=1
        break
    fi
    sleep 1
done

if [ "$ready" -ne 1 ]; then
    echo "ERROR: PostgreSQL no estuvo disponible a tiempo."
    docker compose logs db
    exit 1
fi

echo "==> Iniciando la aplicación en http://localhost:8080"
echo "==> Para finalizar y limpiar el entorno, presioná Ctrl+C."

docker compose up --no-build --abort-on-container-exit --exit-code-from app
