# TPE Programación Web — TP6

Aplicación web para registrar, consultar y eliminar avistamientos de jabalíes.

La entrega se encuentra en la rama `tp6`. Evoluciona la aplicación del TP5 incorporando HTMX para crear y eliminar avistamientos sin recargar toda la página.

El servidor está desarrollado en Go. Las vistas HTML se generan con templ y los datos se almacenan en PostgreSQL mediante consultas tipadas generadas por sqlc. La aplicación y la base se ejecutan en contenedores Docker.

## Requisitos previos

- Git, para clonar el repositorio.
- Go compatible con la versión indicada en `go.mod`: 1.26.5 o superior.
- Docker instalado, en ejecución y accesible para el usuario.
- Docker Compose v2, disponible mediante `docker compose`.
- Make y Bash.
- Puertos 8080 y 5433 disponibles.
- Conexión a Internet para descargar dependencias e imágenes y cargar HTMX y Pico.css desde sus CDN.

No es necesario instalar PostgreSQL, sqlc ni templ de manera global.

El Makefile ejecuta versiones específicas de las herramientas:

- sqlc: v1.31.1.
- templ: v0.3.1100.

## Ejecución de la entrega

Clonar la rama de la entrega:

```bash
git clone --branch tp6 --single-branch https://github.com/TadeoBrun1/TPE_PW.git
cd TPE_PW
```

Ejecutar:

```bash
make run-app
```

También se puede ejecutar el mismo script directamente:

```bash
bash ./run-app.sh
```

El comando realiza automáticamente estas tareas:

1. Comprueba los requisitos básicos y el acceso a Docker.
2. Genera el código Go de sqlc y templ.
3. Construye la imagen Docker, compilando la aplicación.
4. Elimina los contenedores y volúmenes anteriores del proyecto.
5. Levanta PostgreSQL.
6. Espera a que PostgreSQL esté disponible.
7. Inicia la aplicación y muestra los registros de los servicios.
8. Al finalizar, elimina los contenedores, la red de Compose y los volúmenes del proyecto.

Abrir en el navegador:

http://localhost:8080

Para finalizar, presionar Ctrl+C una vez en la terminal y esperar a que termine la limpieza.

**Importante:** `make run-app` elimina los datos anteriores al iniciar y los datos de la sesión al finalizar. Este comportamiento cumple con el ciclo de ejecución solicitado para la entrega. No utilizarlo para conservar información entre ejecuciones.

El script registra la limpieza mediante `trap`, incluyendo la salida por errores y las señales INT y TERM. Un cierre forzado del proceso o del equipo puede impedirla; en ese caso se puede ejecutar `make clean`.

## Uso de la aplicación

La página principal contiene un formulario y una tabla de avistamientos.

### Registrar un avistamiento

Campos obligatorios:

- Fecha y hora del suceso.
- Ubicación.
- Descripción.

Campos opcionales de texto:

- Detalle de los destrozos.
- Referencia a una imagen.
- Nombre de quien reporta.
- Email de quien reporta.

La casilla “¿Hubo destrozos?” se guarda como un booleano: marcada equivale a verdadero y desmarcada a falso.

Los campos opcionales de texto vacíos se almacenan como NULL. El campo imagen almacena una referencia de texto; no implementa la subida de archivos.

Al enviar el formulario, HTMX realiza un POST a `/avistamientos`. El servidor valida los datos, guarda el registro, consulta nuevamente la lista y devuelve únicamente el HTML del componente de listado.

HTMX reemplaza el contenedor `#avistamientos-list` mediante `outerHTML`, sin recargar la página completa.

En esta versión, los campos del formulario conservan sus valores después de un alta exitosa.

### Eliminar un avistamiento

Cada fila incluye un botón “Borrar” con confirmación.

Al confirmar, HTMX envía un DELETE a `/avistamientos/{id}`. El servidor elimina el registro y responde con estado 200 y cuerpo vacío.

