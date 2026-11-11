# TPE Programación Web

Aplicación web para el registro y consulta de avistamientos de jabalíes.

En esta entrega la interfaz se renderiza del lado del servidor con [`templ`](https://templ.guide/): el servidor Go consulta la base de datos (PostgreSQL + `sqlc`) y responde con el HTML ya armado. No se usa JavaScript en el cliente. La aplicación y la base de datos se ejecutan con docker.

## Requisitos

Para ejecutar los tests es necesario disponer de:

* Go
* Docker
* Docker Compose
* Make


No es necesario instalar `sqlc` ni `templ` globalmente: el `Makefile` los ejecuta con `go run` en una versión fija. Tampoco hace falta crear manualmente la base de datos.

Los puertos **8080** (aplicacion) y **5433** (PostgreSQL) deben estar libres.

## Ejecucion

Desde la raiz del proyecto:

```
make run
```

Este comando genera el codigo de `sqlc` y de `templ`,construye la imagen y levanta dos contenedores:
- `db`:PostgreSQL con el esquema de `db/schema/shcema.sql` ya aplicado. Tambien queda accesible desde la computadora en `localhost:5433`.
- `app`:la aplicacion, disponible en [http://localhost:8080](http://localhost:8080). Arranca cuando la base esta lista.

Para detenerla:

```
make down     # conserva los datos
make clean    # además elimina los datos
```
# Comandos disponibles
 
| Comando         | Descripción                                                              |
|-----------------|--------------------------------------------------------------------------|
| `make generate` | Genera el código Go de las consultas (`sqlc`) y de las vistas (`templ`)  |
| `make build`    | Genera el código y compila el proyecto                                   |
| `make run`      | Genera el código y levanta la base de datos y la aplicación con Docker   |
| `make test`     | Ejecuta los tests de persistencia contra una base temporal en Docker     |
| `make down`     | Detiene los contenedores y conserva los datos                            |
| `make clean`    | Detiene los contenedores y elimina los datos                             |
 
> `make test` ejecuta `docker compose down -v`, por lo que elimina los datos de la base. Conviene usarlo con la aplicación detenida.

## Rutas

| Método | Ruta             | Descripción                                                                  |
|--------|------------------|------------------------------------------------------------------------------|
| `GET`  | `/`              | Página principal: formulario y lista de avistamientos                        |
| `POST` | `/avistamientos` | Recibe el formulario, guarda el avistamiento y redirige a `/` (303)  

Al enviar el formulario se aplica el patron Post/Redirect/Get: tras guardar, el servidor redirige a `/`, que muestra la lista actualizada.

Campos del formulario:

- Obligatorios: fecha y hora, ubicación y descripción.
- Opcionales: si hubo destrozos, detalle de los destrozos, imagen, nombre y email de quien reporta. Los campos opcionales vacíos se guardan como `NULL`.
Si faltan datos obligatorios o la fecha es inválida, el servidor responde con un error 400 y no guarda nada.

## Vistas

Los componentes `templ` se encuentran en `views/`:

```
views/layout.templ               # estructura base de la página (HTML, HEAD, BODY)
views/avistamiento_form.templ    # formulario para crear un avistamiento
views/avistamientos_list.templ   # tabla con los avistamientos
views/index_page.templ           # página principal: arma layout, formulario y lista
```

Cada `.templ` se compila a un archivo `_templ.go` con `make generate`. Los estilos provienen de [Pico.css](https://picocss.com/), cargado desde un CDN.


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

```
make test
```

Este comando realiza automáticamente las siguientes tareas:

1. Genera el código Go utilizando `sqlc` y `templ`.
2. Compila el proyecto.
3. Elimina contenedores y volúmenes de pruebas anteriores.
4. Levanta una instancia PostgreSQL mediante Docker Compose.
5. Espera hasta que PostgreSQL esté disponible.
6. Inicializa el esquema de la base de datos.
7. Ejecuta los tests con el paquete `testing` de Go.
8. Elimina los contenedores y volúmenes utilizados al finalizar.

Los tests realizan un ciclo CRUD completo sobre la entidad `Avistamiento`: creación, consulta por identificador, actualización, listado y eliminación. La base utilizada es temporal y se ejecuta dentro de Docker.

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
├── views/
│   ├── layout.templ
│   ├── avistamiento_form.templ
│   ├── avistamientos_list.templ
│   └── index_page.templ
├── docs/
│   └── persistencia.md
├── .dockerignore
├── compose.yaml
├── Dockerfile
├── Makefile
├── sqlc.yaml
├── go.mod
├── go.sum
├── handlers.go
├── main.go
└── README.md
```
Junto a cada `.templ` hay un `_templ.go` generado por `make generate`.

