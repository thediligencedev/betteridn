package server

import (
	"net/http"
)

func MountSwaggerDocs(mux *http.ServeMux) {
	// Serve the openapi.yml file explicitly
	mux.HandleFunc("/api/docs/openapi.yml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/openapi.yml")
	})

	// Serve the swagger-ui files at /api/docs/
	swaggerPath := "./docs/swagger-ui"
	fs := http.FileServer(http.Dir(swaggerPath))
	mux.Handle("/api/docs/", http.StripPrefix("/api/docs/", fs))
}
