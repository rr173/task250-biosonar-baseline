package httpapi

import (
	"net/http"

	"task250-biosonar/internal/service"
)

// Server exposes the biosonar classification HTTP API under the /api prefix.
type Server struct {
	svc *service.Service
}

// NewServer builds an HTTP server around the given service.
func NewServer(svc *service.Service) *Server {
	return &Server{svc: svc}
}

// Handler returns the configured request multiplexer. All routes are under
// /api and use Go 1.22+ method+path patterns with {id} wildcards.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// survey batches
	mux.HandleFunc("POST /api/batches", s.createBatch)
	mux.HandleFunc("GET /api/batches", s.listBatches)
	mux.HandleFunc("GET /api/batches/{id}", s.getBatch)
	mux.HandleFunc("POST /api/batches/{id}/seal", s.sealBatch)
	mux.HandleFunc("GET /api/batches/{id}/classifications", s.listClassifications)

	// echo windows
	mux.HandleFunc("POST /api/echoes", s.ingestEcho)
	mux.HandleFunc("GET /api/echoes", s.listEchoes)
	mux.HandleFunc("GET /api/echoes/{id}", s.getEcho)
	mux.HandleFunc("POST /api/echoes/{id}/exclude", s.excludeEcho)
	mux.HandleFunc("POST /api/echoes/{id}/correct", s.correctEcho)
	mux.HandleFunc("GET /api/echoes/{id}/features", s.echoFeatures)
	mux.HandleFunc("POST /api/echoes/{id}/classify", s.classifyEcho)
	mux.HandleFunc("GET /api/echoes/{id}/classify", s.getClassification)

	// substrate catalogue
	mux.HandleFunc("POST /api/substrates", s.createSubstrate)
	mux.HandleFunc("GET /api/substrates", s.listSubstrates)

	// segment merging
	mux.HandleFunc("POST /api/segments/merge", s.mergeBatch)
	mux.HandleFunc("GET /api/segments", s.listSegments)
	mux.HandleFunc("POST /api/segments/{id}/confirm", s.confirmSegment)
	mux.HandleFunc("POST /api/segments/{id}/reject", s.rejectSegment)

	// interpretation snapshots
	mux.HandleFunc("POST /api/snapshots", s.publishSnapshot)
	mux.HandleFunc("GET /api/snapshots", s.listSnapshots)
	mux.HandleFunc("GET /api/snapshots/{id}", s.getSnapshot)

	// introspection
	mux.HandleFunc("GET /api/stats", s.stats)
	mux.HandleFunc("GET /api/health", s.health)

	return mux
}

// Listen starts the HTTP server on addr.
func (s *Server) Listen(addr string) error {
	return http.ListenAndServe(addr, s.Handler())
}
