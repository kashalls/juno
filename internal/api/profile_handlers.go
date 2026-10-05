package api

import (
	"net/http"

	"github.com/kashalls/juno/internal/profile"
)

type profileAPI struct {
	store *profile.Store
}

// getProfile serves the tracked user's cached Discord user object, wrapped
// as {"user": ...} to match the shape of Discord's GET /users/{id}/profile
// (and proxies of it like dcdn.dstn.to/profile/{id}), minus the
// profile-only fields bots can't read (bio, pronouns, connected accounts).
func (p *profileAPI) getProfile(w http.ResponseWriter, r *http.Request) {
	data := p.store.Get()
	if data == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Success: false, Error: "profile not yet available"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
