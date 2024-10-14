package resp

import (
	"encoding/json"
	"errors"
	"net/http"
	"passkey-golang-backend/logger"
)

func JSONResponse(w http.ResponseWriter, status int, data any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func ErrorResponse(w http.ResponseWriter, r *http.Request, status int, err error) {
	var msg string
	switch status / 100 {
	case 4:
		msg = err.Error()
	case 5:
		msg = http.StatusText(status)
		logger.Error().Err(err).Send()
	}
	JSONResponse(w, status, map[string]any{
		"message": msg,
		"path":    r.URL.Path,
		"method":  r.Method,
		"status":  status,
	})
}

func ErrorResponseWithText(w http.ResponseWriter, r *http.Request, status int) {
	ErrorResponse(w, r, status, errors.New(http.StatusText(status)))
}
