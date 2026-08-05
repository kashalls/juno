package api

import (
	"net/http"

	"github.com/kashalls/juno/internal/profile"
)

type profileAPI struct {
	store *profile.Store
}

// getProfile serves the tracked user's cached Discord profile response
// as-is, so it's a drop-in replacement for third-party proxies of
// Discord's GET /users/{id}/profile endpoint (e.g. dcdn.dstn.to/profile/{id}).
func (p *profileAPI) getProfile(w http.ResponseWriter, r *http.Request) {
	data := p.store.Get()
	if data == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Success: false, Error: "profile not yet available"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
