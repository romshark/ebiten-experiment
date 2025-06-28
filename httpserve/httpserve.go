package httpserve

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"
)

//go:embed dist/*
var content embed.FS

type Server struct {
	log        *slog.Logger
	staticFS   http.Handler
	indexBytes []byte
}

func NewServer(log *slog.Logger) *Server {
	subFS, err := fs.Sub(content, "dist")
	if err != nil {
		log.Error("failed to create sub FS", slog.Any("err", err))
		os.Exit(1)
	}

	// Preload index.html into memory (optional but faster)
	indexData, err := fs.ReadFile(subFS, "index.html")
	if err != nil {
		log.Error("failed to read index.html", slog.Any("err", err))
		os.Exit(1)
	}

	mux := http.NewServeMux()
	s := &Server{
		log:        log,
		staticFS:   http.StripPrefix("/static/", http.FileServer(http.FS(subFS))),
		indexBytes: indexData,
	}

	mux.HandleFunc("/", s.handleIndex)
	mux.Handle("GET /static/", s.staticFS)

	return &Server{
		log:        log,
		staticFS:   mux,
		indexBytes: indexData,
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" || !strings.Contains(path.Base(r.URL.Path), ".") {
		w.Header().Set("Content-Type", "text/html")
		_, err := w.Write(s.indexBytes)
		if err != nil {
			s.log.Error("writing index.html", slog.Any("err", err))
		}
		return
	}

	http.NotFound(w, r)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.staticFS.ServeHTTP(w, r)
}
