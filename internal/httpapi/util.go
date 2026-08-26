package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func readJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func pathID(r *http.Request) (int64, error) {
	s := r.PathValue("id")
	if s == "" {
		return 0, errors.New("missing id")
	}
	return strconv.ParseInt(s, 10, 64)
}

func queryBatchID(r *http.Request) (int64, error) {
	s := r.URL.Query().Get("batch_id")
	if s == "" {
		return 0, errors.New("missing batch_id")
	}
	return strconv.ParseInt(s, 10, 64)
}

func queryInt64(r *http.Request, key string) (int64, error) {
	s := r.URL.Query().Get(key)
	if s == "" {
		return 0, errors.New("missing " + key)
	}
	return strconv.ParseInt(s, 10, 64)
}
