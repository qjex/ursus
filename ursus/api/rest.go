package api

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	log "github.com/go-pkgz/lgr"
	"net/http"
	"time"
	"ursus/store"
)

type Rest struct {
	httpServer *http.Server
	public
}

func NewRest(store store.ProxyStore) *Rest {
	return &Rest{
		public: public{
			store: store,
		},
	}
}

func (s *Rest) Run(port int) {
	s.httpServer = s.makeHTTPServer(port, s.routes())
	s.httpServer.ErrorLog = log.ToStdLogger(log.Default(), "WARN")
	err := s.httpServer.ListenAndServe()
	if err != nil {
		log.Fatalf("Error starting http server: %v", err)
	}
}

func (s *Rest) routes() chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.Throttle(1000), middleware.RealIP, middleware.Recoverer)
	router.Route("/api/v1", func(rapi chi.Router) {

		rapi.Group(func(r chi.Router) {
			r.Use(middleware.Timeout(5 * time.Second))
			r.Use(middleware.NoCache)
			r.Get("/list", s.public.getProxyList)
		})
	})

	return router
}

func (s *Rest) makeHTTPServer(port int, router chi.Router) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
}