El botón utiliza `hx-target="closest tr"` y `hx-swap="outerHTML"`: la fila se reemplaza por la respuesta vacía y desaparece de la tabla.

Cancelar la confirmación no envía la petición ni elimina el registro.

## Rutas HTTP registradas

| Método | Ruta | Comportamiento |
|--------|------|----------------|
| GET | `/` | Devuelve la página completa con formulario y lista. |
| POST | `/avistamientos` | Recibe los campos del formulario, guarda el registro y devuelve la lista actualizada como HTML. |
| DELETE | `/avistamientos/{id}` | Elimina el registro y devuelve 200 con cuerpo vacío. |

El POST ya no utiliza una redirección 303.

Los datos obligatorios ausentes, la fecha inválida o un identificador inválido producen una respuesta 400. Un identificador válido que no existe produce 404 al intentar borrarlo. Los errores internos producen una respuesta 500.

Las validaciones HTML del navegador ayudan a completar el formulario. El servidor también valida los campos obligatorios y la fecha.

Esta versión no incorpora un componente visual específico para mostrar los errores de las peticiones HTMX. Se pueden inspeccionar en las herramientas de desarrollo del navegador.

En `handlers.go` se conservan funciones JSON de entregas anteriores, pero no constituyen rutas disponibles si no están registradas en `main.go`.

## Vistas y HTMX

| Archivo | Responsabilidad |
|---------|-----------------|
| `views/layout.templ` | Estructura HTML general y carga de HTMX y Pico.css. |
| `views/index_page.templ` | Compone el formulario y la lista dentro del layout. |
| `views/avistamiento_form.templ` | Formulario con hx-post, hx-target y hx-swap. |
| `views/avistamientos_list.templ` | Contenedor de la lista, tabla y botones de borrado. |

HTMX 1.9.10 se carga desde:

https://unpkg.com/htmx.org@1.9.10

Los estilos se cargan desde Pico.css mediante CDN.

La entrega no requiere archivos JavaScript `.js` locales ni código JavaScript propio. HTMX es una biblioteca JavaScript externa que interpreta los atributos `hx-*` del HTML.

Cada archivo `.templ` genera un archivo `_templ.go`. Estos archivos generados no deben editarse manualmente.

## Persistencia

PostgreSQL 18 almacena los avistamientos en la base `tpe_pw_test`.

- Esquema: `db/schema/schema.sql`.
- Consultas: `db/queries/avistamientos.sql`.
- Configuración de sqlc: `sqlc.yaml`.
- Código generado: `db/sqlc/`.
- Documentación detallada: [Persistencia de datos](docs/persistencia.md).

Compose aplica el esquema automáticamente al inicializar un volumen vacío.

La aplicación se conecta a `db:5432` dentro de la red de Docker. Desde la computadora, PostgreSQL está disponible en `localhost:5433`.

Las credenciales incluidas en Compose son para el entorno académico local.

## Comandos disponibles

| Comando | Descripción |
|---------|-------------|
| `make run-app` | Comando de entrega: genera, compila, reinicia el entorno, ejecuta y limpia al finalizar. |
| `make generate` | Genera el código Go de sqlc y templ. |
| `make build` | Genera el código y compila el proyecto localmente. |
| `make run` | Alias de `make run-app`: incluye limpieza previa y final de contenedores y volúmenes. |
| `make down` | Elimina los contenedores y la red de Compose, conservando el volumen de datos. |
| `make clean` | Elimina los contenedores, la red y los volúmenes. |
| `make test` | Genera, compila y ejecuta las pruebas de persistencia con PostgreSQL; elimina los datos al comenzar y terminar. |

No ejecutar `make test` al mismo tiempo que la aplicación: utiliza los mismos recursos de Compose y elimina sus datos.

## Pruebas automatizadas

Ejecutar con la aplicación detenida:

```bash
make test
```

El comando genera el código, compila, prepara PostgreSQL, ejecuta las
pruebas y elimina los contenedores y volúmenes al finalizar.

