package http

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/ravernkoh/jabba/http/middleware"
	"github.com/sirupsen/logrus"
)

// Server serves the website.
type Server struct {
	Port     string
	Logger   logrus.FieldLogger
	Database Database
}

// Listen listens for requests, blocking until an error occurs.
func (s *Server) Listen() error {
	r := s.Router()
	s.Logger.WithFields(logrus.Fields{
		"port": s.Port,
	}).Info("http: server started listening")
	return http.ListenAndServe(s.Port, r)
}

// Router mounts all routes on a router and returns it.
func (s *Server) Router() chi.Router {
	r := chi.NewRouter()

	// Mount static assets
	fileServer(r, "/", assets)
	s.Logger.Info("http: mounted static assets")

	// Mount sample route
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		templates.ExecuteTemplate(w, "index.html", nil)
	})

	return r
}

// fileServer conveniently sets up a http.FileServer handler to serve static
// files from a http.FileSystem.
//
// https://github.com/go-chi/chi/blob/master/_examples/fileserver/main.go#L26-L44
//
// Added Neuter for directories.
func fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, middleware.Neuter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})))
}