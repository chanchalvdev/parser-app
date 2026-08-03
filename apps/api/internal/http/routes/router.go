package routes

import (
	"net/http"

	"github.com/enterprise-file-platform/api/internal/http/handlers"
	"github.com/enterprise-file-platform/api/internal/http/middleware"
	"github.com/go-chi/chi/v5"
)

type Router struct {
	router *chi.Mux
}

func NewRouter(
	healthHandler *handlers.HealthHandler,
	versionHandler *handlers.VersionHandler,
	docsHandler *handlers.DocsHandler,
	uploadHandler *handlers.UploadHandler,
	searchHandler *handlers.SearchHandler,
	fileHandler *handlers.FileHandler,
	jobHandler *handlers.JobHandler,
	dashboardHandler *handlers.DashboardHandler,
	adminSettingsHandler *handlers.AdminSettingsHandler,
	appMiddlewares ...func(next http.Handler) http.Handler,
) *Router {
	r := chi.NewRouter()

	r.Use(appMiddlewares...)
	r.Get("/docs", docsHandler.SwaggerUI())
	r.Get("/docs/openapi.json", docsHandler.OpenAPISpec())

	r.Get("/health", healthHandler.Health())
	r.Get("/ready", healthHandler.Ready())

	r.Route("/api/v1", func(router chi.Router) {
		router.Get("/version", versionHandler.Version())
		router.Get("/search", searchHandler.Search())
		router.Post("/search", searchHandler.Search())
		router.Get("/search/suggestions", searchHandler.Suggestions())
		router.Get("/files", fileHandler.List())
		router.Get("/files/{file_id}", fileHandler.GetByID())
		router.Post("/files/{file_id}/password", fileHandler.SubmitPassword())
		router.Get("/files/{file_id}/children", fileHandler.Children())
		router.Get("/files/{file_id}/tree", fileHandler.Tree())
		router.Get("/files/{file_id}/records", fileHandler.Records())
		router.Get("/jobs", jobHandler.List())
		router.Get("/jobs/{job_id}", jobHandler.GetByID())
		router.Get("/jobs/{job_id}/events", jobHandler.Events())
		router.Get("/jobs/{job_id}/records", jobHandler.Records())
		router.Post("/jobs/{job_id}/retry", jobHandler.Retry())
		router.Get("/dashboard/summary", dashboardHandler.Summary())
		router.Get("/dashboard/file-types", dashboardHandler.FileTypes())
		router.Get("/dashboard/processing-status", dashboardHandler.ProcessingStatus())
		router.Get("/dashboard/upload-volume", dashboardHandler.UploadVolume())
		router.Get("/dashboard/error-breakdown", dashboardHandler.ErrorBreakdown())
		router.Get("/dashboard/entities", dashboardHandler.Entities())
		router.Get("/dashboard/processing-duration", dashboardHandler.ProcessingDuration())
		router.Route("/admin", func(router chi.Router) {
			router.With(middleware.RequireRole("admin")).Get("/settings", adminSettingsHandler.Get())
			router.With(middleware.RequireRole("admin")).Put("/settings", adminSettingsHandler.Update())
		})
		router.Route("/uploads", func(r chi.Router) {
			r.Post("/initiate", uploadHandler.InitiateUpload())
			r.Post("/complete", uploadHandler.CompleteUpload())
			r.Get("/{upload_id}", uploadHandler.GetUpload())
		})
	})

	return &Router{router: r}
}

func (r *Router) Handler() http.Handler {
	return r.router
}
