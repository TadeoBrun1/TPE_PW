#!/usr/bin/env bash

BASE_URL="${BASE_URL:-http://localhost:8080}"

# Archivo temporal donde se guarda el cuerpo de cada respuesta.
BODY_FILE="$(mktemp)"
trap 'rm -f "$BODY_FILE"' EXIT

passed=0
failed=0

# request <descripción> <estado esperado> <método> <ruta> [json]
# Deja el código recibido en $status y el cuerpo en $BODY_FILE.
request() {
    local description="$1" expected="$2" method="$3" path="$4" data="${5-}"
    local args=(-s -o "$BODY_FILE" -w '%{http_code}' -X "$method" "$BASE_URL$path")

    if [[ -n "$data" ]]; then
        args+=(-H 'Content-Type: application/json' -d "$data")
    fi

    status="$(curl "${args[@]}")"

    if [[ "$status" == "$expected" ]]; then
        echo "[OK]    $method $path -> $status  ($description)"
        passed=$((passed + 1))
    else
        echo "[ERROR] $method $path -> $status, esperado $expected  ($description)"
        failed=$((failed + 1))
    fi
    if [[ -s "$BODY_FILE" ]]; then
        sed 's/^/        /' "$BODY_FILE"
    fi
}

echo "== Crear =="
request "avistamiento completo" 201 POST /avistamientos '{
  "fecha_hora": "2026-10-10T18:30:00Z",
  "ubicacion": "Ruta 3 km 280, Tandil",
  "descripcion": "Jabalí cruzando la ruta",
  "hubo_destrozos": true,
  "detalle_destrozos": "Alambrado roto",
  "nombre_reportante": "Tomas",
  "email_reportante": "tomas@test.com"
}'
# El id viene en el JSON de la respuesta: {"id":7,...}
id="$(grep -o '"id":[0-9]*' "$BODY_FILE" | head -1 | cut -d: -f2)"

request "solo campos obligatorios" 201 POST /avistamientos '{
  "fecha_hora": "2026-10-11T09:00:00Z",
  "ubicacion": "Cerro El Centinela",
  "descripcion": "Dos jabalíes cerca del sendero"
}'
id_secundario="$(grep -o '"id":[0-9]*' "$BODY_FILE" | head -1 | cut -d: -f2)"

echo
echo "== Crear con datos inválidos =="
request "ubicación vacía" 400 POST /avistamientos '{"fecha_hora":"2026-10-10T18:30:00Z","ubicacion":"","descripcion":"x"}'
request "descripción vacía" 400 POST /avistamientos '{"fecha_hora":"2026-10-10T18:30:00Z","ubicacion":"Tandil","descripcion":""}'
request "sin fecha_hora" 400 POST /avistamientos '{"ubicacion":"Tandil","descripcion":"x"}'
request "JSON mal formado" 400 POST /avistamientos '{"ubicacion":'

echo
echo "== Listar =="
request "todos los avistamientos" 200 GET /avistamientos

echo
echo "== Obtener por ID =="
request "ID existente" 200 GET "/avistamientos/$id"
request "ID inexistente" 404 GET /avistamientos/999999
request "ID no numérico" 400 GET /avistamientos/abc

echo
echo "== Actualizar =="
request "datos nuevos" 200 PUT "/avistamientos/$id" '{
  "fecha_hora": "2026-10-10T18:30:00Z",
  "ubicacion": "Ruta 3 km 281, Tandil",
  "descripcion": "Jabalí en el alambrado",
  "hubo_destrozos": true,
  "detalle_destrozos": "Alambrado roto y cultivo pisado",
  "nombre_reportante": "Tomas",
  "email_reportante": "tomas@test.com"
}'
request "datos inválidos" 400 PUT "/avistamientos/$id" '{"fecha_hora":"2026-10-10T18:30:00Z","ubicacion":"","descripcion":"x"}'
request "ID inexistente" 404 PUT /avistamientos/999999 '{"fecha_hora":"2026-10-10T18:30:00Z","ubicacion":"Tandil","descripcion":"x"}'
request "comprobar el cambio" 200 GET "/avistamientos/$id"

echo
echo "== Eliminar =="
request "ID existente" 204 DELETE "/avistamientos/$id"
request "ya no existe" 404 GET "/avistamientos/$id"
request "borrar de nuevo" 404 DELETE "/avistamientos/$id"
request "limpiar el segundo registro" 204 DELETE "/avistamientos/$id_secundario"

echo
echo "Resultado: $passed correctas, $failed con error"
[[ "$failed" -eq 0 ]]