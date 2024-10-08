package router

import (
	"api/pkg/handlers"
	"api/pkg/middleware"
	"net/http"
)

func Router(audience, domain string) http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("/auth/login",
		middleware.CorsMiddlewareFunc(handlers.LoginHandler),
	)
	router.HandleFunc("/auth/signup",
		middleware.CorsMiddlewareFunc(handlers.SignUpHandler),
	)
	router.HandleFunc("/auth/refresh",
		middleware.CorsMiddlewareFunc(handlers.RefreshHandler),
	)

	// tenant
	router.HandleFunc("/api/tenants", handlers.TenantsHandler)
	router.HandleFunc("/api/tenant", handlers.TenantHandler)

	// test
	router.HandleFunc("/test/api/public", handlers.PublicApiHandler)
	router.Handle("/test/api/private",
		middleware.CorsMiddleware(
			middleware.EnsureValidToken(audience, domain)(
				http.HandlerFunc(handlers.AdminApiHandler),
			),
		),
	)
	router.Handle("/test/api/admin",
		middleware.EnsureValidToken(audience, domain)(
			middleware.ValidatePermissions("read:admin-messages",
				http.HandlerFunc(handlers.ProtectedApiHandler)),
		),
	)

	return router
}
