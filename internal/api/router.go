package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/kashalls/juno/internal/appicons"
	"github.com/kashalls/juno/internal/lanyard"
	"github.com/kashalls/juno/internal/presence"
	"github.com/kashalls/juno/internal/profile"
)

type RouterConfig struct {
	Store             *presence.Store
	Hub               *lanyard.Hub
	ProfileStore      *profile.Store
	AppIconResolver   *appicons.Resolver
	TrustedProxyCIDRs []string
}

// NewRouter serves the Discord/Lanyard presence API: GET /api/users/me
// and the Lanyard-protocol WebSocket at /api/socket, plus GET /api/profile
// for the tracked user's cached Discord profile (bio, badges, banner, ...)
// and GET /api/app-icons/{id} to resolve any application's icon.
func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	if len(cfg.TrustedProxyCIDRs) > 0 {
		r.Use(middleware.ClientIPFromXFF(cfg.TrustedProxyCIDRs...))
	} else {
		r.Use(middleware.ClientIPFromRemoteAddr)
	}
	r.Use(skipHealthzLogger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/healthz", healthHandler)
	r.Get("/api/terms", termsHandler)
	r.Head("/api/terms", termsHandler)
	r.Get("/api/privacy", privacyHandler)
	r.Head("/api/privacy", privacyHandler)

	d := &discordAPI{store: cfg.Store}
	r.Get("/api/users/me", d.getMe)
	r.Get("/api/socket", cfg.Hub.ServeWS)

	p := &profileAPI{store: cfg.ProfileStore}
	r.Get("/api/profile", p.getProfile)

	a := &appIconAPI{resolver: cfg.AppIconResolver}
	r.Get("/api/app-icons/{id}", a.getAppIcon)

	return r
}

// skipHealthzLogger applies the slog-based request logger to every request
// except /healthz, keeping access logs free of health-check noise while
// still logging unmatched routes (404s/405s), which per-route middleware
// would miss.
func skipHealthzLogger(next http.Handler) http.Handler {
	logged := middleware.RequestLogger(slogFormatter{})(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}
		logged.ServeHTTP(w, r)
	})
}
