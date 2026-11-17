# Persistencia de datos — TP6

## Descripción general

La aplicación registra avistamientos de jabalíes utilizando PostgreSQL 18.

El acceso a datos se implementa con `database/sql`, el driver `lib/pq` y código Go tipado generado por sqlc.

HTMX modifica la manera de actualizar la interfaz. Los datos continúan almacenándose en PostgreSQL y las operaciones siguen ejecutándose en el servidor.

## Modelo de datos

El esquema se encuentra en `db/schema/schema.sql`.

La tabla principal es `avistamientos`:

| Campo | Tipo PostgreSQL | Restricciones o valor predeterminado |
|-------|-----------------|--------------------------------------|
| `id` | SERIAL | Clave primaria generada automáticamente. |
| `fecha_hora` | TIMESTAMP WITH TIME ZONE | NOT NULL. |
| `ubicacion` | TEXT | NOT NULL. |
| `descripcion` | TEXT | NOT NULL. |
| `imagen` | TEXT | Permite NULL. |
| `hubo_destrozos` | BOOLEAN | NOT NULL, DEFAULT FALSE. |
| `detalle_destrozos` | TEXT | Permite NULL. |
| `nombre_reportante` | VARCHAR(255) | Permite NULL. |
| `email_reportante` | VARCHAR(255) | Permite NULL. |
| `created_at` | TIMESTAMP WITH TIME ZONE | NOT NULL, DEFAULT CURRENT_TIMESTAMP. |

`fecha_hora` indica cuándo ocurrió el avistamiento. `created_at` indica cuándo se incorporó el registro a la base.

El campo `imagen` almacena una referencia de texto, no el contenido binario de un archivo.

## Campos opcionales y validaciones

Los campos `imagen`, `detalle_destrozos`, `nombre_reportante` y `email_reportante` permiten NULL.

En el formulario, la función `textoOpcional` elimina espacios al principio y al final. Si el resultado está vacío, devuelve un `sql.NullString` con `Valid: false`, que se almacena como NULL.

`hubo_destrozos` no es un texto opcional: se guarda como true si la casilla está marcada y como false si está desmarcada.

El handler del formulario comprueba:

- Que ubicación y descripción no estén vacías después de eliminar espacios.
- Que fecha y hora tengan el formato esperado.

Las restricciones NOT NULL de PostgreSQL impiden almacenar NULL en los campos obligatorios, pero no impiden por sí mismas guardar cadenas vacías. Por eso la validación del servidor es necesaria.

El navegador también aplica las validaciones HTML, como los campos required y el formato del email.

## Tratamiento de fecha y hora

El formulario utiliza `datetime-local`, que envía una fecha y hora sin zona horaria.

Actualmente el handler interpreta ese valor mediante:

```go
time.Parse("2006-01-02T15:04", valor)
```

Al no indicar una zona, Go lo interpreta en UTC. La implementación no realiza una conversión explícita desde la zona horaria del usuario.

La columna TIMESTAMP WITH TIME ZONE representa un instante; PostgreSQL no conserva el nombre de la zona horaria original.

## Consultas y generación con sqlc

Las consultas están definidas en `db/queries/avistamientos.sql`.

| Consulta | Operación |
|----------|-----------|
| `CreateAvistamiento` | Inserta y devuelve el registro creado. |
| `GetAvistamiento` | Busca por identificador. |
| `ListAvistamientos` | Lista los registros por fecha_hora descendente. |
| `UpdateAvistamiento` | Actualiza un registro por identificador. |
| `DeleteAvistamiento` | Elimina un registro por identificador. |

La configuración se encuentra en `sqlc.yaml`.

Para regenerar el código:

```bash
make generate
```

Este comando ejecuta sqlc y también genera las vistas de templ.

El código de acceso a datos se genera en `db/sqlc/`. Los archivos generados deben regenerarse a partir del SQL, no editarse manualmente.

Las consultas utilizan parámetros como `$1` y `$2`; los valores no se concatenan directamente al SQL.

## Integración con la interfaz del TP6

En `main.go` se abre la conexión a partir de `DATABASE_URL`, se comprueba con `Ping` y se construye `db.New(conn)`.

Las rutas registradas utilizan la persistencia de esta manera:

| Ruta | Acceso a datos | Respuesta exitosa |
|------|----------------|------------------|
| `GET /` | `ListAvistamientos` | Página HTML completa. |
| `POST /avistamientos` | `CreateAvistamiento` y luego `ListAvistamientos` | Fragmento HTML de la lista actualizada. |
| `DELETE /avistamientos/{id}` | `GetAvistamiento` y luego `DeleteAvistamiento` | Estado 200 con cuerpo vacío. |

El alta recibe campos de formulario. No espera JSON y ya no redirige a la página principal.

Después del alta, HTMX reemplaza el contenedor de la lista con el HTML devuelto por el servidor.

Después del borrado, HTMX elimina visualmente la fila al reemplazarla por la respuesta vacía.

