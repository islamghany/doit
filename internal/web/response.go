package web

import (
	"encoding/json"
	"net/http"
)

// Response sends a JSON response with the given status code and data.
func Response(w http.ResponseWriter, r *http.Request, statusCode int, data any) error {
	SetStatusCode(r.Context(), statusCode)

	// Set headers BEFORE WriteHeader
	w.Header().Set("Content-Type", "application/json")

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.WriteHeader(statusCode) // Now write status

	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}

func RespondOK(w http.ResponseWriter, r *http.Request, data any) error {
	return Response(w, r, http.StatusOK, data)
}

func RespondCreated(w http.ResponseWriter, r *http.Request, data any) error {
	return Response(w, r, http.StatusCreated, data)
}

func RespondNoContent(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func SetResponseHeader(w http.ResponseWriter, key, value string) {
	w.Header().Set(key, value)
}