Se ejecutan dos grupos de pruebas:

- `db/sqlc/avistamientos_test.go`: verifica creación, consulta individual,
  actualización, listado y eliminación en PostgreSQL.
- `handlers_test.go`: verifica el recorrido HTML mediante `httptest`,
  utilizando los handlers y una base PostgreSQL real.

El recorrido HTML comprueba:

1. Página inicial con formulario y mensaje de lista vacía.
2. Rechazo de un formulario inválido sin guardar datos.
3. Alta correcta con respuesta HTML de la lista y sin redirección.
4. Presencia del registro al consultar nuevamente la página.
5. Borrado con estado 200 y cuerpo vacío.
6. Ausencia del registro después del borrado.
7. Respuesta 404 al intentar borrar un registro inexistente.
8. Respuesta 400 ante un identificador inválido.

La prueba HTML utiliza una transacción que se revierte al finalizar.
El Makefile ejecuta los paquetes secuencialmente mediante `-p 1`.

Estas pruebas no ejecutan JavaScript ni CSS en un navegador. Los reemplazos
de HTMX y la visibilidad del mensaje de lista vacía se verifican manualmente.

## Verificación manual de la entrega

1. Ejecutar `make run-app`.
2. Abrir http://localhost:8080.
3. Crear un avistamiento y comprobar que aparece sin recargar.
4. Recargar manualmente y comprobar que continúa registrado.
5. Cancelar una confirmación de borrado y comprobar que el registro permanece.
6. Confirmar el borrado y comprobar que la fila desaparece sin recargar.
7. Recargar y comprobar que el registro eliminado no reaparece.
8. Finalizar con Ctrl+C y esperar la limpieza.
9. Ejecutar `docker compose ps -a`: no deben quedar contenedores del proyecto.
10. Volver a ejecutar `make run-app`: debe comenzar con una base nueva.

Para comprobar que no hay archivos `.js` versionados:

```bash
git ls-files '*.js'
```

La salida debe estar vacía.

## Organización del proyecto

| Ubicación | Contenido |
|-----------|-----------|
| `main.go` | Conexión a PostgreSQL, registro de rutas e inicio del servidor. |
| `handlers.go` | Procesamiento HTTP, validaciones y renderizado. |
| `views/` | Componentes templ y código Go generado. |
| `db/schema/` | Esquema SQL. |
| `db/queries/` | Consultas utilizadas por sqlc. |
| `db/sqlc/` | Código generado y pruebas de persistencia. |
| `docs/persistencia.md` | Modelo de datos y ciclo de vida de la base. |
| `run-app.sh` | Automatización del ciclo completo de ejecución. |
| `Makefile` | Comandos de generación, compilación, ejecución y pruebas. |
| `compose.yaml` | Servicios app y db, puertos, volumen y comprobación de salud. |
| `Dockerfile` | Compilación de Go e imagen de ejecución. |

## Problemas frecuentes

- **Docker no está disponible:** comprobar que el servicio esté iniciado y que el usuario tenga permisos.
- **Puerto ocupado:** liberar los puertos 8080 y 5433 antes de ejecutar.
- **htmx:targetError:** comprobar que el formulario apunte a `#avistamientos-list` y que ese elemento exista.
- **Los cambios de las vistas no aparecen:** detener la aplicación, volver a ejecutar el comando y recargar el navegador.
- **HTMX no carga:** comprobar la conexión a Internet y los errores de red del navegador.
- **make informa una interrupción al presionar Ctrl+C:** puede deberse al código de salida de la señal; comprobar que haya terminado la limpieza.

## Lista sin avistamientos

Cuando no hay registros se muestra “No hay avistamientos registrados”.

El mensaje se encuentra en una fila especial de la tabla. CSS la muestra
únicamente cuando es la única fila del cuerpo de la tabla mediante
`:only-child`. Al crear un avistamiento queda oculta y al borrar el último
vuelve a mostrarse sin recargar la página y sin JavaScript propio.