Las consultas SQL de actualización y consulta individual siguen disponibles en la capa de persistencia y se utilizan en las pruebas. Su existencia no implica que haya rutas HTTP registradas para todas ellas.

## Borrado y registros inexistentes

`DeleteAvistamiento` está definida con `:exec`, por lo que su método generado no devuelve el número de filas afectadas.

El handler consulta primero el registro con `GetAvistamiento`:

- Identificador inválido: respuesta 400.
- Registro inexistente: respuesta 404.
- Error interno: respuesta 500.
- Borrado exitoso: respuesta 200 sin cuerpo.

La consulta previa y el borrado son operaciones separadas; no se ejecutan dentro de una transacción conjunta.

## Configuración de PostgreSQL en Docker

El servicio `db` utiliza la imagen `postgres:18`.

Configuración local:

| Parámetro | Valor |
|-----------|-------|
| Base | `tpe_pw_test` |
| Usuario | `postgres` |
| Contraseña | `postgres` |
| Dirección desde app | `db:5432` |
| Dirección desde la computadora | `localhost:5433` |

Estas credenciales corresponden al entorno académico local.

Compose proporciona a la aplicación:

```text
DATABASE_URL=postgres://postgres:postgres@db:5432/tpe_pw_test?sslmode=disable
```

No es necesario exportar esta variable manualmente para ejecutar `make run-app`.

## Inicialización del esquema

Compose monta:

```text
./db/schema/schema.sql
```

como:

```text
/docker-entrypoint-initdb.d/001-schema.sql
```

La imagen de PostgreSQL ejecuta ese archivo al inicializar un directorio de datos vacío.

El script no se vuelve a aplicar automáticamente sobre un volumen que ya contiene una base inicializada. Para esta entrega, `make run-app` elimina el volumen anterior y permite reconstruir la base desde el esquema.

El proyecto no incorpora un sistema de migraciones incrementales.

## Volumen y duración de los datos

Compose define el volumen `postgres_test_data` y lo monta en `/var/lib/postgresql`, conforme a la configuración utilizada para PostgreSQL 18.

Docker Compose normalmente agrega el nombre del proyecto como prefijo al nombre físico del volumen.

Mientras la aplicación está en ejecución, los datos están almacenados en PostgreSQL: recargar la página no los elimina.

El comportamiento entre ejecuciones depende del comando:

| Comando | Efecto sobre los datos |
|---------|-----------------------|
| `make run-app` | Elimina el volumen al iniciar y al finalizar. |
| `make run` | Reutiliza el volumen existente; no incluye limpieza automática al salir. |
| `make down` | Conserva el volumen. |
| `make clean` | Elimina el volumen. |
| `make test` | Elimina los datos anteriores y limpia al finalizar las pruebas. |

La eliminación de datos con `make run-app` es intencional y responde al requisito de la entrega final.

## Arranque y limpieza automatizados

`make run-app` ejecuta `run-app.sh`.

El script:

1. Comprueba los requisitos básicos.
2. Genera el código con sqlc y templ.
3. Compila la aplicación mediante la construcción de la imagen Docker.
4. Elimina el entorno anterior del proyecto.
5. Inicia PostgreSQL.
6. Espera su disponibilidad mediante `pg_isready -h localhost`.
7. Inicia la aplicación.
8. Elimina los contenedores y volúmenes al finalizar.

La comprobación con `-h localhost` utiliza TCP. Esto evita tomar como listo el servidor temporal de inicialización que utiliza sockets locales.

Compose también configura una comprobación de salud de PostgreSQL y la dependencia de la aplicación respecto de la base.

La limpieza se registra con `trap` para la salida del script. Las señales INT y TERM provocan una salida que ejecuta esa limpieza. No puede garantizarse ante una terminación forzada con SIGKILL o un apagado del equipo.

## Pruebas de persistencia

Las pruebas se encuentran en:

```text
db/sqlc/avistamientos_test.go
```

Se ejecutan con:

```bash
make test
```

El Makefile genera el código, compila, prepara PostgreSQL, espera su disponibilidad y ejecuta `go test`. Al finalizar elimina los contenedores y volúmenes.

Las pruebas verifican un ciclo CRUD: creación, consulta individual, actualización, listado y eliminación.

La aplicación y las pruebas utilizan los mismos servicios y volumen de Compose. Por eso `make test` debe ejecutarse con la aplicación detenida y teniendo en cuenta que elimina sus datos.

Estas pruebas no verifican el comportamiento de HTMX en el navegador. El alta y borrado sin recarga deben comprobarse también mediante la interfaz.

## Verificación de la limpieza

Después de detener `make run-app` con Ctrl+C y esperar su finalización:

```bash
docker compose ps -a
```

No deben quedar contenedores del proyecto.

Durante la limpieza, Docker debe informar la eliminación del volumen correspondiente. Una nueva ejecución de `make run-app` reconstruye el esquema y comienza sin los avistamientos cargados en la sesión anterior.
