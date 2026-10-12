SHELL := /bin/bash

.PHONY: test test-api

test:
	@set -euo pipefail; \
	cleanup() { \
		echo "==> Limpiando contenedores y volúmenes..."; \
		docker compose down -v >/dev/null 2>&1 || true; \
	}; \
	trap cleanup EXIT; \
	echo "==> Generando código con sqlc..."; \
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate; \
	echo "==> Compilando proyecto..."; \
	go build ./...; \
	echo "==> Eliminando entorno de pruebas anterior..."; \
	docker compose down -v; \
	echo "==> Levantando PostgreSQL..."; \
	docker compose up -d db; \
	echo "==> Esperando a que PostgreSQL esté disponible..."; \
	ready=0; \
	for i in $$(seq 1 30); do \
		if docker compose exec -T db pg_isready -h localhost -U postgres -d tpe_pw_test >/dev/null 2>&1; then \
			ready=1; \
			break; \
		fi; \
		sleep 1; \
	done; \
	if [ "$$ready" -ne 1 ]; then \
		echo "ERROR: PostgreSQL no estuvo disponible a tiempo."; \
		docker compose logs db; \
		exit 1; \
	fi; \
	echo "==> Ejecutando tests..."; \
	TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5433/tpe_pw_test?sslmode=disable" \
		go test -p 1 -v ./...


test-api:
	@set -euo pipefail; \
	cleanup() { \
		echo "==> Limpiando contenedores y volúmenes..."; \
		docker compose down -v >/dev/null 2>&1 || true; \
	}; \
	trap cleanup EXIT; \
	echo "==> Generando código con sqlc..."; \
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate; \
	echo "==> Compilando proyecto..."; \
	go build ./...; \
	echo "==> Eliminando contenedores y volúmenes anteriores..."; \
	docker compose down -v; \
	echo "==> Levantando PostgreSQL..."; \
	docker compose up -d db; \
	echo "==> Esperando a que PostgreSQL esté disponible..."; \
	ready=0; \
	for i in $$(seq 1 30); do \
		if docker compose exec -T db pg_isready -h localhost -U postgres -d tpe_pw_test >/dev/null 2>&1; then \
			ready=1; \
			break; \
		fi; \
		sleep 1; \
	done; \
	if [ "$$ready" -ne 1 ]; then \
		echo "ERROR: PostgreSQL no estuvo disponible a tiempo."; \
		docker compose logs db; \
		exit 1; \
	fi; \
	echo "==> Levantando la aplicación..."; \
	docker compose up -d --build app; \
	echo "==> Esperando a que la API responda..."; \
	ready=0; \
	for i in $$(seq 1 30); do \
		if curl -sf -o /dev/null http://localhost:8080/; then \
			ready=1; \
			break; \
		fi; \
		sleep 1; \
	done; \
	if [ "$$ready" -ne 1 ]; then \
		echo "ERROR: la API no respondió a tiempo."; \
		docker compose logs app; \
		exit 1; \
	fi; \
	echo "==> Probando los endpoints (requests.sh)..."; \
	if ! bash requests.sh; then \
		echo "ERROR: fallaron pruebas de la API."; \
		docker compose logs app; \
		exit 1; \
	fi; \
	echo "==> Todas las pruebas de la API pasaron."