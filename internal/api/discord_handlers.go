package api

import (
	"encoding/json"
	"net/http"

	"github.com/kashalls/juno/internal/presence"
)

type discordAPI struct {
	store *presence.Store
}

type successResponse struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

type errorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func (d *discordAPI) getMe(w http.ResponseWriter, r *http.Request) {
	presence := d.store.Get()
	if presence == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Success: false, Error: "presence not yet available"})
		return
	}

	writeJSON(w, http.StatusOK, successResponse{Success: true, Data: presence})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
