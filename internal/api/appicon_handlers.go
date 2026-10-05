package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/kashalls/juno/internal/appicons"
)

type appIconAPI struct {
	resolver *appicons.Resolver
}

// getAppIcon resolves a Discord application's icon via the application RPC
// endpoint and redirects to it on Discord's CDN.
func (a *appIconAPI) getAppIcon(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "id")
	if !isSnowflake(appID) {
		http.Error(w, "invalid application id", http.StatusBadRequest)
		return
	}

	url, err := a.resolver.IconURL(r.Context(), appID)
	if err != nil {
		if errors.Is(err, appicons.ErrNoIcon) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "failed to resolve application icon", http.StatusBadGateway)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func isSnowflake(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
