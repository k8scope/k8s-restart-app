package api

import (
	_ "embed"
	"net/http"
)

var (
	//go:embed index.go.html
	indexTemplate []byte
)

// Index is the handler for the index page.
// This page is the main page that shows the list of applications to be managed.
func Index(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(indexTemplate)
	if err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
	}
}
