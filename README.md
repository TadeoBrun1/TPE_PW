# TPE Programación Web

Aplicación web para el registro y consulta de avistamientos de jabalíes.

En esta entrega se expone la capa de persistencia (PostgreSQL + `sqlc`) mediante una **API REST con JSON** para la entidad `Avistamiento`. La aplicación, la base de datos y las pruebas se ejecutan con Docker.

## Requisitos

Para ejecutar los tests es necesario disponer de:

* Go
* Docker
* Docker Compose
* Make
* `curl` y `bash`

No es necesario crear manualmente la base de datos ni instalar `sqlc` globalmente. El proceso de testing se encarga de preparar automáticamente el entorno necesario.

Los puertos **8080** (API) y **5433** (PostgreSQL) deben estar libres.

## Persistencia

La entidad principal persistida es `Avistamiento`.

La definición del esquema se encuentra en:

```text
db/schema/schema.sql
```

Las consultas SQL utilizadas por `sqlc` se encuentran en:

```text
db/queries/avistamientos.sql
```

La configuración de `sqlc` se encuentra en:

```text
sqlc.yaml
```

El código Go generado se almacena en:

```text
db/sqlc/
```

La documentación detallada sobre las decisiones de persistencia se encuentra en:

```text
docs/persistencia.md
```

## Ejecución de los tests

Desde la raíz del proyecto ejecutar:

```bash
make test
```

Este comando realiza automáticamente las siguientes tareas:

1. Genera el código Go utilizando `sqlc`.
2. Compila el proyecto.
3. Elimina contenedores y volúmenes de pruebas anteriores.
4. Levanta una instancia PostgreSQL mediante Docker Compose.
5. Espera hasta que PostgreSQL esté disponible.
6. Inicializa el esquema de la base de datos.
7. Ejecuta los tests con el paquete `testing` de Go.
8. Elimina los contenedores y volúmenes utilizados al finalizar.

La limpieza del entorno también se realiza si ocurre un error durante la ejecución.

## Tests de persistencia

Los tests realizan un ciclo CRUD completo sobre la entidad `Avistamiento`:

* creación;
* consulta por identificador;
* actualización;
* listado;
* eliminación.

La base de datos utilizada para los tests es temporal y se ejecuta dentro de Docker, evitando depender de una base PostgreSQL previamente configurada en la computadora.

## API REST

La API expone la entidad `Avistamiento` con los siguientes endpoints:

| Método   | Ruta                  | Descripción                            | Respuestas    |
|----------|-----------------------|----------------------------------------|---------------|
| `GET`    | `/`                   | Página de inicio (`index.html`)        | 200           |
| `POST`   | `/avistamientos`      | Crea un avistamiento                   | 201, 400      |
| `GET`    | `/avistamientos`      | Lista todos (el más reciente primero)  | 200           |
| `GET`    | `/avistamientos/{id}` | Obtiene un avistamiento por ID         | 200, 400, 404 |
| `PUT`    | `/avistamientos/{id}` | Reemplaza los datos de un avistamiento | 200, 400, 404 |
| `DELETE` | `/avistamientos/{id}` | Elimina un avistamiento                | 204, 400, 404 |

Los errores se devuelven como JSON con la forma `{"error":"mensaje"}`.

Ejemplo del JSON de un avistamiento:

```json
{
  "fecha_hora": "2026-10-10T18:30:00Z",
  "ubicacion": "Ruta 3 km 280, Tandil",
  "descripcion": "Jabalí cruzando la ruta",
  "imagen": null,
  "hubo_destrozos": true,
  "detalle_destrozos": "Alambrado roto",
  "nombre_reportante": "Tomas",
  "email_reportante": "tomas@test.com"
}
```
* Obligatorios: `fecha_hora`, `ubicacion` y `descripcion` (no pueden estar vacios).
* Opcionales: `imagen`,`detalle_destrozos`,`nombre_reportante`y`email_reportante`.Pueden omitirse o enviarse como `null`.
* Las respuestas incluyen además `id` y `created_at`, que asigna el servidor.
* `PUT` reemplaza el registro completo: un campo opcional que no se envia queda en `null`.

## Ejecutar la API

Para dejar la API corriendo y usarla manualmente:

```bash
docker compose up --build
```

Se levantan dos contenedores:

* `db`: PostgreSQL con el esquema de `db/schema/schema.sql` ya aplicado. También queda accesible desde la computadora en `localhost:5433`.
* `app`: la API, disponible en <http://localhost:8080>. Arranca cuando la base está lista.

Para detenerla, `Ctrl+C` y luego:

```bash
docker compose down        # conserva los datos
docker compose down -v     # además elimina los datos
```

La aplicación toma la dirección de la base de la variable de entorno `DATABASE_URL`, que ya está definida en `compose.yaml`. Para ejecutarla sin contenedor para la aplicación:

```bash
docker compose up -d db
export DATABASE_URL="postgres://postgres:postgres@localhost:5433/tpe_pw_test?sslmode=disable"
go run .
```

## Prueba de la API

Desde la raíz del proyecto ejecutar:

```bash
make test-api
```

Este comando realiza automáticamente las siguientes tareas:

1. Genera el código Go utilizando `sqlc`.
2. Compila el proyecto.
3. Elimina contenedores y volúmenes de ejecuciones anteriores.
4. Levanta PostgreSQL mediante Docker Compose.
5. Espera hasta que PostgreSQL esté disponible.
6. Construye y levanta la aplicación en un contenedor.
7. Espera hasta que la API responda.
8. Prueba cada uno de los endpoints con `requests.sh`.
9. Elimina los contenedores y volúmenes utilizados al finalizar.

La limpieza del entorno también se realiza si ocurre un error durante la ejecución. Si alguna prueba falla, el comando termina con error y muestra los logs de la aplicación.

Con la API ya corriendo, las pruebas también se pueden ejecutar por separado:

```bash
bash requests.sh
```

El script ejecuta con `curl` todas las operaciones (crear, listar, obtener, actualizar y eliminar, más los casos de error), verifica el código de estado de cada una e imprime `[OK]` o `[ERROR]`. Termina con código distinto de cero si alguna falla. Para apuntar a otra dirección: `BASE_URL=http://otro-host:8080 bash requests.sh`.

Ejemplo de una petición individual:

```bash
curl -i -X POST http://localhost:8080/avistamientos \
  -H "Content-Type: application/json" \
  -d '{"fecha_hora":"2026-10-10T18:30:00Z","ubicacion":"Tandil","descripcion":"Jabalí en el camino"}'
```

> `make test` y `make test-api` ejecutan `docker compose down -v`, por lo que eliminan los datos de la base. Conviene usarlos con la API detenida.


## Estructura principal

```text
.
├── db/
│   ├── queries/
│   │   └── avistamientos.sql
│   ├── schema/
│   │   └── schema.sql
│   └── sqlc/
│       ├── avistamientos.sql.go
│       ├── avistamientos_test.go
│       ├── db.go
│       └── models.go
├── docs/
│   └── persistencia.md
├── .dockerignore
├── compose.yaml
├── Dockerfile
├── Makefile
├── sqlc.yaml
├── go.mod
├── go.sum
├── requests.sh
├── handlers.go
├── main.go
└── README.md
```

## Generación manual con sqlc

Normalmente no es necesario ejecutar este paso manualmente porque `make test` ya lo realiza.

En caso de querer regenerar únicamente el código de acceso a datos se puede ejecutar:

```bash
go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate
```
