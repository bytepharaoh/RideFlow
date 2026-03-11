package server

import (
	"encoding/json"
	"net/http"
)

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func healthHundler(serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("content-type", "application/json")

		if err := json.NewEncoder(w).Encode(healthResponse{
			Status:  "ok",
			Service: serviceName,
		}); err != nil {
			http.Error(w, "failed to encode health response", http.StatusInternalServerError)
		}
	}
}
