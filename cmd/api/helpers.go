package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// get "id" from request context
func (app *application) readIDParam(r *http.Request) (int64, error) {
  params := httprouter.ParamsFromContext(r.Context())

  id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
  if err != nil || id < 1 {
    return 0, errors.New("invalid id parameter")
  }

  return id, nil
}

// write json to responses
func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
  j, err := json.Marshal(data)
  if err != nil {
    return err
  }

  // terminal output niceness
  j = append(j, '\n')

  // write headers
  for key, value := range headers {
    w.Header()[key] = value
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)
  w.Write(j)

  return nil
}
