package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	db "tp1/db/sqlc"
)

func TestRecorridoHTML(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Fatal("Falta TEST_DATABASE_URL. Ejecutá las pruebas con make test.")
	}

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		t.Fatal(err)
	}

	// Los cambios de esta prueba se revierten al terminar.
	tx, err := conn.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(t.Context(), "DELETE FROM avistamientos"); err != nil {
		t.Fatal(err)
	}

	server := &Server{queries: db.New(tx)}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", server.indexHandler)
	mux.HandleFunc("POST /avistamientos", server.crearvistamientoForm)
	mux.HandleFunc("DELETE /avistamientos/{id}", server.deleteAvistamiento)

	pedir := func(metodo, ruta string, datos url.Values) *httptest.ResponseRecorder {
		t.Helper()

		req := httptest.NewRequest(metodo, ruta, strings.NewReader(datos.Encode()))
		if metodo == http.MethodPost {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		if metodo != http.MethodGet {
			req.Header.Set("HX-Request", "true")
		}

		res := httptest.NewRecorder()
		mux.ServeHTTP(res, req)
		return res
	}

	comprobarEstado := func(res *httptest.ResponseRecorder, esperado int) {
		t.Helper()
		if res.Code != esperado {
			t.Fatalf("Estado esperado %d, obtenido %d. Respuesta: %s",
				esperado, res.Code, res.Body.String())
		}
	}

	comprobarContenido := func(res *httptest.ResponseRecorder, texto string) {
		t.Helper()
		if !strings.Contains(res.Body.String(), texto) {
			t.Fatalf("La respuesta no contiene %q. Respuesta: %s",
				texto, res.Body.String())
		}
	}

	// 1. Página inicial con formulario y lista vacía.
	res := pedir(http.MethodGet, "/", nil)
	comprobarEstado(res, http.StatusOK)
	comprobarContenido(res, "<html")
	comprobarContenido(res, `hx-post="/avistamientos"`)
	comprobarContenido(res, `hx-target="#avistamientos-list"`)
	comprobarContenido(res, `id="avistamientos-list"`)
	comprobarContenido(res, "No hay avistamientos registrados.")

	if strings.Contains(res.Body.String(), "hx-delete=") {
		t.Fatal("La página vacía no debería incluir botones de borrado")
	}

	// 2. Un formulario inválido no debe guardar registros.
	res = pedir(http.MethodPost, "/avistamientos", url.Values{
		"fecha_hora":  {"2026-11-21T10:30"},
		"ubicacion":   {"   "},
		"descripcion": {"Prueba inválida"},
	})
	comprobarEstado(res, http.StatusBadRequest)

	items, err := server.queries.ListAvistamientos(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatal("El formulario inválido guardó un registro")
	}

	// 3. Crear mediante el formulario.
	ubicacion := "Tandil - prueba del recorrido HTML"
	res = pedir(http.MethodPost, "/avistamientos", url.Values{
		"fecha_hora":  {"2026-11-21T10:30"},
		"ubicacion":   {ubicacion},
		"descripcion": {"Avistamiento de prueba HTML"},
	})
	comprobarEstado(res, http.StatusOK)
	comprobarContenido(res, ubicacion)
	comprobarContenido(res, `id="avistamientos-list"`)

	if res.Header().Get("Location") != "" {
		t.Fatal("El alta no debe redirigir")
	}
	if !strings.HasPrefix(res.Header().Get("Content-Type"), "text/html") {
		t.Fatal("El alta debe responder HTML")
	}
	if strings.Contains(res.Body.String(), "<html") ||
		strings.Contains(res.Body.String(), "<form") {
		t.Fatal("El alta debe devolver solo la lista, no la página completa")
	}

	items, err = server.queries.ListAvistamientos(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Ubicacion != ubicacion {
		t.Fatalf("No se guardó correctamente el avistamiento: %+v", items)
	}

	rutaBorrado := fmt.Sprintf("/avistamientos/%d", items[0].ID)
	comprobarContenido(res, `hx-delete="`+rutaBorrado+`"`)

	// 4. Una nueva consulta de la página debe incluir el registro.
	res = pedir(http.MethodGet, "/", nil)
	comprobarEstado(res, http.StatusOK)
	comprobarContenido(res, ubicacion)

	// 5. Borrar y comprobar la respuesta que necesita HTMX.
	res = pedir(http.MethodDelete, rutaBorrado, nil)
	comprobarEstado(res, http.StatusOK)
	if res.Body.Len() != 0 {
		t.Fatalf("El borrado debe responder sin cuerpo: %s", res.Body.String())
	}

	items, err = server.queries.ListAvistamientos(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatal("El avistamiento continúa en la base después del borrado")
	}

	// 6. La página vuelve a estar vacía.
	res = pedir(http.MethodGet, "/", nil)
	comprobarEstado(res, http.StatusOK)
	comprobarContenido(res, "No hay avistamientos registrados.")

	if strings.Contains(res.Body.String(), ubicacion) ||
		strings.Contains(res.Body.String(), "hx-delete=") {
		t.Fatal("El registro borrado sigue apareciendo en la página")
	}

	// 7. Intentar borrar nuevamente debe responder 404.
	res = pedir(http.MethodDelete, rutaBorrado, nil)
	comprobarEstado(res, http.StatusNotFound)

	// 8. Un identificador inválido debe responder 400.
	res = pedir(http.MethodDelete, "/avistamientos/abc", nil)
	comprobarEstado(res, http.StatusBadRequest)
}
