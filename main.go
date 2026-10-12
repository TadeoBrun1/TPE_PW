package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	db "tp1/db/sqlc"

	_ "github.com/lib/pq"
)

type Server struct {
	queries *db.Queries
}

func main() {

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("Falta la variable de entorno DATABASE_URL")
	}
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		log.Fatal("No se pudo conectar a la base de datos", err)
	}
	log.Println("Conectando a PostgreSQL")

	server := &Server{queries: db.New(conn)}

	http.HandleFunc("POST /avistamientos", server.createAvistamiento)
	http.HandleFunc("GET /avistamientos", server.listAvistamientos)
	http.HandleFunc("GET /avistamientos/{id}", server.getAvistamiento)
	http.HandleFunc("PUT /avistamientos/{id}", server.updateAvistamiento)
	http.HandleFunc("DELETE /avistamientos/{id}", server.deleteAvistamiento)

	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "./index.html")

	}))

	port := ":8080"
	fmt.Printf("Servidor ESTÁTICO escuchando en http://localhost%s\n", port)

	err = http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Printf("Error al iniciar el servidor: %s\n", err)
	}
}
