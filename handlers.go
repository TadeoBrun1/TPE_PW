package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	db "tp1/db/sqlc"
)

type avistamientoRequest struct {
	FechaHora        time.Time `json:"fecha_hora"`
	Ubicacion        string    `json:"ubicacion"`
	Descripcion      string    `json:"descripcion"`
	Imagen           *string   `json:"imagen"`
	HuboDestrozos    bool      `json:"hubo_destrozos"`
	DetalleDestrozos *string   `json:"detalle_destrozos"`
	NombreReportante *string   `json:"nombre_reportante"`
	EmailReportante  *string   `json:"email_reportante"`
}
type avistamientoResponse struct {
	ID               int32     `json:"id"`
	FechaHora        time.Time `json:"fecha_hora"`
	Ubicacion        string    `json:"ubicacion"`
	Descripcion      string    `json:"descripcion"`
	Imagen           *string   `json:"imagen"`
	HuboDestrozos    bool      `json:"hubo_destrozos"`
	DetalleDestrozos *string   `json:"detalle_destrozos"`
	NombreReportante *string   `json:"nombre_reportante"`
	EmailReportante  *string   `json:"email_reportante"`
	CreatedAt        time.Time `json:"created_at"`
}

// pasa de *string (json) a sql.NullString(sqlc)
func toNullString(text *string) sql.NullString {
	if text == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *text, Valid: true}
}

// paso inverso de tonullstring, pasa de null en la base a null en json
func fromNullString(nullstring sql.NullString) *string {
	if !nullstring.Valid {
		return nil
	}
	return &nullstring.String
}

func toResponse(avistamiento db.Avistamiento) avistamientoResponse {
	return avistamientoResponse{
		ID:               avistamiento.ID,
		FechaHora:        avistamiento.FechaHora,
		Ubicacion:        avistamiento.Ubicacion,
		Descripcion:      avistamiento.Descripcion,
		Imagen:           fromNullString(avistamiento.Imagen),
		HuboDestrozos:    avistamiento.HuboDestrozos,
		DetalleDestrozos: fromNullString(avistamiento.DetalleDestrozos),
		NombreReportante: fromNullString(avistamiento.NombreReportante),
		EmailReportante:  fromNullString(avistamiento.EmailReportante),
		CreatedAt:        avistamiento.CreatedAt,
	}
}
func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(value)
}

func parseID(r *http.Request) (int32, error) {
	parseID, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
	if err != nil || parseID < 1 {
		return 0, errors.New("ID invalido")
	}
	return int32(parseID), nil
}

func (s *Server) findAvistamiento(w http.ResponseWriter, r *http.Request, id int32) (db.Avistamiento, bool) {
	avistamiento, err := s.queries.GetAvistamiento(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "avistamiento no encontrado"})
		return avistamiento, false
	}
	if err != nil {
		log.Println("error al buscar avistamiento: ", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "error interno"})
		return avistamiento, false
	}
	return avistamiento, true
}

func (s *Server) getAvistamiento(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	avistamiento, encontrado := s.findAvistamiento(w, r, id)
	if !encontrado {
		return
	}
	writeJSON(w, http.StatusOK, toResponse(avistamiento))
}

func (s *Server) updateAvistamiento(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	var request avistamientoRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido"})
		return
	}
	if err := validateAvistamiento(request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if _, encontro := s.findAvistamiento(w, r, id); !encontro {
		return
	}
	err = s.queries.UpdateAvistamiento(r.Context(), db.UpdateAvistamientoParams{
		ID:               id,
		FechaHora:        request.FechaHora,
		Ubicacion:        request.Ubicacion,
		Descripcion:      request.Descripcion,
		Imagen:           toNullString(request.Imagen),
		HuboDestrozos:    request.HuboDestrozos,
		DetalleDestrozos: toNullString(request.DetalleDestrozos),
		NombreReportante: toNullString(request.NombreReportante),
		EmailReportante:  toNullString(request.EmailReportante),
	})
	if err != nil {
		log.Println("error al actualizar avistamiento:", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "error interno"})
		return
	}

	actualizado, encontro := s.findAvistamiento(w, r, id)
	if !encontro {
		return
	}
	writeJSON(w, http.StatusOK, toResponse(actualizado))
}

func (s *Server) deleteAvistamiento(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if _, encontrado := s.findAvistamiento(w, r, id); !encontrado {
		return
	}
	if err := s.queries.DeleteAvistamiento(r.Context(), id); err != nil {
		log.Println("error al borrar avistamiento: ", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "error interno"})
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

func (s *Server) listAvistamientos(w http.ResponseWriter, r *http.Request) {
	items, err := s.queries.ListAvistamientos(r.Context())
	if err != nil {
		log.Println("error al listar avistamientos: ", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "error interno"})
		return
	}
	response := make([]avistamientoResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toResponse(item))
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) createAvistamiento(w http.ResponseWriter, r *http.Request) {
	var request avistamientoRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalido"})
		return
	}

	if err := validateAvistamiento(request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	created, err := s.queries.CreateAvistamiento(r.Context(), db.CreateAvistamientoParams{
		FechaHora:        request.FechaHora,
		Ubicacion:        request.Ubicacion,
		Descripcion:      request.Descripcion,
		Imagen:           toNullString(request.Imagen),
		HuboDestrozos:    request.HuboDestrozos,
		DetalleDestrozos: toNullString(request.DetalleDestrozos),
		NombreReportante: toNullString(request.NombreReportante),
		EmailReportante:  toNullString(request.EmailReportante),
	})

	if err != nil {
		log.Println("error al crear el avistamiento")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "error interno"})
		return
	}
	writeJSON(w, http.StatusCreated, toResponse(created))

}

func validateAvistamiento(request avistamientoRequest) error {
	if request.FechaHora.IsZero() {
		return errors.New("fecha_hora es obligatorio")
	}
	if strings.TrimSpace(request.Ubicacion) == "" {
		return errors.New("La ubicacion no puede estar vacia")
	}
	if strings.TrimSpace(request.Descripcion) == "" {
		return errors.New("La descripcion del avistamiento no puede estar vacia")
	}
	return nil
}
